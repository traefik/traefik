package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/middlewares/requestdecorator"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
)

func TestClientIPV2Matcher(t *testing.T) {
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
			desc: "valid ClientIP matcher (many parameters)",
			rule: "ClientIP(`127.0.0.1`, `192.168.1.0/24`)",
			expected: map[string]int{
				"127.0.0.1":   http.StatusOK,
				"192.168.1.1": http.StatusOK,
			},
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

			err = muxer.AddRoute(test.rule, "v2", 0, handler)
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

func TestMethodV2Matcher(t *testing.T) {
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
			desc: "valid Method matcher (many parameters)",
			rule: "Method(`GET`, `POST`)",
			expected: map[string]int{
				http.MethodGet:  http.StatusOK,
				http.MethodPost: http.StatusOK,
			},
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

			err = muxer.AddRoute(test.rule, "v2", 0, handler)
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

func TestHostV2Matcher(t *testing.T) {
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
			desc: "valid Host matcher (many parameters)",
			rule: "Host(`example.com`, `example.org`)",
			expected: map[string]int{
				"https://example.com":      http.StatusOK,
				"https://example.com:8080": http.StatusOK,
				"https://example.com/path": http.StatusOK,
				"https://EXAMPLE.COM/path": http.StatusOK,
				"https://example.org":      http.StatusOK,
				"https://example.org/path": http.StatusOK,
			},
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

			err = muxer.AddRoute(test.rule, "v2", 0, handler)
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

func TestHostRegexpV2Matcher(t *testing.T) {
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
			desc: "valid HostRegexp matcher (invalid regexp)",
			rule: "HostRegexp(`(example.com`)",
			// This is weird.
			expectedError: false,
			expected: map[string]int{
				"https://example.com":      http.StatusNotFound,
				"https://example.com:8080": http.StatusNotFound,
				"https://example.com/path": http.StatusNotFound,
				"https://example.org":      http.StatusNotFound,
				"https://example.org/path": http.StatusNotFound,
			},
		},
		{
			desc: "valid HostRegexp matcher (many parameters)",
			rule: "HostRegexp(`example.com`, `example.org`)",
			expected: map[string]int{
				"https://example.com":      http.StatusOK,
				"https://example.com:8080": http.StatusOK,
				"https://example.com/path": http.StatusOK,
				"https://example.org":      http.StatusOK,
				"https://example.org/path": http.StatusOK,
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
				"https://example.com":      http.StatusOK,
				"https://example.com/path": http.StatusOK,
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

			err = muxer.AddRoute(test.rule, "v2", 0, handler)
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

func TestPathV2Matcher(t *testing.T) {
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
			desc: "valid Path matcher (many parameters)",
			rule: "Path(`/css`, `/js`)",
			expected: map[string]int{
				"https://example.com":              http.StatusNotFound,
				"https://example.com/html":         http.StatusNotFound,
				"https://example.org/css":          http.StatusOK,
				"https://example.com/css":          http.StatusOK,
				"https://example.com/css/":         http.StatusNotFound,
				"https://example.com/css/main.css": http.StatusNotFound,
				"https://example.com/js":           http.StatusOK,
				"https://example.com/js/main.js":   http.StatusNotFound,
			},
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
		{
			desc: "valid Path matcher with regexp",
			rule: "Path(`/css{path:(/.*)?}`)",
			expected: map[string]int{
				"https://example.com":                              http.StatusNotFound,
				"https://example.com/css/main.css":                 http.StatusOK,
				"https://example.org/css/main.css":                 http.StatusOK,
				"https://example.com/css/components/component.css": http.StatusOK,
				"https://example.com/css.css":                      http.StatusNotFound,
				"https://example.com/js/main.js":                   http.StatusNotFound,
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

			err = muxer.AddRoute(test.rule, "v2", 0, handler)
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

func TestPathPrefixV2Matcher(t *testing.T) {
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
			desc: "valid PathPrefix matcher (many parameters)",
			rule: "PathPrefix(`/css`, `/js`)",
			expected: map[string]int{
				"https://example.com":              http.StatusNotFound,
				"https://example.com/html":         http.StatusNotFound,
				"https://example.org/css":          http.StatusOK,
				"https://example.com/css":          http.StatusOK,
				"https://example.com/css/":         http.StatusOK,
				"https://example.com/css/main.css": http.StatusOK,
				"https://example.com/js/":          http.StatusOK,
				"https://example.com/js/main.js":   http.StatusOK,
			},
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
		{
			desc: "valid PathPrefix matcher with regexp",
			rule: "PathPrefix(`/css-{name:[0-9]?}`)",
			expected: map[string]int{
				"https://example.com":                                     http.StatusNotFound,
				"https://example.com/css-1/main.css":                      http.StatusOK,
				"https://example.org/css-222/main.css":                    http.StatusOK,
				"https://example.com/css-333333/components/component.css": http.StatusOK,
				"https://example.com/css.css":                             http.StatusNotFound,
				"https://example.com/js/main.js":                          http.StatusNotFound,
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

			err = muxer.AddRoute(test.rule, "v2", 0, handler)
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

func TestHeadersMatcher(t *testing.T) {
	testCases := []struct {
		desc          string
		rule          string
		expected      map[*http.Header]int
		expectedError bool
	}{
		{
			desc:          "invalid Header matcher (no parameter)",
			rule:          "Headers()",
			expectedError: true,
		},
		{
			desc:          "invalid Header matcher (missing value parameter)",
			rule:          "Headers(`X-Forwarded-Host`)",
			expectedError: true,
		},
		{
			desc:          "invalid Header matcher (missing value parameter)",
			rule:          "Headers(`X-Forwarded-Host`, ``)",
			expectedError: true,
		},
		{
			desc:          "invalid Header matcher (missing key parameter)",
			rule:          "Headers(``, `example.com`)",
			expectedError: true,
		},
		{
			desc:          "invalid Header matcher (too many parameters)",
			rule:          "Headers(`X-Forwarded-Host`, `example.com`, `example.org`)",
			expectedError: true,
		},
		{
			desc: "valid Header matcher",
			rule: "Headers(`X-Forwarded-Proto`, `https`)",
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
			rule: "Headers(`x-forwarded-proto`, `https`)",
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

			err = muxer.AddRoute(test.rule, "v2", 0, handler)
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

func TestHeaderRegexpV2Matcher(t *testing.T) {
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
			rule:          "HeadersRegexp(`X-Forwarded-Host`)",
			expectedError: true,
		},
		{
			desc:          "invalid HeaderRegexp matcher (missing value parameter)",
			rule:          "HeadersRegexp(`X-Forwarded-Host`, ``)",
			expectedError: true,
		},
		{
			desc:          "invalid HeaderRegexp matcher (missing key parameter)",
			rule:          "HeadersRegexp(``, `example.com`)",
			expectedError: true,
		},
		{
			desc:          "invalid HeaderRegexp matcher (invalid regexp)",
			rule:          "HeadersRegexp(`X-Forwarded-Host`,`(example.com`)",
			expectedError: true,
		},
		{
			desc:          "invalid HeaderRegexp matcher (too many parameters)",
			rule:          "HeadersRegexp(`X-Forwarded-Host`, `example.com`, `example.org`)",
			expectedError: true,
		},
		{
			desc: "valid HeaderRegexp matcher",
			rule: "HeadersRegexp(`X-Forwarded-Proto`, `^https?$`)",
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
			rule: "HeadersRegexp(`x-forwarded-proto`, `^https?$`)",
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
			rule: "HeadersRegexp(`X-Forwarded-Proto`, `http{secure:s?}`)",
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

			err = muxer.AddRoute(test.rule, "v2", 0, handler)
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

func TestHostRegexp(t *testing.T) {
	testCases := []struct {
		desc    string
		hostExp string
		urls    map[string]int
	}{
		{
			desc:    "capturing group",
			hostExp: "HostRegexp(`{subdomain:(foo\\.)?bar\\.com}`)",
			urls: map[string]int{
				"http://foo.bar.com": http.StatusOK,
				"http://bar.com":     http.StatusOK,
				"http://fooubar.com": http.StatusNotFound,
				"http://barucom":     http.StatusNotFound,
				"http://barcom":      http.StatusNotFound,
			},
		},
		{
			desc:    "non capturing group",
			hostExp: "HostRegexp(`{subdomain:(?:foo\\.)?bar\\.com}`)",
			urls: map[string]int{
				"http://foo.bar.com": http.StatusOK,
				"http://bar.com":     http.StatusOK,
				"http://fooubar.com": http.StatusNotFound,
				"http://barucom":     http.StatusNotFound,
				"http://barcom":      http.StatusNotFound,
			},
		},
		{
			desc:    "regex insensitive",
			hostExp: "HostRegexp(`{dummy:[A-Za-z-]+\\.bar\\.com}`)",
			urls: map[string]int{
				"http://FOO.bar.com": http.StatusOK,
				"http://foo.bar.com": http.StatusOK,
				"http://fooubar.com": http.StatusNotFound,
				"http://barucom":     http.StatusNotFound,
				"http://barcom":      http.StatusNotFound,
			},
		},
		{
			desc:    "insensitive host",
			hostExp: "HostRegexp(`{dummy:[a-z-]+\\.bar\\.com}`)",
			urls: map[string]int{
				"http://FOO.bar.com": http.StatusOK,
				"http://foo.bar.com": http.StatusOK,
				"http://fooubar.com": http.StatusNotFound,
				"http://barucom":     http.StatusNotFound,
				"http://barcom":      http.StatusNotFound,
			},
		},
		{
			desc:    "insensitive host simple",
			hostExp: "HostRegexp(`foo.bar.com`)",
			urls: map[string]int{
				"http://FOO.bar.com": http.StatusOK,
				"http://foo.bar.com": http.StatusOK,
				"http://fooubar.com": http.StatusNotFound,
				"http://barucom":     http.StatusNotFound,
				"http://barcom":      http.StatusNotFound,
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

			err = muxer.AddRoute(test.hostExp, "v2", 0, handler)
			require.NoError(t, err)

			results := make(map[string]int)
			for calledURL := range test.urls {
				w := httptest.NewRecorder()

				req := httptest.NewRequest(http.MethodGet, calledURL, http.NoBody)

				muxer.ServeHTTP(w, req)
				results[calledURL] = w.Code
			}
			assert.Equal(t, test.urls, results)
		})
	}
}

// This test is a copy from the v2 branch mux_test.go file.
func Test_addRoute(t *testing.T) {
	testCases := []struct {
		desc          string
		rule          string
		headers       map[string]string
		remoteAddr    string
		expected      map[string]int
		expectedError bool
	}{
		{
			desc:          "no tree",
			expectedError: true,
		},
		{
			desc:          "Rule with no matcher",
			rule:          "rulewithnotmatcher",
			expectedError: true,
		},
		{
			desc:          "Host empty",
			rule:          "Host(``)",
			expectedError: true,
		},
		{
			desc:          "PathPrefix empty",
			rule:          "PathPrefix(``)",
			expectedError: true,
		},
		{
			desc: "PathPrefix",
			rule: "PathPrefix(`/foo`)",
			expected: map[string]int{
				"http://localhost/foo": http.StatusOK,
			},
		},
		{
			desc: "wrong PathPrefix",
			rule: "PathPrefix(`/bar`)",
			expected: map[string]int{
				"http://localhost/foo": http.StatusNotFound,
			},
		},
		{
			desc: "Host",
			rule: "Host(`localhost`)",
			expected: map[string]int{
				"http://localhost/foo": http.StatusOK,
			},
		},
		{
			desc: "Host IPv4",
			rule: "Host(`127.0.0.1`)",
			expected: map[string]int{
				"http://127.0.0.1/foo": http.StatusOK,
			},
		},
		{
			desc: "Host IPv6",
			rule: "Host(`10::10`)",
			expected: map[string]int{
				"http://10::10/foo": http.StatusOK,
			},
		},
		{
			desc:          "Non-ASCII Host",
			rule:          "Host(`locàlhost`)",
			expectedError: true,
		},
		{
			desc:          "Non-ASCII HostRegexp",
			rule:          "HostRegexp(`locàlhost`)",
			expectedError: true,
		},
		{
			desc: "HostHeader equivalent to Host",
			rule: "HostHeader(`localhost`)",
			expected: map[string]int{
				"http://localhost/foo": http.StatusOK,
				"http://bar/foo":       http.StatusNotFound,
			},
		},
		{
			desc: "Host with trailing period in rule",
			rule: "Host(`localhost.`)",
			expected: map[string]int{
				"http://localhost/foo": http.StatusOK,
			},
		},
		{
			desc: "Host with trailing period in domain",
			rule: "Host(`localhost`)",
			expected: map[string]int{
				"http://localhost./foo": http.StatusOK,
			},
		},
		{
			desc: "Host with trailing period in domain and rule",
			rule: "Host(`localhost.`)",
			expected: map[string]int{
				"http://localhost./foo": http.StatusOK,
			},
		},
		{
			desc: "wrong Host",
			rule: "Host(`nope`)",
			expected: map[string]int{
				"http://localhost/foo": http.StatusNotFound,
			},
		},
		{
			desc: "Host and PathPrefix",
			rule: "Host(`localhost`) && PathPrefix(`/foo`)",
			expected: map[string]int{
				"http://localhost/foo": http.StatusOK,
			},
		},
		{
			desc: "Host and PathPrefix wrong PathPrefix",
			rule: "Host(`localhost`) && PathPrefix(`/bar`)",
			expected: map[string]int{
				"http://localhost/foo": http.StatusNotFound,
			},
		},
		{
			desc: "Host and PathPrefix wrong Host",
			rule: "Host(`nope`) && PathPrefix(`/foo`)",
			expected: map[string]int{
				"http://localhost/foo": http.StatusNotFound,
			},
		},
		{
			desc: "Host and PathPrefix Host OR, first host",
			rule: "Host(`nope`,`localhost`) && PathPrefix(`/foo`)",
			expected: map[string]int{
				"http://localhost/foo": http.StatusOK,
			},
		},
		{
			desc: "Host and PathPrefix Host OR, second host",
			rule: "Host(`nope`,`localhost`) && PathPrefix(`/foo`)",
			expected: map[string]int{
				"http://nope/foo": http.StatusOK,
			},
		},
		{
			desc: "Host and PathPrefix Host OR, first host and wrong PathPrefix",
			rule: "Host(`nope,localhost`) && PathPrefix(`/bar`)",
			expected: map[string]int{
				"http://localhost/foo": http.StatusNotFound,
			},
		},
		{
			desc: "HostRegexp with capturing group",
			rule: "HostRegexp(`{subdomain:(foo\\.)?bar\\.com}`)",
			expected: map[string]int{
				"http://foo.bar.com": http.StatusOK,
				"http://bar.com":     http.StatusOK,
				"http://fooubar.com": http.StatusNotFound,
				"http://barucom":     http.StatusNotFound,
				"http://barcom":      http.StatusNotFound,
			},
		},
		{
			desc: "HostRegexp with non capturing group",
			rule: "HostRegexp(`{subdomain:(?:foo\\.)?bar\\.com}`)",
			expected: map[string]int{
				"http://foo.bar.com": http.StatusOK,
				"http://bar.com":     http.StatusOK,
				"http://fooubar.com": http.StatusNotFound,
				"http://barucom":     http.StatusNotFound,
				"http://barcom":      http.StatusNotFound,
			},
		},
		{
			desc: "Methods with GET",
			rule: "Method(`GET`)",
			expected: map[string]int{
				"http://localhost/foo": http.StatusOK,
			},
		},
		{
			desc: "Methods with GET and POST",
			rule: "Method(`GET`,`POST`)",
			expected: map[string]int{
				"http://localhost/foo": http.StatusOK,
			},
		},
		{
			desc: "Methods with POST",
			rule: "Method(`POST`)",
			expected: map[string]int{
				// On v2 this test expect a http.StatusMethodNotAllowed status code.
				// This was due to a custom behavior of mux https://github.com/containous/mux/blob/b2dd784e613f218225150a5e8b5742c5733bc1b6/mux.go#L130-L132.
				// Unfortunately, this behavior cannot be ported easily due to the matcher func signature.
				"http://localhost/foo": http.StatusNotFound,
			},
		},
		{
			desc: "Header with matching header",
			rule: "Headers(`Content-Type`,`application/json`)",
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			expected: map[string]int{
				"http://localhost/foo": http.StatusOK,
			},
		},
		{
			desc: "Header without matching header",
			rule: "Headers(`Content-Type`,`application/foo`)",
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			expected: map[string]int{
				"http://localhost/foo": http.StatusNotFound,
			},
		},
		{
			desc: "HeaderRegExp with matching header",
			rule: "HeadersRegexp(`Content-Type`, `application/(text|json)`)",
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			expected: map[string]int{
				"http://localhost/foo": http.StatusOK,
			},
		},
		{
			desc: "HeaderRegExp without matching header",
			rule: "HeadersRegexp(`Content-Type`, `application/(text|json)`)",
			headers: map[string]string{
				"Content-Type": "application/foo",
			},
			expected: map[string]int{
				"http://localhost/foo": http.StatusNotFound,
			},
		},
		{
			desc: "HeaderRegExp with matching second header",
			rule: "HeadersRegexp(`Content-Type`, `application/(text|json)`)",
			headers: map[string]string{
				"Content-Type": "application/text",
			},
			expected: map[string]int{
				"http://localhost/foo": http.StatusOK,
			},
		},
		{
			desc: "Query with multiple params",
			rule: "Query(`foo=bar`, `bar=baz`)",
			expected: map[string]int{
				"http://localhost/foo?foo=bar&bar=baz": http.StatusOK,
				"http://localhost/foo?bar=baz":         http.StatusNotFound,
			},
		},
		{
			desc: "Query with multiple equals",
			rule: "Query(`foo=b=ar`)",
			expected: map[string]int{
				"http://localhost/foo?foo=b=ar": http.StatusOK,
				"http://localhost/foo?foo=bar":  http.StatusNotFound,
			},
		},
		{
			desc: "Rule with simple path",
			rule: `Path("/a")`,
			expected: map[string]int{
				"http://plop/a": http.StatusOK,
			},
		},
		{
			desc: `Rule with a simple host`,
			rule: `Host("plop")`,
			expected: map[string]int{
				"http://plop": http.StatusOK,
			},
		},
		{
			desc: "Rule with Path AND Host",
			rule: `Path("/a") && Host("plop")`,
			expected: map[string]int{
				"http://plop/a":  http.StatusOK,
				"http://plopi/a": http.StatusNotFound,
			},
		},
		{
			desc: "Rule with Host OR Host",
			rule: `Host("tchouk") || Host("pouet")`,
			expected: map[string]int{
				"http://tchouk/toto": http.StatusOK,
				"http://pouet/a":     http.StatusOK,
				"http://plopi/a":     http.StatusNotFound,
			},
		},
		{
			desc: "Rule with host OR (host AND path)",
			rule: `Host("tchouk") || (Host("pouet") && Path("/powpow"))`,
			expected: map[string]int{
				"http://tchouk/toto":   http.StatusOK,
				"http://tchouk/powpow": http.StatusOK,
				"http://pouet/powpow":  http.StatusOK,
				"http://pouet/toto":    http.StatusNotFound,
				"http://plopi/a":       http.StatusNotFound,
			},
		},
		{
			desc: "Rule with host OR host AND path",
			rule: `Host("tchouk") || Host("pouet") && Path("/powpow")`,
			expected: map[string]int{
				"http://tchouk/toto":   http.StatusOK,
				"http://tchouk/powpow": http.StatusOK,
				"http://pouet/powpow":  http.StatusOK,
				"http://pouet/toto":    http.StatusNotFound,
				"http://plopi/a":       http.StatusNotFound,
			},
		},
		{
			desc: "Rule with (host OR host) AND path",
			rule: `(Host("tchouk") || Host("pouet")) && Path("/powpow")`,
			expected: map[string]int{
				"http://tchouk/toto":   http.StatusNotFound,
				"http://tchouk/powpow": http.StatusOK,
				"http://pouet/powpow":  http.StatusOK,
				"http://pouet/toto":    http.StatusNotFound,
				"http://plopi/a":       http.StatusNotFound,
			},
		},
		{
			desc: "Rule with multiple host AND path",
			rule: `(Host("tchouk","pouet")) && Path("/powpow")`,
			expected: map[string]int{
				"http://tchouk/toto":   http.StatusNotFound,
				"http://tchouk/powpow": http.StatusOK,
				"http://pouet/powpow":  http.StatusOK,
				"http://pouet/toto":    http.StatusNotFound,
				"http://plopi/a":       http.StatusNotFound,
			},
		},
		{
			desc: "Rule with multiple host AND multiple path",
			rule: `Host("tchouk","pouet") && Path("/powpow", "/titi")`,
			expected: map[string]int{
				"http://tchouk/toto":   http.StatusNotFound,
				"http://tchouk/powpow": http.StatusOK,
				"http://pouet/powpow":  http.StatusOK,
				"http://tchouk/titi":   http.StatusOK,
				"http://pouet/titi":    http.StatusOK,
				"http://pouet/toto":    http.StatusNotFound,
				"http://plopi/a":       http.StatusNotFound,
			},
		},
		{
			desc: "Rule with (host AND path) OR (host AND path)",
			rule: `(Host("tchouk") && Path("/titi")) || ((Host("pouet")) && Path("/powpow"))`,
			expected: map[string]int{
				"http://tchouk/titi":   http.StatusOK,
				"http://tchouk/powpow": http.StatusNotFound,
				"http://pouet/powpow":  http.StatusOK,
				"http://pouet/toto":    http.StatusNotFound,
				"http://plopi/a":       http.StatusNotFound,
			},
		},
		{
			desc:          "Rule without quote",
			rule:          `Host(tchouk)`,
			expectedError: true,
		},
		{
			desc: "Rule case UPPER",
			rule: `(HOST("tchouk") && PATHPREFIX("/titi"))`,
			expected: map[string]int{
				"http://tchouk/titi":   http.StatusOK,
				"http://tchouk/powpow": http.StatusNotFound,
			},
		},
		{
			desc: "Rule case lower",
			rule: `(host("tchouk") && pathprefix("/titi"))`,
			expected: map[string]int{
				"http://tchouk/titi":   http.StatusOK,
				"http://tchouk/powpow": http.StatusNotFound,
			},
		},
		{
			desc: "Rule case CamelCase",
			rule: `(Host("tchouk") && PathPrefix("/titi"))`,
			expected: map[string]int{
				"http://tchouk/titi":   http.StatusOK,
				"http://tchouk/powpow": http.StatusNotFound,
			},
		},
		{
			desc: "Rule case Title",
			rule: `(Host("tchouk") && Pathprefix("/titi"))`,
			expected: map[string]int{
				"http://tchouk/titi":   http.StatusOK,
				"http://tchouk/powpow": http.StatusNotFound,
			},
		},
		{
			desc:          "Rule Path with error",
			rule:          `Path("titi")`,
			expectedError: true,
		},
		{
			desc:          "Rule PathPrefix with error",
			rule:          `PathPrefix("titi")`,
			expectedError: true,
		},
		{
			desc:          "Rule HostRegexp with error",
			rule:          `HostRegexp("{test")`,
			expectedError: true,
		},
		{
			desc:          "Rule Headers with error",
			rule:          `Headers("titi")`,
			expectedError: true,
		},
		{
			desc:          "Rule HeadersRegexp with error",
			rule:          `HeadersRegexp("titi")`,
			expectedError: true,
		},
		{
			desc:          "Rule Query",
			rule:          `Query("titi")`,
			expectedError: true,
		},
		{
			desc:          "Rule Query with bad syntax",
			rule:          `Query("titi={test")`,
			expectedError: true,
		},
		{
			desc:          "Rule with Path without args",
			rule:          `Host("tchouk") && Path()`,
			expectedError: true,
		},
		{
			desc:          "Rule with an empty path",
			rule:          `Host("tchouk") && Path("")`,
			expectedError: true,
		},
		{
			desc:          "Rule with an empty path",
			rule:          `Host("tchouk") && Path("", "/titi")`,
			expectedError: true,
		},
		{
			desc: "Rule with not",
			rule: `!Host("tchouk")`,
			expected: map[string]int{
				"http://tchouk/titi": http.StatusNotFound,
				"http://test/powpow": http.StatusOK,
			},
		},
		{
			desc: "Rule with not on Path",
			rule: `!Path("/titi")`,
			expected: map[string]int{
				"http://tchouk/titi":   http.StatusNotFound,
				"http://tchouk/powpow": http.StatusOK,
			},
		},
		{
			desc: "Rule with not on multiple route with or",
			rule: `!(Host("tchouk") || Host("toto"))`,
			expected: map[string]int{
				"http://tchouk/titi": http.StatusNotFound,
				"http://toto/powpow": http.StatusNotFound,
				"http://test/powpow": http.StatusOK,
			},
		},
		{
			desc: "Rule with not on multiple route with and",
			rule: `!(Host("tchouk") && Path("/titi"))`,
			expected: map[string]int{
				"http://tchouk/titi": http.StatusNotFound,
				"http://tchouk/toto": http.StatusOK,
				"http://test/titi":   http.StatusOK,
			},
		},
		{
			desc: "Rule with not on multiple route with and another not",
			rule: `!(Host("tchouk") && !Path("/titi"))`,
			expected: map[string]int{
				"http://tchouk/titi": http.StatusOK,
				"http://toto/titi":   http.StatusOK,
				"http://tchouk/toto": http.StatusNotFound,
			},
		},
		{
			desc: "Rule with not on two rule",
			rule: `!Host("tchouk") || !Path("/titi")`,
			expected: map[string]int{
				"http://tchouk/titi": http.StatusNotFound,
				"http://tchouk/toto": http.StatusOK,
				"http://test/titi":   http.StatusOK,
			},
		},
		{
			desc: "Rule case with double not",
			rule: `!(!(Host("tchouk") && Pathprefix("/titi")))`,
			expected: map[string]int{
				"http://tchouk/titi":   http.StatusOK,
				"http://tchouk/powpow": http.StatusNotFound,
				"http://test/titi":     http.StatusNotFound,
			},
		},
		{
			desc: "Rule case with not domain",
			rule: `!Host("tchouk") && Pathprefix("/titi")`,
			expected: map[string]int{
				"http://tchouk/titi":   http.StatusNotFound,
				"http://tchouk/powpow": http.StatusNotFound,
				"http://toto/powpow":   http.StatusNotFound,
				"http://toto/titi":     http.StatusOK,
			},
		},
		{
			desc: "Rule with multiple host AND multiple path AND not",
			rule: `!(Host("tchouk","pouet") && Path("/powpow", "/titi"))`,
			expected: map[string]int{
				"http://tchouk/toto":   http.StatusOK,
				"http://tchouk/powpow": http.StatusNotFound,
				"http://pouet/powpow":  http.StatusNotFound,
				"http://tchouk/titi":   http.StatusNotFound,
				"http://pouet/titi":    http.StatusNotFound,
				"http://pouet/toto":    http.StatusOK,
				"http://plopi/a":       http.StatusOK,
			},
		},
		{
			desc:          "ClientIP empty",
			rule:          "ClientIP(``)",
			expectedError: true,
		},
		{
			desc:          "Invalid ClientIP",
			rule:          "ClientIP(`invalid`)",
			expectedError: true,
		},
		{
			desc:       "Non matching ClientIP",
			rule:       "ClientIP(`10.10.1.1`)",
			remoteAddr: "10.0.0.0",
			expected: map[string]int{
				"http://tchouk/toto": http.StatusNotFound,
			},
		},
		{
			desc:       "Non matching IPv6",
			rule:       "ClientIP(`10::10`)",
			remoteAddr: "::1",
			expected: map[string]int{
				"http://tchouk/toto": http.StatusNotFound,
			},
		},
		{
			desc:       "Matching IP",
			rule:       "ClientIP(`10.0.0.0`)",
			remoteAddr: "10.0.0.0:8456",
			expected: map[string]int{
				"http://tchouk/toto": http.StatusOK,
			},
		},
		{
			desc:       "Matching IPv6",
			rule:       "ClientIP(`10::10`)",
			remoteAddr: "10::10",
			expected: map[string]int{
				"http://tchouk/toto": http.StatusOK,
			},
		},
		{
			desc:       "Matching IP among several IP",
			rule:       "ClientIP(`10.0.0.1`, `10.0.0.0`)",
			remoteAddr: "10.0.0.0",
			expected: map[string]int{
				"http://tchouk/toto": http.StatusOK,
			},
		},
		{
			desc:       "Non Matching IP with CIDR",
			rule:       "ClientIP(`11.0.0.0/24`)",
			remoteAddr: "10.0.0.0",
			expected: map[string]int{
				"http://tchouk/toto": http.StatusNotFound,
			},
		},
		{
			desc:       "Non Matching IPv6 with CIDR",
			rule:       "ClientIP(`11::/16`)",
			remoteAddr: "10::",
			expected: map[string]int{
				"http://tchouk/toto": http.StatusNotFound,
			},
		},
		{
			desc:       "Matching IP with CIDR",
			rule:       "ClientIP(`10.0.0.0/16`)",
			remoteAddr: "10.0.0.0",
			expected: map[string]int{
				"http://tchouk/toto": http.StatusOK,
			},
		},
		{
			desc:       "Matching IPv6 with CIDR",
			rule:       "ClientIP(`10::/16`)",
			remoteAddr: "10::10",
			expected: map[string]int{
				"http://tchouk/toto": http.StatusOK,
			},
		},
		{
			desc:       "Matching IP among several CIDR",
			rule:       "ClientIP(`11.0.0.0/16`, `10.0.0.0/16`)",
			remoteAddr: "10.0.0.0",
			expected: map[string]int{
				"http://tchouk/toto": http.StatusOK,
			},
		},
		{
			desc:       "Matching IP among non matching CIDR and matching IP",
			rule:       "ClientIP(`11.0.0.0/16`, `10.0.0.0`)",
			remoteAddr: "10.0.0.0",
			expected: map[string]int{
				"http://tchouk/toto": http.StatusOK,
			},
		},
		{
			desc:       "Matching IP among matching CIDR and non matching IP",
			rule:       "ClientIP(`11.0.0.0`, `10.0.0.0/16`)",
			remoteAddr: "10.0.0.0",
			expected: map[string]int{
				"http://tchouk/toto": http.StatusOK,
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

			err = muxer.AddRoute(test.rule, "v2", 0, handler)
			if test.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// RequestDecorator is necessary for the hostV2 rule
				reqHost := requestdecorator.New(nil)

				results := make(map[string]int)
				for calledURL := range test.expected {
					w := httptest.NewRecorder()

					req := testhelpers.MustNewRequest(http.MethodGet, calledURL, nil)

					// Useful for the ClientIP matcher
					req.RemoteAddr = test.remoteAddr

					for key, value := range test.headers {
						req.Header.Set(key, value)
					}
					reqHost.ServeHTTP(w, req, muxer.ServeHTTP)
					results[calledURL] = w.Code
				}
				assert.Equal(t, test.expected, results)
			}
		})
	}
}
