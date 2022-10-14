package compress

import (
	"compress/gzip"
	"context"
	"mime"
	"net/http"
	"strings"

	"github.com/klauspost/compress/gzhttp"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/middlewares"
	"github.com/traefik/traefik/v2/pkg/middlewares/compress/brotli"
	"github.com/traefik/traefik/v2/pkg/tracing"
)

const (
	typeName = "Compress"
)

// Compress is a middleware that allows to compress the response.
type compress struct {
	next     http.Handler
	name     string
	excludes []string
	minSize  int
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

	minSize := gzhttp.DefaultMinSize
	if conf.MinResponseBodyBytes > 0 {
		minSize = conf.MinResponseBodyBytes
	}

	return &compress{next: next, name: name, excludes: excludes, minSize: minSize}, nil
}

func (c *compress) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodHead {
		c.next.ServeHTTP(rw, req)
		return
	}

	ctx := middlewares.GetLoggerCtx(req.Context(), c.name, typeName)
	mediaType, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil {
		log.FromContext(ctx).Debug(err)
	}

	if contains(c.excludes, mediaType) {
		c.next.ServeHTTP(rw, req)
		return
	}

	acceptEncoding := strings.TrimSpace(req.Header.Get("Accept-Encoding"))
	if acceptEncoding == "" {
		c.next.ServeHTTP(rw, req)
		return
	}

	if brotli.AcceptsBr(acceptEncoding) {
		c.brotliHandler().ServeHTTP(rw, req)
		return
	}

	c.gzipHandler(ctx).ServeHTTP(rw, req)
}

func (c *compress) GetTracingInformation() (string, ext.SpanKindEnum) {
	return c.name, tracing.SpanKindNoneEnum
}

func (c *compress) gzipHandler(ctx context.Context) http.Handler {
	wrapper, err := gzhttp.NewWrapper(
		gzhttp.ExceptContentTypes(c.excludes),
		gzhttp.CompressionLevel(gzip.DefaultCompression),
		gzhttp.MinSize(c.minSize))
	if err != nil {
		log.FromContext(ctx).Error(err)
	}

	return wrapper(c.next)
}

func (c *compress) brotliHandler() http.Handler {
	return brotli.NewMiddleware(
		brotli.WithMinSize(c.minSize),
	)(c.next)
}

func contains(values []string, val string) bool {
	for _, v := range values {
		if v == val {
			return true
		}
	}
	return false
}
