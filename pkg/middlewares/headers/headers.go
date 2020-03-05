// Package headers Middleware based on https://github.com/unrolled/secure.
package headers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/middlewares"
	"github.com/containous/traefik/v2/pkg/tracing"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/unrolled/secure"
)

const (
	typeName = "Headers"
)

func handleDeprecation(ctx context.Context, cfg *dynamic.Headers) {
	if cfg.AccessControlAllowOrigin != "" {
		log.FromContext(ctx).Warn("accessControlAllowOrigin is deprecated, please use accessControlAllowOriginList instead.")
		cfg.AccessControlAllowOriginList = append(cfg.AccessControlAllowOriginList, cfg.AccessControlAllowOrigin)
		cfg.AccessControlAllowOrigin = ""
	}
}

type headers struct {
	name    string
	handler http.Handler
}

// New creates a Headers middleware.
func New(ctx context.Context, next http.Handler, cfg dynamic.Headers, name string) (http.Handler, error) {
	// HeaderMiddleware -> SecureMiddleWare -> next
	mCtx := middlewares.GetLoggerCtx(ctx, name, typeName)
	logger := log.FromContext(mCtx)
	logger.Debug("Creating middleware")

	handleDeprecation(mCtx, &cfg)

	hasSecureHeaders := cfg.HasSecureHeadersDefined()
	hasCustomHeaders := cfg.HasCustomHeadersDefined()
	hasCorsHeaders := cfg.HasCorsHeadersDefined()

	if !hasSecureHeaders && !hasCustomHeaders && !hasCorsHeaders {
		return nil, errors.New("headers configuration not valid")
	}

	var handler http.Handler
	nextHandler := next

	if hasSecureHeaders {
		logger.Debug("Setting up secureHeaders from %v", cfg)
		handler = newSecure(next, cfg)
		nextHandler = handler
	}

	if hasCustomHeaders || hasCorsHeaders {
		logger.Debug("Setting up customHeaders/Cors from %v", cfg)
		handler = NewHeader(nextHandler, cfg)
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
func newSecure(next http.Handler, cfg dynamic.Headers) *secureHeader {
	opt := secure.Options{
		BrowserXssFilter:        cfg.BrowserXSSFilter,
		ContentTypeNosniff:      cfg.ContentTypeNosniff,
		ForceSTSHeader:          cfg.ForceSTSHeader,
		FrameDeny:               cfg.FrameDeny,
		IsDevelopment:           cfg.IsDevelopment,
		SSLRedirect:             cfg.SSLRedirect,
		SSLForceHost:            cfg.SSLForceHost,
		SSLTemporaryRedirect:    cfg.SSLTemporaryRedirect,
		STSIncludeSubdomains:    cfg.STSIncludeSubdomains,
		STSPreload:              cfg.STSPreload,
		ContentSecurityPolicy:   cfg.ContentSecurityPolicy,
		CustomBrowserXssValue:   cfg.CustomBrowserXSSValue,
		CustomFrameOptionsValue: cfg.CustomFrameOptionsValue,
		PublicKey:               cfg.PublicKey,
		ReferrerPolicy:          cfg.ReferrerPolicy,
		SSLHost:                 cfg.SSLHost,
		AllowedHosts:            cfg.AllowedHosts,
		HostsProxyHeaders:       cfg.HostsProxyHeaders,
		SSLProxyHeaders:         cfg.SSLProxyHeaders,
		STSSeconds:              cfg.STSSeconds,
		FeaturePolicy:           cfg.FeaturePolicy,
	}

	return &secureHeader{
		next:   next,
		secure: secure.New(opt),
	}
}

func (s secureHeader) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	s.secure.HandlerFuncWithNextForRequestOnly(rw, req, s.next.ServeHTTP)
}

// Header is a middleware that helps setup a few basic security features.
// A single headerOptions struct can be provided to configure which features should be enabled,
// and the ability to override a few of the default values.
type Header struct {
	next             http.Handler
	hasCustomHeaders bool
	hasCorsHeaders   bool
	headers          *dynamic.Headers
}

// NewHeader constructs a new header instance from supplied frontend header struct.
func NewHeader(next http.Handler, cfg dynamic.Headers) *Header {
	hasCustomHeaders := cfg.HasCustomHeadersDefined()
	hasCorsHeaders := cfg.HasCorsHeadersDefined()

	ctx := log.With(context.Background(), log.Str(log.MiddlewareType, typeName))
	handleDeprecation(ctx, &cfg)

	return &Header{
		next:             next,
		headers:          &cfg,
		hasCustomHeaders: hasCustomHeaders,
		hasCorsHeaders:   hasCorsHeaders,
	}
}

func (s *Header) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Handle Cors headers and preflight if configured.
	if isPreflight := s.processCorsHeaders(rw, req); isPreflight {
		return
	}

	if s.hasCustomHeaders {
		s.modifyCustomRequestHeaders(req)
	}

	// If there is a next, call it.
	if s.next != nil {
		s.next.ServeHTTP(rw, req)
	}
}

