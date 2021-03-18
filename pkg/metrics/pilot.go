package metrics

import (
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/generic"
)

const (
	// server meta information.
	pilotConfigPrefix                   = "config"
	pilotConfigReloadsTotalName         = pilotConfigPrefix + "ReloadsTotal"
	pilotConfigReloadsFailuresTotalName = pilotConfigPrefix + "ReloadsFailureTotal"
	pilotConfigLastReloadSuccessName    = pilotConfigPrefix + "LastReloadSuccess"
	pilotConfigLastReloadFailureName    = pilotConfigPrefix + "LastReloadFailure"

	// entry point.
	pilotEntryPointPrefix           = "entrypoint"
	pilotEntryPointReqsTotalName    = pilotEntryPointPrefix + "RequestsTotal"
	pilotEntryPointReqsTLSTotalName = pilotEntryPointPrefix + "RequestsTLSTotal"
	pilotEntryPointReqDurationName  = pilotEntryPointPrefix + "RequestDurationSeconds"
	pilotEntryPointOpenConnsName    = pilotEntryPointPrefix + "OpenConnections"

	// service level.
	pilotServicePrefix           = "service"
	pilotServiceReqsTotalName    = pilotServicePrefix + "RequestsTotal"
	pilotServiceReqsTLSTotalName = pilotServicePrefix + "RequestsTLSTotal"
	pilotServiceReqDurationName  = pilotServicePrefix + "RequestDurationSeconds"
	pilotServiceOpenConnsName    = pilotServicePrefix + "OpenConnections"
	pilotServiceRetriesTotalName = pilotServicePrefix + "RetriesTotal"
	pilotServiceServerUpName     = pilotServicePrefix + "ServerUp"
)

const root = "value"

// RegisterPilot registers all Pilot metrics.
func RegisterPilot() *PilotRegistry {
	standardRegistry := &standardRegistry{
		epEnabled:  true,
		svcEnabled: true,
	}

	pr := &PilotRegistry{
		standardRegistry: standardRegistry,
		counters:         make(map[string]*pilotCounter),
		gauges:           make(map[string]*pilotGauge),
		histograms:       make(map[string]*pilotHistogram),
	}

	standardRegistry.configReloadsCounter = pr.newCounter(pilotConfigReloadsTotalName)
	standardRegistry.configReloadsFailureCounter = pr.newCounter(pilotConfigReloadsFailuresTotalName)
	standardRegistry.lastConfigReloadSuccessGauge = pr.newGauge(pilotConfigLastReloadSuccessName)
	standardRegistry.lastConfigReloadFailureGauge = pr.newGauge(pilotConfigLastReloadFailureName)

	standardRegistry.entryPointReqsCounter = pr.newCounter(pilotEntryPointReqsTotalName)
	standardRegistry.entryPointReqsTLSCounter = pr.newCounter(pilotEntryPointReqsTLSTotalName)
	standardRegistry.entryPointReqDurationHistogram, _ = NewHistogramWithScale(pr.newHistogram(pilotEntryPointReqDurationName), time.Millisecond)
	standardRegistry.entryPointOpenConnsGauge = pr.newGauge(pilotEntryPointOpenConnsName)

	standardRegistry.serviceReqsCounter = pr.newCounter(pilotServiceReqsTotalName)
	standardRegistry.serviceReqsTLSCounter = pr.newCounter(pilotServiceReqsTLSTotalName)
	standardRegistry.serviceReqDurationHistogram, _ = NewHistogramWithScale(pr.newHistogram(pilotServiceReqDurationName), time.Millisecond)
	standardRegistry.serviceOpenConnsGauge = pr.newGauge(pilotServiceOpenConnsName)
	standardRegistry.serviceRetriesCounter = pr.newCounter(pilotServiceRetriesTotalName)
	standardRegistry.serviceServerUpGauge = pr.newGauge(pilotServiceServerUpName)

	return pr
}

// PilotMetric is a representation of a metric.
type PilotMetric struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Observations map[string]interface{} `json:"observations"`
}

type pilotHistogramObservation struct {
	Total float64 `json:"total"`
	Count float64 `json:"count"`
}

// PilotRegistry represents the pilots metrics registry.
type PilotRegistry struct {
	counters   map[string]*pilotCounter
	gauges     map[string]*pilotGauge
	histograms map[string]*pilotHistogram

	*standardRegistry
}

// newCounter register and returns a new pilotCounter.
func (pr *PilotRegistry) newCounter(name string) *pilotCounter {
	c := newPilotCounter(name)
	pr.counters[name] = c

	return c
}

// newGauge register and returns a new pilotGauge.
func (pr *PilotRegistry) newGauge(name string) *pilotGauge {
	g := newPilotGauge(name)
	pr.gauges[name] = g

	return g
}

// newHistogram register and returns a new pilotHistogram.
func (pr *PilotRegistry) newHistogram(name string) *pilotHistogram {
	h := newPilotHistogram(name)
	pr.histograms[name] = h

	return h
}

