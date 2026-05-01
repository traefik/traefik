package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"net/textproto"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/containous/alice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/proxy/httputil"
	"github.com/traefik/traefik/v3/pkg/safe"
	"github.com/traefik/traefik/v3/pkg/server/provider"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
)

func pointer[T any](v T) *T { return &v }

func TestGetLoadBalancer(t *testing.T) {
	sm := Manager{
		transportManager: &transportManagerMock{},
	}

	testCases := []struct {
		desc        string
		serviceName string
		service     *dynamic.ServersLoadBalancer
		fwd         http.Handler
		expectError bool
	}{
		{
			desc:        "Fails when provided an invalid URL",
			serviceName: "test",
			service: &dynamic.ServersLoadBalancer{
				Strategy: dynamic.BalancerStrategyWRR,
				Servers: []dynamic.Server{
					{
						URL: ":",
					},
				},
			},
			fwd:         &forwarderMock{},
			expectError: true,
		},
		{
			desc:        "Succeeds when there are no servers",
			serviceName: "test",
			service: &dynamic.ServersLoadBalancer{
				Strategy: dynamic.BalancerStrategyWRR,
			},
			fwd:         &forwarderMock{},
			expectError: false,
		},
		{
			desc:        "Succeeds when sticky.cookie is set",
			serviceName: "test",
			service: &dynamic.ServersLoadBalancer{
				Strategy: dynamic.BalancerStrategyWRR,
				Sticky:   &dynamic.Sticky{Cookie: &dynamic.Cookie{}},
			},
			fwd:         &forwarderMock{},
			expectError: false,
		},
		{
			desc:        "Succeeds when passive health checker is set",
			serviceName: "test",
			service: &dynamic.ServersLoadBalancer{
				Strategy: dynamic.BalancerStrategyWRR,
				PassiveHealthCheck: &dynamic.PassiveServerHealthCheck{
					FailureWindow:     ptypes.Duration(30 * time.Second),
					MaxFailedAttempts: 3,
				},
			},
			fwd:         &forwarderMock{},
			expectError: false,
		},
		{
			desc:        "Fails when unsupported strategy is set",
			serviceName: "test",
			service: &dynamic.ServersLoadBalancer{
				Strategy: "invalid",
				Servers: []dynamic.Server{
					{
						URL: "http://localhost:8080",
					},
				},
			},
			fwd:         &forwarderMock{},
			expectError: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			serviceInfo := &runtime.ServiceInfo{Service: &dynamic.Service{LoadBalancer: test.service}}
			handler, err := sm.getLoadBalancerServiceHandler(t.Context(), test.serviceName, serviceInfo)
			if test.expectError {
				require.Error(t, err)
				assert.Nil(t, handler)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, handler)
			}
		})
	}
}

