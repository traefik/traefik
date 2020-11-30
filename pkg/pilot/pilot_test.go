package pilot

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/metrics"
	"github.com/traefik/traefik/v2/pkg/safe"
)

func TestTick(t *testing.T) {
	receivedConfig := make(chan bool)

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, "invalid method", http.StatusMethodNotAllowed)
			return
		}

		err := json.NewEncoder(rw).Encode(instanceInfo{ID: "123"})
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("/command", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, "invalid method", http.StatusMethodNotAllowed)
			return
		}

		receivedConfig <- true
	})

	pilot, err := New("token", metrics.RegisterPilot(), safe.NewPool(context.Background()))
	require.NoError(t, err)

	pilot.client.baseInstanceInfoURL = server.URL

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go pilot.Tick(ctx)

	pilot.SetDynamicConfiguration(dynamic.Configuration{})
	pilot.SetDynamicConfiguration(dynamic.Configuration{})

	select {
	case <-time.Tick(10 * time.Second):
		t.Fatal("Timeout")
	case <-receivedConfig:
		return
	}
}

func TestClient_SendInstanceInfo(t *testing.T) {
	myToken := "myToken"

	myTokenHash, err := hashToken(myToken)
	require.NoError(t, err)

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, "invalid method", http.StatusMethodNotAllowed)
			return
		}

		tk := req.Header.Get(tokenHeader)
		if tk != myToken {
			http.Error(rw, fmt.Sprintf("invalid token: %s", tk), http.StatusUnauthorized)
		}

		err := json.NewEncoder(rw).Encode(instanceInfo{ID: "123"})
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("/command", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, "invalid method", http.StatusMethodNotAllowed)
			return
		}

		tk := req.Header.Get(tokenHeader)
		if tk != myToken {
			http.Error(rw, fmt.Sprintf("invalid token: %s", tk), http.StatusUnauthorized)
		}

		tkh := req.Header.Get(tokenHashHeader)
		if tkh != myTokenHash {
			http.Error(rw, fmt.Sprintf("invalid token hash: %s", tkh), http.StatusBadRequest)
		}

		defer req.Body.Close()

		info := &instanceInfo{}
		err := json.NewDecoder(req.Body).Decode(info)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		if info.ID != "123" {
			http.Error(rw, fmt.Sprintf("invalid ID: %s", info.ID), http.StatusBadRequest)
		}
	})

	client := client{
		baseInstanceInfoURL: server.URL,
		httpClient:          http.DefaultClient,
		token:               myToken,
		tokenHash:           myTokenHash,
	}

	err = client.SendInstanceInfo(context.Background(), []metrics.PilotMetric{})
	require.NoError(t, err)
}

func TestClient_SendTelemetry(t *testing.T) {
	myToken := "myToken"

	myTokenHash, err := hashToken(myToken)
	require.NoError(t, err)

	mux := http.NewServeMux()
	mux.HandleFunc("/collect", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, "invalid method", http.StatusMethodNotAllowed)
			return
		}

		tk := req.Header.Get(tokenHeader)
		if tk != myToken {
			http.Error(rw, fmt.Sprintf("invalid token: %s", tk), http.StatusUnauthorized)
		}

		tkh := req.Header.Get(tokenHashHeader)
		if tkh != myTokenHash {
			http.Error(rw, fmt.Sprintf("invalid token hash: %s", tkh), http.StatusBadRequest)
		}

		defer req.Body.Close()

		config := &dynamic.Configuration{}
		err := json.NewDecoder(req.Body).Decode(config)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		router, exists := config.HTTP.Routers["foo"]
		if !exists {
			http.Error(rw, "router configuration is missing", http.StatusBadRequest)
			return
		}

		if router.Rule != "xxxx" {
			http.Error(rw, fmt.Sprintf("configuration is not anonymized, got router rule %s", router.Rule), http.StatusBadRequest)
		}
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := client{
		baseTelemetryURL: server.URL,
		httpClient:       http.DefaultClient,
		token:            myToken,
		tokenHash:        myTokenHash,
	}

	config := dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers: map[string]*dynamic.Router{
				"foo": {
					Service: "foo",
					Rule:    "foo.com",
				},
			},
		},
	}

	err = client.SendTelemetry(context.Background(), config)
	require.NoError(t, err)
}

func hashToken(token string) (string, error) {
	tokenHash := fnv.New64a()

	_, err := tokenHash.Write([]byte(token))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", tokenHash.Sum64()), nil
}
