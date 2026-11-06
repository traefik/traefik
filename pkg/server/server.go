package server

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/observability/metrics"
	"github.com/traefik/traefik/v3/pkg/safe"
	"github.com/traefik/traefik/v3/pkg/server/middleware"
)

// Server is the reverse-proxy/load-balancer engine.
type Server struct {
	watcher          *ConfigurationWatcher
	tcpEntryPoints   TCPEntryPoints
	udpEntryPoints   UDPEntryPoints
	observabilityMgr *middleware.ObservabilityMgr

	signals  chan os.Signal
	stopChan chan bool

	routinesPool *safe.Pool
}

// NewServer returns an initialized Server.
func NewServer(routinesPool *safe.Pool, entryPoints TCPEntryPoints, entryPointsUDP UDPEntryPoints, watcher *ConfigurationWatcher, observabilityMgr *middleware.ObservabilityMgr) *Server {
	srv := &Server{
		watcher:          watcher,
		tcpEntryPoints:   entryPoints,
		observabilityMgr: observabilityMgr,
		signals:          make(chan os.Signal, 1),
		stopChan:         make(chan bool, 1),
		routinesPool:     routinesPool,
		udpEntryPoints:   entryPointsUDP,
	}

	srv.configureSignals()

	return srv
}

// Start starts the server and Stop/Close it when context is Done.
func (s *Server) Start(ctx context.Context) error {
	s.watcher.Start()
	if err := s.watcher.WaitInitialConfiguration(ctx); err != nil {
		s.closeUnstartedEntryPoints()
		return err
	}
	if err := ctx.Err(); err != nil {
		s.closeUnstartedEntryPoints()
		return err
	}

	go func() {
		<-ctx.Done()
		logger := log.Ctx(ctx)
		logger.Info().Msg("I have to go...")
		logger.Info().Msg("Stopping server gracefully")
		s.Stop()
	}()

	s.tcpEntryPoints.Start()
	s.udpEntryPoints.Start()

	s.routinesPool.GoCtx(s.listenSignals)
	return nil
}

// Wait blocks until the server shutdown.
func (s *Server) Wait() {
	<-s.stopChan
}

// Stop stops the server.
func (s *Server) Stop() {
	defer log.Info().Msg("Server stopped")

	s.tcpEntryPoints.Stop()
	s.udpEntryPoints.Stop()

	s.stopChan <- true
}

// Close destroys the server.
func (s *Server) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	go func(ctx context.Context) {
		<-ctx.Done()
		if errors.Is(ctx.Err(), context.Canceled) {
			return
		} else if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			panic("Timeout while stopping traefik, killing instance ✝")
		}
	}(ctx)

	stopMetricsClients()

	s.routinesPool.Stop()

	signal.Stop(s.signals)
	close(s.signals)

	close(s.stopChan)

	s.observabilityMgr.Close()

	cancel()
}

// Construction binds sockets and starts HTTP forwarders before the accept loops.
// Startup failure must release these without entering graceful serving shutdown.
func (s *Server) closeUnstartedEntryPoints() {
	for _, ep := range s.tcpEntryPoints {
		_ = ep.listener.Close()
		_ = ep.httpServer.Server.Close()
		_ = ep.httpsServer.Server.Close()
		if ep.http3Server != nil {
			_ = ep.http3Server.Close()
		}
		ep.tracker.Close()
	}
	for _, ep := range s.udpEntryPoints {
		_ = ep.listener.Close()
	}
}

func stopMetricsClients() {
	metrics.StopDatadog()
	metrics.StopStatsd()
	metrics.StopInfluxDB2()
	metrics.StopOpenTelemetry()
}
