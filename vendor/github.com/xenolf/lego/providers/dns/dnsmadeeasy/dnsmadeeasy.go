package dnsmadeeasy

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/xenolf/lego/acme"
)

// DNSProvider is an implementation of the acme.ChallengeProvider interface that uses
// DNSMadeEasy's DNS API to manage TXT records for a domain.
type DNSProvider struct {
	baseURL   string
	apiKey    string
	apiSecret string
}

// Domain holds the DNSMadeEasy API representation of a Domain
type Domain struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Record holds the DNSMadeEasy API representation of a Domain Record
type Record struct {
	ID       int    `json:"id"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Value    string `json:"value"`
	TTL      int    `json:"ttl"`
	SourceID int    `json:"sourceId"`
}

// NewDNSProvider returns a DNSProvider instance configured for DNSMadeEasy DNS.
// Credentials must be passed in the environment variables: DNSMADEEASY_API_KEY
// and DNSMADEEASY_API_SECRET.
func NewDNSProvider() (*DNSProvider, error) {
	dnsmadeeasyAPIKey := os.Getenv("DNSMADEEASY_API_KEY")
	dnsmadeeasyAPISecret := os.Getenv("DNSMADEEASY_API_SECRET")
	dnsmadeeasySandbox := os.Getenv("DNSMADEEASY_SANDBOX")

	var baseURL string

	sandbox, _ := strconv.ParseBool(dnsmadeeasySandbox)
	if sandbox {
		baseURL = "https://api.sandbox.dnsmadeeasy.com/V2.0"
	} else {
		baseURL = "https://api.dnsmadeeasy.com/V2.0"
	}

	return NewDNSProviderCredentials(baseURL, dnsmadeeasyAPIKey, dnsmadeeasyAPISecret)
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for DNSMadeEasy.
func NewDNSProviderCredentials(baseURL, apiKey, apiSecret string) (*DNSProvider, error) {
	if baseURL == "" || apiKey == "" || apiSecret == "" {
		return nil, fmt.Errorf("DNS Made Easy credentials missing")
	}

	return &DNSProvider{
		baseURL:   baseURL,
		apiKey:    apiKey,
		apiSecret: apiSecret,
	}, nil
}

// Present creates a TXT record using the specified parameters
func (d *DNSProvider) Present(domainName, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domainName, keyAuth)

	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return err
	}

	// fetch the domain details
	domain, err := d.getDomain(authZone)
	if err != nil {
		return err
	}

	// create the TXT record
	name := strings.Replace(fqdn, "."+authZone, "", 1)
	record := &Record{Type: "TXT", Name: name, Value: value, TTL: ttl}

	err = d.createRecord(domain, record)
	if err != nil {
		return err
	}

	return nil
}

// CleanUp removes the TXT records matching the specified parameters
func (d *DNSProvider) CleanUp(domainName, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domainName, keyAuth)

	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return err
	}

	// fetch the domain details
	domain, err := d.getDomain(authZone)
	if err != nil {
		return err
	}

	// find matching records
	name := strings.Replace(fqdn, "."+authZone, "", 1)
	records, err := d.getRecords(domain, name, "TXT")
	if err != nil {
		return err
	}

	// delete records
	for _, record := range *records {
		err = d.deleteRecord(record)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DNSProvider) getDomain(authZone string) (*Domain, error) {
	domainName := authZone[0 : len(authZone)-1]
	resource := fmt.Sprintf("%s%s", "/dns/managed/name?domainname=", domainName)

	resp, err := d.sendRequest("GET", resource, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	domain := &Domain{}
	err = json.NewDecoder(resp.Body).Decode(&domain)
	if err != nil {
		return nil, err
	}

	return domain, nil
}

func (d *DNSProvider) getRecords(domain *Domain, recordName, recordType string) (*[]Record, error) {
	resource := fmt.Sprintf("%s/%d/%s%s%s%s", "/dns/managed", domain.ID, "records?recordName=", recordName, "&type=", recordType)

	resp, err := d.sendRequest("GET", resource, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	type recordsResponse struct {
		Records *[]Record `json:"data"`
	}

	records := &recordsResponse{}
	err = json.NewDecoder(resp.Body).Decode(&records)
	if err != nil {
		return nil, err
	}

	return records.Records, nil
}

func (d *DNSProvider) createRecord(domain *Domain, record *Record) error {
	url := fmt.Sprintf("%s/%d/%s", "/dns/managed", domain.ID, "records")

	resp, err := d.sendRequest("POST", url, record)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (d *DNSProvider) deleteRecord(record Record) error {
	resource := fmt.Sprintf("%s/%d/%s/%d", "/dns/managed", record.SourceID, "records", record.ID)

	resp, err := d.sendRequest("DELETE", resource, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (d *DNSProvider) sendRequest(method, resource string, payload interface{}) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", d.baseURL, resource)

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	timestamp := time.Now().UTC().Format(time.RFC1123)
	signature := computeHMAC(timestamp, d.apiSecret)

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-dnsme-apiKey", d.apiKey)
	req.Header.Set("x-dnsme-requestDate", timestamp)
	req.Header.Set("x-dnsme-hmac", signature)
	req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "application/json")

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(10 * time.Second),
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("DNSMadeEasy API request failed with HTTP status code %d", resp.StatusCode)
	}

	return resp, nil
}

func computeHMAC(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha1.New, key)
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}
