package brotli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"

	"github.com/andybalholm/brotli"
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
}

// NewWrapper returns a new Brotli compressing wrapper.
func NewWrapper(cfg Config) (func(http.Handler) http.HandlerFunc, error) {
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

	return func(h http.Handler) http.HandlerFunc {
		return func(rw http.ResponseWriter, r *http.Request) {
			rw.Header().Add(vary, acceptEncoding)

			brw := &responseWriter{
				rw:                   rw,
				bw:                   brotli.NewWriter(rw),
				minSize:              cfg.MinSize,
				statusCode:           http.StatusOK,
				excludedContentTypes: excludedContentTypes,
				includedContentTypes: includedContentTypes,
			}
			defer brw.close()

			h.ServeHTTP(brw, r)
		}
	}, nil
}

// TODO: check whether we want to implement content-type sniffing (as gzip does)
// TODO: check whether we should support Accept-Ranges (as gzip does, see https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept-Ranges)
type responseWriter struct {
	rw http.ResponseWriter
	bw *brotli.Writer

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
		return r.bw.Write(p)
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
				return r.rw.Write(p)
			}
		}

		for _, excludedContentType := range r.excludedContentTypes {
			if excludedContentType.equals(mediaType, params) {
				r.compressionDisabled = true
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

	r.rw.Header().Set(contentEncoding, "br")
	r.rw.WriteHeader(r.statusCode)
	r.headersSent = true

	// Start with sending what we have previously buffered, before actually writing
	// the bytes in argument.
	n, err := r.bw.Write(r.buf)
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
	return r.bw.Write(p)
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
		_ = r.bw.Flush()

		if rw, ok := r.rw.(http.Flusher); ok {
			rw.Flush()
		}
	}()

	// We empty whatever is left of the buffer that Write never took care of.
	n, err := r.bw.Write(r.buf)
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
		return r.bw.Close()
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
	n, err := r.bw.Write(r.buf)
	if err != nil {
		r.bw.Close()
		return err
	}
	if n < len(r.buf) {
		r.bw.Close()
		return io.ErrShortWrite
	}
	return r.bw.Close()
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
