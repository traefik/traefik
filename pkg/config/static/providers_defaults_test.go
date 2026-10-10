package static

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProvidersSetDefaultsDoesNotAliasProviderNames guards against
// SetEffectiveConfiguration, which rewrites Precedence in place, writing
// through into the package-level providerNames slice shared by every
// Configuration.
func TestProvidersSetDefaultsDoesNotAliasProviderNames(t *testing.T) {
	t.Parallel()

	original := slices.Clone(providerNames)

	var providers Providers
	providers.SetDefaults()

	require.Equal(t, original, providers.Precedence)

	providers.Precedence[0] = "MUTATED"

	assert.Equal(t, original, providerNames)
}

func TestProvidersSetDefaultsIsIndependentPerConfiguration(t *testing.T) {
	t.Parallel()

	var first, second Providers
	first.SetDefaults()
	second.SetDefaults()

	first.Precedence[0] = "MUTATED"

	assert.NotEqual(t, first.Precedence[0], second.Precedence[0])
}

// TestSetEffectiveConfigurationDoesNotMutateSharedPrecedence covers the caller
// side: SetEffectiveConfiguration lowercases Precedence, and must not write
// through into a slice the Configuration does not own.
func TestSetEffectiveConfigurationDoesNotMutateSharedPrecedence(t *testing.T) {
	t.Parallel()

	shared := []string{"DOCKER", "File"}
	original := slices.Clone(shared)

	cfg := &Configuration{Providers: &Providers{Precedence: shared}}
	cfg.SetEffectiveConfiguration()

	assert.Equal(t, original, shared)
	assert.Equal(t, []string{"docker", "file"}, cfg.Providers.Precedence)
}
