package ewma

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/server/service/loadbalancer"
)

const (
	// decayConstant is the EWMA decay time constant (in milliseconds).
	// A higher τ makes the average respond more slowly; 10s means past RTTs
	// decay with half-life ≈6.93s.
	decayConstant = 10 * time.Second

	// ewmaStateTTL defines how long an unused EWMA state lives before pruning.
	ewmaStateTTL = 1 * time.Minute

	// initialLatency is the starting EWMA value for new servers.
	// Zero makes brand‑new servers appear “fastest” until measured.
	initialLatency = 0 * time.Second
)

var (
	// ewmaRegistry stores EWMA states for servers, keyed by server name.
	ewmaRegistry = map[string]*serverState{}
	registryMu   sync.RWMutex
)

// serverState holds the EWMA metrics globally between handler recreations.
type serverState struct {
	// ewmaValue stores the EWMA value as a float64 encoded in bits.
	ewmaValue atomic.Uint64

	// lastUpdated stores the last access time in nanoseconds.
	lastUpdated atomic.Uint64
}

// namedHandler wraps an HTTP handler with its name and shared EWMA state.
type namedHandler struct {
	handler http.Handler
	name    string
	state   *serverState
}

type rnd interface {
	Intn(n int) int
}

// Balancer implements a Peak-EWMA P2C load balancer.
type Balancer struct {
	// decayEwma is the EWMA decay time constant in seconds.
	decayEwmaSeconds float64

	// initialEwma is the initial EWMA value for new servers in seconds.
	initialEwmaSeconds float64

	// fenced is the list of terminating yet still serving child services.
	fenced map[string]struct{}

	handlers   []*namedHandler
	handlersMu sync.RWMutex

	healthyHandlers []*namedHandler
	healthyMu       sync.Mutex

	// rand is the random number generator for P2C server selection.
	rand   rnd
	randMu sync.Mutex

	// status is a record of which child services of the Balancer are healthy, keyed
	// by name of child service. A service is initially added to the map when it is
	// created via Add, and it is later removed or added to the map as needed,
	// through the SetStatus method.
	status   map[string]struct{}
	statusMu sync.RWMutex

	// sticky enables sticky session support, if configured.
	sticky *loadbalancer.Sticky

	// updaters is the list of hooks that are run (to update the Balancer
	// parent(s)), whenever the Balancer status changes.
	updaters         []func(bool)
	wantsHealthCheck bool
}

// New creates a Peak-EWMA balancer.
func New(stickyCfg *dynamic.Sticky, wantsHealthCheck bool) *Balancer {
	b := &Balancer{
		status:             make(map[string]struct{}),
		fenced:             make(map[string]struct{}),
		rand:               rand.New(rand.NewSource(time.Now().UnixNano())),
		decayEwmaSeconds:   decayConstant.Seconds(),
		initialEwmaSeconds: initialLatency.Seconds(),
		wantsHealthCheck:   wantsHealthCheck,
	}
	if stickyCfg != nil && stickyCfg.Cookie != nil {
		b.sticky = loadbalancer.NewSticky(*stickyCfg.Cookie)
	}
	return b
}

