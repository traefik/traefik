package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/containous/flaeg"
	"github.com/containous/mux"
	"github.com/containous/traefik/configuration"
	"github.com/containous/traefik/healthcheck"
	"github.com/containous/traefik/middlewares"
	"github.com/containous/traefik/rules"
	th "github.com/containous/traefik/testhelpers"
	"github.com/containous/traefik/tls"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vulcand/oxy/roundrobin"
)

// LocalhostCert is a PEM-encoded TLS cert with SAN IPs
// "127.0.0.1" and "[::1]", expiring at Jan 29 16:00:00 2084 GMT.
// generated from src/crypto/tls:
// go run generate_cert.go  --rsa-bits 1024 --host 127.0.0.1,::1,example.com --ca --start-date "Jan 1 00:00:00 1970" --duration=1000000h
var (
	localhostCert = tls.FileOrContent(`-----BEGIN CERTIFICATE-----
MIICEzCCAXygAwIBAgIQMIMChMLGrR+QvmQvpwAU6zANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQKEwdBY21lIENvMCAXDTcwMDEwMTAwMDAwMFoYDzIwODQwMTI5MTYw
MDAwWjASMRAwDgYDVQQKEwdBY21lIENvMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCB
iQKBgQDuLnQAI3mDgey3VBzWnB2L39JUU4txjeVE6myuDqkM/uGlfjb9SjY1bIw4
iA5sBBZzHi3z0h1YV8QPuxEbi4nW91IJm2gsvvZhIrCHS3l6afab4pZBl2+XsDul
rKBxKKtD1rGxlG4LjncdabFn9gvLZad2bSysqz/qTAUStTvqJQIDAQABo2gwZjAO
BgNVHQ8BAf8EBAMCAqQwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDwYDVR0TAQH/BAUw
AwEB/zAuBgNVHREEJzAlggtleGFtcGxlLmNvbYcEfwAAAYcQAAAAAAAAAAAAAAAA
AAAAATANBgkqhkiG9w0BAQsFAAOBgQCEcetwO59EWk7WiJsG4x8SY+UIAA+flUI9
tyC4lNhbcF2Idq9greZwbYCqTTTr2XiRNSMLCOjKyI7ukPoPjo16ocHj+P3vZGfs
h1fIw3cSS2OolhloGw/XM6RWPWtPAlGykKLciQrBru5NAPvCMsb/I1DAceTiotQM
fblo6RBxUQ==
-----END CERTIFICATE-----`)

	// LocalhostKey is the private key for localhostCert.
	localhostKey = tls.FileOrContent(`-----BEGIN RSA PRIVATE KEY-----
MIICXgIBAAKBgQDuLnQAI3mDgey3VBzWnB2L39JUU4txjeVE6myuDqkM/uGlfjb9
SjY1bIw4iA5sBBZzHi3z0h1YV8QPuxEbi4nW91IJm2gsvvZhIrCHS3l6afab4pZB
l2+XsDulrKBxKKtD1rGxlG4LjncdabFn9gvLZad2bSysqz/qTAUStTvqJQIDAQAB
AoGAGRzwwir7XvBOAy5tM/uV6e+Zf6anZzus1s1Y1ClbjbE6HXbnWWF/wbZGOpet
3Zm4vD6MXc7jpTLryzTQIvVdfQbRc6+MUVeLKwZatTXtdZrhu+Jk7hx0nTPy8Jcb
uJqFk541aEw+mMogY/xEcfbWd6IOkp+4xqjlFLBEDytgbIECQQDvH/E6nk+hgN4H
qzzVtxxr397vWrjrIgPbJpQvBsafG7b0dA4AFjwVbFLmQcj2PprIMmPcQrooz8vp
jy4SHEg1AkEA/v13/5M47K9vCxmb8QeD/asydfsgS5TeuNi8DoUBEmiSJwma7FXY
fFUtxuvL7XvjwjN5B30pNEbc6Iuyt7y4MQJBAIt21su4b3sjXNueLKH85Q+phy2U
fQtuUE9txblTu14q3N7gHRZB4ZMhFYyDy8CKrN2cPg/Fvyt0Xlp/DoCzjA0CQQDU
y2ptGsuSmgUtWj3NM9xuwYPm+Z/F84K6+ARYiZ6PYj013sovGKUFfYAqVXVlxtIX
qyUBnu3X9ps8ZfjLZO7BAkEAlT4R5Yl6cGhaJQYZHOde3JEMhNRcVFMO8dJDaFeo
f9Oeos0UUothgiDktdQHxdNEwLjQf7lJJBzV+5OtwswCWA==
-----END RSA PRIVATE KEY-----`)
)

