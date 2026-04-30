package snippet

import (
	"net/http"
	"net/textproto"
	"strings"

	"github.com/vulcand/oxy/v2/forward"
	"golang.org/x/net/http/httpguts"
)

const (
	connectionHeader = "Connection"
	upgradeHeader    = "Upgrade"
)

const (
	xForwardedURI    = "X-Forwarded-Uri"
	xForwardedMethod = "X-Forwarded-Method"
)

var userAgentHeader = http.CanonicalHeaderKey("User-Agent")

// RemoveConnectionHeaders removes hop-by-hop headers listed in the "Connection" header.
// See RFC 7230, section 6.1.
func RemoveConnectionHeaders(req *http.Request) {
	var reqUpType string
	if httpguts.HeaderValuesContainsToken(req.Header[connectionHeader], upgradeHeader) {
		reqUpType = req.Header.Get(upgradeHeader)
	}

	for _, f := range req.Header[connectionHeader] {
		for sf := range strings.SplitSeq(f, ",") {
			if sf = textproto.TrimString(sf); sf != "" {
				req.Header.Del(sf)
			}
		}
	}

	if reqUpType != "" {
		req.Header.Set(connectionHeader, upgradeHeader)
		req.Header.Set(upgradeHeader, reqUpType)
	} else {
		req.Header.Del(connectionHeader)
	}
}

// hopHeaders Hop-by-hop headers to be removed in the authentication request.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
// Proxy-Authorization header is forwarded to the authentication server (see https://tools.ietf.org/html/rfc7235#section-4.4).
var hopHeaders = []string{
	forward.Connection,
	forward.KeepAlive,
	forward.Te, // canonicalized version of "TE"
	forward.Trailers,
	forward.TransferEncoding,
	forward.Upgrade,
}
