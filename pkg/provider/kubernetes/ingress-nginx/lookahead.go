package ingressnginx

import (
	"regexp"
	"strings"
)

// nginxRegexPrefix is applied to every ingress-nginx path compiled as a regular
// expression: ingress-nginx matches regex locations case-insensitively (the `~*`
// operator) and anchors them at the start of the path.
const nginxRegexPrefix = "(?i)^"

// pathLiteralClass is the set of characters a path may contain while still being
// a literal, that is, while containing no regular expression construct.
const pathLiteralClass = `A-Za-z0-9._~/-`

var (
	// literalPathPrefix matches a path prefix that is a pure literal. A literal
	// prefix matches a determinate number of bytes, which is what lets the
	// exclusion arm anchor at the same offset the lookahead asserted at. A
	// quantifier would make the prefix variable-length and break that.
	literalPathPrefix = regexp.MustCompile(`^/[` + pathLiteralClass + `]*$`)

	// literalAlternation matches a `|`-separated list of non-empty literals.
	// Restricting the lookahead body to this shape rules out several hazards at
	// once: a capture group whose removal would renumber later groups and
	// silently change a rewrite-target replacement, an empty alternative that
	// would exclude every path, and any construct whose interaction with the
	// translation has not been reasoned about.
	literalAlternation = regexp.MustCompile(`^[` + pathLiteralClass + `]+(?:\|[` + pathLiteralClass + `]+)*$`)
)

// splitNegativeLookahead translates an ingress-nginx path that excludes subpaths
// with a negative lookahead into two halves Go's regexp engine can compile.
//
// Go uses RE2, which has no lookahead support, so a path such as
//
//	/api/licensing/((?!_internal).*)
//
// cannot be compiled at all and its router is never created. A lookahead is a
// zero-width assertion, so the path matches if and only if the pattern without
// the assertion matches and the excluded alternation does not match at the
// offset the assertion guarded. That gives two compilable halves:
//
//	keep    /api/licensing/(.*)
//	exclude /api/licensing/(?:_internal)
//
// which a caller combines as PathRegexp(keep) && !PathRegexp(exclude).
//
// Removing the assertion is also what preserves rewrite-target capture groups:
// a zero-width assertion contributes no span, so every group keeps its number
// and its text.
//
// ok reports whether path had the shape this translation is proven correct for.
// It is false both when there is no lookahead to translate and when the pattern
// falls outside that shape; a caller must not translate in either case.
func splitNegativeLookahead(path string) (keep, exclude string, ok bool) {
	// Exactly one negative lookahead. Zero means there is nothing to translate.
	// Several exclusions could become several negated arms, but each one
	// lengthens the rule, and rule length is how Traefik derives router
	// priority, so that is deliberately out of scope here.
	if strings.Count(path, "(?!") != 1 {
		return "", "", false
	}

	// No other zero-width assertion. A positive lookahead is not the negation of
	// a match, so no equivalent rewriting exists, and a lookbehind would have to
	// be anchored from the right. `(?<` covers both `(?<!` and `(?<=`.
	if strings.Contains(path, "(?=") || strings.Contains(path, "(?<") {
		return "", "", false
	}

	start := strings.Index(path, "(?!")

	// The lookahead must open a capture group, so that the offset it asserts at
	// is exactly the length of the literal prefix preceding that group.
	if start < 1 || path[start-1] != '(' {
		return "", "", false
	}

	// Locate the lookahead's own closing parenthesis by counting depth. A regular
	// expression such as `\(\?!.*?\)` would stop at the first `)`, which is wrong
	// as soon as the body contains one.
	end := matchingParen(path, start)
	if end < 0 {
		return "", "", false
	}

	prefix := path[:start-1]
	alternation := path[start+len("(?!") : end]

	if !literalPathPrefix.MatchString(prefix) {
		return "", "", false
	}

	if !literalAlternation.MatchString(alternation) {
		return "", "", false
	}

	keep = path[:start] + path[end+1:]

	// The exclusion arm asks exactly what the assertion asked: does the
	// alternation match starting at the end of the prefix? The group is
	// non-capturing so that `|` binds inside it and no unused group is added.
	exclude = prefix + "(?:" + alternation + ")"

	// Never hand back a translation that cannot be compiled, whatever the checks
	// above may have missed.
	for _, expr := range []string{nginxRegexPrefix + keep, nginxRegexPrefix + exclude} {
		if _, err := regexp.Compile(expr); err != nil {
			return "", "", false
		}
	}

	return keep, exclude, true
}

// matchingParen returns the index of the parenthesis closing the one at start,
// or -1 when it is unbalanced. Escaped parentheses and parentheses inside a
// character class are ignored, so the scan does not rely on the expression
// having been validated beforehand.
func matchingParen(expr string, start int) int {
	var depth int
	var inClass bool

	for i := start; i < len(expr); i++ {
		switch expr[i] {
		case '\\':
			// Skip the escaped byte so that `\(` never changes the depth.
			i++
		case '[':
			inClass = true
		case ']':
			inClass = false
		case '(':
			if !inClass {
				depth++
			}
		case ')':
			if !inClass {
				depth--
				if depth == 0 {
					return i
				}
			}
		}
	}

	return -1
}
