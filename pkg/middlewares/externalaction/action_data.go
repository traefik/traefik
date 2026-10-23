package externalaction

import (
	"context"
	"maps"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type actionData struct {
	Context              context.Context
	Request              actionRequest
	Body                 any
	RawBody              string
	Response             actionResponse
	Item                 any
	ItemIndex            int
	AuthParameters       map[string]string
	LifecycleBinding     actionLifecycleBinding
	Payload              any
	Input                any
	Result               any
	Enrichment           map[string]any
	SourceAuthParameters map[string]string
	LookupTime           time.Time
	OutboundBody         string
	lookupPageQuery      url.Values
	deferredEnrichments  []*compiledEnrichment

	credentialSet              string
	expectedCredentialIdentity [32]byte
}

// actionLifecycleBinding exposes reconciliation state to action templates.
// Created is true only when resolveOrCreate recovered a missing lifecycle
// binding during the current dispatch.
type actionLifecycleBinding struct {
	Created bool
}

type actionRequest struct {
	Method   string
	Scheme   string
	Host     string
	Path     string
	RawQuery string
	Headers  http.Header
	Query    url.Values
	Cookies  map[string]string
}

type actionResponse struct {
	Status      int
	Headers     http.Header
	Bytes       int64
	Body        any
	RawBody     string
	Overflow    bool
	StartedAt   time.Time
	CompletedAt time.Time
	DurationMS  int64
}

func newActionData(req *http.Request, raw []byte, status int, responseHeaders http.Header, started, completed time.Time) *actionData {
	cookies := make(map[string]string)
	for _, cookie := range req.Cookies() {
		cookies[cookie.Name] = cookie.Value
	}
	data := &actionData{
		Context:             req.Context(),
		deferredEnrichments: enrichmentsFromContext(req.Context()),
		Request: actionRequest{
			Method:   req.Method,
			Scheme:   scheme(req),
			Host:     req.Host,
			Path:     req.URL.Path,
			RawQuery: req.URL.RawQuery,
			Headers:  req.Header.Clone(),
			Query:    req.URL.Query(),
			Cookies:  cookies,
		},
		RawBody: string(raw),
		Response: actionResponse{
			Status:      status,
			Headers:     responseHeaders.Clone(),
			StartedAt:   started,
			CompletedAt: completed,
		},
	}
	if !completed.IsZero() {
		data.Response.DurationMS = completed.Sub(started).Milliseconds()
	}
	data.Body = parseBody("", req.Header.Get("Content-Type"), raw)
	return data
}

func (d *actionData) clone() *actionData {
	result := *d
	result.Context = context.WithoutCancel(d.Context)
	result.Request.Headers = d.Request.Headers.Clone()
	result.Request.Query = cloneValues(d.Request.Query)
	result.Request.Cookies = cloneMap(d.Request.Cookies)
	result.Response.Headers = d.Response.Headers.Clone()
	result.AuthParameters = cloneMap(d.AuthParameters)
	return &result
}

func (d *actionData) withBodyFormat(format string) *actionData {
	result := *d
	result.Body = parseBody(format, d.Request.Headers.Get("Content-Type"), []byte(d.RawBody))
	return &result
}

func (d *actionData) withFormats(bodyFormat, responseFormat string) *actionData {
	result := d.withBodyFormat(bodyFormat)
	result.Response.Body = parseBody(responseFormat, result.Response.Headers.Get("Content-Type"), []byte(result.Response.RawBody))
	return result
}

func parseBody(format, contentType string, raw []byte) any {
	if len(raw) == 0 {
		return nil
	}
	format = strings.ToLower(format)
	switch {
	case format == "json" || format == "" && strings.Contains(contentType, "json"):
		if result, err := decodeJSON(raw); err == nil {
			return result
		}
	case format == "form" || format == "" && strings.Contains(contentType, "application/x-www-form-urlencoded"):
		if result, err := url.ParseQuery(string(raw)); err == nil {
			return result
		}
	}
	return string(raw)
}

func scheme(req *http.Request) string {
	if req.TLS != nil {
		return "https"
	}
	return schemeHTTP
}

func cloneValues(source url.Values) url.Values {
	result := make(url.Values, len(source))
	for key, value := range source {
		result[key] = append([]string(nil), value...)
	}
	return result
}

func cloneMap(source map[string]string) map[string]string {
	result := make(map[string]string, len(source))
	maps.Copy(result, source)
	return result
}
