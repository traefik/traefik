package metrics

import (
	"net/http"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/containous/traefik/types"
	"github.com/stvp/go-udp-testing"
)

func TestInflux(t *testing.T) {
	udp.SetAddr(":8089")
	// This is needed to make sure that UDP Listener listens for data a bit longer, otherwise it will quit after a millisecond
	udp.Timeout = 5 * time.Second

	influxRegistry := RegisterInflux(&types.Influx{Address: "localhost:8089", PushInterval: "1s"})
	defer StopInflux()

	if !influxRegistry.IsEnabled() {
		t.Fatalf("Influx registry must be enabled")
	}

	expected := []string{
		"(traefik_requests_total,code=200,method=GET,service=test count=1) [0-9]{19}",
		"(traefik_requests_total,code=404,method=GET,service=test count=1) [0-9]{19}",
		"(traefik_request_duration,code=200,method=GET,service=test p50=10000,p90=10000,p95=10000,p99=10000) [0-9]{19}",
	}

	msg := udp.ReceiveString(t, func() {
		influxRegistry.ReqsCounter().With("service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
		influxRegistry.ReqsCounter().With("service", "test", "code", strconv.Itoa(http.StatusNotFound), "method", http.MethodGet).Add(1)
		influxRegistry.ReqDurationHistogram().With("service", "test", "code", strconv.Itoa(http.StatusOK)).Observe(10000)
	})

	extractAndMatchMessage(t, expected, msg)
}

func extractAndMatchMessage(t *testing.T, patterns []string, msg string) {
	t.Helper()
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(msg)
		if len(match) != 2 {
			t.Errorf("Got %q %v, want %q", msg, match, pattern)
		}
	}
}
