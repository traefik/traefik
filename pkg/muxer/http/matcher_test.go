package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/middlewares/requestdecorator"
)

func TestClientIPMatcher(t *testing.T) {
	testCases := []struct {
		desc          string
		rule          string
		expected      map[string]int
		expectedError bool
	}{
		{
			desc:          "invalid ClientIP matcher",
			rule:          "ClientIP(`1`)",
			expectedError: true,
		},
		{
			desc:          "invalid ClientIP matcher (no parameter)",
			rule:          "ClientIP()",
			expectedError: true,
		},
		{
			desc:          "invalid ClientIP matcher (empty parameter)",
			rule:          "ClientIP(``)",
			expectedError: true,
		},
		{
			desc:          "invalid ClientIP matcher (too many parameters)",
			rule:          "ClientIP(`127.0.0.1`, `192.168.1.0/24`)",
			expectedError: true,
		},
		{
			desc: "valid ClientIP matcher",
			rule: "ClientIP(`127.0.0.1`)",
			expected: map[string]int{
				"127.0.0.1":   http.StatusOK,
				"192.168.1.1": http.StatusNotFound,
			},
		},
		{
			desc: "valid ClientIP matcher but invalid remote address",
			rule: "ClientIP(`127.0.0.1`)",
			expected: map[string]int{
				"1": http.StatusNotFound,
			},
		},
		{
			desc: "valid ClientIP matcher using CIDR",
			rule: "ClientIP(`192.168.1.0/24`)",
			expected: map[string]int{
				"192.168.1.1":   http.StatusOK,
				"192.168.1.100": http.StatusOK,
				"192.168.2.1":   http.StatusNotFound,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			parser, err := NewSyntaxParser()
			require.NoError(t, err)

			muxer := NewMuxer(parser)

			err = muxer.AddRoute(test.rule, "", 0, handler)
			if test.expectedError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			results := make(map[string]int)
			for remoteAddr := range test.expected {
				w := httptest.NewRecorder()

				req := httptest.NewRequest(http.MethodGet, "https://example.com", http.NoBody)
				req.RemoteAddr = remoteAddr

				muxer.ServeHTTP(w, req)
				results[remoteAddr] = w.Code
			}
			assert.Equal(t, test.expected, results)
		})
	}
}

func TestMethodMatcher(t *testing.T) {
	testCases := []struct {
		desc          string
		rule          string
		expected      map[string]int
		expectedError bool
	}{
		{
			desc:          "invalid Method matcher (no parameter)",
			rule:          "Method()",
			expectedError: true,
		},
		{
			desc:          "invalid Method matcher (empty parameter)",
			rule:          "Method(``)",
			expectedError: true,
		},
		{
			desc:          "invalid Method matcher (too many parameters)",
			rule:          "Method(`GET`, `POST`)",
			expectedError: true,
		},
		{
			desc: "valid Method matcher",
			rule: "Method(`GET`)",
			expected: map[string]int{
				http.MethodGet:                  http.StatusOK,
				http.MethodPost:                 http.StatusNotFound,
				strings.ToLower(http.MethodGet): http.StatusNotFound,
			},
		},
		{
			desc: "valid Method matcher (lower case)",
			rule: "Method(`get`)",
			expected: map[string]int{
				http.MethodGet:                  http.StatusOK,
				http.MethodPost:                 http.StatusNotFound,
				strings.ToLower(http.MethodGet): http.StatusNotFound,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			parser, err := NewSyntaxParser()
			require.NoError(t, err)

			muxer := NewMuxer(parser)

			err = muxer.AddRoute(test.rule, "", 0, handler)
			if test.expectedError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			results := make(map[string]int)
			for method := range test.expected {
				w := httptest.NewRecorder()

				req := httptest.NewRequest(method, "https://example.com", http.NoBody)

				muxer.ServeHTTP(w, req)
				results[method] = w.Code
			}
			assert.Equal(t, test.expected, results)
		})
	}
}

