package metrics

import (
	"context"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stvp/go-udp-testing"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v2/pkg/types"
)

func TestStatsD(t *testing.T) {
	udp.SetAddr(":18125")
	// This is needed to make sure that UDP Listener listens for data a bit longer, otherwise it will quit after a millisecond
	udp.Timeout = 5 * time.Second

	statsdRegistry := RegisterStatsd(context.Background(), &types.Statsd{Address: ":18125", PushInterval: ptypes.Duration(time.Second), AddEntryPointsLabels: true, AddServicesLabels: true})
	defer StopStatsd()

	if !statsdRegistry.IsEpEnabled() || !statsdRegistry.IsSvcEnabled() {
		t.Errorf("Statsd registry should return true for IsEnabled()")
	}

	expected := []string{
		// We are only validating counts, as it is nearly impossible to validate latency, since it varies every run
		"traefik.service.request.total:2.000000|c\n",
		"traefik.service.retries.total:2.000000|c\n",
		"traefik.service.request.duration:10000.000000|ms",
		"traefik.config.reload.total:1.000000|c\n",
		"traefik.config.reload.total:1.000000|c\n",
		"traefik.entrypoint.request.total:1.000000|c\n",
		"traefik.entrypoint.request.duration:10000.000000|ms",
		"traefik.entrypoint.connections.open:1.000000|g\n",
		"traefik.service.server.up:1.000000|g\n",
	}

	udp.ShouldReceiveAll(t, expected, func() {
		statsdRegistry.ServiceReqsCounter().With("service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
		statsdRegistry.ServiceReqsCounter().With("service", "test", "code", strconv.Itoa(http.StatusNotFound), "method", http.MethodGet).Add(1)
		statsdRegistry.ServiceRetriesCounter().With("service", "test").Add(1)
		statsdRegistry.ServiceRetriesCounter().With("service", "test").Add(1)
		statsdRegistry.ServiceReqDurationHistogram().With("service", "test", "code", strconv.Itoa(http.StatusOK)).Observe(10000)
		statsdRegistry.ConfigReloadsCounter().Add(1)
		statsdRegistry.ConfigReloadsFailureCounter().Add(1)
		statsdRegistry.EntryPointReqsCounter().With("entrypoint", "test").Add(1)
		statsdRegistry.EntryPointReqDurationHistogram().With("entrypoint", "test").Observe(10000)
		statsdRegistry.EntryPointOpenConnsGauge().With("entrypoint", "test").Set(1)
		statsdRegistry.ServiceServerUpGauge().With("service:test", "url", "http://127.0.0.1").Set(1)
	})
}

func TestStatsDWithPrefix(t *testing.T) {
	udp.SetAddr(":18125")
	// This is needed to make sure that UDP Listener listens for data a bit longer, otherwise it will quit after a millisecond
	udp.Timeout = 5 * time.Second

	statsdRegistry := RegisterStatsd(context.Background(), &types.Statsd{Address: ":18125", PushInterval: ptypes.Duration(time.Second), AddEntryPointsLabels: true, AddServicesLabels: true, Prefix: "testPrefix"})
	defer StopStatsd()

	if !statsdRegistry.IsEpEnabled() || !statsdRegistry.IsSvcEnabled() {
		t.Errorf("Statsd registry should return true for IsEnabled()")
	}

	expected := []string{
		// We are only validating counts, as it is nearly impossible to validate latency, since it varies every run
		"testPrefix.service.request.total:2.000000|c\n",
		"testPrefix.service.retries.total:2.000000|c\n",
		"testPrefix.service.request.duration:10000.000000|ms",
		"testPrefix.config.reload.total:1.000000|c\n",
		"testPrefix.config.reload.total:1.000000|c\n",
		"testPrefix.entrypoint.request.total:1.000000|c\n",
		"testPrefix.entrypoint.request.duration:10000.000000|ms",
		"testPrefix.entrypoint.connections.open:1.000000|g\n",
		"testPrefix.service.server.up:1.000000|g\n",
	}

	udp.ShouldReceiveAll(t, expected, func() {
		statsdRegistry.ServiceReqsCounter().With("service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
		statsdRegistry.ServiceReqsCounter().With("service", "test", "code", strconv.Itoa(http.StatusNotFound), "method", http.MethodGet).Add(1)
		statsdRegistry.ServiceRetriesCounter().With("service", "test").Add(1)
		statsdRegistry.ServiceRetriesCounter().With("service", "test").Add(1)
		statsdRegistry.ServiceReqDurationHistogram().With("service", "test", "code", strconv.Itoa(http.StatusOK)).Observe(10000)
		statsdRegistry.ConfigReloadsCounter().Add(1)
		statsdRegistry.ConfigReloadsFailureCounter().Add(1)
		statsdRegistry.EntryPointReqsCounter().With("entrypoint", "test").Add(1)
		statsdRegistry.EntryPointReqDurationHistogram().With("entrypoint", "test").Observe(10000)
		statsdRegistry.EntryPointOpenConnsGauge().With("entrypoint", "test").Set(1)
		statsdRegistry.ServiceServerUpGauge().With("service:test", "url", "http://127.0.0.1").Set(1)
	})
}
