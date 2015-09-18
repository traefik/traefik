package main

import (
	"github.com/thoas/stats"
	"net/http"
)

func main() {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"hello\": \"world\"}"))
	})

	handler := stats.New().Handler(h)
	http.ListenAndServe(":8080", handler)
}
