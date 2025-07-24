package accesslog

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/containous/alice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/middlewares/capture"
	"github.com/traefik/traefik/v3/pkg/middlewares/observability"
	"github.com/traefik/traefik/v3/pkg/types"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/embedded"
)

const delta float64 = 1e-10

var (
	logFileNameSuffix       = "/traefik/logger/test.log"
	testContent             = "Hello, World"
	testServiceName         = "http://127.0.0.1/testService"
	testRouterName          = "testRouter"
	testStatus              = 123
	testContentSize   int64 = 12
	testHostname            = "TestHost"
	testUsername            = "TestUser"
	testPath                = "testpath"
	testPort                = 8181
	testProto               = "HTTP/0.0"
	testScheme              = "http"
	testMethod              = http.MethodPost
	testReferer             = "testReferer"
	testUserAgent           = "testUserAgent"
	testRetryAttempts       = 2
	testStart               = time.Now()
)

func TestOTelAccessLogWithBody(t *testing.T) {
	testCases := []struct {
		desc        string
		format      string
		bodyCheckFn func(*testing.T, string)
	}{
		{
			desc:   "Common format with log body",
			format: CommonFormat,
			bodyCheckFn: func(t *testing.T, log string) {
				t.Helper()

				// For common format, verify the body contains the CLF formatted string
				assert.Regexp(t, `"body":{"stringValue":".*- /health -.*200.*"}`, log)
			},
		},
		{
			desc:   "JSON format with log body",
			format: JSONFormat,
			bodyCheckFn: func(t *testing.T, log string) {
				t.Helper()

				// For JSON format, verify the body contains the JSON formatted string
				assert.Regexp(t, `"body":{"stringValue":".*DownstreamStatus.*:200.*"}`, log)
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

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

			config := &types.AccessLog{
				Format: test.format,
				OTLP: &types.OTelLog{
					ServiceName:        "test",
					ResourceAttributes: map[string]string{"resource": "attribute"},
					HTTP: &types.OTelHTTP{
						Endpoint: collector.URL,
					},
				},
			}
			logHandler, err := NewHandler(t.Context(), config)
			require.NoError(t, err)
			t.Cleanup(func() {
				err := logHandler.Close()
				require.NoError(t, err)
			})

			req := &http.Request{
				Header: map[string][]string{},
				URL: &url.URL{
					Path: "/health",
				},
			}
			ctx := trace.ContextWithSpanContext(t.Context(), trace.NewSpanContext(trace.SpanContextConfig{
				TraceID: trace.TraceID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
				SpanID:  trace.SpanID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
			}))
			req = req.WithContext(ctx)

			chain := alice.New()
			chain = chain.Append(capture.Wrap)

			// Injection of the observability variables in the request context.
			chain = chain.Append(func(next http.Handler) (http.Handler, error) {
				return observability.WithObservabilityHandler(next, observability.Observability{
					AccessLogsEnabled: true,
				}), nil
			})

			chain = chain.Append(logHandler.AliceConstructor())
			handler, err := chain.Then(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
			}))
			require.NoError(t, err)
			handler.ServeHTTP(httptest.NewRecorder(), req)

			select {
			case <-time.After(5 * time.Second):
				t.Error("AccessLog not exported")

			case log := <-logCh:
				// Verify basic OTLP structure
				assert.Regexp(t, `{"key":"resource","value":{"stringValue":"attribute"}}`, log)
				assert.Regexp(t, `{"key":"service.name","value":{"stringValue":"test"}}`, log)
				assert.Regexp(t, `{"key":"DownstreamStatus","value":{"intValue":"200"}}`, log)
				assert.Regexp(t, `"traceId":"01020304050607080000000000000000","spanId":"0102030405060708"`, log)

				// Most importantly, verify the log body is populated (not empty)
				assert.NotRegexp(t, `"body":{"stringValue":""}`, log, "Log body should not be empty when OTLP is configured")

				// Run format-specific body checks
				test.bodyCheckFn(t, log)
			}
		})
	}
}

func TestLogRotation(t *testing.T) {
	fileName := filepath.Join(t.TempDir(), "traefik.log")
	rotatedFileName := fileName + ".rotated"

	config := &types.AccessLog{FilePath: fileName, Format: CommonFormat}
	logHandler, err := NewHandler(t.Context(), config)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := logHandler.Close()
		require.NoError(t, err)
	})

	chain := alice.New()
	chain = chain.Append(capture.Wrap)

	// Injection of the observability variables in the request context.
	chain = chain.Append(func(next http.Handler) (http.Handler, error) {
		return observability.WithObservabilityHandler(next, observability.Observability{
			AccessLogsEnabled: true,
		}), nil
	})

	chain = chain.Append(logHandler.AliceConstructor())
	handler, err := chain.Then(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)

	iterations := 20
	halfDone := make(chan bool)
	writeDone := make(chan bool)
	go func() {
		for i := range iterations {
			handler.ServeHTTP(recorder, req)
			if i == iterations/2 {
				halfDone <- true
			}
		}
		writeDone <- true
	}()

	<-halfDone
	err = os.Rename(fileName, rotatedFileName)
	if err != nil {
		t.Fatalf("Error renaming file: %s", err)
	}

	err = logHandler.Rotate()
	if err != nil {
		t.Fatalf("Error rotating file: %s", err)
	}

	select {
	case <-writeDone:
		gotLineCount := lineCount(t, fileName) + lineCount(t, rotatedFileName)
		if iterations != gotLineCount {
			t.Errorf("Wanted %d written log lines, got %d", iterations, gotLineCount)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("test timed out")
	}

	close(halfDone)
	close(writeDone)
}

