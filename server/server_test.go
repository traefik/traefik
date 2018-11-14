package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/containous/flaeg/parse"
	"github.com/containous/mux"
	"github.com/containous/traefik/config"
	"github.com/containous/traefik/middlewares"
	"github.com/containous/traefik/old/configuration"
	th "github.com/containous/traefik/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
					IdleTimeout:  parse.Duration(10 * time.Second),
					ReadTimeout:  parse.Duration(12 * time.Second),
					WriteTimeout: parse.Duration(14 * time.Second),
				},
			},
			expectedIdleTimeout:  10 * time.Second,
			expectedReadTimeout:  12 * time.Second,
			expectedWriteTimeout: 14 * time.Second,
		},
		{
			desc:                 "using defaults",
			globalConfig:         configuration.GlobalConfiguration{},
			expectedIdleTimeout:  180 * time.Second,
			expectedReadTimeout:  0 * time.Second,
			expectedWriteTimeout: 0 * time.Second,
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
			httpServer, _, err := srv.prepareServer(context.Background(), entryPointName, entryPoint, router)
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

	server.configurationChan <- config.Message{ProviderName: "kubernetes"}

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
			case conf := <-server.configurationValidatedChan:
				// set the current configuration
				// this is usually done in the processing part of the published configuration
				// so we have to emulate the behavior here
				currentConfigurations := server.currentConfigurations.Get().(config.Configurations)
				currentConfigurations[conf.ProviderName] = conf.Configuration
				server.currentConfigurations.Set(currentConfigurations)

				publishedConfigCount++
				if publishedConfigCount > 1 {
					t.Error("Same configuration should not be published multiple times")
				}
			}
		}
	}()

	conf := th.BuildConfiguration(
		th.WithRouters(th.WithRouter("foo")),
		th.WithLoadBalancerServices(th.WithService("bar")),
	)

	// provide a configuration
	server.configurationChan <- config.Message{ProviderName: "kubernetes", Configuration: conf}

	// give some time so that the configuration can be processed
	time.Sleep(20 * time.Millisecond)

	// provide the same configuration a second time
	server.configurationChan <- config.Message{ProviderName: "kubernetes", Configuration: conf}

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

	conf := th.BuildConfiguration(
		th.WithRouters(th.WithRouter("foo")),
		th.WithLoadBalancerServices(th.WithService("bar")),
	)
	server.configurationChan <- config.Message{ProviderName: "kubernetes", Configuration: conf}
	server.configurationChan <- config.Message{ProviderName: "marathon", Configuration: conf}

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
		ProvidersThrottleDuration: parse.Duration(throttleDuration),
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
		config             func(testServerURL string) *config.Configuration
		expectedStatusCode int
	}{
		{
			desc: "Ok",
			config: func(testServerURL string) *config.Configuration {
				return th.BuildConfiguration(
					th.WithRouters(th.WithRouter("foo",
						th.WithEntryPoints("http"),
						th.WithServiceName("bar"),
						th.WithRule(routeRule)),
					),
					th.WithLoadBalancerServices(th.WithService("bar",
						th.WithLBMethod("wrr"),
						th.WithServers(th.WithServer(testServerURL))),
					),
				)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			desc: "No Frontend",
			config: func(testServerURL string) *config.Configuration {
				return th.BuildConfiguration()
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			desc: "Empty Backend LB-Drr",
			config: func(testServerURL string) *config.Configuration {
				return th.BuildConfiguration(
					th.WithRouters(th.WithRouter("foo",
						th.WithEntryPoints("http"),
						th.WithServiceName("bar"),
						th.WithRule(routeRule)),
					),
					th.WithLoadBalancerServices(th.WithService("bar",
						th.WithLBMethod("drr")),
					),
				)
			},
			expectedStatusCode: http.StatusServiceUnavailable,
		},
		{
			desc: "Empty Backend LB-Drr Sticky",
			config: func(testServerURL string) *config.Configuration {
				return th.BuildConfiguration(
					th.WithRouters(th.WithRouter("foo",
						th.WithEntryPoints("http"),
						th.WithServiceName("bar"),
						th.WithRule(routeRule)),
					),
					th.WithLoadBalancerServices(th.WithService("bar",
						th.WithLBMethod("drr"), th.WithStickiness("test")),
					),
				)
			},
			expectedStatusCode: http.StatusServiceUnavailable,
		},
		{
			desc: "Empty Backend LB-Wrr",
			config: func(testServerURL string) *config.Configuration {
				return th.BuildConfiguration(
					th.WithRouters(th.WithRouter("foo",
						th.WithEntryPoints("http"),
						th.WithServiceName("bar"),
						th.WithRule(routeRule)),
					),
					th.WithLoadBalancerServices(th.WithService("bar",
						th.WithLBMethod("wrr")),
					),
				)
			},
			expectedStatusCode: http.StatusServiceUnavailable,
		},
		{
			desc: "Empty Backend LB-Wrr Sticky",
			config: func(testServerURL string) *config.Configuration {
				return th.BuildConfiguration(
					th.WithRouters(th.WithRouter("foo",
						th.WithEntryPoints("http"),
						th.WithServiceName("bar"),
						th.WithRule(routeRule)),
					),
					th.WithLoadBalancerServices(th.WithService("bar",
						th.WithLBMethod("wrr"), th.WithStickiness("test")),
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
			dynamicConfigs := config.Configurations{"config": test.config(testServer.URL)}

			srv := NewServer(globalConfig, nil, entryPointsConfig)
			entryPoints, _ := srv.loadConfig(dynamicConfigs, globalConfig)

			responseRecorder := &httptest.ResponseRecorder{}
			request := httptest.NewRequest(http.MethodGet, testServer.URL+requestPath, nil)

			entryPoints["http"].ServeHTTP(responseRecorder, request)

			assert.Equal(t, test.expectedStatusCode, responseRecorder.Result().StatusCode, "status code")
		})
	}
}
