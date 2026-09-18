package metrics

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/containous/alice"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/middlewares/capture"
	"github.com/traefik/traefik/v3/pkg/middlewares/observability"
	obsmetrics "github.com/traefik/traefik/v3/pkg/observability/metrics"
	otypes "github.com/traefik/traefik/v3/pkg/observability/types"
)

// scrapeInflight reads the live value of
// traefik_service_inflight_requests{service=svc, method=method, protocol=proto}
// from the Prometheus text exposition served by metrics.PrometheusHandler().
//
// It returns 0 if the series is not present (which is what Prometheus would
// expose after the gauge decrements back to 0 — gauges aren't removed on
// hitting zero, but a series that never existed during this test won't be
// present until something writes to it).
func scrapeInflight(t *testing.T, svc, method, proto string) float64 {
	t.Helper()

	rec := httptest.NewRecorder()
	obsmetrics.PrometheusHandler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	require.Equal(t, http.StatusOK, rec.Code)

	// Match the line of the form:
	//   traefik_service_inflight_requests{method="GET",protocol="http",service="svc"} 1
	// label order is alphabetical in Prometheus text format.
	re := regexp.MustCompile(`(?m)^traefik_service_inflight_requests\{([^}]+)\}\s+([0-9eE+\-.]+)`)
	for _, line := range strings.Split(rec.Body.String(), "\n") {
		m := re.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		labels := m[1]
		if !strings.Contains(labels, `service="`+svc+`"`) {
			continue
		}
		if !strings.Contains(labels, `protocol="`+proto+`"`) {
			continue
		}
		if !strings.Contains(labels, `method="`+method+`"`) {
			continue
		}
		v, err := strconv.ParseFloat(m[2], 64)
		require.NoError(t, err)
		return v
	}
	return 0
}

// waitForInflight polls the gauge until it reaches want, or fails the test
// after deadline. Used for the post-close assertion where the server-side
// handler returning is racy with the client's Close().
func waitForInflight(t *testing.T, svc, method, proto string, want float64, deadline time.Duration) {
	t.Helper()

	timeout := time.After(deadline)
	for {
		if got := scrapeInflight(t, svc, method, proto); got == want {
			return
		}
		select {
		case <-timeout:
			t.Fatalf("traefik_service_inflight_requests{service=%q,method=%q,protocol=%q} did not reach %v within %v (last=%v)",
				svc, method, proto, want, deadline, scrapeInflight(t, svc, method, proto))
		case <-time.After(20 * time.Millisecond):
		}
	}
}

// metricsCtx returns a context that opts requests in to metrics collection,
// the same gate the production server-side observability middleware applies.
func metricsCtx(ctx context.Context) context.Context {
	return observability.WithObservability(ctx, observability.Observability{MetricsEnabled: true})
}

// buildServiceChain wraps the given backend in the production-style chain
// the per-service Alice chain uses: capture.Wrap → service metrics middleware.
func buildServiceChain(t *testing.T, registry obsmetrics.Registry, svc string, backend http.Handler) http.Handler {
	t.Helper()

	chain := alice.New().
		Append(capture.Wrap).
		Append(ServiceMetricsHandler(t.Context(), registry, svc))

	h, err := chain.Then(backend)
	require.NoError(t, err)

	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		h.ServeHTTP(rw, req.WithContext(metricsCtx(req.Context())))
	})
}

// TestServiceInflightRequestsGauge_HTTP asserts the gauge increments on entry
// and decrements when the request returns, for a plain HTTP request. The
// mid-flight assertion is deterministic — a release channel synchronises the
// scrape with the moment the backend is actively serving.
func TestServiceInflightRequestsGauge_HTTP(t *testing.T) {
	const svc = "inflight-test-http"

	reg := obsmetrics.RegisterPrometheus(t.Context(), &otypes.Prometheus{AddServicesLabels: true})
	require.NotNil(t, reg)

	entered := make(chan struct{})
	release := make(chan struct{})

	backend := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		close(entered)
		<-release
		rw.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewServer(buildServiceChain(t, reg, svc, backend))
	t.Cleanup(srv.Close)

	// Pre-flight: gauge is zero.
	assert.Equal(t, float64(0), scrapeInflight(t, svc, http.MethodGet, "http"))

	done := make(chan struct{})
	go func() {
		defer close(done)
		resp, err := http.Get(srv.URL)
		require.NoError(t, err)
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	<-entered
	// Mid-flight: gauge is 1.
	assert.Equal(t, float64(1), scrapeInflight(t, svc, http.MethodGet, "http"))

	close(release)
	<-done

	// Post-flight: gauge is back to 0.
	waitForInflight(t, svc, http.MethodGet, "http", 0, time.Second)
}

// TestServiceInflightRequestsGauge_WebSocket asserts that for an upgraded
// WebSocket session, the gauge:
//   - increments once the backend accepts the upgrade and starts reading,
//   - stays at 1 for the lifetime of the session (this is the gap left by PR
//     #9656 — the histogram-derived rate sees nothing here),
//   - decrements only after the WS handler returns (i.e. at close), not at
//     upgrade-completion.
func TestServiceInflightRequestsGauge_WebSocket(t *testing.T) {
	const svc = "inflight-test-ws"

	reg := obsmetrics.RegisterPrometheus(t.Context(), &otypes.Prometheus{AddServicesLabels: true})
	require.NotNil(t, reg)

	upgrader := websocket.Upgrader{
		CheckOrigin: func(_ *http.Request) bool { return true },
	}

	entered := make(chan struct{})

	backend := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		conn, err := upgrader.Upgrade(rw, req, nil)
		require.NoError(t, err)
		defer conn.Close()

		close(entered)

		// Block until the client closes — mirrors the fast-proxy upgrade
		// path (pkg/proxy/fast/upgrade.go) which waits on the copy
		// goroutines until the connection terminates.
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	})

	srv := httptest.NewServer(buildServiceChain(t, reg, svc, backend))
	t.Cleanup(srv.Close)

	wsURL, err := url.Parse(srv.URL)
	require.NoError(t, err)
	wsURL.Scheme = "ws"

	// Pre-flight: gauge is zero.
	assert.Equal(t, float64(0), scrapeInflight(t, svc, http.MethodGet, "websocket"))

	conn, _, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
	require.NoError(t, err)

	<-entered
	// Mid-session: gauge is 1. This is the case PR #9656's replacement
	// metric (rate of request_duration_seconds_sum) cannot see.
	assert.Equal(t, float64(1), scrapeInflight(t, svc, http.MethodGet, "websocket"))

	// Hold the session open across a scrape to confirm the value persists.
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, float64(1), scrapeInflight(t, svc, http.MethodGet, "websocket"))

	// Close from the client side; the server-side handler returns when
	// ReadMessage errors. The deferred decrement fires then.
	require.NoError(t, conn.Close())

	waitForInflight(t, svc, http.MethodGet, "websocket", 0, time.Second)
}
