package dnsmadeeasy

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	BaseURL            string
	APIKey             string
	APISecret          string
	HTTPClient         *http.Client
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("DNSMADEEASY_TTL", 120),
		PropagationTimeout: env.GetOrDefaultSecond("DNSMADEEASY_PROPAGATION_TIMEOUT", acme.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("DNSMADEEASY_POLLING_INTERVAL", acme.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("DNSMADEEASY_HTTP_TIMEOUT", 10*time.Second),
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface that uses
// DNSMadeEasy's DNS API to manage TXT records for a domain.
type DNSProvider struct {
	config *Config
	client *Client
}

// NewDNSProvider returns a DNSProvider instance configured for DNSMadeEasy DNS.
// Credentials must be passed in the environment variables:
// DNSMADEEASY_API_KEY and DNSMADEEASY_API_SECRET.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("DNSMADEEASY_API_KEY", "DNSMADEEASY_API_SECRET")
	if err != nil {
		return nil, fmt.Errorf("dnsmadeeasy: %v", err)
	}

	var baseURL string
	if sandbox, _ := strconv.ParseBool(os.Getenv("DNSMADEEASY_SANDBOX")); sandbox {
		baseURL = "https://api.sandbox.dnsmadeeasy.com/V2.0"
	} else {
		baseURL = "https://api.dnsmadeeasy.com/V2.0"
	}

	config := NewDefaultConfig()
	config.BaseURL = baseURL
	config.APIKey = values["DNSMADEEASY_API_KEY"]
	config.APISecret = values["DNSMADEEASY_API_SECRET"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderCredentials uses the supplied credentials
// to return a DNSProvider instance configured for DNS Made Easy.
// Deprecated
func NewDNSProviderCredentials(baseURL, apiKey, apiSecret string) (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.BaseURL = baseURL
	config.APIKey = apiKey
	config.APISecret = apiSecret

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for DNS Made Easy.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("dnsmadeeasy: the configuration of the DNS provider is nil")
	}

	if config.BaseURL == "" {
		return nil, fmt.Errorf("dnsmadeeasy: base URL missing")
	}

	client, err := NewClient(config.APIKey, config.APISecret)
	if err != nil {
		return nil, fmt.Errorf("dnsmadeeasy: %v", err)
	}

	client.HTTPClient = config.HTTPClient
	client.BaseURL = config.BaseURL

	return &DNSProvider{
		client: client,
		config: config,
	}, nil
}

// Present creates a TXT record using the specified parameters
func (d *DNSProvider) Present(domainName, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domainName, keyAuth)

	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return err
	}

	// fetch the domain details
	domain, err := d.client.GetDomain(authZone)
	if err != nil {
		return err
	}

	// create the TXT record
	name := strings.Replace(fqdn, "."+authZone, "", 1)
	record := &Record{Type: "TXT", Name: name, Value: value, TTL: d.config.TTL}

	err = d.client.CreateRecord(domain, record)
	return err
}

// CleanUp removes the TXT records matching the specified parameters
func (d *DNSProvider) CleanUp(domainName, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domainName, keyAuth)

	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return err
	}

	// fetch the domain details
	domain, err := d.client.GetDomain(authZone)
	if err != nil {
		return err
	}

	// find matching records
	name := strings.Replace(fqdn, "."+authZone, "", 1)
	records, err := d.client.GetRecords(domain, name, "TXT")
	if err != nil {
		return err
	}

	// delete records
	for _, record := range *records {
		err = d.client.DeleteRecord(record)
		if err != nil {
			return err
		}
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
