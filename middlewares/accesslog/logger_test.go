package accesslog

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/containous/traefik/testhelpers"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
)

type logtestResponseWriter struct{}

const (
	textMessage = `So shaken as we are, so wan with care,
Find we a time for frighted peace to pant
And breathe short-winded accents of new broils
To be commenced in strands afar remote.
No more the thirsty entrance of this soil
Shall daub her lips with her own children’s blood.
Nor more shall trenching war channel her fields,
Nor bruise her flow’rets with the armed hoofs
Of hostile paces: those opposed eyes,
Which, like the meteors of a troubled heaven,
All of one nature, of one substance bred,
Did lately meet in the intestine shock
And furious close of civil butchery
Shall now, in mutual well-beseeming ranks,
March all one way and be no more opposed
Against acquaintance, kindred and allies:`

	testFrontendName = "frontend"
	testBackendName  = "backend"
	testUsername     = "-"
	testReferrer     = "http://example.com/x/y/z"

	// target URL
	testTargetHost  = "test.host.name"
	testTargetPort  = "8181"
	testTarget      = testTargetHost + ":" + testTargetPort
	testTargetPath  = "/a/b/c?q=1#z1"
	testTargetURL   = "http://" + testTarget + testTargetPath
	testBackendAddr = "10.1.2.3:8001"
	testBackendURL  = "http://" + testBackendAddr

	// User agent
	testRemoteHost = "190.190.190.190"
	testRemotePort = "20121"
	testRemoteAddr = testRemoteHost + ":" + testRemotePort
	testUserAgent  = "user-agent-very-very-long-string"
)

var (
	logfileDir       string
	parsedBackendURL *url.URL
	allCoreKeySlice  []string
)

func init() {
	logfileDir = os.TempDir()

	parsedBackendURL = testhelpers.MustParseURL(testBackendURL)

	allCoreKeySlice = defaultCoreKeys[:]
	for k := range allCoreKeys {
		allCoreKeySlice = append(allCoreKeySlice, k)
	}
	sort.StringSlice(allCoreKeySlice).Sort()
}

func logfilePath(suffix string) string {
	return filepath.Join(logfileDir, "traefikTestLogger.log"+suffix)
}

// newRequest constructs a request in a form required by RFC7230 section 5,
// which typically means supplying a Host header after a request line that
// does not include the host:port.
//
// (Note that httptest.NewRequest has subtley different behaviour)
func newRequest(method, path string) *http.Request {
	u, err := url.Parse(path)
	if err != nil {
		panic("Unable to parse path into URL")
	}
	r := &http.Request{
		Method:     method,
		URL:        u,
		Proto:      "HTTP/1.1",
		ProtoMinor: 1,
		Close:      false,
		Host:       testTarget,
		Header: map[string][]string{
			"User-Agent": {testUserAgent},
			"Referer":    {testReferrer},
		},
		RemoteAddr: testRemoteAddr,
	}
	return r
}

//-------------------------------------------------------------------------------------------------

func TestBlankFile(t *testing.T) {
	backend := NewSaveBackend(http.HandlerFunc(simpleHandlerFunc), testBackendName)
	frontend := NewSaveFrontend(backend, testFrontendName)
	settings := &types.AccessLog{}
	la, errs := NewLogAppender(settings)
	assert.Nil(t, errs)
	logger := NewLogHandler(la)

	defer logger.Close()

	r := newRequest("POST", testTargetPath)
	rec := httptest.NewRecorder()

	logger.ServeHTTP(rec, r, frontend.ServeHTTP)

	err := la.Close()
	assert.Nil(t, err, "%v", err)
}

func TestDataCaptureWithBackend(t *testing.T) {
	backend := NewSaveBackend(http.HandlerFunc(simpleHandlerFunc), testBackendName)
	frontend := NewSaveFrontend(backend, testFrontendName)
	settings := &types.AccessLog{CoreFields: allCoreKeySlice}
	buf := &bytes.Buffer{}
	lf := &captureLogFormatter{}
	la := LogAppender{settings: settings, formatter: lf, file: buf, buf: buf}
	logger := NewLogHandler(la)

	defer logger.Close()

	r := newRequest("POST", testTargetPath)
	r.Body = ioutil.NopCloser(bytes.NewBufferString("[1,2,3]"))
	rec := httptest.NewRecorder()

	logger.ServeHTTP(rec, r, swapURLHandler(frontend))

	now := time.Now().UTC()
	assert.Equal(t, 1, len(lf.events))
	e0 := lf.events[0]

	// checks on time-sensitive fields come first
	assert.True(t, e0.Core[StartUTC].(time.Time).Before(now))
	assert.True(t, e0.Core[OriginDuration].(time.Duration) >= 0)
	assert.True(t, e0.Core[Duration].(time.Duration) >= e0.Core[OriginDuration].(time.Duration))
	assert.True(t, e0.Core[Overhead].(time.Duration) >= 0)

	e0.Core[StartUTC] = now
	e0.Core[StartLocal] = now
	e0.Core[Duration] = 0
	e0.Core[OriginDuration] = 0
	e0.Core[Overhead] = 0

	assertCoreEqual(t, CoreLogData{
		StartUTC:              now,
		StartLocal:            now,
		Duration:              0,
		FrontendName:          testFrontendName,
		BackendName:           testBackendName,
		BackendURL:            parsedBackendURL,
		BackendAddr:           testBackendAddr,
		ClientHost:            testRemoteHost,
		ClientPort:            testRemotePort,
		ClientUsername:        testUsername,
		ClientAddr:            testRemoteAddr,
		OriginDuration:        0,
		OriginContentSize:     int64(len(textMessage)),
		RequestAddr:           testTarget,
		RequestHost:           testTargetHost,
		RequestPort:           testTargetPort,
		RequestMethod:         "POST",
		RequestPath:           testTargetPath,
		RequestProtocol:       "HTTP/1.1",
		RequestLine:           "POST " + testTargetPath + " HTTP/1.1",
		RequestContentSize:    int64(7),
		OriginStatus:          200,
		OriginStatusLine:      "200 OK",
		DownstreamStatus:      200,
		DownstreamStatusLine:  "200 OK",
		DownstreamContentSize: int64(len(textMessage)),
		RequestCount:          uint64(1),
		Overhead:              0,
		// n.b. GzipRatio=1 so should be elided.
	}, e0.Core)
}

