package marathon

import (
	"math"
	"testing"

	"github.com/containous/traefik/provider"
	"github.com/gambol99/go-marathon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfiguration(t *testing.T) {
	testCases := []struct {
		desc     string
		app      marathon.Application
		p        Provider
		expected configuration
	}{
		{
			desc: "Empty labels",
			app: marathon.Application{
				Constraints: &[][]string{},
				Labels:      &map[string]string{},
			},
			p: Provider{
				BaseProvider:              provider.BaseProvider{},
				ExposedByDefault:          false,
				FilterMarathonConstraints: false,
			},
			expected: configuration{
				Enable: false,
				Tags:   nil,
				Marathon: specificConfiguration{
					IPAddressIdx: math.MinInt32,
				},
			},
		},
		{
			desc: "label enable",
			app: marathon.Application{
				Constraints: &[][]string{},
				Labels: &map[string]string{
					"traefik.enable": "true",
				},
			},
			p: Provider{
				BaseProvider:              provider.BaseProvider{},
				ExposedByDefault:          false,
				FilterMarathonConstraints: false,
			},
			expected: configuration{
				Enable: true,
				Tags:   nil,
				Marathon: specificConfiguration{
					IPAddressIdx: math.MinInt32,
				},
			},
		},
		{
			desc: "Use ip address index",
			app: marathon.Application{
				Constraints: &[][]string{},
				Labels: &map[string]string{
					"traefik.marathon.IPAddressIdx": "4",
				},
			},
			p: Provider{
				BaseProvider:              provider.BaseProvider{},
				ExposedByDefault:          false,
				FilterMarathonConstraints: false,
			},
			expected: configuration{
				Enable: false,
				Tags:   nil,
				Marathon: specificConfiguration{
					IPAddressIdx: 4,
				},
			},
		},
		{
			desc: "Use marathon constraints",
			app: marathon.Application{
				Constraints: &[][]string{
					{"key", "value"},
				},
				Labels: &map[string]string{},
			},
			p: Provider{
				BaseProvider:              provider.BaseProvider{},
				ExposedByDefault:          false,
				FilterMarathonConstraints: true,
			},
			expected: configuration{
				Enable: false,
				Tags: []string{
					"key:value",
				},
				Marathon: specificConfiguration{
					IPAddressIdx: math.MinInt32,
				},
			},
		},
		{
			desc: "ExposedByDefault and no enable label",
			app: marathon.Application{
				Constraints: &[][]string{},
				Labels:      &map[string]string{},
			},
			p: Provider{
				BaseProvider:              provider.BaseProvider{},
				ExposedByDefault:          true,
				FilterMarathonConstraints: false,
			},
			expected: configuration{
				Enable: true,
				Tags:   nil,
				Marathon: specificConfiguration{
					IPAddressIdx: math.MinInt32,
				},
			},
		},
		{
			desc: "ExposedByDefault and enable label false",
			app: marathon.Application{
				Constraints: &[][]string{},
				Labels: &map[string]string{
					"traefik.enable": "false",
				},
			},
			p: Provider{
				BaseProvider:              provider.BaseProvider{},
				ExposedByDefault:          true,
				FilterMarathonConstraints: false,
			},
			expected: configuration{
				Enable: false,
				Tags:   nil,
				Marathon: specificConfiguration{
					IPAddressIdx: math.MinInt32,
				},
			},
		},
		{
			desc: "Tags in label",
			app: marathon.Application{
				Constraints: &[][]string{},
				Labels: &map[string]string{
					"traefik.tags": "mytags",
				},
			},
			p: Provider{
				BaseProvider:              provider.BaseProvider{},
				ExposedByDefault:          true,
				FilterMarathonConstraints: false,
			},
			expected: configuration{
				Enable: true,
				Tags:   []string{"mytags"},
				Marathon: specificConfiguration{
					IPAddressIdx: math.MinInt32,
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			extraConf, err := test.p.getConfiguration(test.app)
			require.NoError(t, err)

			assert.Equal(t, test.expected, extraConf)
		})
	}
}
