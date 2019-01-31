package server

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/containous/traefik/configuration"
	"github.com/containous/traefik/healthcheck"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/middlewares"
	"github.com/containous/traefik/middlewares/accesslog"
	"github.com/containous/traefik/server/cookie"
	traefiktls "github.com/containous/traefik/tls"
	"github.com/containous/traefik/types"
	"github.com/vulcand/oxy/buffer"
	"github.com/vulcand/oxy/connlimit"
	"github.com/vulcand/oxy/ratelimit"
	"github.com/vulcand/oxy/roundrobin"
	"github.com/vulcand/oxy/utils"
	"golang.org/x/net/http2"
)

type h2cTransportWrapper struct {
	*http2.Transport
}

func (t *h2cTransportWrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	return t.Transport.RoundTrip(req)
}

func (s *Server) buildBalancerMiddlewares(frontendName string, frontend *types.Frontend, backend *types.Backend, fwd http.Handler) (http.Handler, *healthcheck.BackendConfig, error) {
	balancer, err := s.buildLoadBalancer(frontendName, frontend.Backend, backend, fwd)
	if err != nil {
		return nil, nil, err
	}

	// Health Check
	var backendHealthCheck *healthcheck.BackendConfig
	if hcOpts := buildHealthCheckOptions(balancer, frontend.Backend, backend.HealthCheck, s.globalConfiguration.HealthCheck); hcOpts != nil {
		log.Debugf("Setting up backend health check %s", *hcOpts)

		hcOpts.Transport = s.defaultForwardingRoundTripper
		backendHealthCheck = healthcheck.NewBackendConfig(*hcOpts, frontend.Backend)
	}

	// Empty (backend with no servers)
	var lb http.Handler = middlewares.NewEmptyBackendHandler(balancer)

	// Rate Limit
	if frontend.RateLimit != nil && len(frontend.RateLimit.RateSet) > 0 {
		handler, err := buildRateLimiter(lb, frontend.RateLimit)
		if err != nil {
			return nil, nil, fmt.Errorf("error creating rate limiter: %v", err)
		}

		lb = s.wrapHTTPHandlerWithAccessLog(
			s.tracingMiddleware.NewHTTPHandlerWrapper("Rate limit", handler, false),
			fmt.Sprintf("rate limit for %s", frontendName),
		)
	}

	// Max Connections
	if backend.MaxConn != nil && backend.MaxConn.Amount != 0 {
		log.Debugf("Creating load-balancer connection limit")

		handler, err := buildMaxConn(lb, backend.MaxConn)
		if err != nil {
			return nil, nil, err
		}
		lb = s.wrapHTTPHandlerWithAccessLog(handler, fmt.Sprintf("connection limit for %s", frontendName))
	}

	// Retry
	if s.globalConfiguration.Retry != nil {
		handler := s.buildRetryMiddleware(lb, s.globalConfiguration.Retry, len(backend.Servers), frontend.Backend)
		lb = s.tracingMiddleware.NewHTTPHandlerWrapper("Retry", handler, false)
	}

	// Buffering
	if backend.Buffering != nil {
		handler, err := buildBufferingMiddleware(lb, backend.Buffering)
		if err != nil {
			return nil, nil, fmt.Errorf("error setting up buffering middleware: %s", err)
		}

		// TODO refactor ?
		lb = handler
	}

	// Circuit Breaker
	if backend.CircuitBreaker != nil {
		log.Debugf("Creating circuit breaker %s", backend.CircuitBreaker.Expression)

		expression := backend.CircuitBreaker.Expression
		circuitBreaker, err := middlewares.NewCircuitBreaker(lb, expression, middlewares.NewCircuitBreakerOptions(expression))
		if err != nil {
			return nil, nil, fmt.Errorf("error creating circuit breaker: %v", err)
		}

		lb = s.tracingMiddleware.NewHTTPHandlerWrapper("Circuit breaker", circuitBreaker, false)
	}

	return lb, backendHealthCheck, nil
}

