// Package sakuracloud implements a DNS provider for solving the DNS-01 challenge
// using sakuracloud DNS.
package sakuracloud

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/sacloud/libsacloud/api"
	"github.com/sacloud/libsacloud/sacloud"
	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	client *api.Client
}

// NewDNSProvider returns a DNSProvider instance configured for sakuracloud.
// Credentials must be passed in the environment variables: SAKURACLOUD_ACCESS_TOKEN & SAKURACLOUD_ACCESS_TOKEN_SECRET
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("SAKURACLOUD_ACCESS_TOKEN", "SAKURACLOUD_ACCESS_TOKEN_SECRET")
	if err != nil {
		return nil, fmt.Errorf("SakuraCloud: %v", err)
	}

	return NewDNSProviderCredentials(values["SAKURACLOUD_ACCESS_TOKEN"], values["SAKURACLOUD_ACCESS_TOKEN_SECRET"])
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for sakuracloud.
func NewDNSProviderCredentials(token, secret string) (*DNSProvider, error) {
	if token == "" {
		return nil, errors.New("SakuraCloud AccessToken is missing")
	}
	if secret == "" {
		return nil, errors.New("SakuraCloud AccessSecret is missing")
	}

	client := api.NewClient(token, secret, "tk1a")
	client.UserAgent = acme.UserAgent

	return &DNSProvider{client: client}, nil
}

// Present creates a TXT record to fulfil the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)

	zone, err := d.getHostedZone(domain)
	if err != nil {
		return err
	}

	name := d.extractRecordName(fqdn, zone.Name)

	zone.AddRecord(zone.CreateNewRecord(name, "TXT", value, ttl))
	_, err = d.client.GetDNSAPI().Update(zone.ID, zone)
	if err != nil {
		return fmt.Errorf("SakuraCloud API call failed: %v", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	zone, err := d.getHostedZone(domain)
	if err != nil {
		return err
	}

	records, err := d.findTxtRecords(fqdn, zone)
	if err != nil {
		return err
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
		return fmt.Errorf("SakuraCloud API call failed: %v", err)
	}

	return nil
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
		return nil, fmt.Errorf("SakuraCloud API call failed: %v", err)
	}

	for _, zone := range res.CommonServiceDNSItems {
		if zone.Name == zoneName {
			return &zone, nil
		}
	}

	return nil, fmt.Errorf("zone %s not found on SakuraCloud DNS", zoneName)
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
