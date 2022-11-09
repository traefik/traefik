package headers

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
)

// Header is a middleware that helps setup a few basic security features.
// A single headerOptions struct can be provided to configure which features should be enabled,
// and the ability to override a few of the default values.
type Header struct {
	next                                http.Handler
	hasCustomRequestHeaders             bool
	hasCustomResponseHeaders            bool
	hasCorsHeaders                      bool
	hasNewStyleRequestHeaderMiddleware  bool
	hasNewStyleResponseHeaderMiddleware bool
	headers                             *dynamic.Headers
	allowOriginRegexes                  []*regexp.Regexp
}

// NewHeader constructs a new header instance from supplied frontend header struct.
func NewHeader(next http.Handler, cfg dynamic.Headers) (*Header, error) {

	hasOldStyleCustomRequestHeaders := cfg.HasOldStyleCustomRequestHeadersDefined()
	hasNewStyleCustomRequestHeaders := cfg.HasNewStyleCustomRequestHeadersDefined()

	hasOldStyleCustomResponseHeaders := cfg.HasOldStyleCustomResponseHeadersDefined()
	hasNewStyleCustomResponseHeaders := cfg.HasNewStyleCustomResponseHeadersDefined()
	hasCorsHeaders := cfg.HasCorsHeadersDefined()

	ctx := log.With(context.Background(), log.Str(log.MiddlewareType, typeName))
	handleDeprecation(ctx, &cfg)

	regexes := make([]*regexp.Regexp, len(cfg.AccessControlAllowOriginListRegex))
	for i, str := range cfg.AccessControlAllowOriginListRegex {
		reg, err := regexp.Compile(str)
		if err != nil {
			return nil, fmt.Errorf("error occurred during origin parsing: %w", err)
		}
		regexes[i] = reg
	}

	return &Header{
		next:                                next,
		headers:                             &cfg,
		hasCustomRequestHeaders:             hasOldStyleCustomRequestHeaders,
		hasNewStyleRequestHeaderMiddleware:  hasNewStyleCustomRequestHeaders,
		hasCustomResponseHeaders:            hasOldStyleCustomResponseHeaders,
		hasNewStyleResponseHeaderMiddleware: hasNewStyleCustomResponseHeaders,
		hasCorsHeaders:                      hasCorsHeaders,
		allowOriginRegexes:                  regexes,
	}, nil
}

func (s *Header) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Handle Cors headers and preflight if configured.
	if isPreflight := s.processCorsHeaders(rw, req); isPreflight {
		return
	}

	// Old style only used without new style header.
	if s.hasCustomRequestHeaders && !s.hasNewStyleRequestHeaderMiddleware {
		s.modifyCustomRequestHeaders(req)
	}

	if s.hasNewStyleRequestHeaderMiddleware {
		s.modifyNewStyleCustomRequestHeaders(req)
	}

	// If there is a next, call it.
	if s.next != nil {
		if s.hasNewStyleResponseHeaderMiddleware {
			s.next.ServeHTTP(newResponseModifier(rw, req, s.NewStylePostRequestModifyResponseHeaders), req)
		} else {
			s.next.ServeHTTP(newResponseModifier(rw, req, s.PostRequestModifyResponseHeaders), req)
		}

	}
}

// modifyCustomRequestHeaders sets or deletes custom request headers.
func (s *Header) modifyCustomRequestHeaders(req *http.Request) {
	// Loop through Custom request headers
	for header, value := range s.headers.CustomRequestHeaders {
		switch {
		case value == "":
			req.Header.Del(header)

		case strings.EqualFold(header, "Host"):
			req.Host = value

		default:
			req.Header.Set(header, value)
		}
	}
}

func (s *Header) modifyNewStyleCustomRequestHeaders(req *http.Request) {
	// Loop through Custom request headers
	for header, value := range s.headers.AppendRequestHeaders {
		if strings.EqualFold(header, "Host") {
			req.Host = value
		}
		req.Header.Add(header, value)
	}

	for header, value := range s.headers.ReplaceRequestHeaders {
		req.Header.Set(header, value)
	}

	// Support delete header key value instead of delete all.
	for header, value := range s.headers.DeleteRequestHeaders {
		values := req.Header.Values(header)
		req.Header.Del(header)
		if value != "" {
			for _, v := range values {
				if v != value {
					req.Header.Add(header, v)
				}
			}
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

// NewStylePostRequestModifyResponseHeaders set or delete response headers.
// This method is called AFTER the response is generated from the backend
// and can merge/override headers from the backend response.
func (s *Header) NewStylePostRequestModifyResponseHeaders(res *http.Response) error {
	// Loop through Custom response headers
	for header, value := range s.headers.AppendResponseHeaders {
		res.Header.Add(header, value)
	}

	for header, value := range s.headers.ReplaceResponseHeaders {
		res.Header.Set(header, value)
	}

	// Support delete header key value instead of delete all.
	for header, value := range s.headers.DeleteResponseHeaders {
		values := res.Header.Values(header)
		res.Header.Del(header)
		if value != "" {
			for _, v := range values {
				if v != value {
					res.Header.Add(header, v)
				}
			}
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

	for _, regex := range s.allowOriginRegexes {
		if regex.MatchString(origin) {
			return true, origin
		}
	}

	return false, ""
}