func TestGetLoadBalancerServiceHandler(t *testing.T) {
	pb := httputil.NewProxyBuilder(&transportManagerMock{}, nil)
	sm := NewManager(nil, nil, nil, transportManagerMock{}, pb)

	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-From", "first")
	}))
	t.Cleanup(server1.Close)

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-From", "second")
	}))
	t.Cleanup(server2.Close)

	serverPassHost := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-From", "passhost")
		assert.Equal(t, "callme", r.Host)
	}))
	t.Cleanup(serverPassHost.Close)

	serverPassHostFalse := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-From", "passhostfalse")
		assert.NotEqual(t, "callme", r.Host)
	}))
	t.Cleanup(serverPassHostFalse.Close)

	hasNoUserAgent := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.Header.Get("User-Agent"))
	}))
	t.Cleanup(hasNoUserAgent.Close)

	hasUserAgent := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "foobar", r.Header.Get("User-Agent"))
	}))
	t.Cleanup(hasUserAgent.Close)

	type ExpectedResult struct {
		StatusCode     int
		XFrom          string
		LoadBalanced   bool
		SecureCookie   bool
		HTTPOnlyCookie bool
	}

	testCases := []struct {
		desc             string
		serviceName      string
		service          *dynamic.ServersLoadBalancer
		responseModifier func(*http.Response) error
		cookieRawValue   string
		userAgent        string

		expected []ExpectedResult
	}{
		{
			desc:        "Load balances between the two servers",
			serviceName: "test",
			service: &dynamic.ServersLoadBalancer{
				Strategy:       dynamic.BalancerStrategyWRR,
				PassHostHeader: pointer(true),
				Servers: []dynamic.Server{
					{
						URL: server1.URL,
					},
					{
						URL: server2.URL,
					},
				},
			},
			expected: []ExpectedResult{
				{
					StatusCode:   http.StatusOK,
					LoadBalanced: true,
				},
				{
					StatusCode:   http.StatusOK,
					LoadBalanced: true,
				},
			},
		},
		{
			desc:        "StatusBadGateway when the server is not reachable",
			serviceName: "test",
			service: &dynamic.ServersLoadBalancer{
				Strategy: dynamic.BalancerStrategyWRR,
				Servers: []dynamic.Server{
					{
						URL: "http://foo",
					},
				},
			},
			expected: []ExpectedResult{
				{
					StatusCode: http.StatusBadGateway,
				},
			},
		},
		{
			desc:        "ServiceUnavailable when no servers are available",
			serviceName: "test",
			service: &dynamic.ServersLoadBalancer{
				Strategy: dynamic.BalancerStrategyWRR,
				Servers:  []dynamic.Server{},
			},
			expected: []ExpectedResult{
				{
					StatusCode: http.StatusServiceUnavailable,
				},
			},
		},
		{
			desc:        "Always call the same server when sticky.cookie is true",
			serviceName: "test",
			service: &dynamic.ServersLoadBalancer{
				Strategy: dynamic.BalancerStrategyWRR,
				Sticky:   &dynamic.Sticky{Cookie: &dynamic.Cookie{}},
				Servers: []dynamic.Server{
					{
						URL: server1.URL,
					},
					{
						URL: server2.URL,
					},
				},
			},
			expected: []ExpectedResult{
				{
					StatusCode: http.StatusOK,
				},
				{
					StatusCode: http.StatusOK,
				},
			},
		},
		{
			desc:        "Sticky Cookie's options set correctly",
			serviceName: "test",
			service: &dynamic.ServersLoadBalancer{
				Strategy: dynamic.BalancerStrategyWRR,
				Sticky:   &dynamic.Sticky{Cookie: &dynamic.Cookie{HTTPOnly: true, Secure: true}},
				Servers: []dynamic.Server{
					{
						URL: server1.URL,
					},
				},
			},
			expected: []ExpectedResult{
				{
					StatusCode:     http.StatusOK,
					XFrom:          "first",
					SecureCookie:   true,
					HTTPOnlyCookie: true,
				},
			},
		},
		{
			desc:        "PassHost passes the host instead of the IP",
			serviceName: "test",
			service: &dynamic.ServersLoadBalancer{
				Strategy:       dynamic.BalancerStrategyWRR,
				Sticky:         &dynamic.Sticky{Cookie: &dynamic.Cookie{}},
				PassHostHeader: pointer(true),
				Servers: []dynamic.Server{
					{
						URL: serverPassHost.URL,
					},
				},
			},
			expected: []ExpectedResult{
				{
					StatusCode: http.StatusOK,
					XFrom:      "passhost",
				},
			},
		},
		{
			desc:        "PassHost doesn't pass the host instead of the IP",
			serviceName: "test",
			service: &dynamic.ServersLoadBalancer{
				Strategy:       dynamic.BalancerStrategyWRR,
				PassHostHeader: pointer(false),
				Sticky:         &dynamic.Sticky{Cookie: &dynamic.Cookie{}},
				Servers: []dynamic.Server{
					{
						URL: serverPassHostFalse.URL,
					},
				},
			},
			expected: []ExpectedResult{
				{
					StatusCode: http.StatusOK,
					XFrom:      "passhostfalse",
				},
			},
		},
		{
			desc:        "No user-agent",
			serviceName: "test",
			service: &dynamic.ServersLoadBalancer{
				Strategy: dynamic.BalancerStrategyWRR,
				Servers: []dynamic.Server{
					{
						URL: hasNoUserAgent.URL,
					},
				},
			},
			expected: []ExpectedResult{
				{
					StatusCode: http.StatusOK,
				},
			},
		},
		{
			desc:        "Custom user-agent",
			serviceName: "test",
			userAgent:   "foobar",
			service: &dynamic.ServersLoadBalancer{
				Strategy: dynamic.BalancerStrategyWRR,
				Servers: []dynamic.Server{
					{
						URL: hasUserAgent.URL,
					},
				},
			},
			expected: []ExpectedResult{
				{
					StatusCode: http.StatusOK,
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			serviceInfo := &runtime.ServiceInfo{Service: &dynamic.Service{LoadBalancer: test.service}}
			handler, err := sm.getLoadBalancerServiceHandler(t.Context(), test.serviceName, serviceInfo)

			assert.NoError(t, err)
			assert.NotNil(t, handler)

			req := testhelpers.MustNewRequest(http.MethodGet, "http://callme", nil)
			assert.Empty(t, req.Header.Get("User-Agent"))

			if test.userAgent != "" {
				req.Header.Set("User-Agent", test.userAgent)
			}

			if test.cookieRawValue != "" {
				req.Header.Set("Cookie", test.cookieRawValue)
			}

			var prevXFrom string
			for _, expected := range test.expected {
				recorder := httptest.NewRecorder()

				handler.ServeHTTP(recorder, req)

				assert.Equal(t, expected.StatusCode, recorder.Code)

				if expected.XFrom != "" {
					assert.Equal(t, expected.XFrom, recorder.Header().Get("X-From"))
				}

				xFrom := recorder.Header().Get("X-From")
				if prevXFrom != "" {
					if expected.LoadBalanced {
						assert.NotEqual(t, prevXFrom, xFrom)
					} else {
						assert.Equal(t, prevXFrom, xFrom)
					}
				}
				prevXFrom = xFrom

				cookieHeader := recorder.Header().Get("Set-Cookie")
				if len(cookieHeader) > 0 {
					req.Header.Set("Cookie", cookieHeader)
					assert.Equal(t, expected.SecureCookie, strings.Contains(cookieHeader, "Secure"))
					assert.Equal(t, expected.HTTPOnlyCookie, strings.Contains(cookieHeader, "HttpOnly"))
					assert.NotContains(t, cookieHeader, "://")
				}
			}
		})
	}
}

