package accesslog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAccessLog(t *testing.T) {
	testCases := []struct {
		desc     string
		value    string
		expected map[string]string
	}{
		{
			desc:  "full log",
			value: `TestHost - TestUser [13/Apr/2016:07:14:19 -0700] "POST testpath HTTP/0.0" 123 12 "testReferer" "testUserAgent" 1 "testFrontend" "http://127.0.0.1/testBackend" 1ms`,
			expected: map[string]string{
				ClientHost:             "TestHost",
				ClientUsername:         "TestUser",
				StartUTC:               "13/Apr/2016:07:14:19 -0700",
				RequestMethod:          "POST",
				RequestPath:            "testpath",
				RequestProtocol:        "HTTP/0.0",
				OriginStatus:           "123",
				OriginContentSize:      "12",
				RequestRefererHeader:   `"testReferer"`,
				RequestUserAgentHeader: `"testUserAgent"`,
				RequestCount:           "1",
				FrontendName:           `"testFrontend"`,
				BackendURL:             `"http://127.0.0.1/testBackend"`,
				Duration:               "1ms",
			},
		},
		{
			desc:  "log with space",
			value: `127.0.0.1 - - [09/Mar/2018:10:51:32 +0000] "GET / HTTP/1.1" 401 17 "-" "Go-http-client/1.1" 1 "testFrontend with space" - 0ms`,
			expected: map[string]string{
				ClientHost:             "127.0.0.1",
				ClientUsername:         "-",
				StartUTC:               "09/Mar/2018:10:51:32 +0000",
				RequestMethod:          "GET",
				RequestPath:            "/",
				RequestProtocol:        "HTTP/1.1",
				OriginStatus:           "401",
				OriginContentSize:      "17",
				RequestRefererHeader:   `"-"`,
				RequestUserAgentHeader: `"Go-http-client/1.1"`,
				RequestCount:           "1",
				FrontendName:           `"testFrontend with space"`,
				BackendURL:             `-`,
				Duration:               "0ms",
			},
		},
		{
			desc:     "bad log",
			value:    `bad`,
			expected: map[string]string{},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result, err := ParseAccessLog(test.value)
			assert.NoError(t, err)
			assert.Equal(t, len(test.expected), len(result))
			for key, value := range test.expected {
				assert.Equal(t, value, result[key])
			}
		})
	}
}
