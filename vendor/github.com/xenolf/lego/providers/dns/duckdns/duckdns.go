// Package duckdns Adds lego support for http://duckdns.org.
// See http://www.duckdns.org/spec.jsp for more info on updating TXT records.
package duckdns

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

// DNSProvider adds and removes the record for the DNS challenge
type DNSProvider struct {
	// The api token
	token string
}

// NewDNSProvider returns a new DNS provider using
// environment variable DUCKDNS_TOKEN for adding and removing the DNS record.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("DUCKDNS_TOKEN")
	if err != nil {
		return nil, fmt.Errorf("DuckDNS: %v", err)
	}

	return NewDNSProviderCredentials(values["DUCKDNS_TOKEN"])
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for http://duckdns.org .
func NewDNSProviderCredentials(token string) (*DNSProvider, error) {
	if token == "" {
		return nil, errors.New("DuckDNS: credentials missing")
	}

	return &DNSProvider{token: token}, nil
}

// Present creates a TXT record to fulfil the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	_, txtRecord, _ := acme.DNS01Record(domain, keyAuth)
	return updateTxtRecord(domain, d.token, txtRecord, false)
}

// CleanUp clears DuckDNS TXT record
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	return updateTxtRecord(domain, d.token, "", true)
}

// updateTxtRecord Update the domains TXT record
// To update the TXT record we just need to make one simple get request.
// In DuckDNS you only have one TXT record shared with the domain and all sub domains.
func updateTxtRecord(domain, token, txt string, clear bool) error {
	u := fmt.Sprintf("https://www.duckdns.org/update?domains=%s&token=%s&clear=%t&txt=%s", domain, token, clear, txt)

	response, err := acme.HTTPClient.Get(u)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	body := string(bodyBytes)
	if body != "OK" {
		return fmt.Errorf("request to change TXT record for DuckDNS returned the following result (%s) this does not match expectation (OK) used url [%s]", body, u)
	}
	return nil
}
