package plugin

import (
	"github.com/containous/traefik/plugin/proto"
	"golang.org/x/net/context"
)

// GRPCClient is an implementation of KV that talks over RPC.
type GRPCClient struct{ client proto.MiddlewareClient }

// ServeHTTP method implements facade on top of gRPC Client
func (m *GRPCClient) ServeHTTP(req *proto.Request) (*proto.Response, error) {
	resp, err := m.client.ServeHTTP(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GRPCServer is the gRPC server that GRPCClient talks to.
type GRPCServer struct {
	// This is the real implementation
	Impl RemotePluginMiddleware
}

// ServeHTTP method implements facade on top of gRPC Server.
func (m *GRPCServer) ServeHTTP(
	ctx context.Context,
	req *proto.Request) (*proto.Response, error) {
	return m.Impl.ServeHTTP(req)
}
