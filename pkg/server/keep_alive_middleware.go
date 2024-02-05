package server

import (
	"net/http"
	"time"

	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v2/pkg/log"
)

func newKeepAliveMiddleware(next http.Handler, maxRequests int, maxTime ptypes.Duration) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		state, ok := req.Context().Value(connStateKey).(*connState)
		if ok {
			state.HTTPRequestCount++
			if maxRequests > 0 && state.HTTPRequestCount >= maxRequests {
				log.WithoutContext().Debug("Close because of too many requests")
				state.KeepAliveState = "Close because of too many requests"
				rw.Header().Set("Connection", "close")
			}
			if maxTime > 0 && time.Now().After(state.Start.Add(time.Duration(maxTime))) {
				log.WithoutContext().Debug("Close because of too long connection")
				state.KeepAliveState = "Close because of too long connection"
				rw.Header().Set("Connection", "close")
			}
		}
		next.ServeHTTP(rw, req)
	})
}
