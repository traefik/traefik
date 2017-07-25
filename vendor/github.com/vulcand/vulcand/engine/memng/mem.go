// package memng provides in memory engine implementation, mostly used for test purposes
package memng

import (
	"fmt"
	"time"

	"github.com/vulcand/vulcand/engine"
	"github.com/vulcand/vulcand/plugin"

	log "github.com/Sirupsen/logrus"
)

// Mem is exported to provide easy access to its internals
type Mem struct {
	Hosts     map[engine.HostKey]engine.Host
	Frontends map[engine.FrontendKey]engine.Frontend
	Backends  map[engine.BackendKey]engine.Backend
	Listeners map[engine.ListenerKey]engine.Listener

	Middlewares map[engine.FrontendKey][]engine.Middleware
	Servers     map[engine.BackendKey][]engine.Server

	Registry    *plugin.Registry
	ChangesC    chan interface{}
	ErrorsC     chan error
	LogSeverity log.Level
}

func New(r *plugin.Registry) engine.Engine {
	return &Mem{
		Hosts:     map[engine.HostKey]engine.Host{},
		Frontends: map[engine.FrontendKey]engine.Frontend{},
		Backends:  map[engine.BackendKey]engine.Backend{},

		Listeners:   map[engine.ListenerKey]engine.Listener{},
		Middlewares: map[engine.FrontendKey][]engine.Middleware{},
		Servers:     map[engine.BackendKey][]engine.Server{},
		Registry:    r,
		ChangesC:    make(chan interface{}, 1000),
		ErrorsC:     make(chan error),
	}
}

func (m *Mem) emit(val interface{}) {
	select {
	case m.ChangesC <- val:
	default:
	}
}

func (m *Mem) Close() {
}

func (m *Mem) GetLogSeverity() log.Level {
	return m.LogSeverity
}

func (m *Mem) SetLogSeverity(sev log.Level) {
	m.LogSeverity = sev
	log.SetLevel(m.LogSeverity)
}

func (m *Mem) GetRegistry() *plugin.Registry {
	return m.Registry
}

func (m *Mem) GetHosts() ([]engine.Host, error) {
	out := make([]engine.Host, 0, len(m.Hosts))
	for _, h := range m.Hosts {
		out = append(out, h)
	}
	return out, nil
}

func (m *Mem) GetHost(k engine.HostKey) (*engine.Host, error) {
	h, ok := m.Hosts[k]
	if !ok {
		return nil, &engine.NotFoundError{}
	}
	return &h, nil
}

func (m *Mem) UpsertHost(h engine.Host) error {
	m.Hosts[engine.HostKey{Name: h.Name}] = h
	m.emit(&engine.HostUpserted{Host: h})
	return nil
}

func (m *Mem) DeleteHost(k engine.HostKey) error {
	if _, ok := m.Hosts[k]; !ok {
		return &engine.NotFoundError{}
	}
	delete(m.Hosts, k)
	m.emit(&engine.HostDeleted{HostKey: k})
	return nil
}

func (m *Mem) GetListeners() ([]engine.Listener, error) {
	out := make([]engine.Listener, 0, len(m.Listeners))
	for _, l := range m.Listeners {
		out = append(out, l)
	}
	return out, nil
}

func (m *Mem) GetListener(lk engine.ListenerKey) (*engine.Listener, error) {
	val, ok := m.Listeners[lk]
	if !ok {
		return nil, &engine.NotFoundError{}
	}
	return &val, nil
}

func (m *Mem) UpsertListener(l engine.Listener) error {
	defer func() {
		m.emit(&engine.ListenerUpserted{Listener: l})
	}()
	lk := engine.ListenerKey{l.Id}
	m.Listeners[lk] = l
	return nil
}

func (m *Mem) DeleteListener(lk engine.ListenerKey) error {
	if _, ok := m.Listeners[lk]; !ok {
		return &engine.NotFoundError{}
	}
	delete(m.Listeners, lk)
	m.emit(&engine.ListenerDeleted{ListenerKey: lk})
	return nil
}

func (m *Mem) GetFrontends() ([]engine.Frontend, error) {
	out := make([]engine.Frontend, 0, len(m.Frontends))
	for _, h := range m.Frontends {
		out = append(out, h)
	}
	return out, nil
}

func (m *Mem) GetFrontend(k engine.FrontendKey) (*engine.Frontend, error) {
	f, ok := m.Frontends[k]
	if !ok {
		return nil, &engine.NotFoundError{}
	}
	return &f, nil
}

func (m *Mem) UpsertFrontend(f engine.Frontend, d time.Duration) error {
	if _, ok := m.Backends[engine.BackendKey{Id: f.BackendId}]; !ok {
		return &engine.NotFoundError{Message: fmt.Sprintf("backend: %v not found", f.BackendId)}
	}
	m.Frontends[engine.FrontendKey{Id: f.Id}] = f
	m.emit(&engine.FrontendUpserted{Frontend: f})
	return nil
}

func (m *Mem) DeleteFrontend(fk engine.FrontendKey) error {
	if _, ok := m.Frontends[fk]; !ok {
		return &engine.NotFoundError{}
	}
	m.emit(&engine.FrontendDeleted{FrontendKey: fk})
	delete(m.Frontends, fk)
	return nil
}

