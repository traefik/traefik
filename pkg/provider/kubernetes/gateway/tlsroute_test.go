package gateway

import (
	"testing"

	"github.com/stretchr/testify/assert"
	gatev1 "sigs.k8s.io/gateway-api/apis/v1"
)

func Test_hostSNIRule(t *testing.T) {
	testCases := []struct {
		desc             string
		hostnames        []gatev1.Hostname
		expectedRule     string
		expectedPriority int
		expectError      bool
	}{
		{
			desc:             "Empty",
			expectedRule:     "HostSNI(`*`)",
			expectedPriority: 0,
		},
		{
			desc:             "Empty hostname",
			hostnames:        []gatev1.Hostname{""},
			expectedRule:     "HostSNI(`*`)",
			expectedPriority: 0,
		},
		{
			desc:             "Supported wildcard",
			hostnames:        []gatev1.Hostname{"*.foo"},
			expectedRule:     "HostSNIRegexp(`^[a-z0-9-\\.]+\\.foo$`)",
			expectedPriority: 4,
		},
		{
			desc:             "Some empty hostnames",
			hostnames:        []gatev1.Hostname{"foo", "", "bar"},
			expectedRule:     "HostSNI(`foo`) || HostSNI(`bar`)",
			expectedPriority: 3,
		},
		{
			desc:             "Valid hostname",
			hostnames:        []gatev1.Hostname{"foo"},
			expectedRule:     "HostSNI(`foo`)",
			expectedPriority: 3,
		},
		{
			desc:             "Multiple valid hostnames",
			hostnames:        []gatev1.Hostname{"foo", "bar"},
			expectedRule:     "HostSNI(`foo`) || HostSNI(`bar`)",
			expectedPriority: 3,
		},
		{
			desc:             "Multiple valid hostnames with wildcard",
			hostnames:        []gatev1.Hostname{"bar.foo", "foo.foo", "*.foo"},
			expectedRule:     "HostSNI(`bar.foo`) || HostSNI(`foo.foo`) || HostSNIRegexp(`^[a-z0-9-\\.]+\\.foo$`)",
			expectedPriority: 7,
		},
		{
			desc:             "Multiple overlapping hostnames",
			hostnames:        []gatev1.Hostname{"foo", "bar", "foo", "baz"},
			expectedRule:     "HostSNI(`foo`) || HostSNI(`bar`) || HostSNI(`baz`)",
			expectedPriority: 3,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rule, priority := hostSNIRule(test.hostnames)
			assert.Equal(t, test.expectedRule, rule)
			assert.Equal(t, test.expectedPriority, priority)
		})
	}
}
