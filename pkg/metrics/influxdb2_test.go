package metrics

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v2/pkg/types"
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
	defer ts.Close()

	influxDB2Registry := RegisterInfluxDB2(context.Background(),
		&types.InfluxDB2{
			Address:              ts.URL,
			Token:                "test-token",
			BatchSize:            10,
			PushInterval:         ptypes.Duration(time.Millisecond * 10),
			Org:                  "test-org",
			Bucket:               "test-bucket",
			AddEntryPointsLabels: true,
			AddRoutersLabels:     true,
			AddServicesLabels:    true,
		})
	defer StopInfluxDB2()

	if !influxDB2Registry.IsEpEnabled() || !influxDB2Registry.IsRouterEnabled() || !influxDB2Registry.IsSvcEnabled() {
		t.Fatalf("InfluxDB2Registry should return true for IsEnabled(), IsRouterEnabled() and IsSvcEnabled()")
	}

	expectedServer := []string{
		`(traefik\.config\.reload\.total=1) [\d]{19}`,
		`(traefik\.config\.reload\.total\.failure=1) [\d]{19}`,
		`(traefik\.config\.reload\.lastSuccessTimestamp=1) [\d]{19}`,
		`(traefik\.config\.reload\.lastFailureTimestamp=1) [\d]{19}`,
	}

	influxDB2Registry.ConfigReloadsCounter().Add(1)
	influxDB2Registry.ConfigReloadsFailureCounter().Add(1)
	influxDB2Registry.LastConfigReloadSuccessGauge().Set(1)
	influxDB2Registry.LastConfigReloadFailureGauge().Set(1)
	msgServer := <-c

	assertMessage(t, *msgServer, expectedServer)

	expectedTLS := []string{
		`(key=value traefik\.tls\.certs\.notAfterTimestamp=1) [\d]{19}`,
	}

	influxDB2Registry.TLSCertsNotAfterTimestampGauge().With("key", "value").Set(1)
	msgTLS := <-c

	assertMessage(t, *msgTLS, expectedTLS)

	expectedEntrypoint := []string{
		`(entrypoint=test,code=200,method=GET traefik\.entrypoint\.requests\.total=1) [\d]{19}`,
		`(entrypoint=test,tls_version=foo,tls_cipher=bar traefik\.entrypoint\.requests\.tls\.total=1) [\d]{19}`,
		`(entrypoint=test traefik\.entrypoint\.request\.duration=10000) [\d]{19}`,
		`(entrypoint=test traefik\.entrypoint\.connections.open=1) [\d]{19}`,
	}

	influxDB2Registry.EntryPointReqsCounter().With("entrypoint", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
	influxDB2Registry.EntryPointReqsTLSCounter().With("entrypoint", "test", "tls_version", "foo", "tls_cipher", "bar").Add(1)
	influxDB2Registry.EntryPointReqDurationHistogram().With("entrypoint", "test").Observe(10000)
	influxDB2Registry.EntryPointOpenConnsGauge().With("entrypoint", "test").Set(1)
	msgEntrypoint := <-c

	assertMessage(t, *msgEntrypoint, expectedEntrypoint)

	expectedRouter := []string{
		`(router=demo,service=test,code=404,method=GET traefik\.router\.requests\.total=1) [\d]{19}`,
		`(router=demo,service=test,code=200,method=GET traefik\.router\.requests\.total=1) [\d]{19}`,
		`(router=demo,service=test,tls_version=foo,tls_cipher=bar traefik\.router\.requests\.tls\.total=1) [\d]{19}`,
		`(router=demo,service=test,code=200 traefik\.router\.request\.duration=10000) [\d]{19}`,
		`(router=demo,service=test traefik\.router\.connections.open=1) [\d]{19}`,
	}

	influxDB2Registry.RouterReqsCounter().With("router", "demo", "service", "test", "code", strconv.Itoa(http.StatusNotFound), "method", http.MethodGet).Add(1)
	influxDB2Registry.RouterReqsCounter().With("router", "demo", "service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
	influxDB2Registry.RouterReqsTLSCounter().With("router", "demo", "service", "test", "tls_version", "foo", "tls_cipher", "bar").Add(1)
	influxDB2Registry.RouterReqDurationHistogram().With("router", "demo", "service", "test", "code", strconv.Itoa(http.StatusOK)).Observe(10000)
	influxDB2Registry.RouterOpenConnsGauge().With("router", "demo", "service", "test").Set(1)
	msgRouter := <-c

	assertMessage(t, *msgRouter, expectedRouter)

	expectedService := []string{
		`(service=test,code=200,method=GET traefik\.service\.requests\.total=1) [\d]{19}`,
		`(service=test,code=404,method=GET traefik\.service\.requests\.total=1) [\d]{19}`,
		`(service=test,tls_version=foo,tls_cipher=bar traefik\.service\.requests\.tls\.total=1) [\d]{19}`,
		`(service=test,code=200 traefik\.service\.request\.duration=10000) [\d]{19}`,
		`(service=test,url=http://127.0.0.1 traefik\.service\.server\.up=1) [\d]{19}`,
	}

	influxDB2Registry.ServiceReqsCounter().With("service", "test", "code", strconv.Itoa(http.StatusOK), "method", http.MethodGet).Add(1)
	influxDB2Registry.ServiceReqsCounter().With("service", "test", "code", strconv.Itoa(http.StatusNotFound), "method", http.MethodGet).Add(1)
	influxDB2Registry.ServiceReqsTLSCounter().With("service", "test", "tls_version", "foo", "tls_cipher", "bar").Add(1)
	influxDB2Registry.ServiceReqDurationHistogram().With("service", "test", "code", strconv.Itoa(http.StatusOK)).Observe(10000)
	influxDB2Registry.ServiceServerUpGauge().With("service", "test", "url", "http://127.0.0.1").Set(1)
	msgService := <-c

	assertMessage(t, *msgService, expectedService)

	expectedServiceRetries := []string{
		`(service=test traefik\.service\.retries\.total=1) [\d]{19}`,
		`(service=test traefik\.service\.retries\.total=2) [\d]{19}`,
		`(service=foobar traefik\.service\.retries\.total=1) [\d]{19}`,
	}

	influxDB2Registry.ServiceRetriesCounter().With("service", "test").Add(1)
	influxDB2Registry.ServiceRetriesCounter().With("service", "test").Add(1)
	influxDB2Registry.ServiceRetriesCounter().With("service", "foobar").Add(1)

	msgServiceRetries := <-c

	assertMessage(t, *msgServiceRetries, expectedServiceRetries)

	expectedServiceOpenConns := []string{
		`(service=test traefik\.service\.connections\.open=1) [\d]{19}`,
		`(service=test traefik\.service\.connections\.open=2) [\d]{19}`,
		`(service=foobar traefik\.service\.connections\.open=1) [\d]{19}`,
	}

	influxDB2Registry.ServiceOpenConnsGauge().With("service", "test").Add(1)
	influxDB2Registry.ServiceOpenConnsGauge().With("service", "test").Add(1)
	influxDB2Registry.ServiceOpenConnsGauge().With("service", "foobar").Add(1)

	msgServiceOpenConns := <-c

	assertMessage(t, *msgServiceOpenConns, expectedServiceOpenConns)
}
