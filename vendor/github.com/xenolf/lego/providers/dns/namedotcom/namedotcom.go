// Package namedotcom implements a DNS provider for solving the DNS-01 challenge
// using Name.com's DNS service.
package namedotcom

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/namedotcom/go/namecom"
	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	Username           string
	APIToken           string
	Server             string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("NAMECOM_TTL", 120),
		PropagationTimeout: env.GetOrDefaultSecond("NAMECOM_PROPAGATION_TIMEOUT", acme.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("NAMECOM_POLLING_INTERVAL", acme.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("NAMECOM_HTTP_TIMEOUT", 10*time.Second),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	client *namecom.NameCom
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for namedotcom.
// Credentials must be passed in the environment variables:
// NAMECOM_USERNAME and NAMECOM_API_TOKEN
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("NAMECOM_USERNAME", "NAMECOM_API_TOKEN")
	if err != nil {
		return nil, fmt.Errorf("namedotcom: %v", err)
	}

	config := NewDefaultConfig()
	config.Username = values["NAMECOM_USERNAME"]
	config.APIToken = values["NAMECOM_API_TOKEN"]
	config.Server = os.Getenv("NAMECOM_SERVER")

	return NewDNSProviderConfig(config)
}

// NewDNSProviderCredentials uses the supplied credentials
// to return a DNSProvider instance configured for namedotcom.
// Deprecated
func NewDNSProviderCredentials(username, apiToken, server string) (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.Username = username
	config.APIToken = apiToken
	config.Server = server

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for namedotcom.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("namedotcom: the configuration of the DNS provider is nil")
	}

	if config.Username == "" {
		return nil, fmt.Errorf("namedotcom: username is required")
	}

	if config.APIToken == "" {
		return nil, fmt.Errorf("namedotcom: API token is required")
	}

	client := namecom.New(config.Username, config.APIToken)
	client.Client = config.HTTPClient

	if config.Server != "" {
		client.Server = config.Server
	}

	return &DNSProvider{client: client, config: config}, nil
}

// Present creates a TXT record to fulfil the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)

	request := &namecom.Record{
		DomainName: domain,
		Host:       d.extractRecordName(fqdn, domain),
		Type:       "TXT",
		TTL:        uint32(d.config.TTL),
		Answer:     value,
	}

	_, err := d.client.CreateRecord(request)
	if err != nil {
		return fmt.Errorf("namedotcom: API call failed: %v", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	records, err := d.getRecords(domain)
	if err != nil {
		return fmt.Errorf("namedotcom: %v", err)
	}

	for _, rec := range records {
		if rec.Fqdn == fqdn && rec.Type == "TXT" {
			request := &namecom.DeleteRecordRequest{
				DomainName: domain,
				ID:         rec.ID,
			}
			_, err := d.client.DeleteRecord(request)
			if err != nil {
				return fmt.Errorf("namedotcom: %v", err)
			}
		}
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) getRecords(domain string) ([]*namecom.Record, error) {
	request := &namecom.ListRecordsRequest{
		DomainName: domain,
		Page:       1,
	}

	var records []*namecom.Record
	for request.Page > 0 {
		response, err := d.client.ListRecords(request)
		if err != nil {
			return nil, err
		}

		records = append(records, response.Records...)
		request.Page = response.NextPage
	}

	return records, nil
}

func (d *DNSProvider) extractRecordName(fqdn, domain string) string {
	name := acme.UnFqdn(fqdn)
	if idx := strings.Index(name, "."+domain); idx != -1 {
		return name[:idx]
	}
	return name
}
