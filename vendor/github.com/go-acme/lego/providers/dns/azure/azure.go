// Package azure implements a DNS provider for solving the DNS-01 challenge using azure DNS.
// Azure doesn't like trailing dots on domain names, most of the acme code does.
package azure

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/dns/mgmt/2017-09-01/dns"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
)

const defaultMetadataEndpoint = "http://169.254.169.254"

// Config is used to configure the creation of the DNSProvider
type Config struct {
	// optional if using instance metadata service
	ClientID     string
	ClientSecret string
	TenantID     string

	SubscriptionID string
	ResourceGroup  string

	MetadataEndpoint string

	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("AZURE_TTL", 60),
		PropagationTimeout: env.GetOrDefaultSecond("AZURE_PROPAGATION_TIMEOUT", 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond("AZURE_POLLING_INTERVAL", 2*time.Second),
		MetadataEndpoint:   env.GetOrFile("AZURE_METADATA_ENDPOINT"),
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	config     *Config
	authorizer autorest.Authorizer
}

// NewDNSProvider returns a DNSProvider instance configured for azure.
// Credentials can be passed in the environment variables:
// AZURE_CLIENT_ID, AZURE_CLIENT_SECRET, AZURE_SUBSCRIPTION_ID, AZURE_TENANT_ID, AZURE_RESOURCE_GROUP
// If the credentials are _not_ set via the environment,
// then it will attempt to get a bearer token via the instance metadata service.
// see: https://github.com/Azure/go-autorest/blob/v10.14.0/autorest/azure/auth/auth.go#L38-L42
func NewDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.SubscriptionID = env.GetOrFile("AZURE_SUBSCRIPTION_ID")
	config.ResourceGroup = env.GetOrFile("AZURE_RESOURCE_GROUP")

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Azure.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("azure: the configuration of the DNS provider is nil")
	}

	if config.HTTPClient == nil {
		config.HTTPClient = http.DefaultClient
	}

	authorizer, err := getAuthorizer(config)
	if err != nil {
		return nil, err
	}

	if config.SubscriptionID == "" {
		subsID, err := getMetadata(config, "subscriptionId")
		if err != nil {
			return nil, fmt.Errorf("azure: %v", err)
		}

		if subsID == "" {
			return nil, errors.New("azure: SubscriptionID is missing")
		}
		config.SubscriptionID = subsID
	}

	if config.ResourceGroup == "" {
		resGroup, err := getMetadata(config, "resourceGroupName")
		if err != nil {
			return nil, fmt.Errorf("azure: %v", err)
		}

		if resGroup == "" {
			return nil, errors.New("azure: ResourceGroup is missing")
		}
		config.ResourceGroup = resGroup
	}

	return &DNSProvider{config: config, authorizer: authorizer}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	zone, err := d.getHostedZoneID(ctx, fqdn)
	if err != nil {
		return fmt.Errorf("azure: %v", err)
	}

	rsc := dns.NewRecordSetsClient(d.config.SubscriptionID)
	rsc.Authorizer = d.authorizer

	relative := toRelativeRecord(fqdn, dns01.ToFqdn(zone))

	// Get existing record set
	rset, err := rsc.Get(ctx, d.config.ResourceGroup, zone, relative, dns.TXT)
	if err != nil {
		detailedError, ok := err.(autorest.DetailedError)
		if !ok || detailedError.StatusCode != http.StatusNotFound {
			return fmt.Errorf("azure: %v", err)
		}
	}

	// Construct unique TXT records using map
	uniqRecords := map[string]struct{}{value: {}}
	if rset.RecordSetProperties != nil && rset.TxtRecords != nil {
		for _, txtRecord := range *rset.TxtRecords {
			// Assume Value doesn't contain multiple strings
			if txtRecord.Value != nil && len(*txtRecord.Value) > 0 {
				uniqRecords[(*txtRecord.Value)[0]] = struct{}{}
			}
		}
	}

	var txtRecords []dns.TxtRecord
	for txt := range uniqRecords {
		txtRecords = append(txtRecords, dns.TxtRecord{Value: &[]string{txt}})
	}

	rec := dns.RecordSet{
		Name: &relative,
		RecordSetProperties: &dns.RecordSetProperties{
			TTL:        to.Int64Ptr(int64(d.config.TTL)),
			TxtRecords: &txtRecords,
		},
	}

	_, err = rsc.CreateOrUpdate(ctx, d.config.ResourceGroup, zone, relative, dns.TXT, rec, "", "")
	if err != nil {
		return fmt.Errorf("azure: %v", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	zone, err := d.getHostedZoneID(ctx, fqdn)
	if err != nil {
		return fmt.Errorf("azure: %v", err)
	}

	relative := toRelativeRecord(fqdn, dns01.ToFqdn(zone))
	rsc := dns.NewRecordSetsClient(d.config.SubscriptionID)
	rsc.Authorizer = d.authorizer

	_, err = rsc.Delete(ctx, d.config.ResourceGroup, zone, relative, dns.TXT, "")
	if err != nil {
		return fmt.Errorf("azure: %v", err)
	}
	return nil
}

// Checks that azure has a zone for this domain name.
func (d *DNSProvider) getHostedZoneID(ctx context.Context, fqdn string) (string, error) {
	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return "", err
	}

	dc := dns.NewZonesClient(d.config.SubscriptionID)
	dc.Authorizer = d.authorizer

	zone, err := dc.Get(ctx, d.config.ResourceGroup, dns01.UnFqdn(authZone))
	if err != nil {
		return "", err
	}

	// zone.Name shouldn't have a trailing dot(.)
	return to.String(zone.Name), nil
}

// Returns the relative record to the domain
func toRelativeRecord(domain, zone string) string {
	return dns01.UnFqdn(strings.TrimSuffix(domain, zone))
}

func getAuthorizer(config *Config) (autorest.Authorizer, error) {
	if config.ClientID != "" && config.ClientSecret != "" && config.TenantID != "" {
		oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, config.TenantID)
		if err != nil {
			return nil, err
		}

		spt, err := adal.NewServicePrincipalToken(*oauthConfig, config.ClientID, config.ClientSecret, azure.PublicCloud.ResourceManagerEndpoint)
		if err != nil {
			return nil, err
		}

		spt.SetSender(config.HTTPClient)
		return autorest.NewBearerAuthorizer(spt), nil
	}

	return auth.NewAuthorizerFromEnvironment()
}

// Fetches metadata from environment or he instance metadata service
// borrowed from https://github.com/Microsoft/azureimds/blob/master/imdssample.go
func getMetadata(config *Config, field string) (string, error) {
	metadataEndpoint := config.MetadataEndpoint
	if len(metadataEndpoint) == 0 {
		metadataEndpoint = defaultMetadataEndpoint
	}

	resource := fmt.Sprintf("%s/metadata/instance/compute/%s", metadataEndpoint, field)
	req, err := http.NewRequest(http.MethodGet, resource, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Metadata", "True")

	q := req.URL.Query()
	q.Add("format", "text")
	q.Add("api-version", "2017-12-01")
	req.URL.RawQuery = q.Encode()

	resp, err := config.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(respBody), nil
}
