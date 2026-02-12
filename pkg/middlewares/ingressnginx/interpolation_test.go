package ingressnginx

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ReplaceVariables(t *testing.T) {
	reqURL, err := url.Parse("http://baz.com/auth?q=test&val=foo,bar,baz&token=token_1234&test=1&test=2")
	require.NoError(t, err)

	testCases := []struct {
		desc     string
		src      string
		req      *http.Request
		expected string
	}{
		{
			desc: "$host",
			src:  "http://bar.foo.com/external-auth/start?rd=https://$host",
			req: &http.Request{
				Method:     http.MethodGet,
				Host:       "baz.com",
				RequestURI: "/foo/bar?key=value&other=test",
			},
			expected: `http://bar.foo.com/external-auth/start?rd=https://baz.com`,
		},
		{
			desc: "$best_http_host",
			src:  "http://bar.foo.com/external-auth/start?rd=https://$best_http_host",
			req: &http.Request{
				Method:     http.MethodGet,
				Host:       "baz.com",
				RequestURI: "/foo/bar?key=value&other=test",
			},
			expected: `http://bar.foo.com/external-auth/start?rd=https://baz.com`,
		},
		{
			desc: "$hostname",
			src:  "http://bar.foo.com/external-auth/start?rd=https://$hostname",
			req: &http.Request{
				Method:     http.MethodGet,
				Host:       "baz.com",
				RequestURI: "/foo/bar?key=value&other=test",
			},
			expected: `http://bar.foo.com/external-auth/start?rd=https://baz.com`,
		},
		{
			desc: "$http_host",
			src:  "http://bar.foo.com/external-auth/start?rd=https://$http_host",
			req: &http.Request{
				Method:     http.MethodGet,
				Host:       "baz.com",
				RequestURI: "/foo/bar?key=value&other=test",
				Header: map[string][]string{
					"Host": {"foo.com"},
				},
			},
			expected: `http://bar.foo.com/external-auth/start?rd=https://foo.com`,
		},
		{
			desc: "Single Header value in $http_x_api_key",
			src:  "http://bar.foo.com/external-auth/start?rd=https://baz.com/?api=$http_x_api_key",
			req: &http.Request{
				Method:     http.MethodGet,
				Host:       "baz.com",
				RequestURI: "/foo/bar?key=value&other=test",
				Header: map[string][]string{
					"X-Api-Key": {"key"},
				},
			},
			expected: `http://bar.foo.com/external-auth/start?rd=https://baz.com/?api=key`,
		},
		{
			desc: "Multiple Header value in $http_foo",
			src:  "$http_foo",
			req: &http.Request{
				Method:     http.MethodGet,
				Host:       "baz.com",
				RequestURI: "/foo/bar?key=value&other=test",
				Header: map[string][]string{
					"Foo": {"bar", "baz"},
				},
			},
			expected: `bar,baz`,
		},
		{
			desc: "Single arg value in $arg_token",
			src:  "$arg_token",
			req: &http.Request{
				URL:    reqURL,
				Method: http.MethodGet,
				Host:   "baz.com",
			},
			expected: `token_1234`,
		},
		{
			desc: "$arg_val",
			src:  "$arg_val",
			req: &http.Request{
				URL:    reqURL,
				Method: http.MethodGet,
				Host:   "baz.com",
			},
			expected: `foo,bar,baz`,
		},
		{
			desc: "Multiple arg value in $arg_test",
			src:  "$arg_test",
			req: &http.Request{
				URL:    reqURL,
				Method: http.MethodGet,
				Host:   "baz.com",
			},
			expected: `1`,
		},
		{
			desc: "$args",
			src:  "http://bar.foo.com/external-auth/start?rd=https://baz.com/?$args",
			req: &http.Request{
				URL:    reqURL,
				Method: http.MethodGet,
				Host:   "baz.com",
			},
			expected: `http://bar.foo.com/external-auth/start?rd=https://baz.com/?q=test&test=1&test=2&token=token_1234&val=foo%2Cbar%2Cbaz`,
		},
		{
			desc: "$query_string",
			src:  "http://bar.foo.com/external-auth/start?rd=https://baz.com/?$query_string",
			req: &http.Request{
				URL:    reqURL,
				Method: http.MethodGet,
				Host:   "baz.com",
			},
			expected: `http://bar.foo.com/external-auth/start?rd=https://baz.com/?q=test&test=1&test=2&token=token_1234&val=foo%2Cbar%2Cbaz`,
		},
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
		{
			desc: "invalid variable format",
			src:  "https://bar.foo.com/external-auth/start?rd=${invalid}://${foo}${bar}",
			req: &http.Request{
				URL:        &url.URL{Scheme: "http"},
				Method:     http.MethodGet,
				Host:       "baz.com",
				RequestURI: "/foo/bar?key=value&other=test",
			},
			expected: "https://bar.foo.com/external-auth/start?rd=${invalid}://${foo}${bar}",
		},
		{
			desc: "wrong variable", // TODO: should we provide an error ?
			src:  "https://bar.foo.com/external-auth/start?rd=http://$hostpath/foo",
			req: &http.Request{
				URL:        &url.URL{Scheme: "http"},
				Method:     http.MethodGet,
				Host:       "baz.com",
				RequestURI: "/foo/bar?key=value&other=test",
			},
			expected: "https://bar.foo.com/external-auth/start?rd=http://$hostpath/foo",
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