func lineCount(t *testing.T, fileName string) int {
	t.Helper()
	fileContents, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatalf("Error reading from file %s: %s", fileName, err)
	}

	count := 0
	for _, line := range strings.Split(string(fileContents), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		count++
	}

	return count
}

func TestLoggerHeaderFields(t *testing.T) {
	expectedValues := []string{"AAA", "BBB"}

	testCases := []struct {
		desc            string
		accessLogFields types.AccessLogFields
		header          string
		expected        string
	}{
		{
			desc:     "with default mode",
			header:   "User-Agent",
			expected: types.AccessLogDrop,
			accessLogFields: types.AccessLogFields{
				DefaultMode: types.AccessLogDrop,
				Headers: &types.FieldHeaders{
					DefaultMode: types.AccessLogDrop,
					Names:       map[string]string{},
				},
			},
		},
		{
			desc:     "with exact header name",
			header:   "User-Agent",
			expected: types.AccessLogKeep,
			accessLogFields: types.AccessLogFields{
				DefaultMode: types.AccessLogDrop,
				Headers: &types.FieldHeaders{
					DefaultMode: types.AccessLogDrop,
					Names: map[string]string{
						"User-Agent": types.AccessLogKeep,
					},
				},
			},
		},
		{
			desc:     "with case-insensitive match on header name",
			header:   "User-Agent",
			expected: types.AccessLogKeep,
			accessLogFields: types.AccessLogFields{
				DefaultMode: types.AccessLogDrop,
				Headers: &types.FieldHeaders{
					DefaultMode: types.AccessLogDrop,
					Names: map[string]string{
						"user-agent": types.AccessLogKeep,
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			logFile, err := os.CreateTemp(t.TempDir(), "*.log")
			require.NoError(t, err)

			config := &types.AccessLog{
				FilePath: logFile.Name(),
				Format:   CommonFormat,
				Fields:   &test.accessLogFields,
			}

			logger, err := NewHandler(t.Context(), config)
			require.NoError(t, err)
			t.Cleanup(func() {
				err := logger.Close()
				require.NoError(t, err)
			})

			if config.FilePath != "" {
				_, err = os.Stat(config.FilePath)
				require.NoErrorf(t, err, "logger should create %s", config.FilePath)
			}

			req := &http.Request{
				Header: map[string][]string{},
				URL: &url.URL{
					Path: testPath,
				},
			}

			for _, s := range expectedValues {
				req.Header.Add(test.header, s)
			}

			chain := alice.New()
			chain = chain.Append(capture.Wrap)

			// Injection of the observability variables in the request context.
			chain = chain.Append(func(next http.Handler) (http.Handler, error) {
				return observability.WithObservabilityHandler(next, observability.Observability{
					AccessLogsEnabled: true,
				}), nil
			})

			chain = chain.Append(logger.AliceConstructor())
			handler, err := chain.Then(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
			}))
			require.NoError(t, err)
			handler.ServeHTTP(httptest.NewRecorder(), req)

			logData, err := os.ReadFile(logFile.Name())
			require.NoError(t, err)

			if test.expected == types.AccessLogDrop {
				assert.NotContains(t, string(logData), strings.Join(expectedValues, ","))
			} else {
				assert.Contains(t, string(logData), strings.Join(expectedValues, ","))
			}
		})
	}
}

func TestLoggerCLF(t *testing.T) {
	logFilePath := filepath.Join(t.TempDir(), logFileNameSuffix)
	config := &types.AccessLog{FilePath: logFilePath, Format: CommonFormat}
	doLogging(t, config, false)

	logData, err := os.ReadFile(logFilePath)
	require.NoError(t, err)

	expectedLog := ` TestHost - TestUser [13/Apr/2016:07:14:19 -0700] "POST testpath HTTP/0.0" 123 12 "testReferer" "testUserAgent" 1 "testRouter" "http://127.0.0.1/testService" 1ms`
	assertValidLogData(t, expectedLog, logData)
}

