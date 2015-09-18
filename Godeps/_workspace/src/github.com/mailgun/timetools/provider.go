package timetools

import (
	"sync"
	"time"
)

// TimeProvider is an interface we use to mock time in tests.
type TimeProvider interface {
	UtcNow() time.Time
	Sleep(time.Duration)
	After(time.Duration) <-chan time.Time
}

// RealTime is a real clock time, used in production.
type RealTime struct {
}

func (*RealTime) UtcNow() time.Time {
	return time.Now().UTC()
}

func (*RealTime) Sleep(d time.Duration) {
	time.Sleep(d)
}

func (*RealTime) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

// FreezedTime is manually controlled time for use in tests.
type FreezedTime struct {
	CurrentTime time.Time
}

func (t *FreezedTime) UtcNow() time.Time {
	return t.CurrentTime
}

func (t *FreezedTime) Sleep(d time.Duration) {
	t.CurrentTime = t.CurrentTime.Add(d)
}

func (t *FreezedTime) After(d time.Duration) <-chan time.Time {
	t.Sleep(d)
	c := make(chan time.Time, 1)
	c <- t.CurrentTime
	return c
}

type sleepableTime struct {
	currentTime time.Time
	waiters     map[time.Time][]chan time.Time
	mu          sync.Mutex
}

// SleepProvider returns a TimeProvider that has good fakes for
// time.Sleep and time.After. Both functions will behave as if
// time is frozen until you call AdvanceTimeBy, at which point
// any calls to time.Sleep that should return do return and
// any ticks from time.After that should happen do happen.
func SleepProvider(currentTime time.Time) TimeProvider {
	return &sleepableTime{
		currentTime: currentTime,
		waiters:     make(map[time.Time][]chan time.Time),
	}
}

func (t *sleepableTime) UtcNow() time.Time {
	return t.currentTime
}

func (t *sleepableTime) Sleep(d time.Duration) {
	<-t.After(d)
}

func (t *sleepableTime) After(d time.Duration) <-chan time.Time {
	t.mu.Lock()
	defer t.mu.Unlock()

	c := make(chan time.Time, 1)
	until := t.currentTime.Add(d)
	t.waiters[until] = append(t.waiters[until], c)
	return c
}

// AdvanceTimeBy simulates advancing time by some time.Duration d.
// This function panics if st is not the result of a call to
// SleepProvider.
func AdvanceTimeBy(st TimeProvider, d time.Duration) {
	t := st.(*sleepableTime)
	t.mu.Lock()
	defer t.mu.Unlock()

	t.currentTime = t.currentTime.Add(d)
	for k, v := range t.waiters {
		if k.Before(t.currentTime) {
			for _, c := range v {
				c <- t.currentTime
			}
			delete(t.waiters, k)
		}
	}
}
