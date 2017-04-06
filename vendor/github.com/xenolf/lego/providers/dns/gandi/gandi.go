// Package gandi implements a DNS provider for solving the DNS-01
// challenge using Gandi DNS.
package gandi

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/xenolf/lego/acme"
)

// Gandi API reference:       http://doc.rpc.gandi.net/index.html
// Gandi API domain examples: http://doc.rpc.gandi.net/domain/faq.html

var (
	// endpoint is the Gandi XML-RPC endpoint used by Present and
	// CleanUp. It is overridden during tests.
	endpoint = "https://rpc.gandi.net/xmlrpc/"
	// findZoneByFqdn determines the DNS zone of an fqdn. It is overridden
	// during tests.
	findZoneByFqdn = acme.FindZoneByFqdn
)

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
	apiKey              string
	inProgressFQDNs     map[string]inProgressInfo
	inProgressAuthZones map[string]struct{}
	inProgressMu        sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Gandi.
// Credentials must be passed in the environment variable: GANDI_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	apiKey := os.Getenv("GANDI_API_KEY")
	return NewDNSProviderCredentials(apiKey)
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for Gandi.
func NewDNSProviderCredentials(apiKey string) (*DNSProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("No Gandi API Key given")
	}
	return &DNSProvider{
		apiKey:              apiKey,
		inProgressFQDNs:     make(map[string]inProgressInfo),
		inProgressAuthZones: make(map[string]struct{}),
	}, nil
}

// Present creates a TXT record using the specified parameters. It
// does this by creating and activating a new temporary Gandi DNS
// zone. This new zone contains the TXT record.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)
	if ttl < 300 {
		ttl = 300 // 300 is gandi minimum value for ttl
	}
	// find authZone and Gandi zone_id for fqdn
	authZone, err := findZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return fmt.Errorf("Gandi DNS: findZoneByFqdn failure: %v", err)
	}
	zoneID, err := d.getZoneID(authZone)
	if err != nil {
		return err
	}
	// determine name of TXT record
	if !strings.HasSuffix(
		strings.ToLower(fqdn), strings.ToLower("."+authZone)) {
		return fmt.Errorf(
			"Gandi DNS: unexpected authZone %s for fqdn %s", authZone, fqdn)
	}
	name := fqdn[:len(fqdn)-len("."+authZone)]
	// acquire lock and check there is not a challenge already in
	// progress for this value of authZone
	d.inProgressMu.Lock()
	defer d.inProgressMu.Unlock()
	if _, ok := d.inProgressAuthZones[authZone]; ok {
		return fmt.Errorf(
			"Gandi DNS: challenge already in progress for authZone %s",
			authZone)
	}
	// perform API actions to create and activate new gandi zone
	// containing the required TXT record
	newZoneName := fmt.Sprintf(
		"%s [ACME Challenge %s]",
		acme.UnFqdn(authZone), time.Now().Format(time.RFC822Z))
	newZoneID, err := d.cloneZone(zoneID, newZoneName)
	if err != nil {
		return err
	}
	newZoneVersion, err := d.newZoneVersion(newZoneID)
	if err != nil {
		return err
	}
	err = d.addTXTRecord(newZoneID, newZoneVersion, name, value, ttl)
	if err != nil {
		return err
	}
	err = d.setZoneVersion(newZoneID, newZoneVersion)
	if err != nil {
		return err
	}
	err = d.setZone(authZone, newZoneID)
	if err != nil {
		return err
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
		return err
	}
	err = d.deleteZone(newZoneID)
	if err != nil {
		return err
	}
	return nil
}

// Timeout returns the values (40*time.Minute, 60*time.Second) which
// are used by the acme package as timeout and check interval values
// when checking for DNS record propagation with Gandi.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return 40 * time.Minute, 60 * time.Second
}

// types for XML-RPC method calls and parameters

type param interface {
	param()
}
type paramString struct {
	XMLName xml.Name `xml:"param"`
	Value   string   `xml:"value>string"`
}
type paramInt struct {
	XMLName xml.Name `xml:"param"`
	Value   int      `xml:"value>int"`
}

type structMember interface {
	structMember()
}
type structMemberString struct {
	Name  string `xml:"name"`
	Value string `xml:"value>string"`
}
type structMemberInt struct {
	Name  string `xml:"name"`
	Value int    `xml:"value>int"`
}
type paramStruct struct {
	XMLName       xml.Name       `xml:"param"`
	StructMembers []structMember `xml:"value>struct>member"`
}

func (p paramString) param()               {}
func (p paramInt) param()                  {}
func (m structMemberString) structMember() {}
func (m structMemberInt) structMember()    {}
func (p paramStruct) param()               {}

type methodCall struct {
	XMLName    xml.Name `xml:"methodCall"`
	MethodName string   `xml:"methodName"`
	Params     []param  `xml:"params"`
}

// types for XML-RPC responses

type response interface {
	faultCode() int
	faultString() string
}

