package metrics

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/containous/traefik/types"
	"github.com/stvp/go-udp-testing"
)

func TestDatadog(t *testing.T) {
	udp.SetAddr(":18125")
	// This is needed to make sure that UDP Listener listens for data a bit longer, otherwise it will quit after a millisecond
	udp.Timeout = 5 * time.Second

	datadogRegistry := RegisterDatadog(&types.Datadog{Address: ":18125", PushInterval: "1s"})
	defer StopDatadog()

	if !datadogRegistry.IsEnabled() {
		t.Errorf("DatadogRegistry should return true for IsEnabled()")
	}

	expected := []string{
		// We are only validating counts, as it is nearly impossible to validate latency, since it varies every run
		"traefik.backend.request.total:1.000000|c|#service:test,code:404,method:GET\n",
		"traefik.backend.request.total:1.000000|c|#service:test,code:200,method:GET\n",
		"traefik.backend.retries.total:2.000000|c|#service:test\n",
		"traefik.backend.request.duration:10000.000000|h|#service:test,code:200\n",
		"traefik.config.reload.total:1.000000|c\n",
		"traefik.config.reload.total:1.000000|c|#failure:true\n",
		"traefik.entrypoint.request.total:1.000000|c|#entrypoint:test\n",
		"traefik.entrypoint.request.duration:10000.000000|h|#entrypoint:test\n",
		"traefik.entrypoint.connections.open:1.000000|g|#entrypoint:test\n",
		"traefik.backend.server.up:1.000000|g|#backend:test,url:http://127.0.0.1,one:two\n",
	}

	udp.ShouldReceiveAll(t, expected, func() {
		datadogRegistry.BackendReqsCounter().With("service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
		datadogRegistry.BackendReqsCounter().With("service", "test", "code", strconv.Itoa(http.StatusNotFound), "method", http.MethodGet).Add(1)
		datadogRegistry.BackendReqDurationHistogram().With("service", "test", "code", strconv.Itoa(http.StatusOK)).Observe(10000)
		datadogRegistry.BackendRetriesCounter().With("service", "test").Add(1)
		datadogRegistry.BackendRetriesCounter().With("service", "test").Add(1)
		datadogRegistry.ConfigReloadsCounter().Add(1)
		datadogRegistry.ConfigReloadsFailureCounter().Add(1)
		datadogRegistry.EntrypointReqsCounter().With("entrypoint", "test").Add(1)
		datadogRegistry.EntrypointReqDurationHistogram().With("entrypoint", "test").Observe(10000)
		datadogRegistry.EntrypointOpenConnsGauge().With("entrypoint", "test").Set(1)
		datadogRegistry.BackendServerUpGauge().With("backend", "test", "url", "http://127.0.0.1", "one", "two").Set(1)
	})
}
