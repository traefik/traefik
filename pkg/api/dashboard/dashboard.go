package dashboard

import (
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
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

	// allow iframes from our domains only
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/frame-src
	w.Header().Set("Content-Security-Policy", "frame-src 'self' https://traefik.io https://*.traefik.io;")

	// The content type must be guessed by the file server.
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Content-Type-Options
	w.Header().Del("Content-Type")

	if r.URL.Path == "/" {
		indexTemplate, err := template.ParseFS(assets, "index.html")
		if err != nil {
			log.Error().Err(err).Msg("Unable to parse index template")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		apiPath := strings.TrimSuffix(h.BasePath, "/") + "/api/"
		if err = indexTemplate.Execute(w, indexTemplateData{APIUrl: apiPath}); err != nil {
			log.Error().Err(err).Msg("Unable to render index template")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		return
	}

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
				log.Error().Err(err).Msg("Unable to render index template")
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
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
	return nil
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
