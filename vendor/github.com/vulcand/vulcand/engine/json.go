package engine

import (
	"encoding/json"
	"fmt"

	"github.com/vulcand/vulcand/plugin"
	"github.com/vulcand/vulcand/router"
)

type rawServers struct {
	Servers []json.RawMessage
}

type rawBackends struct {
	Backends []json.RawMessage
}

type rawMiddlewares struct {
	Middlewares []json.RawMessage
}

type rawFrontends struct {
	Frontends []json.RawMessage
}

type rawHosts struct {
	Hosts []json.RawMessage
}

type rawListeners struct {
	Listeners []json.RawMessage
}

type rawFrontend struct {
	Id        string
	Route     string
	Type      string
	BackendId string
	Settings  json.RawMessage
	Stats     *RoundTripStats
}

type rawBackend struct {
	Id       string
	Type     string
	Settings json.RawMessage
	Stats    *RoundTripStats
}

type RawMiddleware struct {
	Id         string
	Type       string
	Priority   int
	Middleware json.RawMessage
}

func HostsFromJSON(in []byte) ([]Host, error) {
	var hs rawHosts
	err := json.Unmarshal(in, &hs)
	if err != nil {
		return nil, err
	}
	out := []Host{}
	if len(hs.Hosts) != 0 {
		for _, raw := range hs.Hosts {
			h, err := HostFromJSON(raw)
			if err != nil {
				return nil, err
			}
			out = append(out, *h)
		}
	}
	return out, nil
}

func FrontendsFromJSON(router router.Router, in []byte) ([]Frontend, error) {
	var rf *rawFrontends
	err := json.Unmarshal(in, &rf)
	if err != nil {
		return nil, err
	}
	out := make([]Frontend, len(rf.Frontends))
	for i, raw := range rf.Frontends {
		f, err := FrontendFromJSON(router, raw)
		if err != nil {
			return nil, err
		}
		out[i] = *f
	}
	return out, nil
}

func HostFromJSON(in []byte, name ...string) (*Host, error) {
	var h *Host
	err := json.Unmarshal(in, &h)
	if err != nil {
		return nil, err
	}
	if len(name) != 0 {
		h.Name = name[0]
	}
	return NewHost(h.Name, h.Settings)
}

func ListenerFromJSON(in []byte, id ...string) (*Listener, error) {
	var rl *Listener
	err := json.Unmarshal(in, &rl)
	if err != nil {
		return nil, err
	}
	if len(id) != 0 {
		rl.Id = id[0]
	}
	if rl.Protocol == HTTPS && rl.Settings != nil {
		if _, err = NewTLSConfig(&rl.Settings.TLS); err != nil {
			return nil, err
		}
	}
	return NewListener(rl.Id, rl.Protocol, rl.Address.Network, rl.Address.Address, rl.Scope, rl.Settings)
}

func ListenersFromJSON(in []byte) ([]Listener, error) {
	var rls *rawListeners
	if err := json.Unmarshal(in, &rls); err != nil {
		return nil, err
	}
	out := make([]Listener, len(rls.Listeners))
	if len(out) == 0 {
		return out, nil
	}
	for i, rl := range rls.Listeners {
		l, err := ListenerFromJSON(rl)
		if err != nil {
			return nil, err
		}
		out[i] = *l
	}
	return out, nil
}

func KeyPairFromJSON(in []byte) (*KeyPair, error) {
	var c *KeyPair
	err := json.Unmarshal(in, &c)
	if err != nil {
		return nil, err
	}
	return NewKeyPair(c.Cert, c.Key)
}

func FrontendFromJSON(router router.Router, in []byte, id ...string) (*Frontend, error) {
	var rf *rawFrontend
	if err := json.Unmarshal(in, &rf); err != nil {
		return nil, err
	}
	if rf.Type != HTTP {
		return nil, fmt.Errorf("Unsupported frontend type: %v", rf.Type)
	}
	var s HTTPFrontendSettings
	if rf.Settings != nil {
		if err := json.Unmarshal(rf.Settings, &s); err != nil {
			return nil, err
		}
	}
	if len(id) != 0 {
		rf.Id = id[0]
	}
	f, err := NewHTTPFrontend(router, rf.Id, rf.BackendId, rf.Route, s)
	if err != nil {
		return nil, err
	}
	f.Stats = rf.Stats
	return f, nil
}

func MiddlewareFromJSON(in []byte, getter plugin.SpecGetter, id ...string) (*Middleware, error) {
	var ms *RawMiddleware
	err := json.Unmarshal(in, &ms)
	if err != nil {
		return nil, err
	}
	spec := getter(ms.Type)
	if spec == nil {
		return nil, fmt.Errorf("middleware of type %s is not supported", ms.Type)
	}
	m, err := spec.FromJSON(ms.Middleware)
	if err != nil {
		return nil, err
	}
	if len(id) != 0 {
		ms.Id = id[0]
	}
	return &Middleware{
		Id:         ms.Id,
		Type:       ms.Type,
		Middleware: m,
		Priority:   ms.Priority,
	}, nil
}

func BackendsFromJSON(in []byte) ([]Backend, error) {
	var rbs *rawBackends
	if err := json.Unmarshal(in, &rbs); err != nil {
		return nil, err
	}
	out := make([]Backend, len(rbs.Backends))
	if len(out) == 0 {
		return out, nil
	}
	for i, rb := range rbs.Backends {
		b, err := BackendFromJSON(rb)
		if err != nil {
			return nil, err
		}
		out[i] = *b
	}
	return out, nil
}

func BackendFromJSON(in []byte, id ...string) (*Backend, error) {
	var rb *rawBackend

	if err := json.Unmarshal(in, &rb); err != nil {
		return nil, err
	}
	if rb.Type != HTTP {
		return nil, fmt.Errorf("Unsupported backend type %v", rb.Type)
	}

	var s HTTPBackendSettings
	if rb.Settings != nil {
		if err := json.Unmarshal(rb.Settings, &s); err != nil {
			return nil, err
		}
	}
	if s.TLS != nil {
		if _, err := NewTLSConfig(s.TLS); err != nil {
			return nil, err
		}
	}
	if len(id) != 0 {
		rb.Id = id[0]
	}
	b, err := NewHTTPBackend(rb.Id, s)
	if err != nil {
		return nil, err
	}
	b.Stats = rb.Stats
	return b, nil
}

func ServersFromJSON(in []byte) ([]Server, error) {
	var rs *rawServers
	if err := json.Unmarshal(in, &rs); err != nil {
		return nil, err
	}
	out := make([]Server, len(rs.Servers))
	if len(out) == 0 {
		return out, nil
	}
	for i, rs := range rs.Servers {
		s, err := ServerFromJSON(rs)
		if err != nil {
			return nil, err
		}
		out[i] = *s
	}
	return out, nil
}

func MiddlewaresFromJSON(in []byte, getter plugin.SpecGetter) ([]Middleware, error) {
	var rm *rawMiddlewares
	if err := json.Unmarshal(in, &rm); err != nil {
		return nil, err
	}
	out := make([]Middleware, len(rm.Middlewares))
	if len(out) == 0 {
		return out, nil
	}
	for i, r := range rm.Middlewares {
		m, err := MiddlewareFromJSON(r, getter)
		if err != nil {
			return nil, err
		}
		out[i] = *m
	}
	return out, nil
}

func ServerFromJSON(in []byte, id ...string) (*Server, error) {
	var e *Server
	err := json.Unmarshal(in, &e)
	if err != nil {
		return nil, err
	}
	if len(id) != 0 {
		e.Id = id[0]
	}
	return NewServer(e.Id, e.URL)
}
