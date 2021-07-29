package connectionheader

import (
	"net/http"
	"net/textproto"
	"strings"

	"golang.org/x/net/http/httpguts"
)

// Remove removes hop-by-hop headers listed in the "Connection" header of h.
// See RFC 7230, section 6.1.
func Remove(next http.Handler) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		reqUpType := upgradeType(req.Header)
		removeConnectionHeaders(req.Header)

		if reqUpType != "" {
			req.Header.Set("Connection", "Upgrade")
			req.Header.Set("Upgrade", reqUpType)
		} else {
			req.Header.Del("Connection")
		}

		next.ServeHTTP(rw, req)
	}
}

func removeConnectionHeaders(h http.Header) {
	for _, f := range h["Connection"] {
		for _, sf := range strings.Split(f, ",") {
			if sf = textproto.TrimString(sf); sf != "" {
				h.Del(sf)
			}
		}
	}
}

func upgradeType(h http.Header) string {
	if !httpguts.HeaderValuesContainsToken(h["Connection"], "Upgrade") {
		return ""
	}
	return h.Get("Upgrade")
}
