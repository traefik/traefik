// Package alidns implements a DNS provider for solving the DNS-01 challenge
// using Alibaba Cloud DNS.
package alidns

import (
	"fmt"
	"os"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

const defaultRegionID = "cn-hangzhou"

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	client *alidns.Client
}

// NewDNSProvider returns a DNSProvider instance configured for Alibaba Cloud DNS.
// Credentials must be passed in the environment variables: ALICLOUD_ACCESS_KEY and ALICLOUD_SECRET_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("ALICLOUD_ACCESS_KEY", "ALICLOUD_SECRET_KEY")
	if err != nil {
		return nil, fmt.Errorf("AliDNS: %v", err)
	}

	regionID := os.Getenv("ALICLOUD_REGION_ID")

	return NewDNSProviderCredentials(values["ALICLOUD_ACCESS_KEY"], values["ALICLOUD_SECRET_KEY"], regionID)
}

// NewDNSProviderCredentials uses the supplied credentials to return a DNSProvider instance configured for alidns.
func NewDNSProviderCredentials(apiKey, secretKey, regionID string) (*DNSProvider, error) {
	if apiKey == "" || secretKey == "" {
		return nil, fmt.Errorf("AliDNS: credentials missing")
	}

	if len(regionID) == 0 {
		regionID = defaultRegionID
	}

	client, err := alidns.NewClientWithAccessKey(regionID, apiKey, secretKey)
	if err != nil {
		return nil, fmt.Errorf("AliDNS: credentials failed: %v", err)
	}

	return &DNSProvider{
		client: client,
	}, nil
}

// Present creates a TXT record to fulfil the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)

	_, zoneName, err := d.getHostedZone(domain)
	if err != nil {
		return err
	}

	recordAttributes := d.newTxtRecord(zoneName, fqdn, value, ttl)

	_, err = d.client.AddDomainRecord(recordAttributes)
	if err != nil {
		return fmt.Errorf("AliDNS: API call failed: %v", err)
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

	_, _, err = d.getHostedZone(domain)
	if err != nil {
		return err
	}

	for _, rec := range records {
		request := alidns.CreateDeleteDomainRecordRequest()
		request.RecordId = rec.RecordId
		_, err = d.client.DeleteDomainRecord(request)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DNSProvider) getHostedZone(domain string) (string, string, error) {
	request := alidns.CreateDescribeDomainsRequest()
	zones, err := d.client.DescribeDomains(request)
	if err != nil {
		return "", "", fmt.Errorf("AliDNS: API call failed: %v", err)
	}

	authZone, err := acme.FindZoneByFqdn(acme.ToFqdn(domain), acme.RecursiveNameservers)
	if err != nil {
		return "", "", err
	}

	var hostedZone alidns.Domain
	for _, zone := range zones.Domains.Domain {
		if zone.DomainName == acme.UnFqdn(authZone) {
			hostedZone = zone
		}
	}

	if hostedZone.DomainId == "" {
		return "", "", fmt.Errorf("AliDNS: zone %s not found in AliDNS for domain %s", authZone, domain)
	}
	return fmt.Sprintf("%v", hostedZone.DomainId), hostedZone.DomainName, nil
}

func (d *DNSProvider) newTxtRecord(zone, fqdn, value string, ttl int) *alidns.AddDomainRecordRequest {
	request := alidns.CreateAddDomainRecordRequest()
	request.Type = "TXT"
	request.DomainName = zone
	request.RR = d.extractRecordName(fqdn, zone)
	request.Value = value
	request.TTL = requests.NewInteger(600)
	return request
}

func (d *DNSProvider) findTxtRecords(domain, fqdn string) ([]alidns.Record, error) {
	_, zoneName, err := d.getHostedZone(domain)
	if err != nil {
		return nil, err
	}

	request := alidns.CreateDescribeDomainRecordsRequest()
	request.DomainName = zoneName
	request.PageSize = requests.NewInteger(500)

	var records []alidns.Record

	result, err := d.client.DescribeDomainRecords(request)
	if err != nil {
		return records, fmt.Errorf("AliDNS: API call has failed: %v", err)
	}

	recordName := d.extractRecordName(fqdn, zoneName)
	for _, record := range result.DomainRecords.Record {
		if record.RR == recordName {
			records = append(records, record)
		}
	}
	return records, nil
}

func (d *DNSProvider) extractRecordName(fqdn, domain string) string {
	name := acme.UnFqdn(fqdn)
	if idx := strings.Index(name, "."+domain); idx != -1 {
		return name[:idx]
	}
	return name
}
