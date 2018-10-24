package rules

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/containous/mux"
	"github.com/containous/traefik/middlewares"
	"github.com/containous/traefik/testhelpers"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseOneRule(t *testing.T) {
	reqHostMid := &middlewares.RequestHost{}
	rules := &Rules{
		Route: &types.ServerRoute{
			Route: mux.NewRouter().NewRoute(),
		},
	}

	expression := "Host:foo.bar"

	routeResult, err := rules.Parse(expression)
	require.NoError(t, err, "Error while building route for %s", expression)

	request := testhelpers.MustNewRequest(http.MethodGet, "http://foo.bar", nil)

	reqHostMid.ServeHTTP(nil, request, func(w http.ResponseWriter, r *http.Request) {
		routeMatch := routeResult.Match(r, &mux.RouteMatch{Route: routeResult})
		assert.True(t, routeMatch, "Rule %s don't match.", expression)
	})
}

func TestParseTwoRules(t *testing.T) {
	reqHostMid := &middlewares.RequestHost{}
	rules := &Rules{
		Route: &types.ServerRoute{
			Route: mux.NewRouter().NewRoute(),
		},
	}

	expression := "Host: Foo.Bar ; Path:/FOObar"

	routeResult, err := rules.Parse(expression)
	require.NoError(t, err, "Error while building route for %s.", expression)

	request := testhelpers.MustNewRequest(http.MethodGet, "http://foo.bar/foobar", nil)
	reqHostMid.ServeHTTP(nil, request, func(w http.ResponseWriter, r *http.Request) {
		routeMatch := routeResult.Match(r, &mux.RouteMatch{Route: routeResult})
		assert.False(t, routeMatch, "Rule %s don't match.", expression)
	})

	request = testhelpers.MustNewRequest(http.MethodGet, "http://foo.bar/FOObar", nil)
	reqHostMid.ServeHTTP(nil, request, func(w http.ResponseWriter, r *http.Request) {
		routeMatch := routeResult.Match(r, &mux.RouteMatch{Route: routeResult})
		assert.True(t, routeMatch, "Rule %s don't match.", expression)
	})
}

func TestParseDomains(t *testing.T) {
	rules := &Rules{}

	tests := []struct {
		description   string
		expression    string
		domain        []string
		errorExpected bool
	}{
		{
			description:   "Many host rules",
			expression:    "Host:foo.bar,test.bar",
			domain:        []string{"foo.bar", "test.bar"},
			errorExpected: false,
		},
		{
			description:   "No host rule",
			expression:    "Path:/test",
			errorExpected: false,
		},
		{
			description:   "Host rule and another rule",
			expression:    "Host:foo.bar;Path:/test",
			domain:        []string{"foo.bar"},
			errorExpected: false,
		},
		{
			description:   "Host rule to trim and another rule",
			expression:    "Host: Foo.Bar ;Path:/test",
			domain:        []string{"foo.bar"},
			errorExpected: false,
		},
		{
			description:   "Host rule with no domain",
			expression:    "Host: ;Path:/test",
			errorExpected: true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.expression, func(t *testing.T) {
			t.Parallel()

			domains, err := rules.ParseDomains(test.expression)

			if test.errorExpected {
				require.Errorf(t, err, "unable to parse correctly the domains in the Host rule from %q", test.expression)
			} else {
				require.NoError(t, err, "%s: Error while parsing domain.", test.expression)
			}

			assert.EqualValues(t, test.domain, domains, "%s: Error parsing domains from expression.", test.expression)
		})
	}
}

