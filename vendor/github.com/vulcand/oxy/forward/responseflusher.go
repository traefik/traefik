package forward

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
)

var (
	_ http.Hijacker      = &responseFlusher{}
	_ http.Flusher       = &responseFlusher{}
	_ http.CloseNotifier = &responseFlusher{}
)

type responseFlusher struct {
	http.ResponseWriter
	flush bool
}

func newResponseFlusher(rw http.ResponseWriter, flush bool) *responseFlusher {
	return &responseFlusher{
		ResponseWriter: rw,
		flush:          flush,
	}
}

func (wf *responseFlusher) Write(p []byte) (int, error) {
	written, err := wf.ResponseWriter.Write(p)
	if wf.flush {
		wf.Flush()
	}
	return written, err
}

func (wf *responseFlusher) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := wf.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("the ResponseWriter doesn't support the Hijacker interface")
	}
	return hijacker.Hijack()
}

func (wf *responseFlusher) CloseNotify() <-chan bool {
	return wf.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (wf *responseFlusher) Flush() {
	flusher, ok := wf.ResponseWriter.(http.Flusher)
	if ok {
		flusher.Flush()
	}
}
