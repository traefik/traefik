package types

import "time"

// Clock provides injectable time (supports testing)
type Clock interface {
	Now() time.Time
}

type normalClock struct{}

func (c normalClock) Now() time.Time {
	return time.Now()
}

// TheClock - replaceable during testing
var TheClock Clock = normalClock{}
