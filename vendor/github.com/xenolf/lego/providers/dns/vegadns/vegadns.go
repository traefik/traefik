// Package vegadns implements a DNS provider for solving the DNS-01
// challenge using VegaDNS.
package vegadns

import (
	"fmt"
	"os"
	"strings"
	"time"

	vegaClient "github.com/OpenDNS/vegadns2client"
	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

// DNSProvider describes a provider for VegaDNS
type DNSProvider struct {
	client vegaClient.VegaDNSClient
}

// NewDNSProvider returns a DNSProvider instance configured for VegaDNS.
// Credentials must be passed in the environment variables:
// VEGADNS_URL, SECRET_VEGADNS_KEY, SECRET_VEGADNS_SECRET.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("VEGADNS_URL")
	if err != nil {
		return nil, fmt.Errorf("VegaDNS: %v", err)
	}

	key := os.Getenv("SECRET_VEGADNS_KEY")
	secret := os.Getenv("SECRET_VEGADNS_SECRET")

	return NewDNSProviderCredentials(values["VEGADNS_URL"], key, secret)
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for VegaDNS.
func NewDNSProviderCredentials(vegaDNSURL string, key string, secret string) (*DNSProvider, error) {
	vega := vegaClient.NewVegaDNSClient(vegaDNSURL)
	vega.APIKey = key
	vega.APISecret = secret

	return &DNSProvider{
		client: vega,
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS
// propagation. Adjusting here to cope with spikes in propagation times.
func (r *DNSProvider) Timeout() (timeout, interval time.Duration) {
	timeout = 12 * time.Minute
	interval = 1 * time.Minute
	return
}

// Present creates a TXT record to fulfil the dns-01 challenge
func (r *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)

	_, domainID, err := r.client.GetAuthZone(fqdn)
	if err != nil {
		return fmt.Errorf("can't find Authoritative Zone for %s in Present: %v", fqdn, err)
	}

	return r.client.CreateTXT(domainID, fqdn, value, 10)
}

// CleanUp removes the TXT record matching the specified parameters
func (r *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	_, domainID, err := r.client.GetAuthZone(fqdn)
	if err != nil {
		return fmt.Errorf("can't find Authoritative Zone for %s in CleanUp: %v", fqdn, err)
	}

	txt := strings.TrimSuffix(fqdn, ".")

	recordID, err := r.client.GetRecordID(domainID, txt, "TXT")
	if err != nil {
		return fmt.Errorf("couldn't get Record ID in CleanUp: %s", err)
	}

	return r.client.DeleteRecord(recordID)
}
