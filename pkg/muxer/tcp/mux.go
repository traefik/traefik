package tcp

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/go-acme/lego/v4/challenge/tlsalpn01"
	"github.com/traefik/traefik/v2/pkg/ip"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/rules"
	"github.com/traefik/traefik/v2/pkg/tcp"
	"github.com/traefik/traefik/v2/pkg/types"
	"github.com/vulcand/predicate"
)

var tcpFuncs = map[string]func(*matchersTree, ...string) error{
	"HostSNI":       hostSNI,
	"HostSNIRegexp": hostSNIRegexp,
	"ClientIP":      clientIP,
	"ALPN":          alpn,
}

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
			log.WithoutContext().Warnf("\"ClientIP\" matcher: could not match remote address: %v", err)
			return false
		}
		return ok
	}

	return nil
}

// alpn checks if any of the connection ALPN protocols matches one of the matcher protocols.
func alpn(tree *matchersTree, protos ...string) error {
	if len(protos) == 0 {
		return errors.New("empty value for \"ALPN\" matcher is not allowed")
	}

	for _, proto := range protos {
		if proto == tlsalpn01.ACMETLS1Protocol {
			return fmt.Errorf("invalid protocol value for \"ALPN\" matcher, %q is not allowed", proto)
		}
	}

	tree.matcher = func(meta ConnData) bool {
		for _, proto := range meta.alpnProtos {
			for _, filter := range protos {
				if proto == filter {
					return true
				}
			}
		}

		return false
	}

	return nil
}

var hostOrIP = regexp.MustCompile(`^[[:alnum:]\.\-\:]+$`)

// hostSNI checks if the SNI Host of the connection match the matcher host.
func hostSNI(tree *matchersTree, hosts ...string) error {
	if len(hosts) == 0 {
		return errors.New("empty value for \"HostSNI\" matcher is not allowed")
	}

	for i, host := range hosts {
		// Special case to allow global wildcard
		if host == "*" {
			continue
		}

		if !hostOrIP.MatchString(host) {
			return fmt.Errorf("invalid value for \"HostSNI\" matcher, %q is not a valid hostname or IP", host)
		}

		hosts[i] = strings.ToLower(host)
	}

	tree.matcher = func(meta ConnData) bool {
		// Since a HostSNI(`*`) rule has been provided as catchAll for non-TLS TCP,
		// it allows matching with an empty serverName.
		// Which is why we make sure to take that case into account before
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

// hostSNIRegexp checks if the SNI Host of the connection matches the matcher host regexp.
func hostSNIRegexp(tree *matchersTree, templates ...string) error {
	if len(templates) == 0 {
		return fmt.Errorf("empty value for \"HostSNIRegexp\" matcher is not allowed")
	}

	var regexps []*regexp.Regexp

	for _, template := range templates {
		preparedPattern, err := preparePattern(template)
		if err != nil {
			return fmt.Errorf("invalid pattern value for \"HostSNIRegexp\" matcher, %q is not a valid pattern: %w", template, err)
		}

		regexp, err := regexp.Compile(preparedPattern)
		if err != nil {
			return err
		}

		regexps = append(regexps, regexp)
	}

	tree.matcher = func(meta ConnData) bool {
		for _, regexp := range regexps {
			if regexp.MatchString(meta.serverName) {
				return true
			}
		}

		return false
	}

	return nil
}

// TODO: expose more of containous/mux fork to get rid of the following copied code (https://github.com/containous/mux/blob/8ffa4f6d063c/regexp.go).

// preparePattern builds a regexp pattern from the initial user defined expression.
// This function reuses the code dedicated to host matching of the newRouteRegexp func from the gorilla/mux library.
// https://github.com/containous/mux/tree/8ffa4f6d063c1e2b834a73be6a1515cca3992618.
func preparePattern(template string) (string, error) {
	// Check if it is well-formed.
	idxs, errBraces := braceIndices(template)
	if errBraces != nil {
		return "", errBraces
	}

	defaultPattern := "[^.]+"
	pattern := bytes.NewBufferString("")

	// Host SNI matching is case-insensitive
	_, _ = fmt.Fprint(pattern, "(?i)")

	pattern.WriteByte('^')
	var end int
	for i := 0; i < len(idxs); i += 2 {
		// Set all values we are interested in.
		raw := template[end:idxs[i]]
		end = idxs[i+1]
		parts := strings.SplitN(template[idxs[i]+1:end-1], ":", 2)
		name := parts[0]

		patt := defaultPattern
		if len(parts) == 2 {
			patt = parts[1]
		}

		// Name or pattern can't be empty.
		if name == "" || patt == "" {
			return "", fmt.Errorf("mux: missing name or pattern in %q",
				template[idxs[i]:end])
		}

		// Build the regexp pattern.
		_, _ = fmt.Fprintf(pattern, "%s(?P<%s>%s)", regexp.QuoteMeta(raw), varGroupName(i/2), patt)
	}

	// Add the remaining.
	raw := template[end:]
	pattern.WriteString(regexp.QuoteMeta(raw))
	pattern.WriteByte('$')

	return pattern.String(), nil
}

// varGroupName builds a capturing group name for the indexed variable.
// This function is a copy of varGroupName func from the gorilla/mux library.
// https://github.com/containous/mux/tree/8ffa4f6d063c1e2b834a73be6a1515cca3992618.
func varGroupName(idx int) string {
	return "v" + strconv.Itoa(idx)
}

// braceIndices returns the first level curly brace indices from a string.
// This function is a copy of braceIndices func from the gorilla/mux library.
// https://github.com/containous/mux/tree/8ffa4f6d063c1e2b834a73be6a1515cca3992618.
func braceIndices(s string) ([]int, error) {
	var level, idx int
	var idxs []int
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '{':
			if level++; level == 1 {
				idx = i
			}
		case '}':
			if level--; level == 0 {
				idxs = append(idxs, idx, i+1)
			} else if level < 0 {
				return nil, fmt.Errorf("mux: unbalanced braces in %q", s)
			}
		}
	}
	if level != 0 {
		return nil, fmt.Errorf("mux: unbalanced braces in %q", s)
	}
	return idxs, nil
}
