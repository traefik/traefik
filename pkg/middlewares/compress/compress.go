package compress

import (
	"compress/gzip"
	"context"
	"net/http"
	"strings"

	"github.com/NYTimes/gziphandler"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/middlewares"
	"github.com/containous/traefik/v2/pkg/tracing"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	typeName = "Compress"
)

// Compress is a middleware that allows to compress the response.
type compress struct {
	next http.Handler
	name string
}

// New creates a new compress middleware.
func New(ctx context.Context, next http.Handler, name string) (http.Handler, error) {
	log.FromContext(middlewares.GetLoggerCtx(ctx, name, typeName)).Debug("Creating middleware")

	return &compress{
		next: next,
		name: name,
	}, nil
}

func (c *compress) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	contentType := req.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/grpc") || strings.HasPrefix(contentType, "text/event-stream") {
		c.next.ServeHTTP(rw, req)
	} else {
		ctx := middlewares.GetLoggerCtx(req.Context(), c.name, typeName)
		gzipHandler(ctx, c.next).ServeHTTP(rw, req)
	}
}

func (c *compress) GetTracingInformation() (string, ext.SpanKindEnum) {
	return c.name, tracing.SpanKindNoneEnum
}

func gzipHandler(ctx context.Context, h http.Handler) http.Handler {
	wrapper, err := gziphandler.GzipHandlerWithOpts(
		gziphandler.CompressionLevel(gzip.DefaultCompression),
		gziphandler.MinSize(gziphandler.DefaultMinSize))
	if err != nil {
		log.FromContext(ctx).Error(err)
	}

	return wrapper(h)
}
