package urlrewrite

import (
	"context"
	"net/http"
	"path"
	"strings"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"go.opentelemetry.io/otel/trace"
)

const (
	typeName = "URLRewrite"
)

type urlRewrite struct {
	name       string
	next       http.Handler
	hostname   *string
	path       *string
	pathPrefix *string
}

// NewURLRewrite creates a URL rewrite middleware.
func NewURLRewrite(ctx context.Context, next http.Handler, conf dynamic.URLRewrite, name string) (http.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, typeName)
	logger.Debug().Msg("Creating middleware")

	return urlRewrite{
		name:       name,
		next:       next,
		hostname:   conf.Hostname,
		path:       conf.Path,
		pathPrefix: conf.PathPrefix,
	}, nil
}

func (u urlRewrite) GetTracingInformation() (string, string, trace.SpanKind) {
	return u.name, typeName, trace.SpanKindInternal
}

func (u urlRewrite) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	newPath := req.URL.Path

	if u.path != nil && u.pathPrefix == nil {
		newPath = *u.path
	}

	if u.path != nil && u.pathPrefix != nil {
		newPath = path.Join(*u.path, strings.TrimPrefix(req.URL.Path, *u.pathPrefix))

		// add the trailing slash if needed, as path.Join removes trailing slashes.
		if strings.HasSuffix(req.URL.Path, "/") && !strings.HasSuffix(newPath, "/") {
			newPath += "/"
		}
	}

	req.URL.Path = newPath
	req.URL.RawPath = req.URL.EscapedPath()
	req.RequestURI = req.URL.RequestURI()

	if u.hostname != nil {
		req.Host = *u.hostname
	}

	u.next.ServeHTTP(rw, req)
}
