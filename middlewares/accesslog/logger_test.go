package accesslog

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/containous/traefik/types"
	shellwords "github.com/mattn/go-shellwords"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	logFileNameSuffix       = "/traefik/logger/test.log"
	testContent             = "Hello, World"
	testBackendName         = "http://127.0.0.1/testBackend"
	testFrontendName        = "testFrontend"
	testStatus              = 123
	testContentSize   int64 = 12
	testHostname            = "TestHost"
	testUsername            = "TestUser"
	testPath                = "testpath"
	testPort                = 8181
	testProto               = "HTTP/0.0"
	testMethod              = "POST"
	testReferer             = "testReferer"
	testUserAgent           = "testUserAgent"
	testRetryAttempts       = 2
)

func TestLoggerCLF(t *testing.T) {
	tmpDir := createTempDir(t, CommonFormat)
	defer os.RemoveAll(tmpDir)

	logFilePath := filepath.Join(tmpDir, logFileNameSuffix)
	config := &types.AccessLog{FilePath: logFilePath, Format: CommonFormat}
	doLogging(t, config)

	logData, err := ioutil.ReadFile(logFilePath)
	require.NoError(t, err)

	assertValidLogData(t, logData)
}

func TestLoggerJSON(t *testing.T) {
	tmpDir := createTempDir(t, JSONFormat)
	defer os.RemoveAll(tmpDir)

	logFilePath := filepath.Join(tmpDir, logFileNameSuffix)
	config := &types.AccessLog{FilePath: logFilePath, Format: JSONFormat}
	doLogging(t, config)

	logData, err := ioutil.ReadFile(logFilePath)
	require.NoError(t, err)

	jsonData := make(map[string]interface{})
	err = json.Unmarshal(logData, &jsonData)
	require.NoError(t, err)

	expectedKeys := []string{
		RequestHost,
		RequestAddr,
		RequestMethod,
		RequestPath,
		RequestProtocol,
		RequestPort,
		RequestLine,
		DownstreamStatus,
		DownstreamStatusLine,
		DownstreamContentSize,
		OriginContentSize,
		OriginStatus,
		"request_Referer",
		"request_User-Agent",
		FrontendName,
		BackendURL,
		ClientUsername,
		ClientHost,
		ClientPort,
		ClientAddr,
		"level",
		"msg",
		"downstream_Content-Type",
		RequestCount,
		Duration,
		Overhead,
		RetryAttempts,
		"time",
		"StartLocal",
		"StartUTC",
	}
	containsKeys(t, expectedKeys, jsonData)

	var assertCount int
	assert.Equal(t, testHostname, jsonData[RequestHost])
	assertCount++
	assert.Equal(t, testHostname, jsonData[RequestAddr])
	assertCount++
	assert.Equal(t, testMethod, jsonData[RequestMethod])
	assertCount++
	assert.Equal(t, testPath, jsonData[RequestPath])
	assertCount++
	assert.Equal(t, testProto, jsonData[RequestProtocol])
	assertCount++
	assert.Equal(t, "-", jsonData[RequestPort])
	assertCount++
	assert.Equal(t, fmt.Sprintf("%s %s %s", testMethod, testPath, testProto), jsonData[RequestLine])
	assertCount++
	assert.Equal(t, float64(testStatus), jsonData[DownstreamStatus])
	assertCount++
	assert.Equal(t, fmt.Sprintf("%d ", testStatus), jsonData[DownstreamStatusLine])
	assertCount++
	assert.Equal(t, float64(len(testContent)), jsonData[DownstreamContentSize])
	assertCount++
	assert.Equal(t, float64(len(testContent)), jsonData[OriginContentSize])
	assertCount++
	assert.Equal(t, float64(testStatus), jsonData[OriginStatus])
	assertCount++
	assert.Equal(t, testReferer, jsonData["request_Referer"])
	assertCount++
	assert.Equal(t, testUserAgent, jsonData["request_User-Agent"])
	assertCount++
	assert.Equal(t, testFrontendName, jsonData[FrontendName])
	assertCount++
	assert.Equal(t, testBackendName, jsonData[BackendURL])
	assertCount++
	assert.Equal(t, testUsername, jsonData[ClientUsername])
	assertCount++
	assert.Equal(t, testHostname, jsonData[ClientHost])
	assertCount++
	assert.Equal(t, fmt.Sprintf("%d", testPort), jsonData[ClientPort])
	assertCount++
	assert.Equal(t, fmt.Sprintf("%s:%d", testHostname, testPort), jsonData[ClientAddr])
	assertCount++
	assert.Equal(t, "info", jsonData["level"])
	assertCount++
	assert.Equal(t, "", jsonData["msg"])
	assertCount++
	assert.Equal(t, "text/plain; charset=utf-8", jsonData["downstream_Content-Type"].(string))
	assertCount++
	assert.NotZero(t, jsonData[RequestCount].(float64))
	assertCount++
	assert.NotZero(t, jsonData[Duration].(float64))
	assertCount++
	assert.NotZero(t, jsonData[Overhead].(float64))
	assertCount++
	assert.Equal(t, float64(testRetryAttempts), jsonData[RetryAttempts].(float64))
	assertCount++
	assert.NotEqual(t, "", jsonData["time"].(string))
	assertCount++
	assert.NotEqual(t, "", jsonData["StartLocal"].(string))
	assertCount++
	assert.NotEqual(t, "", jsonData["StartUTC"].(string))
	assertCount++

	assert.Equal(t, len(jsonData), assertCount, string(logData))
}

