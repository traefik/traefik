package ingressnginx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_buildLimitAllowlist(t *testing.T) {
	tests := []struct {
		desc     string
		config   IngressConfig
		expected []string
	}{
		{
			desc:   "no annotation",
			config: IngressConfig{},
		},
		{
			desc:     "canonical annotation",
			config:   IngressConfig{LimitAllowlist: new([]string{"10.0.0.0/24", "192.168.1.1"})},
			expected: []string{"10.0.0.0/24", "192.168.1.1"},
		},
		{
			desc:     "legacy alias",
			config:   IngressConfig{LimitWhitelist: new([]string{"10.0.0.0/24"})},
			expected: []string{"10.0.0.0/24"},
		},
		{
			desc: "canonical annotation takes precedence over the alias",
			config: IngressConfig{
				LimitAllowlist: new([]string{"10.0.0.0/24"}),
				LimitWhitelist: new([]string{"192.168.1.1"}),
			},
			expected: []string{"10.0.0.0/24"},
		},
		{
			desc:     "empty values are dropped",
			config:   IngressConfig{LimitAllowlist: new([]string{"10.0.0.0/24", "", "192.168.1.1"})},
			expected: []string{"10.0.0.0/24", "192.168.1.1"},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var p Provider
			loc := &location{Config: test.config}
			p.buildLimitAllowlist(loc)

			assert.Equal(t, test.expected, loc.LimitAllowlist)
		})
	}
}
