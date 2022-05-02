package collector

import (
	"testing"

	"github.com/instana/testify/require"
	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v2/pkg/config/static"
)

func Test_createBody(t *testing.T) {
	var staticConfiguration static.Configuration

	err := Fill(&staticConfiguration)
	require.NoError(t, err)

	buffer, err := createBody(&staticConfiguration)
	require.NoError(t, err)

	assert.NotEmpty(t, buffer)
}
