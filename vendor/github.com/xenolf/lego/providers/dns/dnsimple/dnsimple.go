// Package dnsimple implements a DNS provider for solving the DNS-01 challenge
// using dnsimple DNS.
package dnsimple

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dnsimple/dnsimple-go/dnsimple"
	"github.com/xenolf/lego/acme"
)

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	client *dnsimple.Client
}

// NewDNSProvider returns a DNSProvider instance configured for dnsimple.
// Credentials must be passed in the environment variables: DNSIMPLE_OAUTH_TOKEN.
//
// See: https://developer.dnsimple.com/v2/#authentication
func NewDNSProvider() (*DNSProvider, error) {
	accessToken := os.Getenv("DNSIMPLE_OAUTH_TOKEN")
	baseURL := os.Getenv("DNSIMPLE_BASE_URL")

	return NewDNSProviderCredentials(accessToken, baseURL)
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for dnsimple.
func NewDNSProviderCredentials(accessToken, baseURL string) (*DNSProvider, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("DNSimple OAuth token is missing")
	}

	client := dnsimple.NewClient(dnsimple.NewOauthTokenCredentials(accessToken))
	client.UserAgent = "lego"

	if baseURL != "" {
		client.BaseURL = baseURL
	}

	return &DNSProvider{client: client}, nil
}

// Present creates a TXT record to fulfil the dns-01 challenge.
func (c *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)

	zoneName, err := c.getHostedZone(domain)

	if err != nil {
		return err
	}

	accountID, err := c.getAccountID()
	if err != nil {
		return err
	}

	recordAttributes := c.newTxtRecord(zoneName, fqdn, value, ttl)
	_, err = c.client.Zones.CreateRecord(accountID, zoneName, *recordAttributes)
	if err != nil {
		return fmt.Errorf("DNSimple API call failed: %v", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (c *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	records, err := c.findTxtRecords(domain, fqdn)
	if err != nil {
		return err
	}

	accountID, err := c.getAccountID()
	if err != nil {
		return err
	}

	for _, rec := range records {
		_, err := c.client.Zones.DeleteRecord(accountID, rec.ZoneID, rec.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *DNSProvider) getHostedZone(domain string) (string, error) {
	authZone, err := acme.FindZoneByFqdn(acme.ToFqdn(domain), acme.RecursiveNameservers)
	if err != nil {
		return "", err
	}

	accountID, err := c.getAccountID()
	if err != nil {
		return "", err
	}

	zoneName := acme.UnFqdn(authZone)

	zones, err := c.client.Zones.ListZones(accountID, &dnsimple.ZoneListOptions{NameLike: zoneName})
	if err != nil {
		return "", fmt.Errorf("DNSimple API call failed: %v", err)
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

func (c *DNSProvider) findTxtRecords(domain, fqdn string) ([]dnsimple.ZoneRecord, error) {
	zoneName, err := c.getHostedZone(domain)
	if err != nil {
		return nil, err
	}

	accountID, err := c.getAccountID()
	if err != nil {
		return nil, err
	}

	recordName := c.extractRecordName(fqdn, zoneName)

	result, err := c.client.Zones.ListRecords(accountID, zoneName, &dnsimple.ZoneRecordListOptions{Name: recordName, Type: "TXT", ListOptions: dnsimple.ListOptions{}})
	if err != nil {
		return []dnsimple.ZoneRecord{}, fmt.Errorf("DNSimple API call has failed: %v", err)
	}

	return result.Data, nil
}

func (c *DNSProvider) newTxtRecord(zoneName, fqdn, value string, ttl int) *dnsimple.ZoneRecord {
	name := c.extractRecordName(fqdn, zoneName)

	return &dnsimple.ZoneRecord{
		Type:    "TXT",
		Name:    name,
		Content: value,
		TTL:     ttl,
	}
}

func (c *DNSProvider) extractRecordName(fqdn, domain string) string {
	name := acme.UnFqdn(fqdn)
	if idx := strings.Index(name, "."+domain); idx != -1 {
		return name[:idx]
	}
	return name
}

func (c *DNSProvider) getAccountID() (string, error) {
	whoamiResponse, err := c.client.Identity.Whoami()
	if err != nil {
		return "", err
	}

	if whoamiResponse.Data.Account == nil {
		return "", fmt.Errorf("DNSimple user tokens are not supported, please use an account token")
	}

	return strconv.FormatInt(whoamiResponse.Data.Account.ID, 10), nil
}
