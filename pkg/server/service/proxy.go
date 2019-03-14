package service

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/containous/flaeg/parse"
	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/log"
)

// StatusClientClosedRequest non-standard HTTP status code for client disconnection
const StatusClientClosedRequest = 499

// StatusClientClosedRequestText non-standard HTTP status for client disconnection
const StatusClientClosedRequestText = "Client Closed Request"

func buildProxy(passHostHeader bool, responseForwarding *config.ResponseForwarding, defaultRoundTripper http.RoundTripper, bufferPool httputil.BufferPool, responseModifier func(*http.Response) error) (http.Handler, error) {
	var flushInterval parse.Duration
	if responseForwarding != nil {
		err := flushInterval.Set(responseForwarding.FlushInterval)
		if err != nil {
			return nil, fmt.Errorf("error creating flush interval: %v", err)
		}
	}
	if flushInterval == 0 {
		flushInterval = parse.Duration(100 * time.Millisecond)
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
			outReq.URL.RawQuery = u.RawQuery
			outReq.RequestURI = "" // Outgoing request should not have RequestURI

			outReq.Proto = "HTTP/1.1"
			outReq.ProtoMajor = 1
			outReq.ProtoMinor = 1

			// Do not pass client Host header unless optsetter PassHostHeader is set.
			if !passHostHeader {
				outReq.Host = outReq.URL.Host
			}

		},
		Transport:      defaultRoundTripper,
		FlushInterval:  time.Duration(flushInterval),
		ModifyResponse: responseModifier,
		BufferPool:     bufferPool,
		ErrorHandler: func(w http.ResponseWriter, request *http.Request, err error) {
			statusCode := http.StatusInternalServerError

			switch {
			case err == io.EOF:
				statusCode = http.StatusBadGateway
			case err == context.Canceled:
				statusCode = StatusClientClosedRequest
			default:
				if e, ok := err.(net.Error); ok {
					if e.Timeout() {
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

func statusText(statusCode int) string {
	if statusCode == StatusClientClosedRequest {
		return StatusClientClosedRequestText
	}
	return http.StatusText(statusCode)
}
