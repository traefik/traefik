package http

import (
	"bufio"
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/canonicalpath"
	"github.com/traefik/traefik/v3/pkg/middlewares/requestdecorator"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
)

func TestMuxer(t *testing.T) {
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
			desc:          "Rule without quote",
			rule:          "Host(example.com)",
			expectedError: true,
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
			desc: "Host and PathPrefix",
			rule: "Host(`localhost`) && PathPrefix(`/css`)",
			expected: map[string]int{
				"https://localhost/css": http.StatusOK,
				"https://localhost/js":  http.StatusNotFound,
			},
		},
		{
			desc: "Rule with Host OR Host",
			rule: "Host(`example.com`) || Host(`example.org`)",
			expected: map[string]int{
				"https://example.com/css": http.StatusOK,
				"https://example.org/js":  http.StatusOK,
				"https://example.eu/html": http.StatusNotFound,
			},
		},
		{
			desc: "Rule with host OR (host AND path)",
			rule: `Host("example.com") || (Host("example.org") && Path("/css"))`,
			expected: map[string]int{
				"https://example.com/css": http.StatusOK,
				"https://example.com/js":  http.StatusOK,
				"https://example.org/css": http.StatusOK,
				"https://example.org/js":  http.StatusNotFound,
				"https://example.eu/css":  http.StatusNotFound,
			},
		},
		{
			desc: "Rule with host OR host AND path",
			rule: `Host("example.com") || Host("example.org") && Path("/css")`,
			expected: map[string]int{
				"https://example.com/css": http.StatusOK,
				"https://example.com/js":  http.StatusOK,
				"https://example.org/css": http.StatusOK,
				"https://example.org/js":  http.StatusNotFound,
				"https://example.eu/css":  http.StatusNotFound,
			},
		},
		{
			desc: "Rule with (host OR host) AND path",
			rule: `(Host("example.com") || Host("example.org")) && Path("/css")`,
			expected: map[string]int{
				"https://example.com/css": http.StatusOK,
				"https://example.com/js":  http.StatusNotFound,
				"https://example.org/css": http.StatusOK,
				"https://example.org/js":  http.StatusNotFound,
				"https://example.eu/css":  http.StatusNotFound,
			},
		},
		{
			desc: "Rule with (host AND path) OR (host AND path)",
			rule: `(Host("example.com") && Path("/js")) || ((Host("example.org")) && Path("/css"))`,
			expected: map[string]int{
				"https://example.com/css": http.StatusNotFound,
				"https://example.com/js":  http.StatusOK,
				"https://example.org/css": http.StatusOK,
				"https://example.org/js":  http.StatusNotFound,
				"https://example.eu/css":  http.StatusNotFound,
			},
		},
		{
			desc: "Rule case UPPER",
			rule: `PATHPREFIX("/css")`,
			expected: map[string]int{
				"https://example.com/css": http.StatusOK,
				"https://example.com/js":  http.StatusNotFound,
			},
		},
		{
			desc: "Rule case lower",
			rule: `pathprefix("/css")`,
			expected: map[string]int{
				"https://example.com/css": http.StatusOK,
				"https://example.com/js":  http.StatusNotFound,
			},
		},
		{
			desc: "Rule case CamelCase",
			rule: `PathPrefix("/css")`,
			expected: map[string]int{
				"https://example.com/css": http.StatusOK,
				"https://example.com/js":  http.StatusNotFound,
			},
		},
		{
			desc: "Rule case Title",
			rule: `Pathprefix("/css")`,
			expected: map[string]int{
				"https://example.com/css": http.StatusOK,
				"https://example.com/js":  http.StatusNotFound,
			},
		},
		{
			desc: "Rule with not",
			rule: `!Host("example.com")`,
			expected: map[string]int{
				"https://example.org": http.StatusOK,
				"https://example.com": http.StatusNotFound,
			},
		},
		{
			desc: "Rule with not on multiple route with or",
			rule: `!(Host("example.com") || Host("example.org"))`,
			expected: map[string]int{
				"https://example.eu/js":   http.StatusOK,
				"https://example.com/css": http.StatusNotFound,
				"https://example.org/js":  http.StatusNotFound,
			},
		},
		{
			desc: "Rule with not on multiple route with and",
			rule: `!(Host("example.com") && Path("/css"))`,
			expected: map[string]int{
				"https://example.com/js":  http.StatusOK,
				"https://example.eu/css":  http.StatusOK,
				"https://example.com/css": http.StatusNotFound,
			},
		},
		{
			desc: "Rule with not on multiple route with and another not",
			rule: `!(Host("example.com") && !Path("/css"))`,
			expected: map[string]int{
				"https://example.com/css": http.StatusOK,
				"https://example.org/css": http.StatusOK,
				"https://example.com/js":  http.StatusNotFound,
			},
		},
		{
			desc: "Rule with not on two rule",
			rule: `!Host("example.com") || !Path("/css")`,
			expected: map[string]int{
				"https://example.com/js":  http.StatusOK,
				"https://example.org/css": http.StatusOK,
				"https://example.com/css": http.StatusNotFound,
			},
		},
		{
			desc: "Rule case with double not",
			rule: `!(!(Host("example.com") && Pathprefix("/css")))`,
			expected: map[string]int{
				"https://example.com/css": http.StatusOK,
				"https://example.com/js":  http.StatusNotFound,
				"https://example.org/css": http.StatusNotFound,
			},
		},
		{
			desc: "Rule case with not domain",
			rule: `!Host("example.com") && Pathprefix("/css")`,
			expected: map[string]int{
				"https://example.org/css": http.StatusOK,
				"https://example.org/js":  http.StatusNotFound,
				"https://example.com/css": http.StatusNotFound,
				"https://example.com/js":  http.StatusNotFound,
			},
		},
		{
			desc: "Rule with multiple host AND multiple path AND not",
			rule: `!(Host("example.com") && Path("/js"))`,
			expected: map[string]int{
				"https://example.com/js":    http.StatusNotFound,
				"https://example.com/html":  http.StatusOK,
				"https://example.org/js":    http.StatusOK,
				"https://example.com/css":   http.StatusOK,
				"https://example.org/css":   http.StatusOK,
				"https://example.org/html":  http.StatusOK,
				"https://example.eu/images": http.StatusOK,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			parser, err := NewSyntaxParser()
			require.NoError(t, err)

			muxer := NewMuxer(parser)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			err = muxer.AddRoute(test.rule, "", 0, handler)
			if test.expectedError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// RequestDecorator is necessary for the host rule
			reqHost := requestdecorator.New(nil)

			results := make(map[string]int)
			for calledURL := range test.expected {
				req := testhelpers.MustNewRequest(http.MethodGet, calledURL, http.NoBody)

				// Useful for the ClientIP matcher
				req.RemoteAddr = test.remoteAddr

				for key, value := range test.headers {
					req.Header.Set(key, value)
				}

				w := httptest.NewRecorder()
				reqHost.ServeHTTP(w, req, muxer.ServeHTTP)
				results[calledURL] = w.Code
			}

			assert.Equal(t, test.expected, results)
		})
	}
}

