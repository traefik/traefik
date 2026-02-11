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
			desc: "fails with unknown directive in configuration snippet",
			config: dynamic.Snippet{
				ConfigurationSnippet: `unknown_directive value;`,
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

func Test_ServeHTTP_responseHeaders(t *testing.T) {
	testCases := []struct {
		desc                 string
		serverSnippet        string
		configurationSnippet string
		expectedHeaders      map[string]string
	}{
		{
			desc:            "add_header server snippet adds simple header",
			serverSnippet:   `add_header X-Custom "custom-value";`,
			expectedHeaders: map[string]string{"X-Custom": "custom-value"},
		},
		{
			desc:            "add_header server snippet adds header without quotes",
			serverSnippet:   `add_header X-Simple simple;`,
			expectedHeaders: map[string]string{"X-Simple": "simple"},
		},
		{
			desc:                 "add_header configuration snippet adds simple header",
			configurationSnippet: `add_header X-Custom "custom-value";`,
			expectedHeaders:      map[string]string{"X-Custom": "custom-value"},
		},
		{
			desc:                 "add_header configuration snippet adds header without quotes",
			configurationSnippet: `add_header X-Simple simple;`,
			expectedHeaders:      map[string]string{"X-Simple": "simple"},
		},
		{
			desc:                 "add_header configuration snippet overrides server snippet",
			serverSnippet:        `add_header X-Server server-value;`,
			configurationSnippet: `add_header X-Config config-value;`,
			expectedHeaders: map[string]string{
				"X-Server": "",
				"X-Config": "config-value",
			},
		},
		{
			desc:            "more_set_headers server snippet sets header",
			serverSnippet:   `more_set_headers "X-Custom:custom-value";`,
			expectedHeaders: map[string]string{"X-Custom": "custom-value"},
		},
		{
			desc:                 "more_set_headers configuration snippet sets header",
			configurationSnippet: `more_set_headers "X-Custom:custom-value";`,
			expectedHeaders:      map[string]string{"X-Custom": "custom-value"},
		},
		{
			desc:                 "more_set_headers both snippets set headers",
			serverSnippet:        `more_set_headers "X-Server:server-value";`,
			configurationSnippet: `more_set_headers "X-Config:config-value";`,
			expectedHeaders: map[string]string{
				"X-Server": "server-value",
				"X-Config": "config-value",
			},
		},
		{
			desc:                 "more_set_headers both snippets override same header",
			serverSnippet:        `more_set_headers "X-Header:server-value";`,
			configurationSnippet: `more_set_headers "X-Header:config-value";`,
			expectedHeaders: map[string]string{
				"X-Header": "config-value",
			},
		},
		{
			desc:                 "more_set_headers with spaces",
			configurationSnippet: `more_set_headers "X-Header: config-value ";`,
			expectedHeaders: map[string]string{
				"X-Header": "config-value",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			config := &dynamic.Snippet{
				ServerSnippet:        test.serverSnippet,
				ConfigurationSnippet: test.configurationSnippet,
			}

			handler, err := New(t.Context(), next, config, "test-snippet")
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rw := httptest.NewRecorder()

			handler.ServeHTTP(rw, req)

			assert.Equal(t, http.StatusOK, rw.Code)
			for header, value := range test.expectedHeaders {
				assert.Equal(t, value, rw.Header().Get(header))
			}
		})
	}
}

func Test_Directives(t *testing.T) {
	testCases := []struct {
		desc                    string
		serverSnippet           string
		configurationSnippet    string
		method                  string
		path                    string
		requestHeaders          map[string]string
		expectedResponseHeaders map[string]string
		expectedRequestHeaders  map[string]string
		expectedStatusCode      int
		expectedBody            string
	}{
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
				"X-Is-Post": "",
				"X-Always":  "present",
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
				"X-Matched": "",
				"X-Always":  "present",
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
				"X-Not-Admin": "",
				"X-Processed": "yes",
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
				"X-Server": "",
				"X-Config": "config-value",
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
			desc: "location without return returns 503",
			serverSnippet: `
location /api {
	add_header X-Location "api";
}
`,
			path:               "/api/users",
			expectedStatusCode: http.StatusServiceUnavailable,
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
			desc: "location with return applies add_header from same block",
			serverSnippet: `
location /blocked {
	add_header X-Block-Header "block-value";
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
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			capturedRequestHeaders := make(map[string]string)
			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
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

			for header, expectedValue := range test.expectedRequestHeaders {
				assert.Equal(t, expectedValue, capturedRequestHeaders[header], "request header %s", header)
			}
		})
	}
}
