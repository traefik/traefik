package kubernetes

import (
	"testing"
	"time"

	"github.com/containous/flaeg"
	"github.com/containous/traefik/provider/label"
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

func responseForwarding(interval string) func(*types.Backend) {
	return func(b *types.Backend) {
		b.ResponseForwarding = &types.ResponseForwarding{}
		b.ResponseForwarding.FlushInterval = interval
	}
}

func buffering(opts ...func(*types.Buffering)) func(*types.Backend) {
	return func(b *types.Backend) {
		if b.Buffering == nil {
			b.Buffering = &types.Buffering{}
		}
		for _, opt := range opts {
			opt(b.Buffering)
		}
	}
}

func maxRequestBodyBytes(value int64) func(*types.Buffering) {
	return func(b *types.Buffering) {
		b.MaxRequestBodyBytes = value
	}
}

func memRequestBodyBytes(value int64) func(*types.Buffering) {
	return func(b *types.Buffering) {
		b.MemRequestBodyBytes = value
	}
}

func maxResponseBodyBytes(value int64) func(*types.Buffering) {
	return func(b *types.Buffering) {
		b.MaxResponseBodyBytes = value
	}
}

func memResponseBodyBytes(value int64) func(*types.Buffering) {
	return func(b *types.Buffering) {
		b.MemResponseBodyBytes = value
	}
}

func retrying(exp string) func(*types.Buffering) {
	return func(b *types.Buffering) {
		b.RetryExpression = exp
	}
}

func maxConnExtractorFunc(exp string) func(*types.Backend) {
	return func(b *types.Backend) {
		if b.MaxConn == nil {
			b.MaxConn = &types.MaxConn{}
		}
		b.MaxConn.ExtractorFunc = exp
	}
}

func maxConnAmount(value int64) func(*types.Backend) {
	return func(b *types.Backend) {
		if b.MaxConn == nil {
			b.MaxConn = &types.MaxConn{}
		}
		b.MaxConn.Amount = value
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
		// related the function frontendName
		name := f.Backend
		f.Backend = backend
		if len(name) > 0 {
			return name
		}
		return backend
	}
}

func frontendName(name string) func(*types.Frontend) {
	return func(f *types.Frontend) {
		// store temporary the frontend name into the backend name
		f.Backend = name
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

// Deprecated
func basicAuthDeprecated(auth ...string) func(*types.Frontend) {
	return func(f *types.Frontend) {
		f.Auth = &types.Auth{Basic: &types.Basic{Users: auth}}
	}
}

func auth(opt func(*types.Auth)) func(*types.Frontend) {
	return func(f *types.Frontend) {
		auth := &types.Auth{}
		opt(auth)
		f.Auth = auth
	}
}

func basicAuth(opts ...func(*types.Basic)) func(*types.Auth) {
	return func(a *types.Auth) {
		basic := &types.Basic{}
		for _, opt := range opts {
			opt(basic)
		}
		a.Basic = basic
	}
}

func baUsers(users ...string) func(*types.Basic) {
	return func(b *types.Basic) {
		b.Users = users
	}
}

func baRemoveHeaders() func(*types.Basic) {
	return func(b *types.Basic) {
		b.RemoveHeader = true
	}
}

func forwardAuth(forwardURL string, opts ...func(*types.Forward)) func(*types.Auth) {
	return func(a *types.Auth) {
		fwd := &types.Forward{Address: forwardURL}
		for _, opt := range opts {
			opt(fwd)
		}
		a.Forward = fwd
	}
}

func fwdAuthResponseHeaders(headers ...string) func(*types.Forward) {
	return func(f *types.Forward) {
		f.AuthResponseHeaders = headers
	}
}

func fwdTrustForwardHeader() func(*types.Forward) {
	return func(f *types.Forward) {
		f.TrustForwardHeader = true
	}
}

func fwdAuthTLS(cert, key string, insecure bool) func(*types.Forward) {
	return func(f *types.Forward) {
		f.TLS = &types.ClientTLS{Cert: cert, Key: key, InsecureSkipVerify: insecure}
	}
}

func whiteList(useXFF bool, ranges ...string) func(*types.Frontend) {
	return func(f *types.Frontend) {
		if f.WhiteList == nil {
			f.WhiteList = &types.WhiteList{}
		}
		f.WhiteList.UseXForwardedFor = useXFF
		f.WhiteList.SourceRange = ranges
	}
}

func priority(value int) func(*types.Frontend) {
	return func(f *types.Frontend) {
		f.Priority = value
	}
}

func headers(h *types.Headers) func(*types.Frontend) {
	return func(f *types.Frontend) {
		f.Headers = h
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

func errorPage(name string, opts ...func(*types.ErrorPage)) func(*types.Frontend) {
	return func(f *types.Frontend) {
		if f.Errors == nil {
			f.Errors = make(map[string]*types.ErrorPage)
		}

		if len(name) > 0 {
			f.Errors[name] = &types.ErrorPage{}
			for _, opt := range opts {
				opt(f.Errors[name])
			}
		}
	}
}

func errorStatus(status ...string) func(*types.ErrorPage) {
	return func(page *types.ErrorPage) {
		page.Status = status
	}
}

func errorQuery(query string) func(*types.ErrorPage) {
	return func(page *types.ErrorPage) {
		page.Query = query
	}
}

func errorBackend(backend string) func(*types.ErrorPage) {
	return func(page *types.ErrorPage) {
		page.Backend = backend
	}
}

func rateLimit(opts ...func(*types.RateLimit)) func(*types.Frontend) {
	return func(f *types.Frontend) {
		if f.RateLimit == nil {
			f.RateLimit = &types.RateLimit{}
		}

		for _, opt := range opts {
			opt(f.RateLimit)
		}
	}
}

func rateExtractorFunc(exp string) func(*types.RateLimit) {
	return func(limit *types.RateLimit) {
		limit.ExtractorFunc = exp
	}
}

func rateSet(name string, opts ...func(*types.Rate)) func(*types.RateLimit) {
	return func(limit *types.RateLimit) {
		if limit.RateSet == nil {
			limit.RateSet = make(map[string]*types.Rate)
		}

		if len(name) > 0 {
			limit.RateSet[name] = &types.Rate{}
			for _, opt := range opts {
				opt(limit.RateSet[name])
			}
		}
	}
}

func limitAverage(avg int64) func(*types.Rate) {
	return func(rate *types.Rate) {
		rate.Average = avg
	}
}

func limitBurst(burst int64) func(*types.Rate) {
	return func(rate *types.Rate) {
		rate.Burst = burst
	}
}

func limitPeriod(period time.Duration) func(*types.Rate) {
	return func(rate *types.Rate) {
		rate.Period = flaeg.Duration(period)
	}
}

// Deprecated
func passTLSCert() func(*types.Frontend) {
	return func(f *types.Frontend) {
		f.PassTLSCert = true
	}
}

func passTLSClientCert() func(*types.Frontend) {
	return func(f *types.Frontend) {
		f.PassTLSClientCert = &types.TLSClientHeaders{
			PEM: true,
			Infos: &types.TLSClientCertificateInfos{
				NotAfter:  true,
				NotBefore: true,
				Subject: &types.TLSCLientCertificateDNInfos{
					CommonName:      true,
					Country:         true,
					DomainComponent: true,
					Locality:        true,
					Organization:    true,
					Province:        true,
					SerialNumber:    true,
				},
				Issuer: &types.TLSCLientCertificateDNInfos{
					CommonName:      true,
					Country:         true,
					DomainComponent: true,
					Locality:        true,
					Organization:    true,
					Province:        true,
					SerialNumber:    true,
				},
				Sans: true,
			},
		}
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

func tlsesSection(opts ...func(*tls.Configuration)) func(*types.Configuration) {
	return func(c *types.Configuration) {
		for _, opt := range opts {
			tlsConf := &tls.Configuration{}
			opt(tlsConf)
			c.TLS = append(c.TLS, tlsConf)
		}
	}
}

func tlsSection(opts ...func(*tls.Configuration)) func(*tls.Configuration) {
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
		tlsesSection(
			tlsSection(
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
						Weight: label.DefaultWeight,
					},
					"http://10.21.0.1:8080": {
						URL:    "http://10.21.0.1:8080",
						Weight: label.DefaultWeight,
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
						Weight: label.DefaultWeight,
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
						Weight: label.DefaultWeight,
					},
					"https://10.15.0.2:9443": {
						URL:    "https://10.15.0.2:9443",
						Weight: label.DefaultWeight,
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
		TLS: []*tls.Configuration{
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
