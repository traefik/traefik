package router

import (
	"crypto/tls"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/middlewares/requestdecorator"
	httpmuxer "github.com/traefik/traefik/v3/pkg/muxer/http"
	"github.com/traefik/traefik/v3/pkg/server/middleware"
	"github.com/traefik/traefik/v3/pkg/server/service"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
	traefiktls "github.com/traefik/traefik/v3/pkg/tls"
)

func TestRouterManager_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	t.Cleanup(func() { server.Close() })

	type expectedResult struct {
		StatusCode     int
		RequestHeaders map[string]string
	}

	testCases := []struct {
		desc              string
		routersConfig     map[string]*dynamic.Router
		serviceConfig     map[string]*dynamic.Service
		middlewaresConfig map[string]*dynamic.Middleware
		entryPoints       []string
		expected          expectedResult
	}{
		{
			desc: "no middleware",
			routersConfig: map[string]*dynamic.Router{
				"foo": {
					EntryPoints: []string{"web"},
					Service:     "foo-service",
					Rule:        "Host(`foo.bar`)",
				},
			},
			serviceConfig: map[string]*dynamic.Service{
				"foo-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Strategy: dynamic.BalancerStrategyWRR,
						Servers: []dynamic.Server{
							{
								URL: server.URL,
							},
						},
					},
				},
			},
			entryPoints: []string{"web"},
			expected:    expectedResult{StatusCode: http.StatusOK},
		},
		{
			desc: "empty host",
			routersConfig: map[string]*dynamic.Router{
				"foo": {
					EntryPoints: []string{"web"},
					Service:     "foo-service",
					Rule:        "Host(``)",
				},
			},
			serviceConfig: map[string]*dynamic.Service{
				"foo-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Strategy: dynamic.BalancerStrategyWRR,
						Servers: []dynamic.Server{
							{
								URL: server.URL,
							},
						},
					},
				},
			},
			entryPoints: []string{"web"},
			expected:    expectedResult{StatusCode: http.StatusNotFound},
		},
		{
			desc: "no load balancer",
			routersConfig: map[string]*dynamic.Router{
				"foo": {
					EntryPoints: []string{"web"},
					Service:     "foo-service",
					Rule:        "Host(`foo.bar`)",
				},
			},
			serviceConfig: map[string]*dynamic.Service{
				"foo-service": {},
			},
			entryPoints: []string{"web"},
			expected:    expectedResult{StatusCode: http.StatusNotFound},
		},
		{
			desc: "no middleware, no matching",
			routersConfig: map[string]*dynamic.Router{
				"foo": {
					EntryPoints: []string{"web"},
					Service:     "foo-service",
					Rule:        "Host(`bar.bar`)",
				},
			},
			serviceConfig: map[string]*dynamic.Service{
				"foo-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Strategy: dynamic.BalancerStrategyWRR,
						Servers: []dynamic.Server{
							{
								URL: server.URL,
							},
						},
					},
				},
			},
			entryPoints: []string{"web"},
			expected:    expectedResult{StatusCode: http.StatusNotFound},
		},
		{
			desc: "middleware: headers > auth",
			routersConfig: map[string]*dynamic.Router{
				"foo": {
					EntryPoints: []string{"web"},
					Middlewares: []string{"headers-middle", "auth-middle"},
					Service:     "foo-service",
					Rule:        "Host(`foo.bar`)",
				},
			},
			serviceConfig: map[string]*dynamic.Service{
				"foo-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Strategy: dynamic.BalancerStrategyWRR,
						Servers: []dynamic.Server{
							{
								URL: server.URL,
							},
						},
					},
				},
			},
			middlewaresConfig: map[string]*dynamic.Middleware{
				"auth-middle": {
					BasicAuth: &dynamic.BasicAuth{
						Users: []string{"toto:titi"},
					},
				},
				"headers-middle": {
					Headers: &dynamic.Headers{
						CustomRequestHeaders: map[string]string{"X-Apero": "beer"},
					},
				},
			},
			entryPoints: []string{"web"},
			expected: expectedResult{
				StatusCode: http.StatusUnauthorized,
				RequestHeaders: map[string]string{
					"X-Apero": "beer",
				},
			},
		},
		{
			desc: "middleware: auth > header",
			routersConfig: map[string]*dynamic.Router{
				"foo": {
					EntryPoints: []string{"web"},
					Middlewares: []string{"auth-middle", "headers-middle"},
					Service:     "foo-service",
					Rule:        "Host(`foo.bar`)",
				},
			},
			serviceConfig: map[string]*dynamic.Service{
				"foo-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Strategy: dynamic.BalancerStrategyWRR,
						Servers: []dynamic.Server{
							{
								URL: server.URL,
							},
						},
					},
				},
			},
			middlewaresConfig: map[string]*dynamic.Middleware{
				"auth-middle": {
					BasicAuth: &dynamic.BasicAuth{
						Users: []string{"toto:titi"},
					},
				},
				"headers-middle": {
					Headers: &dynamic.Headers{
						CustomRequestHeaders: map[string]string{"X-Apero": "beer"},
					},
				},
			},
			entryPoints: []string{"web"},
			expected: expectedResult{
				StatusCode: http.StatusUnauthorized,
				RequestHeaders: map[string]string{
					"X-Apero": "",
				},
			},
		},
		{
			desc: "no middleware with provider name",
			routersConfig: map[string]*dynamic.Router{
				"foo@provider-1": {
					EntryPoints: []string{"web"},
					Service:     "foo-service",
					Rule:        "Host(`foo.bar`)",
				},
			},
			serviceConfig: map[string]*dynamic.Service{
				"foo-service@provider-1": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Strategy: dynamic.BalancerStrategyWRR,
						Servers: []dynamic.Server{
							{
								URL: server.URL,
							},
						},
					},
				},
			},
			entryPoints: []string{"web"},
			expected:    expectedResult{StatusCode: http.StatusOK},
		},
		{
			desc: "no middleware with specified provider name",
			routersConfig: map[string]*dynamic.Router{
				"foo@provider-1": {
					EntryPoints: []string{"web"},
					Service:     "foo-service@provider-2",
					Rule:        "Host(`foo.bar`)",
				},
			},
			serviceConfig: map[string]*dynamic.Service{
				"foo-service@provider-2": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Strategy: dynamic.BalancerStrategyWRR,
						Servers: []dynamic.Server{
							{
								URL: server.URL,
							},
						},
					},
				},
			},
			entryPoints: []string{"web"},
			expected:    expectedResult{StatusCode: http.StatusOK},
		},
		{
			desc: "middleware: chain with provider name",
			routersConfig: map[string]*dynamic.Router{
				"foo@provider-1": {
					EntryPoints: []string{"web"},
					Middlewares: []string{"chain-middle@provider-2", "headers-middle"},
					Service:     "foo-service",
					Rule:        "Host(`foo.bar`)",
				},
			},
			serviceConfig: map[string]*dynamic.Service{
				"foo-service@provider-1": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Strategy: dynamic.BalancerStrategyWRR,
						Servers: []dynamic.Server{
							{
								URL: server.URL,
							},
						},
					},
				},
			},
			middlewaresConfig: map[string]*dynamic.Middleware{
				"chain-middle@provider-2": {
					Chain: &dynamic.Chain{Middlewares: []string{"auth-middle"}},
				},
				"auth-middle@provider-2": {
					BasicAuth: &dynamic.BasicAuth{
						Users: []string{"toto:titi"},
					},
				},
				"headers-middle@provider-1": {
					Headers: &dynamic.Headers{
						CustomRequestHeaders: map[string]string{"X-Apero": "beer"},
					},
				},
			},
			entryPoints: []string{"web"},
			expected: expectedResult{
				StatusCode: http.StatusUnauthorized,
				RequestHeaders: map[string]string{
					"X-Apero": "",
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rtConf := runtime.NewConfig(dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Services:    test.serviceConfig,
					Routers:     test.routersConfig,
					Middlewares: test.middlewaresConfig,
				},
			})

			transportManager := service.NewTransportManager(nil)
			transportManager.Update(map[string]*dynamic.ServersTransport{"default@internal": {}})

			serviceManager := service.NewManager(rtConf.Services, nil, nil, transportManager, proxyBuilderMock{})
			middlewaresBuilder := middleware.NewBuilder(rtConf.Middlewares, serviceManager, nil)
			tlsManager := traefiktls.NewManager(nil)

			parser, err := httpmuxer.NewSyntaxParser()
			require.NoError(t, err)

			routerManager := NewManager(rtConf, serviceManager, middlewaresBuilder, nil, tlsManager, parser)

			handlers := routerManager.BuildHandlers(t.Context(), test.entryPoints, false)

			w := httptest.NewRecorder()
			req := testhelpers.MustNewRequest(http.MethodGet, "http://foo.bar/", nil)

			reqHost := requestdecorator.New(nil)
			reqHost.ServeHTTP(w, req, handlers["web"].ServeHTTP)

			assert.Equal(t, test.expected.StatusCode, w.Code)

			for key, value := range test.expected.RequestHeaders {
				assert.Equal(t, value, req.Header.Get(key))
			}
		})
	}
}

