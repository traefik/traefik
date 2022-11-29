package httputil

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"golang.org/x/net/http/httpguts"
	"golang.org/x/net/http2"
)

// StatusClientClosedRequest non-standard HTTP status code for client disconnection.
const StatusClientClosedRequest = 499

// StatusClientClosedRequestText non-standard HTTP status for client disconnection.
const StatusClientClosedRequestText = "Client Closed Request"

type h2cTransportWrapper struct {
	*http2.Transport
}

func (t *h2cTransportWrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	return t.Transport.RoundTrip(req)
}

// ProxyBuilder handles the http.RoundTripper for httputil reverse proxies.
type ProxyBuilder struct {
	bufferPool *bufferPool
	// lock isn't needed because ProxyBuilder is not called concurrently.
	roundTrippers map[string]http.RoundTripper
}

// NewProxyBuilder creates a new ProxyBuilder.
func NewProxyBuilder() *ProxyBuilder {
	return &ProxyBuilder{
		bufferPool:    newBufferPool(),
		roundTrippers: make(map[string]http.RoundTripper),
	}
}

// Delete deletes the round-tripper corresponding to the given dynamic.HTTPClientConfig.
func (r *ProxyBuilder) Delete(cfgName string) {
	delete(r.roundTrippers, cfgName)
}

// Build builds a new httputil.ReverseProxy with the given configuration.
func (r *ProxyBuilder) Build(cfgName string, cfg *dynamic.HTTPClientConfig, tlsConfig *tls.Config, targetURL *url.URL) (http.Handler, error) {
	roundTripper, ok := r.roundTrippers[cfgName]
	if !ok {
		var err error
		roundTripper, err = createRoundTripper(cfg, tlsConfig)
		if err != nil {
			return nil, err
		}

		r.roundTrippers[cfgName] = roundTripper
	}

	return &httputil.ReverseProxy{
		Transport:    roundTripper,
		BufferPool:   r.bufferPool,
		ErrorHandler: ErrorHandler,
		Director: func(outReq *http.Request) {
			outReq.URL.Scheme = targetURL.Scheme
			outReq.URL.Host = targetURL.Host

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

			// Do not pass client Host header unless PassHostHeader is set.
			if !cfg.PassHostHeader {
				outReq.Host = outReq.URL.Host
			}

			if _, ok := outReq.Header["User-Agent"]; !ok {
				outReq.Header.Set("User-Agent", "")
			}

			cleanWebSocketHeaders(outReq)
		},
	}, nil
}

// createRoundTripper creates a http.RoundTripper configured with the Transport configuration settings.
// For the settings that can't be configured in Traefik it uses the default http.Transport settings.
// An exception to this is the MaxIdleConns setting as we only provide the option MaxIdleConnsPerHost in Traefik at this point in time.
// Setting this value to the default of 100 could lead to confusing behavior and backwards compatibility issues.
func createRoundTripper(cfg *dynamic.HTTPClientConfig, tlsConfig *tls.Config) (http.RoundTripper, error) {
	if cfg == nil {
		return nil, errors.New("no transport configuration given")
	}

	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	if cfg.ForwardingTimeouts != nil {
		dialer.Timeout = time.Duration(cfg.ForwardingTimeouts.DialTimeout)
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		MaxIdleConnsPerHost:   cfg.MaxIdleConnsPerHost,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ReadBufferSize:        64 * 1024,
		WriteBufferSize:       64 * 1024,
		TLSClientConfig:       tlsConfig,
	}

	if cfg.ForwardingTimeouts != nil {
		transport.ResponseHeaderTimeout = time.Duration(cfg.ForwardingTimeouts.ResponseHeaderTimeout)
		transport.IdleConnTimeout = time.Duration(cfg.ForwardingTimeouts.IdleConnTimeout)
	}

	return newSmartRoundTripper(transport, cfg.ForwardingTimeouts)
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

// ErrorHandler is the http.Handler called when something goes wrong when forwarding the request.
func ErrorHandler(w http.ResponseWriter, req *http.Request, err error) {
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

	logger := log.Ctx(req.Context())
	logger.Debug().Err(err).Msgf("%d %s", statusCode, statusText(statusCode))

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
