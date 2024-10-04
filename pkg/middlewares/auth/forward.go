package auth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/middlewares/accesslog"
	"github.com/traefik/traefik/v3/pkg/middlewares/observability"
	"github.com/traefik/traefik/v3/pkg/tracing"
	"github.com/traefik/traefik/v3/pkg/types"
	"github.com/vulcand/oxy/v2/forward"
	"github.com/vulcand/oxy/v2/utils"
	"go.opentelemetry.io/otel/trace"
)

const (
	xForwardedURI    = "X-Forwarded-Uri"
	xForwardedMethod = "X-Forwarded-Method"
)

const typeNameForward = "ForwardAuth"

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
	trustForwardHeader       bool
	authRequestHeaders       []string
	addAuthCookiesToResponse map[string]struct{}
	headerField              string
}

// NewForward creates a forward auth middleware.
func NewForward(ctx context.Context, next http.Handler, config dynamic.ForwardAuth, name string) (http.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, typeNameForward)
	logger.Debug().Msg("Creating middleware")

	addAuthCookiesToResponse := make(map[string]struct{})
	for _, cookieName := range config.AddAuthCookiesToResponse {
		addAuthCookiesToResponse[cookieName] = struct{}{}
	}

	fa := &forwardAuth{
		address:                  config.Address,
		authResponseHeaders:      config.AuthResponseHeaders,
		next:                     next,
		name:                     name,
		trustForwardHeader:       config.TrustForwardHeader,
		authRequestHeaders:       config.AuthRequestHeaders,
		addAuthCookiesToResponse: addAuthCookiesToResponse,
		headerField:              config.HeaderField,
	}

	// Ensure our request client does not follow redirects
	fa.client = http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 30 * time.Second,
	}

	if config.TLS != nil {
		if config.TLS.CAOptional != nil {
			logger.Warn().Msg("CAOptional option is deprecated, TLS client authentication is a server side option, please remove any usage of this option.")
		}

		clientTLS := &types.ClientTLS{
			CA:                 config.TLS.CA,
			Cert:               config.TLS.Cert,
			Key:                config.TLS.Key,
			InsecureSkipVerify: config.TLS.InsecureSkipVerify,
		}

		tlsConfig, err := clientTLS.CreateTLSConfig(ctx)
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

	return fa, nil
}

func (fa *forwardAuth) GetTracingInformation() (string, string, trace.SpanKind) {
	return fa.name, typeNameForward, trace.SpanKindInternal
}

