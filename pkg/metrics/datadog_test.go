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

func TestDatadog(t *testing.T) {
	udp.SetAddr(":18125")
	// This is needed to make sure that UDP Listener listens for data a bit longer, otherwise it will quit after a millisecond
	udp.Timeout = 5 * time.Second

	datadogRegistry := RegisterDatadog(context.Background(), &types.Datadog{Address: ":18125", PushInterval: ptypes.Duration(time.Second), AddEntryPointsLabels: true, AddRoutersLabels: true, AddServicesLabels: true})
	defer StopDatadog()

	if !datadogRegistry.IsEpEnabled() || !datadogRegistry.IsRouterEnabled() || !datadogRegistry.IsSvcEnabled() {
		t.Errorf("DatadogRegistry should return true for IsEnabled(), IsRouterEnabled() and IsSvcEnabled()")
	}
	testDatadogRegistry(t, defaultMetricsPrefix, datadogRegistry)
}

func TestDatadogWithPrefix(t *testing.T) {
	t.Cleanup(func() {
		StopDatadog()
	})

	udp.SetAddr(":18125")
	// This is needed to make sure that UDP Listener listens for data a bit longer, otherwise it will quit after a millisecond
	udp.Timeout = 5 * time.Second

	datadogRegistry := RegisterDatadog(context.Background(), &types.Datadog{Prefix: "testPrefix", Address: ":18125", PushInterval: ptypes.Duration(time.Second), AddEntryPointsLabels: true, AddRoutersLabels: true, AddServicesLabels: true})

	testDatadogRegistry(t, "testPrefix", datadogRegistry)
}

