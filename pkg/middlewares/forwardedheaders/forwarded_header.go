package forwardedheaders

import (
	"net"
	"net/http"
	"net/textproto"
	"os"
	"slices"
	"strings"

	"github.com/traefik/traefik/v2/pkg/ip"
	"github.com/traefik/traefik/v2/pkg/middlewares/forwardedheaders/xheaders"
	"golang.org/x/net/http/httpguts"
)

const (
	connection = "Connection"
	upgrade    = "Upgrade"
)

// XForwarded is an HTTP handler wrapper that sets the X-Forwarded headers,
// and other relevant headers for a reverse-proxy.
// Unless insecure is set,
// it first removes all the existing values for those headers if the remote address is not one of the trusted ones.
type XForwarded struct {
	insecure          bool
	trustedIPs        []string
	connectionHeaders []string
	ipChecker         *ip.Checker
	next              http.Handler
	hostname          string
}

// NewXForwarded creates a new XForwarded.
func NewXForwarded(insecure bool, trustedIPs []string, connectionHeaders []string, next http.Handler) (*XForwarded, error) {
	var ipChecker *ip.Checker
	if len(trustedIPs) > 0 {
		var err error
		ipChecker, err = ip.NewChecker(trustedIPs)
		if err != nil {
			return nil, err
		}
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}

	canonicalConnectionHeaders := make([]string, len(connectionHeaders))
	for i, header := range connectionHeaders {
		canonicalConnectionHeaders[i] = http.CanonicalHeaderKey(header)
	}

	return &XForwarded{
		insecure:          insecure,
		trustedIPs:        trustedIPs,
		connectionHeaders: canonicalConnectionHeaders,
		ipChecker:         ipChecker,
		next:              next,
		hostname:          hostname,
	}, nil
}

// removeIPv6Zone removes the zone if the given IP is an ipv6 address and it has {zone} information in it,
// like "[fe80::d806:a55d:eb1b:49cc%vEthernet (vmxnet3 Ethernet Adapter - Virtual Switch)]:64692".
func removeIPv6Zone(clientIP string) string {
	if before, _, found := strings.Cut(clientIP, "%"); found {
		return before
	}
	return clientIP
}

// isWebsocketRequest returns whether the specified HTTP request is a websocket handshake request.
func isWebsocketRequest(req *http.Request) bool {
	containsHeader := func(name, value string) bool {
		h := unsafeHeader(req.Header).Get(name)
		for {
			before, after, found := strings.Cut(h, ",")
			if strings.EqualFold(value, strings.TrimSpace(before)) {
				return true
			}
			if !found {
				return false
			}
			h = after
		}
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

	if unsafeHeader(req.Header).Get(xheaders.ForwardedProto) == "https" || unsafeHeader(req.Header).Get(xheaders.ForwardedProto) == "wss" {
		return "443"
	}

	if req.TLS != nil {
		return "443"
	}

	return "80"
}

// ServeHTTP implements http.Handler.
func (x *XForwarded) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !x.insecure && !x.isTrustedIP(r.RemoteAddr) {
		// Strip X headers and their underscore variants.
		for key := range xHeadersMap {
			delete(r.Header, key)
		}
	}

	x.rewrite(r)

	x.removeConnectionHeaders(r)

	x.next.ServeHTTP(w, r)
}

func (x *XForwarded) isTrustedIP(ip string) bool {
	if x.ipChecker == nil {
		return false
	}
	return x.ipChecker.IsAuthorized(ip) == nil
}

