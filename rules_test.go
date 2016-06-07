package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"testing"
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

	expression := "Host:foo.bar;Path:/foobar"
	routeResult, err := rules.Parse(expression)

	if err != nil {
		t.Fatal("Error while building route for Host:foo.bar;Path:/foobar")
	}

	request, err := http.NewRequest("GET", "http://foo.bar/foobar", nil)
	routeMatch := routeResult.Match(request, &mux.RouteMatch{Route: routeResult})

	if routeMatch == false {
		t.Log(err)
		t.Fatal("Rule Host:foo.bar;Path:/foobar don't match")
	}
}
