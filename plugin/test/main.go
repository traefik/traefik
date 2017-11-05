package main

import (
	"github.com/containous/traefik/plugin"
	"github.com/containous/traefik/plugin/proto"
	gProto "github.com/golang/protobuf/proto"
	hPlugin "github.com/hashicorp/go-plugin"
)

type MiddlewareTest struct{}

func (MiddlewareTest) ServeHttp(req *proto.Request) (*proto.Response, error) {
	gProto.Merge(req.Request, &proto.HttpRequest{
		Url:        "https://www.google.com",
		RequestUri: "https://www.google.com",
	})
	return &proto.Response{
		Response: &proto.HttpResponse{
			StatusCode: 200,
			Body:       req.Request.Body,
		},
		Request:  req.Request,
		Redirect: true,
	}, nil
}

func main() {
	hPlugin.Serve(&hPlugin.ServeConfig{
		HandshakeConfig: plugin.RemoteHandshake,
		Plugins: map[string]hPlugin.Plugin{
			"middleware": &plugin.RemotePlugin{Impl: &MiddlewareTest{}},
		},

		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: hPlugin.DefaultGRPCServer,
	})
}
