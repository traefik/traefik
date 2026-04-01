package stripprefixregex

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/middlewares/stripprefix"
)

const (
	typeName = "StripPrefixRegex"
)

// StripPrefixRegex is a middleware used to strip prefix from an URL request.
type stripPrefixRegex struct {
	next        http.Handler
	expressions []*regexp.Regexp
	name        string
}

// New builds a new StripPrefixRegex middleware.
func New(ctx context.Context, next http.Handler, config dynamic.StripPrefixRegex, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug().Msg("Creating middleware")

	stripPrefix := stripPrefixRegex{
		next: next,
		name: name,
	}

	for _, exp := range config.Regex {
		reg, err := regexp.Compile(strings.TrimSpace(exp))
		if err != nil {
			return nil, err
		}
		stripPrefix.expressions = append(stripPrefix.expressions, reg)
	}

	return &stripPrefix, nil
}

func (s *stripPrefixRegex) GetTracingInformation() (string, string) {
	return s.name, typeName
}

func (s *stripPrefixRegex) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	for _, exp := range s.expressions {
		parts := exp.FindStringSubmatch(req.URL.Path)
		if len(parts) > 0 && len(parts[0]) > 0 {
			prefix := parts[0]
			if !strings.HasPrefix(req.URL.Path, prefix) {
				continue
			}

			req.Header.Add(stripprefix.ForwardedPrefixHeader, prefix)

			req.URL.Path = ensureLeadingSlash(strings.Replace(req.URL.Path, prefix, "", 1))
			if req.URL.RawPath != "" {
				req.URL.RawPath = ensureLeadingSlash(req.URL.RawPath[encodedPrefixLen(req.URL.RawPath, prefix):])
			}

			req.RequestURI = req.URL.RequestURI()
			s.next.ServeHTTP(rw, req)
			return
		}
	}

	s.next.ServeHTTP(rw, req)
}

// encodedPrefixLen returns the number of bytes in rawPath that correspond to
// the decoded prefix, advancing 3 bytes per %XX sequence and 1 byte otherwise.
func encodedPrefixLen(rawPath, decodedPrefix string) int {
	decoded := 0
	i := 0
	for i < len(rawPath) && decoded < len(decodedPrefix) {
		if rawPath[i] == '%' && i+2 < len(rawPath) {
			i += 3
		} else {
			i++
		}
		decoded++
	}
	return i
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
