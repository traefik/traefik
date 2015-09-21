package cbreaker

import (
	"fmt"
	"time"

	"github.com/mailgun/log"
	"github.com/mailgun/timetools"
)

// ratioController allows passing portions traffic back to the endpoints,
// increasing the amount of passed requests using linear function:
//
//   allowedRequestsRatio = 0.5 * (Now() - Start())/Duration
//
type ratioController struct {
	duration time.Duration
	start    time.Time
	tm       timetools.TimeProvider
	allowed  int
	denied   int
}

func newRatioController(tm timetools.TimeProvider, rampUp time.Duration) *ratioController {
	return &ratioController{
		duration: rampUp,
		tm:       tm,
		start:    tm.UtcNow(),
	}
}

func (r *ratioController) String() string {
	return fmt.Sprintf("RatioController(target=%f, current=%f, allowed=%d, denied=%d)", r.targetRatio(), r.computeRatio(r.allowed, r.denied), r.allowed, r.denied)
}

func (r *ratioController) allowRequest() bool {
	log.Infof("%v", r)
	t := r.targetRatio()
	// This condition answers the question - would we satisfy the target ratio if we allow this request?
	e := r.computeRatio(r.allowed+1, r.denied)
	if e < t {
		r.allowed++
		log.Infof("%v allowed", r)
		return true
	}
	r.denied++
	log.Infof("%v denied", r)
	return false
}

func (r *ratioController) computeRatio(allowed, denied int) float64 {
	if denied+allowed == 0 {
		return 0
	}
	return float64(allowed) / float64(denied+allowed)
}

func (r *ratioController) targetRatio() float64 {
	// Here's why it's 0.5:
	// We are watching the following ratio
	// ratio = a / (a + d)
	// We can notice, that once we get to 0.5
	// 0.5 = a / (a + d)
	// we can evaluate that a = d
	// that means equilibrium, where we would allow all the requests
	// after this point to achieve ratio of 1 (that can never be reached unless d is 0)
	// so we stop from there
	multiplier := 0.5 / float64(r.duration)
	return multiplier * float64(r.tm.UtcNow().Sub(r.start))
}
