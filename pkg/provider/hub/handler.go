package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync/atomic"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

type handler struct {
	mux *http.ServeMux

	client http.Client

	entryPoint string
	port       int
	tlsCfg     *TLS

	// Accessed atomically.
	lastCfgUnixNano int64

	cfgChan chan<- dynamic.Message
}

func newHandler(entryPoint string, port int, cfgChan chan<- dynamic.Message, tlsCfg *TLS, client http.Client) http.Handler {
	h := &handler{
		mux:        http.NewServeMux(),
		entryPoint: entryPoint,
		port:       port,
		cfgChan:    cfgChan,
		tlsCfg:     tlsCfg,
		client:     client,
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
		err = fmt.Errorf("decoding config request: %w", err)
		log.Error().Err(err).Msg("Handling config")
		http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	cfg := payload.Configuration
	patchDynamicConfiguration(cfg, h.entryPoint, h.port, h.tlsCfg)

	// We can safely drop messages here if the other end is not ready to receive them
	// as the agent will re-apply the same configuration.
	select {
	case h.cfgChan <- dynamic.Message{ProviderName: "hub", Configuration: cfg}:
		atomic.StoreInt64(&h.lastCfgUnixNano, payload.UnixNano)
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

	if err := h.doDiscoveryReq(req.Context(), xff, port, nonce); err != nil {
		err = fmt.Errorf("doing discovery request: %w", err)
		log.Error().Err(err).Msg("Handling IP discovery")
		http.Error(rw, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
		return
	}

	if err := json.NewEncoder(rw).Encode(xff); err != nil {
		err = fmt.Errorf("encoding discover ip response: %w", err)
		log.Error().Err(err).Msg("Handling IP discovery")
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (h *handler) doDiscoveryReq(ctx context.Context, ip, port, nonce string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://%s", net.JoinHostPort(ip, port)), http.NoBody)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	q := make(url.Values)
	q.Set("nonce", nonce)
	req.URL.RawQuery = q.Encode()
	req.Host = "agent.traefik"

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("doing request: %w", err)
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
		err = fmt.Errorf("encoding last config received response: %w", err)
		log.Error().Err(err).Msg("Handling state")
		http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
}

func (h *handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	h.mux.ServeHTTP(rw, req)
}
