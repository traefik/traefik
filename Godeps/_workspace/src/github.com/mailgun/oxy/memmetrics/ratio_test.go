package memmetrics

import (
	"testing"
	"time"

	"github.com/mailgun/timetools"
	. "gopkg.in/check.v1"
)

func TestFailrate(t *testing.T) { TestingT(t) }

type FailRateSuite struct {
	tm *timetools.FreezedTime
}

var _ = Suite(&FailRateSuite{})

func (s *FailRateSuite) SetUpSuite(c *C) {
	s.tm = &timetools.FreezedTime{
		CurrentTime: time.Date(2012, 3, 4, 5, 6, 7, 0, time.UTC),
	}
}

func (s *FailRateSuite) TestInvalidParams(c *C) {
	// Bad buckets count
	_, err := NewRatioCounter(0, time.Second, RatioClock(s.tm))
	c.Assert(err, Not(IsNil))

	// Too precise resolution
	_, err = NewRatioCounter(10, time.Millisecond, RatioClock(s.tm))
	c.Assert(err, Not(IsNil))
}

func (s *FailRateSuite) TestNotReady(c *C) {
	// No data
	fr, err := NewRatioCounter(10, time.Second, RatioClock(s.tm))
	c.Assert(err, IsNil)
	c.Assert(fr.IsReady(), Equals, false)
	c.Assert(fr.Ratio(), Equals, 0.0)

	// Not enough data
	fr, err = NewRatioCounter(10, time.Second, RatioClock(s.tm))
	c.Assert(err, IsNil)
	fr.CountA()
	c.Assert(fr.IsReady(), Equals, false)
}

func (s *FailRateSuite) TestNoB(c *C) {
	fr, err := NewRatioCounter(1, time.Second, RatioClock(s.tm))
	c.Assert(err, IsNil)
	fr.IncA(1)
	c.Assert(fr.IsReady(), Equals, true)
	c.Assert(fr.Ratio(), Equals, 1.0)
}

func (s *FailRateSuite) TestNoA(c *C) {
	fr, err := NewRatioCounter(1, time.Second, RatioClock(s.tm))
	c.Assert(err, IsNil)
	fr.IncB(1)
	c.Assert(fr.IsReady(), Equals, true)
	c.Assert(fr.Ratio(), Equals, 0.0)
}

// Make sure that data is properly calculated over several buckets
func (s *FailRateSuite) TestMultipleBuckets(c *C) {
	fr, err := NewRatioCounter(3, time.Second, RatioClock(s.tm))
	c.Assert(err, IsNil)

	fr.IncB(1)
	s.tm.CurrentTime = s.tm.CurrentTime.Add(time.Second)
	fr.IncA(1)

	s.tm.CurrentTime = s.tm.CurrentTime.Add(time.Second)
	fr.IncA(1)

	c.Assert(fr.IsReady(), Equals, true)
	c.Assert(fr.Ratio(), Equals, float64(2)/float64(3))
}

// Make sure that data is properly calculated over several buckets
// When we overwrite old data when the window is rolling
func (s *FailRateSuite) TestOverwriteBuckets(c *C) {
	fr, err := NewRatioCounter(3, time.Second, RatioClock(s.tm))
	c.Assert(err, IsNil)

	fr.IncB(1)

	s.tm.CurrentTime = s.tm.CurrentTime.Add(time.Second)
	fr.IncA(1)

	s.tm.CurrentTime = s.tm.CurrentTime.Add(time.Second)
	fr.IncA(1)

	// This time we should overwrite the old data points
	s.tm.CurrentTime = s.tm.CurrentTime.Add(time.Second)
	fr.IncA(1)
	fr.IncB(2)

	c.Assert(fr.IsReady(), Equals, true)
	c.Assert(fr.Ratio(), Equals, float64(3)/float64(5))
}

// Make sure we cleanup the data after periods of inactivity
// So it does not mess up the stats
func (s *FailRateSuite) TestInactiveBuckets(c *C) {

	fr, err := NewRatioCounter(3, time.Second, RatioClock(s.tm))
	c.Assert(err, IsNil)

	fr.IncB(1)

	s.tm.CurrentTime = s.tm.CurrentTime.Add(time.Second)
	fr.IncA(1)

	s.tm.CurrentTime = s.tm.CurrentTime.Add(time.Second)
	fr.IncA(1)

	// This time we should overwrite the old data points with new data
	s.tm.CurrentTime = s.tm.CurrentTime.Add(time.Second)
	fr.IncA(1)
	fr.IncB(2)

	// Jump to the last bucket and change the data
	s.tm.CurrentTime = s.tm.CurrentTime.Add(time.Second * 2)
	fr.IncB(1)

	c.Assert(fr.IsReady(), Equals, true)
	c.Assert(fr.Ratio(), Equals, float64(1)/float64(4))
}

func (s *FailRateSuite) TestLongPeriodsOfInactivity(c *C) {
	fr, err := NewRatioCounter(2, time.Second, RatioClock(s.tm))
	c.Assert(err, IsNil)

	fr.IncB(1)

	s.tm.CurrentTime = s.tm.CurrentTime.Add(time.Second)
	fr.IncA(1)

	c.Assert(fr.IsReady(), Equals, true)
	c.Assert(fr.Ratio(), Equals, 0.5)

	// This time we should overwrite all data points
	s.tm.CurrentTime = s.tm.CurrentTime.Add(100 * time.Second)
	fr.IncA(1)
	c.Assert(fr.Ratio(), Equals, 1.0)
}

func (s *FailRateSuite) TestReset(c *C) {
	fr, err := NewRatioCounter(1, time.Second, RatioClock(s.tm))
	c.Assert(err, IsNil)

	fr.IncB(1)
	fr.IncA(1)

	c.Assert(fr.IsReady(), Equals, true)
	c.Assert(fr.Ratio(), Equals, 0.5)

	// Reset the counter
	fr.Reset()
	c.Assert(fr.IsReady(), Equals, false)

	// Now add some stats
	fr.IncA(2)

	// We are game again!
	c.Assert(fr.IsReady(), Equals, true)
	c.Assert(fr.Ratio(), Equals, 1.0)
}
