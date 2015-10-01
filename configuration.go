package main

import (
	"errors"
	"strings"
)

type GlobalConfiguration struct {
	Port              string
	GraceTimeOut      int64
	AccessLogsFile    string
	TraefikLogsFile   string
	CertFile, KeyFile string
	LogLevel          string
	Docker            *DockerProvider
	File              *FileProvider
	Web               *WebProvider
	Marathon          *MarathonProvider
	Consul            *ConsulProvider
	Etcd              *EtcdProvider
	Zookeeper         *ZookepperProvider
	Boltdb            *BoltDbProvider
}

func NewGlobalConfiguration() *GlobalConfiguration {
	globalConfiguration := new(GlobalConfiguration)
	// default values
	globalConfiguration.Port = ":80"
	globalConfiguration.GraceTimeOut = 10
	globalConfiguration.LogLevel = "ERROR"

	return globalConfiguration
}

type Backend struct {
	Servers        map[string]Server
	CircuitBreaker *CircuitBreaker
	LoadBalancer   *LoadBalancer
}

type LoadBalancer struct {
	Method string
}

type CircuitBreaker struct {
	Expression string
}

type Server struct {
	URL    string
	Weight int
}

type Route struct {
	Rule  string
	Value string
}

type Frontend struct {
	Backend string
	Routes  map[string]Route
}

type Configuration struct {
	Backends  map[string]*Backend
	Frontends map[string]*Frontend
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
