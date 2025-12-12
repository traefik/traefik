package httputil

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	stdlog "log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/observability/logs"
	"golang.org/x/net/http/httpguts"
)

type key string

const (
	// StatusClientClosedRequest non-standard HTTP status code for client disconnection.
	StatusClientClosedRequest = 499

	// StatusClientClosedRequestText non-standard HTTP status for client disconnection.
	StatusClientClosedRequestText = "Client Closed Request"

	notAppendXFFKey key = "NotAppendXFF"
)

// SetNotAppendXFF indicates xff should not be appended.
func SetNotAppendXFF(ctx context.Context) context.Context {
	return context.WithValue(ctx, notAppendXFFKey, true)
}

// ShouldNotAppendXFF returns whether X-Forwarded-For should not be appended.
func ShouldNotAppendXFF(ctx context.Context) bool {
	val := ctx.Value(notAppendXFFKey)
	if val == nil {
		return false
	}

	notAppendXFF, ok := val.(bool)
	if !ok {
		return false
	}

	return notAppendXFF
}

func buildSingleHostProxy(target *url.URL, passHostHeader bool, preservePath bool, flushInterval time.Duration, roundTripper http.RoundTripper, bufferPool httputil.BufferPool) http.Handler {
	return &httputil.ReverseProxy{
		Rewrite:       rewriteRequestBuilder(target, passHostHeader, preservePath),
		Transport:     roundTripper,
		FlushInterval: flushInterval,
		BufferPool:    bufferPool,
		ErrorLog:      stdlog.New(logs.NoLevel(log.Logger, zerolog.DebugLevel), "", 0),
		ErrorHandler:  ErrorHandler,
	}
}

func rewriteRequestBuilder(target *url.URL, passHostHeader bool, preservePath bool) func(*httputil.ProxyRequest) {
	return func(pr *httputil.ProxyRequest) {
		copyForwardedHeader(pr.Out.Header, pr.In.Header)
		if !ShouldNotAppendXFF(pr.In.Context()) {
			if clientIP, _, err := net.SplitHostPort(pr.In.RemoteAddr); err == nil {
				// If we aren't the first proxy retain prior
				// X-Forwarded-For information as a comma+space
				// separated list and fold multiple headers into one.
				prior, ok := pr.Out.Header["X-Forwarded-For"]
				omit := ok && prior == nil // Issue 38079: nil now means don't populate the header
				if len(prior) > 0 {
					clientIP = strings.Join(prior, ", ") + ", " + clientIP
				}
				if !omit {
					pr.Out.Header.Set("X-Forwarded-For", clientIP)
				}
			}
		}

		pr.Out.URL.Scheme = target.Scheme
		pr.Out.URL.Host = target.Host

		u := pr.Out.URL
		if pr.Out.RequestURI != "" {
			parsedURL, err := url.ParseRequestURI(pr.Out.RequestURI)
			if err == nil {
				u = parsedURL
			}
		}

		pr.Out.URL.Path = u.Path
		pr.Out.URL.RawPath = u.RawPath

		if preservePath {
			pr.Out.URL.Path, pr.Out.URL.RawPath = JoinURLPath(target, u)
		}

		// If a plugin/middleware adds semicolons in query params, they should be urlEncoded.
		pr.Out.URL.RawQuery = strings.ReplaceAll(u.RawQuery, ";", "&")
		pr.Out.RequestURI = "" // Outgoing request should not have RequestURI

		pr.Out.Proto = "HTTP/1.1"
		pr.Out.ProtoMajor = 1
		pr.Out.ProtoMinor = 1

		// Do not pass client Host header unless option PassHostHeader is set.
		if !passHostHeader {
			pr.Out.Host = pr.Out.URL.Host
		}

		if isWebSocketUpgrade(pr.Out) {
			cleanWebSocketHeaders(pr.Out)
		}
	}
}

// copyForwardedHeader copies header that are removed by the reverseProxy when a rewriteRequest is used.
func copyForwardedHeader(dst, src http.Header) {
	prior, ok := src["X-Forwarded-For"]
	if ok {
		dst["X-Forwarded-For"] = prior
	}
	prior, ok = src["Forwarded"]
	if ok {
		dst["Forwarded"] = prior
	}
	prior, ok = src["X-Forwarded-Host"]
	if ok {
		dst["X-Forwarded-Host"] = prior
	}
	prior, ok = src["X-Forwarded-Proto"]
	if ok {
		dst["X-Forwarded-Proto"] = prior
	}
}

