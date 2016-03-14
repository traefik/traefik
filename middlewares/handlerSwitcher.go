package middlewares

import (
	"github.com/gorilla/mux"
	"net/http"
	"sync"
)

// HandlerSwitcher allows hot switching of http.ServeMux
type HandlerSwitcher struct {
	handler     *mux.Router
	handlerLock *sync.Mutex
}

// NewHandlerSwitcher builds a new instance of HandlerSwitcher
func NewHandlerSwitcher(newHandler *mux.Router) (hs *HandlerSwitcher) {
	return &HandlerSwitcher{
		handler:     newHandler,
		handlerLock: &sync.Mutex{},
	}
}

func (hs *HandlerSwitcher) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	hs.handlerLock.Lock()
	handlerBackup := hs.handler
	hs.handlerLock.Unlock()
	handlerBackup.ServeHTTP(rw, r)
}

// GetHandler returns the current http.ServeMux
func (hs *HandlerSwitcher) GetHandler() (newHandler *mux.Router) {
	return hs.handler
}

// UpdateHandler safely updates the current http.ServeMux with a new one
func (hs *HandlerSwitcher) UpdateHandler(newHandler *mux.Router) {
	hs.handlerLock.Lock()
	hs.handler = newHandler
	defer hs.handlerLock.Unlock()
}