func (fa *forwardAuth) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := middlewares.GetLogger(req.Context(), fa.name, typeNameForward)

	forwardReq, err := http.NewRequestWithContext(req.Context(), http.MethodGet, fa.address, nil)
	if err != nil {
		logger.Debug().Msgf("Error calling %s. Cause %s", fa.address, err)
		observability.SetStatusErrorf(req.Context(), "Error calling %s. Cause %s", fa.address, err)

		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeHeader(req, forwardReq, fa.trustForwardHeader, fa.authRequestHeaders)

	var forwardSpan trace.Span
	var tracer *tracing.Tracer
	if tracer = tracing.TracerFromContext(req.Context()); tracer != nil {
		var tracingCtx context.Context
		tracingCtx, forwardSpan = tracer.Start(req.Context(), "AuthRequest", trace.WithSpanKind(trace.SpanKindClient))
		defer forwardSpan.End()

		forwardReq = forwardReq.WithContext(tracingCtx)

		tracing.InjectContextIntoCarrier(forwardReq)
		tracer.CaptureClientRequest(forwardSpan, forwardReq)
	}

	forwardResponse, forwardErr := fa.client.Do(forwardReq)
	if forwardErr != nil {
		logger.Debug().Msgf("Error calling %s. Cause: %s", fa.address, forwardErr)
		observability.SetStatusErrorf(req.Context(), "Error calling %s. Cause: %s", fa.address, forwardErr)

		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer forwardResponse.Body.Close()

	body, readError := io.ReadAll(forwardResponse.Body)
	if readError != nil {
		logger.Debug().Msgf("Error reading body %s. Cause: %s", fa.address, readError)
		observability.SetStatusErrorf(req.Context(), "Error reading body %s. Cause: %s", fa.address, readError)

		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Ending the forward request span as soon as the response is handled.
	// If any errors happen earlier, this span will be close by the defer instruction.
	if forwardSpan != nil {
		forwardSpan.End()
	}

	if fa.headerField != "" {
		if elems := forwardResponse.Header[http.CanonicalHeaderKey(fa.headerField)]; len(elems) > 0 {
			logData := accesslog.GetLogData(req)
			if logData != nil {
				logData.Core[accesslog.ClientUsername] = elems[0]
			}
		}
	}

	// Pass the forward response's body and selected headers if it
	// didn't return a response within the range of [200, 300).
	if forwardResponse.StatusCode < http.StatusOK || forwardResponse.StatusCode >= http.StatusMultipleChoices {
		logger.Debug().Msgf("Remote error %s. StatusCode: %d", fa.address, forwardResponse.StatusCode)

		utils.CopyHeaders(rw.Header(), forwardResponse.Header)
		utils.RemoveHeaders(rw.Header(), hopHeaders...)

		// Grab the location header, if any.
		redirectURL, err := forwardResponse.Location()

		if err != nil {
			if !errors.Is(err, http.ErrNoLocation) {
				logger.Debug().Msgf("Error reading response location header %s. Cause: %s", fa.address, err)
				observability.SetStatusErrorf(req.Context(), "Error reading response location header %s. Cause: %s", fa.address, err)

				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else if redirectURL.String() != "" {
			// Set the location in our response if one was sent back.
			rw.Header().Set("Location", redirectURL.String())
		}

		tracer.CaptureResponse(forwardSpan, forwardResponse.Header, forwardResponse.StatusCode, trace.SpanKindClient)
		rw.WriteHeader(forwardResponse.StatusCode)

		if _, err = rw.Write(body); err != nil {
			logger.Error().Err(err).Send()
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

	tracer.CaptureResponse(forwardSpan, forwardResponse.Header, forwardResponse.StatusCode, trace.SpanKindClient)

	req.RequestURI = req.URL.RequestURI()

	authCookies := forwardResponse.Cookies()
	if len(authCookies) == 0 {
		fa.next.ServeHTTP(rw, req)
		return
	}

	fa.next.ServeHTTP(middlewares.NewResponseModifier(rw, req, fa.buildModifier(authCookies)), req)
}

func (fa *forwardAuth) buildModifier(authCookies []*http.Cookie) func(res *http.Response) error {
	return func(res *http.Response) error {
		cookies := res.Cookies()
		res.Header.Del("Set-Cookie")

		for _, cookie := range cookies {
			if _, found := fa.addAuthCookiesToResponse[cookie.Name]; !found {
				res.Header.Add("Set-Cookie", cookie.String())
			}
		}

		for _, cookie := range authCookies {
			if _, found := fa.addAuthCookiesToResponse[cookie.Name]; found {
				res.Header.Add("Set-Cookie", cookie.String())
			}
		}

		return nil
	}
}

func writeHeader(req, forwardReq *http.Request, trustForwardHeader bool, allowedHeaders []string) {
	utils.CopyHeaders(forwardReq.Header, req.Header)

	RemoveConnectionHeaders(forwardReq)
	utils.RemoveHeaders(forwardReq.Header, hopHeaders...)

	forwardReq.Header = filterForwardRequestHeaders(forwardReq.Header, allowedHeaders)

	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		if trustForwardHeader {
			if prior, ok := req.Header[forward.XForwardedFor]; ok {
				clientIP = strings.Join(prior, ", ") + ", " + clientIP
			}
		}
		forwardReq.Header.Set(forward.XForwardedFor, clientIP)
	}

	xMethod := req.Header.Get(xForwardedMethod)
	switch {
	case xMethod != "" && trustForwardHeader:
		forwardReq.Header.Set(xForwardedMethod, xMethod)
	case req.Method != "":
		forwardReq.Header.Set(xForwardedMethod, req.Method)
	default:
		forwardReq.Header.Del(xForwardedMethod)
	}

	xfp := req.Header.Get(forward.XForwardedProto)
	switch {
	case xfp != "" && trustForwardHeader:
		forwardReq.Header.Set(forward.XForwardedProto, xfp)
	case req.TLS != nil:
		forwardReq.Header.Set(forward.XForwardedProto, "https")
	default:
		forwardReq.Header.Set(forward.XForwardedProto, "http")
	}

	if xfp := req.Header.Get(forward.XForwardedPort); xfp != "" && trustForwardHeader {
		forwardReq.Header.Set(forward.XForwardedPort, xfp)
	}

	xfh := req.Header.Get(forward.XForwardedHost)
	switch {
	case xfh != "" && trustForwardHeader:
		forwardReq.Header.Set(forward.XForwardedHost, xfh)
	case req.Host != "":
		forwardReq.Header.Set(forward.XForwardedHost, req.Host)
	default:
		forwardReq.Header.Del(forward.XForwardedHost)
	}

	xfURI := req.Header.Get(xForwardedURI)
	switch {
	case xfURI != "" && trustForwardHeader:
		forwardReq.Header.Set(xForwardedURI, xfURI)
	case req.URL.RequestURI() != "":
		forwardReq.Header.Set(xForwardedURI, req.URL.RequestURI())
	default:
		forwardReq.Header.Del(xForwardedURI)
	}
}

func filterForwardRequestHeaders(forwardRequestHeaders http.Header, allowedHeaders []string) http.Header {
	if len(allowedHeaders) == 0 {
		return forwardRequestHeaders
	}

	filteredHeaders := http.Header{}
	for _, headerName := range allowedHeaders {
		values := forwardRequestHeaders.Values(headerName)
		if len(values) > 0 {
			filteredHeaders[http.CanonicalHeaderKey(headerName)] = append([]string(nil), values...)
		}
	}

	return filteredHeaders
}
