package blockpathregex

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/middlewares"
	"github.com/containous/traefik/pkg/tracing"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	typeName = "BlockPathRegex"
)

// BlockPathRegex is a middleware used to replace the path of a URL request with a regular expression.
type blockPathRegex struct {
	next         http.Handler
	regexp       *regexp.Regexp
	responseCode int
	message      string
	name         string
}

// New creates a new block path regex middleware.
func New(ctx context.Context, next http.Handler, config config.BlockPathRegex, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug("Creating middleware")

	exp, err := regexp.Compile(strings.TrimSpace(config.Regex))
	if err != nil {
		return nil, fmt.Errorf("error compiling regular expression %s: %s", config.Regex, err)
	}

	if config.ResponseCode < 0 || config.ResponseCode > 599 {
		return nil, fmt.Errorf("response code is out of bounds (0-599): %d", config.ResponseCode)
	}

	return &blockPathRegex{
		regexp:       exp,
		responseCode: config.ResponseCode,
		message:      config.Message,
		next:         next,
		name:         name,
	}, nil
}

func (b *blockPathRegex) GetTracingInformation() (string, ext.SpanKindEnum) {
	return b.name, tracing.SpanKindNoneEnum
}

func (b *blockPathRegex) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if b.regexp != nil && b.responseCode != 0 && b.regexp.MatchString(req.URL.Path) {
		rw.WriteHeader(b.responseCode)
		if len(b.message) > 0 {
			_, err := rw.Write([]byte(b.message))
			if err != nil {
				return
			}
		}
		return
	}
	b.next.ServeHTTP(rw, req)
}