func testDatadogRegistry(t *testing.T, metricsPrefix string, datadogRegistry Registry) {
	t.Helper()

	expected := []string{
		metricsPrefix + ".config.reload.total:1.000000|c\n",
		metricsPrefix + ".config.reload.total:1.000000|c|#failure:true\n",
		metricsPrefix + ".config.reload.lastSuccessTimestamp:1.000000|g\n",
		metricsPrefix + ".config.reload.lastFailureTimestamp:1.000000|g\n",

		metricsPrefix + ".tls.certs.notAfterTimestamp:1.000000|g|#key:value\n",

		metricsPrefix + ".entrypoint.request.total:1.000000|c|#entrypoint:test\n",
		metricsPrefix + ".entrypoint.request.tls.total:1.000000|c|#entrypoint:test,tls_version:foo,tls_cipher:bar\n",
		metricsPrefix + ".entrypoint.request.duration:10000.000000|h|#entrypoint:test\n",
		metricsPrefix + ".entrypoint.connections.open:1.000000|g|#entrypoint:test\n",
		metricsPrefix + ".entrypoint.requests.bytes.total:1.000000|c|#entrypoint:test\n",
		metricsPrefix + ".entrypoint.responses.bytes.total:1.000000|c|#entrypoint:test\n",

		metricsPrefix + ".router.request.total:1.000000|c|#router:demo,service:test,code:404,method:GET\n",
		metricsPrefix + ".router.request.total:1.000000|c|#router:demo,service:test,code:200,method:GET\n",
		metricsPrefix + ".router.request.tls.total:1.000000|c|#router:demo,service:test,tls_version:foo,tls_cipher:bar\n",
		metricsPrefix + ".router.request.duration:10000.000000|h|#router:demo,service:test,code:200\n",
		metricsPrefix + ".router.connections.open:1.000000|g|#router:demo,service:test\n",
		metricsPrefix + ".router.requests.bytes.total:1.000000|c|#router:demo,service:test,code:200,method:GET\n",
		metricsPrefix + ".router.responses.bytes.total:1.000000|c|#router:demo,service:test,code:200,method:GET\n",

		metricsPrefix + ".service.request.total:1.000000|c|#service:test,code:404,method:GET\n",
		metricsPrefix + ".service.request.total:1.000000|c|#service:test,code:200,method:GET\n",
		metricsPrefix + ".service.request.tls.total:1.000000|c|#service:test,tls_version:foo,tls_cipher:bar\n",
		metricsPrefix + ".service.request.duration:10000.000000|h|#service:test,code:200\n",
		metricsPrefix + ".service.connections.open:1.000000|g|#service:test\n",
		metricsPrefix + ".service.retries.total:2.000000|c|#service:test\n",
		metricsPrefix + ".service.request.duration:10000.000000|h|#service:test,code:200\n",
		metricsPrefix + ".service.server.up:1.000000|g|#service:test,url:http://127.0.0.1,one:two\n",
		metricsPrefix + ".service.requests.bytes.total:1.000000|c|#service:test,code:200,method:GET\n",
		metricsPrefix + ".service.responses.bytes.total:1.000000|c|#service:test,code:200,method:GET\n",
	}

	udp.ShouldReceiveAll(t, expected, func() {
		datadogRegistry.ConfigReloadsCounter().Add(1)
		datadogRegistry.ConfigReloadsFailureCounter().Add(1)
		datadogRegistry.LastConfigReloadSuccessGauge().Add(1)
		datadogRegistry.LastConfigReloadFailureGauge().Add(1)

		datadogRegistry.TLSCertsNotAfterTimestampGauge().With("key", "value").Set(1)

		datadogRegistry.EntryPointReqsCounter().With(nil, "entrypoint", "test").Add(1)
		datadogRegistry.EntryPointReqsTLSCounter().With("entrypoint", "test", "tls_version", "foo", "tls_cipher", "bar").Add(1)
		datadogRegistry.EntryPointReqDurationHistogram().With("entrypoint", "test").Observe(10000)
		datadogRegistry.EntryPointOpenConnsGauge().With("entrypoint", "test").Set(1)
		datadogRegistry.EntryPointReqsBytesCounter().With("entrypoint", "test").Add(1)
		datadogRegistry.EntryPointRespsBytesCounter().With("entrypoint", "test").Add(1)

		datadogRegistry.RouterReqsCounter().With(nil, "router", "demo", "service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
		datadogRegistry.RouterReqsCounter().With(nil, "router", "demo", "service", "test", "code", strconv.Itoa(http.StatusNotFound), "method", http.MethodGet).Add(1)
		datadogRegistry.RouterReqsTLSCounter().With("router", "demo", "service", "test", "tls_version", "foo", "tls_cipher", "bar").Add(1)
		datadogRegistry.RouterReqDurationHistogram().With("router", "demo", "service", "test", "code", strconv.Itoa(http.StatusOK)).Observe(10000)
		datadogRegistry.RouterOpenConnsGauge().With("router", "demo", "service", "test").Set(1)
		datadogRegistry.RouterReqsBytesCounter().With("router", "demo", "service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
		datadogRegistry.RouterRespsBytesCounter().With("router", "demo", "service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)

		datadogRegistry.ServiceReqsCounter().With(nil, "service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
		datadogRegistry.ServiceReqsCounter().With(nil, "service", "test", "code", strconv.Itoa(http.StatusNotFound), "method", http.MethodGet).Add(1)
		datadogRegistry.ServiceReqsTLSCounter().With("service", "test", "tls_version", "foo", "tls_cipher", "bar").Add(1)
		datadogRegistry.ServiceReqDurationHistogram().With("service", "test", "code", strconv.Itoa(http.StatusOK)).Observe(10000)
		datadogRegistry.ServiceOpenConnsGauge().With("service", "test").Set(1)
		datadogRegistry.ServiceRetriesCounter().With("service", "test").Add(1)
		datadogRegistry.ServiceRetriesCounter().With("service", "test").Add(1)
		datadogRegistry.ServiceServerUpGauge().With("service", "test", "url", "http://127.0.0.1", "one", "two").Set(1)
		datadogRegistry.ServiceReqsBytesCounter().With("service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
		datadogRegistry.ServiceRespsBytesCounter().With("service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
	})
}
