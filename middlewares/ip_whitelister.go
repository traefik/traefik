package middlewares

import (
	"fmt"
	"net/http"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/whitelist"
	"github.com/pkg/errors"
	"github.com/urfave/negroni"
)

// IPWhiteLister is a middleware that provides Checks of the Requesting IP against a set of Whitelists
type IPWhiteLister struct {
	handler     negroni.Handler
	whiteLister *whitelist.IP
	trustProxy  *whitelist.IP
}

// NewIPWhitelister builds a new IPWhiteLister given a list of CIDR-Strings to whitelist
func NewIPWhitelister(whitelistStrings []string, whitelistTrustProxy []string) (*IPWhiteLister, error) {

	if len(whitelistStrings) == 0 {
		return nil, errors.New("no whitelists provided")
	}

	whiteLister := IPWhiteLister{}

	ip, err := whitelist.NewIP(whitelistStrings, false)
	if err != nil {
		return nil, fmt.Errorf("parsing CIDR whitelist %s: %v", whitelistStrings, err)
	}
	whiteLister.whiteLister = ip

	if len(whitelistTrustProxy) > 0 {
		ip, err := whitelist.NewIP(whitelistTrustProxy, false)
		if err != nil {
			return nil, fmt.Errorf("parsing CIDR whitelist %s: %v", whitelistTrustProxy, err)
		}
		whiteLister.trustProxy = ip
	}

	whiteLister.handler = negroni.HandlerFunc(whiteLister.handle)
	log.Debugf("configured %u IP whitelists: %s", len(whitelistStrings), whitelistStrings)

	return &whiteLister, nil
}

func (wl *IPWhiteLister) handle(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if ip, err := whitelist.GetRemoteIp(r, wl.trustProxy); err == nil {
		allowed, _ := wl.whiteLister.ContainsIP(ip)
		if allowed {
			log.Debugf("source-IP %s matched whitelist %s - passing", ip.String(), wl.whiteLister)
			next.ServeHTTP(w, r)
			return
		}
		log.Debugf("source-IP %s matched none of the whitelists - rejecting", ip.String())
		reject(w)
	} else {
		log.Debugf("source-IP %s matched none of the whitelists - rejecting", r.RemoteAddr)
		reject(w)
	}
}

func (wl *IPWhiteLister) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	wl.handler.ServeHTTP(rw, r, next)
}

func reject(w http.ResponseWriter) {
	statusCode := http.StatusForbidden

	w.WriteHeader(statusCode)
	w.Write([]byte(http.StatusText(statusCode)))
}
