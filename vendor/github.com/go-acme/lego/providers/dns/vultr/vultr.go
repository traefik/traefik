// Package vultr implements a DNS provider for solving the DNS-01 challenge using the vultr DNS.
// See https://www.vultr.com/api/#dns
package vultr

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	vultr "github.com/JamesClonk/vultr/lib"
	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
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
		TTL:                env.GetOrDefaultInt("VULTR_TTL", dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond("VULTR_PROPAGATION_TIMEOUT", dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("VULTR_POLLING_INTERVAL", dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("VULTR_HTTP_TIMEOUT", 0),
			// from Vultr Client
			Transport: &http.Transport{
				TLSNextProto: make(map[string]func(string, *tls.Conn) http.RoundTripper),
			},
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	config *Config
	client *vultr.Client
}

// NewDNSProvider returns a DNSProvider instance with a configured Vultr client.
// Authentication uses the VULTR_API_KEY environment variable.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("VULTR_API_KEY")
	if err != nil {
		return nil, fmt.Errorf("vultr: %v", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values["VULTR_API_KEY"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Vultr.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("vultr: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("vultr: credentials missing")
	}

	options := &vultr.Options{
		HTTPClient: config.HTTPClient,
	}
	client := vultr.NewClient(config.APIKey, options)

	return &DNSProvider{client: client, config: config}, nil
}

// Present creates a TXT record to fulfill the DNS-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	zoneDomain, err := d.getHostedZone(domain)
	if err != nil {
		return fmt.Errorf("vultr: %v", err)
	}

	name := d.extractRecordName(fqdn, zoneDomain)

	err = d.client.CreateDNSRecord(zoneDomain, name, "TXT", `"`+value+`"`, 0, d.config.TTL)
	if err != nil {
		return fmt.Errorf("vultr: API call failed: %v", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	zoneDomain, records, err := d.findTxtRecords(domain, fqdn)
	if err != nil {
		return fmt.Errorf("vultr: %v", err)
	}

	var allErr []string
	for _, rec := range records {
		err := d.client.DeleteDNSRecord(zoneDomain, rec.RecordID)
		if err != nil {
			allErr = append(allErr, err.Error())
		}
	}

	if len(allErr) > 0 {
		return errors.New(strings.Join(allErr, ": "))
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) getHostedZone(domain string) (string, error) {
	domains, err := d.client.GetDNSDomains()
	if err != nil {
		return "", fmt.Errorf("API call failed: %v", err)
	}

	var hostedDomain vultr.DNSDomain
	for _, dom := range domains {
		if strings.HasSuffix(domain, dom.Domain) {
			if len(dom.Domain) > len(hostedDomain.Domain) {
				hostedDomain = dom
			}
		}
	}
	if hostedDomain.Domain == "" {
		return "", fmt.Errorf("no matching Vultr domain found for domain %s", domain)
	}

	return hostedDomain.Domain, nil
}

func (d *DNSProvider) findTxtRecords(domain, fqdn string) (string, []vultr.DNSRecord, error) {
	zoneDomain, err := d.getHostedZone(domain)
	if err != nil {
		return "", nil, err
	}

	var records []vultr.DNSRecord
	result, err := d.client.GetDNSRecords(zoneDomain)
	if err != nil {
		return "", records, fmt.Errorf("API call has failed: %v", err)
	}

	recordName := d.extractRecordName(fqdn, zoneDomain)
	for _, record := range result {
		if record.Type == "TXT" && record.Name == recordName {
			records = append(records, record)
		}
	}

	return zoneDomain, records, nil
}

func (d *DNSProvider) extractRecordName(fqdn, domain string) string {
	name := dns01.UnFqdn(fqdn)
	if idx := strings.Index(name, "."+domain); idx != -1 {
		return name[:idx]
	}
	return name
}