func TestHostMatcher(t *testing.T) {
	testCases := []struct {
		desc          string
		rule          string
		expected      map[string]int
		expectedError bool
	}{
		{
			desc:          "invalid Host matcher (no parameter)",
			rule:          "Host()",
			expectedError: true,
		},
		{
			desc:          "invalid Host matcher (empty parameter)",
			rule:          "Host(``)",
			expectedError: true,
		},
		{
			desc:          "invalid Host matcher (non-ASCII)",
			rule:          "Host(`🦭.com`)",
			expectedError: true,
		},
		{
			desc:          "invalid Host matcher (too many parameters)",
			rule:          "Host(`example.com`, `example.org`)",
			expectedError: true,
		},
		{
			desc: "valid Host matcher",
			rule: "Host(`example.com`)",
			expected: map[string]int{
				"https://example.com":      http.StatusOK,
				"https://example.com:8080": http.StatusOK,
				"https://example.com/path": http.StatusOK,
				"https://EXAMPLE.COM/path": http.StatusOK,
				"https://example.org":      http.StatusNotFound,
				"https://example.org/path": http.StatusNotFound,
			},
		},
		{
			desc: "valid Host matcher - matcher ending with a dot",
			rule: "Host(`example.com.`)",
			expected: map[string]int{
				"https://example.com":       http.StatusOK,
				"https://example.com/path":  http.StatusOK,
				"https://example.org":       http.StatusNotFound,
				"https://example.org/path":  http.StatusNotFound,
				"https://example.com.":      http.StatusOK,
				"https://example.com./path": http.StatusOK,
				"https://example.org.":      http.StatusNotFound,
				"https://example.org./path": http.StatusNotFound,
			},
		},
		{
			desc: "valid Host matcher - URL ending with a dot",
			rule: "Host(`example.com`)",
			expected: map[string]int{
				"https://example.com.":      http.StatusOK,
				"https://example.com./path": http.StatusOK,
				"https://example.org.":      http.StatusNotFound,
				"https://example.org./path": http.StatusNotFound,
			},
		},
		{
			desc: "valid Host matcher - matcher with UPPER case",
			rule: "Host(`EXAMPLE.COM`)",
			expected: map[string]int{
				"https://example.com":      http.StatusOK,
				"https://example.com/path": http.StatusOK,
				"https://example.org":      http.StatusNotFound,
				"https://example.org/path": http.StatusNotFound,
			},
		},
		{
			desc: "valid Host matcher - puny-coded emoji",
			rule: "Host(`xn--9t9h.com`)",
			expected: map[string]int{
				"https://xn--9t9h.com":      http.StatusOK,
				"https://xn--9t9h.com/path": http.StatusOK,
				"https://example.com":       http.StatusNotFound,
				"https://example.com/path":  http.StatusNotFound,
				// The request's sender must use puny-code.
				"https://🦭.com": http.StatusNotFound,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			parser, err := NewSyntaxParser()
			require.NoError(t, err)

			muxer := NewMuxer(parser)

			err = muxer.AddRoute(test.rule, "", 0, handler)
			if test.expectedError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			// RequestDecorator is necessary for the Host matcher
			reqHost := requestdecorator.New(nil)

			results := make(map[string]int)
			for calledURL := range test.expected {
				w := httptest.NewRecorder()

				req := httptest.NewRequest(http.MethodGet, calledURL, http.NoBody)

				reqHost.ServeHTTP(w, req, muxer.ServeHTTP)
				results[calledURL] = w.Code
			}
			assert.Equal(t, test.expected, results)
		})
	}
}

