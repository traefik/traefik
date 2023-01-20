package headers

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
)

// Header is a middleware that helps setup a few basic security features.
// A single headerOptions struct can be provided to configure which features should be enabled,
// and the ability to override a few of the default values.
type Header struct {
	next                     http.Handler
	hasModifyRequestHeaders  bool
	hasModifyResponseHeaders bool
	hasCorsHeaders           bool
	headers                  *dynamic.Headers
	allowOriginRegexes       []*regexp.Regexp
}

// NewHeader constructs a new header instance from supplied frontend header struct.
func NewHeader(next http.Handler, cfg dynamic.Headers) (*Header, error) {

	securityHeaders := cfg.SecurityHeaders
	hasCorsHeaders := securityHeaders.HasCorsHeadersDefined()

	regexes := make([]*regexp.Regexp, len(securityHeaders.AccessControlAllowOriginListRegex))
	for i, str := range securityHeaders.AccessControlAllowOriginListRegex {
		reg, err := regexp.Compile(str)
		if err != nil {
			return nil, fmt.Errorf("error occurred during origin parsing: %w", err)
		}
		regexes[i] = reg
	}

	return &Header{
		next:                     next,
		headers:                  &cfg,
		hasModifyRequestHeaders:  cfg.RequestHeaders.IsDefined(),
		hasModifyResponseHeaders: cfg.ResponseHeaders.IsDefined(),
		hasCorsHeaders:           hasCorsHeaders,
		allowOriginRegexes:       regexes,
	}, nil
}

func (s *Header) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Handle Cors headers and preflight if configured.
	if isPreflight := s.processCorsHeaders(rw, req); isPreflight {
		return
	}

	if s.hasModifyRequestHeaders {
		s.modifyRequestHeaders(req)
	}

	// If there is a next, call it.
	if s.next != nil {
		s.next.ServeHTTP(newResponseModifier(rw, req, s.modifyResponseHeaders), req)
	}
}

// modifyRequestHeaders modify request headers.
func (s *Header) modifyRequestHeaders(req *http.Request) {
	rh := s.headers.RequestHeaders
	if !rh.IsDefined() {
		return
	}

	for key, value := range rh.Append {
		if strings.EqualFold(key, "Host") {
			req.Host = value
		} else {
			req.Header.Add(key, value)
		}

	}

	for key, value := range rh.Set {
		if strings.EqualFold(key, "Host") {
			req.Host = value
		} else {
			req.Header.Set(key, value)
		}
	}

	for _, key := range rh.Delete {
		req.Header.Del(key)
	}
}

// modifyResponseHeaders modify response headers.
// This method is called AFTER the response is generated from the backend
// and can merge/override headers from the backend response.
func (s *Header) modifyResponseHeaders(res *http.Response) error {
	rh := s.headers.ResponseHeaders

	for key, value := range rh.Append {
		res.Header.Add(key, value)
	}

	for key, value := range rh.Set {
		res.Header.Set(key, value)
	}

	for _, key := range rh.Delete {
		res.Header.Del(key)
	}

	if res != nil && res.Request != nil {
		originHeader := res.Request.Header.Get("Origin")
		allowed, match := s.isOriginAllowed(originHeader)

		if allowed {
			res.Header.Set("Access-Control-Allow-Origin", match)
		}
	}

	securityHeaders := s.headers.SecurityHeaders

	if securityHeaders.AccessControlAllowCredentials {
		res.Header.Set("Access-Control-Allow-Credentials", "true")
	}

	if len(securityHeaders.AccessControlExposeHeaders) > 0 {
		exposeHeaders := strings.Join(securityHeaders.AccessControlExposeHeaders, ",")
		res.Header.Set("Access-Control-Expose-Headers", exposeHeaders)
	}

	if !securityHeaders.AddVaryHeader {
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
		if s.headers.SecurityHeaders.AccessControlAllowCredentials {
			rw.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		allowHeaders := strings.Join(s.headers.SecurityHeaders.AccessControlAllowHeaders, ",")
		if allowHeaders != "" {
			rw.Header().Set("Access-Control-Allow-Headers", allowHeaders)
		}

		allowMethods := strings.Join(s.headers.SecurityHeaders.AccessControlAllowMethods, ",")
		if allowMethods != "" {
			rw.Header().Set("Access-Control-Allow-Methods", allowMethods)
		}

		allowed, match := s.isOriginAllowed(originHeader)
		if allowed {
			rw.Header().Set("Access-Control-Allow-Origin", match)
		}

		rw.Header().Set("Access-Control-Max-Age", strconv.Itoa(int(s.headers.SecurityHeaders.AccessControlMaxAge)))
		return true
	}

	return false
}

func (s *Header) isOriginAllowed(origin string) (bool, string) {
	for _, item := range s.headers.SecurityHeaders.AccessControlAllowOriginList {
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
