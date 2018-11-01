package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/containous/flaeg"
	"github.com/containous/mux"
	"github.com/containous/traefik/configuration"
	"github.com/containous/traefik/healthcheck"
	"github.com/containous/traefik/metrics"
	"github.com/containous/traefik/middlewares"
	"github.com/containous/traefik/testhelpers"
	"github.com/containous/traefik/tls"
	"github.com/containous/traefik/types"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/negroni"
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

func (lb *testLoadBalancer) RemoveServer(u *url.URL) error {
	return nil
}

func (lb *testLoadBalancer) UpsertServer(u *url.URL, options ...roundrobin.ServerOption) error {
	return nil
}

func (lb *testLoadBalancer) Servers() []*url.URL {
	return []*url.URL{}
}

func TestPrepareServerTimeouts(t *testing.T) {
	tests := []struct {
		desc             string
		globalConfig     configuration.GlobalConfiguration
		wantIdleTimeout  time.Duration
		wantReadTimeout  time.Duration
		wantWriteTimeout time.Duration
	}{
		{
			desc: "full configuration",
			globalConfig: configuration.GlobalConfiguration{
				RespondingTimeouts: &configuration.RespondingTimeouts{
					IdleTimeout:  flaeg.Duration(10 * time.Second),
					ReadTimeout:  flaeg.Duration(12 * time.Second),
					WriteTimeout: flaeg.Duration(14 * time.Second),
				},
			},
			wantIdleTimeout:  time.Duration(10 * time.Second),
			wantReadTimeout:  time.Duration(12 * time.Second),
			wantWriteTimeout: time.Duration(14 * time.Second),
		},
		{
			desc:             "using defaults",
			globalConfig:     configuration.GlobalConfiguration{},
			wantIdleTimeout:  time.Duration(180 * time.Second),
			wantReadTimeout:  time.Duration(0 * time.Second),
			wantWriteTimeout: time.Duration(0 * time.Second),
		},
		{
			desc: "deprecated IdleTimeout configured",
			globalConfig: configuration.GlobalConfiguration{
				IdleTimeout: flaeg.Duration(45 * time.Second),
			},
			wantIdleTimeout:  time.Duration(45 * time.Second),
			wantReadTimeout:  time.Duration(0 * time.Second),
			wantWriteTimeout: time.Duration(0 * time.Second),
		},
		{
			desc: "deprecated and new IdleTimeout configured",
			globalConfig: configuration.GlobalConfiguration{
				IdleTimeout: flaeg.Duration(45 * time.Second),
				RespondingTimeouts: &configuration.RespondingTimeouts{
					IdleTimeout: flaeg.Duration(80 * time.Second),
				},
			},
			wantIdleTimeout:  time.Duration(45 * time.Second),
			wantReadTimeout:  time.Duration(0 * time.Second),
			wantWriteTimeout: time.Duration(0 * time.Second),
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			entryPointName := "http"
			entryPoint := &configuration.EntryPoint{
				Address:          "localhost:0",
				ForwardedHeaders: &configuration.ForwardedHeaders{Insecure: true},
			}
			router := middlewares.NewHandlerSwitcher(mux.NewRouter())

			srv := NewServer(test.globalConfig)
			httpServer, _, err := srv.prepareServer(entryPointName, entryPoint, router, nil, nil)
			if err != nil {
				t.Fatalf("Unexpected error when preparing srv: %s", err)
			}

			if httpServer.IdleTimeout != test.wantIdleTimeout {
				t.Errorf("Got %s as IdleTimeout, want %s", httpServer.IdleTimeout, test.wantIdleTimeout)
			}
			if httpServer.ReadTimeout != test.wantReadTimeout {
				t.Errorf("Got %s as ReadTimeout, want %s", httpServer.ReadTimeout, test.wantReadTimeout)
			}
			if httpServer.WriteTimeout != test.wantWriteTimeout {
				t.Errorf("Got %s as WriteTimeout, want %s", httpServer.WriteTimeout, test.wantWriteTimeout)
			}
		})
	}
}

