package http

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/rules"
	"github.com/vulcand/predicate"
)

// Muxer handles routing with rules.
type Muxer struct {
	routes         routes
	parser         predicate.Parser
	parserV2       predicate.Parser
	defaultHandler http.Handler
}

// NewMuxer returns a new muxer instance.
func NewMuxer() (*Muxer, error) {
	var matchers []string
	for matcher := range httpFuncs {
		matchers = append(matchers, matcher)
	}

	parser, err := rules.NewParser(matchers)
	if err != nil {
		return nil, fmt.Errorf("error while creating parser: %w", err)
	}

	var matchersV2 []string
	for matcher := range httpFuncsV2 {
		matchersV2 = append(matchersV2, matcher)
	}

	parserV2, err := rules.NewParser(matchersV2)
	if err != nil {
		return nil, fmt.Errorf("error while creating v2 parser: %w", err)
	}

	return &Muxer{
		parser:         parser,
		parserV2:       parserV2,
		defaultHandler: http.NotFoundHandler(),
	}, nil
}

// ServeHTTP forwards the connection to the matching HTTP handler.
// Serves 404 if no handler is found.
func (m *Muxer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	for _, route := range m.routes {
		if route.matchers.match(req) {
			route.handler.ServeHTTP(rw, req)
			return
		}
	}

	m.defaultHandler.ServeHTTP(rw, req)
}

// SetDefaultHandler sets the muxer default handler.
func (m *Muxer) SetDefaultHandler(handler http.Handler) {
	m.defaultHandler = handler
}

// GetRulePriority computes the priority for a given rule.
// The priority is calculated using the length of rule.
func GetRulePriority(rule string) int {
	return len(rule)
}

// AddRoute add a new route to the router.
func (m *Muxer) AddRoute(rule string, syntax string, priority int, handler http.Handler) error {
	var parse interface{}
	var err error
	var matcherFuncs map[string]func(*matchersTree, ...string) error

	switch syntax {
	case "v2":
		parse, err = m.parserV2.Parse(rule)
		if err != nil {
			return fmt.Errorf("error while parsing rule %s: %w", rule, err)
		}

		matcherFuncs = httpFuncsV2
	default:
		parse, err = m.parser.Parse(rule)
		if err != nil {
			return fmt.Errorf("error while parsing rule %s: %w", rule, err)
		}

		matcherFuncs = httpFuncs
	}

	buildTree, ok := parse.(rules.TreeBuilder)
	if !ok {
		return fmt.Errorf("error while parsing rule %s", rule)
	}

	var matchers matchersTree
	err = matchers.addRule(buildTree(), matcherFuncs)
	if err != nil {
		return fmt.Errorf("error while adding rule %s: %w", rule, err)
	}

	m.routes = append(m.routes, &route{
		handler:  handler,
		matchers: matchers,
		priority: priority,
	})

	sort.Sort(m.routes)

	return nil
}

// ParseDomains extract domains from rule.
func ParseDomains(rule string) ([]string, error) {
	var matchers []string
	for matcher := range httpFuncs {
		matchers = append(matchers, matcher)
	}
	for matcher := range httpFuncsV2 {
		matchers = append(matchers, matcher)
	}

	parser, err := rules.NewParser(matchers)
	if err != nil {
		return nil, fmt.Errorf("error while creating parser: %w", err)
	}

	parse, err := parser.Parse(rule)
	if err != nil {
		return nil, fmt.Errorf("error while parsing rule %s: %w", rule, err)
	}

	buildTree, ok := parse.(rules.TreeBuilder)
	if !ok {
		return nil, fmt.Errorf("error while parsing rule %s", rule)
	}

	return buildTree().ParseMatchers([]string{"Host"}), nil
}

// routes implements sort.Interface.
type routes []*route

// Len implements sort.Interface.
func (r routes) Len() int { return len(r) }

// Swap implements sort.Interface.
func (r routes) Swap(i, j int) { r[i], r[j] = r[j], r[i] }

// Less implements sort.Interface.
func (r routes) Less(i, j int) bool { return r[i].priority > r[j].priority }

// route holds the matchers to match HTTP route,
// and the handler that will serve the request.
type route struct {
	// matchers tree structure reflecting the rule.
	matchers matchersTree
	// handler responsible for handling the route.
	handler http.Handler
	// priority is used to disambiguate between two (or more) rules that would all match for a given request.
	// Computed from the matching rule length, if not user-set.
	priority int
}

// matchersTree represents the matchers tree structure.
type matchersTree struct {
	// matcher is a matcher func used to match HTTP request properties.
	// If matcher is not nil, it means that this matcherTree is a leaf of the tree.
	// It is therefore mutually exclusive with left and right.
	matcher func(*http.Request) bool
	// operator to combine the evaluation of left and right leaves.
	operator string
	// Mutually exclusive with matcher.
	left  *matchersTree
	right *matchersTree
}

func (m *matchersTree) match(req *http.Request) bool {
	if m == nil {
		// This should never happen as it should have been detected during parsing.
		log.Warn().Msg("Rule matcher is nil")
		return false
	}

	if m.matcher != nil {
		return m.matcher(req)
	}

	switch m.operator {
	case "or":
		return m.left.match(req) || m.right.match(req)
	case "and":
		return m.left.match(req) && m.right.match(req)
	default:
		// This should never happen as it should have been detected during parsing.
		log.Warn().Str("operator", m.operator).Msg("Invalid rule operator")
		return false
	}
}

type matcherFuncs map[string]func(*matchersTree, ...string) error

func (m *matchersTree) addRule(rule *rules.Tree, funcs matcherFuncs) error {
	switch rule.Matcher {
	case "and", "or":
		m.operator = rule.Matcher
		m.left = &matchersTree{}
		err := m.left.addRule(rule.RuleLeft, funcs)
		if err != nil {
			return fmt.Errorf("error while adding rule %s: %w", rule.Matcher, err)
		}

		m.right = &matchersTree{}
		return m.right.addRule(rule.RuleRight, funcs)
	default:
		err := rules.CheckRule(rule)
		if err != nil {
			return fmt.Errorf("error while checking rule %s: %w", rule.Matcher, err)
		}

		err = funcs[rule.Matcher](m, rule.Value...)
		if err != nil {
			return fmt.Errorf("error while adding rule %s: %w", rule.Matcher, err)
		}

		if rule.Not {
			matcherFunc := m.matcher
			m.matcher = func(req *http.Request) bool {
				return !matcherFunc(req)
			}
		}
	}

	return nil
}
