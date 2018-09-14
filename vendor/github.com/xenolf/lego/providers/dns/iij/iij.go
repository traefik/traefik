// Package iij implements a DNS provider for solving the DNS-01 challenge using IIJ DNS.
package iij

import (
	"fmt"
	"strings"
	"time"

	"github.com/iij/doapi"
	"github.com/iij/doapi/protocol"
	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	AccessKey     string
	SecretKey     string
	DoServiceCode string
}

// DNSProvider implements the acme.ChallengeProvider interface
type DNSProvider struct {
	api    *doapi.API
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for IIJ DO
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("IIJ_API_ACCESS_KEY", "IIJ_API_SECRET_KEY", "IIJ_DO_SERVICE_CODE")
	if err != nil {
		return nil, fmt.Errorf("IIJ: %v", err)
	}

	return NewDNSProviderConfig(&Config{
		AccessKey:     values["IIJ_API_ACCESS_KEY"],
		SecretKey:     values["IIJ_API_SECRET_KEY"],
		DoServiceCode: values["IIJ_DO_SERVICE_CODE"],
	})
}

// NewDNSProviderConfig takes a given config ans returns a custom configured
// DNSProvider instance
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	return &DNSProvider{
		api:    doapi.NewAPI(config.AccessKey, config.SecretKey),
		config: config,
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
func (p *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return time.Minute * 2, time.Second * 4
}

// Present creates a TXT record using the specified parameters
func (p *DNSProvider) Present(domain, token, keyAuth string) error {
	_, value, _ := acme.DNS01Record(domain, keyAuth)
	return p.addTxtRecord(domain, value)
}

// CleanUp removes the TXT record matching the specified parameters
func (p *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	_, value, _ := acme.DNS01Record(domain, keyAuth)
	return p.deleteTxtRecord(domain, value)
}

func (p *DNSProvider) addTxtRecord(domain, value string) error {
	zones, err := p.listZones()
	if err != nil {
		return err
	}

	owner, zone, err := splitDomain(domain, zones)
	if err != nil {
		return err
	}

	request := protocol.RecordAdd{
		DoServiceCode: p.config.DoServiceCode,
		ZoneName:      zone,
		Owner:         owner,
		TTL:           "300",
		RecordType:    "TXT",
		RData:         value,
	}

	response := &protocol.RecordAddResponse{}

	if err := doapi.Call(*p.api, request, response); err != nil {
		return err
	}

	return p.commit()
}

func (p *DNSProvider) deleteTxtRecord(domain, value string) error {
	zones, err := p.listZones()
	if err != nil {
		return err
	}

	owner, zone, err := splitDomain(domain, zones)
	if err != nil {
		return err
	}

	id, err := p.findTxtRecord(owner, zone, value)
	if err != nil {
		return err
	}

	request := protocol.RecordDelete{
		DoServiceCode: p.config.DoServiceCode,
		ZoneName:      zone,
		RecordID:      id,
	}

	response := &protocol.RecordDeleteResponse{}

	if err := doapi.Call(*p.api, request, response); err != nil {
		return err
	}

	return p.commit()
}

func (p *DNSProvider) commit() error {
	request := protocol.Commit{
		DoServiceCode: p.config.DoServiceCode,
	}

	response := &protocol.CommitResponse{}

	return doapi.Call(*p.api, request, response)
}

func (p *DNSProvider) findTxtRecord(owner, zone, value string) (string, error) {
	request := protocol.RecordListGet{
		DoServiceCode: p.config.DoServiceCode,
		ZoneName:      zone,
	}

	response := &protocol.RecordListGetResponse{}

	if err := doapi.Call(*p.api, request, response); err != nil {
		return "", err
	}

	var id string

	for _, record := range response.RecordList {
		if record.Owner == owner && record.RecordType == "TXT" && record.RData == "\""+value+"\"" {
			id = record.Id
		}
	}

	if id == "" {
		return "", fmt.Errorf("%s record in %s not found", owner, zone)
	}

	return id, nil
}

func (p *DNSProvider) listZones() ([]string, error) {
	request := protocol.ZoneListGet{
		DoServiceCode: p.config.DoServiceCode,
	}

	response := &protocol.ZoneListGetResponse{}

	if err := doapi.Call(*p.api, request, response); err != nil {
		return nil, err
	}

	return response.ZoneList, nil
}

func splitDomain(domain string, zones []string) (string, string, error) {
	parts := strings.Split(strings.Trim(domain, "."), ".")

	var owner string
	var zone string

	for i := 0; i < len(parts)-1; i++ {
		zone = strings.Join(parts[i:], ".")
		if zoneContains(zone, zones) {
			baseOwner := strings.Join(parts[0:i], ".")
			if len(baseOwner) > 0 {
				baseOwner = "." + baseOwner
			}
			owner = "_acme-challenge" + baseOwner
			break
		}
	}

	if len(owner) == 0 {
		return "", "", fmt.Errorf("%s not found", domain)
	}

	return owner, zone, nil
}

func zoneContains(zone string, zones []string) bool {
	for _, z := range zones {
		if zone == z {
			return true
		}
	}
	return false
}