func TestListenProvidersSkipsEmptyConfigs(t *testing.T) {
	server, stop, invokeStopChan := setupListenProvider(10 * time.Millisecond)
	defer invokeStopChan()

	go func() {
		for {
			select {
			case <-stop:
				return
			case <-server.configurationValidatedChan:
				t.Error("An empty configuration was published but it should not")
			}
		}
	}()

	server.configurationChan <- types.ConfigMessage{ProviderName: "kubernetes"}

	// give some time so that the configuration can be processed
	time.Sleep(100 * time.Millisecond)
}

func TestListenProvidersSkipsSameConfigurationForProvider(t *testing.T) {
	server, stop, invokeStopChan := setupListenProvider(10 * time.Millisecond)
	defer invokeStopChan()

	publishedConfigCount := 0
	go func() {
		for {
			select {
			case <-stop:
				return
			case config := <-server.configurationValidatedChan:
				// set the current configuration
				// this is usually done in the processing part of the published configuration
				// so we have to emulate the behaviour here
				currentConfigurations := server.currentConfigurations.Get().(types.Configurations)
				currentConfigurations[config.ProviderName] = config.Configuration
				server.currentConfigurations.Set(currentConfigurations)

				publishedConfigCount++
				if publishedConfigCount > 1 {
					t.Error("Same configuration should not be published multiple times")
				}
			}
		}
	}()

	config := buildDynamicConfig(
		withFrontend("frontend", buildFrontend()),
		withBackend("backend", buildBackend()),
	)

	// provide a configuration
	server.configurationChan <- types.ConfigMessage{ProviderName: "kubernetes", Configuration: config}

	// give some time so that the configuration can be processed
	time.Sleep(20 * time.Millisecond)

	// provide the same configuration a second time
	server.configurationChan <- types.ConfigMessage{ProviderName: "kubernetes", Configuration: config}

	// give some time so that the configuration can be processed
	time.Sleep(100 * time.Millisecond)
}

func TestListenProvidersPublishesConfigForEachProvider(t *testing.T) {
	server, stop, invokeStopChan := setupListenProvider(10 * time.Millisecond)
	defer invokeStopChan()

	publishedProviderConfigCount := map[string]int{}
	publishedConfigCount := 0
	consumePublishedConfigsDone := make(chan bool)
	go func() {
		for {
			select {
			case <-stop:
				return
			case newConfig := <-server.configurationValidatedChan:
				publishedProviderConfigCount[newConfig.ProviderName]++
				publishedConfigCount++
				if publishedConfigCount == 2 {
					consumePublishedConfigsDone <- true
					return
				}
			}
		}
	}()

	config := buildDynamicConfig(
		withFrontend("frontend", buildFrontend()),
		withBackend("backend", buildBackend()),
	)
	server.configurationChan <- types.ConfigMessage{ProviderName: "kubernetes", Configuration: config}
	server.configurationChan <- types.ConfigMessage{ProviderName: "marathon", Configuration: config}

	select {
	case <-consumePublishedConfigsDone:
		if val := publishedProviderConfigCount["kubernetes"]; val != 1 {
			t.Errorf("Got %d configuration publication(s) for provider %q, want 1", val, "kubernetes")
		}
		if val := publishedProviderConfigCount["marathon"]; val != 1 {
			t.Errorf("Got %d configuration publication(s) for provider %q, want 1", val, "marathon")
		}
	case <-time.After(100 * time.Millisecond):
		t.Errorf("Published configurations were not consumed in time")
	}
}