type testLoadBalancer struct{}

func (lb *testLoadBalancer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// noop
}

func (lb *testLoadBalancer) RemoveServer(u *url.URL) error {
	return nil
}

func (lb *testLoadBalancer) UpsertServer(u *url.URL, options ...roundrobin.ServerOption) error {
	return nil
}

func (lb *testLoadBalancer) Servers() []*url.URL {
	return []*url.URL{}
}

func TestServerLoadConfigHealthCheckOptions(t *testing.T) {
	healthChecks := []*types.HealthCheck{
		nil,
		{
			Path: "/path",
		},
	}

	for _, lbMethod := range []string{"Wrr", "Drr"} {
		for _, healthCheck := range healthChecks {
			t.Run(fmt.Sprintf("%s/hc=%t", lbMethod, healthCheck != nil), func(t *testing.T) {
				globalConfig := configuration.GlobalConfiguration{
					HealthCheck: &configuration.HealthCheckConfig{Interval: flaeg.Duration(5 * time.Second)},
				}
				entryPoints := map[string]EntryPoint{
					"http": {
						Configuration: &configuration.EntryPoint{
							ForwardedHeaders: &configuration.ForwardedHeaders{Insecure: true},
						},
					},
				}

				dynamicConfigs := types.Configurations{
					"config": &types.Configuration{
						Frontends: map[string]*types.Frontend{
							"frontend": {
								EntryPoints: []string{"http"},
								Backend:     "backend",
							},
						},
						Backends: map[string]*types.Backend{
							"backend": {
								Servers: map[string]types.Server{
									"server": {
										URL: "http://localhost",
									},
								},
								LoadBalancer: &types.LoadBalancer{
									Method: lbMethod,
								},
								HealthCheck: healthCheck,
							},
						},
						TLS: []*tls.Configuration{
							{
								Certificate: &tls.Certificate{
									CertFile: localhostCert,
									KeyFile:  localhostKey,
								},
								EntryPoints: []string{"http"},
							},
						},
					},
				}

				srv := NewServer(globalConfig, nil, entryPoints)

				_ = srv.loadConfig(dynamicConfigs, globalConfig)

				expectedNumHealthCheckBackends := 0
				if healthCheck != nil {
					expectedNumHealthCheckBackends = 1
				}
				assert.Len(t, healthcheck.GetHealthCheck(th.NewCollectingHealthCheckMetrics()).Backends, expectedNumHealthCheckBackends, "health check backends")
			})
		}
	}
}

func TestServerLoadConfigEmptyBasicAuth(t *testing.T) {
	globalConfig := configuration.GlobalConfiguration{
		EntryPoints: configuration.EntryPoints{
			"http": &configuration.EntryPoint{ForwardedHeaders: &configuration.ForwardedHeaders{Insecure: true}},
		},
	}

	dynamicConfigs := types.Configurations{
		"config": &types.Configuration{
			Frontends: map[string]*types.Frontend{
				"frontend": {
					EntryPoints: []string{"http"},
					Backend:     "backend",
					BasicAuth:   []string{""},
				},
			},
			Backends: map[string]*types.Backend{
				"backend": {
					Servers: map[string]types.Server{
						"server": {
							URL: "http://localhost",
						},
					},
					LoadBalancer: &types.LoadBalancer{
						Method: "Wrr",
					},
				},
			},
		},
	}

	entryPoints := map[string]EntryPoint{}
	for key, value := range globalConfig.EntryPoints {
		entryPoints[key] = EntryPoint{
			Configuration: value,
		}
	}

	srv := NewServer(globalConfig, nil, entryPoints)
	_ = srv.loadConfig(dynamicConfigs, globalConfig)
}

