package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/safe"
	"github.com/containous/traefik/v2/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestProvider_Init(t *testing.T) {
	provider := &Provider{}
	assert.Error(t, provider.Init())

	provider = &Provider{
		Endpoint: "localhost",
	}
	assert.NoError(t, provider.Init())
}

func TestProvider_GetDataFromEndpoint(t *testing.T) {
	configString := "{OK}"
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	defer server.Close()
	mux.HandleFunc("/endpoint", func(rw http.ResponseWriter, req *http.Request) {
		_, _ = rw.Write([]byte(configString))
	})

	provider := Provider{
		Endpoint:     server.URL + "/endpoint",
		PollInterval: types.Duration(1 * time.Second),
		PollTimeout:  types.Duration(1 * time.Second),
	}

	assert.NoError(t, provider.Init())

	data, err := provider.getDataFromEndpoint(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, configString, string(data))
}

func TestProvider_Provide(t *testing.T) {
	c, _ := json.Marshal(dynamic.Configuration{})
	configString := string(c)
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	defer server.Close()
	mux.HandleFunc("/endpoint", func(rw http.ResponseWriter, req *http.Request) {
		_, _ = rw.Write([]byte(configString))
	})

	provider := Provider{
		Endpoint:     server.URL + "/endpoint",
		PollTimeout:  types.Duration(1 * time.Second),
		PollInterval: types.Duration(100 * time.Millisecond),
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
		assert.Equal(t, &dynamic.Configuration{}, conf.Configuration)
	case <-timeout:
		t.Errorf("timeout while waiting for config")
	}
}
