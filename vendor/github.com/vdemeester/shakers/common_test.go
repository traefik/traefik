package shakers

import (
	"reflect"
	"testing"
	"time"

	"github.com/go-check/check"
)

func init() {
	check.Suite(&CommonCheckerS{})
}

type CommonCheckerS struct{}

func (s *CommonCheckerS) TestEqualsInfo(c *check.C) {
	testInfo(c, Equals, "Equals", []string{"obtained", "expected"})
}

func (s *CommonCheckerS) TestEqualsValidsEquals(c *check.C) {
	myTime, err := time.Parse("2006-01-02", "2018-01-01")
	if err != nil {
		c.Fatal(err)
	}

	testCheck(c, Equals, true, "", "string", "string")
	testCheck(c, Equals, true, "", 0, 0)
	testCheck(c, Equals, true, "", anEqualer{1}, anEqualer{1})
	testCheck(c, Equals, true, "", myTime, myTime)
	testCheck(c, Equals, true, "", myTime, "2018-01-01")
	testCheck(c, Equals, true, "", myTime, "2018-01-01T00:00:00Z")
}

func (s *CommonCheckerS) TestEqualsValidsDifferent(c *check.C) {
	myTime1, err := time.Parse("2006-01-02", "2018-01-01")
	if err != nil {
		c.Fatal(err)
	}
	myTime2, err := time.Parse("2006-01-02", "2018-01-02")
	if err != nil {
		c.Fatal(err)
	}

	testCheck(c, Equals, false, "", "string", "astring")
	testCheck(c, Equals, false, "", 0, 1)
	testCheck(c, Equals, false, "", anEqualer{1}, anEqualer{0})
	testCheck(c, Equals, false, "", myTime1, myTime2)
	testCheck(c, Equals, false, "", myTime1, "2018-01-02")
	testCheck(c, Equals, false, "", myTime1, "2018-01-02T00:00:00Z")
}

func (s *CommonCheckerS) TestEqualsInvalids(c *check.C) {
	// Incompatible type time.Time
	testCheck(c, Equals, false, "obtained value and expected value have not the same type.", "2015-01-01", time.Now())
	testCheck(c, Equals, false, "expected must be a Time struct, or parseable.", time.Now(), 0)

	// Incompatible type Equaler
	testCheck(c, Equals, false, "expected value must be an Equaler - implementing Equal(Equaler).", anEqualer{0}, 0)

	testCheck(c, Equals, false, "obtained value and expected value have not the same type.", 0, anEqualer{0})

	// Nils
	testCheck(c, Equals, false, "obtained value and expected value have not the same type.", nil, 0)
	testCheck(c, Equals, false, "obtained value and expected value have not the same type.", 0, nil)
}

type anEqualer struct {
	value int
}

func (a anEqualer) Equal(b Equaler) bool {
	if bEqualer, ok := b.(anEqualer); ok {
		return a.value == bEqualer.value
	}
	return false
}

func (s *CommonCheckerS) TestGreaterThanInfo(c *check.C) {
	testInfo(c, GreaterThan, "GreaterThan", []string{"obtained", "expected"})
}

func (s *CommonCheckerS) TestGreaterThanValidsGreater(c *check.C) {
	myTime, err := time.Parse("2006-01-02", "2018-01-01")
	if err != nil {
		c.Fatal(err)
	}

	testCheck(c, GreaterThan, true, "", 2, 1)
	testCheck(c, GreaterThan, true, "", 0, -1)
	testCheck(c, GreaterThan, true, "", float32(2), float32(0))
	testCheck(c, GreaterThan, true, "", float64(2), float64(0))
	testCheck(c, GreaterThan, true, "", byte(2), byte(1))
	testCheck(c, GreaterThan, true, "", int(2), int(1))
	testCheck(c, GreaterThan, true, "", int8(2), int8(1))
	testCheck(c, GreaterThan, true, "", int16(2), int16(1))
	testCheck(c, GreaterThan, true, "", int32(2), int32(1))
	testCheck(c, GreaterThan, true, "", int64(2), int64(1))
	testCheck(c, GreaterThan, true, "", uint(2), uint(1))
	testCheck(c, GreaterThan, true, "", uint8(2), uint8(1))
	testCheck(c, GreaterThan, true, "", uint16(2), uint16(1))
	testCheck(c, GreaterThan, true, "", uint32(2), uint32(1))
	testCheck(c, GreaterThan, true, "", uint64(2), uint64(1))
	testCheck(c, GreaterThan, true, "", myTime, "2017-12-31")
}

