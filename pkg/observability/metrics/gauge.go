package metrics

import (
	"slices"
	"strings"
	"sync"

	"github.com/go-kit/kit/metrics"
)

// statefulGauge keeps the value of a gauge in memory,
// and sets the whole value on the wrapped gauge on every Add or Set.
// The StatsD and InfluxDB stores of go-kit reset their gauges at every push,
// so an Add on their gauges only sums the deltas of the current push interval,
// and pushes that partial sum as the value of the gauge.
type statefulGauge struct {
	state       *gaugeState
	gauge       metrics.Gauge
	labelValues []string
}

type gaugeState struct {
	dropLabels bool

	mu     sync.Mutex
	values map[string]float64
}

// newStatefulGauge wraps a gauge of a store which resets its gauges at every push.
// dropLabels must be true when the store ignores the labels (StatsD):
// all the label sets then share a single value.
func newStatefulGauge(gauge metrics.Gauge, dropLabels bool) metrics.Gauge {
	return &statefulGauge{
		state: &gaugeState{
			dropLabels: dropLabels,
			values:     make(map[string]float64),
		},
		gauge: gauge,
	}
}

func (g *statefulGauge) With(labelValues ...string) metrics.Gauge {
	if g.state.dropLabels {
		return g
	}

	return &statefulGauge{
		state:       g.state,
		gauge:       g.gauge.With(labelValues...),
		labelValues: append(slices.Clone(g.labelValues), labelValues...),
	}
}

func (g *statefulGauge) Set(value float64) {
	g.update(func(float64) float64 { return value })
}

func (g *statefulGauge) Add(delta float64) {
	g.update(func(value float64) float64 { return value + delta })
}

func (g *statefulGauge) update(fn func(float64) float64) {
	// The wrapped gauge is set under the lock,
	// so that concurrent updates reach it in the same order as the stored values.
	g.state.mu.Lock()
	defer g.state.mu.Unlock()

	key := strings.Join(g.labelValues, "\x00")
	value := fn(g.state.values[key])
	g.state.values[key] = value

	g.gauge.Set(value)
}
