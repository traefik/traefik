package server

import (
	"crypto/tls"
	"net/http"

	"golang.org/x/net/http/httpguts"
	"golang.org/x/net/http2"
)

func newSmartRoundTripper(transport *http.Transport) (http.RoundTripper, error) {
	transportHTTP1 := transport.Clone()

	err := http2.ConfigureTransport(transport)
	if err != nil {
		return nil, err
	}

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
	// If we have a connection upgrade, we don't use HTTP2
	if httpguts.HeaderValuesContainsToken(req.Header["Connection"], "Upgrade") {
		return m.http.RoundTrip(req)
	}

	return m.http2.RoundTrip(req)
}

func (m *smartRoundTripper) GetTLSClientConfig() *tls.Config {
	return m.http2.TLSClientConfig
}