func (s *CommonCheckerS) TestGreaterThanValidsDifferent(c *check.C) {
	myTime, err := time.Parse("2006-01-02", "2018-01-01")
	if err != nil {
		c.Fatal(err)
	}

	testCheck(c, GreaterThan, false, "", 1, 2)
	testCheck(c, GreaterThan, false, "", -1, 0)
	testCheck(c, GreaterThan, false, "", float32(0), float32(2))
	testCheck(c, GreaterThan, false, "", float64(0), float64(2))
	testCheck(c, GreaterThan, false, "", byte(1), byte(2))
	testCheck(c, GreaterThan, false, "", int(1), int(2))
	testCheck(c, GreaterThan, false, "", int8(1), int8(2))
	testCheck(c, GreaterThan, false, "", int16(1), int16(2))
	testCheck(c, GreaterThan, false, "", int32(1), int32(2))
	testCheck(c, GreaterThan, false, "", int64(1), int64(2))
	testCheck(c, GreaterThan, false, "", uint(1), uint(2))
	testCheck(c, GreaterThan, false, "", uint8(1), uint8(2))
	testCheck(c, GreaterThan, false, "", uint16(1), uint16(2))
	testCheck(c, GreaterThan, false, "", uint32(1), uint32(2))
	testCheck(c, GreaterThan, false, "", uint64(1), uint64(2))
	testCheck(c, GreaterThan, false, "", myTime, "2018-01-01")
	testCheck(c, GreaterThan, false, "", myTime, "2018-01-02")
}

func (s *CommonCheckerS) TestGreaterThanInvalids(c *check.C) {
	myTime, err := time.Parse("2006-01-02", "2018-01-01")
	if err != nil {
		c.Fatal(err)
	}

	testCheck(c, GreaterThan, false, "obtained value and expected value have not the same type.", 1, "string")
	testCheck(c, GreaterThan, false, "obtained value and expected value have not the same type.", "string", 1)
	testCheck(c, GreaterThan, false, "obtained value and expected value have not the same type.", 1, complex128(1+2i))
	testCheck(c, GreaterThan, false, "expected must be a Time struct, or parseable.", myTime, 1)
	testCheck(c, GreaterThan, false, "expected must be a Time struct, or parseable.", myTime, "invalid")
}

func (s *CommonCheckerS) TestGreaterOrEqualThanInfo(c *check.C) {
	testInfo(c, GreaterOrEqualThan, "GreaterOrEqualThan", []string{"obtained", "expected"})
}

func (s *CommonCheckerS) TestGreaterOrEqualThanValidsGreaterOrEquals(c *check.C) {
	myTime, err := time.Parse("2006-01-02", "2018-01-01")
	if err != nil {
		c.Fatal(err)
	}

	testCheck(c, GreaterOrEqualThan, true, "", 2, 1)
	testCheck(c, GreaterOrEqualThan, true, "", 0, -1)
	testCheck(c, GreaterOrEqualThan, true, "", float32(2), float32(0))
	testCheck(c, GreaterOrEqualThan, true, "", float64(2), float64(0))
	testCheck(c, GreaterOrEqualThan, true, "", byte(2), byte(1))
	testCheck(c, GreaterOrEqualThan, true, "", int(2), int(1))
	testCheck(c, GreaterOrEqualThan, true, "", int8(2), int8(1))
	testCheck(c, GreaterOrEqualThan, true, "", int16(2), int16(1))
	testCheck(c, GreaterOrEqualThan, true, "", int32(2), int32(1))
	testCheck(c, GreaterOrEqualThan, true, "", int64(2), int64(1))
	testCheck(c, GreaterOrEqualThan, true, "", uint(2), uint(1))
	testCheck(c, GreaterOrEqualThan, true, "", uint8(2), uint8(1))
	testCheck(c, GreaterOrEqualThan, true, "", uint16(2), uint16(1))
	testCheck(c, GreaterOrEqualThan, true, "", uint32(2), uint32(1))
	testCheck(c, GreaterOrEqualThan, true, "", uint64(2), uint64(1))
	testCheck(c, GreaterOrEqualThan, true, "", myTime, "2017-12-31")
	testCheck(c, GreaterOrEqualThan, true, "", 2, 2)
	testCheck(c, GreaterOrEqualThan, true, "", -1, -1)
	testCheck(c, GreaterOrEqualThan, true, "", float32(2), float32(2))
	testCheck(c, GreaterOrEqualThan, true, "", float64(2), float64(2))
	testCheck(c, GreaterOrEqualThan, true, "", byte(2), byte(2))
	testCheck(c, GreaterOrEqualThan, true, "", int(2), int(2))
	testCheck(c, GreaterOrEqualThan, true, "", int8(2), int8(2))
	testCheck(c, GreaterOrEqualThan, true, "", int16(2), int16(2))
	testCheck(c, GreaterOrEqualThan, true, "", int32(2), int32(2))
	testCheck(c, GreaterOrEqualThan, true, "", int64(2), int64(2))
	testCheck(c, GreaterOrEqualThan, true, "", uint(2), uint(2))
	testCheck(c, GreaterOrEqualThan, true, "", uint8(2), uint8(2))
	testCheck(c, GreaterOrEqualThan, true, "", uint16(2), uint16(2))
	testCheck(c, GreaterOrEqualThan, true, "", uint32(2), uint32(2))
	testCheck(c, GreaterOrEqualThan, true, "", uint64(2), uint64(2))
	// testCheck(c, GreaterOrEqualThan, true, "", myTime, "2018-01-01")
}

