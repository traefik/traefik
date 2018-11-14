package accesslog

import (
	"net/http"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
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
				StartUTC:               time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
				Duration:               123 * time.Second,
				ClientHost:             "10.0.0.1",
				ClientUsername:         "Client",
				RequestMethod:          http.MethodGet,
				RequestPath:            "/foo",
				RequestProtocol:        "http",
				OriginStatus:           nil,
				OriginContentSize:      nil,
				RequestRefererHeader:   "",
				RequestUserAgentHeader: "",
				RequestCount:           0,
				FrontendName:           "",
				BackendURL:             "",
			},
			expectedLog: `10.0.0.1 - Client [10/Nov/2009:23:00:00 +0000] "GET /foo http" - - "-" "-" 0 - - 123000ms
`,
		},
		{
			name: "all data",
			data: map[string]interface{}{
				StartUTC:               time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
				Duration:               123 * time.Second,
				ClientHost:             "10.0.0.1",
				ClientUsername:         "Client",
				RequestMethod:          http.MethodGet,
				RequestPath:            "/foo",
				RequestProtocol:        "http",
				OriginStatus:           123,
				OriginContentSize:      132,
				RequestRefererHeader:   "referer",
				RequestUserAgentHeader: "agent",
				RequestCount:           nil,
				FrontendName:           "foo",
				BackendURL:             "http://10.0.0.2/toto",
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
		desc         string
		fields       logrus.Fields
		fieldName    string
		defaultValue string
		quoted       bool
		expectedLog  interface{}
	}{
		{
			desc: "Should return int 1",
			fields: logrus.Fields{
				"Powpow": 1,
			},
			fieldName:    "Powpow",
			defaultValue: defaultValue,
			quoted:       false,
			expectedLog:  1,
		},
		{
			desc: "Should return string foo",
			fields: logrus.Fields{
				"Powpow": "foo",
			},
			fieldName:    "Powpow",
			defaultValue: defaultValue,
			quoted:       true,
			expectedLog:  `"foo"`,
		},
		{
			desc: "Should return defaultValue if fieldName does not exist",
			fields: logrus.Fields{
				"Powpow": "foo",
			},
			fieldName:    "",
			defaultValue: defaultValue,
			quoted:       false,
			expectedLog:  "-",
		},
		{
			desc:         "Should return defaultValue if fields is nil",
			fields:       nil,
			fieldName:    "",
			defaultValue: defaultValue,
			quoted:       false,
			expectedLog:  "-",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			lg := toLog(test.fields, test.fieldName, defaultValue, test.quoted)

			assert.Equal(t, test.expectedLog, lg)
		})
	}
}
