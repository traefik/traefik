package main

import (
	"errors"
	"strings"
	"time"
)

type GlobalConfiguration struct {
	Port                      string
	GraceTimeOut              int64
	AccessLogsFile            string
	TraefikLogsFile           string
	CertFile, KeyFile         string
	LogLevel                  string
	ProvidersThrottleDuration time.Duration
	Docker                    *DockerProvider
	File                      *FileProvider
	Web                       *WebProvider
	Marathon                  *MarathonProvider
	Consul                    *ConsulProvider
	Etcd                      *EtcdProvider
	Zookeeper                 *ZookepperProvider
	Boltdb                    *BoltDbProvider
}

func NewGlobalConfiguration() *GlobalConfiguration {
	globalConfiguration := new(GlobalConfiguration)
	// default values
	globalConfiguration.Port = ":80"
	globalConfiguration.GraceTimeOut = 10
	globalConfiguration.LogLevel = "ERROR"
	globalConfiguration.ProvidersThrottleDuration = time.Duration(2 * time.Second)

	return globalConfiguration
}

// Backend configuration
type Backend struct {
	Servers        map[string]Server `json:"servers,omitempty"`
	CircuitBreaker *CircuitBreaker   `json:"circuitBreaker,omitempty"`
	LoadBalancer   *LoadBalancer     `json:"loadBalancer,omitempty"`
}

// LoadBalancer configuration
type LoadBalancer struct {
	Method string `json:"method,omitempty"`
}

// CircuitBreaker configuration
type CircuitBreaker struct {
	Expression string `json:"expression,omitempty"`
}

// Server configuration
type Server struct {
	URL    string `json:"url,omitempty"`
	Weight int    `json:"weight,omitempty"`
}

// Route configuration
type Route struct {
	Rule  string `json:"rule,omitempty"`
	Value string `json:"value,omitempty"`
}

// Frontend configuration
type Frontend struct {
	Backend string           `json:"backend,omitempty"`
	Routes  map[string]Route `json:"routes,omitempty"`
}

// Configuration of a provider
type Configuration struct {
	Backends  map[string]*Backend  `json:"backends,omitempty"`
	Frontends map[string]*Frontend `json:"frontends,omitempty"`
}

// Load Balancer Method
type LoadBalancerMethod uint8

const (
	// wrr (default) = Weighted Round Robin
	wrr LoadBalancerMethod = iota
	// drr = Dynamic Round Robin
	drr
)

var loadBalancerMethodNames = []string{
	"wrr",
	"drr",
}

func NewLoadBalancerMethod(loadBalancer *LoadBalancer) (LoadBalancerMethod, error) {
	if loadBalancer != nil {
		for i, name := range loadBalancerMethodNames {
			if strings.EqualFold(name, loadBalancer.Method) {
				return LoadBalancerMethod(i), nil
			}
		}
	}
	return wrr, ErrInvalidLoadBalancerMethod
}

var ErrInvalidLoadBalancerMethod = errors.New("Invalid method, using default")

type configMessage struct {
	providerName  string
	configuration *Configuration
}

type configs map[string]*Configuration
