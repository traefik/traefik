// Package ns1 implements a DNS provider for solving the DNS-01 challenge
// using NS1 DNS.
package ns1

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
	"gopkg.in/ns1/ns1-go.v2/rest"
	"gopkg.in/ns1/ns1-go.v2/rest/model/dns"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	APIKey             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("NS1_TTL", 120),
		PropagationTimeout: env.GetOrDefaultSecond("NS1_PROPAGATION_TIMEOUT", acme.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("NS1_POLLING_INTERVAL", acme.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("NS1_HTTP_TIMEOUT", 10*time.Second),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	client *rest.Client
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for NS1.
// Credentials must be passed in the environment variables: NS1_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("NS1_API_KEY")
	if err != nil {
		return nil, fmt.Errorf("ns1: %v", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values["NS1_API_KEY"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderCredentials uses the supplied credentials
// to return a DNSProvider instance configured for NS1.
// Deprecated
func NewDNSProviderCredentials(key string) (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.APIKey = key

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for NS1.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("ns1: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("ns1: credentials missing")
	}

	client := rest.NewClient(config.HTTPClient, rest.SetAPIKey(config.APIKey))

	return &DNSProvider{client: client, config: config}, nil
}

// Present creates a TXT record to fulfil the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)

	zone, err := d.getHostedZone(domain)
	if err != nil {
		return fmt.Errorf("ns1: %v", err)
	}

	record := d.newTxtRecord(zone, fqdn, value, d.config.TTL)
	_, err = d.client.Records.Create(record)
	if err != nil && err != rest.ErrRecordExists {
		return fmt.Errorf("ns1: %v", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	zone, err := d.getHostedZone(domain)
	if err != nil {
		return fmt.Errorf("ns1: %v", err)
	}

	name := acme.UnFqdn(fqdn)
	_, err = d.client.Records.Delete(zone.Zone, name, "TXT")
	return fmt.Errorf("ns1: %v", err)
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) getHostedZone(domain string) (*dns.Zone, error) {
	authZone, err := getAuthZone(domain)
	if err != nil {
		return nil, fmt.Errorf("ns1: %v", err)
	}

	zone, _, err := d.client.Zones.Get(authZone)
	if err != nil {
		return nil, fmt.Errorf("ns1: %v", err)
	}

	return zone, nil
}

func getAuthZone(fqdn string) (string, error) {
	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return "", err
	}

	if strings.HasSuffix(authZone, ".") {
		authZone = authZone[:len(authZone)-len(".")]
	}

	return authZone, err
}

func (d *DNSProvider) newTxtRecord(zone *dns.Zone, fqdn, value string, ttl int) *dns.Record {
	name := acme.UnFqdn(fqdn)

	return &dns.Record{
		Type:   "TXT",
		Zone:   zone.Zone,
		Domain: name,
		TTL:    ttl,
		Answers: []*dns.Answer{
			{Rdata: []string{value}},
		},
	}
}
