package memmetrics

import (
	"time"

	"github.com/mailgun/timetools"
	. "gopkg.in/check.v1"
)

type RRSuite struct {
	tm *timetools.FreezedTime
}

var _ = Suite(&RRSuite{})

func (s *RRSuite) SetUpSuite(c *C) {
	s.tm = &timetools.FreezedTime{
		CurrentTime: time.Date(2012, 3, 4, 5, 6, 7, 0, time.UTC),
	}
}

func (s *RRSuite) TestDefaults(c *C) {
	rr, err := NewRTMetrics(RTClock(s.tm))
	c.Assert(err, IsNil)
	c.Assert(rr, NotNil)

	rr.Record(200, time.Second)
	rr.Record(502, 2*time.Second)
	rr.Record(200, time.Second)
	rr.Record(200, time.Second)

	c.Assert(rr.NetworkErrorCount(), Equals, int64(1))
	c.Assert(rr.TotalCount(), Equals, int64(4))
	c.Assert(rr.StatusCodesCounts(), DeepEquals, map[int]int64{502: 1, 200: 3})
	c.Assert(rr.NetworkErrorRatio(), Equals, float64(1)/float64(4))
	c.Assert(rr.ResponseCodeRatio(500, 503, 200, 300), Equals, 1.0/3.0)

	h, err := rr.LatencyHistogram()
	c.Assert(err, IsNil)
	c.Assert(int(h.LatencyAtQuantile(100)/time.Second), Equals, 2)

	rr.Reset()
	c.Assert(rr.NetworkErrorCount(), Equals, int64(0))
	c.Assert(rr.TotalCount(), Equals, int64(0))
	c.Assert(rr.StatusCodesCounts(), DeepEquals, map[int]int64{})
	c.Assert(rr.NetworkErrorRatio(), Equals, float64(0))
	c.Assert(rr.ResponseCodeRatio(500, 503, 200, 300), Equals, float64(0))

	h, err = rr.LatencyHistogram()
	c.Assert(err, IsNil)
	c.Assert(h.LatencyAtQuantile(100), Equals, time.Duration(0))

}

func (s *RRSuite) TestAppend(c *C) {
	rr, err := NewRTMetrics(RTClock(s.tm))
	c.Assert(err, IsNil)
	c.Assert(rr, NotNil)

	rr.Record(200, time.Second)
	rr.Record(502, 2*time.Second)
	rr.Record(200, time.Second)
	rr.Record(200, time.Second)

	rr2, err := NewRTMetrics(RTClock(s.tm))
	c.Assert(err, IsNil)
	c.Assert(rr2, NotNil)

	rr2.Record(200, 3*time.Second)
	rr2.Record(501, 3*time.Second)
	rr2.Record(200, 3*time.Second)
	rr2.Record(200, 3*time.Second)

	c.Assert(rr2.Append(rr), IsNil)
	c.Assert(rr2.StatusCodesCounts(), DeepEquals, map[int]int64{501: 1, 502: 1, 200: 6})
	c.Assert(rr2.NetworkErrorCount(), Equals, int64(1))

	h, err := rr2.LatencyHistogram()
	c.Assert(err, IsNil)
	c.Assert(int(h.LatencyAtQuantile(100)/time.Second), Equals, 3)
}
