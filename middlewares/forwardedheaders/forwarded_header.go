package forwardedheaders

import (
	"net/http"

	"github.com/containous/traefik/ip"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/utils"
)

// XForwarded filter for XForwarded headers
type XForwarded struct {
	insecure   bool
	trustedIps []string
	ipChecker  *ip.Checker
}

// NewXforwarded creates a new XForwarded
func NewXforwarded(insecure bool, trustedIps []string) (*XForwarded, error) {
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
	}, nil
}

func (x *XForwarded) isTrustedIP(ip string) bool {
	if x.ipChecker == nil {
		return false
	}
	return x.ipChecker.IsAuthorized(ip) == nil
}

func (x *XForwarded) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if !x.insecure && !x.isTrustedIP(r.RemoteAddr) {
		utils.RemoveHeaders(r.Header, forward.XHeaders...)
	}

	// If there is a next, call it.
	if next != nil {
		next(w, r)
	}
}
