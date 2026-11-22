package mirror

import (
	"bytes"
	"errors"
	"io"
	"sync"
)

// streamingBufferSize is the maximum size of the request body buffered for a mirror reading it slower than the main handler,
// unless a single read of the main handler is larger.
// Beyond it, the mirrored request is aborted, so that a slow mirror never slows the main request down.
const streamingBufferSize = 1024 * 1024

var (
	errMirrorTooSlow     = errors.New("mirror too slow to read the request body")
	errBodyNotFullyRead  = errors.New("request body not fully read by the main handler")
	errStreamedBodyClose = errors.New("mirrored request body closed")
)

// teeBody copies the request body to the mirrors while the main handler reads it.
type teeBody struct {
	io.ReadCloser

	mirrors []*streamedBody
}

func (t *teeBody) Read(p []byte) (int, error) {
	n, err := t.ReadCloser.Read(p)
	for _, mirror := range t.mirrors {
		if n > 0 {
			mirror.write(p[:n])
		}
		if err != nil {
			// io.EOF ends the mirrored request body normally, any other error aborts the mirrored request.
			mirror.closeWithError(err)
		}
	}

	return n, err
}

// streamedBody is the body of a mirrored request.
// It buffers the bytes read by the main handler until the mirror reads them,
// and aborts the mirrored request when the mirror falls behind by more than its buffer size.
type streamedBody struct {
	mu   sync.Mutex
	cond *sync.Cond
	buf  bytes.Buffer
	err  error
}

func newStreamedBody() *streamedBody {
	b := &streamedBody{}
	b.cond = sync.NewCond(&b.mu)
	return b
}

func (b *streamedBody) Read(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for b.buf.Len() == 0 && b.err == nil {
		b.cond.Wait()
	}

	if b.buf.Len() > 0 {
		return b.buf.Read(p)
	}

	return 0, b.err
}

// Close is called once the mirrored request is done, so the next writes are dropped.
func (b *streamedBody) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err == nil {
		b.err = errStreamedBodyClose
	}
	b.buf.Reset()
	b.cond.Broadcast()

	return nil
}

func (b *streamedBody) write(p []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return
	}

	// A chunk is always accepted when the mirror has read everything,
	// so that a mirror keeping up is not aborted by a read larger than the buffer size.
	if b.buf.Len() > 0 && b.buf.Len()+len(p) > streamingBufferSize {
		b.err = errMirrorTooSlow
		b.buf.Reset()
		b.cond.Broadcast()
		return
	}

	b.buf.Write(p)
	b.cond.Broadcast()
}

// closeWithError ends the mirrored request body once its buffered bytes are read.
// The first error wins, so it is a no-op when the body already ended.
func (b *streamedBody) closeWithError(err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err == nil {
		b.err = err
		b.cond.Broadcast()
	}
}
