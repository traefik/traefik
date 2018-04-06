package testhelpers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/containous/traefik/types"
)

// Intp returns a pointer to the given integer value.
func Intp(i int) *int {
	return &i
}

// Stringp returns a pointer to the given string value.
func Stringp(s string) *string {
	return &s
}

// MustNewRequest creates a new http get request or panics if it can't
func MustNewRequest(method, urlStr string, body io.Reader) *http.Request {
	request, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		panic(fmt.Sprintf("failed to create HTTP %s Request for '%s': %s", method, urlStr, err))
	}
	return request
}

// MustParseURL parses a URL or panics if it can't
func MustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(fmt.Sprintf("failed to parse URL '%s': %s", rawURL, err))
	}
	return u
}

// BuildDynamicConfig is a helper to create a configuration with the builder pattern.
func BuildDynamicConfig(dynamicConfigBuilders ...func(*types.Configuration)) *types.Configuration {
	config := &types.Configuration{
		Frontends: make(map[string]*types.Frontend),
		Backends:  make(map[string]*types.Backend),
	}
	for _, build := range dynamicConfigBuilders {
		build(config)
	}
	return config
}

// WithFrontend builds a function that adds the passed frontend to a configuration.
func WithFrontend(frontendName string, frontend *types.Frontend) func(*types.Configuration) {
	return func(config *types.Configuration) {
		config.Frontends[frontendName] = frontend
	}
}

// WithBackend builds a function that adds the passed backend to a configuration.
func WithBackend(backendName string, backend *types.Backend) func(*types.Configuration) {
	return func(config *types.Configuration) {
		config.Backends[backendName] = backend
	}
}

// BuildFrontend builds a frontend with some default configuration and allows for
// further configuration by passing in frontend builders.
func BuildFrontend(frontendBuilders ...func(*types.Frontend)) *types.Frontend {
	fe := &types.Frontend{
		EntryPoints: []string{"http"},
		Backend:     "backend",
		Routes:      make(map[string]types.Route),
	}
	for _, build := range frontendBuilders {
		build(fe)
	}
	return fe
}

// WithRoute builds a function that adds the passed route to a frontend.
func WithRoute(routeName, rule string) func(*types.Frontend) {
	return func(fe *types.Frontend) {
		fe.Routes[routeName] = types.Route{Rule: rule}
	}
}

// WithEntrypoint builds a function that adds the passed entrypoint to a frontend.
func WithEntrypoint(entrypointName string) func(*types.Frontend) {
	return func(fe *types.Frontend) {
		fe.EntryPoints = append(fe.EntryPoints, entrypointName)
	}
}

// BuildBackend builds a backend with some default configuration and allows for
// further configuration by passing in backend builders.
func BuildBackend(backendBuilders ...func(*types.Backend)) *types.Backend {
	be := &types.Backend{
		Servers:      make(map[string]types.Server),
		LoadBalancer: &types.LoadBalancer{Method: "Wrr"},
	}
	for _, build := range backendBuilders {
		build(be)
	}
	return be
}

// WithServer builds a function that adds the passed server to a backend.
func WithServer(name, url string) func(backend *types.Backend) {
	return func(be *types.Backend) {
		be.Servers[name] = types.Server{URL: url}
	}
}

// WithLoadBalancer builds a function that sets the passed load balancer configuration to a backend.
func WithLoadBalancer(method string, sticky bool) func(*types.Backend) {
	return func(be *types.Backend) {
		if sticky {
			be.LoadBalancer = &types.LoadBalancer{Method: method, Stickiness: &types.Stickiness{CookieName: "test"}}
		} else {
			be.LoadBalancer = &types.LoadBalancer{Method: method}
		}
	}
}
