package job

import (
	"testing"
	"time"

	"github.com/cenk/backoff"
)

func TestJobBackOff(t *testing.T) {
	var (
		testInitialInterval     = 500 * time.Millisecond
		testRandomizationFactor = 0.1
		testMultiplier          = 2.0
		testMaxInterval         = 5 * time.Second
		testMinJobInterval      = 1 * time.Second
	)

	exp := NewBackOff(backoff.NewExponentialBackOff())
	exp.InitialInterval = testInitialInterval
	exp.RandomizationFactor = testRandomizationFactor
	exp.Multiplier = testMultiplier
	exp.MaxInterval = testMaxInterval
	exp.MinJobInterval = testMinJobInterval
	exp.Reset()

	var expectedResults = []time.Duration{500, 500, 500, 1000, 2000, 4000, 5000, 5000, 500, 1000, 2000, 4000, 5000, 5000}
	for i, d := range expectedResults {
		expectedResults[i] = d * time.Millisecond
	}

	for i, expected := range expectedResults {
		// Assert that the next backoff falls in the expected range.
		var minInterval = expected - time.Duration(testRandomizationFactor*float64(expected))
		var maxInterval = expected + time.Duration(testRandomizationFactor*float64(expected))
		if i < 3 || i == 8 {
			time.Sleep(2 * time.Second)
		}
		var actualInterval = exp.NextBackOff()
		if !(minInterval <= actualInterval && actualInterval <= maxInterval) {
			t.Error("error")
		}
		// assertEquals(t, expected, exp.currentInterval)
	}
}
