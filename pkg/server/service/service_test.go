package service

import (
	"context"
	"net/http"
	"net/http/httptest"
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
						URL:    ":",
						Weight: 0,
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
		StatusCode int
		XFrom      string
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
						URL:    server1.URL,
						Weight: 50,
					},
					{
						URL:    server2.URL,
						Weight: 50,
					},
				},
				Method: "wrr",
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
						URL:    "http://foo",
						Weight: 1,
					},
				},
				Method: "wrr",
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
				Method:  "wrr",
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
						URL:    server1.URL,
						Weight: 1,
					},
					{
						URL:    server2.URL,
						Weight: 1,
					},
				},
				Method: "wrr",
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
			desc:        "PassHost passes the host instead of the IP",
			serviceName: "test",
			service: &config.LoadBalancerService{
				Stickiness:     &config.Stickiness{},
				PassHostHeader: true,
				Servers: []config.Server{
					{
						URL:    serverPassHost.URL,
						Weight: 1,
					},
				},
				Method: "wrr",
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
						URL:    serverPassHostFalse.URL,
						Weight: 1,
					},
				},
				Method: "wrr",
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

				if len(recorder.Header().Get("Set-Cookie")) > 0 {
					req.Header.Set("Cookie", recorder.Header().Get("Set-Cookie"))
				}
			}
		})
	}
}

func TestManager_Build(t *testing.T) {
	testCases := []struct {
		desc         string
		serviceName  string
		configs      map[string]*config.Service
		providerName string
	}{
		{
			desc:        "Simple service name",
			serviceName: "serviceName",
			configs: map[string]*config.Service{
				"serviceName": {
					LoadBalancer: &config.LoadBalancerService{Method: "wrr"},
				},
			},
		},
		{
			desc:        "Service name with provider",
			serviceName: "provider-1.serviceName",
			configs: map[string]*config.Service{
				"provider-1.serviceName": {
					LoadBalancer: &config.LoadBalancerService{Method: "wrr"},
				},
			},
		},
		{
			desc:        "Service name with provider in context",
			serviceName: "serviceName",
			configs: map[string]*config.Service{
				"provider-1.serviceName": {
					LoadBalancer: &config.LoadBalancerService{Method: "wrr"},
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
				ctx = internal.AddProviderInContext(ctx, test.providerName+".foobar")
			}

			_, err := manager.BuildHTTP(ctx, test.serviceName, nil)
			require.NoError(t, err)
		})
	}
}

// FIXME Add healthcheck tests
