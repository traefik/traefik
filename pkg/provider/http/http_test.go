package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/safe"
	"github.com/traefik/traefik/v2/pkg/tls"
	"k8s.io/utils/ptr"
)

func TestProvider_Init(t *testing.T) {
	tests := []struct {
		desc         string
		endpoint     string
		pollInterval ptypes.Duration
		expErr       bool
	}{
		{
			desc:   "should return an error if no endpoint is configured",
			expErr: true,
		},
		{
			desc:     "should return an error if pollInterval is equal to 0",
			endpoint: "http://localhost:8080",
			expErr:   true,
		},
		{
			desc:         "should not return an error",
			endpoint:     "http://localhost:8080",
			pollInterval: ptypes.Duration(time.Second),
			expErr:       false,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			provider := &Provider{
				Endpoint:     test.endpoint,
				PollInterval: test.pollInterval,
			}

			err := provider.Init()
			if test.expErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestProvider_SetDefaults(t *testing.T) {
	provider := &Provider{}

	provider.SetDefaults()

	assert.Equal(t, provider.PollInterval, ptypes.Duration(5*time.Second))
	assert.Equal(t, provider.PollTimeout, ptypes.Duration(5*time.Second))
	assert.Equal(t, int64(-1), provider.MaxResponseBodySize)
}

func TestProvider_fetchConfigurationData(t *testing.T) {
	tests := []struct {
		desc                string
		handler             func(rw http.ResponseWriter, req *http.Request)
		expData             []byte
		expErr              bool
		maxResponseBodySize *int64
	}{
		{
			desc:    "should return the fetched configuration data",
			expData: []byte("{}"),
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprintf(rw, "{}")
			},
		},
		{
			desc:   "should return an error if endpoint does not return an OK status code",
			expErr: true,
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusNoContent)
			},
		},
		{
			desc:                "should return an error response body is too long when maxResponseBodySize is 0",
			maxResponseBodySize: ptr.To(int64(0)),
			expErr:              true,
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprintf(rw, "{}")
			},
		},
		{
			desc:                "should return an error response body is too long when response is longer than maxResponseBodySize",
			maxResponseBodySize: ptr.To(int64(1)),
			expErr:              true,
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprintf(rw, "{}")
			},
		},
		{
			desc:                "should return the fetched configuration data when response is the same length with maxResponseBodySize",
			maxResponseBodySize: ptr.To(int64(2)),
			expData:             []byte("{}"),
			expErr:              false,
			handler: func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprintf(rw, "{}")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(test.handler))
			defer server.Close()

			var provider Provider
			provider.SetDefaults()

			provider.Endpoint = server.URL
			provider.PollTimeout = ptypes.Duration(1 * time.Second)
			provider.PollInterval = ptypes.Duration(100 * time.Millisecond)
			if test.maxResponseBodySize != nil {
				provider.MaxResponseBodySize = *test.maxResponseBodySize
			}

			err := provider.Init()
			require.NoError(t, err)

			configData, err := provider.fetchConfigurationData()
			if test.expErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expData, configData)
		})
	}
}

func TestProvider_decodeConfiguration(t *testing.T) {
	tests := []struct {
		desc       string
		configData []byte
		expConfig  *dynamic.Configuration
		expErr     bool
	}{
		{
			desc:       "should return an error if the configuration data cannot be decoded",
			expErr:     true,
			configData: []byte("{"),
		},
		{
			desc:       "should return the decoded dynamic configuration",
			configData: []byte("{\"tcp\":{\"routers\":{\"foo\":{}}}}"),
			expConfig: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           make(map[string]*dynamic.Router),
					Middlewares:       make(map[string]*dynamic.Middleware),
					Services:          make(map[string]*dynamic.Service),
					ServersTransports: make(map[string]*dynamic.ServersTransport),
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"foo": {},
					},
					Services: make(map[string]*dynamic.TCPService),
				},
				TLS: &dynamic.TLSConfiguration{
					Stores:  make(map[string]tls.Store),
					Options: make(map[string]tls.Options),
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  make(map[string]*dynamic.UDPRouter),
					Services: make(map[string]*dynamic.UDPService),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			configuration, err := decodeConfiguration(test.configData)
			if test.expErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expConfig, configuration)
		})
	}
}

func TestProvider_Provide(t *testing.T) {
	handler := func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(rw, "{}")
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	var provider Provider
	provider.SetDefaults()

	provider.Endpoint = server.URL
	provider.PollTimeout = ptypes.Duration(1 * time.Second)
	provider.PollInterval = ptypes.Duration(100 * time.Millisecond)

	err := provider.Init()
	require.NoError(t, err)

	configurationChan := make(chan dynamic.Message)

	expConfiguration := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers:           make(map[string]*dynamic.Router),
			Middlewares:       make(map[string]*dynamic.Middleware),
			Services:          make(map[string]*dynamic.Service),
			ServersTransports: make(map[string]*dynamic.ServersTransport),
		},
		TCP: &dynamic.TCPConfiguration{
			Routers:  make(map[string]*dynamic.TCPRouter),
			Services: make(map[string]*dynamic.TCPService),
		},
		TLS: &dynamic.TLSConfiguration{
			Stores:  make(map[string]tls.Store),
			Options: make(map[string]tls.Options),
		},
		UDP: &dynamic.UDPConfiguration{
			Routers:  make(map[string]*dynamic.UDPRouter),
			Services: make(map[string]*dynamic.UDPService),
		},
	}

	err = provider.Provide(configurationChan, safe.NewPool(t.Context()))
	require.NoError(t, err)

	timeout := time.After(time.Second)

	select {
	case configuration := <-configurationChan:
		assert.NotNil(t, configuration.Configuration)
		assert.Equal(t, expConfiguration, configuration.Configuration)
	case <-timeout:
		t.Errorf("timeout while waiting for config")
	}
}

func TestProvider_ProvideConfigurationOnlyOnceIfUnchanged(t *testing.T) {
	handler := func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(rw, "{}")
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	var provider Provider
	provider.SetDefaults()

	provider.Endpoint = server.URL + "/endpoint"
	provider.PollTimeout = ptypes.Duration(1 * time.Second)
	provider.PollInterval = ptypes.Duration(100 * time.Millisecond)

	err := provider.Init()
	require.NoError(t, err)

	configurationChan := make(chan dynamic.Message, 10)

	err = provider.Provide(configurationChan, safe.NewPool(t.Context()))
	require.NoError(t, err)

	time.Sleep(time.Second)

	assert.Len(t, configurationChan, 1)
}
