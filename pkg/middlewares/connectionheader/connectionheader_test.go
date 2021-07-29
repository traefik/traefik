package connectionheader

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemove(t *testing.T) {
	testCases := []struct {
		desc       string
		reqHeaders map[string]string
		expected   http.Header
	}{
		{
			desc: "simple remove",
			reqHeaders: map[string]string{
				"Foo":        "bar",
				"Connection": "foo",
			},
			expected: http.Header{},
		},
		{
			desc: "remove and Upgrade",
			reqHeaders: map[string]string{
				"Upgrade":    "test",
				"Foo":        "bar",
				"Connection": "Upgrade,foo",
			},
			expected: http.Header{
				"Upgrade":    []string{"test"},
				"Connection": []string{"Upgrade"},
			},
		},
		{
			desc: "no remove",
			reqHeaders: map[string]string{
				"Foo":        "bar",
				"Connection": "fii",
			},
			expected: http.Header{
				"Foo": []string{"bar"},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

			h := Remove(next)

			req := httptest.NewRequest(http.MethodGet, "https://localhost", nil)

			for k, v := range test.reqHeaders {
				req.Header.Set(k, v)
			}

			rw := httptest.NewRecorder()

			h.ServeHTTP(rw, req)

			assert.Equal(t, test.expected, req.Header)
		})
	}
}
