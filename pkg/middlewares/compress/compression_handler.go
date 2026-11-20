package compress

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"slices"
	"strconv"
	"sync"

	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/middlewares/observability"
)

const (
	vary            = "Vary"
	acceptEncoding  = "Accept-Encoding"
	acceptRanges    = "Accept-Ranges"
	contentEncoding = "Content-Encoding"
	contentLength   = "Content-Length"
	contentRange    = "Content-Range"
	contentType     = "Content-Type"
)

// sniffLen is the number of bytes http.DetectContentType looks at.
const sniffLen = 512

// CompressionWriter compresses the written bytes.
type CompressionWriter interface {
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
	// Reset reinitializes the state of the encoder, allowing it to be reused.
	Reset(w io.Writer)
}

// NewCompressionWriter returns a new CompressionWriter with its corresponding algorithm.
type NewCompressionWriter func(rw http.ResponseWriter) (CompressionWriter, string, error)

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
	// MiddlewareName use for logging purposes
	MiddlewareName string
}

// CompressionHandler handles Gzip, Brotli and Zstd compression.
type CompressionHandler struct {
	cfg                  Config
	excludedContentTypes []parsedContentType
	includedContentTypes []parsedContentType
	next                 http.Handler

	writerPool sync.Pool
	newWriter  NewCompressionWriter
}

// NewCompressionHandler returns a new compressing handler.
func NewCompressionHandler(cfg Config, newWriter NewCompressionWriter, next http.Handler) (http.Handler, error) {
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
		newWriter:            newWriter,
	}, nil
}

