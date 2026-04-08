package service

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"golang.org/x/net/http/httpguts"
	"golang.org/x/net/http2"
)

func newSmartRoundTripper(transport *http.Transport, forwardingTimeouts *dynamic.ForwardingTimeouts) (*smartRoundTripper, error) {
	transportHTTP1 := transport.Clone()

	transportHTTP2, err := http2.ConfigureTransports(transport)
	if err != nil {
		return nil, err
	}

	if forwardingTimeouts != nil {
		transportHTTP2.ReadIdleTimeout = time.Duration(forwardingTimeouts.ReadIdleTimeout)
		transportHTTP2.PingTimeout = time.Duration(forwardingTimeouts.PingTimeout)
	}

	transportH2C := &http2.Transport{
		DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			return net.Dial(network, addr)
		},
		AllowHTTP: true,
	}

	if forwardingTimeouts != nil {
		transportH2C.ReadIdleTimeout = time.Duration(forwardingTimeouts.ReadIdleTimeout)
		transportH2C.PingTimeout = time.Duration(forwardingTimeouts.PingTimeout)
	}

	return &smartRoundTripper{
		http2: transport,
		http:  transportHTTP1,
		h2c:   transportH2C,
	}, nil
}

// smartRoundTripper implements RoundTrip while making sure that HTTP/2 is not used
// with protocols that start with a Connection Upgrade, such as SPDY or Websocket.
type smartRoundTripper struct {
	http2 *http.Transport
	http  *http.Transport
	h2c   *http2.Transport
}

func (m *smartRoundTripper) Clone() http.RoundTripper {
	h := m.http.Clone()
	h2 := m.http2.Clone()
	// TODO: this clone looses the "h2c" protocol registration. This was already the case before the fix for
	// https://github.com/traefik/traefik/issues/7465, when "h2c" was registered with
	// m.http2.RegisterProtocol("h2c", transportH2C).
	// We should switch to the new http.Protocols.SetUnencryptedHTTP2 on http.Transport, which is part of the stdlib and
	// supports Clone(). But this switch should probably be done with a new minor release.
	return &smartRoundTripper{http: h, http2: h2}
}

func (m *smartRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	isH2c := (m.h2c != nil && req.URL.Scheme == "h2c")
	if isH2c {
		req.URL.Scheme = "http"
	}

	// If we have a connection upgrade, we don't use HTTP/2 or HTTP/2 cleartext
	if httpguts.HeaderValuesContainsToken(req.Header["Connection"], "Upgrade") {
		return m.http.RoundTrip(req)
	} else if isH2c {
		return m.h2c.RoundTrip(req)
	}

	return m.http2.RoundTrip(req)
}
