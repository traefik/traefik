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
