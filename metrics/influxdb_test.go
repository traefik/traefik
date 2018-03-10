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

func TestInfluxDB(t *testing.T) {
	udp.SetAddr(":8089")
	// This is needed to make sure that UDP Listener listens for data a bit longer, otherwise it will quit after a millisecond
	udp.Timeout = 5 * time.Second

	influxDBRegistry := RegisterInfluxDB(&types.InfluxDB{Address: ":8089", PushInterval: "1s"})
	defer StopInfluxDB()

	if !influxDBRegistry.IsEnabled() {
		t.Fatalf("InfluxDB registry must be enabled")
	}

	expectedBackend := []string{
		`(traefik\.backend\.requests\.total,code=200,method=GET,service=test count=1) [\d]{19}`,
		`(traefik\.backend\.requests\.total,code=404,method=GET,service=test count=1) [\d]{19}`,
		`(traefik\.backend\.request\.duration(?:,backend=test)?,code=200,method=GET,service=test(?:,url=http://127.0.0.1)? p50=10000,p90=10000,p95=10000,p99=10000) [\d]{19}`,
		`(traefik\.backend\.retries\.total(?:,code=[\d]{3},method=GET)?,service=test count=2) [\d]{19}`,
		`(traefik\.config\.reload\.total(?:[a-z=0-9A-Z,]+)? count=1) [\d]{19}`,
		`(traefik\.config\.reload\.total\.failure(?:[a-z=0-9A-Z,]+)? count=1) [\d]{19}`,
		`(traefik\.backend\.server\.up,backend=test(?:[a-z=0-9A-Z,]+)?,url=http://127.0.0.1 value=1) [\d]{19}`,
	}

	msgBackend := udp.ReceiveString(t, func() {
		influxDBRegistry.BackendReqsCounter().With("service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
		influxDBRegistry.BackendReqsCounter().With("service", "test", "code", strconv.Itoa(http.StatusNotFound), "method", http.MethodGet).Add(1)
		influxDBRegistry.BackendRetriesCounter().With("service", "test").Add(1)
		influxDBRegistry.BackendRetriesCounter().With("service", "test").Add(1)
		influxDBRegistry.BackendReqDurationHistogram().With("service", "test", "code", strconv.Itoa(http.StatusOK)).Observe(10000)
		influxDBRegistry.ConfigReloadsCounter().Add(1)
		influxDBRegistry.ConfigReloadsFailureCounter().Add(1)
		influxDBRegistry.BackendServerUpGauge().With("backend", "test", "url", "http://127.0.0.1").Set(1)
	})

	assertMessage(t, msgBackend, expectedBackend)

	expectedEntrypoint := []string{
		`(traefik\.entrypoint\.requests\.total(?:,backend=test,code=[\d]{3})?,entrypoint=test(?:[a-z=0-9A-Z,:/.]+)? count=1) [\d]{19}`,
		`(traefik\.entrypoint\.request\.duration(?:,backend=test,code=[\d]{3})?,entrypoint=test(?:[a-z=0-9A-Z,:/.]+)? p50=10000,p90=10000,p95=10000,p99=10000) [\d]{19}`,
		`(traefik\.entrypoint\.connections\.open(?:[a-z=0-9A-Z,]+)?,entrypoint=test,(?:[a-z=0-9A-Z,:.//]+)? value=1) [\d]{19}`,
	}

	msgEntrypoint := udp.ReceiveString(t, func() {
		influxDBRegistry.EntrypointReqsCounter().With("entrypoint", "test").Add(1)
		influxDBRegistry.EntrypointReqDurationHistogram().With("entrypoint", "test").Observe(10000)
		influxDBRegistry.EntrypointOpenConnsGauge().With("entrypoint", "test").Set(1)

	})

	assertMessage(t, msgEntrypoint, expectedEntrypoint)
}

func assertMessage(t *testing.T, msg string, patterns []string) {
	t.Helper()
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(msg)
		if len(match) != 2 {
			t.Errorf("Got %q %v, want %q", msg, match, pattern)
		}
	}
}
