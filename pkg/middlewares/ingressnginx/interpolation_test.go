package ingressnginx

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ReplaceVariables(t *testing.T) {
	testCases := []struct {
		desc     string
		src      string
		req      *http.Request
		vars     map[string]string
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
			desc: "$hostname",
			src:  "val=$hostname",
			req:  httptest.NewRequest(http.MethodGet, "http://baz.com/foo/bar?key=value&other=test", http.NoBody),
			// $hostname returns the machine hostname (os.Hostname), not the HTTP Host header.
			expected: "val=" + mustHostname(t),
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
			expected: `val=q=test&test=1&test=2&token=token_1234&val=foo,bar,baz`,
		},
		{
			desc:     "$query_string",
			src:      "val=$query_string",
			req:      httptest.NewRequest(http.MethodGet, "http://baz.com?q=test&test=1&test=2&token=token_1234&val=foo,bar,baz", http.NoBody),
			expected: `val=q=test&test=1&test=2&token=token_1234&val=foo,bar,baz`,
		},
		{
			desc:     "$host && $escaped_request_uri",
			src:      "val=$host$escaped_request_uri",
			req:      httptest.NewRequest(http.MethodGet, "http://baz.com/foo/bar?key=value&other=test", http.NoBody),
			expected: `val=baz.com%2Ffoo%2Fbar%3Fkey%3Dvalue%26other%3Dtest`,
		},
		{
			desc:     "variable + text",
			src:      "val=${host}-text",
			req:      httptest.NewRequest(http.MethodGet, "http://baz.com/foo/bar?key=value&other=test", http.NoBody),
			expected: "val=baz.com-text",
		},
		{
			desc:     "variable + text (alternative)",
			src:      "val=$host-text",
			req:      httptest.NewRequest(http.MethodGet, "http://baz.com/foo/bar?key=value&other=test", http.NoBody),
			expected: "val=baz.com-text",
		},
		{
			desc:     "variable + text (alternative 2)",
			src:      "val=${host}test",
			req:      httptest.NewRequest(http.MethodGet, "http://baz.com/foo/bar?key=value&other=test", http.NoBody),
			expected: "val=baz.comtest",
		},
		{
			desc:     "non existing variable",
			src:      "val=$invalid",
			req:      httptest.NewRequest(http.MethodGet, "http://baz.com/foo/bar?key=value&other=test", http.NoBody),
			expected: "val=$invalid",
		},
		{
			desc:     "invalid variable format",
			src:      "val=${invalid-text",
			req:      httptest.NewRequest(http.MethodGet, "http://baz.com/foo/bar?key=value&other=test", http.NoBody),
			expected: "val=${invalid-text",
		},
		{
			desc:     "invalid variable format 2",
			src:      "val=$invalid}-text",
			req:      httptest.NewRequest(http.MethodGet, "http://baz.com/foo/bar?key=value&other=test", http.NoBody),
			expected: "val=$invalid}-text",
		},
		{
			desc:     "invalid variable format 3",
			src:      "val=$hosttext",
			req:      httptest.NewRequest(http.MethodGet, "http://baz.com/foo/bar?key=value&other=test", http.NoBody),
			expected: "val=$hosttext",
		},
		{
			desc:     "$scheme http",
			src:      "val=$scheme",
			req:      httptest.NewRequest(http.MethodGet, "http://baz.com/foo/bar", http.NoBody),
			expected: "val=http",
		},
		{
			desc:     "$scheme https",
			src:      "val=$scheme",
			req:      mustNewRequestWithTLS(t, http.MethodGet, "https://baz.com/foo/bar"),
			expected: "val=https",
		},
		{
			desc:     "$request_uri",
			src:      "val=$request_uri",
			req:      httptest.NewRequest(http.MethodGet, "http://baz.com/foo/bar?key=value&other=test", http.NoBody),
			expected: "val=/foo/bar?key=value&other=test",
		},
		{
			desc:     "$remote_addr",
			src:      "val=$remote_addr",
			req:      mustNewRequestWithRemoteAddr(t, http.MethodGet, "http://baz.com/foo/bar", "192.168.1.1:12345"),
			expected: "val=192.168.1.1",
		},
		{
			desc:     "$remote_addr without port",
			src:      "val=$remote_addr",
			req:      mustNewRequestWithRemoteAddr(t, http.MethodGet, "http://baz.com/foo/bar", "192.168.1.1"),
			expected: "val=192.168.1.1",
		},
		{
			desc:     "custom vars",
			src:      "val=$custom_var",
			req:      httptest.NewRequest(http.MethodGet, "http://baz.com/foo/bar", http.NoBody),
			vars:     map[string]string{"$custom_var": "custom_value"},
			expected: "val=custom_value",
		},
		{
			desc:     "$uri",
			src:      "val=$uri",
			req:      httptest.NewRequest(http.MethodGet, "http://baz.com/foo/bar?key=value", http.NoBody),
			expected: "val=/foo/bar",
		},
		{
			desc:     "$document_uri",
			src:      "val=$document_uri",
			req:      httptest.NewRequest(http.MethodGet, "http://baz.com/foo/bar?key=value", http.NoBody),
			expected: "val=/foo/bar",
		},
		{
			desc:     "$server_name with port",
			src:      "val=$server_name",
			req:      mustNewRequestWithHost(t, http.MethodGet, "http://baz.com:8080/foo", "baz.com:8080"),
			expected: "val=baz.com",
		},
		{
			desc:     "$server_name without port",
			src:      "val=$server_name",
			req:      httptest.NewRequest(http.MethodGet, "http://baz.com/foo", http.NoBody),
			expected: "val=baz.com",
		},
		{
			desc:     "$server_port with explicit port",
			src:      "val=$server_port",
			req:      mustNewRequestWithHost(t, http.MethodGet, "http://baz.com:8080/foo", "baz.com:8080"),
			expected: "val=8080",
		},
		{
			desc:     "$server_port without port http",
			src:      "val=$server_port",
			req:      httptest.NewRequest(http.MethodGet, "http://baz.com/foo", http.NoBody),
			expected: "val=80",
		},
		{
			desc:     "$server_port without port https",
			src:      "val=$server_port",
			req:      mustNewRequestWithTLS(t, http.MethodGet, "https://baz.com/foo"),
			expected: "val=443",
		},
		{
			desc: "$content_type",
			src:  "val=$content_type",
			req: mustNewRequestWithHeaders(t, http.MethodGet, "http://baz.com/foo", map[string][]string{
				"Content-Type": {"application/json"},
			}),
			expected: "val=application/json",
		},
		{
			desc: "$content_length",
			src:  "val=$content_length",
			req: mustNewRequestWithHeaders(t, http.MethodGet, "http://baz.com/foo", map[string][]string{
				"Content-Length": {"42"},
			}),
			expected: "val=42",
		},
		{
			desc:     "$cookie_session",
			src:      "val=$cookie_session",
			req:      mustNewRequestWithCookie(t, http.MethodGet, "http://baz.com/foo", "session", "abc123"),
			expected: "val=abc123",
		},
		{
			desc:     "$cookie_* not found",
			src:      "val=$cookie_missing",
			req:      httptest.NewRequest(http.MethodGet, "http://baz.com/foo", http.NoBody),
			expected: "val=",
		},
		{
			desc:     "$is_args with query string",
			src:      "val=$is_args",
			req:      httptest.NewRequest(http.MethodGet, "http://baz.com/foo?key=value", http.NoBody),
			expected: "val=?",
		},
		{
			desc:     "$is_args without query string",
			src:      "val=$is_args",
			req:      httptest.NewRequest(http.MethodGet, "http://baz.com/foo", http.NoBody),
			expected: "val=",
		},
		{
			desc:     "$host strips port",
			src:      "val=$host",
			req:      mustNewRequestWithHost(t, http.MethodGet, "http://baz.com:8080/foo", "baz.com:8080"),
			expected: "val=baz.com",
		},
		{
			desc:     "$host lowercased",
			src:      "val=$host",
			req:      mustNewRequestWithHost(t, http.MethodGet, "http://BAZ.COM/foo", "BAZ.COM"),
			expected: "val=baz.com",
		},
		{
			desc:     "$best_http_host preserves port",
			src:      "val=$best_http_host",
			req:      mustNewRequestWithHost(t, http.MethodGet, "http://baz.com:8080/foo", "baz.com:8080"),
			expected: "val=baz.com:8080",
		},
		{
			desc:     "$server_name with IPv6 and port",
			src:      "val=$server_name",
			req:      mustNewRequestWithHost(t, http.MethodGet, "http://[::1]:8080/foo", "[::1]:8080"),
			expected: "val=::1",
		},
		{
			desc:     "$server_port with IPv6 and port",
			src:      "val=$server_port",
			req:      mustNewRequestWithHost(t, http.MethodGet, "http://[::1]:8080/foo", "[::1]:8080"),
			expected: "val=8080",
		},
		{
			desc:     "$proxy_add_x_forwarded_for without existing header",
			src:      "val=$proxy_add_x_forwarded_for",
			req:      mustNewRequestWithRemoteAddr(t, http.MethodGet, "http://baz.com/foo", "192.168.1.1:12345"),
			expected: "val=192.168.1.1",
		},
		{
			desc: "$proxy_add_x_forwarded_for with existing header",
			src:  "val=$proxy_add_x_forwarded_for",
			req: func() *http.Request {
				r := mustNewRequestWithRemoteAddr(t, http.MethodGet, "http://baz.com/foo", "10.0.0.1:9999")
				r.Header.Set("X-Forwarded-For", "203.0.113.50, 70.41.3.18")
				return r
			}(),
			expected: "val=203.0.113.50, 70.41.3.18, 10.0.0.1",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			t.Parallel()

			got := ReplaceVariables(testCase.src, testCase.req, nil, testCase.vars)
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

func mustNewRequestWithTLS(t *testing.T, method, target string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(method, target, http.NoBody)
	req.TLS = &tls.ConnectionState{}

	return req
}

func mustNewRequestWithRemoteAddr(t *testing.T, method, target, remoteAddr string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(method, target, http.NoBody)
	req.RemoteAddr = remoteAddr

	return req
}

func mustNewRequestWithHost(t *testing.T, method, target, host string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(method, target, http.NoBody)
	req.Host = host

	return req
}

func mustNewRequestWithCookie(t *testing.T, method, target, cookieName, cookieValue string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(method, target, http.NoBody)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: cookieValue})

	return req
}

func mustHostname(t *testing.T) string {
	t.Helper()

	h, err := os.Hostname()
	require.NoError(t, err)

	return h
}