func TestLoggerCLFWithBufferingSize(t *testing.T) {
	logFilePath := filepath.Join(t.TempDir(), logFileNameSuffix)
	config := &types.AccessLog{FilePath: logFilePath, Format: CommonFormat, BufferingSize: 1024}
	doLogging(t, config, false)

	// wait a bit for the buffer to be written in the file.
	time.Sleep(50 * time.Millisecond)

	logData, err := os.ReadFile(logFilePath)
	require.NoError(t, err)

	expectedLog := ` TestHost - TestUser [13/Apr/2016:07:14:19 -0700] "POST testpath HTTP/0.0" 123 12 "testReferer" "testUserAgent" 1 "testRouter" "http://127.0.0.1/testService" 1ms`
	assertValidLogData(t, expectedLog, logData)
}

func assertString(exp string) func(t *testing.T, actual interface{}) {
	return func(t *testing.T, actual interface{}) {
		t.Helper()

		assert.Equal(t, exp, actual)
	}
}

func assertNotEmpty() func(t *testing.T, actual interface{}) {
	return func(t *testing.T, actual interface{}) {
		t.Helper()

		assert.NotEmpty(t, actual)
	}
}

func assertFloat64(exp float64) func(t *testing.T, actual interface{}) {
	return func(t *testing.T, actual interface{}) {
		t.Helper()

		assert.InDelta(t, exp, actual, delta)
	}
}

func assertFloat64NotZero() func(t *testing.T, actual interface{}) {
	return func(t *testing.T, actual interface{}) {
		t.Helper()

		assert.NotZero(t, actual)
	}
}

