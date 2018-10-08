package retry // import "gopkg.in/retry.v1"

import "time"

// Clock represents a virtual clock interface that
// can be replaced for testing.
type Clock interface {
	Now() time.Time
	After(time.Duration) <-chan time.Time
}

// WallClock exposes wall-clock time as returned by time.Now.
type wallClock struct{}

// Now implements Clock.Now.
func (wallClock) Now() time.Time {
	return time.Now()
}

// After implements Clock.After.
func (wallClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}
