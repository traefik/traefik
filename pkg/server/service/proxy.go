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

	"github.com/rs/zerolog/log"
	"golang.org/x/net/http/httpguts"
)

// StatusClientClosedRequest non-standard HTTP status code for client disconnection.
const StatusClientClosedRequest = 499

// StatusClientClosedRequestText non-standard HTTP status for client disconnection.
const StatusClientClosedRequestText = "Client Closed Request"

// StatusCloseConnection non-standard HTTP sttatus code for close connection.
const StatusCloseConnection = 444

// StatusCloseConnectionText non-standardHTTP status for close connection.
const StatusCloseConnectionText = "No Response"

// ErrStatusCloseConnection non-standard erreor for close connection
var ErrStatusCloseConnection = errors.New(StatusCloseConnectionText)

func buildSingleHostProxy(target *url.URL, passHostHeader bool, flushInterval time.Duration, roundTripper http.RoundTripper, bufferPool httputil.BufferPool) http.Handler {
	return &httputil.ReverseProxy{
		Director:       directorBuilder(target, passHostHeader),
		Transport:      roundTripper,
		FlushInterval:  flushInterval,
		BufferPool:     bufferPool,
		ModifyResponse: modifyResponse,
		ErrorHandler:   errorHandler,
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
		// If a plugin/middleware adds semicolons in query params, they should be urlEncoded.
		outReq.URL.RawQuery = strings.ReplaceAll(u.RawQuery, ";", "&")
		outReq.RequestURI = "" // Outgoing request should not have RequestURI

		outReq.Proto = "HTTP/1.1"
		outReq.ProtoMajor = 1
		outReq.ProtoMinor = 1

		// Do not pass client Host header unless optsetter PassHostHeader is set.
		if !passHostHeader {
			outReq.Host = outReq.URL.Host
		}

		cleanWebSocketHeaders(outReq)
	}
}

// cleanWebSocketHeaders Even if the websocket RFC says that headers should be case-insensitive,
// some servers need Sec-WebSocket-Key, Sec-WebSocket-Extensions, Sec-WebSocket-Accept,
// Sec-WebSocket-Protocol and Sec-WebSocket-Version to be case-sensitive.
// https://tools.ietf.org/html/rfc6455#page-20
func cleanWebSocketHeaders(req *http.Request) {
	if !isWebSocketUpgrade(req) {
		return
	}

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

func errorHandler(w http.ResponseWriter, req *http.Request, err error) {
	statusCode := http.StatusInternalServerError

	switch {
	case errors.Is(err, io.EOF):
		statusCode = http.StatusBadGateway
	case errors.Is(err, context.Canceled):
		statusCode = StatusClientClosedRequest
	case errors.Is(err, ErrStatusCloseConnection):
		statusCode = StatusCloseConnection
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

	logger := log.Ctx(req.Context())
	logger.Debug().Err(err).Msgf("%d %s", statusCode, statusText(statusCode))

	if statusCode == StatusCloseConnection {
		conn, _, hjErr := w.(http.Hijacker).Hijack()
		if hjErr == nil {
			conn.Close()
			return
		}
	}

	w.WriteHeader(statusCode)
	if _, werr := w.Write([]byte(statusText(statusCode))); werr != nil {
		logger.Debug().Err(werr).Msg("Error while writing status code")
	}
}

func modifyResponse(response *http.Response) error {
	switch response.StatusCode {
	case StatusCloseConnection:
		return ErrStatusCloseConnection
	default:
		return nil
	}
}

func statusText(statusCode int) string {
	switch statusCode {
	case StatusClientClosedRequest:
		return StatusClientClosedRequestText
	case StatusCloseConnection:
		return StatusCloseConnectionText
	default:
		return http.StatusText(statusCode)
	}
}
