// Package hostingde implements a DNS provider for solving the DNS-01
// challenge using hosting.de.
package hostingde

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

const defaultBaseURL = "https://secure.hosting.de/api/dns/v1/json"

// Config is used to configure the creation of the DNSProvider
type Config struct {
	APIKey             string
	ZoneName           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("HOSTINGDE_TTL", 120),
		PropagationTimeout: env.GetOrDefaultSecond("HOSTINGDE_PROPAGATION_TIMEOUT", 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond("HOSTINGDE_POLLING_INTERVAL", 2*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("HOSTINGDE_HTTP_TIMEOUT", 30*time.Second),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	config      *Config
	recordIDs   map[string]string
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for hosting.de.
// Credentials must be passed in the environment variables:
// HOSTINGDE_ZONE_NAME and HOSTINGDE_API_KEY
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("HOSTINGDE_API_KEY", "HOSTINGDE_ZONE_NAME")
	if err != nil {
		return nil, fmt.Errorf("hostingde: %v", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values["HOSTINGDE_API_KEY"]
	config.ZoneName = values["HOSTINGDE_ZONE_NAME"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for hosting.de.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("hostingde: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, errors.New("hostingde: API key missing")
	}

	if config.ZoneName == "" {
		return nil, errors.New("hostingde: Zone Name missing")
	}

	return &DNSProvider{
		config:    config,
		recordIDs: make(map[string]string),
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)

	rec := []RecordsAddRequest{{
		Type:    "TXT",
		Name:    acme.UnFqdn(fqdn),
		Content: value,
		TTL:     d.config.TTL,
	}}

	req := ZoneUpdateRequest{
		AuthToken: d.config.APIKey,
		ZoneConfigSelector: ZoneConfigSelector{
			Name: d.config.ZoneName,
		},
		RecordsToAdd: rec,
	}

	resp, err := d.updateZone(req)
	if err != nil {
		return fmt.Errorf("hostingde: %v", err)
	}

	for _, record := range resp.Response.Records {
		if record.Name == acme.UnFqdn(fqdn) && record.Content == fmt.Sprintf(`"%s"`, value) {
			d.recordIDsMu.Lock()
			d.recordIDs[fqdn] = record.ID
			d.recordIDsMu.Unlock()
		}
	}

	if d.recordIDs[fqdn] == "" {
		return fmt.Errorf("hostingde: error getting ID of just created record, for domain %s", domain)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)

	// get the record's unique ID from when we created it
	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[fqdn]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("hostingde: unknown record ID for %q", fqdn)
	}

	rec := []RecordsDeleteRequest{{
		Type:    "TXT",
		Name:    acme.UnFqdn(fqdn),
		Content: value,
		ID:      recordID,
	}}

	req := ZoneUpdateRequest{
		AuthToken: d.config.APIKey,
		ZoneConfigSelector: ZoneConfigSelector{
			Name: d.config.ZoneName,
		},
		RecordsToDelete: rec,
	}

	// Delete record ID from map
	d.recordIDsMu.Lock()
	delete(d.recordIDs, fqdn)
	d.recordIDsMu.Unlock()

	_, err := d.updateZone(req)
	if err != nil {
		return fmt.Errorf("hostingde: %v", err)
	}
	return nil
}

func (d *DNSProvider) updateZone(updateRequest ZoneUpdateRequest) (*ZoneUpdateResponse, error) {
	body, err := json.Marshal(updateRequest)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, defaultBaseURL+"/zoneUpdate", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error querying API: %v", err)
	}

	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(toUnreadableBodyMessage(req, content))
	}

	// Everything looks good; but we'll need the ID later to delete the record
	updateResponse := &ZoneUpdateResponse{}
	err = json.Unmarshal(content, updateResponse)
	if err != nil {
		return nil, fmt.Errorf("%v: %s", err, toUnreadableBodyMessage(req, content))
	}

	if updateResponse.Status != "success" && updateResponse.Status != "pending" {
		return updateResponse, errors.New(toUnreadableBodyMessage(req, content))
	}

	return updateResponse, nil
}

func toUnreadableBodyMessage(req *http.Request, rawBody []byte) string {
	return fmt.Sprintf("the request %s sent a response with a body which is an invalid format: %q", req.URL, string(rawBody))
}
