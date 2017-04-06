// Package dyn implements a DNS provider for solving the DNS-01 challenge
// using Dyn Managed DNS.
package dyn

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/xenolf/lego/acme"
)

var dynBaseURL = "https://api.dynect.net/REST"

type dynResponse struct {
	// One of 'success', 'failure', or 'incomplete'
	Status string `json:"status"`

	// The structure containing the actual results of the request
	Data json.RawMessage `json:"data"`

	// The ID of the job that was created in response to a request.
	JobID int `json:"job_id"`

	// A list of zero or more messages
	Messages json.RawMessage `json:"msgs"`
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface that uses
// Dyn's Managed DNS API to manage TXT records for a domain.
type DNSProvider struct {
	customerName string
	userName     string
	password     string
	token        string
}

// NewDNSProvider returns a DNSProvider instance configured for Dyn DNS.
// Credentials must be passed in the environment variables: DYN_CUSTOMER_NAME,
// DYN_USER_NAME and DYN_PASSWORD.
func NewDNSProvider() (*DNSProvider, error) {
	customerName := os.Getenv("DYN_CUSTOMER_NAME")
	userName := os.Getenv("DYN_USER_NAME")
	password := os.Getenv("DYN_PASSWORD")
	return NewDNSProviderCredentials(customerName, userName, password)
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for Dyn DNS.
func NewDNSProviderCredentials(customerName, userName, password string) (*DNSProvider, error) {
	if customerName == "" || userName == "" || password == "" {
		return nil, fmt.Errorf("DynDNS credentials missing")
	}

	return &DNSProvider{
		customerName: customerName,
		userName:     userName,
		password:     password,
	}, nil
}

func (d *DNSProvider) sendRequest(method, resource string, payload interface{}) (*dynResponse, error) {
	url := fmt.Sprintf("%s/%s", dynBaseURL, resource)

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if len(d.token) > 0 {
		req.Header.Set("Auth-Token", d.token)
	}

	client := &http.Client{Timeout: time.Duration(10 * time.Second)}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Dyn API request failed with HTTP status code %d", resp.StatusCode)
	} else if resp.StatusCode == 307 {
		// TODO add support for HTTP 307 response and long running jobs
		return nil, fmt.Errorf("Dyn API request returned HTTP 307. This is currently unsupported")
	}

	var dynRes dynResponse
	err = json.NewDecoder(resp.Body).Decode(&dynRes)
	if err != nil {
		return nil, err
	}

	if dynRes.Status == "failure" {
		// TODO add better error handling
		return nil, fmt.Errorf("Dyn API request failed: %s", dynRes.Messages)
	}

	return &dynRes, nil
}

// Starts a new Dyn API Session. Authenticates using customerName, userName,
// password and receives a token to be used in for subsequent requests.
func (d *DNSProvider) login() error {
	type creds struct {
		Customer string `json:"customer_name"`
		User     string `json:"user_name"`
		Pass     string `json:"password"`
	}

	type session struct {
		Token   string `json:"token"`
		Version string `json:"version"`
	}

	payload := &creds{Customer: d.customerName, User: d.userName, Pass: d.password}
	dynRes, err := d.sendRequest("POST", "Session", payload)
	if err != nil {
		return err
	}

	var s session
	err = json.Unmarshal(dynRes.Data, &s)
	if err != nil {
		return err
	}

	d.token = s.Token

	return nil
}

// Destroys Dyn Session
func (d *DNSProvider) logout() error {
	if len(d.token) == 0 {
		// nothing to do
		return nil
	}

	url := fmt.Sprintf("%s/Session", dynBaseURL)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Auth-Token", d.token)

	client := &http.Client{Timeout: time.Duration(10 * time.Second)}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Dyn API request failed to delete session with HTTP status code %d", resp.StatusCode)
	}

	d.token = ""

	return nil
}

// Present creates a TXT record using the specified parameters
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)

	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return err
	}

	err = d.login()
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"rdata": map[string]string{
			"txtdata": value,
		},
		"ttl": strconv.Itoa(ttl),
	}

	resource := fmt.Sprintf("TXTRecord/%s/%s/", authZone, fqdn)
	_, err = d.sendRequest("POST", resource, data)
	if err != nil {
		return err
	}

	err = d.publish(authZone, "Added TXT record for ACME dns-01 challenge using lego client")
	if err != nil {
		return err
	}

	err = d.logout()
	if err != nil {
		return err
	}

	return nil
}

func (d *DNSProvider) publish(zone, notes string) error {
	type publish struct {
		Publish bool   `json:"publish"`
		Notes   string `json:"notes"`
	}

	pub := &publish{Publish: true, Notes: notes}
	resource := fmt.Sprintf("Zone/%s/", zone)
	_, err := d.sendRequest("PUT", resource, pub)
	if err != nil {
		return err
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	authZone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return err
	}

	err = d.login()
	if err != nil {
		return err
	}

	resource := fmt.Sprintf("TXTRecord/%s/%s/", authZone, fqdn)
	url := fmt.Sprintf("%s/%s", dynBaseURL, resource)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Auth-Token", d.token)

	client := &http.Client{Timeout: time.Duration(10 * time.Second)}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Dyn API request failed to delete TXT record HTTP status code %d", resp.StatusCode)
	}

	err = d.publish(authZone, "Removed TXT record for ACME dns-01 challenge using lego client")
	if err != nil {
		return err
	}

	err = d.logout()
	if err != nil {
		return err
	}

	return nil
}
