package snippet

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func Test_New(t *testing.T) {
	testCases := []struct {
		desc        string
		config      dynamic.Snippet
		expectError bool
	}{
		{
			desc:        "fails when both snippets are empty",
			config:      dynamic.Snippet{},
			expectError: true,
		},
		{
			desc: "succeeds with valid server snippet",
			config: dynamic.Snippet{
				ServerSnippet: `add_header X-Test "value";`,
			},
			expectError: false,
		},
		{
			desc: "succeeds with valid always server snippet",
			config: dynamic.Snippet{
				ServerSnippet: `add_header X-Test "value" always;`,
			},
			expectError: false,
		},
		{
			desc: "succeeds with valid configuration snippet",
			config: dynamic.Snippet{
				ConfigurationSnippet: `add_header X-Test "value";`,
			},
			expectError: false,
		},
		{
			desc: "succeeds with both snippets",
			config: dynamic.Snippet{
				ServerSnippet:        `add_header X-Server "server";`,
				ConfigurationSnippet: `add_header X-Config "config";`,
			},
			expectError: false,
		},
		{
			desc: "fails with invalid server snippet syntax",
			config: dynamic.Snippet{
				ServerSnippet: `add_header X-Test`,
			},
			expectError: true,
		},
		{
			desc: "fails with invalid server snippet syntax",
			config: dynamic.Snippet{
				ConfigurationSnippet: `add_header X-Test`,
			},
			expectError: true,
		},
		{
			desc: "fails with unknown directive in server snippet",
			config: dynamic.Snippet{
				ServerSnippet: `unknown_directive value;`,
			},
			expectError: true,
		},
		{
			desc: "fails on context when proxy_set_headers in if in server snippet",
			config: dynamic.Snippet{
				ServerSnippet: `if ( $request_method = "GET") {
  proxy_set_header X-Test "value";
}`,
			},
			expectError: true,
		},
		{
			desc: "fails on context when proxy_set_headers in if in configuration snippet",
			config: dynamic.Snippet{
				ConfigurationSnippet: `if ( $request_method = "GET") {
  proxy_set_header X-Test "value";
}`,
			},
			expectError: true,
		},
		{
			desc: "fails on context when add_header in if in server snippet",
			config: dynamic.Snippet{
				ServerSnippet: `if ( $request_method = "GET") {
  add_header X-Test "value";
}`,
			},
			expectError: true,
		},
		{
			desc: "valid context when add_headers in if in configuration snippet",
			config: dynamic.Snippet{
				ConfigurationSnippet: `if ( $request_method = "GET") {
  add_header X-Test "value";
}`,
			},
			expectError: false,
		},

		{
			desc: "succeeds with valid more_clear_headers",
			config: dynamic.Snippet{
				ConfigurationSnippet: `more_clear_headers "X-Test";`,
			},
			expectError: false,
		},
		{
			desc: "succeeds with valid more_clear_input_headers",
			config: dynamic.Snippet{
				ConfigurationSnippet: `more_clear_input_headers "X-Test";`,
			},
			expectError: false,
		},
		{
			desc: "fails with unknown directive in configuration snippet",
			config: dynamic.Snippet{
				ConfigurationSnippet: `unknown_directive value;`,
			},
			expectError: true,
		},
		{
			desc: "succeeds with valid rewrite in server snippet",
			config: dynamic.Snippet{
				ServerSnippet: `rewrite ^/old$ /new last;`,
			},
			expectError: false,
		},
		{
			desc: "succeeds with valid rewrite in configuration snippet",
			config: dynamic.Snippet{
				ConfigurationSnippet: `rewrite ^/old$ /new break;`,
			},
			expectError: false,
		},
		{
			desc: "succeeds with rewrite without flag",
			config: dynamic.Snippet{
				ConfigurationSnippet: `rewrite ^/old$ /new;`,
			},
			expectError: false,
		},
		{
			desc: "fails with rewrite with invalid regex",
			config: dynamic.Snippet{
				ConfigurationSnippet: `rewrite ^/old[$ /new last;`,
			},
			expectError: true,
		},
		{
			desc: "fails with rewrite with invalid flag",
			config: dynamic.Snippet{
				ConfigurationSnippet: `rewrite ^/old$ /new invalid;`,
			},
			expectError: true,
		},
		{
			desc: "fails with rewrite with missing parameters",
			config: dynamic.Snippet{
				ConfigurationSnippet: `rewrite ^/old$;`,
			},
			expectError: true,
		},
		{
			desc: "succeeds with allow directive",
			config: dynamic.Snippet{
				ConfigurationSnippet: `allow 10.0.0.0/8;`,
			},
			expectError: false,
		},
		{
			desc: "succeeds with deny directive",
			config: dynamic.Snippet{
				ConfigurationSnippet: `deny all;`,
			},
			expectError: false,
		},
		{
			desc: "fails with deny with invalid IP",
			config: dynamic.Snippet{
				ConfigurationSnippet: `deny invalid-ip;`,
			},
			expectError: true,
		},
		{
			desc: "succeeds with proxy_hide_header directive",
			config: dynamic.Snippet{
				ServerSnippet: `proxy_hide_header X-Powered-By;`,
			},
			expectError: false,
		},
		{
			desc: "succeeds with expires directive",
			config: dynamic.Snippet{
				ConfigurationSnippet: `expires 24h;`,
			},
			expectError: false,
		},
		{
			desc: "succeeds with expires epoch",
			config: dynamic.Snippet{
				ConfigurationSnippet: `expires epoch;`,
			},
			expectError: false,
		},
		{
			desc: "fails with expires with invalid duration",
			config: dynamic.Snippet{
				ConfigurationSnippet: `expires invalid;`,
			},
			expectError: true,
		},
		{
			desc: "fails with -s flag on more_set_input_headers",
			config: dynamic.Snippet{
				ConfigurationSnippet: `more_set_input_headers -s "200" "X-Custom: value";`,
			},
			expectError: true,
		},
		{
			desc: "fails with -s flag on more_clear_input_headers",
			config: dynamic.Snippet{
				ConfigurationSnippet: `more_clear_input_headers -s "200" "X-Custom";`,
			},
			expectError: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			_, err := New(t.Context(), next, &test.config, "test-snippet")
			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_Directives(t *testing.T) {
	testCases := []struct {
		desc                      string
		serverSnippet             string
		configurationSnippet      string
		method                    string
		path                      string
		remoteAddr                string
		requestHeaders            map[string]string
		expectedResponseHeaders   map[string]string
		unexpectedResponseHeaders []string
		expectedRequestHeaders    map[string]string
		expectedStatusCode        int
		expectedBody              string
		expectedPath              string
		expectedQuery             string
		expectedRedirectURL       string
	}{
		{
			desc:                    "add_header server snippet adds simple header",
			serverSnippet:           `add_header X-Custom "custom-value";`,
			expectedResponseHeaders: map[string]string{"X-Custom": "custom-value"},
		},
		{
			desc:                    "add_header server snippet adds header without quotes",
			serverSnippet:           `add_header X-Simple simple;`,
			expectedResponseHeaders: map[string]string{"X-Simple": "simple"},
		},
		{
			desc:                    "add_header configuration snippet adds simple header",
			configurationSnippet:    `add_header X-Custom "custom-value";`,
			expectedResponseHeaders: map[string]string{"X-Custom": "custom-value"},
		},
		{
			desc:                    "add_header configuration snippet adds header without quotes",
			configurationSnippet:    `add_header X-Simple simple;`,
			expectedResponseHeaders: map[string]string{"X-Simple": "simple"},
		},
		{
			desc:                 "add_header configuration snippet overrides server snippet",
			serverSnippet:        `add_header X-Server server-value;`,
			configurationSnippet: `add_header X-Config config-value;`,
			expectedResponseHeaders: map[string]string{
				"X-Config": "config-value",
			},
			unexpectedResponseHeaders: []string{"X-Server"},
		},
		{
			desc:                    "more_set_headers server snippet sets header",
			serverSnippet:           `more_set_headers "X-Custom:custom-value";`,
			expectedResponseHeaders: map[string]string{"X-Custom": "custom-value"},
		},
		{
			desc:                    "more_set_headers configuration snippet sets header",
			configurationSnippet:    `more_set_headers "X-Custom:custom-value";`,
			expectedResponseHeaders: map[string]string{"X-Custom": "custom-value"},
		},
		{
			desc:                 "more_set_headers both snippets set headers",
			serverSnippet:        `more_set_headers "X-Server:server-value";`,
			configurationSnippet: `more_set_headers "X-Config:config-value";`,
			expectedResponseHeaders: map[string]string{
				"X-Server": "server-value",
				"X-Config": "config-value",
			},
		},
		{
			desc:                 "more_set_headers both snippets override same header",
			serverSnippet:        `more_set_headers "X-Header:server-value";`,
			configurationSnippet: `more_set_headers "X-Header:config-value";`,
			expectedResponseHeaders: map[string]string{
				"X-Header": "config-value",
			},
		},
		{
			desc:                 "more_set_headers with spaces",
			configurationSnippet: `more_set_headers "X-Header: config-value ";`,
			expectedResponseHeaders: map[string]string{
				"X-Header": "config-value",
			},
		},
		{
			desc:                 "more_set_headers with multiple headers in one directive",
			configurationSnippet: `more_set_headers "X-First: first-value" "X-Second: second-value";`,
			expectedResponseHeaders: map[string]string{
				"X-First":  "first-value",
				"X-Second": "second-value",
			},
		},
		{
			desc: "add_header with variable interpolation",
			configurationSnippet: `
add_header X-Method $request_method;
add_header X-Uri $request_uri;
`,
			expectedResponseHeaders: map[string]string{
				"X-Method": "GET",
				"X-Uri":    "/test",
			},
		},
		{
			desc: "more_set_headers directive",
			configurationSnippet: `
more_set_headers "X-Custom-Header:custom-value";
more_set_headers "X-Another:another-value";
`,
			expectedResponseHeaders: map[string]string{
				"X-Custom-Header": "custom-value",
				"X-Another":       "another-value",
			},
		},
		{
			desc: "proxy_set_header directive",
			configurationSnippet: `
proxy_set_header X-Custom-Method $request_method;
proxy_set_header X-Custom-Uri $request_uri;
`,
			expectedRequestHeaders: map[string]string{
				"X-Custom-Method": "GET",
				"X-Custom-Uri":    "/test",
			},
		},
		{
			desc: "proxy_set_header with empty value removes header",
			configurationSnippet: `
proxy_set_header Accept-Encoding "";
`,
			requestHeaders: map[string]string{
				"Accept-Encoding": "gzip, deflate",
			},
			unexpectedResponseHeaders: []string{
				"Accept-Encoding",
			},
		},
		{
			desc: "set directive creates variable",
			configurationSnippet: `
set $my_var "hello";
add_header X-My-Var $my_var;
`,
			expectedResponseHeaders: map[string]string{
				"X-My-Var": "hello",
			},
		},
		{
			desc: "set directive with variable interpolation",
			configurationSnippet: `
set $combined "$request_method-$request_uri";
add_header X-Combined $combined;
`,
			expectedResponseHeaders: map[string]string{
				"X-Combined": "GET-/test",
			},
		},
		{
			desc: "if directive with matching condition",
			configurationSnippet: `
if ($request_method = GET) {
	add_header X-Is-Get "true";
}
`,
			method: http.MethodGet,
			expectedResponseHeaders: map[string]string{
				"X-Is-Get": "true",
			},
		},
		{
			desc: "if directive with non-matching condition",
			configurationSnippet: `
if ($request_method = POST) {
	add_header X-Is-Post "true";
}
add_header X-Always "present";
`,
			method: http.MethodGet,
			expectedResponseHeaders: map[string]string{
				"X-Always": "present",
			},
			unexpectedResponseHeaders: []string{
				"X-Is-Post",
			},
		},
		{
			desc: "if directive with header check",
			configurationSnippet: `
if ($http_x_custom = "expected") {
	add_header X-Matched "yes";
}
`,
			requestHeaders: map[string]string{
				"X-Custom": "expected",
			},
			expectedResponseHeaders: map[string]string{
				"X-Matched": "yes",
			},
		},
		{
			desc: "if directive with regex match",
			configurationSnippet: `
if ($request_uri ~ "^/api") {
	add_header X-Is-Api "true";
}
`,
			path: "/api/users",
			expectedResponseHeaders: map[string]string{
				"X-Is-Api": "true",
			},
		},
		{
			desc: "if directive with case-insensitive regex match - matching",
			configurationSnippet: `
if ($http_x_custom ~* "^test") {
	add_header X-Matched "yes";
}
`,
			requestHeaders: map[string]string{
				"X-Custom": "TEST-value",
			},
			expectedResponseHeaders: map[string]string{
				"X-Matched": "yes",
			},
		},
		{
			desc: "if directive with case-insensitive regex match - not matching",
			configurationSnippet: `
if ($http_x_custom ~* "^test") {
	add_header X-Matched "yes";
}
add_header X-Always "present";
`,
			requestHeaders: map[string]string{
				"X-Custom": "other-value",
			},
			expectedResponseHeaders: map[string]string{
				"X-Always": "present",
			},
			unexpectedResponseHeaders: []string{
				"X-Matched",
			},
		},
		{
			desc: "if directive with negative case-insensitive regex match",
			configurationSnippet: `
if ($http_x_custom !~* "^admin") {
	add_header X-Not-Admin "true";
}
`,
			requestHeaders: map[string]string{
				"X-Custom": "user-request",
			},
			expectedResponseHeaders: map[string]string{
				"X-Not-Admin": "true",
			},
		},
		{
			desc: "if directive with negative case-insensitive regex match - should not match",
			configurationSnippet: `
if ($http_x_custom !~* "^admin") {
	add_header X-Not-Admin "true";
}
add_header X-Processed "yes";
`,
			requestHeaders: map[string]string{
				"X-Custom": "ADMIN-request",
			},
			expectedResponseHeaders: map[string]string{
				"X-Processed": "yes",
			},
			unexpectedResponseHeaders: []string{
				"X-Not-Admin",
			},
		},
		{
			desc: "if directive with set variable check",
			configurationSnippet: `
set $flag "enabled";
if ($flag) {
	add_header X-Flag-Set "yes";
}
`,
			expectedResponseHeaders: map[string]string{
				"X-Flag-Set": "yes",
			},
		},
		{
			desc: "all directives combined",
			configurationSnippet: `
set $backend_type "api";
proxy_set_header X-Backend-Type $backend_type;
if ($request_method = GET) {
	add_header X-Read-Only "true";
	more_set_headers "X-Cache-Control:public";
}
add_header X-Powered-By "traefik";
`,
			method: http.MethodGet,
			expectedResponseHeaders: map[string]string{
				"X-Read-Only":     "true",
				"X-Cache-Control": "public",
			},
			expectedRequestHeaders: map[string]string{
				"X-Backend-Type": "api",
			},
		},
		{
			desc: "server and configuration snippets interaction",
			serverSnippet: `
add_header X-Server "server-value";
set $shared "from-server";
`,
			configurationSnippet: `
add_header X-Config "config-value";
`,
			expectedResponseHeaders: map[string]string{
				"X-Config": "config-value",
			},
			unexpectedResponseHeaders: []string{
				"X-Server",
			},
		},
		{
			desc: "return directive with status code and text",
			configurationSnippet: `
return 403 "Forbidden";
`,
			expectedStatusCode: http.StatusForbidden,
			expectedBody:       "Forbidden",
		},
		{
			desc: "return directive with 200 status",
			configurationSnippet: `
return 200 "OK";
`,
			expectedStatusCode: http.StatusOK,
			expectedBody:       "OK",
		},
		{
			desc: "return directive inside if block - condition matches",
			configurationSnippet: `
if ($request_method = POST) {
	return 405 "Method Not Allowed";
}
add_header X-Allowed "true";
`,
			method:             http.MethodPost,
			expectedStatusCode: http.StatusMethodNotAllowed,
			expectedBody:       "Method Not Allowed",
		},
		{
			desc: "return directive inside if block - condition does not match",
			configurationSnippet: `
if ($request_method = POST) {
	return 405 "Method Not Allowed";
}
add_header X-Allowed "true";
`,
			method:             http.MethodGet,
			expectedStatusCode: http.StatusOK,
			expectedResponseHeaders: map[string]string{
				"X-Allowed": "true",
			},
		},
		{
			desc: "return directive doesn't stop processing headers",
			configurationSnippet: `
return 204 "";
add_header X-Should-Appear "value";
`,
			expectedStatusCode: http.StatusNoContent,
			expectedBody:       "",
			expectedResponseHeaders: map[string]string{
				"X-Should-Appear": "value",
			},
		},
		{
			desc: "location without return passes through to next handler",
			serverSnippet: `
location /api {
	add_header X-Location "api";
}
`,
			path: "/api/users",
			expectedResponseHeaders: map[string]string{
				"X-Location": "api",
			},
		},
		{
			desc: "location directive with prefix match - not matching continues to next",
			serverSnippet: `
location /api {
	return 200 "OK";
}
add_header X-Always "present";
`,
			path: "/web/users",
			expectedResponseHeaders: map[string]string{
				"X-Always": "present",
			},
		},
		{
			desc: "location directive with exact match and return",
			serverSnippet: `
location = /exact {
	return 200 "exact";
}
`,
			path:               "/exact",
			expectedStatusCode: http.StatusOK,
			expectedBody:       "exact",
		},
		{
			desc: "location directive with exact match - not matching continues to next",
			serverSnippet: `
location = /exact {
	return 200 "exact";
}
add_header X-Always "present";
`,
			path: "/exact/more",
			expectedResponseHeaders: map[string]string{
				"X-Always": "present",
			},
		},
		{
			desc: "location directive with regex match and return",
			serverSnippet: `
location ~ ^/api/v[0-9]+/ {
	return 200 "versioned";
}
`,
			path:               "/api/v2/users",
			expectedStatusCode: http.StatusOK,
			expectedBody:       "versioned",
		},
		{
			desc: "location directive with regex match - not matching continues to next",
			serverSnippet: `
location ~ ^/api/v[0-9]+/ {
	return 200 "versioned";
}
add_header X-Always "present";
`,
			path: "/api/latest/users",
			expectedResponseHeaders: map[string]string{
				"X-Always": "present",
			},
		},
		{
			desc: "location with return applies add_header always from same block",
			serverSnippet: `
location /blocked {
	add_header X-Block-Header "block-value" always;
	return 403 "Blocked";
}
`,
			path:               "/blocked/path",
			expectedStatusCode: http.StatusForbidden,
			expectedBody:       "Blocked",
			expectedResponseHeaders: map[string]string{
				"X-Block-Header": "block-value",
			},
		},
		{
			desc: "add_header without always skips non-success status codes",
			serverSnippet: `
location /blocked {
	add_header X-Block-Header "block-value";
	return 403 "Blocked";
}
`,
			path:                      "/blocked/path",
			expectedStatusCode:        http.StatusForbidden,
			expectedBody:              "Blocked",
			unexpectedResponseHeaders: []string{"X-Block-Header"},
		},
		{
			desc: "location with return applies more_set_headers from same block",
			serverSnippet: `
location /blocked {
	more_set_headers "X-More-Header:more-value";
	return 403 "Blocked";
}
`,
			path:               "/blocked/path",
			expectedStatusCode: http.StatusForbidden,
			expectedBody:       "Blocked",
			expectedResponseHeaders: map[string]string{
				"X-More-Header": "more-value",
			},
		},
		{
			desc: "location with return applies both add_header and more_set_headers",
			serverSnippet: `
location /api {
	add_header X-Add "add-value";
	more_set_headers "X-More:more-value";
	return 200 "OK";
}
`,
			path:               "/api/endpoint",
			expectedStatusCode: http.StatusOK,
			expectedBody:       "OK",
			expectedResponseHeaders: map[string]string{
				"X-Add":  "add-value",
				"X-More": "more-value",
			},
		},
		{
			desc: "add_header only applied in deepest block - location overrides root",
			serverSnippet: `
add_header X-Level "root";
location /api {
	add_header X-Level "location";
	return 200 "OK";
}
`,
			path:               "/api/endpoint",
			expectedStatusCode: http.StatusOK,
			expectedBody:       "OK",
			expectedResponseHeaders: map[string]string{
				"X-Level": "location",
			},
		},
		{
			desc: "add_header only applied in deepest block - nested if inside location",
			serverSnippet: `
add_header X-Level "root";
location /api {
	add_header X-Level "location";
	if ($request_method = GET) {
		add_header X-Level "if-block";
		return 200 "OK";
	}
}
`,
			path:               "/api/endpoint",
			method:             http.MethodGet,
			expectedStatusCode: http.StatusOK,
			expectedBody:       "OK",
			expectedResponseHeaders: map[string]string{
				"X-Level": "if-block",
			},
		},
		{
			desc: "add_header from location when if condition not matched",
			serverSnippet: `
add_header X-Level "root";
location /api {
	add_header X-Level "location";
	if ($request_method = POST) {
		add_header X-Level "if-block";
		return 200 "POST";
	}
	return 200 "OTHER";
}
`,
			path:               "/api/endpoint",
			method:             http.MethodGet,
			expectedStatusCode: http.StatusOK,
			expectedBody:       "OTHER",
			expectedResponseHeaders: map[string]string{
				"X-Level": "location",
			},
		},
		{
			desc: "root add_header applied when location not matched",
			serverSnippet: `
add_header X-Level "root";
location /api {
	add_header X-Level "location";
	return 200 "API";
}
`,
			path: "/web/endpoint",
			expectedResponseHeaders: map[string]string{
				"X-Level": "root",
			},
		},
		{
			desc: "more_set_input_headers sets request header",
			configurationSnippet: `
more_set_input_headers "X-Custom-Input:input-value";
`,
			expectedRequestHeaders: map[string]string{
				"X-Custom-Input": "input-value",
			},
		},
		{
			desc: "more_set_input_headers with variable interpolation",
			configurationSnippet: `
more_set_input_headers "X-Method-Input:$request_method";
`,
			expectedRequestHeaders: map[string]string{
				"X-Method-Input": "GET",
			},
		},
		{
			desc: "more_set_headers with multiple headers per directive",
			configurationSnippet: `
more_set_headers "X-Foo: bar" "X-Baz: qux";
`,
			expectedResponseHeaders: map[string]string{
				"X-Foo": "bar",
				"X-Baz": "qux",
			},
		},
		{
			desc: "more_set_headers clearing header with colon",
			serverSnippet: `
more_set_headers "X-Foo: server-value";
`,
			configurationSnippet: `
more_set_headers "X-Foo:";
`,
			unexpectedResponseHeaders: []string{
				"X-Foo",
			},
		},
		{
			desc: "more_set_headers clearing header without colon",
			serverSnippet: `
more_set_headers "X-Foo: server-value";
`,
			configurationSnippet: `
more_set_headers "X-Foo";
`,
			unexpectedResponseHeaders: []string{
				"X-Foo",
			},
		},
		{
			desc: "more_set_input_headers with multiple headers per directive",
			configurationSnippet: `
more_set_input_headers "X-Foo: bar" "X-Baz: qux";
`,
			expectedRequestHeaders: map[string]string{
				"X-Foo": "bar",
				"X-Baz": "qux",
			},
		},
		{
			desc: "more_set_input_headers clearing request header",
			configurationSnippet: `
more_set_input_headers "X-Foo";
`,
			requestHeaders: map[string]string{
				"X-Foo": "original-value",
			},
			unexpectedResponseHeaders: []string{
				"X-Foo",
			},
		},
		{
			desc: "more_clear_headers clears a single response header",
			configurationSnippet: `
more_set_headers "X-Remove-Me: some-value";
more_clear_headers "X-Remove-Me";
`,
			unexpectedResponseHeaders: []string{
				"X-Remove-Me",
			},
		},
		{
			desc: "more_clear_headers clears multiple response headers",
			configurationSnippet: `
more_set_headers "X-Foo: foo-val";
more_set_headers "X-Bar: bar-val";
more_set_headers "X-Keep: keep-val";
more_clear_headers Foo Bar;
`,
			expectedResponseHeaders: map[string]string{
				"X-Foo":  "foo-val",
				"X-Bar":  "bar-val",
				"X-Keep": "keep-val",
			},
		},
		{
			desc: "more_clear_headers clears multiple headers by exact name",
			configurationSnippet: `
more_set_headers "X-Foo: foo-val";
more_set_headers "X-Bar: bar-val";
more_set_headers "X-Keep: keep-val";
more_clear_headers "X-Foo" "X-Bar";
`,
			expectedResponseHeaders: map[string]string{
				"X-Foo":  "",
				"X-Bar":  "",
				"X-Keep": "keep-val",
			},
		},
		{
			desc: "more_clear_headers with wildcard pattern",
			configurationSnippet: `
more_set_headers "X-Hidden-One: val1";
more_set_headers "X-Hidden-Two: val2";
more_set_headers "X-Visible: visible";
more_clear_headers "X-Hidden-*";
`,
			expectedResponseHeaders: map[string]string{
				"X-Visible": "visible",
			},
			unexpectedResponseHeaders: []string{
				"X-Hidden-One",
				"X-Hidden-Two",
			},
		},
		{
			desc: "more_clear_input_headers removes a request header",
			configurationSnippet: `
more_clear_input_headers "X-Secret";
`,
			requestHeaders: map[string]string{
				"X-Secret": "secret-value",
			},
			unexpectedResponseHeaders: []string{
				"X-Secret",
			},
		},
		{
			desc: "more_clear_input_headers with wildcard removes matching request headers",
			configurationSnippet: `
more_clear_input_headers "X-Custom-*";
`,
			requestHeaders: map[string]string{
				"X-Custom-One": "val1",
				"X-Custom-Two": "val2",
				"X-Other":      "other",
			},
			expectedRequestHeaders: map[string]string{
				"X-Other": "other",
			},
			unexpectedResponseHeaders: []string{
				"X-Custom-One",
				"X-Custom-Two",
			},
		},
		{
			desc: "more_clear_headers in configuration-snippet clears server-snippet header",
			serverSnippet: `
more_set_headers "X-Server-Header: server-value";
`,
			configurationSnippet: `
more_clear_headers "X-Server-Header";
`,
			unexpectedResponseHeaders: []string{
				"X-Server-Header",
			},
		},
		{
			desc: "rewrite with capture groups and last flag",
			serverSnippet: `
rewrite ^/old/(.*)$ /new/$1 last;
`,
			path:         "/old/page",
			expectedPath: "/new/page",
		},
		{
			desc: "rewrite with permanent redirect",
			serverSnippet: `
rewrite ^ https://example.com$request_uri? permanent;
`,
			path:                "/some/path",
			expectedStatusCode:  http.StatusMovedPermanently,
			expectedRedirectURL: "https://example.com/some/path",
		},
		{
			desc: "rewrite with break flag",
			configurationSnippet: `
rewrite ^/api/v1/(.*)$ /api/v2/$1 break;
`,
			path:         "/api/v1/users",
			expectedPath: "/api/v2/users",
		},
		{
			desc: "rewrite with redirect flag",
			configurationSnippet: `
rewrite ^/old$ /new redirect;
`,
			path:                "/old",
			expectedStatusCode:  http.StatusFound,
			expectedRedirectURL: "/new",
		},
		{
			desc: "rewrite with no match passes through",
			configurationSnippet: `
rewrite ^/nomatch /other last;
`,
			path:         "/test",
			expectedPath: "/test",
		},
		{
			desc: "rewrite preserves query string",
			configurationSnippet: `
rewrite ^/search$ /find last;
`,
			path:          "/search?q=test",
			expectedPath:  "/find",
			expectedQuery: "q=test",
		},
		{
			desc: "rewrite with ? suffix suppresses query string",
			configurationSnippet: `
rewrite ^/search$ /find? last;
`,
			path:          "/search?q=test",
			expectedPath:  "/find",
			expectedQuery: "",
		},
		{
			desc: "rewrite in configuration-snippet (location context)",
			configurationSnippet: `
rewrite ^/old/(.*)$ /new/$1 last;
`,
			path:         "/old/resource",
			expectedPath: "/new/resource",
		},
		{
			desc: "rewrite in if block",
			configurationSnippet: `
if ($request_method = GET) {
	rewrite ^/old/(.*)$ /new/$1 last;
}
`,
			method:       http.MethodGet,
			path:         "/old/page",
			expectedPath: "/new/page",
		},
		{
			desc: "rewrite with multiple capture groups",
			serverSnippet: `
rewrite ^/download/(.*)/media/(.*)\..*$ /download/$1/mp3/$2.mp3 last;
`,
			path:         "/download/music/media/song.flac",
			expectedPath: "/download/music/mp3/song.mp3",
		},
		{
			desc: "rewrite with no flag continues processing",
			configurationSnippet: `
rewrite ^/a$ /b;
rewrite ^/b$ /c last;
`,
			path:         "/a",
			expectedPath: "/c",
		},
		{
			desc: "rewrite with URL-based redirect (http://)",
			configurationSnippet: `
rewrite ^/old$ http://other.example.com/new last;
`,
			path:                "/old",
			expectedStatusCode:  http.StatusFound,
			expectedRedirectURL: "http://other.example.com/new",
		},
		// --- add_header always tests ---
		{
			desc: "add_header with always applies to 200 status",
			configurationSnippet: `
add_header X-Custom "always-value" always;
`,
			expectedResponseHeaders: map[string]string{
				"X-Custom": "always-value",
			},
		},
		{
			desc: "add_header without always applies to 200 status",
			configurationSnippet: `
add_header X-Custom "no-always-value";
`,
			expectedResponseHeaders: map[string]string{
				"X-Custom": "no-always-value",
			},
		},
		{
			desc: "add_header without always skips 404 status",
			serverSnippet: `
location /missing {
	add_header X-Custom "no-always-value";
	return 404 "Not Found";
}
`,
			path:                      "/missing",
			expectedStatusCode:        http.StatusNotFound,
			unexpectedResponseHeaders: []string{"X-Custom"},
		},
		{
			desc: "add_header with always applies to 404 status",
			serverSnippet: `
location /missing {
	add_header X-Custom "always-value" always;
	return 404 "Not Found";
}
`,
			path:               "/missing",
			expectedStatusCode: http.StatusNotFound,
			expectedResponseHeaders: map[string]string{
				"X-Custom": "always-value",
			},
		},
		// --- deny/allow tests ---
		{
			desc: "deny all blocks request",
			configurationSnippet: `
deny all;
`,
			remoteAddr:         "192.168.1.1:12345",
			expectedStatusCode: http.StatusForbidden,
			expectedBody:       "403 Forbidden",
		},
		{
			desc: "allow all permits request",
			configurationSnippet: `
allow all;
`,
			remoteAddr: "192.168.1.1:12345",
		},
		{
			desc: "deny specific IP blocks matching request",
			configurationSnippet: `
deny 10.0.0.1;
`,
			remoteAddr:         "10.0.0.1:12345",
			expectedStatusCode: http.StatusForbidden,
			expectedBody:       "403 Forbidden",
		},
		{
			desc: "deny specific IP allows non-matching request",
			configurationSnippet: `
deny 10.0.0.1;
`,
			remoteAddr: "10.0.0.2:12345",
		},
		{
			desc: "deny CIDR blocks matching request",
			configurationSnippet: `
deny 10.0.0.0/24;
`,
			remoteAddr:         "10.0.0.50:12345",
			expectedStatusCode: http.StatusForbidden,
			expectedBody:       "403 Forbidden",
		},
		{
			desc: "deny CIDR allows non-matching request",
			configurationSnippet: `
deny 10.0.0.0/24;
`,
			remoteAddr: "10.0.1.1:12345",
		},
		{
			desc: "allow then deny all permits allowed IP",
			configurationSnippet: `
allow 192.168.1.0/24;
deny all;
`,
			remoteAddr: "192.168.1.50:12345",
		},
		{
			desc: "allow then deny all blocks non-allowed IP",
			configurationSnippet: `
allow 192.168.1.0/24;
deny all;
`,
			remoteAddr:         "10.0.0.1:12345",
			expectedStatusCode: http.StatusForbidden,
			expectedBody:       "403 Forbidden",
		},
		// --- proxy_hide_header tests ---
		{
			desc: "proxy_hide_header removes response header",
			configurationSnippet: `
proxy_hide_header X-Powered-By;
`,
			unexpectedResponseHeaders: []string{"X-Powered-By"},
		},
		// --- expires tests ---
		{
			desc: "expires epoch sets no-cache",
			configurationSnippet: `
expires epoch;
`,
			expectedResponseHeaders: map[string]string{
				"Expires":       "Thu, 01 Jan 1970 00:00:01 GMT",
				"Cache-Control": "no-cache",
			},
		},
		{
			desc: "expires max sets far-future cache",
			configurationSnippet: `
expires max;
`,
			expectedResponseHeaders: map[string]string{
				"Expires":       "Thu, 31 Dec 2037 23:55:55 GMT",
				"Cache-Control": "max-age=315360000",
			},
		},
		{
			desc: "expires off does not set cache headers",
			configurationSnippet: `
expires off;
`,
			unexpectedResponseHeaders: []string{"Expires", "Cache-Control"},
		},
		// --- more_set_headers with flags ---
		{
			desc: "more_set_headers with -a append flag",
			configurationSnippet: `
more_set_headers "X-Custom: first";
more_set_headers -a "X-Custom: second";
`,
			expectedResponseHeaders: map[string]string{
				"X-Custom": "first",
			},
		},
		{
			desc: "more_set_headers clearing header with empty value",
			configurationSnippet: `
more_set_headers "X-Remove:";
`,
			unexpectedResponseHeaders: []string{"X-Remove"},
		},
		{
			desc: "more_set_headers clearing header without colon",
			configurationSnippet: `
more_set_headers "X-Remove";
`,
			unexpectedResponseHeaders: []string{"X-Remove"},
		},
		// --- more_set_input_headers with -r restrict flag ---
		{
			desc: "more_set_input_headers with -r only sets existing header",
			configurationSnippet: `
more_set_input_headers -r "X-Existing: new-value";
`,
			requestHeaders: map[string]string{
				"X-Existing": "old-value",
			},
			expectedRequestHeaders: map[string]string{
				"X-Existing": "new-value",
			},
		},
		{
			desc: "more_set_input_headers with -r skips non-existing header",
			configurationSnippet: `
more_set_input_headers -r "X-Missing: new-value";
`,
			unexpectedResponseHeaders: []string{
				"X-Missing",
			},
		},
		// --- more_set_headers with multiple headers per directive ---
		{
			desc: "more_set_headers with multiple headers in single directive",
			configurationSnippet: `
more_set_headers "X-One: val1" "X-Two: val2";
`,
			expectedResponseHeaders: map[string]string{
				"X-One": "val1",
				"X-Two": "val2",
			},
		},
		// --- more_clear_input_headers ---
		{
			desc: "more_clear_input_headers removes request header",
			configurationSnippet: `
more_clear_input_headers "X-Remove";
`,
			requestHeaders: map[string]string{
				"X-Remove": "should-be-removed",
			},
			unexpectedResponseHeaders: []string{
				"X-Remove",
			},
		},
		// --- rewrite last/break forwards to upstream ---
		{
			desc: "rewrite break in server snippet skips configuration snippet",
			serverSnippet: `
rewrite ^/old/(.*)$ /new/$1 break;
`,
			configurationSnippet: `
add_header X-Config "config-value" always;
`,
			path:                      "/old/page",
			expectedPath:              "/new/page",
			unexpectedResponseHeaders: []string{"X-Config"},
		},
		{
			desc: "rewrite last in server snippet allows configuration snippet to run",
			serverSnippet: `
rewrite ^/old/(.*)$ /new/$1 last;
`,
			configurationSnippet: `
add_header X-Config "config-value" always;
`,
			path:         "/old/page",
			expectedPath: "/new/page",
			expectedResponseHeaders: map[string]string{
				"X-Config": "config-value",
			},
		},
		// --- if single variable check with built-in variables ---
		{
			desc: "if single variable check with built-in variable",
			configurationSnippet: `
if ($request_method) {
	add_header X-Has-Method "yes";
}
`,
			expectedResponseHeaders: map[string]string{
				"X-Has-Method": "yes",
			},
		},
		// --- if regex capture groups ---
		{
			desc: "if regex capture groups stored as $1",
			configurationSnippet: `
if ($request_uri ~ "^/api/(.*)") {
	add_header X-Captured "$1";
}
`,
			path: "/api/users",
			expectedResponseHeaders: map[string]string{
				"X-Captured": "users",
			},
		},
		// --- location ~* case-insensitive regex ---
		{
			desc: "location with case-insensitive regex matches",
			serverSnippet: `
location ~* \.css$ {
	add_header X-Type "css" always;
	return 200 "CSS";
}
`,
			path:               "/style/main.CSS",
			expectedStatusCode: http.StatusOK,
			expectedBody:       "CSS",
			expectedResponseHeaders: map[string]string{
				"X-Type": "css",
			},
		},
		{
			desc: "location with case-insensitive regex does not match",
			serverSnippet: `
location ~* \.css$ {
	return 200 "CSS";
}
add_header X-Fallback "yes";
`,
			path: "/style/main.js",
			expectedResponseHeaders: map[string]string{
				"X-Fallback": "yes",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			capturedRequestHeaders := make(map[string]string)
			var capturedPath, capturedQuery string
			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				capturedPath = r.URL.Path
				capturedQuery = r.URL.RawQuery
				for header := range test.expectedRequestHeaders {
					capturedRequestHeaders[header] = r.Header.Get(header)
				}
				w.WriteHeader(http.StatusOK)
			})

			config := &dynamic.Snippet{
				ServerSnippet:        test.serverSnippet,
				ConfigurationSnippet: test.configurationSnippet,
			}

			handler, err := New(t.Context(), next, config, "test-snippet")
			require.NoError(t, err)

			method := test.method
			if method == "" {
				method = http.MethodGet
			}
			path := test.path
			if path == "" {
				path = "/test"
			}

			req := httptest.NewRequest(method, "http://example.com"+path, nil)
			if test.remoteAddr != "" {
				req.RemoteAddr = test.remoteAddr
			}
			for k, v := range test.requestHeaders {
				req.Header.Set(k, v)
			}
			rw := httptest.NewRecorder()

			handler.ServeHTTP(rw, req)

			expectedStatusCode := test.expectedStatusCode
			if expectedStatusCode == 0 {
				expectedStatusCode = http.StatusOK
			}
			assert.Equal(t, expectedStatusCode, rw.Code)

			if test.expectedBody != "" {
				assert.Equal(t, test.expectedBody, rw.Body.String())
			}

			// If a return directive was used, next should not be called
			if test.expectedStatusCode != 0 && test.expectedStatusCode != http.StatusOK {
				assert.False(t, nextCalled, "next handler should not be called when return directive is used")
			}

			for header, expectedValue := range test.expectedResponseHeaders {
				assert.Equal(t, expectedValue, rw.Header().Get(header), "response header %s", header)
			}

			for _, header := range test.unexpectedResponseHeaders {
				assert.Empty(t, rw.Header().Get(header), "response header %s should not be set", header)
			}

			for header, expectedValue := range test.expectedRequestHeaders {
				assert.Equal(t, expectedValue, capturedRequestHeaders[header], "request header %s", header)
			}

			if test.expectedPath != "" && nextCalled {
				assert.Equal(t, test.expectedPath, capturedPath, "rewritten path")
			}

			if test.expectedQuery != "" && nextCalled {
				assert.Equal(t, test.expectedQuery, capturedQuery, "rewritten query")
			} else if test.expectedQuery == "" && test.expectedPath != "" && nextCalled {
				assert.Empty(t, capturedQuery, "query string should be empty")
			}

			if test.expectedRedirectURL != "" {
				assert.Contains(t, rw.Header().Get("Location"), test.expectedRedirectURL, "redirect URL")
			}
		})
	}
}
