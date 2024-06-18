package compress

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/middlewares/observability"
)

const (
	vary            = "Vary"
	acceptEncoding  = "Accept-Encoding"
	contentEncoding = "Content-Encoding"
	contentLength   = "Content-Length"
	contentType     = "Content-Type"
)

// Config is the Brotli handler configuration.
type Config struct {
	// ExcludedContentTypes is the list of content types for which we should not compress.
	// Mutually exclusive with the IncludedContentTypes option.
	ExcludedContentTypes []string
	// IncludedContentTypes is the list of content types for which compression should be exclusively enabled.
	// Mutually exclusive with the ExcludedContentTypes option.
	IncludedContentTypes []string
	// MinSize is the minimum size (in bytes) required to enable compression.
	MinSize int
	// Algorithm used for the compression (currently Brotli and Zstandard)
	Algorithm string
	// MiddlewareName use for logging purposes
	MiddlewareName string
}

// CompressionHandler handles Brolti and Zstd compression.
type CompressionHandler struct {
	cfg                  Config
	excludedContentTypes []parsedContentType
	includedContentTypes []parsedContentType
	next                 http.Handler
}

// NewCompressionHandler returns a new compressing handler.
func NewCompressionHandler(cfg Config, next http.Handler) (http.Handler, error) {
	if cfg.Algorithm == "" {
		return nil, errors.New("compression algorithm undefined")
	}

	if cfg.MinSize < 0 {
		return nil, errors.New("minimum size must be greater than or equal to zero")
	}

	if len(cfg.ExcludedContentTypes) > 0 && len(cfg.IncludedContentTypes) > 0 {
		return nil, errors.New("excludedContentTypes and includedContentTypes options are mutually exclusive")
	}

	var excludedContentTypes []parsedContentType
	for _, v := range cfg.ExcludedContentTypes {
		mediaType, params, err := mime.ParseMediaType(v)
		if err != nil {
			return nil, fmt.Errorf("parsing excluded media type: %w", err)
		}

		excludedContentTypes = append(excludedContentTypes, parsedContentType{mediaType, params})
	}

	var includedContentTypes []parsedContentType
	for _, v := range cfg.IncludedContentTypes {
		mediaType, params, err := mime.ParseMediaType(v)
		if err != nil {
			return nil, fmt.Errorf("parsing included media type: %w", err)
		}

		includedContentTypes = append(includedContentTypes, parsedContentType{mediaType, params})
	}

	return &CompressionHandler{
		cfg:                  cfg,
		excludedContentTypes: excludedContentTypes,
		includedContentTypes: includedContentTypes,
		next:                 next,
	}, nil
}

