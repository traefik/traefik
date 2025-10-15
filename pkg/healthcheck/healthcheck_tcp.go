package healthcheck

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/tcp"
)

const (
	MaxPayloadSize = 65535 // Maximum size of the payload to send during health checks.
)

type ServiceTCPHealthChecker struct {
	dialerManager *tcp.DialerManager
	balancer      StatusSetter
	info          *runtime.TCPServiceInfo

	config            *dynamic.TCPServerHealthCheck
	interval          time.Duration
	unhealthyInterval time.Duration
	timeout           time.Duration

	healthyTargets   chan *TCPHealthCheckTarget
	unhealthyTargets chan *TCPHealthCheckTarget

	serviceName string
}

type TCPHealthCheckTarget struct {
	Address string
	TLS     bool
	Dialer  tcp.Dialer
}

func NewServiceTCPHealthChecker(ctx context.Context, config *dynamic.TCPServerHealthCheck, service StatusSetter, info *runtime.TCPServiceInfo, targets []TCPHealthCheckTarget, serviceName string) *ServiceTCPHealthChecker {
	logger := log.Ctx(ctx)
	interval := time.Duration(config.Interval)
	if interval <= 0 {
		logger.Error().Msg("Health check interval smaller than zero")
		interval = time.Duration(dynamic.DefaultHealthCheckInterval)
	}

	// If the unhealthyInterval option is not set, we use the interval option value,
	// to check the unhealthy targets as often as the healthy ones.
	var unhealthyInterval time.Duration
	if config.UnhealthyInterval == nil {
		unhealthyInterval = interval
	} else {
		unhealthyInterval = time.Duration(*config.UnhealthyInterval)
		if unhealthyInterval <= 0 {
			logger.Error().Msg("Health check unhealthy interval smaller than zero, default value will be used instead.")
			unhealthyInterval = time.Duration(dynamic.DefaultHealthCheckInterval)
		}
	}

	timeout := time.Duration(config.Timeout)
	if timeout <= 0 {
		logger.Error().Msg("Health check timeout smaller than zero, default value will be used instead.")
		timeout = time.Duration(dynamic.DefaultHealthCheckTimeout)
	}

	if config.Send != "" && len(config.Send) > MaxPayloadSize {
		logger.Error().Msgf("Health check payload size exceeds maximum allowed size of %d bytes, falling back to connect only check.", MaxPayloadSize)
		config.Send = ""
	}

	if config.Expect != "" && len(config.Expect) > MaxPayloadSize {
		logger.Error().Msgf("Health check expected response size exceeds maximum allowed size of %d bytes, falling back to close without response.", MaxPayloadSize)
		config.Expect = ""
	}

	healthyTargets := make(chan *TCPHealthCheckTarget, len(targets))
	for _, target := range targets {
		healthyTargets <- &target
	}
	unhealthyTargets := make(chan *TCPHealthCheckTarget, len(targets))

	return &ServiceTCPHealthChecker{
		balancer:          service,
		info:              info,
		config:            config,
		interval:          interval,
		unhealthyInterval: unhealthyInterval,
		timeout:           timeout,
		healthyTargets:    healthyTargets,
		unhealthyTargets:  unhealthyTargets,
		serviceName:       serviceName,
	}
}

func (thc *ServiceTCPHealthChecker) Launch(ctx context.Context) {
	go thc.healthcheck(ctx, thc.unhealthyTargets, thc.unhealthyInterval)

	thc.healthcheck(ctx, thc.healthyTargets, thc.interval)
}

func (thc *ServiceTCPHealthChecker) healthcheck(ctx context.Context, targets chan *TCPHealthCheckTarget, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			// We collect the targets to check once for all,
			// to avoid rechecking a target that has been moved during the health check.
			var targetsToCheck []*TCPHealthCheckTarget
			hasMoreTargets := true
			for hasMoreTargets {
				select {
				case <-ctx.Done():
					return
				case target := <-targets:
					targetsToCheck = append(targetsToCheck, target)
				default:
					hasMoreTargets = false
				}
			}

			// Now we can check the targets.
			for _, target := range targetsToCheck {
				select {
				case <-ctx.Done():
					return
				default:
				}

				up := true

				if err := thc.executeHealthCheck(ctx, thc.config, target); err != nil {
					// The context is canceled when the dynamic configuration is refreshed.
					if errors.Is(err, context.Canceled) {
						return
					}

					log.Ctx(ctx).Warn().
						Str("targetAddress", target.Address).
						Err(err).
						Msg("Health check failed.")

					up = false
				}

				thc.balancer.SetStatus(ctx, target.Address, up)

				var statusStr string
				if up {
					statusStr = runtime.StatusUp
					thc.healthyTargets <- target
				} else {
					statusStr = runtime.StatusDown
					thc.unhealthyTargets <- target
				}

				thc.info.UpdateServerStatus(target.Address, statusStr)

				// TODO: add a TCP server up metric (like for HTTP)
			}
		}
	}
}

func (thc *ServiceTCPHealthChecker) executeHealthCheck(ctx context.Context, config *dynamic.TCPServerHealthCheck, target *TCPHealthCheckTarget) error {
	conn, err := target.Dialer.Dial("tcp", target.Address, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("Failed to close health check connection")
		}
	}()

	if config.Send != "" {
		_, err = conn.Write([]byte(config.Send))
		if err != nil {
			return err
		}
	}

	if config.Expect != "" {
		err := conn.SetReadDeadline(time.Now().Add(thc.timeout))
		if err != nil {
			return err
		}

		buf := make([]byte, len(config.Expect))
		_, err = conn.Read(buf)
		if err != nil {
			return err
		}

		if string(buf) != config.Expect {
			return errors.New("unexpected response")
		}
	}

	return nil
}
