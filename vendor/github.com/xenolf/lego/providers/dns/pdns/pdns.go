// Package pdns implements a DNS provider for solving the DNS-01
// challenge using PowerDNS nameserver.
package pdns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/xenolf/lego/acme"
)

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	apiKey     string
	host       *url.URL
	apiVersion int
}

// NewDNSProvider returns a DNSProvider instance configured for pdns.
// Credentials must be passed in the environment variable:
// PDNS_API_URL and PDNS_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	key := os.Getenv("PDNS_API_KEY")
	hostUrl, err := url.Parse(os.Getenv("PDNS_API_URL"))
	if err != nil {
		return nil, err
	}

	return NewDNSProviderCredentials(hostUrl, key)
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for pdns.
func NewDNSProviderCredentials(host *url.URL, key string) (*DNSProvider, error) {
	if key == "" {
		return nil, fmt.Errorf("PDNS API key missing")
	}

	if host == nil || host.Host == "" {
		return nil, fmt.Errorf("PDNS API URL missing")
	}

	provider := &DNSProvider{
		host:   host,
		apiKey: key,
	}
	provider.getAPIVersion()

	return provider, nil
}

// Timeout returns the timeout and interval to use when checking for DNS
// propagation. Adjusting here to cope with spikes in propagation times.
func (c *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return 120 * time.Second, 2 * time.Second
}

// Present creates a TXT record to fulfil the dns-01 challenge
func (c *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)
	zone, err := c.getHostedZone(fqdn)
	if err != nil {
		return err
	}

	name := fqdn

	// pre-v1 API wants non-fqdn
	if c.apiVersion == 0 {
		name = acme.UnFqdn(fqdn)
	}

	rec := pdnsRecord{
		Content:  "\"" + value + "\"",
		Disabled: false,

		// pre-v1 API
		Type: "TXT",
		Name: name,
		TTL:  120,
	}

	rrsets := rrSets{
		RRSets: []rrSet{
			rrSet{
				Name:       name,
				ChangeType: "REPLACE",
				Type:       "TXT",
				Kind:       "Master",
				TTL:        120,
				Records:    []pdnsRecord{rec},
			},
		},
	}

	body, err := json.Marshal(rrsets)
	if err != nil {
		return err
	}

	_, err = c.makeRequest("PATCH", zone.URL, bytes.NewReader(body))
	if err != nil {
		fmt.Println("here")
		return err
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (c *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	zone, err := c.getHostedZone(fqdn)
	if err != nil {
		return err
	}

	set, err := c.findTxtRecord(fqdn)
	if err != nil {
		return err
	}

	rrsets := rrSets{
		RRSets: []rrSet{
			rrSet{
				Name:       set.Name,
				Type:       set.Type,
				ChangeType: "DELETE",
			},
		},
	}
	body, err := json.Marshal(rrsets)
	if err != nil {
		return err
	}

	_, err = c.makeRequest("PATCH", zone.URL, bytes.NewReader(body))
	if err != nil {
		return err
	}

	return nil
}

func (c *DNSProvider) getHostedZone(fqdn string) (*hostedZone, error) {
	var zone hostedZone
	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return nil, err
	}

	url := "/servers/localhost/zones"
	result, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	zones := []hostedZone{}
	err = json.Unmarshal(result, &zones)
	if err != nil {
		return nil, err
	}

	url = ""
	for _, zone := range zones {
		if acme.UnFqdn(zone.Name) == acme.UnFqdn(authZone) {
			url = zone.URL
		}
	}

	result, err = c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(result, &zone)
	if err != nil {
		return nil, err
	}

	// convert pre-v1 API result
	if len(zone.Records) > 0 {
		zone.RRSets = []rrSet{}
		for _, record := range zone.Records {
			set := rrSet{
				Name:    record.Name,
				Type:    record.Type,
				Records: []pdnsRecord{record},
			}
			zone.RRSets = append(zone.RRSets, set)
		}
	}

	return &zone, nil
}

func (c *DNSProvider) findTxtRecord(fqdn string) (*rrSet, error) {
	zone, err := c.getHostedZone(fqdn)
	if err != nil {
		return nil, err
	}

	_, err = c.makeRequest("GET", zone.URL, nil)
	if err != nil {
		return nil, err
	}

	for _, set := range zone.RRSets {
		if (set.Name == acme.UnFqdn(fqdn) || set.Name == fqdn) && set.Type == "TXT" {
			return &set, nil
		}
	}

	return nil, fmt.Errorf("No existing record found for %s", fqdn)
}

func (c *DNSProvider) getAPIVersion() {
	type APIVersion struct {
		URL     string `json:"url"`
		Version int    `json:"version"`
	}

	result, err := c.makeRequest("GET", "/api", nil)
	if err != nil {
		return
	}

	var versions []APIVersion
	err = json.Unmarshal(result, &versions)
	if err != nil {
		return
	}

	latestVersion := 0
	for _, v := range versions {
		if v.Version > latestVersion {
			latestVersion = v.Version
		}
	}
	c.apiVersion = latestVersion
}

func (c *DNSProvider) makeRequest(method, uri string, body io.Reader) (json.RawMessage, error) {
	type APIError struct {
		Error string `json:"error"`
	}
	var path = ""
	if c.host.Path != "/" {
		path = c.host.Path
	}
	if c.apiVersion > 0 {
		if !strings.HasPrefix(uri, "api/v") {
			uri = "/api/v" + strconv.Itoa(c.apiVersion) + uri
		} else {
			uri = "/" + uri
		}
	}
	url := c.host.Scheme + "://" + c.host.Host + path + uri
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", c.apiKey)

	client := http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error talking to PDNS API -> %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 422 && (resp.StatusCode < 200 || resp.StatusCode >= 300) {
		return nil, fmt.Errorf("Unexpected HTTP status code %d when fetching '%s'", resp.StatusCode, url)
	}

	var msg json.RawMessage
	err = json.NewDecoder(resp.Body).Decode(&msg)
	switch {
	case err == io.EOF:
		// empty body
		return nil, nil
	case err != nil:
		// other error
		return nil, err
	}

	// check for PowerDNS error message
	if len(msg) > 0 && msg[0] == '{' {
		var apiError APIError
		err = json.Unmarshal(msg, &apiError)
		if err != nil {
			return nil, err
		}
		if apiError.Error != "" {
			return nil, fmt.Errorf("Error talking to PDNS API -> %v", apiError.Error)
		}
	}
	return msg, nil
}

type pdnsRecord struct {
	Content  string `json:"content"`
	Disabled bool   `json:"disabled"`

	// pre-v1 API
	Name string `json:"name"`
	Type string `json:"type"`
	TTL  int    `json:"ttl,omitempty"`
}

type hostedZone struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	URL    string  `json:"url"`
	RRSets []rrSet `json:"rrsets"`

	// pre-v1 API
	Records []pdnsRecord `json:"records"`
}

type rrSet struct {
	Name       string       `json:"name"`
	Type       string       `json:"type"`
	Kind       string       `json:"kind"`
	ChangeType string       `json:"changetype"`
	Records    []pdnsRecord `json:"records"`
	TTL        int          `json:"ttl,omitempty"`
}

type rrSets struct {
	RRSets []rrSet `json:"rrsets"`
}
