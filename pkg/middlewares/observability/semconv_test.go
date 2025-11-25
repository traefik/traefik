package observability

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/middlewares/capture"
	"github.com/traefik/traefik/v3/pkg/observability/metrics"
	otypes "github.com/traefik/traefik/v3/pkg/observability/types"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/metric/metricdata/metricdatatest"
)

type stringerAddr string

func (s stringerAddr) String() string {
	return string(s)
}

func TestSemConvServerMetrics(t *testing.T) {
	tests := []struct {
		desc           string
		statusCode     int
		url            string
		forwardedProto string
		route          string
		tls            bool
		localAddr      any
		wantScheme     string
		baseAttrs      []attribute.KeyValue
	}{
		{
			desc:           "http default port not found",
			statusCode:     http.StatusNotFound,
			url:            "http://example.com/search?q=Opentelemetry",
			forwardedProto: "http",
			route:          "Path(`/search`)",
			wantScheme:     "http",
			baseAttrs: []attribute.KeyValue{
				attribute.Key("error.type").String("404"),
				attribute.Key("http.request.method").String("GET"),
				attribute.Key("http.response.status_code").Int(404),
				attribute.Key("network.protocol.name").String("http/1.1"),
				attribute.Key("network.protocol.version").String("1.1"),
				attribute.Key("http.route").String("Path(`/search`)"),
				attribute.Key("server.address").String("example.com"),
				attribute.Key("server.port").Int(80),
			},
		},
		{
			desc:           "http default port created",
			statusCode:     http.StatusCreated,
			url:            "http://example.com/search?q=Opentelemetry",
			forwardedProto: "http",
			route:          "Path(`/search`)",
			wantScheme:     "http",
			baseAttrs: []attribute.KeyValue{
				attribute.Key("http.request.method").String("GET"),
				attribute.Key("http.response.status_code").Int(201),
				attribute.Key("network.protocol.name").String("http/1.1"),
				attribute.Key("network.protocol.version").String("1.1"),
				attribute.Key("http.route").String("Path(`/search`)"),
				attribute.Key("server.address").String("example.com"),
				attribute.Key("server.port").Int(80),
			},
		},
		{
			desc:           "https default port ok",
			statusCode:     http.StatusOK,
			url:            "https://example.com/secure",
			forwardedProto: "https",
			route:          "Path(`/secure`)",
			tls:            true,
			wantScheme:     "https",
			baseAttrs: []attribute.KeyValue{
				attribute.Key("http.request.method").String("GET"),
				attribute.Key("http.response.status_code").Int(200),
				attribute.Key("network.protocol.name").String("http/1.1"),
				attribute.Key("network.protocol.version").String("1.1"),
				attribute.Key("http.route").String("Path(`/secure`)"),
				attribute.Key("server.address").String("example.com"),
				attribute.Key("server.port").Int(443),
			},
		},
		{
			desc:       "explicit host port",
			statusCode: http.StatusOK,
			url:        "http://example.com:8080/search",
			route:      "Path(`/search`)",
			wantScheme: "",
			baseAttrs: []attribute.KeyValue{
				attribute.Key("http.request.method").String("GET"),
				attribute.Key("http.response.status_code").Int(200),
				attribute.Key("network.protocol.name").String("http/1.1"),
				attribute.Key("network.protocol.version").String("1.1"),
				attribute.Key("http.route").String("Path(`/search`)"),
				attribute.Key("server.address").String("example.com"),
				attribute.Key("server.port").Int(8080),
			},
		},
		{
			desc:       "ipv6 host explicit port",
			statusCode: http.StatusOK,
			url:        "http://[2001:db8::1]:8443/secure",
			route:      "Path(`/secure`)",
			wantScheme: "",
			baseAttrs: []attribute.KeyValue{
				attribute.Key("http.request.method").String("GET"),
				attribute.Key("http.response.status_code").Int(200),
				attribute.Key("network.protocol.name").String("http/1.1"),
				attribute.Key("network.protocol.version").String("1.1"),
				attribute.Key("http.route").String("Path(`/secure`)"),
				attribute.Key("server.address").String("2001:db8::1"),
				attribute.Key("server.port").Int(8443),
			},
		},
		{
			desc:       "ipv6 host default port",
			statusCode: http.StatusOK,
			url:        "http://[2001:db8::1]/secure",
			route:      "Path(`/secure`)",
			wantScheme: "",
			baseAttrs: []attribute.KeyValue{
				attribute.Key("http.request.method").String("GET"),
				attribute.Key("http.response.status_code").Int(200),
				attribute.Key("network.protocol.name").String("http/1.1"),
				attribute.Key("network.protocol.version").String("1.1"),
				attribute.Key("http.route").String("Path(`/secure`)"),
				attribute.Key("server.address").String("2001:db8::1"),
				attribute.Key("server.port").Int(80),
			},
		},
		{
			desc:           "forwarded proto overrides scheme",
			statusCode:     http.StatusOK,
			url:            "http://example.com/forwarded",
			forwardedProto: " HTTPS , http",
			route:          "Path(`/forwarded`)",
			wantScheme:     " HTTPS , http",
			baseAttrs: []attribute.KeyValue{
				attribute.Key("http.request.method").String("GET"),
				attribute.Key("http.response.status_code").Int(200),
				attribute.Key("network.protocol.name").String("http/1.1"),
				attribute.Key("network.protocol.version").String("1.1"),
				attribute.Key("http.route").String("Path(`/forwarded`)"),
				attribute.Key("server.address").String("example.com"),
				attribute.Key("server.port").Int(443),
			},
		},
		{
			desc:       "local addr from context",
			statusCode: http.StatusOK,
			url:        "http://example.com/context",
			localAddr:  &net.TCPAddr{IP: net.ParseIP("192.0.2.1"), Port: 9090},
			wantScheme: "",
			baseAttrs: []attribute.KeyValue{
				attribute.Key("http.request.method").String("GET"),
				attribute.Key("http.response.status_code").Int(200),
				attribute.Key("network.protocol.name").String("http/1.1"),
				attribute.Key("network.protocol.version").String("1.1"),
				attribute.Key("server.address").String("example.com"),
				attribute.Key("server.port").Int(9090),
			},
		},
		{
			desc:       "local addr string",
			statusCode: http.StatusOK,
			url:        "http://example.com/string",
			route:      "Path(`/string`)",
			localAddr:  ":7070",
			wantScheme: "",
			baseAttrs: []attribute.KeyValue{
				attribute.Key("http.request.method").String("GET"),
				attribute.Key("http.response.status_code").Int(200),
				attribute.Key("network.protocol.name").String("http/1.1"),
				attribute.Key("network.protocol.version").String("1.1"),
				attribute.Key("http.route").String("Path(`/string`)"),
				attribute.Key("server.address").String("example.com"),
				attribute.Key("server.port").Int(7070),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var cfg otypes.OTLP
			(&cfg).SetDefaults()
			cfg.AddRoutersLabels = true
			cfg.PushInterval = ptypes.Duration(10 * time.Millisecond)
			rdr := sdkmetric.NewManualReader()

			meterProvider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(rdr))
			metrics.SetMeterProvider(meterProvider)

			semConvMetricRegistry, err := metrics.NewSemConvMetricRegistry(t.Context(), &cfg)
			require.NoError(t, err)
			require.NotNil(t, semConvMetricRegistry)

			req := httptest.NewRequest(http.MethodGet, test.url, nil)
			rw := httptest.NewRecorder()
			req.RemoteAddr = "10.0.0.1:1234"
			req.Header.Set("User-Agent", "entrypoint-test")
			if test.forwardedProto != "" {
				req.Header.Set("X-Forwarded-Proto", test.forwardedProto)
			}

			if test.tls {
				req.TLS = &tls.ConnectionState{}
			}

			if test.route != "" {
				req = req.WithContext(WithHTTPRoute(req.Context(), test.route))
			}

			if test.localAddr != nil {
				req = req.WithContext(context.WithValue(req.Context(), http.LocalAddrContextKey, test.localAddr))
			}

			next := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
				rw.WriteHeader(test.statusCode)
			})

			handler := newServerMetricsSemConv(t.Context(), semConvMetricRegistry, next)
			handler, err = capture.Wrap(handler)
			require.NoError(t, err)

			handler = WithObservabilityHandler(handler, Observability{SemConvMetricsEnabled: true})
			handler.ServeHTTP(rw, req)

			attrs := append([]attribute.KeyValue(nil), test.baseAttrs...)
			attrs = append(attrs, attribute.Key("url.scheme").String(test.wantScheme))
			attrSet := attribute.NewSet(attrs...)

			got := metricdata.ResourceMetrics{}
			err = rdr.Collect(t.Context(), &got)
			require.NoError(t, err)
			require.Len(t, got.ScopeMetrics, 1)

			expected := metricdata.Metrics{
				Name:        "http.server.request.duration",
				Description: "Duration of HTTP server requests.",
				Unit:        "s",
				Data: metricdata.Histogram[float64]{
					DataPoints: []metricdata.HistogramDataPoint[float64]{
						{
							Attributes:   attrSet,
							Count:        1,
							Bounds:       []float64{0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10},
							BucketCounts: []uint64{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
							Min:          metricdata.NewExtrema[float64](1),
							Max:          metricdata.NewExtrema[float64](1),
							Sum:          1,
						},
					},
					Temporality: metricdata.CumulativeTemporality,
				},
			}

			metricdatatest.AssertEqual[metricdata.Metrics](t, expected, got.ScopeMetrics[0].Metrics[0], metricdatatest.IgnoreTimestamp(), metricdatatest.IgnoreValue())
		})
	}
}

