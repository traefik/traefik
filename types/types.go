package types

import (
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/ryanuber/go-glob"
	"reflect"
	"strings"
)

// Backend holds backend configuration.
type Backend struct {
	Servers        map[string]Server `json:"servers,omitempty"`
	CircuitBreaker *CircuitBreaker   `json:"circuitBreaker,omitempty"`
	LoadBalancer   *LoadBalancer     `json:"loadBalancer,omitempty"`
	MaxConn        *MaxConn          `json:"maxConn,omitempty"`
}

// MaxConn holds maximum connection configuration
type MaxConn struct {
	Amount        int64  `json:"amount,omitempty"`
	ExtractorFunc string `json:"extractorFunc,omitempty"`
}

// LoadBalancer holds load balancing configuration.
type LoadBalancer struct {
	Method string `json:"method,omitempty"`
}

// CircuitBreaker holds circuit breaker configuration.
type CircuitBreaker struct {
	Expression string `json:"expression,omitempty"`
}

// Server holds server configuration.
type Server struct {
	URL    string `json:"url,omitempty"`
	Weight int    `json:"weight,omitempty"`
}

// Route holds route configuration.
type Route struct {
	Rule string `json:"rule,omitempty"`
	// ⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠
	// TODO: backwards compatibility with DEPRECATED rule.Value
	Value string `json:"value,omitempty"`
	// ⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠
}

// Frontend holds frontend configuration.
type Frontend struct {
	EntryPoints    []string         `json:"entryPoints,omitempty"`
	Backend        string           `json:"backend,omitempty"`
	Routes         map[string]Route `json:"routes,omitempty"`
	PassHostHeader bool             `json:"passHostHeader,omitempty"`
}

// LoadBalancerMethod holds the method of load balancing to use.
type LoadBalancerMethod uint8

const (
	// Wrr (default) = Weighted Round Robin
	Wrr LoadBalancerMethod = iota
	// Drr = Dynamic Round Robin
	Drr
)

var loadBalancerMethodNames = []string{
	"Wrr",
	"Drr",
}

// NewLoadBalancerMethod create a new LoadBalancerMethod from a given LoadBalancer.
func NewLoadBalancerMethod(loadBalancer *LoadBalancer) (LoadBalancerMethod, error) {
	if loadBalancer != nil {
		for i, name := range loadBalancerMethodNames {
			if strings.EqualFold(name, loadBalancer.Method) {
				return LoadBalancerMethod(i), nil
			}
		}
	}
	return Wrr, ErrInvalidLoadBalancerMethod
}

// ErrInvalidLoadBalancerMethod is thrown when the specified load balancing method is invalid.
var ErrInvalidLoadBalancerMethod = errors.New("Invalid method, using default")

// Configuration of a provider.
type Configuration struct {
	Backends  map[string]*Backend  `json:"backends,omitempty"`
	Frontends map[string]*Frontend `json:"frontends,omitempty"`
}

// ConfigMessage hold configuration information exchanged between parts of traefik.
type ConfigMessage struct {
	ProviderName  string
	Configuration *Configuration
}

// Constraint hold a parsed constraint expresssion
type Constraint struct {
	Key string
	// MustMatch is true if operator is "==" or false if operator is "!="
	MustMatch bool
	Regex     string
}

// NewConstraint receive a string and return a *Constraint, after checking syntax and parsing the constraint expression
func NewConstraint(exp string) (*Constraint, error) {
	sep := ""
	constraint := &Constraint{}

	if strings.Contains(exp, "==") {
		sep = "=="
		constraint.MustMatch = true
	} else if strings.Contains(exp, "!=") {
		sep = "!="
		constraint.MustMatch = false
	} else {
		return nil, errors.New("Constraint expression missing valid operator: '==' or '!='")
	}

	kv := strings.SplitN(exp, sep, 2)
	if len(kv) == 2 {
		// At the moment, it only supports tags
		if kv[0] != "tag" {
			return nil, errors.New("Constraint must be tag-based. Syntax: tag==us-*")
		}

		constraint.Key = kv[0]
		constraint.Regex = kv[1]
		return constraint, nil
	}

	return nil, errors.New("Incorrect constraint expression: " + exp)
}

func (c *Constraint) String() string {
	if c.MustMatch {
		return c.Key + "==" + c.Regex
	}
	return c.Key + "!=" + c.Regex
}

// MatchConstraintWithAtLeastOneTag tests a constraint for one single service
func (c *Constraint) MatchConstraintWithAtLeastOneTag(tags []string) bool {
	for _, tag := range tags {
		if glob.Glob(c.Regex, tag) {
			return true
		}
	}
	return false
}

// StringToConstraintHookFunc returns a DecodeHookFunc that converts strings to Constraint.
// This hook is triggered during the configuration file unmarshal-ing
func StringToConstraintHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(&Constraint{}) {
			return data, nil
		}

		constraint, err := NewConstraint(data.(string))
		if err != nil {
			return data, err
		}
		return constraint, nil
	}
}

// Constraints own a pointer on globalConfiguration.Constraints and supports a Set() method (not possible on a slice)
// interface:
type Constraints struct {
	value   *[]*Constraint
	changed bool
}

// Set receive a cli argument and add it to globalConfiguration
func (cs *Constraints) Set(value string) error {
	exps := strings.Split(value, ",")
	if len(exps) == 0 {
		return errors.New("Bad Constraint format: " + value)
	}
	for _, exp := range exps {
		constraint, err := NewConstraint(exp)
		if err != nil {
			return err
		}
		*cs.value = append(*cs.value, constraint)
	}
	return nil
}

// Type exports the Constraints type as a string
func (cs *Constraints) Type() string {
	return "constraints"
}

func (cs *Constraints) String() string {
	return fmt.Sprintln("%v", *cs.value)
}

// NewConstraintSliceValue make an alias of []*Constraint to Constraints for the command line
// Viper does not supprt SliceVar value types
// Constraints.Set called by viper will fill the []*Constraint slice
func NewConstraintSliceValue(p *[]*Constraint) *Constraints {
	cs := new(Constraints)
	cs.value = p
	if p == nil {
		*cs.value = []*Constraint{}
	}
	return cs
}
