package tcp

import (
	"fmt"
	"net"
	"sort"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v2/pkg/rules"
	"github.com/traefik/traefik/v2/pkg/tcp"
	"github.com/traefik/traefik/v2/pkg/types"
	"github.com/vulcand/predicate"
)

// ParseHostSNI extracts the HostSNIs declared in a rule.
// This is a first naive implementation used in TCP routing.
func ParseHostSNI(rule string) ([]string, error) {
	var matchers []string
	for matcher := range tcpFuncs {
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

	return buildTree().ParseMatchers([]string{"HostSNI"}), nil
}

// ConnData contains TCP connection metadata.
type ConnData struct {
	serverName string
	remoteIP   string
	alpnProtos []string
}

// NewConnData builds a connData struct from the given parameters.
func NewConnData(serverName string, conn tcp.WriteCloser, alpnProtos []string) (ConnData, error) {
	remoteIP, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		return ConnData{}, fmt.Errorf("error while parsing remote address %q: %w", conn.RemoteAddr().String(), err)
	}

	// as per https://datatracker.ietf.org/doc/html/rfc6066:
	// > The hostname is represented as a byte string using ASCII encoding without a trailing dot.
	// so there is no need to trim a potential trailing dot
	serverName = types.CanonicalDomain(serverName)

	return ConnData{
		serverName: types.CanonicalDomain(serverName),
		remoteIP:   remoteIP,
		alpnProtos: alpnProtos,
	}, nil
}

// Muxer defines a muxer that handles TCP routing with rules.
type Muxer struct {
	routes []*route
	parser predicate.Parser
}

// NewMuxer returns a TCP muxer.
func NewMuxer() (*Muxer, error) {
	var matcherNames []string
	for matcherName := range tcpFuncs {
		matcherNames = append(matcherNames, matcherName)
	}

	parser, err := rules.NewParser(matcherNames)
	if err != nil {
		return nil, fmt.Errorf("error while creating rules parser: %w", err)
	}

	return &Muxer{parser: parser}, nil
}

// Match returns the handler of the first route matching the connection metadata,
// and whether the match is exactly from the rule HostSNI(*).
func (m Muxer) Match(meta ConnData) (tcp.Handler, bool) {
	for _, route := range m.routes {
		if route.matchers.match(meta) {
			return route.handler, route.catchAll
		}
	}

	return nil, false
}

// AddRoute adds a new route, associated to the given handler, at the given
// priority, to the muxer.
func (m *Muxer) AddRoute(rule string, priority int, handler tcp.Handler) error {
	parse, err := m.parser.Parse(rule)
	if err != nil {
		return fmt.Errorf("error while parsing rule %s: %w", rule, err)
	}

	buildTree, ok := parse.(rules.TreeBuilder)
	if !ok {
		return fmt.Errorf("error while parsing rule %s", rule)
	}

	ruleTree := buildTree()

	var matchers matchersTree
	err = addRule(&matchers, ruleTree)
	if err != nil {
		return err
	}

	var catchAll bool
	if ruleTree.RuleLeft == nil && ruleTree.RuleRight == nil && len(ruleTree.Value) == 1 {
		catchAll = ruleTree.Value[0] == "*" && strings.EqualFold(ruleTree.Matcher, "HostSNI")
	}

	// Special case for when the catchAll fallback is present.
	// When no user-defined priority is found, the lowest computable priority minus one is used,
	// in order to make the fallback the last to be evaluated.
	if priority == 0 && catchAll {
		priority = -1
	}

	// Default value, which means the user has not set it, so we'll compute it.
	if priority == 0 {
		priority = len(rule)
	}

	newRoute := &route{
		handler:  handler,
		matchers: matchers,
		catchAll: catchAll,
		priority: priority,
	}
	m.routes = append(m.routes, newRoute)

	sort.Sort(routes(m.routes))

	return nil
}

func addRule(tree *matchersTree, rule *rules.Tree) error {
	switch rule.Matcher {
	case "and", "or":
		tree.operator = rule.Matcher
		tree.left = &matchersTree{}
		err := addRule(tree.left, rule.RuleLeft)
		if err != nil {
			return err
		}

		tree.right = &matchersTree{}
		return addRule(tree.right, rule.RuleRight)
	default:
		err := rules.CheckRule(rule)
		if err != nil {
			return err
		}

		err = tcpFuncs[rule.Matcher](tree, rule.Value...)
		if err != nil {
			return err
		}

		if rule.Not {
			matcherFunc := tree.matcher
			tree.matcher = func(meta ConnData) bool {
				return !matcherFunc(meta)
			}
		}
	}

	return nil
}

// HasRoutes returns whether the muxer has routes.
func (m *Muxer) HasRoutes() bool {
	return len(m.routes) > 0
}

// routes implements sort.Interface.
type routes []*route

// Len implements sort.Interface.
func (r routes) Len() int { return len(r) }

// Swap implements sort.Interface.
func (r routes) Swap(i, j int) { r[i], r[j] = r[j], r[i] }

// Less implements sort.Interface.
func (r routes) Less(i, j int) bool { return r[i].priority > r[j].priority }

// route holds the matchers to match TCP route,
// and the handler that will serve the connection.
type route struct {
	// matchers tree structure reflecting the rule.
	matchers matchersTree
	// handler responsible for handling the route.
	handler tcp.Handler
	// catchAll indicates whether the route rule has exactly the catchAll value (HostSNI(`*`)).
	catchAll bool
	// priority is used to disambiguate between two (or more) rules that would
	// all match for a given request.
	// Computed from the matching rule length, if not user-set.
	priority int
}

// matcher is a matcher func used to match connection properties.
type matcher func(meta ConnData) bool

// matchersTree represents the matchers tree structure.
type matchersTree struct {
	// If matcher is not nil, it means that this matcherTree is a leaf of the tree.
	// It is therefore mutually exclusive with left and right.
	matcher matcher
	// operator to combine the evaluation of left and right leaves.
	operator string
	// Mutually exclusive with matcher.
	left  *matchersTree
	right *matchersTree
}

func (m *matchersTree) match(meta ConnData) bool {
	if m == nil {
		// This should never happen as it should have been detected during parsing.
		log.Warn().Msg("Rule matcher is nil")
		return false
	}

	if m.matcher != nil {
		return m.matcher(meta)
	}

	switch m.operator {
	case "or":
		return m.left.match(meta) || m.right.match(meta)
	case "and":
		return m.left.match(meta) && m.right.match(meta)
	default:
		// This should never happen as it should have been detected during parsing.
		log.Warn().Str("operator", m.operator).Msg("Invalid rule operator")
		return false
	}
}
