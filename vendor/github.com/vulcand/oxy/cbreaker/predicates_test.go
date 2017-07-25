package cbreaker

import (
	"github.com/vulcand/oxy/memmetrics"
	"time"

	. "gopkg.in/check.v1"
)

type PredicatesSuite struct {
}

var _ = Suite(&PredicatesSuite{})

func (s *PredicatesSuite) TestTripped(c *C) {
	predicates := []struct {
		Expression string
		M          *memmetrics.RTMetrics
		V          bool
	}{
		{
			Expression: "NetworkErrorRatio() > 0.5",
			M:          statsNetErrors(0.6),
			V:          true,
		},
		{
			Expression: "NetworkErrorRatio() < 0.5",
			M:          statsNetErrors(0.6),
			V:          false,
		},
		{
			Expression: "LatencyAtQuantileMS(50.0) > 50",
			M:          statsLatencyAtQuantile(50, time.Millisecond*51),
			V:          true,
		},
		{
			Expression: "LatencyAtQuantileMS(50.0) < 50",
			M:          statsLatencyAtQuantile(50, time.Millisecond*51),
			V:          false,
		},
		{
			Expression: "ResponseCodeRatio(500, 600, 0, 600) > 0.5",
			M:          statsResponseCodes(statusCode{Code: 200, Count: 5}, statusCode{Code: 500, Count: 6}),
			V:          true,
		},
		{
			Expression: "ResponseCodeRatio(500, 600, 0, 600) > 0.5",
			M:          statsResponseCodes(statusCode{Code: 200, Count: 5}, statusCode{Code: 500, Count: 4}),
			V:          false,
		},
	}
	for _, t := range predicates {
		p, err := parseExpression(t.Expression)
		c.Assert(err, IsNil)
		c.Assert(p, NotNil)

		c.Assert(p(&CircuitBreaker{metrics: t.M}), Equals, t.V)
	}
}

func (s *PredicatesSuite) TestErrors(c *C) {
	predicates := []struct {
		Expression string
		M          *memmetrics.RTMetrics
	}{
		{
			Expression: "LatencyAtQuantileMS(40.0) > 50", // quantile not defined
			M:          statsNetErrors(0.6),
		},
	}
	for _, t := range predicates {
		p, err := parseExpression(t.Expression)
		c.Assert(err, IsNil)
		c.Assert(p, NotNil)

		c.Assert(p(&CircuitBreaker{metrics: t.M}), Equals, false)
	}
}
