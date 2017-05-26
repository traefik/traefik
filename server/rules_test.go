package server

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/containous/mux"
)

func TestParseOneRule(t *testing.T) {
	router := mux.NewRouter()
	route := router.NewRoute()
	serverRoute := &serverRoute{route: route}
	rules := &Rules{route: serverRoute}

	expression := "Host:foo.bar"
	routeResult, err := rules.Parse(expression)

	if err != nil {
		t.Fatalf("Error while building route for Host:foo.bar: %s", err)
	}

	request, err := http.NewRequest("GET", "http://foo.bar", nil)
	routeMatch := routeResult.Match(request, &mux.RouteMatch{Route: routeResult})

	if !routeMatch {
		t.Fatalf("Rule Host:foo.bar don't match: %s", err)
	}
}

func TestParseTwoRules(t *testing.T) {
	router := mux.NewRouter()
	route := router.NewRoute()
	serverRoute := &serverRoute{route: route}
	rules := &Rules{route: serverRoute}

	expression := "Host: Foo.Bar ; Path:/FOObar"
	routeResult, err := rules.Parse(expression)

	if err != nil {
		t.Fatalf("Error while building route for Host:foo.bar;Path:/FOObar: %s", err)
	}

	request, err := http.NewRequest("GET", "http://foo.bar/foobar", nil)
	routeMatch := routeResult.Match(request, &mux.RouteMatch{Route: routeResult})

	if routeMatch {
		t.Fatalf("Rule Host:foo.bar;Path:/FOObar don't match: %s", err)
	}

	request, err = http.NewRequest("GET", "http://foo.bar/FOObar", nil)
	routeMatch = routeResult.Match(request, &mux.RouteMatch{Route: routeResult})

	if !routeMatch {
		t.Fatalf("Rule Host:foo.bar;Path:/FOObar don't match: %s", err)
	}
}

func TestParseDomains(t *testing.T) {
	rules := &Rules{}
	expressionsSlice := []string{
		"Host:foo.bar,test.bar",
		"Path:/test",
		"Host:foo.bar;Path:/test",
		"Host: Foo.Bar ;Path:/test",
	}
	domainsSlice := [][]string{
		{"foo.bar", "test.bar"},
		{},
		{"foo.bar"},
		{"foo.bar"},
	}
	for i, expression := range expressionsSlice {
		domains, err := rules.ParseDomains(expression)
		if err != nil {
			t.Fatalf("Error while parsing domains: %v", err)
		}
		if !reflect.DeepEqual(domains, domainsSlice[i]) {
			t.Fatalf("Error parsing domains: expected %+v, got %+v", domainsSlice[i], domains)
		}
	}
}

func TestPriorites(t *testing.T) {
	router := mux.NewRouter()
	router.StrictSlash(true)
	rules := &Rules{route: &serverRoute{route: router.NewRoute()}}
	routeFoo, err := rules.Parse("PathPrefix:/foo")
	if err != nil {
		t.Fatalf("Error while building route for PathPrefix:/foo: %s", err)
	}
	fooHandler := &fakeHandler{name: "fooHandler"}
	routeFoo.Handler(fooHandler)

	if !router.Match(&http.Request{URL: &url.URL{
		Path: "/foo",
	}}, &mux.RouteMatch{}) {
		t.Fatal("Error matching route")
	}

	if router.Match(&http.Request{URL: &url.URL{
		Path: "/fo",
	}}, &mux.RouteMatch{}) {
		t.Fatal("Error matching route")
	}

	multipleRules := &Rules{route: &serverRoute{route: router.NewRoute()}}
	routeFoobar, err := multipleRules.Parse("PathPrefix:/foobar")
	if err != nil {
		t.Fatalf("Error while building route for PathPrefix:/foobar: %s", err)
	}
	foobarHandler := &fakeHandler{name: "foobarHandler"}
	routeFoobar.Handler(foobarHandler)
	if !router.Match(&http.Request{URL: &url.URL{
		Path: "/foo",
	}}, &mux.RouteMatch{}) {
		t.Fatal("Error matching route")
	}
	fooMatcher := &mux.RouteMatch{}
	if !router.Match(&http.Request{URL: &url.URL{
		Path: "/foobar",
	}}, fooMatcher) {
		t.Fatal("Error matching route")
	}

	if fooMatcher.Handler == foobarHandler {
		t.Fatal("Error matching priority")
	}

	if fooMatcher.Handler != fooHandler {
		t.Fatal("Error matching priority")
	}

	routeFoo.Priority(1)
	routeFoobar.Priority(10)
	router.SortRoutes()

	foobarMatcher := &mux.RouteMatch{}
	if !router.Match(&http.Request{URL: &url.URL{
		Path: "/foobar",
	}}, foobarMatcher) {
		t.Fatal("Error matching route")
	}

	if foobarMatcher.Handler != foobarHandler {
		t.Fatal("Error matching priority")
	}

	if foobarMatcher.Handler == fooHandler {
		t.Fatal("Error matching priority")
	}
}

type fakeHandler struct {
	name string
}

func (h *fakeHandler) ServeHTTP(http.ResponseWriter, *http.Request) {}
