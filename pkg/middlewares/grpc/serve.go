package grpc

import (
	"context"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/middlewares"
	"net/http"
)

func (g *grpc) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := log.FromContext(middlewares.GetLoggerCtx(req.Context(), g.name, typeName))

	ctx, _ := context.WithTimeout(req.Context(), g.timeout)
	request := g.buildRequest(req)

	resp, err := g.grpcClient.Handle(ctx, request, g.grpcCallOptions...)
	if err != nil {
		logger.Warnf("error send request to grpc middleware: %v", err)
		g.next.ServeHTTP(rw, req)
		return
	}

	if resp.Actions == nil {
		g.next.ServeHTTP(rw, req)
		return
	}

	g.actionSetRequestHeaders(resp.Actions.SetRequestHeaders, req)
	g.actionRemoveRequestHeaders(resp.Actions.RemoveRequestHeaders, req)
	g.actionSetResponseHeaders(resp.Actions.SetResponseHeaders, rw)

	if cancel := g.actionCancelRequest(resp.Actions.CancelRequest, rw); cancel {
		logger.Debugf("request cancelled by middleware")
		return
	}

	g.next.ServeHTTP(rw, req)
}