func (s *Server) buildLoadBalancer(frontendName string, backendName string, backend *types.Backend, fwd http.Handler) (healthcheck.BalancerHandler, error) {
	var rr *roundrobin.RoundRobin
	var saveFrontend http.Handler

	if s.accessLoggerMiddleware != nil {
		saveUsername := accesslog.NewSaveUsername(fwd)
		saveBackend := accesslog.NewSaveBackend(saveUsername, backendName)
		saveFrontend = accesslog.NewSaveFrontend(saveBackend, frontendName)
		rr, _ = roundrobin.New(saveFrontend)
	} else {
		rr, _ = roundrobin.New(fwd)
	}

	var stickySession *roundrobin.StickySession
	var cookieName string
	if stickiness := backend.LoadBalancer.Stickiness; stickiness != nil {
		cookieName = cookie.GetName(stickiness.CookieName, backendName)
		stickySession = roundrobin.NewStickySession(cookieName)
	}

	lbMethod, err := types.NewLoadBalancerMethod(backend.LoadBalancer)
	if err != nil {
		return nil, fmt.Errorf("error loading load balancer method '%+v' for frontend %s: %v", backend.LoadBalancer, frontendName, err)
	}

	var lb healthcheck.BalancerHandler

	switch lbMethod {
	case types.Drr:
		log.Debug("Creating load-balancer drr")

		if stickySession != nil {
			log.Debugf("Sticky session with cookie %v", cookieName)

			lb, err = roundrobin.NewRebalancer(rr, roundrobin.RebalancerStickySession(stickySession))
			if err != nil {
				return nil, err
			}
		} else {
			lb, err = roundrobin.NewRebalancer(rr)
			if err != nil {
				return nil, err
			}
		}
	case types.Wrr:
		log.Debug("Creating load-balancer wrr")

		if stickySession != nil {
			log.Debugf("Sticky session with cookie %v", cookieName)

			if s.accessLoggerMiddleware != nil {
				lb, err = roundrobin.New(saveFrontend, roundrobin.EnableStickySession(stickySession))
				if err != nil {
					return nil, err
				}
			} else {
				lb, err = roundrobin.New(fwd, roundrobin.EnableStickySession(stickySession))
				if err != nil {
					return nil, err
				}
			}
		} else {
			lb = rr
		}
	default:
		return nil, fmt.Errorf("invalid load-balancing method %q", lbMethod)
	}

	if err := s.configureLBServers(lb, backend, backendName); err != nil {
		return nil, fmt.Errorf("error configuring load balancer for frontend %s: %v", frontendName, err)
	}

	return lb, nil
}

func (s *Server) configureLBServers(lb healthcheck.BalancerHandler, backend *types.Backend, backendName string) error {
	for name, srv := range backend.Servers {
		u, err := url.Parse(srv.URL)
		if err != nil {
			return fmt.Errorf("error parsing server URL %s: %v", srv.URL, err)
		}

		log.Debugf("Creating server %s at %s with weight %d", name, u, srv.Weight)

		if err := lb.UpsertServer(u, roundrobin.Weight(srv.Weight)); err != nil {
			return fmt.Errorf("error adding server %s to load balancer: %v", srv.URL, err)
		}

		s.metricsRegistry.BackendServerUpGauge().With("backend", backendName, "url", srv.URL).Set(1)
	}
	return nil
}

// getRoundTripper will either use server.defaultForwardingRoundTripper or create a new one
// given a custom TLS configuration is passed and the passTLSCert option is set to true.
func (s *Server) getRoundTripper(entryPointName string, passTLSCert bool, tls *traefiktls.TLS) (http.RoundTripper, error) {
	if passTLSCert {
		tlsConfig, err := createClientTLSConfig(entryPointName, tls)
		if err != nil {
			return nil, fmt.Errorf("failed to create TLSClientConfig: %v", err)
		}
		tlsConfig.InsecureSkipVerify = s.globalConfiguration.InsecureSkipVerify

		transport, err := createHTTPTransport(s.globalConfiguration)
		if err != nil {
			return nil, fmt.Errorf("failed to create HTTP transport: %v", err)
		}

		transport.TLSClientConfig = tlsConfig
		return transport, nil
	}

	return s.defaultForwardingRoundTripper, nil
}

