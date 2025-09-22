package compress

import (
	"context"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"slices"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/gzhttp"
	"github.com/klauspost/compress/zstd"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
)

const typeName = "Compress"

// defaultMinSize is the default minimum size (in bytes) required to enable compression.
// See https://github.com/klauspost/compress/blob/9559b037e79ad673c71f6ef7c732c00949014cd2/gzhttp/compress.go#L47.
const defaultMinSize = 1024

var defaultSupportedEncodings = []string{gzipName, brotliName, zstdName}

// Compress is a middleware that allows to compress the response.
type compress struct {
	next            http.Handler
	name            string
	excludes        []string
	includes        []string
	minSize         int
	encodings       []string
	defaultEncoding string
	// supportedEncodings is a map of supported encodings and their priority.
	supportedEncodings map[string]int

	brotliHandler http.Handler
	gzipHandler   http.Handler
	zstdHandler   http.Handler
}

// New creates a new compress middleware.
func New(ctx context.Context, next http.Handler, conf dynamic.Compress, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug().Msg("Creating middleware")

	if len(conf.ExcludedContentTypes) > 0 && len(conf.IncludedContentTypes) > 0 {
		return nil, errors.New("excludedContentTypes and includedContentTypes options are mutually exclusive")
	}

	excludes := []string{"application/grpc"}
	for _, v := range conf.ExcludedContentTypes {
		mediaType, _, err := mime.ParseMediaType(v)
		if err != nil {
			return nil, fmt.Errorf("parsing excluded media type: %w", err)
		}

		excludes = append(excludes, mediaType)
	}

	var includes []string
	for _, v := range conf.IncludedContentTypes {
		mediaType, _, err := mime.ParseMediaType(v)
		if err != nil {
			return nil, fmt.Errorf("parsing included media type: %w", err)
		}

		includes = append(includes, mediaType)
	}

	minSize := defaultMinSize
	if conf.MinResponseBodyBytes > 0 {
		minSize = conf.MinResponseBodyBytes
	}

	if len(conf.Encodings) == 0 {
		return nil, errors.New("at least one encoding must be specified")
	}
	for _, encoding := range conf.Encodings {
		if !slices.Contains(defaultSupportedEncodings, encoding) {
			return nil, fmt.Errorf("unsupported encoding: %s", encoding)
		}
	}
	if conf.DefaultEncoding != "" && !slices.Contains(conf.Encodings, conf.DefaultEncoding) {
		return nil, fmt.Errorf("unsupported default encoding: %s", conf.DefaultEncoding)
	}

	c := &compress{
		next:               next,
		name:               name,
		excludes:           excludes,
		includes:           includes,
		minSize:            minSize,
		encodings:          conf.Encodings,
		defaultEncoding:    conf.DefaultEncoding,
		supportedEncodings: buildSupportedEncodings(conf.Encodings),
	}

	var err error

	c.zstdHandler, err = c.newZstdHandler(name)
	if err != nil {
		return nil, err
	}

	c.brotliHandler, err = c.newBrotliHandler(name)
	if err != nil {
		return nil, err
	}

	c.gzipHandler, err = c.newGzipHandler()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func buildSupportedEncodings(encodings []string) map[string]int {
	supportedEncodings := map[string]int{
		// the most permissive first.
		wildcardName: -1,
		// the less permissive last.
		identityName: len(encodings),
	}
	for i, encoding := range encodings {
		supportedEncodings[encoding] = i
	}
	return supportedEncodings
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
	if slices.Contains(c.excludes, mediaType) {
		c.next.ServeHTTP(rw, req)
		return
	}

	acceptEncoding, ok := req.Header[acceptEncodingHeader]
	if !ok {
		if c.defaultEncoding != "" {
			// RFC says: "If no Accept-Encoding header field is in the request, any content coding is considered acceptable by the user agent."
			// https://www.rfc-editor.org/rfc/rfc9110#field.accept-encoding
			c.chooseHandler(c.defaultEncoding, rw, req)
			return
		}

		// Client doesn't specify a preferred encoding, for compatibility don't encode the request
		// See https://github.com/traefik/traefik/issues/9734
		c.next.ServeHTTP(rw, req)
		return
	}

	c.chooseHandler(c.getCompressionEncoding(acceptEncoding), rw, req)
}

func (c *compress) chooseHandler(typ string, rw http.ResponseWriter, req *http.Request) {
	switch typ {
	case zstdName:
		c.zstdHandler.ServeHTTP(rw, req)
	case brotliName:
		c.brotliHandler.ServeHTTP(rw, req)
	case gzipName:
		c.gzipHandler.ServeHTTP(rw, req)
	default:
		c.next.ServeHTTP(rw, req)
	}
}

func (c *compress) GetTracingInformation() (string, string) {
	return c.name, typeName
}

func (c *compress) newGzipHandler() (http.Handler, error) {
	var wrapper func(http.Handler) http.HandlerFunc
	var err error

	if len(c.includes) > 0 {
		wrapper, err = gzhttp.NewWrapper(
			gzhttp.ContentTypes(c.includes),
			gzhttp.MinSize(c.minSize),
		)
	} else {
		wrapper, err = gzhttp.NewWrapper(
			gzhttp.ExceptContentTypes(c.excludes),
			gzhttp.MinSize(c.minSize),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("new gzip wrapper: %w", err)
	}

	return wrapper(c.next), nil
}

func (c *compress) newBrotliHandler(middlewareName string) (http.Handler, error) {
	cfg := Config{MinSize: c.minSize, MiddlewareName: middlewareName}
	if len(c.includes) > 0 {
		cfg.IncludedContentTypes = c.includes
	} else {
		cfg.ExcludedContentTypes = c.excludes
	}

	newBrotliWriter := func(rw http.ResponseWriter) (CompressionWriter, string, error) {
		return brotli.NewWriter(rw), brotliName, nil
	}
	return NewCompressionHandler(cfg, newBrotliWriter, c.next)
}

func (c *compress) newZstdHandler(middlewareName string) (http.Handler, error) {
	cfg := Config{MinSize: c.minSize, MiddlewareName: middlewareName}
	if len(c.includes) > 0 {
		cfg.IncludedContentTypes = c.includes
	} else {
		cfg.ExcludedContentTypes = c.excludes
	}

	newZstdWriter := func(rw http.ResponseWriter) (CompressionWriter, string, error) {
		writer, err := zstd.NewWriter(rw)
		if err != nil {
			return nil, "", fmt.Errorf("creating zstd writer: %w", err)
		}
		return writer, zstdName, nil
	}
	return NewCompressionHandler(cfg, newZstdWriter, c.next)
}