// This test is an adapted version of net/http/httputil.Test1xxResponses test.
func Test1xxResponses(t *testing.T) {
	pb := httputil.NewProxyBuilder(&transportManagerMock{}, nil)
	sm := NewManager(nil, nil, nil, &transportManagerMock{}, pb)

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Add("Link", "</style.css>; rel=preload; as=style")
		h.Add("Link", "</script.js>; rel=preload; as=script")
		w.WriteHeader(http.StatusEarlyHints)

		h.Add("Link", "</foo.js>; rel=preload; as=script")
		w.WriteHeader(http.StatusProcessing)

		_, _ = w.Write([]byte("Hello"))
	}))
	t.Cleanup(backend.Close)

	info := &runtime.ServiceInfo{
		Service: &dynamic.Service{
			LoadBalancer: &dynamic.ServersLoadBalancer{
				Strategy: dynamic.BalancerStrategyWRR,
				Servers: []dynamic.Server{
					{
						URL: backend.URL,
					},
				},
			},
		},
	}

	handler, err := sm.getLoadBalancerServiceHandler(t.Context(), "foobar", info)
	assert.NoError(t, err)

	frontend := httptest.NewServer(handler)
	t.Cleanup(frontend.Close)
	frontendClient := frontend.Client()

	checkLinkHeaders := func(t *testing.T, expected, got []string) {
		t.Helper()

		if len(expected) != len(got) {
			t.Errorf("Expected %d link headers; got %d", len(expected), len(got))
		}

		for i := range expected {
			if i >= len(got) {
				t.Errorf("Expected %q link header; got nothing", expected[i])

				continue
			}

			if expected[i] != got[i] {
				t.Errorf("Expected %q link header; got %q", expected[i], got[i])
			}
		}
	}

	var respCounter uint8
	trace := &httptrace.ClientTrace{
		Got1xxResponse: func(code int, header textproto.MIMEHeader) error {
			switch code {
			case http.StatusEarlyHints:
				checkLinkHeaders(t, []string{"</style.css>; rel=preload; as=style", "</script.js>; rel=preload; as=script"}, header["Link"])
			case http.StatusProcessing:
				checkLinkHeaders(t, []string{"</style.css>; rel=preload; as=style", "</script.js>; rel=preload; as=script", "</foo.js>; rel=preload; as=script"}, header["Link"])
			default:
				t.Error("Unexpected 1xx response")
			}

			respCounter++

			return nil
		},
	}
	req, _ := http.NewRequestWithContext(httptrace.WithClientTrace(t.Context(), trace), http.MethodGet, frontend.URL, nil)

	res, err := frontendClient.Do(req)
	assert.NoError(t, err)

	defer res.Body.Close()

	if respCounter != 2 {
		t.Errorf("Expected 2 1xx responses; got %d", respCounter)
	}
	checkLinkHeaders(t, []string{"</style.css>; rel=preload; as=style", "</script.js>; rel=preload; as=script", "</foo.js>; rel=preload; as=script"}, res.Header["Link"])

	body, _ := io.ReadAll(res.Body)
	if string(body) != "Hello" {
		t.Errorf("Read body %q; want Hello", body)
	}
}

