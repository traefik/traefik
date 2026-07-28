package job

import (
	"time"

	"github.com/cenkalti/backoff/v4"
)

var _ backoff.BackOff = (*BackOff)(nil)

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

	next := b.ExponentialBackOff.NextBackOff()

	// Defensive guard: if the underlying backoff returns Stop (which can
	// happen despite MaxElapsedTime=0 under certain edge conditions such as
	// the wrapped context being cancelled), reset and restart from initial
	// interval so that long-running jobs never permanently stop retrying.
	if next == backoff.Stop {
		b.Reset()
		return b.InitialInterval
	}

	return next
}
