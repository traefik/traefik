package testhelpers

import (
	"github.com/containous/traefik/config"
)

// BuildConfiguration is a helper to create a configuration.
func BuildConfiguration(dynamicConfigBuilders ...func(*config.HTTPConfiguration)) *config.HTTPConfiguration {
	conf := &config.HTTPConfiguration{}
	for _, build := range dynamicConfigBuilders {
		build(conf)
	}
	return conf
}

// WithRouters is a helper to create a configuration.
func WithRouters(opts ...func(*config.Router) string) func(*config.HTTPConfiguration) {
	return func(c *config.HTTPConfiguration) {
		c.Routers = make(map[string]*config.Router)
		for _, opt := range opts {
			b := &config.Router{}
			name := opt(b)
			c.Routers[name] = b
		}
	}
}

// WithRouter is a helper to create a configuration.
func WithRouter(routerName string, opts ...func(*config.Router)) func(*config.Router) string {
	return func(r *config.Router) string {
		for _, opt := range opts {
			opt(r)
		}
		return routerName
	}
}

// WithRouterMiddlewares is a helper to create a configuration.
func WithRouterMiddlewares(middlewaresName ...string) func(*config.Router) {
	return func(r *config.Router) {
		r.Middlewares = middlewaresName
	}
}

// WithServiceName is a helper to create a configuration.
func WithServiceName(serviceName string) func(*config.Router) {
	return func(r *config.Router) {
		r.Service = serviceName
	}
}

// WithLoadBalancerServices is a helper to create a configuration.
func WithLoadBalancerServices(opts ...func(service *config.LoadBalancerService) string) func(*config.HTTPConfiguration) {
	return func(c *config.HTTPConfiguration) {
		c.Services = make(map[string]*config.Service)
		for _, opt := range opts {
			b := &config.LoadBalancerService{}
			name := opt(b)
			c.Services[name] = &config.Service{
				LoadBalancer: b,
			}
		}
	}
}

// WithService is a helper to create a configuration.
func WithService(name string, opts ...func(*config.LoadBalancerService)) func(*config.LoadBalancerService) string {
	return func(r *config.LoadBalancerService) string {
		for _, opt := range opts {
			opt(r)
		}
		return name
	}
}

// WithMiddlewares is a helper to create a configuration.
func WithMiddlewares(opts ...func(*config.Middleware) string) func(*config.HTTPConfiguration) {
	return func(c *config.HTTPConfiguration) {
		c.Middlewares = make(map[string]*config.Middleware)
		for _, opt := range opts {
			b := &config.Middleware{}
			name := opt(b)
			c.Middlewares[name] = b
		}
	}
}

// WithMiddleware is a helper to create a configuration.
func WithMiddleware(name string, opts ...func(*config.Middleware)) func(*config.Middleware) string {
	return func(r *config.Middleware) string {
		for _, opt := range opts {
			opt(r)
		}
		return name
	}
}

// WithBasicAuth is a helper to create a configuration.
func WithBasicAuth(auth *config.BasicAuth) func(*config.Middleware) {
	return func(r *config.Middleware) {
		r.BasicAuth = auth
	}
}

// WithEntryPoints is a helper to create a configuration.
func WithEntryPoints(eps ...string) func(*config.Router) {
	return func(f *config.Router) {
		f.EntryPoints = eps
	}
}

// WithRule is a helper to create a configuration.
func WithRule(rule string) func(*config.Router) {
	return func(f *config.Router) {
		f.Rule = rule
	}
}

// WithServers is a helper to create a configuration.
func WithServers(opts ...func(*config.Server)) func(*config.LoadBalancerService) {
	return func(b *config.LoadBalancerService) {
		for _, opt := range opts {
			server := config.Server{Weight: 1}
			opt(&server)
			b.Servers = append(b.Servers, server)
		}
	}
}

// WithServer is a helper to create a configuration.
func WithServer(url string, opts ...func(*config.Server)) func(*config.Server) {
	return func(s *config.Server) {
		for _, opt := range opts {
			opt(s)
		}
		s.URL = url
	}
}

// WithLBMethod is a helper to create a configuration.
func WithLBMethod(method string) func(*config.LoadBalancerService) {
	return func(b *config.LoadBalancerService) {
		b.Method = method
	}
}

// WithStickiness is a helper to create a configuration.
func WithStickiness(cookieName string) func(*config.LoadBalancerService) {
	return func(b *config.LoadBalancerService) {
		b.Stickiness = &config.Stickiness{
			CookieName: cookieName,
		}
	}
}
