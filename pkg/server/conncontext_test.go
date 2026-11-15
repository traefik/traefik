package server

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type keyTest string

func TestConnContext(t *testing.T) {
	var connContext multipleConnContext
	connContext.AddConnContextFunc(func(ctx context.Context, _ net.Conn) context.Context {
		return context.WithValue(ctx, keyTest("test"), "test")
	})
	connContext.AddConnContextFunc(func(ctx context.Context, _ net.Conn) context.Context {
		return context.WithValue(ctx, keyTest("test2"), "test2")
	})

	ctx := connContext.Build()(context.Background(), nil)

	require.Equal(t, "test", ctx.Value(keyTest("test")))
	require.Equal(t, "test2", ctx.Value(keyTest("test2")))
}

func TestConnContexts(t *testing.T) {
	var connContexts connContexts

	conn, otherConn := net.Pipe()
	t.Cleanup(func() {
		_ = conn.Close()
		_ = otherConn.Close()
	})

	_, ok := connContexts.remove(conn)
	assert.False(t, ok)

	connContexts.add(context.WithValue(t.Context(), keyTest("conn"), "conn"), conn)

	_, ok = connContexts.remove(otherConn)
	assert.False(t, ok)

	ctx, ok := connContexts.remove(conn)
	require.True(t, ok)
	assert.Equal(t, "conn", ctx.Value(keyTest("conn")))

	_, ok = connContexts.remove(conn)
	assert.False(t, ok)
}
