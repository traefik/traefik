// Package otc implements a DNS provider for solving the DNS-01 challenge using Open Telekom Cloud Managed DNS.
package otc

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
)

const defaultIdentityEndpoint = "https://iam.eu-de.otc.t-systems.com:443/v3/auth/tokens"

// minTTL 300 is otc minimum value for ttl
const minTTL = 300

// Config is used to configure the creation of the DNSProvider
type Config struct {
	IdentityEndpoint   string
	DomainName         string
	ProjectName        string
	UserName           string
	Password           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		IdentityEndpoint:   env.GetOrDefaultString("OTC_IDENTITY_ENDPOINT", defaultIdentityEndpoint),
		PropagationTimeout: env.GetOrDefaultSecond("OTC_PROPAGATION_TIMEOUT", dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("OTC_POLLING_INTERVAL", dns01.DefaultPollingInterval),
		TTL:                env.GetOrDefaultInt("OTC_TTL", minTTL),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("OTC_HTTP_TIMEOUT", 10*time.Second),
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
					DualStack: true,
				}).DialContext,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,

				// Workaround for keep alive bug in otc api
				DisableKeepAlives: true,
			},
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface that uses
// OTC's Managed DNS API to manage TXT records for a domain.
type DNSProvider struct {
	config  *Config
	baseURL string
	token   string
}

// NewDNSProvider returns a DNSProvider instance configured for OTC DNS.
// Credentials must be passed in the environment variables: OTC_USER_NAME,
// OTC_DOMAIN_NAME, OTC_PASSWORD OTC_PROJECT_NAME and OTC_IDENTITY_ENDPOINT.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("OTC_DOMAIN_NAME", "OTC_USER_NAME", "OTC_PASSWORD", "OTC_PROJECT_NAME")
	if err != nil {
		return nil, fmt.Errorf("otc: %v", err)
	}

	config := NewDefaultConfig()
	config.DomainName = values["OTC_DOMAIN_NAME"]
	config.UserName = values["OTC_USER_NAME"]
	config.Password = values["OTC_PASSWORD"]
	config.ProjectName = values["OTC_PROJECT_NAME"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for OTC DNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("otc: the configuration of the DNS provider is nil")
	}

	if config.DomainName == "" || config.UserName == "" || config.Password == "" || config.ProjectName == "" {
		return nil, fmt.Errorf("otc: credentials missing")
	}

	if config.TTL < minTTL {
		return nil, fmt.Errorf("otc: invalid TTL, TTL (%d) must be greater than %d", config.TTL, minTTL)
	}

	if config.IdentityEndpoint == "" {
		config.IdentityEndpoint = defaultIdentityEndpoint
	}

	return &DNSProvider{config: config}, nil
}

// Present creates a TXT record using the specified parameters
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("otc: %v", err)
	}

	err = d.login()
	if err != nil {
		return fmt.Errorf("otc: %v", err)
	}

	zoneID, err := d.getZoneID(authZone)
	if err != nil {
		return fmt.Errorf("otc: unable to get zone: %s", err)
	}

	resource := fmt.Sprintf("zones/%s/recordsets", zoneID)

	r1 := &recordset{
		Name:        fqdn,
		Description: "Added TXT record for ACME dns-01 challenge using lego client",
		Type:        "TXT",
		TTL:         d.config.TTL,
		Records:     []string{fmt.Sprintf("\"%s\"", value)},
	}

	_, err = d.sendRequest(http.MethodPost, resource, r1)
	if err != nil {
		return fmt.Errorf("otc: %v", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return fmt.Errorf("otc: %v", err)
	}

	err = d.login()
	if err != nil {
		return fmt.Errorf("otc: %v", err)
	}

	zoneID, err := d.getZoneID(authZone)
	if err != nil {
		return fmt.Errorf("otc: %v", err)
	}

	recordID, err := d.getRecordSetID(zoneID, fqdn)
	if err != nil {
		return fmt.Errorf("otc: unable go get record %s for zone %s: %s", fqdn, domain, err)
	}

	err = d.deleteRecordSet(zoneID, recordID)
	if err != nil {
		return fmt.Errorf("otc: %v", err)
	}
	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
