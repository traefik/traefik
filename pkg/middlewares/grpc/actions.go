package grpc

import (
	proto "github.com/containous/traefik/v2/api/middleware/grpc"
	"net/http"
)

func (g *grpc) actionSetRequestHeaders(headers map[string]string, req *http.Request) {
	if headers == nil {
		return
	}

	for name, value := range headers {
		req.Header.Add(name, value)
	}
}

func (g *grpc) actionRemoveRequestHeaders(headers []string, req *http.Request) {
	for i := 0; i < len(headers); i++ {
		req.Header.Del(headers[i])
	}
}

func (g *grpc) actionSetResponseHeaders(headers map[string]string, rw http.ResponseWriter) {
	if headers == nil {
		return
	}

	for name, value := range headers {
		rw.Header().Add(name, value)
	}
}

func (g *grpc) actionCancelRequest(cancelRequest *proto.CancelRequest, rw http.ResponseWriter) bool {
	if cancelRequest == nil {
		return false
	}

	rw.WriteHeader(int(cancelRequest.Code))

	i, err := rw.Write([]byte(cancelRequest.Body))
	if err != nil || i != len(cancelRequest.Body) {
		// todo: check error and log it? Pass logger to function?
	}

	return true
}
