package ingressnginx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildFromToWwwRedirect(t *testing.T) {
	testCases := []struct {
		desc     string
		hostname string
		enabled  bool
		allHosts map[string]bool
		expected *middlewareFromToWwwRedirect
	}{
		{
			desc:     "annotation not set",
			hostname: "example.com",
		},
		{
			desc:     "default backend has no hostname",
			hostname: "",
			enabled:  true,
		},
		{
			desc:     "non-www host redirects to www",
			hostname: "example.com",
			enabled:  true,
			expected: &middlewareFromToWwwRedirect{
				ExtraRouterRule: `Host("www.example.com")`,
				TargetHostname:  "example.com",
			},
		},
		{
			desc:     "www host redirects to www",
			hostname: "www.example.com",
			enabled:  true,
			expected: &middlewareFromToWwwRedirect{
				ExtraRouterRule: `Host("example.com")`,
				TargetHostname:  "www.example.com",
			},
		},
		{
			desc:     "www counterpart already served",
			hostname: "example.com",
			enabled:  true,
			allHosts: map[string]bool{"www.example.com": true},
		},
		{
			desc:     "non-www counterpart already served",
			hostname: "www.example.com",
			enabled:  true,
			allHosts: map[string]bool{"example.com": true},
		},
		{
			desc:     "wildcard host",
			hostname: "*.example.com",
			enabled:  true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			loc := &location{Config: IngressConfig{}}
			if test.enabled {
				loc.Config.FromToWwwRedirect = new(true)
			}

			p := &Provider{}
			p.buildFromToWwwRedirect(loc, test.hostname, test.allHosts)

			assert.Equal(t, test.expected, loc.FromToWwwRedirect)
		})
	}
}
