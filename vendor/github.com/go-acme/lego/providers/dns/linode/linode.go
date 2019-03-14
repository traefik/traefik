// Package linode implements a DNS provider for solving the DNS-01 challenge using Linode DNS.
package linode

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
	"github.com/timewasted/linode/dns"
)

const (
	minTTL             = 300
	dnsUpdateFreqMins  = 15
	dnsUpdateFudgeSecs = 120
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	APIKey          string
	PollingInterval time.Duration
	TTL             int
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		PollingInterval: env.GetOrDefaultSecond("LINODE_POLLING_INTERVAL", 15*time.Second),
		TTL:             env.GetOrDefaultInt("LINODE_TTL", minTTL),
	}
}

type hostedZoneInfo struct {
	domainID     int
	resourceName string
}

// DNSProvider implements the acme.ChallengeProvider interface.
type DNSProvider struct {
	config *Config
	client *dns.DNS
}

// NewDNSProvider returns a DNSProvider instance configured for Linode.
// Credentials must be passed in the environment variable: LINODE_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("LINODE_API_KEY")
	if err != nil {
		return nil, fmt.Errorf("linode: %v", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values["LINODE_API_KEY"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Linode.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("linode: the configuration of the DNS provider is nil")
	}

	if len(config.APIKey) == 0 {
		return nil, errors.New("linode: credentials missing")
	}

	if config.TTL < minTTL {
		return nil, fmt.Errorf("linode: invalid TTL, TTL (%d) must be greater than %d", config.TTL, minTTL)
	}

	return &DNSProvider{
		config: config,
		client: dns.New(config.APIKey),
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS
// propagation.  Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	// Since Linode only updates their zone files every X minutes, we need
	// to figure out how many minutes we have to wait until we hit the next
	// interval of X.  We then wait another couple of minutes, just to be
	// safe.  Hopefully at some point during all of this, the record will
	// have propagated throughout Linode's network.
	minsRemaining := dnsUpdateFreqMins - (time.Now().Minute() % dnsUpdateFreqMins)

	timeout = (time.Duration(minsRemaining) * time.Minute) +
		(minTTL * time.Second) +
		(dnsUpdateFudgeSecs * time.Second)
	interval = d.config.PollingInterval
	return
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	zone, err := d.getHostedZoneInfo(fqdn)
	if err != nil {
		return err
	}

	if _, err = d.client.CreateDomainResourceTXT(zone.domainID, dns01.UnFqdn(fqdn), value, d.config.TTL); err != nil {
		return err
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	zone, err := d.getHostedZoneInfo(fqdn)
	if err != nil {
		return err
	}

	// Get all TXT records for the specified domain.
	resources, err := d.client.GetResourcesByType(zone.domainID, "TXT")
	if err != nil {
		return err
	}

	// Remove the specified resource, if it exists.
	for _, resource := range resources {
		if resource.Name == zone.resourceName && resource.Target == value {
			resp, err := d.client.DeleteDomainResource(resource.DomainID, resource.ResourceID)
			if err != nil {
				return err
			}

			if resp.ResourceID != resource.ResourceID {
				return errors.New("error deleting resource: resource IDs do not match")
			}
			break
		}
	}

	return nil
}

func (d *DNSProvider) getHostedZoneInfo(fqdn string) (*hostedZoneInfo, error) {
	// Lookup the zone that handles the specified FQDN.
	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return nil, err
	}

	resourceName := strings.TrimSuffix(fqdn, "."+authZone)

	// Query the authority zone.
	domain, err := d.client.GetDomain(dns01.UnFqdn(authZone))
	if err != nil {
		return nil, err
	}

	return &hostedZoneInfo{
		domainID:     domain.DomainID,
		resourceName: resourceName,
	}, nil
}
