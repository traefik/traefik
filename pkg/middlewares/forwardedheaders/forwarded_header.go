package forwardedheaders

import (
	"net/http"

	"github.com/containous/traefik/pkg/ip"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/utils"
)

// XForwarded filter for XForwarded headers.
type XForwarded struct {
	insecure   bool
	trustedIps []string
	ipChecker  *ip.Checker
	next       http.Handler
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

	return &XForwarded{
		insecure:   insecure,
		trustedIps: trustedIps,
		ipChecker:  ipChecker,
		next:       next,
	}, nil
}

func (x *XForwarded) isTrustedIP(ip string) bool {
	if x.ipChecker == nil {
		return false
	}
	return x.ipChecker.IsAuthorized(ip) == nil
}

func (x *XForwarded) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !x.insecure && !x.isTrustedIP(r.RemoteAddr) {
		utils.RemoveHeaders(r.Header, forward.XHeaders...)
	}

	x.next.ServeHTTP(w, r)
}
