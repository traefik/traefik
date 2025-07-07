package headermodifier

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
)

func TestRequestHeaderModifier(t *testing.T) {
	testCases := []struct {
		desc            string
		config          dynamic.HeaderModifier
		requestHeaders  http.Header
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
			requestHeaders:  map[string][]string{"Foo": {"Baz"}, "Bar": {"Foo"}},
			expectedHeaders: map[string][]string{"Foo": {"Bar"}, "Bar": {"Foo"}},
		},
		{
			desc: "set multiple headers with existing headers",
			config: dynamic.HeaderModifier{
				Set: map[string]string{"Foo": "Bar", "Bar": "Foo"},
			},
			requestHeaders:  map[string][]string{"Foo": {"Baz"}, "Bar": {"Foobar"}},
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
			requestHeaders:  map[string][]string{"Foo": {"Baz"}, "Bar": {"Foo"}},
			expectedHeaders: map[string][]string{"Foo": {"Baz", "Bar"}, "Bar": {"Foo"}},
		},
		{
			desc: "add multiple headers with existing headers",
			config: dynamic.HeaderModifier{
				Add: map[string]string{"Foo": "Bar", "Bar": "Foo"},
			},
			requestHeaders:  map[string][]string{"Foo": {"Baz"}, "Bar": {"Foobar"}},
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
			requestHeaders:  map[string][]string{"Foo": {"Baz"}, "Bar": {"Foo"}},
			expectedHeaders: map[string][]string{"Bar": {"Foo"}},
		},
		{
			desc: "remove multiple headers with existing headers",
			config: dynamic.HeaderModifier{
				Remove: []string{"Foo", "Bar"},
			},
			requestHeaders:  map[string][]string{"Foo": {"Bar"}, "Bar": {"Foo"}, "Baz": {"Bar"}},
			expectedHeaders: map[string][]string{"Baz": {"Bar"}},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var gotHeaders http.Header
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotHeaders = r.Header
			})

			handler := NewRequestHeaderModifier(t.Context(), next, test.config, "foo-request-header-modifier")

			req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost", nil)
			for h, v := range test.requestHeaders {
				req.Header[h] = v
			}
			resp := httptest.NewRecorder()

			handler.ServeHTTP(resp, req)

			assert.Equal(t, test.expectedHeaders, gotHeaders)
		})
	}
}
