package router

import (
	"context"
	"crypto/tls"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/containous/alice"
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

func TestManager_ComputeMultiLayerRouting(t *testing.T) {
	testCases := []struct {
		desc                string
		routers             map[string]*dynamic.Router
		expectedStatuses    map[string]string
		expectedChildRefs   map[string][]string
		expectedErrors      map[string][]string
		expectedErrorCounts map[string]int
	}{
		{
			desc: "Simple router",
			routers: map[string]*dynamic.Router{
				"A": {
					Service: "A-service",
				},
			},
			expectedStatuses: map[string]string{
				"A": runtime.StatusEnabled,
			},
			expectedChildRefs: map[string][]string{
				"A": {},
			},
		},
		{
			// A->B1
			//  ->B2
			desc: "Router with two children",
			routers: map[string]*dynamic.Router{
				"A": {},
				"B1": {
					ParentRefs: []string{"A"},
					Service:    "B1-service",
				},
				"B2": {
					ParentRefs: []string{"A"},
					Service:    "B2-service",
				},
			},
			expectedStatuses: map[string]string{
				"A":  runtime.StatusEnabled,
				"B1": runtime.StatusEnabled,
				"B2": runtime.StatusEnabled,
			},
			expectedChildRefs: map[string][]string{
				"A":  {"B1", "B2"},
				"B1": nil,
				"B2": nil,
			},
		},
		{
			desc: "Non-root router with TLS config",
			routers: map[string]*dynamic.Router{
				"A": {},
				"B": {
					ParentRefs: []string{"A"},
					Service:    "B-service",
					TLS:        &dynamic.RouterTLSConfig{},
				},
			},
			expectedStatuses: map[string]string{
				"A": runtime.StatusEnabled,
				"B": runtime.StatusDisabled,
			},
			expectedChildRefs: map[string][]string{
				"A": {"B"},
				"B": nil,
			},
			expectedErrors: map[string][]string{
				"B": {"non-root router cannot have TLS configuration"},
			},
		},
		{
			desc: "Non-root router with observability config",
			routers: map[string]*dynamic.Router{
				"A": {},
				"B": {
					ParentRefs:    []string{"A"},
					Service:       "B-service",
					Observability: &dynamic.RouterObservabilityConfig{},
				},
			},
			expectedStatuses: map[string]string{
				"A": runtime.StatusEnabled,
				"B": runtime.StatusDisabled,
			},
			expectedChildRefs: map[string][]string{
				"A": {"B"},
				"B": nil,
			},
			expectedErrors: map[string][]string{
				"B": {"non-root router cannot have Observability configuration"},
			},
		},
		{
			desc: "Non-root router with EntryPoints config",
			routers: map[string]*dynamic.Router{
				"A": {},
				"B": {
					ParentRefs:  []string{"A"},
					Service:     "B-service",
					EntryPoints: []string{"web"},
				},
			},
			expectedStatuses: map[string]string{
				"A": runtime.StatusEnabled,
				"B": runtime.StatusDisabled,
			},
			expectedChildRefs: map[string][]string{
				"A": {"B"},
				"B": nil,
			},
			expectedErrors: map[string][]string{
				"B": {"non-root router cannot have Entrypoints configuration"},
			},
		},

		{
			desc: "Router with non-existing parent",
			routers: map[string]*dynamic.Router{
				"B": {
					ParentRefs: []string{"A"},
					Service:    "B-service",
				},
			},
			expectedStatuses: map[string]string{
				"B": runtime.StatusDisabled,
			},
			expectedChildRefs: map[string][]string{
				"B": nil,
			},
			expectedErrors: map[string][]string{
				"B": {"parent router \"A\" does not exist", "router is not reachable"},
			},
		},
		{
			desc: "Dead-end router with no child and no service",
			routers: map[string]*dynamic.Router{
				"A": {},
			},
			expectedStatuses: map[string]string{
				"A": runtime.StatusDisabled,
			},
			expectedChildRefs: map[string][]string{
				"A": {},
			},
			expectedErrors: map[string][]string{
				"A": {"router has no service and no child routers"},
			},
		},
		{
			// A->B->A
			desc: "Router is not reachable",
			routers: map[string]*dynamic.Router{
				"A": {
					ParentRefs: []string{"B"},
				},
				"B": {
					ParentRefs: []string{"A"},
				},
			},
			expectedStatuses: map[string]string{
				"A": runtime.StatusDisabled,
				"B": runtime.StatusDisabled,
			},
			expectedChildRefs: map[string][]string{
				"A": {"B"},
				"B": {"A"},
			},
			// Cycle detection does not visit unreachable routers (it avoids computing the cycle dependency graph for unreachable routers).
			expectedErrors: map[string][]string{
				"A": {"router is not reachable"},
				"B": {"router is not reachable"},
			},
		},
		{
			// A->B->C->D->B
			desc: "Router creating a cycle is a dead-end and should be disabled",
			routers: map[string]*dynamic.Router{
				"A": {},
				"B": {
					ParentRefs: []string{"A", "D"},
				},
				"C": {
					ParentRefs: []string{"B"},
				},
				"D": {
					ParentRefs: []string{"C"},
				},
			},
			expectedStatuses: map[string]string{
				"A": runtime.StatusEnabled,
				"B": runtime.StatusEnabled,
				"C": runtime.StatusEnabled,
				"D": runtime.StatusDisabled, // Dead-end router is disabled, because the cycle error broke the link with B.
			},
			expectedChildRefs: map[string][]string{
				"A": {"B"},
				"B": {"C"},
				"C": {"D"},
				"D": {},
			},
			expectedErrors: map[string][]string{
				"D": {
					"cyclic reference detected in router tree: B -> C -> D -> B",
					"router has no service and no child routers",
				},
			},
		},
		{
			// A->B->C->D->B
			//           ->E
			desc: "Router creating a cycle A->B->C->D->B but which is referenced elsewhere, must be set to warning status",
			routers: map[string]*dynamic.Router{
				"A": {},
				"B": {
					ParentRefs: []string{"A", "D"},
				},
				"C": {
					ParentRefs: []string{"B"},
				},
				"D": {
					ParentRefs: []string{"C"},
				},
				"E": {
					ParentRefs: []string{"D"},
					Service:    "E-service",
				},
			},
			expectedStatuses: map[string]string{
				"A": runtime.StatusEnabled,
				"B": runtime.StatusEnabled,
				"C": runtime.StatusEnabled,
				"D": runtime.StatusWarning,
				"E": runtime.StatusEnabled,
			},
			expectedChildRefs: map[string][]string{
				"A": {"B"},
				"B": {"C"},
				"C": {"D"},
				"D": {"E"},
			},
			expectedErrors: map[string][]string{
				"D": {"cyclic reference detected in router tree: B -> C -> D -> B"},
			},
		},
		{
			desc: "Parent router with all children having errors",
			routers: map[string]*dynamic.Router{
				"parent": {},
				"child-a": {
					ParentRefs: []string{"parent"},
					Service:    "child-a-service",
					TLS:        &dynamic.RouterTLSConfig{}, // Invalid: non-root cannot have TLS
				},
				"child-b": {
					ParentRefs: []string{"parent"},
					Service:    "child-b-service",
					TLS:        &dynamic.RouterTLSConfig{}, // Invalid: non-root cannot have TLS
				},
			},
			expectedStatuses: map[string]string{
				"parent":  runtime.StatusEnabled, // Enabled during ParseRouterTree (no config errors). Would be disabled during handler building when empty muxer is detected.
				"child-a": runtime.StatusDisabled,
				"child-b": runtime.StatusDisabled,
			},
			expectedChildRefs: map[string][]string{
				"parent":  {"child-a", "child-b"},
				"child-a": nil,
				"child-b": nil,
			},
			expectedErrors: map[string][]string{
				"child-a": {"non-root router cannot have TLS configuration"},
				"child-b": {"non-root router cannot have TLS configuration"},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			// Create runtime routers
			runtimeRouters := make(map[string]*runtime.RouterInfo)
			for name, router := range test.routers {
				runtimeRouters[name] = &runtime.RouterInfo{
					Router: router,
					Status: runtime.StatusEnabled,
				}
			}

			conf := &runtime.Configuration{
				Routers: runtimeRouters,
			}

			manager := &Manager{
				conf: conf,
			}

			// Execute the function we're testing
			manager.ParseRouterTree()

			// Verify ChildRefs are populated correctly
			for routerName, expectedChildren := range test.expectedChildRefs {
				router := runtimeRouters[routerName]
				assert.ElementsMatch(t, expectedChildren, router.ChildRefs)
			}

			// Verify statuses are set correctly
			var gotStatuses map[string]string
			for routerName, router := range runtimeRouters {
				if gotStatuses == nil {
					gotStatuses = make(map[string]string)
				}
				gotStatuses[routerName] = router.Status
			}
			assert.Equal(t, test.expectedStatuses, gotStatuses)

			// Verify errors are added correctly
			var gotErrors map[string][]string
			for routerName, router := range runtimeRouters {
				for _, err := range router.Err {
					if gotErrors == nil {
						gotErrors = make(map[string][]string)
					}
					gotErrors[routerName] = append(gotErrors[routerName], err)
				}
			}
			assert.Equal(t, test.expectedErrors, gotErrors)
		})
	}
}

func TestManager_buildChildRoutersMuxer(t *testing.T) {
	testCases := []struct {
		desc             string
		childRefs        []string
		routers          map[string]*dynamic.Router
		services         map[string]*dynamic.Service
		middlewares      map[string]*dynamic.Middleware
		expectedError    string
		expectedRequests []struct {
			path       string
			statusCode int
		}
	}{
		{
			desc:      "simple child router with service",
			childRefs: []string{"child1"},
			routers: map[string]*dynamic.Router{
				"child1": {
					Rule:    "Path(`/api`)",
					Service: "child1-service",
				},
			},
			services: map[string]*dynamic.Service{
				"child1-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Servers: []dynamic.Server{{URL: "http://localhost:8080"}},
					},
				},
			},
			expectedRequests: []struct {
				path       string
				statusCode int
			}{
				{path: "/api", statusCode: http.StatusOK},
				{path: "/unknown", statusCode: http.StatusNotFound},
			},
		},
		{
			desc:      "multiple child routers with different rules",
			childRefs: []string{"child1", "child2"},
			routers: map[string]*dynamic.Router{
				"child1": {
					Rule:    "Path(`/api`)",
					Service: "child1-service",
				},
				"child2": {
					Rule:    "Path(`/web`)",
					Service: "child2-service",
				},
			},
			services: map[string]*dynamic.Service{
				"child1-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Servers: []dynamic.Server{{URL: "http://localhost:8080"}},
					},
				},
				"child2-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Servers: []dynamic.Server{{URL: "http://localhost:8081"}},
					},
				},
			},
			expectedRequests: []struct {
				path       string
				statusCode int
			}{
				{path: "/api", statusCode: http.StatusOK},
				{path: "/web", statusCode: http.StatusOK},
				{path: "/unknown", statusCode: http.StatusNotFound},
			},
		},
		{
			desc:      "child router with middleware",
			childRefs: []string{"child1"},
			routers: map[string]*dynamic.Router{
				"child1": {
					Rule:        "Path(`/api`)",
					Service:     "child1-service",
					Middlewares: []string{"test-middleware"},
				},
			},
			services: map[string]*dynamic.Service{
				"child1-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Servers: []dynamic.Server{{URL: "http://localhost:8080"}},
					},
				},
			},
			middlewares: map[string]*dynamic.Middleware{
				"test-middleware": {
					Headers: &dynamic.Headers{
						CustomRequestHeaders: map[string]string{"X-Test": "value"},
					},
				},
			},
			expectedRequests: []struct {
				path       string
				statusCode int
			}{
				{path: "/api", statusCode: http.StatusOK},
				{path: "/unknown", statusCode: http.StatusNotFound},
			},
		},
		{
			desc:      "nested child routers (child with its own children)",
			childRefs: []string{"intermediate"},
			routers: map[string]*dynamic.Router{
				"intermediate": {
					Rule: "PathPrefix(`/api`)",
					// No service - this will have its own children
				},
				"leaf1": {
					Rule:       "Path(`/api/v1`)",
					Service:    "leaf1-service",
					ParentRefs: []string{"intermediate"},
				},
				"leaf2": {
					Rule:       "Path(`/api/v2`)",
					Service:    "leaf2-service",
					ParentRefs: []string{"intermediate"},
				},
			},
			services: map[string]*dynamic.Service{
				"leaf1-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Servers: []dynamic.Server{{URL: "http://localhost:8080"}},
					},
				},
				"leaf2-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Servers: []dynamic.Server{{URL: "http://localhost:8081"}},
					},
				},
			},
			expectedRequests: []struct {
				path       string
				statusCode int
			}{
				{path: "/api/v1", statusCode: http.StatusOK},
				{path: "/api/v2", statusCode: http.StatusOK},
				{path: "/unknown", statusCode: http.StatusNotFound},
			},
		},
		{
			desc:      "all child routers have errors - should return error",
			childRefs: []string{"child1", "child2"},
			routers: map[string]*dynamic.Router{
				"child1": {
					Rule:       "Path(`/api`)",
					Service:    "child1-service",
					ParentRefs: []string{"parent"},
					TLS:        &dynamic.RouterTLSConfig{}, // Invalid: non-root router cannot have TLS
				},
				"child2": {
					Rule:       "Path(`/web`)",
					Service:    "child2-service",
					ParentRefs: []string{"parent"},
					TLS:        &dynamic.RouterTLSConfig{}, // Invalid: non-root router cannot have TLS
				},
			},
			services: map[string]*dynamic.Service{
				"child1-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Servers: []dynamic.Server{{URL: "http://localhost:8080"}},
					},
				},
				"child2-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Servers: []dynamic.Server{{URL: "http://localhost:8081"}},
					},
				},
			},
			expectedError: "no child routers could be added to muxer (2 skipped)",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			// Create runtime routers
			runtimeRouters := make(map[string]*runtime.RouterInfo)
			for name, router := range test.routers {
				runtimeRouters[name] = &runtime.RouterInfo{
					Router: router,
				}
			}

			// Create runtime services
			runtimeServices := make(map[string]*runtime.ServiceInfo)
			for name, service := range test.services {
				runtimeServices[name] = &runtime.ServiceInfo{
					Service: service,
				}
			}

			// Create runtime middlewares
			runtimeMiddlewares := make(map[string]*runtime.MiddlewareInfo)
			for name, middleware := range test.middlewares {
				runtimeMiddlewares[name] = &runtime.MiddlewareInfo{
					Middleware: middleware,
				}
			}

			conf := &runtime.Configuration{
				Routers:     runtimeRouters,
				Services:    runtimeServices,
				Middlewares: runtimeMiddlewares,
			}

			// Set up the manager with mocks
			serviceManager := &mockServiceManager{}
			middlewareBuilder := &mockMiddlewareBuilder{}
			parser, err := httpmuxer.NewSyntaxParser()
			require.NoError(t, err)

			manager := NewManager(conf, serviceManager, middlewareBuilder, nil, nil, parser)

			// Compute multi-layer routing to populate ChildRefs
			manager.ParseRouterTree()

			// Build the child routers muxer
			ctx := t.Context()
			muxer, err := manager.buildChildRoutersMuxer(ctx, test.childRefs)

			if test.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedError)
				return
			}

			if len(test.childRefs) == 0 {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, muxer)

			// Test that the muxer routes requests correctly
			for _, req := range test.expectedRequests {
				recorder := httptest.NewRecorder()
				request := httptest.NewRequest(http.MethodGet, req.path, nil)
				muxer.ServeHTTP(recorder, request)

				assert.Equal(t, req.statusCode, recorder.Code, "unexpected status code for path %s", req.path)
			}
		})
	}
}

