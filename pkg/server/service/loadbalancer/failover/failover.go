package failover

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
)

// Failover is an http.Handler that can forward requests to the fallback handler
// when the main handler status is down.
type Failover struct {
	wantsHealthCheck bool
	handler          http.Handler
	fallbackHandler  http.Handler
	// updaters is the list of hooks that are run (to update the Failover
	// parent(s)), whenever the Failover status changes.
	updaters []func(bool)

	handlerStatusMu sync.RWMutex
	handlerStatus   bool

	fallbackStatusMu sync.RWMutex
	fallbackStatus   bool
}

// New creates a new Failover handler.
func New(hc *dynamic.HealthCheck) *Failover {
	return &Failover{
		wantsHealthCheck: hc != nil,
	}
}

// RegisterStatusUpdater adds fn to the list of hooks that are run when the
// status of the Failover changes.
// Not thread safe.
func (f *Failover) RegisterStatusUpdater(fn func(up bool)) error {
	if !f.wantsHealthCheck {
		return errors.New("healthCheck not enabled in config for this failover service")
	}

	f.updaters = append(f.updaters, fn)

	return nil
}

func (f *Failover) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	f.handlerStatusMu.RLock()
	handlerStatus := f.handlerStatus
	f.handlerStatusMu.RUnlock()

	if handlerStatus {
		f.handler.ServeHTTP(w, req)
		return
	}

	f.fallbackStatusMu.RLock()
	fallbackStatus := f.fallbackStatus
	f.fallbackStatusMu.RUnlock()

	if fallbackStatus {
		f.fallbackHandler.ServeHTTP(w, req)
		return
	}

	http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
}

// SetHandler sets the main http.Handler.
func (f *Failover) SetHandler(handler http.Handler) {
	f.handlerStatusMu.Lock()
	defer f.handlerStatusMu.Unlock()

	f.handler = handler
	f.handlerStatus = true
}

// SetHandlerStatus sets the main handler status.
func (f *Failover) SetHandlerStatus(ctx context.Context, up bool) {
	f.handlerStatusMu.Lock()
	defer f.handlerStatusMu.Unlock()

	status := "DOWN"
	if up {
		status = "UP"
	}

	if up == f.handlerStatus {
		// We're still with the same status, no need to propagate.
		log.FromContext(ctx).Debugf("Still %s, no need to propagate", status)
		return
	}

	log.FromContext(ctx).Debugf("Propagating new %s status", status)
	f.handlerStatus = up

	for _, fn := range f.updaters {
		// Failover service status is set to DOWN
		// when main and fallback handlers have a DOWN status.
		fn(f.handlerStatus || f.fallbackStatus)
	}
}

// SetFallbackHandler sets the fallback http.Handler.
func (f *Failover) SetFallbackHandler(handler http.Handler) {
	f.fallbackStatusMu.Lock()
	defer f.fallbackStatusMu.Unlock()

	f.fallbackHandler = handler
	f.fallbackStatus = true
}

// SetFallbackHandlerStatus sets the fallback handler status.
func (f *Failover) SetFallbackHandlerStatus(ctx context.Context, up bool) {
	f.fallbackStatusMu.Lock()
	defer f.fallbackStatusMu.Unlock()

	status := "DOWN"
	if up {
		status = "UP"
	}

	if up == f.fallbackStatus {
		// We're still with the same status, no need to propagate.
		log.FromContext(ctx).Debugf("Still %s, no need to propagate", status)
		return
	}

	log.FromContext(ctx).Debugf("Propagating new %s status", status)
	f.fallbackStatus = up

	for _, fn := range f.updaters {
		// Failover service status is set to DOWN
		// when main and fallback handlers have a DOWN status.
		fn(f.handlerStatus || f.fallbackStatus)
	}
}
