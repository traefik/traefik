// cbreaker package implements circuit breaker similar to  https://github.com/Netflix/Hystrix/wiki/How-it-Works
//
// Vulcand circuit breaker watches the error condtion to match
// after what it activates the fallback scenario, e.g. returns the response code
// or redirects the request to another location

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
// It is possible to define actions of transitions (Standby -> Tripped) and (Recovering -> Standby)
// using handlers 'OnTripped' and 'OnStandby', e.g. issuing webhook calls.
//

package cbreaker

import (
	"fmt"
	"net/http"
	"time"

	"github.com/codegangsta/cli"
	"github.com/vulcand/oxy/cbreaker"
	"github.com/vulcand/vulcand/plugin"
)

const Type = "cbreaker"

func GetSpec() *plugin.MiddlewareSpec {
	return &plugin.MiddlewareSpec{
		Type:      Type,
		FromOther: FromOther,
		FromCli:   FromCli,
		CliFlags:  CliFlags(),
	}
}

// Spec defines circuit breaker middleware parameters
type Spec struct {
	// Condition is a JSON dictionary formula to set circuit breaker in "Tripped" state
	Condition string
	// Fallback is a JSON dictionary with fallback action, such as response or redirect
	Fallback interface{}

	// OnTripped defines JSON dict with action executed after (Standby -> Tripped) transition takes place
	OnTripped interface{}
	// OnStandby defines JSON dict with action executed after (Recovering -> Standby) transition takes place)
	OnStandby interface{}

	// FallbackDuration defines time period for circuit breaker to activate fallback scenario for all requests
	FallbackDuration time.Duration

	// Recovery duration defines time period for circuit breaker to increase traffic to the original upstream.
	RecoveryDuration time.Duration

	// CheckPeriod defines the period between circuit breaker checks
	CheckPeriod time.Duration
}

func fromSpec(spec *Spec, next http.Handler) (*cbreaker.CircuitBreaker, error) {
	b, err := toBytes(spec.Fallback)
	if err != nil {
		return nil, err
	}

	fallback, err := actionFromJSON(b)
	if err != nil {
		return nil, err
	}
	var onTripped, onStandby cbreaker.SideEffect
	if spec.OnTripped != nil {
		b, err := toBytes(spec.OnTripped)
		if err != nil {
			return nil, err
		}
		if len(b) != 0 {
			v, err := sideEffectFromJSON(b)
			if err != nil {
				return nil, err
			}
			onTripped = v
		}
	}

	if spec.OnStandby != nil {
		b, err := toBytes(spec.OnStandby)
		if err != nil {
			return nil, err
		}
		if len(b) != 0 {
			v, err := sideEffectFromJSON(b)
			if err != nil {
				return nil, err
			}
			onStandby = v
		}
	}

	return cbreaker.New(next, spec.Condition,
		cbreaker.Fallback(fallback),
		cbreaker.FallbackDuration(spec.FallbackDuration),
		cbreaker.RecoveryDuration(spec.RecoveryDuration),
		cbreaker.OnTripped(onTripped),
		cbreaker.OnStandby(onStandby),
		cbreaker.CheckPeriod(spec.CheckPeriod))
}

// NewMiddleware vulcan library compatible middleware
func (c *Spec) NewHandler(next http.Handler) (http.Handler, error) {
	return fromSpec(c, next)
}

// NewSpec check parameters and returns new specification for the middleware
func NewSpec(condition string, fallback, onTripped, onStandby interface{}, fallbackDuration, recoveryDuration, checkPeriod time.Duration) (*Spec, error) {
	spec := &Spec{
		Condition:        condition,
		Fallback:         fallback,
		OnTripped:        onTripped,
		OnStandby:        onStandby,
		RecoveryDuration: recoveryDuration,
		FallbackDuration: fallbackDuration,
		CheckPeriod:      checkPeriod,
	}
	if _, err := fromSpec(spec, nil); err != nil {
		return nil, err
	}
	return spec, nil
}

func (c *Spec) String() string {
	return fmt.Sprintf("condition=%s, fallback=%v, recovery=%v, period=%v", c.Condition, c.FallbackDuration, c.RecoveryDuration, c.CheckPeriod)
}

// FromOther is used to read spec from the serialized format
func FromOther(c Spec) (plugin.Middleware, error) {
	return NewSpec(c.Condition, c.Fallback, c.OnTripped, c.OnStandby, c.FallbackDuration, c.RecoveryDuration, c.CheckPeriod)
}

// FromCli constructs the middleware from the command line arguments
func FromCli(c *cli.Context) (plugin.Middleware, error) {
	return NewSpec(c.String("condition"), c.String("fallback"), c.String("onTripped"), c.String("onStandby"), c.Duration("fallbackDuration"), c.Duration("recoveryDuration"), c.Duration("checkPeriod"))
}

func CliFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{Name: "condition", Usage: "Condition defines a formula for setting the cbreaker to 'Tripped' state e.g. 'LatencyAtQuantileMS(50) > 40'"},
		cli.StringFlag{Name: "fallback", Usage: `Fallback action e.g. {"Type": "response", Action: {"StatusCode": 400, "Body": "Come back later"}}`},

		cli.StringFlag{Name: "onTripped", Usage: `Action executed when circuit breaker is tripped e.g. {"Type": "webhook", Action: {"Method": "POST", "Form": {"Key": ["Val"]}}}`},
		cli.StringFlag{Name: "onStandby", Usage: `Action executed when circuit breaker is transitioned back to standby mode e.g. {"Type": "webhook", Action: {"Method": "POST", "Form": {"Key": ["Val"]}}}`},

		cli.DurationFlag{Name: "fallbackDuration", Usage: "Circuit breaker will default to fallback during this period without checking the backend status", Value: defaultFallbackDuration},
		cli.DurationFlag{Name: "recoveryDuration", Usage: "Circuit breaker will start passing some traffic through to the upstreams ramping up to full speed", Value: defaultRecoveryDuration},

		cli.DurationFlag{Name: "checkPeriod", Usage: "Period between circuit breaker checks", Value: defaultCheckPeriod},
	}
}

const (
	defaultFallbackDuration = 10 * time.Second
	defaultRecoveryDuration = 10 * time.Second
	defaultCheckPeriod      = 100 * time.Millisecond
)