// createHTTPTransport creates an http.Transport configured with the GlobalConfiguration settings.
// For the settings that can't be configured in Traefik it uses the default http.Transport settings.
// An exception to this is the MaxIdleConns setting as we only provide the option MaxIdleConnsPerHost
// in Traefik at this point in time. Setting this value to the default of 100 could lead to confusing
// behaviour and backwards compatibility issues.
func createHTTPTransport(globalConfiguration configuration.GlobalConfiguration) (*http.Transport, error) {
	dialer := &net.Dialer{
		Timeout:   configuration.DefaultDialTimeout,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}

	if globalConfiguration.ForwardingTimeouts != nil {
		dialer.Timeout = time.Duration(globalConfiguration.ForwardingTimeouts.DialTimeout)
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		MaxIdleConnsPerHost:   globalConfiguration.MaxIdleConnsPerHost,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	transport.RegisterProtocol("h2c", &h2cTransportWrapper{
		Transport: &http2.Transport{
			DialTLS: func(netw, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(netw, addr)
			},
			AllowHTTP: true,
		},
	})

	if globalConfiguration.ForwardingTimeouts != nil {
		transport.ResponseHeaderTimeout = time.Duration(globalConfiguration.ForwardingTimeouts.ResponseHeaderTimeout)
	}

	if globalConfiguration.InsecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if len(globalConfiguration.RootCAs) > 0 {
		transport.TLSClientConfig = &tls.Config{
			RootCAs: createRootCACertPool(globalConfiguration.RootCAs),
		}
	}

	err := http2.ConfigureTransport(transport)
	if err != nil {
		return nil, err
	}

	return transport, nil
}

func createRootCACertPool(rootCAs traefiktls.FilesOrContents) *x509.CertPool {
	roots := x509.NewCertPool()

	for _, cert := range rootCAs {
		certContent, err := cert.Read()
		if err != nil {
			log.Error("Error while read RootCAs", err)
			continue
		}
		roots.AppendCertsFromPEM(certContent)
	}

	return roots
}

func createClientTLSConfig(entryPointName string, tlsOption *traefiktls.TLS) (*tls.Config, error) {
	if tlsOption == nil {
		return nil, errors.New("no TLS provided")
	}

	config, err := tlsOption.Certificates.CreateTLSConfig(entryPointName)
	if err != nil {
		return nil, err
	}

	if len(tlsOption.ClientCAFiles) > 0 {
		log.Warnf("Deprecated configuration found during client TLS configuration creation: %s. Please use %s (which allows to make the CA Files optional).", "tls.ClientCAFiles", "tls.ClientCA.files")
		tlsOption.ClientCA.Files = tlsOption.ClientCAFiles
		tlsOption.ClientCA.Optional = false
	}

	if len(tlsOption.ClientCA.Files) > 0 {
		pool := x509.NewCertPool()
		for _, caFile := range tlsOption.ClientCA.Files {
			data, err := caFile.Read()
			if err != nil {
				return nil, err
			}

			if !pool.AppendCertsFromPEM(data) {
				return nil, fmt.Errorf("invalid certificate(s) in %s", caFile)
			}
		}
		config.RootCAs = pool
	}

	config.BuildNameToCertificate()

	return config, nil
}

func (s *Server) buildRetryMiddleware(handler http.Handler, retry *configuration.Retry, countServers int, backendName string) http.Handler {
	retryListeners := middlewares.RetryListeners{}
	if s.metricsRegistry.IsEnabled() {
		retryListeners = append(retryListeners, middlewares.NewMetricsRetryListener(s.metricsRegistry, backendName))
	}
	if s.accessLoggerMiddleware != nil {
		retryListeners = append(retryListeners, &accesslog.SaveRetries{})
	}

	retryAttempts := countServers
	if retry.Attempts > 0 {
		retryAttempts = retry.Attempts
	}

	log.Debugf("Creating retries max attempts %d", retryAttempts)

	return middlewares.NewRetry(retryAttempts, handler, retryListeners)
}

func buildRateLimiter(handler http.Handler, rlConfig *types.RateLimit) (http.Handler, error) {
	extractFunc, err := utils.NewExtractor(rlConfig.ExtractorFunc)
	if err != nil {
		return nil, err
	}

	log.Debugf("Creating load-balancer rate limiter")

	rateSet := ratelimit.NewRateSet()
	for _, rate := range rlConfig.RateSet {
		if err := rateSet.Add(time.Duration(rate.Period), rate.Average, rate.Burst); err != nil {
			return nil, err
		}
	}

	return ratelimit.New(handler, extractFunc, rateSet)
}

func buildBufferingMiddleware(handler http.Handler, config *types.Buffering) (http.Handler, error) {
	log.Debugf("Setting up buffering: request limits: %d (mem), %d (max), response limits: %d (mem), %d (max) with retry: '%s'",
		config.MemRequestBodyBytes, config.MaxRequestBodyBytes, config.MemResponseBodyBytes,
		config.MaxResponseBodyBytes, config.RetryExpression)

	return buffer.New(
		handler,
		buffer.MemRequestBodyBytes(config.MemRequestBodyBytes),
		buffer.MaxRequestBodyBytes(config.MaxRequestBodyBytes),
		buffer.MemResponseBodyBytes(config.MemResponseBodyBytes),
		buffer.MaxResponseBodyBytes(config.MaxResponseBodyBytes),
		buffer.CondSetter(len(config.RetryExpression) > 0, buffer.Retry(config.RetryExpression)),
	)
}

func buildMaxConn(lb http.Handler, maxConns *types.MaxConn) (http.Handler, error) {
	extractFunc, err := utils.NewExtractor(maxConns.ExtractorFunc)
	if err != nil {
		return nil, fmt.Errorf("error creating connection limit: %v", err)
	}

	log.Debugf("Creating load-balancer connection limit")

	handler, err := connlimit.New(lb, extractFunc, maxConns.Amount)
	if err != nil {
		return nil, fmt.Errorf("error creating connection limit: %v", err)
	}

	return handler, nil
}

func buildHealthCheckOptions(lb healthcheck.BalancerHandler, backend string, hc *types.HealthCheck, hcConfig *configuration.HealthCheckConfig) *healthcheck.Options {
	if hc == nil || hc.Path == "" || hcConfig == nil {
		return nil
	}

	interval := time.Duration(hcConfig.Interval)
	if hc.Interval != "" {
		intervalOverride, err := time.ParseDuration(hc.Interval)
		if err != nil {
			log.Errorf("Illegal health check interval for backend '%s': %s", backend, err)
		} else if intervalOverride <= 0 {
			log.Errorf("Health check interval smaller than zero for backend '%s', backend", backend)
		} else {
			interval = intervalOverride
		}
	}

	return &healthcheck.Options{
		Scheme:   hc.Scheme,
		Path:     hc.Path,
		Port:     hc.Port,
		Interval: interval,
		LB:       lb,
		Hostname: hc.Hostname,
		Headers:  hc.Headers,
	}
}
