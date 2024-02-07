package observability

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/metrics"
	"github.com/traefik/traefik/v3/pkg/types"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
)

func TestEntryPointMiddleware(t *testing.T) {
	type expected struct {
		name       string
		attributes []attribute.KeyValue
	}

	testCases := []struct {
		desc       string
		entryPoint string
		expected   expected
	}{
		{
			desc:       "basic test",
			entryPoint: "test",
			expected: expected{
				name: "EntryPoint",
				attributes: []attribute.KeyValue{
					attribute.String("span.kind", "server"),
					attribute.String("entry_point", "test"),
					attribute.String("http.request.method", "GET"),
					attribute.String("network.protocol.version", "1.1"),
					attribute.Int64("http.request.body.size", int64(0)),
					attribute.String("url.path", "/search"),
					attribute.String("url.query", "q=Opentelemetry"),
					attribute.String("url.scheme", "http"),
					attribute.String("user_agent.original", "entrypoint-test"),
					attribute.String("server.address", "www.test.com"),
					attribute.String("network.peer.address", "10.0.0.1"),
					attribute.String("network.peer.port", "1234"),
					attribute.String("client.address", "10.0.0.1"),
					attribute.Int64("client.port", int64(1234)),
					attribute.String("client.socket.address", ""),
					attribute.Int64("http.response.status_code", int64(404)),
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://www.test.com/search?q=Opentelemetry", nil)
			rw := httptest.NewRecorder()
			req.RemoteAddr = "10.0.0.1:1234"
			req.Header.Set("User-Agent", "entrypoint-test")
			req.Header.Set("X-Forwarded-Proto", "http")

			next := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
				rw.WriteHeader(http.StatusNotFound)
			})

			tracer := &mockTracer{}

			handler := newEntryPoint(context.Background(), tracer, nil, test.entryPoint, next)
			handler.ServeHTTP(rw, req)

			for _, span := range tracer.spans {
				assert.Equal(t, test.expected.name, span.name)
				assert.Equal(t, test.expected.attributes, span.attributes)
			}
		})
	}
}

func TestName(t *testing.T) {
	c := make(chan *string, 5)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gzr, err := gzip.NewReader(r.Body)
		require.NoError(t, err)

		body, err := io.ReadAll(gzr)
		require.NoError(t, err)

		req := pmetricotlp.NewExportRequest()
		err = req.UnmarshalProto(body)
		require.NoError(t, err)

		marshalledReq, err := json.Marshal(req)
		require.NoError(t, err)

		bodyStr := string(marshalledReq)
		c <- &bodyStr

		w.WriteHeader(http.StatusOK)
	}))

	t.Cleanup(func() {
		ts.Close()
	})

	var cfg types.OTLP
	(&cfg).SetDefaults()
	cfg.AddRoutersLabels = true
	cfg.HTTP = &types.OtelHTTP{
		Endpoint: ts.URL,
	}
	cfg.PushInterval = ptypes.Duration(10 * time.Millisecond)

	semConvMetricRegistry, err := metrics.SemConvMetricRegistry(context.Background(), &cfg)
	require.NoError(t, err)
	require.NotNil(t, semConvMetricRegistry)

	req := httptest.NewRequest(http.MethodGet, "http://www.test.com/search?q=Opentelemetry", nil)
	rw := httptest.NewRecorder()
	req.RemoteAddr = "10.0.0.1:1234"
	req.Header.Set("User-Agent", "entrypoint-test")
	req.Header.Set("X-Forwarded-Proto", "http")

	next := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusNotFound)
	})

	handler := newEntryPoint(context.Background(), nil, semConvMetricRegistry, "test", next)
	handler.ServeHTTP(rw, req)

	tryAssertMessage(t, c, []string{"powpow"})
}

func tryAssertMessage(t *testing.T, c chan *string, expected []string) {
	t.Helper()

	var errs []error
	timeout := time.After(1 * time.Second)
	for {
		select {
		case <-timeout:
			for _, err := range errs {
				t.Error(err)
			}
		case msg := <-c:
			errs = verifyMessage(*msg, expected)
			if len(errs) == 0 {
				return
			}
		}
	}
}

func verifyMessage(msg string, expected []string) []error {
	var errs []error
	for _, pattern := range expected {
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(msg)
		if len(match) != 2 {
			errs = append(errs, fmt.Errorf("Got %q %v, want %q", msg, match, pattern))
		}
	}
	return errs
}
