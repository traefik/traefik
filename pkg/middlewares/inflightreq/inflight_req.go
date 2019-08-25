package inflightreq

import (
	"context"
	"fmt"
	"net/http"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/middlewares"
	"github.com/containous/traefik/v2/pkg/tracing"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/vulcand/oxy/connlimit"
)

const (
	typeName = "InFlightReq"
)

type inFlightReq struct {
	handler http.Handler
	name    string
}

// New creates a max request middleware.
func New(ctx context.Context, next http.Handler, config dynamic.InFlightReq, name string) (http.Handler, error) {
	ctxLog := log.With(ctx, log.Str(log.MiddlewareName, name), log.Str(log.MiddlewareType, typeName))
	log.FromContext(ctxLog).Debug("Creating middleware")

	if config.SourceCriterion == nil ||
		config.SourceCriterion.IPStrategy == nil &&
			config.SourceCriterion.RequestHeaderName == "" && !config.SourceCriterion.RequestHost {
		config.SourceCriterion = &dynamic.SourceCriterion{
			RequestHost: true,
		}
	}

	sourceMatcher, err := middlewares.GetSourceExtractor(ctxLog, config.SourceCriterion)
	if err != nil {
		return nil, fmt.Errorf("error creating requests limiter: %v", err)
	}

	handler, err := connlimit.New(next, sourceMatcher, config.Amount)
	if err != nil {
		return nil, fmt.Errorf("error creating connection limit: %v", err)
	}

	return &inFlightReq{handler: handler, name: name}, nil
}

func (mc *inFlightReq) GetTracingInformation() (string, ext.SpanKindEnum) {
	return mc.name, tracing.SpanKindNoneEnum
}

func (mc *inFlightReq) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	mc.handler.ServeHTTP(rw, req)
}