func TestDataCaptureWithBackendAndGzip(t *testing.T) {
	backend := NewSaveBackend(http.HandlerFunc(simpleHandlerFunc), testBackendName)
	frontend := NewSaveFrontend(backend, testFrontendName)
	settings := &types.AccessLog{CoreFields: allCoreKeySlice}
	buf := &bytes.Buffer{}
	lf := &captureLogFormatter{}
	la := LogAppender{settings: settings, formatter: lf, file: buf, buf: buf}
	logger := NewLogHandler(la)

	defer logger.Close()

	r := newRequest("POST", testTargetPath)
	r.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	logger.ServeHTTP(rec, r, swapURLHandler(gziphandler.GzipHandler(frontend)))

	now := time.Now().UTC()
	assert.Equal(t, 1, len(lf.events))
	e0 := lf.events[0]

	// checks on time-sensitive fields come first
	assert.True(t, e0.Core[StartUTC].(time.Time).Before(now))
	assert.True(t, e0.Core[OriginDuration].(time.Duration) >= 0)
	assert.True(t, e0.Core[Duration].(time.Duration) >= e0.Core[OriginDuration].(time.Duration))
	assert.True(t, e0.Core[Overhead].(time.Duration) >= 0)

	e0.Core[StartUTC] = now
	e0.Core[StartLocal] = now
	e0.Core[Duration] = 0
	e0.Core[OriginDuration] = 0
	e0.Core[Overhead] = 0

	assertCoreEqual(t, CoreLogData{
		StartUTC:              now,
		StartLocal:            now,
		Duration:              0,
		FrontendName:          testFrontendName,
		BackendName:           testBackendName,
		BackendURL:            parsedBackendURL,
		BackendAddr:           testBackendAddr,
		ClientHost:            testRemoteHost,
		ClientPort:            testRemotePort,
		ClientUsername:        testUsername,
		ClientAddr:            testRemoteAddr,
		OriginDuration:        0,
		OriginContentSize:     int64(len(textMessage)),
		RequestAddr:           testTarget,
		RequestHost:           testTargetHost,
		RequestPort:           testTargetPort,
		RequestMethod:         "POST",
		RequestPath:           testTargetPath,
		RequestProtocol:       "HTTP/1.1",
		RequestLine:           "POST " + testTargetPath + " HTTP/1.1",
		OriginStatus:          200,
		OriginStatusLine:      "200 OK",
		DownstreamStatus:      200,
		DownstreamStatusLine:  "200 OK",
		DownstreamContentSize: int64(434),
		RequestCount:          uint64(2),
		GzipRatio:             float64(1.5806451612903225),
		Overhead:              0,
	}, e0.Core)
}

