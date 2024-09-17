package auth

import (
	"net/http"
	"net/textproto"
	"strings"

	"golang.org/x/net/http/httpguts"
)

const (
	connectionHeader = "Connection"
	upgradeHeader    = "Upgrade"
)

// RemoveConnectionHeaders removes hop-by-hop headers listed in the "Connection" header.
// See RFC 7230, section 6.1.
func RemoveConnectionHeaders(req *http.Request) {
	var reqUpType string
	if httpguts.HeaderValuesContainsToken(req.Header[connectionHeader], upgradeHeader) {
		reqUpType = req.Header.Get(upgradeHeader)
	}

	for _, f := range req.Header[connectionHeader] {
		for _, sf := range strings.Split(f, ",") {
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