func TestHostRegexpMatcher(t *testing.T) {
	testCases := []struct {
		desc          string
		rule          string
		expected      map[string]int
		expectedError bool
	}{
		{
			desc:          "invalid HostRegexp matcher (no parameter)",
			rule:          "HostRegexp()",
			expectedError: true,
		},
		{
			desc:          "invalid HostRegexp matcher (empty parameter)",
			rule:          "HostRegexp(``)",
			expectedError: true,
		},
		{
			desc:          "invalid HostRegexp matcher (non-ASCII)",
			rule:          "HostRegexp(`🦭.com`)",
			expectedError: true,
		},
		{
			desc:          "invalid HostRegexp matcher (invalid regexp)",
			rule:          "HostRegexp(`(example.com`)",
			expectedError: true,
		},
		{
			desc:          "invalid HostRegexp matcher (too many parameters)",
			rule:          "HostRegexp(`example.com`, `example.org`)",
			expectedError: true,
		},
		{
			desc: "valid HostRegexp matcher",
			rule: "HostRegexp(`^[a-zA-Z-]+\\.com$`)",
			expected: map[string]int{
				"https://example.com":      http.StatusOK,
				"https://example.com:8080": http.StatusOK,
				"https://example.com/path": http.StatusOK,
				"https://example.org":      http.StatusNotFound,
				"https://example.org/path": http.StatusNotFound,
			},
		},
		{
			desc: "valid HostRegexp matcher with case sensitive regexp",
			rule: "HostRegexp(`^[A-Z]+\\.com$`)",
			expected: map[string]int{
				"https://example.com":      http.StatusNotFound,
				"https://EXAMPLE.com":      http.StatusNotFound,
				"https://example.com/path": http.StatusNotFound,
				"https://example.org":      http.StatusNotFound,
				"https://example.org/path": http.StatusNotFound,
			},
		},
		{
			desc: "valid HostRegexp matcher with Traefik v2 syntax",
			rule: "HostRegexp(`{domain:[a-zA-Z-]+\\.com}`)",
			expected: map[string]int{
				"https://example.com":      http.StatusNotFound,
				"https://example.com/path": http.StatusNotFound,
				"https://example.org":      http.StatusNotFound,
				"https://example.org/path": http.StatusNotFound,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			parser, err := NewSyntaxParser()
			require.NoError(t, err)

			muxer := NewMuxer(parser)

			err = muxer.AddRoute(test.rule, "", 0, handler)
			if test.expectedError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// RequestDecorator is necessary for the HostRegexp matcher
			reqHost := requestdecorator.New(nil)

			results := make(map[string]int)
			for calledURL := range test.expected {
				w := httptest.NewRecorder()

				req := httptest.NewRequest(http.MethodGet, calledURL, http.NoBody)

				reqHost.ServeHTTP(w, req, muxer.ServeHTTP)
				results[calledURL] = w.Code
			}
			assert.Equal(t, test.expected, results)
		})
	}
}

func TestPathMatcher(t *testing.T) {
	testCases := []struct {
		desc          string
		rule          string
		expected      map[string]int
		expectedError bool
	}{
		{
			desc:          "invalid Path matcher (no parameter)",
			rule:          "Path()",
			expectedError: true,
		},
		{
			desc:          "invalid Path matcher (empty parameter)",
			rule:          "Path(``)",
			expectedError: true,
		},
		{
			desc:          "invalid Path matcher (no leading /)",
			rule:          "Path(`css`)",
			expectedError: true,
		},
		{
			desc:          "invalid Path matcher (too many parameters)",
			rule:          "Path(`/css`, `/js`)",
			expectedError: true,
		},
		{
			desc: "valid Path matcher",
			rule: "Path(`/css`)",
			expected: map[string]int{
				"https://example.com":              http.StatusNotFound,
				"https://example.com/html":         http.StatusNotFound,
				"https://example.org/css":          http.StatusOK,
				"https://example.com/css":          http.StatusOK,
				"https://example.com/css/":         http.StatusNotFound,
				"https://example.com/css/main.css": http.StatusNotFound,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			parser, err := NewSyntaxParser()
			require.NoError(t, err)

			muxer := NewMuxer(parser)

			err = muxer.AddRoute(test.rule, "", 0, handler)
			if test.expectedError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			results := make(map[string]int)
			for calledURL := range test.expected {
				w := httptest.NewRecorder()

				req := httptest.NewRequest(http.MethodGet, calledURL, http.NoBody)

				muxer.ServeHTTP(w, req)
				results[calledURL] = w.Code
			}
			assert.Equal(t, test.expected, results)
		})
	}
}

