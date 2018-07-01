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
	"strconv"
	"strings"
	"time"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/log"
	"github.com/xenolf/lego/platform/config/env"
)

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	apiKey     string
	host       *url.URL
	apiVersion int
	client     *http.Client
}

// NewDNSProvider returns a DNSProvider instance configured for pdns.
// Credentials must be passed in the environment variable:
// PDNS_API_URL and PDNS_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("PDNS_API_KEY", "PDNS_API_URL")
	if err != nil {
		return nil, fmt.Errorf("PDNS: %v", err)
	}

	hostURL, err := url.Parse(values["PDNS_API_URL"])
	if err != nil {
		return nil, fmt.Errorf("PDNS: %v", err)
	}

	return NewDNSProviderCredentials(hostURL, values["PDNS_API_KEY"])
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

	d := &DNSProvider{
		host:   host,
		apiKey: key,
		client: &http.Client{Timeout: 30 * time.Second},
	}

	apiVersion, err := d.getAPIVersion()
	if err != nil {
		log.Warnf("PDNS: failed to get API version %v", err)
	}
	d.apiVersion = apiVersion

	return d, nil
}

// Timeout returns the timeout and interval to use when checking for DNS
// propagation. Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return 120 * time.Second, 2 * time.Second
}

// Present creates a TXT record to fulfil the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)
	zone, err := d.getHostedZone(fqdn)
	if err != nil {
		return err
	}

	name := fqdn

	// pre-v1 API wants non-fqdn
	if d.apiVersion == 0 {
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
			{
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

	_, err = d.makeRequest(http.MethodPatch, zone.URL, bytes.NewReader(body))
	return err
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	zone, err := d.getHostedZone(fqdn)
	if err != nil {
		return err
	}

	set, err := d.findTxtRecord(fqdn)
	if err != nil {
		return err
	}

	rrsets := rrSets{
		RRSets: []rrSet{
			{
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

	_, err = d.makeRequest(http.MethodPatch, zone.URL, bytes.NewReader(body))
	return err
}

func (d *DNSProvider) getHostedZone(fqdn string) (*hostedZone, error) {
	var zone hostedZone
	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return nil, err
	}

	url := "/servers/localhost/zones"
	result, err := d.makeRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var zones []hostedZone
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

	result, err = d.makeRequest(http.MethodGet, url, nil)
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

func (d *DNSProvider) findTxtRecord(fqdn string) (*rrSet, error) {
	zone, err := d.getHostedZone(fqdn)
	if err != nil {
		return nil, err
	}

	_, err = d.makeRequest(http.MethodGet, zone.URL, nil)
	if err != nil {
		return nil, err
	}

	for _, set := range zone.RRSets {
		if (set.Name == acme.UnFqdn(fqdn) || set.Name == fqdn) && set.Type == "TXT" {
			return &set, nil
		}
	}

	return nil, fmt.Errorf("no existing record found for %s", fqdn)
}

func (d *DNSProvider) getAPIVersion() (int, error) {
	type APIVersion struct {
		URL     string `json:"url"`
		Version int    `json:"version"`
	}

	result, err := d.makeRequest(http.MethodGet, "/api", nil)
	if err != nil {
		return 0, err
	}

	var versions []APIVersion
	err = json.Unmarshal(result, &versions)
	if err != nil {
		return 0, err
	}

	latestVersion := 0
	for _, v := range versions {
		if v.Version > latestVersion {
			latestVersion = v.Version
		}
	}

	return latestVersion, err
}

func (d *DNSProvider) makeRequest(method, uri string, body io.Reader) (json.RawMessage, error) {
	type APIError struct {
		Error string `json:"error"`
	}

	var path = ""
	if d.host.Path != "/" {
		path = d.host.Path
	}

	if !strings.HasPrefix(uri, "/") {
		uri = "/" + uri
	}

	if d.apiVersion > 0 && !strings.HasPrefix(uri, "/api/v") {
		uri = "/api/v" + strconv.Itoa(d.apiVersion) + uri
	}

	url := d.host.Scheme + "://" + d.host.Host + path + uri
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", d.apiKey)

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error talking to PDNS API -> %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity && (resp.StatusCode < 200 || resp.StatusCode >= 300) {
		return nil, fmt.Errorf("unexpected HTTP status code %d when fetching '%s'", resp.StatusCode, url)
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
			return nil, fmt.Errorf("error talking to PDNS API -> %v", apiError.Error)
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
