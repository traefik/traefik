package server

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/network"
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

	// networks gives the entryPoints what the host network does not.
	// It is nil when Traefik serves on the host network only.
	networks network.Extension

	signals  chan os.Signal
	stopChan chan bool

	routinesPool *safe.Pool
}

// NewServer returns an initialized Server.
// The networks extension is the one the entryPoints were given, whose lifecycle the server drives;
// it is nil when there is none.
func NewServer(routinesPool *safe.Pool, entryPoints TCPEntryPoints, entryPointsUDP UDPEntryPoints, watcher *ConfigurationWatcher, observabilityMgr *middleware.ObservabilityMgr, networks network.Extension) *Server {
	srv := &Server{
		watcher:          watcher,
		tcpEntryPoints:   entryPoints,
		observabilityMgr: observabilityMgr,
		signals:          make(chan os.Signal, 1),
		stopChan:         make(chan bool, 1),
		routinesPool:     routinesPool,
		udpEntryPoints:   entryPointsUDP,
		networks:         networks,
	}

	srv.configureSignals()

	return srv
}

// networksWithdrawTimeout bounds the withdrawal of the networks extension that starts a shutdown.
// The entryPoints do not start draining until it returns, so it is kept short.
const networksWithdrawTimeout = 5 * time.Second

// Start starts the server and Stop/Close it when context is Done.
func (s *Server) Start(ctx context.Context) {
	go func() {
		<-ctx.Done()
		logger := log.Ctx(ctx)
		logger.Info().Msg("I have to go...")

		if s.networks != nil {
			// First, before any entryPoint stops accepting:
			// the extension stops attracting new connections while this instance still answers,
			// so that peers have moved elsewhere by the time the listeners close.
			withdrawCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), networksWithdrawTimeout)
			s.networks.Withdraw(withdrawCtx)
			cancel()
		}

		logger.Info().Msg("Stopping server gracefully")
		s.Stop()
	}()

	if s.networks != nil {
		// Before the entryPoints: what the extension brings up may be what they accept on.
		s.networks.Start(ctx)
	}

	s.tcpEntryPoints.Start()
	s.udpEntryPoints.Start()
	s.watcher.Start()

	s.routinesPool.GoCtx(s.listenSignals)
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

	if s.networks != nil {
		// Last: the extension backs the entryPoint listeners that have just been shut down.
		s.networks.Close()
	}

	cancel()
}

func stopMetricsClients() {
	metrics.StopDatadog()
	metrics.StopStatsd()
	metrics.StopInfluxDB2()
	metrics.StopOpenTelemetry()
}