func TestPathRegexpMatcher(t *testing.T) {
	testCases := []struct {
		desc          string
		rule          string
		expected      map[string]int
		expectedError bool
	}{
		{
			desc:          "invalid PathRegexp matcher (no parameter)",
			rule:          "PathRegexp()",
			expectedError: true,
		},
		{
			desc:          "invalid PathRegexp matcher (empty parameter)",
			rule:          "PathRegexp(``)",
			expectedError: true,
		},
		{
			desc:          "invalid PathRegexp matcher (invalid regexp)",
			rule:          "PathRegexp(`/(css`)",
			expectedError: true,
		},
		{
			desc:          "invalid PathRegexp matcher (too many parameters)",
			rule:          "PathRegexp(`/css`, `/js`)",
			expectedError: true,
		},
		{
			desc: "valid PathRegexp matcher",
			rule: "PathRegexp(`^/(css|js)`)",
			expected: map[string]int{
				"https://example.com":              http.StatusNotFound,
				"https://example.com/html":         http.StatusNotFound,
				"https://example.org/css":          http.StatusOK,
				"https://example.com/CSS":          http.StatusNotFound,
				"https://example.com/css":          http.StatusOK,
				"https://example.com/css/":         http.StatusOK,
				"https://example.com/css/main.css": http.StatusOK,
				"https://example.com/js":           http.StatusOK,
				"https://example.com/js/":          http.StatusOK,
				"https://example.com/js/main.js":   http.StatusOK,
			},
		},
		{
			desc: "valid PathRegexp matcher with Traefik v2 syntax",
			rule: `PathRegexp("/{path:(css|js)}")`,
			expected: map[string]int{
				"https://example.com":                 http.StatusNotFound,
				"https://example.com/html":            http.StatusNotFound,
				"https://example.org/css":             http.StatusNotFound,
				"https://example.com/{path:css}":      http.StatusOK,
				"https://example.com/{path:css}/":     http.StatusOK,
				"https://example.com/%7Bpath:css%7D":  http.StatusOK,
				"https://example.com/%7Bpath:css%7D/": http.StatusOK,
				"https://example.com/{path:js}":       http.StatusOK,
				"https://example.com/{path:js}/":      http.StatusOK,
				"https://example.com/%7Bpath:js%7D":   http.StatusOK,
				"https://example.com/%7Bpath:js%7D/":  http.StatusOK,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			parser, err := NewSyntaxParser()
			require.NoError(t, err)

			muxer := NewMuxer(parser)

			err = muxer.AddRoute(test.rule, "", 0, handler)
			if test.expectedError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			results := make(map[string]int)
			for calledURL := range test.expected {
				w := httptest.NewRecorder()

				req := httptest.NewRequest(http.MethodGet, calledURL, http.NoBody)

				muxer.ServeHTTP(w, req)
				results[calledURL] = w.Code
			}
			assert.Equal(t, test.expected, results)
		})
	}
}

func TestPathPrefixMatcher(t *testing.T) {
	testCases := []struct {
		desc          string
		rule          string
		expected      map[string]int
		expectedError bool
	}{
		{
			desc:          "invalid PathPrefix matcher (no parameter)",
			rule:          "PathPrefix()",
			expectedError: true,
		},
		{
			desc:          "invalid PathPrefix matcher (empty parameter)",
			rule:          "PathPrefix(``)",
			expectedError: true,
		},
		{
			desc:          "invalid PathPrefix matcher (no leading /)",
			rule:          "PathPrefix(`css`)",
			expectedError: true,
		},
		{
			desc:          "invalid PathPrefix matcher (too many parameters)",
			rule:          "PathPrefix(`/css`, `/js`)",
			expectedError: true,
		},
		{
			desc: "valid PathPrefix matcher",
			rule: `PathPrefix("/css")`,
			expected: map[string]int{
				"https://example.com":              http.StatusNotFound,
				"https://example.com/html":         http.StatusNotFound,
				"https://example.org/css":          http.StatusOK,
				"https://example.com/css":          http.StatusOK,
				"https://example.com/css/":         http.StatusOK,
				"https://example.com/css/main.css": http.StatusOK,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			parser, err := NewSyntaxParser()
			require.NoError(t, err)

			muxer := NewMuxer(parser)

			err = muxer.AddRoute(test.rule, "", 0, handler)
			if test.expectedError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			results := make(map[string]int)
			for calledURL := range test.expected {
				w := httptest.NewRecorder()

				req := httptest.NewRequest(http.MethodGet, calledURL, http.NoBody)

				muxer.ServeHTTP(w, req)
				results[calledURL] = w.Code
			}
			assert.Equal(t, test.expected, results)
		})
	}
}

