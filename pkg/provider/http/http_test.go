package http

import (
	"context"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
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
	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, _ = rw.Write([]byte("{OK}"))
	})
	router := mux.NewRouter()

	router.HandleFunc("/endpoint", handler)

	go http.ListenAndServe("127.0.0.1:9000", router)

	provider := Provider{
		endpoint: "http://127.0.0.1:9000/endpoint",
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