// Data exports the metrics: metrics name -> labels -> values.
func (pr *PilotRegistry) Data() []PilotMetric {
	var pilotMetrics []PilotMetric

	for name, counter := range pr.counters {
		pilotMetric := PilotMetric{
			Name:         name,
			Type:         "COUNTER",
			Observations: make(map[string]interface{}),
		}
		pilotMetrics = append(pilotMetrics, pilotMetric)

		counter.counters.Range(func(key, value interface{}) bool {
			labels := key.(string)
			pc := value.(*pilotCounter)

			if labels == "" {
				labels = root
			}

			if labels == root || len(pc.c.LabelValues())%2 == 0 {
				pilotMetric.Observations[labels] = pc.c.Value()
			}

			return true
		})
	}

	for name, gauge := range pr.gauges {
		pilotMetric := PilotMetric{
			Name:         name,
			Type:         "GAUGE",
			Observations: make(map[string]interface{}),
		}
		pilotMetrics = append(pilotMetrics, pilotMetric)

		gauge.gauges.Range(func(key, value interface{}) bool {
			labels := key.(string)
			pg := value.(*pilotGauge)

			if labels == "" {
				labels = root
			}

			if labels == root || len(pg.g.LabelValues())%2 == 0 {
				pilotMetric.Observations[labels] = pg.g.Value()
			}

			return true
		})
	}

	for name, histogram := range pr.histograms {
		pilotMetric := PilotMetric{
			Name:         name,
			Type:         "HISTOGRAM",
			Observations: make(map[string]interface{}),
		}
		pilotMetrics = append(pilotMetrics, pilotMetric)

		histogram.histograms.Range(func(key, value interface{}) bool {
			labels := key.(string)
			ph := value.(*pilotHistogram)

			if labels == "" {
				labels = root
			}

			if labels == root || len(ph.labels)%2 == 0 {
				pilotMetric.Observations[labels] = &pilotHistogramObservation{
					Total: ph.total.Value(),
					Count: ph.count.Value(),
				}
			}

			return true
		})
	}

	return pilotMetrics
}

type pilotCounter struct {
	c        *generic.Counter
	counters *sync.Map
}

func newPilotCounter(name string) *pilotCounter {
	return &pilotCounter{
		c:        generic.NewCounter(name),
		counters: &sync.Map{},
	}
}

// With returns a new pilotCounter with the given labels.
func (c *pilotCounter) With(labels ...string) metrics.Counter {
	newCounter := c.c.With(labels...).(*generic.Counter)
	newCounter.ValueReset()

	return &pilotCounter{
		c:        newCounter,
		counters: c.counters,
	}
}

// Add adds the given delta to the counter.
func (c *pilotCounter) Add(delta float64) {
	labelsKey := strings.Join(c.c.LabelValues(), ",")

	pc, _ := c.counters.LoadOrStore(labelsKey, c)

	pc.(*pilotCounter).c.Add(delta)
}

type pilotGauge struct {
	g      *generic.Gauge
	gauges *sync.Map
}

func newPilotGauge(name string) *pilotGauge {
	return &pilotGauge{
		g:      generic.NewGauge(name),
		gauges: &sync.Map{},
	}
}

// With returns a new pilotGauge with the given labels.
func (g *pilotGauge) With(labels ...string) metrics.Gauge {
	newGauge := g.g.With(labels...).(*generic.Gauge)
	newGauge.Set(0)

	return &pilotGauge{
		g:      newGauge,
		gauges: g.gauges,
	}
}

// Set sets the given value to the gauge.
func (g *pilotGauge) Set(value float64) {
	labelsKey := strings.Join(g.g.LabelValues(), ",")

	pg, _ := g.gauges.LoadOrStore(labelsKey, g)

	pg.(*pilotGauge).g.Set(value)
}

// Add adds the given delta to the gauge.
func (g *pilotGauge) Add(delta float64) {
	labelsKey := strings.Join(g.g.LabelValues(), ",")

	pg, _ := g.gauges.LoadOrStore(labelsKey, g)

	pg.(*pilotGauge).g.Add(delta)
}

type pilotHistogram struct {
	name       string
	labels     []string
	count      *generic.Counter
	total      *generic.Counter
	histograms *sync.Map
}

func newPilotHistogram(name string) *pilotHistogram {
	return &pilotHistogram{
		name:       name,
		labels:     make([]string, 0),
		count:      &generic.Counter{},
		total:      &generic.Counter{},
		histograms: &sync.Map{},
	}
}

// With returns a new pilotHistogram with the given labels.
func (h *pilotHistogram) With(labels ...string) metrics.Histogram {
	var newLabels []string

	newLabels = append(newLabels, h.labels...)
	newLabels = append(newLabels, labels...)

	return &pilotHistogram{
		name:       h.name,
		labels:     newLabels,
		count:      &generic.Counter{},
		total:      &generic.Counter{},
		histograms: h.histograms,
	}
}

// Observe records a new value into the histogram.
func (h *pilotHistogram) Observe(value float64) {
	labelsKey := strings.Join(h.labels, ",")

	ph, _ := h.histograms.LoadOrStore(labelsKey, h)

	pHisto := ph.(*pilotHistogram)

	pHisto.count.Add(1)
	pHisto.total.Add(value)
}
