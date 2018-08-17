// Package cbreaker implements circuit breaker similar to  https://github.com/Netflix/Hystrix/wiki/How-it-Works
//
// Vulcan circuit breaker watches the error condtion to match
// after which it activates the fallback scenario, e.g. returns the response code
// or redirects the request to another location
//
// Circuit breakers start in the Standby state first, observing responses and watching location metrics.
//
// Once the Circuit breaker condition is met, it enters the "Tripped" state, where it activates fallback scenario
// for all requests during the FallbackDuration time period and reset the stats for the location.
//
// After FallbackDuration time period passes, Circuit breaker enters "Recovering" state, during that state it will
// start passing some traffic back to the endpoints, increasing the amount of passed requests using linear function:
//
//    allowedRequestsRatio = 0.5 * (Now() - StartRecovery())/RecoveryDuration
//
// Two scenarios are possible in the "Recovering" state:
// 1. Condition matches again, this will reset the state to "Tripped" and reset the timer.
// 2. Condition does not match, circuit breaker enters "Standby" state
//
// It is possible to define actions (e.g. webhooks) of transitions between states:
//
// * OnTripped action is called on transition (Standby -> Tripped)
// * OnStandby action is called on transition (Recovering -> Standby)
//
package cbreaker

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/mailgun/timetools"
	log "github.com/sirupsen/logrus"
	"github.com/vulcand/oxy/memmetrics"
	"github.com/vulcand/oxy/utils"
)

// CircuitBreaker is http.Handler that implements circuit breaker pattern
type CircuitBreaker struct {
	m       *sync.RWMutex
	metrics *memmetrics.RTMetrics

	condition hpredicate

	fallbackDuration time.Duration
	recoveryDuration time.Duration

	onTripped SideEffect
	onStandby SideEffect

	state cbState
	until time.Time

	rc *ratioController

	checkPeriod time.Duration
	lastCheck   time.Time

	fallback http.Handler
	next     http.Handler

	clock timetools.TimeProvider

	log *log.Logger
}

// New creates a new CircuitBreaker middleware
func New(next http.Handler, expression string, options ...CircuitBreakerOption) (*CircuitBreaker, error) {
	cb := &CircuitBreaker{
		m:    &sync.RWMutex{},
		next: next,
		// Default values. Might be overwritten by options below.
		clock:            &timetools.RealTime{},
		checkPeriod:      defaultCheckPeriod,
		fallbackDuration: defaultFallbackDuration,
		recoveryDuration: defaultRecoveryDuration,
		fallback:         defaultFallback,
		log:              log.StandardLogger(),
	}

	for _, s := range options {
		if err := s(cb); err != nil {
			return nil, err
		}
	}

	condition, err := parseExpression(expression)
	if err != nil {
		return nil, err
	}
	cb.condition = condition

	mt, err := memmetrics.NewRTMetrics()
	if err != nil {
		return nil, err
	}
	cb.metrics = mt

	return cb, nil
}

// Logger defines the logger the circuit breaker will use.
//
// It defaults to logrus.StandardLogger(), the global logger used by logrus.
func Logger(l *log.Logger) CircuitBreakerOption {
	return func(c *CircuitBreaker) error {
		c.log = l
		return nil
	}
}

func (c *CircuitBreaker) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if c.log.Level >= log.DebugLevel {
		logEntry := c.log.WithField("Request", utils.DumpHttpRequest(req))
		logEntry.Debug("vulcand/oxy/circuitbreaker: begin ServeHttp on request")
		defer logEntry.Debug("vulcand/oxy/circuitbreaker: completed ServeHttp on request")
	}
	if c.activateFallback(w, req) {
		c.fallback.ServeHTTP(w, req)
		return
	}
	c.serve(w, req)
}

// Wrap sets the next handler to be called by circuit breaker handler.
func (c *CircuitBreaker) Wrap(next http.Handler) {
	c.next = next
}

// updateState updates internal state and returns true if fallback should be used and false otherwise
func (c *CircuitBreaker) activateFallback(w http.ResponseWriter, req *http.Request) bool {
	// Quick check with read locks optimized for normal operation use-case
	if c.isStandby() {
		return false
	}
	// Circuit breaker is in tripped or recovering state
	c.m.Lock()
	defer c.m.Unlock()

	c.log.Warnf("%v is in error state", c)

	switch c.state {
	case stateStandby:
		// someone else has set it to standby just now
		return false
	case stateTripped:
		if c.clock.UtcNow().Before(c.until) {
			return true
		}
		// We have been in active state enough, enter recovering state
		c.setRecovering()
		fallthrough
	case stateRecovering:
		// We have been in recovering state enough, enter standby and allow request
		if c.clock.UtcNow().After(c.until) {
			c.setState(stateStandby, c.clock.UtcNow())
			return false
		}
		// ratio controller allows this request
		if c.rc.allowRequest() {
			return false
		}
		return true
	}
	return false
}

func (c *CircuitBreaker) serve(w http.ResponseWriter, req *http.Request) {
	start := c.clock.UtcNow()
	p := utils.NewProxyWriterWithLogger(w, c.log)

	c.next.ServeHTTP(p, req)

	latency := c.clock.UtcNow().Sub(start)
	c.metrics.Record(p.StatusCode(), latency)

	// Note that this call is less expensive than it looks -- checkCondition only performs the real check
	// periodically. Because of that we can afford to call it here on every single response.
	c.checkAndSet()
}