// setupListenProvider configures the Server and starts listenProviders
func setupListenProvider(throttleDuration time.Duration) (server *Server, stop chan bool, invokeStopChan func()) {
	stop = make(chan bool)
	invokeStopChan = func() {
		stop <- true
	}

	globalConfig := configuration.GlobalConfiguration{
		EntryPoints: configuration.EntryPoints{
			"http": &configuration.EntryPoint{},
		},
		ProvidersThrottleDuration: flaeg.Duration(throttleDuration),
	}

	server = NewServer(globalConfig)
	go server.listenProviders(stop)

	return server, stop, invokeStopChan
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
	server := NewServer(globalConfig)

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
	wantPublishedConfigCount := 2
	if publishedConfigCount != wantPublishedConfigCount {
		t.Errorf("%d times configs were published, want %d times", publishedConfigCount, wantPublishedConfigCount)
	}

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
	cases := []struct {
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

	for _, test := range cases {
		test := test
		t.Run(test.expression, func(t *testing.T) {
			t.Parallel()

			router := mux.NewRouter()
			route := router.NewRoute()
			serverRoute := &serverRoute{route: route}
			rules := &Rules{route: serverRoute}

			expression := test.expression
			routeResult, err := rules.Parse(expression)

			if err != nil {
				t.Fatalf("Error while building route for %s: %+v", expression, err)
			}

			request := testhelpers.MustNewRequest(http.MethodGet, test.requestURL, nil)
			routeMatch := routeResult.Match(request, &mux.RouteMatch{Route: routeResult})

			if !routeMatch {
				t.Fatalf("Rule %s doesn't match", expression)
			}

			server := new(Server)

			server.wireFrontendBackend(serverRoute, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.String() != test.expectedURL {
					t.Fatalf("got URL %s, expected %s", r.URL.String(), test.expectedURL)
				}
			}))
			serverRoute.route.GetHandler().ServeHTTP(nil, request)
		})
	}
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
					EntryPoints: configuration.EntryPoints{
						"http": &configuration.EntryPoint{
							ForwardedHeaders: &configuration.ForwardedHeaders{Insecure: true},
						},
					},
					HealthCheck: &configuration.HealthCheckConfig{Interval: flaeg.Duration(5 * time.Second)},
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

				srv := NewServer(globalConfig)
				if _, err := srv.loadConfig(dynamicConfigs, globalConfig); err != nil {
					t.Fatalf("got error: %s", err)
				}

				wantNumHealthCheckBackends := 0
				if healthCheck != nil {
					wantNumHealthCheckBackends = 1
				}
				gotNumHealthCheckBackends := len(healthcheck.GetHealthCheck().Backends)
				if gotNumHealthCheckBackends != wantNumHealthCheckBackends {
					t.Errorf("got %d health check backends, want %d", gotNumHealthCheckBackends, wantNumHealthCheckBackends)
				}
			})
		}
	}
}

