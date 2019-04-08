package addprefix

import (
	"context"
	"fmt"
	"net/http"

	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/middlewares"
	"github.com/containous/traefik/pkg/tracing"
	"github.com/opentracing/opentracing-go/ext"
)

type key string

const (
	typeName = "AddPrefix"
	// AddPrefixKey is the context key for storing the added prefix
	AddPrefixKey key = "AddPrefix"
)

// AddPrefix is a middleware used to add prefix to an URL request.
type addPrefix struct {
	next   http.Handler
	prefix string
	name   string
}

// New creates a new handler.
func New(ctx context.Context, next http.Handler, config config.AddPrefix, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug("Creating middleware")
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

func (ap *addPrefix) GetTracingInformation() (string, ext.SpanKindEnum) {
	return ap.name, tracing.SpanKindNoneEnum
}

func (ap *addPrefix) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := middlewares.GetLogger(req.Context(), ap.name, typeName)

	oldURLPath := req.URL.Path
	req.URL.Path = ap.prefix + req.URL.Path
	logger.Debugf("URL.Path is now %s (was %s).", req.URL.Path, oldURLPath)

	if req.URL.RawPath != "" {
		oldURLRawPath := req.URL.RawPath
		req.URL.RawPath = ap.prefix + req.URL.RawPath
		logger.Debugf("URL.RawPath is now %s (was %s).", req.URL.RawPath, oldURLRawPath)
	}
	req.RequestURI = req.URL.RequestURI()
	req = req.WithContext(context.WithValue(req.Context(), AddPrefixKey, ap.prefix))

	ap.next.ServeHTTP(rw, req)
}
