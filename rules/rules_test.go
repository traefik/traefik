package rules

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/containous/mux"
	"github.com/containous/traefik/testhelpers"
	"github.com/containous/traefik/types"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHashRangeRule(t *testing.T) {

	testCases := []struct {
		desc        string
		expectMatch bool
		request     func() *http.Request
		rule        string
	}{
		{
			desc: "Match: header in range",
			request: func() *http.Request {
				request := testhelpers.MustNewRequest(http.MethodGet, "http://foo.bar", nil)
				request.Header.Add("x-partitionheader", "hello")
				return request
			},
			expectMatch: true,
			rule:        "HashedRange: type:header value:x-partitionheader match:50-200 range:0-300",
		},
		{
			desc: "Don't match: header not in range",
			request: func() *http.Request {
				request := testhelpers.MustNewRequest(http.MethodGet, "http://foo.bar", nil)
				request.Header.Add("x-partitionheader", "hello outside range please")
				return request
			},
			expectMatch: false,
			rule:        "HashedRange: type:header value:x-partitionheader match:50-200 range:0-300",
		},
		{
			desc: "Error case: missing part",
			request: func() *http.Request {
				request := testhelpers.MustNewRequest(http.MethodGet, "http://foo.bar", nil)
				return request
			},
			expectMatch: false,
			rule:        "HashedRange: type:header match:50-200 range:0-300",
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			router := mux.NewRouter()
			route := router.NewRoute()
			serverRoute := &types.ServerRoute{Route: route}
			rules := &Rules{Route: serverRoute}
			routeResult, err := rules.Parse(test.rule)
			require.NoError(t, err, "Error while building route for %s", test.rule)

			routeMatch := routeResult.Match(test.request(), &mux.RouteMatch{Route: routeResult})

			if test.expectMatch {
				assert.True(t, routeMatch, "Rule %s don't match.", test.rule)
			} else {
				assert.False(t, routeMatch, "Rule %s matched when not expected.", test.rule)

			}
		})
	}
}

func TestHashRangeRuleDistribution(t *testing.T) {
	router := mux.NewRouter()
	route := router.NewRoute()
	serverRoute := &types.ServerRoute{Route: route}
	rules := &Rules{Route: serverRoute}

	expression := "HashedRange: type:header value:x-partitionheader match:0-100 range:0-500"
	routeResult, err := rules.Parse(expression)
	require.NoError(t, err, "Error while building route for %s", expression)

	i := 0
	matches := 0
	attempts := 1000
	allowedVariance := (float64(attempts) / 100) * 5
	for i < attempts {
		request := testhelpers.MustNewRequest(http.MethodGet, "http://foo.bar", nil)
		request.Header.Add("x-partitionheader", uuid.NewV4().String())
		routeMatch := routeResult.Match(request, &mux.RouteMatch{Route: routeResult})
		if routeMatch {
			matches++
		}
		i++
	}

	// For a range 0-500 matching 0-100 we expect 1/5th of requests to match this rule
	expectedEvenDistribution := float64(attempts) / 5
	lowAcceptableMatch := expectedEvenDistribution - allowedVariance
	highAcceptableMatch := expectedEvenDistribution + allowedVariance
	if float64(matches) < lowAcceptableMatch || float64(matches) > highAcceptableMatch {
		assert.Fail(t, fmt.Sprintf("Matched attempts: %v fell outside acceptable range: %v -> %v for even distrubtion", matches, lowAcceptableMatch, highAcceptableMatch))
	}
}

func TestParseOneRule(t *testing.T) {
	router := mux.NewRouter()
	route := router.NewRoute()
	serverRoute := &types.ServerRoute{Route: route}
	rules := &Rules{Route: serverRoute}

	expression := "Host:foo.bar"
	routeResult, err := rules.Parse(expression)
	require.NoError(t, err, "Error while building route for %s", expression)

	request := testhelpers.MustNewRequest(http.MethodGet, "http://foo.bar", nil)
	routeMatch := routeResult.Match(request, &mux.RouteMatch{Route: routeResult})

	assert.True(t, routeMatch, "Rule %s don't match.", expression)
}

func TestParseTwoRules(t *testing.T) {
	router := mux.NewRouter()
	route := router.NewRoute()
	serverRoute := &types.ServerRoute{Route: route}
	rules := &Rules{Route: serverRoute}

	expression := "Host: Foo.Bar ; Path:/FOObar"
	routeResult, err := rules.Parse(expression)

	require.NoError(t, err, "Error while building route for %s.", expression)

	request := testhelpers.MustNewRequest(http.MethodGet, "http://foo.bar/foobar", nil)
	routeMatch := routeResult.Match(request, &mux.RouteMatch{Route: routeResult})

	assert.False(t, routeMatch, "Rule %s don't match.", expression)

	request = testhelpers.MustNewRequest(http.MethodGet, "http://foo.bar/FOObar", nil)
	routeMatch = routeResult.Match(request, &mux.RouteMatch{Route: routeResult})

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
