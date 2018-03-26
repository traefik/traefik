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
	IPs, err := whitelist.NewIP(trustedIPs, insecure, true)
	if err != nil {
		return nil, err
	}

	h, err := os.Hostname()
	if err != nil {
		h = "localhost"
	}

	return &headerRewriter{
		secureRewriter:   &forward.HeaderRewriter{TrustForwardHeader: true, Hostname: h},
		insecureRewriter: &forward.HeaderRewriter{TrustForwardHeader: false, Hostname: h},
		ips:              IPs,
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
	authorized, _, err := h.ips.IsAuthorized(req)
	if err != nil {
		log.Error(err)
		h.secureRewriter.Rewrite(req)
		return
	}

	if h.insecure || authorized {
		h.secureRewriter.Rewrite(req)
	} else {
		h.insecureRewriter.Rewrite(req)
	}
}
