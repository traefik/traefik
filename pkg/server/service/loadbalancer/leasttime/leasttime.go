package leasttime

import (
	"context"
	"errors"
	"math"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/server/service/loadbalancer"
)

const (
	sampleSize = 100 // Number of response time samples to track.
)

// namedHandler wraps an HTTP handler with metrics and server information.
// Tracks response time (TTFB) and inflight request count for load balancing decisions.
type namedHandler struct {
	http.Handler
	name   string
	weight float64

	deadline float64 // WRR tie-breaking (EDF scheduling).

	// Metrics (protected by mutex).
	mu              sync.RWMutex
	responseTimes   [sampleSize]float64 // Fixed-size ring buffer (TTFB measurements in ms).
	responseTimeIdx int                 // Current position in ring buffer.
	responseTimeSum float64             // Sum of all values in buffer.
	sampleCount     int                 // Number of samples collected so far.
	inflightCount   atomic.Int64        // Number of inflight requests.
}

// updateResponseTime updates the average response time for this server using a ring buffer.
func (s *namedHandler) updateResponseTime(elapsed time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

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
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.sampleCount == 0 {
		return 0
	}
	return s.responseTimeSum / float64(s.sampleCount)
}

// responseTracker wraps http.ResponseWriter to capture Time To First Byte (TTFB).
type responseTracker struct {
	http.ResponseWriter
	headerTime    *time.Time
	headerWritten bool
}

// WriteHeader intercepts the first byte written to capture TTFB.
func (r *responseTracker) WriteHeader(statusCode int) {
	if !r.headerWritten {
		*r.headerTime = time.Now()
		r.headerWritten = true
	}
	r.ResponseWriter.WriteHeader(statusCode)
}

// Write intercepts writes to capture implicit WriteHeader call.
// This is necessary because if WriteHeader is not explicitly called,
// the first Write will trigger an implicit WriteHeader(200).
func (r *responseTracker) Write(b []byte) (int, error) {
	if !r.headerWritten {
		*r.headerTime = time.Now()
		r.headerWritten = true
	}
	return r.ResponseWriter.Write(b)
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
	updaters []func(bool)

	sticky *loadbalancer.Sticky

	// deadlineMu protects EDF scheduling state (curDeadline and all handler deadline fields).
	// Separate from handlersMu to reduce lock contention during tie-breaking.
	deadlineMu sync.RWMutex
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
		return errors.New("healthCheck not enabled in config for this weighted service")
	}
	b.updaters = append(b.updaters, fn)
	return nil
}

var errNoAvailableServer = errors.New("no available server")

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

	var selected *namedHandler
	minDeadline := math.MaxFloat64

	// Find handler with earliest deadline.
	b.deadlineMu.RLock()
	for _, h := range candidates {
		if h.deadline < minDeadline {
			minDeadline = h.deadline
			selected = h
		}
	}
	b.deadlineMu.RUnlock()

	if selected == nil {
		selected = candidates[0]
	}

	b.deadlineMu.Lock()
	selected.deadline = b.curDeadline + 1/selected.weight
	b.curDeadline = selected.deadline
	b.deadlineMu.Unlock()

	return selected
}

// nextServer selects the next server to handle the request.
// Implements least-time algorithm: selects server with minimum score.
// Score = (avgResponseTime × (1 + inflightCount)) / weight
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

	// Wrap response writer to capture TTFB.
	startTime := time.Now()
	headerTime := startTime // Track when header is written.
	tracked := &responseTracker{
		ResponseWriter: rw,
		headerTime:     &headerTime,
	}

	// Serve request.
	server.ServeHTTP(tracked, req)

	// Update average response time (TTFB).
	// If headerTime changed, use it; otherwise use current time.
	elapsed := headerTime.Sub(startTime)
	server.updateResponseTime(elapsed)
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

	b.deadlineMu.Lock()
	h.deadline = b.curDeadline + 1/h.weight
	b.deadlineMu.Unlock()

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
