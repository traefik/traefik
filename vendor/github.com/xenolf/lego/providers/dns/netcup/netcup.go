// Package netcup implements a DNS Provider for solving the DNS-01 challenge using the netcup DNS API.
package netcup

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	Key                string
	Password           string
	Customer           string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("NETCUP_TTL", 120),
		PropagationTimeout: env.GetOrDefaultSecond("NETCUP_PROPAGATION_TIMEOUT", acme.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("NETCUP_POLLING_INTERVAL", acme.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("NETCUP_HTTP_TIMEOUT", 10*time.Second),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	client *Client
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for netcup.
// Credentials must be passed in the environment variables:
// NETCUP_CUSTOMER_NUMBER, NETCUP_API_KEY, NETCUP_API_PASSWORD
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("NETCUP_CUSTOMER_NUMBER", "NETCUP_API_KEY", "NETCUP_API_PASSWORD")
	if err != nil {
		return nil, fmt.Errorf("netcup: %v", err)
	}

	config := NewDefaultConfig()
	config.Customer = values["NETCUP_CUSTOMER_NUMBER"]
	config.Key = values["NETCUP_API_KEY"]
	config.Password = values["NETCUP_API_PASSWORD"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderCredentials uses the supplied credentials
// to return a DNSProvider instance configured for netcup.
// Deprecated
func NewDNSProviderCredentials(customer, key, password string) (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.Customer = customer
	config.Key = key
	config.Password = password

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for netcup.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("netcup: the configuration of the DNS provider is nil")
	}

	if config.Customer == "" || config.Key == "" || config.Password == "" {
		return nil, fmt.Errorf("netcup: netcup credentials missing")
	}

	client := NewClient(config.Customer, config.Key, config.Password)
	client.HTTPClient = config.HTTPClient

	return &DNSProvider{client: client, config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domainName, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domainName, keyAuth)

	zone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return fmt.Errorf("netcup: failed to find DNSZone, %v", err)
	}

	sessionID, err := d.client.Login()
	if err != nil {
		return fmt.Errorf("netcup: %v", err)
	}

	hostname := strings.Replace(fqdn, "."+zone, "", 1)
	record := CreateTxtRecord(hostname, value, d.config.TTL)

	err = d.client.UpdateDNSRecord(sessionID, acme.UnFqdn(zone), record)
	if err != nil {
		if errLogout := d.client.Logout(sessionID); errLogout != nil {
			return fmt.Errorf("netcup: failed to add TXT-Record: %v; %v", err, errLogout)
		}
		return fmt.Errorf("netcup: failed to add TXT-Record: %v", err)
	}

	err = d.client.Logout(sessionID)
	if err != nil {
		return fmt.Errorf("netcup: %v", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domainname, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domainname, keyAuth)

	zone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return fmt.Errorf("netcup: failed to find DNSZone, %v", err)
	}

	sessionID, err := d.client.Login()
	if err != nil {
		return fmt.Errorf("netcup: %v", err)
	}

	hostname := strings.Replace(fqdn, "."+zone, "", 1)

	zone = acme.UnFqdn(zone)

	records, err := d.client.GetDNSRecords(zone, sessionID)
	if err != nil {
		return fmt.Errorf("netcup: %v", err)
	}

	record := CreateTxtRecord(hostname, value, 0)

	idx, err := GetDNSRecordIdx(records, record)
	if err != nil {
		return fmt.Errorf("netcup: %v", err)
	}

	records[idx].DeleteRecord = true

	err = d.client.UpdateDNSRecord(sessionID, zone, records[idx])
	if err != nil {
		if errLogout := d.client.Logout(sessionID); errLogout != nil {
			return fmt.Errorf("netcup: %v; %v", err, errLogout)
		}
		return fmt.Errorf("netcup: %v", err)
	}

	err = d.client.Logout(sessionID)
	if err != nil {
		return fmt.Errorf("netcup: %v", err)
	}
	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
