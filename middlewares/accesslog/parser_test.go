package accesslog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseLog(t *testing.T) {
	data := `TestHost - TestUser [13/Apr/2016:07:14:19 -0700] "POST testpath HTTP/0.0" 123 12 "testReferer" "testUserAgent" 1 "testFrontend" "http://127.0.0.1/testBackend" 1ms`

	result, err := ParseAccessLog(data)
	assert.NoError(t, err)
	assert.Equal(t, "TestHost", result[ClientHost])
	assert.Equal(t, "TestUser", result[ClientUsername])
	assert.Equal(t, "13/Apr/2016:07:14:19 -0700", result[StartUTC])
	assert.Equal(t, "POST", result[RequestMethod])
	assert.Equal(t, "testpath", result[RequestPath])
	assert.Equal(t, "HTTP/0.0", result[RequestProtocol])
	assert.Equal(t, "123", result[OriginStatus])
	assert.Equal(t, "12", result[OriginContentSize])
	assert.Equal(t, `"testReferer"`, result["request_Referer"])
	assert.Equal(t, `"testUserAgent"`, result["request_User-Agent"])
	assert.Equal(t, "1", result[RequestCount])
	assert.Equal(t, `"testFrontend"`, result[FrontendName])
	assert.Equal(t, `"http://127.0.0.1/testBackend"`, result[BackendURL])
	assert.Equal(t, "1ms", result[Duration])
}

func TestParseLogWithSpace(t *testing.T) {
	data := `127.0.0.1 - - [09/Mar/2018:10:51:32 +0000] "GET / HTTP/1.1" 401 17 "-" "Go-http-client/1.1" 1 "testFrontend with space" - 0ms`

	result, err := ParseAccessLog(data)
	assert.NoError(t, err)
	assert.Equal(t, "127.0.0.1", result[ClientHost])
	assert.Equal(t, "-", result[ClientUsername])
	assert.Equal(t, "09/Mar/2018:10:51:32 +0000", result[StartUTC])
	assert.Equal(t, "GET", result[RequestMethod])
	assert.Equal(t, "/", result[RequestPath])
	assert.Equal(t, "HTTP/1.1", result[RequestProtocol])
	assert.Equal(t, "401", result[OriginStatus])
	assert.Equal(t, "17", result[OriginContentSize])
	assert.Equal(t, `"-"`, result["request_Referer"])
	assert.Equal(t, `"Go-http-client/1.1"`, result["request_User-Agent"])
	assert.Equal(t, "1", result[RequestCount])
	assert.Equal(t, `"testFrontend with space"`, result[FrontendName])
	assert.Equal(t, `-`, result[BackendURL])
	assert.Equal(t, "0ms", result[Duration])
}

func TestParseLogError(t *testing.T) {
	data := `bad`

	result, err := ParseAccessLog(data)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
}
