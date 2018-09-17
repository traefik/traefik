// Package gandi implements a DNS provider for solving the DNS-01
// challenge using Gandi DNS.
package gandi

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

// Gandi API reference:       http://doc.rpc.gandi.net/index.html
// Gandi API domain examples: http://doc.rpc.gandi.net/domain/faq.html

const (
	// defaultBaseURL Gandi XML-RPC endpoint used by Present and CleanUp
	defaultBaseURL = "https://rpc.gandi.net/xmlrpc/"
	minTTL         = 300
)

// findZoneByFqdn determines the DNS zone of an fqdn.
// It is overridden during tests.
var findZoneByFqdn = acme.FindZoneByFqdn

// Config is used to configure the creation of the DNSProvider
type Config struct {
	BaseURL            string
	APIKey             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TTL:                env.GetOrDefaultInt("GANDI_TTL", minTTL),
		PropagationTimeout: env.GetOrDefaultSecond("GANDI_PROPAGATION_TIMEOUT", 40*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond("GANDI_POLLING_INTERVAL", 60*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("GANDI_HTTP_TIMEOUT", 60*time.Second),
		},
	}
}

// inProgressInfo contains information about an in-progress challenge
type inProgressInfo struct {
	zoneID    int    // zoneID of gandi zone to restore in CleanUp
	newZoneID int    // zoneID of temporary gandi zone containing TXT record
	authZone  string // the domain name registered at gandi with trailing "."
}

// DNSProvider is an implementation of the
// acme.ChallengeProviderTimeout interface that uses Gandi's XML-RPC
// API to manage TXT records for a domain.
type DNSProvider struct {
	inProgressFQDNs     map[string]inProgressInfo
	inProgressAuthZones map[string]struct{}
	inProgressMu        sync.Mutex
	config              *Config
}

// NewDNSProvider returns a DNSProvider instance configured for Gandi.
// Credentials must be passed in the environment variable: GANDI_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("GANDI_API_KEY")
	if err != nil {
		return nil, fmt.Errorf("gandi: %v", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values["GANDI_API_KEY"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderCredentials uses the supplied credentials
// to return a DNSProvider instance configured for Gandi.
// Deprecated
func NewDNSProviderCredentials(apiKey string) (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.APIKey = apiKey

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Gandi.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("gandi: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("gandi: no API Key given")
	}

	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}

	return &DNSProvider{
		config:              config,
		inProgressFQDNs:     make(map[string]inProgressInfo),
		inProgressAuthZones: make(map[string]struct{}),
	}, nil
}

// Present creates a TXT record using the specified parameters. It
// does this by creating and activating a new temporary Gandi DNS
// zone. This new zone contains the TXT record.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)

	if d.config.TTL < minTTL {
		d.config.TTL = minTTL // 300 is gandi minimum value for ttl
	}

	// find authZone and Gandi zone_id for fqdn
	authZone, err := findZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return fmt.Errorf("gandi: findZoneByFqdn failure: %v", err)
	}

	zoneID, err := d.getZoneID(authZone)
	if err != nil {
		return fmt.Errorf("gandi: %v", err)
	}

	// determine name of TXT record
	if !strings.HasSuffix(
		strings.ToLower(fqdn), strings.ToLower("."+authZone)) {
		return fmt.Errorf("gandi: unexpected authZone %s for fqdn %s", authZone, fqdn)
	}
	name := fqdn[:len(fqdn)-len("."+authZone)]

	// acquire lock and check there is not a challenge already in
	// progress for this value of authZone
	d.inProgressMu.Lock()
	defer d.inProgressMu.Unlock()

	if _, ok := d.inProgressAuthZones[authZone]; ok {
		return fmt.Errorf("gandi: challenge already in progress for authZone %s", authZone)
	}

	// perform API actions to create and activate new gandi zone
	// containing the required TXT record
	newZoneName := fmt.Sprintf("%s [ACME Challenge %s]", acme.UnFqdn(authZone), time.Now().Format(time.RFC822Z))

	newZoneID, err := d.cloneZone(zoneID, newZoneName)
	if err != nil {
		return err
	}

	newZoneVersion, err := d.newZoneVersion(newZoneID)
	if err != nil {
		return fmt.Errorf("gandi: %v", err)
	}

	err = d.addTXTRecord(newZoneID, newZoneVersion, name, value, d.config.TTL)
	if err != nil {
		return fmt.Errorf("gandi: %v", err)
	}

	err = d.setZoneVersion(newZoneID, newZoneVersion)
	if err != nil {
		return fmt.Errorf("gandi: %v", err)
	}

	err = d.setZone(authZone, newZoneID)
	if err != nil {
		return fmt.Errorf("gandi: %v", err)
	}

	// save data necessary for CleanUp
	d.inProgressFQDNs[fqdn] = inProgressInfo{
		zoneID:    zoneID,
		newZoneID: newZoneID,
		authZone:  authZone,
	}
	d.inProgressAuthZones[authZone] = struct{}{}

	return nil
}

