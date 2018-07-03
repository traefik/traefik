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
func NewDNSProviderCredentials(duckdnsToken string) (*DNSProvider, error) {
	if duckdnsToken == "" {
		return nil, errors.New("DuckDNS: credentials missing")
	}

	return &DNSProvider{token: duckdnsToken}, nil
}

// makeDuckdnsURL creates a url to clear the set or unset the TXT record.
// txt == "" will clear the TXT record.
func makeDuckdnsURL(domain, token, txt string) string {
	requestBase := fmt.Sprintf("https://www.duckdns.org/update?domains=%s&token=%s", domain, token)
	if txt == "" {
		return requestBase + "&clear=true"
	}
	return requestBase + "&txt=" + txt
}

func issueDuckdnsRequest(url string) error {
	response, err := acme.HTTPClient.Get(url)
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
		return fmt.Errorf("Request to change TXT record for duckdns returned the following result (%s) this does not match expectation (OK) used url [%s]", body, url)
	}
	return nil
}

// Present creates a TXT record to fulfil the dns-01 challenge.
// In duckdns you only have one TXT record shared with
// the domain and all sub domains.
//
// To update the TXT record we just need to make one simple get request.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	_, txtRecord, _ := acme.DNS01Record(domain, keyAuth)
	url := makeDuckdnsURL(domain, d.token, txtRecord)
	return issueDuckdnsRequest(url)
}

// CleanUp clears duckdns TXT record
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	url := makeDuckdnsURL(domain, d.token, "")
	return issueDuckdnsRequest(url)
}
