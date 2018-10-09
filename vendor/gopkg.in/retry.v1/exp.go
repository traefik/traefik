package retry // import "gopkg.in/retry.v1"

import (
	"math/rand"
	"sync"
	"time"
)

var (
	// randomMu guards random.
	randomMu sync.Mutex
	// random is used as a random number source for jitter.
	// We avoid using the global math/rand source
	// as we don't want to be responsible for seeding it,
	// and its lock may be more contended.
	random = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// Exponential represents an exponential backoff retry strategy.
// To limit the number of attempts or their overall duration, wrap
// this in LimitCount or LimitDuration.
type Exponential struct {
	// Initial holds the initial delay.
	Initial time.Duration
	// Factor holds the factor that the delay time will be multiplied
	// by on each iteration. If this is zero, a factor of two will be used.
	Factor float64
	// MaxDelay holds the maximum delay between the start
	// of attempts. If this is zero, there is no maximum delay.
	MaxDelay time.Duration
	// Jitter specifies whether jitter should be added to the
	// retry interval. The algorithm used is described as "Full Jitter"
	// in https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/
	Jitter bool
}

type exponentialTimer struct {
	strategy Exponential
	start    time.Time
	end      time.Time
	delay    time.Duration
}

// NewTimer implements Strategy.NewTimer.
func (r Exponential) NewTimer(now time.Time) Timer {
	if r.Factor <= 0 {
		r.Factor = 2
	}
	return &exponentialTimer{
		strategy: r,
		start:    now,
		delay:    r.Initial,
	}
}

// NextSleep implements Timer.NextSleep.
func (a *exponentialTimer) NextSleep(now time.Time) (time.Duration, bool) {
	sleep := a.delay - now.Sub(a.start)
	if sleep <= 0 {
		sleep = 0
	}
	if a.strategy.Jitter {
		sleep = randDuration(sleep)
	}
	// Set the start of the next try.
	a.start = now.Add(sleep)
	a.delay = time.Duration(float64(a.delay) * a.strategy.Factor)
	if a.strategy.MaxDelay > 0 && a.delay > a.strategy.MaxDelay {
		a.delay = a.strategy.MaxDelay
	}
	return sleep, true
}

func randDuration(max time.Duration) time.Duration {
	if max <= 0 {
		return 0
	}
	randomMu.Lock()
	defer randomMu.Unlock()
	return time.Duration(random.Int63n(int64(max)))
}