func (s *CommonCheckerS) TestGreaterOrEqualThanValidsDifferent(c *check.C) {
	myTime, err := time.Parse("2006-01-02", "2018-01-01")
	if err != nil {
		c.Fatal(err)
	}

	testCheck(c, GreaterOrEqualThan, false, "", 1, 2)
	testCheck(c, GreaterOrEqualThan, false, "", -1, 0)
	testCheck(c, GreaterOrEqualThan, false, "", float32(0), float32(2))
	testCheck(c, GreaterOrEqualThan, false, "", float64(0), float64(2))
	testCheck(c, GreaterOrEqualThan, false, "", byte(1), byte(2))
	testCheck(c, GreaterOrEqualThan, false, "", int(1), int(2))
	testCheck(c, GreaterOrEqualThan, false, "", int8(1), int8(2))
	testCheck(c, GreaterOrEqualThan, false, "", int16(1), int16(2))
	testCheck(c, GreaterOrEqualThan, false, "", int32(1), int32(2))
	testCheck(c, GreaterOrEqualThan, false, "", int64(1), int64(2))
	testCheck(c, GreaterOrEqualThan, false, "", uint(1), uint(2))
	testCheck(c, GreaterOrEqualThan, false, "", uint8(1), uint8(2))
	testCheck(c, GreaterOrEqualThan, false, "", uint16(1), uint16(2))
	testCheck(c, GreaterOrEqualThan, false, "", uint32(1), uint32(2))
	testCheck(c, GreaterOrEqualThan, false, "", uint64(1), uint64(2))
	testCheck(c, GreaterOrEqualThan, false, "", myTime, "2018-01-01")
	testCheck(c, GreaterOrEqualThan, false, "", myTime, "2018-01-02")
}

func (s *CommonCheckerS) TestGreaterOrEqualThanInvalids(c *check.C) {
	myTime, err := time.Parse("2006-01-02", "2018-01-01")
	if err != nil {
		c.Fatal(err)
	}

	testCheck(c, GreaterOrEqualThan, false, "obtained value and expected value have not the same type.", 1, "string")
	testCheck(c, GreaterOrEqualThan, false, "obtained value and expected value have not the same type.", "string", 1)
	testCheck(c, GreaterOrEqualThan, false, "obtained value and expected value have not the same type.", 1, complex128(1+2i))
	testCheck(c, GreaterOrEqualThan, false, "expected must be a Time struct, or parseable.", myTime, 1)
	testCheck(c, GreaterOrEqualThan, false, "expected must be a Time struct, or parseable.", myTime, "invalid")
}

func (s *CommonCheckerS) TestLessThanInfo(c *check.C) {
	testInfo(c, LessThan, "LessThan", []string{"obtained", "expected"})
}

func (s *CommonCheckerS) TestLessThanValidsLess(c *check.C) {
	myTime, err := time.Parse("2006-01-02", "2018-01-01")
	if err != nil {
		c.Fatal(err)
	}

	testCheck(c, LessThan, true, "", 1, 2)
	testCheck(c, LessThan, true, "", -1, 0)
	testCheck(c, LessThan, true, "", float32(0), float32(2))
	testCheck(c, LessThan, true, "", float64(0), float64(2))
	testCheck(c, LessThan, true, "", byte(1), byte(2))
	testCheck(c, LessThan, true, "", int(1), int(2))
	testCheck(c, LessThan, true, "", int8(1), int8(2))
	testCheck(c, LessThan, true, "", int16(1), int16(2))
	testCheck(c, LessThan, true, "", int32(1), int32(2))
	testCheck(c, LessThan, true, "", int64(1), int64(2))
	testCheck(c, LessThan, true, "", uint(1), uint(2))
	testCheck(c, LessThan, true, "", uint8(1), uint8(2))
	testCheck(c, LessThan, true, "", uint16(1), uint16(2))
	testCheck(c, LessThan, true, "", uint32(1), uint32(2))
	testCheck(c, LessThan, true, "", uint64(1), uint64(2))
	testCheck(c, LessThan, true, "", myTime, "2018-01-02")
}

