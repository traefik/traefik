package pilot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/config/runtime"
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

	pilot := New("token", metrics.RegisterPilot(), safe.NewPool(context.Background()))
	pilot.client.baseURL = server.URL

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go pilot.Tick(ctx)

	pilot.SetRuntimeConfiguration(&runtime.Configuration{})
	pilot.SetRuntimeConfiguration(&runtime.Configuration{})

	select {
	case <-time.Tick(10 * time.Second):
		t.Fatal("Timeout")
	case <-receivedConfig:
		return
	}
}

func TestClient_SendConfiguration(t *testing.T) {
	myToken := "myToken"

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
		baseURL:    server.URL,
		httpClient: http.DefaultClient,
		token:      myToken,
	}

	err := client.SendData(context.Background(), RunTimeRepresentation{}, []metrics.PilotMetric{})
	require.NoError(t, err)
}
