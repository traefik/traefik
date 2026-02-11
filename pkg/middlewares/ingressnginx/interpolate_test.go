package ingressnginx

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReplaceNginxVariables(t *testing.T) {
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
			desc: "$http_x_api_key",
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
			desc: "$http_foo",
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
			desc: "$arg_token",
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
			desc: "$arg_test",
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

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result := ReplaceNginxVariables(tc.src, tc.req)
			require.Equal(t, tc.expected, result)
		})
	}
}