func TestServerParseHealthCheckOptions(t *testing.T) {
	lb := &testLoadBalancer{}
	globalInterval := 15 * time.Second

	tests := []struct {
		desc     string
		hc       *types.HealthCheck
		wantOpts *healthcheck.Options
	}{
		{
			desc:     "nil health check",
			hc:       nil,
			wantOpts: nil,
		},
		{
			desc: "empty path",
			hc: &types.HealthCheck{
				Path: "",
			},
			wantOpts: nil,
		},
		{
			desc: "unparseable interval",
			hc: &types.HealthCheck{
				Path:     "/path",
				Interval: "unparseable",
			},
			wantOpts: &healthcheck.Options{
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
			wantOpts: &healthcheck.Options{
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
			wantOpts: &healthcheck.Options{
				Path:     "/path",
				Interval: 5 * time.Minute,
				LB:       lb,
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			gotOpts := parseHealthCheckOptions(lb, "backend", test.hc, &configuration.HealthCheckConfig{Interval: flaeg.Duration(globalInterval)})
			if !reflect.DeepEqual(gotOpts, test.wantOpts) {
				t.Errorf("got health check options %+v, want %+v", gotOpts, test.wantOpts)
			}
		})
	}
}

func TestNewServerWithWhitelistSourceRange(t *testing.T) {
	cases := []struct {
		desc                 string
		whitelistStrings     []string
		middlewareConfigured bool
		errMessage           string
	}{
		{
			desc:                 "no whitelists configued",
			whitelistStrings:     nil,
			middlewareConfigured: false,
			errMessage:           "",
		}, {
			desc: "whitelists configued",
			whitelistStrings: []string{
				"1.2.3.4/24",
				"fe80::/16",
			},
			middlewareConfigured: true,
			errMessage:           "",
		}, {
			desc: "invalid whitelists configued",
			whitelistStrings: []string{
				"foo",
			},
			middlewareConfigured: false,
			errMessage:           "parsing CIDR whitelist [foo]: parsing CIDR whitelist <nil>: invalid CIDR address: foo",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			middleware, err := configureIPWhitelistMiddleware(tc.whitelistStrings)

			if tc.errMessage != "" {
				require.EqualError(t, err, tc.errMessage)
			} else {
				assert.NoError(t, err)

				if tc.middlewareConfigured {
					require.NotNil(t, middleware, "not expected middleware to be configured")
				} else {
					require.Nil(t, middleware, "expected middleware to be configured")
				}
			}
		})
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

	srv := NewServer(globalConfig)
	if _, err := srv.loadConfig(dynamicConfigs, globalConfig); err != nil {
		t.Fatalf("got error: %s", err)
	}
}

func TestConfigureBackends(t *testing.T) {
	validMethod := "Drr"
	defaultMethod := "wrr"

	tests := []struct {
		desc           string
		lb             *types.LoadBalancer
		wantMethod     string
		wantStickiness *types.Stickiness
	}{
		{
			desc: "valid load balancer method with sticky enabled",
			lb: &types.LoadBalancer{
				Method:     validMethod,
				Stickiness: &types.Stickiness{},
			},
			wantMethod:     validMethod,
			wantStickiness: &types.Stickiness{},
		},
		{
			desc: "valid load balancer method with sticky disabled",
			lb: &types.LoadBalancer{
				Method:     validMethod,
				Stickiness: nil,
			},
			wantMethod: validMethod,
		},
		{
			desc: "invalid load balancer method with sticky enabled",
			lb: &types.LoadBalancer{
				Method:     "Invalid",
				Stickiness: &types.Stickiness{},
			},
			wantMethod:     defaultMethod,
			wantStickiness: &types.Stickiness{},
		},
		{
			desc: "invalid load balancer method with sticky disabled",
			lb: &types.LoadBalancer{
				Method:     "Invalid",
				Stickiness: nil,
			},
			wantMethod: defaultMethod,
		},
		{
			desc:       "missing load balancer",
			lb:         nil,
			wantMethod: defaultMethod,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			backend := &types.Backend{
				LoadBalancer: test.lb,
			}

			srv := Server{}
			srv.configureBackends(map[string]*types.Backend{
				"backend": backend,
			})

			wantLB := types.LoadBalancer{
				Method:     test.wantMethod,
				Stickiness: test.wantStickiness,
			}
			if !reflect.DeepEqual(*backend.LoadBalancer, wantLB) {
				t.Errorf("got backend load-balancer\n%v\nwant\n%v\n", spew.Sdump(backend.LoadBalancer), spew.Sdump(wantLB))
			}
		})
	}
}

func TestServerEntryPointWhitelistConfig(t *testing.T) {
	tests := []struct {
		desc           string
		entrypoint     *configuration.EntryPoint
		wantMiddleware bool
	}{
		{
			desc: "no whitelist middleware if no config on entrypoint",
			entrypoint: &configuration.EntryPoint{
				Address:          ":0",
				ForwardedHeaders: &configuration.ForwardedHeaders{Insecure: true},
			},
			wantMiddleware: false,
		},
		{
			desc: "whitelist middleware should be added if configured on entrypoint",
			entrypoint: &configuration.EntryPoint{
				Address: ":0",
				WhitelistSourceRange: []string{
					"127.0.0.1/32",
				},
				ForwardedHeaders: &configuration.ForwardedHeaders{Insecure: true},
			},
			wantMiddleware: true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			srv := Server{
				globalConfiguration: configuration.GlobalConfiguration{
					EntryPoints: map[string]*configuration.EntryPoint{
						"test": test.entrypoint,
					},
				},
				metricsRegistry: metrics.NewVoidRegistry(),
			}

			srv.serverEntryPoints = srv.buildEntryPoints(srv.globalConfiguration)
			srvEntryPoint := srv.setupServerEntryPoint("test", srv.serverEntryPoints["test"])
			handler := srvEntryPoint.httpServer.Handler.(*mux.Router).NotFoundHandler.(*negroni.Negroni)
			found := false
			for _, handler := range handler.Handlers() {
				if reflect.TypeOf(handler) == reflect.TypeOf((*middlewares.IPWhiteLister)(nil)) {
					found = true
				}
			}

			if found && !test.wantMiddleware {
				t.Errorf("ip whitelist middleware was installed even though it should not")
			}

			if !found && test.wantMiddleware {
				t.Errorf("ip whitelist middleware was not installed even though it should have")
			}
		})
	}
}

