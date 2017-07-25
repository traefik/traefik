package proxy

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"

	log "github.com/Sirupsen/logrus"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/roundrobin"
	"github.com/vulcand/oxy/stream"
	"github.com/vulcand/oxy/utils"
	"github.com/vulcand/vulcand/engine"
)

type frontend struct {
	key         engine.FrontendKey
	mux         *mux
	frontend    engine.Frontend
	lb          *roundrobin.Rebalancer
	handler     http.Handler
	watcher     *RTWatcher
	backend     *backend
	middlewares map[engine.MiddlewareKey]engine.Middleware
	log         utils.Logger
}

func newFrontend(m *mux, f engine.Frontend, b *backend) (*frontend, error) {
	fr := &frontend{
		key:         engine.FrontendKey{Id: f.Id},
		frontend:    f,
		mux:         m,
		backend:     b,
		middlewares: make(map[engine.MiddlewareKey]engine.Middleware),
		log:         log.StandardLogger(),
	}

	if err := fr.rebuild(); err != nil {
		return nil, err
	}
	b.linkFrontend(engine.FrontendKey{f.Id}, fr)
	return fr, nil
}

func (f *frontend) String() string {
	return fmt.Sprintf("%v frontend(wrap=%v)", f.mux, &f.frontend)
}

// syncs backend servers and rebalancer state
func syncServers(m *mux, rb *roundrobin.Rebalancer, backend *backend, w *RTWatcher) error {
	// First, collect and parse servers to add
	newServers := map[string]*url.URL{}
	for _, s := range backend.servers {
		u, err := url.Parse(s.URL)
		if err != nil {
			return fmt.Errorf("failed to parse url %v", s.URL)
		}
		newServers[s.URL] = u
	}

	// Memorize what endpoints exist in load balancer at the moment
	existingServers := map[string]*url.URL{}
	for _, s := range rb.Servers() {
		existingServers[s.String()] = s
	}

	// First, add endpoints, that should be added and are not in lb
	for _, s := range newServers {
		if _, exists := existingServers[s.String()]; !exists {
			if err := rb.UpsertServer(s); err != nil {
				log.Errorf("%v failed to add %v, err: %s", m, s, err)
			} else {
				log.Infof("%v add %v", m, s)
			}
			w.upsertServer(s)
		}
	}

	// Second, remove endpoints that should not be there any more
	for k, v := range existingServers {
		if _, exists := newServers[k]; !exists {
			if err := rb.RemoveServer(v); err != nil {
				log.Errorf("%v failed to remove %v, err: %v", m, v, err)
			} else {
				log.Infof("%v removed %v", m, v)
			}
			w.removeServer(v)
		}
	}
	return nil
}

func (f *frontend) updateTransport(t *http.Transport) error {
	return f.rebuild()
}

func (f *frontend) sortedMiddlewares() []engine.Middleware {
	vals := make([]engine.Middleware, 0, len(f.middlewares))
	for _, m := range f.middlewares {
		vals = append(vals, m)
	}
	sort.Sort(sort.Reverse(&middlewareSorter{ms: vals}))
	return vals
}

func (f *frontend) rebuild() error {
	settings := f.frontend.HTTPSettings()

	// set up forwarder
	fwd, err := forward.New(
		forward.Logger(f.log),
		forward.RoundTripper(f.backend.transport),
		forward.Rewriter(
			&forward.HeaderRewriter{
				Hostname:           settings.Hostname,
				TrustForwardHeader: settings.TrustForwardHeader,
			}),
		forward.PassHostHeader(settings.PassHostHeader))

	// rtwatcher will be observing and aggregating metrics
	watcher, err := NewWatcher(fwd)
	if err != nil {
		return err
	}

	// Create a load balancer
	rr, err := roundrobin.New(watcher)
	if err != nil {
		return err
	}

	// Rebalancer will readjust load balancer weights based on error ratios
	rb, err := roundrobin.NewRebalancer(rr, roundrobin.RebalancerLogger(f.log))
	if err != nil {
		return err
	}

	// create middlewares sorted by priority and chain them
	middlewares := f.sortedMiddlewares()
	handlers := make([]http.Handler, len(middlewares))
	for i, m := range middlewares {
		var prev http.Handler
		if i == 0 {
			prev = rb
		} else {
			prev = handlers[i-1]
		}
		h, err := m.Middleware.NewHandler(prev)
		if err != nil {
			return err
		}
		handlers[i] = h
	}

	var next http.Handler
	if len(handlers) != 0 {
		next = handlers[len(handlers)-1]
	} else {
		next = rb
	}

	// stream will retry and replay requests, fix encodings
	if settings.FailoverPredicate == "" {
		settings.FailoverPredicate = `IsNetworkError() && RequestMethod() == "GET" && Attempts() < 2`
	}
	str, err := stream.New(next,
		stream.Logger(f.log),
		stream.Retry(settings.FailoverPredicate),
		stream.MaxRequestBodyBytes(settings.Limits.MaxBodyBytes),
		stream.MemRequestBodyBytes(settings.Limits.MaxMemBodyBytes))
	if err != nil {
		return err
	}

	if err := syncServers(f.mux, rb, f.backend, watcher); err != nil {
		return err
	}

	// Add the frontend to the router
	if err := f.mux.router.Handle(f.frontend.Route, str); err != nil {
		return err
	}

	f.lb = rb
	f.handler = str
	f.watcher = watcher
	return nil
}

func (f *frontend) upsertMiddleware(fk engine.FrontendKey, mi engine.Middleware) error {
	f.middlewares[engine.MiddlewareKey{FrontendKey: fk, Id: mi.Id}] = mi
	return f.rebuild()
}

func (f *frontend) deleteMiddleware(mk engine.MiddlewareKey) error {
	delete(f.middlewares, mk)
	return f.rebuild()
}

func (f *frontend) updateBackend(b *backend) error {
	oldb := f.backend
	f.backend = b

	// Switching backends, set the new transport and perform switch
	if b.backend.Id != oldb.backend.Id {
		log.Infof("%v updating backend from %v to %v", f, &oldb, &f.backend)
		oldb.unlinkFrontend(f.key)
		b.linkFrontend(f.key, f)
		return f.rebuild()
	}
	return syncServers(f.mux, f.lb, f.backend, f.watcher)
}

// TODO: implement rollback in case of suboperation failure
func (f *frontend) update(ef engine.Frontend, b *backend) error {
	oldf := f.frontend
	f.frontend = ef

	if err := f.updateBackend(b); err != nil {
		return err
	}

	if oldf.Route != ef.Route {
		log.Infof("%v updating route from %v to %v", oldf.Route, ef.Route)
		if err := f.mux.router.Handle(ef.Route, f.handler); err != nil {
			return err
		}
		if err := f.mux.router.Remove(oldf.Route); err != nil {
			return err
		}
	}

	olds := oldf.HTTPSettings()
	news := ef.HTTPSettings()
	if !olds.Equals(news) {
		if err := f.rebuild(); err != nil {
			return err
		}
	}

	return nil
}

func (f *frontend) remove() error {
	f.backend.unlinkFrontend(f.key)
	return f.mux.router.Remove(f.frontend.Route)
}

type middlewareSorter struct {
	ms []engine.Middleware
}

func (s *middlewareSorter) Len() int {
	return len(s.ms)
}

func (s *middlewareSorter) Swap(i, j int) {
	s.ms[i], s.ms[j] = s.ms[j], s.ms[i]
}

func (s *middlewareSorter) Less(i, j int) bool {
	return s.ms[i].Priority < s.ms[j].Priority
}
