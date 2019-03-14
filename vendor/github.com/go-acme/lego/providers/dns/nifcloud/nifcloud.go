// Package nifcloud implements a DNS provider for solving the DNS-01 challenge using NIFCLOUD DNS.
package nifcloud

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-acme/lego/providers/dns/nifcloud/internal"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
	"github.com/go-acme/lego/platform/wait"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	BaseURL            string
	AccessKey          string
	SecretKey          string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("NIFCLOUD_TTL", dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond("NIFCLOUD_PROPAGATION_TIMEOUT", dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("NIFCLOUD_POLLING_INTERVAL", dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("NIFCLOUD_HTTP_TIMEOUT", 30*time.Second),
		},
	}
}

// DNSProvider implements the acme.ChallengeProvider interface
type DNSProvider struct {
	client *internal.Client
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for the NIFCLOUD DNS service.
// Credentials must be passed in the environment variables:
// NIFCLOUD_ACCESS_KEY_ID and NIFCLOUD_SECRET_ACCESS_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("NIFCLOUD_ACCESS_KEY_ID", "NIFCLOUD_SECRET_ACCESS_KEY")
	if err != nil {
		return nil, fmt.Errorf("nifcloud: %v", err)
	}

	config := NewDefaultConfig()
	config.BaseURL = env.GetOrFile("NIFCLOUD_DNS_ENDPOINT")
	config.AccessKey = values["NIFCLOUD_ACCESS_KEY_ID"]
	config.SecretKey = values["NIFCLOUD_SECRET_ACCESS_KEY"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for NIFCLOUD.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("nifcloud: the configuration of the DNS provider is nil")
	}

	client, err := internal.NewClient(config.AccessKey, config.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("nifcloud: %v", err)
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	}

	if len(config.BaseURL) > 0 {
		client.BaseURL = config.BaseURL
	}

	return &DNSProvider{client: client, config: config}, nil
}

// Present creates a TXT record using the specified parameters
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	err := d.changeRecord("CREATE", fqdn, value, domain, d.config.TTL)
	if err != nil {
		return fmt.Errorf("nifcloud: %v", err)
	}
	return err
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	err := d.changeRecord("DELETE", fqdn, value, domain, d.config.TTL)
	if err != nil {
		return fmt.Errorf("nifcloud: %v", err)
	}
	return err
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider) changeRecord(action, fqdn, value, domain string, ttl int) error {
	name := dns01.UnFqdn(fqdn)

	reqParams := internal.ChangeResourceRecordSetsRequest{
		XMLNs: internal.XMLNs,
		ChangeBatch: internal.ChangeBatch{
			Comment: "Managed by Lego",
			Changes: internal.Changes{
				Change: []internal.Change{
					{
						Action: action,
						ResourceRecordSet: internal.ResourceRecordSet{
							Name: name,
							Type: "TXT",
							TTL:  ttl,
							ResourceRecords: internal.ResourceRecords{
								ResourceRecord: []internal.ResourceRecord{
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

	return wait.For("nifcloud", 120*time.Second, 4*time.Second, func() (bool, error) {
		resp, err := d.client.GetChange(statusID)
		if err != nil {
			return false, fmt.Errorf("failed to query NIFCLOUD DNS change status: %v", err)
		}
		return resp.ChangeInfo.Status == "INSYNC", nil
	})
}
