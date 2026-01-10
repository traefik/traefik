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

	healthyTargets   chan *TCPHealthCheckTarget
	unhealthyTargets chan *TCPHealthCheckTarget

	serviceName string

	// Threshold configuration
	failsThreshold  int
	passesThreshold int

	// State tracking
	stateMu sync.RWMutex
	states  map[string]*targetState
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

	if config.Send != "" && len(config.Send) > maxPayloadSize {
		logger.Error().Msgf("Health check payload size exceeds maximum allowed size of %d bytes, falling back to connect only check.", maxPayloadSize)
		config.Send = ""
	}

	if config.Expect != "" && len(config.Expect) > maxPayloadSize {
		logger.Error().Msgf("Health check expected response size exceeds maximum allowed size of %d bytes, falling back to close without response.", maxPayloadSize)
		config.Expect = ""
	}

	healthyTargets := make(chan *TCPHealthCheckTarget, len(targets))
	states := make(map[string]*targetState, len(targets))
	for _, target := range targets {
		healthyTargets <- &target
		// All targets start as healthy
		states[target.Address] = &targetState{healthy: true}
	}
	unhealthyTargets := make(chan *TCPHealthCheckTarget, len(targets))

	// Use default thresholds if not configured (0 means use default of 1)
	failsThreshold := config.FailsThreshold
	if failsThreshold <= 0 {
		failsThreshold = 1
	}
	passesThreshold := config.PassesThreshold
	if passesThreshold <= 0 {
		passesThreshold = 1
	}

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
		failsThreshold:    failsThreshold,
		passesThreshold:   passesThreshold,
		states:            states,
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

				checkPassed := true

				if err := thc.executeHealthCheck(ctx, thc.config, target); err != nil {
					// The context is canceled when the dynamic configuration is refreshed.
					if errors.Is(err, context.Canceled) {
						return
					}

					log.Ctx(ctx).Warn().
						Str("targetAddress", target.Address).
						Err(err).
						Msg("Health check failed.")

					checkPassed = false
				}

				// Apply threshold logic and determine current perceived health status
				nowHealthy := thc.updateTargetState(target.Address, checkPassed)

				thc.balancer.SetStatus(ctx, target.Address, nowHealthy)

				var statusStr string
				if nowHealthy {
					statusStr = runtime.StatusUp
				} else {
					statusStr = runtime.StatusDown
				}

				thc.info.UpdateServerStatus(target.Address, statusStr)

				// Route target to appropriate channel based on current health status
				if nowHealthy {
					thc.healthyTargets <- target
				} else {
					thc.unhealthyTargets <- target
				}

				// TODO: add a TCP server up metric (like for HTTP).
			}
		}
	}
}

// updateTargetState updates the target state based on the health check result and thresholds.
// Returns the current perceived health status after applying threshold logic.
func (thc *ServiceTCPHealthChecker) updateTargetState(targetAddress string, checkPassed bool) bool {
	thc.stateMu.Lock()
	defer thc.stateMu.Unlock()

	state, ok := thc.states[targetAddress]
	if !ok {
		// This shouldn't happen, but handle gracefully
		state = &targetState{healthy: true}
		thc.states[targetAddress] = state
	}

	if checkPassed {
		state.consecutiveFails = 0
		if !state.healthy {
			state.consecutivePasses++
			if state.consecutivePasses >= thc.passesThreshold {
				state.healthy = true
				state.consecutivePasses = 0
			}
		}
	} else {
		state.consecutivePasses = 0
		if state.healthy {
			state.consecutiveFails++
			if state.consecutiveFails >= thc.failsThreshold {
				state.healthy = false
				state.consecutiveFails = 0
			}
		}
	}

	return state.healthy
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
