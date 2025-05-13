package server

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	ptypes "github.com/traefik/paerser/types"
)

func newKeepAliveMiddleware(next http.Handler, maxRequests int, maxTime ptypes.Duration) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		state, ok := req.Context().Value(connStateKey).(*connState)
		if ok {
			state.HTTPRequestCount++
			if maxRequests > 0 && state.HTTPRequestCount >= maxRequests {
				log.Debug().Msg("Close because of too many requests")
				state.KeepAliveState = "Close because of too many requests"
				rw.Header().Set("Connection", "close")
			}
			if maxTime > 0 && time.Now().After(state.Start.Add(time.Duration(maxTime))) {
				log.Debug().Msg("Close because of too long connection")
				state.KeepAliveState = "Close because of too long connection"
				rw.Header().Set("Connection", "close")
			}
		}
		next.ServeHTTP(rw, req)
	})
}