func TestManager_ServiceBuilders(t *testing.T) {
	var internalHandler internalHandler

	manager := NewManager(map[string]*runtime.ServiceInfo{
		"test@test": {
			Service: &dynamic.Service{
				LoadBalancer: &dynamic.ServersLoadBalancer{
					Strategy: dynamic.BalancerStrategyWRR,
				},
			},
		},
	}, nil, nil, &TransportManager{
		roundTrippers: map[string]http.RoundTripper{
			"default@internal": http.DefaultTransport,
		},
	}, nil, serviceBuilderFunc(func(rootCtx context.Context, serviceName string) (http.Handler, error) {
		if strings.HasSuffix(serviceName, "@internal") {
			return internalHandler, nil
		}
		return nil, nil
	}))

	h, err := manager.BuildHTTP(t.Context(), "test@internal")
	require.NoError(t, err)
	assert.Equal(t, internalHandler, h)

	h, err = manager.BuildHTTP(t.Context(), "test@test")
	require.NoError(t, err)
	assert.NotNil(t, h)

	_, err = manager.BuildHTTP(t.Context(), "wrong@test")
	assert.Error(t, err)
}

func TestManager_Build(t *testing.T) {
	testCases := []struct {
		desc         string
		serviceName  string
		configs      map[string]*runtime.ServiceInfo
		providerName string
	}{
		{
			desc:        "Simple service name",
			serviceName: "serviceName",
			configs: map[string]*runtime.ServiceInfo{
				"serviceName": {
					Service: &dynamic.Service{
						LoadBalancer: &dynamic.ServersLoadBalancer{
							Strategy: dynamic.BalancerStrategyWRR,
						},
					},
				},
			},
		},
		{
			desc:        "Service name with provider",
			serviceName: "serviceName@provider-1",
			configs: map[string]*runtime.ServiceInfo{
				"serviceName@provider-1": {
					Service: &dynamic.Service{
						LoadBalancer: &dynamic.ServersLoadBalancer{
							Strategy: dynamic.BalancerStrategyWRR,
						},
					},
				},
			},
		},
		{
			desc:        "Service name with provider in context",
			serviceName: "serviceName",
			configs: map[string]*runtime.ServiceInfo{
				"serviceName@provider-1": {
					Service: &dynamic.Service{
						LoadBalancer: &dynamic.ServersLoadBalancer{
							Strategy: dynamic.BalancerStrategyWRR,
						},
					},
				},
			},
			providerName: "provider-1",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			manager := NewManager(test.configs, nil, nil, &transportManagerMock{}, nil)

			ctx := t.Context()
			if len(test.providerName) > 0 {
				ctx = provider.AddInContext(ctx, "foobar@"+test.providerName)
			}

			_, err := manager.BuildHTTP(ctx, test.serviceName)
			require.NoError(t, err)
		})
	}
}

