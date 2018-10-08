package retry // import "gopkg.in/retry.v1"

import (
	"time"
)

type strategyFunc func(now time.Time) Timer

// NewTimer implements Strategy.NewTimer.
func (f strategyFunc) NewTimer(now time.Time) Timer {
	return f(now)
}

// LimitCount limits the number of attempts that the given
// strategy will perform to n. Note that all strategies
// will allow at least one attempt.
func LimitCount(n int, strategy Strategy) Strategy {
	return strategyFunc(func(now time.Time) Timer {
		return &countLimitTimer{
			timer:  strategy.NewTimer(now),
			remain: n,
		}
	})
}

type countLimitTimer struct {
	timer  Timer
	remain int
}

func (t *countLimitTimer) NextSleep(now time.Time) (time.Duration, bool) {
	if t.remain--; t.remain <= 0 {
		return 0, false
	}
	return t.timer.NextSleep(now)
}

// LimitTime limits the given strategy such that no attempt will
// made after the given duration has elapsed.
func LimitTime(limit time.Duration, strategy Strategy) Strategy {
	return strategyFunc(func(now time.Time) Timer {
		return &timeLimitTimer{
			timer: strategy.NewTimer(now),
			end:   now.Add(limit),
		}
	})
}

type timeLimitTimer struct {
	timer Timer
	end   time.Time
}

func (t *timeLimitTimer) NextSleep(now time.Time) (time.Duration, bool) {
	sleep, ok := t.timer.NextSleep(now)
	if ok && now.Add(sleep).After(t.end) {
		return 0, false
	}
	return sleep, ok
}
