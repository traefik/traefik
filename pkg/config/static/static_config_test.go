package static

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v3/pkg/provider/hub"
)

func TestHasEntrypoint(t *testing.T) {
	tests := []struct {
		desc        string
		entryPoints map[string]*EntryPoint
		assert      assert.BoolAssertionFunc
	}{
		{
			desc:   "no user defined entryPoints",
			assert: assert.False,
		},
		{
			desc: "user defined entryPoints",
			entryPoints: map[string]*EntryPoint{
				"foo": {},
			},
			assert: assert.True,
		},
		{
			desc: "user defined entryPoints + hub entryPoint (tunnel)",
			entryPoints: map[string]*EntryPoint{
				"foo":                {},
				hub.TunnelEntrypoint: {},
			},
			assert: assert.True,
		},
		{
			desc: "hub entryPoint (tunnel)",
			entryPoints: map[string]*EntryPoint{
				hub.TunnelEntrypoint: {},
			},
			assert: assert.False,
		},
		{
			desc: "user defined entryPoints + hub entryPoint (api)",
			entryPoints: map[string]*EntryPoint{
				"foo":             {},
				hub.APIEntrypoint: {},
			},
			assert: assert.True,
		},
		{
			desc: "hub entryPoint (api)",
			entryPoints: map[string]*EntryPoint{
				hub.APIEntrypoint: {},
			},
			assert: assert.True,
		},
		{
			desc: "user defined entryPoints + hub entryPoints (tunnel, api)",
			entryPoints: map[string]*EntryPoint{
				"foo":                {},
				hub.TunnelEntrypoint: {},
				hub.APIEntrypoint:    {},
			},
			assert: assert.True,
		},
		{
			desc: "hub entryPoints (tunnel, api)",
			entryPoints: map[string]*EntryPoint{
				hub.TunnelEntrypoint: {},
				hub.APIEntrypoint:    {},
			},
			assert: assert.False,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			cfg := &Configuration{
				EntryPoints: test.entryPoints,
			}

			test.assert(t, cfg.hasUserDefinedEntrypoint())
		})
	}
}