func (c *CircuitBreaker) isStandby() bool {
	c.m.RLock()
	defer c.m.RUnlock()
	return c.state == stateStandby
}

// String returns log-friendly representation of the circuit breaker state
func (c *CircuitBreaker) String() string {
	switch c.state {
	case stateTripped, stateRecovering:
		return fmt.Sprintf("CircuitBreaker(state=%v, until=%v)", c.state, c.until)
	default:
		return fmt.Sprintf("CircuitBreaker(state=%v)", c.state)
	}
}

// exec executes side effect
func (c *CircuitBreaker) exec(s SideEffect) {
	if s == nil {
		return
	}
	go func() {
		if err := s.Exec(); err != nil {
			c.log.Errorf("%v side effect failure: %v", c, err)
		}
	}()
}

func (c *CircuitBreaker) setState(new cbState, until time.Time) {
	c.log.Debugf("%v setting state to %v, until %v", c, new, until)
	c.state = new
	c.until = until
	switch new {
	case stateTripped:
		c.exec(c.onTripped)
	case stateStandby:
		c.exec(c.onStandby)
	}
}

func (c *CircuitBreaker) timeToCheck() bool {
	c.m.RLock()
	defer c.m.RUnlock()
	return c.clock.UtcNow().After(c.lastCheck)
}

// Checks if tripping condition matches and sets circuit breaker to the tripped state
func (c *CircuitBreaker) checkAndSet() {
	if !c.timeToCheck() {
		return
	}

	c.m.Lock()
	defer c.m.Unlock()

	// Other goroutine could have updated the lastCheck variable before we grabbed mutex
	if !c.clock.UtcNow().After(c.lastCheck) {
		return
	}
	c.lastCheck = c.clock.UtcNow().Add(c.checkPeriod)

	if c.state == stateTripped {
		c.log.Debugf("%v skip set tripped", c)
		return
	}

	if !c.condition(c) {
		return
	}

	c.setState(stateTripped, c.clock.UtcNow().Add(c.fallbackDuration))
	c.metrics.Reset()
}

func (c *CircuitBreaker) setRecovering() {
	c.setState(stateRecovering, c.clock.UtcNow().Add(c.recoveryDuration))
	c.rc = newRatioController(c.clock, c.recoveryDuration, c.log)
}

// CircuitBreakerOption represents an option you can pass to New.
// See the documentation for the individual options below.
type CircuitBreakerOption func(*CircuitBreaker) error

// Clock allows you to fake che CircuitBreaker's view of the current time.
// Intended for unit tests.
func Clock(clock timetools.TimeProvider) CircuitBreakerOption {
	return func(c *CircuitBreaker) error {
		c.clock = clock
		return nil
	}
}

// FallbackDuration is how long the CircuitBreaker will remain in the Tripped
// state before trying to recover.
func FallbackDuration(d time.Duration) CircuitBreakerOption {
	return func(c *CircuitBreaker) error {
		c.fallbackDuration = d
		return nil
	}
}

// RecoveryDuration is how long the CircuitBreaker will take to ramp up
// requests during the Recovering state.
func RecoveryDuration(d time.Duration) CircuitBreakerOption {
	return func(c *CircuitBreaker) error {
		c.recoveryDuration = d
		return nil
	}
}

// CheckPeriod is how long the CircuitBreaker will wait between successive
// checks of the breaker condition.
func CheckPeriod(d time.Duration) CircuitBreakerOption {
	return func(c *CircuitBreaker) error {
		c.checkPeriod = d
		return nil
	}
}

// OnTripped sets a SideEffect to run when entering the Tripped state.
// Only one SideEffect can be set for this hook.
func OnTripped(s SideEffect) CircuitBreakerOption {
	return func(c *CircuitBreaker) error {
		c.onTripped = s
		return nil
	}
}

// OnStandby sets a SideEffect to run when entering the Standby state.
// Only one SideEffect can be set for this hook.
func OnStandby(s SideEffect) CircuitBreakerOption {
	return func(c *CircuitBreaker) error {
		c.onStandby = s
		return nil
	}
}

// Fallback defines the http.Handler that the CircuitBreaker should route
// requests to when it prevents a request from taking its normal path.
func Fallback(h http.Handler) CircuitBreakerOption {
	return func(c *CircuitBreaker) error {
		c.fallback = h
		return nil
	}
}

// cbState is the state of the circuit breaker
type cbState int

func (s cbState) String() string {
	switch s {
	case stateStandby:
		return "standby"
	case stateTripped:
		return "tripped"
	case stateRecovering:
		return "recovering"
	}
	return "undefined"
}

const (
	// CircuitBreaker is passing all requests and watching stats
	stateStandby = iota
	// CircuitBreaker activates fallback scenario for all requests
	stateTripped
	// CircuitBreaker passes some requests to go through, rejecting others
	stateRecovering
)

const (
	defaultFallbackDuration = 10 * time.Second
	defaultRecoveryDuration = 10 * time.Second
	defaultCheckPeriod      = 100 * time.Millisecond
)

var defaultFallback = &fallback{}

type fallback struct{}

func (f *fallback) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusServiceUnavailable)
	w.Write([]byte(http.StatusText(http.StatusServiceUnavailable)))
}