// AddServer registers a new backend server with its HTTP handler and configuration,
// reusing its EWMA state if it exists and is fresh.
func (b *Balancer) AddServer(name string, handler http.Handler, server dynamic.Server) {
	b.handlersMu.RLock()
	for _, h := range b.handlers {
		if h.name == name {
			b.handlersMu.RUnlock()
			return
		}
	}
	existing := make([]string, len(b.handlers))
	for i, h := range b.handlers {
		existing[i] = h.name
	}
	b.handlersMu.RUnlock()

	names := append(existing, name)

	b.syncHandlers(names, name, handler)

	b.statusMu.Lock()
	b.status[name] = struct{}{}
	if server.Fenced {
		b.fenced[name] = struct{}{}
	}
	b.statusMu.Unlock()

	if b.sticky != nil {
		b.sticky.AddHandler(name, handler)
	}
	b.updateHealthyHandlers()
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

// SetStatus sets on the balancer that its given child is now of the given status.
func (b *Balancer) SetStatus(ctx context.Context, name string, up bool) {
	status := "DOWN"
	b.statusMu.Lock()
	if up {
		status = "UP"
		b.status[name] = struct{}{}
	} else {
		delete(b.status, name)
	}
	b.statusMu.Unlock()
	b.updateHealthyHandlers()
	log.Ctx(ctx).Debug().Msgf("Setting status of %s to %v", name, status)
}

// ServeHTTP implements the http.Handler interface, balancing the request
// across backend servers using Peak-EWMA P2C and updating EWMA metrics.
func (b *Balancer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if b.tryStickyHandler(rw, req) {
		return
	}

	b.handleNewRequest(rw, req)
}

// buildHandlers constructs a slice of namedHandler for the given names,
// reusing or initializing EWMA states and selecting the appropriate HTTP handler.
func (b *Balancer) buildHandlers(names []string, newName string, newHandler http.Handler) []*namedHandler {
	initEwma := b.computeAverageEWMA(names)

	b.handlersMu.RLock()
	existing := make(map[string]http.Handler, len(b.handlers))
	for _, h := range b.handlers {
		existing[h.name] = h.handler
	}
	b.handlersMu.RUnlock()

	registryMu.Lock()
	defer registryMu.Unlock()

	handlers := make([]*namedHandler, 0, len(names))
	for _, name := range names {
		st, ok := ewmaRegistry[name]
		if !ok {
			st = &serverState{}
			startupPenalty := 0.1
			st.ewmaValue.Store(math.Float64bits(initEwma + startupPenalty))
			st.lastUpdated.Store(uint64(time.Now().UnixNano()))
			ewmaRegistry[name] = st
		}

		realHandler := newHandler
		if name != newName {
			realHandler = http.NotFoundHandler()
			if h, ok := existing[name]; ok {
				realHandler = h
			}
		}

		handlers = append(handlers, &namedHandler{
			name:    name,
			state:   st,
			handler: realHandler,
		})
	}
	return handlers
}

// cleanupRegistry removes EWMA states older than the configured TTL.
func (b *Balancer) cleanupRegistry() {
	cutoff := time.Now().UnixNano() - ewmaStateTTL.Nanoseconds()
	registryMu.Lock()
	defer registryMu.Unlock()

	for name, st := range ewmaRegistry {
		if int64(st.lastUpdated.Load()) < cutoff {
			delete(ewmaRegistry, name)
		}
	}
}

// computeAverageEWMA computes the average EWMA across existing states for the given names.
func (b *Balancer) computeAverageEWMA(names []string) float64 {
	registryMu.RLock()
	defer registryMu.RUnlock()

	var sum float64
	var count int
	for _, name := range names {
		if st, ok := ewmaRegistry[name]; ok {
			sum += math.Float64frombits(st.ewmaValue.Load())
			count++
		}
	}
	if count > 0 {
		return sum / float64(count)
	}
	return b.initialEwmaSeconds
}

// handleNewRequest selects a backend, serves the request, and updates its EWMA.
func (b *Balancer) handleNewRequest(rw http.ResponseWriter, req *http.Request) {
	server, err := b.nextServer()
	if err != nil {
		code := http.StatusInternalServerError
		if errors.Is(err, errNoAvailableServer) {
			code = http.StatusServiceUnavailable
		}
		http.Error(rw, err.Error(), code)
		return
	}

	if b.sticky != nil {
		if err := b.sticky.WriteStickyCookie(rw, server.name); err != nil {
			log.Error().Err(err).Msg("writing sticky cookie")
		}
	}
	start := time.Now()
	server.handler.ServeHTTP(rw, req)
	b.updateEWMA(server, time.Since(start))
}

// IsServerHealthy checks if the server with the given name is healthy and not fenced.
func (b *Balancer) isServerHealthy(name string) bool {
	b.statusMu.RLock()
	_, ok := b.status[name]
	_, fenced := b.fenced[name]
	b.statusMu.RUnlock()
	return ok && !fenced
}

var errNoAvailableServer = errors.New("no available server")

func (b *Balancer) nextServer() (*namedHandler, error) {
	b.healthyMu.Lock()

	n := len(b.healthyHandlers)
	if n == 0 {
		b.healthyMu.Unlock()
		return nil, errNoAvailableServer
	}
	if n == 1 {
		b.healthyMu.Unlock()
		return b.healthyHandlers[0], nil
	}

	pickSetSize := 2
	b.randMu.Lock()
	for i := 0; i < pickSetSize; i++ {
		j := i + b.rand.Intn(n-i)
		b.healthyHandlers[i], b.healthyHandlers[j] = b.healthyHandlers[j], b.healthyHandlers[i]
	}
	srv1, srv2 := b.healthyHandlers[0], b.healthyHandlers[1]

	b.healthyMu.Unlock()
	b.randMu.Unlock()

	best, bestScore := srv1, math.MaxFloat64
	now := time.Now().UnixNano()
	for _, srv := range []*namedHandler{srv1, srv2} {
		ewma := math.Float64frombits(srv.state.ewmaValue.Load())
		lt := int64(srv.state.lastUpdated.Load())
		dt := time.Duration(now - lt).Seconds()
		score := computeEWMA(ewma, 0, dt, b.decayEwmaSeconds)
		if score < bestScore {
			bestScore, best = score, srv
		}
	}

	return best, nil
}

// syncHandlers rebuilds balancer handlers, preserving EWMA states.
func (b *Balancer) syncHandlers(names []string, newName string, newHandler http.Handler) {
	b.cleanupRegistry()
	handlers := b.buildHandlers(names, newName, newHandler)

	b.handlersMu.Lock()
	b.handlers = handlers
	b.handlersMu.Unlock()
}

// tryStickyHandler attempts sticky routing and returns true if request handled.
func (b *Balancer) tryStickyHandler(rw http.ResponseWriter, req *http.Request) bool {
	if b.sticky == nil {
		return false
	}

	h, rewrite, err := b.sticky.StickyHandler(req)
	if err != nil {
		log.Error().Err(err).Msg("getting sticky handler")
		return false
	}
	if h == nil {
		return false
	}
	if !b.isServerHealthy(h.Name) {
		return false
	}

	if rewrite {
		if err := b.sticky.WriteStickyCookie(rw, h.Name); err != nil {
			log.Error().Err(err).Msg("writing sticky cookie")
		}
	}

	h.ServeHTTP(rw, req)
	return true
}

// updateEWMA updates the EWMA state for the given server based on the observed request latency.
func (b *Balancer) updateEWMA(nh *namedHandler, latency time.Duration) {
	now := time.Now().UnixNano()
	old := math.Float64frombits(nh.state.ewmaValue.Load())
	lt := nh.state.lastUpdated.Load()
	newEwma := old

	dt := time.Duration(now - int64(lt)).Seconds()
	newEwma = computeEWMA(old, latency.Seconds(), dt, b.decayEwmaSeconds)

	nh.state.ewmaValue.Store(math.Float64bits(newEwma))
	nh.state.lastUpdated.Store(uint64(now))
}

// updateHealthyHandlers refreshes the list of currently healthy handlers.
func (b *Balancer) updateHealthyHandlers() {
	b.handlersMu.RLock()
	handlers := b.handlers
	b.handlersMu.RUnlock()

	healthy := make([]*namedHandler, 0)
	for _, h := range handlers {
		if b.isServerHealthy(h.name) {
			healthy = append(healthy, h)
		}
	}

	b.healthyMu.Lock()
	b.healthyHandlers = healthy
	b.healthyMu.Unlock()
}

// computeEWMA computes the EWMA value using the formula: old * exp(-dt/τ) + rtt * (1 - exp(-dt/τ)).
func computeEWMA(prevEWMA, latency, deltaTime, decayConstant float64) float64 {
	weight := math.Exp(-deltaTime / decayConstant)
	return prevEWMA*weight + latency*(1-weight)
}
