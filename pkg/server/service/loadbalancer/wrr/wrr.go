package wrr

import (
	"container/heap"
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/server/service/loadbalancer"
)

type namedHandler struct {
	http.Handler
	name     string
	weight   float64
	deadline float64
}

// Balancer is a WeightedRoundRobin load balancer based on Earliest Deadline First (EDF).
// (https://en.wikipedia.org/wiki/Earliest_deadline_first_scheduling)
// Each pick from the schedule has the earliest deadline entry selected.
// Entries have deadlines set at currentDeadline + 1 / weight,
// providing weighted round-robin behavior with floating point weights and an O(log n) pick time.
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

	curDeadline float64
}

// New creates a new load balancer.
func New(sticky *dynamic.Sticky, wantsHealthCheck bool) *Balancer {
	balancer := &Balancer{
		status:           make(map[string]struct{}),
		fenced:           make(map[string]struct{}),
		wantsHealthCheck: wantsHealthCheck,
	}
	if sticky != nil && sticky.Cookie != nil {
		balancer.sticky = loadbalancer.NewSticky(*sticky.Cookie)
	}

	return balancer
}

// Len implements heap.Interface/sort.Interface.
func (b *Balancer) Len() int { return len(b.handlers) }

// Less implements heap.Interface/sort.Interface.
func (b *Balancer) Less(i, j int) bool {
	return b.handlers[i].deadline < b.handlers[j].deadline
}

// Swap implements heap.Interface/sort.Interface.
func (b *Balancer) Swap(i, j int) {
	b.handlers[i], b.handlers[j] = b.handlers[j], b.handlers[i]
}

// Push implements heap.Interface for pushing an item into the heap.
func (b *Balancer) Push(x interface{}) {
	h, ok := x.(*namedHandler)
	if !ok {
		return
	}

	b.handlers = append(b.handlers, h)
}

// Pop implements heap.Interface for popping an item from the heap.
// It panics if b.Len() < 1.
func (b *Balancer) Pop() interface{} {
	h := b.handlers[len(b.handlers)-1]
	b.handlers = b.handlers[0 : len(b.handlers)-1]
	return h
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
	b.handlersMu.Lock()
	defer b.handlersMu.Unlock()

	if len(b.handlers) == 0 || len(b.status) == 0 || len(b.fenced) == len(b.handlers) {
		return nil, errNoAvailableServer
	}

	var handler *namedHandler
	for {
		// Pick handler with closest deadline.
		handler = heap.Pop(b).(*namedHandler)

		// curDeadline should be handler's deadline so that new added entry would have a fair competition environment with the old ones.
		b.curDeadline = handler.deadline
		handler.deadline += 1 / handler.weight

		heap.Push(b, handler)
		if _, ok := b.status[handler.name]; ok {
			if _, ok := b.fenced[handler.name]; !ok {
				// do not select a fenced handler.
				break
			}
		}
	}

	log.Debug().Msgf("Service selected by WRR: %s", handler.name)
	return handler, nil
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
	h.deadline = b.curDeadline + 1/h.weight
	heap.Push(b, h)
	b.status[name] = struct{}{}
	if fenced {
		b.fenced[name] = struct{}{}
	}
	b.handlersMu.Unlock()

	if b.sticky != nil {
		b.sticky.AddHandler(name, handler)
	}
}
