// Package namecheap implements a DNS provider for solving the DNS-01 challenge using namecheap DNS.
package namecheap

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/log"
	"github.com/go-acme/lego/platform/config/env"
)

// Notes about namecheap's tool API:
// 1. Using the API requires registration. Once registered, use your account
//    name and API key to access the API.
// 2. There is no API to add or modify a single DNS record. Instead you must
//    read the entire list of records, make modifications, and then write the
//    entire updated list of records.  (Yuck.)
// 3. Namecheap's DNS updates can be slow to propagate. I've seen them take
//    as long as an hour.
// 4. Namecheap requires you to whitelist the IP address from which you call
//    its APIs. It also requires all API calls to include the whitelisted IP
//    address as a form or query string value. This code uses a namecheap
//    service to query the client's IP address.

const (
	defaultBaseURL = "https://api.namecheap.com/xml.response"
	getIPURL       = "https://dynamicdns.park-your-domain.com/getip"
)

// A challenge represents all the data needed to specify a dns-01 challenge
// to lets-encrypt.
type challenge struct {
	domain   string
	key      string
	keyFqdn  string
	keyValue string
	tld      string
	sld      string
	host     string
}

// Config is used to configure the creation of the DNSProvider
type Config struct {
	Debug              bool
	BaseURL            string
	APIUser            string
	APIKey             string
	ClientIP           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		BaseURL:            defaultBaseURL,
		Debug:              env.GetOrDefaultBool("NAMECHEAP_DEBUG", false),
		TTL:                env.GetOrDefaultInt("NAMECHEAP_TTL", dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond("NAMECHEAP_PROPAGATION_TIMEOUT", 60*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond("NAMECHEAP_POLLING_INTERVAL", 15*time.Second),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond("NAMECHEAP_HTTP_TIMEOUT", 60*time.Second),
		},
	}
}

// DNSProvider is an implementation of the ChallengeProviderTimeout interface
// that uses Namecheap's tool API to manage TXT records for a domain.
type DNSProvider struct {
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for namecheap.
// Credentials must be passed in the environment variables:
// NAMECHEAP_API_USER and NAMECHEAP_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("NAMECHEAP_API_USER", "NAMECHEAP_API_KEY")
	if err != nil {
		return nil, fmt.Errorf("namecheap: %v", err)
	}

	config := NewDefaultConfig()
	config.APIUser = values["NAMECHEAP_API_USER"]
	config.APIKey = values["NAMECHEAP_API_KEY"]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for namecheap.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("namecheap: the configuration of the DNS provider is nil")
	}

	if config.APIUser == "" || config.APIKey == "" {
		return nil, fmt.Errorf("namecheap: credentials missing")
	}

	if len(config.ClientIP) == 0 {
		clientIP, err := getClientIP(config.HTTPClient, config.Debug)
		if err != nil {
			return nil, fmt.Errorf("namecheap: %v", err)
		}
		config.ClientIP = clientIP
	}

	return &DNSProvider{config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Namecheap can sometimes take a long time to complete an update, so wait up to 60 minutes for the update to propagate.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present installs a TXT record for the DNS challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	tlds, err := d.getTLDs()
	if err != nil {
		return fmt.Errorf("namecheap: %v", err)
	}

	ch, err := newChallenge(domain, keyAuth, tlds)
	if err != nil {
		return fmt.Errorf("namecheap: %v", err)
	}

	records, err := d.getHosts(ch.sld, ch.tld)
	if err != nil {
		return fmt.Errorf("namecheap: %v", err)
	}

	record := Record{
		Name:    ch.key,
		Type:    "TXT",
		Address: ch.keyValue,
		MXPref:  "10",
		TTL:     strconv.Itoa(d.config.TTL),
	}

	records = append(records, record)

	if d.config.Debug {
		for _, h := range records {
			log.Printf("%-5.5s %-30.30s %-6s %-70.70s", h.Type, h.Name, h.TTL, h.Address)
		}
	}

	err = d.setHosts(ch.sld, ch.tld, records)
	if err != nil {
		return fmt.Errorf("namecheap: %v", err)
	}
	return nil
}

// CleanUp removes a TXT record used for a previous DNS challenge.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	tlds, err := d.getTLDs()
	if err != nil {
		return fmt.Errorf("namecheap: %v", err)
	}

	ch, err := newChallenge(domain, keyAuth, tlds)
	if err != nil {
		return fmt.Errorf("namecheap: %v", err)
	}

	records, err := d.getHosts(ch.sld, ch.tld)
	if err != nil {
		return fmt.Errorf("namecheap: %v", err)
	}

	// Find the challenge TXT record and remove it if found.
	var found bool
	var newRecords []Record
	for _, h := range records {
		if h.Name == ch.key && h.Type == "TXT" {
			found = true
		} else {
			newRecords = append(newRecords, h)
		}
	}

	if !found {
		return nil
	}

	err = d.setHosts(ch.sld, ch.tld, newRecords)
	if err != nil {
		return fmt.Errorf("namecheap: %v", err)
	}
	return nil
}

// getClientIP returns the client's public IP address.
// It uses namecheap's IP discovery service to perform the lookup.
func getClientIP(client *http.Client, debug bool) (addr string, err error) {
	resp, err := client.Get(getIPURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	clientIP, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if debug {
		log.Println("Client IP:", string(clientIP))
	}
	return string(clientIP), nil
}

// newChallenge builds a challenge record from a domain name, a challenge
// authentication key, and a map of available TLDs.
func newChallenge(domain, keyAuth string, tlds map[string]string) (*challenge, error) {
	domain = dns01.UnFqdn(domain)
	parts := strings.Split(domain, ".")

	// Find the longest matching TLD.
	longest := -1
	for i := len(parts); i > 0; i-- {
		t := strings.Join(parts[i-1:], ".")
		if _, found := tlds[t]; found {
			longest = i - 1
		}
	}
	if longest < 1 {
		return nil, fmt.Errorf("invalid domain name %q", domain)
	}

	tld := strings.Join(parts[longest:], ".")
	sld := parts[longest-1]

	var host string
	if longest >= 1 {
		host = strings.Join(parts[:longest-1], ".")
	}

	fqdn, value := dns01.GetRecord(domain, keyAuth)

	return &challenge{
		domain:   domain,
		key:      "_acme-challenge." + host,
		keyFqdn:  fqdn,
		keyValue: value,
		tld:      tld,
		sld:      sld,
		host:     host,
	}, nil
}
