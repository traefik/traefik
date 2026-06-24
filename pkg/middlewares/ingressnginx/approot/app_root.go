package approot

import (
	"context"
	"errors"
	"net/http"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/middlewares/ingressnginx"
)

const typeName = "appRoot"

type appRoot struct {
	name    string
	next    http.Handler
	appRoot string
}

func New(ctx context.Context, next http.Handler, config dynamic.AppRoot, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug().Msg("Creating middleware")

	if config.AppRoot == "" {
		return nil, errors.New("app root cannot be empty")
	}

	return &appRoot{
		name:    name,
		next:    next,
		appRoot: config.AppRoot,
	}, nil
}

func (ar *appRoot) GetTracingInformation() (string, string) {
	return ar.name, typeName
}

func (ar *appRoot) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// request is not to the app root
	if req.URL.Path != "/" {
		ar.next.ServeHTTP(rw, req)
		return
	}

	path := ingressnginx.ReplaceVariables(ar.appRoot, req, nil, nil)
	http.Redirect(rw, req, path, http.StatusFound)
}
