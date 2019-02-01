package middlewares

import (
	"fmt"
	"net/http"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/middlewares/tracing"
	"github.com/containous/traefik/whitelist"
	"github.com/pkg/errors"
	"github.com/urfave/negroni"
)

// IPWhiteLister is a middleware that provides Checks of the Requesting IP against a set of Whitelists
type IPBlackLister struct {
	handler     negroni.Handler
	blackLister *whitelist.IP
}

// NewIPBlackLister builds a new IPBlackLister given a list of CIDR-Strings to whitelist
func NewIPBlackLister(blackList []string, useXForwardedFor bool) (*IPBlackLister, error) {
	if len(blackList) == 0 {
		return nil, errors.New("no black list provided")
	}

	blackLister := IPBlackLister{}

	ip, err := whitelist.NewIP(blackList, false, useXForwardedFor)
	if err != nil {
		return nil, fmt.Errorf("parsing CIDR blacklist %s: %v", blackList, err)
	}
	blackLister.blackLister = ip

	blackLister.handler = negroni.HandlerFunc(blackLister.handle)
	log.Debugf("configured IP blacklist: %s", blackList)

	return &blackLister, nil
}

func (bl *IPBlackLister) handle(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	err := bl.blackLister.ContainsReq(r)
	if err == nil {
		tracing.SetErrorAndDebugLog(r, "request %+v - rejecting: %v", r, err)
		whitelist.Reject(w)
		return
	}

	tracing.SetErrorAndDebugLog(r, "request %+v matched blacklist %v - passing", r, bl.blackLister)
	next.ServeHTTP(w, r)
}

func (bl *IPBlackLister) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	bl.handler.ServeHTTP(rw, r, next)
}
