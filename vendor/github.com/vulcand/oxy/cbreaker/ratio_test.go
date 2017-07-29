package cbreaker

import (
	"math"
	"time"

	"github.com/mailgun/timetools"
	. "gopkg.in/check.v1"
)

type RatioSuite struct {
	tm *timetools.FreezedTime
}

var _ = Suite(&RatioSuite{
	tm: &timetools.FreezedTime{
		CurrentTime: time.Date(2012, 3, 4, 5, 6, 7, 0, time.UTC),
	},
})

func (s *RatioSuite) advanceTime(d time.Duration) {
	s.tm.CurrentTime = s.tm.CurrentTime.Add(d)
}

func (s *RatioSuite) TestRampUp(c *C) {
	duration := 10 * time.Second
	rc := newRatioController(s.tm, duration)

	allowed, denied := 0, 0
	for i := 0; i < int(duration/time.Millisecond); i++ {
		ratio := s.sendRequest(&allowed, &denied, rc)
		expected := rc.targetRatio()
		diff := math.Abs(expected - ratio)
		c.Assert(round(diff, 0.5, 1), Equals, float64(0))
		s.advanceTime(time.Millisecond)
	}
}

func (s *RatioSuite) sendRequest(allowed, denied *int, rc *ratioController) float64 {
	if rc.allowRequest() {
		*allowed++
	} else {
		*denied++
	}
	if *allowed+*denied == 0 {
		return 0
	}
	return float64(*allowed) / float64(*allowed+*denied)
}

func round(val float64, roundOn float64, places int) float64 {
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	var round float64
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	return round / pow
}