func TestServerResponseEmptyBackend(t *testing.T) {
	const requestPath = "/path"
	const routeRule = "Path:" + requestPath

	testCases := []struct {
		desc           string
		dynamicConfig  func(testServerURL string) *types.Configuration
		wantStatusCode int
	}{
		{
			desc: "Ok",
			dynamicConfig: func(testServerURL string) *types.Configuration {
				return buildDynamicConfig(
					withFrontend("frontend", buildFrontend(withRoute(requestPath, routeRule))),
					withBackend("backend", buildBackend(withServer("testServer", testServerURL))),
				)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			desc: "No Frontend",
			dynamicConfig: func(testServerURL string) *types.Configuration {
				return buildDynamicConfig()
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			desc: "Empty Backend LB-Drr",
			dynamicConfig: func(testServerURL string) *types.Configuration {
				return buildDynamicConfig(
					withFrontend("frontend", buildFrontend(withRoute(requestPath, routeRule))),
					withBackend("backend", buildBackend(withLoadBalancer("Drr", false))),
				)
			},
			wantStatusCode: http.StatusServiceUnavailable,
		},
		{
			desc: "Empty Backend LB-Drr Sticky",
			dynamicConfig: func(testServerURL string) *types.Configuration {
				return buildDynamicConfig(
					withFrontend("frontend", buildFrontend(withRoute(requestPath, routeRule))),
					withBackend("backend", buildBackend(withLoadBalancer("Drr", true))),
				)
			},
			wantStatusCode: http.StatusServiceUnavailable,
		},
		{
			desc: "Empty Backend LB-Wrr",
			dynamicConfig: func(testServerURL string) *types.Configuration {
				return buildDynamicConfig(
					withFrontend("frontend", buildFrontend(withRoute(requestPath, routeRule))),
					withBackend("backend", buildBackend(withLoadBalancer("Wrr", false))),
				)
			},
			wantStatusCode: http.StatusServiceUnavailable,
		},
		{
			desc: "Empty Backend LB-Wrr Sticky",
			dynamicConfig: func(testServerURL string) *types.Configuration {
				return buildDynamicConfig(
					withFrontend("frontend", buildFrontend(withRoute(requestPath, routeRule))),
					withBackend("backend", buildBackend(withLoadBalancer("Wrr", true))),
				)
			},
			wantStatusCode: http.StatusServiceUnavailable,
		},
	}

	for _, test := range testCases {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			testServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
			}))
			defer testServer.Close()

			globalConfig := configuration.GlobalConfiguration{
				EntryPoints: configuration.EntryPoints{
					"http": &configuration.EntryPoint{ForwardedHeaders: &configuration.ForwardedHeaders{Insecure: true}},
				},
			}
			dynamicConfigs := types.Configurations{"config": test.dynamicConfig(testServer.URL)}

			srv := NewServer(globalConfig)
			entryPoints, err := srv.loadConfig(dynamicConfigs, globalConfig)
			if err != nil {
				t.Fatalf("error loading config: %s", err)
			}

			responseRecorder := &httptest.ResponseRecorder{}
			request := httptest.NewRequest(http.MethodGet, testServer.URL+requestPath, nil)

			entryPoints["http"].httpRouter.ServeHTTP(responseRecorder, request)

			if responseRecorder.Result().StatusCode != test.wantStatusCode {
				t.Errorf("got status code %d, want %d", responseRecorder.Result().StatusCode, test.wantStatusCode)
			}
		})
	}
}

