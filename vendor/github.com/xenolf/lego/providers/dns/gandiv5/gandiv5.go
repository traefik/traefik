// Package gandiv5 implements a DNS provider for solving the DNS-01
// challenge using Gandi LiveDNS api.
package gandiv5

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/log"
	"github.com/xenolf/lego/platform/config/env"
)

// Gandi API reference:       http://doc.livedns.gandi.net/

const (
	// defaultBaseURL endpoint is the Gandi API endpoint used by Present and CleanUp.
	defaultBaseURL = "https://dns.api.gandi.net/api/v5"
	minTTL         = 300
)

// findZoneByFqdn determines the DNS zone of an fqdn.
// It is overridden during tests.
var findZoneByFqdn = acme.FindZoneByFqdn

// inProgressInfo contains information about an in-progress challenge
type inProgressInfo struct {
	fieldName string
	authZone  string
}

// Config is used to configure the creation of the DNSProvider
type Config struct {
	BaseURL            string
	APIKey             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("GANDIV5_TTL", minTTL),
		PropagationTimeout: env.GetOrDefaultSecond("GANDIV5_PROPAGATION_TIMEOUT", 20*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond("GANDIV5_POLLING_INTERVAL", 20*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("GANDIV5_HTTP_TIMEOUT", 10*time.Second),
		},
	}
}

// DNSProvider is an implementation of the
// acme.ChallengeProviderTimeout interface that uses Gandi's LiveDNS
// API to manage TXT records for a domain.
type DNSProvider struct {
	config          *Config
	inProgressFQDNs map[string]inProgressInfo
	inProgressMu    sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Gandi.
// Credentials must be passed in the environment variable: GANDIV5_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("GANDIV5_API_KEY")
	if err != nil {
		return nil, fmt.Errorf("gandi: %v", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values["GANDIV5_API_KEY"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderCredentials uses the supplied credentials
// to return a DNSProvider instance configured for Gandi.
// Deprecated
func NewDNSProviderCredentials(apiKey string) (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.APIKey = apiKey

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Gandi.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("gandiv5: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("gandiv5: no API Key given")
	}

	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}

	return &DNSProvider{
		config:          config,
		inProgressFQDNs: make(map[string]inProgressInfo),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)

	if d.config.TTL < minTTL {
		d.config.TTL = minTTL // 300 is gandi minimum value for ttl
	}

	// find authZone
	authZone, err := findZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return fmt.Errorf("gandiv5: findZoneByFqdn failure: %v", err)
	}

	// determine name of TXT record
	if !strings.HasSuffix(
		strings.ToLower(fqdn), strings.ToLower("."+authZone)) {
		return fmt.Errorf("gandiv5: unexpected authZone %s for fqdn %s", authZone, fqdn)
	}
	name := fqdn[:len(fqdn)-len("."+authZone)]

	// acquire lock and check there is not a challenge already in
	// progress for this value of authZone
	d.inProgressMu.Lock()
	defer d.inProgressMu.Unlock()

	// add TXT record into authZone
	err = d.addTXTRecord(acme.UnFqdn(authZone), name, value, d.config.TTL)
	if err != nil {
		return err
	}

	// save data necessary for CleanUp
	d.inProgressFQDNs[fqdn] = inProgressInfo{
		authZone:  authZone,
		fieldName: name,
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	// acquire lock and retrieve authZone
	d.inProgressMu.Lock()
	defer d.inProgressMu.Unlock()
	if _, ok := d.inProgressFQDNs[fqdn]; !ok {
		// if there is no cleanup information then just return
		return nil
	}

	fieldName := d.inProgressFQDNs[fqdn].fieldName
	authZone := d.inProgressFQDNs[fqdn].authZone
	delete(d.inProgressFQDNs, fqdn)

	// delete TXT record from authZone
	err := d.deleteTXTRecord(acme.UnFqdn(authZone), fieldName)
	if err != nil {
		return fmt.Errorf("gandiv5: %v", err)
	}
	return nil
}

// Timeout returns the values (20*time.Minute, 20*time.Second) which
// are used by the acme package as timeout and check interval values
// when checking for DNS record propagation with Gandi.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// functions to perform API actions

func (d *DNSProvider) addTXTRecord(domain string, name string, value string, ttl int) error {
	target := fmt.Sprintf("domains/%s/records/%s/TXT", domain, name)
	response, err := d.sendRequest(http.MethodPut, target, addFieldRequest{
		RRSetTTL:    ttl,
		RRSetValues: []string{value},
	})
	if response != nil {
		log.Infof("gandiv5: %s", response.Message)
	}
	return err
}

func (d *DNSProvider) deleteTXTRecord(domain string, name string) error {
	target := fmt.Sprintf("domains/%s/records/%s/TXT", domain, name)
	response, err := d.sendRequest(http.MethodDelete, target, deleteFieldRequest{
		Delete: true,
	})
	if response != nil && response.Message == "" {
		log.Infof("gandiv5: Zone record deleted")
	}
	return err
}

func (d *DNSProvider) sendRequest(method string, resource string, payload interface{}) (*responseStruct, error) {
	url := fmt.Sprintf("%s/%s", d.config.BaseURL, resource)

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if len(d.config.APIKey) > 0 {
		req.Header.Set("X-Api-Key", d.config.APIKey)
	}

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request failed with HTTP status code %d", resp.StatusCode)
	}

	var response responseStruct
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil && method != http.MethodDelete {
		return nil, err
	}

	return &response, nil
}
