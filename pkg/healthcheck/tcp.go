package healthcheck

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
)

type ServiceTCPHealthChecker struct {
	balancer StatusSetter
	info     *runtime.TCPServiceInfo

	config   *dynamic.TCPServerHealthCheck
	interval time.Duration
	timeout  time.Duration

	metrics metricsHealthCheck

	targets     map[string]*net.TCPAddr
	serviceName string
}

func NewServiceTCPHealthChecker(metrics metricsHealthCheck, config *dynamic.TCPServerHealthCheck, service StatusSetter, info *runtime.TCPServiceInfo, targets map[string]*net.TCPAddr, serviceName string) *ServiceTCPHealthChecker {
	return &ServiceTCPHealthChecker{
		balancer:    service,
		info:        info,
		config:      config,
		interval:    time.Duration(config.Interval),
		timeout:     time.Duration(config.Timeout),
		metrics:     metrics,
		targets:     targets,
		serviceName: serviceName,
	}
}

func (thc *ServiceTCPHealthChecker) Launch(ctx context.Context) {
	ticker := time.NewTicker(thc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			for proxyName, target := range thc.targets {
				select {
				case <-ctx.Done():
					return
				default:
				}

				isUp := true
				serverUpMetricValue := float64(1)

				if err := thc.executeHealthCheck(ctx, thc.config, target); err != nil {
					// The context is canceled when the dynamic configuration is refreshed.
					if errors.Is(err, context.Canceled) {
						return
					}

					log.Ctx(ctx).Warn().
						Str("targetURL", target.String()).
						Err(err).
						Msg("Health check failed.")

					isUp = false
					serverUpMetricValue = float64(0)
				}

				thc.balancer.SetStatus(ctx, proxyName, isUp)

				thc.info.UpdateServerStatus(target.String(), isUp)

				thc.metrics.ServiceServerUpGauge().
					With("service", thc.serviceName, "url", target.String()).
					Set(serverUpMetricValue)
			}
		}
	}
}

func (thc *ServiceTCPHealthChecker) executeHealthCheck(ctx context.Context, config *dynamic.TCPServerHealthCheck, target *net.TCPAddr) error {
	ctx, cancel := context.WithTimeout(ctx, thc.timeout)
	defer cancel()

	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", target.String())
	if err != nil {
		return err
	}

	defer conn.Close()

	if config.Payload != "" {
		_, err = conn.Write([]byte(config.Payload))
		if err != nil {
			return err
		}
	}

	if config.Expected != "" {
		buf := make([]byte, len(config.Expected))
		_, err = conn.Read(buf)
		if err != nil {
			return err
		}

		if string(buf) != config.Expected {
			return errors.New("unexpected response")
		}
	}

	return nil
}
