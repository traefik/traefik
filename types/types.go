package types

import (
	"errors"
	"strings"
)

// Backend holds backend configuration.
type Backend struct {
	Servers        map[string]Server `json:"servers,omitempty"`
	CircuitBreaker *CircuitBreaker   `json:"circuitBreaker,omitempty"`
	LoadBalancer   *LoadBalancer     `json:"loadBalancer,omitempty"`
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
	Rule  string `json:"rule,omitempty"`
	Value string `json:"value,omitempty"`
}

// Frontend holds frontend configuration.
type Frontend struct {
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
