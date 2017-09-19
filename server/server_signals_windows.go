// +build windows

package server

import (
	"os/signal"
	"syscall"

	"github.com/containous/traefik/log"
)

func (server *Server) configureSignals() {
	signal.Notify(server.signals, syscall.SIGINT, syscall.SIGTERM)
}

func (server *Server) listenSignals() {
	for {
		sig := <-server.signals
		switch sig {
		default:
			log.Infof("I have to go... %+v", sig)
			log.Info("Stopping server")
			server.Stop()
		}
	}
}
