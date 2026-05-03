package affinity

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

var errNoAvailableServer = errors.New("no available server")

type namedHandler struct {
	http.Handler
	name string
}

// Balancer consistently routes requests to the same backend
// based on a key extracted from the request path with a regex or a header.
type Balancer struct {
	wantsHealthCheck bool

	handlersMu sync.RWMutex
	handlers   []*namedHandler
	// status is a record of which child services of the Balancer are healthy, keyed
	// by name of child service. A service is initially added to the map when it is
	// created via Add, and it is later removed or added to the map as needed,
	// through the SetStatus method.
	status map[string]struct{}
	// updaters is the list of hooks that are run (to update the Balancer
	// parent(s)), whenever the Balancer status changes.
	// No mutex is needed, as it is modified only during the configuration build.
	updaters []func(bool)

	ringMu sync.Mutex
	ring   map[string]string // affinity key → server name

	regex  *regexp.Regexp
	header string
}

// New creates a new affinity load balancer.
func New(cfg *dynamic.AffinityConfig, wantsHealthCheck bool) *Balancer {
	b := &Balancer{
		status:           make(map[string]struct{}),
		ring:             make(map[string]string),
		wantsHealthCheck: wantsHealthCheck,
	}

	if cfg == nil {
		return b
	}

	b.header = cfg.HeaderName

	if cfg.Regex != "" {
		re, err := regexp.Compile(cfg.Regex)
		if err != nil {
			log.Error().Err(err).Msg("Invalid affinity regex, falling back to full path")
		} else {
			b.regex = re
		}
	}

	return b
}

// SetStatus sets on the balancer that its given child is now of the given
// status. childName is only needed for logging purposes.
func (b *Balancer) SetStatus(ctx context.Context, childName string, up bool) {
	b.handlersMu.Lock()
	defer b.handlersMu.Unlock()

	upBefore := len(b.status) > 0

	status := "DOWN"
	if up {
		status = "UP"
	}

	log.Ctx(ctx).Debug().Msgf("Setting status of %s to %v", childName, status)

	if up {
		b.status[childName] = struct{}{}
	} else {
		delete(b.status, childName)

		// Evict any pinned sessions that pointed at this server.
		b.ringMu.Lock()
		for key, name := range b.ring {
			if name == childName {
				delete(b.ring, key)
			}
		}
		b.ringMu.Unlock()
	}

	upAfter := len(b.status) > 0
	status = "DOWN"
	if upAfter {
		status = "UP"
	}

	// No Status Change.
	if upBefore == upAfter {
		log.Ctx(ctx).Debug().Msgf("Still %s, no need to propagate", status)
		return
	}

	// Status Change.
	log.Ctx(ctx).Debug().Msgf("Propagating new %s status", status)
	for _, fn := range b.updaters {
		fn(upAfter)
	}
}

// RegisterStatusUpdater adds fn to the list of hooks that are run when the
// status of the Balancer changes.
// Not thread safe.
func (b *Balancer) RegisterStatusUpdater(fn func(up bool)) error {
	if !b.wantsHealthCheck {
		return errors.New("healthCheck not enabled in config for this Affinity service")
	}
	b.updaters = append(b.updaters, fn)
	return nil
}

func (b *Balancer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	server, err := b.nextServer(req)
	if err != nil {
		if errors.Is(err, errNoAvailableServer) {
			http.Error(rw, errNoAvailableServer.Error(), http.StatusServiceUnavailable)
		} else {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	log.Debug().Msgf("Service selected by affinity: %s", server.name)
	server.ServeHTTP(rw, req)
}

// AddServer adds a handler with a server.
func (b *Balancer) AddServer(name string, handler http.Handler, _ dynamic.Server) {
	h := &namedHandler{Handler: handler, name: name}

	b.handlersMu.Lock()
	b.handlers = append(b.handlers, h)
	b.status[name] = struct{}{}
	b.handlersMu.Unlock()
}

func (b *Balancer) nextServer(req *http.Request) (*namedHandler, error) {
	b.handlersMu.RLock()
	var healthy []*namedHandler
	for _, h := range b.handlers {
		if _, ok := b.status[h.name]; ok {
			healthy = append(healthy, h)
		}
	}
	b.handlersMu.RUnlock()

	if len(healthy) == 0 {
		return nil, errNoAvailableServer
	}

	key := b.extractKey(req)
	if key == "" {
		// No affinity key — fall back to full path hash, still deterministic.
		key = req.URL.Path
	}

	b.ringMu.Lock()
	defer b.ringMu.Unlock()

	if name, ok := b.ring[key]; ok {
		// Check the pinned server is still healthy.
		for _, h := range healthy {
			if h.name == name {
				return h, nil
			}
		}
		// Pinned server no longer healthy — re-assign.
		delete(b.ring, key)
	}

	h := healthy[jumpHash(key, len(healthy))]
	b.ring[key] = h.name
	return h, nil
}

func (b *Balancer) extractKey(req *http.Request) string {
	if b.regex != nil {
		if m := b.regex.FindStringSubmatch(req.URL.Path); len(m) >= 2 {
			return m[1]
		}
	}
	if b.header != "" {
		return req.Header.Get(b.header)
	}
	return ""
}

func jumpHash(key string, n int) int {
	var h uint64
	for _, c := range []byte(key) {
		h = h*31 + uint64(c)
	}
	var b, j int64 = -1, 0
	for j < int64(n) {
		b = j
		h = h*2862933555777941757 + 1
		j = int64(float64(b+1) * (float64(int64(1)<<31) / float64((h>>33)+1)))
	}
	return int(b)
}