func (x *XForwarded) rewrite(outreq *http.Request) {
	if clientIP, _, err := net.SplitHostPort(outreq.RemoteAddr); err == nil {
		clientIP = removeIPv6Zone(clientIP)

		if unsafeHeader(outreq.Header).Get(xheaders.RealIP) == "" {
			unsafeHeader(outreq.Header).Set(xheaders.RealIP, clientIP)
		}
	}

	xfProto := unsafeHeader(outreq.Header).Get(xheaders.ForwardedProto)
	if xfProto == "" {
		// TODO: is this expected to set the X-Forwarded-Proto header value to
		// ws(s) as the underlying request used to upgrade the connection is
		// made over HTTP(S)?
		if isWebsocketRequest(outreq) {
			if outreq.TLS != nil {
				unsafeHeader(outreq.Header).Set(xheaders.ForwardedProto, "wss")
			} else {
				unsafeHeader(outreq.Header).Set(xheaders.ForwardedProto, "ws")
			}
		} else {
			if outreq.TLS != nil {
				unsafeHeader(outreq.Header).Set(xheaders.ForwardedProto, "https")
			} else {
				unsafeHeader(outreq.Header).Set(xheaders.ForwardedProto, "http")
			}
		}
	}

	if xfPort := unsafeHeader(outreq.Header).Get(xheaders.ForwardedPort); xfPort == "" {
		unsafeHeader(outreq.Header).Set(xheaders.ForwardedPort, forwardedPort(outreq))
	}

	if xfHost := unsafeHeader(outreq.Header).Get(xheaders.ForwardedHost); xfHost == "" && outreq.Host != "" {
		unsafeHeader(outreq.Header).Set(xheaders.ForwardedHost, outreq.Host)
	}

	// Per https://www.rfc-editor.org/rfc/rfc2616#section-4.2, the Forwarded IPs list is in
	// the same order as the values in the X-Forwarded-For header(s).
	if xffs := unsafeHeader(outreq.Header).Values(xheaders.ForwardedFor); len(xffs) > 0 {
		unsafeHeader(outreq.Header).Set(xheaders.ForwardedFor, strings.Join(xffs, ", "))
	}

	if x.hostname != "" {
		unsafeHeader(outreq.Header).Set(xheaders.ForwardedServer, x.hostname)
	}
}

func (x *XForwarded) removeConnectionHeaders(req *http.Request) {
	var reqUpType string
	if httpguts.HeaderValuesContainsToken(req.Header[connection], upgrade) {
		reqUpType = unsafeHeader(req.Header).Get(upgrade)
	}

	var connectionHopByHopHeaders []string
	for _, f := range req.Header[connection] {
		for sf := range strings.SplitSeq(f, ",") {
			if sf = textproto.TrimString(sf); sf != "" {
				key := http.CanonicalHeaderKey(sf)
				// Connection header cannot dictate to remove X- headers managed by Traefik,
				// as per rfc7230 https://datatracker.ietf.org/doc/html/rfc7230#section-6.1,
				// A proxy or gateway MUST ... and then remove the Connection header field itself
				// (or replace it with the intermediary's own connection options for the forwarded message).
				if _, ok := xHeadersMap[key]; ok {
					continue
				}

				// Keep headers allowed through the middleware chain.
				if slices.Contains(x.connectionHeaders, key) {
					connectionHopByHopHeaders = append(connectionHopByHopHeaders, key)
					continue
				}

				// Apply Connection header option.
				delete(req.Header, key)
			}
		}
	}

	if reqUpType != "" {
		connectionHopByHopHeaders = append(connectionHopByHopHeaders, upgrade)
		unsafeHeader(req.Header).Set(upgrade, reqUpType)
	}
	if len(connectionHopByHopHeaders) > 0 {
		unsafeHeader(req.Header).Set(connection, strings.Join(connectionHopByHopHeaders, ","))
		return
	}

	unsafeHeader(req.Header).Del(connection)
}

// unsafeHeader allows to manage Header values.
// Must be used only when the header name is already a canonical key.
type unsafeHeader map[string][]string

func (h unsafeHeader) Set(key, value string) {
	h[key] = []string{value}
}

func (h unsafeHeader) Get(key string) string {
	if len(h[key]) == 0 {
		return ""
	}
	return h[key][0]
}

func (h unsafeHeader) Values(key string) []string {
	return h[key]
}

func (h unsafeHeader) Del(key string) {
	delete(h, key)
}