func Test_addRoutePriority(t *testing.T) {
	type Case struct {
		xFrom    string
		rule     string
		priority int
	}

	testCases := []struct {
		desc     string
		path     string
		cases    []Case
		expected string
	}{
		{
			desc: "Higher priority on second rule",
			path: "/my",
			cases: []Case{
				{
					xFrom:    "header1",
					rule:     "PathPrefix(`/my`)",
					priority: 10,
				},
				{
					xFrom:    "header2",
					rule:     "PathPrefix(`/my`)",
					priority: 20,
				},
			},
			expected: "header2",
		},
		{
			desc: "Higher priority on first rule",
			path: "/my",
			cases: []Case{
				{
					xFrom:    "header1",
					rule:     "PathPrefix(`/my`)",
					priority: 20,
				},
				{
					xFrom:    "header2",
					rule:     "PathPrefix(`/my`)",
					priority: 10,
				},
			},
			expected: "header1",
		},
		{
			desc: "Higher priority on second rule with different rule",
			path: "/mypath",
			cases: []Case{
				{
					xFrom:    "header1",
					rule:     "PathPrefix(`/mypath`)",
					priority: 10,
				},
				{
					xFrom:    "header2",
					rule:     "PathPrefix(`/my`)",
					priority: 20,
				},
			},
			expected: "header2",
		},
		{
			desc: "Higher priority on longest rule (longest first)",
			path: "/mypath",
			cases: []Case{
				{
					xFrom: "header1",
					rule:  "PathPrefix(`/mypath`)",
				},
				{
					xFrom: "header2",
					rule:  "PathPrefix(`/my`)",
				},
			},
			expected: "header1",
		},
		{
			desc: "Higher priority on longest rule (longest second)",
			path: "/mypath",
			cases: []Case{
				{
					xFrom: "header1",
					rule:  "PathPrefix(`/my`)",
				},
				{
					xFrom: "header2",
					rule:  "PathPrefix(`/mypath`)",
				},
			},
			expected: "header2",
		},
		{
			desc: "Higher priority on longest rule (longest third)",
			path: "/mypath",
			cases: []Case{
				{
					xFrom: "header1",
					rule:  "PathPrefix(`/my`)",
				},
				{
					xFrom: "header2",
					rule:  "PathPrefix(`/mypa`)",
				},
				{
					xFrom: "header3",
					rule:  "PathPrefix(`/mypath`)",
				},
			},
			expected: "header3",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			parser, err := NewSyntaxParser()
			require.NoError(t, err)

			muxer := NewMuxer(parser)

			for _, route := range test.cases {
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("X-From", route.xFrom)
				})

				if route.priority == 0 {
					route.priority = GetRulePriority(route.rule)
				}

				err := muxer.AddRoute(route.rule, "", route.priority, handler)
				require.NoError(t, err, route.rule)
			}

			w := httptest.NewRecorder()
			req := testhelpers.MustNewRequest(http.MethodGet, test.path, http.NoBody)

			muxer.ServeHTTP(w, req)

			assert.Equal(t, test.expected, w.Header().Get("X-From"))
		})
	}
}

