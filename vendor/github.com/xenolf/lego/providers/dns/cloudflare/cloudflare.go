// Package cloudflare implements a DNS provider for solving the DNS-01
// challenge using cloudflare DNS.
package cloudflare

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

// CloudFlareAPIURL represents the API endpoint to call.
// TODO: Unexport?
const CloudFlareAPIURL = "https://api.cloudflare.com/client/v4"

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	authEmail string
	authKey   string
	client    *http.Client
}

// NewDNSProvider returns a DNSProvider instance configured for cloudflare.
// Credentials must be passed in the environment variables: CLOUDFLARE_EMAIL
// and CLOUDFLARE_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("CLOUDFLARE_EMAIL", "CLOUDFLARE_API_KEY")
	if err != nil {
		return nil, fmt.Errorf("CloudFlare: %v", err)
	}

	return NewDNSProviderCredentials(values["CLOUDFLARE_EMAIL"], values["CLOUDFLARE_API_KEY"])
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for cloudflare.
func NewDNSProviderCredentials(email, key string) (*DNSProvider, error) {
	if email == "" || key == "" {
		return nil, errors.New("CloudFlare: some credentials information are missing")
	}

	return &DNSProvider{
		authEmail: email,
		authKey:   key,
		client:    &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS
// propagation. Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return 120 * time.Second, 2 * time.Second
}

// Present creates a TXT record to fulfil the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)
	zoneID, err := d.getHostedZoneID(fqdn)
	if err != nil {
		return err
	}

	rec := cloudFlareRecord{
		Type:    "TXT",
		Name:    acme.UnFqdn(fqdn),
		Content: value,
		TTL:     ttl,
	}

	body, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	_, err = d.doRequest(http.MethodPost, fmt.Sprintf("/zones/%s/dns_records", zoneID), bytes.NewReader(body))
	return err
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	record, err := d.findTxtRecord(fqdn)
	if err != nil {
		return err
	}

	_, err = d.doRequest(http.MethodDelete, fmt.Sprintf("/zones/%s/dns_records/%s", record.ZoneID, record.ID), nil)
	return err
}

func (d *DNSProvider) getHostedZoneID(fqdn string) (string, error) {
	// HostedZone represents a CloudFlare DNS zone
	type HostedZone struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return "", err
	}

	result, err := d.doRequest(http.MethodGet, "/zones?name="+acme.UnFqdn(authZone), nil)
	if err != nil {
		return "", err
	}

	var hostedZone []HostedZone
	err = json.Unmarshal(result, &hostedZone)
	if err != nil {
		return "", err
	}

	if len(hostedZone) != 1 {
		return "", fmt.Errorf("zone %s not found in CloudFlare for domain %s", authZone, fqdn)
	}

	return hostedZone[0].ID, nil
}

func (d *DNSProvider) findTxtRecord(fqdn string) (*cloudFlareRecord, error) {
	zoneID, err := d.getHostedZoneID(fqdn)
	if err != nil {
		return nil, err
	}

	result, err := d.doRequest(
		http.MethodGet,
		fmt.Sprintf("/zones/%s/dns_records?per_page=1000&type=TXT&name=%s", zoneID, acme.UnFqdn(fqdn)),
		nil,
	)
	if err != nil {
		return nil, err
	}

	var records []cloudFlareRecord
	err = json.Unmarshal(result, &records)
	if err != nil {
		return nil, err
	}

	for _, rec := range records {
		if rec.Name == acme.UnFqdn(fqdn) {
			return &rec, nil
		}
	}

	return nil, fmt.Errorf("no existing record found for %s", fqdn)
}

func (d *DNSProvider) doRequest(method, uri string, body io.Reader) (json.RawMessage, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", CloudFlareAPIURL, uri), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Auth-Email", d.authEmail)
	req.Header.Set("X-Auth-Key", d.authKey)

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error querying Cloudflare API -> %v", err)
	}

	defer resp.Body.Close()

	var r APIResponse
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, err
	}

	if !r.Success {
		if len(r.Errors) > 0 {
			errStr := ""
			for _, apiErr := range r.Errors {
				errStr += fmt.Sprintf("\t Error: %d: %s", apiErr.Code, apiErr.Message)
				for _, chainErr := range apiErr.ErrorChain {
					errStr += fmt.Sprintf("<- %d: %s", chainErr.Code, chainErr.Message)
				}
			}
			return nil, fmt.Errorf("Cloudflare API Error \n%s", errStr)
		}
		strBody := "Unreadable body"
		if body, err := ioutil.ReadAll(resp.Body); err == nil {
			strBody = string(body)
		}
		return nil, fmt.Errorf("Cloudflare API error: the request %s sent a response with a body which is not in JSON format: %s", req.URL.String(), strBody)
	}

	return r.Result, nil
}

// APIError contains error details for failed requests
type APIError struct {
	Code       int        `json:"code,omitempty"`
	Message    string     `json:"message,omitempty"`
	ErrorChain []APIError `json:"error_chain,omitempty"`
}

// APIResponse represents a response from CloudFlare API
type APIResponse struct {
	Success bool            `json:"success"`
	Errors  []*APIError     `json:"errors"`
	Result  json.RawMessage `json:"result"`
}

// cloudFlareRecord represents a CloudFlare DNS record
type cloudFlareRecord struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	ID      string `json:"id,omitempty"`
	TTL     int    `json:"ttl,omitempty"`
	ZoneID  string `json:"zone_id,omitempty"`
}
