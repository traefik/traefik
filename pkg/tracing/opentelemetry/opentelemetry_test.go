package opentelemetry_test

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/containous/alice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/static"
	tracingMiddle "github.com/traefik/traefik/v3/pkg/middlewares/tracing"
	"github.com/traefik/traefik/v3/pkg/tracing"
	"github.com/traefik/traefik/v3/pkg/tracing/opentelemetry"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
)

func TestTracing(t *testing.T) {
	tests := []struct {
		desc     string
		headers  map[string]string
		assertFn func(*testing.T, string)
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
			desc: "context propagation",
			headers: map[string]string{
				"traceparent": "00-00000000000000000000000000000001-0000000000000001-01",
				"tracestate":  "foo=bar",
			},
			assertFn: func(t *testing.T, trace string) {
				t.Helper()

				assert.Regexp(t, `("traceId":"00000000000000000000000000000001")`, trace)
				assert.Regexp(t, `("parentSpanId":"0000000000000001")`, trace)
				assert.Regexp(t, `("traceState":"foo=bar")`, trace)
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

	tracingConfig := &static.Tracing{
		ServiceName: "traefik",
		SampleRate:  1.0,
		OTLP: &opentelemetry.Config{
			HTTP: &opentelemetry.HTTP{
				Endpoint: collector.URL,
			},
		},
	}

	newTracing, closer, err := tracing.NewTracing(tracingConfig)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = closer.Close()
	})

	chain := alice.New(tracingMiddle.WrapEntryPointHandler(context.Background(), newTracing, "test"))
	epHandler, err := chain.Then(http.NotFoundHandler())
	require.NoError(t, err)

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
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
				assert.Equal(t, http.StatusNotFound, rw.Code)
				test.assertFn(t, trace)
			}
		})
	}
}
