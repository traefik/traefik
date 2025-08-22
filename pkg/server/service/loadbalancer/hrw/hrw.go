package hrw

import (
	"context"
	"errors"
	"hash/fnv"
	"math"
	"net/http"
	"strconv"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/ip"
)

type namedHandler struct {
	http.Handler
	name   string
	weight float64
}

// Balancer is a HRW load balancer based on RendezVous hashing Algorithm (HRW).
// (https://en.m.wikipedia.org/wiki/Rendezvous_hashing)
// providing weighted stateless sticky session behavior with floating point weights and an O(n) pick time.
// Client connects to the same server each time based on their IP source
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

// getNodeScore calcul the score of the couple of src and handler name.
func getNodeScore(handler *namedHandler, src string) float64 {
	h := fnv.New32a()
	h.Write([]byte(src + (*handler).name))
	sum := h.Sum32()
	score := float32(sum) / float32(math.Pow(2, 32))
	log_score := 1.0 / -math.Log(float64(score))
	log_weighted_score := log_score * (*handler).weight
	return log_weighted_score
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

func (b *Balancer) nextServer(ip string) (*namedHandler, error) {
	b.handlersMu.Lock()
	defer b.handlersMu.Unlock()

	if len(b.handlers) == 0 || len(b.status) == 0 {
		log.Debug().Msg("nextServer() len = 0")
		return nil, errNoAvailableServer
	}

	var handler *namedHandler
	score := 0.0
	for _, h := range b.handlers {
		// if handler healthy we calcul score
		// fmt.Printf("b.status = %s\n", b.status[h.name])
		if _, ok := b.status[h.name]; ok {
			s := getNodeScore(h, ip)
			if s > score {
				handler = h
				score = s
			}
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

func hash(input string) string {
	hasher := fnv.New64()
	// We purposely ignore the error because the implementation always returns nil.
	_, _ = hasher.Write([]byte(input))

	return strconv.FormatUint(hasher.Sum64(), 16)
}
