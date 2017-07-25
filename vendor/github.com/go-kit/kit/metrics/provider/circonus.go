package provider

import (
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/circonus"
)

type circonusProvider struct {
	c *circonus.Circonus
}

// NewCirconusProvider takes the given Circonnus object and returns a Provider
// that produces Circonus metrics.
func NewCirconusProvider(c *circonus.Circonus) Provider {
	return &circonusProvider{
		c: c,
	}
}

// NewCounter implements Provider.
func (p *circonusProvider) NewCounter(name string) metrics.Counter {
	return p.c.NewCounter(name)
}

// NewGauge implements Provider.
func (p *circonusProvider) NewGauge(name string) metrics.Gauge {
	return p.c.NewGauge(name)
}

// NewHistogram implements Provider. The buckets parameter is ignored.
func (p *circonusProvider) NewHistogram(name string, _ int) metrics.Histogram {
	return p.c.NewHistogram(name)
}

// Stop implements Provider, but is a no-op.
func (p *circonusProvider) Stop() {}