func TestRuntimeConfiguration(t *testing.T) {
	testCases := []struct {
		desc             string
		serviceConfig    map[string]*dynamic.Service
		routerConfig     map[string]*dynamic.Router
		middlewareConfig map[string]*dynamic.Middleware
		tlsOptions       map[string]traefiktls.Options
		expectedError    int
	}{
		{
			desc: "No error",
			serviceConfig: map[string]*dynamic.Service{
				"foo-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Strategy: dynamic.BalancerStrategyWRR,
						Servers: []dynamic.Server{
							{
								URL: "http://127.0.0.1:8085",
							},
							{
								URL: "http://127.0.0.1:8086",
							},
						},
						HealthCheck: &dynamic.ServerHealthCheck{
							Interval: ptypes.Duration(500 * time.Millisecond),
							Path:     "/health",
						},
					},
				},
			},
			routerConfig: map[string]*dynamic.Router{
				"foo": {
					EntryPoints: []string{"web"},
					Service:     "foo-service",
					Rule:        "Host(`bar.foo`)",
				},
				"bar": {
					EntryPoints: []string{"web"},
					Service:     "foo-service",
					Rule:        "Host(`foo.bar`)",
				},
			},
			expectedError: 0,
		},
		{
			desc: "One router with wrong rule",
			serviceConfig: map[string]*dynamic.Service{
				"foo-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Strategy: dynamic.BalancerStrategyWRR,
						Servers: []dynamic.Server{
							{
								URL: "http://127.0.0.1",
							},
						},
					},
				},
			},
			routerConfig: map[string]*dynamic.Router{
				"foo": {
					EntryPoints: []string{"web"},
					Service:     "foo-service",
					Rule:        "WrongRule(`bar.foo`)",
				},
				"bar": {
					EntryPoints: []string{"web"},
					Service:     "foo-service",
					Rule:        "Host(`foo.bar`)",
				},
			},
			expectedError: 1,
		},
		{
			desc: "All router with wrong rule",
			serviceConfig: map[string]*dynamic.Service{
				"foo-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Strategy: dynamic.BalancerStrategyWRR,
						Servers: []dynamic.Server{
							{
								URL: "http://127.0.0.1",
							},
						},
					},
				},
			},
			routerConfig: map[string]*dynamic.Router{
				"foo": {
					EntryPoints: []string{"web"},
					Service:     "foo-service",
					Rule:        "WrongRule(`bar.foo`)",
				},
				"bar": {
					EntryPoints: []string{"web"},
					Service:     "foo-service",
					Rule:        "WrongRule(`foo.bar`)",
				},
			},
			expectedError: 2,
		},
		{
			desc: "Router with unknown service",
			serviceConfig: map[string]*dynamic.Service{
				"foo-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Strategy: dynamic.BalancerStrategyWRR,
						Servers: []dynamic.Server{
							{
								URL: "http://127.0.0.1",
							},
						},
					},
				},
			},
			routerConfig: map[string]*dynamic.Router{
				"foo": {
					EntryPoints: []string{"web"},
					Service:     "wrong-service",
					Rule:        "Host(`bar.foo`)",
				},
				"bar": {
					EntryPoints: []string{"web"},
					Service:     "foo-service",
					Rule:        "Host(`foo.bar`)",
				},
			},
			expectedError: 1,
		},
		{
			desc: "Router with broken service",
			serviceConfig: map[string]*dynamic.Service{
				"foo-service": {
					LoadBalancer: nil,
				},
			},
			routerConfig: map[string]*dynamic.Router{
				"bar": {
					EntryPoints: []string{"web"},
					Service:     "foo-service",
					Rule:        "Host(`foo.bar`)",
				},
			},
			expectedError: 2,
		},
		{
			desc: "Router with middleware",
			serviceConfig: map[string]*dynamic.Service{
				"foo-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Strategy: dynamic.BalancerStrategyWRR,
						Servers: []dynamic.Server{
							{
								URL: "http://127.0.0.1",
							},
						},
					},
				},
			},
			middlewareConfig: map[string]*dynamic.Middleware{
				"auth": {
					BasicAuth: &dynamic.BasicAuth{
						Users: []string{"admin:admin"},
					},
				},
				"addPrefixTest": {
					AddPrefix: &dynamic.AddPrefix{
						Prefix: "/toto",
					},
				},
			},
			routerConfig: map[string]*dynamic.Router{
				"bar": {
					EntryPoints: []string{"web"},
					Service:     "foo-service",
					Rule:        "Host(`foo.bar`)",
					Middlewares: []string{"auth", "addPrefixTest"},
				},
				"test": {
					EntryPoints: []string{"web"},
					Service:     "foo-service",
					Rule:        "Host(`foo.bar.other`)",
					Middlewares: []string{"addPrefixTest", "auth"},
				},
			},
		},
		{
			desc: "Router with unknown middleware",
			serviceConfig: map[string]*dynamic.Service{
				"foo-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Strategy: dynamic.BalancerStrategyWRR,
						Servers: []dynamic.Server{
							{
								URL: "http://127.0.0.1",
							},
						},
					},
				},
			},
			middlewareConfig: map[string]*dynamic.Middleware{
				"auth": {
					BasicAuth: &dynamic.BasicAuth{
						Users: []string{"admin:admin"},
					},
				},
			},
			routerConfig: map[string]*dynamic.Router{
				"bar": {
					EntryPoints: []string{"web"},
					Service:     "foo-service",
					Rule:        "Host(`foo.bar`)",
					Middlewares: []string{"unknown"},
				},
			},
			expectedError: 1,
		},
		{
			desc: "Router with broken middleware",
			serviceConfig: map[string]*dynamic.Service{
				"foo-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Strategy: dynamic.BalancerStrategyWRR,
						Servers: []dynamic.Server{
							{
								URL: "http://127.0.0.1",
							},
						},
					},
				},
			},
			middlewareConfig: map[string]*dynamic.Middleware{
				"auth": {
					BasicAuth: &dynamic.BasicAuth{
						Users: []string{"foo"},
					},
				},
			},
			routerConfig: map[string]*dynamic.Router{
				"bar": {
					EntryPoints: []string{"web"},
					Service:     "foo-service",
					Rule:        "Host(`foo.bar`)",
					Middlewares: []string{"auth"},
				},
			},
			expectedError: 2,
		},
		{
			desc: "Router priority exceeding max user-defined priority",
			serviceConfig: map[string]*dynamic.Service{
				"foo-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Strategy: dynamic.BalancerStrategyWRR,
						Servers: []dynamic.Server{
							{
								URL: "http://127.0.0.1",
							},
						},
					},
				},
			},
			middlewareConfig: map[string]*dynamic.Middleware{},
			routerConfig: map[string]*dynamic.Router{
				"bar": {
					EntryPoints: []string{"web"},
					Service:     "foo-service",
					Rule:        "Host(`foo.bar`)",
					Priority:    math.MaxInt,
					TLS:         &dynamic.RouterTLSConfig{},
				},
			},
			tlsOptions:    map[string]traefiktls.Options{},
			expectedError: 1,
		},
		{
			desc: "Router with broken tlsOption",
			serviceConfig: map[string]*dynamic.Service{
				"foo-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Strategy: dynamic.BalancerStrategyWRR,
						Servers: []dynamic.Server{
							{
								URL: "http://127.0.0.1",
							},
						},
					},
				},
			},
			middlewareConfig: map[string]*dynamic.Middleware{},
			routerConfig: map[string]*dynamic.Router{
				"bar": {
					EntryPoints: []string{"web"},
					Service:     "foo-service",
					Rule:        "Host(`foo.bar`)",
					TLS: &dynamic.RouterTLSConfig{
						Options: "broken-tlsOption",
					},
				},
			},
			tlsOptions: map[string]traefiktls.Options{
				"broken-tlsOption": {
					ClientAuth: traefiktls.ClientAuth{
						ClientAuthType: "foobar",
					},
				},
			},
			expectedError: 1,
		},
		{
			desc: "Router with broken default tlsOption",
			serviceConfig: map[string]*dynamic.Service{
				"foo-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Strategy: dynamic.BalancerStrategyWRR,
						Servers: []dynamic.Server{
							{
								URL: "http://127.0.0.1",
							},
						},
					},
				},
			},
			middlewareConfig: map[string]*dynamic.Middleware{},
			routerConfig: map[string]*dynamic.Router{
				"bar": {
					EntryPoints: []string{"web"},
					Service:     "foo-service",
					Rule:        "Host(`foo.bar`)",
					TLS:         &dynamic.RouterTLSConfig{},
				},
			},
			tlsOptions: map[string]traefiktls.Options{
				"default": {
					ClientAuth: traefiktls.ClientAuth{
						ClientAuthType: "foobar",
					},
				},
			},
			expectedError: 1,
		},
	}
	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			entryPoints := []string{"web"}

			rtConf := runtime.NewConfig(dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Services:    test.serviceConfig,
					Routers:     test.routerConfig,
					Middlewares: test.middlewareConfig,
				},
				TLS: &dynamic.TLSConfiguration{
					Options: test.tlsOptions,
				},
			})

			transportManager := service.NewTransportManager(nil)
			transportManager.Update(map[string]*dynamic.ServersTransport{"default@internal": {}})

			serviceManager := service.NewManager(rtConf.Services, nil, nil, transportManager, proxyBuilderMock{})
			middlewaresBuilder := middleware.NewBuilder(rtConf.Middlewares, serviceManager, nil)
			tlsManager := traefiktls.NewManager(nil)
			tlsManager.UpdateConfigs(t.Context(), nil, test.tlsOptions, nil)

			parser, err := httpmuxer.NewSyntaxParser()
			require.NoError(t, err)

			routerManager := NewManager(rtConf, serviceManager, middlewaresBuilder, nil, tlsManager, parser)

			_ = routerManager.BuildHandlers(t.Context(), entryPoints, false)
			_ = routerManager.BuildHandlers(t.Context(), entryPoints, true)

			// even though rtConf was passed by argument to the manager builders above,
			// it's ok to use it as the result we check, because everything worth checking
			// can be accessed by pointers in it.
			var allErrors int
			for _, v := range rtConf.Services {
				if v.Err != nil {
					allErrors++
				}
			}
			for _, v := range rtConf.Routers {
				if len(v.Err) > 0 {
					allErrors++
				}
			}
			for _, v := range rtConf.Middlewares {
				if v.Err != nil {
					allErrors++
				}
			}
			assert.Equal(t, test.expectedError, allErrors)
		})
	}
}