// modifyCustomRequestHeaders sets or deletes custom request headers.
func (s *Header) modifyCustomRequestHeaders(req *http.Request) {
	// Loop through Custom request headers
	for header, value := range s.headers.CustomRequestHeaders {
		if value == "" {
			req.Header.Del(header)
		} else {
			req.Header.Set(header, value)
		}
	}
}

// PostRequestModifyResponseHeaders set or delete response headers.
// This method is called AFTER the response is generated from the backend
// and can merge/override headers from the backend response.
func (s *Header) PostRequestModifyResponseHeaders(res *http.Response) error {
	// Loop through Custom response headers
	for header, value := range s.headers.CustomResponseHeaders {
		if value == "" {
			res.Header.Del(header)
		} else {
			res.Header.Set(header, value)
		}
	}

	if res != nil && res.Request != nil {
		originHeader := res.Request.Header.Get("Origin")
		allowed, match := s.isOriginAllowed(originHeader)

		if allowed {
			res.Header.Set("Access-Control-Allow-Origin", match)
		}
	}

	if s.headers.AccessControlAllowCredentials {
		res.Header.Set("Access-Control-Allow-Credentials", "true")
	}

	if len(s.headers.AccessControlExposeHeaders) > 0 {
		exposeHeaders := strings.Join(s.headers.AccessControlExposeHeaders, ",")
		res.Header.Set("Access-Control-Expose-Headers", exposeHeaders)
	}

	if !s.headers.AddVaryHeader {
		return nil
	}

	varyHeader := res.Header.Get("Vary")
	if varyHeader == "Origin" {
		return nil
	}

	if varyHeader != "" {
		varyHeader += ","
	}
	varyHeader += "Origin"

	res.Header.Set("Vary", varyHeader)
	return nil
}

// processCorsHeaders processes the incoming request,
// and returns if it is a preflight request.
// If not a preflight, it handles the preRequestModifyCorsResponseHeaders.
func (s *Header) processCorsHeaders(rw http.ResponseWriter, req *http.Request) bool {
	if !s.hasCorsHeaders {
		return false
	}

	reqAcMethod := req.Header.Get("Access-Control-Request-Method")
	originHeader := req.Header.Get("Origin")

	if reqAcMethod != "" && originHeader != "" && req.Method == http.MethodOptions {
		// If the request is an OPTIONS request with an Access-Control-Request-Method header,
		// and Origin headers, then it is a CORS preflight request,
		// and we need to build a custom response: https://www.w3.org/TR/cors/#preflight-request
		if s.headers.AccessControlAllowCredentials {
			rw.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		allowHeaders := strings.Join(s.headers.AccessControlAllowHeaders, ",")
		if allowHeaders != "" {
			rw.Header().Set("Access-Control-Allow-Headers", allowHeaders)
		}

		allowMethods := strings.Join(s.headers.AccessControlAllowMethods, ",")
		if allowMethods != "" {
			rw.Header().Set("Access-Control-Allow-Methods", allowMethods)
		}

		allowed, match := s.isOriginAllowed(originHeader)
		if allowed {
			rw.Header().Set("Access-Control-Allow-Origin", match)
		}

		rw.Header().Set("Access-Control-Max-Age", strconv.Itoa(int(s.headers.AccessControlMaxAge)))
		return true
	}

	return false
}

func (s *Header) isOriginAllowed(origin string) (bool, string) {
	for _, item := range s.headers.AccessControlAllowOriginList {
		if item == "*" || item == origin {
			return true, item
		}
	}

	return false, ""
}
