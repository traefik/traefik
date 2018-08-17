package cbreaker

import (
	"fmt"
	"time"

	"github.com/vulcand/predicate"
)

type hpredicate func(*CircuitBreaker) bool

// parseExpression parses expression in the go language into predicates.
func parseExpression(in string) (hpredicate, error) {
	p, err := predicate.NewParser(predicate.Def{
		Operators: predicate.Operators{
			AND: and,
			OR:  or,
			EQ:  eq,
			NEQ: neq,
			LT:  lt,
			LE:  le,
			GT:  gt,
			GE:  ge,
		},
		Functions: map[string]interface{}{
			"LatencyAtQuantileMS": latencyAtQuantile,
			"NetworkErrorRatio":   networkErrorRatio,
			"ResponseCodeRatio":   responseCodeRatio,
		},
	})
	if err != nil {
		return nil, err
	}
	out, err := p.Parse(in)
	if err != nil {
		return nil, err
	}
	pr, ok := out.(hpredicate)
	if !ok {
		return nil, fmt.Errorf("expected predicate, got %T", out)
	}
	return pr, nil
}

type toInt func(c *CircuitBreaker) int
type toFloat64 func(c *CircuitBreaker) float64

func latencyAtQuantile(quantile float64) toInt {
	return func(c *CircuitBreaker) int {
		h, err := c.metrics.LatencyHistogram()
		if err != nil {
			c.log.Errorf("Failed to get latency histogram, for %v error: %v", c, err)
			return 0
		}
		return int(h.LatencyAtQuantile(quantile) / time.Millisecond)
	}
}

func networkErrorRatio() toFloat64 {
	return func(c *CircuitBreaker) float64 {
		return c.metrics.NetworkErrorRatio()
	}
}

func responseCodeRatio(startA, endA, startB, endB int) toFloat64 {
	return func(c *CircuitBreaker) float64 {
		return c.metrics.ResponseCodeRatio(startA, endA, startB, endB)
	}
}

// or returns predicate by joining the passed predicates with logical 'or'
func or(fns ...hpredicate) hpredicate {
	return func(c *CircuitBreaker) bool {
		for _, fn := range fns {
			if fn(c) {
				return true
			}
		}
		return false
	}
}

// and returns predicate by joining the passed predicates with logical 'and'
func and(fns ...hpredicate) hpredicate {
	return func(c *CircuitBreaker) bool {
		for _, fn := range fns {
			if !fn(c) {
				return false
			}
		}
		return true
	}
}

// not creates negation of the passed predicate
func not(p hpredicate) hpredicate {
	return func(c *CircuitBreaker) bool {
		return !p(c)
	}
}

// eq returns predicate that tests for equality of the value of the mapper and the constant
func eq(m interface{}, value interface{}) (hpredicate, error) {
	switch mapper := m.(type) {
	case toInt:
		return intEQ(mapper, value)
	case toFloat64:
		return float64EQ(mapper, value)
	}
	return nil, fmt.Errorf("eq: unsupported argument: %T", m)
}

// neq returns predicate that tests for inequality of the value of the mapper and the constant
func neq(m interface{}, value interface{}) (hpredicate, error) {
	p, err := eq(m, value)
	if err != nil {
		return nil, err
	}
	return not(p), nil
}

// lt returns predicate that tests that value of the mapper function is less than the constant
func lt(m interface{}, value interface{}) (hpredicate, error) {
	switch mapper := m.(type) {
	case toInt:
		return intLT(mapper, value)
	case toFloat64:
		return float64LT(mapper, value)
	}
	return nil, fmt.Errorf("lt: unsupported argument: %T", m)
}

// le returns predicate that tests that value of the mapper function is less or equal than the constant
func le(m interface{}, value interface{}) (hpredicate, error) {
	l, err := lt(m, value)
	if err != nil {
		return nil, err
	}
	e, err := eq(m, value)
	if err != nil {
		return nil, err
	}
	return func(c *CircuitBreaker) bool {
		return l(c) || e(c)
	}, nil
}

// gt returns predicate that tests that value of the mapper function is greater than the constant
func gt(m interface{}, value interface{}) (hpredicate, error) {
	switch mapper := m.(type) {
	case toInt:
		return intGT(mapper, value)
	case toFloat64:
		return float64GT(mapper, value)
	}
	return nil, fmt.Errorf("gt: unsupported argument: %T", m)
}

// ge returns predicate that tests that value of the mapper function is less or equal than the constant
func ge(m interface{}, value interface{}) (hpredicate, error) {
	g, err := gt(m, value)
	if err != nil {
		return nil, err
	}
	e, err := eq(m, value)
	if err != nil {
		return nil, err
	}
	return func(c *CircuitBreaker) bool {
		return g(c) || e(c)
	}, nil
}

func intEQ(m toInt, val interface{}) (hpredicate, error) {
	value, ok := val.(int)
	if !ok {
		return nil, fmt.Errorf("expected int, got %T", val)
	}
	return func(c *CircuitBreaker) bool {
		return m(c) == value
	}, nil
}

func float64EQ(m toFloat64, val interface{}) (hpredicate, error) {
	value, ok := val.(float64)
	if !ok {
		return nil, fmt.Errorf("expected float64, got %T", val)
	}
	return func(c *CircuitBreaker) bool {
		return m(c) == value
	}, nil
}

func intLT(m toInt, val interface{}) (hpredicate, error) {
	value, ok := val.(int)
	if !ok {
		return nil, fmt.Errorf("expected int, got %T", val)
	}
	return func(c *CircuitBreaker) bool {
		return m(c) < value
	}, nil
}

func intGT(m toInt, val interface{}) (hpredicate, error) {
	value, ok := val.(int)
	if !ok {
		return nil, fmt.Errorf("expected int, got %T", val)
	}
	return func(c *CircuitBreaker) bool {
		return m(c) > value
	}, nil
}

func float64LT(m toFloat64, val interface{}) (hpredicate, error) {
	value, ok := val.(float64)
	if !ok {
		return nil, fmt.Errorf("expected int, got %T", val)
	}
	return func(c *CircuitBreaker) bool {
		return m(c) < value
	}, nil
}

func float64GT(m toFloat64, val interface{}) (hpredicate, error) {
	value, ok := val.(float64)
	if !ok {
		return nil, fmt.Errorf("expected int, got %T", val)
	}
	return func(c *CircuitBreaker) bool {
		return m(c) > value
	}, nil
}