func TestProviderOnMiddlewares(t *testing.T) {
	entryPoints := []string{"web"}

	rtConf := runtime.NewConfig(dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Services: map[string]*dynamic.Service{
				"test@file": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Strategy: dynamic.BalancerStrategyWRR,
						Servers:  []dynamic.Server{},
					},
				},
			},
			Routers: map[string]*dynamic.Router{
				"router@file": {
					EntryPoints: []string{"web"},
					Rule:        "Host(`test`)",
					Service:     "test@file",
					Middlewares: []string{"chain@file", "m1"},
				},
				"router@docker": {
					EntryPoints: []string{"web"},
					Rule:        "Host(`test`)",
					Service:     "test@file",
					Middlewares: []string{"chain", "m1@file"},
				},
			},
			Middlewares: map[string]*dynamic.Middleware{
				"chain@file": {
					Chain: &dynamic.Chain{Middlewares: []string{"m1", "m2", "m1@file"}},
				},
				"chain@docker": {
					Chain: &dynamic.Chain{Middlewares: []string{"m1", "m2", "m1@file"}},
				},
				"m1@file":   {AddPrefix: &dynamic.AddPrefix{Prefix: "/m1"}},
				"m2@file":   {AddPrefix: &dynamic.AddPrefix{Prefix: "/m2"}},
				"m1@docker": {AddPrefix: &dynamic.AddPrefix{Prefix: "/m1"}},
				"m2@docker": {AddPrefix: &dynamic.AddPrefix{Prefix: "/m2"}},
			},
		},
	})

	transportManager := service.NewTransportManager(nil)
	transportManager.Update(map[string]*dynamic.ServersTransport{"default@internal": {}})

	serviceManager := service.NewManager(rtConf.Services, nil, nil, transportManager, nil)
	middlewaresBuilder := middleware.NewBuilder(rtConf.Middlewares, serviceManager, nil)
	tlsManager := traefiktls.NewManager(nil)

	parser, err := httpmuxer.NewSyntaxParser()
	require.NoError(t, err)

	routerManager := NewManager(rtConf, serviceManager, middlewaresBuilder, nil, tlsManager, parser)

	_ = routerManager.BuildHandlers(t.Context(), entryPoints, false)

	assert.Equal(t, []string{"chain@file", "m1@file"}, rtConf.Routers["router@file"].Middlewares)
	assert.Equal(t, []string{"m1@file", "m2@file", "m1@file"}, rtConf.Middlewares["chain@file"].Chain.Middlewares)
	assert.Equal(t, []string{"chain@docker", "m1@file"}, rtConf.Routers["router@docker"].Middlewares)
	assert.Equal(t, []string{"m1@docker", "m2@docker", "m1@file"}, rtConf.Middlewares["chain@docker"].Chain.Middlewares)
}

