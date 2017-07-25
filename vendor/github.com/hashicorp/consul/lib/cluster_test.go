package lib

import (
	"testing"
	"time"
)

func TestDurationMinusBuffer(t *testing.T) {
	tests := []struct {
		Duration time.Duration
		Buffer   time.Duration
		Jitter   int64
	}{
		{
			Duration: 1 * time.Minute,
			Buffer:   10 * time.Second,
			Jitter:   16,
		},
		{
			Duration: 1 * time.Second,
			Buffer:   500 * time.Millisecond,
			Jitter:   4,
		},
		{
			Duration: 1 * time.Second,
			Buffer:   1 * time.Second,
			Jitter:   4,
		},
		{
			Duration: 1 * time.Second,
			Buffer:   1 * time.Second,
			Jitter:   0,
		},
		{
			Duration: 1 * time.Second,
			Buffer:   1 * time.Second,
			Jitter:   1,
		},
	}

	for _, test := range tests {
		min, max := DurationMinusBufferDomain(test.Duration, test.Buffer, test.Jitter)
		for i := 0; i < 10; i++ {
			d := DurationMinusBuffer(test.Duration, test.Buffer, test.Jitter)
			if d < min || d > max {
				t.Fatalf("Bad: %v", d)
			}
		}
	}
}

func TestDurationMinusBufferDomain(t *testing.T) {
	tests := []struct {
		Duration time.Duration
		Buffer   time.Duration
		Jitter   int64
		Min      time.Duration
		Max      time.Duration
	}{
		{
			Duration: 60 * time.Second,
			Buffer:   10 * time.Second,
			Jitter:   16,
			Min:      46*time.Second + 875*time.Millisecond,
			Max:      50 * time.Second,
		},
		{
			Duration: 60 * time.Second,
			Buffer:   0 * time.Second,
			Jitter:   16,
			Min:      56*time.Second + 250*time.Millisecond,
			Max:      60 * time.Second,
		},
		{
			Duration: 60 * time.Second,
			Buffer:   0 * time.Second,
			Jitter:   0,
			Min:      60 * time.Second,
			Max:      60 * time.Second,
		},
		{
			Duration: 60 * time.Second,
			Buffer:   0 * time.Second,
			Jitter:   1,
			Min:      0 * time.Second,
			Max:      60 * time.Second,
		},
		{
			Duration: 60 * time.Second,
			Buffer:   0 * time.Second,
			Jitter:   2,
			Min:      30 * time.Second,
			Max:      60 * time.Second,
		},
		{
			Duration: 60 * time.Second,
			Buffer:   0 * time.Second,
			Jitter:   4,
			Min:      45 * time.Second,
			Max:      60 * time.Second,
		},
		{
			Duration: 0 * time.Second,
			Buffer:   0 * time.Second,
			Jitter:   0,
			Min:      0 * time.Second,
			Max:      0 * time.Second,
		},
		{
			Duration: 60 * time.Second,
			Buffer:   120 * time.Second,
			Jitter:   8,
			Min:      -1 * (52*time.Second + 500*time.Millisecond),
			Max:      -1 * 60 * time.Second,
		},
	}

	for _, test := range tests {
		min, max := DurationMinusBufferDomain(test.Duration, test.Buffer, test.Jitter)
		if min != test.Min {
			t.Errorf("Bad min: %v != %v", min, test.Min)
		}

		if max != test.Max {
			t.Errorf("Bad max: %v != %v", max, test.Max)
		}
	}
}

func TestRandomStagger(t *testing.T) {
	intv := time.Minute
	for i := 0; i < 10; i++ {
		stagger := RandomStagger(intv)
		if stagger < 0 || stagger >= intv {
			t.Fatalf("Bad: %v", stagger)
		}
	}
}

func TestRateScaledInterval(t *testing.T) {
	const min = 1 * time.Second
	rate := 200.0
	if v := RateScaledInterval(rate, min, 0); v != min {
		t.Fatalf("Bad: %v", v)
	}
	if v := RateScaledInterval(rate, min, 100); v != min {
		t.Fatalf("Bad: %v", v)
	}
	if v := RateScaledInterval(rate, min, 200); v != 1*time.Second {
		t.Fatalf("Bad: %v", v)
	}
	if v := RateScaledInterval(rate, min, 1000); v != 5*time.Second {
		t.Fatalf("Bad: %v", v)
	}
	if v := RateScaledInterval(rate, min, 5000); v != 25*time.Second {
		t.Fatalf("Bad: %v", v)
	}
	if v := RateScaledInterval(rate, min, 10000); v != 50*time.Second {
		t.Fatalf("Bad: %v", v)
	}
	if v := RateScaledInterval(0, min, 10000); v != min {
		t.Fatalf("Bad: %v", v)
	}
	if v := RateScaledInterval(0.0, min, 10000); v != min {
		t.Fatalf("Bad: %v", v)
	}
	if v := RateScaledInterval(-1, min, 10000); v != min {
		t.Fatalf("Bad: %v", v)
	}
}
