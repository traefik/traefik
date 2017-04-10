package middlewares

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/containous/traefik/middlewares/accesslog"
	shellwords "github.com/mattn/go-shellwords"
	"github.com/stretchr/testify/assert"
)

type logtestResponseWriter struct{}

var (
	logger                  *Logger
	logfileName             = "traefikTestLogger.log"
	logfilePath             string
	helloWorld              = "Hello, World"
	testBackendName         = "http://127.0.0.1/testBackend"
	testFrontendName        = "testFrontend"
	testStatus              = 123
	testHostname            = "TestHost"
	testUsername            = "TestUser"
	testPath                = "http://testpath"
	testPort                = 8181
	testProto               = "HTTP/0.0"
	testMethod              = "POST"
	testReferer             = "testReferer"
	testUserAgent           = "testUserAgent"
	testBackend2FrontendMap = map[string]string{
		testBackendName: testFrontendName,
	}
	printedLogdata bool
)

func TestLogger(t *testing.T) {
	if runtime.GOOS == "windows" {
		logfilePath = filepath.Join(os.Getenv("TEMP"), logfileName)
	} else {
		logfilePath = filepath.Join("/tmp", logfileName)
	}

	logger = NewLogger(logfilePath)
	defer cleanup()
	SetBackend2FrontendMap(&testBackend2FrontendMap)

	r := &http.Request{
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
	}

	// Temporary - until new access logger is fully implemented
	// create the data table and populate frontend and backend
	core := make(accesslog.CoreLogData)
	logDataTable := &accesslog.LogData{Core: core, Request: r.Header}
	logDataTable.Core[accesslog.FrontendName] = testFrontendName
	logDataTable.Core[accesslog.BackendURL] = testBackendName
	req := r.WithContext(context.WithValue(r.Context(), accesslog.DataTableKey, logDataTable))

	logger.ServeHTTP(&logtestResponseWriter{}, req, LogWriterTestHandlerFunc)

	if logdata, err := ioutil.ReadFile(logfilePath); err != nil {
		fmt.Printf("%s\n%s\n", string(logdata), err.Error())
		assert.Nil(t, err)
	} else if tokens, err := shellwords.Parse(string(logdata)); err != nil {
		fmt.Printf("%s\n", err.Error())
		assert.Nil(t, err)
	} else if assert.Equal(t, 14, len(tokens), printLogdata(logdata)) {
		assert.Equal(t, testHostname, tokens[0], printLogdata(logdata))
		assert.Equal(t, testUsername, tokens[2], printLogdata(logdata))
		assert.Equal(t, fmt.Sprintf("%s %s %s", testMethod, testPath, testProto), tokens[5], printLogdata(logdata))
		assert.Equal(t, fmt.Sprintf("%d", testStatus), tokens[6], printLogdata(logdata))
		assert.Equal(t, fmt.Sprintf("%d", len(helloWorld)), tokens[7], printLogdata(logdata))
		assert.Equal(t, testReferer, tokens[8], printLogdata(logdata))
		assert.Equal(t, testUserAgent, tokens[9], printLogdata(logdata))
		assert.Equal(t, "1", tokens[10], printLogdata(logdata))
		assert.Equal(t, testFrontendName, tokens[11], printLogdata(logdata))
		assert.Equal(t, testBackendName, tokens[12], printLogdata(logdata))
	}
}

func cleanup() {
	logger.Close()
	os.Remove(logfilePath)
}

func printLogdata(logdata []byte) string {
	return fmt.Sprintf(
		"\nExpected: %s\n"+
			"Actual:   %s",
		"TestHost - TestUser [13/Apr/2016:07:14:19 -0700] \"POST http://testpath HTTP/0.0\" 123 12 \"testReferer\" \"testUserAgent\" 1 \"testFrontend\" \"http://127.0.0.1/testBackend\" 1ms",
		string(logdata))
}

func LogWriterTestHandlerFunc(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte(helloWorld))
	rw.WriteHeader(testStatus)
	saveBackendNameForLogger(r, testBackendName)
}

func (lrw *logtestResponseWriter) Header() http.Header {
	return map[string][]string{}
}

func (lrw *logtestResponseWriter) Write(b []byte) (int, error) {
	return len(b), nil
}

func (lrw *logtestResponseWriter) WriteHeader(s int) {
}
