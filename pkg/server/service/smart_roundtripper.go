package service

import (
	"net/http"

	"golang.org/x/net/http/httpguts"
	"golang.org/x/net/http2"
)

func newSmartRoundTripper(transport *http.Transport, disableHTTP2 bool) (http.RoundTripper, error) {
	transportHTTP1 := transport.Clone()

	err := http2.ConfigureTransport(transport)
	if err != nil {
		return nil, err
	}

	return &smartRoundTripper{
		http2:        transport,
		http:         transportHTTP1,
		disableHTTP2: disableHTTP2,
	}, nil
}

type smartRoundTripper struct {
	http2        *http.Transport
	http         *http.Transport
	disableHTTP2 bool
}

// smartRoundTripper implements RoundTrip while making sure that HTTP/2 is not used
// if it is disabled through config or with protocols that start with a Connection
// Upgrade, such as SPDY or Websocket.
func (m *smartRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// If we have disabled HTTP/2 through config or this is a connection upgrade, we don't use HTTP/2
	if m.disableHTTP2 || httpguts.HeaderValuesContainsToken(req.Header["Connection"], "Upgrade") {
		return m.http.RoundTrip(req)
	}

	return m.http2.RoundTrip(req)
}
