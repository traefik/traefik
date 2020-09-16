package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/traefik/traefik/v2/pkg/config/runtime"
)

type serviceManager interface {
	BuildHTTP(rootCtx context.Context, serviceName string) (http.Handler, error)
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

// BuildHTTP builds an HTTP handler.
func (m *InternalHandlers) BuildHTTP(rootCtx context.Context, serviceName string) (http.Handler, error) {
	if !strings.HasSuffix(serviceName, "@internal") {
		return m.serviceManager.BuildHTTP(rootCtx, serviceName)
	}

	internalHandler, err := m.get(serviceName)
	if err != nil {
		return nil, err
	}

	return internalHandler, nil
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
