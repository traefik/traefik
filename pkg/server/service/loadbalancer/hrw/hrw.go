package hrw

import (
	"context"
	"errors"
	"hash/fnv"
	"math"
	"net/http"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/ip"
)

var errNoAvailableServer = errors.New("no available server")

type namedHandler struct {
	http.Handler
	name   string
	weight float64
}

// Balancer implements the Rendezvous Hashing algorithm for load balancing.
// The idea is to compute a score for each available backend using a hash of the client's
// source (for example, IP) combined with the backend's identifier, and assign the client
// to the backend with the highest score. This ensures that each client consistently
// connects to the same backend while distributing load evenly across all backends.
type Balancer struct {
	wantsHealthCheck bool

	strategy ip.RemoteAddrStrategy

	handlersMu sync.RWMutex
	// References all the handlers by name and also by the hashed value of the name.
	handlers []*namedHandler
	// status is a record of which child services of the Balancer are healthy, keyed
	// by name of child service. A service is initially added to the map when it is
	// created via Add, and it is later removed or added to the map as needed,
	// through the SetStatus method.
	status map[string]struct{}
	// updaters is the list of hooks that are run (to update the Balancer
	// parent(s)), whenever the Balancer status changes.
	// No mutex is needed, as it is modified only during the configuration build.
	updaters []func(bool)
	// fenced is the list of terminating yet still serving child services.
	fenced map[string]struct{}
}

// New creates a new load balancer.
func New(wantHealthCheck bool) *Balancer {
	balancer := &Balancer{
		status:           make(map[string]struct{}),
		fenced:           make(map[string]struct{}),
		wantsHealthCheck: wantHealthCheck,
		strategy:         ip.RemoteAddrStrategy{},
	}

	return balancer
}

// getNodeScore calculates the score of the couple of src and handler name.
func getNodeScore(handler *namedHandler, src string) float64 {
	h := fnv.New64a()
	h.Write([]byte(src + handler.name))
	sum := h.Sum64()
	score := float64(sum) / math.Pow(2, 64)
	logScore := 1.0 / -math.Log(score)

	return logScore * handler.weight
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
		return errors.New("healthCheck not enabled in config for this HRW service")
	}
	b.updaters = append(b.updaters, fn)

	return nil
}

func (b *Balancer) nextServer(ip string) (*namedHandler, error) {
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

	var handler *namedHandler
	score := 0.0
	for _, h := range healthy {
		s := getNodeScore(h, ip)
		if s > score {
			handler = h
			score = s
		}
	}

	log.Debug().Msgf("Service selected by HRW: %s", handler.name)

	return handler, nil
}

func (b *Balancer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// give ip fetched to b.nextServer
	clientIP := b.strategy.GetIP(req)
	log.Debug().Msgf("ServeHTTP() clientIP=%s", clientIP)

	server, err := b.nextServer(clientIP)
	if err != nil {
		if errors.Is(err, errNoAvailableServer) {
			http.Error(w, errNoAvailableServer.Error(), http.StatusServiceUnavailable)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	server.ServeHTTP(w, req)
}

// AddServer adds a handler with a server.
func (b *Balancer) AddServer(name string, handler http.Handler, server dynamic.Server) {
	b.Add(name, handler, server.Weight, server.Fenced)
}

// Add adds a handler.
// A handler with a non-positive weight is ignored.
func (b *Balancer) Add(name string, handler http.Handler, weight *int, fenced bool) {
	w := 1
	if weight != nil {
		w = *weight
	}

	if w <= 0 { // non-positive weight is meaningless
		return
	}

	h := &namedHandler{Handler: handler, name: name, weight: float64(w)}

	b.handlersMu.Lock()
	b.handlers = append(b.handlers, h)
	b.status[name] = struct{}{}
	if fenced {
		b.fenced[name] = struct{}{}
	}
	b.handlersMu.Unlock()
}
