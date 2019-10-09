package log

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLog(t *testing.T) {
	testCases := []struct {
		desc     string
		fields   map[string]string
		expected string
	}{
		{
			desc: "Log with one field",
			fields: map[string]string{
				"foo": "bar",
			},
			expected: ` level=error msg="message test" foo=bar$`,
		},
		{
			desc: "Log with two fields",
			fields: map[string]string{
				"foo": "bar",
				"oof": "rab",
			},
			expected: ` level=error msg="message test" foo=bar oof=rab$`,
		},
		{
			desc:     "Log without field",
			fields:   map[string]string{},
			expected: ` level=error msg="message test"$`,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			var buffer bytes.Buffer
			SetOutput(&buffer)

			ctx := context.Background()

			for key, value := range test.fields {
				ctx = With(ctx, Str(key, value))
			}

			FromContext(ctx).Error("message test")

			assert.Regexp(t, test.expected, strings.TrimSpace(buffer.String()))
		})
	}
}