func TestLoggerJSON(t *testing.T) {
	testCases := []struct {
		desc     string
		config   *types.AccessLog
		tls      bool
		tracing  bool
		expected map[string]func(t *testing.T, value interface{})
	}{
		{
			desc: "default config without tracing",
			config: &types.AccessLog{
				FilePath: "",
				Format:   JSONFormat,
			},
			expected: map[string]func(t *testing.T, value interface{}){
				RequestContentSize:        assertFloat64(0),
				RequestHost:               assertString(testHostname),
				RequestAddr:               assertString(testHostname),
				RequestMethod:             assertString(testMethod),
				RequestPath:               assertString(testPath),
				RequestProtocol:           assertString(testProto),
				RequestScheme:             assertString(testScheme),
				RequestPort:               assertString("-"),
				DownstreamStatus:          assertFloat64(float64(testStatus)),
				DownstreamContentSize:     assertFloat64(float64(len(testContent))),
				OriginContentSize:         assertFloat64(float64(len(testContent))),
				OriginStatus:              assertFloat64(float64(testStatus)),
				RequestRefererHeader:      assertString(testReferer),
				RequestUserAgentHeader:    assertString(testUserAgent),
				RouterName:                assertString(testRouterName),
				ServiceURL:                assertString(testServiceName),
				ClientUsername:            assertString(testUsername),
				ClientHost:                assertString(testHostname),
				ClientPort:                assertString(strconv.Itoa(testPort)),
				ClientAddr:                assertString(fmt.Sprintf("%s:%d", testHostname, testPort)),
				"level":                   assertString("info"),
				"msg":                     assertString(""),
				"downstream_Content-Type": assertString("text/plain; charset=utf-8"),
				RequestCount:              assertFloat64NotZero(),
				Duration:                  assertFloat64NotZero(),
				Overhead:                  assertFloat64NotZero(),
				RetryAttempts:             assertFloat64(float64(testRetryAttempts)),
				"time":                    assertNotEmpty(),
				"StartLocal":              assertNotEmpty(),
				"StartUTC":                assertNotEmpty(),
			},
		},
		{
			desc: "default config with tracing",
			config: &types.AccessLog{
				FilePath: "",
				Format:   JSONFormat,
			},
			tracing: true,
			expected: map[string]func(t *testing.T, value interface{}){
				RequestContentSize:        assertFloat64(0),
				RequestHost:               assertString(testHostname),
				RequestAddr:               assertString(testHostname),
				RequestMethod:             assertString(testMethod),
				RequestPath:               assertString(testPath),
				RequestProtocol:           assertString(testProto),
				RequestScheme:             assertString(testScheme),
				RequestPort:               assertString("-"),
				DownstreamStatus:          assertFloat64(float64(testStatus)),
				DownstreamContentSize:     assertFloat64(float64(len(testContent))),
				OriginContentSize:         assertFloat64(float64(len(testContent))),
				OriginStatus:              assertFloat64(float64(testStatus)),
				RequestRefererHeader:      assertString(testReferer),
				RequestUserAgentHeader:    assertString(testUserAgent),
				RouterName:                assertString(testRouterName),
				ServiceURL:                assertString(testServiceName),
				ClientUsername:            assertString(testUsername),
				ClientHost:                assertString(testHostname),
				ClientPort:                assertString(strconv.Itoa(testPort)),
				ClientAddr:                assertString(fmt.Sprintf("%s:%d", testHostname, testPort)),
				"level":                   assertString("info"),
				"msg":                     assertString(""),
				"downstream_Content-Type": assertString("text/plain; charset=utf-8"),
				RequestCount:              assertFloat64NotZero(),
				Duration:                  assertFloat64NotZero(),
				Overhead:                  assertFloat64NotZero(),
				RetryAttempts:             assertFloat64(float64(testRetryAttempts)),
				"time":                    assertNotEmpty(),
				"StartLocal":              assertNotEmpty(),
				"StartUTC":                assertNotEmpty(),
				TraceID:                   assertString("01000000000000000000000000000000"),
				SpanID:                    assertString("0100000000000000"),
			},
		},
		{
			desc: "default config, with TLS request",
			config: &types.AccessLog{
				FilePath: "",
				Format:   JSONFormat,
			},
			tls: true,
			expected: map[string]func(t *testing.T, value interface{}){
				RequestContentSize:        assertFloat64(0),
				RequestHost:               assertString(testHostname),
				RequestAddr:               assertString(testHostname),
				RequestMethod:             assertString(testMethod),
				RequestPath:               assertString(testPath),
				RequestProtocol:           assertString(testProto),
				RequestScheme:             assertString("https"),
				RequestPort:               assertString("-"),
				DownstreamStatus:          assertFloat64(float64(testStatus)),
				DownstreamContentSize:     assertFloat64(float64(len(testContent))),
				OriginContentSize:         assertFloat64(float64(len(testContent))),
				OriginStatus:              assertFloat64(float64(testStatus)),
				RequestRefererHeader:      assertString(testReferer),
				RequestUserAgentHeader:    assertString(testUserAgent),
				RouterName:                assertString(testRouterName),
				ServiceURL:                assertString(testServiceName),
				ClientUsername:            assertString(testUsername),
				ClientHost:                assertString(testHostname),
				ClientPort:                assertString(strconv.Itoa(testPort)),
				ClientAddr:                assertString(fmt.Sprintf("%s:%d", testHostname, testPort)),
				"level":                   assertString("info"),
				"msg":                     assertString(""),
				"downstream_Content-Type": assertString("text/plain; charset=utf-8"),
				RequestCount:              assertFloat64NotZero(),
				Duration:                  assertFloat64NotZero(),
				Overhead:                  assertFloat64NotZero(),
				RetryAttempts:             assertFloat64(float64(testRetryAttempts)),
				TLSClientSubject:          assertString("CN=foobar"),
				TLSVersion:                assertString("1.3"),
				TLSCipher:                 assertString("TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256"),
				"time":                    assertNotEmpty(),
				StartLocal:                assertNotEmpty(),
				StartUTC:                  assertNotEmpty(),
			},
		},
		{
			desc: "default config drop all fields",
			config: &types.AccessLog{
				FilePath: "",
				Format:   JSONFormat,
				Fields: &types.AccessLogFields{
					DefaultMode: "drop",
				},
			},
			expected: map[string]func(t *testing.T, value interface{}){
				"level":                   assertString("info"),
				"msg":                     assertString(""),
				"time":                    assertNotEmpty(),
				"downstream_Content-Type": assertString("text/plain; charset=utf-8"),
				RequestRefererHeader:      assertString(testReferer),
				RequestUserAgentHeader:    assertString(testUserAgent),
			},
		},
		{
			desc: "default config drop all fields and headers",
			config: &types.AccessLog{
				FilePath: "",
				Format:   JSONFormat,
				Fields: &types.AccessLogFields{
					DefaultMode: "drop",
					Headers: &types.FieldHeaders{
						DefaultMode: "drop",
					},
				},
			},
			expected: map[string]func(t *testing.T, value interface{}){
				"level": assertString("info"),
				"msg":   assertString(""),
				"time":  assertNotEmpty(),
			},
		},
		{
			desc: "default config drop all fields and redact headers",
			config: &types.AccessLog{
				FilePath: "",
				Format:   JSONFormat,
				Fields: &types.AccessLogFields{
					DefaultMode: "drop",
					Headers: &types.FieldHeaders{
						DefaultMode: "redact",
					},
				},
			},
			expected: map[string]func(t *testing.T, value interface{}){
				"level":                   assertString("info"),
				"msg":                     assertString(""),
				"time":                    assertNotEmpty(),
				"downstream_Content-Type": assertString("REDACTED"),
				RequestRefererHeader:      assertString("REDACTED"),
				RequestUserAgentHeader:    assertString("REDACTED"),
			},
		},
		{
			desc: "default config drop all fields and headers but kept someone",
			config: &types.AccessLog{
				FilePath: "",
				Format:   JSONFormat,
				Fields: &types.AccessLogFields{
					DefaultMode: "drop",
					Names: map[string]string{
						RequestHost: "keep",
					},
					Headers: &types.FieldHeaders{
						DefaultMode: "drop",
						Names: map[string]string{
							"Referer": "keep",
						},
					},
				},
			},
			expected: map[string]func(t *testing.T, value interface{}){
				RequestHost:          assertString(testHostname),
				"level":              assertString("info"),
				"msg":                assertString(""),
				"time":               assertNotEmpty(),
				RequestRefererHeader: assertString(testReferer),
			},
		},
		{
			desc: "fields and headers with unconventional letter case",
			config: &types.AccessLog{
				FilePath: "",
				Format:   JSONFormat,
				Fields: &types.AccessLogFields{
					DefaultMode: "drop",
					Names: map[string]string{
						"rEqUeStHoSt": "keep",
					},
					Headers: &types.FieldHeaders{
						DefaultMode: "drop",
						Names: map[string]string{
							"ReFeReR": "keep",
						},
					},
				},
			},
			expected: map[string]func(t *testing.T, value interface{}){
				RequestHost:          assertString(testHostname),
				"level":              assertString("info"),
				"msg":                assertString(""),
				"time":               assertNotEmpty(),
				RequestRefererHeader: assertString(testReferer),
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			logFilePath := filepath.Join(t.TempDir(), logFileNameSuffix)

			test.config.FilePath = logFilePath
			if test.tls {
				doLoggingTLS(t, test.config, test.tracing)
			} else {
				doLogging(t, test.config, test.tracing)
			}

			logData, err := os.ReadFile(logFilePath)
			require.NoError(t, err)

			jsonData := make(map[string]interface{})
			err = json.Unmarshal(logData, &jsonData)
			require.NoError(t, err)

			assert.Len(t, jsonData, len(test.expected))

			for field, assertion := range test.expected {
				assertion(t, jsonData[field])
			}
		})
	}
}

