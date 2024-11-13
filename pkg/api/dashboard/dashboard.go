package dashboard

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/traefik/traefik/v3/webui"
)

// Handler expose dashboard routes.
type Handler struct {
	assets fs.FS // optional assets, to override the webui.FS default
}

// Append adds dashboard routes on the given router, optionally using the given
// assets (or webui.FS otherwise).
func Append(router *mux.Router, customAssets fs.FS) {
	assets := customAssets
	if assets == nil {
		assets = webui.FS
	}
	// Expose dashboard
	router.Methods(http.MethodGet).
		Path("/").
		HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			prefix := strings.TrimSuffix(req.Header.Get("X-Forwarded-Prefix"), "/")
			http.Redirect(resp, req, prefix+"/dashboard/", http.StatusFound)
		})

	router.Methods(http.MethodGet).
		PathPrefix("/dashboard/").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// allow iframes from our domains only
			// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/frame-src
			w.Header().Set("Content-Security-Policy", "frame-src 'self' https://traefik.io https://*.traefik.io;")

			// The content type must be guessed by the file server.
			// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Content-Type-Options
			w.Header().Del("Content-Type")

			http.StripPrefix("/dashboard/", http.FileServerFS(assets)).ServeHTTP(w, r)
		})
}

func (g Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	assets := g.assets
	if assets == nil {
		assets = webui.FS
	}
	// allow iframes from our domains only
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/frame-src
	w.Header().Set("Content-Security-Policy", "frame-src 'self' https://traefik.io https://*.traefik.io;")

	// The content type must be guessed by the file server.
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Content-Type-Options
	w.Header().Del("Content-Type")

	http.FileServerFS(assets).ServeHTTP(w, r)
}
