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
		rw.WriteHeader(http.StatusOK)
		return
	}

	if s.hasCustomHeaders {
		s.modifyCustomRequestHeaders(req)
	}

	// If there is a next, call it.
	if s.next != nil {
		s.next.ServeHTTP(newResponseModifier(rw, req, s.PostRequestModifyResponseHeaders), req)
	}
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

	for _, regex := range s.allowOriginRegexes {
		if regex.MatchString(origin) {
			return true, origin
		}
	}

	return false, ""
}
