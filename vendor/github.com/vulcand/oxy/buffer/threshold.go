package buffer

import (
	"fmt"
	"net/http"

	"github.com/vulcand/predicate"
)

// IsValidExpression check if it's a valid expression
func IsValidExpression(expr string) bool {
	_, err := parseExpression(expr)
	return err == nil
}

type context struct {
	r            *http.Request
	attempt      int
	responseCode int
}

type hpredicate func(*context) bool

// Parses expression in the go language into Failover predicates
func parseExpression(in string) (hpredicate, error) {
	p, err := predicate.NewParser(predicate.Def{
		Operators: predicate.Operators{
			AND: and,
			OR:  or,
			EQ:  eq,
			NEQ: neq,
			LT:  lt,
			GT:  gt,
			LE:  le,
			GE:  ge,
		},
		Functions: map[string]interface{}{
			"RequestMethod":  requestMethod,
			"IsNetworkError": isNetworkError,
			"Attempts":       attempts,
			"ResponseCode":   responseCode,
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

type toString func(c *context) string
type toInt func(c *context) int

// RequestMethod returns mapper of the request to its method e.g. POST
func requestMethod() toString {
	return func(c *context) string {
		return c.r.Method
	}
}

// Attempts returns mapper of the request to the number of proxy attempts
func attempts() toInt {
	return func(c *context) int {
		return c.attempt
	}
}

// ResponseCode returns mapper of the request to the last response code, returns 0 if there was no response code.
func responseCode() toInt {
	return func(c *context) int {
		return c.responseCode
	}
}

// IsNetworkError returns a predicate that returns true if last attempt ended with network error.
func isNetworkError() hpredicate {
	return func(c *context) bool {
		return c.responseCode == http.StatusBadGateway || c.responseCode == http.StatusGatewayTimeout
	}
}

// and returns predicate by joining the passed predicates with logical 'and'
func and(fns ...hpredicate) hpredicate {
	return func(c *context) bool {
		for _, fn := range fns {
			if !fn(c) {
				return false
			}
		}
		return true
	}
}

// or returns predicate by joining the passed predicates with logical 'or'
func or(fns ...hpredicate) hpredicate {
	return func(c *context) bool {
		for _, fn := range fns {
			if fn(c) {
				return true
			}
		}
		return false
	}
}

// not creates negation of the passed predicate
func not(p hpredicate) hpredicate {
	return func(c *context) bool {
		return !p(c)
	}
}

// eq returns predicate that tests for equality of the value of the mapper and the constant
func eq(m interface{}, value interface{}) (hpredicate, error) {
	switch mapper := m.(type) {
	case toString:
		return stringEQ(mapper, value)
	case toInt:
		return intEQ(mapper, value)
	}
	return nil, fmt.Errorf("unsupported argument: %T", m)
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
	}
	return nil, fmt.Errorf("unsupported argument: %T", m)
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
	return func(c *context) bool {
		return l(c) || e(c)
	}, nil
}

// gt returns predicate that tests that value of the mapper function is greater than the constant
func gt(m interface{}, value interface{}) (hpredicate, error) {
	switch mapper := m.(type) {
	case toInt:
		return intGT(mapper, value)
	}
	return nil, fmt.Errorf("unsupported argument: %T", m)
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
	return func(c *context) bool {
		return g(c) || e(c)
	}, nil
}

func stringEQ(m toString, val interface{}) (hpredicate, error) {
	value, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("expected string, got %T", val)
	}
	return func(c *context) bool {
		return m(c) == value
	}, nil
}

func intEQ(m toInt, val interface{}) (hpredicate, error) {
	value, ok := val.(int)
	if !ok {
		return nil, fmt.Errorf("expected int, got %T", val)
	}
	return func(c *context) bool {
		return m(c) == value
	}, nil
}

func intLT(m toInt, val interface{}) (hpredicate, error) {
	value, ok := val.(int)
	if !ok {
		return nil, fmt.Errorf("expected int, got %T", val)
	}
	return func(c *context) bool {
		return m(c) < value
	}, nil
}

func intGT(m toInt, val interface{}) (hpredicate, error) {
	value, ok := val.(int)
	if !ok {
		return nil, fmt.Errorf("expected int, got %T", val)
	}
	return func(c *context) bool {
		return m(c) > value
	}, nil
}