type staticTransportManager struct {
	res *http.Response
}

func (s staticTransportManager) GetRoundTripper(_ string) (http.RoundTripper, error) {
	return &staticTransport{res: s.res}, nil
}

func (s staticTransportManager) GetTLSConfig(_ string) (*tls.Config, error) {
	panic("implement me")
}

func (s staticTransportManager) Get(_ string) (*dynamic.ServersTransport, error) {
	panic("implement me")
}

type staticTransport struct {
	res *http.Response
}

func (t *staticTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	return t.res, nil
}

func BenchmarkRouterServe(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	b.Cleanup(func() { server.Close() })

	res := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("")),
	}

	routersConfig := map[string]*dynamic.Router{
		"foo": {
			EntryPoints: []string{"web"},
			Service:     "foo-service",
			Rule:        "Host(`foo.bar`) && Path(`/`)",
		},
	}
	serviceConfig := map[string]*dynamic.Service{
		"foo-service": {
			LoadBalancer: &dynamic.ServersLoadBalancer{
				Servers: []dynamic.Server{
					{
						URL: server.URL,
					},
				},
			},
		},
	}
	entryPoints := []string{"web"}

	rtConf := runtime.NewConfig(dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Services:    serviceConfig,
			Routers:     routersConfig,
			Middlewares: map[string]*dynamic.Middleware{},
		},
	})

	serviceManager := service.NewManager(rtConf.Services, nil, nil, staticTransportManager{res}, nil)
	middlewaresBuilder := middleware.NewBuilder(rtConf.Middlewares, serviceManager, nil)
	tlsManager := traefiktls.NewManager(nil)

	parser, err := httpmuxer.NewSyntaxParser()
	require.NoError(b, err)

	routerManager := NewManager(rtConf, serviceManager, middlewaresBuilder, nil, tlsManager, parser)

	handlers := routerManager.BuildHandlers(b.Context(), entryPoints, false)

	w := httptest.NewRecorder()
	req := testhelpers.MustNewRequest(http.MethodGet, "http://foo.bar/", nil)

	reqHost := requestdecorator.New(nil)
	b.ReportAllocs()
	for range b.N {
		reqHost.ServeHTTP(w, req, handlers["web"].ServeHTTP)
	}
}

func BenchmarkService(b *testing.B) {
	res := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("")),
	}

	serviceConfig := map[string]*dynamic.Service{
		"foo-service": {
			LoadBalancer: &dynamic.ServersLoadBalancer{
				Servers: []dynamic.Server{
					{
						URL: "tchouk",
					},
				},
			},
		},
	}

	rtConf := runtime.NewConfig(dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Services: serviceConfig,
		},
	})

	serviceManager := service.NewManager(rtConf.Services, nil, nil, staticTransportManager{res}, nil)
	w := httptest.NewRecorder()
	req := testhelpers.MustNewRequest(http.MethodGet, "http://foo.bar/", nil)

	handler, _ := serviceManager.BuildHTTP(b.Context(), "foo-service")
	b.ReportAllocs()
	for range b.N {
		handler.ServeHTTP(w, req)
	}
}

type proxyBuilderMock struct{}

func (p proxyBuilderMock) Build(_ string, _ *url.URL, _, _ bool, _ time.Duration) (http.Handler, error) {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, req *http.Request) {}), nil
}

func (p proxyBuilderMock) Update(_ map[string]*dynamic.ServersTransport) {
	panic("implement me")
}
