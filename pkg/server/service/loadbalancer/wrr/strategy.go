package wrr

import (
	"net/http"
	"net/url"
	"sort"
)

func init() {
	var _ LBStrategy = new(CompositeStrategy)
}
func Strategy() LBStrategy {
	return strategies
}

func Provide(lbs LBStrategy) {
	strategies.Add(lbs)
}

var strategies = new(CompositeStrategy)

type Server interface {
	http.Handler

	Name() string

	Deadline() float64

	// URL server url. maybe nil
	URL() *url.URL

	// Weight Relative weight for the endpoint to other endpoints in the load balancer.
	Weight() float64

	// Set the weight.
	Set(weight float64)
}

type LBStrategy interface {

	// Name is the strategy name.
	Name() string

	// Priority more than has more priority.
	Priority() int

	// Next servers
	// Load balancer extension for custom rules filter.
	Next(w http.ResponseWriter, req *http.Request, servers []Server) []Server
}

type CompositeStrategy struct {
	strategies []LBStrategy
}

func (that *CompositeStrategy) Add(lbs LBStrategy) *CompositeStrategy {
	that.strategies = append(that.strategies, lbs)
	sort.Slice(that.strategies, func(i, j int) bool { return that.strategies[i].Priority() < that.strategies[j].Priority() })
	return that
}

func (that *CompositeStrategy) Name() string {
	return "composite"
}

func (that *CompositeStrategy) Priority() int {
	return 0
}

func (that *CompositeStrategy) Next(w http.ResponseWriter, req *http.Request, servers []Server) []Server {
	for _, strategy := range that.strategies {
		servers = strategy.Next(w, req, servers)
	}
	return servers
}
