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
		t.Errorf("PrometheusRegistry should return true for IsEnabled()")
	}

	expected := []string{
		// We are only validating counts, as it is nearly impossible to validate latency, since it varies every run
		"traefik.requests.total:2.000000|c\n",
		"traefik.backend.retries.total:2.000000|c\n",
		"traefik.request.duration:10000.000000|ms",
	}

	udp.ShouldReceiveAll(t, expected, func() {
		statsdRegistry.ReqsCounter().With("service", "test", "code", string(http.StatusOK), "method", http.MethodGet).Add(1)
		statsdRegistry.ReqsCounter().With("service", "test", "code", string(http.StatusNotFound), "method", http.MethodGet).Add(1)
		statsdRegistry.RetriesCounter().With("service", "test").Add(1)
		statsdRegistry.RetriesCounter().With("service", "test").Add(1)
		statsdRegistry.ReqDurationHistogram().With("service", "test", "code", string(http.StatusOK)).Observe(10000)
	})
}
