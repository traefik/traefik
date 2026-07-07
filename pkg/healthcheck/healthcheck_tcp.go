package healthcheck

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/tcp"
)

// maxPayloadSize is the maximum payload size that can be sent during health checks.
const maxPayloadSize = 65535

type TCPHealthCheckTarget struct {
	Address string
	TLS     bool
	Dialer  tcp.Dialer
}
type ServiceTCPHealthChecker struct {
	balancer StatusSetter
	info     *runtime.TCPServiceInfo

	config            *dynamic.TCPServerHealthCheck
	interval          time.Duration
	unhealthyInterval time.Duration
	timeout           time.Duration
	fails             int
	passes            int

	healthyTargets   chan *TCPHealthCheckTarget
	unhealthyTargets chan *TCPHealthCheckTarget
	targetsMu        sync.Mutex
	targets          map[string]*healthStatus

	serviceName string
}

func NewServiceTCPHealthChecker(ctx context.Context, config *dynamic.TCPServerHealthCheck, service StatusSetter, info *runtime.TCPServiceInfo, targets []TCPHealthCheckTarget, serviceName string) *ServiceTCPHealthChecker {
	logger := log.Ctx(ctx)
	interval := time.Duration(config.Interval)
	if interval <= 0 {
		logger.Error().Msg("Health check interval smaller than zero, default value will be used instead.")
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

	fails := config.Fails
	if fails <= 0 {
		logger.Error().Msg("Health check fails smaller than one, default value will be used instead.")
		fails = 1
	}

	passes := config.Passes
	if passes <= 0 {
		logger.Error().Msg("Health check passes smaller than one, default value will be used instead.")
		passes = 1
	}

	if config.Send != "" && len(config.Send) > maxPayloadSize {
		logger.Error().Msgf("Health check payload size exceeds maximum allowed size of %d bytes, falling back to connect only check.", maxPayloadSize)
		config.Send = ""
	}

	if config.Expect != "" && len(config.Expect) > maxPayloadSize {
		logger.Error().Msgf("Health check expected response size exceeds maximum allowed size of %d bytes, falling back to close without response.", maxPayloadSize)
		config.Expect = ""
	}

	healthyTargets := make(chan *TCPHealthCheckTarget, len(targets))
	targetStatuses := make(map[string]*healthStatus, len(targets))
	for _, target := range targets {
		healthyTargets <- &target
		targetStatuses[target.Address] = &healthStatus{up: true}
	}
	unhealthyTargets := make(chan *TCPHealthCheckTarget, len(targets))

	return &ServiceTCPHealthChecker{
		balancer:          service,
		info:              info,
		config:            config,
		interval:          interval,
		unhealthyInterval: unhealthyInterval,
		timeout:           timeout,
		fails:             fails,
		passes:            passes,
		healthyTargets:    healthyTargets,
		unhealthyTargets:  unhealthyTargets,
		targets:           targetStatuses,
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

				if err := thc.executeHealthCheck(ctx, thc.config, target); err != nil {
					// The context is canceled when the dynamic configuration is refreshed.
					if errors.Is(err, context.Canceled) {
						return
					}

					result := thc.handleHealthCheckResult(ctx, target, false, targets)

					event := log.Ctx(ctx).Warn().
						Str("targetAddress", target.Address).
						Err(err)
					if !result.update {
						event.
							Int("currentFailures", result.count).
							Int("failureThreshold", result.threshold).
							Msg("Health check failed, failure threshold not reached.")
						continue
					}

					event.Msg("Health check failed.")
					continue
				}

				result := thc.handleHealthCheckResult(ctx, target, true, targets)
				if !result.update {
					log.Ctx(ctx).Info().
						Str("targetAddress", target.Address).
						Int("currentPasses", result.count).
						Int("passThreshold", result.threshold).
						Msg("Health check succeeded, pass threshold not reached.")
				}
			}
		}
	}
}

func (thc *ServiceTCPHealthChecker) handleHealthCheckResult(ctx context.Context, target *TCPHealthCheckTarget, up bool, currentTargets chan *TCPHealthCheckTarget) healthStatusResult {
	result := thc.observeTargetStatus(target.Address, up)
	if !result.update {
		currentTargets <- target
		return result
	}

	thc.balancer.SetStatus(ctx, target.Address, up)

	statusStr := runtime.StatusDown
	nextTargets := thc.unhealthyTargets
	if up {
		statusStr = runtime.StatusUp
		nextTargets = thc.healthyTargets
	}

	nextTargets <- target
	thc.info.UpdateServerStatus(target.Address, statusStr)

	// TODO: add a TCP server up metric (like for HTTP).

	return result
}

func (thc *ServiceTCPHealthChecker) observeTargetStatus(address string, up bool) healthStatusResult {
	thc.targetsMu.Lock()
	defer thc.targetsMu.Unlock()

	if thc.targets == nil {
		thc.targets = make(map[string]*healthStatus)
	}

	status := thc.targets[address]
	if status == nil {
		status = &healthStatus{up: true}
		thc.targets[address] = status
	}

	return status.observe(up, thc.fails, thc.passes)
}

func (thc *ServiceTCPHealthChecker) executeHealthCheck(ctx context.Context, config *dynamic.TCPServerHealthCheck, target *TCPHealthCheckTarget) error {
	addr := target.Address
	if config.Port != 0 {
		host, _, err := net.SplitHostPort(target.Address)
		if err != nil {
			return fmt.Errorf("parsing address %q: %w", target.Address, err)
		}

		addr = net.JoinHostPort(host, strconv.Itoa(config.Port))
	}

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Duration(config.Timeout)))
	defer cancel()

	conn, err := target.Dialer.DialContext(ctx, "tcp", addr, nil)
	if err != nil {
		return fmt.Errorf("connecting to %s: %w", addr, err)
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(thc.timeout)); err != nil {
		return fmt.Errorf("setting timeout to %s: %w", thc.timeout, err)
	}

	if config.Send != "" {
		if _, err = conn.Write([]byte(config.Send)); err != nil {
			return fmt.Errorf("sending to %s: %w", addr, err)
		}
	}

	if config.Expect != "" {
		buf := make([]byte, len(config.Expect))
		if _, err = conn.Read(buf); err != nil {
			return fmt.Errorf("reading from %s: %w", addr, err)
		}

		if string(buf) != config.Expect {
			return errors.New("unexpected heath check response")
		}
	}

	return nil
}
