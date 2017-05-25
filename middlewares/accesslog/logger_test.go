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
	"testing"

	"github.com/containous/traefik/types"
	shellwords "github.com/mattn/go-shellwords"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	logger            *LogHandler
	logFileNameSuffix       = "/traefik/logger/test.log"
	helloWorld              = "Hello, World"
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
)

func TestLoggerCLF(t *testing.T) {
	tmpDir, logFilePath := doLogging(t, CommonFormat)
	defer os.RemoveAll(tmpDir)

	logData, err := ioutil.ReadFile(logFilePath)
	require.NoError(t, err)

	tokens, err := shellwords.Parse(string(logData))
	require.NoError(t, err)

	assert.Equal(t, 14, len(tokens), printLogData(logData))
	assert.Equal(t, testHostname, tokens[0], printLogData(logData))
	assert.Equal(t, testUsername, tokens[2], printLogData(logData))
	assert.Equal(t, fmt.Sprintf("%s %s %s", testMethod, testPath, testProto), tokens[5], printLogData(logData))
	assert.Equal(t, fmt.Sprintf("%d", testStatus), tokens[6], printLogData(logData))
	assert.Equal(t, fmt.Sprintf("%d", len(helloWorld)), tokens[7], printLogData(logData))
	assert.Equal(t, testReferer, tokens[8], printLogData(logData))
	assert.Equal(t, testUserAgent, tokens[9], printLogData(logData))
	assert.Equal(t, "1", tokens[10], printLogData(logData))
	assert.Equal(t, testFrontendName, tokens[11], printLogData(logData))
	assert.Equal(t, testBackendName, tokens[12], printLogData(logData))
}

func TestLoggerJSON(t *testing.T) {
	tmpDir, logFilePath := doLogging(t, JSONFormat)
	defer os.RemoveAll(tmpDir)

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
	assert.Equal(t, float64(len(helloWorld)), jsonData[DownstreamContentSize])
	assertCount++
	assert.Equal(t, float64(len(helloWorld)), jsonData[OriginContentSize])
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
	assert.NotEqual(t, "", jsonData["time"].(string))
	assertCount++
	assert.NotEqual(t, "", jsonData["StartLocal"].(string))
	assertCount++
	assert.NotEqual(t, "", jsonData["StartUTC"].(string))
	assertCount++

	assert.Equal(t, len(jsonData), assertCount, string(logData))
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

func doLogging(t *testing.T, format string) (string, string) {
	tmp, err := ioutil.TempDir("", format)
	if err != nil {
		t.Fatalf("failed to create temp dir: %s", err)
	}

	logFilePath := filepath.Join(tmp, logFileNameSuffix)

	config := types.AccessLog{FilePath: logFilePath, Format: format}

	logger, err = NewLogHandler(&config)
	defer logger.Close()
	require.NoError(t, err)

	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		t.Fatalf("logger should create %s", logFilePath)
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

	rw := httptest.NewRecorder()
	logger.ServeHTTP(rw, req, logWriterTestHandlerFunc)
	return tmp, logFilePath
}

func printLogData(logdata []byte) string {
	return fmt.Sprintf(`
	Expected: TestHost - TestUser [13/Apr/2016:07:14:19 -0700] "POST testpath HTTP/0.0" 123 12 "testReferer" "testUserAgent" 1 "testFrontend" "http://127.0.0.1/testBackend" 1ms
	Actual:   %s
	`,
		string(logdata))
}

func logWriterTestHandlerFunc(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte(helloWorld))
	rw.WriteHeader(testStatus)

	logDataTable := GetLogDataTable(r)
	logDataTable.Core[FrontendName] = testFrontendName
	logDataTable.Core[BackendURL] = testBackendName
	logDataTable.Core[OriginStatus] = testStatus
	logDataTable.Core[OriginContentSize] = testContentSize
}