func TestManager_buildHTTPHandler_WithChildRouters(t *testing.T) {
	testCases := []struct {
		desc             string
		router           *runtime.RouterInfo
		childRouters     map[string]*dynamic.Router
		services         map[string]*dynamic.Service
		expectedError    string
		expectedRequests []struct {
			path       string
			statusCode int
		}
	}{
		{
			desc: "router with child routers",
			router: &runtime.RouterInfo{
				Router: &dynamic.Router{
					Rule: "PathPrefix(`/api`)",
				},
				ChildRefs: []string{"child1", "child2"},
			},
			childRouters: map[string]*dynamic.Router{
				"child1": {
					Rule:    "Path(`/api/v1`)",
					Service: "child1-service",
				},
				"child2": {
					Rule:    "Path(`/api/v2`)",
					Service: "child2-service",
				},
			},
			services: map[string]*dynamic.Service{
				"child1-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Servers: []dynamic.Server{{URL: "http://localhost:8080"}},
					},
				},
				"child2-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Servers: []dynamic.Server{{URL: "http://localhost:8081"}},
					},
				},
			},
			expectedRequests: []struct {
				path       string
				statusCode int
			}{
				{path: "/unknown", statusCode: http.StatusNotFound},
			},
		},
		{
			desc: "router with service (normal case)",
			router: &runtime.RouterInfo{
				Router: &dynamic.Router{
					Rule:    "PathPrefix(`/api`)",
					Service: "main-service",
				},
			},
			services: map[string]*dynamic.Service{
				"main-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Servers: []dynamic.Server{{URL: "http://localhost:8080"}},
					},
				},
			},
			expectedRequests: []struct {
				path       string
				statusCode int
			}{},
		},
		{
			desc: "router with neither service nor child routers - error",
			router: &runtime.RouterInfo{
				Router: &dynamic.Router{
					Rule: "PathPrefix(`/api`)",
				},
			},
			expectedError: "router must have either a service or child routers",
		},
		{
			desc: "router with child routers but missing child - error",
			router: &runtime.RouterInfo{
				Router: &dynamic.Router{
					Rule: "PathPrefix(`/api`)",
				},
				ChildRefs: []string{"nonexistent"},
			},
			expectedError: "child router \"nonexistent\" does not exist",
		},
		{
			desc: "router with all children having errors - returns empty muxer error",
			router: &runtime.RouterInfo{
				Router: &dynamic.Router{
					Rule: "PathPrefix(`/api`)",
				},
				ChildRefs: []string{"child1", "child2"},
			},
			childRouters: map[string]*dynamic.Router{
				"child1": {
					Rule:       "Path(`/api/v1`)",
					Service:    "child1-service",
					ParentRefs: []string{"parent"},
					TLS:        &dynamic.RouterTLSConfig{}, // Invalid for non-root
				},
				"child2": {
					Rule:       "Path(`/api/v2`)",
					Service:    "child2-service",
					ParentRefs: []string{"parent"},
					TLS:        &dynamic.RouterTLSConfig{}, // Invalid for non-root
				},
			},
			services: map[string]*dynamic.Service{
				"child1-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Servers: []dynamic.Server{{URL: "http://localhost:8080"}},
					},
				},
				"child2-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Servers: []dynamic.Server{{URL: "http://localhost:8081"}},
					},
				},
			},
			expectedError: "no child routers could be added to muxer (2 skipped)",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			// Create runtime routers
			runtimeRouters := make(map[string]*runtime.RouterInfo)
			runtimeRouters["test-router"] = test.router
			for name, router := range test.childRouters {
				runtimeRouters[name] = &runtime.RouterInfo{
					Router: router,
				}
			}

			// Create runtime services
			runtimeServices := make(map[string]*runtime.ServiceInfo)
			for name, service := range test.services {
				runtimeServices[name] = &runtime.ServiceInfo{
					Service: service,
				}
			}

			conf := &runtime.Configuration{
				Routers:  runtimeRouters,
				Services: runtimeServices,
			}

			// Set up the manager with mocks
			serviceManager := &mockServiceManager{}
			middlewareBuilder := &mockMiddlewareBuilder{}
			parser, err := httpmuxer.NewSyntaxParser()
			require.NoError(t, err)

			manager := NewManager(conf, serviceManager, middlewareBuilder, nil, nil, parser)

			// Run ParseRouterTree to validate configuration and populate ChildRefs/errors
			manager.ParseRouterTree()

			// Build the HTTP handler
			ctx := t.Context()
			handler, err := manager.buildHTTPHandler(ctx, test.router, "test-router")

			if test.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedError)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, handler)

			// Test that the handler routes requests correctly
			for _, req := range test.expectedRequests {
				recorder := httptest.NewRecorder()
				request := httptest.NewRequest(http.MethodGet, req.path, nil)
				handler.ServeHTTP(recorder, request)

				assert.Equal(t, req.statusCode, recorder.Code, "unexpected status code for path %s", req.path)
			}
		})
	}
}

