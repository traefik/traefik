// Package dnsimple implements a DNS provider for solving the DNS-01 challenge using dnsimple DNS.
package dnsimple

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dnsimple/dnsimple-go/dnsimple"
	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
	"golang.org/x/oauth2"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	AccessToken        string
	BaseURL            string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("DNSIMPLE_TTL", dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond("DNSIMPLE_PROPAGATION_TIMEOUT", dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("DNSIMPLE_POLLING_INTERVAL", dns01.DefaultPollingInterval),
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	config *Config
	client *dnsimple.Client
}

// NewDNSProvider returns a DNSProvider instance configured for dnsimple.
// Credentials must be passed in the environment variables: DNSIMPLE_OAUTH_TOKEN.
//
// See: https://developer.dnsimple.com/v2/#authentication
func NewDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.AccessToken = env.GetOrFile("DNSIMPLE_OAUTH_TOKEN")
	config.BaseURL = env.GetOrFile("DNSIMPLE_BASE_URL")

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for DNSimple.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("dnsimple: the configuration of the DNS provider is nil")
	}

	if config.AccessToken == "" {
		return nil, fmt.Errorf("dnsimple: OAuth token is missing")
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: config.AccessToken})
	client := dnsimple.NewClient(oauth2.NewClient(context.Background(), ts))

	if config.BaseURL != "" {
		client.BaseURL = config.BaseURL
	}

	return &DNSProvider{client: client, config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	zoneName, err := d.getHostedZone(domain)
	if err != nil {
		return fmt.Errorf("dnsimple: %v", err)
	}

	accountID, err := d.getAccountID()
	if err != nil {
		return fmt.Errorf("dnsimple: %v", err)
	}

	recordAttributes := newTxtRecord(zoneName, fqdn, value, d.config.TTL)
	_, err = d.client.Zones.CreateRecord(accountID, zoneName, recordAttributes)
	if err != nil {
		return fmt.Errorf("dnsimple: API call failed: %v", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	records, err := d.findTxtRecords(domain, fqdn)
	if err != nil {
		return fmt.Errorf("dnsimple: %v", err)
	}

	accountID, err := d.getAccountID()
	if err != nil {
		return fmt.Errorf("dnsimple: %v", err)
	}

	var lastErr error
	for _, rec := range records {
		_, err := d.client.Zones.DeleteRecord(accountID, rec.ZoneID, rec.ID)
		if err != nil {
			lastErr = fmt.Errorf("dnsimple: %v", err)
		}
	}

	return lastErr
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) getHostedZone(domain string) (string, error) {
	authZone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(domain))
	if err != nil {
		return "", err
	}

	accountID, err := d.getAccountID()
	if err != nil {
		return "", err
	}

	zoneName := dns01.UnFqdn(authZone)

	zones, err := d.client.Zones.ListZones(accountID, &dnsimple.ZoneListOptions{NameLike: zoneName})
	if err != nil {
		return "", fmt.Errorf("API call failed: %v", err)
	}

	var hostedZone dnsimple.Zone
	for _, zone := range zones.Data {
		if zone.Name == zoneName {
			hostedZone = zone
		}
	}

	if hostedZone.ID == 0 {
		return "", fmt.Errorf("zone %s not found in DNSimple for domain %s", authZone, domain)
	}

	return hostedZone.Name, nil
}

func (d *DNSProvider) findTxtRecords(domain, fqdn string) ([]dnsimple.ZoneRecord, error) {
	zoneName, err := d.getHostedZone(domain)
	if err != nil {
		return nil, err
	}

	accountID, err := d.getAccountID()
	if err != nil {
		return nil, err
	}

	recordName := extractRecordName(fqdn, zoneName)

	result, err := d.client.Zones.ListRecords(accountID, zoneName, &dnsimple.ZoneRecordListOptions{Name: recordName, Type: "TXT", ListOptions: dnsimple.ListOptions{}})
	if err != nil {
		return nil, fmt.Errorf("API call has failed: %v", err)
	}

	return result.Data, nil
}

func newTxtRecord(zoneName, fqdn, value string, ttl int) dnsimple.ZoneRecord {
	name := extractRecordName(fqdn, zoneName)

	return dnsimple.ZoneRecord{
		Type:    "TXT",
		Name:    name,
		Content: value,
		TTL:     ttl,
	}
}

func extractRecordName(fqdn, domain string) string {
	name := dns01.UnFqdn(fqdn)
	if idx := strings.Index(name, "."+domain); idx != -1 {
		return name[:idx]
	}
	return name
}

func (d *DNSProvider) getAccountID() (string, error) {
	whoamiResponse, err := d.client.Identity.Whoami()
	if err != nil {
		return "", err
	}

	if whoamiResponse.Data.Account == nil {
		return "", fmt.Errorf("user tokens are not supported, please use an account token")
	}

	return strconv.FormatInt(whoamiResponse.Data.Account.ID, 10), nil
}
