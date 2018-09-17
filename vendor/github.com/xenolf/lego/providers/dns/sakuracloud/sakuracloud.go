// Package sakuracloud implements a DNS provider for solving the DNS-01 challenge
// using sakuracloud DNS.
package sakuracloud

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sacloud/libsacloud/api"
	"github.com/sacloud/libsacloud/sacloud"
	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	Token              string
	Secret             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("SAKURACLOUD_TTL", 120),
		PropagationTimeout: env.GetOrDefaultSecond("SAKURACLOUD_PROPAGATION_TIMEOUT", acme.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("SAKURACLOUD_POLLING_INTERVAL", acme.DefaultPollingInterval),
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	config *Config
	client *api.Client
}

// NewDNSProvider returns a DNSProvider instance configured for sakuracloud.
// Credentials must be passed in the environment variables: SAKURACLOUD_ACCESS_TOKEN & SAKURACLOUD_ACCESS_TOKEN_SECRET
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("SAKURACLOUD_ACCESS_TOKEN", "SAKURACLOUD_ACCESS_TOKEN_SECRET")
	if err != nil {
		return nil, fmt.Errorf("sakuracloud: %v", err)
	}

	config := NewDefaultConfig()
	config.Token = values["SAKURACLOUD_ACCESS_TOKEN"]
	config.Secret = values["SAKURACLOUD_ACCESS_TOKEN_SECRET"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderCredentials uses the supplied credentials
// to return a DNSProvider instance configured for sakuracloud.
// Deprecated
func NewDNSProviderCredentials(token, secret string) (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.Token = token
	config.Secret = secret

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for GleSYS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("sakuracloud: the configuration of the DNS provider is nil")
	}

	if config.Token == "" {
		return nil, errors.New("sakuracloud: AccessToken is missing")
	}

	if config.Secret == "" {
		return nil, errors.New("sakuracloud: AccessSecret is missing")
	}

	client := api.NewClient(config.Token, config.Secret, "tk1a")
	client.UserAgent = acme.UserAgent

	return &DNSProvider{client: client}, nil
}

// Present creates a TXT record to fulfil the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)

	zone, err := d.getHostedZone(domain)
	if err != nil {
		return fmt.Errorf("sakuracloud: %v", err)
	}

	name := d.extractRecordName(fqdn, zone.Name)

	zone.AddRecord(zone.CreateNewRecord(name, "TXT", value, d.config.TTL))
	_, err = d.client.GetDNSAPI().Update(zone.ID, zone)
	if err != nil {
		return fmt.Errorf("sakuracloud: API call failed: %v", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	zone, err := d.getHostedZone(domain)
	if err != nil {
		return fmt.Errorf("sakuracloud: %v", err)
	}

	records, err := d.findTxtRecords(fqdn, zone)
	if err != nil {
		return fmt.Errorf("sakuracloud: %v", err)
	}

	for _, record := range records {
		var updRecords []sacloud.DNSRecordSet
		for _, r := range zone.Settings.DNS.ResourceRecordSets {
			if !(r.Name == record.Name && r.Type == record.Type && r.RData == record.RData) {
				updRecords = append(updRecords, r)
			}
		}
		zone.Settings.DNS.ResourceRecordSets = updRecords
	}

	_, err = d.client.GetDNSAPI().Update(zone.ID, zone)
	if err != nil {
		return fmt.Errorf("sakuracloud: API call failed: %v", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) getHostedZone(domain string) (*sacloud.DNS, error) {
	authZone, err := acme.FindZoneByFqdn(acme.ToFqdn(domain), acme.RecursiveNameservers)
	if err != nil {
		return nil, err
	}

	zoneName := acme.UnFqdn(authZone)

	res, err := d.client.GetDNSAPI().WithNameLike(zoneName).Find()
	if err != nil {
		if notFound, ok := err.(api.Error); ok && notFound.ResponseCode() == http.StatusNotFound {
			return nil, fmt.Errorf("zone %s not found on SakuraCloud DNS: %v", zoneName, err)
		}
		return nil, fmt.Errorf("API call failed: %v", err)
	}

	for _, zone := range res.CommonServiceDNSItems {
		if zone.Name == zoneName {
			return &zone, nil
		}
	}

	return nil, fmt.Errorf("zone %s not found", zoneName)
}

func (d *DNSProvider) findTxtRecords(fqdn string, zone *sacloud.DNS) ([]sacloud.DNSRecordSet, error) {
	recordName := d.extractRecordName(fqdn, zone.Name)

	var res []sacloud.DNSRecordSet
	for _, record := range zone.Settings.DNS.ResourceRecordSets {
		if record.Name == recordName && record.Type == "TXT" {
			res = append(res, record)
		}
	}
	return res, nil
}

func (d *DNSProvider) extractRecordName(fqdn, domain string) string {
	name := acme.UnFqdn(fqdn)
	if idx := strings.Index(name, "."+domain); idx != -1 {
		return name[:idx]
	}
	return name
}
