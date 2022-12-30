package http

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/traefik/traefik/v2/pkg/rules"
	"github.com/vulcand/predicate"
)

// Muxer handles routing with rules.
type Muxer struct {
	*mux.Router
	parser predicate.Parser
}

// NewMuxer returns a new muxer instance.
func NewMuxer() (*Muxer, error) {
	var matchers []string
	for matcher := range httpFuncs {
		matchers = append(matchers, matcher)
	}

	parser, err := rules.NewParser(matchers)
	if err != nil {
		return nil, err
	}

	return &Muxer{
		Router: mux.NewRouter().SkipClean(true),
		parser: parser,
	}, nil
}

// AddRoute add a new route to the router.
func (r *Muxer) AddRoute(rule string, priority int, handler http.Handler) error {
	parse, err := r.parser.Parse(rule)
	if err != nil {
		return fmt.Errorf("error while parsing rule %s: %w", rule, err)
	}

	buildTree, ok := parse.(rules.TreeBuilder)
	if !ok {
		return fmt.Errorf("error while parsing rule %s", rule)
	}

	if priority == 0 {
		priority = len(rule)
	}

	route := r.NewRoute().Handler(handler).Priority(priority)

	err = addRuleOnRoute(route, buildTree())
	if err != nil {
		route.BuildOnly()
		return err
	}

	return nil
}

func addRuleOnRouter(router *mux.Router, rule *rules.Tree) error {
	switch rule.Matcher {
	case "and":
		route := router.NewRoute()
		err := addRuleOnRoute(route, rule.RuleLeft)
		if err != nil {
			return err
		}

		return addRuleOnRoute(route, rule.RuleRight)
	case "or":
		err := addRuleOnRouter(router, rule.RuleLeft)
		if err != nil {
			return err
		}

		return addRuleOnRouter(router, rule.RuleRight)
	default:
		err := rules.CheckRule(rule)
		if err != nil {
			return err
		}

		if rule.Not {
			return not(httpFuncs[rule.Matcher])(router.NewRoute(), rule.Value...)
		}

		return httpFuncs[rule.Matcher](router.NewRoute(), rule.Value...)
	}
}

func addRuleOnRoute(route *mux.Route, rule *rules.Tree) error {
	switch rule.Matcher {
	case "and":
		err := addRuleOnRoute(route, rule.RuleLeft)
		if err != nil {
			return err
		}

		return addRuleOnRoute(route, rule.RuleRight)
	case "or":
		subRouter := route.Subrouter()

		err := addRuleOnRouter(subRouter, rule.RuleLeft)
		if err != nil {
			return err
		}

		return addRuleOnRouter(subRouter, rule.RuleRight)
	default:
		err := rules.CheckRule(rule)
		if err != nil {
			return err
		}

		if rule.Not {
			return not(httpFuncs[rule.Matcher])(route, rule.Value...)
		}

		return httpFuncs[rule.Matcher](route, rule.Value...)
	}
}

func not(m func(*mux.Route, ...string) error) func(*mux.Route, ...string) error {
	return func(r *mux.Route, v ...string) error {
		router := mux.NewRouter()

		err := m(router.NewRoute(), v...)
		if err != nil {
			return err
		}

		r.MatcherFunc(func(req *http.Request, ma *mux.RouteMatch) bool {
			return !router.Match(req, ma)
		})

		return nil
	}
}

// ParseDomains extract domains from rule.
func ParseDomains(rule string) ([]string, error) {
	var matchers []string
	for matcher := range httpFuncs {
		matchers = append(matchers, matcher)
	}

	parser, err := rules.NewParser(matchers)
	if err != nil {
		return nil, err
	}

	parse, err := parser.Parse(rule)
	if err != nil {
		return nil, err
	}

	buildTree, ok := parse.(rules.TreeBuilder)
	if !ok {
		return nil, fmt.Errorf("error while parsing rule %s", rule)
	}

	return buildTree().ParseMatchers([]string{"Host"}), nil
}
