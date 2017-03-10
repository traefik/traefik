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
}

// NewDNSProvider returns a DNSProvider instance configured for rfc2136
// dynamic update. Credentials must be passed in the environment variables:
// RFC2136_NAMESERVER, RFC2136_TSIG_ALGORITHM, RFC2136_TSIG_KEY and
// RFC2136_TSIG_SECRET. To disable TSIG authentication, leave the TSIG
// variables unset. RFC2136_NAMESERVER must be a network address in the form
// "host" or "host:port".
func NewDNSProvider() (*DNSProvider, error) {
	nameserver := os.Getenv("RFC2136_NAMESERVER")
	tsigAlgorithm := os.Getenv("RFC2136_TSIG_ALGORITHM")
	tsigKey := os.Getenv("RFC2136_TSIG_KEY")
	tsigSecret := os.Getenv("RFC2136_TSIG_SECRET")
	return NewDNSProviderCredentials(nameserver, tsigAlgorithm, tsigKey, tsigSecret)
}

// NewDNSProviderCredentials uses the supplied credentials to return a
// DNSProvider instance configured for rfc2136 dynamic update. To disable TSIG
// authentication, leave the TSIG parameters as empty strings.
// nameserver must be a network address in the form "host" or "host:port".
func NewDNSProviderCredentials(nameserver, tsigAlgorithm, tsigKey, tsigSecret string) (*DNSProvider, error) {
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

	return d, nil
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
