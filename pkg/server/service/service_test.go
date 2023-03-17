package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"net/textproto"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/config/runtime"
	"github.com/traefik/traefik/v2/pkg/server/provider"
	"github.com/traefik/traefik/v2/pkg/testhelpers"
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
	sm := NewManager(nil, nil, nil, &RoundTripperManager{
		roundTrippers: map[string]http.RoundTripper{
			"default@internal": http.DefaultTransport,
		},
	})

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
		{
			desc:        "Cookie value is backward compatible",
			serviceName: "test",
			service: &dynamic.ServersLoadBalancer{
				Sticky: &dynamic.Sticky{
					Cookie: &dynamic.Cookie{},
				},
				Servers: []dynamic.Server{
					{
						URL: server1.URL,
					},
					{
						URL: server2.URL,
					},
				},
			},
			cookieRawValue: "_6f743=" + server1.URL,
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
			desc:        "No user-agent",
			serviceName: "test",
			service: &dynamic.ServersLoadBalancer{
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
		test := test
		t.Run(test.desc, func(t *testing.T) {
			handler, err := sm.getLoadBalancerServiceHandler(context.Background(), test.serviceName, test.service)

			assert.NoError(t, err)
			assert.NotNil(t, handler)

			req := testhelpers.MustNewRequest(http.MethodGet, "http://callme", nil)
			assert.Equal(t, "", req.Header.Get("User-Agent"))

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
	sm := NewManager(nil, nil, nil, &RoundTripperManager{
		roundTrippers: map[string]http.RoundTripper{
			"default@internal": http.DefaultTransport,
		},
	})

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

	config := &dynamic.ServersLoadBalancer{
		Servers: []dynamic.Server{
			{
				URL: backend.URL,
			},
		},
	}
	handler, err := sm.getLoadBalancerServiceHandler(context.Background(), "foobar", config)
	assert.Nil(t, err)

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
	req, _ := http.NewRequestWithContext(httptrace.WithClientTrace(context.Background(), trace), http.MethodGet, frontend.URL, nil)

	res, err := frontendClient.Do(req)
	assert.Nil(t, err)

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
