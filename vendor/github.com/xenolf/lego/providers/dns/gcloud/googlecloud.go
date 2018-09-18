// Package gcloud implements a DNS provider for solving the DNS-01
// challenge using Google Cloud DNS.
package gcloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/dns/v1"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	Project            string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("GCE_TTL", 120),
		PropagationTimeout: env.GetOrDefaultSecond("GCE_PROPAGATION_TIMEOUT", 180*time.Second),
		PollingInterval:    env.GetOrDefaultSecond("GCE_POLLING_INTERVAL", 5*time.Second),
	}
}

// DNSProvider is an implementation of the DNSProvider interface.
type DNSProvider struct {
	config *Config
	client *dns.Service
}

// NewDNSProvider returns a DNSProvider instance configured for Google Cloud DNS.
// Project name must be passed in the environment variable: GCE_PROJECT.
// A Service Account file can be passed in the environment variable: GCE_SERVICE_ACCOUNT_FILE
func NewDNSProvider() (*DNSProvider, error) {
	if saFile, ok := os.LookupEnv("GCE_SERVICE_ACCOUNT_FILE"); ok {
		return NewDNSProviderServiceAccount(saFile)
	}

	project := os.Getenv("GCE_PROJECT")
	return NewDNSProviderCredentials(project)
}

// NewDNSProviderCredentials uses the supplied credentials
// to return a DNSProvider instance configured for Google Cloud DNS.
func NewDNSProviderCredentials(project string) (*DNSProvider, error) {
	if project == "" {
		return nil, fmt.Errorf("googlecloud: project name missing")
	}

	client, err := google.DefaultClient(context.Background(), dns.NdevClouddnsReadwriteScope)
	if err != nil {
		return nil, fmt.Errorf("googlecloud: unable to get Google Cloud client: %v", err)
	}

	config := NewDefaultConfig()
	config.Project = project
	config.HTTPClient = client

	return NewDNSProviderConfig(config)
}

// NewDNSProviderServiceAccount uses the supplied service account JSON file
// to return a DNSProvider instance configured for Google Cloud DNS.
func NewDNSProviderServiceAccount(saFile string) (*DNSProvider, error) {
	if saFile == "" {
		return nil, fmt.Errorf("googlecloud: Service Account file missing")
	}

	dat, err := ioutil.ReadFile(saFile)
	if err != nil {
		return nil, fmt.Errorf("googlecloud: unable to read Service Account file: %v", err)
	}

	// read project id from service account file
	var datJSON struct {
		ProjectID string `json:"project_id"`
	}
	err = json.Unmarshal(dat, &datJSON)
	if err != nil || datJSON.ProjectID == "" {
		return nil, fmt.Errorf("googlecloud: project ID not found in Google Cloud Service Account file")
	}
	project := datJSON.ProjectID

	conf, err := google.JWTConfigFromJSON(dat, dns.NdevClouddnsReadwriteScope)
	if err != nil {
		return nil, fmt.Errorf("googlecloud: unable to acquire config: %v", err)
	}
	client := conf.Client(context.Background())

	config := NewDefaultConfig()
	config.Project = project
	config.HTTPClient = client

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Google Cloud DNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("googlecloud: the configuration of the DNS provider is nil")
	}

	svc, err := dns.New(config.HTTPClient)
	if err != nil {
		return nil, fmt.Errorf("googlecloud: unable to create Google Cloud DNS service: %v", err)
	}

	return &DNSProvider{config: config, client: svc}, nil
}

// Present creates a TXT record to fulfil the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)

	zone, err := d.getHostedZone(domain)
	if err != nil {
		return fmt.Errorf("googlecloud: %v", err)
	}

	rec := &dns.ResourceRecordSet{
		Name:    fqdn,
		Rrdatas: []string{value},
		Ttl:     int64(d.config.TTL),
		Type:    "TXT",
	}
	change := &dns.Change{
		Additions: []*dns.ResourceRecordSet{rec},
	}

	// Look for existing records.
	existing, err := d.findTxtRecords(zone, fqdn)
	if err != nil {
		return fmt.Errorf("googlecloud: %v", err)
	}
	if len(existing) > 0 {
		// Attempt to delete the existing records when adding our new one.
		change.Deletions = existing
	}

	chg, err := d.client.Changes.Create(d.config.Project, zone, change).Do()
	if err != nil {
		return fmt.Errorf("googlecloud: %v", err)
	}

	// wait for change to be acknowledged
	for chg.Status == "pending" {
		time.Sleep(time.Second)

		chg, err = d.client.Changes.Get(d.config.Project, zone, chg.Id).Do()
		if err != nil {
			return fmt.Errorf("googlecloud: %v", err)
		}
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	zone, err := d.getHostedZone(domain)
	if err != nil {
		return fmt.Errorf("googlecloud: %v", err)
	}

	records, err := d.findTxtRecords(zone, fqdn)
	if err != nil {
		return fmt.Errorf("googlecloud: %v", err)
	}

	if len(records) == 0 {
		return nil
	}

	_, err = d.client.Changes.Create(d.config.Project, zone, &dns.Change{Deletions: records}).Do()
	return fmt.Errorf("googlecloud: %v", err)
}

// Timeout customizes the timeout values used by the ACME package for checking
// DNS record validity.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// getHostedZone returns the managed-zone
func (d *DNSProvider) getHostedZone(domain string) (string, error) {
	authZone, err := acme.FindZoneByFqdn(acme.ToFqdn(domain), acme.RecursiveNameservers)
	if err != nil {
		return "", err
	}

	zones, err := d.client.ManagedZones.
		List(d.config.Project).
		DnsName(authZone).
		Do()
	if err != nil {
		return "", fmt.Errorf("API call failed: %v", err)
	}

	if len(zones.ManagedZones) == 0 {
		return "", fmt.Errorf("no matching domain found for domain %s", authZone)
	}

	return zones.ManagedZones[0].Name, nil
}

func (d *DNSProvider) findTxtRecords(zone, fqdn string) ([]*dns.ResourceRecordSet, error) {
	recs, err := d.client.ResourceRecordSets.List(d.config.Project, zone).Name(fqdn).Type("TXT").Do()
	if err != nil {
		return nil, err
	}

	return recs.Rrsets, nil
}
