// Package rfc2136 implements a DNS provider for solving the DNS-01 challenge using the rfc2136 dynamic update.
package rfc2136

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/go-acme/lego/challenge/dns01"
	"github.com/go-acme/lego/platform/config/env"
	"github.com/miekg/dns"
)

// Config is used to configure the creation of the DNSProvider
type Config struct {
	Nameserver         string
	TSIGAlgorithm      string
	TSIGKey            string
	TSIGSecret         string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
	SequenceInterval   time.Duration
	DNSTimeout         time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider
func NewDefaultConfig() *Config {
	return &Config{
		TSIGAlgorithm:      env.GetOrDefaultString("RFC2136_TSIG_ALGORITHM", dns.HmacMD5),
		TTL:                env.GetOrDefaultInt("RFC2136_TTL", dns01.DefaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond("RFC2136_PROPAGATION_TIMEOUT", env.GetOrDefaultSecond("RFC2136_TIMEOUT", 60*time.Second)),
		PollingInterval:    env.GetOrDefaultSecond("RFC2136_POLLING_INTERVAL", 2*time.Second),
		SequenceInterval:   env.GetOrDefaultSecond("RFC2136_SEQUENCE_INTERVAL", dns01.DefaultPropagationTimeout),
		DNSTimeout:         env.GetOrDefaultSecond("RFC2136_DNS_TIMEOUT", 10*time.Second),
	}
}

// DNSProvider is an implementation of the acme.ChallengeProvider interface that
// uses dynamic DNS updates (RFC 2136) to create TXT records on a nameserver.
type DNSProvider struct {
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for rfc2136
// dynamic update. Configured with environment variables:
// RFC2136_NAMESERVER: Network address in the form "host" or "host:port".
// RFC2136_TSIG_ALGORITHM: Defaults to hmac-md5.sig-alg.reg.int. (HMAC-MD5).
// See https://github.com/miekg/dns/blob/master/tsig.go for supported values.
// RFC2136_TSIG_KEY: Name of the secret key as defined in DNS server configuration.
// RFC2136_TSIG_SECRET: Secret key payload.
// RFC2136_PROPAGATION_TIMEOUT: DNS propagation timeout in time.ParseDuration format. (60s)
// To disable TSIG authentication, leave the RFC2136_TSIG* variables unset.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("RFC2136_NAMESERVER")
	if err != nil {
		return nil, fmt.Errorf("rfc2136: %v", err)
	}

	config := NewDefaultConfig()
	config.Nameserver = values["RFC2136_NAMESERVER"]
	config.TSIGKey = env.GetOrFile("RFC2136_TSIG_KEY")
	config.TSIGSecret = env.GetOrFile("RFC2136_TSIG_SECRET")

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for rfc2136.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("rfc2136: the configuration of the DNS provider is nil")
	}

	if config.Nameserver == "" {
		return nil, fmt.Errorf("rfc2136: nameserver missing")
	}

	if config.TSIGAlgorithm == "" {
		config.TSIGAlgorithm = dns.HmacMD5
	}

	// Append the default DNS port if none is specified.
	if _, _, err := net.SplitHostPort(config.Nameserver); err != nil {
		if strings.Contains(err.Error(), "missing port") {
			config.Nameserver = net.JoinHostPort(config.Nameserver, "53")
		} else {
			return nil, fmt.Errorf("rfc2136: %v", err)
		}
	}

	if len(config.TSIGKey) == 0 && len(config.TSIGSecret) > 0 ||
		len(config.TSIGKey) > 0 && len(config.TSIGSecret) == 0 {
		config.TSIGKey = ""
		config.TSIGSecret = ""
	}

	return &DNSProvider{config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Sequential All DNS challenges for this provider will be resolved sequentially.
// Returns the interval between each iteration.
func (d *DNSProvider) Sequential() time.Duration {
	return d.config.SequenceInterval
}

// Present creates a TXT record using the specified parameters
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	err := d.changeRecord("INSERT", fqdn, value, d.config.TTL)
	if err != nil {
		return fmt.Errorf("rfc2136: failed to insert: %v", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	err := d.changeRecord("REMOVE", fqdn, value, d.config.TTL)
	if err != nil {
		return fmt.Errorf("rfc2136: failed to remove: %v", err)
	}
	return nil
}

func (d *DNSProvider) changeRecord(action, fqdn, value string, ttl int) error {
	// Find the zone for the given fqdn
	zone, err := dns01.FindZoneByFqdnCustom(fqdn, []string{d.config.Nameserver})
	if err != nil {
		return err
	}

	// Create RR
	rr := new(dns.TXT)
	rr.Hdr = dns.RR_Header{Name: fqdn, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: uint32(ttl)}
	rr.Txt = []string{value}
	rrs := []dns.RR{rr}

	// Create dynamic update packet
	m := new(dns.Msg)
	m.SetUpdate(zone)
	switch action {
	case "INSERT":
		// Always remove old challenge left over from who knows what.
		m.RemoveRRset(rrs)
		m.Insert(rrs)
	case "REMOVE":
		m.Remove(rrs)
	default:
		return fmt.Errorf("unexpected action: %s", action)
	}

	// Setup client
	c := &dns.Client{Timeout: d.config.DNSTimeout}
	c.SingleInflight = true

	// TSIG authentication / msg signing
	if len(d.config.TSIGKey) > 0 && len(d.config.TSIGSecret) > 0 {
		m.SetTsig(dns.Fqdn(d.config.TSIGKey), d.config.TSIGAlgorithm, 300, time.Now().Unix())
		c.TsigSecret = map[string]string{dns.Fqdn(d.config.TSIGKey): d.config.TSIGSecret}
	}

	// Send the query
	reply, _, err := c.Exchange(m, d.config.Nameserver)
	if err != nil {
		return fmt.Errorf("DNS update failed: %v", err)
	}
	if reply != nil && reply.Rcode != dns.RcodeSuccess {
		return fmt.Errorf("DNS update failed: server replied: %s", dns.RcodeToString[reply.Rcode])
	}

	return nil
}
