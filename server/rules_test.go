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
		t.Fatal("Error while building route for Host:foo.bar")
	}

	request, err := http.NewRequest("GET", "http://foo.bar", nil)
	routeMatch := routeResult.Match(request, &mux.RouteMatch{Route: routeResult})

	if routeMatch == false {
		t.Log(err)
		t.Fatal("Rule Host:foo.bar don't match")
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
		t.Fatal("Error while building route for Host:foo.bar;Path:/FOObar")
	}

	request, err := http.NewRequest("GET", "http://foo.bar/foobar", nil)
	routeMatch := routeResult.Match(request, &mux.RouteMatch{Route: routeResult})

	if routeMatch == true {
		t.Log(err)
		t.Fatal("Rule Host:foo.bar;Path:/FOObar don't match")
	}

	request, err = http.NewRequest("GET", "http://foo.bar/FOObar", nil)
	routeMatch = routeResult.Match(request, &mux.RouteMatch{Route: routeResult})

	if routeMatch == false {
		t.Log(err)
		t.Fatal("Rule Host:foo.bar;Path:/FOObar don't match")
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
		t.Fatal("Error while building route for PathPrefix:/foo")
	}
	fooHandler := &fakeHandler{name: "fooHandler"}
	routeFoo.Handler(fooHandler)

	if !router.Match(&http.Request{URL: &url.URL{
		Path: "/foo",
	}}, &mux.RouteMatch{}) {
		t.Fatalf("Error matching route")
	}

	if router.Match(&http.Request{URL: &url.URL{
		Path: "/fo",
	}}, &mux.RouteMatch{}) {
		t.Fatalf("Error matching route")
	}

	multipleRules := &Rules{route: &serverRoute{route: router.NewRoute()}}
	routeFoobar, err := multipleRules.Parse("PathPrefix:/foobar")
	if err != nil {
		t.Fatal("Error while building route for PathPrefix:/foobar")
	}
	foobarHandler := &fakeHandler{name: "foobarHandler"}
	routeFoobar.Handler(foobarHandler)
	if !router.Match(&http.Request{URL: &url.URL{
		Path: "/foo",
	}}, &mux.RouteMatch{}) {
		t.Fatalf("Error matching route")
	}
	fooMatcher := &mux.RouteMatch{}
	if !router.Match(&http.Request{URL: &url.URL{
		Path: "/foobar",
	}}, fooMatcher) {
		t.Fatalf("Error matching route")
	}

	if fooMatcher.Handler == foobarHandler {
		t.Fatalf("Error matching priority")
	}

	if fooMatcher.Handler != fooHandler {
		t.Fatalf("Error matching priority")
	}

	routeFoo.Priority(1)
	routeFoobar.Priority(10)
	router.SortRoutes()

	foobarMatcher := &mux.RouteMatch{}
	if !router.Match(&http.Request{URL: &url.URL{
		Path: "/foobar",
	}}, foobarMatcher) {
		t.Fatalf("Error matching route")
	}

	if foobarMatcher.Handler != foobarHandler {
		t.Fatalf("Error matching priority")
	}

	if foobarMatcher.Handler == fooHandler {
		t.Fatalf("Error matching priority")
	}
}

type fakeHandler struct {
	name string
}

func (h *fakeHandler) ServeHTTP(http.ResponseWriter, *http.Request) {

}
