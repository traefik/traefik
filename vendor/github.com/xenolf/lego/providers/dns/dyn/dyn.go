// Package dyn implements a DNS provider for solving the DNS-01 challenge
// using Dyn Managed DNS.
package dyn

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	CustomerName       string
	UserName           string
	Password           string
	HTTPClient         *http.Client
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("DYN_TTL", 120),
		PropagationTimeout: env.GetOrDefaultSecond("DYN_PROPAGATION_TIMEOUT", acme.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("DYN_POLLING_INTERVAL", acme.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("DYN_HTTP_TIMEOUT", 10*time.Second),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface that uses
// Dyn's Managed DNS API to manage TXT records for a domain.
type DNSProvider struct {
	config *Config
	token  string
}

// NewDNSProvider returns a DNSProvider instance configured for Dyn DNS.
// Credentials must be passed in the environment variables:
// DYN_CUSTOMER_NAME, DYN_USER_NAME and DYN_PASSWORD.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("DYN_CUSTOMER_NAME", "DYN_USER_NAME", "DYN_PASSWORD")
	if err != nil {
		return nil, fmt.Errorf("dyn: %v", err)
	}

	config := NewDefaultConfig()
	config.CustomerName = values["DYN_CUSTOMER_NAME"]
	config.UserName = values["DYN_USER_NAME"]
	config.Password = values["DYN_PASSWORD"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderCredentials uses the supplied credentials
// to return a DNSProvider instance configured for Dyn DNS.
// Deprecated
func NewDNSProviderCredentials(customerName, userName, password string) (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.CustomerName = customerName
	config.UserName = userName
	config.Password = password

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Dyn DNS
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("dyn: the configuration of the DNS provider is nil")
	}

	if config.CustomerName == "" || config.UserName == "" || config.Password == "" {
		return nil, fmt.Errorf("dyn: credentials missing")
	}

	return &DNSProvider{config: config}, nil
}

// Present creates a TXT record using the specified parameters
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)

	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return fmt.Errorf("dyn: %v", err)
	}

	err = d.login()
	if err != nil {
		return fmt.Errorf("dyn: %v", err)
	}

	data := map[string]interface{}{
		"rdata": map[string]string{
			"txtdata": value,
		},
		"ttl": strconv.Itoa(d.config.TTL),
	}

	resource := fmt.Sprintf("TXTRecord/%s/%s/", authZone, fqdn)
	_, err = d.sendRequest(http.MethodPost, resource, data)
	if err != nil {
		return fmt.Errorf("dyn: %v", err)
	}

	err = d.publish(authZone, "Added TXT record for ACME dns-01 challenge using lego client")
	if err != nil {
		return fmt.Errorf("dyn: %v", err)
	}

	return d.logout()
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return fmt.Errorf("dyn: %v", err)
	}

	err = d.login()
	if err != nil {
		return fmt.Errorf("dyn: %v", err)
	}

	resource := fmt.Sprintf("TXTRecord/%s/%s/", authZone, fqdn)
	url := fmt.Sprintf("%s/%s", defaultBaseURL, resource)

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("dyn: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Auth-Token", d.token)

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("dyn: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("dyn: API request failed to delete TXT record HTTP status code %d", resp.StatusCode)
	}

	err = d.publish(authZone, "Removed TXT record for ACME dns-01 challenge using lego client")
	if err != nil {
		return fmt.Errorf("dyn: %v", err)
	}

	return d.logout()
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Starts a new Dyn API Session. Authenticates using customerName, userName,
// password and receives a token to be used in for subsequent requests.
func (d *DNSProvider) login() error {
	payload := &creds{Customer: d.config.CustomerName, User: d.config.UserName, Pass: d.config.Password}
	dynRes, err := d.sendRequest(http.MethodPost, "Session", payload)
	if err != nil {
		return err
	}

	var s session
	err = json.Unmarshal(dynRes.Data, &s)
	if err != nil {
		return err
	}

	d.token = s.Token

	return nil
}

// Destroys Dyn Session
func (d *DNSProvider) logout() error {
	if len(d.token) == 0 {
		// nothing to do
		return nil
	}

	url := fmt.Sprintf("%s/Session", defaultBaseURL)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Auth-Token", d.token)

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed to delete session with HTTP status code %d", resp.StatusCode)
	}

	d.token = ""

	return nil
}

func (d *DNSProvider) publish(zone, notes string) error {
	pub := &publish{Publish: true, Notes: notes}
	resource := fmt.Sprintf("Zone/%s/", zone)

	_, err := d.sendRequest(http.MethodPut, resource, pub)
	return err
}

func (d *DNSProvider) sendRequest(method, resource string, payload interface{}) (*dynResponse, error) {
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
	if len(d.token) > 0 {
		req.Header.Set("Auth-Token", d.token)
	}

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("API request failed with HTTP status code %d", resp.StatusCode)
	}

	var dynRes dynResponse
	err = json.NewDecoder(resp.Body).Decode(&dynRes)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API request failed with HTTP status code %d: %s", resp.StatusCode, dynRes.Messages)
	} else if resp.StatusCode == 307 {
		// TODO add support for HTTP 307 response and long running jobs
		return nil, fmt.Errorf("API request returned HTTP 307. This is currently unsupported")
	}

	if dynRes.Status == "failure" {
		// TODO add better error handling
		return nil, fmt.Errorf("API request failed: %s", dynRes.Messages)
	}

	return &dynRes, nil
}
