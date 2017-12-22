package server

import (
	"net/http"
)

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}
