package main

import (
	"github.com/containous/traefik/plugin"
	"github.com/containous/traefik/plugin/proto"
	//gProto "github.com/golang/protobuf/proto"
	"github.com/hashicorp/go-hclog"
	hPlugin "github.com/hashicorp/go-plugin"
	"os"
	"strings"
)

// MiddlewareTest is a test struct
type MiddlewareTest struct {
	logger hclog.Logger
}

func (m *MiddlewareTest) ServeHTTP(req *proto.Request) (*proto.Response, error) {
	//gProto.Merge(req.Request, &proto.HttpRequest{
	//	Url:        "https://www.google.com",
	//	RequestUri: "https://www.google.com",
	//})
	redirect := strings.Contains(req.Request.Url, "redirect=true")
	renderContent := strings.Contains(req.Request.Url, "renderContent=true")
	stopChain := strings.Contains(req.Request.Url, "stopChain=true")
	m.logger.Debug("Processing plugin request: from within plugin...")
	return &proto.Response{
		Response: &proto.HttpResponse{
			StatusCode: 200,
			Body:       req.Request.Body,
			Header:     map[string]*proto.ValueList{"X-Remote-Plugin-Header": {[]string{"Plugin was called"}}},
		},
		Request:       req.Request,
		Redirect:      redirect,
		RenderContent: renderContent,
		StopChain:     stopChain,
	}, nil
}

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Trace,
		Output:     os.Stderr,
		JSONFormat: true,
	})

	hPlugin.Serve(&hPlugin.ServeConfig{
		HandshakeConfig: plugin.RemoteHandshake,
		Plugins: map[string]hPlugin.Plugin{
			"middleware": &plugin.RemotePlugin{Impl: &MiddlewareTest{logger: logger}},
		},

		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: hPlugin.DefaultGRPCServer,
	})
}