func TestHeaderMatcher(t *testing.T) {
	testCases := []struct {
		desc          string
		rule          string
		expected      map[*http.Header]int
		expectedError bool
	}{
		{
			desc:          "invalid Header matcher (no parameter)",
			rule:          "Header()",
			expectedError: true,
		},
		{
			desc:          "invalid Header matcher (missing value parameter)",
			rule:          "Header(`X-Forwarded-Host`)",
			expectedError: true,
		},
		{
			desc:          "invalid Header matcher (missing value parameter)",
			rule:          "Header(`X-Forwarded-Host`, ``)",
			expectedError: true,
		},
		{
			desc:          "invalid Header matcher (missing key parameter)",
			rule:          "Header(``, `example.com`)",
			expectedError: true,
		},
		{
			desc:          "invalid Header matcher (too many parameters)",
			rule:          "Header(`X-Forwarded-Host`, `example.com`, `example.org`)",
			expectedError: true,
		},
		{
			desc: "valid Header matcher",
			rule: "Header(`X-Forwarded-Proto`, `https`)",
			expected: map[*http.Header]int{
				{"X-Forwarded-Proto": []string{"https"}}:         http.StatusOK,
				{"x-forwarded-proto": []string{"https"}}:         http.StatusNotFound,
				{"X-Forwarded-Proto": []string{"http", "https"}}: http.StatusOK,
				{"X-Forwarded-Proto": []string{"https", "http"}}: http.StatusOK,
				{"X-Forwarded-Host": []string{"example.com"}}:    http.StatusNotFound,
			},
		},
		{
			desc: "valid Header matcher (non-canonical form)",
			rule: "Header(`x-forwarded-proto`, `https`)",
			expected: map[*http.Header]int{
				{"X-Forwarded-Proto": []string{"https"}}:         http.StatusOK,
				{"x-forwarded-proto": []string{"https"}}:         http.StatusNotFound,
				{"X-Forwarded-Proto": []string{"http", "https"}}: http.StatusOK,
				{"X-Forwarded-Proto": []string{"https", "http"}}: http.StatusOK,
				{"X-Forwarded-Host": []string{"example.com"}}:    http.StatusNotFound,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			parser, err := NewSyntaxParser()
			require.NoError(t, err)

			muxer := NewMuxer(parser)

			err = muxer.AddRoute(test.rule, "", 0, handler)
			if test.expectedError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			for headers := range test.expected {
				w := httptest.NewRecorder()

				req := httptest.NewRequest(http.MethodGet, "https://example.com", http.NoBody)
				req.Header = *headers

				muxer.ServeHTTP(w, req)
				assert.Equal(t, test.expected[headers], w.Code, headers)
			}
		})
	}
}

func TestHeaderRegexpMatcher(t *testing.T) {
	testCases := []struct {
		desc          string
		rule          string
		expected      map[*http.Header]int
		expectedError bool
	}{
		{
			desc:          "invalid HeaderRegexp matcher (no parameter)",
			rule:          "HeaderRegexp()",
			expectedError: true,
		},
		{
			desc:          "invalid HeaderRegexp matcher (missing value parameter)",
			rule:          "HeaderRegexp(`X-Forwarded-Host`)",
			expectedError: true,
		},
		{
			desc:          "invalid HeaderRegexp matcher (missing value parameter)",
			rule:          "HeaderRegexp(`X-Forwarded-Host`, ``)",
			expectedError: true,
		},
		{
			desc:          "invalid HeaderRegexp matcher (missing key parameter)",
			rule:          "HeaderRegexp(``, `example.com`)",
			expectedError: true,
		},
		{
			desc:          "invalid HeaderRegexp matcher (invalid regexp)",
			rule:          "HeaderRegexp(`X-Forwarded-Host`,`(example.com`)",
			expectedError: true,
		},
		{
			desc:          "invalid HeaderRegexp matcher (too many parameters)",
			rule:          "HeaderRegexp(`X-Forwarded-Host`, `example.com`, `example.org`)",
			expectedError: true,
		},
		{
			desc: "valid HeaderRegexp matcher",
			rule: "HeaderRegexp(`X-Forwarded-Proto`, `^https?$`)",
			expected: map[*http.Header]int{
				{"X-Forwarded-Proto": []string{"http"}}:        http.StatusOK,
				{"x-forwarded-proto": []string{"http"}}:        http.StatusNotFound,
				{"X-Forwarded-Proto": []string{"https"}}:       http.StatusOK,
				{"X-Forwarded-Proto": []string{"HTTPS"}}:       http.StatusNotFound,
				{"X-Forwarded-Proto": []string{"ws", "https"}}: http.StatusOK,
				{"X-Forwarded-Host": []string{"example.com"}}:  http.StatusNotFound,
			},
		},
		{
			desc: "valid HeaderRegexp matcher (non-canonical form)",
			rule: "HeaderRegexp(`x-forwarded-proto`, `^https?$`)",
			expected: map[*http.Header]int{
				{"X-Forwarded-Proto": []string{"http"}}:        http.StatusOK,
				{"x-forwarded-proto": []string{"http"}}:        http.StatusNotFound,
				{"X-Forwarded-Proto": []string{"https"}}:       http.StatusOK,
				{"X-Forwarded-Proto": []string{"HTTPS"}}:       http.StatusNotFound,
				{"X-Forwarded-Proto": []string{"ws", "https"}}: http.StatusOK,
				{"X-Forwarded-Host": []string{"example.com"}}:  http.StatusNotFound,
			},
		},
		{
			desc: "valid HeaderRegexp matcher with Traefik v2 syntax",
			rule: "HeaderRegexp(`X-Forwarded-Proto`, `http{secure:s?}`)",
			expected: map[*http.Header]int{
				{"X-Forwarded-Proto": []string{"http"}}:                 http.StatusNotFound,
				{"X-Forwarded-Proto": []string{"https"}}:                http.StatusNotFound,
				{"X-Forwarded-Proto": []string{"http{secure:}"}}:        http.StatusOK,
				{"X-Forwarded-Proto": []string{"HTTP{secure:}"}}:        http.StatusNotFound,
				{"X-Forwarded-Proto": []string{"http{secure:s}"}}:       http.StatusOK,
				{"X-Forwarded-Proto": []string{"http{secure:S}"}}:       http.StatusNotFound,
				{"X-Forwarded-Proto": []string{"HTTPS"}}:                http.StatusNotFound,
				{"X-Forwarded-Proto": []string{"ws", "http{secure:s}"}}: http.StatusOK,
				{"X-Forwarded-Host": []string{"example.com"}}:           http.StatusNotFound,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			parser, err := NewSyntaxParser()
			require.NoError(t, err)

			muxer := NewMuxer(parser)

			err = muxer.AddRoute(test.rule, "", 0, handler)
			if test.expectedError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			for headers := range test.expected {
				w := httptest.NewRecorder()

				req := httptest.NewRequest(http.MethodGet, "https://example.com", http.NoBody)
				req.Header = *headers

				muxer.ServeHTTP(w, req)
				assert.Equal(t, test.expected[headers], w.Code, *headers)
			}
		})
	}
}

