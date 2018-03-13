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
	// BackendAddr is the map key used for the IP:port of the Traefik backend (extracted from BackendURL)
	BackendAddr = "BackendAddr"
	// ClientAddr is the map key used for the remote address in its original form (usually IP:port).
	ClientAddr = "ClientAddr"
	// ClientHost is the map key used for the remote IP address from which the client request was received.
	ClientHost = "ClientHost"
	// ClientPort is the map key used for the remote TCP port from which the client request was received.
	ClientPort = "ClientPort"
	// ClientUsername is the map key used for the username provided in the URL, if present.
	ClientUsername = "ClientUsername"
	// RequestAddr is the map key used for the HTTP Host header (usually IP:port). This is treated as not a header by the Go API.
	RequestAddr = "RequestAddr"
	// RequestHost is the map key used for the HTTP Host server name (not including port).
	RequestHost = "RequestHost"
	// RequestPort is the map key used for the TCP port from the HTTP Host.
	RequestPort = "RequestPort"
	// RequestMethod is the map key used for the HTTP method.
	RequestMethod = "RequestMethod"
	// RequestPath is the map key used for the HTTP request URI, not including the scheme, host or port.
	RequestPath = "RequestPath"
	// RequestProtocol is the map key used for the version of HTTP requested.
	RequestProtocol = "RequestProtocol"
	// RequestLine is the original request line
	RequestLine = "RequestLine"
	// RequestContentSize is the map key used for the number of bytes in the request entity (a.k.a. body) sent by the client.
	RequestContentSize = "RequestContentSize"
	// RequestRefererHeader is the Referer header in the request
	RequestRefererHeader = "request_Referer"
	// RequestUserAgentHeader is the User-Agent header in the request
	RequestUserAgentHeader = "request_User-Agent"
	// OriginDuration is the map key used for the time taken by the origin server ('upstream') to return its response.
	OriginDuration = "OriginDuration"
	// OriginContentSize is the map key used for the content length specified by the origin server, or 0 if unspecified.
	OriginContentSize = "OriginContentSize"
	// OriginStatus is the map key used for the HTTP status code returned by the origin server.
	// If the request was handled by this Traefik instance (e.g. with a redirect), then this value will be absent.
	OriginStatus = "OriginStatus"
	// OriginStatusLine is the map key used for the HTTP status code and corresponding descriptive string.
	// If the request was handled by this Traefik instance (e.g. with a redirect), then this value will be absent.
	// Note that the actual message string might be different to what is reported here, depending on server behaviour.
	OriginStatusLine = "OriginStatusLine"
	// DownstreamStatus is the map key used for the HTTP status code returned to the client.
	DownstreamStatus = "DownstreamStatus"
	// DownstreamStatusLine is the map key used for the HTTP status line returned to the client.
	DownstreamStatusLine = "DownstreamStatusLine"
	// DownstreamContentSize is the map key used for the number of bytes in the response entity returned to the client.
	// This is in addition to the "Content-Length" header, which may be present in the origin response.
	DownstreamContentSize = "DownstreamContentSize"
	// RequestCount is the map key used for the number of requests received since the Traefik instance started.
	RequestCount = "RequestCount"
	// GzipRatio is the map key used for the response body compression ratio achieved.
	GzipRatio = "GzipRatio"
	// Overhead is the map key used for the processing time overhead caused by Traefik.
	Overhead = "Overhead"
	// RetryAttempts is the map key used for the amount of attempts the request was retried.
	RetryAttempts = "RetryAttempts"
)

// These are written out in the default case when no config is provided to specify keys of interest.
var defaultCoreKeys = [...]string{
	StartUTC,
	Duration,
	FrontendName,
	BackendName,
	BackendURL,
	ClientHost,
	ClientPort,
	ClientUsername,
	RequestHost,
	RequestPort,
	RequestMethod,
	RequestPath,
	RequestProtocol,
	RequestContentSize,
	OriginDuration,
	OriginContentSize,
	OriginStatus,
	DownstreamStatus,
	DownstreamContentSize,
	RequestCount,
}

// This contains the set of all keys, i.e. all the default keys plus all non-default keys.
var allCoreKeys = make(map[string]struct{})

func init() {
	for _, k := range defaultCoreKeys {
		allCoreKeys[k] = struct{}{}
	}
	allCoreKeys[BackendAddr] = struct{}{}
	allCoreKeys[ClientAddr] = struct{}{}
	allCoreKeys[RequestAddr] = struct{}{}
	allCoreKeys[RequestLine] = struct{}{}
	allCoreKeys[OriginStatusLine] = struct{}{}
	allCoreKeys[DownstreamStatusLine] = struct{}{}
	allCoreKeys[GzipRatio] = struct{}{}
	allCoreKeys[StartLocal] = struct{}{}
	allCoreKeys[Overhead] = struct{}{}
	allCoreKeys[RetryAttempts] = struct{}{}
}

// CoreLogData holds the fields computed from the request/response.
type CoreLogData map[string]interface{}

// LogData is the data captured by the middleware so that it can be logged.
type LogData struct {
	Core               CoreLogData
	Request            http.Header
	OriginResponse     http.Header
	DownstreamResponse http.Header
}
