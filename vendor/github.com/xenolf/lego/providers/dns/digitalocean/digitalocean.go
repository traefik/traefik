// Package digitalocean implements a DNS provider for solving the DNS-01
// challenge using digitalocean DNS.
package digitalocean

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/xenolf/lego/acme"
)

// DNSProvider is an implementation of the acme.ChallengeProvider interface
// that uses DigitalOcean's REST API to manage TXT records for a domain.
type DNSProvider struct {
	apiAuthToken string
	recordIDs    map[string]int
	recordIDsMu  sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for Digital
// Ocean. Credentials must be passed in the environment variable:
// DO_AUTH_TOKEN.
func NewDNSProvider() (*DNSProvider, error) {
	apiAuthToken := os.Getenv("DO_AUTH_TOKEN")
	return NewDNSProviderCredentials(apiAuthToken)
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for Digital Ocean.
func NewDNSProviderCredentials(apiAuthToken string) (*DNSProvider, error) {
	if apiAuthToken == "" {
		return nil, fmt.Errorf("DigitalOcean credentials missing")
	}
	return &DNSProvider{
		apiAuthToken: apiAuthToken,
		recordIDs:    make(map[string]int),
	}, nil
}

// Present creates a TXT record using the specified parameters
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	// txtRecordRequest represents the request body to DO's API to make a TXT record
	type txtRecordRequest struct {
		RecordType string `json:"type"`
		Name       string `json:"name"`
		Data       string `json:"data"`
	}

	// txtRecordResponse represents a response from DO's API after making a TXT record
	type txtRecordResponse struct {
		DomainRecord struct {
			ID   int    `json:"id"`
			Type string `json:"type"`
			Name string `json:"name"`
			Data string `json:"data"`
		} `json:"domain_record"`
	}

	fqdn, value, _ := acme.DNS01Record(domain, keyAuth)

	authZone, err := acme.FindZoneByFqdn(acme.ToFqdn(domain), acme.RecursiveNameservers)
	if err != nil {
		return fmt.Errorf("Could not determine zone for domain: '%s'. %s", domain, err)
	}

	authZone = acme.UnFqdn(authZone)

	reqURL := fmt.Sprintf("%s/v2/domains/%s/records", digitalOceanBaseURL, authZone)
	reqData := txtRecordRequest{RecordType: "TXT", Name: fqdn, Data: value}
	body, err := json.Marshal(reqData)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", reqURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.apiAuthToken))

	client := http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errInfo digitalOceanAPIError
		json.NewDecoder(resp.Body).Decode(&errInfo)
		return fmt.Errorf("HTTP %d: %s: %s", resp.StatusCode, errInfo.ID, errInfo.Message)
	}

	// Everything looks good; but we'll need the ID later to delete the record
	var respData txtRecordResponse
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return err
	}
	d.recordIDsMu.Lock()
	d.recordIDs[fqdn] = respData.DomainRecord.ID
	d.recordIDsMu.Unlock()

	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	// get the record's unique ID from when we created it
	d.recordIDsMu.Lock()
	recordID, ok := d.recordIDs[fqdn]
	d.recordIDsMu.Unlock()
	if !ok {
		return fmt.Errorf("unknown record ID for '%s'", fqdn)
	}

	authZone, err := acme.FindZoneByFqdn(acme.ToFqdn(domain), acme.RecursiveNameservers)
	if err != nil {
		return fmt.Errorf("Could not determine zone for domain: '%s'. %s", domain, err)
	}

	authZone = acme.UnFqdn(authZone)

	reqURL := fmt.Sprintf("%s/v2/domains/%s/records/%d", digitalOceanBaseURL, authZone, recordID)
	req, err := http.NewRequest("DELETE", reqURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.apiAuthToken))

	client := http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errInfo digitalOceanAPIError
		json.NewDecoder(resp.Body).Decode(&errInfo)
		return fmt.Errorf("HTTP %d: %s: %s", resp.StatusCode, errInfo.ID, errInfo.Message)
	}

	// Delete record ID from map
	d.recordIDsMu.Lock()
	delete(d.recordIDs, fqdn)
	d.recordIDsMu.Unlock()

	return nil
}

type digitalOceanAPIError struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

var digitalOceanBaseURL = "https://api.digitalocean.com"
