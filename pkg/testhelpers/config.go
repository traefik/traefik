package testhelpers

import (
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
)

// BuildConfiguration is a helper to create a configuration.
func BuildConfiguration(dynamicConfigBuilders ...func(*dynamic.HTTPConfiguration)) *dynamic.HTTPConfiguration {
	conf := &dynamic.HTTPConfiguration{
		Models:            map[string]*dynamic.Model{},
		ServersTransports: map[string]*dynamic.ServersTransport{},
	}

	for _, build := range dynamicConfigBuilders {
		build(conf)
	}
	return conf
}

// WithRouters is a helper to create a configuration.
func WithRouters(opts ...func(*dynamic.Router) string) func(*dynamic.HTTPConfiguration) {
	return func(c *dynamic.HTTPConfiguration) {
		c.Routers = make(map[string]*dynamic.Router)
		for _, opt := range opts {
			b := &dynamic.Router{}
			name := opt(b)
			c.Routers[name] = b
		}
	}
}

// WithRouter is a helper to create a configuration.
func WithRouter(routerName string, opts ...func(*dynamic.Router)) func(*dynamic.Router) string {
	return func(r *dynamic.Router) string {
		for _, opt := range opts {
			opt(r)
		}
		return routerName
	}
}

// WithRouterMiddlewares is a helper to create a configuration.
func WithRouterMiddlewares(middlewaresName ...string) func(*dynamic.Router) {
	return func(r *dynamic.Router) {
		r.Middlewares = middlewaresName
	}
}

// WithServiceName is a helper to create a configuration.
func WithServiceName(serviceName string) func(*dynamic.Router) {
	return func(r *dynamic.Router) {
		r.Service = serviceName
	}
}

// WithLoadBalancerServices is a helper to create a configuration.
func WithLoadBalancerServices(opts ...func(service *dynamic.ServersLoadBalancer) string) func(*dynamic.HTTPConfiguration) {
	return func(c *dynamic.HTTPConfiguration) {
		c.Services = make(map[string]*dynamic.Service)
		for _, opt := range opts {
			b := &dynamic.ServersLoadBalancer{}
			name := opt(b)
			c.Services[name] = &dynamic.Service{
				LoadBalancer: b,
			}
		}
	}
}

// WithService is a helper to create a configuration.
func WithService(name string, opts ...func(*dynamic.ServersLoadBalancer)) func(*dynamic.ServersLoadBalancer) string {
	return func(r *dynamic.ServersLoadBalancer) string {
		for _, opt := range opts {
			opt(r)
		}
		return name
	}
}

// WithMiddlewares is a helper to create a configuration.
func WithMiddlewares(opts ...func(*dynamic.Middleware) string) func(*dynamic.HTTPConfiguration) {
	return func(c *dynamic.HTTPConfiguration) {
		c.Middlewares = make(map[string]*dynamic.Middleware)
		for _, opt := range opts {
			b := &dynamic.Middleware{}
			name := opt(b)
			c.Middlewares[name] = b
		}
	}
}

// WithMiddleware is a helper to create a configuration.
func WithMiddleware(name string, opts ...func(*dynamic.Middleware)) func(*dynamic.Middleware) string {
	return func(r *dynamic.Middleware) string {
		for _, opt := range opts {
			opt(r)
		}
		return name
	}
}

// WithBasicAuth is a helper to create a configuration.
func WithBasicAuth(auth *dynamic.BasicAuth) func(*dynamic.Middleware) {
	return func(r *dynamic.Middleware) {
		r.BasicAuth = auth
	}
}

// WithEntryPoints is a helper to create a configuration.
func WithEntryPoints(eps ...string) func(*dynamic.Router) {
	return func(f *dynamic.Router) {
		f.EntryPoints = eps
	}
}

// WithRule is a helper to create a configuration.
func WithRule(rule string) func(*dynamic.Router) {
	return func(f *dynamic.Router) {
		f.Rule = rule
	}
}

// WithServers is a helper to create a configuration.
func WithServers(opts ...func(*dynamic.Server)) func(*dynamic.ServersLoadBalancer) {
	return func(b *dynamic.ServersLoadBalancer) {
		for _, opt := range opts {
			server := dynamic.Server{}
			opt(&server)
			b.Servers = append(b.Servers, server)
		}
	}
}

// WithServer is a helper to create a configuration.
func WithServer(url string, opts ...func(*dynamic.Server)) func(*dynamic.Server) {
	return func(s *dynamic.Server) {
		for _, opt := range opts {
			opt(s)
		}
		s.URL = url
	}
}

// WithSticky is a helper to create a configuration.
func WithSticky(cookieName string) func(*dynamic.ServersLoadBalancer) {
	return func(b *dynamic.ServersLoadBalancer) {
		b.Sticky = &dynamic.Sticky{
			Cookie: &dynamic.Cookie{Name: cookieName},
		}
	}
}
