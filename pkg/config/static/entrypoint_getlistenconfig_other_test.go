//go:build !(linux || freebsd || openbsd || darwin)

package static

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEntryPointGetListenConfig(t *testing.T) {
	ep := EntryPoint{Address: ":0"}
	listenConfig := ep.GetListenConfig()
	require.Nil(t, listenConfig.Control)
	require.Zero(t, listenConfig.KeepAlive)

	l1, err := listenConfig.Listen(context.Background(), "tcp", ep.Address)
	require.NoError(t, err)
	require.NotNil(t, l1)
	defer l1.Close()

	l2, err := listenConfig.Listen(context.Background(), "tcp", l1.Addr().String())
	require.Error(t, err)
	require.ErrorContains(t, err, "address already in use")
	require.Nil(t, l2)

	ep = EntryPoint{Address: ":0", ReusePort: true}
	listenConfig = ep.GetListenConfig()
	require.Nil(t, listenConfig.Control)
	require.Zero(t, listenConfig.KeepAlive)

	l3, err := listenConfig.Listen(context.Background(), "tcp", ep.Address)
	require.NoError(t, err)
	require.NotNil(t, l3)
	defer l3.Close()

	l4, err := listenConfig.Listen(context.Background(), "tcp", l3.Addr().String())
	require.Error(t, err)
	require.ErrorContains(t, err, "address already in use")
	require.Nil(t, l4)
}
