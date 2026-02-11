package middlewares

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReplaceNginxVariables(t *testing.T) {
	testCases := []struct {
		desc     string
		src      string
		req      *http.Request
		expected string
	}{
		{
			desc: "$host && $escaped_request_uri",
			src:  "http://bar.foo.com/external-auth/start?rd=https://$host$escaped_request_uri",
			req: &http.Request{
				Method:     http.MethodGet,
				Host:       "baz.com",
				RequestURI: "/foo/bar?key=value&other=test",
			},
			expected: `http://bar.foo.com/external-auth/start?rd=https://baz.com%2Ffoo%2Fbar%3Fkey%3Dvalue%26other%3Dtest`,
		},
		{
			desc: "$host, $scheme, $request_uri",
			src:  "$scheme://bar.foo.com/external-auth/start?rd=$scheme://$host$request_uri",
			req: &http.Request{
				URL:    &url.URL{Scheme: "http", Path: "/foo/bar", RawQuery: "key=value&other=test"},
				Method: http.MethodGet,
				Host:   "baz.com",
			},
			expected: `http://bar.foo.com/external-auth/start?rd=http://baz.com/foo/bar?key=value&other=test`,
		},
		{
			desc: "invalid variable",
			src:  "https://bar.foo.com/external-auth/start?rd=$invalid://$foo$bar",
			req: &http.Request{
				URL:        &url.URL{Scheme: "http"},
				Method:     http.MethodGet,
				Host:       "baz.com",
				RequestURI: "/foo/bar?key=value&other=test",
			},
			expected: "https://bar.foo.com/external-auth/start?rd=$invalid://$foo$bar",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := ReplaceNginxVariables(tc.src, tc.req)
			require.Equal(t, tc.expected, result)
		})
	}
}
