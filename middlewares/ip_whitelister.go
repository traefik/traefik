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
type IPWhiteLister struct {
	handler     negroni.Handler
	whiteLister *whitelist.IP
}

// NewIPWhiteLister builds a new IPWhiteLister given a list of CIDR-Strings to whitelist
func NewIPWhiteLister(whiteList []string, useXForwardedFor bool) (*IPWhiteLister, error) {
	if len(whiteList) == 0 {
		return nil, errors.New("no white list provided")
	}

	whiteLister := IPWhiteLister{}

	ip, err := whitelist.NewIP(whiteList, false, useXForwardedFor)
	if err != nil {
		return nil, fmt.Errorf("parsing CIDR whitelist %s: %v", whiteList, err)
	}
	whiteLister.whiteLister = ip

	whiteLister.handler = negroni.HandlerFunc(whiteLister.handle)
	log.Debugf("configured IP white list: %s", whiteList)

	return &whiteLister, nil
}

func (wl *IPWhiteLister) handle(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	err := wl.whiteLister.IsAuthorized(r)
	if err != nil {
		tracing.SetErrorAndDebugLog(r, "request %+v - rejecting: %v", r, err)
		reject(w)
		return
	}

	tracing.SetErrorAndDebugLog(r, "request %+v matched white list %v - passing", r, wl.whiteLister)
	next.ServeHTTP(w, r)
}

func (wl *IPWhiteLister) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	wl.handler.ServeHTTP(rw, r, next)
}

func reject(w http.ResponseWriter) {
	statusCode := http.StatusForbidden

	w.WriteHeader(statusCode)
	_, err := w.Write([]byte(http.StatusText(statusCode)))
	if err != nil {
		log.Error(err)
	}
}
