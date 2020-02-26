package job

import (
	"time"

	"github.com/cenkalti/backoff/v4"
)

var (
	_ backoff.BackOff = (*BackOff)(nil)
)

const (
	defaultMinJobInterval = 30 * time.Second
)

// BackOff is an exponential backoff implementation for long running jobs.
// In long running jobs, an operation() that fails after a long Duration should not increments the backoff period.
// If operation() takes more than MinJobInterval, Reset() is called in NextBackOff().
type BackOff struct {
	*backoff.ExponentialBackOff
	MinJobInterval time.Duration
}

// NewBackOff creates an instance of BackOff using default values.
func NewBackOff(backOff *backoff.ExponentialBackOff) *BackOff {
	backOff.MaxElapsedTime = 0
	return &BackOff{
		ExponentialBackOff: backOff,
		MinJobInterval:     defaultMinJobInterval,
	}
}

// NextBackOff calculates the next backoff interval.
func (b *BackOff) NextBackOff() time.Duration {
	if b.GetElapsedTime() >= b.MinJobInterval {
		b.Reset()
	}
	return b.ExponentialBackOff.NextBackOff()
}
