package externalaction

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
)

func readAndRestoreBody(req *http.Request, limit int64) ([]byte, bool, error) {
	if req.Body == nil || limit <= 0 {
		return nil, false, nil
	}
	original := req.Body
	value, err := io.ReadAll(io.LimitReader(original, limit+1))
	if err != nil {
		return nil, false, err
	}
	if int64(len(value)) > limit {
		req.Body = &multiReadCloser{
			Reader: io.MultiReader(bytes.NewReader(value), original),
			closer: original,
		}
		return value, true, nil
	}
	if err := original.Close(); err != nil {
		return nil, false, err
	}
	req.Body = io.NopCloser(bytes.NewReader(value))
	return value, false, nil
}

type multiReadCloser struct {
	io.Reader

	closer io.Closer
}

func (m *multiReadCloser) Close() error {
	return m.closer.Close()
}

func overflowFor(rule *compiledRule, raw []byte, globalOverflow bool) bool {
	if strings.EqualFold(rule.config.Body.ReadMode, "none") {
		return false
	}
	limit := bodyLimit(rule.config.Body.MaxBytes)
	return globalOverflow || int64(len(raw)) > limit
}

func bodyLimit(configured int64) int64 {
	if configured <= 0 {
		return 1024 * 1024
	}
	return configured
}

func responseBodyLimit(configured int64) int64 {
	if configured <= 0 {
		return 1024 * 1024
	}
	return configured
}

func responseOverflowFor(rule *compiledRule, responseSize int64) bool {
	if !strings.EqualFold(rule.config.ResponseBody.ReadMode, "capture") {
		return false
	}
	return responseSize > responseBodyLimit(rule.config.ResponseBody.MaxBytes)
}

// decodeCapturedResponseBody decodes only the middleware's private response
// copy. The original encoded bytes have already been written to the client by
// responseRecorder and are never changed. A bounded read prevents a small
// compressed response from expanding beyond the configured capture limit.
func decodeCapturedResponseBody(contentEncoding string, raw []byte, limit int64) ([]byte, bool, error) {
	if len(raw) == 0 || strings.TrimSpace(contentEncoding) == "" {
		return raw, false, nil
	}
	if limit <= 0 {
		return nil, true, nil
	}

	encodings := strings.Split(contentEncoding, ",")
	value := raw
	for _, encoding := range slices.Backward(encodings) {
		encoding := strings.ToLower(strings.TrimSpace(strings.SplitN(encoding, ";", 2)[0]))
		if encoding == "" || encoding == "identity" {
			continue
		}

		var reader io.Reader
		var closeReader func()
		switch encoding {
		case "gzip", "x-gzip":
			decoder, err := gzip.NewReader(bytes.NewReader(value))
			if err != nil {
				return nil, false, fmt.Errorf("initializing gzip response decoder: %w", err)
			}
			reader = decoder
			closeReader = func() { _ = decoder.Close() }
		case "deflate":
			decoder, err := zlib.NewReader(bytes.NewReader(value))
			if err != nil {
				return nil, false, fmt.Errorf("initializing deflate response decoder: %w", err)
			}
			reader = decoder
			closeReader = func() { _ = decoder.Close() }
		case "br":
			reader = brotli.NewReader(bytes.NewReader(value))
		case "zstd":
			decoder, err := zstd.NewReader(bytes.NewReader(value))
			if err != nil {
				return nil, false, fmt.Errorf("initializing zstd response decoder: %w", err)
			}
			reader = decoder
			closeReader = decoder.Close
		default:
			return nil, false, fmt.Errorf("unsupported response content encoding %q", encoding)
		}

		decoded, err := io.ReadAll(io.LimitReader(reader, limit+1))
		if closeReader != nil {
			closeReader()
		}
		if err != nil {
			return nil, false, fmt.Errorf("decoding %s response body: %w", encoding, err)
		}
		if int64(len(decoded)) > limit {
			return decoded[:limit], true, nil
		}
		value = decoded
	}
	return value, false, nil
}

type responseRecorder struct {
	http.ResponseWriter

	status    int
	size      int64
	body      bytes.Buffer
	bodyLimit int64
	overflow  bool
}

func newResponseRecorder(rw http.ResponseWriter, bodyLimit int64) *responseRecorder {
	return &responseRecorder{ResponseWriter: rw, bodyLimit: bodyLimit}
}

func (r *responseRecorder) WriteHeader(status int) {
	if r.status != 0 {
		return
	}
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(value []byte) (int, error) {
	if r.status == 0 {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.ResponseWriter.Write(value)
	r.size += int64(n)
	if r.bodyLimit > 0 && n > 0 {
		remaining := r.bodyLimit - int64(r.body.Len())
		if remaining > 0 {
			captured := n
			if int64(captured) > remaining {
				captured = int(remaining)
			}
			_, _ = r.body.Write(value[:captured])
		}
		if int64(r.body.Len()) < r.size {
			r.overflow = true
		}
	}
	return n, err
}

func (r *responseRecorder) Status() int {
	return r.status
}

func (r *responseRecorder) Size() int64 {
	return r.size
}

func (r *responseRecorder) Body() []byte {
	return r.body.Bytes()
}

func (r *responseRecorder) Overflow() bool {
	return r.overflow
}

func (r *responseRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func (r *responseRecorder) Flush() {
	if value, ok := r.ResponseWriter.(http.Flusher); ok {
		value.Flush()
	}
}

func (r *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	value, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("response writer %T does not support hijacking", r.ResponseWriter)
	}
	return value.Hijack()
}

func (r *responseRecorder) Push(target string, options *http.PushOptions) error {
	value, ok := r.ResponseWriter.(http.Pusher)
	if !ok {
		return http.ErrNotSupported
	}
	return value.Push(target, options)
}
