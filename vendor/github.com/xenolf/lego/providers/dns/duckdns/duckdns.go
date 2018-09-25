// Package duckdns Adds lego support for http://duckdns.org.
// See http://www.duckdns.org/spec.jsp for more info on updating TXT records.
package duckdns

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	Token              string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	client := acme.HTTPClient
	client.Timeout = env.GetOrDefaultSecond("DUCKDNS_HTTP_TIMEOUT", 30*time.Second)

	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond("DUCKDNS_PROPAGATION_TIMEOUT", acme.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("DUCKDNS_POLLING_INTERVAL", acme.DefaultPollingInterval),
		HTTPClient:         &client,
	}
}

// DNSProvider adds and removes the record for the DNS challenge
type DNSProvider struct {
	config *Config
}

// NewDNSProvider returns a new DNS provider using
// environment variable DUCKDNS_TOKEN for adding and removing the DNS record.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("DUCKDNS_TOKEN")
	if err != nil {
		return nil, fmt.Errorf("duckdns: %v", err)
	}

	config := NewDefaultConfig()
	config.Token = values["DUCKDNS_TOKEN"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderCredentials uses the supplied credentials
// to return a DNSProvider instance configured for http://duckdns.org
// Deprecated
func NewDNSProviderCredentials(token string) (*DNSProvider, error) {
	config := NewDefaultConfig()
	config.Token = token

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for DuckDNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("duckdns: the configuration of the DNS provider is nil")
	}

	if config.Token == "" {
		return nil, errors.New("duckdns: credentials missing")
	}

	return &DNSProvider{config: config}, nil
}

// Present creates a TXT record to fulfil the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	_, txtRecord, _ := acme.DNS01Record(domain, keyAuth)
	return updateTxtRecord(domain, d.config.Token, txtRecord, false)
}

// CleanUp clears DuckDNS TXT record
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	return updateTxtRecord(domain, d.config.Token, "", true)
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
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
