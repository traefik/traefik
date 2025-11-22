package observability

import (
	"context"
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

func TestSemConvServerMetrics(t *testing.T) {
	tests := []struct {
		desc           string
		statusCode     int
		host           string
		httpRoute      string
		localAddr      net.Addr
		wantAttributes attribute.Set
	}{
		{
			desc:       "not found status",
			statusCode: http.StatusNotFound,
			host:       "www.test.com",
			wantAttributes: attribute.NewSet(
				attribute.Key("error.type").String("404"),
				attribute.Key("http.request.method").String("GET"),
				attribute.Key("http.response.status_code").Int(404),
				attribute.Key("network.protocol.name").String("http/1.1"),
				attribute.Key("network.protocol.version").String("1.1"),
				attribute.Key("server.address").String("www.test.com"),
				attribute.Key("url.scheme").String("http"),
			),
		},
		{
			desc:       "created status",
			statusCode: http.StatusCreated,
			host:       "www.test.com",
			wantAttributes: attribute.NewSet(
				attribute.Key("http.request.method").String("GET"),
				attribute.Key("http.response.status_code").Int(201),
				attribute.Key("network.protocol.name").String("http/1.1"),
				attribute.Key("network.protocol.version").String("1.1"),
				attribute.Key("server.address").String("www.test.com"),
				attribute.Key("url.scheme").String("http"),
			),
		},
		{
			desc:       "with http.route and server.port from Host header",
			statusCode: http.StatusOK,
			host:       "example.com:443",
			httpRoute:  "/api/banking",
			wantAttributes: attribute.NewSet(
				attribute.Key("http.request.method").String("GET"),
				attribute.Key("http.response.status_code").Int(200),
				attribute.Key("http.route").String("/api/banking"),
				attribute.Key("network.protocol.name").String("http/1.1"),
				attribute.Key("network.protocol.version").String("1.1"),
				attribute.Key("server.address").String("example.com:443"),
				attribute.Key("server.port").Int(443),
				attribute.Key("url.scheme").String("http"),
			),
		},
		{
			desc:       "default HTTPS port 443 from local address",
			statusCode: http.StatusOK,
			host:       "example.com",
			localAddr:  &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: 443},
			wantAttributes: attribute.NewSet(
				attribute.Key("http.request.method").String("GET"),
				attribute.Key("http.response.status_code").Int(200),
				attribute.Key("network.protocol.name").String("http/1.1"),
				attribute.Key("network.protocol.version").String("1.1"),
				attribute.Key("server.address").String("example.com"),
				attribute.Key("server.port").Int(443),
				attribute.Key("url.scheme").String("https"),
			),
		},
		{
			desc:       "default HTTP port 80 from local address",
			statusCode: http.StatusOK,
			host:       "example.com",
			localAddr:  &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: 80},
			wantAttributes: attribute.NewSet(
				attribute.Key("http.request.method").String("GET"),
				attribute.Key("http.response.status_code").Int(200),
				attribute.Key("network.protocol.name").String("http/1.1"),
				attribute.Key("network.protocol.version").String("1.1"),
				attribute.Key("server.address").String("example.com"),
				attribute.Key("server.port").Int(80),
				attribute.Key("url.scheme").String("http"),
			),
		},
		{
			desc:       "non-standard port 8443 from local address with http.route",
			statusCode: http.StatusOK,
			host:       "api.example.com",
			httpRoute:  "/v1/users",
			localAddr:  &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8443},
			wantAttributes: attribute.NewSet(
				attribute.Key("http.request.method").String("GET"),
				attribute.Key("http.response.status_code").Int(200),
				attribute.Key("http.route").String("/v1/users"),
				attribute.Key("network.protocol.name").String("http/1.1"),
				attribute.Key("network.protocol.version").String("1.1"),
				attribute.Key("server.address").String("api.example.com"),
				attribute.Key("server.port").Int(8443),
				attribute.Key("url.scheme").String("https"),
			),
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
			// force the meter provider with manual reader to collect metrics for the test.
			metrics.SetMeterProvider(meterProvider)

			semConvMetricRegistry, err := metrics.NewSemConvMetricRegistry(t.Context(), &cfg)
			require.NoError(t, err)
			require.NotNil(t, semConvMetricRegistry)

			req := httptest.NewRequest(http.MethodGet, "http://"+test.host+"/search?q=Opentelemetry", nil)
			rw := httptest.NewRecorder()
			req.Host = test.host
			req.RemoteAddr = "10.0.0.1:1234"
			req.Header.Set("User-Agent", "entrypoint-test")

			// Set X-Forwarded-Proto based on the local address port or default to http.
			scheme := "http"
			if test.localAddr != nil {
				ctx := context.WithValue(req.Context(), http.LocalAddrContextKey, test.localAddr)
				req = req.WithContext(ctx)

				// Determine scheme based on port for the test.
				if tcpAddr, ok := test.localAddr.(*net.TCPAddr); ok {
					if tcpAddr.Port == 443 || tcpAddr.Port == 8443 {
						scheme = "https"
					}
				}
			}
			req.Header.Set("X-Forwarded-Proto", scheme)

			// Inject http.route into context if provided.
			if test.httpRoute != "" {
				req = req.WithContext(WithHTTPRoute(req.Context(), test.httpRoute))
			}

			next := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
				rw.WriteHeader(test.statusCode)
			})

			handler := newServerMetricsSemConv(t.Context(), semConvMetricRegistry, next)

			handler, err = capture.Wrap(handler)
			require.NoError(t, err)

			// Injection of the observability variables in the request context.
			handler = WithObservabilityHandler(handler, Observability{
				SemConvMetricsEnabled: true,
			})

			handler.ServeHTTP(rw, req)

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
							Attributes:   test.wantAttributes,
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
