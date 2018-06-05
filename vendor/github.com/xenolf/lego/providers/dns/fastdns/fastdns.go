package fastdns

import (
	"fmt"
	"os"
	"reflect"

	configdns "github.com/akamai/AkamaiOPEN-edgegrid-golang/configdns-v1"
	"github.com/akamai/AkamaiOPEN-edgegrid-golang/edgegrid"
	"github.com/xenolf/lego/acme"
)

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	config edgegrid.Config
}

// NewDNSProvider uses the supplied environment variables to return a DNSProvider instance:
// AKAMAI_HOST, AKAMAI_CLIENT_TOKEN, AKAMAI_CLIENT_SECRET, AKAMAI_ACCESS_TOKEN
func NewDNSProvider() (*DNSProvider, error) {
	host := os.Getenv("AKAMAI_HOST")
	clientToken := os.Getenv("AKAMAI_CLIENT_TOKEN")
	clientSecret := os.Getenv("AKAMAI_CLIENT_SECRET")
	accessToken := os.Getenv("AKAMAI_ACCESS_TOKEN")

	return NewDNSProviderClient(host, clientToken, clientSecret, accessToken)
}

// NewDNSProviderClient uses the supplied parameters to return a DNSProvider instance
// configured for FastDNS.
func NewDNSProviderClient(host, clientToken, clientSecret, accessToken string) (*DNSProvider, error) {
	if clientToken == "" || clientSecret == "" || accessToken == "" || host == "" {
		return nil, fmt.Errorf("Akamai FastDNS credentials missing")
	}
	config := edgegrid.Config{
		Host:         host,
		ClientToken:  clientToken,
		ClientSecret: clientSecret,
		AccessToken:  accessToken,
		MaxBody:      131072,
	}

	return &DNSProvider{
		config: config,
	}, nil
}

// Present creates a TXT record to fullfil the dns-01 challenge.
func (c *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)
	zoneName, recordName, err := c.findZoneAndRecordName(fqdn, domain)
	if err != nil {
		return err
	}

	configdns.Init(c.config)

	zone, err := configdns.GetZone(zoneName)
	if err != nil {
		return err
	}

	record := configdns.NewTxtRecord()
	record.SetField("name", recordName)
	record.SetField("ttl", ttl)
	record.SetField("target", value)
	record.SetField("active", true)

	existingRecord := c.findExistingRecord(zone, recordName)

	if existingRecord != nil {
		if reflect.DeepEqual(existingRecord.ToMap(), record.ToMap()) {
			return nil
		}
		zone.RemoveRecord(existingRecord)
		return c.createRecord(zone, record)
	}

	return c.createRecord(zone, record)
}

// CleanUp removes the record matching the specified parameters.
func (c *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)
	zoneName, recordName, err := c.findZoneAndRecordName(fqdn, domain)
	if err != nil {
		return err
	}

	configdns.Init(c.config)

	zone, err := configdns.GetZone(zoneName)
	if err != nil {
		return err
	}

	existingRecord := c.findExistingRecord(zone, recordName)

	if existingRecord != nil {
		err := zone.RemoveRecord(existingRecord)
		if err != nil {
			return err
		}
		return zone.Save()
	}

	return nil
}

func (c *DNSProvider) findZoneAndRecordName(fqdn, domain string) (string, string, error) {
	zone, err := acme.FindZoneByFqdn(acme.ToFqdn(domain), acme.RecursiveNameservers)
	if err != nil {
		return "", "", err
	}
	zone = acme.UnFqdn(zone)
	name := acme.UnFqdn(fqdn)
	name = name[:len(name)-len("."+zone)]

	return zone, name, nil
}

func (c *DNSProvider) findExistingRecord(zone *configdns.Zone, recordName string) *configdns.TxtRecord {
	for _, r := range zone.Zone.Txt {
		if r.Name == recordName {
			return r
		}
	}

	return nil
}

func (c *DNSProvider) createRecord(zone *configdns.Zone, record *configdns.TxtRecord) error {
	err := zone.AddRecord(record)
	if err != nil {
		return err
	}

	return zone.Save()
}
