package middlewares

import (
	"bufio"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/containous/traefik/log"
	"github.com/vulcand/oxy/utils"
)

// Compile time validation responseRecorder implements http interfaces correctly.
var (
	_                      Stateful  = &retryResponseRecorder{}
	_                      io.Reader = (*retryResponseRecorder)(nil)
	ErrorInconsistentWrite           = errors.New("inconsistent write")
	ErrorNotEnoughSpace              = errors.New("not enough space")
)

// RetrySettings is a settings for Retry middleware
type RetrySettings struct {
	Attempts             int           // Number of attempts
	CacheInitialCapacity int           // Initial capacity of cache when creating (in bytes)
	CacheMaxCapacity     int           // Maximum allowed capacity for cache (in bytes) when upstream data is bigger than this value, retry starts caching data into file. if set to 0 all data will be cached in memory
	TempDir              string        // Place to store temporary files
	RetryInterval        time.Duration // Retry interval in milliseconds before next try
}

// Retry is a middleware that retries requests
type Retry struct {
	next     http.Handler
	listener RetryListener
	conf     *RetrySettings
}

// NewRetry returns a new Retry instance
func NewRetry(conf *RetrySettings, next http.Handler, listener RetryListener) *Retry {
	return &Retry{
		next:     next,
		listener: listener,
		conf:     conf,
	}
}

func (retry *Retry) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// if we might make multiple attempts, swap the body for an ioutil.NopCloser
	// cf https://github.com/containous/traefik/issues/1008
	if retry.conf.Attempts > 1 {
		body := r.Body
		defer body.Close()
		r.Body = ioutil.NopCloser(body)
	}
	attempts := 1

	recorder := newRetryResponseRecorder(int64(retry.conf.CacheInitialCapacity), int64(retry.conf.CacheMaxCapacity), &retry.conf.TempDir)
	recorder.responseWriter = rw
	defer recorder.Reset()

	for {
		netErrorOccurred := false
		// We pass in a pointer to netErrorOccurred so that we can set it to true on network errors
		// when proxying the HTTP requests to the backends. This happens in the custom RecordingErrorHandler.
		newCtx := context.WithValue(r.Context(), defaultNetErrCtxKey, &netErrorOccurred)

		retry.next.ServeHTTP(recorder, r.WithContext(newCtx))

		// It's a stream request and the body gets already sent to the client.
		// Therefore we should not send the response a second time.
		if recorder.streamingResponseStarted {
			break
		}

		if !netErrorOccurred || attempts >= retry.conf.Attempts {
			utils.CopyHeaders(rw.Header(), recorder.Header())
			rw.WriteHeader(recorder.Code)
			recorder.Seek(0, 0)
			io.Copy(rw, recorder)
			break
		}
		attempts++
		log.Debugf("New attempt %d for request: %v", attempts, r.URL)
		retry.listener.Retried(r, attempts)
		recorder.Reset()

		if retry.conf.RetryInterval != 0 {
			<-time.After(retry.conf.RetryInterval * time.Millisecond)
		}
	}
}

// netErrorCtxKey is a custom type that is used as key for the context.
type netErrorCtxKey string

// defaultNetErrCtxKey is the actual key which value is used to record network errors.
var defaultNetErrCtxKey netErrorCtxKey = "NetErrCtxKey"

// NetErrorRecorder is an interface to record net errors.
type NetErrorRecorder interface {
	// Record can be used to signal the retry middleware that an network error happened
	// and therefore the request should be retried.
	Record(ctx context.Context)
}

// DefaultNetErrorRecorder is the default NetErrorRecorder implementation.
type DefaultNetErrorRecorder struct{}

// Record is recording network errors by setting the context value for the defaultNetErrCtxKey to true.
func (DefaultNetErrorRecorder) Record(ctx context.Context) {
	val := ctx.Value(defaultNetErrCtxKey)

	if netErrorOccurred, isBoolPointer := val.(*bool); isBoolPointer {
		*netErrorOccurred = true
	}
}

// RetryListener is used to inform about retry attempts.
type RetryListener interface {
	// Retried will be called when a retry happens, with the request attempt passed to it.
	// For the first retry this will be attempt 2.
	Retried(req *http.Request, attempt int)
}

// RetryListeners is a convenience type to construct a list of RetryListener and notify
// each of them about a retry attempt.
type RetryListeners []RetryListener

// Retried exists to implement the RetryListener interface. It calls Retried on each of its slice entries.
func (l RetryListeners) Retried(req *http.Request, attempt int) {
	for _, retryListener := range l {
		retryListener.Retried(req, attempt)
	}
}

// retryResponseRecorder is an implementation of http.ResponseWriter that
// records its mutations for later inspection.
type retryResponseRecorder struct {
	Code      int         // the HTTP response code from WriteHeader
	HeaderMap http.Header // the HTTP response headers

	responseWriter           http.ResponseWriter
	err                      error
	streamingResponseStarted bool

	// hot
	hotMem          []byte
	initialCapacity int64
	hotMemCapacity  int64
	readPos         int64

	// cold
	fileReader *bufio.Reader
	fileWriter *bufio.Writer
	fileDir    *string
	file       *os.File
}