func (s *CommonCheckerS) TestLessThanValidsDifferent(c *check.C) {
	myTime, err := time.Parse("2006-01-02", "2018-01-01")
	if err != nil {
		c.Fatal(err)
	}

	testCheck(c, LessThan, false, "", 2, 1)
	testCheck(c, LessThan, false, "", 0, -1)
	testCheck(c, LessThan, false, "", float32(2), float32(0))
	testCheck(c, LessThan, false, "", float64(2), float64(0))
	testCheck(c, LessThan, false, "", byte(2), byte(1))
	testCheck(c, LessThan, false, "", int(2), int(1))
	testCheck(c, LessThan, false, "", int8(2), int8(1))
	testCheck(c, LessThan, false, "", int16(2), int16(1))
	testCheck(c, LessThan, false, "", int32(2), int32(1))
	testCheck(c, LessThan, false, "", int64(2), int64(1))
	testCheck(c, LessThan, false, "", uint(2), uint(1))
	testCheck(c, LessThan, false, "", uint8(2), uint8(1))
	testCheck(c, LessThan, false, "", uint16(2), uint16(1))
	testCheck(c, LessThan, false, "", uint32(2), uint32(1))
	testCheck(c, LessThan, false, "", uint64(2), uint64(1))
	testCheck(c, LessThan, false, "", myTime, "2018-01-01")
	testCheck(c, LessThan, false, "", myTime, "2017-12-31")
}

func (s *CommonCheckerS) TestLessThanInvalids(c *check.C) {
	myTime, err := time.Parse("2006-01-02", "2018-01-01")
	if err != nil {
		c.Fatal(err)
	}

	testCheck(c, LessThan, false, "obtained value and expected value have not the same type.", 1, "string")
	testCheck(c, LessThan, false, "obtained value and expected value have not the same type.", "string", 1)
	testCheck(c, LessThan, false, "obtained value and expected value have not the same type.", 1, complex128(1+2i))
	testCheck(c, LessThan, false, "expected must be a Time struct, or parseable.", myTime, 1)
	testCheck(c, LessThan, false, "expected must be a Time struct, or parseable.", myTime, "invalid")
}

func (s *CommonCheckerS) TestLessEqualThanInfo(c *check.C) {
	testInfo(c, LessOrEqualThan, "LessOrEqualThan", []string{"obtained", "expected"})
}

func (s *CommonCheckerS) TestLessOrEqualThanValidsLessOrEquals(c *check.C) {
	myTime, err := time.Parse("2006-01-02", "2018-01-01")
	if err != nil {
		c.Fatal(err)
	}

	testCheck(c, LessOrEqualThan, true, "", 1, 2)
	testCheck(c, LessOrEqualThan, true, "", -1, 0)
	testCheck(c, LessOrEqualThan, true, "", float32(0), float32(2))
	testCheck(c, LessOrEqualThan, true, "", float64(0), float64(2))
	testCheck(c, LessOrEqualThan, true, "", byte(1), byte(2))
	testCheck(c, LessOrEqualThan, true, "", int(1), int(2))
	testCheck(c, LessOrEqualThan, true, "", int8(1), int8(2))
	testCheck(c, LessOrEqualThan, true, "", int16(1), int16(2))
	testCheck(c, LessOrEqualThan, true, "", int32(1), int32(2))
	testCheck(c, LessOrEqualThan, true, "", int64(1), int64(2))
	testCheck(c, LessOrEqualThan, true, "", uint(1), uint(2))
	testCheck(c, LessOrEqualThan, true, "", uint8(1), uint8(2))
	testCheck(c, LessOrEqualThan, true, "", uint16(1), uint16(2))
	testCheck(c, LessOrEqualThan, true, "", uint32(1), uint32(2))
	testCheck(c, LessOrEqualThan, true, "", uint64(1), uint64(2))
	testCheck(c, LessOrEqualThan, true, "", myTime, "2018-01-02")
	testCheck(c, LessOrEqualThan, true, "", 1, 1)
	testCheck(c, LessOrEqualThan, true, "", -1, -1)
	testCheck(c, LessOrEqualThan, true, "", float32(2), float32(2))
	testCheck(c, LessOrEqualThan, true, "", float64(2), float64(2))
	testCheck(c, LessOrEqualThan, true, "", byte(2), byte(2))
	testCheck(c, LessOrEqualThan, true, "", int(2), int(2))
	testCheck(c, LessOrEqualThan, true, "", int8(2), int8(2))
	testCheck(c, LessOrEqualThan, true, "", int16(2), int16(2))
	testCheck(c, LessOrEqualThan, true, "", int32(2), int32(2))
	testCheck(c, LessOrEqualThan, true, "", int64(2), int64(2))
	testCheck(c, LessOrEqualThan, true, "", uint(2), uint(2))
	testCheck(c, LessOrEqualThan, true, "", uint8(2), uint8(2))
	testCheck(c, LessOrEqualThan, true, "", uint16(2), uint16(2))
	testCheck(c, LessOrEqualThan, true, "", uint32(2), uint32(2))
	testCheck(c, LessOrEqualThan, true, "", uint64(2), uint64(2))
	// testCheck(c, LessOrEqualThan, true, "", myTime, "2018-01-02")
}