func TestNewLogHandlerOutputStdout(t *testing.T) {
	file, restoreStdout := captureStdout(t)
	defer restoreStdout()

	config := &types.AccessLog{FilePath: "", Format: CommonFormat}
	doLogging(t, config)

	written, err := ioutil.ReadFile(file.Name())
	require.NoError(t, err, "unable to read captured stdout from file")
	require.NotZero(t, len(written), "expected access log message on stdout")
	assertValidLogData(t, written)
}

func assertValidLogData(t *testing.T, logData []byte) {
	tokens, err := shellwords.Parse(string(logData))
	require.NoError(t, err)

	formatErrMessage := fmt.Sprintf(`
		Expected: TestHost - TestUser [13/Apr/2016:07:14:19 -0700] "POST testpath HTTP/0.0" 123 12 "testReferer" "testUserAgent" 1 "testFrontend" "http://127.0.0.1/testBackend" 1ms
		Actual:   %s
		`,
		string(logData))
	require.Equal(t, 14, len(tokens), formatErrMessage)
	assert.Equal(t, testHostname, tokens[0], formatErrMessage)
	assert.Equal(t, testUsername, tokens[2], formatErrMessage)
	assert.Equal(t, fmt.Sprintf("%s %s %s", testMethod, testPath, testProto), tokens[5], formatErrMessage)
	assert.Equal(t, fmt.Sprintf("%d", testStatus), tokens[6], formatErrMessage)
	assert.Equal(t, fmt.Sprintf("%d", len(testContent)), tokens[7], formatErrMessage)
	assert.Equal(t, testReferer, tokens[8], formatErrMessage)
	assert.Equal(t, testUserAgent, tokens[9], formatErrMessage)
	assert.Regexp(t, regexp.MustCompile("[0-9]*"), tokens[10], formatErrMessage)
	assert.Equal(t, testFrontendName, tokens[11], formatErrMessage)
	assert.Equal(t, testBackendName, tokens[12], formatErrMessage)
}

func captureStdout(t *testing.T) (out *os.File, restoreStdout func()) {
	file, err := ioutil.TempFile("", "testlogger")
	require.NoError(t, err, "failed to create temp file")

	original := os.Stdout
	os.Stdout = file

	restoreStdout = func() {
		os.Stdout = original
	}

	return file, restoreStdout
}

func createTempDir(t *testing.T, prefix string) string {
	tmpDir, err := ioutil.TempDir("", prefix)
	require.NoError(t, err, "failed to create temp dir")

	return tmpDir
}

func doLogging(t *testing.T, config *types.AccessLog) {
	logger, err := NewLogHandler(config)
	require.NoError(t, err)
	defer logger.Close()

	if config.FilePath != "" {
		_, err = os.Stat(config.FilePath)
		require.NoError(t, err, fmt.Sprintf("logger should create %s", config.FilePath))
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
	}

	logger.ServeHTTP(httptest.NewRecorder(), req, logWriterTestHandlerFunc)
}

func containsKeys(t *testing.T, expectedKeys []string, data map[string]interface{}) {
	for key, value := range data {
		if !contains(expectedKeys, key) {
			t.Errorf("Unexpected log key: %s [value: %s]", key, value)
		}
	}
	for _, k := range expectedKeys {
		if _, ok := data[k]; !ok {
			t.Errorf("the expected key '%s' is not present in the map. %+v", k, data)
		}
	}
}

func contains(values []string, value string) bool {
	for _, v := range values {
		if value == v {
			return true
		}
	}
	return false
}

func logWriterTestHandlerFunc(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte(testContent))
	rw.WriteHeader(testStatus)

	logDataTable := GetLogDataTable(r)
	logDataTable.Core[FrontendName] = testFrontendName
	logDataTable.Core[BackendURL] = testBackendName
	logDataTable.Core[OriginStatus] = testStatus
	logDataTable.Core[OriginContentSize] = testContentSize
	logDataTable.Core[RetryAttempts] = testRetryAttempts
}
