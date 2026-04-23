package accesslog

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/containous/alice"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/middlewares/capture"
	"github.com/traefik/traefik/v3/pkg/middlewares/observability"
	otypes "github.com/traefik/traefik/v3/pkg/observability/types"
)

func TestCommonLogFormatter_Format(t *testing.T) {
	clf := CommonLogFormatter{}

	testCases := []struct {
		name        string
		data        map[string]any
		expectedLog string
	}{
		{
			name: "DownstreamStatus & DownstreamContentSize are nil",
			data: map[string]any{
				StartUTC:               time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
				Duration:               123 * time.Second,
				ClientHost:             "10.0.0.1",
				ClientUsername:         "Client",
				RequestMethod:          http.MethodGet,
				RequestPath:            "/foo",
				RequestProtocol:        "http",
				DownstreamStatus:       nil,
				DownstreamContentSize:  nil,
				RequestRefererHeader:   "",
				RequestUserAgentHeader: "",
				RequestCount:           0,
				RouterName:             "",
				ServiceURL:             "",
			},
			expectedLog: `10.0.0.1 - Client [10/Nov/2009:23:00:00 +0000] "GET /foo http" - - "-" "-" 0 "-" "-" 123000ms
`,
		},
		{
			name: "all data",
			data: map[string]any{
				StartUTC:               time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
				Duration:               123 * time.Second,
				ClientHost:             "10.0.0.1",
				ClientUsername:         "Client",
				RequestMethod:          http.MethodGet,
				RequestPath:            "/foo",
				RequestProtocol:        "http",
				DownstreamStatus:       123,
				DownstreamContentSize:  132,
				RequestRefererHeader:   "referer",
				RequestUserAgentHeader: "agent",
				RequestCount:           nil,
				RouterName:             "foo",
				ServiceURL:             "http://10.0.0.2/toto",
			},
			expectedLog: `10.0.0.1 - Client [10/Nov/2009:23:00:00 +0000] "GET /foo http" 123 132 "referer" "agent" - "foo" "http://10.0.0.2/toto" 123000ms
`,
		},
		{
			name: "all data with local time",
			data: map[string]any{
				StartLocal:             time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
				Duration:               123 * time.Second,
				ClientHost:             "10.0.0.1",
				ClientUsername:         "Client",
				RequestMethod:          http.MethodGet,
				RequestPath:            "/foo",
				RequestProtocol:        "http",
				DownstreamStatus:       123,
				DownstreamContentSize:  132,
				RequestRefererHeader:   "referer",
				RequestUserAgentHeader: "agent",
				RequestCount:           nil,
				RouterName:             "foo",
				ServiceURL:             "http://10.0.0.2/toto",
			},
			expectedLog: `10.0.0.1 - Client [10/Nov/2009:14:00:00 -0900] "GET /foo http" 123 132 "referer" "agent" - "foo" "http://10.0.0.2/toto" 123000ms
`,
		},
	}

	var err error
	time.Local, err = time.LoadLocation("Etc/GMT+9")
	require.NoError(t, err)

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			entry := &logrus.Entry{Data: test.data}

			raw, err := clf.Format(entry)
			assert.NoError(t, err)

			assert.Equal(t, test.expectedLog, string(raw))
		})
	}
}

