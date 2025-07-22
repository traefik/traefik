package dashboard

import (
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/webui"
)

type indexTemplateData struct {
	APIUrl string
}

// Handler expose dashboard routes.
type Handler struct {
	BasePath string

	assets fs.FS // optional assets, to override the webui.FS default
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	assets := h.assets
	if assets == nil {
		assets = webui.FS
	}

	// Allow iframes from traefik domains only.
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/frame-src
	w.Header().Set("Content-Security-Policy", "frame-src 'self' https://traefik.io https://*.traefik.io;")

	if r.RequestURI == "/" {
		indexTemplate, err := template.ParseFS(assets, "index.html")
		if err != nil {
			log.Error().Err(err).Msg("Unable to parse index template")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		apiPath := strings.TrimSuffix(h.BasePath, "/") + "/api/"
		if err = indexTemplate.Execute(w, indexTemplateData{APIUrl: apiPath}); err != nil {
			log.Error().Err(err).Msg("Unable to render index template")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		return
	}

	// The content type must be guessed by the file server.
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Content-Type-Options
	w.Header().Del("Content-Type")

	http.FileServerFS(assets).ServeHTTP(w, r)
}

// Append adds dashboard routes on the given router, optionally using the given
// assets (or webui.FS otherwise).
func Append(router *mux.Router, basePath string, customAssets fs.FS) error {
	assets := customAssets
	if assets == nil {
		assets = webui.FS
	}

	indexTemplate, err := template.ParseFS(assets, "index.html")
	if err != nil {
		return fmt.Errorf("parsing index template: %w", err)
	}

	dashboardPath := strings.TrimSuffix(basePath, "/") + "/dashboard/"

	// Expose dashboard
	router.Methods(http.MethodGet).
		Path(basePath).
		HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			prefix := strings.TrimSuffix(req.Header.Get("X-Forwarded-Prefix"), "/")
			http.Redirect(resp, req, prefix+dashboardPath, http.StatusFound)
		})

	router.Methods(http.MethodGet).
		Path(dashboardPath).
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Allow iframes from our domains only.
			// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/frame-src
			w.Header().Set("Content-Security-Policy", "frame-src 'self' https://traefik.io https://*.traefik.io;")

			w.Header().Set("Content-Type", "text/html; charset=utf-8")

			apiPath := strings.TrimSuffix(basePath, "/") + "/api/"
			if err = indexTemplate.Execute(w, indexTemplateData{APIUrl: apiPath}); err != nil {
				log.Error().Err(err).Msg("Unable to render index template")
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		})

	router.Methods(http.MethodGet).
		PathPrefix(dashboardPath).
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Allow iframes from traefik domains only.
			// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/frame-src
			w.Header().Set("Content-Security-Policy", "frame-src 'self' https://traefik.io https://*.traefik.io;")

			// The content type must be guessed by the file server.
			// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Content-Type-Options
			w.Header().Del("Content-Type")

			http.StripPrefix(dashboardPath, http.FileServerFS(assets)).ServeHTTP(w, r)
		})

	return nil
}
