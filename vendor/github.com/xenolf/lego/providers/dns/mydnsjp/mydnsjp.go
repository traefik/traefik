// Package mydnsjp implements a DNS provider for solving the DNS-01
// challenge using MyDNS.jp.
package mydnsjp

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/platform/config/env"
)

const defaultBaseURL = "https://www.mydns.jp/directedit.html"

// Config is used to configure the creation of the DNSProvider
type Config struct {
	MasterID           string
	Password           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond("MYDNSJP_PROPAGATION_TIMEOUT", 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond("MYDNSJP_POLLING_INTERVAL", 2*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("MYDNSJP_HTTP_TIMEOUT", 30*time.Second),
		},
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface
type DNSProvider struct {
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for MyDNS.jp.
// Credentials must be passed in the environment variables: MYDNSJP_MASTER_ID and MYDNSJP_PASSWORD.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("MYDNSJP_MASTER_ID", "MYDNSJP_PASSWORD")
	if err != nil {
		return nil, fmt.Errorf("mydnsjp: %v", err)
	}

	config := NewDefaultConfig()
	config.MasterID = values["MYDNSJP_MASTER_ID"]
	config.Password = values["MYDNSJP_PASSWORD"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for MyDNS.jp.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("mydnsjp: the configuration of the DNS provider is nil")
	}

	if config.MasterID == "" || config.Password == "" {
		return nil, errors.New("mydnsjp: some credentials information are missing")
	}

	return &DNSProvider{config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	_, value, _ := acme.DNS01Record(domain, keyAuth)
	err := d.doRequest(domain, value, "REGIST")
	if err != nil {
		return fmt.Errorf("mydnsjp: %v", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	_, value, _ := acme.DNS01Record(domain, keyAuth)
	err := d.doRequest(domain, value, "DELETE")
	if err != nil {
		return fmt.Errorf("mydnsjp: %v", err)
	}
	return nil
}

func (d *DNSProvider) doRequest(domain, value string, cmd string) error {
	req, err := d.buildRequest(domain, value, cmd)
	if err != nil {
		return err
	}

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error querying API: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var content []byte
		content, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("request %s failed [status code %d]: %s", req.URL, resp.StatusCode, string(content))
	}

	return nil
}

func (d *DNSProvider) buildRequest(domain, value string, cmd string) (*http.Request, error) {
	params := url.Values{}
	params.Set("CERTBOT_DOMAIN", domain)
	params.Set("CERTBOT_VALIDATION", value)
	params.Set("EDIT_CMD", cmd)

	req, err := http.NewRequest(http.MethodPost, defaultBaseURL, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("invalid request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(d.config.MasterID, d.config.Password)

	return req, nil
}
