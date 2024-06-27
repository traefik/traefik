package tracing

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_safeFullURL(t *testing.T) {
	testCases := []struct {
		desc                  string
		unRedactedQueryParams []string
		originalURL           *url.URL
		expectedURL           *url.URL
	}{
		{
			desc:        "Nil URL",
			originalURL: nil,
			expectedURL: nil,
		},
		{
			desc:        "No query parameters",
			originalURL: &url.URL{Scheme: "https", Host: "example.com"},
			expectedURL: &url.URL{Scheme: "https", Host: "example.com"},
		},
		{
			desc:        "All query parameters redacted",
			originalURL: &url.URL{Scheme: "https", Host: "example.com", RawQuery: "foo=bar&baz=qux"},
			expectedURL: &url.URL{Scheme: "https", Host: "example.com", RawQuery: "baz=REDACTED&foo=REDACTED"},
		},
		{
			desc:                  "Some query parameters unredacted",
			unRedactedQueryParams: []string{"foo"},
			originalURL:           &url.URL{Scheme: "https", Host: "example.com", RawQuery: "foo=bar&baz=qux"},
			expectedURL:           &url.URL{Scheme: "https", Host: "example.com", RawQuery: "baz=REDACTED&foo=bar"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			tr := NewTracer(nil, nil, nil, test.unRedactedQueryParams)

			gotURL := tr.safeURL(test.originalURL)

			assert.Equal(t, test.expectedURL, gotURL)
		})
	}
}
