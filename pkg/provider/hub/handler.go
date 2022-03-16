package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync/atomic"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
)

type handler struct {
	mux *http.ServeMux

	entryPoint string
	port       int

	// Accessed atomically.
	lastCfgUnixNano int64

	cfgChan chan<- dynamic.Message
}

func newHandler(entryPoint string, port int, cfgChan chan<- dynamic.Message) http.Handler {
	h := &handler{
		mux:        http.NewServeMux(),
		entryPoint: entryPoint,
		port:       port,
		cfgChan:    cfgChan,
	}

	h.mux.HandleFunc("/config", h.handleConfig)
	h.mux.HandleFunc("/discover-ip", h.handleDiscoverIP)
	h.mux.HandleFunc("/state", h.handleState)

	return h
}

type configRequest struct {
	UnixNano      int64                  `json:"unixNano"`
	Configuration *dynamic.Configuration `json:"configuration"`
}

func (h *handler) handleConfig(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	payload := &configRequest{Configuration: emptyDynamicConfiguration()}
	if err := json.NewDecoder(req.Body).Decode(payload); err != nil {
		err = fmt.Errorf("decode config request: %w", err)
		log.WithoutContext().Errorf("Handle config: %v", err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	atomic.StoreInt64(&h.lastCfgUnixNano, payload.UnixNano)

	cfg := payload.Configuration
	patchDynamicConfiguration(cfg, h.entryPoint, h.port)

	// We can safely drop messages here if the other end is not ready to receive them
	// as the agent will re-apply the same configuration.
	select {
	case h.cfgChan <- dynamic.Message{ProviderName: "hub", Configuration: cfg}:
	default:
	}
}

func (h *handler) handleDiscoverIP(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	xff := req.Header.Get("X-Forwarded-For")
	port := req.URL.Query().Get("port")
	nonce := req.URL.Query().Get("nonce")

	if err := doDiscoveryReq(req.Context(), xff, port, nonce); err != nil {
		err = fmt.Errorf("do discovery request: %w", err)
		log.WithoutContext().Errorf("Discover ip: %v", err)
		http.Error(rw, err.Error(), http.StatusBadGateway)
		return
	}

	if err := json.NewEncoder(rw).Encode(xff); err != nil {
		err = fmt.Errorf("encode discover ip response: %w", err)
		log.WithoutContext().Errorf("Discover ip: %v", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

func doDiscoveryReq(ctx context.Context, ip, port, nonce string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://%s:%s", ip, port), http.NoBody)
	if err != nil {
		return fmt.Errorf("make request: %w", err)
	}

	q := make(url.Values)
	q.Set("nonce", nonce)
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return nil
}

type stateResponse struct {
	LastConfigUnixNano int64 `json:"lastConfigUnixNano"`
}

func (h *handler) handleState(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	resp := stateResponse{
		LastConfigUnixNano: atomic.LoadInt64(&h.lastCfgUnixNano),
	}
	if err := json.NewEncoder(rw).Encode(resp); err != nil {
		err = fmt.Errorf("encode last config received response: %w", err)
		log.WithoutContext().Errorf("Last config received: %v", err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	h.mux.ServeHTTP(rw, req)
}
