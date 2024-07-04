package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/middlewares/forwardedheaders"
	"golang.org/x/net/http/httpguts"
)

// StatusClientClosedRequest non-standard HTTP status code for client disconnection.
const StatusClientClosedRequest = 499

// StatusClientClosedRequestText non-standard HTTP status for client disconnection.
const StatusClientClosedRequestText = "Client Closed Request"

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
		Rewrite: func(req *httputil.ProxyRequest) {
			u := req.In.URL
			if req.In.RequestURI != "" {
				parsedURL, err := url.ParseRequestURI(req.In.RequestURI)
				if err == nil {
					u = parsedURL
				}
			}

			req.Out.URL.Path = u.Path
			req.Out.URL.RawPath = u.RawPath
			// If a plugin/middleware adds semicolons in query params, they should be urlEncoded.
			req.Out.URL.RawQuery = strings.ReplaceAll(u.RawQuery, ";", "&")
			req.Out.RequestURI = "" // Outgoing request should not have RequestURI

			// Do not pass client Host header unless PassHostHeader is set.
			if passHostHeader != nil && !*passHostHeader {
				req.Out.Host = req.In.URL.Host
			}

			// Add back removed Forwarded Headers.
			req.Out.Header["Forwarded"] = req.In.Header["Forwarded"]
			req.Out.Header["X-Forwarded-For"] = req.In.Header["X-Forwarded-For"]
			req.Out.Header["X-Forwarded-Host"] = req.In.Header["X-Forwarded-Host"]
			req.Out.Header["X-Forwarded-Proto"] = req.In.Header["X-Forwarded-Proto"]

			// In case of a ProxyProtocol connection the http.Request#RemoteAddr is the Client one.
			// To populate the X-Forwarded-For header we have to use the peer socket address.
			// Adapted from httputil.ReverseProxy
			remoteAddr := req.In.RemoteAddr
			if xForwardedForAddr, ok := req.In.Context().Value(forwardedheaders.XForwardedForAddr).(string); ok {
				remoteAddr = xForwardedForAddr
			}
			if clientIP, _, err := net.SplitHostPort(remoteAddr); err == nil {
				// If we aren't the first proxy retain prior
				// X-Forwarded-For information as a comma+space
				// separated list and fold multiple headers into one.
				prior, ok := req.In.Header["X-Forwarded-For"]
				omit := ok && prior == nil // nil means don't populate the header.
				if len(prior) > 0 {
					clientIP = strings.Join(prior, ", ") + ", " + clientIP
				}
				if !omit {
					req.Out.Header.Set("X-Forwarded-For", clientIP)
				}
			}

			// Even if the websocket RFC says that headers should be case-insensitive,
			// some servers need Sec-WebSocket-Key, Sec-WebSocket-Extensions, Sec-WebSocket-Accept,
			// Sec-WebSocket-Protocol and Sec-WebSocket-Version to be case-sensitive.
			// https://tools.ietf.org/html/rfc6455#page-20
			if isWebSocketUpgrade(req.In) {
				req.Out.Header["Sec-WebSocket-Key"] = req.In.Header["Sec-Websocket-Key"]
				req.Out.Header["Sec-WebSocket-Extensions"] = req.In.Header["Sec-Websocket-Extensions"]
				req.Out.Header["Sec-WebSocket-Accept"] = req.In.Header["Sec-Websocket-Accept"]
				req.Out.Header["Sec-WebSocket-Protocol"] = req.In.Header["Sec-Websocket-Protocol"]
				req.Out.Header["Sec-WebSocket-Version"] = req.In.Header["Sec-Websocket-Version"]
				delete(req.Out.Header, "Sec-Websocket-Key")
				delete(req.Out.Header, "Sec-Websocket-Extensions")
				delete(req.Out.Header, "Sec-Websocket-Accept")
				delete(req.Out.Header, "Sec-Websocket-Protocol")
				delete(req.Out.Header, "Sec-Websocket-Version")
			}
		},
		Transport:     roundTripper,
		FlushInterval: time.Duration(flushInterval),
		BufferPool:    bufferPool,
		ErrorHandler: func(w http.ResponseWriter, request *http.Request, err error) {
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

			log.Debugf("'%d %s' caused by: %v", statusCode, statusText(statusCode), err)
			w.WriteHeader(statusCode)
			_, werr := w.Write([]byte(statusText(statusCode)))
			if werr != nil {
				log.Debugf("Error while writing status code", werr)
			}
		},
	}

	return proxy, nil
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