func TestMultipleTypeOnBuildHTTP(t *testing.T) {
	services := map[string]*runtime.ServiceInfo{
		"test@file": {
			Service: &dynamic.Service{
				LoadBalancer: &dynamic.ServersLoadBalancer{},
				Weighted:     &dynamic.WeightedRoundRobin{},
			},
		},
	}

	manager := NewManager(services, nil, nil, &transportManagerMock{}, nil)

	_, err := manager.BuildHTTP(t.Context(), "test@file")
	assert.Error(t, err, "cannot create service: multi-types service not supported, consider declaring two different pieces of service instead")
}

func TestGetServiceHandler_Headers(t *testing.T) {
	pb := httputil.NewProxyBuilder(&transportManagerMock{}, nil)

	testCases := []struct {
		desc            string
		service         dynamic.WRRService
		userAgent       string
		expectedHeaders map[string]string
	}{
		{
			desc: "Service with custom headers",
			service: dynamic.WRRService{
				Name: "target-service",
				Headers: map[string]string{
					"X-Custom-Header": "custom-value",
					"X-Service-Type":  "knative-service",
					"Authorization":   "bearer token123",
				},
			},
			userAgent: "test-agent",
			expectedHeaders: map[string]string{
				"X-Custom-Header": "custom-value",
				"X-Service-Type":  "knative-service",
				"Authorization":   "bearer token123",
			},
		},
		{
			desc: "Service with empty headers map",
			service: dynamic.WRRService{
				Name:    "target-service",
				Headers: map[string]string{},
			},
			userAgent:       "test-agent",
			expectedHeaders: map[string]string{},
		},
		{
			desc: "Service with nil headers",
			service: dynamic.WRRService{
				Name:    "target-service",
				Headers: nil,
			},
			userAgent:       "test-agent",
			expectedHeaders: map[string]string{},
		},
		{
			desc: "Service with headers that override existing request headers",
			service: dynamic.WRRService{
				Name: "target-service",
				Headers: map[string]string{
					"User-Agent": "overridden-agent",
					"Accept":     "application/json",
				},
			},
			userAgent: "original-agent",
			expectedHeaders: map[string]string{
				"User-Agent": "overridden-agent",
				"Accept":     "application/json",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			// Create a test server that will verify the headers are properly set for this specific test case
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify expected headers are present
				for key, expectedValue := range test.expectedHeaders {
					actualValue := r.Header.Get(key)
					assert.Equal(t, expectedValue, actualValue, "Header %s should be %s", key, expectedValue)
				}

				w.Header().Set("X-Response", "success")
				w.WriteHeader(http.StatusOK)
			}))
			t.Cleanup(testServer.Close)

			// Create the target service that the WRRService will point to
			targetServiceInfo := &runtime.ServiceInfo{
				Service: &dynamic.Service{
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Strategy: dynamic.BalancerStrategyWRR,
						Servers: []dynamic.Server{
							{URL: testServer.URL},
						},
					},
				},
			}

			// Create a fresh manager for each test case
			sm := NewManager(map[string]*runtime.ServiceInfo{
				"target-service": targetServiceInfo,
			}, nil, nil, &transportManagerMock{}, pb)

			// Get the service handler
			handler, err := sm.getServiceHandler(t.Context(), test.service)
			require.NoError(t, err)
			require.NotNil(t, handler)

			// Create a test request
			req := testhelpers.MustNewRequest(http.MethodGet, "http://test.example.com/path", nil)
			if test.userAgent != "" {
				req.Header.Set("User-Agent", test.userAgent)
			}

			// Execute the request
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			// Verify the response was successful
			assert.Equal(t, http.StatusOK, recorder.Code)
		})
	}
}

