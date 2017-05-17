package accesslog

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func fixtureHeaders(kv ...string) http.Header {
	if len(kv)%2 != 0 {
		panic("Odd number of input parameters to fixtureHeaders")
	}
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
			RequestMethod:         http.MethodGet,
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
