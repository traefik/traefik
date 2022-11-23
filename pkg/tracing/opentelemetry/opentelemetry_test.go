package opentelemetry

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mtracing "github.com/traefik/traefik/v2/pkg/middlewares/tracing"
	"github.com/traefik/traefik/v2/pkg/tracing"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
)

func TestTraceContextPropagation(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gzr, err := gzip.NewReader(r.Body)
		require.NoError(t, err)

		body, err := io.ReadAll(gzr)
		require.NoError(t, err)

		req := ptraceotlp.NewExportRequest()
		err = req.UnmarshalProto(body)
		require.NoError(t, err)

		marshalledReq, err := json.Marshal(req)
		require.NoError(t, err)

		bodyStr := string(marshalledReq)
		assert.Regexp(t, `("traceId":"00000000000000000000000000000001")`, bodyStr)
		assert.Regexp(t, `("parentSpanId":"0000000000000001")`, bodyStr)
		assert.Regexp(t, `("traceState":"foo=bar")`, bodyStr)
	}))
	defer ts.Close()

	cfg := Config{
		Address:  strings.TrimPrefix(ts.URL, "http://"),
		Insecure: true,
	}

	newTracing, err := tracing.NewTracing("", 0, &cfg)
	require.NoError(t, err)
	defer newTracing.Close()

	req := httptest.NewRequest(http.MethodGet, "http://www.test.com", nil)
	req.Header.Set("traceparent", "00-00000000000000000000000000000001-0000000000000001-00")
	req.Header.Set("tracestate", "foo=bar")
	rw := httptest.NewRecorder()

	var forwarded bool
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		forwarded = true
	})

	handler := mtracing.NewEntryPoint(context.Background(), newTracing, "test", next)
	handler.ServeHTTP(rw, req)

	require.True(t, forwarded)
}
