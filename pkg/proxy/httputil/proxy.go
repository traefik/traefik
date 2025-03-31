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
	"github.com/traefik/traefik/v3/pkg/logs"
	"golang.org/x/net/http/httpguts"
)

const (
	// StatusClientClosedRequest non-standard HTTP status code for client disconnection.
	StatusClientClosedRequest = 499

	// StatusClientClosedRequestText non-standard HTTP status for client disconnection.
	StatusClientClosedRequestText = "Client Closed Request"
)

func buildSingleHostProxy(target *url.URL, passHostHeader bool, preservePath bool, flushInterval time.Duration, roundTripper http.RoundTripper, bufferPool httputil.BufferPool) http.Handler {
	return &httputil.ReverseProxy{
		Director:      directorBuilder(target, passHostHeader, preservePath),
		Transport:     roundTripper,
		FlushInterval: flushInterval,
		BufferPool:    bufferPool,
		ErrorLog:      stdlog.New(logs.NoLevel(log.Logger, zerolog.DebugLevel), "", 0),
		ErrorHandler:  ErrorHandler,
	}
}

func directorBuilder(target *url.URL, passHostHeader bool, preservePath bool) func(req *http.Request) {
	return func(outReq *http.Request) {
		outReq.URL.Scheme = target.Scheme
		outReq.URL.Host = target.Host

		u := outReq.URL
		if outReq.RequestURI != "" {
			parsedURL, err := url.ParseRequestURI(outReq.RequestURI)
			if err == nil {
				u = parsedURL
			}
		}

		outReq.URL.Path = u.Path
		outReq.URL.RawPath = u.RawPath

		if preservePath {
			outReq.URL.Path, outReq.URL.RawPath = JoinURLPath(target, u)
		}

		// If a plugin/middleware adds semicolons in query params, they should be urlEncoded.
		outReq.URL.RawQuery = strings.ReplaceAll(u.RawQuery, ";", "&")
		outReq.RequestURI = "" // Outgoing request should not have RequestURI

		outReq.Proto = "HTTP/1.1"
		outReq.ProtoMajor = 1
		outReq.ProtoMinor = 1

		// Do not pass client Host header unless option PassHostHeader is set.
		if !passHostHeader {
			outReq.Host = outReq.URL.Host
		}

		if isWebSocketUpgrade(outReq) {
			cleanWebSocketHeaders(outReq)
		}
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
