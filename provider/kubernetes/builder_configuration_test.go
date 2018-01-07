package kubernetes

import (
	"testing"

	"github.com/containous/traefik/tls"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
)

func buildConfiguration(opts ...func(*types.Configuration)) *types.Configuration {
	conf := &types.Configuration{}
	for _, opt := range opts {
		opt(conf)
	}
	return conf
}

// Backend

func backends(opts ...func(*types.Backend) string) func(*types.Configuration) {
	return func(c *types.Configuration) {
		c.Backends = make(map[string]*types.Backend)
		for _, opt := range opts {
			b := &types.Backend{}
			name := opt(b)
			c.Backends[name] = b
		}
	}
}

func backend(name string, opts ...func(*types.Backend)) func(*types.Backend) string {
	return func(b *types.Backend) string {
		for _, opt := range opts {
			opt(b)
		}
		return name
	}
}

func servers(opts ...func(*types.Server) string) func(*types.Backend) {
	return func(b *types.Backend) {
		b.Servers = make(map[string]types.Server)
		for _, opt := range opts {
			s := &types.Server{}
			name := opt(s)
			b.Servers[name] = *s
		}
	}
}

func server(url string, opts ...func(*types.Server)) func(*types.Server) string {
	return func(s *types.Server) string {
		for _, opt := range opts {
			opt(s)
		}
		s.URL = url
		return url
	}
}

func weight(value int) func(*types.Server) {
	return func(s *types.Server) {
		s.Weight = value
	}
}

func lbMethod(method string) func(*types.Backend) {
	return func(b *types.Backend) {
		if b.LoadBalancer == nil {
			b.LoadBalancer = &types.LoadBalancer{}
		}
		b.LoadBalancer.Method = method
	}
}

func lbSticky() func(*types.Backend) {
	return func(b *types.Backend) {
		if b.LoadBalancer == nil {
			b.LoadBalancer = &types.LoadBalancer{}
		}
		b.LoadBalancer.Sticky = true
	}
}

func circuitBreaker(exp string) func(*types.Backend) {
	return func(b *types.Backend) {
		b.CircuitBreaker = &types.CircuitBreaker{}
		b.CircuitBreaker.Expression = exp
	}
}

// Frontend

func buildFrontends(opts ...func(*types.Frontend) string) map[string]*types.Frontend {
	fronts := make(map[string]*types.Frontend)
	for _, opt := range opts {
		f := &types.Frontend{}
		name := opt(f)
		fronts[name] = f
	}
	return fronts
}

func frontends(opts ...func(*types.Frontend) string) func(*types.Configuration) {
	return func(c *types.Configuration) {
		c.Frontends = make(map[string]*types.Frontend)
		for _, opt := range opts {
			f := &types.Frontend{}
			name := opt(f)
			c.Frontends[name] = f
		}
	}
}

func frontend(backend string, opts ...func(*types.Frontend)) func(*types.Frontend) string {
	return func(f *types.Frontend) string {
		for _, opt := range opts {
			opt(f)
		}
		f.Backend = backend
		return backend
	}
}

func passHostHeader() func(*types.Frontend) {
	return func(f *types.Frontend) {
		f.PassHostHeader = true
	}
}

func entryPoints(eps ...string) func(*types.Frontend) {
	return func(f *types.Frontend) {
		f.EntryPoints = eps
	}
}

func basicAuth(auth ...string) func(*types.Frontend) {
	return func(f *types.Frontend) {
		f.BasicAuth = auth
	}
}

func whitelistSourceRange(ranges ...string) func(*types.Frontend) {
	return func(f *types.Frontend) {
		f.WhitelistSourceRange = ranges
	}
}

func priority(value int) func(*types.Frontend) {
	return func(f *types.Frontend) {
		f.Priority = value
	}
}

func headers() func(*types.Frontend) {
	return func(f *types.Frontend) {
		f.Headers = &types.Headers{}
	}
}

func redirectEntryPoint(name string) func(*types.Frontend) {
	return func(f *types.Frontend) {
		if f.Redirect == nil {
			f.Redirect = &types.Redirect{}
		}
		f.Redirect.EntryPoint = name
	}
}

func redirectRegex(regex, replacement string) func(*types.Frontend) {
	return func(f *types.Frontend) {
		if f.Redirect == nil {
			f.Redirect = &types.Redirect{}
		}
		f.Redirect.Regex = regex
		f.Redirect.Replacement = replacement
	}
}

func passTLSCert() func(*types.Frontend) {
	return func(f *types.Frontend) {
		f.PassTLSCert = true
	}
}

func routes(opts ...func(*types.Route) string) func(*types.Frontend) {
	return func(f *types.Frontend) {
		f.Routes = make(map[string]types.Route)
		for _, opt := range opts {
			s := &types.Route{}
			name := opt(s)
			f.Routes[name] = *s
		}
	}
}

