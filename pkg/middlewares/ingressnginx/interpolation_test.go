package ingressnginx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ReplaceVariables(t *testing.T) {
	testCases := []struct {
		desc     string
		src      string
		req      *http.Request
		expected string
	}{
		{
			desc:     "$host",
			src:      "val=$host",
			req:      httptest.NewRequest(http.MethodGet, "http://baz.com/foo/bar?key=value&other=test", http.NoBody),
			expected: `val=baz.com`,
		},
		{
			desc:     "$best_http_host",
			src:      "val=$best_http_host",
			req:      httptest.NewRequest(http.MethodGet, "http://baz.com/foo/bar?key=value&other=test", http.NoBody),
			expected: `val=baz.com`,
		},
		{
			desc:     "$hostname",
			src:      "val=$hostname",
			req:      httptest.NewRequest(http.MethodGet, "http://baz.com/foo/bar?key=value&other=test", http.NoBody),
			expected: `val=baz.com`,
		},
		{
			desc: "$http_*",
			src:  "val=$http_x_api_key",
			req: mustNewRequestWithHeaders(t, http.MethodGet, "http://baz.com/foo/bar?key=value&other=test", map[string][]string{
				"X-Api-Key": {"key"},
			}),
			expected: `val=key`,
		},
		{
			desc: "Multiple Header value in $http_*",
			src:  "val=$http_foo",
			req: mustNewRequestWithHeaders(t, http.MethodGet, "http://baz.com/foo/bar?key=value&other=test", map[string][]string{
				"Foo": {"bar", "baz"},
			}),
			expected: `val=bar,baz`,
		},
		{
			desc:     "$arg_*",
			src:      "val=$arg_token",
			req:      httptest.NewRequest(http.MethodGet, "http://baz.com?token=token_1234", http.NoBody),
			expected: `val=token_1234`,
		},
		{
			desc:     "$args",
			src:      "val=$args",
			req:      httptest.NewRequest(http.MethodGet, "http://baz.com?q=test&test=1&test=2&token=token_1234&val=foo,bar,baz", http.NoBody),
			expected: `val=q=test&test=1&test=2&token=token_1234&val=foo%2Cbar%2Cbaz`,
		},
		{
			desc:     "$query_string",
			src:      "val=$query_string",
			req:      httptest.NewRequest(http.MethodGet, "http://baz.com?q=test&test=1&test=2&token=token_1234&val=foo,bar,baz", http.NoBody),
			expected: `val=q=test&test=1&test=2&token=token_1234&val=foo%2Cbar%2Cbaz`,
		},
		{
			desc:     "$host && $escaped_request_uri",
			src:      "val=$host$escaped_request_uri",
			req:      httptest.NewRequest(http.MethodGet, "http://baz.com/foo/bar?key=value&other=test", http.NoBody),
			expected: `val=baz.com%2Ffoo%2Fbar%3Fkey%3Dvalue%26other%3Dtest`,
		},
		{
			desc:     "non existing variable",
			src:      "val=$invalid",
			req:      httptest.NewRequest(http.MethodGet, "http://baz.com/foo/bar?key=value&other=test", http.NoBody),
			expected: "val=$invalid",
		},
		{
			desc:     "invalid variable format",
			src:      "val=${invalid}",
			req:      httptest.NewRequest(http.MethodGet, "http://baz.com/foo/bar?key=value&other=test", http.NoBody),
			expected: "val=${invalid}",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			t.Parallel()

			got := ReplaceVariables(testCase.src, testCase.req)
			require.Equal(t, testCase.expected, got)
		})
	}
}

func mustNewRequestWithHeaders(t *testing.T, method, target string, headers map[string][]string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(method, target, http.NoBody)
	req.Header = headers

	return req
}
