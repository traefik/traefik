package metrics

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stvp/go-udp-testing"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/types"
)

func TestStatsD(t *testing.T) {
	t.Cleanup(func() {
		StopStatsd()
	})

	udp.SetAddr(":18125")
	// This is needed to make sure that UDP Listener listens for data a bit longer, otherwise it will quit after a millisecond
	udp.Timeout = 5 * time.Second

	statsdRegistry := RegisterStatsd(t.Context(), &types.Statsd{Address: ":18125", PushInterval: ptypes.Duration(time.Second), AddEntryPointsLabels: true, AddRoutersLabels: true, AddServicesLabels: true})

	testRegistry(t, defaultMetricsPrefix, statsdRegistry)
}

func TestStatsDWithPrefix(t *testing.T) {
	t.Cleanup(func() {
		StopStatsd()
	})

	udp.SetAddr(":18125")
	// This is needed to make sure that UDP Listener listens for data a bit longer, otherwise it will quit after a millisecond
	udp.Timeout = 5 * time.Second

	statsdRegistry := RegisterStatsd(t.Context(), &types.Statsd{Address: ":18125", PushInterval: ptypes.Duration(time.Second), AddEntryPointsLabels: true, AddRoutersLabels: true, AddServicesLabels: true, Prefix: "testPrefix"})

	testRegistry(t, "testPrefix", statsdRegistry)
}

func testRegistry(t *testing.T, metricsPrefix string, registry Registry) {
	t.Helper()

	if !registry.IsEpEnabled() || !registry.IsRouterEnabled() || !registry.IsSvcEnabled() {
		t.Errorf("Statsd registry should return true for IsEnabled(), IsRouterEnabled() and IsSvcEnabled()")
	}

	expected := []string{
		metricsPrefix + ".config.reload.total:1.000000|c\n",
		metricsPrefix + ".config.reload.lastSuccessTimestamp:1.000000|g\n",
		metricsPrefix + ".open.connections:1.000000|g\n",

		metricsPrefix + ".tls.certs.notAfterTimestamp:1.000000|g\n",

		metricsPrefix + ".entrypoint.request.total:1.000000|c\n",
		metricsPrefix + ".entrypoint.request.tls.total:1.000000|c\n",
		metricsPrefix + ".entrypoint.request.duration:10000.000000|ms",
		metricsPrefix + ".entrypoint.requests.bytes.total:1.000000|c\n",
		metricsPrefix + ".entrypoint.responses.bytes.total:1.000000|c\n",

		metricsPrefix + ".router.request.total:2.000000|c\n",
		metricsPrefix + ".router.request.tls.total:1.000000|c\n",
		metricsPrefix + ".router.request.duration:10000.000000|ms",
		metricsPrefix + ".router.requests.bytes.total:1.000000|c\n",
		metricsPrefix + ".router.responses.bytes.total:1.000000|c\n",

		metricsPrefix + ".service.request.total:2.000000|c\n",
		metricsPrefix + ".service.request.tls.total:1.000000|c\n",
		metricsPrefix + ".service.request.duration:10000.000000|ms",
		metricsPrefix + ".service.retries.total:2.000000|c\n",
		metricsPrefix + ".service.server.up:1.000000|g\n",
		metricsPrefix + ".service.requests.bytes.total:1.000000|c\n",
		metricsPrefix + ".service.responses.bytes.total:1.000000|c\n",
	}

	udp.ShouldReceiveAll(t, expected, func() {
		registry.ConfigReloadsCounter().Add(1)
		registry.LastConfigReloadSuccessGauge().Set(1)
		registry.OpenConnectionsGauge().With("entrypoint", "test", "protocol", "TCP").Set(1)

		registry.TLSCertsNotAfterTimestampGauge().With("key", "value").Set(1)

		registry.EntryPointReqsCounter().With(nil, "entrypoint", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
		registry.EntryPointReqsTLSCounter().With("entrypoint", "test", "tls_version", "foo", "tls_cipher", "bar").Add(1)
		registry.EntryPointReqDurationHistogram().With("entrypoint", "test").Observe(10000)
		registry.EntryPointReqsBytesCounter().With("entrypoint", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
		registry.EntryPointRespsBytesCounter().With("entrypoint", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)

		registry.RouterReqsCounter().With(nil, "router", "demo", "service", "test", "code", strconv.Itoa(http.StatusNotFound), "method", http.MethodGet).Add(1)
		registry.RouterReqsCounter().With(nil, "router", "demo", "service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
		registry.RouterReqsTLSCounter().With("router", "demo", "service", "test", "tls_version", "foo", "tls_cipher", "bar").Add(1)
		registry.RouterReqDurationHistogram().With("router", "demo", "service", "test", "code", strconv.Itoa(http.StatusOK)).Observe(10000)
		registry.RouterReqsBytesCounter().With("router", "demo", "service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
		registry.RouterRespsBytesCounter().With("router", "demo", "service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)

		registry.ServiceReqsCounter().With(nil, "service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
		registry.ServiceReqsCounter().With(nil, "service", "test", "code", strconv.Itoa(http.StatusNotFound), "method", http.MethodGet).Add(1)
		registry.ServiceReqsTLSCounter().With("service", "test", "tls_version", "foo", "tls_cipher", "bar").Add(1)
		registry.ServiceReqDurationHistogram().With("service", "test", "code", strconv.Itoa(http.StatusOK)).Observe(10000)
		registry.ServiceRetriesCounter().With("service", "test").Add(1)
		registry.ServiceRetriesCounter().With("service", "test").Add(1)
		registry.ServiceServerUpGauge().With("service:test", "url", "http://127.0.0.1").Set(1)
		registry.ServiceReqsBytesCounter().With("service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
		registry.ServiceRespsBytesCounter().With("service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
	})
}
