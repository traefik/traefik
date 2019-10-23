package grpc

import (
	proto "github.com/containous/traefik/v2/api/middleware/grpc"
	"net/http"
)

func (g *grpc) buildRequest(req *http.Request) *proto.Request {
	request := &proto.Request{
		Headers: map[string]string{},
	}

	if g.passToClient.Method {
		request.Method = req.Method
	}

	if g.passToClient.RequestURI {
		request.RequestURI = req.RequestURI
	}

	if g.passToClient.RemoteAddr {
		request.RemoteAddr = req.RemoteAddr
	}

	for i := 0; i < len(g.passToClient.Headers); i++ {
		request.Headers[g.passToClient.Headers[i]] = req.Header.Get(g.passToClient.Headers[i])
	}

	return request
}
