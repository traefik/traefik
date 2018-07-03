// Package godaddy implements a DNS provider for solving the DNS-01 challenge using godaddy DNS.
package godaddy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

// GoDaddyAPIURL represents the API endpoint to call.
const apiURL = "https://api.godaddy.com"

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	apiKey    string
	apiSecret string
	client    *http.Client
}

// NewDNSProvider returns a DNSProvider instance configured for godaddy.
// Credentials must be passed in the environment variables: GODADDY_API_KEY
// and GODADDY_API_SECRET.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("GODADDY_API_KEY", "GODADDY_API_SECRET")
	if err != nil {
		return nil, fmt.Errorf("GoDaddy: %v", err)
	}

	return NewDNSProviderCredentials(values["GODADDY_API_KEY"], values["GODADDY_API_SECRET"])
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for godaddy.
func NewDNSProviderCredentials(apiKey, apiSecret string) (*DNSProvider, error) {
	if apiKey == "" || apiSecret == "" {
		return nil, fmt.Errorf("GoDaddy credentials missing")
	}

	return &DNSProvider{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		client:    &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS
// propagation. Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return 120 * time.Second, 2 * time.Second
}

func (d *DNSProvider) extractRecordName(fqdn, domain string) string {
	name := acme.UnFqdn(fqdn)
	if idx := strings.Index(name, "."+domain); idx != -1 {
		return name[:idx]
	}
	return name
}

// Present creates a TXT record to fulfil the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)
	domainZone, err := d.getZone(fqdn)
	if err != nil {
		return err
	}

	if ttl < 600 {
		ttl = 600
	}

	recordName := d.extractRecordName(fqdn, domainZone)
	rec := []DNSRecord{
		{
			Type: "TXT",
			Name: recordName,
			Data: value,
			TTL:  ttl,
		},
	}

	return d.updateRecords(rec, domainZone, recordName)
}

func (d *DNSProvider) updateRecords(records []DNSRecord, domainZone string, recordName string) error {
	body, err := json.Marshal(records)
	if err != nil {
		return err
	}

	var resp *http.Response
	resp, err = d.makeRequest(http.MethodPut, fmt.Sprintf("/v1/domains/%s/records/TXT/%s", domainZone, recordName), bytes.NewReader(body))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("could not create record %v; Status: %v; Body: %s", string(body), resp.StatusCode, string(bodyBytes))
	}
	return nil
}

// CleanUp sets null value in the TXT DNS record as GoDaddy has no proper DELETE record method
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)
	domainZone, err := d.getZone(fqdn)
	if err != nil {
		return err
	}

	recordName := d.extractRecordName(fqdn, domainZone)
	rec := []DNSRecord{
		{
			Type: "TXT",
			Name: recordName,
			Data: "null",
		},
	}

	return d.updateRecords(rec, domainZone, recordName)
}

func (d *DNSProvider) getZone(fqdn string) (string, error) {
	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return "", err
	}

	return acme.UnFqdn(authZone), nil
}

func (d *DNSProvider) makeRequest(method, uri string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", apiURL, uri), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("sso-key %s:%s", d.apiKey, d.apiSecret))

	return d.client.Do(req)
}

// DNSRecord a DNS record
type DNSRecord struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Data     string `json:"data"`
	Priority int    `json:"priority,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
}
