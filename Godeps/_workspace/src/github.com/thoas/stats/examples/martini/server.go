package main

import (
	"encoding/json"
	"github.com/go-martini/martini"
	"github.com/thoas/stats"
	"net/http"
)

func main() {
	middleware := stats.New()

	m := martini.Classic()
	m.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"hello\": \"world\"}"))
	})
	m.Get("/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		stats := middleware.Data()

		b, _ := json.Marshal(stats)

		w.Write(b)
	})

	m.Use(func(c martini.Context, w http.ResponseWriter, r *http.Request) {
		beginning, recorder := middleware.Begin(w)

		c.Next()

		middleware.End(beginning, recorder)
	})
	m.Run()
}
