package accesslog

import (
	"net/http"
)

const (
	// StartUTC is the map key used for the time at which request processing started.
	StartUTC = "StartUTC"
	// StartLocal is the map key used for the local time at which request processing started.
	StartLocal = "StartLocal"
	// Duration is the map key used for the total time taken by processing the response, including the origin server's time but
	// not the log writing time.
	Duration = "Duration"
	// FrontendName is the map key used for the name of the Traefik frontend.
	FrontendName = "FrontendName"
	// BackendName is the map key used for the name of the Traefik backend.
	BackendName = "BackendName"
	// BackendURL is the map key used for the URL of the Traefik backend.
	BackendURL = "BackendURL"
	// ClientRemoteAddr is the map key used for the remote address in its original form (usually IP:port).
	ClientRemoteAddr = "ClientRemoteAddr"
	// ClientHost is the map key used for the remote IP address from which the client request was received.
	ClientHost = "ClientHost"
	// ClientPort is the map key used for the remote TCP port from which the client request was received.
	ClientPort = "ClientPort"
	// ClientUsername is the map key used for the username provided in the URL, if present.
	ClientUsername = "ClientUsername"
	// HTTPHost is the map key used for the HTTP Host header (which is treated as not a header by the Go API).
	HTTPHost = "HTTPHost"
	// HTTPMethod is the map key used for the HTTP method.
	HTTPMethod = "HTTPMethod"
	// HTTPRequestPath is the map key used for the HTTP request URI, not including the scheme, host or port.
	HTTPRequestPath = "HTTPRequestPath"
	// HTTPProtocol is the map key used for the version of HTTP requested.
	HTTPProtocol = "HTTPProtocol"
	// HTTPRequestLine is the original request line
	HTTPRequestLine = "HTTPRequestLine"
	// OriginDuration is the map key used for the time taken by the origin server ('upstream') to return its response.
	OriginDuration = "OriginDuration"
	// OriginContentSize is the map key used for the content length specified by the origin server, or 0 if unspecified.
	OriginContentSize = "OriginContentSize"
	// OriginStatus is the map key used for the HTTP status code returned by the origin server.
	// If the request was handled by this Traefik instance (e.g. with a redirect), then this value will be zero.
	OriginStatus = "OriginStatus"
	// DownstreamStatus is the map key used for the HTTP status code returned to the client.
	DownstreamStatus = "DownstreamStatus"
	// DownstreamContentSize is the map key used for the number of bytes in the response entity returned to the client.
	DownstreamContentSize = "DownstreamContentSize"
	// RequestCount is the map key used for the number of requests received since the Traefik instance started.
	RequestCount = "RequestCount"
	// GzipRatio is the map key used for the response body compression ratio achieved.
	GzipRatio = "GzipRatio"
	// Overhead is the map key used for the processing time overhead caused by Traefik.
	Overhead = "Overhead"
)

var defaultCoreKeys = []string{
	StartUTC,
	Duration,
	FrontendName,
	BackendName,
	BackendURL,
	ClientRemoteAddr,
	ClientHost,
	ClientPort,
	ClientUsername,
	HTTPHost,
	HTTPMethod,
	HTTPRequestPath,
	HTTPRequestLine,
	HTTPProtocol,
	OriginDuration,
	OriginContentSize,
	OriginStatus,
	DownstreamStatus,
	DownstreamContentSize,
	RequestCount,
}

var allCoreKeys = make(map[string]struct{})

func init() {
	for _, k := range defaultCoreKeys {
		allCoreKeys[k] = struct{}{}
	}
	allCoreKeys[GzipRatio] = struct{}{}
	allCoreKeys[StartLocal] = struct{}{}
	allCoreKeys[Overhead] = struct{}{}
}

// CoreLogData is analysed by reflection - don't add any nested elements (but some methods are supported, below)
type CoreLogData map[string]interface{}

// LogData is the data captured by the middleware so that it can be logged.
type LogData struct {
	Core               CoreLogData
	Request            http.Header
	OriginResponse     http.Header
	DownstreamResponse http.Header
}
