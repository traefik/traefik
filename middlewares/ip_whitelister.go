package middlewares

import (
	"fmt"
	"net"
	"net/http"

	"github.com/containous/traefik/log"
	"github.com/pkg/errors"
	"github.com/urfave/negroni"
)

// IPWhitelister is a middleware that provides Checks of the Requesting IP against a set of Whitelists
type IPWhitelister struct {
	handler    negroni.Handler
	whitelists []*net.IPNet
}

// NewIPWhitelister builds a new IPWhitelister given a list of CIDR-Strings to whitelist
func NewIPWhitelister(whitelistStrings []string) (*IPWhitelister, error) {

	if len(whitelistStrings) == 0 {
		return nil, errors.New("no whitelists provided")
	}

	whitelister := IPWhitelister{}

	for _, whitelistString := range whitelistStrings {
		_, whitelist, err := net.ParseCIDR(whitelistString)
		if err != nil {
			return nil, fmt.Errorf("parsing CIDR whitelist %s: %v", whitelist, err)
		}
		whitelister.whitelists = append(whitelister.whitelists, whitelist)
	}

	whitelister.handler = negroni.HandlerFunc(whitelister.handle)
	log.Debugf("configured %u IP whitelists: %s", len(whitelister.whitelists), whitelister.whitelists)

	return &whitelister, nil
}

func (whitelister *IPWhitelister) handle(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	remoteIP, err := ipFromRemoteAddr(r.RemoteAddr)
	if err != nil {
		log.Warnf("unable to parse remote-address from header: %s - rejecting", r.RemoteAddr)
		reject(w)
		return
	}

	for _, whitelist := range whitelister.whitelists {
		if whitelist.Contains(*remoteIP) {
			log.Debugf("source-IP %s matched whitelist %s - passing", remoteIP, whitelist)
			next.ServeHTTP(w, r)
			return
		}
	}

	log.Debugf("source-IP %s matched none of the whitelists - rejecting", remoteIP)
	reject(w)
}

func reject(w http.ResponseWriter) {
	statusCode := http.StatusForbidden

	w.WriteHeader(statusCode)
	w.Write([]byte(http.StatusText(statusCode)))
}

func ipFromRemoteAddr(addr string) (*net.IP, error) {
	ip, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("can't extract IP/Port from address %s: %s", addr, err)
	}

	userIP := net.ParseIP(ip)
	if userIP == nil {
		return nil, fmt.Errorf("can't parse IP from address %s", ip)
	}

	return &userIP, nil
}

func (whitelister *IPWhitelister) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	whitelister.handler.ServeHTTP(rw, r, next)
}