func TestPriorites(t *testing.T) {
	router := mux.NewRouter()
	router.StrictSlash(true)

	rules := &Rules{Route: &types.ServerRoute{Route: router.NewRoute()}}
	expression01 := "PathPrefix:/foo"

	routeFoo, err := rules.Parse(expression01)
	require.NoError(t, err, "Error while building route for %s", expression01)

	fooHandler := &fakeHandler{name: "fooHandler"}
	routeFoo.Handler(fooHandler)

	routeMatch := router.Match(&http.Request{URL: &url.URL{Path: "/foo"}}, &mux.RouteMatch{})
	assert.True(t, routeMatch, "Error matching route")

	routeMatch = router.Match(&http.Request{URL: &url.URL{Path: "/fo"}}, &mux.RouteMatch{})
	assert.False(t, routeMatch, "Error matching route")

	multipleRules := &Rules{Route: &types.ServerRoute{Route: router.NewRoute()}}
	expression02 := "PathPrefix:/foobar"

	routeFoobar, err := multipleRules.Parse(expression02)
	require.NoError(t, err, "Error while building route for %s", expression02)

	foobarHandler := &fakeHandler{name: "foobarHandler"}
	routeFoobar.Handler(foobarHandler)
	routeMatch = router.Match(&http.Request{URL: &url.URL{Path: "/foo"}}, &mux.RouteMatch{})

	assert.True(t, routeMatch, "Error matching route")

	fooMatcher := &mux.RouteMatch{}
	routeMatch = router.Match(&http.Request{URL: &url.URL{Path: "/foobar"}}, fooMatcher)

	assert.True(t, routeMatch, "Error matching route")
	assert.NotEqual(t, fooMatcher.Handler, foobarHandler, "Error matching priority")
	assert.Equal(t, fooMatcher.Handler, fooHandler, "Error matching priority")

	routeFoo.Priority(1)
	routeFoobar.Priority(10)
	router.SortRoutes()

	foobarMatcher := &mux.RouteMatch{}
	routeMatch = router.Match(&http.Request{URL: &url.URL{Path: "/foobar"}}, foobarMatcher)

	assert.True(t, routeMatch, "Error matching route")
	assert.Equal(t, foobarMatcher.Handler, foobarHandler, "Error matching priority")
	assert.NotEqual(t, foobarMatcher.Handler, fooHandler, "Error matching priority")
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

			rls := &Rules{
				Route: &types.ServerRoute{
					Route: &mux.Route{},
				},
			}

			rt := rls.hostRegexp(test.hostExp)

			for testURL, match := range test.urls {
				req := testhelpers.MustNewRequest(http.MethodGet, testURL, nil)
				assert.Equal(t, match, rt.Match(req, &mux.RouteMatch{}), testURL)
			}
		})
	}
}

func TestParseInvalidSyntax(t *testing.T) {
	router := mux.NewRouter()
	router.StrictSlash(true)

	rules := &Rules{Route: &types.ServerRoute{Route: router.NewRoute()}}
	expression01 := "Path: /path1;Query:param_one=true, /path2"

	routeFoo, err := rules.Parse(expression01)
	require.Error(t, err)
	assert.Nil(t, routeFoo)
}

func TestPathPrefix(t *testing.T) {
	testCases := []struct {
		desc string
		path string
		urls map[string]bool
	}{
		{
			desc: "leading slash",
			path: "/bar",
			urls: map[string]bool{
				"http://foo.com/bar":  true,
				"http://foo.com/bar/": true,
			},
		},
		{
			desc: "leading trailing slash",
			path: "/bar/",
			urls: map[string]bool{
				"http://foo.com/bar":  false,
				"http://foo.com/bar/": true,
			},
		},
		{
			desc: "no slash",
			path: "bar",
			urls: map[string]bool{
				"http://foo.com/bar":  false,
				"http://foo.com/bar/": false,
			},
		},
		{
			desc: "trailing slash",
			path: "bar/",
			urls: map[string]bool{
				"http://foo.com/bar":  false,
				"http://foo.com/bar/": false,
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rls := &Rules{
				Route: &types.ServerRoute{
					Route: &mux.Route{},
				},
			}

			rt := rls.pathPrefix(test.path)

			for testURL, expectedMatch := range test.urls {
				req := testhelpers.MustNewRequest(http.MethodGet, testURL, nil)
				match := rt.Match(req, &mux.RouteMatch{})
				if match != expectedMatch {
					t.Errorf("Error matching %s with %s, got %v expected %v", test.path, testURL, match, expectedMatch)
				}
			}
		})
	}
}

type fakeHandler struct {
	name string
}

func (h *fakeHandler) ServeHTTP(http.ResponseWriter, *http.Request) {}
