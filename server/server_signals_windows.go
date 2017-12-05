// +build windows

package server

import (
	"os/signal"
	"syscall"

	"github.com/containous/traefik/log"
)

func (s *Server) configureSignals() {
	signal.Notify(s.signals, syscall.SIGINT, syscall.SIGTERM)
}

func (s *Server) listenSignals() {
	for {
		sig := <-s.signals
		switch sig {
		default:
			log.Infof("I have to go... %+v", sig)
			log.Info("Stopping server")
			s.Stop()
		}
	}
}
