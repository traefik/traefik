package metrics

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/types"
)

func TestInfluxDB2(t *testing.T) {
	c := make(chan *string)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		bodyStr := string(body)
		c <- &bodyStr
		_, _ = fmt.Fprintln(w, "ok")
	}))

	influxDB2Registry := RegisterInfluxDB2(t.Context(),
		&types.InfluxDB2{
			Address:              ts.URL,
			Token:                "test-token",
			PushInterval:         ptypes.Duration(10 * time.Millisecond),
			Org:                  "test-org",
			Bucket:               "test-bucket",
			AddEntryPointsLabels: true,
			AddRoutersLabels:     true,
			AddServicesLabels:    true,
		})

	t.Cleanup(func() {
		StopInfluxDB2()
		ts.Close()
	})

	if !influxDB2Registry.IsEpEnabled() || !influxDB2Registry.IsRouterEnabled() || !influxDB2Registry.IsSvcEnabled() {
		t.Fatalf("InfluxDB2Registry should return true for IsEnabled(), IsRouterEnabled() and IsSvcEnabled()")
	}

	expectedServer := []string{
		`(traefik\.config\.reload\.total count=1) [\d]{19}`,
		`(traefik\.config\.reload\.lastSuccessTimestamp value=1) [\d]{19}`,
		`(traefik\.open\.connections,entrypoint=test,protocol=TCP value=1) [\d]{19}`,
	}

	influxDB2Registry.ConfigReloadsCounter().Add(1)
	influxDB2Registry.LastConfigReloadSuccessGauge().Set(1)
	influxDB2Registry.OpenConnectionsGauge().With("entrypoint", "test", "protocol", "TCP").Set(1)
	msgServer := <-c

	assertMessage(t, *msgServer, expectedServer)

	expectedTLS := []string{
		`(traefik\.tls\.certs\.notAfterTimestamp,key=value value=1) [\d]{19}`,
	}

	influxDB2Registry.TLSCertsNotAfterTimestampGauge().With("key", "value").Set(1)
	msgTLS := <-c

	assertMessage(t, *msgTLS, expectedTLS)

	expectedEntrypoint := []string{
		`(traefik\.entrypoint\.requests\.total,code=200,entrypoint=test,method=GET count=1) [\d]{19}`,
		`(traefik\.entrypoint\.requests\.tls\.total,entrypoint=test,tls_cipher=bar,tls_version=foo count=1) [\d]{19}`,
		`(traefik\.entrypoint\.request\.duration(?:,code=[\d]{3})?,entrypoint=test p50=10000,p90=10000,p95=10000,p99=10000) [\d]{19}`,
		`(traefik\.entrypoint\.requests\.bytes\.total,code=200,entrypoint=test,method=GET count=1) [\d]{19}`,
		`(traefik\.entrypoint\.responses\.bytes\.total,code=200,entrypoint=test,method=GET count=1) [\d]{19}`,
	}

	influxDB2Registry.EntryPointReqsCounter().With(nil, "entrypoint", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
	influxDB2Registry.EntryPointReqsTLSCounter().With("entrypoint", "test", "tls_version", "foo", "tls_cipher", "bar").Add(1)
	influxDB2Registry.EntryPointReqDurationHistogram().With("entrypoint", "test").Observe(10000)
	influxDB2Registry.EntryPointReqsBytesCounter().With("entrypoint", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
	influxDB2Registry.EntryPointRespsBytesCounter().With("entrypoint", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
	msgEntrypoint := <-c

	assertMessage(t, *msgEntrypoint, expectedEntrypoint)

	expectedRouter := []string{
		`(traefik\.router\.requests\.total,code=200,method=GET,router=demo,service=test count=1) [\d]{19}`,
		`(traefik\.router\.requests\.total,code=404,method=GET,router=demo,service=test count=1) [\d]{19}`,
		`(traefik\.router\.requests\.tls\.total,router=demo,service=test,tls_cipher=bar,tls_version=foo count=1) [\d]{19}`,
		`(traefik\.router\.request\.duration,code=200,router=demo,service=test p50=10000,p90=10000,p95=10000,p99=10000) [\d]{19}`,
		`(traefik\.router\.requests\.bytes\.total,code=200,method=GET,router=demo,service=test count=1) [\d]{19}`,
		`(traefik\.router\.responses\.bytes\.total,code=200,method=GET,router=demo,service=test count=1) [\d]{19}`,
	}

	influxDB2Registry.RouterReqsCounter().With(nil, "router", "demo", "service", "test", "code", strconv.Itoa(http.StatusNotFound), "method", http.MethodGet).Add(1)
	influxDB2Registry.RouterReqsCounter().With(nil, "router", "demo", "service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
	influxDB2Registry.RouterReqsTLSCounter().With("router", "demo", "service", "test", "tls_version", "foo", "tls_cipher", "bar").Add(1)
	influxDB2Registry.RouterReqDurationHistogram().With("router", "demo", "service", "test", "code", strconv.Itoa(http.StatusOK)).Observe(10000)
	influxDB2Registry.RouterReqsBytesCounter().With("router", "demo", "service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
	influxDB2Registry.RouterRespsBytesCounter().With("router", "demo", "service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
	msgRouter := <-c

	assertMessage(t, *msgRouter, expectedRouter)

	expectedService := []string{
		`(traefik\.service\.requests\.total,code=200,method=GET,service=test count=1) [\d]{19}`,
		`(traefik\.service\.requests\.total,code=404,method=GET,service=test count=1) [\d]{19}`,
		`(traefik\.service\.requests\.tls\.total,service=test,tls_cipher=bar,tls_version=foo count=1) [\d]{19}`,
		`(traefik\.service\.request\.duration,code=200,service=test p50=10000,p90=10000,p95=10000,p99=10000) [\d]{19}`,
		`(traefik\.service\.server\.up,service=test,url=http://127.0.0.1 value=1) [\d]{19}`,
		`(traefik\.service\.requests\.bytes\.total,code=200,method=GET,service=test count=1) [\d]{19}`,
		`(traefik\.service\.responses\.bytes\.total,code=200,method=GET,service=test count=1) [\d]{19}`,
	}

	influxDB2Registry.ServiceReqsCounter().With(nil, "service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
	influxDB2Registry.ServiceReqsCounter().With(nil, "service", "test", "code", strconv.Itoa(http.StatusNotFound), "method", http.MethodGet).Add(1)
	influxDB2Registry.ServiceReqsTLSCounter().With("service", "test", "tls_version", "foo", "tls_cipher", "bar").Add(1)
	influxDB2Registry.ServiceReqDurationHistogram().With("service", "test", "code", strconv.Itoa(http.StatusOK)).Observe(10000)
	influxDB2Registry.ServiceServerUpGauge().With("service", "test", "url", "http://127.0.0.1").Set(1)
	influxDB2Registry.ServiceReqsBytesCounter().With("service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
	influxDB2Registry.ServiceRespsBytesCounter().With("service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
	msgService := <-c

	assertMessage(t, *msgService, expectedService)

	expectedServiceRetries := []string{
		`(traefik\.service\.retries\.total,service=test count=2) [\d]{19}`,
		`(traefik\.service\.retries\.total,service=foobar count=1) [\d]{19}`,
	}

	influxDB2Registry.ServiceRetriesCounter().With("service", "test").Add(1)
	influxDB2Registry.ServiceRetriesCounter().With("service", "test").Add(1)
	influxDB2Registry.ServiceRetriesCounter().With("service", "foobar").Add(1)

	msgServiceRetries := <-c

	assertMessage(t, *msgServiceRetries, expectedServiceRetries)
}
