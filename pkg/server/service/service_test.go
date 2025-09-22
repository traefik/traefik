package service

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"net/textproto"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/proxy/httputil"
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
				PassHostHeader: boolPtr(true),
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

type serviceBuilderFunc func(ctx context.Context, serviceName string) (http.Handler, error)

func (s serviceBuilderFunc) BuildHTTP(ctx context.Context, serviceName string) (http.Handler, error) {
	return s(ctx, serviceName)
}

type internalHandler struct{}

func (internalHandler) ServeHTTP(_ http.ResponseWriter, _ *http.Request) {}

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

func boolPtr(v bool) *bool { return &v }

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
