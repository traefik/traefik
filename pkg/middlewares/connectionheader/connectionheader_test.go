package connectionheader

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemover(t *testing.T) {
	testCases := []struct {
		desc       string
		reqHeaders map[string]string
		expected   http.Header
	}{
		{
			desc: "simple remove",
			reqHeaders: map[string]string{
				"Foo":            "bar",
				connectionHeader: "foo",
			},
			expected: http.Header{},
		},
		{
			desc: "remove and Upgrade",
			reqHeaders: map[string]string{
				upgradeHeader:    "test",
				"Foo":            "bar",
				connectionHeader: "Upgrade,foo",
			},
			expected: http.Header{
				upgradeHeader:    []string{"test"},
				connectionHeader: []string{"Upgrade"},
			},
		},
		{
			desc: "no remove",
			reqHeaders: map[string]string{
				"Foo":            "bar",
				connectionHeader: "fii",
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

			h := Remover(next)

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
