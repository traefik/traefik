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
		Director: func(outReq *http.Request) {
			u := outReq.URL
			if outReq.RequestURI != "" {
				parsedURL, err := url.ParseRequestURI(outReq.RequestURI)
				if err == nil {
					u = parsedURL
				}
			}

			outReq.URL.Path = u.Path
			outReq.URL.RawPath = u.RawPath
			// If a plugin/middleware adds semicolons in query params, they should be urlEncoded.
			outReq.URL.RawQuery = strings.ReplaceAll(u.RawQuery, ";", "&")
			outReq.RequestURI = "" // Outgoing request should not have RequestURI

			outReq.Proto = "HTTP/1.1"
			outReq.ProtoMajor = 1
			outReq.ProtoMinor = 1

			// Do not pass client Host header unless optsetter PassHostHeader is set.
			if passHostHeader != nil && !*passHostHeader {
				outReq.Host = outReq.URL.Host
			}

			// Even if the websocket RFC says that headers should be case-insensitive,
			// some servers need Sec-WebSocket-Key, Sec-WebSocket-Extensions, Sec-WebSocket-Accept,
			// Sec-WebSocket-Protocol and Sec-WebSocket-Version to be case-sensitive.
			// https://tools.ietf.org/html/rfc6455#page-20
			if isWebSocketUpgrade(outReq) {
				outReq.Header["Sec-WebSocket-Key"] = outReq.Header["Sec-Websocket-Key"]
				outReq.Header["Sec-WebSocket-Extensions"] = outReq.Header["Sec-Websocket-Extensions"]
				outReq.Header["Sec-WebSocket-Accept"] = outReq.Header["Sec-Websocket-Accept"]
				outReq.Header["Sec-WebSocket-Protocol"] = outReq.Header["Sec-Websocket-Protocol"]
				outReq.Header["Sec-WebSocket-Version"] = outReq.Header["Sec-Websocket-Version"]
				delete(outReq.Header, "Sec-Websocket-Key")
				delete(outReq.Header, "Sec-Websocket-Extensions")
				delete(outReq.Header, "Sec-Websocket-Accept")
				delete(outReq.Header, "Sec-Websocket-Protocol")
				delete(outReq.Header, "Sec-Websocket-Version")
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
				log.Debugf("Error while writing status code", werr)
			}
		},
	}

	return proxy, nil
}

// isTLSError returns true if the error is a TLS error which is related to configuration.
// We assume that if the error is a tls.RecordHeaderError or a tls.CertificateVerificationError,
// it is related to configuration, because the client should not send a TLS request to a non-TLS server,
// and the client configuration should allow to verify the server certificate.
func isTLSConfigError(err error) bool {
	// tls.RecordHeaderError is returned when the client sends a TLS request to a non-TLS server.
	if _, ok := err.(tls.RecordHeaderError); ok {
		return true
	}

	// tls.CertificateVerificationError is returned when the server certificate cannot be verified.
	var tlsCertErr *tls.CertificateVerificationError
	if errors.As(err, &tlsCertErr) {
		return true
	}

	return false
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
