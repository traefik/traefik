package accesslog

import (
	"bytes"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

type stubExiter struct {
	callCount int
}

func (exiter *stubExiter) Exit(code int) {
	exiter.callCount++
}

func init() {
	exiter = &stubExiter{}
}

func fixtureHeaders(kv ...string) http.Header {
	h := http.Header{}
	for i := 0; i < len(kv)-1; i += 2 {
		h.Set(kv[i], kv[i+1])
	}
	return h
}

var now = time.Now()

func fixtureLogDataTable(count uint64) *LogData {
	return &LogData{
		CoreLogData{
			StartUTC:              now,
			Duration:              2 * time.Millisecond,
			BackendName:           testBackendName,
			BackendURL:            testTargetURL,
			FrontendName:          testFrontendName,
			ClientHost:            testRemoteHost,
			ClientPort:            testRemotePort,
			ClientUsername:        testUsername,
			OriginDuration:        time.Millisecond,
			OriginContentSize:     102,
			RequestAddr:           testTarget,
			RequestHost:           testTargetHost,
			RequestPort:           testTargetPort,
			RequestMethod:         "GET",
			RequestPath:           "/y/xy/z",
			RequestProtocol:       "HTTP/1.1",
			OriginStatus:          200,
			DownstreamStatus:      200,
			DownstreamContentSize: 82,
			RequestCount:          count,
		},
		fixtureHeaders(
			"User-Agent", testUserAgent,
			"Accept", "*/*",
			"Referer", testReferrer),
		fixtureHeaders("Content-Length", "123",
			"Content-Type", "text/html; charset=utf-8",
			"Etag", `W/"32fa7-MkUKYk6PW5j6qz0Msr+EwA"`,
			"Set-Cookie", "UID=d5782b86abcb6f05cd7870f6e191cadd60999c9b57d42476ca7027322d74114c0; expires=Sun, 28-Feb-21 12:17:57 GMT; path=/; domain=.example.com",
			"Cache-Control", "private, max-age=0, must-revalidate",
			"Server", "foobar v1"),
		fixtureHeaders("Content-Length", "80",
			"Content-Type", "text/html; charset=utf-8",
			"Etag", `W/"32fa7-MkUKYk6PW5j6qz0Msr+EwA"`,
			"Set-Cookie", "UID=d5782b86abcb6f05cd7870f6e191cadd60999c9b57d42476ca7027322d74114c0; expires=Sun, 28-Feb-21 12:17:57 GMT; path=/; domain=.example.com",
			"Cache-Control", "private, max-age=0, must-revalidate",
			"Server", "foobar v1",
			"Location", "http://somewhere.else/a/b"),
	}
}

func TestCommonLogFormatter(t *testing.T) {
	la := commonLogFormatter{commonLogTimeFormat}
	buf := &bytes.Buffer{}
	la.Write(buf, fixtureLogDataTable(12345))
	s := buf.String()
	assert.Equal(t,
		`190.190.190.190 - - [`,
		s[:21])
	assert.Equal(t,
		`] "GET /y/xy/z HTTP/1.1" 200 102 "http://example.com/x/y/z" "user-agent-very-very-long-string" 12345 "frontend" "http://test.host.name:8181/a/b/c?q=1#z1" 2ms`+"\n",
		s[47:])
}

func TestJsonLogFormatter(t *testing.T) {
	exiter.(*stubExiter).callCount = 0

	jlf := newJSONLogFormatter(&types.AccessLog{
		TimeFormat: commonLogTimeFormat,
		CoreFields: []string{
			"StartUTC:StartUTC",
			"Duration:Duration",
			"FrontendName:thefrontend",
			"BackendName:BackendName",
			"BackendURL:BackendURL",
			"OriginDuration:OriginDuration",
			"OriginContentSize:OriginContentSize",
			"RequestAddr:RequestAddr",
			"RequestMethod:RequestMethod",
			"RequestPath:RequestPath",
			"RequestProtocol:RequestProtocol",
			"OriginStatus:OriginStatus",
			"DownstreamContentSize:DownstreamContentSize",
			"RequestCount:RequestCount",
			"ClientHost:ClHost",
			"ClientPort:ClPort",
			"ClientUsername:ClUsername",
		},
		RequestHeaders:            []string{"User-Agent: user_agent", "Referer: referrer"},
		OriginResponseHeaders:     []string{"Server: upstream_http_server"},
		DownstreamResponseHeaders: []string{"Location: sent_http_location"},
	})

	assert.Equal(t, 0, exiter.(*stubExiter).callCount)

	buf := &bytes.Buffer{}
	jlf.Write(buf, fixtureLogDataTable(12345))
	s := buf.String()
	assert.Equal(t,
		`{"StartUTC":"`,
		s[:13])
	assert.Equal(t,
		`","Duration":0.002,"thefrontend":"frontend","BackendName":"backend","BackendURL":"http://test.host.name:8181/a/b/c?q=1#z1","OriginDuration":0.001,`+
			`"OriginContentSize":102,`+
			`"RequestAddr":"test.host.name:8181","RequestMethod":"GET","RequestPath":"/y/xy/z","RequestProtocol":"HTTP/1.1",`+
			`"OriginStatus":200,"DownstreamContentSize":82,"RequestCount":12345,`+
			`"ClHost":"190.190.190.190","ClPort":"20121","ClUsername":"-",`+
			`"user_agent":"user-agent-very-very-long-string","referrer":"http://example.com/x/y/z",`+
			`"upstream_http_server":"foobar v1",`+
			`"sent_http_location":"http://somewhere.else/a/b"`+"}\n",
		s[39:])
}

func TestGzipRatioDivideByZero(t *testing.T) {
	jlf := newJSONLogFormatter(&types.AccessLog{
		TimeFormat: commonLogTimeFormat,
		CoreFields: []string{"FrontendName", "GzipRatio", "BackendName"},
	})

	buf := &bytes.Buffer{}
	data := fixtureLogDataTable(12345)
	data.Core[DownstreamContentSize] = 0 // cause the calculation to reach infinity
	jlf.Write(buf, data)
	s := buf.String()
	assert.Equal(t, `{"FrontendName":"frontend","BackendName":"backend"`+"}\n", s)
}

func TestNewJsonLogFormatterValidation(t *testing.T) {
	exiter.(*stubExiter).callCount = 0
	buf := &bytes.Buffer{}
	log.SetOutput(buf)
	defer log.SetOutput(os.Stderr)

	newJSONLogFormatter(&types.AccessLog{
		CoreFields:                []string{"Duration", "StartUTC", "FrontendName:frontend", "NonExistent"},
		RequestHeaders:            []string{"Host: http_host"},
		OriginResponseHeaders:     []string{"Server: upstream_http_server"},
		DownstreamResponseHeaders: []string{"Location: sent_http_location"},
	})

	assert.Equal(t, 1, exiter.(*stubExiter).callCount)

	logEntry := buf.String()

	strings.Index(logEntry, "level=")
	assert.True(t, strings.Contains(logEntry, "Unsupported access log fields: [NonExistent]"), "%s", logEntry)
}

func TestNewJsonLogFormatterValidationNonBlank(t *testing.T) {
	exiter.(*stubExiter).callCount = 0
	buf := &bytes.Buffer{}
	log.SetOutput(buf)
	defer log.SetOutput(os.Stderr)

	newJSONLogFormatter(&types.AccessLog{
		CoreFields:                []string{" ", " : "},
		RequestHeaders:            []string{" ", " : "},
		OriginResponseHeaders:     []string{" ", " : "},
		DownstreamResponseHeaders: []string{" ", " : "},
	})

	assert.Equal(t, 1, exiter.(*stubExiter).callCount)

	logEntry := buf.String()

	strings.Index(logEntry, "level=")
	assert.True(t, strings.Contains(logEntry, "Unsupported access log fields: [ ]"), "%s", logEntry)
}

func TestNewJsonLogFormatterValidationNoDuplicates(t *testing.T) {
	exiter.(*stubExiter).callCount = 0
	buf := &bytes.Buffer{}
	log.SetOutput(buf)
	defer log.SetOutput(os.Stderr)

	newJSONLogFormatter(&types.AccessLog{
		CoreFields:                []string{"Duration:x", "Duration:y", "FrontendName:end", "BackendName:end"},
		RequestHeaders:            []string{"Host: http_host", "Host: http_host"},
		OriginResponseHeaders:     []string{"Server: server", "Server: server"},
		DownstreamResponseHeaders: []string{"Location: location", "Location: location"},
	})

	assert.Equal(t, 1, exiter.(*stubExiter).callCount)

	logEntry := buf.String()

	strings.Index(logEntry, "level=")
	assert.True(t, strings.Contains(logEntry, "Duplicate access log fields: Duration end"), "%s", logEntry)
	assert.True(t, strings.Contains(logEntry, "Duplicate access log fields: Host http_host"), "%s", logEntry)
	assert.True(t, strings.Contains(logEntry, "Duplicate access log fields: Server server"), "%s", logEntry)
	assert.True(t, strings.Contains(logEntry, "Duplicate access log fields: Location location"), "%s", logEntry)
}
