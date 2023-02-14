package http

import (
	"bufio"
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/middlewares/requestdecorator"
	"github.com/traefik/traefik/v2/pkg/testhelpers"
)

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
				"http://localhost/foo": http.StatusMethodNotAllowed,
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
		test := test

		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			muxer, err := NewMuxer()
			require.NoError(t, err)

			err = muxer.AddRoute(test.rule, 0, handler)
			if test.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// RequestDecorator is necessary for the host rule
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
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			muxer, err := NewMuxer()
			require.NoError(t, err)

			for _, route := range test.cases {
				route := route
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("X-From", route.xFrom)
				})

				err := muxer.AddRoute(route.rule, route.priority, handler)
				require.NoError(t, err, route.rule)
			}

			muxer.SortRoutes()

			w := httptest.NewRecorder()
			req := testhelpers.MustNewRequest(http.MethodGet, test.path, nil)

			muxer.ServeHTTP(w, req)

			assert.Equal(t, test.expected, w.Header().Get("X-From"))
		})
	}
}

func TestHostRegexp(t *testing.T) {
	testCases := []struct {
		desc    string
		hostExp string
		urls    map[string]bool
	}{
		{
			desc:    "capturing group",
			hostExp: "{subdomain:(foo\\.)?bar\\.com}",
			urls: map[string]bool{
				"http://foo.bar.com": true,
				"http://bar.com":     true,
				"http://fooubar.com": false,
				"http://barucom":     false,
				"http://barcom":      false,
			},
		},
		{
			desc:    "non capturing group",
			hostExp: "{subdomain:(?:foo\\.)?bar\\.com}",
			urls: map[string]bool{
				"http://foo.bar.com": true,
				"http://bar.com":     true,
				"http://fooubar.com": false,
				"http://barucom":     false,
				"http://barcom":      false,
			},
		},
		{
			desc:    "regex insensitive",
			hostExp: "{dummy:[A-Za-z-]+\\.bar\\.com}",
			urls: map[string]bool{
				"http://FOO.bar.com": true,
				"http://foo.bar.com": true,
				"http://fooubar.com": false,
				"http://barucom":     false,
				"http://barcom":      false,
			},
		},
		{
			desc:    "insensitive host",
			hostExp: "{dummy:[a-z-]+\\.bar\\.com}",
			urls: map[string]bool{
				"http://FOO.bar.com": true,
				"http://foo.bar.com": true,
				"http://fooubar.com": false,
				"http://barucom":     false,
				"http://barcom":      false,
			},
		},
		{
			desc:    "insensitive host simple",
			hostExp: "foo.bar.com",
			urls: map[string]bool{
				"http://FOO.bar.com": true,
				"http://foo.bar.com": true,
				"http://fooubar.com": false,
				"http://barucom":     false,
				"http://barcom":      false,
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rt := &mux.Route{}
			err := hostRegexp(rt, test.hostExp)
			require.NoError(t, err)

			for testURL, match := range test.urls {
				req := testhelpers.MustNewRequest(http.MethodGet, testURL, nil)
				assert.Equal(t, match, rt.Match(req, &mux.RouteMatch{}), testURL)
			}
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
			description: "Several host rules",
			expression:  "Host(`foo.bar`,`test.bar`)",
			domain:      []string{"foo.bar", "test.bar"},
		},
		{
			description: "Several host rules upper",
			expression:  "HOST(`foo.bar`,`test.bar`)",
			domain:      []string{"foo.bar", "test.bar"},
		},
		{
			description: "Several host rules lower",
			expression:  "host(`foo.bar`,`test.bar`)",
			domain:      []string{"foo.bar", "test.bar"},
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
			expression:  "Host(`Foo.Bar`) && Path(`/test`)",
			domain:      []string{"foo.bar"},
		},
		{
			description: "Host rule with no domain",
			expression:  "Host() && Path(`/test`)",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.expression, func(t *testing.T) {
			t.Parallel()

			domains, err := ParseDomains(test.expression)

			if test.errorExpected {
				require.Errorf(t, err, "unable to parse correctly the domains in the Host rule from %q", test.expression)
			} else {
				require.NoError(t, err, "%s: Error while parsing domain.", test.expression)
			}

			assert.EqualValues(t, test.domain, domains, "%s: Error parsing domains from expression.", test.expression)
		})
	}
}

func TestAbsoluteFormURL(t *testing.T) {
	testCases := []struct {
		desc     string
		request  string
		rule     string
		expected int
	}{
		{
			desc:     "!HostRegexp with absolute-form URL with empty host with non-matching host header",
			request:  "GET http://@/ HTTP/1.1\r\nHost: test.localhost\r\n\r\n",
			rule:     "!HostRegexp(`test.localhost`)",
			expected: http.StatusNotFound,
		},
		{
			desc:     "!Host with absolute-form URL with empty host with non-matching host header",
			request:  "GET http://@/ HTTP/1.1\r\nHost: test.localhost\r\n\r\n",
			rule:     "!Host(`test.localhost`)",
			expected: http.StatusNotFound,
		},
		{
			desc:     "!HostRegexp with absolute-form URL with matching host header",
			request:  "GET http://test.localhost/ HTTP/1.1\r\nHost: toto.localhost\r\n\r\n",
			rule:     "!HostRegexp(`test.localhost`)",
			expected: http.StatusNotFound,
		},
		{
			desc:     "!Host with absolute-form URL with matching host header",
			request:  "GET http://test.localhost/ HTTP/1.1\r\nHost: toto.localhost\r\n\r\n",
			rule:     "!Host(`test.localhost`)",
			expected: http.StatusNotFound,
		},
		{
			desc:     "!HostRegexp with absolute-form URL with non-matching host header",
			request:  "GET http://test.localhost/ HTTP/1.1\r\nHost: toto.localhost\r\n\r\n",
			rule:     "!HostRegexp(`toto.localhost`)",
			expected: http.StatusOK,
		},
		{
			desc:     "!Host with absolute-form URL with non-matching host header",
			request:  "GET http://test.localhost/ HTTP/1.1\r\nHost: toto.localhost\r\n\r\n",
			rule:     "!Host(`toto.localhost`)",
			expected: http.StatusOK,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			muxer, err := NewMuxer()
			require.NoError(t, err)

			err = muxer.AddRoute(test.rule, 0, handler)
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