func TestQueryMatcher(t *testing.T) {
	testCases := []struct {
		desc          string
		rule          string
		expected      map[string]int
		expectedError bool
	}{
		{
			desc:          "invalid Query matcher (no parameter)",
			rule:          "Query()",
			expectedError: true,
		},
		{
			desc:          "invalid Query matcher (empty key, one parameter)",
			rule:          "Query(``)",
			expectedError: true,
		},
		{
			desc:          "invalid Query matcher (empty key)",
			rule:          "Query(``, `traefik`)",
			expectedError: true,
		},
		{
			desc:          "invalid Query matcher (empty value)",
			rule:          "Query(`q`, ``)",
			expectedError: true,
		},
		{
			desc:          "invalid Query matcher (too many parameters)",
			rule:          "Query(`q`, `traefik`, `proxy`)",
			expectedError: true,
		},
		{
			desc: "valid Query matcher",
			rule: "Query(`q`, `traefik`)",
			expected: map[string]int{
				"https://example.com":                     http.StatusNotFound,
				"https://example.com?q=traefik":           http.StatusOK,
				"https://example.com?rel=ddg&q=traefik":   http.StatusOK,
				"https://example.com?q=traefik&q=proxy":   http.StatusOK,
				"https://example.com?q=awesome&q=traefik": http.StatusOK,
				"https://example.com?q=nginx":             http.StatusNotFound,
				"https://example.com?rel=ddg":             http.StatusNotFound,
				"https://example.com?q=TRAEFIK":           http.StatusNotFound,
				"https://example.com?Q=traefik":           http.StatusNotFound,
				"https://example.com?rel=traefik":         http.StatusNotFound,
			},
		},
		{
			desc: "valid Query matcher with empty value",
			rule: "Query(`mobile`)",
			expected: map[string]int{
				"https://example.com":             http.StatusNotFound,
				"https://example.com?mobile":      http.StatusOK,
				"https://example.com?mobile=true": http.StatusNotFound,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			parser, err := NewSyntaxParser()
			require.NoError(t, err)

			muxer := NewMuxer(parser)

			err = muxer.AddRoute(test.rule, "", 0, handler)
			if test.expectedError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			results := make(map[string]int)
			for calledURL := range test.expected {
				w := httptest.NewRecorder()

				req := httptest.NewRequest(http.MethodGet, calledURL, http.NoBody)

				muxer.ServeHTTP(w, req)
				results[calledURL] = w.Code
			}
			assert.Equal(t, test.expected, results)
		})
	}
}

