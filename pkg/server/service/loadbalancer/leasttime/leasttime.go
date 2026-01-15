package leasttime

import (
	"context"
	"errors"
	"math"
	"net/http"
	"net/http/httptrace"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/server/service/loadbalancer"
)

const sampleSize = 100 // Number of response time samples to track.

var errNoAvailableServer = errors.New("no available server")

// namedHandler wraps an HTTP handler with metrics and server information.
// Tracks response time (TTFB) and inflight request count for load balancing decisions.
type namedHandler struct {
	http.Handler
	name   string
	weight float64

	deadlineMu sync.RWMutex
	deadline   float64 // WRR tie-breaking (EDF scheduling).

	inflightCount atomic.Int64 // Number of inflight requests.

	responseTimeMu  sync.RWMutex
	responseTimes   [sampleSize]float64 // Fixed-size ring buffer (TTFB measurements in ms).
	responseTimeIdx int                 // Current position in ring buffer.
	responseTimeSum float64             // Sum of all values in buffer.
	sampleCount     int                 // Number of samples collected so far.
}

// updateResponseTime updates the average response time for this server using a ring buffer.
func (s *namedHandler) updateResponseTime(elapsed time.Duration) {
	s.responseTimeMu.Lock()
	defer s.responseTimeMu.Unlock()

	ms := float64(elapsed.Milliseconds())

	if s.sampleCount < sampleSize {
		// Still filling the buffer.
		s.responseTimes[s.responseTimeIdx] = ms
		s.responseTimeSum += ms
		s.sampleCount++
	} else {
		// Buffer is full, replace oldest value.
		oldValue := s.responseTimes[s.responseTimeIdx]
		s.responseTimes[s.responseTimeIdx] = ms
		s.responseTimeSum = s.responseTimeSum - oldValue + ms
	}

	s.responseTimeIdx = (s.responseTimeIdx + 1) % sampleSize
}

// getAvgResponseTime returns the average response time in milliseconds.
// Returns 0 if no samples have been collected yet (cold start).
func (s *namedHandler) getAvgResponseTime() float64 {
	s.responseTimeMu.RLock()
	defer s.responseTimeMu.RUnlock()

	if s.sampleCount == 0 {
		return 0
	}
	return s.responseTimeSum / float64(s.sampleCount)
}

func (s *namedHandler) getDeadline() float64 {
	s.deadlineMu.RLock()
	defer s.deadlineMu.RUnlock()
	return s.deadline
}

func (s *namedHandler) setDeadline(deadline float64) {
	s.deadlineMu.Lock()
	defer s.deadlineMu.Unlock()
	s.deadline = deadline
}

// Balancer implements the least-time load balancing algorithm.
// It selects the server with the lowest average response time (TTFB) and fewest active connections.
type Balancer struct {
	wantsHealthCheck bool

	// handlersMu protects the handlers slice, the status and the fenced maps.
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
	// No mutex is needed, as it is modified only during the configuration build.
	updaters []func(bool)

	sticky *loadbalancer.Sticky

	// deadlineMu protects EDF scheduling state (curDeadline and all handler deadline fields).
	// Separate from handlersMu to reduce lock contention during tie-breaking.
	curDeadlineMu sync.RWMutex
	// curDeadline is used for WRR tie-breaking (EDF scheduling).
	curDeadline float64
}

// New creates a new least-time load balancer.
func New(stickyConfig *dynamic.Sticky, wantsHealthCheck bool) *Balancer {
	balancer := &Balancer{
		status:           make(map[string]struct{}),
		fenced:           make(map[string]struct{}),
		wantsHealthCheck: wantsHealthCheck,
	}
	if stickyConfig != nil && stickyConfig.Cookie != nil {
		balancer.sticky = loadbalancer.NewSticky(*stickyConfig.Cookie)
	}

	return balancer
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
	}

	upAfter := len(b.status) > 0
	status = "DOWN"
	if upAfter {
		status = "UP"
	}

	// No Status Change.
	if upBefore == upAfter {
		// We're still with the same status, no need to propagate.
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
		return errors.New("healthCheck not enabled in config for this LeastTime service")
	}
	b.updaters = append(b.updaters, fn)
	return nil
}

