package maxconnection

import (
	"context"
	"fmt"
	"net/http"

	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/middlewares"
	"github.com/containous/traefik/pkg/tracing"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/vulcand/oxy/connlimit"
	"github.com/vulcand/oxy/utils"
)

const (
	typeName = "MaxConnection"
)

type maxConnection struct {
	handler http.Handler
	name    string
}

// New creates a max connection middleware.
func New(ctx context.Context, next http.Handler, maxConns config.MaxConn, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug("Creating middleware")

	extractFunc, err := utils.NewExtractor(maxConns.ExtractorFunc)
	if err != nil {
		return nil, fmt.Errorf("error creating connection limit: %v", err)
	}

	handler, err := connlimit.New(next, extractFunc, maxConns.Amount)
	if err != nil {
		return nil, fmt.Errorf("error creating connection limit: %v", err)
	}

	return &maxConnection{handler: handler, name: name}, nil
}

func (mc *maxConnection) GetTracingInformation() (string, ext.SpanKindEnum) {
	return mc.name, tracing.SpanKindNoneEnum
}

func (mc *maxConnection) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	mc.handler.ServeHTTP(rw, req)
}