type serviceBuilderFunc func(ctx context.Context, serviceName string) (http.Handler, error)

func (s serviceBuilderFunc) BuildHTTP(ctx context.Context, serviceName string) (http.Handler, error) {
	return s(ctx, serviceName)
}

func TestGetServiceHandler_HealthCheck(t *testing.T) {
	pb := httputil.NewProxyBuilder(&transportManagerMock{}, nil)

	testCases := []struct {
		desc                       string
		withServiceMiddleware      bool
		withChildServiceMiddleware bool
	}{
		{
			desc: "without middleware",
		},
		{
			desc:                  "with service middleware",
			withServiceMiddleware: true,
		},
		{
			desc:                       "with child service middleware",
			withChildServiceMiddleware: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			t.Cleanup(backend.Close)

			childSvc := &dynamic.Service{
				LoadBalancer: &dynamic.ServersLoadBalancer{
					Servers: []dynamic.Server{{URL: backend.URL}},
					HealthCheck: &dynamic.ServerHealthCheck{
						Path: "/health",
					},
				},
			}
			if test.withServiceMiddleware {
				childSvc.Middlewares = []string{"add-header@file"}
			}

			wrrChild := dynamic.WRRService{Name: "child@file", Weight: pointer(1)}
			if test.withChildServiceMiddleware {
				wrrChild.Middlewares = []string{"add-header@file"}
			}

			configs := map[string]*runtime.ServiceInfo{
				"child@file": {Service: childSvc},
				"wrr@file": {
					Service: &dynamic.Service{
						Weighted: &dynamic.WeightedRoundRobin{
							Services:    []dynamic.WRRService{wrrChild},
							HealthCheck: &dynamic.HealthCheck{},
						},
					},
				},
			}

			manager := NewManager(configs, nil, nil, &transportManagerMock{}, pb)
			if test.withServiceMiddleware || test.withChildServiceMiddleware {
				manager.SetMiddlewareChainBuilder(&noopMiddlewareChainBuilder{})
			}

			_, err := manager.BuildHTTP(t.Context(), "wrr@file")
			require.NoError(t, err)
		})
	}
}

// noopMiddlewareChainBuilder wraps a handler in a plain http.HandlerFunc,
// simulating the effect of service-level middlewares without needing real middleware config.
type noopMiddlewareChainBuilder struct{}

func (n *noopMiddlewareChainBuilder) BuildMiddlewareChain(_ context.Context, _ []string) *alice.Chain {
	chain := alice.New(func(next http.Handler) (http.Handler, error) {
		return http.HandlerFunc(next.ServeHTTP), nil
	})
	return &chain
}

// requestHeaderMiddlewareChainBuilder sets X-Middleware on the *request* per
// middleware name, so the upstream backend can observe which chain ran.
type requestHeaderMiddlewareChainBuilder struct{}

func (r *requestHeaderMiddlewareChainBuilder) BuildMiddlewareChain(_ context.Context, names []string) *alice.Chain {
	chain := alice.New()
	for _, name := range names {
		chain = chain.Append(func(next http.Handler) (http.Handler, error) {
			return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				req.Header.Set("X-Middleware", name)
				next.ServeHTTP(rw, req)
			}), nil
		})
	}
	return &chain
}

