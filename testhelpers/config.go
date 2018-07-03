package testhelpers

import (
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/types"
)

// BuildConfiguration is a helper to create a configuration.
func BuildConfiguration(dynamicConfigBuilders ...func(*types.Configuration)) *types.Configuration {
	config := &types.Configuration{}
	for _, build := range dynamicConfigBuilders {
		build(config)
	}
	return config
}

// -- Backend

// WithBackends is a helper to create a configuration
func WithBackends(opts ...func(*types.Backend) string) func(*types.Configuration) {
	return func(c *types.Configuration) {
		c.Backends = make(map[string]*types.Backend)
		for _, opt := range opts {
			b := &types.Backend{}
			name := opt(b)
			c.Backends[name] = b
		}
	}
}

// WithBackendNew is a helper to create a configuration
func WithBackendNew(name string, opts ...func(*types.Backend)) func(*types.Backend) string {
	return func(b *types.Backend) string {
		for _, opt := range opts {
			opt(b)
		}
		return name
	}
}

// WithServersNew is a helper to create a configuration
func WithServersNew(opts ...func(*types.Server) string) func(*types.Backend) {
	return func(b *types.Backend) {
		b.Servers = make(map[string]types.Server)
		for _, opt := range opts {
			s := &types.Server{Weight: 1}
			name := opt(s)
			b.Servers[name] = *s
		}
	}
}

// WithServerNew is a helper to create a configuration
func WithServerNew(url string, opts ...func(*types.Server)) func(*types.Server) string {
	return func(s *types.Server) string {
		for _, opt := range opts {
			opt(s)
		}
		s.URL = url
		return provider.Normalize(url)
	}
}

// WithLBMethod is a helper to create a configuration
func WithLBMethod(method string) func(*types.Backend) {
	return func(b *types.Backend) {
		if b.LoadBalancer == nil {
			b.LoadBalancer = &types.LoadBalancer{}
		}
		b.LoadBalancer.Method = method
	}
}

// -- Frontend

// WithFrontends is a helper to create a configuration
func WithFrontends(opts ...func(*types.Frontend) string) func(*types.Configuration) {
	return func(c *types.Configuration) {
		c.Frontends = make(map[string]*types.Frontend)
		for _, opt := range opts {
			f := &types.Frontend{}
			name := opt(f)
			c.Frontends[name] = f
		}
	}
}

// WithFrontend is a helper to create a configuration
func WithFrontend(backend string, opts ...func(*types.Frontend)) func(*types.Frontend) string {
	return func(f *types.Frontend) string {
		for _, opt := range opts {
			opt(f)
		}

		// related the function WithFrontendName
		name := f.Backend
		f.Backend = backend
		if len(name) > 0 {
			return name
		}
		return backend
	}
}

// WithFrontendName is a helper to create a configuration
func WithFrontendName(name string) func(*types.Frontend) {
	return func(f *types.Frontend) {
		// store temporary the frontend name into the backend name
		f.Backend = name
	}
}

// WithEntryPoints is a helper to create a configuration
func WithEntryPoints(eps ...string) func(*types.Frontend) {
	return func(f *types.Frontend) {
		f.EntryPoints = eps
	}
}

// WithRoutes is a helper to create a configuration
func WithRoutes(opts ...func(*types.Route) string) func(*types.Frontend) {
	return func(f *types.Frontend) {
		f.Routes = make(map[string]types.Route)
		for _, opt := range opts {
			s := &types.Route{}
			name := opt(s)
			f.Routes[name] = *s
		}
	}
}

// WithRoute is a helper to create a configuration
func WithRoute(name string, rule string) func(*types.Route) string {
	return func(r *types.Route) string {
		r.Rule = rule
		return name
	}
}

// WithBasicAuth is a helper to create a configuration
// Deprecated
func WithBasicAuth(username string, password string) func(*types.Frontend) {
	return func(fe *types.Frontend) {
		fe.BasicAuth = []string{username + ":" + password}
	}
}

// WithFrontEndAuth is a helper to create a configuration
func WithFrontEndAuth(auth *types.Auth) func(*types.Frontend) {
	return func(fe *types.Frontend) {
		fe.Auth = auth
	}
}

// WithLBSticky is a helper to create a configuration
func WithLBSticky(cookieName string) func(*types.Backend) {
	return func(b *types.Backend) {
		if b.LoadBalancer == nil {
			b.LoadBalancer = &types.LoadBalancer{}
		}
		b.LoadBalancer.Stickiness = &types.Stickiness{CookieName: cookieName}
	}
}