// CleanUp removes the TXT record matching the specified
// parameters. It does this by restoring the old Gandi DNS zone and
// removing the temporary one created by Present.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	// acquire lock and retrieve zoneID, newZoneID and authZone
	d.inProgressMu.Lock()
	defer d.inProgressMu.Unlock()

	if _, ok := d.inProgressFQDNs[fqdn]; !ok {
		// if there is no cleanup information then just return
		return nil
	}

	zoneID := d.inProgressFQDNs[fqdn].zoneID
	newZoneID := d.inProgressFQDNs[fqdn].newZoneID
	authZone := d.inProgressFQDNs[fqdn].authZone
	delete(d.inProgressFQDNs, fqdn)
	delete(d.inProgressAuthZones, authZone)

	// perform API actions to restore old gandi zone for authZone
	err := d.setZone(authZone, zoneID)
	if err != nil {
		return fmt.Errorf("gandi: %v", err)
	}

	return d.deleteZone(newZoneID)
}

// Timeout returns the values (40*time.Minute, 60*time.Second) which
// are used by the acme package as timeout and check interval values
// when checking for DNS record propagation with Gandi.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// rpcCall makes an XML-RPC call to Gandi's RPC endpoint by
// marshalling the data given in the call argument to XML and sending
// that via HTTP Post to Gandi. The response is then unmarshalled into
// the resp argument.
func (d *DNSProvider) rpcCall(call *methodCall, resp response) error {
	// marshal
	b, err := xml.MarshalIndent(call, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal error: %v", err)
	}

	// post
	b = append([]byte(`<?xml version="1.0"?>`+"\n"), b...)
	respBody, err := d.httpPost(d.config.BaseURL, "text/xml", bytes.NewReader(b))
	if err != nil {
		return err
	}

	// unmarshal
	err = xml.Unmarshal(respBody, resp)
	if err != nil {
		return fmt.Errorf("unmarshal error: %v", err)
	}
	if resp.faultCode() != 0 {
		return rpcError{
			faultCode: resp.faultCode(), faultString: resp.faultString()}
	}
	return nil
}

// functions to perform API actions

func (d *DNSProvider) getZoneID(domain string) (int, error) {
	resp := &responseStruct{}
	err := d.rpcCall(&methodCall{
		MethodName: "domain.info",
		Params: []param{
			paramString{Value: d.config.APIKey},
			paramString{Value: domain},
		},
	}, resp)
	if err != nil {
		return 0, err
	}

	var zoneID int
	for _, member := range resp.StructMembers {
		if member.Name == "zone_id" {
			zoneID = member.ValueInt
		}
	}

	if zoneID == 0 {
		return 0, fmt.Errorf("could not determine zone_id for %s", domain)
	}
	return zoneID, nil
}

