package server

import (
	"context"
	"net"
	"testing"

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
