package ratelimit

import (
	"testing"
	"time"

	"github.com/mailgun/timetools"
	. "gopkg.in/check.v1"
)

func TestTokenBucket(t *testing.T) { TestingT(t) }

type BucketSuite struct {
	clock *timetools.FreezedTime
}

var _ = Suite(&BucketSuite{})

func (s *BucketSuite) SetUpSuite(c *C) {
	s.clock = &timetools.FreezedTime{
		CurrentTime: time.Date(2012, 3, 4, 5, 6, 7, 0, time.UTC),
	}
}

func (s *BucketSuite) TestConsumeSingleToken(c *C) {
	tb := newTokenBucket(&rate{time.Second, 1, 1}, s.clock)

	// First request passes
	delay, err := tb.consume(1)
	c.Assert(err, IsNil)
	c.Assert(delay, Equals, time.Duration(0))

	// Next request does not pass the same second
	delay, err = tb.consume(1)
	c.Assert(err, IsNil)
	c.Assert(delay, Equals, time.Second)

	// Second later, the request passes
	s.clock.Sleep(time.Second)
	delay, err = tb.consume(1)
	c.Assert(err, IsNil)
	c.Assert(delay, Equals, time.Duration(0))

	// Five seconds later, still only one request is allowed
	// because maxBurst is 1
	s.clock.Sleep(5 * time.Second)
	delay, err = tb.consume(1)
	c.Assert(err, IsNil)
	c.Assert(delay, Equals, time.Duration(0))

	// The next one is forbidden
	delay, err = tb.consume(1)
	c.Assert(err, IsNil)
	c.Assert(delay, Equals, time.Second)
}

func (s *BucketSuite) TestFastConsumption(c *C) {
	tb := newTokenBucket(&rate{time.Second, 1, 1}, s.clock)

	// First request passes
	delay, err := tb.consume(1)
	c.Assert(err, IsNil)
	c.Assert(delay, Equals, time.Duration(0))

	// Try 200 ms later
	s.clock.Sleep(time.Millisecond * 200)
	delay, err = tb.consume(1)
	c.Assert(err, IsNil)
	c.Assert(delay, Equals, time.Second)

	// Try 700 ms later
	s.clock.Sleep(time.Millisecond * 700)
	delay, err = tb.consume(1)
	c.Assert(err, IsNil)
	c.Assert(delay, Equals, time.Second)

	// Try 100 ms later, success!
	s.clock.Sleep(time.Millisecond * 100)
	delay, err = tb.consume(1)
	c.Assert(err, IsNil)
	c.Assert(delay, Equals, time.Duration(0))
}

func (s *BucketSuite) TestConsumeMultipleTokens(c *C) {
	tb := newTokenBucket(&rate{time.Second, 3, 5}, s.clock)

	delay, err := tb.consume(3)
	c.Assert(err, IsNil)
	c.Assert(delay, Equals, time.Duration(0))

	delay, err = tb.consume(2)
	c.Assert(err, IsNil)
	c.Assert(delay, Equals, time.Duration(0))

	delay, err = tb.consume(1)
	c.Assert(err, IsNil)
	c.Assert(delay, Not(Equals), time.Duration(0))
}

func (s *BucketSuite) TestDelayIsCorrect(c *C) {
	tb := newTokenBucket(&rate{time.Second, 3, 5}, s.clock)

	// Exhaust initial capacity
	delay, err := tb.consume(5)
	c.Assert(err, IsNil)
	c.Assert(delay, Equals, time.Duration(0))

	delay, err = tb.consume(3)
	c.Assert(err, IsNil)
	c.Assert(delay, Not(Equals), time.Duration(0))

	// Now wait provided delay and make sure we can consume now
	s.clock.Sleep(delay)
	delay, err = tb.consume(3)
	c.Assert(err, IsNil)
	c.Assert(delay, Equals, time.Duration(0))
}

// Make sure requests that exceed burst size are not allowed
func (s *BucketSuite) TestExceedsBurst(c *C) {
	tb := newTokenBucket(&rate{time.Second, 1, 10}, s.clock)

	_, err := tb.consume(11)
	c.Assert(err, NotNil)
}

func (s *BucketSuite) TestConsumeBurst(c *C) {
	tb := newTokenBucket(&rate{time.Second, 2, 5}, s.clock)

	// In two seconds we would have 5 tokens
	s.clock.Sleep(2 * time.Second)

	// Lets consume 5 at once
	delay, err := tb.consume(5)
	c.Assert(delay, Equals, time.Duration(0))
	c.Assert(err, IsNil)
}

func (s *BucketSuite) TestConsumeEstimate(c *C) {
	tb := newTokenBucket(&rate{time.Second, 2, 4}, s.clock)

	// Consume all burst at once
	delay, err := tb.consume(4)
	c.Assert(err, IsNil)
	c.Assert(delay, Equals, time.Duration(0))

	// Now try to consume it and face delay
	delay, err = tb.consume(4)
	c.Assert(err, IsNil)
	c.Assert(delay, Equals, time.Duration(2)*time.Second)
}

