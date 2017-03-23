package accesslog

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"bufio"
	"bytes"
	"github.com/NYTimes/gziphandler"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
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
Of hostile paces.`

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
	logFileCount     int64
	parsedBackendURL *url.URL
	allCoreKeySlice  []string
)

func init() {
	if runtime.GOOS == "windows" {
		logfileDir = os.Getenv("TEMP")
	} else {
		logfileDir = "/tmp"
	}
	parsedBackendURL, _ = url.Parse(testBackendURL)

	allCoreKeySlice = defaultCoreKeys
	for k := range allCoreKeys {
		allCoreKeySlice = append(allCoreKeySlice, k)
	}
	sort.StringSlice(allCoreKeySlice).Sort()
}

func logfilePath() string {
	// each test gets a unique file name, avoiding conflict should tests fail.
	n := atomic.AddInt64(&logFileCount, 1)
	name := fmt.Sprintf("traefikTestLogger%d.log", n)
	path := filepath.Join(logfileDir, name)
	os.Remove(path) // if present; ignore any error otherwise
	return path
}

// newRequest constructs a request in a form required by RFC7230 section 5,
// which typically means supplying a Host header after a request line that
// does not include the host:port.
//
// (Note that httptest.NewRequest has subtley different behaviour)
func newRequest(method, path string) *http.Request {
	u, _ := url.Parse(path)
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
	la, err := NewLogAppender(settings)
	assert.Nil(t, err)
	logger := NewLogHandler(la)

	defer logger.Close()

	r := newRequest("POST", testTargetPath)
	rec := httptest.NewRecorder()

	logger.ServeHTTP(rec, r, frontend.ServeHTTP)

	err = la.Close()
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
		DownstreamContentSize: int64(263),
		RequestCount:          uint64(2),
		GzipRatio:             float64(1.4334600760456273),
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
	file := logfilePath()
	backend := NewSaveBackend(http.HandlerFunc(simpleHandlerFunc), testBackendName)
	frontend := NewSaveFrontend(backend, testFrontendName)
	settings := &types.AccessLog{File: file}
	la, err := NewLogAppender(settings)
	assert.Nil(t, err)
	logger := NewLogHandler(la)

	defer os.Remove(file)

	r := newRequest("POST", testTargetPath)
	rec := httptest.NewRecorder()

	logger.ServeHTTP(rec, r, swapURLHandler(frontend))

	err = logger.Close()
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
	time.Sleep(time.Millisecond)
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

func BenchmarkCommonLogFormatToFile(b *testing.B) {
	logDataTable := fixtureLogDataTable(1234)
	benchmarks := []struct {
		name      string
		file      string
		gzipLevel int
		buff      string
		async     bool
		chansize  int
	}{
		{"plain unbuffered sync", "tmp.log", 0, "", false, 0},
		{"plain unbuffered async 0", "tmp.log", 0, "", true, 0},
		{"plain unbuffered async 100", "tmp.log", 0, "", true, 100},
		{"plain buffer=256B sync", "tmp.log", 0, "256B", false, 0},
		{"plain buffer=4KiB sync", "tmp.log", 0, "4KiB", false, 0},
		{"plain buffer=4KiB async 0", "tmp.log", 0, "4KiB", true, 0},
		{"plain buffer=4KiB async 100", "tmp.log", 0, "4KiB", true, 100},
		{"plain buffer=64KiB sync", "tmp.log", 0, "64KiB", false, 0},
		{"plain buffer=64KiB async 0", "tmp.log", 0, "64KiB", true, 0},
		{"plain buffer=64KiB async 100", "tmp.log", 0, "64KiB", true, 100},
		{"plain buffer=512KiB sync", "tmp.log", 0, "512KiB", false, 0},
		{"plain buffer=512KiB async 0", "tmp.log", 0, "512KiB", true, 0},
		{"plain buffer=512KiB async 100", "tmp.log", 0, "512KiB", true, 100},
		{"gzip -1 unbuffered sync", "tmp.log.gz", -1, "", false, 0},
		{"gzip 9 unbuffered sync", "tmp.log.gz", 9, "", false, 0},
		{"gzip 2 unbuffered sync", "tmp.log.gz", 2, "", false, 0},
		{"gzip 2 unbuffered async 0", "tmp.log.gz", 2, "", true, 0},
		{"gzip 2 unbuffered async 100", "tmp.log.gz", 2, "", true, 100},
	}
	for _, bm := range benchmarks {
		defer os.Remove(bm.file)
		b.Run(bm.name, func(b *testing.B) {
			cfg := &types.AccessLog{File: bm.file, GzipLevel: bm.gzipLevel, BufferSize: bm.buff, Async: bm.async, ChannelBuffer: bm.chansize}
			la, err := NewLogAppender(cfg)
			if err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				la.Write(logDataTable)
			}
			la.Close()
		})
	}
}

func BenchmarkCommonLogFormatToDevNull(b *testing.B) {
	logDataTable := fixtureLogDataTable(1234)
	file := "/dev/null"

	la, err := NewLogAppender(&types.AccessLog{File: file})
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		la.Write(logDataTable)
	}
	la.Close()
}

func BenchmarkCommonLogFormatToDiscard(b *testing.B) {
	logDataTable := fixtureLogDataTable(1234)
	settings := &types.AccessLog{}

	la := LogAppender{settings: settings, formatter: commonLogFormatter{}, file: ioutil.Discard, buf: ioutil.Discard}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		la.Write(logDataTable)
	}
	la.Close()
}

func BenchmarkNoJsonToFile(b *testing.B) {
	logDataTable := fixtureLogDataTable(1234)
	file := "tmp.log"
	defer os.Remove(file)

	jlf := jsonLogFormatter{}
	la, err := NewLogAppender(&types.AccessLog{File: file, Format: "json"})
	if err != nil {
		b.Fatal(err)
	}
	la.formatter = jlf // empty settings are required
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		la.Write(logDataTable)
	}
	la.Close()
}

func BenchmarkNoJsonToDiscard(b *testing.B) {
	logDataTable := fixtureLogDataTable(1234)

	settings := &types.AccessLog{}
	jlf := jsonLogFormatter{}
	la := LogAppender{settings: settings, formatter: jlf, file: ioutil.Discard, buf: ioutil.Discard}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		la.Write(logDataTable)
	}
	la.Close()
}

func BenchmarkAllJsonToDiscard(b *testing.B) {
	logDataTable := fixtureLogDataTable(1234)

	cfg := &types.AccessLog{
		RequestHeaders:            []string{"Host", "Accept", "Referer", "User-Agent"},
		OriginResponseHeaders:     []string{"Content-Type", "Content-Length", "Server"},
		DownstreamResponseHeaders: []string{"Content-Type", "Content-Length", "Server", "Location"},
	}
	settings := &types.AccessLog{}
	jlf := newJSONLogFormatter(cfg)
	la := LogAppender{settings: settings, formatter: jlf, file: ioutil.Discard, buf: ioutil.Discard}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		la.Write(logDataTable)
	}
	la.Close()
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
