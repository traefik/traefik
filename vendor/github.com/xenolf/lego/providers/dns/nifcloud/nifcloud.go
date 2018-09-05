// Package nifcloud implements a DNS provider for solving the DNS-01 challenge
// using NIFCLOUD DNS.
package nifcloud

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

// DNSProvider implements the acme.ChallengeProvider interface
type DNSProvider struct {
	client *Client
}

// NewDNSProvider returns a DNSProvider instance configured for the NIFCLOUD DNS service.
// Credentials must be passed in the environment variables: NIFCLOUD_ACCESS_KEY_ID and NIFCLOUD_SECRET_ACCESS_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("NIFCLOUD_ACCESS_KEY_ID", "NIFCLOUD_SECRET_ACCESS_KEY")
	if err != nil {
		return nil, fmt.Errorf("NIFCLOUD: %v", err)
	}

	endpoint := os.Getenv("NIFCLOUD_DNS_ENDPOINT")
	if endpoint == "" {
		endpoint = defaultEndpoint
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}

	return NewDNSProviderCredentials(httpClient, endpoint, values["NIFCLOUD_ACCESS_KEY_ID"], values["NIFCLOUD_SECRET_ACCESS_KEY"])
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for NIFCLOUD.
func NewDNSProviderCredentials(httpClient *http.Client, endpoint, accessKey, secretKey string) (*DNSProvider, error) {
	client := newClient(httpClient, accessKey, secretKey, endpoint)

	return &DNSProvider{
		client: client,
	}, nil
}

// Present creates a TXT record using the specified parameters
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)
	return d.changeRecord("CREATE", fqdn, value, domain, ttl)
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)
	return d.changeRecord("DELETE", fqdn, value, domain, ttl)
}

func (d *DNSProvider) changeRecord(action, fqdn, value, domain string, ttl int) error {
	name := acme.UnFqdn(fqdn)

	reqParams := ChangeResourceRecordSetsRequest{
		XMLNs: xmlNs,
		ChangeBatch: ChangeBatch{
			Comment: "Managed by Lego",
			Changes: Changes{
				Change: []Change{
					{
						Action: action,
						ResourceRecordSet: ResourceRecordSet{
							Name: name,
							Type: "TXT",
							TTL:  ttl,
							ResourceRecords: ResourceRecords{
								ResourceRecord: []ResourceRecord{
									{
										Value: value,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	resp, err := d.client.ChangeResourceRecordSets(domain, reqParams)
	if err != nil {
		return fmt.Errorf("failed to change NIFCLOUD record set: %v", err)
	}

	statusID := resp.ChangeInfo.ID

	return acme.WaitFor(120*time.Second, 4*time.Second, func() (bool, error) {
		resp, err := d.client.GetChange(statusID)
		if err != nil {
			return false, fmt.Errorf("failed to query NIFCLOUD DNS change status: %v", err)
		}
		return resp.ChangeInfo.Status == "INSYNC", nil
	})
}
