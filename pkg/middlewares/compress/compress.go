package compress

import (
	"context"
	"mime"
	"net/http"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/middlewares"
	"github.com/containous/traefik/v2/pkg/tracing"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/containous/traefik/v2/pkg/brotlihandler"
)

const (
	typeName = "Compress"
)

// Compress is a middleware that allows to compress the response.
type compress struct {
	next     http.Handler
	name     string
	excludes []string
}

// New creates a new compress middleware.
func New(ctx context.Context, next http.Handler, conf dynamic.Compress, name string) (http.Handler, error) {
	log.FromContext(middlewares.GetLoggerCtx(ctx, name, typeName)).Debug("Creating middleware")

	excludes := []string{"application/grpc"}
	for _, v := range conf.ExcludedContentTypes {
		mediaType, _, err := mime.ParseMediaType(v)
		if err != nil {
			return nil, err
		}

		excludes = append(excludes, mediaType)
	}

	return &compress{next: next, name: name, excludes: excludes}, nil
}

func (c *compress) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	mediaType, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil {
		log.FromContext(middlewares.GetLoggerCtx(context.Background(), c.name, typeName)).Debug(err)
	}

	if contains(c.excludes, mediaType) {
		c.next.ServeHTTP(rw, req)
	} else {
		brotlihandler.CompressHandler(c.next).ServeHTTP(rw, req)
	}
}

func (c *compress) GetTracingInformation() (string, ext.SpanKindEnum) {
	return c.name, tracing.SpanKindNoneEnum
}

func contains(values []string, val string) bool {
	for _, v := range values {
		if v == val {
			return true
		}
	}
	return false
}
