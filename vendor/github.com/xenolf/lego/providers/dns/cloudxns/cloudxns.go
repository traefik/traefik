// Package cloudxns implements a DNS provider for solving the DNS-01 challenge
// using cloudxns DNS.
package cloudxns

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/xenolf/lego/acme"
)

const cloudXNSBaseURL = "https://www.cloudxns.net/api2/"

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	apiKey    string
	secretKey string
}

// NewDNSProvider returns a DNSProvider instance configured for cloudxns.
// Credentials must be passed in the environment variables: CLOUDXNS_API_KEY
// and CLOUDXNS_SECRET_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	apiKey := os.Getenv("CLOUDXNS_API_KEY")
	secretKey := os.Getenv("CLOUDXNS_SECRET_KEY")
	return NewDNSProviderCredentials(apiKey, secretKey)
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for cloudxns.
func NewDNSProviderCredentials(apiKey, secretKey string) (*DNSProvider, error) {
	if apiKey == "" || secretKey == "" {
		return nil, fmt.Errorf("CloudXNS credentials missing")
	}

	return &DNSProvider{
		apiKey:    apiKey,
		secretKey: secretKey,
	}, nil
}

// Present creates a TXT record to fulfil the dns-01 challenge.
func (c *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)
	zoneID, err := c.getHostedZoneID(fqdn)
	if err != nil {
		return err
	}

	return c.addTxtRecord(zoneID, fqdn, value, ttl)
}

// CleanUp removes the TXT record matching the specified parameters.
func (c *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)
	zoneID, err := c.getHostedZoneID(fqdn)
	if err != nil {
		return err
	}

	recordID, err := c.findTxtRecord(zoneID, fqdn)
	if err != nil {
		return err
	}

	return c.delTxtRecord(recordID, zoneID)
}

func (c *DNSProvider) getHostedZoneID(fqdn string) (string, error) {
	type Data struct {
		ID     string `json:"id"`
		Domain string `json:"domain"`
	}

	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return "", err
	}

	result, err := c.makeRequest("GET", "domain", nil)
	if err != nil {
		return "", err
	}

	var domains []Data
	err = json.Unmarshal(result, &domains)
	if err != nil {
		return "", err
	}

	for _, data := range domains {
		if data.Domain == authZone {
			return data.ID, nil
		}
	}

	return "", fmt.Errorf("zone %s not found in cloudxns for domain %s", authZone, fqdn)
}

func (c *DNSProvider) findTxtRecord(zoneID, fqdn string) (string, error) {
	result, err := c.makeRequest("GET", fmt.Sprintf("record/%s?host_id=0&offset=0&row_num=2000", zoneID), nil)
	if err != nil {
		return "", err
	}

	var records []cloudXNSRecord
	err = json.Unmarshal(result, &records)
	if err != nil {
		return "", err
	}

	for _, record := range records {
		if record.Host == acme.UnFqdn(fqdn) && record.Type == "TXT" {
			return record.RecordID, nil
		}
	}

	return "", fmt.Errorf("no existing record found for %s", fqdn)
}

func (c *DNSProvider) addTxtRecord(zoneID, fqdn, value string, ttl int) error {
	id, err := strconv.Atoi(zoneID)
	if err != nil {
		return err
	}

	payload := cloudXNSRecord{
		ID:     id,
		Host:   acme.UnFqdn(fqdn),
		Value:  value,
		Type:   "TXT",
		LineID: 1,
		TTL:    ttl,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = c.makeRequest("POST", "record", body)
	return err
}

func (c *DNSProvider) delTxtRecord(recordID, zoneID string) error {
	_, err := c.makeRequest("DELETE", fmt.Sprintf("record/%s/%s", recordID, zoneID), nil)
	return err
}

func (c *DNSProvider) hmac(url, date, body string) string {
	sum := md5.Sum([]byte(c.apiKey + url + body + date + c.secretKey))
	return hex.EncodeToString(sum[:])
}

func (c *DNSProvider) makeRequest(method, uri string, body []byte) (json.RawMessage, error) {
	type APIResponse struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data,omitempty"`
	}

	url := cloudXNSBaseURL + uri
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	requestDate := time.Now().Format(time.RFC1123Z)

	req.Header.Set("API-KEY", c.apiKey)
	req.Header.Set("API-REQUEST-DATE", requestDate)
	req.Header.Set("API-HMAC", c.hmac(url, requestDate, string(body)))
	req.Header.Set("API-FORMAT", "json")

	resp, err := acme.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var r APIResponse
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, err
	}

	if r.Code != 1 {
		return nil, fmt.Errorf("CloudXNS API Error: %s", r.Message)
	}
	return r.Data, nil
}

type cloudXNSRecord struct {
	ID       int    `json:"domain_id,omitempty"`
	RecordID string `json:"record_id,omitempty"`

	Host   string `json:"host"`
	Value  string `json:"value"`
	Type   string `json:"type"`
	LineID int    `json:"line_id,string"`
	TTL    int    `json:"ttl,string"`
}
