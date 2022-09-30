package compress

import (
	"compress/gzip"
	"context"
	"mime"
	"net/http"

	"github.com/klauspost/compress/gzhttp"
	"github.com/opentracing/opentracing-go/ext"
	accept "github.com/timewasted/go-accept-headers"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/middlewares"
	"github.com/traefik/traefik/v2/pkg/middlewares/compress/brotli"
	"github.com/traefik/traefik/v2/pkg/tracing"
)

const (
	typeName = "Compress"

	encodingBrotli = "br"
	encodingGzip   = "gzip"
)

var supportedEncodings = []string{encodingBrotli, encodingGzip}

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
	acceptEncoding := req.Header.Get("Accept-Encoding")
	// Set a default encoding value so `Negotiate` doesn't pick a compression method where none was desired.
	if acceptEncoding == "" {
		acceptEncoding = "identity"
	}

	// If no `Accept-Encoding` header was set, or it was set to `*` encoding will be `supportedEncodings[0]`.
	// If a header was set to a value we don't know about, the negotiated encoding will be "".
	encoding, err := accept.Negotiate(acceptEncoding, supportedEncodings...)
	if err != nil {
		log.FromContext(middlewares.GetLoggerCtx(context.Background(), c.name, typeName)).Debug(err)
	}

	if encoding == "" || contains(c.excludes, parseMediaTypeOrLog(req.Header.Get("Content-Type"), c.name)) {
		c.next.ServeHTTP(rw, req)
		return
	}

	ctx := middlewares.GetLoggerCtx(req.Context(), c.name, typeName)
	switch encoding {
	case encodingBrotli:
		c.brotliHandler().ServeHTTP(rw, req)
	default:
		c.gzipHandler(ctx).ServeHTTP(rw, req)
	}
}

func (c *compress) GetTracingInformation() (string, ext.SpanKindEnum) {
	return c.name, tracing.SpanKindNoneEnum
}

func parseMediaTypeOrLog(contentType, name string) string {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		log.FromContext(middlewares.GetLoggerCtx(context.Background(), name, typeName)).Debug(err)
	}
	return mediaType
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