func TestBuildEntryPointRedirect(t *testing.T) {
	srv := Server{
		globalConfiguration: configuration.GlobalConfiguration{
			EntryPoints: configuration.EntryPoints{
				"http":  &configuration.EntryPoint{Address: ":80"},
				"https": &configuration.EntryPoint{Address: ":443", TLS: &tls.TLS{}},
			},
		},
	}

	testCases := []struct {
		desc              string
		srcEntryPointName string
		url               string
		entryPoint        *configuration.EntryPoint
		redirect          *types.Redirect
		expectedURL       string
	}{
		{
			desc:              "redirect regex",
			srcEntryPointName: "http",
			url:               "http://foo.com",
			redirect: &types.Redirect{
				Regex:       `^(?:http?:\/\/)(foo)(\.com)$`,
				Replacement: "https://$1{{\"bar\"}}$2",
			},
			entryPoint: &configuration.EntryPoint{
				Address: ":80",
				Redirect: &types.Redirect{
					Regex:       `^(?:http?:\/\/)(foo)(\.com)$`,
					Replacement: "https://$1{{\"bar\"}}$2",
				},
			},
			expectedURL: "https://foobar.com",
		},
		{
			desc:              "redirect entry point",
			srcEntryPointName: "http",
			url:               "http://foo:80",
			redirect: &types.Redirect{
				EntryPoint: "https",
			},
			entryPoint: &configuration.EntryPoint{
				Address: ":80",
				Redirect: &types.Redirect{
					EntryPoint: "https",
				},
			},
			expectedURL: "https://foo:443",
		},
		{
			desc:              "redirect entry point with regex (ignored)",
			srcEntryPointName: "http",
			url:               "http://foo.com:80",
			redirect: &types.Redirect{
				EntryPoint:  "https",
				Regex:       `^(?:http?:\/\/)(foo)(\.com)$`,
				Replacement: "https://$1{{\"bar\"}}$2",
			},
			entryPoint: &configuration.EntryPoint{
				Address: ":80",
				Redirect: &types.Redirect{
					EntryPoint:  "https",
					Regex:       `^(?:http?:\/\/)(foo)(\.com)$`,
					Replacement: "https://$1{{\"bar\"}}$2",
				},
			},
			expectedURL: "https://foo.com:443",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rewrite, err := srv.buildRedirectHandler(test.srcEntryPointName, test.redirect)
			require.NoError(t, err)

			req := testhelpers.MustNewRequest(http.MethodGet, test.url, nil)
			recorder := httptest.NewRecorder()

			rewrite.ServeHTTP(recorder, req, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Location", "fail")
			}))

			location, err := recorder.Result().Location()
			require.NoError(t, err)
			assert.Equal(t, test.expectedURL, location.String())
		})
	}
}

func TestServerBuildEntryPointRedirect(t *testing.T) {
	srv := Server{
		globalConfiguration: configuration.GlobalConfiguration{
			EntryPoints: configuration.EntryPoints{
				"http":  &configuration.EntryPoint{Address: ":80"},
				"https": &configuration.EntryPoint{Address: ":443", TLS: &tls.TLS{}},
			},
		},
	}

	testCases := []struct {
		desc               string
		srcEntryPointName  string
		redirectEntryPoint string
		url                string
		expectedURL        string
		errorExpected      bool
	}{
		{
			desc:               "existing redirect entry point",
			srcEntryPointName:  "http",
			redirectEntryPoint: "https",
			url:                "http://foo:80",
			expectedURL:        "https://foo:443",
		},
		{
			desc:               "non-existing redirect entry point",
			srcEntryPointName:  "http",
			redirectEntryPoint: "foo",
			url:                "http://foo:80",
			errorExpected:      true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rewrite, err := srv.buildEntryPointRedirect(test.srcEntryPointName, test.redirectEntryPoint)
			if test.errorExpected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				recorder := httptest.NewRecorder()
				r := testhelpers.MustNewRequest(http.MethodGet, test.url, nil)
				rewrite.ServeHTTP(recorder, r, nil)

				location, err := recorder.Result().Location()
				require.NoError(t, err)

				assert.Equal(t, test.expectedURL, location.String())
			}
		})
	}
}

