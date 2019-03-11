package oraclecloud

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/dns"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	CompartmentID      string
	OCIConfigProvider  common.ConfigurationProvider
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("OCI_TTL", dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond("OCI_PROPAGATION_TIMEOUT", dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("OCI_POLLING_INTERVAL", dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("OCI_HTTP_TIMEOUT", 60*time.Second),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	client *dns.DnsClient
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for OracleCloud.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(ociPrivkey, ociTenancyOCID, ociUserOCID, ociPubkeyFingerprint, ociRegion, "OCI_COMPARTMENT_OCID")
	if err != nil {
		return nil, fmt.Errorf("oraclecloud: %v", err)
	}

	config := NewDefaultConfig()
	config.CompartmentID = values["OCI_COMPARTMENT_OCID"]
	config.OCIConfigProvider = newConfigProvider(values)

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for OracleCloud.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("oraclecloud: the configuration of the DNS provider is nil")
	}

	if config.CompartmentID == "" {
		return nil, errors.New("oraclecloud: CompartmentID is missing")
	}

	if config.OCIConfigProvider == nil {
		return nil, errors.New("oraclecloud: OCIConfigProvider is missing")
	}

	client, err := dns.NewDnsClientWithConfigurationProvider(config.OCIConfigProvider)
	if err != nil {
		return nil, fmt.Errorf("oraclecloud: %v", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	return &DNSProvider{client: &client, config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	// generate request to dns.PatchDomainRecordsRequest
	recordOperation := dns.RecordOperation{
		Domain:      common.String(dns01.UnFqdn(fqdn)),
		Rdata:       common.String(value),
		Rtype:       common.String("TXT"),
		Ttl:         common.Int(d.config.TTL),
		IsProtected: common.Bool(false),
	}

	request := dns.PatchDomainRecordsRequest{
		CompartmentId: common.String(d.config.CompartmentID),
		ZoneNameOrId:  common.String(domain),
		Domain:        common.String(dns01.UnFqdn(fqdn)),
		PatchDomainRecordsDetails: dns.PatchDomainRecordsDetails{
			Items: []dns.RecordOperation{recordOperation},
		},
	}

	_, err := d.client.PatchDomainRecords(context.Background(), request)
	if err != nil {
		return fmt.Errorf("oraclecloud: %v", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	// search to TXT record's hash to delete
	getRequest := dns.GetDomainRecordsRequest{
		ZoneNameOrId:  common.String(domain),
		Domain:        common.String(dns01.UnFqdn(fqdn)),
		CompartmentId: common.String(d.config.CompartmentID),
		Rtype:         common.String("TXT"),
	}

	ctx := context.Background()

	domainRecords, err := d.client.GetDomainRecords(ctx, getRequest)
	if err != nil {
		return fmt.Errorf("oraclecloud: %v", err)
	}

	if *domainRecords.OpcTotalItems == 0 {
		return fmt.Errorf("oraclecloud: no record to CleanUp")
	}

	var deleteHash *string
	for _, record := range domainRecords.RecordCollection.Items {
		if record.Rdata != nil && *record.Rdata == `"`+value+`"` {
			deleteHash = record.RecordHash
			break
		}
	}

	if deleteHash == nil {
		return fmt.Errorf("oraclecloud: no record to CleanUp")
	}

	recordOperation := dns.RecordOperation{
		RecordHash: deleteHash,
		Operation:  dns.RecordOperationOperationRemove,
	}

	patchRequest := dns.PatchDomainRecordsRequest{
		ZoneNameOrId: common.String(domain),
		Domain:       common.String(dns01.UnFqdn(fqdn)),
		PatchDomainRecordsDetails: dns.PatchDomainRecordsDetails{
			Items: []dns.RecordOperation{recordOperation},
		},
		CompartmentId: common.String(d.config.CompartmentID),
	}

	_, err = d.client.PatchDomainRecords(ctx, patchRequest)
	if err != nil {
		return fmt.Errorf("oraclecloud: %v", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
