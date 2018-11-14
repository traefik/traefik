package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/mux"
	"github.com/containous/traefik/middlewares/requestdecorator"
	"github.com/containous/traefik/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_addRoute(t *testing.T) {
	testCases := []struct {
		desc     string
		rule     string
		headers  map[string]string
		expected map[string]int
	}{
		{
			desc: "no rule",
			expected: map[string]int{
				"http://localhost/foo": http.StatusOK,
			},
		},
		{
			desc: "PathPrefix",
			rule: "PathPrefix:/foo",
			expected: map[string]int{
				"http://localhost/foo": http.StatusOK,
			},
		},
		{
			desc: "wrong PathPrefix",
			rule: "PathPrefix:/bar",
			expected: map[string]int{
				"http://localhost/foo": http.StatusNotFound,
			},
		},
		{
			desc: "Host",
			rule: "Host:localhost",
			expected: map[string]int{
				"http://localhost/foo": http.StatusOK,
			},
		},
		{
			desc: "wrong Host",
			rule: "Host:nope",
			expected: map[string]int{
				"http://localhost/foo": http.StatusNotFound,
			},
		},
		{
			desc: "Host and PathPrefix",
			rule: "Host:localhost;PathPrefix:/foo",
			expected: map[string]int{
				"http://localhost/foo": http.StatusOK,
			},
		},
		{
			desc: "Host and PathPrefix: wrong PathPrefix",
			rule: "Host:localhost;PathPrefix:/bar",
			expected: map[string]int{
				"http://localhost/foo": http.StatusNotFound,
			},
		},
		{
			desc: "Host and PathPrefix: wrong Host",
			rule: "Host:nope;PathPrefix:/bar",
			expected: map[string]int{
				"http://localhost/foo": http.StatusNotFound,
			},
		},
		{
			desc: "Host and PathPrefix: Host OR, first host",
			rule: "Host:nope,localhost;PathPrefix:/foo",
			expected: map[string]int{
				"http://localhost/foo": http.StatusOK,
			},
		},
		{
			desc: "Host and PathPrefix: Host OR, second host",
			rule: "Host:nope,localhost;PathPrefix:/foo",
			expected: map[string]int{
				"http://nope/foo": http.StatusOK,
			},
		},
		{
			desc: "Host and PathPrefix: Host OR, first host and wrong PathPrefix",
			rule: "Host:nope,localhost;PathPrefix:/bar",
			expected: map[string]int{
				"http://localhost/foo": http.StatusNotFound,
			},
		},
		{
			desc: "HostRegexp with capturing group",
			rule: "HostRegexp: {subdomain:(foo\\.)?bar\\.com}",
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
			rule: "HostRegexp: {subdomain:(?:foo\\.)?bar\\.com}",
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
			rule: "Method: GET",
			expected: map[string]int{
				"http://localhost/foo": http.StatusOK,
			},
		},
		{
			desc: "Methods with GET and POST",
			rule: "Method: GET,POST",
			expected: map[string]int{
				"http://localhost/foo": http.StatusOK,
			},
		},
		{
			desc: "Methods with POST",
			rule: "Method: POST",
			expected: map[string]int{
				"http://localhost/foo": http.StatusMethodNotAllowed,
			},
		},
		{
			desc: "Header with matching header",
			rule: "Headers: Content-Type,application/json",
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			expected: map[string]int{
				"http://localhost/foo": http.StatusOK,
			},
		},
		{
			desc: "Header without matching header",
			rule: "Headers: Content-Type,application/foo",
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			expected: map[string]int{
				"http://localhost/foo": http.StatusNotFound,
			},
		},
		{
			desc: "HeaderRegExp with matching header",
			rule: "HeadersRegexp: Content-Type, application/(text|json)",
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			expected: map[string]int{
				"http://localhost/foo": http.StatusOK,
			},
		},
		{
			desc: "HeaderRegExp without matching header",
			rule: "HeadersRegexp: Content-Type, application/(text|json)",
			headers: map[string]string{
				"Content-Type": "application/foo",
			},
			expected: map[string]int{
				"http://localhost/foo": http.StatusNotFound,
			},
		},
		{
			desc: "HeaderRegExp with matching second header",
			rule: "HeadersRegexp: Content-Type, application/(text|json)",
			headers: map[string]string{
				"Content-Type": "application/text",
			},
			expected: map[string]int{
				"http://localhost/foo": http.StatusOK,
			},
		},
		{
			desc: "Query with multiple params",
			rule: "Query: foo=bar, bar=baz",
			expected: map[string]int{
				"http://localhost/foo?foo=bar&bar=baz": http.StatusOK,
				"http://localhost/foo?bar=baz":         http.StatusNotFound,
			},
		},
		{
			desc: "Invalid rule syntax",
			rule: "Query:param_one=true, /path2;Path: /path1",
			expected: map[string]int{
				"http://localhost/foo?bar=baz": http.StatusNotFound,
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

			router := mux.NewRouter()
			router.SkipClean(true)

			err := addRoute(context.Background(), router, test.rule, 0, handler)
			require.NoError(t, err)

			// RequestDecorator is necessary for the host rule
			reqHost := requestdecorator.New(nil)

			results := make(map[string]int)
			for calledURL := range test.expected {
				w := httptest.NewRecorder()

				req := testhelpers.MustNewRequest(http.MethodGet, calledURL, nil)
				for key, value := range test.headers {
					req.Header.Set(key, value)
				}
				reqHost.ServeHTTP(w, req, router.ServeHTTP)
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
					rule:     "PathPrefix:/my",
					priority: 10,
				},
				{
					xFrom:    "header2",
					rule:     "PathPrefix:/my",
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
					rule:     "PathPrefix:/my",
					priority: 20,
				},
				{
					xFrom:    "header2",
					rule:     "PathPrefix:/my",
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
					rule:     "PathPrefix:/mypath",
					priority: 10,
				},
				{
					xFrom:    "header2",
					rule:     "PathPrefix:/my",
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
					rule:  "PathPrefix:/mypath",
				},
				{
					xFrom: "header2",
					rule:  "PathPrefix:/my",
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
					rule:  "PathPrefix:/my",
				},
				{
					xFrom: "header2",
					rule:  "PathPrefix:/mypath",
				},
			},
			expected: "header2",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			router := mux.NewRouter()

			for _, route := range test.cases {
				route := route
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("X-From", route.xFrom)
				})
				err := addRoute(context.Background(), router, route.rule, route.priority, handler)
				require.NoError(t, err, route)
			}

			router.SortRoutes()

			w := httptest.NewRecorder()
			req := testhelpers.MustNewRequest(http.MethodGet, test.path, nil)

			router.ServeHTTP(w, req)

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
			hostRegexp(rt, test.hostExp)

			for testURL, match := range test.urls {
				req := testhelpers.MustNewRequest(http.MethodGet, testURL, nil)
				assert.Equal(t, match, rt.Match(req, &mux.RouteMatch{}), testURL)
			}
		})
	}
}
