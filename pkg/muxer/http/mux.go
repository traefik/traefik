package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/rules"
)

type matcherBuilderFuncs map[string]matcherBuilderFunc

type matcherBuilderFunc func(*matchersTree, ...string) error

type MatcherFunc func(*http.Request) bool

// Muxer handles routing with rules.
type Muxer struct {
	routes         routes
	parser         SyntaxParser
	defaultHandler http.Handler
}

// NewMuxer returns a new muxer instance.
func NewMuxer(parser SyntaxParser) *Muxer {
	return &Muxer{
		parser:         parser,
		defaultHandler: http.NotFoundHandler(),
	}
}

// ServeHTTP forwards the connection to the matching HTTP handler.
// Serves 404 if no handler is found.
func (m *Muxer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := log.Ctx(req.Context())

	var err error
	req, err = withRoutingPath(req)
	if err != nil {
		logger.Debug().Err(err).Msg("Unable to add routing path to request context")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

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
	matchers, err := m.parser.parse(syntax, rule)
	if err != nil {
		return fmt.Errorf("error while parsing rule %s: %w", rule, err)
	}

	m.routes = append(m.routes, &route{
		handler:  handler,
		matchers: matchers,
		priority: priority,
	})

	sort.Sort(m.routes)

	return nil
}

// reservedCharacters contains the mapping of the percent-encoded form to the ASCII form
// of the reserved characters according to https://datatracker.ietf.org/doc/html/rfc3986#section-2.2.
// By extension to https://datatracker.ietf.org/doc/html/rfc3986#section-2.1 the percent character is also considered a reserved character.
// Because decoding the percent character would change the meaning of the URL.
var reservedCharacters = map[string]rune{
	"%3A": ':',
	"%2F": '/',
	"%3F": '?',
	"%23": '#',
	"%5B": '[',
	"%5D": ']',
	"%40": '@',
	"%21": '!',
	"%24": '$',
	"%26": '&',
	"%27": '\'',
	"%28": '(',
	"%29": ')',
	"%2A": '*',
	"%2B": '+',
	"%2C": ',',
	"%3B": ';',
	"%3D": '=',
	"%25": '%',
}

// getRoutingPath retrieves the routing path from the request context.
// It returns nil if the routing path is not set in the context.
func getRoutingPath(req *http.Request) *string {
	routingPath := req.Context().Value(mux.RoutingPathKey)
	if routingPath != nil {
		rp := routingPath.(string)
		return &rp
	}
	return nil
}

// withRoutingPath decodes non-allowed characters in the EscapedPath and stores it in the request context to be able to use it for routing.
// This allows using the decoded version of the non-allowed characters in the routing rules for a better UX.
// For example, the rule PathPrefix(`/foo bar`) will match the following request path `/foo%20bar`.
func withRoutingPath(req *http.Request) (*http.Request, error) {
	escapedPath := req.URL.EscapedPath()

	var routingPathBuilder strings.Builder
	for i := 0; i < len(escapedPath); i++ {
		if escapedPath[i] != '%' {
			routingPathBuilder.WriteString(string(escapedPath[i]))
			continue
		}

		// This should never happen as the standard library will reject requests containing invalid percent-encodings.
		// This discards URLs with a percent character at the end.
		if i+2 >= len(escapedPath) {
			return nil, errors.New("invalid percent-encoding at the end of the URL path")
		}

		encodedCharacter := escapedPath[i : i+3]
		if _, reserved := reservedCharacters[encodedCharacter]; reserved {
			routingPathBuilder.WriteString(encodedCharacter)
		} else {
			// This should never happen as the standard library will reject requests containing invalid percent-encodings.
			decodedCharacter, err := url.PathUnescape(encodedCharacter)
			if err != nil {
				return nil, errors.New("invalid percent-encoding in URL path")
			}
			routingPathBuilder.WriteString(decodedCharacter)
		}

		i += 2
	}

	return req.WithContext(
		context.WithValue(
			req.Context(),
			mux.RoutingPathKey,
			routingPathBuilder.String(),
		),
	), nil
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
	matcher MatcherFunc
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

func (m *matchersTree) addRule(rule *rules.Tree, funcs matcherBuilderFuncs) error {
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