func TestGenericCLFLogFormatter_Format(t *testing.T) {
	clf := GenericCLFLogFormatter{}

	testCases := []struct {
		name        string
		data        map[string]any
		expectedLog string
	}{
		{
			name: "DownstreamStatus & DownstreamContentSize are nil",
			data: map[string]any{
				StartUTC:               time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
				Duration:               123 * time.Second,
				ClientHost:             "10.0.0.1",
				ClientUsername:         "Client",
				RequestMethod:          http.MethodGet,
				RequestPath:            "/foo",
				RequestProtocol:        "http",
				DownstreamStatus:       nil,
				DownstreamContentSize:  nil,
				RequestRefererHeader:   "",
				RequestUserAgentHeader: "",
				RequestCount:           0,
				RouterName:             "",
				ServiceURL:             "",
			},
			expectedLog: `10.0.0.1 - Client [10/Nov/2009:23:00:00 +0000] "GET /foo http" - - "-" "-"
`,
		},
		{
			name: "all data",
			data: map[string]any{
				StartUTC:               time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
				Duration:               123 * time.Second,
				ClientHost:             "10.0.0.1",
				ClientUsername:         "Client",
				RequestMethod:          http.MethodGet,
				RequestPath:            "/foo",
				RequestProtocol:        "http",
				DownstreamStatus:       123,
				DownstreamContentSize:  132,
				RequestRefererHeader:   "referer",
				RequestUserAgentHeader: "agent",
				RequestCount:           nil,
				RouterName:             "foo",
				ServiceURL:             "http://10.0.0.2/toto",
			},
			expectedLog: `10.0.0.1 - Client [10/Nov/2009:23:00:00 +0000] "GET /foo http" 123 132 "referer" "agent"
`,
		},
		{
			name: "all data with local time",
			data: map[string]any{
				StartLocal:             time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
				Duration:               123 * time.Second,
				ClientHost:             "10.0.0.1",
				ClientUsername:         "Client",
				RequestMethod:          http.MethodGet,
				RequestPath:            "/foo",
				RequestProtocol:        "http",
				DownstreamStatus:       123,
				DownstreamContentSize:  132,
				RequestRefererHeader:   "referer",
				RequestUserAgentHeader: "agent",
				RequestCount:           nil,
				RouterName:             "foo",
				ServiceURL:             "http://10.0.0.2/toto",
			},
			expectedLog: `10.0.0.1 - Client [10/Nov/2009:14:00:00 -0900] "GET /foo http" 123 132 "referer" "agent"
`,
		},
	}

	var err error
	time.Local, err = time.LoadLocation("Etc/GMT+9")
	require.NoError(t, err)

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			entry := &logrus.Entry{Data: test.data}

			raw, err := clf.Format(entry)
			assert.NoError(t, err)

			assert.Equal(t, test.expectedLog, string(raw))
		})
	}
}

func Test_toLog(t *testing.T) {
	testCases := []struct {
		desc         string
		fields       logrus.Fields
		fieldName    string
		defaultValue string
		quoted       bool
		expectedLog  any
	}{
		{
			desc: "Should return int 1",
			fields: logrus.Fields{
				"Powpow": 1,
			},
			fieldName:    "Powpow",
			defaultValue: defaultValue,
			quoted:       false,
			expectedLog:  1,
		},
		{
			desc: "Should return string foo",
			fields: logrus.Fields{
				"Powpow": "foo",
			},
			fieldName:    "Powpow",
			defaultValue: defaultValue,
			quoted:       true,
			expectedLog:  `"foo"`,
		},
		{
			desc: "Should return defaultValue if fieldName does not exist",
			fields: logrus.Fields{
				"Powpow": "foo",
			},
			fieldName:    "",
			defaultValue: defaultValue,
			quoted:       false,
			expectedLog:  "-",
		},
		{
			desc:         "Should return defaultValue if fields is nil",
			fields:       nil,
			fieldName:    "",
			defaultValue: defaultValue,
			quoted:       false,
			expectedLog:  "-",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			lg := toLog(test.fields, test.fieldName, defaultValue, test.quoted)

			assert.Equal(t, test.expectedLog, lg)
		})
	}
}

func TestNewTemplateJSONFormatter_InvalidTemplate(t *testing.T) {
	_, err := NewTemplateJSONFormatter(`{{ .Unclosed`)
	require.Error(t, err)
}

func TestTemplateJSONFormatter_Format(t *testing.T) {
	f, err := NewTemplateJSONFormatter(`{"status":{{ index . "DownstreamStatus" }},"method":"{{ index . "RequestMethod" }}"}`)
	require.NoError(t, err)

	entry := &logrus.Entry{
		Logger: logrus.New(),
		Data: logrus.Fields{
			DownstreamStatus: 200,
			RequestMethod:    "GET",
		},
	}

	out, err := f.Format(entry)
	require.NoError(t, err)
	assert.JSONEq(t, `{"status":200,"method":"GET"}`, strings.TrimSuffix(string(out), "\n"))
}

func TestTemplateJSONFormatter_FormatIncludesBuiltins(t *testing.T) {
	f, err := NewTemplateJSONFormatter(`{"level":"{{ index . "level" }}","msg":"{{ index . "msg" }}"}`)
	require.NoError(t, err)

	entry := &logrus.Entry{
		Logger:  logrus.New(),
		Level:   logrus.InfoLevel,
		Message: "hello",
		Data:    logrus.Fields{},
	}

	out, err := f.Format(entry)
	require.NoError(t, err)
	assert.JSONEq(t, `{"level":"info","msg":"hello"}`, strings.TrimSuffix(string(out), "\n"))
}