func TestLogger_AbortedRequest(t *testing.T) {
	expected := map[string]func(t *testing.T, value interface{}){
		RequestContentSize:             assertFloat64(0),
		RequestHost:                    assertString(testHostname),
		RequestAddr:                    assertString(testHostname),
		RequestMethod:                  assertString(testMethod),
		RequestPath:                    assertString(""),
		RequestProtocol:                assertString(testProto),
		RequestScheme:                  assertString(testScheme),
		RequestPort:                    assertString("-"),
		DownstreamStatus:               assertFloat64(float64(200)),
		DownstreamContentSize:          assertFloat64(float64(40)),
		RequestRefererHeader:           assertString(testReferer),
		RequestUserAgentHeader:         assertString(testUserAgent),
		ServiceURL:                     assertString("http://stream"),
		ServiceAddr:                    assertString("127.0.0.1"),
		ServiceName:                    assertString("stream"),
		ClientUsername:                 assertString(testUsername),
		ClientHost:                     assertString(testHostname),
		ClientPort:                     assertString(strconv.Itoa(testPort)),
		ClientAddr:                     assertString(fmt.Sprintf("%s:%d", testHostname, testPort)),
		"level":                        assertString("info"),
		"msg":                          assertString(""),
		RequestCount:                   assertFloat64NotZero(),
		Duration:                       assertFloat64NotZero(),
		Overhead:                       assertFloat64NotZero(),
		RetryAttempts:                  assertFloat64(float64(0)),
		"time":                         assertNotEmpty(),
		StartLocal:                     assertNotEmpty(),
		StartUTC:                       assertNotEmpty(),
		"downstream_Content-Type":      assertString("text/plain"),
		"downstream_Transfer-Encoding": assertString("chunked"),
		"downstream_Cache-Control":     assertString("no-cache"),
	}

	config := &types.AccessLog{
		FilePath: filepath.Join(t.TempDir(), logFileNameSuffix),
		Format:   JSONFormat,
	}
	doLoggingWithAbortedStream(t, config)

	logData, err := os.ReadFile(config.FilePath)
	require.NoError(t, err)

	jsonData := make(map[string]interface{})
	err = json.Unmarshal(logData, &jsonData)
	require.NoError(t, err)

	assert.Len(t, jsonData, len(expected))

	for field, assertion := range expected {
		assertion(t, jsonData[field])
		if t.Failed() {
			return
		}
	}
}

