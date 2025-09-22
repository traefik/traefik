package observability

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/metrics"
	"github.com/traefik/traefik/v3/pkg/middlewares/capture"
	"github.com/traefik/traefik/v3/pkg/types"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/metric/metricdata/metricdatatest"
)

func TestSemConvServerMetrics(t *testing.T) {
	tests := []struct {
		desc           string
		statusCode     int
		wantAttributes attribute.Set
	}{
		{
			desc:       "not found status",
			statusCode: http.StatusNotFound,
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
			wantAttributes: attribute.NewSet(
				attribute.Key("http.request.method").String("GET"),
				attribute.Key("http.response.status_code").Int(201),
				attribute.Key("network.protocol.name").String("http/1.1"),
				attribute.Key("network.protocol.version").String("1.1"),
				attribute.Key("server.address").String("www.test.com"),
				attribute.Key("url.scheme").String("http"),
			),
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var cfg types.OTLP
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

			req := httptest.NewRequest(http.MethodGet, "http://www.test.com/search?q=Opentelemetry", nil)
			rw := httptest.NewRecorder()
			req.RemoteAddr = "10.0.0.1:1234"
			req.Header.Set("User-Agent", "entrypoint-test")
			req.Header.Set("X-Forwarded-Proto", "http")

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
