package route

import (
	"net/http"
	"strings"
)

// RawPath returns escaped url path section
func rawPath(r *http.Request) string {
	// If there are no escape symbols, don't extract raw path
	if !strings.ContainsRune(r.RequestURI, '%') {
		if len(r.URL.Path) == 0 {
			return "/"
		}
		return r.URL.Path
	}
	path := r.RequestURI
	if path == "" {
		path = "/"
	}
	// This is absolute URI, split host and port
	if strings.Contains(path, "://") {
		vals := strings.SplitN(path, r.URL.Host, 2)
		if len(vals) == 2 {
			path = vals[1]
		}
	}
	idx := strings.IndexRune(path, '?')
	if idx == -1 {
		return path
	}
	return path[:idx]
}
