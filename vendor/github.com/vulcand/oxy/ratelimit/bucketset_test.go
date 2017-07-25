package ratelimit

import (
	"time"

	"github.com/mailgun/timetools"
	. "gopkg.in/check.v1"
)

type BucketSetSuite struct {
	clock *timetools.FreezedTime
}

var _ = Suite(&BucketSetSuite{})

func (s *BucketSetSuite) SetUpSuite(c *C) {
	s.clock = &timetools.FreezedTime{
		CurrentTime: time.Date(2012, 3, 4, 5, 6, 7, 0, time.UTC),
	}
}

// A value returned by `MaxPeriod` corresponds to the longest bucket time period.
func (s *BucketSetSuite) TestLongestPeriod(c *C) {
	// Given
	rates := NewRateSet()
	rates.Add(1*time.Second, 10, 20)
	rates.Add(7*time.Second, 10, 20)
	rates.Add(5*time.Second, 11, 21)
	// When
	tbs := NewTokenBucketSet(rates, s.clock)
	// Then
	c.Assert(tbs.maxPeriod, Equals, 7*time.Second)
}

// Successful token consumption updates state of all buckets in the set.
func (s *BucketSetSuite) TestConsume(c *C) {
	// Given
	rates := NewRateSet()
	rates.Add(1*time.Second, 10, 20)
	rates.Add(10*time.Second, 20, 50)
	tbs := NewTokenBucketSet(rates, s.clock)
	// When
	delay, err := tbs.Consume(15)
	// Then
	c.Assert(delay, Equals, time.Duration(0))
	c.Assert(err, IsNil)
	c.Assert(tbs.debugState(), Equals, "{1s: 5}, {10s: 35}")
}

// As time goes by all set buckets are refilled with appropriate rates.
func (s *BucketSetSuite) TestConsumeRefill(c *C) {
	// Given
	rates := NewRateSet()
	rates.Add(10*time.Second, 10, 20)
	rates.Add(100*time.Second, 20, 50)
	tbs := NewTokenBucketSet(rates, s.clock)
	tbs.Consume(15)
	c.Assert(tbs.debugState(), Equals, "{10s: 5}, {1m40s: 35}")
	// When
	s.clock.Sleep(10 * time.Second)
	delay, err := tbs.Consume(0) // Consumes nothing but forces an internal state update.
	// Then
	c.Assert(delay, Equals, time.Duration(0))
	c.Assert(err, IsNil)
	c.Assert(tbs.debugState(), Equals, "{10s: 15}, {1m40s: 37}")
}

// If the first bucket in the set has no enough tokens to allow desired
// consumption then an appropriate delay is returned.
func (s *BucketSetSuite) TestConsumeLimitedBy1st(c *C) {
	// Given
	rates := NewRateSet()
	rates.Add(10*time.Second, 10, 10)
	rates.Add(100*time.Second, 20, 20)
	tbs := NewTokenBucketSet(rates, s.clock)
	tbs.Consume(5)
	c.Assert(tbs.debugState(), Equals, "{10s: 5}, {1m40s: 15}")
	// When
	delay, err := tbs.Consume(10)
	// Then
	c.Assert(delay, Equals, 5*time.Second)
	c.Assert(err, IsNil)
	c.Assert(tbs.debugState(), Equals, "{10s: 5}, {1m40s: 15}")
}

// If the second bucket in the set has no enough tokens to allow desired
// consumption then an appropriate delay is returned.
func (s *BucketSetSuite) TestConsumeLimitedBy2st(c *C) {
	// Given
	rates := NewRateSet()
	rates.Add(10*time.Second, 10, 10)
	rates.Add(100*time.Second, 20, 20)
	tbs := NewTokenBucketSet(rates, s.clock)
	tbs.Consume(10)
	s.clock.Sleep(10 * time.Second)
	tbs.Consume(10)
	s.clock.Sleep(5 * time.Second)
	tbs.Consume(0)
	c.Assert(tbs.debugState(), Equals, "{10s: 5}, {1m40s: 3}")
	// When
	delay, err := tbs.Consume(10)
	// Then
	c.Assert(delay, Equals, 7*(5*time.Second))
	c.Assert(err, IsNil)
	c.Assert(tbs.debugState(), Equals, "{10s: 5}, {1m40s: 3}")
}