func TestNewLogHandlerOutputStdout(t *testing.T) {
	testCases := []struct {
		desc        string
		config      *types.AccessLog
		expectedLog string
	}{
		{
			desc: "default config",
			config: &types.AccessLog{
				FilePath: "",
				Format:   CommonFormat,
			},
			expectedLog: `TestHost - TestUser [13/Apr/2016:07:14:19 -0700] "POST testpath HTTP/0.0" 123 12 "testReferer" "testUserAgent" 23 "testRouter" "http://127.0.0.1/testService" 1ms`,
		},
		{
			desc: "default config with empty filters",
			config: &types.AccessLog{
				FilePath: "",
				Format:   CommonFormat,
				Filters:  &types.AccessLogFilters{},
			},
			expectedLog: `TestHost - TestUser [13/Apr/2016:07:14:19 -0700] "POST testpath HTTP/0.0" 123 12 "testReferer" "testUserAgent" 23 "testRouter" "http://127.0.0.1/testService" 1ms`,
		},
		{
			desc: "Status code filter not matching",
			config: &types.AccessLog{
				FilePath: "",
				Format:   CommonFormat,
				Filters: &types.AccessLogFilters{
					StatusCodes: []string{"200"},
				},
			},
			expectedLog: ``,
		},
		{
			desc: "Status code filter matching",
			config: &types.AccessLog{
				FilePath: "",
				Format:   CommonFormat,
				Filters: &types.AccessLogFilters{
					StatusCodes: []string{"123"},
				},
			},
			expectedLog: `TestHost - TestUser [13/Apr/2016:07:14:19 -0700] "POST testpath HTTP/0.0" 123 12 "testReferer" "testUserAgent" 23 "testRouter" "http://127.0.0.1/testService" 1ms`,
		},
		{
			desc: "Duration filter not matching",
			config: &types.AccessLog{
				FilePath: "",
				Format:   CommonFormat,
				Filters: &types.AccessLogFilters{
					MinDuration: ptypes.Duration(1 * time.Hour),
				},
			},
			expectedLog: ``,
		},
		{
			desc: "Duration filter matching",
			config: &types.AccessLog{
				FilePath: "",
				Format:   CommonFormat,
				Filters: &types.AccessLogFilters{
					MinDuration: ptypes.Duration(1 * time.Millisecond),
				},
			},
			expectedLog: `TestHost - TestUser [13/Apr/2016:07:14:19 -0700] "POST testpath HTTP/0.0" 123 12 "testReferer" "testUserAgent" 23 "testRouter" "http://127.0.0.1/testService" 1ms`,
		},
		{
			desc: "Retry attempts filter matching",
			config: &types.AccessLog{
				FilePath: "",
				Format:   CommonFormat,
				Filters: &types.AccessLogFilters{
					RetryAttempts: true,
				},
			},
			expectedLog: `TestHost - TestUser [13/Apr/2016:07:14:19 -0700] "POST testpath HTTP/0.0" 123 12 "testReferer" "testUserAgent" 23 "testRouter" "http://127.0.0.1/testService" 1ms`,
		},
		{
			desc: "Default mode keep",
			config: &types.AccessLog{
				FilePath: "",
				Format:   CommonFormat,
				Fields: &types.AccessLogFields{
					DefaultMode: "keep",
				},
			},
			expectedLog: `TestHost - TestUser [13/Apr/2016:07:14:19 -0700] "POST testpath HTTP/0.0" 123 12 "testReferer" "testUserAgent" 23 "testRouter" "http://127.0.0.1/testService" 1ms`,
		},
		{
			desc: "Default mode keep with override",
			config: &types.AccessLog{
				FilePath: "",
				Format:   CommonFormat,
				Fields: &types.AccessLogFields{
					DefaultMode: "keep",
					Names: map[string]string{
						ClientHost: "drop",
					},
				},
			},
			expectedLog: `- - TestUser [13/Apr/2016:07:14:19 -0700] "POST testpath HTTP/0.0" 123 12 "testReferer" "testUserAgent" 23 "testRouter" "http://127.0.0.1/testService" 1ms`,
		},
		{
			desc: "Default mode drop",
			config: &types.AccessLog{
				FilePath: "",
				Format:   CommonFormat,
				Fields: &types.AccessLogFields{
					DefaultMode: "drop",
				},
			},
			expectedLog: `- - - [-] "- - -" - - "testReferer" "testUserAgent" - "-" "-" 0ms`,
		},
		{
			desc: "Default mode drop with override",
			config: &types.AccessLog{
				FilePath: "",
				Format:   CommonFormat,
				Fields: &types.AccessLogFields{
					DefaultMode: "drop",
					Names: map[string]string{
						ClientHost:     "drop",
						ClientUsername: "keep",
					},
				},
			},
			expectedLog: `- - TestUser [-] "- - -" - - "testReferer" "testUserAgent" - "-" "-" 0ms`,
		},
		{
			desc: "Default mode drop with header dropped",
			config: &types.AccessLog{
				FilePath: "",
				Format:   CommonFormat,
				Fields: &types.AccessLogFields{
					DefaultMode: "drop",
					Names: map[string]string{
						ClientHost:     "drop",
						ClientUsername: "keep",
					},
					Headers: &types.FieldHeaders{
						DefaultMode: "drop",
					},
				},
			},
			expectedLog: `- - TestUser [-] "- - -" - - "-" "-" - "-" "-" 0ms`,
		},
		{
			desc: "Default mode drop with header redacted",
			config: &types.AccessLog{
				FilePath: "",
				Format:   CommonFormat,
				Fields: &types.AccessLogFields{
					DefaultMode: "drop",
					Names: map[string]string{
						ClientHost:     "drop",
						ClientUsername: "keep",
					},
					Headers: &types.FieldHeaders{
						DefaultMode: "redact",
					},
				},
			},
			expectedLog: `- - TestUser [-] "- - -" - - "REDACTED" "REDACTED" - "-" "-" 0ms`,
		},
		{
			desc: "Default mode drop with header redacted",
			config: &types.AccessLog{
				FilePath: "",
				Format:   CommonFormat,
				Fields: &types.AccessLogFields{
					DefaultMode: "drop",
					Names: map[string]string{
						ClientHost:     "drop",
						ClientUsername: "keep",
					},
					Headers: &types.FieldHeaders{
						DefaultMode: "keep",
						Names: map[string]string{
							"Referer": "redact",
						},
					},
				},
			},
			expectedLog: `- - TestUser [-] "- - -" - - "REDACTED" "testUserAgent" - "-" "-" 0ms`,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			// NOTE: It is not possible to run these cases in parallel because we capture Stdout

			file, restoreStdout := captureStdout(t)
			defer restoreStdout()

			doLogging(t, test.config, false)

			written, err := os.ReadFile(file.Name())
			require.NoError(t, err, "unable to read captured stdout from file")
			assertValidLogData(t, test.expectedLog, written)
		})
	}
}

