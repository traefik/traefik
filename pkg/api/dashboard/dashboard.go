package dashboard

import (
	"io/fs"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/traefik/traefik/v2/webui"
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
			http.Redirect(resp, req, safePrefix(req)+"/dashboard/", http.StatusFound)
		})

	router.Methods(http.MethodGet).
		PathPrefix("/dashboard/").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// allow iframes from our domains only
			// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/frame-src
			w.Header().Set("Content-Security-Policy", "frame-src 'self' https://traefik.io https://*.traefik.io;")
			http.StripPrefix("/dashboard/", http.FileServer(http.FS(assets))).ServeHTTP(w, r)
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
	http.FileServer(http.FS(assets)).ServeHTTP(w, r)
}

func safePrefix(req *http.Request) string {
	prefix := req.Header.Get("X-Forwarded-Prefix")
	if prefix == "" {
		return ""
	}

	parse, err := url.Parse(prefix)
	if err != nil {
		return ""
	}

	if parse.Host != "" {
		return ""
	}

	return parse.Path
}
