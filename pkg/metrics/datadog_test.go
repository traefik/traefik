package metrics

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stvp/go-udp-testing"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/types"
)

func TestDatadog(t *testing.T) {
	t.Cleanup(StopDatadog)

	udp.SetAddr(":18125")
	// This is needed to make sure that UDP Listener listens for data a bit longer, otherwise it will quit after a millisecond
	udp.Timeout = 5 * time.Second

	datadogRegistry := RegisterDatadog(t.Context(), &types.Datadog{Address: ":18125", PushInterval: ptypes.Duration(time.Second), AddEntryPointsLabels: true, AddRoutersLabels: true, AddServicesLabels: true})

	if !datadogRegistry.IsEpEnabled() || !datadogRegistry.IsRouterEnabled() || !datadogRegistry.IsSvcEnabled() {
		t.Errorf("DatadogRegistry should return true for IsEnabled(), IsRouterEnabled() and IsSvcEnabled()")
	}
	testDatadogRegistry(t, defaultMetricsPrefix, datadogRegistry)
}

func TestDatadogWithPrefix(t *testing.T) {
	t.Cleanup(StopDatadog)

	udp.SetAddr(":18125")
	// This is needed to make sure that UDP Listener listens for data a bit longer, otherwise it will quit after a millisecond
	udp.Timeout = 5 * time.Second

	datadogRegistry := RegisterDatadog(t.Context(), &types.Datadog{Prefix: "testPrefix", Address: ":18125", PushInterval: ptypes.Duration(time.Second), AddEntryPointsLabels: true, AddRoutersLabels: true, AddServicesLabels: true})

	testDatadogRegistry(t, "testPrefix", datadogRegistry)
}

func TestDatadog_parseDatadogAddress(t *testing.T) {
	tests := []struct {
		desc       string
		address    string
		expNetwork string
		expAddress string
	}{
		{
			desc:       "empty address",
			expNetwork: "udp",
			expAddress: "localhost:8125",
		},
		{
			desc:       "udp address",
			address:    "127.0.0.1:8080",
			expNetwork: "udp",
			expAddress: "127.0.0.1:8080",
		},
		{
			desc:       "unix address",
			address:    "unix:///path/to/datadog.socket",
			expNetwork: "unix",
			expAddress: "/path/to/datadog.socket",
		},
		{
			desc:       "unixgram address",
			address:    "unixgram:///path/to/datadog.socket",
			expNetwork: "unixgram",
			expAddress: "/path/to/datadog.socket",
		},
		{
			desc:       "unixstream address",
			address:    "unixstream:///path/to/datadog.socket",
			expNetwork: "unixstream",
			expAddress: "/path/to/datadog.socket",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			gotNetwork, gotAddress := parseDatadogAddress(test.address)
			assert.Equal(t, test.expNetwork, gotNetwork)
			assert.Equal(t, test.expAddress, gotAddress)
		})
	}
}