func TestDataCaptureWithRedirect(t *testing.T) {
	frontend := NewSaveFrontend(http.HandlerFunc(redirHandlerFunc), testFrontendName)
	settings := &types.AccessLog{CoreFields: allCoreKeySlice}
	buf := &bytes.Buffer{}
	lf := &captureLogFormatter{}
	la := LogAppender{settings: settings, formatter: lf, file: buf, buf: buf}
	logger := NewLogHandler(la)

	defer logger.Close()

	r := newRequest("POST", testTargetPath)
	rec := httptest.NewRecorder()

	logger.ServeHTTP(rec, r, frontend.ServeHTTP)

	now := time.Now().UTC()
	assert.Equal(t, 1, len(lf.events))
	e0 := lf.events[0]

	// checks on time-sensitive fields come first
	assert.True(t, e0.Core[StartUTC].(time.Time).Before(now))
	assert.Nil(t, e0.Core[OriginDuration])
	assert.True(t, e0.Core[Duration].(time.Duration) > 0)
	assert.True(t, e0.Core[Overhead].(time.Duration) > 0)

	e0.Core[StartUTC] = now
	e0.Core[StartLocal] = now
	e0.Core[Duration] = 0
	e0.Core[Overhead] = 0

	assertCoreEqual(t, CoreLogData{
		StartUTC:              now,
		StartLocal:            now,
		Duration:              0,
		FrontendName:          testFrontendName,
		ClientHost:            testRemoteHost,
		ClientPort:            testRemotePort,
		ClientUsername:        testUsername,
		ClientAddr:            testRemoteAddr,
		RequestAddr:           testTarget,
		RequestHost:           testTargetHost,
		RequestPort:           testTargetPort,
		RequestMethod:         "POST",
		RequestPath:           testTargetPath,
		RequestProtocol:       "HTTP/1.1",
		RequestLine:           "POST " + testTargetPath + " HTTP/1.1",
		DownstreamStatus:      307,
		DownstreamStatusLine:  "307 Temporary Redirect",
		DownstreamContentSize: int64(0),
		RequestCount:          uint64(3),
		Overhead:              0,
	}, e0.Core)
}

func TestCLFLogger(t *testing.T) {
	file := logfilePath("")
	backend := NewSaveBackend(http.HandlerFunc(simpleHandlerFunc), testBackendName)
	frontend := NewSaveFrontend(backend, testFrontendName)
	settings := &types.AccessLog{File: file}
	la, errs := NewLogAppender(settings)
	assert.Len(t, errs, 0)
	logger := NewLogHandler(la)

	defer os.Remove(file)

	r := newRequest("POST", testTargetPath)
	rec := httptest.NewRecorder()

	logger.ServeHTTP(rec, r, swapURLHandler(frontend))

	err := logger.Close()
	assert.Nil(t, err, "%v", err)

	if logdata, err := os.Open(file); err != nil {
		assert.Fail(t, "Failed to read logfile", "%v", err)
	} else {
		scanner := bufio.NewScanner(logdata)
		assert.True(t, scanner.Scan())
		line := scanner.Text()
		tokens := strings.Split(line, " ")
		if assert.Equal(t, 16, len(tokens), line) {
			assert.Equal(t, testRemoteHost, tokens[0], line)
			assert.Equal(t, "-", tokens[2], line)
			assert.Equal(t, fmt.Sprintf("\"%s", r.Method), tokens[5], line)
			assert.Equal(t, fmt.Sprintf("%s", r.URL.String()), tokens[6], line)
			assert.Equal(t, fmt.Sprintf("%s\"", r.Proto), tokens[7], line)
			assert.Equal(t, fmt.Sprintf("%d", 200), tokens[8], line)
			assert.Equal(t, fmt.Sprintf("%d", len(textMessage)), tokens[9], line)
			assert.Equal(t, quoted(testReferrer), tokens[10], line)
			assert.Equal(t, quoted(testUserAgent), tokens[11], line)
			assert.Equal(t, "4", tokens[12], line)
			assert.Equal(t, quoted(testFrontendName), tokens[13], line)
			assert.Equal(t, quoted(testBackendURL), tokens[14], line)
		}
		assert.False(t, scanner.Scan())
	}
}

//-------------------------------------------------------------------------------------------------

func simpleHandlerFunc(rw http.ResponseWriter, r *http.Request) {
	// Simple way to force a read of the body
	// so that captureRequestReader count is incremented
	if r.Body != nil {
		io.Copy(ioutil.Discard, r.Body)
	}
	time.Sleep(2 * time.Millisecond)
	rw.Header().Set("Content-Length", strconv.Itoa(len(textMessage)))
	rw.Header().Set("Content-Type", "text/html")
	rw.Write([]byte(textMessage))
	rw.WriteHeader(200)
}

func redirHandlerFunc(rw http.ResponseWriter, r *http.Request) {
	time.Sleep(1 * time.Millisecond)
	rw.Header().Set("Location", testTargetPath)
	rw.WriteHeader(307)
}

func swapURLHandler(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL = parsedBackendURL
		next.ServeHTTP(w, r)
	})
}

//-------------------------------------------------------------------------------------------------

type captureLogFormatter struct {
	events []*LogData
	err    error
}

func (c *captureLogFormatter) Write(w io.Writer, event *LogData) error {
	c.events = append(c.events, event)
	if len(c.events) > 1 {
		panic(event)
	}
	return c.err
}

//-------------------------------------------------------------------------------------------------

func assertCoreEqual(t *testing.T, expected, actual CoreLogData) {
	if len(expected) < len(actual) {
		for k, a := range actual {
			e, ok := expected[k]
			if ok {
				assert.Equal(t, e, a, "for key:%s", k)
			} else {
				assert.Fail(t, "unexpected", "key:%s value:%v", k, a)
			}
		}
	} else {
		for k, e := range expected {
			a, ok := actual[k]
			if ok {
				assert.Equal(t, e, a, "for key:%s", k)
			} else {
				assert.Fail(t, "missing", "key:%s value:%v", k, e)
			}
		}
	}
}
