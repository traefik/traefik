package api

import (
	"expvar"
	"fmt"
	"net/http"
	"net/http/pprof"
	"runtime"

	"github.com/gorilla/mux"
)

func init() {
	// TODO Goroutines2 -> Goroutines
	expvar.Publish("Goroutines2", expvar.Func(goroutines))
}

func goroutines() interface{} {
	return runtime.NumGoroutine()
}

// DebugHandler expose debug routes.
type DebugHandler struct{}

// Append add debug routes on a router.
func (g DebugHandler) Append(router *mux.Router) {
	router.Methods(http.MethodGet).Path("/debug/vars").
		HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			fmt.Fprint(w, "{\n")
			first := true
			expvar.Do(func(kv expvar.KeyValue) {
				if !first {
					fmt.Fprint(w, ",\n")
				}
				first = false
				fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
			})
			fmt.Fprint(w, "\n}\n")
		})

	runtime.SetBlockProfileRate(1)
	runtime.SetMutexProfileFraction(5)
	router.Methods(http.MethodGet).PathPrefix("/debug/pprof/cmdline").HandlerFunc(pprof.Cmdline)
	router.Methods(http.MethodGet).PathPrefix("/debug/pprof/profile").HandlerFunc(pprof.Profile)
	router.Methods(http.MethodGet).PathPrefix("/debug/pprof/symbol").HandlerFunc(pprof.Symbol)
	router.Methods(http.MethodGet).PathPrefix("/debug/pprof/trace").HandlerFunc(pprof.Trace)
	router.Methods(http.MethodGet).PathPrefix("/debug/pprof/").HandlerFunc(pprof.Index)
}