func TestGetWRRServiceHandler_WithChildServiceMiddleware(t *testing.T) {
	newBackend := func(expected string) *httptest.Server {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, expected, r.Header.Get("X-Middleware"))
			w.WriteHeader(http.StatusOK)
		}))
		t.Cleanup(srv.Close)
		return srv
	}

	backendA := newBackend("mw-a")
	backendB := newBackend("mw-b")

	configs := map[string]*runtime.ServiceInfo{
		"child-a@file": {Service: &dynamic.Service{LoadBalancer: &dynamic.ServersLoadBalancer{Servers: []dynamic.Server{{URL: backendA.URL}}}}},
		"child-b@file": {Service: &dynamic.Service{LoadBalancer: &dynamic.ServersLoadBalancer{Servers: []dynamic.Server{{URL: backendB.URL}}}}},
		"wrr@file": {
			Service: &dynamic.Service{
				Weighted: &dynamic.WeightedRoundRobin{
					Services: []dynamic.WRRService{
						{Name: "child-a@file", Weight: pointer(1), Middlewares: []string{"mw-a"}},
						{Name: "child-b@file", Weight: pointer(1), Middlewares: []string{"mw-b"}},
					},
				},
			},
		},
	}

	pb := httputil.NewProxyBuilder(&transportManagerMock{}, nil)
	manager := NewManager(configs, nil, nil, &transportManagerMock{}, pb)
	manager.SetMiddlewareChainBuilder(&requestHeaderMiddlewareChainBuilder{})

	handler, err := manager.BuildHTTP(t.Context(), "wrr@file")
	require.NoError(t, err)

	// 8 calls makes missing a child statistically negligible (2 * 0.5^8 ≈ 0.78%).
	for range 8 {
		req := testhelpers.MustNewRequest(http.MethodGet, "http://test.example.com/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestGetHRWServiceHandler_WithChildServiceMiddleware(t *testing.T) {
	newBackend := func(expected string) *httptest.Server {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, expected, r.Header.Get("X-Middleware"))
			w.WriteHeader(http.StatusOK)
		}))
		t.Cleanup(srv.Close)
		return srv
	}

	backendA := newBackend("mw-a")
	backendB := newBackend("mw-b")

	configs := map[string]*runtime.ServiceInfo{
		"child-a@file": {Service: &dynamic.Service{LoadBalancer: &dynamic.ServersLoadBalancer{Servers: []dynamic.Server{{URL: backendA.URL}}}}},
		"child-b@file": {Service: &dynamic.Service{LoadBalancer: &dynamic.ServersLoadBalancer{Servers: []dynamic.Server{{URL: backendB.URL}}}}},
		"hrw@file": {
			Service: &dynamic.Service{
				HighestRandomWeight: &dynamic.HighestRandomWeight{
					Services: []dynamic.HRWService{
						{Name: "child-a@file", Weight: pointer(1), Middlewares: []string{"mw-a"}},
						{Name: "child-b@file", Weight: pointer(1), Middlewares: []string{"mw-b"}},
					},
				},
			},
		},
	}

	pb := httputil.NewProxyBuilder(&transportManagerMock{}, nil)
	manager := NewManager(configs, nil, nil, &transportManagerMock{}, pb)
	manager.SetMiddlewareChainBuilder(&requestHeaderMiddlewareChainBuilder{})

	handler, err := manager.BuildHTTP(t.Context(), "hrw@file")
	require.NoError(t, err)

	// HRW is deterministic per client IP; varying RemoteAddr across calls covers both children.
	for i := range 16 {
		req := testhelpers.MustNewRequest(http.MethodGet, "http://test.example.com/", nil)
		req.RemoteAddr = fmt.Sprintf("10.0.0.%d:1234", i+1)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestGetMirrorServiceHandler_WithChildServiceMiddleware(t *testing.T) {
	newBackend := func(expected string) *httptest.Server {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, expected, r.Header.Get("X-Middleware"))
			w.WriteHeader(http.StatusOK)
		}))
		t.Cleanup(srv.Close)
		return srv
	}

	mainBackend := newBackend("mw-main")
	mirrorBackend := newBackend("mw-mirror")

	configs := map[string]*runtime.ServiceInfo{
		"main@file":   {Service: &dynamic.Service{LoadBalancer: &dynamic.ServersLoadBalancer{Servers: []dynamic.Server{{URL: mainBackend.URL}}}}},
		"mirror@file": {Service: &dynamic.Service{LoadBalancer: &dynamic.ServersLoadBalancer{Servers: []dynamic.Server{{URL: mirrorBackend.URL}}}}},
		"mirroring@file": {
			Service: &dynamic.Service{
				Mirroring: &dynamic.Mirroring{
					Service:     "main@file",
					Middlewares: []string{"mw-main"},
					Mirrors: []dynamic.MirrorService{
						{Name: "mirror@file", Percent: 100, Middlewares: []string{"mw-mirror"}},
					},
				},
			},
		},
	}

	pool := safe.NewPool(t.Context())
	pb := httputil.NewProxyBuilder(&transportManagerMock{}, nil)
	manager := NewManager(configs, nil, pool, &transportManagerMock{}, pb)
	manager.SetMiddlewareChainBuilder(&requestHeaderMiddlewareChainBuilder{})

	handler, err := manager.BuildHTTP(t.Context(), "mirroring@file")
	require.NoError(t, err)

	req := testhelpers.MustNewRequest(http.MethodGet, "http://test.example.com/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	// Wait for the mirror goroutine to run its request through the mirror backend.
	pool.Stop()
}

