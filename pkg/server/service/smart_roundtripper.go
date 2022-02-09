package service

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"golang.org/x/net/http/httpguts"
	"golang.org/x/net/http2"
)

func newSmartRoundTripper(transport *http.Transport, forwardingTimeouts *dynamic.ForwardingTimeouts) (http.RoundTripper, error) {
	transportHTTP1 := transport.Clone()

	transportHTTP2, err := http2.ConfigureTransports(transport)
	if err != nil {
		return nil, err
	}

	if forwardingTimeouts != nil {
		transportHTTP2.ReadIdleTimeout = time.Duration(forwardingTimeouts.ReadIdleTimeout)
		transportHTTP2.PingTimeout = time.Duration(forwardingTimeouts.PingTimeout)
	}

	transportH2C := &h2cTransportWrapper{
		Transport: &http2.Transport{
			DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(network, addr)
			},
			AllowHTTP: true,
		},
	}

	if forwardingTimeouts != nil {
		transportH2C.ReadIdleTimeout = time.Duration(forwardingTimeouts.ReadIdleTimeout)
		transportH2C.PingTimeout = time.Duration(forwardingTimeouts.PingTimeout)
	}

	transport.RegisterProtocol("h2c", transportH2C)

	return &smartRoundTripper{
		http2: transport,
		http:  transportHTTP1,
	}, nil
}

// smartRoundTripper implements RoundTrip while making sure that HTTP/2 is not used
// with protocols that start with a Connection Upgrade, such as SPDY or Websocket.
type smartRoundTripper struct {
	http2 *http.Transport
	http  *http.Transport
}

func (m *smartRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// If we have a connection upgrade, we don't use HTTP/2
	if httpguts.HeaderValuesContainsToken(req.Header["Connection"], "Upgrade") {
		return m.http.RoundTrip(req)
	}

	return m.http2.RoundTrip(req)
}