// cleanWebSocketHeaders Even if the websocket RFC says that headers should be case-insensitive,
// some servers need Sec-WebSocket-Key, Sec-WebSocket-Extensions, Sec-WebSocket-Accept,
// Sec-WebSocket-Protocol and Sec-WebSocket-Version to be case-sensitive.
// https://tools.ietf.org/html/rfc6455#page-20
func cleanWebSocketHeaders(req *http.Request) {
	req.Header["Sec-WebSocket-Key"] = req.Header["Sec-Websocket-Key"]
	delete(req.Header, "Sec-Websocket-Key")

	req.Header["Sec-WebSocket-Extensions"] = req.Header["Sec-Websocket-Extensions"]
	delete(req.Header, "Sec-Websocket-Extensions")

	req.Header["Sec-WebSocket-Accept"] = req.Header["Sec-Websocket-Accept"]
	delete(req.Header, "Sec-Websocket-Accept")

	req.Header["Sec-WebSocket-Protocol"] = req.Header["Sec-Websocket-Protocol"]
	delete(req.Header, "Sec-Websocket-Protocol")

	req.Header["Sec-WebSocket-Version"] = req.Header["Sec-Websocket-Version"]
	delete(req.Header, "Sec-Websocket-Version")
}

func isWebSocketUpgrade(req *http.Request) bool {
	return httpguts.HeaderValuesContainsToken(req.Header["Connection"], "Upgrade") &&
		strings.EqualFold(req.Header.Get("Upgrade"), "websocket")
}

// ErrorHandler is the http.Handler called when something goes wrong when forwarding the request.
func ErrorHandler(w http.ResponseWriter, req *http.Request, err error) {
	ErrorHandlerWithContext(req.Context(), w, err)
}

// ErrorHandlerWithContext is the http.Handler called when something goes wrong when forwarding the request.
func ErrorHandlerWithContext(ctx context.Context, w http.ResponseWriter, err error) {
	statusCode := ComputeStatusCode(err)

	logger := log.Ctx(ctx)

	// Log the error with error level if it is a TLS error related to configuration.
	if isTLSConfigError(err) {
		logger.Error().Err(err).Msgf("%d %s", statusCode, statusText(statusCode))
	} else {
		logger.Debug().Err(err).Msgf("%d %s", statusCode, statusText(statusCode))
	}

	w.WriteHeader(statusCode)
	if _, werr := w.Write([]byte(statusText(statusCode))); werr != nil {
		logger.Debug().Err(werr).Msg("Error while writing status code")
	}
}

func statusText(statusCode int) string {
	if statusCode == StatusClientClosedRequest {
		return StatusClientClosedRequestText
	}
	return http.StatusText(statusCode)
}

// isTLSConfigError returns true if the error is a TLS error which is related to configuration.
// We assume that if the error is a tls.RecordHeaderError or a tls.CertificateVerificationError,
// it is related to configuration, because the client should not send a TLS request to a non-TLS server,
// and the client configuration should allow to verify the server certificate.
func isTLSConfigError(err error) bool {
	// tls.RecordHeaderError is returned when the client sends a TLS request to a non-TLS server.
	var recordHeaderErr tls.RecordHeaderError
	if errors.As(err, &recordHeaderErr) {
		return true
	}

	// tls.CertificateVerificationError is returned when the server certificate cannot be verified.
	var certVerificationErr *tls.CertificateVerificationError
	return errors.As(err, &certVerificationErr)
}

// ComputeStatusCode computes the HTTP status code according to the given error.
func ComputeStatusCode(err error) int {
	switch {
	case errors.Is(err, io.EOF):
		return http.StatusBadGateway
	case errors.Is(err, context.Canceled):
		return StatusClientClosedRequest
	default:
		var netErr net.Error
		if errors.As(err, &netErr) {
			if netErr.Timeout() {
				return http.StatusGatewayTimeout
			}

			return http.StatusBadGateway
		}
	}

	return http.StatusInternalServerError
}

// JoinURLPath computes the joined path and raw path of the given URLs.
// From https://github.com/golang/go/blob/b521ebb55a9b26c8824b219376c7f91f7cda6ec2/src/net/http/httputil/reverseproxy.go#L221
func JoinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}

	// Same as singleJoiningSlash, but uses EscapedPath to determine
	// whether a slash should be added
	apath := a.EscapedPath()
	bpath := b.EscapedPath()

	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")

	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
	return a.Path + b.Path, apath + bpath
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
