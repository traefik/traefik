package proxy

import (
	"fmt"
	"net"
	"net/http"

	"github.com/vulcand/vulcand/engine"
)

type backend struct {
	mux     *mux
	backend engine.Backend

	frontends map[engine.FrontendKey]*frontend
	servers   []engine.Server
	transport *http.Transport
}

func newBackend(m *mux, b engine.Backend) (*backend, error) {
	s, err := m.transportSettings(b)
	if err != nil {
		return nil, err
	}
	return &backend{
		mux:       m,
		backend:   b,
		transport: newTransport(s),
		servers:   []engine.Server{},
		frontends: make(map[engine.FrontendKey]*frontend),
	}, nil
}

func (b *backend) String() string {
	return fmt.Sprintf("%v upstream(wrap=%v)", b.mux, &b.backend)
}

func (b *backend) linkFrontend(key engine.FrontendKey, f *frontend) {
	b.frontends[key] = f
}

func (b *backend) unlinkFrontend(key engine.FrontendKey) {
	delete(b.frontends, key)
}

func (b *backend) Close() error {
	b.transport.CloseIdleConnections()
	return nil
}

func (b *backend) update(be engine.Backend) error {
	if err := b.updateSettings(be); err != nil {
		return err
	}
	b.backend = be
	return nil
}

func (b *backend) updateSettings(be engine.Backend) error {
	olds := b.backend.HTTPSettings()
	news := be.HTTPSettings()

	// Nothing changed in transport options
	if news.Equals(olds) {
		return nil
	}
	s, err := b.mux.transportSettings(be)
	if err != nil {
		return err
	}
	t := newTransport(s)
	b.transport.CloseIdleConnections()
	b.transport = t
	for _, f := range b.frontends {
		f.updateTransport(t)
	}
	return nil
}

func (b *backend) indexOfServer(id string) int {
	for i := range b.servers {
		if b.servers[i].Id == id {
			return i
		}
	}
	return -1
}

func (b *backend) findServer(sk engine.ServerKey) (*engine.Server, bool) {
	i := b.indexOfServer(sk.Id)
	if i == -1 {
		return nil, false
	}
	return &b.servers[i], true
}

func (b *backend) upsertServer(s engine.Server) error {
	if i := b.indexOfServer(s.Id); i != -1 {
		b.servers[i] = s
	} else {
		b.servers = append(b.servers, s)
	}
	return b.updateFrontends()
}

func (b *backend) deleteServer(sk engine.ServerKey) error {
	i := b.indexOfServer(sk.Id)
	if i == -1 {
		return fmt.Errorf("%v not found %v", b, sk)
	}
	b.servers = append(b.servers[:i], b.servers[i+1:]...)
	return b.updateFrontends()
}

func (b *backend) updateFrontends() error {
	for _, f := range b.frontends {
		if err := f.updateBackend(b); err != nil {
			return err
		}
	}
	return nil
}

func newTransport(s *engine.TransportSettings) *http.Transport {
	return &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   s.Timeouts.Dial,
			KeepAlive: s.KeepAlive.Period,
		}).Dial,
		ResponseHeaderTimeout: s.Timeouts.Read,
		TLSHandshakeTimeout:   s.Timeouts.TLSHandshake,
		MaxIdleConnsPerHost:   s.KeepAlive.MaxIdleConnsPerHost,
		TLSClientConfig:       s.TLS,
	}
}
