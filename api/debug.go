package api

import (
	"expvar"
	"fmt"

	"github.com/containous/mux"
	"net/http"
	"runtime"
)

func init() {
	expvar.Publish("Goroutines", expvar.Func(goroutines))
}

func goroutines() interface{} {
	return runtime.NumGoroutine()
}

type DebugHandler struct{}

func (g DebugHandler) AddRoutes(router *mux.Router) {
	router.Methods("GET").Path("/debug/vars").HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
}
