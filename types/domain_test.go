package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDomain_ToStrArray(t *testing.T) {
	testCases := []struct {
		desc     string
		domain   Domain
		expected []string
	}{
		{
			desc: "with Main and SANs",
			domain: Domain{
				Main: "foo.com",
				SANs: []string{"bar.foo.com", "bir.foo.com"},
			},
			expected: []string{"foo.com", "bar.foo.com", "bir.foo.com"},
		},
		{
			desc: "without SANs",
			domain: Domain{
				Main: "foo.com",
			},
			expected: []string{"foo.com"},
		},
		{
			desc: "without Main",
			domain: Domain{
				SANs: []string{"bar.foo.com", "bir.foo.com"},
			},
			expected: []string{"bar.foo.com", "bir.foo.com"},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			domains := test.domain.ToStrArray()
			assert.EqualValues(t, test.expected, domains)
		})
	}
}

func TestDomain_Set(t *testing.T) {
	testCases := []struct {
		desc       string
		rawDomains []string
		expected   Domain
	}{
		{
			desc:       "with 3 domains",
			rawDomains: []string{"foo.com", "bar.foo.com", "bir.foo.com"},
			expected: Domain{
				Main: "foo.com",
				SANs: []string{"bar.foo.com", "bir.foo.com"},
			},
		},
		{
			desc:       "with 1 domain",
			rawDomains: []string{"foo.com"},
			expected: Domain{
				Main: "foo.com",
				SANs: []string{},
			},
		},
		{
			desc:       "",
			rawDomains: nil,
			expected:   Domain{},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			domain := Domain{}
			domain.Set(test.rawDomains)

			assert.Equal(t, test.expected, domain)
		})
	}
}

func TestMatchDomain(t *testing.T) {
	testCases := []struct {
		desc       string
		certDomain string
		domain     string
		expected   bool
	}{
		{
			desc:       "exact match",
			certDomain: "traefik.wtf",
			domain:     "traefik.wtf",
			expected:   true,
		},
		{
			desc:       "wildcard and root domain",
			certDomain: "*.traefik.wtf",
			domain:     "traefik.wtf",
			expected:   false,
		},
		{
			desc:       "wildcard and sub domain",
			certDomain: "*.traefik.wtf",
			domain:     "sub.traefik.wtf",
			expected:   true,
		},
		{
			desc:       "wildcard and sub sub domain",
			certDomain: "*.traefik.wtf",
			domain:     "sub.sub.traefik.wtf",
			expected:   false,
		},
		{
			desc:       "double wildcard and sub sub domain",
			certDomain: "*.*.traefik.wtf",
			domain:     "sub.sub.traefik.wtf",
			expected:   true,
		},
		{
			desc:       "sub sub domain and invalid wildcard",
			certDomain: "sub.*.traefik.wtf",
			domain:     "sub.sub.traefik.wtf",
			expected:   false,
		},
		{
			desc:       "sub sub domain and valid wildcard",
			certDomain: "*.sub.traefik.wtf",
			domain:     "sub.sub.traefik.wtf",
			expected:   true,
		},
		{
			desc:       "dot replaced by a cahr",
			certDomain: "sub.sub.traefik.wtf",
			domain:     "sub.sub.traefikiwtf",
			expected:   false,
		},
		{
			desc:       "*",
			certDomain: "*",
			domain:     "sub.sub.traefik.wtf",
			expected:   false,
		},
		{
			desc:       "?",
			certDomain: "?",
			domain:     "sub.sub.traefik.wtf",
			expected:   false,
		},
		{
			desc:       "...................",
			certDomain: "...................",
			domain:     "sub.sub.traefik.wtf",
			expected:   false,
		},
		{
			desc:       "wildcard and *",
			certDomain: "*.traefik.wtf",
			domain:     "*.*.traefik.wtf",
			expected:   false,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			domains := MatchDomain(test.domain, test.certDomain)
			assert.Equal(t, test.expected, domains)
		})
	}
}
