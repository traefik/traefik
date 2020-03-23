package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/safe"
	"github.com/stretchr/testify/assert"
)

func TestProviderInit(t *testing.T) {
	provider := &Provider{}
	assert.Error(t, provider.Init())

	provider = &Provider{
		endpoint: "localhost",
	}
	assert.NoError(t, provider.Init())
}

func TestGetDataFromEndpoint(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("{OK}"))
	}))
	defer ts.Close()

	provider := Provider{
		endpoint: ts.URL,
	}

	assert.NoError(t, provider.Init())

	data, err := provider.getDataFromEndpoint(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "{OK}", string(data))
}

func TestBuildConfiguration(t *testing.T) {
	provider := Provider{
		endpoint: "http://127.0.0.1:9000/endpoint",
	}

	assert.NoError(t, provider.Init())

	config := provider.buildConfiguration(context.Background(), []byte("{}"))
	assert.NotEqual(t, nil, config)

}

func TestProvide(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("{}"))
	}))
	defer ts.Close()

	provider := Provider{
		endpoint:     ts.URL,
		pollTimeout:  1 * time.Second,
		pollInterval: 100 * time.Millisecond,
	}

	assert.NoError(t, provider.Init())

	configChan := make(chan dynamic.Message)

	go func() {
		err := provider.Provide(configChan, safe.NewPool(context.Background()))
		assert.NoError(t, err)
	}()

	timeout := time.After(time.Second)
	select {
	case conf := <-configChan:
		assert.NotNil(t, conf.Configuration)
	case <-timeout:
		t.Errorf("timeout while waiting for config")
	}

}
