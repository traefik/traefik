package forward

import (
	"net"
	"net/http"
	"strings"

	"github.com/vulcand/oxy/utils"
)

// Rewriter is responsible for removing hop-by-hop headers and setting forwarding headers
type HeaderRewriter struct {
	TrustForwardHeader bool
	Hostname           string
}

func (rw *HeaderRewriter) Rewrite(req *http.Request) {
	if !rw.TrustForwardHeader {
		utils.RemoveHeaders(req.Header, XHeaders...)
	}

	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		if prior, ok := req.Header[XForwardedFor]; ok {
			req.Header.Set(XForwardedFor, strings.Join(prior, ", ")+", "+clientIP)
		} else {
			req.Header.Set(XForwardedFor, clientIP)
		}

		if req.Header.Get(XRealIp) == "" {
			req.Header.Set(XRealIp, clientIP)
		}
	}

	xfProto := req.Header.Get(XForwardedProto)
	if xfProto == "" {
		if req.TLS != nil {
			req.Header.Set(XForwardedProto, "https")
		} else {
			req.Header.Set(XForwardedProto, "http")
		}
	}

	if xfp := req.Header.Get(XForwardedPort); xfp == "" {
		req.Header.Set(XForwardedPort, forwardedPort(req))
	}

	if xfHost := req.Header.Get(XForwardedHost); xfHost == "" && req.Host != "" {
		req.Header.Set(XForwardedHost, req.Host)
	}

	if rw.Hostname != "" {
		req.Header.Set(XForwardedServer, rw.Hostname)
	}

	// Remove hop-by-hop headers to the backend.  Especially important is "Connection" because we want a persistent
	// connection, regardless of what the client sent to us.
	utils.RemoveHeaders(req.Header, HopHeaders...)
}

func forwardedPort(req *http.Request) string {
	if req == nil {
		return ""
	}

	if _, port, err := net.SplitHostPort(req.Host); err == nil && port != "" {
		return port
	}

	if req.TLS != nil {
		return "443"
	}

	return "80"
}
