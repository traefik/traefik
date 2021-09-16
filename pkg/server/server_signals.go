//go:build !windows
// +build !windows

package server

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/traefik/traefik/v2/pkg/log"
)

func (s *Server) configureSignals() {
	signal.Notify(s.signals, syscall.SIGUSR1)
}

func (s *Server) listenSignals(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case sig := <-s.signals:
			if sig == syscall.SIGUSR1 {
				log.WithoutContext().Infof("Closing and re-opening log files for rotation: %+v", sig)

				if s.accessLoggerMiddleware != nil {
					if err := s.accessLoggerMiddleware.Rotate(); err != nil {
						log.WithoutContext().Errorf("Error rotating access log: %v", err)
					}
				}

				if err := log.RotateFile(); err != nil {
					log.WithoutContext().Errorf("Error rotating traefik log: %v", err)
				}
			}
		}
	}
}