func assertValidLogData(t *testing.T, expected string, logData []byte) {
	t.Helper()

	if len(expected) == 0 {
		assert.Empty(t, logData)
		t.Log(string(logData))
		return
	}

	result, err := ParseAccessLog(string(logData))
	require.NoError(t, err)

	resultExpected, err := ParseAccessLog(expected)
	require.NoError(t, err)

	formatErrMessage := fmt.Sprintf("Expected:\t%q\nActual:\t%q", expected, string(logData))

	require.Len(t, result, len(resultExpected), formatErrMessage)
	assert.Equal(t, resultExpected[ClientHost], result[ClientHost], formatErrMessage)
	assert.Equal(t, resultExpected[ClientUsername], result[ClientUsername], formatErrMessage)
	assert.Equal(t, resultExpected[RequestMethod], result[RequestMethod], formatErrMessage)
	assert.Equal(t, resultExpected[RequestPath], result[RequestPath], formatErrMessage)
	assert.Equal(t, resultExpected[RequestProtocol], result[RequestProtocol], formatErrMessage)
	assert.Equal(t, resultExpected[OriginStatus], result[OriginStatus], formatErrMessage)
	assert.Equal(t, resultExpected[OriginContentSize], result[OriginContentSize], formatErrMessage)
	assert.Equal(t, resultExpected[RequestRefererHeader], result[RequestRefererHeader], formatErrMessage)
	assert.Equal(t, resultExpected[RequestUserAgentHeader], result[RequestUserAgentHeader], formatErrMessage)
	assert.Regexp(t, `\d*`, result[RequestCount], formatErrMessage)
	assert.Equal(t, resultExpected[RouterName], result[RouterName], formatErrMessage)
	assert.Equal(t, resultExpected[ServiceURL], result[ServiceURL], formatErrMessage)
	assert.Regexp(t, `\d*ms`, result[Duration], formatErrMessage)
}

func captureStdout(t *testing.T) (out *os.File, restoreStdout func()) {
	t.Helper()

	file, err := os.CreateTemp(t.TempDir(), "testlogger")
	require.NoError(t, err, "failed to create temp file")

	original := os.Stdout
	os.Stdout = file

	restoreStdout = func() {
		os.Stdout = original
		_ = os.RemoveAll(file.Name())
	}

	return file, restoreStdout
}

func doLoggingTLSOpt(t *testing.T, config *types.AccessLog, enableTLS, tracing bool) {
	t.Helper()
	logger, err := NewHandler(t.Context(), config)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := logger.Close()
		require.NoError(t, err)
	})

	if config.FilePath != "" {
		_, err = os.Stat(config.FilePath)
		require.NoErrorf(t, err, "logger should create %s", config.FilePath)
	}

	req := &http.Request{
		Header: map[string][]string{
			"User-Agent": {testUserAgent},
			"Referer":    {testReferer},
		},
		Proto:      testProto,
		Host:       testHostname,
		Method:     testMethod,
		RemoteAddr: fmt.Sprintf("%s:%d", testHostname, testPort),
		URL: &url.URL{
			User: url.UserPassword(testUsername, ""),
			Path: testPath,
		},
		Body: io.NopCloser(bytes.NewReader([]byte("bar"))),
	}
	if enableTLS {
		req.TLS = &tls.ConnectionState{
			Version:     tls.VersionTLS13,
			CipherSuite: tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			PeerCertificates: []*x509.Certificate{{
				Subject: pkix.Name{CommonName: "foobar"},
			}},
		}
	}

	if tracing {
		contextWithSpan := trace.ContextWithSpan(req.Context(), &mockSpan{})
		req = req.WithContext(contextWithSpan)
	}

	chain := alice.New()
	chain = chain.Append(capture.Wrap)

	// Injection of the observability variables in the request context.
	chain = chain.Append(func(next http.Handler) (http.Handler, error) {
		return observability.WithObservabilityHandler(next, observability.Observability{
			AccessLogsEnabled: true,
		}), nil
	})

	chain = chain.Append(logger.AliceConstructor())
	handler, err := chain.Then(http.HandlerFunc(logWriterTestHandlerFunc))
	require.NoError(t, err)

	handler.ServeHTTP(httptest.NewRecorder(), req)
}

