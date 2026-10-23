package docker

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/safe"
)

func TestProviderReconnectsAfterEventRefreshError(t *testing.T) {
	t.Parallel()

	var versionCalls atomic.Int32
	var containerListCalls atomic.Int32
	var eventCalls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")

		switch {
		case req.URL.Path == "/_ping":
			rw.Header().Set("API-Version", "1.54")
			_, _ = fmt.Fprint(rw, "OK")

		case strings.HasSuffix(req.URL.Path, "/version"):
			versionCalls.Add(1)
			_, _ = fmt.Fprint(rw, `{"Version":"test","ApiVersion":"1.54","MinAPIVersion":"1.24"}`)

		case strings.HasSuffix(req.URL.Path, "/containers/json"):
			if containerListCalls.Add(1) == 2 {
				rw.WriteHeader(http.StatusInternalServerError)
				_, _ = fmt.Fprint(rw, `{"message":"transient list failure"}`)
				return
			}
			_, _ = fmt.Fprint(rw, `[]`)

		case strings.HasSuffix(req.URL.Path, "/events"):
			if eventCalls.Add(1) == 1 {
				_, _ = fmt.Fprintln(rw, `{"Type":"container","Action":"start","Actor":{"ID":"test"}}`)
			}
			if flusher, ok := rw.(http.Flusher); ok {
				flusher.Flush()
			}
			<-req.Context().Done()

		default:
			t.Errorf("unexpected Docker API request: %s %s", req.Method, req.URL.Path)
			http.NotFound(rw, req)
		}
	}))
	t.Cleanup(server.Close)

	provider := &Provider{}
	provider.SetDefaults()
	provider.Endpoint = "tcp://" + strings.TrimPrefix(server.URL, "http://")
	require.NoError(t, provider.Init())

	ctx, cancel := context.WithCancel(t.Context())
	pool := safe.NewPool(ctx)
	t.Cleanup(func() {
		cancel()
		pool.Stop()
	})

	require.NoError(t, provider.Provide(make(chan dynamic.Message, 2), pool))
	require.Eventually(t, func() bool {
		return containerListCalls.Load() >= 2
	}, 2*time.Second, 10*time.Millisecond)
	require.Eventually(t, func() bool {
		return versionCalls.Load() >= 2 && eventCalls.Load() >= 2
	}, 5*time.Second, 10*time.Millisecond)
}
