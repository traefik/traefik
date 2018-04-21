package server

import (
	"net/http"
	"os"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/whitelist"
	"github.com/vulcand/oxy/forward"
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
	authorized, err := h.ips.IsAuthorized(req)
	if err != nil {
		log.Error(err)
		h.insecureRewriter.Rewrite(req)
		return
	}

	if h.insecure || authorized {
		h.insecureRewriter.Rewrite(req)
	} else {
		h.secureRewriter.Rewrite(req)
	}
}
