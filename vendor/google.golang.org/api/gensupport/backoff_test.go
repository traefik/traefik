package gensupport

import (
	"testing"
	"time"
)

func TestBackoff(t *testing.T) {
	eb := &ExponentialBackoff{Base: time.Millisecond, Max: time.Second}

	var total time.Duration
	for n, max := 0, 2*time.Millisecond; ; n, max = n+1, max*2 {
		if n > 100 {
			// There's less than 1 in 10^28 of taking longer than 100 iterations,
			// so this is just to check we don't have an infinite loop.
			t.Fatalf("Failed to timeout after 100 iterations.")
		}
		pause, retry := eb.Pause()
		if !retry {
			break
		}

		if 0 > pause || pause >= max {
			t.Errorf("Iteration %d: pause = %v; want in range [0, %v)", n, pause, max)
		}
		total += pause
	}

	if total < time.Second {
		t.Errorf("Total time = %v; want > %v", total, time.Second)
	}
}
