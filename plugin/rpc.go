package plugin

import (
	"github.com/containous/traefik/plugin/proto"
	"net/rpc"
)

// RPCClient is an implementation of KV that talks over RPC.
type RPCClient struct{ client *rpc.Client }

func (m *RPCClient) ServeHttp(req *proto.Request) (*proto.Response, error) {
	var resp = &proto.Response{}

	err := m.client.Call("Plugin.ServeHttp", &req, &resp)

	return resp, err
}

// Here is the RPC server that RPCClient talks to, conforming to
// the requirements of net/rpc
type RPCServer struct {
	// This is the real implementation
	Impl RemotePluginMiddleware
}

func (m *RPCServer) ServeHttp(req *proto.Request) (*proto.Response, error) {
	return m.Impl.ServeHttp(req)
}
