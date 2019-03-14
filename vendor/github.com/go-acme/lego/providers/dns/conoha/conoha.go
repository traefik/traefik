// Package conoha implements a DNS provider for solving the DNS-01 challenge using ConoHa DNS.
package conoha

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
	"github.com/go-acme/lego/providers/dns/conoha/internal"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	Region             string
	TenantID           string
	Username           string
	Password           string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		Region:             env.GetOrDefaultString("CONOHA_REGION", "tyo1"),
		TTL:                env.GetOrDefaultInt("CONOHA_TTL", 60),
		PropagationTimeout: env.GetOrDefaultSecond("CONOHA_PROPAGATION_TIMEOUT", dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("CONOHA_POLLING_INTERVAL", dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("CONOHA_HTTP_TIMEOUT", 30*time.Second),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	config *Config
	client *internal.Client
}

// NewDNSProvider returns a DNSProvider instance configured for ConoHa DNS.
// Credentials must be passed in the environment variables: CONOHA_TENANT_ID, CONOHA_API_USERNAME, CONOHA_API_PASSWORD
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("CONOHA_TENANT_ID", "CONOHA_API_USERNAME", "CONOHA_API_PASSWORD")
	if err != nil {
		return nil, fmt.Errorf("conoha: %v", err)
	}

	config := NewDefaultConfig()
	config.TenantID = values["CONOHA_TENANT_ID"]
	config.Username = values["CONOHA_API_USERNAME"]
	config.Password = values["CONOHA_API_PASSWORD"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for ConoHa DNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("conoha: the configuration of the DNS provider is nil")
	}

	if config.TenantID == "" || config.Username == "" || config.Password == "" {
		return nil, errors.New("conoha: some credentials information are missing")
	}

	auth := internal.Auth{
		TenantID: config.TenantID,
		PasswordCredentials: internal.PasswordCredentials{
			Username: config.Username,
			Password: config.Password,
		},
	}

	client, err := internal.NewClient(config.Region, auth, config.HTTPClient)
	if err != nil {
		return nil, fmt.Errorf("conoha: failed to create client: %v", err)
	}

	return &DNSProvider{config: config, client: client}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return err
	}

	id, err := d.client.GetDomainID(authZone)
	if err != nil {
		return fmt.Errorf("conoha: failed to get domain ID: %v", err)
	}

	record := internal.Record{
		Name: fqdn,
		Type: "TXT",
		Data: value,
		TTL:  d.config.TTL,
	}

	err = d.client.CreateRecord(id, record)
	if err != nil {
		return fmt.Errorf("conoha: failed to create record: %v", err)
	}

	return nil
}

// CleanUp clears ConoHa DNS TXT record
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return err
	}

	domID, err := d.client.GetDomainID(authZone)
	if err != nil {
		return fmt.Errorf("conoha: failed to get domain ID: %v", err)
	}

	recID, err := d.client.GetRecordID(domID, fqdn, "TXT", value)
	if err != nil {
		return fmt.Errorf("conoha: failed to get record ID: %v", err)
	}

	err = d.client.DeleteRecord(domID, recID)
	if err != nil {
		return fmt.Errorf("conoha: failed to delete record: %v", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
