/*
Copyright
*/
package middlewares

import (
	log "github.com/Sirupsen/logrus"
	"github.com/mailgun/oxy/roundrobin"
	"net/http"
	"strings"
	"time"
)

type WebsocketUpgrader struct {
	rr *roundrobin.RoundRobin
}

func NewWebsocketUpgrader(rr *roundrobin.RoundRobin) *WebsocketUpgrader {
	wu := WebsocketUpgrader{
		rr: rr,
	}
	return &wu
}

func (u *WebsocketUpgrader) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// If request is websocket, serve with golang websocket server to do protocol handshake
	if strings.Join(req.Header["Upgrade"], "") == "websocket" {
		start := time.Now().UTC()
		url, err := u.rr.NextServer()
		if err != nil {
			log.Errorf("Can't round robin in websocket middleware")
			return
		}
		log.Debugf("Websocket forward to %s", url.String())
		NewProxy(url).ServeHTTP(w, req)

		if req.TLS != nil {
			log.Debugf("Round trip: %v, duration: %v tls:version: %x, tls:resume:%t, tls:csuite:%x, tls:server:%v",
				req.URL, time.Now().UTC().Sub(start),
				req.TLS.Version,
				req.TLS.DidResume,
				req.TLS.CipherSuite,
				req.TLS.ServerName)
		} else {
			log.Debugf("Round trip: %v, duration: %v",
				req.URL, time.Now().UTC().Sub(start))
		}

		return
	}
	u.rr.ServeHTTP(w, req)
}
