// Package OVH implements a DNS provider for solving the DNS-01
// challenge using OVH DNS.
package ovh

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/ovh/go-ovh/ovh"
	"github.com/xenolf/lego/acme"
)

// OVH API reference:       https://eu.api.ovh.com/
// Create a Token:					https://eu.api.ovh.com/createToken/

// DNSProvider is an implementation of the acme.ChallengeProvider interface
// that uses OVH's REST API to manage TXT records for a domain.
type DNSProvider struct {
	client      *ovh.Client
	recordIDs   map[string]int
	recordIDsMu sync.Mutex
}

// NewDNSProvider returns a DNSProvider instance configured for OVH
// Credentials must be passed in the environment variable:
// OVH_ENDPOINT : it must be ovh-eu or ovh-ca
// OVH_APPLICATION_KEY
// OVH_APPLICATION_SECRET
// OVH_CONSUMER_KEY
func NewDNSProvider() (*DNSProvider, error) {
	apiEndpoint := os.Getenv("OVH_ENDPOINT")
	applicationKey := os.Getenv("OVH_APPLICATION_KEY")
	applicationSecret := os.Getenv("OVH_APPLICATION_SECRET")
	consumerKey := os.Getenv("OVH_CONSUMER_KEY")
	return NewDNSProviderCredentials(apiEndpoint, applicationKey, applicationSecret, consumerKey)
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for OVH.
func NewDNSProviderCredentials(apiEndpoint, applicationKey, applicationSecret, consumerKey string) (*DNSProvider, error) {
	if apiEndpoint == "" || applicationKey == "" || applicationSecret == "" || consumerKey == "" {
		return nil, fmt.Errorf("OVH credentials missing")
	}

	ovhClient, _ := ovh.NewClient(
		apiEndpoint,
		applicationKey,
		applicationSecret,
		consumerKey,
	)

	return &DNSProvider{
		client:    ovhClient,
		recordIDs: make(map[string]int),
	}, nil
}

// Present creates a TXT record to fulfil the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {

	// txtRecordRequest represents the request body to DO's API to make a TXT record
	type txtRecordRequest struct {
		FieldType string `json:"fieldType"`
		SubDomain string `json:"subDomain"`
		Target    string `json:"target"`
		TTL       int    `json:"ttl"`
	}

	// txtRecordResponse represents a response from DO's API after making a TXT record
	type txtRecordResponse struct {
		ID        int    `json:"id"`
		FieldType string `json:"fieldType"`
		SubDomain string `json:"subDomain"`
		Target    string `json:"target"`
		TTL       int    `json:"ttl"`
		Zone      string `json:"zone"`
	}

	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)

	// Parse domain name
	authZone, err := acme.FindZoneByFqdn(acme.ToFqdn(domain), acme.RecursiveNameservers)
	if err != nil {
		return fmt.Errorf("Could not determine zone for domain: '%s'. %s", domain, err)
	}

	authZone = acme.UnFqdn(authZone)
	subDomain := d.extractRecordName(fqdn, authZone)

	reqURL := fmt.Sprintf("/domain/zone/%s/record", authZone)
	reqData := txtRecordRequest{FieldType: "TXT", SubDomain: subDomain, Target: value, TTL: ttl}
	var respData txtRecordResponse

	// Create TXT record
	err = d.client.Post(reqURL, reqData, &respData)
	if err != nil {
		fmt.Printf("Error when call OVH api to add record : %q \n", err)
		return err
	}

	// Apply the change
	reqURL = fmt.Sprintf("/domain/zone/%s/refresh", authZone)
	err = d.client.Post(reqURL, nil, nil)
	if err != nil {
		fmt.Printf("Error when call OVH api to refresh zone : %q \n", err)
		return err
	}

	d.recordIDsMu.Lock()
	d.recordIDs[fqdn] = respData.ID
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

	reqURL := fmt.Sprintf("/domain/zone/%s/record/%d", authZone, recordID)

	err = d.client.Delete(reqURL, nil)
	if err != nil {
		fmt.Printf("Error when call OVH api to delete challenge record : %q \n", err)
		return err
	}

	// Delete record ID from map
	d.recordIDsMu.Lock()
	delete(d.recordIDs, fqdn)
	d.recordIDsMu.Unlock()

	return nil
}

func (d *DNSProvider) extractRecordName(fqdn, domain string) string {
	name := acme.UnFqdn(fqdn)
	if idx := strings.Index(name, "."+domain); idx != -1 {
		return name[:idx]
	}
	return name
}