func route(name string, rule string) func(*types.Route) string {
	return func(r *types.Route) string {
		r.Rule = rule
		return name
	}
}

func tlsConfigurations(opts ...func(*tls.Configuration)) func(*types.Configuration) {
	return func(c *types.Configuration) {
		for _, opt := range opts {
			tlsConf := &tls.Configuration{}
			opt(tlsConf)
			c.TLSConfiguration = append(c.TLSConfiguration, tlsConf)
		}
	}
}

func tlsConfiguration(opts ...func(*tls.Configuration)) func(*tls.Configuration) {
	return func(c *tls.Configuration) {
		for _, opt := range opts {
			opt(c)
		}
	}
}

func tlsEntryPoints(entryPoints ...string) func(*tls.Configuration) {
	return func(c *tls.Configuration) {
		c.EntryPoints = entryPoints
	}
}

func certificate(cert string, key string) func(*tls.Configuration) {
	return func(c *tls.Configuration) {
		c.Certificate = &tls.Certificate{
			CertFile: tls.FileOrContent(cert),
			KeyFile:  tls.FileOrContent(key),
		}
	}
}

// Test

func TestBuildConfiguration(t *testing.T) {
	actual := buildConfiguration(
		backends(
			backend("foo/bar",
				servers(
					server("http://10.10.0.1:8080", weight(1)),
					server("http://10.21.0.1:8080", weight(1)),
				),
				lbMethod("wrr"),
			),
			backend("foo/namedthing",
				servers(server("https://example.com", weight(1))),
				lbMethod("wrr"),
			),
			backend("bar",
				servers(
					server("https://10.15.0.1:8443", weight(1)),
					server("https://10.15.0.2:9443", weight(1)),
				),
				lbMethod("wrr"),
			),
		),
		frontends(
			frontend("foo/bar",
				passHostHeader(),
				routes(
					route("/bar", "PathPrefix:/bar"),
					route("foo", "Host:foo"),
				),
			),
			frontend("foo/namedthing",
				passHostHeader(),
				routes(
					route("/namedthing", "PathPrefix:/namedthing"),
					route("foo", "Host:foo"),
				),
			),
			frontend("bar",
				passHostHeader(),
				routes(
					route("bar", "Host:bar"),
				),
			),
		),
		tlsConfigurations(
			tlsConfiguration(
				tlsEntryPoints("https"),
				certificate("certificate", "key"),
			),
		),
	)

	assert.EqualValues(t, sampleConfiguration(), actual)
}

func sampleConfiguration() *types.Configuration {
	return &types.Configuration{
		Backends: map[string]*types.Backend{
			"foo/bar": {
				Servers: map[string]types.Server{
					"http://10.10.0.1:8080": {
						URL:    "http://10.10.0.1:8080",
						Weight: 1,
					},
					"http://10.21.0.1:8080": {
						URL:    "http://10.21.0.1:8080",
						Weight: 1,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer: &types.LoadBalancer{
					Method: "wrr",
				},
			},
			"foo/namedthing": {
				Servers: map[string]types.Server{
					"https://example.com": {
						URL:    "https://example.com",
						Weight: 1,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer: &types.LoadBalancer{
					Method: "wrr",
				},
			},
			"bar": {
				Servers: map[string]types.Server{
					"https://10.15.0.1:8443": {
						URL:    "https://10.15.0.1:8443",
						Weight: 1,
					},
					"https://10.15.0.2:9443": {
						URL:    "https://10.15.0.2:9443",
						Weight: 1,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer: &types.LoadBalancer{
					Method: "wrr",
				},
			},
		},
		Frontends: map[string]*types.Frontend{
			"foo/bar": {
				Backend:        "foo/bar",
				PassHostHeader: true,
				Routes: map[string]types.Route{
					"/bar": {
						Rule: "PathPrefix:/bar",
					},
					"foo": {
						Rule: "Host:foo",
					},
				},
			},
			"foo/namedthing": {
				Backend:        "foo/namedthing",
				PassHostHeader: true,
				Routes: map[string]types.Route{
					"/namedthing": {
						Rule: "PathPrefix:/namedthing",
					},
					"foo": {
						Rule: "Host:foo",
					},
				},
			},
			"bar": {
				Backend:        "bar",
				PassHostHeader: true,
				Routes: map[string]types.Route{
					"bar": {
						Rule: "Host:bar",
					},
				},
			},
		},
		TLSConfiguration: []*tls.Configuration{
			{
				EntryPoints: []string{"https"},
				Certificate: &tls.Certificate{
					CertFile: tls.FileOrContent("certificate"),
					KeyFile:  tls.FileOrContent("key"),
				},
			},
		},
	}
}
