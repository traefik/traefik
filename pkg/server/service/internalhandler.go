package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
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
func NewInternalHandlers(api func(configuration *runtime.Configuration) http.Handler, configuration *runtime.Configuration, rest http.Handler, metricsHandler http.Handler, pingHandler http.Handler, dashboard http.Handler, next serviceManager) *InternalHandlers {
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

// internalWithModifier wraps an internal handler together with a response modifier.
// Its goal is to apply the modifier on the response that would be served by the internal handler,
// but before the response is actually written to the original response writer.
// It is safe to call ServeHTTP on an internalWithModifier with a nil modifier.
type internalWithModifier struct {
	internal http.Handler
	modifier func(*http.Response) error
}

func (imh internalWithModifier) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if imh.modifier == nil {
		imh.internal.ServeHTTP(rw, req)
		return
	}

	rec := httptest.NewRecorder()
	imh.internal.ServeHTTP(rec, req)

	resp := rec.Result()
	resp.Request = req

	if err := imh.modifier(resp); err != nil {
		log.FromContext(req.Context()).Error(err)
		http.Error(rw, "error while applying response modifier", http.StatusInternalServerError)
		return
	}

	if err := resp.Write(rw); err != nil {
		log.FromContext(req.Context()).Error(err)
		return
	}
}

// BuildHTTP builds an HTTP handler.
func (m *InternalHandlers) BuildHTTP(rootCtx context.Context, serviceName string, responseModifier func(*http.Response) error) (http.Handler, error) {
	if !strings.HasSuffix(serviceName, "@internal") {
		return m.serviceManager.BuildHTTP(rootCtx, serviceName, responseModifier)
	}

	internalHandler, err := m.get(serviceName)
	if err != nil {
		return nil, err
	}

	return internalWithModifier{
		internal: internalHandler,
		modifier: responseModifier,
	}, nil
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
