package hub

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"testing"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
)

func TestHandleConfig(t *testing.T) {
	cfgChan := make(chan dynamic.Message, 1)

	h := newHandler("traefik-hub-ep", 42, cfgChan)

	cfg := emptyDynamicConfiguration()
	cfg.HTTP.Routers["foo"] = &dynamic.Router{
		EntryPoints: []string{"ep"},
		Service:     "bar",
		Rule:        "Host(`foo.com`)",
	}

	req := configRequest{Configuration: cfg}
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Unable to serialize configuration: %v", err)
	}

	server := httptest.NewServer(h)
	t.Cleanup(server.Close)

	resp, err := http.Post(server.URL+"/config", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("Unable to send request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Invalid status code, expected: %d, got: %d", http.StatusOK, resp.StatusCode)
	}

	// NOTE: Not using select here as it doesn't work well with Yaegi yet.
	gotCfgRaw := <-cfgChan

	patchDynamicConfiguration(cfg, "traefik-hub-ep", 42)

	if !reflect.DeepEqual(cfg, gotCfgRaw.Configuration) {
		t.Fatalf("Configurations are not equal, expected: %v, got %v", cfg, gotCfgRaw)
	}
}

func TestHandle_Config_MethodNotAllowed(t *testing.T) {
	cfgChan := make(chan dynamic.Message, 1)
	h := newHandler("traefik-hub-ep", 42, cfgChan)

	server := httptest.NewServer(h)
	t.Cleanup(server.Close)

	resp, err := http.Get(server.URL + "/config")
	if err != nil {
		t.Fatalf("Unable to create req: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("Invalid status code, expected: %d, got: %d", http.StatusMethodNotAllowed, resp.StatusCode)
	}
}

func TestHandle_DiscoverIP(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Cannot listen: %v", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	nonce := "XVlBzgbaiCMRAjWw"

	mux := http.NewServeMux()

	var handlerCallCount int
	mux.HandleFunc("/", func(_ http.ResponseWriter, req *http.Request) {
		handlerCallCount++
		if req.URL.Query().Get("nonce") != nonce {
			t.Errorf("Wrong nonce: expected %q; got %q", nonce, req.URL.Query().Get("nonce"))
		}
	})

	s := &http.Server{Handler: mux}
	t.Cleanup(func() { _ = s.Close() })

	rdy := make(chan struct{})

	go func(s *http.Server) {
		close(rdy)
		if err = s.Serve(listener); errors.Is(err, http.ErrServerClosed) {
			return
		}
	}(s)

	<-rdy

	cfgChan := make(chan dynamic.Message, 1)
	h := newHandler("traefik-hub-ep", 42, cfgChan)

	server := httptest.NewServer(h)
	t.Cleanup(server.Close)

	req, err := http.NewRequest(http.MethodGet, server.URL+"/discover-ip", http.NoBody)
	if err != nil {
		t.Fatalf("Unable to create req: %v", err)
	}

	q := make(url.Values)
	q.Set("port", strconv.Itoa(port))
	q.Set("nonce", nonce)
	req.URL.RawQuery = q.Encode()

	// Simulate a call from behind different proxies.
	req.Header.Add("X-Forwarded-For", "127.0.0.1")
	req.Header.Add("X-Forwarded-For", "10.10.0.13")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Unable to send req: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if handlerCallCount != 1 {
		t.Fatalf("Expected handler to be called exactly once, got %d call(s)", handlerCallCount)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Invalid status code, expected: %d, got: %d", http.StatusOK, resp.StatusCode)
	}

	var ip string
	if err = json.NewDecoder(resp.Body).Decode(&ip); err != nil {
		t.Fatalf("Unable to decode response: %v", err)
	}

	if ip != "127.0.0.1" {
		t.Fatalf("expected ip: %s, got: %s", "127.0.0.1", ip)
	}
}

func TestHandle_DiscoverIP_MethodNotAllowed(t *testing.T) {
	cfgChan := make(chan dynamic.Message, 1)
	h := newHandler("traefik-hub-ep", 42, cfgChan)

	server := httptest.NewServer(h)
	t.Cleanup(server.Close)

	resp, err := http.Post(server.URL+"/discover-ip", "", http.NoBody)
	if err != nil {
		t.Fatalf("Unable to create req: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("Invalid status code, expected: %d, got: %d", http.StatusMethodNotAllowed, resp.StatusCode)
	}
}
