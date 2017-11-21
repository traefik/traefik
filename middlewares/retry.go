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
	"sync"
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
	CacheMaxCapacity     int           // Maximum allowed capacity for cache (in bytes) when upstream data is bigger than this value, retry starts caching data into file
	TempDir              string        // Place to store temporary files
	RetryInterval        time.Duration // Retry interval in milliseconds before next try
}

// Retry is a middleware that retries requests
type Retry struct {
	next       http.Handler
	listener   RetryListener
	hotMemPool sync.Pool
	conf       *RetrySettings
}

// NewRetry returns a new Retry instance
func NewRetry(conf *RetrySettings, next http.Handler, listener RetryListener) *Retry {
	retry := &Retry{
		next:     next,
		listener: listener,
		conf:     conf,
	}

	retry.hotMemPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, retry.conf.CacheInitialCapacity)
		},
	}

	return retry
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

	buf := retry.hotMemPool.Get().([]byte)
	defer retry.hotMemPool.Put(buf)

	recorder := newRetryResponseRecorder(buf, int64(retry.conf.CacheMaxCapacity), &retry.conf.TempDir)
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
	hotMem            []byte
	hotMemCapacity    int64
	writePos, readPos int64

	// cold
	fileReader *bufio.Reader
	fileWriter *bufio.Writer
	fileDir    *string
	file       *os.File
}

// newRetryResponseRecorder returns an initialized retryResponseRecorder.
func newRetryResponseRecorder(buff []byte, maxCapacity int64, fileDir *string) *retryResponseRecorder {
	if fileDir == nil {
		maxCapacity *= 100
	}

	return &retryResponseRecorder{
		HeaderMap:      make(http.Header),
		Code:           http.StatusOK,
		hotMemCapacity: maxCapacity,
		hotMem:         buff,
		writePos:       0,
		readPos:        0,
		fileDir:        fileDir,
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

	if whence == io.SeekStart {
		rw.readPos = offset
	} else if whence == io.SeekCurrent {
		rw.readPos = rw.readPos + offset
	} else if whence == io.SeekEnd {
		rw.readPos = rw.writePos + offset
	}

	if rw.readPos > rw.writePos {
		rw.readPos = rw.writePos
	} else if rw.readPos < 0 {
		rw.readPos = 0
	}
	return rw.readPos, nil
}

func (rw *retryResponseRecorder) Read(buf []byte) (n int, err error) {
	if rw.fileWriter != nil {
		return rw.fileReader.Read(buf)
	}

	n = copy(buf, rw.hotMem[rw.readPos:rw.writePos])
	rw.readPos += int64(n)
	err = nil
	if rw.readPos == rw.writePos {
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
	if int64(len(buf))+rw.writePos < rw.hotMemCapacity {
		// enough capacity for buf in hotMem
		if int64(len(buf)) > int64(len(rw.hotMem))-rw.writePos {
			// not enough space in current hotMem
			prevPos := rw.writePos
			n = copy(rw.hotMem[rw.writePos:], buf)
			rw.writePos += int64(n)
			rw.hotMem = append(rw.hotMem, buf[n:]...)
			rw.writePos = int64(len(rw.hotMem))
			n = int(rw.writePos - prevPos)
		} else {
			n = copy(rw.hotMem[rw.writePos:], buf)
			rw.writePos += int64(n)
		}
		return n, nil
	} else {
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

		n, err := rw.fileWriter.Write(rw.hotMem[:rw.writePos])
		if err != nil {
			return 0, err
		}
		if int64(n) != rw.writePos {
			return 0, ErrorInconsistentWrite
		}
		rw.fileWriter.Flush()
		rw.writePos = 0

		return rw.Write(buf)
	}
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
	rw.writePos = 0
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