func testDatadogRegistry(t *testing.T, metricsPrefix string, datadogRegistry Registry) {
	t.Helper()

	expected := []string{
		metricsPrefix + ".config.reload.total:1.000000|c\n",
		metricsPrefix + ".config.reload.lastSuccessTimestamp:1.000000|g\n",
		metricsPrefix + ".open.connections:1.000000|g|#entrypoint:test,protocol:TCP\n",

		metricsPrefix + ".tls.certs.notAfterTimestamp:1.000000|g|#key:value\n",

		metricsPrefix + ".entrypoint.request.total:1.000000|c|#entrypoint:test\n",
		metricsPrefix + ".entrypoint.request.tls.total:1.000000|c|#entrypoint:test,tls_version:foo,tls_cipher:bar\n",
		metricsPrefix + ".entrypoint.request.duration:10000.000000|h|#entrypoint:test\n",
		metricsPrefix + ".entrypoint.requests.bytes.total:1.000000|c|#entrypoint:test\n",
		metricsPrefix + ".entrypoint.responses.bytes.total:1.000000|c|#entrypoint:test\n",

		metricsPrefix + ".router.request.total:1.000000|c|#router:demo,service:test,code:404,method:GET\n",
		metricsPrefix + ".router.request.total:1.000000|c|#router:demo,service:test,code:200,method:GET\n",
		metricsPrefix + ".router.request.tls.total:1.000000|c|#router:demo,service:test,tls_version:foo,tls_cipher:bar\n",
		metricsPrefix + ".router.request.duration:10000.000000|h|#router:demo,service:test,code:200\n",
		metricsPrefix + ".router.requests.bytes.total:1.000000|c|#router:demo,service:test,code:200,method:GET\n",
		metricsPrefix + ".router.responses.bytes.total:1.000000|c|#router:demo,service:test,code:200,method:GET\n",

		metricsPrefix + ".service.request.total:1.000000|c|#service:test,code:404,method:GET\n",
		metricsPrefix + ".service.request.total:1.000000|c|#service:test,code:200,method:GET\n",
		metricsPrefix + ".service.request.tls.total:1.000000|c|#service:test,tls_version:foo,tls_cipher:bar\n",
		metricsPrefix + ".service.request.duration:10000.000000|h|#service:test,code:200\n",
		metricsPrefix + ".service.retries.total:2.000000|c|#service:test\n",
		metricsPrefix + ".service.request.duration:10000.000000|h|#service:test,code:200\n",
		metricsPrefix + ".service.server.up:1.000000|g|#service:test,url:http://127.0.0.1,one:two\n",
		metricsPrefix + ".service.requests.bytes.total:1.000000|c|#service:test,code:200,method:GET\n",
		metricsPrefix + ".service.responses.bytes.total:1.000000|c|#service:test,code:200,method:GET\n",
	}

	udp.ShouldReceiveAll(t, expected, func() {
		datadogRegistry.ConfigReloadsCounter().Add(1)
		datadogRegistry.LastConfigReloadSuccessGauge().Add(1)
		datadogRegistry.OpenConnectionsGauge().With("entrypoint", "test", "protocol", "TCP").Add(1)

		datadogRegistry.TLSCertsNotAfterTimestampGauge().With("key", "value").Set(1)

		datadogRegistry.EntryPointReqsCounter().With(nil, "entrypoint", "test").Add(1)
		datadogRegistry.EntryPointReqsTLSCounter().With("entrypoint", "test", "tls_version", "foo", "tls_cipher", "bar").Add(1)
		datadogRegistry.EntryPointReqDurationHistogram().With("entrypoint", "test").Observe(10000)
		datadogRegistry.EntryPointReqsBytesCounter().With("entrypoint", "test").Add(1)
		datadogRegistry.EntryPointRespsBytesCounter().With("entrypoint", "test").Add(1)

		datadogRegistry.RouterReqsCounter().With(nil, "router", "demo", "service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
		datadogRegistry.RouterReqsCounter().With(nil, "router", "demo", "service", "test", "code", strconv.Itoa(http.StatusNotFound), "method", http.MethodGet).Add(1)
		datadogRegistry.RouterReqsTLSCounter().With("router", "demo", "service", "test", "tls_version", "foo", "tls_cipher", "bar").Add(1)
		datadogRegistry.RouterReqDurationHistogram().With("router", "demo", "service", "test", "code", strconv.Itoa(http.StatusOK)).Observe(10000)
		datadogRegistry.RouterReqsBytesCounter().With("router", "demo", "service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
		datadogRegistry.RouterRespsBytesCounter().With("router", "demo", "service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)

		datadogRegistry.ServiceReqsCounter().With(nil, "service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
		datadogRegistry.ServiceReqsCounter().With(nil, "service", "test", "code", strconv.Itoa(http.StatusNotFound), "method", http.MethodGet).Add(1)
		datadogRegistry.ServiceReqsTLSCounter().With("service", "test", "tls_version", "foo", "tls_cipher", "bar").Add(1)
		datadogRegistry.ServiceReqDurationHistogram().With("service", "test", "code", strconv.Itoa(http.StatusOK)).Observe(10000)
		datadogRegistry.ServiceRetriesCounter().With("service", "test").Add(1)
		datadogRegistry.ServiceRetriesCounter().With("service", "test").Add(1)
		datadogRegistry.ServiceServerUpGauge().With("service", "test", "url", "http://127.0.0.1", "one", "two").Set(1)
		datadogRegistry.ServiceReqsBytesCounter().With("service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
		datadogRegistry.ServiceRespsBytesCounter().With("service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
	})
}
