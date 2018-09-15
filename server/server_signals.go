// +build !windows

package server

import (
	"os/signal"
	"syscall"

	"github.com/containous/traefik/log"
)

func (s *Server) configureSignals() {
	signal.Notify(s.signals, syscall.SIGUSR1)
}

func (s *Server) listenSignals(stop chan bool) {
	for {
		select {
		case <-stop:
			return
		case sig := <-s.signals:
			switch sig {
			case syscall.SIGUSR1:
				log.Infof("Closing and re-opening log files for rotation: %+v", sig)

				if s.accessLoggerMiddleware != nil {
					if err := s.accessLoggerMiddleware.Rotate(); err != nil {
						log.Errorf("Error rotating access log: %v", err)
					}
				}

				if err := log.RotateFile(); err != nil {
					log.Errorf("Error rotating traefik log: %v", err)
				}
			}
		}
	}
}
