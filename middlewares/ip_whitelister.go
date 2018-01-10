package middlewares

import (
	"fmt"
	"net"
	"net/http"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/middlewares/tracing"
	"github.com/containous/traefik/whitelist"
	"github.com/pkg/errors"
	"github.com/urfave/negroni"
)

// IPWhiteLister is a middleware that provides Checks of the Requesting IP against a set of Whitelists
type IPWhiteLister struct {
	handler     negroni.Handler
	whiteLister *whitelist.IP
}

// NewIPWhitelister builds a new IPWhiteLister given a list of CIDR-Strings to whitelist
func NewIPWhitelister(whitelistStrings []string) (*IPWhiteLister, error) {

	if len(whitelistStrings) == 0 {
		return nil, errors.New("no whitelists provided")
	}

	whiteLister := IPWhiteLister{}

	ip, err := whitelist.NewIP(whitelistStrings, false)
	if err != nil {
		return nil, fmt.Errorf("parsing CIDR whitelist %s: %v", whitelistStrings, err)
	}
	whiteLister.whiteLister = ip

	whiteLister.handler = negroni.HandlerFunc(whiteLister.handle)
	log.Debugf("configured %u IP whitelists: %s", len(whitelistStrings), whitelistStrings)

	return &whiteLister, nil
}

func (wl *IPWhiteLister) handle(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	ipAddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		tracing.SetErrorAndWarnLog(r, "unable to parse remote-address from header: %s - rejecting", r.RemoteAddr)
		reject(w)
		return
	}

	allowed, ip, err := wl.whiteLister.Contains(ipAddress)
	if err != nil {
		tracing.SetErrorAndDebugLog(r, "source-IP %s matched none of the whitelists - rejecting", ipAddress)
		reject(w)
		return
	}

	if allowed {
		tracing.SetErrorAndDebugLog(r, "source-IP %s matched whitelist %s - passing", ipAddress, wl.whiteLister)
		next.ServeHTTP(w, r)
		return
	}

	tracing.SetErrorAndDebugLog(r, "source-IP %s matched none of the whitelists - rejecting", ip)
	reject(w)
}

func (wl *IPWhiteLister) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	wl.handler.ServeHTTP(rw, r, next)
}

func reject(w http.ResponseWriter) {
	statusCode := http.StatusForbidden

	w.WriteHeader(statusCode)
	w.Write([]byte(http.StatusText(statusCode)))
}
