// Package headers Middleware based on https://github.com/unrolled/secure.
package headers

import (
	"context"
	"errors"
	"net/http"

	"github.com/containous/traefik/config"
	"github.com/containous/traefik/middlewares"
	"github.com/containous/traefik/tracing"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/unrolled/secure"
)

const (
	typeName = "Headers"
)

type headers struct {
	name    string
	handler http.Handler
}

// New creates a Headers middleware.
func New(ctx context.Context, next http.Handler, config config.Headers, name string) (http.Handler, error) {
	// HeaderMiddleware -> SecureMiddleWare -> next
	logger := middlewares.GetLogger(ctx, name, typeName)
	logger.Debug("Creating middleware")

	if !config.HasSecureHeadersDefined() && !config.HasCustomHeadersDefined() {
		return nil, errors.New("headers configuration not valid")
	}

	var handler http.Handler
	nextHandler := next

	if config.HasSecureHeadersDefined() {
		logger.Debug("Setting up secureHeaders from %v", config)
		handler = newSecure(next, config)
		nextHandler = handler
	}

	if config.HasCustomHeadersDefined() {
		logger.Debug("Setting up customHeaders from %v", config)
		handler = newHeader(nextHandler, config)
	}

	return &headers{
		handler: handler,
		name:    name,
	}, nil
}

func (h *headers) GetTracingInformation() (string, ext.SpanKindEnum) {
	return h.name, tracing.SpanKindNoneEnum
}

func (h *headers) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	h.handler.ServeHTTP(rw, req)
}

type secureHeader struct {
	next   http.Handler
	secure *secure.Secure
}

// newSecure constructs a new secure instance with supplied options.
func newSecure(next http.Handler, headers config.Headers) *secureHeader {
	opt := secure.Options{
		BrowserXssFilter:        headers.BrowserXSSFilter,
		ContentTypeNosniff:      headers.ContentTypeNosniff,
		ForceSTSHeader:          headers.ForceSTSHeader,
		FrameDeny:               headers.FrameDeny,
		IsDevelopment:           headers.IsDevelopment,
		SSLRedirect:             headers.SSLRedirect,
		SSLForceHost:            headers.SSLForceHost,
		SSLTemporaryRedirect:    headers.SSLTemporaryRedirect,
		STSIncludeSubdomains:    headers.STSIncludeSubdomains,
		STSPreload:              headers.STSPreload,
		ContentSecurityPolicy:   headers.ContentSecurityPolicy,
		CustomBrowserXssValue:   headers.CustomBrowserXSSValue,
		CustomFrameOptionsValue: headers.CustomFrameOptionsValue,
		PublicKey:               headers.PublicKey,
		ReferrerPolicy:          headers.ReferrerPolicy,
		SSLHost:                 headers.SSLHost,
		AllowedHosts:            headers.AllowedHosts,
		HostsProxyHeaders:       headers.HostsProxyHeaders,
		SSLProxyHeaders:         headers.SSLProxyHeaders,
		STSSeconds:              headers.STSSeconds,
	}

	return &secureHeader{
		next:   next,
		secure: secure.New(opt),
	}
}

func (s secureHeader) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	s.secure.HandlerFuncWithNextForRequestOnly(rw, req, s.next.ServeHTTP)
}

// Header is a middleware that helps setup a few basic security features. A single headerOptions struct can be
// provided to configure which features should be enabled, and the ability to override a few of the default values.
type header struct {
	next http.Handler
	// If Custom request headers are set, these will be added to the request
	customRequestHeaders map[string]string
}

// NewHeader constructs a new header instance from supplied frontend header struct.
func newHeader(next http.Handler, headers config.Headers) *header {
	return &header{
		next:                 next,
		customRequestHeaders: headers.CustomRequestHeaders,
	}
}

func (s *header) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	s.modifyRequestHeaders(req)
	s.next.ServeHTTP(rw, req)
}

// modifyRequestHeaders set or delete request headers.
func (s *header) modifyRequestHeaders(req *http.Request) {
	// Loop through Custom request headers
	for header, value := range s.customRequestHeaders {
		if value == "" {
			req.Header.Del(header)
		} else {
			req.Header.Set(header, value)
		}
	}
}
