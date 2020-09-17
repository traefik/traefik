package addprefix

import (
	"context"
	"fmt"
	"net/http"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/middlewares"
	"github.com/traefik/traefik/v2/pkg/tracing"
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
	log.FromContext(middlewares.GetLoggerCtx(ctx, name, typeName)).Debug("Creating middleware")
	var result *addPrefix

	if len(config.Prefix) > 0 {
		result = &addPrefix{
			prefix: config.Prefix,
			next:   next,
			name:   name,
		}
	} else {
		return nil, fmt.Errorf("prefix cannot be empty")
	}

	return result, nil
}

func (a *addPrefix) GetTracingInformation() (string, ext.SpanKindEnum) {
	return a.name, tracing.SpanKindNoneEnum
}

func (a *addPrefix) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := log.FromContext(middlewares.GetLoggerCtx(req.Context(), a.name, typeName))

	oldURLPath := req.URL.Path
	req.URL.Path = ensureLeadingSlash(a.prefix + req.URL.Path)
	logger.Debugf("URL.Path is now %s (was %s).", req.URL.Path, oldURLPath)

	if req.URL.RawPath != "" {
		oldURLRawPath := req.URL.RawPath
		req.URL.RawPath = ensureLeadingSlash(a.prefix + req.URL.RawPath)
		logger.Debugf("URL.RawPath is now %s (was %s).", req.URL.RawPath, oldURLRawPath)
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