// newRetryResponseRecorder returns an initialized retryResponseRecorder.
func newRetryResponseRecorder(initialCapacity int64, maxCapacity int64, fileDir *string) *retryResponseRecorder {
	if fileDir == nil {
		maxCapacity = 0
	}

	return &retryResponseRecorder{
		HeaderMap:       make(http.Header),
		Code:            http.StatusOK,
		initialCapacity: initialCapacity,
		hotMemCapacity:  maxCapacity,
		hotMem:          make([]byte, 0, initialCapacity),
		readPos:         0,
		fileDir:         fileDir,
	}
}

// Header returns the response headers.
func (rw *retryResponseRecorder) Header() http.Header {
	m := rw.HeaderMap
	if m == nil {
		m = make(http.Header)
		rw.HeaderMap = m
	}
	return m
}

// Seek sets the offset for the next Read or Write on file to offset, interpreted
// according to whence: 0 means relative to the origin of the file, 1 means
// relative to the current offset, and 2 means relative to the end.
// It returns the new offset and an error, if any.
// The behavior of Seek on a file opened with O_APPEND is not specified.
func (rw *retryResponseRecorder) Seek(offset int64, whence int) (int64, error) {
	if rw.file != nil {
		return rw.file.Seek(offset, whence)
	}

	hotMemLen := int64(len(rw.hotMem))

	if whence == io.SeekStart {
		rw.readPos = offset
	} else if whence == io.SeekCurrent {
		rw.readPos = rw.readPos + offset
	} else if whence == io.SeekEnd {
		rw.readPos = hotMemLen + offset
	}

	if rw.readPos > hotMemLen {
		rw.readPos = hotMemLen
	} else if rw.readPos < 0 {
		rw.readPos = 0
	}
	return rw.readPos, nil
}

func (rw *retryResponseRecorder) Read(buf []byte) (n int, err error) {
	if rw.fileWriter != nil {
		return rw.fileReader.Read(buf)
	}

	n = copy(buf, rw.hotMem[rw.readPos:])
	rw.readPos += int64(n)
	err = nil
	if rw.readPos == int64(len(rw.hotMem)) {
		err = io.EOF
	}
	return n, err
}

// Write always succeeds and writes to rw.Body, if not nil.
func (rw *retryResponseRecorder) Write(buf []byte) (n int, err error) {
	if rw.err != nil {
		return 0, rw.err
	}

	if rw.fileWriter != nil {
		// cold mem is already enabled
		return rw.fileWriter.Write(buf)
	}

	// still writing to hot mem
	if rw.hotMemCapacity == 0 || int64(len(buf)+len(rw.hotMem)) < rw.hotMemCapacity {
		// enough capacity for buf in hotMem
		oldLen := len(rw.hotMem)
		rw.hotMem = append(rw.hotMem, buf...)
		return len(rw.hotMem) - oldLen, nil
	}

	// create file
	if rw.fileDir == nil {
		return 0, ErrorNotEnoughSpace
	}
	rw.file, err = ioutil.TempFile(*rw.fileDir, "traefik-")
	if err != nil {
		log.Errorf("Failed to open temp file in retryResponseRecorder in folder(%s): %s", *rw.fileDir, err.Error())
		return 0, err
	}
	rw.fileWriter = bufio.NewWriter(rw.file)
	rw.fileReader = bufio.NewReader(rw.file)

	n, err = rw.fileWriter.Write(rw.hotMem)
	if err != nil {
		return 0, err
	}
	if n != len(rw.hotMem) {
		return 0, ErrorInconsistentWrite
	}
	rw.fileWriter.Flush()
	rw.hotMem = make([]byte, 0, 0)

	return rw.Write(buf)
}

// WriteHeader sets rw.Code.
func (rw *retryResponseRecorder) WriteHeader(code int) {
	rw.Code = code
}

// Hijack hijacks the connection
func (rw *retryResponseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return rw.responseWriter.(http.Hijacker).Hijack()
}

// CloseNotify returns a channel that receives at most a
// single value (true) when the client connection has gone
// away.
func (rw *retryResponseRecorder) CloseNotify() <-chan bool {
	return rw.responseWriter.(http.CloseNotifier).CloseNotify()
}

// Flush sends any buffered data to the client.
func (rw *retryResponseRecorder) Flush() {
	if !rw.streamingResponseStarted {
		utils.CopyHeaders(rw.responseWriter.Header(), rw.Header())
		rw.responseWriter.WriteHeader(rw.Code)
		rw.streamingResponseStarted = true
	}

	rw.Seek(0, 0)
	_, err := io.Copy(rw.responseWriter, rw)
	if err != nil {
		log.Errorf("Error writing response in retryResponseRecorder: %s", err)
		rw.err = err
	}
	rw.Reset()
	// rw.Body.Reset()
	flusher, ok := rw.responseWriter.(http.Flusher)
	if ok {
		flusher.Flush()
	}
}

func (rw *retryResponseRecorder) Reset() {
	rw.hotMem = make([]byte, 0, rw.initialCapacity)
	rw.readPos = 0

	rw.fileReader = nil
	rw.fileWriter = nil
	rw.fileDir = nil

	if rw.file != nil {
		rw.file.Close()
		os.Remove(rw.file.Name())
		rw.file = nil
	}
}
