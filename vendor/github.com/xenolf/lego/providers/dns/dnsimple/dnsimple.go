// Package dnsimple implements a DNS provider for solving the DNS-01 challenge
// using dnsimple DNS.
package dnsimple

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dnsimple/dnsimple-go/dnsimple"
	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
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
		TTL:                env.GetOrDefaultInt("DNSIMPLE_TTL", 120),
		PropagationTimeout: env.GetOrDefaultSecond("DNSIMPLE_PROPAGATION_TIMEOUT", acme.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("DNSIMPLE_POLLING_INTERVAL", acme.DefaultPollingInterval),
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
	config.AccessToken = os.Getenv("DNSIMPLE_OAUTH_TOKEN")
	config.BaseURL = os.Getenv("DNSIMPLE_BASE_URL")

	return NewDNSProviderConfig(config)
}

// NewDNSProviderCredentials uses the supplied credentials
// to return a DNSProvider instance configured for DNSimple.
// Deprecated
func NewDNSProviderCredentials(accessToken, baseURL string) (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.AccessToken = accessToken
	config.BaseURL = baseURL

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

	client := dnsimple.NewClient(dnsimple.NewOauthTokenCredentials(config.AccessToken))
	client.UserAgent = acme.UserAgent

	if config.BaseURL != "" {
		client.BaseURL = config.BaseURL
	}

	return &DNSProvider{client: client}, nil
}

// Present creates a TXT record to fulfil the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)

	zoneName, err := d.getHostedZone(domain)
	if err != nil {
		return err
	}

	accountID, err := d.getAccountID()
	if err != nil {
		return err
	}

	recordAttributes := d.newTxtRecord(zoneName, fqdn, value, d.config.TTL)
	_, err = d.client.Zones.CreateRecord(accountID, zoneName, recordAttributes)
	if err != nil {
		return fmt.Errorf("API call failed: %v", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	records, err := d.findTxtRecords(domain, fqdn)
	if err != nil {
		return err
	}

	accountID, err := d.getAccountID()
	if err != nil {
		return err
	}

	for _, rec := range records {
		_, err := d.client.Zones.DeleteRecord(accountID, rec.ZoneID, rec.ID)
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

func (d *DNSProvider) getHostedZone(domain string) (string, error) {
	authZone, err := acme.FindZoneByFqdn(acme.ToFqdn(domain), acme.RecursiveNameservers)
	if err != nil {
		return "", err
	}

	accountID, err := d.getAccountID()
	if err != nil {
		return "", err
	}

	zoneName := acme.UnFqdn(authZone)

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

	recordName := d.extractRecordName(fqdn, zoneName)

	result, err := d.client.Zones.ListRecords(accountID, zoneName, &dnsimple.ZoneRecordListOptions{Name: recordName, Type: "TXT", ListOptions: dnsimple.ListOptions{}})
	if err != nil {
		return []dnsimple.ZoneRecord{}, fmt.Errorf("API call has failed: %v", err)
	}

	return result.Data, nil
}

func (d *DNSProvider) newTxtRecord(zoneName, fqdn, value string, ttl int) dnsimple.ZoneRecord {
	name := d.extractRecordName(fqdn, zoneName)

	return dnsimple.ZoneRecord{
		Type:    "TXT",
		Name:    name,
		Content: value,
		TTL:     ttl,
	}
}

func (d *DNSProvider) extractRecordName(fqdn, domain string) string {
	name := acme.UnFqdn(fqdn)
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
