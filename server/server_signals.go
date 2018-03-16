// +build !windows

package server

import (
	"os/signal"
	"syscall"
	"time"

	"github.com/containous/traefik/log"
)

func (s *Server) configureSignals() {
	signal.Notify(s.signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)
}

func (s *Server) listenSignals() {
	for {
		sig := <-s.signals
		switch sig {
		case syscall.SIGUSR1:
			log.Infof("Closing and re-opening log files for rotation: %+v", sig)

			if s.accessLoggerMiddleware != nil {
				if err := s.accessLoggerMiddleware.Rotate(); err != nil {
					log.Errorf("Error rotating access log: %s", err)
				}
			}

			if err := log.RotateFile(); err != nil {
				log.Errorf("Error rotating traefik log: %s", err)
			}
		default:
			log.Infof("I have to go... %+v", sig)

			reqAcceptGraceTimeOut := time.Duration(s.globalConfiguration.LifeCycle.RequestAcceptGraceTimeout)
			if s.globalConfiguration.Ping != nil && reqAcceptGraceTimeOut > 0 {
				s.globalConfiguration.Ping.SetTerminating()
			}
			if reqAcceptGraceTimeOut > 0 {
				log.Infof("Waiting %s for incoming requests to cease", reqAcceptGraceTimeOut)
				time.Sleep(reqAcceptGraceTimeOut)
			}
			log.Info("Stopping server gracefully")
			s.Stop()
		}
	}
}
