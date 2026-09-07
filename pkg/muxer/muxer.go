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

// DomainMatchHostExpression returns true if the domain matches the host expression.
// The host expression can be a wildcard, in which case it will match any subdomain of the domain.
// For example, if the domain is "example.com" and the host expression is "*.example.com", this function will return true.
// If the host expression is "example.com", this function will also return true.
func DomainMatchHostExpression(domain string, hostExpr string) bool {
	if strings.HasPrefix(hostExpr, "*") {
		labels := strings.Split(domain, ".")
		labels[0] = "*"
		return strings.EqualFold(hostExpr, strings.Join(labels, "."))
	}

	return strings.EqualFold(domain, hostExpr)
}
