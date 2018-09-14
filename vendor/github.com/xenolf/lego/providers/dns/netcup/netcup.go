// Package netcup implements a DNS Provider for solving the DNS-01 challenge using the netcup DNS API.
package netcup

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	client *Client
}

// NewDNSProvider returns a DNSProvider instance configured for netcup.
// Credentials must be passed in the environment variables: NETCUP_CUSTOMER_NUMBER,
// NETCUP_API_KEY, NETCUP_API_PASSWORD
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("NETCUP_CUSTOMER_NUMBER", "NETCUP_API_KEY", "NETCUP_API_PASSWORD")
	if err != nil {
		return nil, fmt.Errorf("netcup: %v", err)
	}

	return NewDNSProviderCredentials(values["NETCUP_CUSTOMER_NUMBER"], values["NETCUP_API_KEY"], values["NETCUP_API_PASSWORD"])
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for netcup.
func NewDNSProviderCredentials(customer, key, password string) (*DNSProvider, error) {
	if customer == "" || key == "" || password == "" {
		return nil, fmt.Errorf("netcup: netcup credentials missing")
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	return &DNSProvider{
		client: NewClient(httpClient, customer, key, password),
	}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domainName, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domainName, keyAuth)

	zone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return fmt.Errorf("netcup: failed to find DNSZone, %v", err)
	}

	sessionID, err := d.client.Login()
	if err != nil {
		return err
	}

	hostname := strings.Replace(fqdn, "."+zone, "", 1)
	record := CreateTxtRecord(hostname, value)

	err = d.client.UpdateDNSRecord(sessionID, acme.UnFqdn(zone), record)
	if err != nil {
		if errLogout := d.client.Logout(sessionID); errLogout != nil {
			return fmt.Errorf("failed to add TXT-Record: %v; %v", err, errLogout)
		}
		return fmt.Errorf("failed to add TXT-Record: %v", err)
	}

	return d.client.Logout(sessionID)
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domainname, token, keyAuth string) error {
	fqdn, value, _ := acme.DNS01Record(domainname, keyAuth)

	zone, err := acme.FindZoneByFqdn(fqdn, acme.RecursiveNameservers)
	if err != nil {
		return fmt.Errorf("failed to find DNSZone, %v", err)
	}

	sessionID, err := d.client.Login()
	if err != nil {
		return err
	}

	hostname := strings.Replace(fqdn, "."+zone, "", 1)

	zone = acme.UnFqdn(zone)

	records, err := d.client.GetDNSRecords(zone, sessionID)
	if err != nil {
		return err
	}

	record := CreateTxtRecord(hostname, value)

	idx, err := GetDNSRecordIdx(records, record)
	if err != nil {
		return err
	}

	records[idx].DeleteRecord = true

	err = d.client.UpdateDNSRecord(sessionID, zone, records[idx])
	if err != nil {
		if errLogout := d.client.Logout(sessionID); errLogout != nil {
			return fmt.Errorf("%v; %v", err, errLogout)
		}
		return err
	}

	return d.client.Logout(sessionID)
}
