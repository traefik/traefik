// Package digitalocean implements a DNS provider for solving the DNS-01
// challenge using digitalocean DNS.
package digitalocean

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	BaseURL            string
	AuthToken          string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		BaseURL:            defaultBaseURL,
		TTL:                env.GetOrDefaultInt("DO_TTL", 30),
		PropagationTimeout: env.GetOrDefaultSecond("DO_PROPAGATION_TIMEOUT", 60*time.Second),
		PollingInterval:    env.GetOrDefaultSecond("DO_POLLING_INTERVAL", 5*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("DO_HTTP_TIMEOUT", 30*time.Second),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface
// that uses DigitalOcean's REST API to manage TXT records for a domain.
type DNSProvider struct {
	config      *Config
	recordIDs   map[string]int
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Digital
// Ocean. Credentials must be passed in the environment variable:
// DO_AUTH_TOKEN.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("DO_AUTH_TOKEN")
	if err != nil {
		return nil, fmt.Errorf("digitalocean: %v", err)
	}

	config := NewDefaultConfig()
	config.AuthToken = values["DO_AUTH_TOKEN"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderCredentials uses the supplied credentials
// to return a DNSProvider instance configured for Digital Ocean.
// Deprecated
func NewDNSProviderCredentials(apiAuthToken string) (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.AuthToken = apiAuthToken

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Digital Ocean.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("digitalocean: the configuration of the DNS provider is nil")
	}

	if config.AuthToken == "" {
		return nil, fmt.Errorf("digitalocean: credentials missing")
	}

	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}

	return &DNSProvider{
		config:    config,
		recordIDs: make(map[string]int),
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)

	respData, err := d.addTxtRecord(domain, fqdn, value)
	if err != nil {
		return fmt.Errorf("digitalocean: %v", err)
	}

	d.recordIDsMu.Lock()
	d.recordIDs[fqdn] = respData.DomainRecord.ID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	// get the record's unique ID from when we created it
	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[fqdn]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("digitalocean: unknown record ID for '%s'", fqdn)
	}

	err := d.removeTxtRecord(domain, recordID)
	if err != nil {
		return fmt.Errorf("digitalocean: %v", err)
	}

	// Delete record ID from map
	d.recordIDsMu.Lock()
	delete(d.recordIDs, fqdn)
	d.recordIDsMu.Unlock()

	return nil
}

func (d *DNSProvider) removeTxtRecord(domain string, recordID int) error {
	authZone, err := acme.FindZoneByFqdn(acme.ToFqdn(domain), acme.RecursiveNameservers)
	if err != nil {
		return fmt.Errorf("could not determine zone for domain: '%s'. %s", domain, err)
	}

	reqURL := fmt.Sprintf("%s/v2/domains/%s/records/%d", d.config.BaseURL, acme.UnFqdn(authZone), recordID)
	req, err := d.newRequest(http.MethodDelete, reqURL, nil)
	if err != nil {
		return err
	}

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return readError(req, resp)
	}

	return nil
}

func (d *DNSProvider) addTxtRecord(domain, fqdn, value string) (*txtRecordResponse, error) {
	authZone, err := acme.FindZoneByFqdn(acme.ToFqdn(domain), acme.RecursiveNameservers)
	if err != nil {
		return nil, fmt.Errorf("could not determine zone for domain: '%s'. %s", domain, err)
	}

	reqData := txtRecordRequest{RecordType: "TXT", Name: fqdn, Data: value, TTL: d.config.TTL}
	body, err := json.Marshal(reqData)
	if err != nil {
		return nil, err
	}

	reqURL := fmt.Sprintf("%s/v2/domains/%s/records", d.config.BaseURL, acme.UnFqdn(authZone))
	req, err := d.newRequest(http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, readError(req, resp)
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(toUnreadableBodyMessage(req, content))
	}

	// Everything looks good; but we'll need the ID later to delete the record
	respData := &txtRecordResponse{}
	err = json.Unmarshal(content, respData)
	if err != nil {
		return nil, fmt.Errorf("%v: %s", err, toUnreadableBodyMessage(req, content))
	}

	return respData, nil
}

func (d *DNSProvider) newRequest(method, reqURL string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.config.AuthToken))

	return req, nil
}

func readError(req *http.Request, resp *http.Response) error {
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New(toUnreadableBodyMessage(req, content))
	}

	var errInfo digitalOceanAPIError
	err = json.Unmarshal(content, &errInfo)
	if err != nil {
		return fmt.Errorf("digitalOceanAPIError unmarshaling error: %v: %s", err, toUnreadableBodyMessage(req, content))
	}

	return fmt.Errorf("HTTP %d: %s: %s", resp.StatusCode, errInfo.ID, errInfo.Message)
}

func toUnreadableBodyMessage(req *http.Request, rawBody []byte) string {
	return fmt.Sprintf("the request %s sent a response with a body which is an invalid format: %q", req.URL, string(rawBody))
}
