package tracer

import (
	"encoding/json"
	"io"
	"math"
	"sync"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
)

// Sampler is the generic interface of any sampler. It must be safe for concurrent use.
type Sampler interface {
	// Sample returns true if the given span should be sampled.
	Sample(span Span) bool
}

// RateSampler is a sampler implementation which randomly selects spans using a
// provided rate. For example, a rate of 0.75 will permit 75% of the spans.
// RateSampler implementations should be safe for concurrent use.
type RateSampler interface {
	Sampler

	// Rate returns the current sample rate.
	Rate() float64

	// SetRate sets a new sample rate.
	SetRate(rate float64)
}

// rateSampler samples from a sample rate.
type rateSampler struct {
	sync.RWMutex
	rate float64
}

// NewAllSampler is a short-hand for NewRateSampler(1). It is all-permissive.
func NewAllSampler() RateSampler { return NewRateSampler(1) }

// NewRateSampler returns an initialized RateSampler with a given sample rate.
func NewRateSampler(rate float64) RateSampler {
	return &rateSampler{rate: rate}
}

// Rate returns the current rate of the sampler.
func (r *rateSampler) Rate() float64 {
	r.RLock()
	defer r.RUnlock()
	return r.rate
}

// SetRate sets a new sampling rate.
func (r *rateSampler) SetRate(rate float64) {
	r.Lock()
	r.rate = rate
	r.Unlock()
}

// constants used for the Knuth hashing, same as agent.
const knuthFactor = uint64(1111111111111111111)

// Sample returns true if the given span should be sampled.
func (r *rateSampler) Sample(spn ddtrace.Span) bool {
	if r.rate == 1 {
		// fast path
		return true
	}
	s, ok := spn.(*span)
	if !ok {
		return false
	}
	r.RLock()
	defer r.RUnlock()
	return sampledByRate(s.TraceID, r.rate)
}

// sampledByRate verifies if the number n should be sampled at the specified
// rate.
func sampledByRate(n uint64, rate float64) bool {
	if rate < 1 {
		return n*knuthFactor < uint64(rate*math.MaxUint64)
	}
	return true
}

// prioritySampler holds a set of per-service sampling rates and applies
// them to spans.
type prioritySampler struct {
	mu          sync.RWMutex
	rates       map[string]float64
	defaultRate float64
}

func newPrioritySampler() *prioritySampler {
	return &prioritySampler{
		rates:       make(map[string]float64),
		defaultRate: 1.,
	}
}

// readRatesJSON will try to read the rates as JSON from the given io.ReadCloser.
func (ps *prioritySampler) readRatesJSON(rc io.ReadCloser) error {
	var payload struct {
		Rates map[string]float64 `json:"rate_by_service"`
	}
	if err := json.NewDecoder(rc).Decode(&payload); err != nil {
		return err
	}
	rc.Close()
	const defaultRateKey = "service:,env:"
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.rates = payload.Rates
	if v, ok := ps.rates[defaultRateKey]; ok {
		ps.defaultRate = v
		delete(ps.rates, defaultRateKey)
	}
	return nil
}

// getRate returns the sampling rate to be used for the given span. Callers must
// guard the span.
func (ps *prioritySampler) getRate(spn *span) float64 {
	key := "service:" + spn.Service + ",env:" + spn.Meta[ext.Environment]
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	if rate, ok := ps.rates[key]; ok {
		return rate
	}
	return ps.defaultRate
}

// apply applies sampling priority to the given span. Caller must ensure it is safe
// to modify the span.
func (ps *prioritySampler) apply(spn *span) {
	rate := ps.getRate(spn)
	if sampledByRate(spn.TraceID, rate) {
		spn.SetTag(ext.SamplingPriority, ext.PriorityAutoKeep)
	} else {
		spn.SetTag(ext.SamplingPriority, ext.PriorityAutoReject)
	}
	spn.SetTag(keySamplingPriorityRate, rate)
}
