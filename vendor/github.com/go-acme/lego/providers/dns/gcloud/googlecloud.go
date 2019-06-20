// Package gcloud implements a DNS provider for solving the DNS-01 challenge using Google Cloud DNS.
package gcloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/log"
	"github.com/go-acme/lego/platform/config/env"
	"github.com/go-acme/lego/platform/wait"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/dns/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

const (
	changeStatusDone = "done"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	Debug              bool
	Project            string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		Debug:              env.GetOrDefaultBool("GCE_DEBUG", false),
		TTL:                env.GetOrDefaultInt("GCE_TTL", dns01.DefaultTTL),
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
// A Service Account can be passed in the environment variable: GCE_SERVICE_ACCOUNT
// or by specifying the keyfile location: GCE_SERVICE_ACCOUNT_FILE
func NewDNSProvider() (*DNSProvider, error) {
	// Use a service account file if specified via environment variable.
	if saKey := env.GetOrFile("GCE_SERVICE_ACCOUNT"); len(saKey) > 0 {
		return NewDNSProviderServiceAccountKey([]byte(saKey))
	}

	// Use default credentials.
	project := env.GetOrDefaultString("GCE_PROJECT", "")
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

// NewDNSProviderServiceAccountKey uses the supplied service account JSON
// to return a DNSProvider instance configured for Google Cloud DNS.
func NewDNSProviderServiceAccountKey(saKey []byte) (*DNSProvider, error) {
	if len(saKey) == 0 {
		return nil, fmt.Errorf("googlecloud: Service Account is missing")
	}

	// If GCE_PROJECT is non-empty it overrides the project in the service
	// account file.
	project := env.GetOrDefaultString("GCE_PROJECT", "")
	if project == "" {
		// read project id from service account file
		var datJSON struct {
			ProjectID string `json:"project_id"`
		}
		err := json.Unmarshal(saKey, &datJSON)
		if err != nil || datJSON.ProjectID == "" {
			return nil, fmt.Errorf("googlecloud: project ID not found in Google Cloud Service Account file")
		}
		project = datJSON.ProjectID
	}

	conf, err := google.JWTConfigFromJSON(saKey, dns.NdevClouddnsReadwriteScope)
	if err != nil {
		return nil, fmt.Errorf("googlecloud: unable to acquire config: %v", err)
	}
	client := conf.Client(context.Background())

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

	saKey, err := ioutil.ReadFile(saFile)
	if err != nil {
		return nil, fmt.Errorf("googlecloud: unable to read Service Account file: %v", err)
	}

	return NewDNSProviderServiceAccountKey(saKey)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Google Cloud DNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("googlecloud: the configuration of the DNS provider is nil")
	}
	if config.HTTPClient == nil {
		return nil, fmt.Errorf("googlecloud: unable to create Google Cloud DNS service: client is nil")
	}

	svc, err := dns.NewService(context.Background(), option.WithHTTPClient(config.HTTPClient))
	if err != nil {
		return nil, fmt.Errorf("googlecloud: unable to create Google Cloud DNS service: %v", err)
	}

	return &DNSProvider{config: config, client: svc}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	zone, err := d.getHostedZone(fqdn)
	if err != nil {
		return fmt.Errorf("googlecloud: %v", err)
	}

	// Look for existing records.
	existingRrSet, err := d.findTxtRecords(zone, fqdn)
	if err != nil {
		return fmt.Errorf("googlecloud: %v", err)
	}

	for _, rrSet := range existingRrSet {
		var rrd []string
		for _, rr := range rrSet.Rrdatas {
			data := mustUnquote(rr)
			rrd = append(rrd, data)

			if data == value {
				log.Printf("skip: the record already exists: %s", value)
				return nil
			}
		}
		rrSet.Rrdatas = rrd
	}

	// Attempt to delete the existing records before adding the new one.
	if len(existingRrSet) > 0 {
		if err = d.applyChanges(zone, &dns.Change{Deletions: existingRrSet}); err != nil {
			return fmt.Errorf("googlecloud: %v", err)
		}
	}

	rec := &dns.ResourceRecordSet{
		Name:    fqdn,
		Rrdatas: []string{value},
		Ttl:     int64(d.config.TTL),
		Type:    "TXT",
	}

	// Append existing TXT record data to the new TXT record data
	for _, rrSet := range existingRrSet {
		for _, rr := range rrSet.Rrdatas {
			if rr != value {
				rec.Rrdatas = append(rec.Rrdatas, rr)
			}
		}
	}

	change := &dns.Change{
		Additions: []*dns.ResourceRecordSet{rec},
	}

	if err = d.applyChanges(zone, change); err != nil {
		return fmt.Errorf("googlecloud: %v", err)
	}

	return nil
}

func (d *DNSProvider) applyChanges(zone string, change *dns.Change) error {
	if d.config.Debug {
		data, _ := json.Marshal(change)
		log.Printf("change (Create): %s", string(data))
	}

	chg, err := d.client.Changes.Create(d.config.Project, zone, change).Do()
	if err != nil {
		if v, ok := err.(*googleapi.Error); ok {
			if v.Code == http.StatusNotFound {
				return nil
			}
		}

		data, _ := json.Marshal(change)
		return fmt.Errorf("failed to perform changes [zone %s, change %s]: %v", zone, string(data), err)
	}

	if chg.Status == changeStatusDone {
		return nil
	}

	chgID := chg.Id

	// wait for change to be acknowledged
	return wait.For("apply change", 30*time.Second, 3*time.Second, func() (bool, error) {
		if d.config.Debug {
			data, _ := json.Marshal(change)
			log.Printf("change (Get): %s", string(data))
		}

		chg, err = d.client.Changes.Get(d.config.Project, zone, chgID).Do()
		if err != nil {
			data, _ := json.Marshal(change)
			return false, fmt.Errorf("failed to get changes [zone %s, change %s]: %v", zone, string(data), err)
		}

		if chg.Status == changeStatusDone {
			return true, nil
		}

		return false, fmt.Errorf("status: %s", chg.Status)
	})
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	zone, err := d.getHostedZone(fqdn)
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
	if err != nil {
		return fmt.Errorf("googlecloud: %v", err)
	}
	return nil
}

// Timeout customizes the timeout values used by the ACME package for checking
// DNS record validity.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// getHostedZone returns the managed-zone
func (d *DNSProvider) getHostedZone(domain string) (string, error) {
	authZone, err := dns01.FindZoneByFqdn(dns01.ToFqdn(domain))
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

	for _, z := range zones.ManagedZones {
		if z.Visibility == "public" || z.Visibility == "" {
			return z.Name, nil
		}
	}

	return "", fmt.Errorf("no public zone found for domain %s", authZone)
}

func (d *DNSProvider) findTxtRecords(zone, fqdn string) ([]*dns.ResourceRecordSet, error) {
	recs, err := d.client.ResourceRecordSets.List(d.config.Project, zone).Name(fqdn).Type("TXT").Do()
	if err != nil {
		return nil, err
	}

	return recs.Rrsets, nil
}

func mustUnquote(raw string) string {
	clean, err := strconv.Unquote(raw)
	if err != nil {
		return raw
	}
	return clean
}