func TestQueryRegexpMatcher(t *testing.T) {
	testCases := []struct {
		desc          string
		rule          string
		expected      map[string]int
		expectedError bool
	}{
		{
			desc:          "invalid QueryRegexp matcher (no parameter)",
			rule:          "QueryRegexp()",
			expectedError: true,
		},
		{
			desc:          "invalid QueryRegexp matcher (empty parameter)",
			rule:          "QueryRegexp(``)",
			expectedError: true,
		},
		{
			desc:          "invalid QueryRegexp matcher (invalid regexp)",
			rule:          "QueryRegexp(`q`, `(traefik`)",
			expectedError: true,
		},
		{
			desc:          "invalid QueryRegexp matcher (too many parameters)",
			rule:          "QueryRegexp(`q`, `traefik`, `proxy`)",
			expectedError: true,
		},
		{
			desc: "valid QueryRegexp matcher",
			rule: "QueryRegexp(`q`, `^(traefik|nginx)$`)",
			expected: map[string]int{
				"https://example.com":                     http.StatusNotFound,
				"https://example.com?q=traefik":           http.StatusOK,
				"https://example.com?rel=ddg&q=traefik":   http.StatusOK,
				"https://example.com?q=traefik&q=proxy":   http.StatusOK,
				"https://example.com?q=awesome&q=traefik": http.StatusOK,
				"https://example.com?q=TRAEFIK":           http.StatusNotFound,
				"https://example.com?Q=traefik":           http.StatusNotFound,
				"https://example.com?rel=traefik":         http.StatusNotFound,
				"https://example.com?q=nginx":             http.StatusOK,
				"https://example.com?rel=ddg&q=nginx":     http.StatusOK,
				"https://example.com?q=nginx&q=proxy":     http.StatusOK,
				"https://example.com?q=awesome&q=nginx":   http.StatusOK,
				"https://example.com?q=NGINX":             http.StatusNotFound,
				"https://example.com?Q=nginx":             http.StatusNotFound,
				"https://example.com?rel=nginx":           http.StatusNotFound,
				"https://example.com?q=haproxy":           http.StatusNotFound,
				"https://example.com?rel=ddg":             http.StatusNotFound,
			},
		},
		{
			desc: "valid QueryRegexp matcher",
			rule: "QueryRegexp(`q`, `^.*$`)",
			expected: map[string]int{
				"https://example.com":                     http.StatusNotFound,
				"https://example.com?q=traefik":           http.StatusOK,
				"https://example.com?rel=ddg&q=traefik":   http.StatusOK,
				"https://example.com?q=traefik&q=proxy":   http.StatusOK,
				"https://example.com?q=awesome&q=traefik": http.StatusOK,
				"https://example.com?q=TRAEFIK":           http.StatusOK,
				"https://example.com?Q=traefik":           http.StatusNotFound,
				"https://example.com?rel=traefik":         http.StatusNotFound,
				"https://example.com?q=nginx":             http.StatusOK,
				"https://example.com?rel=ddg&q=nginx":     http.StatusOK,
				"https://example.com?q=nginx&q=proxy":     http.StatusOK,
				"https://example.com?q=awesome&q=nginx":   http.StatusOK,
				"https://example.com?q=NGINX":             http.StatusOK,
				"https://example.com?Q=nginx":             http.StatusNotFound,
				"https://example.com?rel=nginx":           http.StatusNotFound,
				"https://example.com?q=haproxy":           http.StatusOK,
				"https://example.com?rel=ddg":             http.StatusNotFound,
			},
		},
		{
			desc: "valid QueryRegexp matcher with Traefik v2 syntax",
			rule: "QueryRegexp(`q`, `{value:(traefik|nginx)}`)",
			expected: map[string]int{
				"https://example.com?q=traefik":         http.StatusNotFound,
				"https://example.com?q={value:traefik}": http.StatusOK,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			parser, err := NewSyntaxParser()
			require.NoError(t, err)

			muxer := NewMuxer(parser)

			err = muxer.AddRoute(test.rule, "", 0, handler)
			if test.expectedError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			results := make(map[string]int)
			for calledURL := range test.expected {
				w := httptest.NewRecorder()

				req := httptest.NewRequest(http.MethodGet, calledURL, http.NoBody)

				muxer.ServeHTTP(w, req)
				results[calledURL] = w.Code
			}
			assert.Equal(t, test.expected, results)
		})
	}
}
