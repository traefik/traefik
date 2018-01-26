package cony

import (
	"math/rand"
	"time"
)

// DefaultBackoff See: http://blog.gopheracademy.com/advent-2014/backoff/
var DefaultBackoff Backoffer = BackoffPolicy{
	[]int{0, 10, 100, 200, 500, 1000, 2000, 3000, 5000},
}

// Backoffer is interface to hold Backoff strategy
type Backoffer interface {
	Backoff(int) time.Duration
}

// BackoffPolicy is a default Backoffer implementation
type BackoffPolicy struct {
	ms []int
}

// Backoff implements Backoffer
func (b BackoffPolicy) Backoff(n int) time.Duration {
	if n >= len(b.ms) {
		n = len(b.ms) - 1
	}

	return time.Duration(jitter(b.ms[n])) * time.Millisecond
}

func jitter(ms int) int {
	if ms == 0 {
		return 0
	}

	return ms/2 + rand.Intn(ms)
}
