package accesslog

import (
	"net/http"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCommonLogFormatter_Format(t *testing.T) {
	clf := CommonLogFormatter{}

	testCases := []struct {
		name        string
		data        map[string]interface{}
		expectedLog string
	}{
		{
			name: "OriginStatus & OriginContentSize are nil",
			data: map[string]interface{}{
				StartUTC:             time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
				Duration:             123 * time.Second,
				ClientHost:           "10.0.0.1",
				ClientUsername:       "Client",
				RequestMethod:        http.MethodGet,
				RequestPath:          "/foo",
				RequestProtocol:      "http",
				OriginStatus:         nil,
				OriginContentSize:    nil,
				"request_Referer":    "",
				"request_User-Agent": "",
				RequestCount:         0,
				FrontendName:         "",
				BackendURL:           "",
			},
			expectedLog: `10.0.0.1 - Client [10/Nov/2009:23:00:00 +0000] "GET /foo http" - - - - 0 - - 123000ms
`,
		},
		{
			name: "all data",
			data: map[string]interface{}{
				StartUTC:             time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
				Duration:             123 * time.Second,
				ClientHost:           "10.0.0.1",
				ClientUsername:       "Client",
				RequestMethod:        http.MethodGet,
				RequestPath:          "/foo",
				RequestProtocol:      "http",
				OriginStatus:         123,
				OriginContentSize:    132,
				"request_Referer":    "referer",
				"request_User-Agent": "agent",
				RequestCount:         nil,
				FrontendName:         "foo",
				BackendURL:           "http://10.0.0.2/toto",
			},
			expectedLog: `10.0.0.1 - Client [10/Nov/2009:23:00:00 +0000] "GET /foo http" 123 132 "referer" "agent" - "foo" "http://10.0.0.2/toto" 123000ms
`,
		},
	}

	for _, test := range testCases {
		test := test
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
		name        string
		value       interface{}
		expectedLog interface{}
	}{
		{
			name:        "",
			value:       1,
			expectedLog: 1,
		},
		{
			name:        "",
			value:       "foo",
			expectedLog: `"foo"`,
		},
		{
			name:        "",
			value:       nil,
			expectedLog: "-",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			lg := toLog(test.value)

			assert.Equal(t, test.expectedLog, lg)
		})
	}
}
