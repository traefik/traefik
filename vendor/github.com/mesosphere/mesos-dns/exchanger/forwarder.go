package exchanger

import (
	"fmt"
	"net"

	"github.com/miekg/dns"
)

// A Forwarder is a DNS message forwarder that transparently proxies messages
// to DNS servers.
type Forwarder func(*dns.Msg, string) (*dns.Msg, error)

// Forward is an utility method that calls f itself.
func (f Forwarder) Forward(m *dns.Msg, proto string) (*dns.Msg, error) {
	return f(m, proto)
}

// NewForwarder returns a new Forwarder for the given addrs with the given
// Exchangers map which maps network protocols to Exchangers.
//
// Every message will be exchanged with each address until no error is returned.
// If no addresses or no matching protocol exchanger exist, a *ForwardError will
// be returned.
func NewForwarder(addrs []string, exs map[string]Exchanger) Forwarder {
	return func(m *dns.Msg, proto string) (r *dns.Msg, err error) {
		ex, ok := exs[proto]
		if !ok || len(addrs) == 0 {
			return nil, &ForwardError{Addrs: addrs, Proto: proto}
		}
		for _, a := range addrs {
			if r, _, err = ex.Exchange(m, net.JoinHostPort(a, "53")); err == nil {
				break
			}
		}
		return
	}
}

// A ForwardError is returned by Forwarders when they can't forward.
type ForwardError struct {
	Addrs []string
	Proto string
}

// Error implements the error interface.
func (e ForwardError) Error() string {
	return fmt.Sprintf("can't forward to %v over %q", e.Addrs, e.Proto)
}