func (s *CommonCheckerS) TestLessOrEqualThanValidsDifferent(c *check.C) {
	myTime, err := time.Parse("2006-01-02", "2018-01-01")
	if err != nil {
		c.Fatal(err)
	}

	testCheck(c, LessOrEqualThan, false, "", 2, 1)
	testCheck(c, LessOrEqualThan, false, "", 0, -1)
	testCheck(c, LessOrEqualThan, false, "", float32(2), float32(0))
	testCheck(c, LessOrEqualThan, false, "", float64(2), float64(0))
	testCheck(c, LessOrEqualThan, false, "", byte(2), byte(1))
	testCheck(c, LessOrEqualThan, false, "", int(2), int(1))
	testCheck(c, LessOrEqualThan, false, "", int8(2), int8(1))
	testCheck(c, LessOrEqualThan, false, "", int16(2), int16(1))
	testCheck(c, LessOrEqualThan, false, "", int32(2), int32(1))
	testCheck(c, LessOrEqualThan, false, "", int64(2), int64(1))
	testCheck(c, LessOrEqualThan, false, "", uint(2), uint(1))
	testCheck(c, LessOrEqualThan, false, "", uint8(2), uint8(1))
	testCheck(c, LessOrEqualThan, false, "", uint16(2), uint16(1))
	testCheck(c, LessOrEqualThan, false, "", uint32(2), uint32(1))
	testCheck(c, LessOrEqualThan, false, "", uint64(2), uint64(1))
	testCheck(c, LessOrEqualThan, false, "", myTime, "2018-01-01")
	testCheck(c, LessOrEqualThan, false, "", myTime, "2017-12-31")
}

func (s *CommonCheckerS) TestLessOrEqualThanInvalids(c *check.C) {
	myTime, err := time.Parse("2006-01-02", "2018-01-01")
	if err != nil {
		c.Fatal(err)
	}

	testCheck(c, LessOrEqualThan, false, "obtained value and expected value have not the same type.", 1, "string")
	testCheck(c, LessOrEqualThan, false, "obtained value and expected value have not the same type.", "string", 1)
	testCheck(c, LessOrEqualThan, false, "obtained value and expected value have not the same type.", 1, complex128(1+2i))
	testCheck(c, LessOrEqualThan, false, "expected must be a Time struct, or parseable.", myTime, 1)
	testCheck(c, LessOrEqualThan, false, "expected must be a Time struct, or parseable.", myTime, "invalid")
}

func Test(t *testing.T) {
	check.TestingT(t)
}

func testInfo(c *check.C, checker check.Checker, name string, paramNames []string) {
	info := checker.Info()
	if info.Name != name {
		c.Fatalf("Got name %s, expected %s", info.Name, name)
	}
	if !reflect.DeepEqual(info.Params, paramNames) {
		c.Fatalf("Got param names %#v, expected %#v", info.Params, paramNames)
	}
}

func testCheck(c *check.C, checker check.Checker, expectedResult bool, expectedError string, params ...interface{}) ([]interface{}, []string) {
	info := checker.Info()
	if len(params) != len(info.Params) {
		c.Fatalf("unexpected param count in test; expected %d got %d", len(info.Params), len(params))
	}
	names := append([]string{}, info.Params...)
	result, error := checker.Check(params, names)
	if result != expectedResult || error != expectedError {
		c.Fatalf("%s.Check(%#v) returned (%#v, %#v) rather than (%#v, %#v)",
			info.Name, params, result, error, expectedResult, expectedError)
	}
	return params, names
}