func doLoggingTLS(t *testing.T, config *types.AccessLog, tracing bool) {
	t.Helper()

	doLoggingTLSOpt(t, config, true, tracing)
}

func doLogging(t *testing.T, config *types.AccessLog, tracing bool) {
	t.Helper()

	doLoggingTLSOpt(t, config, false, tracing)
}

func logWriterTestHandlerFunc(rw http.ResponseWriter, r *http.Request) {
	if _, err := rw.Write([]byte(testContent)); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	logData := GetLogData(r)
	if logData != nil {
		logData.Core[RouterName] = testRouterName
		logData.Core[ServiceURL] = testServiceName
		logData.Core[OriginStatus] = testStatus
		logData.Core[OriginContentSize] = testContentSize
		logData.Core[RetryAttempts] = testRetryAttempts
		logData.Core[StartUTC] = testStart.UTC()
		logData.Core[StartLocal] = testStart.Local()
	} else {
		http.Error(rw, "LogData is nil", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(testStatus)
}

func doLoggingWithAbortedStream(t *testing.T, config *types.AccessLog) {
	t.Helper()

	logger, err := NewHandler(t.Context(), config)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := logger.Close()
		require.NoError(t, err)
	})

	if config.FilePath != "" {
		_, err = os.Stat(config.FilePath)
		require.NoError(t, err, "logger should create "+config.FilePath)
	}

	reqContext, cancelRequest := context.WithCancel(t.Context())

	req := &http.Request{
		Header: map[string][]string{
			"User-Agent": {testUserAgent},
			"Referer":    {testReferer},
		},
		Proto:      testProto,
		Host:       testHostname,
		Method:     testMethod,
		RemoteAddr: fmt.Sprintf("%s:%d", testHostname, testPort),
		URL: &url.URL{
			User: url.UserPassword(testUsername, ""),
		},
		Body: nil,
	}

	req = req.WithContext(reqContext)

	chain := alice.New()

	chain = chain.Append(func(next http.Handler) (http.Handler, error) {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			defer func() {
				_ = recover() // ignore the stream backend panic to avoid the test to fail.
			}()
			next.ServeHTTP(rw, req)
		}), nil
	})
	chain = chain.Append(capture.Wrap)

	// Injection of the observability variables in the request context.
	chain = chain.Append(func(next http.Handler) (http.Handler, error) {
		return observability.WithObservabilityHandler(next, observability.Observability{
			AccessLogsEnabled: true,
		}), nil
	})

	chain = chain.Append(logger.AliceConstructor())

	service := NewFieldHandler(http.HandlerFunc(streamBackend), ServiceURL, "http://stream", nil)
	service = NewFieldHandler(service, ServiceAddr, "127.0.0.1", nil)
	service = NewFieldHandler(service, ServiceName, "stream", AddServiceFields)

	handler, err := chain.Then(service)
	require.NoError(t, err)

	go func() {
		time.Sleep(499 * time.Millisecond)
		cancelRequest()
	}()

	handler.ServeHTTP(httptest.NewRecorder(), req)
}

func streamBackend(rw http.ResponseWriter, r *http.Request) {
	// Get the Flusher to flush the response to the client
	flusher, ok := rw.(http.Flusher)
	if !ok {
		http.Error(rw, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// Set the headers for streaming
	rw.Header().Set("Content-Type", "text/plain")
	rw.Header().Set("Transfer-Encoding", "chunked")
	rw.Header().Set("Cache-Control", "no-cache")

	for {
		time.Sleep(100 * time.Millisecond)

		select {
		case <-r.Context().Done():
			panic(http.ErrAbortHandler)

		default:
			if _, err := fmt.Fprint(rw, "FOOBAR!!!!"); err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
			flusher.Flush()
		}
	}
}

// mockSpan is an implementation of Span that preforms no operations.
type mockSpan struct {
	embedded.Span
}

var _ trace.Span = &mockSpan{}

func (*mockSpan) SpanContext() trace.SpanContext {
	return trace.NewSpanContext(trace.SpanContextConfig{TraceID: trace.TraceID{1}, SpanID: trace.SpanID{1}})
}
func (*mockSpan) IsRecording() bool                             { return true }
func (s *mockSpan) SetStatus(_ codes.Code, _ string)            {}
func (s *mockSpan) SetAttributes(...attribute.KeyValue)         {}
func (s *mockSpan) End(...trace.SpanEndOption)                  {}
func (s *mockSpan) RecordError(_ error, _ ...trace.EventOption) {}
func (s *mockSpan) AddEvent(_ string, _ ...trace.EventOption)   {}
func (s *mockSpan) AddLink(_ trace.Link)                        {}

func (s *mockSpan) SetName(_ string) {}

func (s *mockSpan) TracerProvider() trace.TracerProvider {
	return nil
}
