package mirror

import (
	"context"
	"net/http"
	"sync"

	"github.com/containous/traefik/v2/pkg/safe"
)

// Mirroring is an http.Handler that can mirror requests
type Mirroring struct {
	handler        http.Handler
	mirrorHandlers []*mirrorHandler
	rw             http.ResponseWriter
	routinePool    *safe.Pool

	lock  sync.RWMutex
	total uint64
}

// New return new instance of *Mirroring
func New(handler http.Handler, pool *safe.Pool) *Mirroring {
	return &Mirroring{
		routinePool: pool,
		handler:     handler,
		rw:          blackholeResponseWriter{},
	}
}

func (m *Mirroring) inc() uint64 {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.total++
	return m.total
}

type mirrorHandler struct {
	http.Handler
	percent int

	lock  sync.RWMutex
	count uint64
}

func (m *Mirroring) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	m.handler.ServeHTTP(rw, req)

	select {
	case <-req.Context().Done():
		// No mirroring if request has been canceled during main handler ServeHTTP
		return
	default:
	}

	m.routinePool.GoCtx(func(_ context.Context) {
		total := m.inc()
		for _, handler := range m.mirrorHandlers {
			handler.lock.Lock()
			if handler.count*100 < total*uint64(handler.percent) {
				handler.count++
				handler.lock.Unlock()
				handler.ServeHTTP(m.rw, req.WithContext(contextStopPropagation{req.Context()}))
			} else {
				handler.lock.Unlock()
			}
		}
	})
}

// AddMirror adds on httpHandler to mirror on.
func (m *Mirroring) AddMirror(handler http.Handler, percent int) {
	m.mirrorHandlers = append(m.mirrorHandlers, &mirrorHandler{Handler: handler, percent: percent})
}

type blackholeResponseWriter struct{}

func (b blackholeResponseWriter) Header() http.Header {
	return http.Header{}
}

func (b blackholeResponseWriter) Write(bytes []byte) (int, error) {
	return len(bytes), nil
}

func (b blackholeResponseWriter) WriteHeader(statusCode int) {
}

type contextStopPropagation struct {
	context.Context
}

func (c contextStopPropagation) Done() <-chan struct{} {
	return make(chan struct{})
}