func TestManager_BuildHandlers_WithChildRouters(t *testing.T) {
	testCases := []struct {
		desc               string
		routers            map[string]*dynamic.Router
		services           map[string]*dynamic.Service
		entryPoints        []string
		expectedEntryPoint string
		expectedRequests   []struct {
			path       string
			statusCode int
		}
	}{
		{
			desc: "parent router with child routers",
			routers: map[string]*dynamic.Router{
				"parent": {
					EntryPoints: []string{"web"},
					Rule:        "PathPrefix(`/api`)",
				},
				"child1": {
					Rule:       "Path(`/api/v1`)",
					Service:    "child1-service",
					ParentRefs: []string{"parent"},
				},
				"child2": {
					Rule:       "Path(`/api/v2`)",
					Service:    "child2-service",
					ParentRefs: []string{"parent"},
				},
			},
			services: map[string]*dynamic.Service{
				"child1-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Servers: []dynamic.Server{{URL: "http://localhost:8080"}},
					},
				},
				"child2-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Servers: []dynamic.Server{{URL: "http://localhost:8081"}},
					},
				},
			},
			entryPoints:        []string{"web"},
			expectedEntryPoint: "web",
			expectedRequests: []struct {
				path       string
				statusCode int
			}{
				{path: "/unknown", statusCode: http.StatusNotFound},
			},
		},
		{
			desc: "multiple parent routers with children",
			routers: map[string]*dynamic.Router{
				"api-parent": {
					EntryPoints: []string{"web"},
					Rule:        "PathPrefix(`/api`)",
				},
				"web-parent": {
					EntryPoints: []string{"web"},
					Rule:        "PathPrefix(`/web`)",
				},
				"api-child": {
					Rule:       "Path(`/api/v1`)",
					Service:    "api-service",
					ParentRefs: []string{"api-parent"},
				},
				"web-child": {
					Rule:       "Path(`/web/index`)",
					Service:    "web-service",
					ParentRefs: []string{"web-parent"},
				},
			},
			services: map[string]*dynamic.Service{
				"api-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Servers: []dynamic.Server{{URL: "http://localhost:8080"}},
					},
				},
				"web-service": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Servers: []dynamic.Server{{URL: "http://localhost:8081"}},
					},
				},
			},
			entryPoints:        []string{"web"},
			expectedEntryPoint: "web",
			expectedRequests: []struct {
				path       string
				statusCode int
			}{},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			// Create runtime routers
			runtimeRouters := make(map[string]*runtime.RouterInfo)
			for name, router := range test.routers {
				runtimeRouters[name] = &runtime.RouterInfo{
					Router: router,
				}
			}

			// Create runtime services
			runtimeServices := make(map[string]*runtime.ServiceInfo)
			for name, service := range test.services {
				runtimeServices[name] = &runtime.ServiceInfo{
					Service: service,
				}
			}

			conf := &runtime.Configuration{
				Routers:  runtimeRouters,
				Services: runtimeServices,
			}

			// Set up the manager with mocks
			serviceManager := &mockServiceManager{}
			middlewareBuilder := &mockMiddlewareBuilder{}
			parser, err := httpmuxer.NewSyntaxParser()
			require.NoError(t, err)

			manager := NewManager(conf, serviceManager, middlewareBuilder, nil, nil, parser)

			// Compute multi-layer routing to set up parent-child relationships
			manager.ParseRouterTree()

			// Build handlers
			ctx := t.Context()
			handlers := manager.BuildHandlers(ctx, test.entryPoints, false)

			require.Contains(t, handlers, test.expectedEntryPoint)
			handler := handlers[test.expectedEntryPoint]
			require.NotNil(t, handler)

			// Test that the handler routes requests correctly
			for _, req := range test.expectedRequests {
				recorder := httptest.NewRecorder()
				request := httptest.NewRequest(http.MethodGet, req.path, nil)
				request.Host = "test.com"
				handler.ServeHTTP(recorder, request)

				assert.Equal(t, req.statusCode, recorder.Code, "unexpected status code for path %s", req.path)
			}
		})
	}
}

// Mock implementations for testing

type mockServiceManager struct{}

func (m *mockServiceManager) BuildHTTP(_ context.Context, _ string) (http.Handler, error) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("mock service response"))
	}), nil
}

func (m *mockServiceManager) LaunchHealthCheck(_ context.Context) {}

type mockMiddlewareBuilder struct{}

func (m *mockMiddlewareBuilder) BuildChain(_ context.Context, _ []string) *alice.Chain {
	chain := alice.New()
	return &chain
}

type proxyBuilderMock struct{}

func (p proxyBuilderMock) Build(_ string, _ *url.URL, _, _ bool, _ time.Duration) (http.Handler, error) {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, req *http.Request) {}), nil
}

func (p proxyBuilderMock) Update(_ map[string]*dynamic.ServersTransport) {
	panic("implement me")
}
