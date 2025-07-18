package redirect

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"path"
	"strings"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
)

const typeName = "RequestRedirect"

type redirect struct {
	name string
	next http.Handler

	scheme     *string
	hostname   *string
	port       *string
	path       *string
	pathPrefix *string
	statusCode int
}

// NewRequestRedirect creates a redirect middleware.
func NewRequestRedirect(ctx context.Context, next http.Handler, conf dynamic.RequestRedirect, name string) (http.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, typeName)
	logger.Debug().Msg("Creating middleware")

	statusCode := conf.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusFound
	}
	if statusCode != http.StatusMovedPermanently && statusCode != http.StatusFound {
		return nil, fmt.Errorf("unsupported status code: %d", statusCode)
	}

	return redirect{
		name:       name,
		next:       next,
		scheme:     conf.Scheme,
		hostname:   conf.Hostname,
		port:       conf.Port,
		path:       conf.Path,
		pathPrefix: conf.PathPrefix,
		statusCode: statusCode,
	}, nil
}

func (r redirect) GetTracingInformation() (string, string) {
	return r.name, typeName
}

func (r redirect) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	redirectURL := *req.URL
	redirectURL.Host = req.Host

	if r.scheme != nil {
		redirectURL.Scheme = *r.scheme
	}

	host := redirectURL.Hostname()
	if r.hostname != nil {
		host = *r.hostname
	}

	port := redirectURL.Port()
	if r.port != nil {
		port = *r.port
	}

	if port != "" {
		host = net.JoinHostPort(host, port)
	}
	redirectURL.Host = host

	if r.path != nil && r.pathPrefix == nil {
		redirectURL.Path = *r.path
	}

	if r.path != nil && r.pathPrefix != nil {
		redirectURL.Path = path.Join(*r.path, strings.TrimPrefix(req.URL.Path, *r.pathPrefix))

		// add the trailing slash if needed, as path.Join removes trailing slashes.
		if strings.HasSuffix(req.URL.Path, "/") && !strings.HasSuffix(redirectURL.Path, "/") {
			redirectURL.Path += "/"
		}
	}

	rw.Header().Set("Location", redirectURL.String())

	rw.WriteHeader(r.statusCode)
	if _, err := rw.Write([]byte(http.StatusText(r.statusCode))); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}
