package hub

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/tls/generate"
)

func TestHandleConfig(t *testing.T) {
	cfgChan := make(chan dynamic.Message, 1)

	client, err := createAgentClient(&TLS{Insecure: true})
	require.NoError(t, err)
	h := newHandler("traefik-hub-ep", 42, cfgChan, nil, client)

	cfg := emptyDynamicConfiguration()
	cfg.HTTP.Routers["foo"] = &dynamic.Router{
		EntryPoints: []string{"ep"},
		Service:     "bar",
		Rule:        "Host(`foo.com`)",
	}

	req := configRequest{Configuration: cfg}

	b, err := json.Marshal(req)
	require.NoError(t, err)

	server := httptest.NewServer(h)
	t.Cleanup(server.Close)

	resp, err := http.Post(server.URL+"/config", "application/json", bytes.NewReader(b))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	select {
	case gotCfgRaw := <-cfgChan:
		patchDynamicConfiguration(cfg, "traefik-hub-ep", 42, nil)
		assert.Equal(t, cfg, gotCfgRaw.Configuration)

	case <-time.After(time.Second):
		t.Fatal("Configuration not received")
	}
}

func TestHandle_Config_MethodNotAllowed(t *testing.T) {
	cfgChan := make(chan dynamic.Message, 1)
	client, err := createAgentClient(&TLS{Insecure: true})
	require.NoError(t, err)
	h := newHandler("traefik-hub-ep", 42, cfgChan, nil, client)

	server := httptest.NewServer(h)
	t.Cleanup(server.Close)

	resp, err := http.Get(server.URL + "/config")
	require.NoError(t, err)

	err = resp.Body.Close()
	require.NoError(t, err)

	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestHandle_DiscoverIP(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	port := listener.Addr().(*net.TCPAddr).Port
	nonce := "XVlBzgbaiCMRAjWw"

	mux := http.NewServeMux()

	var handlerCallCount int
	mux.HandleFunc("/", func(_ http.ResponseWriter, req *http.Request) {
		handlerCallCount++
		assert.Equal(t, nonce, req.URL.Query().Get("nonce"))
	})

	certificate, err := generate.DefaultCertificate()
	require.NoError(t, err)
	agentServer := &http.Server{
		Handler: mux,
		TLSConfig: &tls.Config{
			Certificates:       []tls.Certificate{*certificate},
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS13,
		},
	}
	t.Cleanup(func() { _ = agentServer.Close() })

	rdy := make(chan struct{})

	go func(s *http.Server) {
		close(rdy)
		if err = s.ServeTLS(listener, "", ""); errors.Is(err, http.ErrServerClosed) {
			return
		}
	}(agentServer)

	<-rdy

	cfgChan := make(chan dynamic.Message, 1)
	client, err := createAgentClient(&TLS{Insecure: true})
	require.NoError(t, err)
	h := newHandler("traefik-hub-ep", 42, cfgChan, nil, client)

	traefikServer := httptest.NewServer(h)
	t.Cleanup(traefikServer.Close)

	req, err := http.NewRequest(http.MethodGet, traefikServer.URL+"/discover-ip", http.NoBody)
	require.NoError(t, err)

	q := make(url.Values)
	q.Set("port", strconv.Itoa(port))
	q.Set("nonce", nonce)
	req.URL.RawQuery = q.Encode()

	// Simulate a call from behind different proxies.
	req.Header.Add("X-Forwarded-For", "127.0.0.1")
	req.Header.Add("X-Forwarded-For", "10.10.0.13")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	defer func() {
		err = resp.Body.Close()
		require.NoError(t, err)
	}()

	assert.Equal(t, 1, handlerCallCount)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var ip string
	err = json.NewDecoder(resp.Body).Decode(&ip)
	require.NoError(t, err)

	assert.Equal(t, "127.0.0.1", ip)
}

func TestHandle_DiscoverIP_MethodNotAllowed(t *testing.T) {
	cfgChan := make(chan dynamic.Message, 1)
	client, err := createAgentClient(&TLS{Insecure: true})
	require.NoError(t, err)
	h := newHandler("traefik-hub-ep", 42, cfgChan, nil, client)

	server := httptest.NewServer(h)
	t.Cleanup(server.Close)

	resp, err := http.Post(server.URL+"/discover-ip", "", http.NoBody)
	require.NoError(t, err)

	err = resp.Body.Close()
	require.NoError(t, err)

	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}
