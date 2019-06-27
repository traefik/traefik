package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/server/internal"
	"github.com/containous/traefik/pkg/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockForwarder struct{}

func (MockForwarder) ServeHTTP(http.ResponseWriter, *http.Request) {
	panic("implement me")
}

func TestGetLoadBalancer(t *testing.T) {
	sm := Manager{}

	testCases := []struct {
		desc        string
		serviceName string
		service     *config.LoadBalancerService
		fwd         http.Handler
		expectError bool
	}{
		{
			desc:        "Fails when provided an invalid URL",
			serviceName: "test",
			service: &config.LoadBalancerService{
				Servers: []config.Server{
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
			service:     &config.LoadBalancerService{},
			fwd:         &MockForwarder{},
			expectError: false,
		},
		{
			desc:        "Succeeds when stickiness is set",
			serviceName: "test",
			service: &config.LoadBalancerService{
				Stickiness: &config.Stickiness{},
			},
			fwd:         &MockForwarder{},
			expectError: false,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			handler, err := sm.getLoadBalancer(context.Background(), test.serviceName, test.service, test.fwd)
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
	sm := NewManager(nil, http.DefaultTransport)

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
		SecureCookie   bool
		HTTPOnlyCookie bool
	}

	testCases := []struct {
		desc             string
		serviceName      string
		service          *config.LoadBalancerService
		responseModifier func(*http.Response) error

		expected []ExpectedResult
	}{
		{
			desc:        "Load balances between the two servers",
			serviceName: "test",
			service: &config.LoadBalancerService{
				Servers: []config.Server{
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
					XFrom:      "first",
				},
				{
					StatusCode: http.StatusOK,
					XFrom:      "second",
				},
			},
		},
		{
			desc:        "StatusBadGateway when the server is not reachable",
			serviceName: "test",
			service: &config.LoadBalancerService{
				Servers: []config.Server{
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
			service: &config.LoadBalancerService{
				Servers: []config.Server{},
			},
			expected: []ExpectedResult{
				{
					StatusCode: http.StatusServiceUnavailable,
				},
			},
		},
		{
			desc:        "Always call the same server when stickiness is true",
			serviceName: "test",
			service: &config.LoadBalancerService{
				Stickiness: &config.Stickiness{},
				Servers: []config.Server{
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
					XFrom:      "first",
				},
				{
					StatusCode: http.StatusOK,
					XFrom:      "first",
				},
			},
		},
		{
			desc:        "Sticky Cookie's options set correctly",
			serviceName: "test",
			service: &config.LoadBalancerService{
				Stickiness: &config.Stickiness{HTTPOnlyCookie: true, SecureCookie: true},
				Servers: []config.Server{
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
			service: &config.LoadBalancerService{
				Stickiness:     &config.Stickiness{},
				PassHostHeader: true,
				Servers: []config.Server{
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
			desc:        "PassHost doesn't passe the host instead of the IP",
			serviceName: "test",
			service: &config.LoadBalancerService{
				Stickiness: &config.Stickiness{},
				Servers: []config.Server{
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

			handler, err := sm.getLoadBalancerServiceHandler(context.Background(), test.serviceName, test.service, test.responseModifier)

			assert.NoError(t, err)
			assert.NotNil(t, handler)

			req := testhelpers.MustNewRequest(http.MethodGet, "http://callme", nil)
			for _, expected := range test.expected {
				recorder := httptest.NewRecorder()

				handler.ServeHTTP(recorder, req)

				assert.Equal(t, expected.StatusCode, recorder.Code)
				assert.Equal(t, expected.XFrom, recorder.Header().Get("X-From"))

				cookieHeader := recorder.Header().Get("Set-Cookie")
				if len(cookieHeader) > 0 {
					req.Header.Set("Cookie", cookieHeader)
					assert.Equal(t, expected.SecureCookie, strings.Contains(cookieHeader, "Secure"))
					assert.Equal(t, expected.HTTPOnlyCookie, strings.Contains(cookieHeader, "HttpOnly"))
				}
			}
		})
	}
}

func TestManager_Build(t *testing.T) {
	testCases := []struct {
		desc         string
		serviceName  string
		configs      map[string]*config.ServiceInfo
		providerName string
	}{
		{
			desc:        "Simple service name",
			serviceName: "serviceName",
			configs: map[string]*config.ServiceInfo{
				"serviceName": {
					Service: &config.Service{
						LoadBalancer: &config.LoadBalancerService{},
					},
				},
			},
		},
		{
			desc:        "Service name with provider",
			serviceName: "serviceName@provider-1",
			configs: map[string]*config.ServiceInfo{
				"serviceName@provider-1": {
					Service: &config.Service{
						LoadBalancer: &config.LoadBalancerService{},
					},
				},
			},
		},
		{
			desc:        "Service name with provider in context",
			serviceName: "serviceName",
			configs: map[string]*config.ServiceInfo{
				"serviceName@provider-1": {
					Service: &config.Service{
						LoadBalancer: &config.LoadBalancerService{},
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

			manager := NewManager(test.configs, http.DefaultTransport)

			ctx := context.Background()
			if len(test.providerName) > 0 {
				ctx = internal.AddProviderInContext(ctx, "foobar@"+test.providerName)
			}

			_, err := manager.BuildHTTP(ctx, test.serviceName, nil)
			require.NoError(t, err)
		})
	}
}

// FIXME Add healthcheck tests
