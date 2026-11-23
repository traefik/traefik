package acme

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProviderNameALPNCannotCollideWithResolverProvider(t *testing.T) {
	require.False(t, strings.HasSuffix(providerNameALPN, resolverSuffix))
}