func TestTemplateJSONFormatter_AutoEscape(t *testing.T) {
	f, err := NewTemplateJSONFormatter(`{"path":"{{ index . "RequestPath" }}"}`)
	require.NoError(t, err)

	entry := &logrus.Entry{
		Logger: logrus.New(),
		Data:   logrus.Fields{RequestPath: `/foo"bar\baz`},
	}

	out, err := f.Format(entry)
	require.NoError(t, err)
	assert.JSONEq(t, `{"path":"/foo\"bar\\baz"}`, strings.TrimSuffix(string(out), "\n"))
}

func TestTemplateJSONFormatter_JSONHelper_SafeEscaping(t *testing.T) {
	f, err := NewTemplateJSONFormatter(`{"path":{{ json (index . "RequestPath") }}}`)
	require.NoError(t, err)

	entry := &logrus.Entry{
		Logger: logrus.New(),
		Data:   logrus.Fields{RequestPath: `/foo"bar\baz`},
	}

	out, err := f.Format(entry)
	require.NoError(t, err)
	assert.JSONEq(t, `{"path":"/foo\"bar\\baz"}`, strings.TrimSuffix(string(out), "\n"))
}

func TestTemplateJSONFormatter_InvalidJSONOutput(t *testing.T) {
	f, err := NewTemplateJSONFormatter(`not json at all`)
	require.NoError(t, err)

	entry := &logrus.Entry{Logger: logrus.New(), Data: logrus.Fields{}}
	_, err = f.Format(entry)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid JSON")
}

func TestTemplateJSONFormatter_NilField(t *testing.T) {
	f, err := NewTemplateJSONFormatter(`{"v":{{ json (index . "Field") }}}`)
	require.NoError(t, err)

	entry := &logrus.Entry{
		Logger: logrus.New(),
		Data:   logrus.Fields{"Field": nil},
	}

	out, err := f.Format(entry)
	require.NoError(t, err)
	assert.Equal(t, `{"v":null}`+"\n", string(out))
}

func TestTemplateJSONFormatter_MissingField(t *testing.T) {
	f, err := NewTemplateJSONFormatter(`{"v":{{ json (index . "NoSuchField") }}}`)
	require.NoError(t, err)

	entry := &logrus.Entry{Logger: logrus.New(), Data: logrus.Fields{}}

	out, err := f.Format(entry)
	require.NoError(t, err)
	// missing map key → nil → JSON null
	assert.Equal(t, `{"v":null}`+"\n", string(out))
}

func TestNewHandler_JSONTemplateInvalidFails(t *testing.T) {
	config := &otypes.AccessLog{
		Format:       JSONFormat,
		JSONTemplate: `{{ .Unclosed`,
	}
	_, err := NewHandler(context.Background(), config)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid accessLog.jsonTemplate")
}

func TestNewHandler_JSONTemplateWrongFormatFails(t *testing.T) {
	config := &otypes.AccessLog{
		Format:       CommonFormat,
		JSONTemplate: `{"status":{{ index . "DownstreamStatus" }}}`,
	}
	_, err := NewHandler(context.Background(), config)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "jsonTemplate requires format")
}

func TestHandler_JSONTemplateOutput(t *testing.T) {
	logFile := t.TempDir() + "/access.log"
	config := &otypes.AccessLog{
		FilePath:     logFile,
		Format:       JSONFormat,
		JSONTemplate: `{"s":{{ index . "DownstreamStatus" }},"m":"{{ index . "RequestMethod" }}"}`,
	}
	h, err := NewHandler(context.Background(), config)
	require.NoError(t, err)
	t.Cleanup(func() { _ = h.Close() })

	chain := alice.New()
	chain = chain.Append(capture.Wrap)
	chain = chain.Append(func(next http.Handler) (http.Handler, error) {
		return observability.WithObservabilityHandler(next, observability.Observability{
			AccessLogsEnabled: true,
		}), nil
	})
	chain = chain.Append(h.AliceConstructor())
	handler, err := chain.Then(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	raw, err := os.ReadFile(logFile)
	require.NoError(t, err)
	content := string(raw)
	assert.Contains(t, content, `"s":200`, "expected status 200 in output: %s", content)
	assert.Contains(t, content, `"m":"GET"`, "expected method GET in output: %s", content)
}