func (c *CompressionHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add(vary, acceptEncoding)

	compressionWriter, err := c.getCompressionWriter(rw)
	if err != nil {
		logger := middlewares.GetLogger(r.Context(), c.cfg.MiddlewareName, typeName)
		logger.Debug().Msgf("Create compression handler: %v", err)
		observability.SetStatusErrorf(r.Context(), "Create compression handler: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer c.putCompressionWriter(compressionWriter)

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

func (c *CompressionHandler) getCompressionWriter(rw http.ResponseWriter) (*compressionWriterWrapper, error) {
	if writer, ok := c.writerPool.Get().(*compressionWriterWrapper); ok {
		writer.Reset(rw)
		return writer, nil
	}

	writer, algo, err := c.newWriter(rw)
	if err != nil {
		return nil, fmt.Errorf("creating compression writer: %w", err)
	}
	return &compressionWriterWrapper{CompressionWriter: writer, algo: algo}, nil
}

func (c *CompressionHandler) putCompressionWriter(writer *compressionWriterWrapper) {
	writer.Reset(nil)
	c.writerPool.Put(writer)
}

type compressionWriterWrapper struct {
	CompressionWriter

	algo string
}

func (c *compressionWriterWrapper) ContentEncoding() string {
	return c.algo
}

// writerState tracks the compress-or-not decision of a response.
// The headers are held back until the decision is made, as compressing changes them.
type writerState int

const (
	// stateHeadersPending means nothing was written or flushed yet.
	stateHeadersPending writerState = iota
	// stateBuffering means the headers allow compression, and the body is buffered until it reaches minSize.
	stateBuffering
	// statePlain means the headers are sent and the body goes out uncompressed.
	statePlain
	// stateCompressing means the headers are sent and the body goes through the compression writer.
	stateCompressing
)

type responseWriter struct {
	rw                http.ResponseWriter
	compressionWriter *compressionWriterWrapper

	minSize              int
	excludedContentTypes []parsedContentType
	includedContentTypes []parsedContentType

	state    writerState
	buf      []byte
	hijacked bool

	// declaredSize is the body size announced by the Content-Length header, or -1 when unknown.
	declaredSize int

	statusCodeSet bool
	statusCode    int
}

func (r *responseWriter) Header() http.Header {
	return r.rw.Header()
}

func (r *responseWriter) WriteHeader(statusCode int) {
	// Handle informational headers
	// This is gated to not forward 1xx responses on builds prior to go1.20.
	if statusCode >= 100 && statusCode <= 199 {
		r.rw.WriteHeader(statusCode)
		return
	}

	if !r.statusCodeSet {
		r.statusCode = statusCode
		r.statusCodeSet = true
	}
}

func (r *responseWriter) Write(p []byte) (int, error) {
	if r.state == stateHeadersPending {
		r.checkHeaders()
	}

	// The headers are already handled, and some writers reject even an empty body for statuses such as 204.
	if len(p) == 0 {
		return 0, nil
	}

	switch r.state {
	case stateCompressing:
		return r.compressionWriter.Write(p)
	case statePlain:
		return r.rw.Write(p)
	}

	// Only what is needed to reach minSize and to sniff the content type is buffered,
	// so the rest of a large write is not copied.
	n := min(len(p), max(r.minSize, sniffLen)-len(r.buf))
	r.buf = append(r.buf, p[:n]...)

	compress := r.declaredSize >= r.minSize
	if r.declaredSize < 0 {
		if len(r.buf) < r.minSize {
			// The whole write fits in the buffer, as it can hold at least minSize bytes.
			return len(p), nil
		}

		compress = true
	}

	if err := r.start(compress); err != nil {
		return 0, err
	}

	if n == len(p) {
		return n, nil
	}

	m, err := r.Write(p[n:])
	return n + m, err
}

// FlushError sends the headers and everything written so far to the client.
func (r *responseWriter) FlushError() error {
	if r.state == stateHeadersPending {
		r.checkHeaders()
	}

	if r.state == stateBuffering {
		// The handler wants what it wrote so far, headers included, to reach the client now.
		// Streams such as Kubernetes watches flush their headers and may then send nothing for minutes,
		// so the decision cannot wait for minSize bytes: without a Content-Length,
		// the body is a stream of unknown size, which is worth compressing.
		if err := r.start(r.declaredSize < 0 || r.declaredSize >= r.minSize); err != nil {
			return err
		}
	}

	if r.state == stateCompressing {
		if err := r.compressionWriter.Flush(); err != nil {
			return err
		}
	}

	return http.NewResponseController(r.rw).Flush()
}

func (r *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	conn, brw, err := http.NewResponseController(r.rw).Hijack()
	if errors.Is(err, http.ErrNotSupported) {
		return nil, nil, err
	}

	// We only make use of r.hijacked in close (and not in Write/WriteHeader)
	// because we want to let the stdlib catch the error on writes, as
	// they already do a good job of logging it.
	r.hijacked = true

	return conn, brw, err
}

// Unwrap exposes the underlying ResponseWriter for http.ResponseController.
func (r *responseWriter) Unwrap() http.ResponseWriter {
	return r.rw
}

// checkHeaders runs on the first write or flush, as the handler cannot change the headers after that.
func (r *responseWriter) checkHeaders() {
	r.declaredSize = -1
	if size, err := strconv.Atoi(r.rw.Header().Get(contentLength)); err == nil && size >= 0 {
		r.declaredSize = size
	}

	if !r.compressionAllowed() {
		// Nothing is buffered yet, so this only sends the headers and cannot fail.
		_ = r.start(false)
		return
	}

	r.state = stateBuffering
}

// compressionAllowed reports whether the response headers allow compression.
func (r *responseWriter) compressionAllowed() bool {
	header := r.rw.Header()

	// Compressing an already encoded body would need a stacked Content-Encoding, which few clients decode.
	if header.Get(contentEncoding) != "" {
		return false
	}

	// A partial body is a byte range of the uncompressed representation, compressing it would break the range.
	if header.Get(contentRange) != "" {
		return false
	}

	if r.statusCode == http.StatusNoContent || r.statusCode == http.StatusNotModified {
		return false
	}

	// Without a Content-Type, the type is sniffed from the body once it is buffered.
	ct := header.Get(contentType)
	if ct == "" {
		return true
	}

	return r.contentTypeAllowed(ct)
}

// contentTypeAllowed reports whether the configured content types allow compressing this one.
func (r *responseWriter) contentTypeAllowed(ct string) bool {
	mediaType, params, err := mime.ParseMediaType(ct)
	// An unparsable type cannot be matched against the configured ones, so it is left uncompressed.
	if err != nil {
		return false
	}

	matches := func(pct parsedContentType) bool {
		return pct.equals(mediaType, params)
	}

	if len(r.includedContentTypes) > 0 {
		return slices.ContainsFunc(r.includedContentTypes, matches)
	}

	return !slices.ContainsFunc(r.excludedContentTypes, matches)
}

// start makes the compress-or-not decision final, and sends the headers followed by the buffered body.
func (r *responseWriter) start(compress bool) error {
	header := r.rw.Header()

	if compress && header.Get(contentType) == "" && len(r.buf) > 0 {
		ct := http.DetectContentType(r.buf)
		// The stdlib only sniffs the type of bodies that are not encoded, so it has to be set here.
		if _, ok := header[contentType]; !ok {
			header.Set(contentType, ct)
		}

		compress = r.contentTypeAllowed(ct)
	}

	var w io.Writer = r.rw
	r.state = statePlain

	if compress {
		w = r.compressionWriter
		r.state = stateCompressing

		// The compressed size is only known at the end,
		// and a range request would apply to the uncompressed representation.
		header.Del(contentLength)
		header.Del(acceptRanges)
		header.Set(contentEncoding, r.compressionWriter.ContentEncoding())
	}

	r.rw.WriteHeader(r.statusCode)

	if len(r.buf) == 0 {
		return nil
	}

	n, err := w.Write(r.buf)
	if err == nil && n < len(r.buf) {
		err = io.ErrShortWrite
	}

	r.buf = r.buf[:0]

	return err
}

// close sends whatever the handler left pending, and terminates the compressed stream.
// The compression writer is only closed once compression started,
// as closing it writes end-of-stream bytes.
func (r *responseWriter) close() error {
	if r.hijacked {
		return nil
	}

	switch r.state {
	case stateHeadersPending:
		r.rw.WriteHeader(r.statusCode)
		return nil

	case stateBuffering:
		// The body ended before reaching minSize.
		return r.start(false)

	case stateCompressing:
		return r.compressionWriter.Close()

	default:
		return nil
	}
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
