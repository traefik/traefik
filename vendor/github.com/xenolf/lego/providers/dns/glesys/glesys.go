// Package glesys implements a DNS provider for solving the DNS-01
// challenge using GleSYS api.
package glesys

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/xenolf/lego/acmev2"
)

// GleSYS API reference: https://github.com/GleSYS/API/wiki/API-Documentation

// domainAPI is the GleSYS API endpoint used by Present and CleanUp.
const domainAPI = "https://api.glesys.com/domain"

var (
	// Logger is used to log API communication results;
	// if nil, the default log.Logger is used.
	Logger *log.Logger
)

// logf writes a log entry. It uses Logger if not
// nil, otherwise it uses the default log.Logger.
func logf(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}

// DNSProvider is an implementation of the
// acmev2.ChallengeProviderTimeout interface that uses GleSYS
// API to manage TXT records for a domain.
type DNSProvider struct {
	apiUser       string
	apiKey        string
	activeRecords map[string]int
	inProgressMu  sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for GleSYS.
// Credentials must be passed in the environment variables: GLESYS_API_USER
// and GLESYS_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	apiUser := os.Getenv("GLESYS_API_USER")
	apiKey := os.Getenv("GLESYS_API_KEY")
	return NewDNSProviderCredentials(apiUser, apiKey)
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for GleSYS.
func NewDNSProviderCredentials(apiUser string, apiKey string) (*DNSProvider, error) {
	if apiUser == "" || apiKey == "" {
		return nil, fmt.Errorf("GleSYS DNS: Incomplete credentials provided")
	}
	return &DNSProvider{
		apiUser:       apiUser,
		apiKey:        apiKey,
		activeRecords: make(map[string]int),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acmev2.DNS01Record(domain, keyAuth)
	if ttl < 60 {
		ttl = 60 // 60 is GleSYS minimum value for ttl
	}
	// find authZone
	authZone, err := acmev2.FindZoneByFqdn(fqdn, acmev2.RecursiveNameservers)
	if err != nil {
		return fmt.Errorf("GleSYS DNS: findZoneByFqdn failure: %v", err)
	}
	// determine name of TXT record
	if !strings.HasSuffix(
		strings.ToLower(fqdn), strings.ToLower("."+authZone)) {
		return fmt.Errorf(
			"GleSYS DNS: unexpected authZone %s for fqdn %s", authZone, fqdn)
	}
	name := fqdn[:len(fqdn)-len("."+authZone)]
	// acquire lock and check there is not a challenge already in
	// progress for this value of authZone
	d.inProgressMu.Lock()
	defer d.inProgressMu.Unlock()
	// add TXT record into authZone
	recordId, err := d.addTXTRecord(domain, acmev2.UnFqdn(authZone), name, value, ttl)
	if err != nil {
		return err
	}
	// save data necessary for CleanUp
	d.activeRecords[fqdn] = recordId
	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acmev2.DNS01Record(domain, keyAuth)
	// acquire lock and retrieve authZone
	d.inProgressMu.Lock()
	defer d.inProgressMu.Unlock()
	if _, ok := d.activeRecords[fqdn]; !ok {
		// if there is no cleanup information then just return
		return nil
	}
	recordId := d.activeRecords[fqdn]
	delete(d.activeRecords, fqdn)
	// delete TXT record from authZone
	err := d.deleteTXTRecord(domain, recordId)
	if err != nil {
		return err
	}
	return nil
}

// Timeout returns the values (20*time.Minute, 20*time.Second) which
// are used by the acme package as timeout and check interval values
// when checking for DNS record propagation with GleSYS.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return 20 * time.Minute, 20 * time.Second
}

// types for JSON method calls, parameters, and responses

type addRecordRequest struct {
	Domainname string `json:"domainname"`
	Host       string `json:"host"`
	Type       string `json:"type"`
	Data       string `json:"data"`
	Ttl        int    `json:"ttl,omitempty"`
}

type deleteRecordRequest struct {
	Recordid int `json:"recordid"`
}

type responseStruct struct {
	Response struct {
		Status struct {
			Code int `json:"code"`
		} `json:"status"`
		Record deleteRecordRequest `json:"record"`
	} `json:"response"`
}

// POSTing/Marshalling/Unmarshalling

func (d *DNSProvider) sendRequest(method string, resource string, payload interface{}) (*responseStruct, error) {
	url := fmt.Sprintf("%s/%s", domainAPI, resource)

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(d.apiUser, d.apiKey)

	client := &http.Client{Timeout: time.Duration(10 * time.Second)}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("GleSYS DNS: request failed with HTTP status code %d", resp.StatusCode)
	}
	var response responseStruct
	err = json.NewDecoder(resp.Body).Decode(&response)

	return &response, err
}

// functions to perform API actions

func (d *DNSProvider) addTXTRecord(fqdn string, domain string, name string, value string, ttl int) (int, error) {
	response, err := d.sendRequest("POST", "addrecord", addRecordRequest{
		Domainname: domain,
		Host:       name,
		Type:       "TXT",
		Data:       value,
		Ttl:        ttl,
	})
	if response != nil && response.Response.Status.Code == 200 {
		logf("[INFO][%s] GleSYS DNS: Successfully created recordid %d", fqdn, response.Response.Record.Recordid)
		return response.Response.Record.Recordid, nil
	}
	return 0, err
}

func (d *DNSProvider) deleteTXTRecord(fqdn string, recordid int) error {
	response, err := d.sendRequest("POST", "deleterecord", deleteRecordRequest{
		Recordid: recordid,
	})
	if response != nil && response.Response.Status.Code == 200 {
		logf("[INFO][%s] GleSYS DNS: Successfully deleted recordid %d", fqdn, recordid)
	}
	return err
}
