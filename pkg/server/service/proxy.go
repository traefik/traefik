package service

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/traefik/traefik/v2/pkg/log"
	"golang.org/x/net/http/httpguts"
)

// StatusClientClosedRequest non-standard HTTP status code for client disconnection.
const StatusClientClosedRequest = 499

// StatusClientClosedRequestText non-standard HTTP status for client disconnection.
const StatusClientClosedRequestText = "Client Closed Request"

func buildSingleHostProxy(target *url.URL, passHostHeader bool, flushInterval time.Duration, roundTripper http.RoundTripper, bufferPool httputil.BufferPool) http.Handler {
	return &httputil.ReverseProxy{
		Director:      directorBuilder(target, passHostHeader),
		Transport:     roundTripper,
		FlushInterval: flushInterval,
		BufferPool:    bufferPool,
		ErrorHandler:  errorHandler,
	}
}

func directorBuilder(target *url.URL, passHostHeader bool) func(req *http.Request) {
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
		outReq.URL.RawQuery = strings.ReplaceAll(u.RawQuery, ";", "&")
		outReq.RequestURI = "" // Outgoing request should not have RequestURI

		outReq.Proto = "HTTP/1.1"
		outReq.ProtoMajor = 1
		outReq.ProtoMinor = 1

		if _, ok := outReq.Header["User-Agent"]; !ok {
			outReq.Header.Set("User-Agent", "")
		}

		// Do not pass client Host header unless PassHostHeader is set.
		if !passHostHeader {
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
	}
}

func errorHandler(w http.ResponseWriter, req *http.Request, err error) {
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

	logger := log.FromContext(req.Context())
	logger.Debugf("'%d %s' caused by: %v", statusCode, statusText(statusCode), err)

	w.WriteHeader(statusCode)
	if _, werr := w.Write([]byte(statusText(statusCode))); werr != nil {
		logger.Debugf("Error while writing status code", werr)
	}
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
