package muxer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoubleWildcardPenalty(t *testing.T) {
	assert.Equal(t, 0, DoubleWildcardPenalty("example.com"))
	assert.Equal(t, 0, DoubleWildcardPenalty("*"))
	assert.Equal(t, 0, DoubleWildcardPenalty("*.example.com"))
	assert.Equal(t, 2, DoubleWildcardPenalty("**.example.com"))
}

func TestDomainMatchHostExpression(t *testing.T) {
	testCases := []struct {
		desc     string
		domain   string
		hostExpr string
		expected bool
	}{
		{desc: "exact host", domain: "example.com", hostExpr: "example.com", expected: true},
		{desc: "exact host, case insensitive", domain: "Example.COM", hostExpr: "example.com", expected: true},
		{desc: "exact host, other domain", domain: "foo.example.com", hostExpr: "example.com", expected: false},
		{desc: "wildcard, direct subdomain", domain: "foo.example.com", hostExpr: "*.example.com", expected: true},
		{desc: "wildcard, nested subdomain", domain: "bar.foo.example.com", hostExpr: "*.example.com", expected: false},
		{desc: "wildcard, apex", domain: "example.com", hostExpr: "*.example.com", expected: false},
		{desc: "double wildcard, direct subdomain", domain: "foo.example.com", hostExpr: "**.example.com", expected: true},
		{desc: "double wildcard, nested subdomain", domain: "bar.foo.example.com", hostExpr: "**.example.com", expected: true},
		{desc: "double wildcard, case insensitive", domain: "Bar.Foo.Example.com", hostExpr: "**.example.com", expected: true},
		{desc: "double wildcard, apex", domain: "example.com", hostExpr: "**.example.com", expected: false},
		{desc: "double wildcard, empty label", domain: ".example.com", hostExpr: "**.example.com", expected: false},
		{desc: "double wildcard, other domain", domain: "foo.example.org", hostExpr: "**.example.com", expected: false},
		{desc: "double wildcard, suffix within a label", domain: "fooexample.com", hostExpr: "**.example.com", expected: false},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.expected, DomainMatchHostExpression(test.domain, test.hostExpr))
		})
	}
}
