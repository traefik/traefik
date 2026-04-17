package auth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"net"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/middlewares"
	"github.com/traefik/traefik/v2/pkg/middlewares/forwardedheaders"
	"github.com/traefik/traefik/v2/pkg/tracing"
	"github.com/vulcand/oxy/v2/forward"
	"github.com/vulcand/oxy/v2/utils"
)

const forwardedTypeName = "ForwardedAuthType"

var errResponseBodyTooLarge = errors.New("response body too large")

// hopHeaders Hop-by-hop headers to be removed in the authentication request.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
// Proxy-Authorization header is forwarded to the authentication server (see https://tools.ietf.org/html/rfc7235#section-4.4).
var hopHeaders = []string{
	forward.Connection,
	forward.KeepAlive,
	forward.Te, // canonicalized version of "TE"
	forward.Trailers,
	forward.TransferEncoding,
	forward.Upgrade,
}

type forwardAuth struct {
	address                  string
	authResponseHeaders      []string
	authResponseHeadersRegex *regexp.Regexp
	next                     http.Handler
	name                     string
	client                   http.Client
	trustForwardHeader       *bool
	authRequestHeaders       []string
	maxResponseBodySize      int64
}

// NewForward creates a forward auth middleware.
func NewForward(ctx context.Context, next http.Handler, config dynamic.ForwardAuth, name string) (http.Handler, error) {
	logger := log.FromContext(middlewares.GetLoggerCtx(ctx, name, forwardedTypeName))
	logger.Debug("Creating middleware")

	fa := &forwardAuth{
		address:             config.Address,
		authResponseHeaders: config.AuthResponseHeaders,
		next:                next,
		name:                name,
		trustForwardHeader:  config.TrustForwardHeader,
		authRequestHeaders:  config.AuthRequestHeaders,
	}

	if config.MaxResponseBodySize != nil {
		fa.maxResponseBodySize = *config.MaxResponseBodySize
	} else {
		fa.maxResponseBodySize = -1
		logger.Warn("ForwardAuth 'maxResponseBodySize' is not configured, allowing unlimited response body size which can lead to DoS attacks and memory exhaustion. Please set an appropriate limit.")
	}

	// Ensure our request client does not follow redirects
	fa.client = http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 30 * time.Second,
	}

	if config.TLS != nil {
		tlsConfig, err := config.TLS.CreateTLSConfig(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to create client TLS configuration: %w", err)
		}

		tr := http.DefaultTransport.(*http.Transport).Clone()
		tr.TLSClientConfig = tlsConfig
		fa.client.Transport = tr
	}

	if config.AuthResponseHeadersRegex != "" {
		re, err := regexp.Compile(config.AuthResponseHeadersRegex)
		if err != nil {
			return nil, fmt.Errorf("error compiling regular expression %s: %w", config.AuthResponseHeadersRegex, err)
		}
		fa.authResponseHeadersRegex = re
	}

	if config.TrustForwardHeader == nil {
		logger.Warn("TrustForwardHeader is not set: this creates an inconsistent security behavior where some X-Forwarded headers (e.g. X-Forwarded-For, X-Forwarded-Proto) are removed but others (e.g. X-Forwarded-Prefix) are forwarded untouched. Set it to false to remove all X-Forwarded headers, or true to trust them all.")
	} else if *config.TrustForwardHeader && len(fa.authRequestHeaders) > 0 {
		fa.authRequestHeaders = append(fa.authRequestHeaders, slices.Collect(maps.Keys(forwardedheaders.XHeadersSet))...)
	}

	return fa, nil
}

func (fa *forwardAuth) GetTracingInformation() (string, ext.SpanKindEnum) {
	return fa.name, ext.SpanKindRPCClientEnum
}