func TestServerAddress(t *testing.T) {
	tests := []struct {
		desc string
		host string
		want string
	}{
		{desc: "empty host", host: "", want: ""},
		{desc: "hostname", host: "example.com", want: "example.com"},
		{desc: "host with port", host: "example.com:8443", want: "example.com"},
		{desc: "ipv6", host: "[2001:db8::1]", want: "2001:db8::1"},
		{desc: "ipv6 with port", host: "[2001:db8::1]:8080", want: "2001:db8::1"},
		{desc: "ipv6 zone", host: "[fe80::1%eth0]:9090", want: "fe80::1%eth0"},
		{desc: "wildcard", host: "*:8080", want: "*"},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
			req.Host = tc.host
			require.Equal(t, tc.want, serverAddress(req))
		})
	}
}

func TestServerPort(t *testing.T) {
	tests := []struct {
		desc           string
		url            string
		forwardedProto string
		tls            bool
		localAddr      any
		want           int
		ok             bool
	}{
		{desc: "explicit host port", url: "http://example.com:8081/test", want: 8081, ok: true},
		{desc: "ipv6 with port", url: "http://[2001:db8::1]:8443/test", want: 8443, ok: true},
		{desc: "tls default", url: "http://example.com/test", tls: true, want: 443, ok: true},
		{desc: "forwarded proto", url: "http://example.com/test", forwardedProto: "HTTPS", want: 443, ok: true},
		{desc: "forwarded multi", url: "http://example.com/test", forwardedProto: " https , http", want: 443, ok: true},
		{desc: "local addr net", url: "http://example.com/test", localAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9090}, want: 9090, ok: true},
		{desc: "local addr string", url: "http://example.com/test", localAddr: ":7070", want: 7070, ok: true},
		{desc: "local addr stringer", url: "http://example.com/test", localAddr: stringerAddr(":6060"), want: 6060, ok: true},
		{desc: "default http", url: "http://example.com/test", want: 80, ok: true},
		{desc: "invalid string", url: "http://example.com/test", localAddr: "invalid", want: 80, ok: true},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			if tc.forwardedProto != "" {
				req.Header.Set("X-Forwarded-Proto", tc.forwardedProto)
			}
			if tc.tls {
				req.TLS = &tls.ConnectionState{}
			}
			if tc.localAddr != nil {
				req = req.WithContext(context.WithValue(req.Context(), http.LocalAddrContextKey, tc.localAddr))
			}

			port, ok := serverPort(req)
			if !tc.ok {
				require.False(t, ok)
				return
			}

			require.True(t, ok)
			require.Equal(t, tc.want, port)
		})
	}
}
