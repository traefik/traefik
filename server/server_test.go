package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/containous/flaeg"
	"github.com/containous/mux"
	"github.com/containous/traefik/configuration"
	"github.com/containous/traefik/middlewares"
	th "github.com/containous/traefik/testhelpers"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unrolled/secure"
)

func TestPrepareServerTimeouts(t *testing.T) {
	testCases := []struct {
		desc                 string
		globalConfig         configuration.GlobalConfiguration
		expectedIdleTimeout  time.Duration
		expectedReadTimeout  time.Duration
		expectedWriteTimeout time.Duration
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
			expectedIdleTimeout:  time.Duration(10 * time.Second),
			expectedReadTimeout:  time.Duration(12 * time.Second),
			expectedWriteTimeout: time.Duration(14 * time.Second),
		},
		{
			desc:                 "using defaults",
			globalConfig:         configuration.GlobalConfiguration{},
			expectedIdleTimeout:  time.Duration(180 * time.Second),
			expectedReadTimeout:  time.Duration(0 * time.Second),
			expectedWriteTimeout: time.Duration(0 * time.Second),
		},
		{
			desc: "deprecated IdleTimeout configured",
			globalConfig: configuration.GlobalConfiguration{
				IdleTimeout: flaeg.Duration(45 * time.Second),
			},
			expectedIdleTimeout:  time.Duration(45 * time.Second),
			expectedReadTimeout:  time.Duration(0 * time.Second),
			expectedWriteTimeout: time.Duration(0 * time.Second),
		},
		{
			desc: "deprecated and new IdleTimeout configured",
			globalConfig: configuration.GlobalConfiguration{
				IdleTimeout: flaeg.Duration(45 * time.Second),
				RespondingTimeouts: &configuration.RespondingTimeouts{
					IdleTimeout: flaeg.Duration(80 * time.Second),
				},
			},
			expectedIdleTimeout:  time.Duration(45 * time.Second),
			expectedReadTimeout:  time.Duration(0 * time.Second),
			expectedWriteTimeout: time.Duration(0 * time.Second),
		},
	}

	for _, test := range testCases {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			entryPointName := "http"
			entryPoint := &configuration.EntryPoint{
				Address:          "localhost:0",
				ForwardedHeaders: &configuration.ForwardedHeaders{Insecure: true},
			}
			router := middlewares.NewHandlerSwitcher(mux.NewRouter())

			srv := NewServer(test.globalConfig, nil, nil)
			httpServer, _, err := srv.prepareServer(entryPointName, entryPoint, router, nil)
			require.NoError(t, err, "Unexpected error when preparing srv")

			assert.Equal(t, test.expectedIdleTimeout, httpServer.IdleTimeout, "IdleTimeout")
			assert.Equal(t, test.expectedReadTimeout, httpServer.ReadTimeout, "ReadTimeout")
			assert.Equal(t, test.expectedWriteTimeout, httpServer.WriteTimeout, "WriteTimeout")
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

	config := th.BuildConfiguration(
		th.WithFrontends(th.WithFrontend("backend")),
		th.WithBackends(th.WithBackendNew("backend")),
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

	config := th.BuildConfiguration(
		th.WithFrontends(th.WithFrontend("backend")),
		th.WithBackends(th.WithBackendNew("backend")),
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

	server = NewServer(globalConfig, nil, nil)
	go server.listenProviders(stop)

	return server, stop, invokeStopChan
}

func TestServerResponseEmptyBackend(t *testing.T) {
	const requestPath = "/path"
	const routeRule = "Path:" + requestPath

	testCases := []struct {
		desc               string
		config             func(testServerURL string) *types.Configuration
		expectedStatusCode int
	}{
		{
			desc: "Ok",
			config: func(testServerURL string) *types.Configuration {
				return th.BuildConfiguration(
					th.WithFrontends(th.WithFrontend("backend",
						th.WithEntryPoints("http"),
						th.WithRoutes(th.WithRoute(requestPath, routeRule))),
					),
					th.WithBackends(th.WithBackendNew("backend",
						th.WithLBMethod("wrr"),
						th.WithServersNew(th.WithServerNew(testServerURL))),
					),
				)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			desc: "No Frontend",
			config: func(testServerURL string) *types.Configuration {
				return th.BuildConfiguration()
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			desc: "Empty Backend LB-Drr",
			config: func(testServerURL string) *types.Configuration {
				return th.BuildConfiguration(
					th.WithFrontends(th.WithFrontend("backend",
						th.WithEntryPoints("http"),
						th.WithRoutes(th.WithRoute(requestPath, routeRule))),
					),
					th.WithBackends(th.WithBackendNew("backend",
						th.WithLBMethod("drr")),
					),
				)
			},
			expectedStatusCode: http.StatusServiceUnavailable,
		},
		{
			desc: "Empty Backend LB-Drr Sticky",
			config: func(testServerURL string) *types.Configuration {
				return th.BuildConfiguration(
					th.WithFrontends(th.WithFrontend("backend",
						th.WithEntryPoints("http"),
						th.WithRoutes(th.WithRoute(requestPath, routeRule))),
					),
					th.WithBackends(th.WithBackendNew("backend",
						th.WithLBMethod("drr"), th.WithLBSticky("test")),
					),
				)
			},
			expectedStatusCode: http.StatusServiceUnavailable,
		},
		{
			desc: "Empty Backend LB-Wrr",
			config: func(testServerURL string) *types.Configuration {
				return th.BuildConfiguration(
					th.WithFrontends(th.WithFrontend("backend",
						th.WithEntryPoints("http"),
						th.WithRoutes(th.WithRoute(requestPath, routeRule))),
					),
					th.WithBackends(th.WithBackendNew("backend",
						th.WithLBMethod("wrr")),
					),
				)
			},
			expectedStatusCode: http.StatusServiceUnavailable,
		},
		{
			desc: "Empty Backend LB-Wrr Sticky",
			config: func(testServerURL string) *types.Configuration {
				return th.BuildConfiguration(
					th.WithFrontends(th.WithFrontend("backend",
						th.WithEntryPoints("http"),
						th.WithRoutes(th.WithRoute(requestPath, routeRule))),
					),
					th.WithBackends(th.WithBackendNew("backend",
						th.WithLBMethod("wrr"), th.WithLBSticky("test")),
					),
				)
			},
			expectedStatusCode: http.StatusServiceUnavailable,
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

			globalConfig := configuration.GlobalConfiguration{}
			entryPointsConfig := map[string]EntryPoint{
				"http": {Configuration: &configuration.EntryPoint{ForwardedHeaders: &configuration.ForwardedHeaders{Insecure: true}}},
			}
			dynamicConfigs := types.Configurations{"config": test.config(testServer.URL)}

			srv := NewServer(globalConfig, nil, entryPointsConfig)
			entryPoints := srv.loadConfig(dynamicConfigs, globalConfig)

			responseRecorder := &httptest.ResponseRecorder{}
			request := httptest.NewRequest(http.MethodGet, testServer.URL+requestPath, nil)

			entryPoints["http"].httpRouter.ServeHTTP(responseRecorder, request)

			assert.Equal(t, test.expectedStatusCode, responseRecorder.Result().StatusCode, "status code")
		})
	}
}

type mockContext struct {
	headers http.Header
}

func (c mockContext) Deadline() (deadline time.Time, ok bool) {
	return deadline, ok
}

func (c mockContext) Done() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

func (c mockContext) Err() error {
	return context.DeadlineExceeded
}

func (c mockContext) Value(key interface{}) interface{} {
	return c.headers
}

func TestNewServerWithResponseModifiers(t *testing.T) {
	testCases := []struct {
		desc             string
		headerMiddleware *middlewares.HeaderStruct
		secureMiddleware *secure.Secure
		ctx              context.Context
		expected         map[string]string
	}{
		{
			desc:             "header and secure nil",
			headerMiddleware: nil,
			secureMiddleware: nil,
			ctx:              mockContext{},
			expected: map[string]string{
				"X-Default":       "powpow",
				"Referrer-Policy": "same-origin",
			},
		},
		{
			desc: "header middleware not nil",
			headerMiddleware: middlewares.NewHeaderFromStruct(&types.Headers{
				CustomResponseHeaders: map[string]string{
					"X-Default": "powpow",
				},
			}),
			secureMiddleware: nil,
			ctx:              mockContext{},
			expected: map[string]string{
				"X-Default":       "powpow",
				"Referrer-Policy": "same-origin",
			},
		},
		{
			desc:             "secure middleware not nil",
			headerMiddleware: nil,
			secureMiddleware: middlewares.NewSecure(&types.Headers{
				ReferrerPolicy: "no-referrer",
			}),
			ctx: mockContext{
				headers: http.Header{"Referrer-Policy": []string{"no-referrer"}},
			},
			expected: map[string]string{
				"X-Default":       "powpow",
				"Referrer-Policy": "no-referrer",
			},
		},
		{
			desc: "header and secure middleware not nil",
			headerMiddleware: middlewares.NewHeaderFromStruct(&types.Headers{
				CustomResponseHeaders: map[string]string{
					"Referrer-Policy": "powpow",
				},
			}),
			secureMiddleware: middlewares.NewSecure(&types.Headers{
				ReferrerPolicy: "no-referrer",
			}),
			ctx: mockContext{
				headers: http.Header{"Referrer-Policy": []string{"no-referrer"}},
			},
			expected: map[string]string{
				"X-Default":       "powpow",
				"Referrer-Policy": "powpow",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			headers := make(http.Header)
			headers.Add("X-Default", "powpow")
			headers.Add("Referrer-Policy", "same-origin")

			req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1", nil)

			res := &http.Response{
				Request: req.WithContext(test.ctx),
				Header:  headers,
			}

			responseModifier := buildModifyResponse(test.secureMiddleware, test.headerMiddleware)
			err := responseModifier(res)

			assert.NoError(t, err)
			assert.Equal(t, len(test.expected), len(res.Header))

			for k, v := range test.expected {
				assert.Equal(t, v, res.Header.Get(k))
			}
		})
	}
}
