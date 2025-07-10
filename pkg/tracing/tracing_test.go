package tracing

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/containous/alice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/static"
	"github.com/traefik/traefik/v3/pkg/types"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"go.opentelemetry.io/otel/trace"
)

func Test_safeFullURL(t *testing.T) {
	testCases := []struct {
		desc            string
		safeQueryParams []string
		originalURL     *url.URL
		expectedURL     *url.URL
	}{
		{
			desc:        "Nil URL",
			originalURL: nil,
			expectedURL: nil,
		},
		{
			desc:        "No query parameters",
			originalURL: &url.URL{Scheme: "https", Host: "example.com"},
			expectedURL: &url.URL{Scheme: "https", Host: "example.com"},
		},
		{
			desc:        "All query parameters redacted",
			originalURL: &url.URL{Scheme: "https", Host: "example.com", RawQuery: "foo=bar&baz=qux"},
			expectedURL: &url.URL{Scheme: "https", Host: "example.com", RawQuery: "baz=REDACTED&foo=REDACTED"},
		},
		{
			desc:            "Some query parameters unredacted",
			safeQueryParams: []string{"foo"},
			originalURL:     &url.URL{Scheme: "https", Host: "example.com", RawQuery: "foo=bar&baz=qux"},
			expectedURL:     &url.URL{Scheme: "https", Host: "example.com", RawQuery: "baz=REDACTED&foo=bar"},
		},
		{
			desc:            "User info and some query parameters redacted",
			safeQueryParams: []string{"foo"},
			originalURL:     &url.URL{Scheme: "https", Host: "example.com", User: url.UserPassword("username", "password"), RawQuery: "foo=bar&baz=qux"},
			expectedURL:     &url.URL{Scheme: "https", Host: "example.com", User: url.UserPassword("REDACTED", "REDACTED"), RawQuery: "baz=REDACTED&foo=bar"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			tr := NewTracer(nil, nil, nil, test.safeQueryParams)

			gotURL := tr.safeURL(test.originalURL)

			assert.Equal(t, test.expectedURL, gotURL)
		})
	}
}

