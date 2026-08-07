package headers

import (
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/vulcand/oxy/v2/forward"
)

// Header is a middleware that helps setup a few basic security features.
// A single headerOptions struct can be provided to configure which features should be enabled,
// and the ability to override a few of the default values.
type Header struct {
	next               http.Handler
	hasCustomHeaders   bool
	hasCorsHeaders     bool
	headers            *dynamic.Headers
	allowOriginRegexes []*regexp.Regexp
}

// NewHeader constructs a new header instance from supplied frontend header struct.
func NewHeader(next http.Handler, cfg dynamic.Headers) (*Header, error) {
	hasCustomHeaders := cfg.HasCustomHeadersDefined()
	hasCorsHeaders := cfg.HasCorsHeadersDefined()

	regexes := make([]*regexp.Regexp, len(cfg.AccessControlAllowOriginListRegex))
	for i, str := range cfg.AccessControlAllowOriginListRegex {
		reg, err := regexp.Compile(str)
		if err != nil {
			return nil, fmt.Errorf("error occurred during origin parsing: %w", err)
		}
		regexes[i] = reg
	}

	return &Header{
		next:               next,
		headers:            &cfg,
		hasCustomHeaders:   hasCustomHeaders,
		hasCorsHeaders:     hasCorsHeaders,
		allowOriginRegexes: regexes,
	}, nil
}

func (s *Header) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Handle Cors headers and preflight if configured.
	if isPreflight := s.processCorsHeaders(rw, req); isPreflight {
		rw.Header().Set("Content-Length", "0")
		rw.WriteHeader(http.StatusOK)
		return
	}

	if s.hasCustomHeaders {
		s.modifyCustomRequestHeaders(req)
	}

	// If there is a next, call it.
	if s.next != nil {
		s.next.ServeHTTP(middlewares.NewResponseModifier(rw, req, s.PostRequestModifyResponseHeaders), req)
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
		match, ok := s.matchOrigin(res.Request.Header.Get("Origin"))
		if ok {
			res.Header.Set("Access-Control-Allow-Origin", match)
		}
	}

	if s.headers.AccessControlAllowCredentials {
		res.Header.Set("Access-Control-Allow-Credentials", "true")
	}

	if len(s.headers.AccessControlExposeHeaders) > 0 {
		exposeHeaders := strings.Join(s.headers.AccessControlExposeHeaders, ",")
		if !(slices.Contains(s.headers.AccessControlExposeHeaders, "*") && s.headers.AccessControlAllowCredentials) {
			res.Header.Set("Access-Control-Expose-Headers", exposeHeaders)
		}
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

// modifyCustomRequestHeaders sets or deletes custom request headers.
func (s *Header) modifyCustomRequestHeaders(req *http.Request) {
	// Loop through Custom request headers
	for header, value := range s.headers.CustomRequestHeaders {
		switch {
		// Handling https://github.com/golang/go/commit/ecdbffd4ec68b509998792f120868fec319de59b.
		case value == "" && header == forward.XForwardedFor:
			req.Header[header] = nil

		case value == "":
			req.Header.Del(header)

		case strings.EqualFold(header, "Host"):
			req.Host = value

		default:
			req.Header.Set(header, value)
		}
	}
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
		if slices.Contains(s.headers.AccessControlAllowHeaders, "*") && s.headers.AccessControlAllowCredentials {
			if acrh := req.Header.Get("Access-Control-Request-Headers"); acrh != "" {
				rw.Header().Set("Access-Control-Allow-Headers", acrh)
			}
		} else if allowHeaders != "" {
			rw.Header().Set("Access-Control-Allow-Headers", allowHeaders)
		}

		allowMethods := strings.Join(s.headers.AccessControlAllowMethods, ",")
		if slices.Contains(s.headers.AccessControlAllowMethods, "*") && s.headers.AccessControlAllowCredentials {
			rw.Header().Set("Access-Control-Allow-Methods", reqAcMethod)
		} else if allowMethods != "" {
			rw.Header().Set("Access-Control-Allow-Methods", allowMethods)
		}

		match, ok := s.matchOrigin(originHeader)
		if ok {
			rw.Header().Set("Access-Control-Allow-Origin", match)
		}

		if s.headers.AccessControlMaxAge != nil {
			rw.Header().Set("Access-Control-Max-Age", strconv.FormatInt(*s.headers.AccessControlMaxAge, 10))
		}

		if len(s.headers.AccessControlExposeHeaders) > 0 {
			exposeHeaders := strings.Join(s.headers.AccessControlExposeHeaders, ",")
			if !(slices.Contains(s.headers.AccessControlExposeHeaders, "*") && s.headers.AccessControlAllowCredentials) {
				rw.Header().Set("Access-Control-Expose-Headers", exposeHeaders)
			}
		}

		return true
	}

	return false
}

func (s *Header) matchOrigin(origin string) (string, bool) {
	for _, item := range s.headers.AccessControlAllowOriginList {
		switch item {
		case origin:
			return item, true
		case "*":
			if s.headers.AccessControlAllowCredentials {
				return origin, true // Can't use wildcard with credentials.
			}
			return item, true
		}
	}

	for _, regex := range s.allowOriginRegexes {
		if regex.MatchString(origin) {
			return origin, true
		}
	}

	return "", false
}
