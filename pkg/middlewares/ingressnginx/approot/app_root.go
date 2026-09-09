package approot

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/middlewares/ingressnginx"
)

const typeName = "AppRoot"

type appRoot struct {
	name string
	next http.Handler
	path string
}

func New(ctx context.Context, next http.Handler, config dynamic.AppRoot, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug().Msg("Creating middleware")

	if config.Path == "" {
		return nil, errors.New("path cannot be empty")
	}
	if !strings.HasPrefix(config.Path, "/") {
		return nil, errors.New("path should start with /")
	}

	return &appRoot{
		name: name,
		next: next,
		path: config.Path,
	}, nil
}

func (ar *appRoot) GetTracingInformation() (string, string) {
	return ar.name, typeName
}

func (ar *appRoot) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Request is not to the app root.
	if req.URL.Path != "/" {
		ar.next.ServeHTTP(rw, req)
		return
	}

	path := ingressnginx.ReplaceVariables("$scheme://$best_http_host"+ar.path, req, nil, nil)
	http.Redirect(rw, req, path, http.StatusFound)
}