func (d *DNSProvider) cloneZone(zoneID int, name string) (int, error) {
	resp := &responseStruct{}
	err := d.rpcCall(&methodCall{
		MethodName: "domain.zone.clone",
		Params: []param{
			paramString{Value: d.config.APIKey},
			paramInt{Value: zoneID},
			paramInt{Value: 0},
			paramStruct{
				StructMembers: []structMember{
					structMemberString{
						Name:  "name",
						Value: name,
					}},
			},
		},
	}, resp)
	if err != nil {
		return 0, err
	}

	var newZoneID int
	for _, member := range resp.StructMembers {
		if member.Name == "id" {
			newZoneID = member.ValueInt
		}
	}

	if newZoneID == 0 {
		return 0, fmt.Errorf("could not determine cloned zone_id")
	}
	return newZoneID, nil
}

func (d *DNSProvider) newZoneVersion(zoneID int) (int, error) {
	resp := &responseInt{}
	err := d.rpcCall(&methodCall{
		MethodName: "domain.zone.version.new",
		Params: []param{
			paramString{Value: d.config.APIKey},
			paramInt{Value: zoneID},
		},
	}, resp)
	if err != nil {
		return 0, err
	}

	if resp.Value == 0 {
		return 0, fmt.Errorf("could not create new zone version")
	}
	return resp.Value, nil
}

func (d *DNSProvider) addTXTRecord(zoneID int, version int, name string, value string, ttl int) error {
	resp := &responseStruct{}
	err := d.rpcCall(&methodCall{
		MethodName: "domain.zone.record.add",
		Params: []param{
			paramString{Value: d.config.APIKey},
			paramInt{Value: zoneID},
			paramInt{Value: version},
			paramStruct{
				StructMembers: []structMember{
					structMemberString{
						Name:  "type",
						Value: "TXT",
					}, structMemberString{
						Name:  "name",
						Value: name,
					}, structMemberString{
						Name:  "value",
						Value: value,
					}, structMemberInt{
						Name:  "ttl",
						Value: ttl,
					}},
			},
		},
	}, resp)
	return err
}

func (d *DNSProvider) setZoneVersion(zoneID int, version int) error {
	resp := &responseBool{}
	err := d.rpcCall(&methodCall{
		MethodName: "domain.zone.version.set",
		Params: []param{
			paramString{Value: d.config.APIKey},
			paramInt{Value: zoneID},
			paramInt{Value: version},
		},
	}, resp)
	if err != nil {
		return err
	}

	if !resp.Value {
		return fmt.Errorf("could not set zone version")
	}
	return nil
}

func (d *DNSProvider) setZone(domain string, zoneID int) error {
	resp := &responseStruct{}
	err := d.rpcCall(&methodCall{
		MethodName: "domain.zone.set",
		Params: []param{
			paramString{Value: d.config.APIKey},
			paramString{Value: domain},
			paramInt{Value: zoneID},
		},
	}, resp)
	if err != nil {
		return err
	}

	var respZoneID int
	for _, member := range resp.StructMembers {
		if member.Name == "zone_id" {
			respZoneID = member.ValueInt
		}
	}

	if respZoneID != zoneID {
		return fmt.Errorf("could not set new zone_id for %s", domain)
	}
	return nil
}

func (d *DNSProvider) deleteZone(zoneID int) error {
	resp := &responseBool{}
	err := d.rpcCall(&methodCall{
		MethodName: "domain.zone.delete",
		Params: []param{
			paramString{Value: d.config.APIKey},
			paramInt{Value: zoneID},
		},
	}, resp)
	if err != nil {
		return err
	}

	if !resp.Value {
		return fmt.Errorf("could not delete zone_id")
	}
	return nil
}

func (d *DNSProvider) httpPost(url string, bodyType string, body io.Reader) ([]byte, error) {
	resp, err := d.config.HTTPClient.Post(url, bodyType, body)
	if err != nil {
		return nil, fmt.Errorf("HTTP Post Error: %v", err)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("HTTP Post Error: %v", err)
	}

	return b, nil
}
