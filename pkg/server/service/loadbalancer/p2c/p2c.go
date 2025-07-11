package p2c

import (
	"context"
	"errors"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/server/service/loadbalancer"
)

type namedHandler struct {
	http.Handler

	// name is the handler name.
	name string
	// inflight is the number of inflight requests.
	// It is used to implement the "power-of-two-random-choices" algorithm.
	inflight atomic.Int64
}

func (h *namedHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	h.inflight.Add(1)
	defer h.inflight.Add(-1)

	h.Handler.ServeHTTP(rw, req)
}

type rnd interface {
	Intn(n int) int
}

// Balancer implements the power-of-two-random-choices algorithm for load balancing.
// The idea is to randomly select two of the available backends and choose the one with the fewest in-flight requests.
// This algorithm balances the load more effectively than a round-robin approach, while maintaining a constant time for the selection:
// The strategy also has more advantageous "herd" behavior than the "fewest connections" algorithm, especially when the load balancer
// doesn't have perfect knowledge of the global number of connections to the backend, for example, when running in a distributed fashion.
type Balancer struct {
	wantsHealthCheck bool

	// handlersMu is a mutex to protect the handlers slice, the status and the fenced maps.
	handlersMu sync.RWMutex
	handlers   []*namedHandler
	// status is a record of which child services of the Balancer are healthy, keyed
	// by name of child service. A service is initially added to the map when it is
	// created via Add, and it is later removed or added to the map as needed,
	// through the SetStatus method.
	status map[string]struct{}
	// fenced is the list of terminating yet still serving child services.
	fenced map[string]struct{}

	// updaters is the list of hooks that are run (to update the Balancer
	// parent(s)), whenever the Balancer status changes.
	updaters []func(bool)

	sticky *loadbalancer.Sticky

	randMu sync.Mutex
	rand   rnd
}

// New creates a new power-of-two-random-choices load balancer.
func New(stickyConfig *dynamic.Sticky, wantsHealthCheck bool) *Balancer {
	balancer := &Balancer{
		status:           make(map[string]struct{}),
		fenced:           make(map[string]struct{}),
		wantsHealthCheck: wantsHealthCheck,
		rand:             rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	if stickyConfig != nil && stickyConfig.Cookie != nil {
		balancer.sticky = loadbalancer.NewSticky(*stickyConfig.Cookie)
	}

	return balancer
}

// SetStatus sets on the balancer that its given child is now of the given
// status. balancerName is only needed for logging purposes.
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
	}

	upAfter := len(b.status) > 0
	status = "DOWN"
	if upAfter {
		status = "UP"
	}

	// No Status Change
	if upBefore == upAfter {
		// We're still with the same status, no need to propagate
		log.Ctx(ctx).Debug().Msgf("Still %s, no need to propagate", status)
		return
	}

	// Status Change
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
		return errors.New("healthCheck not enabled in config for this weighted service")
	}
	b.updaters = append(b.updaters, fn)
	return nil
}

var errNoAvailableServer = errors.New("no available server")

func (b *Balancer) nextServer() (*namedHandler, error) {
	// We kept the same representation (map) as in the WRR strategy to improve maintainability.
	// However, with the P2C strategy, we only need a slice of healthy servers.
	b.handlersMu.RLock()
	var healthy []*namedHandler
	for _, h := range b.handlers {
		if _, ok := b.status[h.name]; ok {
			if _, fenced := b.fenced[h.name]; !fenced {
				healthy = append(healthy, h)
			}
		}
	}
	b.handlersMu.RUnlock()

	if len(healthy) == 0 {
		return nil, errNoAvailableServer
	}

	// If there is only one healthy server, return it.
	if len(healthy) == 1 {
		return healthy[0], nil
	}
	// In order to not get the same backend twice, we make the second call to s.rand.Intn one fewer
	// than the length of the slice. We then have to shift over the second index if it is equal or
	// greater than the first index, wrapping round if needed.
	b.randMu.Lock()
	n1, n2 := b.rand.Intn(len(healthy)), b.rand.Intn(len(healthy))
	b.randMu.Unlock()

	if n2 == n1 {
		n2 = (n2 + 1) % len(healthy)
	}

	h1, h2 := healthy[n1], healthy[n2]
	// Ensure h1 has fewer inflight requests than h2.
	if h2.inflight.Load() < h1.inflight.Load() {
		log.Debug().Msgf("Service selected by P2C: %s", h2.name)
		return h2, nil
	}

	log.Debug().Msgf("Service selected by P2C: %s", h1.name)
	return h1, nil
}

func (b *Balancer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if b.sticky != nil {
		h, rewrite, err := b.sticky.StickyHandler(req)
		if err != nil {
			log.Error().Err(err).Msg("Error while getting sticky handler")
		} else if h != nil {
			b.handlersMu.RLock()
			_, ok := b.status[h.Name]
			b.handlersMu.RUnlock()
			if ok {
				if rewrite {
					if err := b.sticky.WriteStickyCookie(rw, h.Name); err != nil {
						log.Error().Err(err).Msg("Writing sticky cookie")
					}
				}

				h.ServeHTTP(rw, req)
				return
			}
		}
	}

	server, err := b.nextServer()
	if err != nil {
		if errors.Is(err, errNoAvailableServer) {
			http.Error(rw, errNoAvailableServer.Error(), http.StatusServiceUnavailable)
		} else {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if b.sticky != nil {
		if err := b.sticky.WriteStickyCookie(rw, server.name); err != nil {
			log.Error().Err(err).Msg("Error while writing sticky cookie")
		}
	}

	server.ServeHTTP(rw, req)
}

// AddServer adds a handler with a server.
func (b *Balancer) AddServer(name string, handler http.Handler, server dynamic.Server) {
	h := &namedHandler{Handler: handler, name: name}

	b.handlersMu.Lock()
	b.handlers = append(b.handlers, h)
	b.status[name] = struct{}{}
	if server.Fenced {
		b.fenced[name] = struct{}{}
	}
	b.handlersMu.Unlock()

	if b.sticky != nil {
		b.sticky.AddHandler(name, h)
	}
}
