package utils

import (
	"fmt"
	"net/http"
	"strings"
)

// SourceExtractor extracts the source from the request, e.g. that may be client ip, or particular header that
// identifies the source. amount stands for amount of connections the source consumes, usually 1 for connection limiters
// error should be returned when source can not be identified
type SourceExtractor interface {
	Extract(req *http.Request) (token string, amount int64, err error)
}

// ExtractorFunc extractor function type
type ExtractorFunc func(req *http.Request) (token string, amount int64, err error)

// Extract extract from request
func (f ExtractorFunc) Extract(req *http.Request) (string, int64, error) {
	return f(req)
}

// ExtractSource extract source function type
type ExtractSource func(req *http.Request)

// NewExtractor creates a new SourceExtractor
func NewExtractor(variable string) (SourceExtractor, error) {
	if variable == "client.ip" {
		return ExtractorFunc(extractClientIP), nil
	}
	if variable == "request.host" {
		return ExtractorFunc(extractHost), nil
	}
	if strings.HasPrefix(variable, "request.header.") {
		header := strings.TrimPrefix(variable, "request.header.")
		if len(header) == 0 {
			return nil, fmt.Errorf("wrong header: %s", header)
		}
		return makeHeaderExtractor(header), nil
	}
	return nil, fmt.Errorf("unsupported limiting variable: '%s'", variable)
}

func extractClientIP(req *http.Request) (string, int64, error) {
	vals := strings.SplitN(req.RemoteAddr, ":", 2)
	if len(vals[0]) == 0 {
		return "", 0, fmt.Errorf("failed to parse client IP: %v", req.RemoteAddr)
	}
	return vals[0], 1, nil
}

func extractHost(req *http.Request) (string, int64, error) {
	return req.Host, 1, nil
}

func makeHeaderExtractor(header string) SourceExtractor {
	return ExtractorFunc(func(req *http.Request) (string, int64, error) {
		return req.Header.Get(header), 1, nil
	})
}
