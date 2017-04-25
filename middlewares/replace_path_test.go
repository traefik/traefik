package middlewares_test

import (
	"net/http"
	"testing"

	"github.com/containous/traefik/middlewares"
)

func TestReplacePath(t *testing.T) {
	const replacementPath = "/replacement-path"

	paths := []string{
		"/example",
		"/some/really/long/path",
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			var newPath, oldPath string
			handler := &middlewares.ReplacePath{
				Path: replacementPath,
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					newPath = r.URL.Path
					oldPath = r.Header.Get("X-Replaced-Path")
				}),
			}

			req, err := http.NewRequest("GET", "http://localhost"+path, nil)
			if err != nil {
				t.Error(err)
			}

			handler.ServeHTTP(nil, req)
			if newPath != replacementPath {
				t.Fatalf("new path should be '%s'", replacementPath)
			}

			if oldPath != path {
				t.Fatalf("old path should be '%s'", path)
			}
		})
	}
}
