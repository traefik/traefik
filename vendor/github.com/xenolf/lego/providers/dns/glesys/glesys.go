// Package glesys implements a DNS provider for solving the DNS-01
// challenge using GleSYS api.
package glesys

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

// GleSYS API reference: https://github.com/GleSYS/API/wiki/API-Documentation

const (
	// defaultBaseURL is the GleSYS API endpoint used by Present and CleanUp.
	defaultBaseURL = "https://api.glesys.com/domain"
	minTTL         = 60
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
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
		TTL:                env.GetOrDefaultInt("GLESYS_TTL", minTTL),
		PropagationTimeout: env.GetOrDefaultSecond("GLESYS_PROPAGATION_TIMEOUT", 20*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond("GLESYS_POLLING_INTERVAL", 20*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("GLESYS_HTTP_TIMEOUT", 10*time.Second),
		},
	}
}

// DNSProvider is an implementation of the
// acme.ChallengeProviderTimeout interface that uses GleSYS
// API to manage TXT records for a domain.
type DNSProvider struct {
	config        *Config
	activeRecords map[string]int
	inProgressMu  sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for GleSYS.
// Credentials must be passed in the environment variables:
// GLESYS_API_USER and GLESYS_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("GLESYS_API_USER", "GLESYS_API_KEY")
	if err != nil {
		return nil, fmt.Errorf("glesys: %v", err)
	}

	config := NewDefaultConfig()
	config.APIUser = values["GLESYS_API_USER"]
	config.APIKey = values["GLESYS_API_KEY"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderCredentials uses the supplied credentials
// to return a DNSProvider instance configured for GleSYS.
// Deprecated
func NewDNSProviderCredentials(apiUser string, apiKey string) (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.APIUser = apiUser
	config.APIKey = apiKey

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for GleSYS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("glesys: the configuration of the DNS provider is nil")
	}

	if config.APIUser == "" || config.APIKey == "" {
		return nil, fmt.Errorf("glesys: incomplete credentials provided")
	}

	return &DNSProvider{
		activeRecords: make(map[string]int),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)

	if d.config.TTL < minTTL {
		d.config.TTL = minTTL // 60 is GleSYS minimum value for ttl
	}
	// find authZone
	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return fmt.Errorf("glesys: findZoneByFqdn failure: %v", err)
	}

	// determine name of TXT record
	if !strings.HasSuffix(
		strings.ToLower(fqdn), strings.ToLower("."+authZone)) {
		return fmt.Errorf("glesys: unexpected authZone %s for fqdn %s", authZone, fqdn)
	}
	name := fqdn[:len(fqdn)-len("."+authZone)]

	// acquire lock and check there is not a challenge already in
	// progress for this value of authZone
	d.inProgressMu.Lock()
	defer d.inProgressMu.Unlock()

	// add TXT record into authZone
	recordID, err := d.addTXTRecord(domain, acme.UnFqdn(authZone), name, value, d.config.TTL)
	if err != nil {
		return err
	}

	// save data necessary for CleanUp
	d.activeRecords[fqdn] = recordID
	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	// acquire lock and retrieve authZone
	d.inProgressMu.Lock()
	defer d.inProgressMu.Unlock()
	if _, ok := d.activeRecords[fqdn]; !ok {
		// if there is no cleanup information then just return
		return nil
	}

	recordID := d.activeRecords[fqdn]
	delete(d.activeRecords, fqdn)

	// delete TXT record from authZone
	return d.deleteTXTRecord(domain, recordID)
}

// Timeout returns the values (20*time.Minute, 20*time.Second) which
// are used by the acme package as timeout and check interval values
// when checking for DNS record propagation with GleSYS.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// POSTing/Marshalling/Unmarshalling

func (d *DNSProvider) sendRequest(method string, resource string, payload interface{}) (*responseStruct, error) {
	url := fmt.Sprintf("%s/%s", defaultBaseURL, resource)

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(d.config.APIUser, d.config.APIKey)

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

	return &response, err
}

// functions to perform API actions

func (d *DNSProvider) addTXTRecord(fqdn string, domain string, name string, value string, ttl int) (int, error) {
	response, err := d.sendRequest(http.MethodPost, "addrecord", addRecordRequest{
		DomainName: domain,
		Host:       name,
		Type:       "TXT",
		Data:       value,
		TTL:        ttl,
	})

	if response != nil && response.Response.Status.Code == http.StatusOK {
		log.Infof("[%s]: Successfully created record id %d", fqdn, response.Response.Record.RecordID)
		return response.Response.Record.RecordID, nil
	}
	return 0, err
}

func (d *DNSProvider) deleteTXTRecord(fqdn string, recordid int) error {
	response, err := d.sendRequest(http.MethodPost, "deleterecord", deleteRecordRequest{
		RecordID: recordid,
	})
	if response != nil && response.Response.Status.Code == 200 {
		log.Infof("[%s]: Successfully deleted record id %d", fqdn, recordid)
	}
	return err
}
