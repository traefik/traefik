package rackspace

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-acme/lego/challenge/dns01"
)

// APIKeyCredentials API credential
type APIKeyCredentials struct {
	Username string `json:"username"`
	APIKey   string `json:"apiKey"`
}

// Auth auth credentials
type Auth struct {
	APIKeyCredentials `json:"RAX-KSKEY:apiKeyCredentials"`
}

// AuthData Auth data
type AuthData struct {
	Auth `json:"auth"`
}

// Identity Identity
type Identity struct {
	Access Access `json:"access"`
}

// Access Access
type Access struct {
	ServiceCatalog []ServiceCatalog `json:"serviceCatalog"`
	Token          Token            `json:"token"`
}

// Token Token
type Token struct {
	ID string `json:"id"`
}

// ServiceCatalog ServiceCatalog
type ServiceCatalog struct {
	Endpoints []Endpoint `json:"endpoints"`
	Name      string     `json:"name"`
}

// Endpoint Endpoint
type Endpoint struct {
	PublicURL string `json:"publicURL"`
	TenantID  string `json:"tenantId"`
}

// ZoneSearchResponse represents the response when querying Rackspace DNS zones
type ZoneSearchResponse struct {
	TotalEntries int          `json:"totalEntries"`
	HostedZones  []HostedZone `json:"domains"`
}

// HostedZone HostedZone
type HostedZone struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Records is the list of records sent/received from the DNS API
type Records struct {
	Record []Record `json:"records"`
}

// Record represents a Rackspace DNS record
type Record struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Data string `json:"data"`
	TTL  int    `json:"ttl,omitempty"`
	ID   string `json:"id,omitempty"`
}

// getHostedZoneID performs a lookup to get the DNS zone which needs
// modifying for a given FQDN
func (d *DNSProvider) getHostedZoneID(fqdn string) (int, error) {
	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return 0, err
	}

	result, err := d.makeRequest(http.MethodGet, fmt.Sprintf("/domains?name=%s", dns01.UnFqdn(authZone)), nil)
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
	result, err := d.makeRequest(http.MethodGet, fmt.Sprintf("/domains/%d/records?type=TXT&name=%s", zoneID, dns01.UnFqdn(fqdn)), nil)
	if err != nil {
		return nil, err
	}

	var records Records
	err = json.Unmarshal(result, &records)
	if err != nil {
		return nil, err
	}

	switch len(records.Record) {
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

func login(config *Config) (*Identity, error) {
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

	resp, err := config.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error querying Identity API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("authentication failed: response code: %d", resp.StatusCode)
	}

	var identity Identity
	err = json.NewDecoder(resp.Body).Decode(&identity)
	if err != nil {
		return nil, err
	}

	return &identity, nil
}
