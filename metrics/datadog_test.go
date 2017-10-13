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
		"traefik.requests.total:1.000000|c|#service:test,code:404,method:GET\n",
		"traefik.requests.total:1.000000|c|#service:test,code:200,method:GET\n",
		"traefik.backend.retries.total:2.000000|c|#service:test\n",
		"traefik.request.duration:10000.000000|h|#service:test,code:200",
	}

	udp.ShouldReceiveAll(t, expected, func() {
		datadogRegistry.ReqsCounter().With("service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
		datadogRegistry.ReqsCounter().With("service", "test", "code", strconv.Itoa(http.StatusNotFound), "method", http.MethodGet).Add(1)
		datadogRegistry.ReqDurationHistogram().With("service", "test", "code", strconv.Itoa(http.StatusOK)).Observe(10000)
		datadogRegistry.RetriesCounter().With("service", "test").Add(1)
		datadogRegistry.RetriesCounter().With("service", "test").Add(1)
	})
}
