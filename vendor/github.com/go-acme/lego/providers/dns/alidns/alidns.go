// Package alidns implements a DNS provider for solving the DNS-01 challenge using Alibaba Cloud DNS.
package alidns

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
)

const defaultRegionID = "cn-hangzhou"

// Config is used to configure the creation of the DNSProvider
type Config struct {
	APIKey             string
	SecretKey          string
	RegionID           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPTimeout        time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("ALICLOUD_TTL", 600),
		PropagationTimeout: env.GetOrDefaultSecond("ALICLOUD_PROPAGATION_TIMEOUT", dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("ALICLOUD_POLLING_INTERVAL", dns01.DefaultPollingInterval),
		HTTPTimeout:        env.GetOrDefaultSecond("ALICLOUD_HTTP_TIMEOUT", 10*time.Second),
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	config *Config
	client *alidns.Client
}

// NewDNSProvider returns a DNSProvider instance configured for Alibaba Cloud DNS.
// Credentials must be passed in the environment variables: ALICLOUD_ACCESS_KEY and ALICLOUD_SECRET_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("ALICLOUD_ACCESS_KEY", "ALICLOUD_SECRET_KEY")
	if err != nil {
		return nil, fmt.Errorf("alicloud: %v", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values["ALICLOUD_ACCESS_KEY"]
	config.SecretKey = values["ALICLOUD_SECRET_KEY"]
	config.RegionID = env.GetOrFile("ALICLOUD_REGION_ID")

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for alidns.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("alicloud: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" || config.SecretKey == "" {
		return nil, fmt.Errorf("alicloud: credentials missing")
	}

	if len(config.RegionID) == 0 {
		config.RegionID = defaultRegionID
	}

	conf := sdk.NewConfig().WithTimeout(config.HTTPTimeout)
	credential := credentials.NewAccessKeyCredential(config.APIKey, config.SecretKey)

	client, err := alidns.NewClientWithOptions(config.RegionID, conf, credential)
	if err != nil {
		return nil, fmt.Errorf("alicloud: credentials failed: %v", err)
	}

	return &DNSProvider{config: config, client: client}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	_, zoneName, err := d.getHostedZone(domain)
	if err != nil {
		return fmt.Errorf("alicloud: %v", err)
	}

	recordAttributes := d.newTxtRecord(zoneName, fqdn, value)

	_, err = d.client.AddDomainRecord(recordAttributes)
	if err != nil {
		return fmt.Errorf("alicloud: API call failed: %v", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	records, err := d.findTxtRecords(domain, fqdn)
	if err != nil {
		return fmt.Errorf("alicloud: %v", err)
	}

	_, _, err = d.getHostedZone(domain)
	if err != nil {
		return fmt.Errorf("alicloud: %v", err)
	}

	for _, rec := range records {
		request := alidns.CreateDeleteDomainRecordRequest()
		request.RecordId = rec.RecordId
		_, err = d.client.DeleteDomainRecord(request)
		if err != nil {
			return fmt.Errorf("alicloud: %v", err)
		}
	}
	return nil
}

func (d *DNSProvider) getHostedZone(domain string) (string, string, error) {
	request := alidns.CreateDescribeDomainsRequest()

	var domains []alidns.Domain
	startPage := 1

	for {
		request.PageNumber = requests.NewInteger(startPage)

		response, err := d.client.DescribeDomains(request)
		if err != nil {
			return "", "", fmt.Errorf("API call failed: %v", err)
		}

		domains = append(domains, response.Domains.Domain...)

		if response.PageNumber*response.PageSize >= response.TotalCount {
			break
		}

		startPage++
	}

	authZone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(domain))
	if err != nil {
		return "", "", err
	}

	var hostedZone alidns.Domain
	for _, zone := range domains {
		if zone.DomainName == dns01.UnFqdn(authZone) {
			hostedZone = zone
		}
	}

	if hostedZone.DomainId == "" {
		return "", "", fmt.Errorf("zone %s not found in AliDNS for domain %s", authZone, domain)
	}
	return fmt.Sprintf("%v", hostedZone.DomainId), hostedZone.DomainName, nil
}

func (d *DNSProvider) newTxtRecord(zone, fqdn, value string) *alidns.AddDomainRecordRequest {
	request := alidns.CreateAddDomainRecordRequest()
	request.Type = "TXT"
	request.DomainName = zone
	request.RR = d.extractRecordName(fqdn, zone)
	request.Value = value
	request.TTL = requests.NewInteger(d.config.TTL)
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
		return records, fmt.Errorf("API call has failed: %v", err)
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
	name := dns01.UnFqdn(fqdn)
	if idx := strings.Index(name, "."+domain); idx != -1 {
		return name[:idx]
	}
	return name
}