// An attempt to consume more tokens then the smallest bucket capacity results
// in error.
func (s *BucketSetSuite) TestConsumeMoreThenBurst(c *C) {
	// Given
	rates := NewRateSet()
	rates.Add(1*time.Second, 10, 20)
	rates.Add(10*time.Second, 50, 100)
	tbs := NewTokenBucketSet(rates, s.clock)
	tbs.Consume(5)
	c.Assert(tbs.debugState(), Equals, "{1s: 15}, {10s: 95}")
	// When
	_, err := tbs.Consume(21)
	//Then
	c.Assert(tbs.debugState(), Equals, "{1s: 15}, {10s: 95}")
	c.Assert(err, NotNil)
}

// Update operation can add buckets.
func (s *BucketSetSuite) TestUpdateMore(c *C) {
	// Given
	rates := NewRateSet()
	rates.Add(1*time.Second, 10, 20)
	rates.Add(10*time.Second, 20, 50)
	rates.Add(20*time.Second, 45, 90)
	tbs := NewTokenBucketSet(rates, s.clock)
	tbs.Consume(5)
	c.Assert(tbs.debugState(), Equals, "{1s: 15}, {10s: 45}, {20s: 85}")
	rates = NewRateSet()
	rates.Add(10*time.Second, 30, 40)
	rates.Add(11*time.Second, 30, 40)
	rates.Add(12*time.Second, 30, 40)
	rates.Add(13*time.Second, 30, 40)
	// When
	tbs.Update(rates)
	// Then
	c.Assert(tbs.debugState(), Equals, "{10s: 40}, {11s: 40}, {12s: 40}, {13s: 40}")
	c.Assert(tbs.maxPeriod, Equals, 13*time.Second)
}

// Update operation can remove buckets.
func (s *BucketSetSuite) TestUpdateLess(c *C) {
	// Given
	rates := NewRateSet()
	rates.Add(1*time.Second, 10, 20)
	rates.Add(10*time.Second, 20, 50)
	rates.Add(20*time.Second, 45, 90)
	rates.Add(30*time.Second, 50, 100)
	tbs := NewTokenBucketSet(rates, s.clock)
	tbs.Consume(5)
	c.Assert(tbs.debugState(), Equals, "{1s: 15}, {10s: 45}, {20s: 85}, {30s: 95}")
	rates = NewRateSet()
	rates.Add(10*time.Second, 25, 20)
	rates.Add(20*time.Second, 30, 21)
	// When
	tbs.Update(rates)
	// Then
	c.Assert(tbs.debugState(), Equals, "{10s: 20}, {20s: 21}")
	c.Assert(tbs.maxPeriod, Equals, 20*time.Second)
}

// Update operation can remove buckets.
func (s *BucketSetSuite) TestUpdateAllDifferent(c *C) {
	// Given
	rates := NewRateSet()
	rates.Add(10*time.Second, 20, 50)
	rates.Add(30*time.Second, 50, 100)
	tbs := NewTokenBucketSet(rates, s.clock)
	tbs.Consume(5)
	c.Assert(tbs.debugState(), Equals, "{10s: 45}, {30s: 95}")
	rates = NewRateSet()
	rates.Add(1*time.Second, 10, 40)
	rates.Add(60*time.Second, 100, 150)
	// When
	tbs.Update(rates)
	// Then
	c.Assert(tbs.debugState(), Equals, "{1s: 40}, {1m0s: 150}")
	c.Assert(tbs.maxPeriod, Equals, 60*time.Second)
}
