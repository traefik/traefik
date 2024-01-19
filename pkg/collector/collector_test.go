package collector

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/collector/hydratation"
	"github.com/traefik/traefik/v3/pkg/config/static"
)

func Test_createBody(t *testing.T) {
	var staticConfiguration static.Configuration

	err := hydratation.Hydrate(&staticConfiguration)
	require.NoError(t, err)

	buffer, err := createBody(&staticConfiguration)
	require.NoError(t, err)

	assert.NotEmpty(t, buffer)
}
