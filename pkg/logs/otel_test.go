package logs

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/types"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/otel/trace"
)

func TestLog(t *testing.T) {
	tests := []struct {
		desc     string
		level    zerolog.Level
		assertFn func(*testing.T, string)
		noLog    bool
	}{
		{
			desc:  "no level log",
			level: zerolog.NoLevel,
			assertFn: func(t *testing.T, log string) {
				t.Helper()
				// SeverityUndefined Severity = 0 // UNDEFINED
				assert.NotContains(t, log, `"severityNumber"`)
				assert.Regexp(t, `{"key":"resource","value":{"stringValue":"attribute"}}`, log)
				assert.Regexp(t, `{"key":"service.name","value":{"stringValue":"test"}}`, log)
				assert.Regexp(t, `"body":{"stringValue":"test"}`, log)
				assert.Regexp(t, `{"key":"foo","value":{"stringValue":"bar"}}`, log)
				assert.Regexp(t, `"traceId":"01020304050607080000000000000000","spanId":"0102030405060708"`, log)
			},
		},
		{
			desc:  "trace log",
			level: zerolog.TraceLevel,
			assertFn: func(t *testing.T, log string) {
				t.Helper()
				// SeverityTrace1 Severity = 1 // TRACE
				assert.Contains(t, log, `"severityNumber":1`)
				assert.Regexp(t, `{"key":"resource","value":{"stringValue":"attribute"}}`, log)
				assert.Regexp(t, `{"key":"service.name","value":{"stringValue":"test"}}`, log)
				assert.Regexp(t, `"body":{"stringValue":"test"}`, log)
				assert.Regexp(t, `{"key":"foo","value":{"stringValue":"bar"}}`, log)
				assert.Regexp(t, `"traceId":"01020304050607080000000000000000","spanId":"0102030405060708"`, log)
			},
		},
		{
			desc:  "debug log",
			level: zerolog.DebugLevel,
			assertFn: func(t *testing.T, log string) {
				t.Helper()
				// SeverityDebug1 Severity = 5 // DEBUG
				assert.Contains(t, log, `"severityNumber":5`)
				assert.Regexp(t, `{"key":"resource","value":{"stringValue":"attribute"}}`, log)
				assert.Regexp(t, `{"key":"service.name","value":{"stringValue":"test"}}`, log)
				assert.Regexp(t, `"body":{"stringValue":"test"}`, log)
				assert.Regexp(t, `{"key":"foo","value":{"stringValue":"bar"}}`, log)
				assert.Regexp(t, `"traceId":"01020304050607080000000000000000","spanId":"0102030405060708"`, log)
			},
		},
		{
			desc:  "info log",
			level: zerolog.InfoLevel,
			assertFn: func(t *testing.T, log string) {
				t.Helper()
				// SeverityInfo1 Severity = 9  // INFO
				assert.Contains(t, log, `"severityNumber":9`)
				assert.Regexp(t, `{"key":"resource","value":{"stringValue":"attribute"}}`, log)
				assert.Regexp(t, `{"key":"service.name","value":{"stringValue":"test"}}`, log)
				assert.Regexp(t, `"body":{"stringValue":"test"}`, log)
				assert.Regexp(t, `{"key":"foo","value":{"stringValue":"bar"}}`, log)
				assert.Regexp(t, `"traceId":"01020304050607080000000000000000","spanId":"0102030405060708"`, log)
			},
		},
		{
			desc:  "warn log",
			level: zerolog.WarnLevel,
			assertFn: func(t *testing.T, log string) {
				t.Helper()
				// SeverityWarn1 Severity = 13 // WARN
				assert.Contains(t, log, `"severityNumber":13`)
				assert.Regexp(t, `{"key":"resource","value":{"stringValue":"attribute"}}`, log)
				assert.Regexp(t, `{"key":"service.name","value":{"stringValue":"test"}}`, log)
				assert.Regexp(t, `"body":{"stringValue":"test"}`, log)
				assert.Regexp(t, `{"key":"foo","value":{"stringValue":"bar"}}`, log)
				assert.Regexp(t, `"traceId":"01020304050607080000000000000000","spanId":"0102030405060708"`, log)
			},
		},
		{
			desc:  "error log",
			level: zerolog.ErrorLevel,
			assertFn: func(t *testing.T, log string) {
				t.Helper()
				// SeverityError1 Severity = 17 // ERROR
				assert.Contains(t, log, `"severityNumber":17`)
				assert.Regexp(t, `{"key":"resource","value":{"stringValue":"attribute"}}`, log)
				assert.Regexp(t, `{"key":"service.name","value":{"stringValue":"test"}}`, log)
				assert.Regexp(t, `"body":{"stringValue":"test"}`, log)
				assert.Regexp(t, `{"key":"foo","value":{"stringValue":"bar"}}`, log)
				assert.Regexp(t, `"traceId":"01020304050607080000000000000000","spanId":"0102030405060708"`, log)
			},
		},
		{
			desc:  "fatal log",
			level: zerolog.FatalLevel,
			assertFn: func(t *testing.T, log string) {
				t.Helper()
				// SeverityFatal Severity = 21 // FATAL
				assert.Contains(t, log, `"severityNumber":21`)
				assert.Regexp(t, `{"key":"resource","value":{"stringValue":"attribute"}}`, log)
				assert.Regexp(t, `{"key":"service.name","value":{"stringValue":"test"}}`, log)
				assert.Regexp(t, `"body":{"stringValue":"test"}`, log)
				assert.Regexp(t, `{"key":"foo","value":{"stringValue":"bar"}}`, log)
				assert.Regexp(t, `"traceId":"01020304050607080000000000000000","spanId":"0102030405060708"`, log)
			},
		},
		{
			desc:  "panic log",
			level: zerolog.PanicLevel,
			assertFn: func(t *testing.T, log string) {
				t.Helper()
				// SeverityFatal4 Severity = 24 // FATAL
				assert.Contains(t, log, `"severityNumber":24`)
				assert.Regexp(t, `{"key":"resource","value":{"stringValue":"attribute"}}`, log)
				assert.Regexp(t, `{"key":"service.name","value":{"stringValue":"test"}}`, log)
				assert.Regexp(t, `"body":{"stringValue":"test"}`, log)
				assert.Regexp(t, `{"key":"foo","value":{"stringValue":"bar"}}`, log)
				assert.Regexp(t, `"traceId":"01020304050607080000000000000000","spanId":"0102030405060708"`, log)
			},
		},
	}

	logCh := make(chan string)
	collector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gzr, err := gzip.NewReader(r.Body)
		require.NoError(t, err)

		body, err := io.ReadAll(gzr)
		require.NoError(t, err)

		req := plogotlp.NewExportRequest()
		err = req.UnmarshalProto(body)
		require.NoError(t, err)

		marshalledReq, err := json.Marshal(req)
		require.NoError(t, err)

		logCh <- string(marshalledReq)
	}))
	t.Cleanup(collector.Close)

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			config := &types.OTelLog{
				ServiceName:        "test",
				ResourceAttributes: map[string]string{"resource": "attribute"},
				HTTP: &types.OTelHTTP{
					Endpoint: collector.URL,
				},
			}

			out := zerolog.MultiLevelWriter(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
			logger := zerolog.New(out).With().Caller().Logger()

			logger, err := SetupOTelLogger(logger, config)
			require.NoError(t, err)

			ctx := trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{
				TraceID: trace.TraceID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				SpanID:  trace.SpanID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
			}))
			logger = logger.With().Ctx(ctx).Logger()

			logger.WithLevel(test.level).Str("foo", "bar").Msg("test")

			select {
			case <-time.After(5 * time.Second):
				t.Error("Log not exported")

			case log := <-logCh:
				if test.assertFn != nil {
					test.assertFn(t, log)
				}
			}
		})
	}
}