func TestParseDomains(t *testing.T) {
	testCases := []struct {
		description   string
		expression    string
		domain        []string
		errorExpected bool
	}{
		{
			description:   "Unknown rule",
			expression:    "Foobar(`foo.bar`,`test.bar`)",
			errorExpected: true,
		},
		{
			description: "No host rule",
			expression:  "Path(`/test`)",
		},
		{
			description: "Host rule and another rule",
			expression:  "Host(`foo.bar`) && Path(`/test`)",
			domain:      []string{"foo.bar"},
		},
		{
			description: "Host rule to trim and another rule",
			expression:  "Host(`Foo.Bar`) || Host(`bar.buz`) && Path(`/test`)",
			domain:      []string{"foo.bar", "bar.buz"},
		},
		{
			description: "Host rule to trim and another rule",
			expression:  "Host(`Foo.Bar`) && Path(`/test`)",
			domain:      []string{"foo.bar"},
		},
		{
			description: "Host rule with no domain",
			expression:  "Host() && Path(`/test`)",
		},
	}

	for _, test := range testCases {
		t.Run(test.expression, func(t *testing.T) {
			t.Parallel()

			domains, err := ParseDomains(test.expression)

			if test.errorExpected {
				require.Errorf(t, err, "unable to parse correctly the domains in the Host rule from %q", test.expression)
			} else {
				require.NoError(t, err, "%s: Error while parsing domain.", test.expression)
			}

			assert.Equal(t, test.domain, domains, "%s: Error parsing domains from expression.", test.expression)
		})
	}
}