func (c *CompressionHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add(vary, acceptEncoding)

	compressionWriter, err := newCompressionWriter(c.cfg.Algorithm, rw)
	if err != nil {
		logger := middlewares.GetLogger(r.Context(), c.cfg.MiddlewareName, typeName)
		logMessage := fmt.Sprintf("create compression handler: %v", err)
		logger.Debug().Msg(logMessage)
		observability.SetStatusErrorf(r.Context(), logMessage)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	responseWriter := &responseWriter{
		rw:                   rw,
		compressionWriter:    compressionWriter,
		minSize:              c.cfg.MinSize,
		statusCode:           http.StatusOK,
		excludedContentTypes: c.excludedContentTypes,
		includedContentTypes: c.includedContentTypes,
	}
	defer responseWriter.close()

	c.next.ServeHTTP(responseWriter, r)
}

type compression interface {
	// Write data to the encoder.
	// Input data will be buffered and as the buffer fills up
	// content will be compressed and written to the output.
	// When done writing, use Close to flush the remaining output
	// and write CRC if requested.
	Write(p []byte) (n int, err error)
	// Flush will send the currently written data to output
	// and block until everything has been written.
	// This should only be used on rare occasions where pushing the currently queued data is critical.
	Flush() error
	// Close closes the underlying writers if/when appropriate.
	// Note that the compressed writer should not be closed if we never used it,
	// as it would otherwise send some extra "end of compression" bytes.
	// Close also makes sure to flush whatever was left to write from the buffer.
	Close() error
}

type compressionWriter struct {
	compression
	alg string
}

func newCompressionWriter(algo string, in io.Writer) (*compressionWriter, error) {
	switch algo {
	case brotliName:
		return &compressionWriter{compression: brotli.NewWriter(in), alg: algo}, nil

	case zstdName:
		writer, err := zstd.NewWriter(in)
		if err != nil {
			return nil, fmt.Errorf("creating zstd writer: %w", err)
		}
		return &compressionWriter{compression: writer, alg: algo}, nil

	default:
		return nil, fmt.Errorf("unknown compression algo: %s", algo)
	}
}

func (c *compressionWriter) ContentEncoding() string {
	return c.alg
}

// TODO: check whether we want to implement content-type sniffing (as gzip does)
// TODO: check whether we should support Accept-Ranges (as gzip does, see https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept-Ranges)
type responseWriter struct {
	rw                http.ResponseWriter
	compressionWriter *compressionWriter

	minSize              int
	excludedContentTypes []parsedContentType
	includedContentTypes []parsedContentType

	buf                 []byte
	hijacked            bool
	compressionStarted  bool
	compressionDisabled bool
	headersSent         bool

	// Mostly needed to avoid calling bw.Flush/bw.Close when no data was
	// written in bw.
	seenData bool

	statusCodeSet bool
	statusCode    int
}

func (r *responseWriter) Header() http.Header {
	return r.rw.Header()
}

func (r *responseWriter) WriteHeader(statusCode int) {
	if r.statusCodeSet {
		return
	}

	r.statusCode = statusCode
	r.statusCodeSet = true
}

func (r *responseWriter) Write(p []byte) (int, error) {
	// i.e. has write ever been called at least once with non nil data.
	if !r.seenData && len(p) > 0 {
		r.seenData = true
	}

	// We do not compress, either for contentEncoding or contentType reasons.
	if r.compressionDisabled {
		return r.rw.Write(p)
	}

	// We have already buffered more than minSize,
	// We are now in compression cruise mode until the end of times.
	if r.compressionStarted {
		// If compressionStarted we assume we have sent headers already
		return r.compressionWriter.Write(p)
	}

	// If we detect a contentEncoding, we know we are never going to compress.
	if r.rw.Header().Get(contentEncoding) != "" {
		r.compressionDisabled = true
		r.rw.WriteHeader(r.statusCode)
		return r.rw.Write(p)
	}

	// Disable compression according to user wishes in excludedContentTypes or includedContentTypes.
	if ct := r.rw.Header().Get(contentType); ct != "" {
		mediaType, params, err := mime.ParseMediaType(ct)
		if err != nil {
			return 0, fmt.Errorf("parsing content-type media type: %w", err)
		}

		if len(r.includedContentTypes) > 0 {
			var found bool
			for _, includedContentType := range r.includedContentTypes {
				if includedContentType.equals(mediaType, params) {
					found = true
					break
				}
			}
			if !found {
				r.compressionDisabled = true
				r.rw.WriteHeader(r.statusCode)
				return r.rw.Write(p)
			}
		}

		for _, excludedContentType := range r.excludedContentTypes {
			if excludedContentType.equals(mediaType, params) {
				r.compressionDisabled = true
				r.rw.WriteHeader(r.statusCode)
				return r.rw.Write(p)
			}
		}
	}

	// We buffer until we know whether to compress (i.e. when we reach minSize received).
	if len(r.buf)+len(p) < r.minSize {
		r.buf = append(r.buf, p...)
		return len(p), nil
	}

	// If we ever make it here, we have received at least minSize, which means we want to compress,
	// and we are going to send headers right away.
	r.compressionStarted = true

	// Since we know we are going to compress we will never be able to know the actual length.
	r.rw.Header().Del(contentLength)

	r.rw.Header().Set(contentEncoding, r.compressionWriter.ContentEncoding())
	r.rw.WriteHeader(r.statusCode)
	r.headersSent = true

	// Start with sending what we have previously buffered, before actually writing
	// the bytes in argument.
	n, err := r.compressionWriter.Write(r.buf)
	if err != nil {
		r.buf = r.buf[n:]
		// Return zero because we haven't taken care of the bytes in argument yet.
		return 0, err
	}

	// If we wrote less than what we wanted, we need to reclaim the leftovers + the bytes in argument,
	// and keep them for a subsequent Write.
	if n < len(r.buf) {
		r.buf = r.buf[n:]
		r.buf = append(r.buf, p...)
		return len(p), nil
	}

	// Otherwise just reset the buffer.
	r.buf = r.buf[:0]

	// Now that we emptied the buffer, we can actually write the given bytes.
	return r.compressionWriter.Write(p)
}

// Flush flushes data to the appropriate underlying writer(s), although it does
// not guarantee that all buffered data will be sent.
// If not enough bytes have been written to determine whether to enable compression,
// no flushing will take place.
func (r *responseWriter) Flush() {
	if !r.seenData {
		// we should not flush if there never was any data, because flushing the bw
		// (just like closing) would send some extra end of compressionStarted stream bytes.
		return
	}

	// It was already established by Write that compression is disabled, we only
	// have to flush the uncompressed writer.
	if r.compressionDisabled {
		if rw, ok := r.rw.(http.Flusher); ok {
			rw.Flush()
		}

		return
	}

	// Here, nothing was ever written either to rw or to bw (since we're still
	// waiting to decide whether to compress), so we do not need to flush anything.
	// Note that we diverge with klauspost's gzip behavior, where they instead
	// force compression and flush whatever was in the buffer in this case.
	if !r.compressionStarted {
		return
	}

	// Conversely, we here know that something was already written to bw (or is
	// going to be written right after anyway), so bw will have to be flushed.
	// Also, since we know that bw writes to rw, but (apparently) never flushes it,
	// we have to do it ourselves.
	defer func() {
		// because we also ignore the error returned by Write anyway
		_ = r.compressionWriter.Flush()

		if rw, ok := r.rw.(http.Flusher); ok {
			rw.Flush()
		}
	}()

	// We empty whatever is left of the buffer that Write never took care of.
	n, err := r.compressionWriter.Write(r.buf)
	if err != nil {
		return
	}

	// And just like in Write we also handle "short writes".
	if n < len(r.buf) {
		r.buf = r.buf[n:]
		return
	}

	r.buf = r.buf[:0]
}

func (r *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := r.rw.(http.Hijacker); ok {
		// We only make use of r.hijacked in close (and not in Write/WriteHeader)
		// because we want to let the stdlib catch the error on writes, as
		// they already do a good job of logging it.
		r.hijacked = true
		return hijacker.Hijack()
	}

	return nil, nil, fmt.Errorf("%T is not a http.Hijacker", r.rw)
}

// close closes the underlying writers if/when appropriate.
// Note that the compressed writer should not be closed if we never used it,
// as it would otherwise send some extra "end of compression" bytes.
// Close also makes sure to flush whatever was left to write from the buffer.
func (r *responseWriter) close() error {
	if r.hijacked {
		return nil
	}

	// We have to take care of statusCode ourselves (in case there was never any
	// call to Write or WriteHeader before us) as it's the only header we buffer.
	if !r.headersSent {
		r.rw.WriteHeader(r.statusCode)
		r.headersSent = true
	}

	// Nothing was ever written anywhere, nothing to flush.
	if !r.seenData {
		return nil
	}

	// If compression was disabled, there never was anything in the buffer to flush,
	// and nothing was ever written to bw.
	if r.compressionDisabled {
		return nil
	}

	if len(r.buf) == 0 {
		// If we got here we know compression has started, so we can safely flush on bw.
		return r.compressionWriter.Close()
	}

	// There is still data in the buffer, because we never reached minSize (to
	// determine whether to compress). We therefore flush it uncompressed.
	if !r.compressionStarted {
		n, err := r.rw.Write(r.buf)
		if err != nil {
			return err
		}
		if n < len(r.buf) {
			return io.ErrShortWrite
		}
		return nil
	}

	// There is still data in the buffer, simply because Write did not take care of it all.
	// We flush it to the compressed writer.
	n, err := r.compressionWriter.Write(r.buf)
	if err != nil {
		r.compressionWriter.Close()
		return err
	}
	if n < len(r.buf) {
		r.compressionWriter.Close()
		return io.ErrShortWrite
	}
	return r.compressionWriter.Close()
}

// parsedContentType is the parsed representation of one of the inputs to ContentTypes.
// From https://github.com/klauspost/compress/blob/master/gzhttp/compress.go#L401.
type parsedContentType struct {
	mediaType string
	params    map[string]string
}

// equals returns whether this content type matches another content type.
func (p parsedContentType) equals(mediaType string, params map[string]string) bool {
	if p.mediaType != mediaType {
		return false
	}

	// if p has no params, don't care about other's params
	if len(p.params) == 0 {
		return true
	}

	// if p has any params, they must be identical to other's.
	if len(p.params) != len(params) {
		return false
	}

	for k, v := range p.params {
		if w, ok := params[k]; !ok || v != w {
			return false
		}
	}

	return true
}
