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
	router := mux.NewRouter()
	route := router.NewRoute()
	reqHostMid := &middlewares.RequestHost{}
	serverRoute := &types.ServerRoute{Route: route}
	rules := &Rules{Route: serverRoute}

	expression := "Host:foo.bar"
	routeResult, err := rules.Parse(expression)
	require.NoError(t, err, "Error while building route for %s", expression)

	request := testhelpers.MustNewRequest(http.MethodGet, "http://foo.bar", nil)
	var routeMatch bool
	reqHostMid.ServeHTTP(nil, request, func(w http.ResponseWriter, r *http.Request) {
		routeMatch = routeResult.Match(r, &mux.RouteMatch{Route: routeResult})
	})

	assert.True(t, routeMatch, "Rule %s don't match.", expression)
}

func TestParseTwoRules(t *testing.T) {
	router := mux.NewRouter()
	route := router.NewRoute()
	serverRoute := &types.ServerRoute{Route: route}
	reqHostMid := &middlewares.RequestHost{}
	rules := &Rules{Route: serverRoute}

	expression := "Host: Foo.Bar ; Path:/FOObar"
	routeResult, err := rules.Parse(expression)

	require.NoError(t, err, "Error while building route for %s.", expression)

	request := testhelpers.MustNewRequest(http.MethodGet, "http://foo.bar/foobar", nil)
	var routeMatch bool
	reqHostMid.ServeHTTP(nil, request, func(w http.ResponseWriter, r *http.Request) {
		routeMatch = routeResult.Match(r, &mux.RouteMatch{Route: routeResult})
	})

	assert.False(t, routeMatch, "Rule %s don't match.", expression)

	request = testhelpers.MustNewRequest(http.MethodGet, "http://foo.bar/FOObar", nil)
	reqHostMid.ServeHTTP(nil, request, func(w http.ResponseWriter, r *http.Request) {
		routeMatch = routeResult.Match(r, &mux.RouteMatch{Route: routeResult})
	})

	assert.True(t, routeMatch, "Rule %s don't match.", expression)
}

func TestParseDomains(t *testing.T) {
	rules := &Rules{}

	tests := []struct {
		expression string
		domain     []string
	}{
		{
			expression: "Host:foo.bar,test.bar",
			domain:     []string{"foo.bar", "test.bar"},
		},
		{
			expression: "Path:/test",
			domain:     []string{},
		},
		{
			expression: "Host:foo.bar;Path:/test",
			domain:     []string{"foo.bar"},
		},
		{
			expression: "Host: Foo.Bar ;Path:/test",
			domain:     []string{"foo.bar"},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.expression, func(t *testing.T) {
			t.Parallel()

			domains, err := rules.ParseDomains(test.expression)
			require.NoError(t, err, "%s: Error while parsing domain.", test.expression)

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
				assert.Equal(t, match, rt.Match(req, &mux.RouteMatch{}))
			}
		})
	}
}

type fakeHandler struct {
	name string
}

func (h *fakeHandler) ServeHTTP(http.ResponseWriter, *http.Request) {}

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