// TestEmptyHost is a non regression test for
// https://github.com/traefik/traefik/pull/9131
func TestEmptyHost(t *testing.T) {
	testCases := []struct {
		desc     string
		request  string
		rule     string
		expected int
	}{
		{
			desc:     "HostRegexp with absolute-form URL with empty host with non-matching host header",
			request:  "GET http://@/ HTTP/1.1\r\nHost: example.com\r\n\r\n",
			rule:     "HostRegexp(`example.com`)",
			expected: http.StatusOK,
		},
		{
			desc:     "Host with absolute-form URL with empty host with non-matching host header",
			request:  "GET http://@/ HTTP/1.1\r\nHost: example.com\r\n\r\n",
			rule:     "Host(`example.com`)",
			expected: http.StatusOK,
		},
		{
			desc:     "HostRegexp with absolute-form URL with matching host header",
			request:  "GET http://example.com/ HTTP/1.1\r\nHost: example.org\r\n\r\n",
			rule:     "HostRegexp(`example.com`)",
			expected: http.StatusOK,
		},
		{
			desc:     "Host with absolute-form URL with matching host header",
			request:  "GET http://example.com/ HTTP/1.1\r\nHost: example.org\r\n\r\n",
			rule:     "Host(`example.com`)",
			expected: http.StatusOK,
		},
		{
			desc:     "HostRegexp with absolute-form URL with non-matching host header",
			request:  "GET http://example.com/ HTTP/1.1\r\nHost: example.org\r\n\r\n",
			rule:     "HostRegexp(`example.org`)",
			expected: http.StatusNotFound,
		},
		{
			desc:     "Host with absolute-form URL with non-matching host header",
			request:  "GET http://example.com/ HTTP/1.1\r\nHost: example.org\r\n\r\n",
			rule:     "Host(`example.org`)",
			expected: http.StatusNotFound,
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
			require.NoError(t, err)

			// RequestDecorator is necessary for the host rule
			reqHost := requestdecorator.New(nil)

			w := httptest.NewRecorder()

			req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader([]byte(test.request))))
			require.NoError(t, err)

			reqHost.ServeHTTP(w, req, muxer.ServeHTTP)
			assert.Equal(t, test.expected, w.Code)
		})
	}
}

func TestGetRulePriority(t *testing.T) {
	testCases := []struct {
		desc     string
		rule     string
		expected int
	}{
		{
			desc:     "simple rule",
			rule:     "Host(`example.org`)",
			expected: 19,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.expected, GetRulePriority(test.rule))
		})
	}
}

func TestRoutingPath(t *testing.T) {
	tests := []struct {
		desc                string
		path                string
		expectedRoutingPath string
	}{
		{
			desc:                "unallowed percent-encoded character is decoded",
			path:                "/foo%20bar",
			expectedRoutingPath: "/foo bar",
		},
		{
			desc:                "reserved percent-encoded character is kept encoded",
			path:                "/foo%2Fbar",
			expectedRoutingPath: "/foo%2Fbar",
		},
		{
			desc:                "multiple mixed characters",
			path:                "/foo%20bar%2Fbaz%23qux",
			expectedRoutingPath: "/foo bar%2Fbaz%23qux",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "http://foo"+test.path, http.NoBody)

			var err error
			req, err = withRoutingPath(req)
			require.NoError(t, err)

			gotRoutingPath := getRoutingPath(req)
			assert.NotNil(t, gotRoutingPath)
			assert.Equal(t, test.expectedRoutingPath, *gotRoutingPath)
		})
	}
}

