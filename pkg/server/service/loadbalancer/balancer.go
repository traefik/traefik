package loadbalancer

import (
	"context"
	"errors"
	"hash/fnv"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

type namedHandler struct {
	http.Handler
	name     string
	weight   float64
	deadline float64
	inflight atomic.Int64
}

func (h *namedHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.inflight.Add(1)
	defer h.inflight.Add(-1)
	h.Handler.ServeHTTP(w, req)
}

type stickyCookie struct {
	name     string
	secure   bool
	httpOnly bool
	sameSite string
	maxAge   int
}

func convertSameSite(sameSite string) http.SameSite {
	switch sameSite {
	case "none":
		return http.SameSiteNoneMode
	case "lax":
		return http.SameSiteLaxMode
	case "strict":
		return http.SameSiteStrictMode
	default:
		return http.SameSiteDefaultMode
	}
}

// strategy is an interface that can be used to implement different load balancing strategies
// for the Balancer.
type strategy interface {
	// nextServer returns the next server to serve a request, this is called under the handlersMu lock.
	// Each pick from the schedule has the earliest deadline entry selected. The status param is a
	// map of the currently healthy child services.
	nextServer(status map[string]struct{}) *namedHandler
	// Entries have deadlines set at currentDeadline + 1 / weight,
	// add adds a handler to the balancing algorithm, this is called under the handlersMu lock.
	// providing weighted round-robin behavior with floating point weights and an O(log n) pick time.
	add(h *namedHandler)

	setUp(name string, up bool)

	name() string
	len() int
}

type Balancer struct {
	stickyCookie     *stickyCookie
	wantsHealthCheck bool

	handlersMu sync.RWMutex
	// References all the handlers by name and also by the hashed value of the
	// name.
	handlerMap map[string]*namedHandler
	// status is a record of which child services of the Balancer are healthy,
	// keyed by name of child service. A service is initially added to the map
	// when it is created via Add, and it is later removed or added to the map
	// as needed, through the SetStatus method.
	status map[string]struct{}

	// strategy references the load balancing strategy to be used. The add and
	// nextServer method must be called under the handlersMu lock
	strategy strategy

	// updaters is the list of hooks that are run (to update the Balancer
	// parent(s)), whenever the Balancer status changes.
	updaters []func(bool)
}

// NewWRR creates a WeightedRoundRobin load balancer based on Earliest Deadline
// First (EDF).
// (https://en.wikipedia.org/wiki/Earliest_deadline_first_scheduling)
// Each pick from the schedule has the earliest deadline entry selected.
// Entries have deadlines set at currentDeadline + 1 / weight,
// providing weighted round-robin behavior with floating point weights and an
// O(log n) pick time.
func NewWRR(sticky *dynamic.Sticky, wantHealthCheck bool) *Balancer {
	return newBalancer(sticky, wantHealthCheck, newStrategyWRR())
}

// NewP2C creates a "the power-of-two-random-choices" algorithm for load
// balancing. The idea of this is two take two of the backends at random from
// the available backends, and select the backend that has the fewest in-flight
// requests. This is constant time when picking, and has more beneficial "herd"
// behavior than the "fewest connections" algorithm.
func NewP2C(sticky *dynamic.Sticky, wantHealthCheck bool) *Balancer {
	return newBalancer(sticky, wantHealthCheck, newStrategyP2C())
}

func newBalancer(sticky *dynamic.Sticky, wantHealthCheck bool, strategy strategy) *Balancer {
	balancer := &Balancer{
		status:           make(map[string]struct{}),
		handlerMap:       make(map[string]*namedHandler),
		wantsHealthCheck: wantHealthCheck,
		strategy:         strategy,
	}
	if sticky != nil && sticky.Cookie != nil {
		balancer.stickyCookie = &stickyCookie{
			name:     sticky.Cookie.Name,
			secure:   sticky.Cookie.Secure,
			httpOnly: sticky.Cookie.HTTPOnly,
			sameSite: sticky.Cookie.SameSite,
			maxAge:   sticky.Cookie.MaxAge,
		}
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

	b.strategy.setUp(childName, up)

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

	if b.strategy.len() == 0 || len(b.status) == 0 {
		return nil, errNoAvailableServer
	}

	handler := b.strategy.nextServer(b.status)

	log.Debug().Msgf("Service selected by strategy %q: %s", b.strategy.name(), handler.name)
	return handler, nil
}

func (b *Balancer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if b.stickyCookie != nil {
		cookie, err := req.Cookie(b.stickyCookie.name)

		if err != nil && !errors.Is(err, http.ErrNoCookie) {
			log.Warn().Err(err).Msg("Error while reading cookie")
		}

		if err == nil && cookie != nil {
			b.handlersMu.RLock()
			handler, ok := b.handlerMap[cookie.Value]
			b.handlersMu.RUnlock()

			if ok && handler != nil {
				b.handlersMu.RLock()
				_, isHealthy := b.status[handler.name]
				b.handlersMu.RUnlock()
				if isHealthy {
					handler.ServeHTTP(w, req)
					return
				}
			}
		}
	}

	server, err := b.nextServer()
	if err != nil {
		if errors.Is(err, errNoAvailableServer) {
			http.Error(w, errNoAvailableServer.Error(), http.StatusServiceUnavailable)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if b.stickyCookie != nil {
		cookie := &http.Cookie{
			Name:     b.stickyCookie.name,
			Value:    hash(server.name),
			Path:     "/",
			HttpOnly: b.stickyCookie.httpOnly,
			Secure:   b.stickyCookie.secure,
			SameSite: convertSameSite(b.stickyCookie.sameSite),
			MaxAge:   b.stickyCookie.maxAge,
		}
		http.SetCookie(w, cookie)
	}

	server.ServeHTTP(w, req)
}

// Add adds a handler.
// A handler with a non-positive weight is ignored.
func (b *Balancer) Add(name string, handler http.Handler, weight *int) {
	w := 1
	if weight != nil {
		w = *weight
	}

	if w <= 0 { // non-positive weight is meaningless
		return
	}

	h := &namedHandler{Handler: handler, name: name, weight: float64(w)}

	b.handlersMu.Lock()
	b.strategy.add(h)
	b.status[name] = struct{}{}
	b.handlerMap[name] = h
	b.handlerMap[hash(name)] = h
	b.handlersMu.Unlock()
}

func hash(input string) string {
	hasher := fnv.New64()
	// We purposely ignore the error because the implementation always returns nil.
	_, _ = hasher.Write([]byte(input))

	return strconv.FormatUint(hasher.Sum64(), 16)
}
