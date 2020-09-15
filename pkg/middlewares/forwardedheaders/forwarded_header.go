package forwardedheaders

import (
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/traefik/traefik/v2/pkg/ip"
)

const (
	xForwardedProto             = "X-Forwarded-Proto"
	xForwardedFor               = "X-Forwarded-For"
	xForwardedHost              = "X-Forwarded-Host"
	xForwardedPort              = "X-Forwarded-Port"
	xForwardedServer            = "X-Forwarded-Server"
	xForwardedURI               = "X-Forwarded-Uri"
	xForwardedMethod            = "X-Forwarded-Method"
	xForwardedTLSClientCert     = "X-Forwarded-Tls-Client-Cert"
	xForwardedTLSClientCertInfo = "X-Forwarded-Tls-Client-Cert-Info"
	xRealIP                     = "X-Real-Ip"
	connection                  = "Connection"
	upgrade                     = "Upgrade"
)

var xHeaders = []string{
	xForwardedProto,
	xForwardedFor,
	xForwardedHost,
	xForwardedPort,
	xForwardedServer,
	xForwardedURI,
	xForwardedMethod,
	xForwardedTLSClientCert,
	xForwardedTLSClientCertInfo,
	xRealIP,
}

// XForwarded is an HTTP handler wrapper that sets the X-Forwarded headers,
// and other relevant headers for a reverse-proxy.
// Unless insecure is set,
// it first removes all the existing values for those headers if the remote address is not one of the trusted ones.
type XForwarded struct {
	insecure   bool
	trustedIps []string
	ipChecker  *ip.Checker
	next       http.Handler
	hostname   string
}

// NewXForwarded creates a new XForwarded.
func NewXForwarded(insecure bool, trustedIps []string, next http.Handler) (*XForwarded, error) {
	var ipChecker *ip.Checker
	if len(trustedIps) > 0 {
		var err error
		ipChecker, err = ip.NewChecker(trustedIps)
		if err != nil {
			return nil, err
		}
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}

	return &XForwarded{
		insecure:   insecure,
		trustedIps: trustedIps,
		ipChecker:  ipChecker,
		next:       next,
		hostname:   hostname,
	}, nil
}

func (x *XForwarded) isTrustedIP(ip string) bool {
	if x.ipChecker == nil {
		return false
	}
	return x.ipChecker.IsAuthorized(ip) == nil
}

// removeIPv6Zone removes the zone if the given IP is an ipv6 address and it has {zone} information in it,
// like "[fe80::d806:a55d:eb1b:49cc%vEthernet (vmxnet3 Ethernet Adapter - Virtual Switch)]:64692".
func removeIPv6Zone(clientIP string) string {
	return strings.Split(clientIP, "%")[0]
}

// isWebsocketRequest returns whether the specified HTTP request is a websocket handshake request.
func isWebsocketRequest(req *http.Request) bool {
	containsHeader := func(name, value string) bool {
		items := strings.Split(req.Header.Get(name), ",")
		for _, item := range items {
			if value == strings.ToLower(strings.TrimSpace(item)) {
				return true
			}
		}
		return false
	}
	return containsHeader(connection, "upgrade") && containsHeader(upgrade, "websocket")
}

func forwardedPort(req *http.Request) string {
	if req == nil {
		return ""
	}

	if _, port, err := net.SplitHostPort(req.Host); err == nil && port != "" {
		return port
	}

	if req.Header.Get(xForwardedProto) == "https" || req.Header.Get(xForwardedProto) == "wss" {
		return "443"
	}

	if req.TLS != nil {
		return "443"
	}

	return "80"
}

func (x *XForwarded) rewrite(outreq *http.Request) {
	if clientIP, _, err := net.SplitHostPort(outreq.RemoteAddr); err == nil {
		clientIP = removeIPv6Zone(clientIP)

		if outreq.Header.Get(xRealIP) == "" {
			outreq.Header.Set(xRealIP, clientIP)
		}
	}

	xfProto := outreq.Header.Get(xForwardedProto)
	if xfProto == "" {
		if isWebsocketRequest(outreq) {
			if outreq.TLS != nil {
				outreq.Header.Set(xForwardedProto, "wss")
			} else {
				outreq.Header.Set(xForwardedProto, "ws")
			}
		} else {
			if outreq.TLS != nil {
				outreq.Header.Set(xForwardedProto, "https")
			} else {
				outreq.Header.Set(xForwardedProto, "http")
			}
		}
	}

	if xfPort := outreq.Header.Get(xForwardedPort); xfPort == "" {
		outreq.Header.Set(xForwardedPort, forwardedPort(outreq))
	}

	if xfHost := outreq.Header.Get(xForwardedHost); xfHost == "" && outreq.Host != "" {
		outreq.Header.Set(xForwardedHost, outreq.Host)
	}

	if x.hostname != "" {
		outreq.Header.Set(xForwardedServer, x.hostname)
	}
}

// ServeHTTP implements http.Handler.
func (x *XForwarded) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !x.insecure && !x.isTrustedIP(r.RemoteAddr) {
		for _, h := range xHeaders {
			r.Header.Del(h)
		}
	}

	x.rewrite(r)

	x.next.ServeHTTP(w, r)
}
