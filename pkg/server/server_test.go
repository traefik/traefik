package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/config/runtime"
	"github.com/containous/traefik/v2/pkg/config/static"
	th "github.com/containous/traefik/v2/pkg/testhelpers"
	"github.com/containous/traefik/v2/pkg/types"
	"github.com/stretchr/testify/assert"
)

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

	server.configurationChan <- dynamic.Message{ProviderName: "kubernetes"}

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
				currentConfigurations := server.currentConfigurations.Get().(dynamic.Configurations)
				currentConfigurations[conf.ProviderName] = conf.Configuration
				server.currentConfigurations.Set(currentConfigurations)

				publishedConfigCount++
				if publishedConfigCount > 1 {
					t.Error("Same configuration should not be published multiple times")
				}
			}
		}
	}()
	conf := &dynamic.Configuration{}
	conf.HTTP = th.BuildConfiguration(
		th.WithRouters(th.WithRouter("foo")),
		th.WithLoadBalancerServices(th.WithService("bar")),
	)

	// provide a configuration
	server.configurationChan <- dynamic.Message{ProviderName: "kubernetes", Configuration: conf}

	// give some time so that the configuration can be processed
	time.Sleep(20 * time.Millisecond)

	// provide the same configuration a second time
	server.configurationChan <- dynamic.Message{ProviderName: "kubernetes", Configuration: conf}

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

	conf := &dynamic.Configuration{}
	conf.HTTP = th.BuildConfiguration(
		th.WithRouters(th.WithRouter("foo")),
		th.WithLoadBalancerServices(th.WithService("bar")),
	)
	server.configurationChan <- dynamic.Message{ProviderName: "kubernetes", Configuration: conf}
	server.configurationChan <- dynamic.Message{ProviderName: "marathon", Configuration: conf}

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

	staticConfiguration := static.Configuration{
		Providers: &static.Providers{
			ProvidersThrottleDuration: types.Duration(throttleDuration),
		},
	}

	server = NewServer(staticConfiguration, nil, nil, nil)
	go server.listenProviders(stop)

	return server, stop, invokeStopChan
}

func TestServerResponseEmptyBackend(t *testing.T) {
	const requestPath = "/path"
	const routeRule = "Path(`" + requestPath + "`)"

	testCases := []struct {
		desc               string
		config             func(testServerURL string) *dynamic.HTTPConfiguration
		expectedStatusCode int
	}{
		{
			desc: "Ok",
			config: func(testServerURL string) *dynamic.HTTPConfiguration {
				return th.BuildConfiguration(
					th.WithRouters(th.WithRouter("foo",
						th.WithEntryPoints("http"),
						th.WithServiceName("bar"),
						th.WithRule(routeRule)),
					),
					th.WithLoadBalancerServices(th.WithService("bar",
						th.WithServers(th.WithServer(testServerURL))),
					),
				)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			desc: "No Frontend",
			config: func(testServerURL string) *dynamic.HTTPConfiguration {
				return th.BuildConfiguration()
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			desc: "Empty Backend LB",
			config: func(testServerURL string) *dynamic.HTTPConfiguration {
				return th.BuildConfiguration(
					th.WithRouters(th.WithRouter("foo",
						th.WithEntryPoints("http"),
						th.WithServiceName("bar"),
						th.WithRule(routeRule)),
					),
					th.WithLoadBalancerServices(th.WithService("bar")),
				)
			},
			expectedStatusCode: http.StatusServiceUnavailable,
		},
		{
			desc: "Empty Backend LB Sticky",
			config: func(testServerURL string) *dynamic.HTTPConfiguration {
				return th.BuildConfiguration(
					th.WithRouters(th.WithRouter("foo",
						th.WithEntryPoints("http"),
						th.WithServiceName("bar"),
						th.WithRule(routeRule)),
					),
					th.WithLoadBalancerServices(th.WithService("bar",
						th.WithSticky("test")),
					),
				)
			},
			expectedStatusCode: http.StatusServiceUnavailable,
		},
		{
			desc: "Empty Backend LB",
			config: func(testServerURL string) *dynamic.HTTPConfiguration {
				return th.BuildConfiguration(
					th.WithRouters(th.WithRouter("foo",
						th.WithEntryPoints("http"),
						th.WithServiceName("bar"),
						th.WithRule(routeRule)),
					),
					th.WithLoadBalancerServices(th.WithService("bar")),
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

			globalConfig := static.Configuration{}
			entryPointsConfig := TCPEntryPoints{
				"http": &TCPEntryPoint{},
			}

			srv := NewServer(globalConfig, nil, entryPointsConfig, nil)
			rtConf := runtime.NewConfig(dynamic.Configuration{HTTP: test.config(testServer.URL)})
			entryPoints, _ := srv.createHTTPHandlers(context.Background(), rtConf, []string{"http"})

			responseRecorder := &httptest.ResponseRecorder{}
			request := httptest.NewRequest(http.MethodGet, testServer.URL+requestPath, nil)

			entryPoints["http"].ServeHTTP(responseRecorder, request)

			assert.Equal(t, test.expectedStatusCode, responseRecorder.Result().StatusCode, "status code")
		})
	}
}
