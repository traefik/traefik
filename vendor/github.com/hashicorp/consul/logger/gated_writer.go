package logger

import (
	"io"
	"sync"
)

// GatedWriter is an io.Writer implementation that buffers all of its
// data into an internal buffer until it is told to let data through.
type GatedWriter struct {
	Writer io.Writer

	buf   [][]byte
	flush bool
	lock  sync.RWMutex
}

// Flush tells the GatedWriter to flush any buffered data and to stop
// buffering.
func (w *GatedWriter) Flush() {
	w.lock.Lock()
	w.flush = true
	w.lock.Unlock()

	for _, p := range w.buf {
		w.Write(p)
	}
	w.buf = nil
}

func (w *GatedWriter) Write(p []byte) (n int, err error) {
	// Once we flush we no longer synchronize writers since there's
	// no use of the internal buffer. This is the happy path.
	w.lock.RLock()
	if w.flush {
		w.lock.RUnlock()
		return w.Writer.Write(p)
	}
	w.lock.RUnlock()

	// Now take the write lock.
	w.lock.Lock()
	defer w.lock.Unlock()

	// Things could have changed between the locking operations, so we
	// have to check one more time.
	if w.flush {
		return w.Writer.Write(p)
	}

	// Buffer up the written data.
	p2 := make([]byte, len(p))
	copy(p2, p)
	w.buf = append(w.buf, p2)
	return len(p), nil
}
