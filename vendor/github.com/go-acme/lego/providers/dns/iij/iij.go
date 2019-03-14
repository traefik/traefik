// Package iij implements a DNS provider for solving the DNS-01 challenge using IIJ DNS.
package iij

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
	"github.com/iij/doapi"
	"github.com/iij/doapi/protocol"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	AccessKey          string
	SecretKey          string
	DoServiceCode      string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("IIJ_TTL", 300),
		PropagationTimeout: env.GetOrDefaultSecond("IIJ_PROPAGATION_TIMEOUT", 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond("IIJ_POLLING_INTERVAL", 4*time.Second),
	}
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
		return nil, fmt.Errorf("iij: %v", err)
	}

	config := NewDefaultConfig()
	config.AccessKey = values["IIJ_API_ACCESS_KEY"]
	config.SecretKey = values["IIJ_API_SECRET_KEY"]
	config.DoServiceCode = values["IIJ_DO_SERVICE_CODE"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig takes a given config
// and returns a custom configured DNSProvider instance
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config.SecretKey == "" || config.AccessKey == "" || config.DoServiceCode == "" {
		return nil, fmt.Errorf("iij: credentials missing")
	}

	return &DNSProvider{
		api:    doapi.NewAPI(config.AccessKey, config.SecretKey),
		config: config,
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	_, value := dns01.GetRecord(domain, keyAuth)

	err := d.addTxtRecord(domain, value)
	if err != nil {
		return fmt.Errorf("iij: %v", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	_, value := dns01.GetRecord(domain, keyAuth)

	err := d.deleteTxtRecord(domain, value)
	if err != nil {
		return fmt.Errorf("iij: %v", err)
	}
	return nil
}

func (d *DNSProvider) addTxtRecord(domain, value string) error {
	zones, err := d.listZones()
	if err != nil {
		return err
	}

	owner, zone, err := splitDomain(domain, zones)
	if err != nil {
		return err
	}

	request := protocol.RecordAdd{
		DoServiceCode: d.config.DoServiceCode,
		ZoneName:      zone,
		Owner:         owner,
		TTL:           strconv.Itoa(d.config.TTL),
		RecordType:    "TXT",
		RData:         value,
	}

	response := &protocol.RecordAddResponse{}

	if err := doapi.Call(*d.api, request, response); err != nil {
		return err
	}

	return d.commit()
}

func (d *DNSProvider) deleteTxtRecord(domain, value string) error {
	zones, err := d.listZones()
	if err != nil {
		return err
	}

	owner, zone, err := splitDomain(domain, zones)
	if err != nil {
		return err
	}

	id, err := d.findTxtRecord(owner, zone, value)
	if err != nil {
		return err
	}

	request := protocol.RecordDelete{
		DoServiceCode: d.config.DoServiceCode,
		ZoneName:      zone,
		RecordID:      id,
	}

	response := &protocol.RecordDeleteResponse{}

	if err := doapi.Call(*d.api, request, response); err != nil {
		return err
	}

	return d.commit()
}

func (d *DNSProvider) commit() error {
	request := protocol.Commit{
		DoServiceCode: d.config.DoServiceCode,
	}

	response := &protocol.CommitResponse{}

	return doapi.Call(*d.api, request, response)
}

func (d *DNSProvider) findTxtRecord(owner, zone, value string) (string, error) {
	request := protocol.RecordListGet{
		DoServiceCode: d.config.DoServiceCode,
		ZoneName:      zone,
	}

	response := &protocol.RecordListGetResponse{}

	if err := doapi.Call(*d.api, request, response); err != nil {
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

func (d *DNSProvider) listZones() ([]string, error) {
	request := protocol.ZoneListGet{
		DoServiceCode: d.config.DoServiceCode,
	}

	response := &protocol.ZoneListGetResponse{}

	if err := doapi.Call(*d.api, request, response); err != nil {
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
