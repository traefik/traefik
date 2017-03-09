package manners

import (
	"net"
	"net/http"
	"sync"
)

var (
	defaultServer     *GracefulServer
	defaultServerLock = &sync.Mutex{}
)

func init() {
	defaultServerLock.Lock()
}

// ListenAndServe provides a graceful version of the function provided by the
// net/http package. Call Close() to stop the server.
func ListenAndServe(addr string, handler http.Handler) error {
	defaultServer = NewWithServer(&http.Server{Addr: addr, Handler: handler})
	defaultServerLock.Unlock()
	return defaultServer.ListenAndServe()
}

// ListenAndServeTLS provides a graceful version of the function provided by the
// net/http package. Call Close() to stop the server.
func ListenAndServeTLS(addr string, certFile string, keyFile string, handler http.Handler) error {
	defaultServer = NewWithServer(&http.Server{Addr: addr, Handler: handler})
	defaultServerLock.Unlock()
	return defaultServer.ListenAndServeTLS(certFile, keyFile)
}

// Serve provides a graceful version of the function provided by the net/http
// package. Call Close() to stop the server.
func Serve(l net.Listener, handler http.Handler) error {
	defaultServer = NewWithServer(&http.Server{Handler: handler})
	defaultServerLock.Unlock()
	return defaultServer.Serve(l)
}

// Shuts down the default server used by ListenAndServe, ListenAndServeTLS and
// Serve. It returns true if it's the first time Close is called.
func Close() bool {
	defaultServerLock.Lock()
	return defaultServer.Close()
}
