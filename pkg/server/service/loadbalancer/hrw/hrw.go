package hrw

import (
	// "container/heap"
	"context"
	"errors"
	"hash/fnv"
	"math"
	"net/http"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/ip"
	// "github.com/traefik/traefik/v3/pkg/config/dynamic"
)

type namedHandler struct {
	http.Handler
	name   string
	weight float64
}

// Balancer is a HRW load balancer based on RENDEZVOUS (HRW).
// (https://en.m.wikipedia.org/wiki/Rendezvous_hashing)
// providing weighted round-robin behavior with floating point weights and an O(n) pick time.
type Balancer struct {
	wantsHealthCheck bool
	// checker          ip.Checker
	// strategy         ip.Strategy

	mutex    sync.RWMutex
	handlers []*namedHandler
	// status is a record of which child services of the Balancer are healthy, keyed
	// by name of child service. A service is initially added to the map when it is
	// created via Add, and it is later removed or added to the map as needed,
	// through the SetStatus method.
	status map[string]struct{}
	// updaters is the list of hooks that are run (to update the Balancer
	// parent(s)), whenever the Balancer status changes.
	updaters []func(bool)
}

// New creates a new load balancer.
func New(wantHealthCheck bool) *Balancer {
	log.Debug().Msgf("New() 1")
	balancer := &Balancer{
		status:           make(map[string]struct{}),
		wantsHealthCheck: wantHealthCheck,
	}
	log.Debug().Msgf("New() 2")
	return balancer
}

// // Len implements heap.Interface/sort.Interface.
// func (b *Balancer) Len() int { return len(b.handlers) }

// // Less implements heap.Interface/sort.Interface.
// func (b *Balancer) Less(i, j int) bool {
// 	return b.handlers[i].deadline < b.handlers[j].deadline
// }

// // Swap implements heap.Interface/sort.Interface.
// func (b *Balancer) Swap(i, j int) {
// 	b.handlers[i], b.handlers[j] = b.handlers[j], b.handlers[i]
// }

// Push implements something for pushing an item into the list.
func (b *Balancer) Push(x interface{}) {
	h, ok := x.(*namedHandler)
	if !ok {
		return
	}
	b.handlers = append(b.handlers, h)
}

// // Pop implements heap.Interface for popping an item from the heap.
// // It panics if b.Len() < 1.
// func (b *Balancer) Pop() interface{} {
// 	h := b.handlers[len(b.handlers)-1]
// 	b.handlers = b.handlers[0 : len(b.handlers)-1]
// 	return h
// }

func getNodeScore(handler *namedHandler, src string) float64 {
	log.Debug().Msgf("getNodeScore() name=%s", src+(*handler).name)
	h := fnv.New32a()
	h.Write([]byte(src + (*handler).name))
	sum := h.Sum32()
	log.Debug().Msgf("getNodeScore() sum=%d", sum)
	score := float32(sum) / float32(math.Pow(2, 32))
	log.Debug().Msgf("getNodeScore() score=%f", score)
	log_score := 1.0 / -math.Log(float64(score))
	log.Debug().Msgf("getNodeScore() log_score=%f", score)
	return log_score
}

// SetStatus sets on the balancer that its given child is now of the given
// status. balancerName is only needed for logging purposes.
func (b *Balancer) SetStatus(ctx context.Context, childName string, up bool) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

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

func (b *Balancer) nextServer(ip string) (*namedHandler, error) {
	log.Debug().Msgf("nextServer()")
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if len(b.handlers) == 0 || len(b.status) == 0 {
		log.Debug().Msg("nextServer() len = 0")
		return nil, errNoAvailableServer
	}

	var handler *namedHandler
	score := 0.0
	for _, h := range b.handlers {
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
	log.Debug().Msgf("ServeHTTP()")
	// here give ip fetched to b.nextServer
	sourceRange := []string{}
	checker, _ := ip.NewChecker(sourceRange)

	strategy := &ip.PoolStrategy{
		Checker: checker,
	}
	strategyRM := ip.RemoteAddrStrategy{}

	clientIP := strategy.GetIP(req)
	clientIPRM := strategyRM.GetIP(req)
	log.Debug().Msgf("ServeHTTP() clientIP=%s", clientIP)
	log.Debug().Msgf("ServeHTTP() clientIP=%s", clientIPRM)

	server, err := b.nextServer(clientIP)
	if err != nil {
		log.Debug().Err(err).Msg("ServeHTTP() err")
		if errors.Is(err, errNoAvailableServer) {
			http.Error(w, errNoAvailableServer.Error(), http.StatusServiceUnavailable)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	server.ServeHTTP(w, req)
}

// Add adds a handler.
// A handler with a non-positive weight is ignored.
func (b *Balancer) Add(name string, handler http.Handler, weight *int) {
	log.Debug().Msgf("Add()")
	w := 1
	if weight != nil {
		w = *weight
	}

	if w <= 0 { // non-positive weight is meaningless
		return
	}

	h := &namedHandler{Handler: handler, name: name, weight: float64(w)}

	b.mutex.Lock()
	b.Push(h)
	b.status[name] = struct{}{}
	b.mutex.Unlock()
}
