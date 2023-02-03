package compress

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	"strings"

	"github.com/klauspost/compress/gzhttp"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/middlewares/compress/brotli"
	"github.com/traefik/traefik/v3/pkg/tracing"
)

const typeName = "Compress"

// DefaultMinSize is the default minimum size (in bytes) required to enable compression.
// See https://github.com/klauspost/compress/blob/9559b037e79ad673c71f6ef7c732c00949014cd2/gzhttp/compress.go#L47.
const DefaultMinSize = 1024

// Compress is a middleware that allows to compress the response.
type compress struct {
	next     http.Handler
	name     string
	excludes []string
	minSize  int

	brotliHandler http.Handler
	gzipHandler   http.Handler
}

// New creates a new compress middleware.
func New(ctx context.Context, next http.Handler, conf dynamic.Compress, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug().Msg("Creating middleware")

	excludes := []string{"application/grpc"}
	for _, v := range conf.ExcludedContentTypes {
		mediaType, _, err := mime.ParseMediaType(v)
		if err != nil {
			return nil, err
		}

		excludes = append(excludes, mediaType)
	}

	minSize := DefaultMinSize
	if conf.MinResponseBodyBytes > 0 {
		minSize = conf.MinResponseBodyBytes
	}

	c := &compress{
		next:     next,
		name:     name,
		excludes: excludes,
		minSize:  minSize,
	}

	var err error
	c.brotliHandler, err = c.newBrotliHandler()
	if err != nil {
		return nil, err
	}

	c.gzipHandler, err = c.newGzipHandler()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *compress) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := middlewares.GetLogger(req.Context(), c.name, typeName)

	if req.Method == http.MethodHead {
		c.next.ServeHTTP(rw, req)
		return
	}

	mediaType, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil {
		logger.Debug().Err(err).Msg("Unable to parse MIME type")
	}

	// Notably for text/event-stream requests the response should not be compressed.
	// See https://github.com/traefik/traefik/issues/2576
	if contains(c.excludes, mediaType) {
		c.next.ServeHTTP(rw, req)
		return
	}

	// Client allows us to do whatever we want, so we br compress.
	// See https://www.rfc-editor.org/rfc/rfc9110.html#section-12.5.3
	acceptEncoding, ok := req.Header["Accept-Encoding"]
	if !ok {
		c.brotliHandler.ServeHTTP(rw, req)
		return
	}

	if encodingAccepts(acceptEncoding, "br") {
		c.brotliHandler.ServeHTTP(rw, req)
		return
	}

	if encodingAccepts(acceptEncoding, "gzip") {
		c.gzipHandler.ServeHTTP(rw, req)
		return
	}

	c.next.ServeHTTP(rw, req)
}

func (c *compress) GetTracingInformation() (string, ext.SpanKindEnum) {
	return c.name, tracing.SpanKindNoneEnum
}

func (c *compress) newGzipHandler() (http.Handler, error) {
	wrapper, err := gzhttp.NewWrapper(
		gzhttp.ExceptContentTypes(c.excludes),
		gzhttp.MinSize(c.minSize),
	)
	if err != nil {
		return nil, fmt.Errorf("new gzip wrapper: %w", err)
	}

	return wrapper(c.next), nil
}

func (c *compress) newBrotliHandler() (http.Handler, error) {
	cfg := brotli.Config{
		ExcludedContentTypes: c.excludes,
		MinSize:              c.minSize,
	}

	wrapper, err := brotli.NewWrapper(cfg)
	if err != nil {
		return nil, fmt.Errorf("new brotli wrapper: %w", err)
	}

	return wrapper(c.next), nil
}

func encodingAccepts(acceptEncoding []string, typ string) bool {
	for _, ae := range acceptEncoding {
		for _, e := range strings.Split(ae, ",") {
			parsed := strings.Split(strings.TrimSpace(e), ";")
			if len(parsed) == 0 {
				continue
			}
			if parsed[0] == typ || parsed[0] == "*" {
				return true
			}
		}
	}

	return false
}

func contains(values []string, val string) bool {
	for _, v := range values {
		if v == val {
			return true
		}
	}

	return false
}