type responseFault struct {
	FaultCode   int    `xml:"fault>value>struct>member>value>int"`
	FaultString string `xml:"fault>value>struct>member>value>string"`
}

func (r responseFault) faultCode() int      { return r.FaultCode }
func (r responseFault) faultString() string { return r.FaultString }

type responseStruct struct {
	responseFault
	StructMembers []struct {
		Name     string `xml:"name"`
		ValueInt int    `xml:"value>int"`
	} `xml:"params>param>value>struct>member"`
}

type responseInt struct {
	responseFault
	Value int `xml:"params>param>value>int"`
}

type responseBool struct {
	responseFault
	Value bool `xml:"params>param>value>boolean"`
}

// POSTing/Marshalling/Unmarshalling

type rpcError struct {
	faultCode   int
	faultString string
}

func (e rpcError) Error() string {
	return fmt.Sprintf(
		"Gandi DNS: RPC Error: (%d) %s", e.faultCode, e.faultString)
}

func httpPost(url string, bodyType string, body io.Reader) ([]byte, error) {
	client := http.Client{Timeout: 60 * time.Second}
	resp, err := client.Post(url, bodyType, body)
	if err != nil {
		return nil, fmt.Errorf("Gandi DNS: HTTP Post Error: %v", err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Gandi DNS: HTTP Post Error: %v", err)
	}
	return b, nil
}

// rpcCall makes an XML-RPC call to Gandi's RPC endpoint by
// marshalling the data given in the call argument to XML and sending
// that via HTTP Post to Gandi. The response is then unmarshalled into
// the resp argument.
func rpcCall(call *methodCall, resp response) error {
	// marshal
	b, err := xml.MarshalIndent(call, "", "  ")
	if err != nil {
		return fmt.Errorf("Gandi DNS: Marshal Error: %v", err)
	}
	// post
	b = append([]byte(`<?xml version="1.0"?>`+"\n"), b...)
	respBody, err := httpPost(endpoint, "text/xml", bytes.NewReader(b))
	if err != nil {
		return err
	}
	// unmarshal
	err = xml.Unmarshal(respBody, resp)
	if err != nil {
		return fmt.Errorf("Gandi DNS: Unmarshal Error: %v", err)
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
	err := rpcCall(&methodCall{
		MethodName: "domain.info",
		Params: []param{
			paramString{Value: d.apiKey},
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
		return 0, fmt.Errorf(
			"Gandi DNS: Could not determine zone_id for %s", domain)
	}
	return zoneID, nil
}

func (d *DNSProvider) cloneZone(zoneID int, name string) (int, error) {
	resp := &responseStruct{}
	err := rpcCall(&methodCall{
		MethodName: "domain.zone.clone",
		Params: []param{
			paramString{Value: d.apiKey},
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
		return 0, fmt.Errorf("Gandi DNS: Could not determine cloned zone_id")
	}
	return newZoneID, nil
}

func (d *DNSProvider) newZoneVersion(zoneID int) (int, error) {
	resp := &responseInt{}
	err := rpcCall(&methodCall{
		MethodName: "domain.zone.version.new",
		Params: []param{
			paramString{Value: d.apiKey},
			paramInt{Value: zoneID},
		},
	}, resp)
	if err != nil {
		return 0, err
	}
	if resp.Value == 0 {
		return 0, fmt.Errorf("Gandi DNS: Could not create new zone version")
	}
	return resp.Value, nil
}

func (d *DNSProvider) addTXTRecord(zoneID int, version int, name string, value string, ttl int) error {
	resp := &responseStruct{}
	err := rpcCall(&methodCall{
		MethodName: "domain.zone.record.add",
		Params: []param{
			paramString{Value: d.apiKey},
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
	if err != nil {
		return err
	}
	return nil
}

func (d *DNSProvider) setZoneVersion(zoneID int, version int) error {
	resp := &responseBool{}
	err := rpcCall(&methodCall{
		MethodName: "domain.zone.version.set",
		Params: []param{
			paramString{Value: d.apiKey},
			paramInt{Value: zoneID},
			paramInt{Value: version},
		},
	}, resp)
	if err != nil {
		return err
	}
	if !resp.Value {
		return fmt.Errorf("Gandi DNS: could not set zone version")
	}
	return nil
}

func (d *DNSProvider) setZone(domain string, zoneID int) error {
	resp := &responseStruct{}
	err := rpcCall(&methodCall{
		MethodName: "domain.zone.set",
		Params: []param{
			paramString{Value: d.apiKey},
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
		return fmt.Errorf(
			"Gandi DNS: Could not set new zone_id for %s", domain)
	}
	return nil
}

func (d *DNSProvider) deleteZone(zoneID int) error {
	resp := &responseBool{}
	err := rpcCall(&methodCall{
		MethodName: "domain.zone.delete",
		Params: []param{
			paramString{Value: d.apiKey},
			paramInt{Value: zoneID},
		},
	}, resp)
	if err != nil {
		return err
	}
	if !resp.Value {
		return fmt.Errorf("Gandi DNS: could not delete zone_id")
	}
	return nil
}
