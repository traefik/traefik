package appinsights

// We need to mock out the clock for tests; we'll use this to do it.

import "code.cloudfoundry.org/clock"

var currentClock clock.Clock

func init() {
	currentClock = clock.NewClock()
}