func (fa *forwardAuth) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := log.FromContext(middlewares.GetLoggerCtx(req.Context(), fa.name, forwardedTypeName))

	forwardReq, err := http.NewRequest(http.MethodGet, fa.address, nil)
	tracing.LogRequest(tracing.GetSpan(req), forwardReq)
	if err != nil {
		logger.Debugf("Error calling %s. Cause %s", fa.address, err)
		tracing.SetErrorWithEvent(req, "Error calling %s. Cause %s", fa.address, err)

		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Ensure tracing headers are in the request before we copy the headers to the
	// forwardReq.
	tracing.InjectRequestHeaders(req)

	if fa.trustForwardHeader != nil {
		writeHeader(req, forwardReq, *fa.trustForwardHeader, fa.authRequestHeaders)
	} else {
		oldWriteHeader(req, forwardReq, fa.authRequestHeaders)
	}

	forwardResponse, forwardErr := fa.client.Do(forwardReq)
	if forwardErr != nil {
		logger.Debugf("Error calling %s. Cause: %s", fa.address, forwardErr)
		tracing.SetErrorWithEvent(req, "Error calling %s. Cause: %s", fa.address, forwardErr)

		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer forwardResponse.Body.Close()

	body, readError := fa.readResponseBodyBytes(forwardResponse)
	if readError != nil {
		if errors.Is(readError, errResponseBodyTooLarge) {
			logger.Debugf("Response body is too large, maxResponseBodySize: %d", fa.maxResponseBodySize)

			tracing.SetErrorWithEvent(req, "Response body is too large, maxResponseBodySize: %d", fa.maxResponseBodySize)
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}
		logger.Debugf("Error reading body %s", fa.address)
		tracing.SetErrorWithEvent(req, "Error reading body %s. Cause: %s", fa.address, readError)

		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Pass the forward response's body and selected headers if it
	// didn't return a response within the range of [200, 300).
	if forwardResponse.StatusCode < http.StatusOK || forwardResponse.StatusCode >= http.StatusMultipleChoices {
		logger.Debugf("Remote error %s. StatusCode: %d", fa.address, forwardResponse.StatusCode)

		utils.CopyHeaders(rw.Header(), forwardResponse.Header)
		utils.RemoveHeaders(rw.Header(), hopHeaders...)

		// Grab the location header, if any.
		redirectURL, err := forwardResponse.Location()

		if err != nil {
			if !errors.Is(err, http.ErrNoLocation) {
				logger.Debugf("Error reading response location header %s. Cause: %s", fa.address, err)
				tracing.SetErrorWithEvent(req, "Error reading response location header %s. Cause: %s", fa.address, err)

				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else if redirectURL.String() != "" {
			// Set the location in our response if one was sent back.
			rw.Header().Set("Location", redirectURL.String())
		}

		tracing.LogResponseCode(tracing.GetSpan(req), forwardResponse.StatusCode)
		rw.WriteHeader(forwardResponse.StatusCode)

		if _, err = rw.Write(body); err != nil {
			logger.Error(err)
		}
		return
	}

	for _, headerName := range fa.authResponseHeaders {
		headerKey := http.CanonicalHeaderKey(headerName)
		req.Header.Del(headerKey)
		if len(forwardResponse.Header[headerKey]) > 0 {
			req.Header[headerKey] = append([]string(nil), forwardResponse.Header[headerKey]...)
		}
	}

	if fa.authResponseHeadersRegex != nil {
		for headerKey := range req.Header {
			if fa.authResponseHeadersRegex.MatchString(headerKey) {
				req.Header.Del(headerKey)
			}
		}

		for headerKey, headerValues := range forwardResponse.Header {
			if fa.authResponseHeadersRegex.MatchString(headerKey) {
				req.Header[headerKey] = append([]string(nil), headerValues...)
			}
		}
	}

	req.RequestURI = req.URL.RequestURI()
	fa.next.ServeHTTP(rw, req)
}

func (fa *forwardAuth) readResponseBodyBytes(res *http.Response) ([]byte, error) {
	if fa.maxResponseBodySize < 0 {
		return io.ReadAll(res.Body)
	}

	body := make([]byte, fa.maxResponseBodySize+1)
	n, err := io.ReadFull(res.Body, body)
	if errors.Is(err, io.EOF) {
		return nil, nil
	}
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return nil, fmt.Errorf("reading response body bytes: %w", err)
	}
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return body[:n], nil
	}
	return nil, errResponseBodyTooLarge
}

func writeHeader(req, forwardReq *http.Request, trustForwardHeader bool, allowedHeaders []string) {
	utils.CopyHeaders(forwardReq.Header, req.Header)

	RemoveConnectionHeaders(forwardReq)
	utils.RemoveHeaders(forwardReq.Header, hopHeaders...)

	if !trustForwardHeader {
		forwardedheaders.DeleteXForwardedHeaders(forwardReq.Header)
	}

	forwardReq.Header = filterForwardRequestHeaders(forwardReq.Header, allowedHeaders)

	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		if prior, ok := forwardReq.Header[forwardedheaders.XForwardedFor]; ok {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		forwardReq.Header.Set(forwardedheaders.XForwardedFor, clientIP)
	}

	if _, ok := forwardReq.Header[forwardedheaders.XForwardedMethod]; !ok {
		forwardReq.Header.Set(forwardedheaders.XForwardedMethod, req.Method)
	}

	if _, ok := forwardReq.Header[forwardedheaders.XForwardedProto]; !ok {
		forwardReq.Header.Set(forwardedheaders.XForwardedProto, "http")
		if req.TLS != nil {
			forwardReq.Header.Set(forwardedheaders.XForwardedProto, "https")
		}
	}

	if _, ok := forwardReq.Header[forwardedheaders.XForwardedPort]; !ok {
		forwardReq.Header.Set(forwardedheaders.XForwardedPort, forwardedPort(req))
	}

	if _, ok := forwardReq.Header[forwardedheaders.XForwardedHost]; !ok {
		forwardReq.Header.Set(forwardedheaders.XForwardedHost, req.Host)
	}

	if _, ok := forwardReq.Header[forwardedheaders.XForwardedURI]; !ok {
		forwardReq.Header.Set(forwardedheaders.XForwardedURI, req.URL.RequestURI())
	}
}

// oldWriteHeader is the legacy implementation of writeHeader, which is used when TrustForwardHeader is not set (old false behavior).
// It is kept to avoid breaking existing configurations that rely on the previous behavior.
func oldWriteHeader(req, forwardReq *http.Request, allowedHeaders []string) {
	utils.CopyHeaders(forwardReq.Header, req.Header)

	RemoveConnectionHeaders(forwardReq)
	utils.RemoveHeaders(forwardReq.Header, hopHeaders...)

	forwardReq.Header = filterForwardRequestHeaders(forwardReq.Header, allowedHeaders)

	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		forwardReq.Header.Set(forwardedheaders.XForwardedFor, clientIP)
	}

	proto := "http"
	if req.TLS != nil {
		proto = "https"
	}
	forwardReq.Header.Set(forwardedheaders.XForwardedProto, proto)

	forwardReq.Header.Set(forwardedheaders.XForwardedMethod, req.Method)
	forwardReq.Header.Set(forwardedheaders.XForwardedHost, req.Host)
	forwardReq.Header.Set(forwardedheaders.XForwardedURI, req.URL.RequestURI())
}

func filterForwardRequestHeaders(forwardRequestHeaders http.Header, allowedHeaders []string) http.Header {
	if len(allowedHeaders) == 0 {
		return forwardRequestHeaders
	}

	filteredHeaders := http.Header{}
	for _, headerName := range allowedHeaders {
		if values := forwardRequestHeaders.Values(headerName); len(values) > 0 {
			filteredHeaders[http.CanonicalHeaderKey(headerName)] = values
		}
	}

	return filteredHeaders
}

func forwardedPort(req *http.Request) string {
	if req == nil {
		return ""
	}

	if _, port, err := net.SplitHostPort(req.Host); err == nil && port != "" {
		return port
	}

	if req.Header.Get(forwardedheaders.XForwardedProto) == "https" || req.Header.Get(forwardedheaders.XForwardedProto) == "wss" {
		return "443"
	}

	if req.TLS != nil {
		return "443"
	}

	return "80"
}
