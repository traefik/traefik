// Package pdns implements a DNS provider for solving the DNS-01
// challenge using PowerDNS nameserver.
package pdns

import (
	"bytes"
	"encoding/json"
	"errors"
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

// Config is used to configure the creation of the DNSProvider
type Config struct {
	APIKey             string
	Host               *url.URL
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("PDNS_TTL", 120),
		PropagationTimeout: env.GetOrDefaultSecond("PDNS_PROPAGATION_TIMEOUT", 120*time.Second),
		PollingInterval:    env.GetOrDefaultSecond("PDNS_POLLING_INTERVAL", 2*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("PDNS_HTTP_TIMEOUT", 30*time.Second),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	apiVersion int
	config     *Config
}

// NewDNSProvider returns a DNSProvider instance configured for pdns.
// Credentials must be passed in the environment variable:
// PDNS_API_URL and PDNS_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("PDNS_API_KEY", "PDNS_API_URL")
	if err != nil {
		return nil, fmt.Errorf("pdns: %v", err)
	}

	hostURL, err := url.Parse(values["PDNS_API_URL"])
	if err != nil {
		return nil, fmt.Errorf("pdns: %v", err)
	}

	config := NewDefaultConfig()
	config.Host = hostURL
	config.APIKey = values["PDNS_API_KEY"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderCredentials uses the supplied credentials
// to return a DNSProvider instance configured for pdns.
// Deprecated
func NewDNSProviderCredentials(host *url.URL, key string) (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.Host = host
	config.APIKey = key

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for pdns.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("pdns: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("pdns: API key missing")
	}

	if config.Host == nil || config.Host.Host == "" {
		return nil, fmt.Errorf("pdns: API URL missing")
	}

	d := &DNSProvider{config: config}

	apiVersion, err := d.getAPIVersion()
	if err != nil {
		log.Warnf("pdns: failed to get API version %v", err)
	}
	d.apiVersion = apiVersion

	return d, nil
}

// Timeout returns the timeout and interval to use when checking for DNS
// propagation. Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfil the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)
	zone, err := d.getHostedZone(fqdn)
	if err != nil {
		return fmt.Errorf("pdns: %v", err)
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
		TTL:  d.config.TTL,
	}

	rrsets := rrSets{
		RRSets: []rrSet{
			{
				Name:       name,
				ChangeType: "REPLACE",
				Type:       "TXT",
				Kind:       "Master",
				TTL:        d.config.TTL,
				Records:    []pdnsRecord{rec},
			},
		},
	}

	body, err := json.Marshal(rrsets)
	if err != nil {
		return fmt.Errorf("pdns: %v", err)
	}

	_, err = d.makeRequest(http.MethodPatch, zone.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("pdns: %v", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	zone, err := d.getHostedZone(fqdn)
	if err != nil {
		return fmt.Errorf("pdns: %v", err)
	}

	set, err := d.findTxtRecord(fqdn)
	if err != nil {
		return fmt.Errorf("pdns: %v", err)
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
		return fmt.Errorf("pdns: %v", err)
	}

	_, err = d.makeRequest(http.MethodPatch, zone.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("pdns: %v", err)
	}
	return nil
}

func (d *DNSProvider) getHostedZone(fqdn string) (*hostedZone, error) {
	var zone hostedZone
	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return nil, err
	}

	u := "/servers/localhost/zones"
	result, err := d.makeRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	var zones []hostedZone
	err = json.Unmarshal(result, &zones)
	if err != nil {
		return nil, err
	}

	u = ""
	for _, zone := range zones {
		if acme.UnFqdn(zone.Name) == acme.UnFqdn(authZone) {
			u = zone.URL
		}
	}

	result, err = d.makeRequest(http.MethodGet, u, nil)
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
	if d.config.Host.Path != "/" {
		path = d.config.Host.Path
	}

	if !strings.HasPrefix(uri, "/") {
		uri = "/" + uri
	}

	if d.apiVersion > 0 && !strings.HasPrefix(uri, "/api/v") {
		uri = "/api/v" + strconv.Itoa(d.apiVersion) + uri
	}

	u := d.config.Host.Scheme + "://" + d.config.Host.Host + path + uri
	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", d.config.APIKey)

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error talking to PDNS API -> %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity && (resp.StatusCode < 200 || resp.StatusCode >= 300) {
		return nil, fmt.Errorf("unexpected HTTP status code %d when fetching '%s'", resp.StatusCode, u)
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
