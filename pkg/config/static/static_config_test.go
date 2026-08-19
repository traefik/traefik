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
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			cfg := &Configuration{
				EntryPoints: test.entryPoints,
			}

			test.assert(t, cfg.hasUserDefinedEntrypoint())
		})
	}
}

func TestValidateConfiguration_aliasHeadersStrategy(t *testing.T) {
	testCases := []struct {
		desc        string
		underscore  string
		alias       string
		expectError bool
	}{
		{
			desc: "no strategy configured",
		},
		{
			desc:  "only the new option configured",
			alias: AliasHeadersStrategyDelete,
		},
		{
			desc:       "only the deprecated option configured",
			underscore: AliasHeadersStrategyDelete,
			alias:      AliasHeadersStrategyDelete,
		},
		{
			desc:        "both options configured with different values",
			underscore:  AliasHeadersStrategyDelete,
			alias:       AliasHeadersStrategyReject,
			expectError: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			cfg := &Configuration{
				Providers: &Providers{},
				EntryPoints: EntryPoints{
					"web": &EntryPoint{
						Address: ":80",
						HTTP: HTTPConfig{
							AliasHeadersStrategy:      test.alias,
							UnderscoreHeadersStrategy: test.underscore,
						},
					},
				},
			}

			err := cfg.ValidateConfiguration()
			if test.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}