func TestServerLoadCertificateWithDefaultEntryPoint(t *testing.T) {
	globalConfig := configuration.GlobalConfiguration{
		DefaultEntryPoints: []string{"http", "https"},
	}
	entryPoints := map[string]EntryPoint{
		"https": {Configuration: &configuration.EntryPoint{TLS: &tls.TLS{}}},
		"http":  {Configuration: &configuration.EntryPoint{}},
	}

	dynamicConfigs := types.Configurations{
		"config": &types.Configuration{
			TLS: []*tls.Configuration{
				{
					Certificate: &tls.Certificate{
						CertFile: localhostCert,
						KeyFile:  localhostKey,
					},
				},
			},
		},
	}

	srv := NewServer(globalConfig, nil, entryPoints)

	mapEntryPoints := srv.loadConfig(dynamicConfigs, globalConfig)
	if !mapEntryPoints["https"].certs.ContainsCertificates() {
		t.Fatal("got error: https entryPoint must have TLS certificates.")
	}
}

func TestReuseBackend(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	globalConfig := configuration.GlobalConfiguration{
		DefaultEntryPoints: []string{"http"},
	}

	entryPoints := map[string]EntryPoint{
		"http": {Configuration: &configuration.EntryPoint{
			ForwardedHeaders: &configuration.ForwardedHeaders{Insecure: true},
		}},
	}

	dynamicConfigs := types.Configurations{
		"config": th.BuildConfiguration(
			th.WithFrontends(
				th.WithFrontend("backend",
					th.WithFrontendName("frontend0"),
					th.WithEntryPoints("http"),
					th.WithRoutes(th.WithRoute("/ok", "Path: /ok"))),
				th.WithFrontend("backend",
					th.WithFrontendName("frontend1"),
					th.WithEntryPoints("http"),
					th.WithRoutes(th.WithRoute("/unauthorized", "Path: /unauthorized")),
					th.WithBasicAuth("foo", "bar")),
			),
			th.WithBackends(th.WithBackendNew("backend",
				th.WithLBMethod("wrr"),
				th.WithServersNew(th.WithServerNew(testServer.URL))),
			),
		),
	}

	srv := NewServer(globalConfig, nil, entryPoints)

	serverEntryPoints := srv.loadConfig(dynamicConfigs, globalConfig)

	// Test that the /ok path returns a status 200.
	responseRecorderOk := &httptest.ResponseRecorder{}
	requestOk := httptest.NewRequest(http.MethodGet, testServer.URL+"/ok", nil)
	serverEntryPoints["http"].httpRouter.ServeHTTP(responseRecorderOk, requestOk)

	assert.Equal(t, http.StatusOK, responseRecorderOk.Result().StatusCode, "status code")

	// Test that the /unauthorized path returns a 401 because of
	// the basic authentication defined on the frontend.
	responseRecorderUnauthorized := &httptest.ResponseRecorder{}
	requestUnauthorized := httptest.NewRequest(http.MethodGet, testServer.URL+"/unauthorized", nil)
	serverEntryPoints["http"].httpRouter.ServeHTTP(responseRecorderUnauthorized, requestUnauthorized)

	assert.Equal(t, http.StatusUnauthorized, responseRecorderUnauthorized.Result().StatusCode, "status code")
}