// TestCWE436UnreservedCharacterBypass tests that encoding unreserved characters
// cannot be used to bypass path-based routing rules (CWE-436 mitigation).
//
// Per RFC 3986, /%61dmin and /admin are semantically equivalent because 'a' (0x61)
// is an unreserved character. Both MUST route to the same destination.
//
// Attack scenario: Attacker requests /%61dmin hoping to bypass /admin rules.
// See: https://cwe.mitre.org/data/definitions/436.html
// See: https://github.com/advisories/GHSA-gm3x-23wp-hc2c
func TestCWE436UnreservedCharacterBypass(t *testing.T) {
	testCases := []struct {
		desc        string
		rule        string
		requestPath string
		shouldMatch bool
	}{
		{
			desc:        "direct /admin matches PathPrefix /admin",
			rule:        "PathPrefix(`/admin`)",
			requestPath: "/admin",
			shouldMatch: true,
		},
		{
			desc:        "encoded /%61dmin (a=0x61) MUST also match PathPrefix /admin",
			rule:        "PathPrefix(`/admin`)",
			requestPath: "/%61dmin",
			shouldMatch: true,
		},
		{
			desc:        "fully encoded /%61%64%6D%69%6E MUST match PathPrefix /admin",
			rule:        "PathPrefix(`/admin`)",
			requestPath: "/%61%64%6D%69%6E",
			shouldMatch: true,
		},
		{
			desc:        "uppercase hex /%61%64%6d%69%6e also matches (hex case insensitive)",
			rule:        "PathPrefix(`/admin`)",
			requestPath: "/%61%64%6d%69%6e",
			shouldMatch: true,
		},
		{
			desc:        "encoded space in path /%61dmin%20panel matches /admin panel",
			rule:        "PathPrefix(`/admin panel`)",
			requestPath: "/%61dmin%20panel",
			shouldMatch: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			parser, err := NewSyntaxParser()
			require.NoError(t, err)
			muxer := NewMuxer(parser)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			err = muxer.AddRoute(test.rule, "v3", 0, handler)
			require.NoError(t, err)

			// Create request with the test path
			req := httptest.NewRequest(http.MethodGet, "http://localhost"+test.requestPath, http.NoBody)

			// Apply canonical path middleware (simulating what happens in production)
			pr := canonicalpath.Canonicalize(req.URL.EscapedPath(), canonicalpath.StrategyPreserveReserved)
			req = req.WithContext(canonicalpath.WithPathRepresentation(req.Context(), pr))

			recorder := httptest.NewRecorder()
			muxer.ServeHTTP(recorder, req)

			if test.shouldMatch {
				assert.Equal(t, http.StatusOK, recorder.Code,
					"CWE-436 BYPASS: encoded unreserved character evaded routing rule")
			} else {
				assert.Equal(t, http.StatusNotFound, recorder.Code,
					"request should not match the rule")
			}
		})
	}
}

// TestCWE436ReservedCharacterSemantics tests that encoded reserved characters
// preserve their semantic meaning (not decoded during routing).
//
// Per RFC 3986, /admin%2Fsecret (one segment with literal '/') is semantically
// DIFFERENT from /admin/secret (two segments). They MUST be treated differently.
//
// This is NOT about security bypass - it's about semantic correctness.
// Decoding %2F would break legitimate use cases like GitLab namespaces.
func TestCWE436ReservedCharacterSemantics(t *testing.T) {
	testCases := []struct {
		desc        string
		rule        string
		requestPath string
		shouldMatch bool
	}{
		{
			desc:        "/admin/secret (two segments) matches Path /admin/secret",
			rule:        "Path(`/admin/secret`)",
			requestPath: "/admin/secret",
			shouldMatch: true,
		},
		{
			desc:        "/admin%2Fsecret (one segment) does NOT match Path /admin/secret",
			rule:        "Path(`/admin/secret`)",
			requestPath: "/admin%2Fsecret",
			shouldMatch: false,
		},
		{
			desc:        "/admin%2Fsecret matches PathPrefix /admin (starts with /admin)",
			rule:        "PathPrefix(`/admin`)",
			requestPath: "/admin%2Fsecret",
			shouldMatch: true,
		},
		{
			desc:        "lowercase /admin%2fsecret also does NOT match Path /admin/secret",
			rule:        "Path(`/admin/secret`)",
			requestPath: "/admin%2fsecret",
			shouldMatch: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			parser, err := NewSyntaxParser()
			require.NoError(t, err)
			muxer := NewMuxer(parser)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			err = muxer.AddRoute(test.rule, "v3", 0, handler)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, "http://localhost"+test.requestPath, http.NoBody)

			// Apply canonical path middleware
			pr := canonicalpath.Canonicalize(req.URL.EscapedPath(), canonicalpath.StrategyPreserveReserved)
			req = req.WithContext(canonicalpath.WithPathRepresentation(req.Context(), pr))

			recorder := httptest.NewRecorder()
			muxer.ServeHTTP(recorder, req)

			if test.shouldMatch {
				assert.Equal(t, http.StatusOK, recorder.Code,
					"request should match the rule")
			} else {
				assert.Equal(t, http.StatusNotFound, recorder.Code,
					"RFC 3986: /admin%%2Fsecret (one segment) must NOT equal /admin/secret (two segments)")
			}
		})
	}
}
