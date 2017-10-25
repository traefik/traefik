// Package googlecloud implements a DNS provider for solving the DNS-01
// challenge using Google Cloud DNS.
package googlecloud

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/xenolf/lego/acme"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"google.golang.org/api/dns/v1"
)

// DNSProvider is an implementation of the DNSProvider interface.
type DNSProvider struct {
	project string
	client  *dns.Service
}

// NewDNSProvider returns a DNSProvider instance configured for Google Cloud
// DNS. Project name must be passed in the environment variable: GCE_PROJECT.
// A Service Account file can be passed in the environment variable:
// GCE_SERVICE_ACCOUNT_FILE
func NewDNSProvider() (*DNSProvider, error) {
	project := os.Getenv("GCE_PROJECT")
	if saFile, ok := os.LookupEnv("GCE_SERVICE_ACCOUNT_FILE"); ok {
		return NewDNSProviderServiceAccount(project, saFile)
	}
	return NewDNSProviderCredentials(project)
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for Google Cloud DNS.
func NewDNSProviderCredentials(project string) (*DNSProvider, error) {
	if project == "" {
		return nil, fmt.Errorf("Google Cloud project name missing")
	}

	client, err := google.DefaultClient(context.Background(), dns.NdevClouddnsReadwriteScope)
	if err != nil {
		return nil, fmt.Errorf("Unable to get Google Cloud client: %v", err)
	}
	svc, err := dns.New(client)
	if err != nil {
		return nil, fmt.Errorf("Unable to create Google Cloud DNS service: %v", err)
	}
	return &DNSProvider{
		project: project,
		client:  svc,
	}, nil
}

// NewDNSProviderServiceAccount uses the supplied service account JSON file to
// return a DNSProvider instance configured for Google Cloud DNS.
func NewDNSProviderServiceAccount(project string, saFile string) (*DNSProvider, error) {
	if project == "" {
		return nil, fmt.Errorf("Google Cloud project name missing")
	}
	if saFile == "" {
		return nil, fmt.Errorf("Google Cloud Service Account file missing")
	}

	dat, err := ioutil.ReadFile(saFile)
	if err != nil {
		return nil, fmt.Errorf("Unable to read Service Account file: %v", err)
	}
	conf, err := google.JWTConfigFromJSON(dat, dns.NdevClouddnsReadwriteScope)
	if err != nil {
		return nil, fmt.Errorf("Unable to acquire config: %v", err)
	}
	client := conf.Client(oauth2.NoContext)

	svc, err := dns.New(client)
	if err != nil {
		return nil, fmt.Errorf("Unable to create Google Cloud DNS service: %v", err)
	}
	return &DNSProvider{
		project: project,
		client:  svc,
	}, nil
}

// Present creates a TXT record to fulfil the dns-01 challenge.
func (c *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)

	zone, err := c.getHostedZone(domain)
	if err != nil {
		return err
	}

	rec := &dns.ResourceRecordSet{
		Name:    fqdn,
		Rrdatas: []string{value},
		Ttl:     int64(ttl),
		Type:    "TXT",
	}
	change := &dns.Change{
		Additions: []*dns.ResourceRecordSet{rec},
	}

	// Look for existing records.
	list, err := c.client.ResourceRecordSets.List(c.project, zone).Name(fqdn).Type("TXT").Do()
	if err != nil {
		return err
	}
	if len(list.Rrsets) > 0 {
		// Attempt to delete the existing records when adding our new one.
		change.Deletions = list.Rrsets
	}

	chg, err := c.client.Changes.Create(c.project, zone, change).Do()
	if err != nil {
		return err
	}

	// wait for change to be acknowledged
	for chg.Status == "pending" {
		time.Sleep(time.Second)

		chg, err = c.client.Changes.Get(c.project, zone, chg.Id).Do()
		if err != nil {
			return err
		}
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (c *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	zone, err := c.getHostedZone(domain)
	if err != nil {
		return err
	}

	records, err := c.findTxtRecords(zone, fqdn)
	if err != nil {
		return err
	}

	for _, rec := range records {
		change := &dns.Change{
			Deletions: []*dns.ResourceRecordSet{rec},
		}
		_, err = c.client.Changes.Create(c.project, zone, change).Do()
		if err != nil {
			return err
		}
	}
	return nil
}

// Timeout customizes the timeout values used by the ACME package for checking
// DNS record validity.
func (c *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return 180 * time.Second, 5 * time.Second
}

// getHostedZone returns the managed-zone
func (c *DNSProvider) getHostedZone(domain string) (string, error) {
	authZone, err := acme.FindZoneByFqdn(acme.ToFqdn(domain), acme.RecursiveNameservers)
	if err != nil {
		return "", err
	}

	zones, err := c.client.ManagedZones.
		List(c.project).
		DnsName(authZone).
		Do()
	if err != nil {
		return "", fmt.Errorf("GoogleCloud API call failed: %v", err)
	}

	if len(zones.ManagedZones) == 0 {
		return "", fmt.Errorf("No matching GoogleCloud domain found for domain %s", authZone)
	}

	return zones.ManagedZones[0].Name, nil
}

func (c *DNSProvider) findTxtRecords(zone, fqdn string) ([]*dns.ResourceRecordSet, error) {

	recs, err := c.client.ResourceRecordSets.List(c.project, zone).Do()
	if err != nil {
		return nil, err
	}

	found := []*dns.ResourceRecordSet{}
	for _, r := range recs.Rrsets {
		if r.Type == "TXT" && r.Name == fqdn {
			found = append(found, r)
		}
	}

	return found, nil
}
