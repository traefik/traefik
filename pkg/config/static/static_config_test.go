package static

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v2/pkg/provider/hub"
)

func TestHasEntrypoint(t *testing.T) {
	tests := []struct {
		desc   string
		cfg    *Configuration
		assert assert.BoolAssertionFunc
	}{
		{
			desc:   "no user defined entryPoints",
			cfg:    &Configuration{},
			assert: assert.False,
		},
		{
			desc: "user defined entryPoints",
			cfg: &Configuration{EntryPoints: map[string]*EntryPoint{
				"foo": {},
			}},
			assert: assert.True,
		},
		{
			desc: "user defined entryPoints with hub",
			cfg: &Configuration{
				Hub: &hub.Provider{},
				EntryPoints: map[string]*EntryPoint{
					"foo": {},
				},
			},
			assert: assert.True,
		},
		{
			desc: "user defined entryPoints + hub entryPoint (tunnel)",
			cfg: &Configuration{
				Hub: &hub.Provider{},
				EntryPoints: map[string]*EntryPoint{
					"foo":                {},
					hub.TunnelEntrypoint: {},
				},
			},
			assert: assert.True,
		},
		{
			desc: "hub entryPoint (tunnel)",
			cfg: &Configuration{
				Hub: &hub.Provider{},
				EntryPoints: map[string]*EntryPoint{
					hub.TunnelEntrypoint: {},
				},
			},
			assert: assert.False,
		},
		{
			desc: "user defined entryPoints + hub entryPoints (tunnel, api)",
			cfg: &Configuration{
				Hub: &hub.Provider{},
				EntryPoints: map[string]*EntryPoint{
					"foo":                {},
					hub.TunnelEntrypoint: {},
					hub.APIEntrypoint:    {},
				},
			},
			assert: assert.True,
		},
		{
			desc: "hub entryPoints (tunnel, api)",
			cfg: &Configuration{
				Hub: &hub.Provider{},
				EntryPoints: map[string]*EntryPoint{
					hub.TunnelEntrypoint: {},
					hub.APIEntrypoint:    {},
				},
			},
			assert: assert.False,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			test.assert(t, test.cfg.hasUserDefinedEntrypoint())
		})
	}
}