func TestTracing(t *testing.T) {
	tests := []struct {
		desc                 string
		propagators          string
		headers              map[string]string
		resourceAttributes   map[string]string
		wantServiceHeadersFn func(t *testing.T, headers http.Header)
		assertFn             func(*testing.T, string)
	}{
		{
			desc: "service name and version",
			assertFn: func(t *testing.T, trace string) {
				t.Helper()

				assert.Regexp(t, `({"key":"service.name","value":{"stringValue":"traefik"}})`, trace)
				assert.Regexp(t, `({"key":"service.version","value":{"stringValue":"dev"}})`, trace)
			},
		},
		{
			desc: "resource attributes must be propagated",
			resourceAttributes: map[string]string{
				"service.environment": "custom",
			},
			assertFn: func(t *testing.T, trace string) {
				t.Helper()

				assert.Regexp(t, `({"key":"service.environment","value":{"stringValue":"custom"}})`, trace)
			},
		},
		{
			desc:        "TraceContext propagation",
			propagators: "tracecontext",
			headers: map[string]string{
				"traceparent": "00-00000000000000000000000000000001-0000000000000001-01",
				"tracestate":  "foo=bar",
			},
			wantServiceHeadersFn: func(t *testing.T, headers http.Header) {
				t.Helper()

				assert.Regexp(t, `(00-00000000000000000000000000000001-\w{16}-01)`, headers["Traceparent"][0])
				assert.Equal(t, []string{"foo=bar"}, headers["Tracestate"])
			},
			assertFn: func(t *testing.T, trace string) {
				t.Helper()

				assert.Regexp(t, `("traceId":"00000000000000000000000000000001")`, trace)
				assert.Regexp(t, `("parentSpanId":"0000000000000001")`, trace)
				assert.Regexp(t, `("traceState":"foo=bar")`, trace)
			},
		},
		{
			desc:        "root span TraceContext propagation",
			propagators: "tracecontext",
			wantServiceHeadersFn: func(t *testing.T, headers http.Header) {
				t.Helper()

				assert.Regexp(t, `(00-\w{32}-\w{16}-01)`, headers["Traceparent"][0])
			},
			assertFn: func(t *testing.T, trace string) {
				t.Helper()

				assert.Regexp(t, `("traceId":"\w{32}")`, trace)
				assert.Regexp(t, `("parentSpanId":"\w{16}")`, trace)
			},
		},
		{
			desc:        "B3 propagation",
			propagators: "b3",
			headers: map[string]string{
				"b3": "00000000000000000000000000000001-0000000000000002-1-0000000000000001",
			},
			wantServiceHeadersFn: func(t *testing.T, headers http.Header) {
				t.Helper()

				assert.Regexp(t, `(00000000000000000000000000000001-\w{16}-1)`, headers["B3"][0])
			},
			assertFn: func(t *testing.T, trace string) {
				t.Helper()

				assert.Regexp(t, `("traceId":"00000000000000000000000000000001")`, trace)
				assert.Regexp(t, `("parentSpanId":"0000000000000002")`, trace)
			},
		},
		{
			desc:        "root span B3 propagation",
			propagators: "b3",
			wantServiceHeadersFn: func(t *testing.T, headers http.Header) {
				t.Helper()

				assert.Regexp(t, `(\w{32}-\w{16}-1)`, headers["B3"][0])
			},
			assertFn: func(t *testing.T, trace string) {
				t.Helper()

				assert.Regexp(t, `("traceId":"\w{32}")`, trace)
				assert.Regexp(t, `("parentSpanId":"\w{16}")`, trace)
			},
		},
		{
			desc:        "B3 propagation Multiple Headers",
			propagators: "b3multi",
			headers: map[string]string{
				"x-b3-traceid":      "00000000000000000000000000000001",
				"x-b3-parentspanid": "0000000000000001",
				"x-b3-spanid":       "0000000000000002",
				"x-b3-sampled":      "1",
			},
			wantServiceHeadersFn: func(t *testing.T, headers http.Header) {
				t.Helper()

				assert.Equal(t, "00000000000000000000000000000001", headers["X-B3-Traceid"][0])
				assert.Equal(t, "0000000000000001", headers["X-B3-Parentspanid"][0])
				assert.Equal(t, "1", headers["X-B3-Sampled"][0])
				assert.Len(t, headers["X-B3-Spanid"][0], 16)
			},
			assertFn: func(t *testing.T, trace string) {
				t.Helper()

				assert.Regexp(t, `("traceId":"00000000000000000000000000000001")`, trace)
				assert.Regexp(t, `("parentSpanId":"0000000000000002")`, trace)
			},
		},
		{
			desc:        "root span B3 propagation Multiple Headers",
			propagators: "b3multi",
			wantServiceHeadersFn: func(t *testing.T, headers http.Header) {
				t.Helper()

				assert.Regexp(t, `(\w{32})`, headers["X-B3-Traceid"][0])
				assert.Equal(t, "1", headers["X-B3-Sampled"][0])
				assert.Regexp(t, `(\w{16})`, headers["X-B3-Spanid"][0])
			},
			assertFn: func(t *testing.T, trace string) {
				t.Helper()

				assert.Regexp(t, `("traceId":"\w{32}")`, trace)
				assert.Regexp(t, `("parentSpanId":"")`, trace)
			},
		},
		{
			desc:        "Baggage propagation",
			propagators: "baggage",
			headers: map[string]string{
				"baggage": "userId=id",
			},
			wantServiceHeadersFn: func(t *testing.T, headers http.Header) {
				t.Helper()

				assert.Equal(t, []string{"userId=id"}, headers["Baggage"])
			},
		},
		{
			desc:        "Jaeger propagation",
			propagators: "jaeger",
			headers: map[string]string{
				"uber-trace-id": "00000000000000000000000000000001:0000000000000002:0000000000000001:1",
			},
			wantServiceHeadersFn: func(t *testing.T, headers http.Header) {
				t.Helper()

				assert.Regexp(t, `(00000000000000000000000000000001:\w{16}:0:1)`, headers["Uber-Trace-Id"][0])
			},
			assertFn: func(t *testing.T, trace string) {
				t.Helper()

				assert.Regexp(t, `("traceId":"00000000000000000000000000000001")`, trace)
				assert.Regexp(t, `("parentSpanId":"\w{16}")`, trace)
			},
		},
		{
			desc:        "root span Jaeger propagation",
			propagators: "jaeger",
			wantServiceHeadersFn: func(t *testing.T, headers http.Header) {
				t.Helper()

				assert.Regexp(t, `(\w{32}:\w{16}:0:1)`, headers["Uber-Trace-Id"][0])
			},
			assertFn: func(t *testing.T, trace string) {
				t.Helper()

				assert.Regexp(t, `("traceId":"\w{32}")`, trace)
				assert.Regexp(t, `("parentSpanId":"\w{16}")`, trace)
			},
		},
		{
			desc:        "XRay propagation",
			propagators: "xray",
			headers: map[string]string{
				"X-Amzn-Trace-Id": "Root=1-5759e988-bd862e3fe1be46a994272793;Parent=53995c3f42cd8ad8;Sampled=1",
			},
			wantServiceHeadersFn: func(t *testing.T, headers http.Header) {
				t.Helper()

				assert.Regexp(t, `(Root=1-5759e988-bd862e3fe1be46a994272793;Parent=\w{16};Sampled=1)`, headers["X-Amzn-Trace-Id"][0])
			},
			assertFn: func(t *testing.T, trace string) {
				t.Helper()

				assert.Regexp(t, `("traceId":"5759e988bd862e3fe1be46a994272793")`, trace)
				assert.Regexp(t, `("parentSpanId":"\w{16}")`, trace)
			},
		},
		{
			desc:        "root span XRay propagation",
			propagators: "xray",
			wantServiceHeadersFn: func(t *testing.T, headers http.Header) {
				t.Helper()

				assert.Regexp(t, `(Root=1-\w{8}-\w{24};Parent=\w{16};Sampled=1)`, headers["X-Amzn-Trace-Id"][0])
			},
			assertFn: func(t *testing.T, trace string) {
				t.Helper()

				assert.Regexp(t, `("traceId":"\w{32}")`, trace)
				assert.Regexp(t, `("parentSpanId":"\w{16}")`, trace)
			},
		},
		{
			desc:        "no propagation",
			propagators: "none",
			wantServiceHeadersFn: func(t *testing.T, headers http.Header) {
				t.Helper()

				assert.Empty(t, headers)
			},
			assertFn: func(t *testing.T, trace string) {
				t.Helper()

				assert.Regexp(t, `("traceId":"\w{32}")`, trace)
				assert.Regexp(t, `("parentSpanId":"\w{16}")`, trace)
			},
		},
	}

	traceCh := make(chan string)
	collector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gzr, err := gzip.NewReader(r.Body)
		require.NoError(t, err)

		body, err := io.ReadAll(gzr)
		require.NoError(t, err)

		req := ptraceotlp.NewExportRequest()
		err = req.UnmarshalProto(body)
		require.NoError(t, err)

		marshalledReq, err := json.Marshal(req)
		require.NoError(t, err)

		traceCh <- string(marshalledReq)
	}))
	t.Cleanup(collector.Close)

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			t.Setenv("OTEL_PROPAGATORS", test.propagators)

			service := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tracer := TracerFromContext(r.Context())
				ctx, span := tracer.Start(r.Context(), "service")
				defer span.End()

				r = r.WithContext(ctx)

				InjectContextIntoCarrier(r)

				if test.wantServiceHeadersFn != nil {
					test.wantServiceHeadersFn(t, r.Header)
				}
			})

			tracingConfig := &static.Tracing{
				ServiceName:        "traefik",
				SampleRate:         1.0,
				ResourceAttributes: test.resourceAttributes,
				OTLP: &types.OTelTracing{
					HTTP: &types.OTelHTTP{
						Endpoint: collector.URL,
					},
				},
			}

			tracer, closer, err := NewTracing(t.Context(), tracingConfig)
			require.NoError(t, err)
			t.Cleanup(func() {
				_ = closer.Close()
			})

			chain := alice.New(func(next http.Handler) (http.Handler, error) {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					tracingCtx := ExtractCarrierIntoContext(r.Context(), r.Header)
					start := time.Now()
					tracingCtx, span := tracer.Start(tracingCtx, "test", trace.WithSpanKind(trace.SpanKindServer), trace.WithTimestamp(start))
					end := time.Now()
					span.End(trace.WithTimestamp(end))
					next.ServeHTTP(w, r.WithContext(tracingCtx))
				}), nil
			})

			epHandler, err := chain.Then(service)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, "http://www.test.com", nil)
			for k, v := range test.headers {
				req.Header.Set(k, v)
			}

			rw := httptest.NewRecorder()

			epHandler.ServeHTTP(rw, req)

			select {
			case <-time.After(10 * time.Second):
				t.Error("Trace not exported")

			case trace := <-traceCh:
				assert.Equal(t, http.StatusOK, rw.Code)
				if test.assertFn != nil {
					test.assertFn(t, trace)
				}
			}
		})
	}
}

// TestTracerProvider ensures that Tracer returns a valid TracerProvider
// when using the default Traefik Tracer and a custom one.
func TestTracerProvider(t *testing.T) {
	t.Parallel()

	otlpConfig := &types.OTelTracing{}
	otlpConfig.SetDefaults()

	config := &static.Tracing{OTLP: otlpConfig}
	tracer, closer, err := NewTracing(t.Context(), config)
	if err != nil {
		t.Fatal(err)
	}
	closer.Close()

	_, span := tracer.Start(t.Context(), "test")
	defer span.End()

	span.TracerProvider().Tracer("github.com/traefik/traefik")
	span.TracerProvider().Tracer("other")
}
