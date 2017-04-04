package accesslog

import (
	"io"
	"sync"
)

// linearWriter ensures that writes are not intermingled by applying a lock around every Write call.
type linearWriter struct {
	w  io.Writer
	wl *sync.Mutex
}

// LinearWriter returns a WriteCloser that allows concurrent writing by many goroutines. Each Write
// call is surrounded by a lock so that they are interleaved cleanly.
func LinearWriter(w io.Writer) io.WriteCloser {
	return &linearWriter{w, &sync.Mutex{}}
}

func (lw *linearWriter) Write(p []byte) (n int, err error) {
	lw.wl.Lock()
	defer lw.wl.Unlock()
	return lw.w.Write(p)
}

func (lw *linearWriter) Flush() error {
	if f, ok := lw.w.(Flusher); ok {
		return f.Flush()
	}
	return nil
}

func (lw *linearWriter) Close() error {
	if c, ok := lw.w.(io.Closer); ok {
		return c.Close()
	}
	return nil
}
