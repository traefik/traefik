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

// Headers that ingress-nginx sets in the external auth location (proxy_set_header).
// X-Original-URI and X-Scheme are set only when auth-method is set; the others are always set.
// The client values must never reach the auth service.
// The headers list has been extracted from here: https://github.com/kubernetes/ingress-nginx/blob/main/rootfs/etc/nginx/template/nginx.tmpl
const (
	xOriginalURL         = "X-Original-Url"
	xOriginalURI         = "X-Original-Uri"
	xOriginalMethod      = "X-Original-Method"
	xAuthRequestRedirect = "X-Auth-Request-Redirect"
	xRealIP              = "X-Real-IP"
	xScheme              = "X-Scheme"

	xSentFrom = "X-Sent-From"
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
