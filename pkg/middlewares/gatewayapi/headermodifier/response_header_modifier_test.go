package headermodifier

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
)

func TestResponseHeaderModifier(t *testing.T) {
	testCases := []struct {
		desc            string
		config          dynamic.HeaderModifier
		responseHeaders http.Header
		expectedHeaders http.Header
	}{
		{
			desc:            "no config",
			config:          dynamic.HeaderModifier{},
			expectedHeaders: map[string][]string{},
		},
		{
			desc: "set header",
			config: dynamic.HeaderModifier{
				Set: map[string]string{"Foo": "Bar"},
			},
			expectedHeaders: map[string][]string{"Foo": {"Bar"}},
		},
		{
			desc: "set header with existing headers",
			config: dynamic.HeaderModifier{
				Set: map[string]string{"Foo": "Bar"},
			},
			responseHeaders: map[string][]string{"Foo": {"Baz"}, "Bar": {"Foo"}},
			expectedHeaders: map[string][]string{"Foo": {"Bar"}, "Bar": {"Foo"}},
		},
		{
			desc: "set multiple headers with existing headers",
			config: dynamic.HeaderModifier{
				Set: map[string]string{"Foo": "Bar", "Bar": "Foo"},
			},
			responseHeaders: map[string][]string{"Foo": {"Baz"}, "Bar": {"Foobar"}},
			expectedHeaders: map[string][]string{"Foo": {"Bar"}, "Bar": {"Foo"}},
		},
		{
			desc: "add header",
			config: dynamic.HeaderModifier{
				Add: map[string]string{"Foo": "Bar"},
			},
			expectedHeaders: map[string][]string{"Foo": {"Bar"}},
		},
		{
			desc: "add header with existing headers",
			config: dynamic.HeaderModifier{
				Add: map[string]string{"Foo": "Bar"},
			},
			responseHeaders: map[string][]string{"Foo": {"Baz"}, "Bar": {"Foo"}},
			expectedHeaders: map[string][]string{"Foo": {"Baz", "Bar"}, "Bar": {"Foo"}},
		},
		{
			desc: "add multiple headers with existing headers",
			config: dynamic.HeaderModifier{
				Add: map[string]string{"Foo": "Bar", "Bar": "Foo"},
			},
			responseHeaders: map[string][]string{"Foo": {"Baz"}, "Bar": {"Foobar"}},
			expectedHeaders: map[string][]string{"Foo": {"Baz", "Bar"}, "Bar": {"Foobar", "Foo"}},
		},
		{
			desc: "remove header",
			config: dynamic.HeaderModifier{
				Remove: []string{"Foo"},
			},
			expectedHeaders: map[string][]string{},
		},
		{
			desc: "remove header with existing headers",
			config: dynamic.HeaderModifier{
				Remove: []string{"Foo"},
			},
			responseHeaders: map[string][]string{"Foo": {"Baz"}, "Bar": {"Foo"}},
			expectedHeaders: map[string][]string{"Bar": {"Foo"}},
		},
		{
			desc: "remove multiple headers with existing headers",
			config: dynamic.HeaderModifier{
				Remove: []string{"Foo", "Bar"},
			},
			responseHeaders: map[string][]string{"Foo": {"Bar"}, "Bar": {"Foo"}, "Baz": {"Bar"}},
			expectedHeaders: map[string][]string{"Baz": {"Bar"}},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var nextCallCount int
			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				nextCallCount++
				rw.WriteHeader(http.StatusOK)
			})

			handler := NewResponseHeaderModifier(t.Context(), next, test.config, "foo-response-header-modifier")

			req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost", nil)
			resp := httptest.NewRecorder()
			for k, v := range test.responseHeaders {
				resp.Header()[k] = v
			}

			handler.ServeHTTP(resp, req)

			assert.Equal(t, 1, nextCallCount)
			assert.Equal(t, test.expectedHeaders, resp.Header())
		})
	}
}