func TestServerBuildRedirect(t *testing.T) {
	testCases := []struct {
		desc                   string
		globalConfiguration    configuration.GlobalConfiguration
		redirectEntryPointName string
		expectedReplacement    string
		errorExpected          bool
	}{
		{
			desc:                   "Redirect endpoint http to https with HTTPS protocol",
			redirectEntryPointName: "https",
			globalConfiguration: configuration.GlobalConfiguration{
				EntryPoints: configuration.EntryPoints{
					"http":  &configuration.EntryPoint{Address: ":80"},
					"https": &configuration.EntryPoint{Address: ":443", TLS: &tls.TLS{}},
				},
			},
			expectedReplacement: "https://${1}:443${2}",
		},
		{
			desc:                   "Redirect endpoint http to http02 with HTTP protocol",
			redirectEntryPointName: "http02",
			globalConfiguration: configuration.GlobalConfiguration{
				EntryPoints: configuration.EntryPoints{
					"http":   &configuration.EntryPoint{Address: ":80"},
					"http02": &configuration.EntryPoint{Address: ":88"},
				},
			},
			expectedReplacement: "http://${1}:88${2}",
		},
		{
			desc:                   "Redirect endpoint to non-existent entry point",
			redirectEntryPointName: "foobar",
			globalConfiguration: configuration.GlobalConfiguration{
				EntryPoints: configuration.EntryPoints{
					"http":   &configuration.EntryPoint{Address: ":80"},
					"http02": &configuration.EntryPoint{Address: ":88"},
				},
			},
			errorExpected: true,
		},
		{
			desc:                   "Redirect endpoint to an entry point with a malformed address",
			redirectEntryPointName: "http02",
			globalConfiguration: configuration.GlobalConfiguration{
				EntryPoints: configuration.EntryPoints{
					"http":   &configuration.EntryPoint{Address: ":80"},
					"http02": &configuration.EntryPoint{Address: "88"},
				},
			},
			errorExpected: true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			srv := Server{globalConfiguration: test.globalConfiguration}

			_, replacement, err := srv.buildRedirect(test.redirectEntryPointName)

			require.Equal(t, test.errorExpected, err != nil, "Expected an error but don't have error, or Expected no error but have an error: %v", err)
			assert.Equal(t, test.expectedReplacement, replacement, "build redirect does not return the right replacement pattern")
		})
	}
}

func buildDynamicConfig(dynamicConfigBuilders ...func(*types.Configuration)) *types.Configuration {
	config := &types.Configuration{
		Frontends: make(map[string]*types.Frontend),
		Backends:  make(map[string]*types.Backend),
	}
	for _, build := range dynamicConfigBuilders {
		build(config)
	}
	return config
}

func withFrontend(frontendName string, frontend *types.Frontend) func(*types.Configuration) {
	return func(config *types.Configuration) {
		config.Frontends[frontendName] = frontend
	}
}

func withBackend(backendName string, backend *types.Backend) func(*types.Configuration) {
	return func(config *types.Configuration) {
		config.Backends[backendName] = backend
	}
}

func buildFrontend(frontendBuilders ...func(*types.Frontend)) *types.Frontend {
	fe := &types.Frontend{
		EntryPoints: []string{"http"},
		Backend:     "backend",
		Routes:      make(map[string]types.Route),
	}
	for _, build := range frontendBuilders {
		build(fe)
	}
	return fe
}

func withRoute(routeName, rule string) func(*types.Frontend) {
	return func(fe *types.Frontend) {
		fe.Routes[routeName] = types.Route{Rule: rule}
	}
}

func buildBackend(backendBuilders ...func(*types.Backend)) *types.Backend {
	be := &types.Backend{
		Servers:      make(map[string]types.Server),
		LoadBalancer: &types.LoadBalancer{Method: "Wrr"},
	}
	for _, build := range backendBuilders {
		build(be)
	}
	return be
}

func withServer(name, url string) func(backend *types.Backend) {
	return func(be *types.Backend) {
		be.Servers[name] = types.Server{URL: url}
	}
}

func withLoadBalancer(method string, sticky bool) func(*types.Backend) {
	return func(be *types.Backend) {
		if sticky {
			be.LoadBalancer = &types.LoadBalancer{Method: method, Stickiness: &types.Stickiness{CookieName: "test"}}
		} else {
			be.LoadBalancer = &types.LoadBalancer{Method: method}
		}
	}
}
