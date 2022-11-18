//go:build !windows
// +build !windows

package server

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
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
				log.Info().Msgf("Closing and re-opening log files for rotation: %+v", sig)

				if s.accessLoggerMiddleware != nil {
					if err := s.accessLoggerMiddleware.Rotate(); err != nil {
						log.Error().Err(err).Msg("Error rotating access log")
					}
				}
			}
		}
	}
}
