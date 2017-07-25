package generic_test

// This is package generic_test in order to get around an import cycle: this
// package imports teststat to do its testing, but package teststat imports
// generic to use its Histogram in the Quantiles helper function.

import (
	"math"
	"math/rand"
	"testing"

	"github.com/go-kit/kit/metrics/generic"
	"github.com/go-kit/kit/metrics/teststat"
)

func TestCounter(t *testing.T) {
	counter := generic.NewCounter("my_counter").With("label", "counter").(*generic.Counter)
	value := func() float64 { return counter.Value() }
	if err := teststat.TestCounter(counter, value); err != nil {
		t.Fatal(err)
	}
}

func TestValueReset(t *testing.T) {
	counter := generic.NewCounter("test_value_reset")
	counter.Add(123)
	counter.Add(456)
	counter.Add(789)
	if want, have := float64(123+456+789), counter.ValueReset(); want != have {
		t.Errorf("want %f, have %f", want, have)
	}
	if want, have := float64(0), counter.Value(); want != have {
		t.Errorf("want %f, have %f", want, have)
	}
}

func TestGauge(t *testing.T) {
	gauge := generic.NewGauge("my_gauge").With("label", "gauge").(*generic.Gauge)
	value := func() float64 { return gauge.Value() }
	if err := teststat.TestGauge(gauge, value); err != nil {
		t.Fatal(err)
	}
}

func TestHistogram(t *testing.T) {
	histogram := generic.NewHistogram("my_histogram", 50).With("label", "histogram").(*generic.Histogram)
	quantiles := func() (float64, float64, float64, float64) {
		return histogram.Quantile(0.50), histogram.Quantile(0.90), histogram.Quantile(0.95), histogram.Quantile(0.99)
	}
	if err := teststat.TestHistogram(histogram, quantiles, 0.01); err != nil {
		t.Fatal(err)
	}
}

func TestSimpleHistogram(t *testing.T) {
	histogram := generic.NewSimpleHistogram().With("label", "simple_histogram").(*generic.SimpleHistogram)
	var (
		sum   int
		count = 1234 // not too big
	)
	for i := 0; i < count; i++ {
		value := rand.Intn(1000)
		sum += value
		histogram.Observe(float64(value))
	}

	var (
		want      = float64(sum) / float64(count)
		have      = histogram.ApproximateMovingAverage()
		tolerance = 0.001 // real real slim
	)
	if math.Abs(want-have)/want > tolerance {
		t.Errorf("want %f, have %f", want, have)
	}
}
