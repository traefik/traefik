package dnstest

import (
	"net"

	"github.com/miekg/dns"
)

// ResponseRecorder implements the dns.ResponseWriter interface. It's used in
// tests only.
type ResponseRecorder struct {
	Local, Remote net.IPAddr
	Msg           *dns.Msg
}

// LocalAddr returns the internal Local net.IPAddr.
func (r ResponseRecorder) LocalAddr() net.Addr { return &r.Local }

// RemoteAddr returns the internal Remote net.IPAddr
func (r ResponseRecorder) RemoteAddr() net.Addr { return &r.Remote }

// WriteMsg sets the internal Msg to the given Msg and returns nil.
func (r *ResponseRecorder) WriteMsg(m *dns.Msg) error {
	r.Msg = m
	return nil
}

// Write is not implemented.
func (r *ResponseRecorder) Write([]byte) (int, error) { return 0, nil }

// Close is not implemented.
func (r *ResponseRecorder) Close() error { return nil }

// TsigStatus is not implemented.
func (r ResponseRecorder) TsigStatus() error { return nil }

// TsigTimersOnly is not implemented.
func (r ResponseRecorder) TsigTimersOnly(bool) {}

// Hijack is not implemented.
func (r *ResponseRecorder) Hijack() {}