// If a rate with different period is passed to the `update` method, then an
// error is returned but the state of the bucket remains valid and unchanged.
func (s *BucketSuite) TestUpdateInvalidPeriod(c *C) {
	// Given
	tb := newTokenBucket(&rate{time.Second, 10, 20}, s.clock)
	tb.consume(15) // 5 tokens available
	// When
	err := tb.update(&rate{time.Second + 1, 30, 40}) // still 5 tokens available
	// Then
	c.Assert(err, NotNil)

	// ...check that rate did not change
	s.clock.Sleep(500 * time.Millisecond)
	delay, err := tb.consume(11)
	c.Assert(err, IsNil)
	c.Assert(delay, Equals, 100*time.Millisecond)
	delay, err = tb.consume(10)
	c.Assert(err, IsNil)
	c.Assert(delay, Equals, time.Duration(0)) // 0 available

	// ...check that burst did not change
	s.clock.Sleep(40 * time.Second)
	delay, err = tb.consume(21)
	c.Assert(err, NotNil)
	delay, err = tb.consume(20)
	c.Assert(err, IsNil)
	c.Assert(delay, Equals, time.Duration(0)) // 0 available
}

// If the capacity of the bucket is increased by the update then it takes some
// time to fill the bucket with tokens up to the new capacity.
func (s *BucketSuite) TestUpdateBurstIncreased(c *C) {
	// Given
	tb := newTokenBucket(&rate{time.Second, 10, 20}, s.clock)
	tb.consume(15) // 5 tokens available
	// When
	err := tb.update(&rate{time.Second, 10, 50}) // still 5 tokens available
	// Then
	c.Assert(err, IsNil)
	delay, err := tb.consume(50)
	c.Assert(err, IsNil)
	c.Assert(delay, Equals, time.Duration(time.Second/10*45))
}

// If the capacity of the bucket is increased by the update then it takes some
// time to fill the bucket with tokens up to the new capacity.
func (s *BucketSuite) TestUpdateBurstDecreased(c *C) {
	// Given
	tb := newTokenBucket(&rate{time.Second, 10, 50}, s.clock)
	tb.consume(15) // 35 tokens available
	// When
	err := tb.update(&rate{time.Second, 10, 20}) // the number of available tokens reduced to 20.
	// Then
	c.Assert(err, IsNil)
	delay, err := tb.consume(21)
	c.Assert(err, NotNil)
	c.Assert(delay, Equals, time.Duration(-1))
}

// If rate is updated then it affects the bucket refill speed.
func (s *BucketSuite) TestUpdateRateChanged(c *C) {
	// Given
	tb := newTokenBucket(&rate{time.Second, 10, 20}, s.clock)
	tb.consume(15) // 5 tokens available
	// When
	err := tb.update(&rate{time.Second, 20, 20}) // still 5 tokens available
	// Then
	delay, err := tb.consume(20)
	c.Assert(err, IsNil)
	c.Assert(delay, Equals, time.Duration(time.Second/20*15))
}

// Only the most recent consumption is reverted by `Rollback`.
func (s *BucketSuite) TestRollback(c *C) {
	// Given
	tb := newTokenBucket(&rate{time.Second, 10, 20}, s.clock)
	tb.consume(8) // 12 tokens available
	tb.consume(7) // 5 tokens available
	// When
	tb.rollback() // 12 tokens available
	// Then
	delay, err := tb.consume(12)
	c.Assert(err, IsNil)
	c.Assert(delay, Equals, time.Duration(0))
	delay, err = tb.consume(1)
	c.Assert(err, IsNil)
	c.Assert(delay, Equals, 100*time.Millisecond)
}

// It is safe to call `Rollback` several times. The second and all subsequent
// calls just do nothing.
func (s *BucketSuite) TestRollbackSeveralTimes(c *C) {
	// Given
	tb := newTokenBucket(&rate{time.Second, 10, 20}, s.clock)
	tb.consume(8) // 12 tokens available
	tb.rollback() // 20 tokens available
	// When
	tb.rollback() // still 20 tokens available
	tb.rollback() // still 20 tokens available
	tb.rollback() // still 20 tokens available
	// Then: all 20 tokens can be consumed
	delay, err := tb.consume(20)
	c.Assert(err, IsNil)
	c.Assert(delay, Equals, time.Duration(0))
	delay, err = tb.consume(1)
	c.Assert(err, IsNil)
	c.Assert(delay, Equals, 100*time.Millisecond)
}

// If previous consumption returned a delay due to an attempt to consume more
// tokens then there are available, then `Rollback` has no effect.
func (s *BucketSuite) TestRollbackAfterAvailableExceeded(c *C) {
	// Given
	tb := newTokenBucket(&rate{time.Second, 10, 20}, s.clock)
	tb.consume(8)                // 12 tokens available
	delay, err := tb.consume(15) // still 12 tokens available
	c.Assert(err, IsNil)
	c.Assert(delay, Equals, 300*time.Millisecond)
	// When
	tb.rollback() // Previous operation consumed 0 tokens, so rollback has no effect.
	// Then
	delay, err = tb.consume(12)
	c.Assert(err, IsNil)
	c.Assert(delay, Equals, time.Duration(0))
	delay, err = tb.consume(1)
	c.Assert(err, IsNil)
	c.Assert(delay, Equals, 100*time.Millisecond)
}

// If previous consumption returned a error due to an attempt to consume more
// tokens then the bucket's burst size, then `Rollback` has no effect.
func (s *BucketSuite) TestRollbackAfterError(c *C) {
	// Given
	tb := newTokenBucket(&rate{time.Second, 10, 20}, s.clock)
	tb.consume(8)                // 12 tokens available
	delay, err := tb.consume(21) // still 12 tokens available
	c.Assert(err, NotNil)
	c.Assert(delay, Equals, time.Duration(-1))
	// When
	tb.rollback() // Previous operation consumed 0 tokens, so rollback has no effect.
	// Then
	delay, err = tb.consume(12)
	c.Assert(err, IsNil)
	c.Assert(delay, Equals, time.Duration(0))
	delay, err = tb.consume(1)
	c.Assert(err, IsNil)
	c.Assert(delay, Equals, 100*time.Millisecond)
}
