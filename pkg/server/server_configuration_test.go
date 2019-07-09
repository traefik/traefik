package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/containous/traefik/pkg/config/dynamic"
	"github.com/containous/traefik/pkg/config/static"
	th "github.com/containous/traefik/pkg/testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestReuseService(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	entryPoints := TCPEntryPoints{
		"http": &TCPEntryPoint{},
	}

	staticConfig := static.Configuration{}

	dynamicConfigs := th.BuildConfiguration(
		th.WithRouters(
			th.WithRouter("foo",
				th.WithServiceName("bar"),
				th.WithRule("Path(`/ok`)")),
			th.WithRouter("foo2",
				th.WithEntryPoints("http"),
				th.WithRule("Path(`/unauthorized`)"),
				th.WithServiceName("bar"),
				th.WithRouterMiddlewares("basicauth")),
		),
		th.WithMiddlewares(th.WithMiddleware("basicauth",
			th.WithBasicAuth(&dynamic.BasicAuth{Users: []string{"foo:bar"}}),
		)),
		th.WithLoadBalancerServices(th.WithService("bar",
			th.WithServers(th.WithServer(testServer.URL))),
		),
	)

	srv := NewServer(staticConfig, nil, entryPoints, nil)

	rtConf := dynamic.NewRuntimeConfig(dynamic.Configuration{HTTP: dynamicConfigs})
	entrypointsHandlers, _ := srv.createHTTPHandlers(context.Background(), rtConf, []string{"http"})

	// Test that the /ok path returns a status 200.
	responseRecorderOk := &httptest.ResponseRecorder{}
	requestOk := httptest.NewRequest(http.MethodGet, testServer.URL+"/ok", nil)
	entrypointsHandlers["http"].ServeHTTP(responseRecorderOk, requestOk)

	assert.Equal(t, http.StatusOK, responseRecorderOk.Result().StatusCode, "status code")

	// Test that the /unauthorized path returns a 401 because of
	// the basic authentication defined on the frontend.
	responseRecorderUnauthorized := &httptest.ResponseRecorder{}
	requestUnauthorized := httptest.NewRequest(http.MethodGet, testServer.URL+"/unauthorized", nil)
	entrypointsHandlers["http"].ServeHTTP(responseRecorderUnauthorized, requestUnauthorized)

	assert.Equal(t, http.StatusUnauthorized, responseRecorderUnauthorized.Result().StatusCode, "status code")
}

func TestThrottleProviderConfigReload(t *testing.T) {
	throttleDuration := 30 * time.Millisecond
	publishConfig := make(chan dynamic.Message)
	providerConfig := make(chan dynamic.Message)
	stop := make(chan bool)
	defer func() {
		stop <- true
	}()

	staticConfiguration := static.Configuration{}
	server := NewServer(staticConfiguration, nil, nil, nil)

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
		providerConfig <- dynamic.Message{}
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
