package service

import (
	"net/http"
	"time"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"golang.org/x/net/http/httpguts"
	"golang.org/x/net/http2"
)

func newSmartRoundTripper(transport *http.Transport, forwardingTimeouts *dynamic.ForwardingTimeouts) (http.RoundTripper, error) {
	transportHTTP1 := transport.Clone()

	t, err := http2.ConfigureTransports(transport)
	if err != nil {
		return nil, err
	}

	if forwardingTimeouts != nil {
		t.ReadIdleTimeout = time.Duration(forwardingTimeouts.ReadIdleTimeout)
		t.PingTimeout = time.Duration(forwardingTimeouts.PingTimeout)
	}

	return &smartRoundTripper{
		http2: transport,
		http:  transportHTTP1,
	}, nil
}

type smartRoundTripper struct {
	http2 *http.Transport
	http  *http.Transport
}

// smartRoundTripper implements RoundTrip while making sure that HTTP/2 is not used
// with protocols that start with a Connection Upgrade, such as SPDY or Websocket.
func (m *smartRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// If we have a connection upgrade, we don't use HTTP/2
	if httpguts.HeaderValuesContainsToken(req.Header["Connection"], "Upgrade") {
		return m.http.RoundTrip(req)
	}

	return m.http2.RoundTrip(req)
}
