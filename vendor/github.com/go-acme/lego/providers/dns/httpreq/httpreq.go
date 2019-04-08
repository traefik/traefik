// Package httpreq implements a DNS provider for solving the DNS-01 challenge through a HTTP server.
package httpreq

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
)

type message struct {
	FQDN  string `json:"fqdn"`
	Value string `json:"value"`
}

type messageRaw struct {
	Domain  string `json:"domain"`
	Token   string `json:"token"`
	KeyAuth string `json:"keyAuth"`
}

// Config is used to configure the creation of the DNSProvider
type Config struct {
	Endpoint           *url.URL
	Mode               string
	Username           string
	Password           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond("HTTPREQ_PROPAGATION_TIMEOUT", dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond("HTTPREQ_POLLING_INTERVAL", dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("HTTPREQ_HTTP_TIMEOUT", 30*time.Second),
		},
	}
}

// DNSProvider describes a provider for acme-proxy
type DNSProvider struct {
	config *Config
}

// NewDNSProvider returns a DNSProvider instance.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("HTTPREQ_ENDPOINT")
	if err != nil {
		return nil, fmt.Errorf("httpreq: %v", err)
	}

	endpoint, err := url.Parse(values["HTTPREQ_ENDPOINT"])
	if err != nil {
		return nil, fmt.Errorf("httpreq: %v", err)
	}

	config := NewDefaultConfig()
	config.Mode = os.Getenv("HTTPREQ_MODE")
	config.Username = os.Getenv("HTTPREQ_USERNAME")
	config.Password = os.Getenv("HTTPREQ_PASSWORD")
	config.Endpoint = endpoint
	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider .
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("httpreq: the configuration of the DNS provider is nil")
	}

	if config.Endpoint == nil {
		return nil, errors.New("httpreq: the endpoint is missing")
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
	if d.config.Mode == "RAW" {
		msg := &messageRaw{
			Domain:  domain,
			Token:   token,
			KeyAuth: keyAuth,
		}

		err := d.doPost("/present", msg)
		if err != nil {
			return fmt.Errorf("httpreq: %v", err)
		}
		return nil
	}

	fqdn, value := dns01.GetRecord(domain, keyAuth)
	msg := &message{
		FQDN:  fqdn,
		Value: value,
	}

	err := d.doPost("/present", msg)
	if err != nil {
		return fmt.Errorf("httpreq: %v", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	if d.config.Mode == "RAW" {
		msg := &messageRaw{
			Domain:  domain,
			Token:   token,
			KeyAuth: keyAuth,
		}

		err := d.doPost("/cleanup", msg)
		if err != nil {
			return fmt.Errorf("httpreq: %v", err)
		}
		return nil
	}

	fqdn, value := dns01.GetRecord(domain, keyAuth)
	msg := &message{
		FQDN:  fqdn,
		Value: value,
	}

	err := d.doPost("/cleanup", msg)
	if err != nil {
		return fmt.Errorf("httpreq: %v", err)
	}
	return nil
}

func (d *DNSProvider) doPost(uri string, msg interface{}) error {
	reqBody := &bytes.Buffer{}
	err := json.NewEncoder(reqBody).Encode(msg)
	if err != nil {
		return err
	}

	newURI := path.Join(d.config.Endpoint.EscapedPath(), uri)
	endpoint, err := d.config.Endpoint.Parse(newURI)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, endpoint.String(), reqBody)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	if len(d.config.Username) > 0 && len(d.config.Password) > 0 {
		req.SetBasicAuth(d.config.Username, d.config.Password)
	}

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("%d: failed to read response body: %v", resp.StatusCode, err)
		}

		return fmt.Errorf("%d: request failed: %v", resp.StatusCode, string(body))
	}

	return nil
}
