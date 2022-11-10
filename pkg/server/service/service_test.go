package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/config/runtime"
	"github.com/traefik/traefik/v2/pkg/server/provider"
	"github.com/traefik/traefik/v2/pkg/testhelpers"
)

func TestGetLoadBalancer(t *testing.T) {
	sm := Manager{
		roundTripperManager: newRtMock(),
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
				Servers: []dynamic.Server{
					{
						URL: ":",
					},
				},
			},
			fwd:         &MockForwarder{},
			expectError: true,
		},
		{
			desc:        "Succeeds when there are no servers",
			serviceName: "test",
			service:     &dynamic.ServersLoadBalancer{},
			fwd:         &MockForwarder{},
			expectError: false,
		},
		{
			desc:        "Succeeds when sticky.cookie is set",
			serviceName: "test",
			service: &dynamic.ServersLoadBalancer{
				Sticky: &dynamic.Sticky{Cookie: &dynamic.Cookie{}},
			},
			fwd:         &MockForwarder{},
			expectError: false,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			serviceInfo := &runtime.ServiceInfo{Service: &dynamic.Service{LoadBalancer: test.service}}
			handler, err := sm.getLoadBalancerServiceHandler(context.Background(), test.serviceName, serviceInfo)
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
	sm := NewManager(nil, nil, nil, &RoundTripperManager{
		roundTrippers: map[string]http.RoundTripper{
			"default@internal": http.DefaultTransport,
		},
	})

	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-From", "first")
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-From", "second")
	}))
	defer server2.Close()

	serverPassHost := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-From", "passhost")
		assert.Equal(t, "callme", r.Host)
	}))
	defer serverPassHost.Close()

	serverPassHostFalse := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-From", "passhostfalse")
		assert.NotEqual(t, "callme", r.Host)
	}))
	defer serverPassHostFalse.Close()

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

		expected []ExpectedResult
	}{
		{
			desc:        "Load balances between the two servers",
			serviceName: "test",
			service: &dynamic.ServersLoadBalancer{
				PassHostHeader: Bool(true),
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
				Servers: []dynamic.Server{},
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
				Sticky: &dynamic.Sticky{Cookie: &dynamic.Cookie{}},
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
				Sticky: &dynamic.Sticky{Cookie: &dynamic.Cookie{HTTPOnly: true, Secure: true}},
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
				Sticky:         &dynamic.Sticky{Cookie: &dynamic.Cookie{}},
				PassHostHeader: func(v bool) *bool { return &v }(true),
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
				PassHostHeader: Bool(false),
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
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			serviceInfo := &runtime.ServiceInfo{Service: &dynamic.Service{LoadBalancer: test.service}}
			handler, err := sm.getLoadBalancerServiceHandler(context.Background(), test.serviceName, serviceInfo)

			assert.NoError(t, err)
			assert.NotNil(t, handler)

			req := testhelpers.MustNewRequest(http.MethodGet, "http://callme", nil)
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
						LoadBalancer: &dynamic.ServersLoadBalancer{},
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
						LoadBalancer: &dynamic.ServersLoadBalancer{},
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
						LoadBalancer: &dynamic.ServersLoadBalancer{},
					},
				},
			},
			providerName: "provider-1",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			manager := NewManager(test.configs, nil, nil, &RoundTripperManager{
				roundTrippers: map[string]http.RoundTripper{
					"default@internal": http.DefaultTransport,
				},
			})

			ctx := context.Background()
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

	manager := NewManager(services, nil, nil, &RoundTripperManager{
		roundTrippers: map[string]http.RoundTripper{
			"default@internal": http.DefaultTransport,
		},
	})

	_, err := manager.BuildHTTP(context.Background(), "test@file")
	assert.Error(t, err, "cannot create service: multi-types service not supported, consider declaring two different pieces of service instead")
}

func Bool(v bool) *bool { return &v }

type MockForwarder struct{}

func (MockForwarder) ServeHTTP(http.ResponseWriter, *http.Request) {
	panic("not available")
}

type rtMock struct{}

func newRtMock() RoundTripperGetter {
	return &rtMock{}
}

func (r *rtMock) Get(_ string) (http.RoundTripper, error) {
	return http.DefaultTransport, nil
}
