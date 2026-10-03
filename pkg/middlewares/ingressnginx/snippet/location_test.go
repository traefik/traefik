package snippet

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func Test_LocationSelection(t *testing.T) {
	testCases := []struct {
		desc     string
		snippet  string
		path     string
		wantBody string
		wantNext bool
	}{
		{
			desc: "exact before prefix and regex",
			snippet: `location = /api { return 200 exact; }
location / { return 200 prefix; }
location ~ ^/api { return 200 regex; }`,
			path: "/api", wantBody: "exact",
		},
		{
			desc: "exact after regex and prefix",
			snippet: `location ~ ^/api { return 200 regex; }
location / { return 200 prefix; }
location = /api { return 200 exact; }`,
			path: "/api", wantBody: "exact",
		},
		{
			desc: "longest prefix first",
			snippet: `location /api/v1 { return 200 longest; }
location /api { return 200 shorter; }
location / { return 200 root; }`,
			path: "/api/v1/users", wantBody: "longest",
		},
		{
			desc: "longest prefix last",
			snippet: `location / { return 200 root; }
location /api { return 200 shorter; }
location /api/v1 { return 200 longest; }`,
			path: "/api/v1/users", wantBody: "longest",
		},
		{
			desc: "preferred prefix before regex",
			snippet: `location ^~ /api { return 200 prefix; }
location ~ ^/api { return 200 regex; }`,
			path: "/api/users", wantBody: "prefix",
		},
		{
			desc: "preferred prefix after regex",
			snippet: `location ~ ^/api { return 200 regex; }
location ^~ /api { return 200 prefix; }`,
			path: "/api/users", wantBody: "prefix",
		},
		{
			desc: "shorter preferred prefix does not suppress regex",
			snippet: `location /api/v1 { return 200 longest; }
location ~ ^/api { return 200 regex; }
location ^~ /api { return 200 shorter; }`,
			path: "/api/v1/users", wantBody: "regex",
		},
		{
			desc: "first matching regex wins over later regex and prefix",
			snippet: `location ~ ^/other { return 200 unrelated; }
location ~* ^/API { return 200 first; }
location ~ ^/api { return 200 second; }
location /api { return 200 prefix; }`,
			path: "/api/users", wantBody: "first",
		},
		{
			desc: "regex wins without a matching prefix",
			snippet: `location ~ ^/api { return 200 regex; }
location /other { return 200 other; }`,
			path: "/api/users", wantBody: "regex",
		},
		{
			desc: "nonmatching regex falls back to longest prefix",
			snippet: `location /api { return 200 prefix; }
location / { return 200 root; }
location ~ ^/other { return 200 regex; }`,
			path: "/api/users", wantBody: "prefix",
		},
		{
			desc: "no matching location reaches backend",
			snippet: `location = /api { return 200 exact; }
location /other { return 200 prefix; }
location ~ ^/v1 { return 200 regex; }`,
			path: "/api/users", wantBody: "backend", wantNext: true,
		},
		{
			desc: "selected location without return does not run shorter prefix",
			snippet: `location /api { add_header X-Selected yes; }
location / { return 200 root; }`,
			path: "/api/users", wantBody: "backend", wantNext: true,
		},
		{
			desc: "server return prevents location execution",
			snippet: `return 200 server;
location /api { return 200 location; }`,
			path: "/api", wantBody: "server",
		},
		{
			desc: "server set runs before location regardless of declaration order",
			snippet: `location /api { return 200 $result; }
set $result server;`,
			path: "/api", wantBody: "server",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			nextCalled := false
			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				nextCalled = true
				_, _ = rw.Write([]byte("backend"))
			})
			handler, err := New(t.Context(), next, &dynamic.Snippet{ServerSnippet: test.snippet}, "test-snippet")
			require.NoError(t, err)

			rw := httptest.NewRecorder()
			handler.ServeHTTP(rw, httptest.NewRequest(http.MethodGet, test.path, nil))

			assert.Equal(t, http.StatusOK, rw.Code)
			assert.Equal(t, test.wantBody, rw.Body.String())
			assert.Equal(t, test.wantNext, nextCalled)
		})
	}
}

func Test_LocationSelectionHeaders(t *testing.T) {
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Empty(t, req.Header.Get("X-Unselected"))
		assert.Equal(t, "yes", req.Header.Get("X-Selected"))
		rw.WriteHeader(http.StatusNoContent)
	})
	handler, err := New(t.Context(), next, &dynamic.Snippet{ServerSnippet: `
location /api {
    add_header X-Selected yes;
    more_set_input_headers "X-Selected: yes";
}
location / {
    add_header X-Unselected yes;
    more_set_headers "X-Unselected-More: yes";
    more_set_input_headers "X-Unselected: yes";
}
add_header X-Server yes;
more_set_headers "X-Server-More: yes";
`}, "test-snippet")
	require.NoError(t, err)

	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, httptest.NewRequest(http.MethodGet, "/api/users", nil))

	assert.Equal(t, http.StatusNoContent, rw.Code)
	assert.Equal(t, "yes", rw.Header().Get("X-Selected"))
	assert.Equal(t, "yes", rw.Header().Get("X-Server-More"))
	assert.Empty(t, rw.Header().Get("X-Unselected"))
	assert.Empty(t, rw.Header().Get("X-Unselected-More"))
	assert.Empty(t, rw.Header().Get("X-Server"))
}
