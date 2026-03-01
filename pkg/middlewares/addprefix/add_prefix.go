package addprefix

import (
	"context"
	"errors"
	"net/http"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
)

const (
	typeName = "AddPrefix"
)

// AddPrefix is a middleware used to add prefix to an URL request.
type addPrefix struct {
	next   http.Handler
	prefix string
	name   string
}

// New creates a new handler.
func New(ctx context.Context, next http.Handler, config dynamic.AddPrefix, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug().Msg("Creating middleware")
	var result *addPrefix

	if len(config.Prefix) > 0 {
		result = &addPrefix{
			prefix: config.Prefix,
			next:   next,
			name:   name,
		}
	} else {
		return nil, errors.New("prefix cannot be empty")
	}

	return result, nil
}

func (a *addPrefix) GetTracingInformation() (string, string) {
	return a.name, typeName
}

func (a *addPrefix) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := middlewares.GetLogger(req.Context(), a.name, typeName)

	oldURLPath := req.URL.Path
	req.URL.Path = ensureLeadingSlash(a.prefix + req.URL.Path)
	logger.Debug().Msgf("URL.Path is now %s (was %s).", req.URL.Path, oldURLPath)

	if req.URL.RawPath != "" {
		oldURLRawPath := req.URL.RawPath
		req.URL.RawPath = ensureLeadingSlash(a.prefix + req.URL.RawPath)
		logger.Debug().Msgf("URL.RawPath is now %s (was %s).", req.URL.RawPath, oldURLRawPath)
	}
	req.RequestURI = req.URL.RequestURI()

	a.next.ServeHTTP(rw, req)
}

func ensureLeadingSlash(str string) string {
	if str == "" {
		return str
	}

	if str[0] == '/' {
		return str
	}

	return "/" + str
}
