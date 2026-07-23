package service

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	stdlog "log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"golang.org/x/net/http/httpguts"
)

// StatusClientClosedRequest non-standard HTTP status code for client disconnection.
const StatusClientClosedRequest = 499

// StatusClientClosedRequestText non-standard HTTP status for client disconnection.
const StatusClientClosedRequestText = "Client Closed Request"

// errorLogger is a logger instance used to log proxy errors.
// This logger is a shared instance as having one instance by proxy introduces a memory and go routine leak.
// The writer go routine is never stopped as the finalizer is never called.
// See https://github.com/sirupsen/logrus/blob/d1e6332644483cfee14de11099f03645561d55f8/writer.go#L57).
var errorLogger = stdlog.New(log.WithoutContext().WriterLevel(logrus.DebugLevel), "", 0)

func buildProxy(passHostHeader *bool, responseForwarding *dynamic.ResponseForwarding, roundTripper http.RoundTripper, bufferPool httputil.BufferPool) (http.Handler, error) {
	var flushInterval ptypes.Duration
	if responseForwarding != nil {
		err := flushInterval.Set(responseForwarding.FlushInterval)
		if err != nil {
			return nil, fmt.Errorf("error creating flush interval: %w", err)
		}
	}
	if flushInterval == 0 {
		flushInterval = ptypes.Duration(100 * time.Millisecond)
	}

	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			copyForwardedHeader(pr.Out.Header, pr.In.Header)

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

			u := pr.Out.URL
			if pr.Out.RequestURI != "" {
				parsedURL, err := url.ParseRequestURI(pr.Out.RequestURI)
				if err == nil {
					u = parsedURL
				}
			}

			pr.Out.URL.Path = u.Path
			pr.Out.URL.RawPath = u.RawPath
			// If a plugin/middleware adds semicolons in query params, they should be urlEncoded.
			pr.Out.URL.RawQuery = strings.ReplaceAll(u.RawQuery, ";", "&")
			pr.Out.RequestURI = "" // Outgoing request should not have RequestURI

			pr.Out.Proto = "HTTP/1.1"
			pr.Out.ProtoMajor = 1
			pr.Out.ProtoMinor = 1

			// Adding the "Connection: close" header to the request ensures that we are not reusing the connection for
			// subsequent requests in case the backend does not support CONNECT and returns a 2xx response.
			if pr.Out.Method == http.MethodConnect {
				pr.Out.Close = true
			}

			// Do not pass client Host header unless optsetter PassHostHeader is set.
			if passHostHeader != nil && !*passHostHeader {
				pr.Out.Host = pr.Out.URL.Host
			}

			// Even if the websocket RFC says that headers should be case-insensitive,
			// some servers need Sec-WebSocket-Key, Sec-WebSocket-Extensions, Sec-WebSocket-Accept,
			// Sec-WebSocket-Protocol and Sec-WebSocket-Version to be case-sensitive.
			// https://tools.ietf.org/html/rfc6455#page-20
			if isWebSocketUpgrade(pr.Out) {
				pr.Out.Header["Sec-WebSocket-Key"] = pr.Out.Header["Sec-Websocket-Key"]
				pr.Out.Header["Sec-WebSocket-Extensions"] = pr.Out.Header["Sec-Websocket-Extensions"]
				pr.Out.Header["Sec-WebSocket-Accept"] = pr.Out.Header["Sec-Websocket-Accept"]
				pr.Out.Header["Sec-WebSocket-Protocol"] = pr.Out.Header["Sec-Websocket-Protocol"]
				pr.Out.Header["Sec-WebSocket-Version"] = pr.Out.Header["Sec-Websocket-Version"]
				delete(pr.Out.Header, "Sec-Websocket-Key")
				delete(pr.Out.Header, "Sec-Websocket-Extensions")
				delete(pr.Out.Header, "Sec-Websocket-Accept")
				delete(pr.Out.Header, "Sec-Websocket-Protocol")
				delete(pr.Out.Header, "Sec-Websocket-Version")
			}
		},
		Transport:     roundTripper,
		FlushInterval: time.Duration(flushInterval),
		BufferPool:    bufferPool,
		ErrorLog:      errorLogger,
		ErrorHandler: func(w http.ResponseWriter, request *http.Request, err error) {
			logger := log.FromContext(request.Context())

			statusCode := http.StatusInternalServerError
			switch {
			case errors.Is(err, io.EOF):
				statusCode = http.StatusBadGateway
			case errors.Is(err, context.Canceled):
				statusCode = StatusClientClosedRequest
			default:
				var netErr net.Error
				if errors.As(err, &netErr) {
					if netErr.Timeout() {
						statusCode = http.StatusGatewayTimeout
					} else {
						statusCode = http.StatusBadGateway
					}
				}
			}

			// Log the error with error level if it is a TLS error related to configuration.
			if isTLSConfigError(err) {
				logger.Errorf("'%d %s' caused by: %v", statusCode, statusText(statusCode), err)
			} else {
				logger.Debugf("'%d %s' caused by: %v", statusCode, statusText(statusCode), err)
			}

			w.WriteHeader(statusCode)
			_, werr := w.Write([]byte(statusText(statusCode)))
			if werr != nil {
				log.WithoutContext().Debugf("Error while writing status code", werr)
			}
		},
	}

	return newConnectHandler(proxy), nil
}

// isTLSError returns true if the error is a TLS error which is related to configuration.
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

func isWebSocketUpgrade(req *http.Request) bool {
	if !httpguts.HeaderValuesContainsToken(req.Header["Connection"], "Upgrade") {
		return false
	}

	return strings.EqualFold(req.Header.Get("Upgrade"), "websocket")
}

func statusText(statusCode int) string {
	if statusCode == StatusClientClosedRequest {
		return StatusClientClosedRequestText
	}
	return http.StatusText(statusCode)
}

// copyForwardedHeader copies headers that are removed by ReverseProxy when Rewrite is used.
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
