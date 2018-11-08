package server

import (
	"net/http"
	"os"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/whitelist"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/utils"
)

const (
	xForwardedURI    = "X-Forwarded-Uri"
	xForwardedMethod = "X-Forwarded-Method"
)

// NewHeaderRewriter Create a header rewriter
func NewHeaderRewriter(trustedIPs []string, insecure bool) (forward.ReqRewriter, error) {
	ips, err := whitelist.NewIP(trustedIPs, insecure, true)
	if err != nil {
		return nil, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}

	return &headerRewriter{
		secureRewriter:   &forward.HeaderRewriter{TrustForwardHeader: false, Hostname: hostname},
		insecureRewriter: &forward.HeaderRewriter{TrustForwardHeader: true, Hostname: hostname},
		ips:              ips,
		insecure:         insecure,
	}, nil
}

type headerRewriter struct {
	secureRewriter   forward.ReqRewriter
	insecureRewriter forward.ReqRewriter
	insecure         bool
	ips              *whitelist.IP
}

func (h *headerRewriter) Rewrite(req *http.Request) {
	if h.insecure {
		h.insecureRewriter.Rewrite(req)
		return
	}

	err := h.ips.IsAuthorized(req)
	if err != nil {
		log.Debug(err)
		// Remove additional X-Forwarded Headers which are used by the forward authentication
		utils.RemoveHeaders(req.Header, xForwardedURI, xForwardedMethod)
		h.secureRewriter.Rewrite(req)
		return
	}

	h.insecureRewriter.Rewrite(req)
}
