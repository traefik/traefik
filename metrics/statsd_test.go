package metrics

import (
	"net/http"
	"testing"
	"time"

	"github.com/containous/traefik/types"
	"github.com/stvp/go-udp-testing"
)

func TestStatsD(t *testing.T) {
	udp.SetAddr(":18125")
	// This is needed to make sure that UDP Listener listens for data a bit longer, otherwise it will quit after a millisecond
	udp.Timeout = 5 * time.Second

	statsdRegistry := RegisterStatsd(&types.Statsd{Address: ":18125", PushInterval: "1s"})
	defer StopStatsd()

	if !statsdRegistry.IsEnabled() {
		t.Errorf("Statsd registry should return true for IsEnabled()")
	}

	expected := []string{
		// We are only validating counts, as it is nearly impossible to validate latency, since it varies every run
		"traefik.backend.request.total:2.000000|c\n",
		"traefik.backend.retries.total:2.000000|c\n",
		"traefik.backend.request.duration:10000.000000|ms",
		"traefik.config.reload.total:1.000000|c\n",
		"traefik.config.reload.total:1.000000|c\n",
		"traefik.entrypoint.request.total:1.000000|c\n",
		"traefik.entrypoint.request.duration:10000.000000|ms",
		"traefik.entrypoint.connections.open:1.000000|g\n",
		"traefik.backend.server.up:1.000000|g\n",
	}

	udp.ShouldReceiveAll(t, expected, func() {
		statsdRegistry.BackendReqsCounter().With("service", "test", "code", string(http.StatusOK), "method", http.MethodGet).Add(1)
		statsdRegistry.BackendReqsCounter().With("service", "test", "code", string(http.StatusNotFound), "method", http.MethodGet).Add(1)
		statsdRegistry.BackendRetriesCounter().With("service", "test").Add(1)
		statsdRegistry.BackendRetriesCounter().With("service", "test").Add(1)
		statsdRegistry.BackendReqDurationHistogram().With("service", "test", "code", string(http.StatusOK)).Observe(10000)
		statsdRegistry.ConfigReloadsCounter().Add(1)
		statsdRegistry.ConfigReloadsFailureCounter().Add(1)
		statsdRegistry.EntrypointReqsCounter().With("entrypoint", "test").Add(1)
		statsdRegistry.EntrypointReqDurationHistogram().With("entrypoint", "test").Observe(10000)
		statsdRegistry.EntrypointOpenConnsGauge().With("entrypoint", "test").Set(1)
		statsdRegistry.BackendServerUpGauge().With("backend:test", "url", "http://127.0.0.1").Set(1)
	})
}