func (m *Mem) GetMiddlewares(fk engine.FrontendKey) ([]engine.Middleware, error) {
	vals, ok := m.Middlewares[fk]
	if !ok {
		return []engine.Middleware{}, nil
	}
	return vals, nil
}

func (m *Mem) GetMiddleware(mk engine.MiddlewareKey) (*engine.Middleware, error) {
	vals, ok := m.Middlewares[mk.FrontendKey]
	if !ok {
		return nil, &engine.NotFoundError{Message: fmt.Sprintf("'%v' not found", mk.FrontendKey)}
	}
	for _, v := range vals {
		if v.Id == mk.Id {
			return &v, nil
		}
	}
	return nil, &engine.NotFoundError{Message: fmt.Sprintf("'%v' not found", mk)}
}

func (m *Mem) UpsertMiddleware(fk engine.FrontendKey, md engine.Middleware, d time.Duration) error {
	if _, ok := m.Frontends[fk]; !ok {
		return &engine.NotFoundError{Message: fmt.Sprintf("'%v' not found", fk)}
	}
	defer func() {
		m.emit(&engine.MiddlewareUpserted{FrontendKey: fk, Middleware: md})
	}()
	vals, ok := m.Middlewares[fk]
	if !ok {
		m.Middlewares[fk] = []engine.Middleware{md}
		return nil
	}
	for i, v := range vals {
		if v.Id == md.Id {
			vals[i] = md
			return nil
		}
	}
	vals = append(vals, md)
	m.Middlewares[fk] = vals
	return nil
}

func (m *Mem) DeleteMiddleware(mk engine.MiddlewareKey) error {
	vals, ok := m.Middlewares[mk.FrontendKey]
	if !ok {
		return &engine.NotFoundError{}
	}
	for i, v := range vals {
		if v.Id == mk.Id {
			vals = append(vals[:i], vals[i+1:]...)
			m.Middlewares[mk.FrontendKey] = vals
			m.emit(&engine.MiddlewareDeleted{MiddlewareKey: mk})
			return nil
		}
	}
	return &engine.NotFoundError{}
}

func (m *Mem) GetBackends() ([]engine.Backend, error) {
	out := make([]engine.Backend, 0, len(m.Backends))
	for _, h := range m.Backends {
		out = append(out, h)
	}
	return out, nil
}

func (m *Mem) GetBackend(bk engine.BackendKey) (*engine.Backend, error) {
	f, ok := m.Backends[bk]
	if !ok {
		return nil, &engine.NotFoundError{}
	}
	return &f, nil
}

func (m *Mem) UpsertBackend(b engine.Backend) error {
	m.emit(&engine.BackendUpserted{Backend: b})
	m.Backends[engine.BackendKey{Id: b.Id}] = b
	return nil
}

func (m *Mem) DeleteBackend(bk engine.BackendKey) error {
	for _, f := range m.Frontends {
		if f.BackendId == bk.Id {
			return fmt.Errorf("Backend is in use by %v", f)
		}
	}
	if _, ok := m.Backends[bk]; !ok {
		return &engine.NotFoundError{}
	}
	m.emit(&engine.BackendDeleted{BackendKey: bk})
	delete(m.Backends, bk)
	return nil
}

func (m *Mem) GetServers(bk engine.BackendKey) ([]engine.Server, error) {
	vals, ok := m.Servers[bk]
	if !ok {
		return []engine.Server{}, nil
	}
	return vals, nil
}

func (m *Mem) GetServer(sk engine.ServerKey) (*engine.Server, error) {
	vals, ok := m.Servers[sk.BackendKey]
	if !ok {
		return nil, &engine.NotFoundError{}
	}
	for _, v := range vals {
		if v.Id == sk.Id {
			return &v, nil
		}
	}
	return nil, &engine.NotFoundError{}
}

func (m *Mem) UpsertServer(bk engine.BackendKey, srv engine.Server, d time.Duration) error {
	defer func() {
		m.emit(&engine.ServerUpserted{BackendKey: bk, Server: srv})
	}()
	vals, ok := m.Servers[bk]
	if !ok {
		m.Servers[bk] = []engine.Server{srv}
		return nil
	}
	for i, v := range vals {
		if v.Id == srv.Id {
			m.Servers[bk][i] = srv
			return nil
		}
	}
	m.Servers[bk] = append(vals, srv)
	return nil
}

func (m *Mem) DeleteServer(sk engine.ServerKey) error {
	vals, ok := m.Servers[sk.BackendKey]
	if !ok {
		return &engine.NotFoundError{}
	}
	for i, v := range vals {
		if v.Id == sk.Id {
			vals = append(vals[:i], vals[i+1:]...)
			m.Servers[sk.BackendKey] = vals
			m.emit(&engine.ServerDeleted{ServerKey: sk})
			return nil
		}
	}
	return &engine.NotFoundError{}
}

func (m *Mem) Subscribe(changes chan interface{}, cancelC chan bool) error {
	for {
		select {
		case <-cancelC:
			return nil
		case change := <-m.ChangesC:
			log.Infof("Got change channel: %v", change)
			select {
			case changes <- change:
				log.Infof("Got changes from changes channel: %v", changes)
			case err := <-m.ErrorsC:
				log.Infof("Returning error (while reading changes from channel): %v", err)
				return err
			}
		case err := <-m.ErrorsC:
			log.Infof("Returning error (before changes channel was acquired.): %v", err)
			return err
		}
	}
}