func TestGetFailoverServiceHandler_WithChildServiceMiddleware(t *testing.T) {
	var mainStatus atomic.Int32
	mainStatus.Store(http.StatusOK)

	var gotService, gotFallback string
	serviceBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotService = r.Header.Get("X-Middleware")
		w.WriteHeader(int(mainStatus.Load()))
	}))
	t.Cleanup(serviceBackend.Close)

	fallbackBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotFallback = r.Header.Get("X-Middleware")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(fallbackBackend.Close)

	configs := map[string]*runtime.ServiceInfo{
		"service@file":  {Service: &dynamic.Service{LoadBalancer: &dynamic.ServersLoadBalancer{Servers: []dynamic.Server{{URL: serviceBackend.URL}}}}},
		"fallback@file": {Service: &dynamic.Service{LoadBalancer: &dynamic.ServersLoadBalancer{Servers: []dynamic.Server{{URL: fallbackBackend.URL}}}}},
		"failover@file": {
			Service: &dynamic.Service{
				Failover: &dynamic.Failover{
					Service:             "service@file",
					Middlewares:         []string{"mw-service"},
					Fallback:            "fallback@file",
					FallbackMiddlewares: []string{"mw-fallback"},
					Errors: &dynamic.FailoverError{
						Status: []string{"500-599"},
					},
				},
			},
		},
	}

	pb := httputil.NewProxyBuilder(&transportManagerMock{}, nil)
	manager := NewManager(configs, nil, nil, &transportManagerMock{}, pb)
	manager.SetMiddlewareChainBuilder(&requestHeaderMiddlewareChainBuilder{})

	handler, err := manager.BuildHTTP(t.Context(), "failover@file")
	require.NoError(t, err)

	// Main service healthy: the service-edge middleware is applied.
	req := testhelpers.MustNewRequest(http.MethodGet, "http://test.example.com/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "mw-service", gotService)

	// Main service fails: failover triggers on the configured status range, fallback-edge middleware is applied.
	mainStatus.Store(http.StatusInternalServerError)
	req = testhelpers.MustNewRequest(http.MethodGet, "http://test.example.com/", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "mw-fallback", gotFallback)
}

type internalHandler struct{}

func (internalHandler) ServeHTTP(_ http.ResponseWriter, _ *http.Request) {}

type forwarderMock struct{}

func (forwarderMock) ServeHTTP(http.ResponseWriter, *http.Request) {
	panic("not available")
}

type transportManagerMock struct{}

func (t transportManagerMock) GetRoundTripper(_ string) (http.RoundTripper, error) {
	return &http.Transport{}, nil
}

func (t transportManagerMock) GetTLSConfig(_ string) (*tls.Config, error) {
	return nil, nil
}

func (t transportManagerMock) Get(_ string) (*dynamic.ServersTransport, error) {
	return &dynamic.ServersTransport{}, nil
}
