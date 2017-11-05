package plugin

import (
	"github.com/containous/traefik/plugin/proto"
	"golang.org/x/net/context"
)

// GRPCClient is an implementation of KV that talks over RPC.
type GRPCClient struct{ client proto.MiddlewareClient }

func (m *GRPCClient) ServeHttp(req *proto.Request) (*proto.Response, error) {
	resp, err := m.client.ServeHttp(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Here is the gRPC server that GRPCClient talks to.
type GRPCServer struct {
	// This is the real implementation
	Impl RemotePluginMiddleware
}

func (m *GRPCServer) ServeHttp(
	ctx context.Context,
	req *proto.Request) (*proto.Response, error) {
	return m.Impl.ServeHttp(req)
}
