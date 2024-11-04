package dashboard

import (
	"io/fs"
	"net/http"
	"net/url"
	"strings"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/webui"
)

// Handler expose dashboard routes.
type Handler struct {
	assets   fs.FS // optional assets, to override the webui.FS default
	BasePath string
}

// Append adds dashboard routes on the given router, optionally using the given
// assets (or webui.FS otherwise).
func Append(router *mux.Router, basePath string, customAssets fs.FS) {
	assets := customAssets
	if assets == nil {
		assets = webui.FS
	}

	indexTemplate, err := template.ParseFS(assets, "index.html")
	if err != nil {
		log.Error().Err(err).Msg("unable to load index.html")
	}

	dashboardPath := strings.TrimSuffix(basePath, "/") + "/dashboard/"

	// Expose dashboard
	router.Methods(http.MethodGet).
		Path(strings.TrimSuffix(dashboardPath, "/")).
		HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			http.Redirect(resp, req, safePrefix(req)+dashboardPath, http.StatusFound)
		})

	router.Methods(http.MethodGet).
		Path(dashboardPath).
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// allow iframes from our domains only
			// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/frame-src
			w.Header().Set("Content-Security-Policy", "frame-src 'self' https://traefik.io https://*.traefik.io;")

			// The content type must be guessed by the file server.
			// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Content-Type-Options
			w.Header().Del("Content-Type")

			apiPath := strings.TrimSuffix(basePath, "/") + "/api/"
			if err = indexTemplate.Execute(w, indexTemplateData{APIUrl: apiPath}); err != nil {
				log.Error().Err(err).Msg("Unable to serve APIPortal index.html page")
			}
		})

	router.Methods(http.MethodGet).
		PathPrefix(dashboardPath).
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// allow iframes from our domains only
			// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/frame-src
			w.Header().Set("Content-Security-Policy", "frame-src 'self' https://traefik.io https://*.traefik.io;")

			// The content type must be guessed by the file server.
			// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Content-Type-Options
			w.Header().Del("Content-Type")

			http.StripPrefix(dashboardPath, http.FileServerFS(assets)).ServeHTTP(w, r)
		})
}

type indexTemplateData struct {
	APIUrl string
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

	if r.RequestURI == "/" {
		indexTemplate, err := template.ParseFS(assets, "index.html")
		if err != nil {
			log.Error().Err(err).Msg("Unable to serve APIPortal index.html page")
		}

		apiPath := strings.TrimSuffix(g.BasePath, "/") + "/api/"
		if err = indexTemplate.Execute(w, indexTemplateData{APIUrl: apiPath}); err != nil {
			log.Error().Err(err).Msg("Unable to serve APIPortal index.html page")
		}

		return
	}

	http.FileServerFS(assets).ServeHTTP(w, r)
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
