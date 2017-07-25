package memmetrics

import (
	"time"

	"github.com/mailgun/timetools"
	. "gopkg.in/check.v1"
)

type HistogramSuite struct {
	tm *timetools.FreezedTime
}

var _ = Suite(&HistogramSuite{})

func (s *HistogramSuite) SetUpSuite(c *C) {
	s.tm = &timetools.FreezedTime{
		CurrentTime: time.Date(2012, 3, 4, 5, 6, 7, 0, time.UTC),
	}
}

func (s *HistogramSuite) TestMerge(c *C) {
	a, err := NewHDRHistogram(1, 3600000, 2)
	c.Assert(err, IsNil)

	a.RecordValues(1, 2)

	b, err := NewHDRHistogram(1, 3600000, 2)
	c.Assert(err, IsNil)

	b.RecordValues(2, 1)

	c.Assert(a.Merge(b), IsNil)

	c.Assert(a.ValueAtQuantile(50), Equals, int64(1))
	c.Assert(a.ValueAtQuantile(100), Equals, int64(2))
}

func (s *HistogramSuite) TestInvalidParams(c *C) {
	_, err := NewHDRHistogram(1, 3600000, 0)
	c.Assert(err, NotNil)
}

func (s *HistogramSuite) TestMergeNil(c *C) {
	a, err := NewHDRHistogram(1, 3600000, 1)
	c.Assert(err, IsNil)

	c.Assert(a.Merge(nil), NotNil)
}

func (s *HistogramSuite) TestRotation(c *C) {
	h, err := NewRollingHDRHistogram(
		1,           // min value
		3600000,     // max value
		3,           // significant figurwes
		time.Second, // 1 second is a rolling period
		2,           // 2 histograms in a window
		RollingClock(s.tm))

	c.Assert(err, IsNil)
	c.Assert(h, NotNil)

	h.RecordValues(5, 1)

	m, err := h.Merged()
	c.Assert(err, IsNil)
	c.Assert(m.ValueAtQuantile(100), Equals, int64(5))

	s.tm.CurrentTime = s.tm.CurrentTime.Add(time.Second)
	h.RecordValues(2, 1)
	h.RecordValues(1, 1)

	m, err = h.Merged()
	c.Assert(err, IsNil)
	c.Assert(m.ValueAtQuantile(100), Equals, int64(5))

	// rotate, this means that the old value would evaporate
	s.tm.CurrentTime = s.tm.CurrentTime.Add(time.Second)
	h.RecordValues(1, 1)
	m, err = h.Merged()
	c.Assert(err, IsNil)
	c.Assert(m.ValueAtQuantile(100), Equals, int64(2))
}

func (s *HistogramSuite) TestReset(c *C) {
	h, err := NewRollingHDRHistogram(
		1,           // min value
		3600000,     // max value
		3,           // significant figurwes
		time.Second, // 1 second is a rolling period
		2,           // 2 histograms in a window
		RollingClock(s.tm))

	c.Assert(err, IsNil)
	c.Assert(h, NotNil)

	h.RecordValues(5, 1)

	m, err := h.Merged()
	c.Assert(err, IsNil)
	c.Assert(m.ValueAtQuantile(100), Equals, int64(5))

	s.tm.CurrentTime = s.tm.CurrentTime.Add(time.Second)
	h.RecordValues(2, 1)
	h.RecordValues(1, 1)

	m, err = h.Merged()
	c.Assert(err, IsNil)
	c.Assert(m.ValueAtQuantile(100), Equals, int64(5))

	h.Reset()

	h.RecordValues(5, 1)

	m, err = h.Merged()
	c.Assert(err, IsNil)
	c.Assert(m.ValueAtQuantile(100), Equals, int64(5))

	s.tm.CurrentTime = s.tm.CurrentTime.Add(time.Second)
	h.RecordValues(2, 1)
	h.RecordValues(1, 1)

	m, err = h.Merged()
	c.Assert(err, IsNil)
	c.Assert(m.ValueAtQuantile(100), Equals, int64(5))

}