// getHealthyServers returns the list of healthy, non-fenced servers.
func (b *Balancer) getHealthyServers() []*namedHandler {
	b.handlersMu.RLock()
	defer b.handlersMu.RUnlock()

	var healthy []*namedHandler
	for _, h := range b.handlers {
		if _, ok := b.status[h.name]; ok {
			if _, fenced := b.fenced[h.name]; !fenced {
				healthy = append(healthy, h)
			}
		}
	}
	return healthy
}

// selectWRR selects a server from candidates using Weighted Round Robin (EDF scheduling).
// This is used for tie-breaking when multiple servers have identical scores.
func (b *Balancer) selectWRR(candidates []*namedHandler) *namedHandler {
	if len(candidates) == 0 {
		return nil
	}

	selected := candidates[0]
	minDeadline := math.MaxFloat64

	// Find handler with earliest deadline.
	for _, h := range candidates {
		handlerDeadline := h.getDeadline()
		if handlerDeadline < minDeadline {
			minDeadline = handlerDeadline
			selected = h
		}
	}

	// Update deadline based on when this server was selected (minDeadline),
	// not the global curDeadline. This ensures proper weighted distribution.
	newDeadline := minDeadline + 1/selected.weight
	selected.setDeadline(newDeadline)

	// Track the maximum deadline assigned for initializing new servers.
	b.curDeadlineMu.Lock()
	if newDeadline > b.curDeadline {
		b.curDeadline = newDeadline
	}
	b.curDeadlineMu.Unlock()

	return selected
}

// Score = (avgResponseTime × (1 + inflightCount)) / weight.
func (b *Balancer) nextServer() (*namedHandler, error) {
	healthy := b.getHealthyServers()

	if len(healthy) == 0 {
		return nil, errNoAvailableServer
	}

	if len(healthy) == 1 {
		return healthy[0], nil
	}

	// Calculate scores and find minimum.
	minScore := math.MaxFloat64
	var candidates []*namedHandler

	for _, h := range healthy {
		avgRT := h.getAvgResponseTime()
		inflight := float64(h.inflightCount.Load())
		score := (avgRT * (1 + inflight)) / h.weight

		if score < minScore {
			minScore = score
			candidates = []*namedHandler{h}
		} else if score == minScore {
			candidates = append(candidates, h)
		}
	}

	if len(candidates) == 1 {
		return candidates[0], nil
	}

	// Multiple servers with same score: use WRR (EDF) tie-breaking.
	selected := b.selectWRR(candidates)
	if selected == nil {
		return nil, errNoAvailableServer
	}

	return selected, nil
}

func (b *Balancer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Handle sticky sessions first.
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

	// Track inflight requests.
	server.inflightCount.Add(1)
	defer server.inflightCount.Add(-1)

	startTime := time.Now()
	trace := &httptrace.ClientTrace{
		GotFirstResponseByte: func() {
			// Update average response time (TTFB).
			server.updateResponseTime(time.Since(startTime))
		},
	}
	traceCtx := httptrace.WithClientTrace(req.Context(), trace)
	server.ServeHTTP(rw, req.WithContext(traceCtx))
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

	if w <= 0 { // non-positive weight is meaningless.
		return
	}

	h := &namedHandler{Handler: handler, name: name, weight: float64(w)}

	// Initialize deadline by adding 1/weight to current deadline.
	// This staggers servers to prevent all starting at the same time.
	var deadline float64
	b.curDeadlineMu.RLock()
	deadline = b.curDeadline + 1/h.weight
	b.curDeadlineMu.RUnlock()

	h.setDeadline(deadline)

	// Update balancer's current deadline with the new server's deadline.
	b.curDeadlineMu.Lock()
	b.curDeadline = deadline
	b.curDeadlineMu.Unlock()

	b.handlersMu.Lock()
	b.handlers = append(b.handlers, h)
	b.status[name] = struct{}{}
	if fenced {
		b.fenced[name] = struct{}{}
	}
	b.handlersMu.Unlock()

	if b.sticky != nil {
		b.sticky.AddHandler(name, handler)
	}
}
