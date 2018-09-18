// Package rackspace implements a DNS provider for solving the DNS-01
// challenge using rackspace DNS.
package rackspace

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

// defaultBaseURL represents the Identity API endpoint to call
const defaultBaseURL = "https://identity.api.rackspacecloud.com/v2.0/tokens"

// Config is used to configure the creation of the DNSProvider
type Config struct {
	BaseURL            string
	APIUser            string
	APIKey             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		BaseURL:            defaultBaseURL,
		TTL:                env.GetOrDefaultInt("RACKSPACE_TTL", 300),
		PropagationTimeout: env.GetOrDefaultSecond("RACKSPACE_PROPAGATION_TIMEOUT", acme.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("RACKSPACE_POLLING_INTERVAL", acme.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("RACKSPACE_HTTP_TIMEOUT", 30*time.Second),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface
// used to store the reusable token and DNS API endpoint
type DNSProvider struct {
	config           *Config
	token            string
	cloudDNSEndpoint string
}

// NewDNSProvider returns a DNSProvider instance configured for Rackspace.
// Credentials must be passed in the environment variables:
// RACKSPACE_USER and RACKSPACE_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("RACKSPACE_USER", "RACKSPACE_API_KEY")
	if err != nil {
		return nil, fmt.Errorf("rackspace: %v", err)
	}

	config := NewDefaultConfig()
	config.APIUser = values["RACKSPACE_USER"]
	config.APIKey = values["RACKSPACE_API_KEY"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderCredentials uses the supplied credentials
// to return a DNSProvider instance configured for Rackspace.
// It authenticates against the API, also grabbing the DNS Endpoint.
// Deprecated
func NewDNSProviderCredentials(user, key string) (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.APIUser = user
	config.APIKey = key

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Rackspace.
// It authenticates against the API, also grabbing the DNS Endpoint.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("rackspace: the configuration of the DNS provider is nil")
	}

	if config.APIUser == "" || config.APIKey == "" {
		return nil, fmt.Errorf("rackspace: credentials missing")
	}

	authData := AuthData{
		Auth: Auth{
			APIKeyCredentials: APIKeyCredentials{
				Username: config.APIUser,
				APIKey:   config.APIKey,
			},
		},
	}

	body, err := json.Marshal(authData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, config.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// client := &http.Client{Timeout: 30 * time.Second}
	resp, err := config.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("rackspace: error querying Identity API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rackspace: authentication failed: response code: %d", resp.StatusCode)
	}

	var identity Identity
	err = json.NewDecoder(resp.Body).Decode(&identity)
	if err != nil {
		return nil, fmt.Errorf("rackspace: %v", err)
	}

	// Iterate through the Service Catalog to get the DNS Endpoint
	var dnsEndpoint string
	for _, service := range identity.Access.ServiceCatalog {
		if service.Name == "cloudDNS" {
			dnsEndpoint = service.Endpoints[0].PublicURL
			break
		}
	}
	if dnsEndpoint == "" {
		return nil, fmt.Errorf("rackspace: failed to populate DNS endpoint, check Rackspace API for changes")
	}

	return &DNSProvider{
		config:           config,
		token:            identity.Access.Token.ID,
		cloudDNSEndpoint: dnsEndpoint,
	}, nil

}

// Present creates a TXT record to fulfil the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)
	zoneID, err := d.getHostedZoneID(fqdn)
	if err != nil {
		return fmt.Errorf("rackspace: %v", err)
	}

	rec := Records{
		Record: []Record{{
			Name: acme.UnFqdn(fqdn),
			Type: "TXT",
			Data: value,
			TTL:  d.config.TTL,
		}},
	}

	body, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("rackspace: %v", err)
	}

	_, err = d.makeRequest(http.MethodPost, fmt.Sprintf("/domains/%d/records", zoneID), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("rackspace: %v", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)
	zoneID, err := d.getHostedZoneID(fqdn)
	if err != nil {
		return fmt.Errorf("rackspace: %v", err)
	}

	record, err := d.findTxtRecord(fqdn, zoneID)
	if err != nil {
		return fmt.Errorf("rackspace: %v", err)
	}

	_, err = d.makeRequest(http.MethodDelete, fmt.Sprintf("/domains/%d/records?id=%s", zoneID, record.ID), nil)
	if err != nil {
		return fmt.Errorf("rackspace: %v", err)
	}
	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// getHostedZoneID performs a lookup to get the DNS zone which needs
// modifying for a given FQDN
func (d *DNSProvider) getHostedZoneID(fqdn string) (int, error) {
	// HostedZones represents the response when querying Rackspace DNS zones
	type ZoneSearchResponse struct {
		TotalEntries int `json:"totalEntries"`
		HostedZones  []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"domains"`
	}

	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return 0, err
	}

	result, err := d.makeRequest(http.MethodGet, fmt.Sprintf("/domains?name=%s", acme.UnFqdn(authZone)), nil)
	if err != nil {
		return 0, err
	}

	var zoneSearchResponse ZoneSearchResponse
	err = json.Unmarshal(result, &zoneSearchResponse)
	if err != nil {
		return 0, err
	}

	// If nothing was returned, or for whatever reason more than 1 was returned (the search uses exact match, so should not occur)
	if zoneSearchResponse.TotalEntries != 1 {
		return 0, fmt.Errorf("found %d zones for %s in Rackspace for domain %s", zoneSearchResponse.TotalEntries, authZone, fqdn)
	}

	return zoneSearchResponse.HostedZones[0].ID, nil
}

// findTxtRecord searches a DNS zone for a TXT record with a specific name
func (d *DNSProvider) findTxtRecord(fqdn string, zoneID int) (*Record, error) {
	result, err := d.makeRequest(http.MethodGet, fmt.Sprintf("/domains/%d/records?type=TXT&name=%s", zoneID, acme.UnFqdn(fqdn)), nil)
	if err != nil {
		return nil, err
	}

	var records Records
	err = json.Unmarshal(result, &records)
	if err != nil {
		return nil, err
	}

	recordsLength := len(records.Record)
	switch recordsLength {
	case 1:
	case 0:
		return nil, fmt.Errorf("no TXT record found for %s", fqdn)
	default:
		return nil, fmt.Errorf("more than 1 TXT record found for %s", fqdn)
	}

	return &records.Record[0], nil
}

// makeRequest is a wrapper function used for making DNS API requests
func (d *DNSProvider) makeRequest(method, uri string, body io.Reader) (json.RawMessage, error) {
	url := d.cloudDNSEndpoint + uri
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Auth-Token", d.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error querying DNS API: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("request failed for %s %s. Response code: %d", method, url, resp.StatusCode)
	}

	var r json.RawMessage
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, fmt.Errorf("JSON decode failed for %s %s. Response code: %d", method, url, resp.StatusCode)
	}

	return r, nil
}
