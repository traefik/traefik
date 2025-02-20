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
	name       string
	hashedName string

	// inflight is the number of inflight requests.
	// It is used to implement the "power-of-two-random-choices" algorithm.
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
	path     string
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

// Balancer is a WeightedRoundRobin load balancer based on Earliest Deadline First (EDF).
// (https://en.wikipedia.org/wiki/Earliest_deadline_first_scheduling)
// Each pick from the schedule has the earliest deadline entry selected.
// Entries have deadlines set at currentDeadline + 1 / weight,
// providing weighted round-robin behavior with floating point weights and an O(log n) pick time.
type Balancer struct {
	stickyCookie     *stickyCookie
	wantsHealthCheck bool

	handlersMu sync.RWMutex
	// References all the handlers by name and also by the hashed value of the name.
	stickyMap              map[string]*namedHandler
	compatibilityStickyMap map[string]*namedHandler
	handlers               []*namedHandler
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

	rand rnd
}

type rnd interface {
	Intn(n int) int
}

// New creates a new "the power-of-two-random-choices" load balancer.
// strategyPowerOfTwoChoices implements "the power-of-two-random-choices" algorithm for load balancing.
// The idea of this is two take two of the backends at random from the available backends, and select
// the backend that has the fewest in-flight requests. This algorithm more effectively balances the
// load than a round-robin approach, while also being constant time when picking: The strategy also
// has more beneficial "herd" behavior than the "fewest connections" algorithm, especially when the
// load balancer doesn't have perfect knowledge about the global number of connections to the backend,
// for example, when running in a distributed fashion.
func New(sticky *dynamic.Sticky, wantHealthCheck bool) *Balancer {
	balancer := &Balancer{
		status:           make(map[string]struct{}),
		fenced:           make(map[string]struct{}),
		wantsHealthCheck: wantHealthCheck,
		rand:             rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	if sticky != nil && sticky.Cookie != nil {
		balancer.stickyCookie = &stickyCookie{
			name:     sticky.Cookie.Name,
			secure:   sticky.Cookie.Secure,
			httpOnly: sticky.Cookie.HTTPOnly,
			sameSite: sticky.Cookie.SameSite,
			maxAge:   sticky.Cookie.MaxAge,
			path:     "/",
		}
		if sticky.Cookie.Path != nil {
			balancer.stickyCookie.path = *sticky.Cookie.Path
		}

		balancer.stickyMap = make(map[string]*namedHandler)
		balancer.compatibilityStickyMap = make(map[string]*namedHandler)
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
	// In order to not get the same backend twice, we make the second call to s.rand.IntN one fewer
	// than the length of the slice. We then have to shift over the second index if it is equal or
	// greater than the first index, wrapping round if needed.
	n1, n2 := b.rand.Intn(len(healthy)), b.rand.Intn(len(healthy))
	if n2 == n1 {
		n2 = (n2 + 1) % len(healthy)
	}

	h1, h2 := healthy[n1], healthy[n2]
	// Ensure h1 has fewer inflight requests than h2.
	if h2.inflight.Load() < h1.inflight.Load() {
		log.Debug().Msgf("Service selected by P2C: %s", h1.name)
		return h2, nil
	}

	log.Debug().Msgf("Service selected by P2C: %s", h1.name)
	return h1, nil
}

func (b *Balancer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if b.stickyCookie != nil {
		cookie, err := req.Cookie(b.stickyCookie.name)

		if err != nil && !errors.Is(err, http.ErrNoCookie) {
			log.Warn().Err(err).Msg("Error while reading cookie")
		}

		if err == nil && cookie != nil {
			b.handlersMu.RLock()
			handler, ok := b.stickyMap[cookie.Value]
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

			b.handlersMu.RLock()
			handler, ok = b.compatibilityStickyMap[cookie.Value]
			b.handlersMu.RUnlock()

			if ok && handler != nil {
				b.handlersMu.RLock()
				_, isHealthy := b.status[handler.name]
				b.handlersMu.RUnlock()
				if isHealthy {
					b.writeStickyCookie(w, handler)

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
		b.writeStickyCookie(w, server)
	}

	server.ServeHTTP(w, req)
}

func (b *Balancer) writeStickyCookie(w http.ResponseWriter, handler *namedHandler) {
	cookie := &http.Cookie{
		Name:     b.stickyCookie.name,
		Value:    handler.hashedName,
		Path:     b.stickyCookie.path,
		HttpOnly: b.stickyCookie.httpOnly,
		Secure:   b.stickyCookie.secure,
		SameSite: convertSameSite(b.stickyCookie.sameSite),
		MaxAge:   b.stickyCookie.maxAge,
	}
	http.SetCookie(w, cookie)
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

	if b.stickyCookie != nil {
		sha256HashedName := loadbalancer.Sha256Hash(name)
		h.hashedName = sha256HashedName

		b.stickyMap[sha256HashedName] = h
		b.compatibilityStickyMap[name] = h

		hashedName := loadbalancer.FnvHash(name)
		b.compatibilityStickyMap[hashedName] = h

		// server.URL was fnv hashed in service.Manager
		// so we can have "double" fnv hash in already existing cookies
		hashedName = loadbalancer.FnvHash(hashedName)
		b.compatibilityStickyMap[hashedName] = h
	}
	b.handlersMu.Unlock()
}