func TestThrottleProviderConfigReload(t *testing.T) {
	throttleDuration := 30 * time.Millisecond
	publishConfig := make(chan types.ConfigMessage)
	providerConfig := make(chan types.ConfigMessage)
	stop := make(chan bool)
	defer func() {
		stop <- true
	}()

	globalConfig := configuration.GlobalConfiguration{}
	server := NewServer(globalConfig, nil, nil)

	go server.throttleProviderConfigReload(throttleDuration, publishConfig, providerConfig, stop)

	publishedConfigCount := 0
	stopConsumeConfigs := make(chan bool)
	go func() {
		for {
			select {
			case <-stop:
				return
			case <-stopConsumeConfigs:
				return
			case <-publishConfig:
				publishedConfigCount++
			}
		}
	}()

	// publish 5 new configs, one new config each 10 milliseconds
	for i := 0; i < 5; i++ {
		providerConfig <- types.ConfigMessage{}
		time.Sleep(10 * time.Millisecond)
	}

	// after 50 milliseconds 5 new configs were published
	// with a throttle duration of 30 milliseconds this means, we should have received 2 new configs
	assert.Equal(t, 2, publishedConfigCount, "times configs were published")

	stopConsumeConfigs <- true

	select {
	case <-publishConfig:
		// There should be exactly one more message that we receive after ~60 milliseconds since the start of the test.
		select {
		case <-publishConfig:
			t.Error("extra config publication found")
		case <-time.After(100 * time.Millisecond):
			return
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Last config was not published in time")
	}
}

func TestServerMultipleFrontendRules(t *testing.T) {
	testCases := []struct {
		expression  string
		requestURL  string
		expectedURL string
	}{
		{
			expression:  "Host:foo.bar",
			requestURL:  "http://foo.bar",
			expectedURL: "http://foo.bar",
		},
		{
			expression:  "PathPrefix:/management;ReplacePath:/health",
			requestURL:  "http://foo.bar/management",
			expectedURL: "http://foo.bar/health",
		},
		{
			expression:  "Host:foo.bar;AddPrefix:/blah",
			requestURL:  "http://foo.bar/baz",
			expectedURL: "http://foo.bar/blah/baz",
		},
		{
			expression:  "PathPrefixStripRegex:/one/{two}/{three:[0-9]+}",
			requestURL:  "http://foo.bar/one/some/12345/four",
			expectedURL: "http://foo.bar/four",
		},
		{
			expression:  "PathPrefixStripRegex:/one/{two}/{three:[0-9]+};AddPrefix:/zero",
			requestURL:  "http://foo.bar/one/some/12345/four",
			expectedURL: "http://foo.bar/zero/four",
		},
		{
			expression:  "AddPrefix:/blah;ReplacePath:/baz",
			requestURL:  "http://foo.bar/hello",
			expectedURL: "http://foo.bar/baz",
		},
		{
			expression:  "PathPrefixStrip:/management;ReplacePath:/health",
			requestURL:  "http://foo.bar/management",
			expectedURL: "http://foo.bar/health",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.expression, func(t *testing.T) {
			t.Parallel()

			router := mux.NewRouter()
			route := router.NewRoute()
			serverRoute := &types.ServerRoute{Route: route}
			reqHostMid := &middlewares.RequestHost{}
			rls := &rules.Rules{Route: serverRoute}

			expression := test.expression
			routeResult, err := rls.Parse(expression)

			if err != nil {
				t.Fatalf("Error while building route for %s: %+v", expression, err)
			}

			request := th.MustNewRequest(http.MethodGet, test.requestURL, nil)
			var routeMatch bool
			reqHostMid.ServeHTTP(nil, request, func(w http.ResponseWriter, r *http.Request) {
				routeMatch = routeResult.Match(r, &mux.RouteMatch{Route: routeResult})
			})

			if !routeMatch {
				t.Fatalf("Rule %s doesn't match", expression)
			}

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, test.expectedURL, r.URL.String(), "URL")
			})

			hd := buildMatcherMiddlewares(serverRoute, handler)
			serverRoute.Route.Handler(hd)

			serverRoute.Route.GetHandler().ServeHTTP(nil, request)
		})
	}
}

func TestServerBuildHealthCheckOptions(t *testing.T) {
	lb := &testLoadBalancer{}
	globalInterval := 15 * time.Second

	testCases := []struct {
		desc         string
		hc           *types.HealthCheck
		expectedOpts *healthcheck.Options
	}{
		{
			desc:         "nil health check",
			hc:           nil,
			expectedOpts: nil,
		},
		{
			desc: "empty path",
			hc: &types.HealthCheck{
				Path: "",
			},
			expectedOpts: nil,
		},
		{
			desc: "unparseable interval",
			hc: &types.HealthCheck{
				Path:     "/path",
				Interval: "unparseable",
			},
			expectedOpts: &healthcheck.Options{
				Path:     "/path",
				Interval: globalInterval,
				LB:       lb,
			},
		},
		{
			desc: "sub-zero interval",
			hc: &types.HealthCheck{
				Path:     "/path",
				Interval: "-42s",
			},
			expectedOpts: &healthcheck.Options{
				Path:     "/path",
				Interval: globalInterval,
				LB:       lb,
			},
		},
		{
			desc: "parseable interval",
			hc: &types.HealthCheck{
				Path:     "/path",
				Interval: "5m",
			},
			expectedOpts: &healthcheck.Options{
				Path:     "/path",
				Interval: 5 * time.Minute,
				LB:       lb,
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			opts := buildHealthCheckOptions(lb, "backend", test.hc, &configuration.HealthCheckConfig{Interval: flaeg.Duration(globalInterval)})
			assert.Equal(t, test.expectedOpts, opts, "health check options")
		})
	}
}
