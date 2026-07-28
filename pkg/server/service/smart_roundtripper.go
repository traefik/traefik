package service

import (
	"net/http"

	"golang.org/x/net/http/httpguts"
)

func newSmartRoundTripper(transport *http.Transport) *smartRoundTripper {
	// HTTP/1 only transport for requests with a Connection: Upgrade header.
	transportHTTP1 := transport.Clone()
	transportHTTP1.Protocols = new(http.Protocols)
	transportHTTP1.Protocols.SetHTTP1(true)

	// Transport switching automatically to HTTP/2 with TLS ALPN.
	transportHTTP2 := transport.Clone()
	transportHTTP2.Protocols = new(http.Protocols)
	transportHTTP2.Protocols.SetHTTP1(true)
	transportHTTP2.Protocols.SetHTTP2(true)

	// Transport speaking HTTP/2 with prior knowledge on unencrypted connections.
	transportH2C := transport.Clone()
	transportH2C.Protocols = new(http.Protocols)
	transportH2C.Protocols.SetUnencryptedHTTP2(true)

	return &smartRoundTripper{
		http2: transportHTTP2,
		http:  transportHTTP1,
		h2c:   transportH2C,
	}
}

// smartRoundTripper implements RoundTrip while making sure that HTTP/2 is not used
// with protocols that start with a Connection Upgrade, such as SPDY or Websocket.
type smartRoundTripper struct {
	http2 *http.Transport
	http  *http.Transport
	h2c   *http.Transport
}

func (m *smartRoundTripper) Clone() http.RoundTripper {
	return &smartRoundTripper{
		http2: m.http2.Clone(),
		http:  m.http.Clone(),
		h2c:   m.h2c.Clone(),
	}
}

func (m *smartRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	h2c := req.URL.Scheme == "h2c"
	if h2c {
		req.URL.Scheme = "http"
	}

	// Connection upgrades cannot be carried over HTTP/2, they always use HTTP/1.
	if httpguts.HeaderValuesContainsToken(req.Header["Connection"], "Upgrade") {
		return m.http.RoundTrip(req)
	}

	if h2c {
		return m.h2c.RoundTrip(req)
	}

	return m.http2.RoundTrip(req)
}
