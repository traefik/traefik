package route

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

type matcher interface {
	match(*http.Request) *match
	setMatch(match *match)

	canMerge(matcher) bool
	merge(matcher) (matcher, error)

	canChain(matcher) bool
	chain(matcher) (matcher, error)
}

func hostTrieMatcher(hostname string) (matcher, error) {
	return newTrieMatcher(strings.ToLower(hostname), &hostMapper{}, &match{})
}

func hostRegexpMatcher(hostname string) (matcher, error) {
	return newRegexpMatcher(strings.ToLower(hostname), &hostMapper{}, &match{})
}

func methodTrieMatcher(method string) (matcher, error) {
	return newTrieMatcher(method, &methodMapper{}, &match{})
}

func methodRegexpMatcher(method string) (matcher, error) {
	return newRegexpMatcher(method, &methodMapper{}, &match{})
}

func pathTrieMatcher(path string) (matcher, error) {
	return newTrieMatcher(path, &pathMapper{}, &match{})
}

func pathRegexpMatcher(path string) (matcher, error) {
	return newRegexpMatcher(path, &pathMapper{}, &match{})
}

func headerTrieMatcher(name, value string) (matcher, error) {
	return newTrieMatcher(value, &headerMapper{header: name}, &match{})
}

func headerRegexpMatcher(name, value string) (matcher, error) {
	return newRegexpMatcher(value, &headerMapper{header: name}, &match{})
}

type match struct {
	val interface{}
}

type andMatcher struct {
	a matcher
	b matcher
}

func newAndMatcher(a, b matcher) matcher {
	if a.canChain(b) {
		m, err := a.chain(b)
		if err == nil {
			return m
		}
	}
	return &andMatcher{
		a: a, b: b,
	}
}

func (a *andMatcher) canChain(matcher) bool {
	return false
}

func (a *andMatcher) chain(matcher) (matcher, error) {
	return nil, fmt.Errorf("not supported")
}

func (a *andMatcher) String() string {
	return fmt.Sprintf("andMatcher(%v, %v)", a.a, a.b)
}

func (a *andMatcher) setMatch(m *match) {
	a.a.setMatch(m)
	a.b.setMatch(m)
}

func (a *andMatcher) canMerge(o matcher) bool {
	return false
}

func (a *andMatcher) merge(o matcher) (matcher, error) {
	return nil, fmt.Errorf("Method not supported")
}

func (a *andMatcher) match(req *http.Request) *match {
	result := a.a.match(req)
	if result == nil {
		return nil
	}
	return a.b.match(req)
}

// Regular expression matcher, takes a regular expression and requestMapper
type regexpMatcher struct {
	// Uses this mapper to extract a string from a request to match against
	mapper requestMapper
	// Compiled regular expression
	expr *regexp.Regexp
	// match result
	result *match
}

func (r *regexpMatcher) canChain(matcher) bool {
	return false
}

func (r *regexpMatcher) chain(matcher) (matcher, error) {
	return nil, fmt.Errorf("not supported")
}

func (m *regexpMatcher) String() string {
	return fmt.Sprintf("regexpMatcher(%v)", m.expr)
}

func (m *regexpMatcher) setMatch(result *match) {
	m.result = result
}

func newRegexpMatcher(expr string, mapper requestMapper, m *match) (matcher, error) {
	r, err := regexp.Compile(expr)

	if err != nil {
		return nil, fmt.Errorf("Bad regular expression: %s %s", expr, err)
	}
	return &regexpMatcher{expr: r, mapper: mapper, result: m}, nil
}

func (m *regexpMatcher) canMerge(matcher) bool {
	return false
}

func (m *regexpMatcher) merge(matcher) (matcher, error) {
	return nil, fmt.Errorf("Method not supported")
}

func (m *regexpMatcher) match(req *http.Request) *match {
	if m.expr.MatchString(m.mapper.mapRequest(req)) {
		return m.result
	}
	return nil
}
