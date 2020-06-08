package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogs_KeepHeader(t *testing.T) {
	testCases := []struct {
		desc            string
		accessLogFields AccessLogFields
		header          string
		expected        string
	}{
		{
			desc:     "with default mode",
			header:   "X-Forwarded-For",
			expected: AccessLogDrop,
			accessLogFields: AccessLogFields{
				DefaultMode: AccessLogDrop,
				Headers: &FieldHeaders{
					DefaultMode: AccessLogDrop,
					Names:       map[string]string{},
				},
			},
		},
		{
			desc:     "with exact header name",
			header:   "X-Forwarded-For",
			expected: AccessLogKeep,
			accessLogFields: AccessLogFields{
				DefaultMode: AccessLogDrop,
				Headers: &FieldHeaders{
					DefaultMode: AccessLogDrop,
					Names: map[string]string{
						"X-Forwarded-For": AccessLogKeep,
					},
				},
			},
		},
		{
			desc:     "with case insensitive match on header name",
			header:   "X-Forwarded-For",
			expected: AccessLogKeep,
			accessLogFields: AccessLogFields{
				DefaultMode: AccessLogDrop,
				Headers: &FieldHeaders{
					DefaultMode: AccessLogDrop,
					Names: map[string]string{
						"x-forwarded-for": AccessLogKeep,
					},
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := test.accessLogFields.KeepHeader(test.header)
			assert.EqualValues(t, test.expected, result)
		})
	}
}
