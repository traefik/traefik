package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/containous/traefik/v2/pkg/config/runtime"
)

type serviceManager interface {
	BuildHTTP(rootCtx context.Context, serviceName string, responseModifier func(*http.Response) error) (http.Handler, error)
	LaunchHealthCheck()
}

// InternalHandlers is the internal HTTP handlers builder.
type InternalHandlers struct {
	api  http.Handler
	rest http.Handler
	serviceManager
}

// NewInternalHandlers creates a new InternalHandlers.
func NewInternalHandlers(api func(configuration *runtime.Configuration) http.Handler, configuration *runtime.Configuration, rest http.Handler, next serviceManager) *InternalHandlers {
	var apiHandler http.Handler
	if api != nil {
		apiHandler = api(configuration)
	}

	return &InternalHandlers{
		api:            apiHandler,
		rest:           rest,
		serviceManager: next,
	}
}

// BuildHTTP builds an HTTP handler.
func (m *InternalHandlers) BuildHTTP(rootCtx context.Context, serviceName string, responseModifier func(*http.Response) error) (http.Handler, error) {
	if strings.HasSuffix(serviceName, "@internal") {
		return m.get(serviceName)
	}

	return m.serviceManager.BuildHTTP(rootCtx, serviceName, responseModifier)
}

func (m *InternalHandlers) get(serviceName string) (http.Handler, error) {
	if serviceName == "api@internal" {
		if m.api == nil {
			return nil, errors.New("api is not enabled")
		}
		return m.api, nil
	}

	if serviceName == "rest@internal" {
		if m.rest == nil {
			return nil, errors.New("rest is not enabled")
		}
		return m.rest, nil
	}

	return nil, fmt.Errorf("unknown internal service %s", serviceName)
}
