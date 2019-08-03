package compress

import (
	"compress/gzip"
	"context"
	"net/http"
	"strings"

	"github.com/NYTimes/gziphandler"
	"github.com/containous/traefik/v2/pkg/middlewares"
	"github.com/containous/traefik/v2/pkg/tracing"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sirupsen/logrus"
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
	middlewares.GetLogger(ctx, name, typeName).Debug("Creating middleware")

	return &compress{
		next: next,
		name: name,
	}, nil
}

func (c *compress) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	contentType := req.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/grpc") {
		c.next.ServeHTTP(rw, req)
	} else {
		gzipHandler(c.next, middlewares.GetLogger(req.Context(), c.name, typeName)).ServeHTTP(rw, req)
	}
}

func (c *compress) GetTracingInformation() (string, ext.SpanKindEnum) {
	return c.name, tracing.SpanKindNoneEnum
}

func gzipHandler(h http.Handler, logger logrus.FieldLogger) http.Handler {
	wrapper, err := gziphandler.GzipHandlerWithOpts(
		gziphandler.CompressionLevel(gzip.DefaultCompression),
		gziphandler.MinSize(gziphandler.DefaultMinSize))
	if err != nil {
		logger.Error(err)
	}

	return wrapper(h)
}
