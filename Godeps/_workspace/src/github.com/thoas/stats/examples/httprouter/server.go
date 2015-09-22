package main

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"github.com/thoas/stats"
	"net/http"
)

func main() {
	router := httprouter.New()
	s := stats.New()
	router.GET("/stats", func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		s, err := json.Marshal(s.Data())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Write(s)
	})
	http.ListenAndServe(":8080", s.Handler(router))
}
