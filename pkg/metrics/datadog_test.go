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

	datadogRegistry := RegisterDatadog(context.Background(), &types.Datadog{Address: ":18125", PushInterval: ptypes.Duration(time.Second), AddEntryPointsLabels: true, AddServicesLabels: true})
	defer StopDatadog()

	if !datadogRegistry.IsEpEnabled() || !datadogRegistry.IsSvcEnabled() {
		t.Errorf("DatadogRegistry should return true for IsEnabled()")
	}

	expected := []string{
		// We are only validating counts, as it is nearly impossible to validate latency, since it varies every run
		"traefik.service.request.total:1.000000|c|#service:test,code:404,method:GET\n",
		"traefik.service.request.total:1.000000|c|#service:test,code:200,method:GET\n",
		"traefik.service.retries.total:2.000000|c|#service:test\n",
		"traefik.service.request.duration:10000.000000|h|#service:test,code:200\n",
		"traefik.config.reload.total:1.000000|c\n",
		"traefik.config.reload.total:1.000000|c|#failure:true\n",
		"traefik.entrypoint.request.total:1.000000|c|#entrypoint:test\n",
		"traefik.entrypoint.request.duration:10000.000000|h|#entrypoint:test\n",
		"traefik.entrypoint.connections.open:1.000000|g|#entrypoint:test\n",
		"traefik.service.server.up:1.000000|g|#service:test,url:http://127.0.0.1,one:two\n",
	}

	udp.ShouldReceiveAll(t, expected, func() {
		datadogRegistry.ServiceReqsCounter().With("service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
		datadogRegistry.ServiceReqsCounter().With("service", "test", "code", strconv.Itoa(http.StatusNotFound), "method", http.MethodGet).Add(1)
		datadogRegistry.ServiceReqDurationHistogram().With("service", "test", "code", strconv.Itoa(http.StatusOK)).Observe(10000)
		datadogRegistry.ServiceRetriesCounter().With("service", "test").Add(1)
		datadogRegistry.ServiceRetriesCounter().With("service", "test").Add(1)
		datadogRegistry.ConfigReloadsCounter().Add(1)
		datadogRegistry.ConfigReloadsFailureCounter().Add(1)
		datadogRegistry.EntryPointReqsCounter().With("entrypoint", "test").Add(1)
		datadogRegistry.EntryPointReqDurationHistogram().With("entrypoint", "test").Observe(10000)
		datadogRegistry.EntryPointOpenConnsGauge().With("entrypoint", "test").Set(1)
		datadogRegistry.ServiceServerUpGauge().With("service", "test", "url", "http://127.0.0.1", "one", "two").Set(1)
	})
}
