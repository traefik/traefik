package tcp

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"sort"
	"strings"

	"github.com/traefik/traefik/v2/pkg/ip"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/rules"
	"github.com/traefik/traefik/v2/pkg/types"
	"github.com/vulcand/predicate"
)

var funcs = map[string]func(*matchersTree, ...string) error{
	"HostSNI":  hostSNI,
	"ClientIP": clientIP,
}

// ParseHostSNI extracts the HostSNIs declared in a rule.
// This is a first naive implementation used in TCP routing.
func ParseHostSNI(rule string) ([]string, error) {
	var matchers []string
	for matcher := range funcs {
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
		return nil, errors.New("cannot parse")
	}

	return buildTree().ParseMatchers([]string{"HostSNI"}), nil
}

// ConnData contains TCP connection metadata.
type ConnData struct {
	serverName string
	remoteIP   string
}

// NewConnData builds a connData struct from the given parameters.
func NewConnData(serverName string, conn WriteCloser) (ConnData, error) {
	ip, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		return ConnData{}, fmt.Errorf("error while parsing remote address %q: %w", conn.RemoteAddr().String(), err)
	}

	// as per https://datatracker.ietf.org/doc/html/rfc6066:
	// > The hostname is represented as a byte string using ASCII encoding without a trailing dot.
	// so there is no need to trim a potential trailing dot
	serverName = types.CanonicalDomain(serverName)

	return ConnData{
		serverName: types.CanonicalDomain(serverName),
		remoteIP:   ip,
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
	for matcherName := range funcs {
		matcherNames = append(matcherNames, matcherName)
	}

	parser, err := rules.NewParser(matcherNames)
	if err != nil {
		return nil, fmt.Errorf("error while creating rules parser: %w", err)
	}

	return &Muxer{parser: parser}, nil
}

// Match returns the handler of the first route matching the connection metadata.
func (m Muxer) Match(meta ConnData) Handler {
	for _, route := range m.routes {
		if route.matchers.match(meta) {
			return route.handler
		}
	}
	return nil
}

// AddRoute adds a new route, associated to the given handler, at the given
// priority, to the muxer.
func (m *Muxer) AddRoute(rule string, priority int, handler Handler) error {
	// Special case for when the catchAll fallback is present.
	// When no user-defined priority is found, the lowest computable priority minus one is used,
	// in order to make the fallback the last to be evaluated.
	if priority == 0 && rule == "HostSNI(`*`)" {
		priority = -1
	}

	// Default value, which means the user has not set it, so we'll compute it.
	if priority == 0 {
		priority = len(rule)
	}

	parse, err := m.parser.Parse(rule)
	if err != nil {
		return fmt.Errorf("error while parsing rule %s: %w", rule, err)
	}

	buildTree, ok := parse.(rules.TreeBuilder)
	if !ok {
		return fmt.Errorf("error while parsing rule %s", rule)
	}

	var matchers matchersTree
	err = addRule(&matchers, buildTree())
	if err != nil {
		return err
	}

	newRoute := &route{
		handler:  handler,
		priority: priority,
		matchers: matchers,
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

		err = funcs[rule.Matcher](tree, rule.Value...)
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

func (m *Muxer) hasRoutes() bool {
	return len(m.routes) > 0
}

type routes []*route

func (r routes) Len() int      { return len(r) }
func (r routes) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r routes) Less(i, j int) bool {
	return r[i].priority > r[j].priority
}

// route holds the matchers to match TCP route,
// and the handler that will serve the connection.
type route struct {
	// The matchers tree structure reflecting the rule.
	matchers matchersTree
	// The handler responsible for handling the route.
	handler Handler

	// Used to disambiguate between two (or more) rules that would both match for a
	// given request.
	// Defaults to 0.
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
		log.WithoutContext().Warnf("Rule matcher is nil")
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
		log.WithoutContext().Warnf("Invalid rule operator %s", m.operator)
		return false
	}
}

func clientIP(tree *matchersTree, clientIPs ...string) error {
	checker, err := ip.NewChecker(clientIPs)
	if err != nil {
		return fmt.Errorf("could not initialize IP Checker for \"ClientIP\" matcher: %w", err)
	}

	tree.matcher = func(meta ConnData) bool {
		if meta.remoteIP == "" {
			return false
		}

		ok, err := checker.Contains(meta.remoteIP)
		if err != nil {
			log.WithoutContext().Warnf("\"ClientIP\" matcher: could not match remote address : %w", err)
			return false
		}
		if ok {
			return true
		}

		return false
	}

	return nil
}

var almostFQDN = regexp.MustCompile(`^[[:alnum:]\.-]+$`)

// hostSNI checks if the SNI Host of the connection match the matcher host.
func hostSNI(tree *matchersTree, hosts ...string) error {
	if len(hosts) == 0 {
		return fmt.Errorf("empty value for \"HostSNI\" matcher is not allowed")
	}

	for i, host := range hosts {
		// Special case to allow global wildcard
		if host == "*" {
			continue
		}

		if !almostFQDN.MatchString(host) {
			return fmt.Errorf("invalid value for \"HostSNI\" matcher, %q is not a valid hostname", host)
		}

		hosts[i] = strings.ToLower(host)
	}

	tree.matcher = func(meta ConnData) bool {
		// Since a HostSNI(`*`) rule has been provided as catchAll for non-TLS TCP,
		// it allows matching with an empty serverName.
		// Which is why we make sure to take that case into account before before
		// checking meta.serverName.
		if hosts[0] == "*" {
			return true
		}

		if meta.serverName == "" {
			return false
		}

		for _, host := range hosts {
			if host == "*" {
				return true
			}

			if host == meta.serverName {
				return true
			}

			// trim trailing period in case of FQDN
			host = strings.TrimSuffix(host, ".")
			if host == meta.serverName {
				return true
			}
		}

		return false
	}

	return nil
}
