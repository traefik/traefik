// Package azure implements a DNS provider for solving the DNS-01
// challenge using azure DNS.
// Azure doesn't like trailing dots on domain names, most of the acme code does.
package azure

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/dns/mgmt/2017-09-01/dns"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	ClientID           string
	ClientSecret       string
	SubscriptionID     string
	TenantID           string
	ResourceGroup      string
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
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for azure.
// Credentials must be passed in the environment variables: AZURE_CLIENT_ID,
// AZURE_CLIENT_SECRET, AZURE_SUBSCRIPTION_ID, AZURE_TENANT_ID, AZURE_RESOURCE_GROUP
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("AZURE_CLIENT_ID", "AZURE_CLIENT_SECRET", "AZURE_SUBSCRIPTION_ID", "AZURE_TENANT_ID", "AZURE_RESOURCE_GROUP")
	if err != nil {
		return nil, fmt.Errorf("azure: %v", err)
	}

	config := NewDefaultConfig()
	config.ClientID = values["AZURE_CLIENT_ID"]
	config.ClientSecret = values["AZURE_CLIENT_SECRET"]
	config.SubscriptionID = values["AZURE_SUBSCRIPTION_ID"]
	config.TenantID = values["AZURE_TENANT_ID"]
	config.ResourceGroup = values["AZURE_RESOURCE_GROUP"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderCredentials uses the supplied credentials
// to return a DNSProvider instance configured for azure.
// Deprecated
func NewDNSProviderCredentials(clientID, clientSecret, subscriptionID, tenantID, resourceGroup string) (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.ClientID = clientID
	config.ClientSecret = clientSecret
	config.SubscriptionID = subscriptionID
	config.TenantID = tenantID
	config.ResourceGroup = resourceGroup

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Azure.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("azure: the configuration of the DNS provider is nil")
	}

	if config.ClientID == "" || config.ClientSecret == "" || config.SubscriptionID == "" || config.TenantID == "" || config.ResourceGroup == "" {
		return nil, errors.New("azure: some credentials information are missing")
	}

	return &DNSProvider{config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS
// propagation. Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfil the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)

	zone, err := d.getHostedZoneID(ctx, fqdn)
	if err != nil {
		return fmt.Errorf("azure: %v", err)
	}

	rsc := dns.NewRecordSetsClient(d.config.SubscriptionID)
	spt, err := d.newServicePrincipalToken(azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		return fmt.Errorf("azure: %v", err)
	}

	rsc.Authorizer = autorest.NewBearerAuthorizer(spt)

	relative := toRelativeRecord(fqdn, acme.ToFqdn(zone))
	rec := dns.RecordSet{
		Name: &relative,
		RecordSetProperties: &dns.RecordSetProperties{
			TTL:        to.Int64Ptr(int64(d.config.TTL)),
			TxtRecords: &[]dns.TxtRecord{{Value: &[]string{value}}},
		},
	}

	_, err = rsc.CreateOrUpdate(ctx, d.config.ResourceGroup, zone, relative, dns.TXT, rec, "", "")
	return fmt.Errorf("azure: %v", err)
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	zone, err := d.getHostedZoneID(ctx, fqdn)
	if err != nil {
		return fmt.Errorf("azure: %v", err)
	}

	relative := toRelativeRecord(fqdn, acme.ToFqdn(zone))
	rsc := dns.NewRecordSetsClient(d.config.SubscriptionID)
	spt, err := d.newServicePrincipalToken(azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		return fmt.Errorf("azure: %v", err)
	}

	rsc.Authorizer = autorest.NewBearerAuthorizer(spt)

	_, err = rsc.Delete(ctx, d.config.ResourceGroup, zone, relative, dns.TXT, "")
	return fmt.Errorf("azure: %v", err)
}

// Checks that azure has a zone for this domain name.
func (d *DNSProvider) getHostedZoneID(ctx context.Context, fqdn string) (string, error) {
	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return "", err
	}

	// Now we want to to Azure and get the zone.
	spt, err := d.newServicePrincipalToken(azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		return "", err
	}

	dc := dns.NewZonesClient(d.config.SubscriptionID)
	dc.Authorizer = autorest.NewBearerAuthorizer(spt)

	zone, err := dc.Get(ctx, d.config.ResourceGroup, acme.UnFqdn(authZone))
	if err != nil {
		return "", err
	}

	// zone.Name shouldn't have a trailing dot(.)
	return to.String(zone.Name), nil
}

// NewServicePrincipalTokenFromCredentials creates a new ServicePrincipalToken using values of the
// passed credentials map.
func (d *DNSProvider) newServicePrincipalToken(scope string) (*adal.ServicePrincipalToken, error) {
	oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, d.config.TenantID)
	if err != nil {
		return nil, err
	}
	return adal.NewServicePrincipalToken(*oauthConfig, d.config.ClientID, d.config.ClientSecret, scope)
}

// Returns the relative record to the domain
func toRelativeRecord(domain, zone string) string {
	return acme.UnFqdn(strings.TrimSuffix(domain, zone))
}
