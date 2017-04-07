// Package azure implements a DNS provider for solving the DNS-01
// challenge using azure DNS.
// Azure doesn't like trailing dots on domain names, most of the acme code does.
package azure

import (
	"fmt"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/arm/dns"

	"strings"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/xenolf/lego/acme"
)

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	clientId       string
	clientSecret   string
	subscriptionId string
	tenantId       string
	resourceGroup  string
}

// NewDNSProvider returns a DNSProvider instance configured for azure.
// Credentials must be passed in the environment variables: AZURE_CLIENT_ID,
// AZURE_CLIENT_SECRET, AZURE_SUBSCRIPTION_ID, AZURE_TENANT_ID
func NewDNSProvider() (*DNSProvider, error) {
	clientId := os.Getenv("AZURE_CLIENT_ID")
	clientSecret := os.Getenv("AZURE_CLIENT_SECRET")
	subscriptionId := os.Getenv("AZURE_SUBSCRIPTION_ID")
	tenantId := os.Getenv("AZURE_TENANT_ID")
	resourceGroup := os.Getenv("AZURE_RESOURCE_GROUP")
	return NewDNSProviderCredentials(clientId, clientSecret, subscriptionId, tenantId, resourceGroup)
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for azure.
func NewDNSProviderCredentials(clientId, clientSecret, subscriptionId, tenantId, resourceGroup string) (*DNSProvider, error) {
	if clientId == "" || clientSecret == "" || subscriptionId == "" || tenantId == "" || resourceGroup == "" {
		return nil, fmt.Errorf("Azure configuration missing")
	}

	return &DNSProvider{
		clientId:       clientId,
		clientSecret:   clientSecret,
		subscriptionId: subscriptionId,
		tenantId:       tenantId,
		resourceGroup:  resourceGroup,
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS
// propagation. Adjusting here to cope with spikes in propagation times.
func (c *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return 120 * time.Second, 2 * time.Second
}

// Present creates a TXT record to fulfil the dns-01 challenge
func (c *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)
	zone, err := c.getHostedZoneID(fqdn)
	if err != nil {
		return err
	}

	rsc := dns.NewRecordSetsClient(c.subscriptionId)
	rsc.Authorizer, err = c.newServicePrincipalTokenFromCredentials(azure.PublicCloud.ResourceManagerEndpoint)
	relative := toRelativeRecord(fqdn, acme.ToFqdn(zone))
	rec := dns.RecordSet{
		Name: &relative,
		RecordSetProperties: &dns.RecordSetProperties{
			TTL:        to.Int64Ptr(60),
			TxtRecords: &[]dns.TxtRecord{dns.TxtRecord{Value: &[]string{value}}},
		},
	}
	_, err = rsc.CreateOrUpdate(c.resourceGroup, zone, relative, dns.TXT, rec, "", "")

	if err != nil {
		return err
	}

	return nil
}

// Returns the relative record to the domain
func toRelativeRecord(domain, zone string) string {
	return acme.UnFqdn(strings.TrimSuffix(domain, zone))
}

// CleanUp removes the TXT record matching the specified parameters
func (c *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	zone, err := c.getHostedZoneID(fqdn)
	if err != nil {
		return err
	}

	relative := toRelativeRecord(fqdn, acme.ToFqdn(zone))
	rsc := dns.NewRecordSetsClient(c.subscriptionId)
	rsc.Authorizer, err = c.newServicePrincipalTokenFromCredentials(azure.PublicCloud.ResourceManagerEndpoint)
	_, err = rsc.Delete(c.resourceGroup, zone, relative, dns.TXT, "")
	if err != nil {
		return err
	}

	return nil
}

// Checks that azure has a zone for this domain name.
func (c *DNSProvider) getHostedZoneID(fqdn string) (string, error) {
	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return "", err
	}

	// Now we want to to Azure and get the zone.
	dc := dns.NewZonesClient(c.subscriptionId)
	dc.Authorizer, err = c.newServicePrincipalTokenFromCredentials(azure.PublicCloud.ResourceManagerEndpoint)
	zone, err := dc.Get(c.resourceGroup, acme.UnFqdn(authZone))

	if err != nil {
		return "", err
	}

	// zone.Name shouldn't have a trailing dot(.)
	return to.String(zone.Name), nil
}

// NewServicePrincipalTokenFromCredentials creates a new ServicePrincipalToken using values of the
// passed credentials map.
func (c *DNSProvider) newServicePrincipalTokenFromCredentials(scope string) (*azure.ServicePrincipalToken, error) {
	oauthConfig, err := azure.PublicCloud.OAuthConfigForTenant(c.tenantId)
	if err != nil {
		panic(err)
	}
	return azure.NewServicePrincipalToken(*oauthConfig, c.clientId, c.clientSecret, scope)
}
