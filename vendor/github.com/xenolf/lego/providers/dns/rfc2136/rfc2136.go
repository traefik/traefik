// Package rfc2136 implements a DNS provider for solving the DNS-01 challenge
// using the rfc2136 dynamic update.
package rfc2136

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/xenolf/lego/acme"
)

// DNSProvider is an implementation of the acme.ChallengeProvider interface that
// uses dynamic DNS updates (RFC 2136) to create TXT records on a nameserver.
type DNSProvider struct {
	nameserver    string
	tsigAlgorithm string
	tsigKey       string
	tsigSecret    string
	timeout       time.Duration
}

// NewDNSProvider returns a DNSProvider instance configured for rfc2136
// dynamic update. Configured with environment variables:
// RFC2136_NAMESERVER: Network address in the form "host" or "host:port".
// RFC2136_TSIG_ALGORITHM: Defaults to hmac-md5.sig-alg.reg.int. (HMAC-MD5).
// See https://github.com/miekg/dns/blob/master/tsig.go for supported values. 
// RFC2136_TSIG_KEY: Name of the secret key as defined in DNS server configuration.
// RFC2136_TSIG_SECRET: Secret key payload.
// RFC2136_TIMEOUT: DNS propagation timeout in time.ParseDuration format. (60s)
// To disable TSIG authentication, leave the RFC2136_TSIG* variables unset.
func NewDNSProvider() (*DNSProvider, error) {
	nameserver := os.Getenv("RFC2136_NAMESERVER")
	tsigAlgorithm := os.Getenv("RFC2136_TSIG_ALGORITHM")
	tsigKey := os.Getenv("RFC2136_TSIG_KEY")
	tsigSecret := os.Getenv("RFC2136_TSIG_SECRET")
	timeout := os.Getenv("RFC2136_TIMEOUT")
	return NewDNSProviderCredentials(nameserver, tsigAlgorithm, tsigKey, tsigSecret, timeout)
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for rfc2136 dynamic update. To disable TSIG
// authentication, leave the TSIG parameters as empty strings.
// nameserver must be a network address in the form "host" or "host:port".
func NewDNSProviderCredentials(nameserver, tsigAlgorithm, tsigKey, tsigSecret, timeout string) (*DNSProvider, error) {
	if nameserver == "" {
		return nil, fmt.Errorf("RFC2136 nameserver missing")
	}

	// Append the default DNS port if none is specified.
	if _, _, err := net.SplitHostPort(nameserver); err != nil {
		if strings.Contains(err.Error(), "missing port") {
			nameserver = net.JoinHostPort(nameserver, "53")
		} else {
			return nil, err
		}
	}
	d := &DNSProvider{
		nameserver: nameserver,
	}
	if tsigAlgorithm == "" {
		tsigAlgorithm = dns.HmacMD5
	}
	d.tsigAlgorithm = tsigAlgorithm
	if len(tsigKey) > 0 && len(tsigSecret) > 0 {
		d.tsigKey = tsigKey
		d.tsigSecret = tsigSecret
	}

	if timeout == "" {
		d.timeout = 60 * time.Second
	} else {
		t, err := time.ParseDuration(timeout)
		if err != nil {
			return nil, err
		} else if t < 0 {
			return nil, fmt.Errorf("Invalid/negative RFC2136_TIMEOUT: %v", timeout)
		} else {
			d.timeout = t
		}
	}

	return d, nil
}

// Returns the timeout configured with RFC2136_TIMEOUT, or 60s.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
    return d.timeout, 2 * time.Second
}

// Present creates a TXT record using the specified parameters
func (r *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)
	return r.changeRecord("INSERT", fqdn, value, ttl)
}

// CleanUp removes the TXT record matching the specified parameters
func (r *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)
	return r.changeRecord("REMOVE", fqdn, value, ttl)
}

func (r *DNSProvider) changeRecord(action, fqdn, value string, ttl int) error {
	// Find the zone for the given fqdn
	zone, err := acme.FindZoneByFqdn(fqdn, []string{r.nameserver})
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
		return fmt.Errorf("Unexpected action: %s", action)
	}

	// Setup client
	c := new(dns.Client)
	c.SingleInflight = true
	// TSIG authentication / msg signing
	if len(r.tsigKey) > 0 && len(r.tsigSecret) > 0 {
		m.SetTsig(dns.Fqdn(r.tsigKey), r.tsigAlgorithm, 300, time.Now().Unix())
		c.TsigSecret = map[string]string{dns.Fqdn(r.tsigKey): r.tsigSecret}
	}

	// Send the query
	reply, _, err := c.Exchange(m, r.nameserver)
	if err != nil {
		return fmt.Errorf("DNS update failed: %v", err)
	}
	if reply != nil && reply.Rcode != dns.RcodeSuccess {
		return fmt.Errorf("DNS update failed. Server replied: %s", dns.RcodeToString[reply.Rcode])
	}

	return nil
}
