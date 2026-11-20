package redis

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProvider_Init_EmptyEndpoints(t *testing.T) {
	provider := &Provider{}
	provider.SetDefaults()
	provider.Endpoints = nil

	require.EqualError(t, provider.Init(), "no endpoint defined")
}
