package muxer

import (
	"strings"
	"unicode"
)

// IsASCII checks if the given string contains only ASCII characters.
func IsASCII(s string) bool {
	for i := range len(s) {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}

	return true
}

// DoubleWildcardPenalty returns what a double wildcard host expression costs to the default priority
// of its rule: its stars do not count, so that the expression ranks below the single-label wildcard
// and the exact hosts it covers, whatever their lengths. Any other expression costs nothing.
func DoubleWildcardPenalty(hostExpr string) int {
	if strings.HasPrefix(hostExpr, "**.") {
		return len("**")
	}

	return 0
}

// DomainMatchHostExpression returns whether the domain matches the host expression.
// The host expression is either an exact host or a wildcard one:
// "*.example.com" matches the direct subdomains of example.com (exactly one label),
// "**.example.com" matches all its subdomains (one or more labels),
// and none of them matches example.com itself.
func DomainMatchHostExpression(domain string, hostExpr string) bool {
	if suffix, ok := strings.CutPrefix(hostExpr, "**."); ok {
		// At least one label has to precede the suffix.
		return len(domain) > len(suffix)+1 && strings.EqualFold(domain[len(domain)-len(suffix)-1:], "."+suffix)
	}

	if strings.HasPrefix(hostExpr, "*") {
		labels := strings.Split(domain, ".")
		labels[0] = "*"
		return strings.EqualFold(hostExpr, strings.Join(labels, "."))
	}

	return strings.EqualFold(domain, hostExpr)
}
