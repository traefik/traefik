package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/containous/traefik/v2/pkg/config/runtime"
	"github.com/containous/traefik/v2/pkg/log"
)

type serviceManager interface {
	BuildHTTP(rootCtx context.Context, serviceName string, responseModifier func(*http.Response) error) (http.Handler, error)
	LaunchHealthCheck()
}

// InternalHandlers is the internal HTTP handlers builder.
type InternalHandlers struct {
	api        http.Handler
	dashboard  http.Handler
	rest       http.Handler
	prometheus http.Handler
	ping       http.Handler
	serviceManager
}

// NewInternalHandlers creates a new InternalHandlers.
func NewInternalHandlers(api func(configuration *runtime.Configuration) http.Handler, configuration *runtime.Configuration, rest, metricsHandler, pingHandler, dashboard http.Handler, next serviceManager) *InternalHandlers {
	var apiHandler http.Handler
	if api != nil {
		apiHandler = api(configuration)
	}

	return &InternalHandlers{
		api:            apiHandler,
		dashboard:      dashboard,
		rest:           rest,
		prometheus:     metricsHandler,
		ping:           pingHandler,
		serviceManager: next,
	}
}

type responseModifier struct {
	r *http.Request
	w http.ResponseWriter

	headersSent bool // whether headers have already been sent
	code        int  // status code, must default to 200

	modifier    func(*http.Response) error // can be nil
	modified    bool                       // whether modifier has already been called for the current request
	modifierErr error                      // returned by modifier call
}

// modifier can be nil.
func newResponseModifier(w http.ResponseWriter, r *http.Request, modifier func(*http.Response) error) *responseModifier {
	return &responseModifier{
		r:        r,
		w:        w,
		modifier: modifier,
		code:     http.StatusOK,
	}
}

func (w *responseModifier) WriteHeader(code int) {
	if w.headersSent {
		return
	}
	defer func() {
		w.code = code
		w.headersSent = true
	}()

	if w.modifier == nil || w.modified {
		w.w.WriteHeader(code)
		return
	}

	resp := http.Response{
		Header:  w.w.Header(),
		Request: w.r,
	}

	if err := w.modifier(&resp); err != nil {
		w.modifierErr = err
		// we are propagating when we are called in Write, but we're logging anyway,
		// because we could be called from another place which does not take care of
		// checking w.modifierErr.
		log.Errorf("Error when applying response modifier: %v", err)
		w.w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.modified = true
	w.w.WriteHeader(code)
}

func (w *responseModifier) Header() http.Header {
	return w.w.Header()
}

func (w *responseModifier) Write(b []byte) (int, error) {
	w.WriteHeader(w.code)
	if w.modifierErr != nil {
		return 0, w.modifierErr
	}

	return w.w.Write(b)
}

// BuildHTTP builds an HTTP handler.
func (m *InternalHandlers) BuildHTTP(rootCtx context.Context, serviceName string, respModifier func(*http.Response) error) (http.Handler, error) {
	if !strings.HasSuffix(serviceName, "@internal") {
		return m.serviceManager.BuildHTTP(rootCtx, serviceName, respModifier)
	}

	internalHandler, err := m.get(serviceName)
	if err != nil {
		return nil, err
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		internalHandler.ServeHTTP(newResponseModifier(w, r, respModifier), r)
	}), nil
}

func (m *InternalHandlers) get(serviceName string) (http.Handler, error) {
	switch serviceName {
	case "noop@internal":
		return http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			rw.WriteHeader(http.StatusTeapot)
		}), nil

	case "api@internal":
		if m.api == nil {
			return nil, errors.New("api is not enabled")
		}
		return m.api, nil

	case "dashboard@internal":
		if m.dashboard == nil {
			return nil, errors.New("dashboard is not enabled")
		}
		return m.dashboard, nil

	case "rest@internal":
		if m.rest == nil {
			return nil, errors.New("rest is not enabled")
		}
		return m.rest, nil

	case "ping@internal":
		if m.ping == nil {
			return nil, errors.New("ping is not enabled")
		}
		return m.ping, nil

	case "prometheus@internal":
		if m.prometheus == nil {
			return nil, errors.New("prometheus is not enabled")
		}
		return m.prometheus, nil

	default:
		return nil, fmt.Errorf("unknown internal service %s", serviceName)
	}
}
