package compress

import (
	"compress/gzip"
	"context"
	"mime"
	"net/http"
	"strings"

	"github.com/NYTimes/gziphandler"
	"github.com/containous/traefik/v2/pkg/config/dynamic"
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

type compressResponseWriter struct {
	excluded          func() []string
	exclusionComputed bool
	compressWriter    http.ResponseWriter
	originalWriter    http.ResponseWriter
}

func (w *compressResponseWriter) Flush() {
	if w.exclusionComputed && w.originalWriter != nil {
		if fw, ok := w.originalWriter.(http.Flusher); ok {
			fw.Flush()
		}
		return
	}

	if fw, ok := w.compressWriter.(http.Flusher); ok {
		fw.Flush()
	}
}

func (w *compressResponseWriter) WriteHeader(code int) {
	if !w.exclusionComputed {
		w.exclusionComputed = true
		contentType := w.Header().Get("Content-Type")

		if !contains(w.excluded(), contentType) {
			w.originalWriter = nil
		} else {
			// Copy headers to original response writer fallback
			for key, values := range w.compressWriter.Header() {
				for _, value := range values {
					w.originalWriter.Header().Add(key, value)
				}
			}
		}
	}

	if w.originalWriter != nil {
		w.originalWriter.WriteHeader(code)
		return
	}

	w.compressWriter.WriteHeader(code)
}

func (w *compressResponseWriter) Header() http.Header {
	// Fallback for trailer headers
	if w.exclusionComputed && w.originalWriter != nil {
		return w.originalWriter.Header()
	}

	return w.compressWriter.Header()
}

func (w *compressResponseWriter) Write(b []byte) (int, error) {
	// Reproduce standard write behavior
	// If has not yet been called, Write calls
	// WriteHeader(http.StatusOK) before writing the data
	if !w.exclusionComputed {
		w.WriteHeader(http.StatusOK)
	}

	if w.originalWriter != nil {
		return w.originalWriter.Write(b)
	}

	return w.compressWriter.Write(b)
}

func (c *compress) handleExclusions(h http.Handler, ow http.ResponseWriter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cw := &compressResponseWriter{
			excluded: func() []string {
				return c.excludes
			},
			originalWriter: ow,
			compressWriter: w,
		}
		h.ServeHTTP(cw, r)
	})
}

func (c *compress) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	mediaType, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil {
		log.FromContext(middlewares.GetLoggerCtx(context.Background(), c.name, typeName)).Debug(err)
	}

	if contains(c.excludes, mediaType) {
		c.next.ServeHTTP(rw, req)
	} else {
		ctx := middlewares.GetLoggerCtx(req.Context(), c.name, typeName)
		gzipHandler(ctx, c.handleExclusions(c.next, rw)).ServeHTTP(rw, req)
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

func contains(values []string, val string) bool {
	for _, v := range values {
		if strings.Contains(val, v) {
			return true
		}
	}
	return false
}
