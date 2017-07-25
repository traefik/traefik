package metrics_test

import (
	"math"
	"testing"

	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/generic"
)

func TestTimerFast(t *testing.T) {
	h := generic.NewSimpleHistogram()
	metrics.NewTimer(h).ObserveDuration()

	tolerance := 0.050
	if want, have := 0.000, h.ApproximateMovingAverage(); math.Abs(want-have) > tolerance {
		t.Errorf("want %.3f, have %.3f", want, have)
	}
}

func TestTimerSlow(t *testing.T) {
	h := generic.NewSimpleHistogram()
	timer := metrics.NewTimer(h)
	time.Sleep(250 * time.Millisecond)
	timer.ObserveDuration()

	tolerance := 0.050
	if want, have := 0.250, h.ApproximateMovingAverage(); math.Abs(want-have) > tolerance {
		t.Errorf("want %.3f, have %.3f", want, have)
	}
}
