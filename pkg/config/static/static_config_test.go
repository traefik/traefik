package static

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
