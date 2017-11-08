package plugin

import (
	"github.com/containous/traefik/plugin/proto"
	"net/rpc"
)

// RPCClient is an implementation of KV that talks over RPC.
type RPCClient struct{ client *rpc.Client }

// ServeHTTP method implements facade on top of NetRPC Client
func (m *RPCClient) ServeHTTP(req *proto.Request) (*proto.Response, error) {
	var resp = &proto.Response{}

	err := m.client.Call("Plugin.ServeHTTP", &req, &resp)

	return resp, err
}

// RPCServer is the RPC server that RPCClient talks to, conforming to
// the requirements of net/rpc
type RPCServer struct {
	// This is the real implementation
	Impl RemotePluginMiddleware
}

func (m *RPCServer) ServeHTTP(req *proto.Request) (*proto.Response, error) {
	return m.Impl.ServeHTTP(req)
}
