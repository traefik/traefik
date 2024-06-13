package headermodifier

import (
	"context"
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
		config          dynamic.RequestHeaderModifier
		requestHeaders  http.Header
		expectedHeaders http.Header
	}{
		{
			desc:            "no config",
			config:          dynamic.RequestHeaderModifier{},
			expectedHeaders: map[string][]string{},
		},
		{
			desc: "set header",
			config: dynamic.RequestHeaderModifier{
				Set: map[string]string{"Foo": "Bar"},
			},
			expectedHeaders: map[string][]string{"Foo": {"Bar"}},
		},
		{
			desc: "set header with existing headers",
			config: dynamic.RequestHeaderModifier{
				Set: map[string]string{"Foo": "Bar"},
			},
			requestHeaders:  map[string][]string{"Foo": {"Baz"}, "Bar": {"Foo"}},
			expectedHeaders: map[string][]string{"Foo": {"Bar"}, "Bar": {"Foo"}},
		},
		{
			desc: "set multiple headers with existing headers",
			config: dynamic.RequestHeaderModifier{
				Set: map[string]string{"Foo": "Bar", "Bar": "Foo"},
			},
			requestHeaders:  map[string][]string{"Foo": {"Baz"}, "Bar": {"Foobar"}},
			expectedHeaders: map[string][]string{"Foo": {"Bar"}, "Bar": {"Foo"}},
		},
		{
			desc: "add header",
			config: dynamic.RequestHeaderModifier{
				Add: map[string]string{"Foo": "Bar"},
			},
			expectedHeaders: map[string][]string{"Foo": {"Bar"}},
		},
		{
			desc: "add header with existing headers",
			config: dynamic.RequestHeaderModifier{
				Add: map[string]string{"Foo": "Bar"},
			},
			requestHeaders:  map[string][]string{"Foo": {"Baz"}, "Bar": {"Foo"}},
			expectedHeaders: map[string][]string{"Foo": {"Baz", "Bar"}, "Bar": {"Foo"}},
		},
		{
			desc: "add multiple headers with existing headers",
			config: dynamic.RequestHeaderModifier{
				Add: map[string]string{"Foo": "Bar", "Bar": "Foo"},
			},
			requestHeaders:  map[string][]string{"Foo": {"Baz"}, "Bar": {"Foobar"}},
			expectedHeaders: map[string][]string{"Foo": {"Baz", "Bar"}, "Bar": {"Foobar", "Foo"}},
		},
		{
			desc: "remove header",
			config: dynamic.RequestHeaderModifier{
				Remove: []string{"Foo"},
			},
			expectedHeaders: map[string][]string{},
		},
		{
			desc: "remove header with existing headers",
			config: dynamic.RequestHeaderModifier{
				Remove: []string{"Foo"},
			},
			requestHeaders:  map[string][]string{"Foo": {"Baz"}, "Bar": {"Foo"}},
			expectedHeaders: map[string][]string{"Bar": {"Foo"}},
		},
		{
			desc: "remove multiple headers with existing headers",
			config: dynamic.RequestHeaderModifier{
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

			handler := NewRequestHeaderModifier(context.Background(), next, test.config, "foo-request-header-modifier")

			req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost", nil)
			if test.requestHeaders != nil {
				req.Header = test.requestHeaders
			}

			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			assert.Equal(t, test.expectedHeaders, gotHeaders)
		})
	}
}
