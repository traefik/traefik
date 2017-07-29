package shakers

import (
	"time"

	"github.com/go-check/check"
)

func init() {
	check.Suite(&TimeCheckerS{})
}

type TimeCheckerS struct{}

type randomStruct struct {
	foo string
	baz int
}

func (s *TimeCheckerS) TestIsBefore(c *check.C) {
	testInfo(c, IsBefore, "IsBefore", []string{"obtained", "expected"})
}

func (s *TimeCheckerS) TestIsAfter(c *check.C) {
	testInfo(c, IsAfter, "IsAfter", []string{"obtained", "expected"})
}

func (s *TimeCheckerS) TestIsBetweenInfo(c *check.C) {
	testInfo(c, IsBetween, "IsBetween", []string{"obtained", "start", "end"})
}

func (s *TimeCheckerS) TestTimeEquals(c *check.C) {
	testInfo(c, TimeEquals, "TimeEquals", []string{"obtained", "expected"})
}

func (s *TimeCheckerS) TestTimeIgnore(c *check.C) {
	testInfo(c, TimeIgnore(TimeEquals, time.Second), "TimeIgnore(TimeEquals, 1s)", []string{"obtained", "expected"})
	testInfo(c, TimeIgnore(IsBefore, time.Minute), "TimeIgnore(IsBefore, 1m0s)", []string{"obtained", "expected"})
	testInfo(c, TimeIgnore(IsBetween, time.Hour), "TimeIgnore(IsBetween, 1h0m0s)", []string{"obtained", "start", "end"})
}

func (s *TimeCheckerS) TestIsBeforeValid(c *check.C) {
	before := []struct {
		format   string
		obtained string
		t        interface{}
	}{
		{
			format:   "2006-01-02",
			obtained: "2018-01-01",
			t:        "2018-01-02",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-01T15:04:05Z",
			t:        "2018-01-02T15:04:05Z",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:03:05Z",
			t:        "2018-01-02T15:04:05Z",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-01T15:04:05+07:00",
			t:        "2018-01-02T15:04:05+07:00",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:03:05+07:00",
			t:        "2018-01-02T15:04:05+07:00",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:04:05+08:00",
			t:        "2018-01-02T15:04:05+07:00",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:04:05-06:00",
			t:        "2018-01-02T15:04:05-07:00",
		},
		{
			format:   time.RFC3339Nano,
			obtained: "2018-01-01T15:04:05.999999999Z",
			t:        "2018-01-02T15:04:05.999999999Z",
		},
		{
			format:   time.RFC3339Nano,
			obtained: "2018-01-02T15:03:05.999999999Z",
			t:        "2018-01-02T15:04:05.999999999Z",
		},
		{
			format:   time.RFC3339Nano,
			obtained: "2018-01-01T15:04:05.999999999+07:00",
			t:        "2018-01-02T15:04:05.999999999+07:00",
		},
		{
			format:   time.RFC3339Nano,
			obtained: "2018-01-02T15:03:05.999999999+07:00",
			t:        "2018-01-02T15:04:05.999999999+07:00",
		},
		{
			format:   time.RFC3339Nano,
			obtained: "2018-01-02T15:04:05.999999999+08:00",
			t:        "2018-01-02T15:04:05.999999999+07:00",
		},
		{
			format:   time.RFC3339Nano,
			obtained: "2018-01-02T15:04:05.999999999-06:00",
			t:        "2018-01-02T15:04:05.999999999-07:00",
		},
		{
			format:   time.RFC822,
			obtained: "01 Jan 18 15:04 MST",
			t:        "02 Jan 18 15:04 MST",
		},
		{
			format:   time.RFC822,
			obtained: "01 Jan 18 15:03 MST",
			t:        "01 Jan 18 15:04 MST",
		},
		{
			format:   time.RFC822Z,
			obtained: "01 Jan 18 15:04 +0700",
			t:        "02 Jan 18 15:04 +0700",
		},
		{
			format:   time.RFC822Z,
			obtained: "01 Jan 18 15:03 +0700",
			t:        "01 Jan 18 15:04 +0700",
		},
		{
			format:   time.RFC822Z,
			obtained: "01 Jan 18 15:04 +0800",
			t:        "01 Jan 18 15:04 +0700",
		},
		{
			format:   time.RFC822Z,
			obtained: "01 Jan 18 15:04 -0600",
			t:        "01 Jan 18 15:04 -0700",
		},
	}
	for _, d := range before {
		obtained, err := time.Parse(d.format, d.obtained)
		c.Assert(err, check.IsNil)
		testCheck(c, IsBefore, true, "", obtained, d.t)
	}
}

func (s *TimeCheckerS) TestIsBeforeValidAfter(c *check.C) {
	after := []struct {
		format   string
		obtained string
		t        interface{}
	}{
		{
			format:   "2006-01-02",
			obtained: "2018-01-12",
			t:        "2018-01-02",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-12T15:04:05Z",
			t:        "2018-01-02T15:04:05Z",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:05:05Z",
			t:        "2018-01-02T15:04:05Z",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:04:05+06:00",
			t:        "2018-01-02T15:04:05+07:00",
		},
		{
			format:   time.RFC3339Nano,
			obtained: "2018-01-02T15:04:05.999999995Z",
			t:        "2018-01-02T15:04:05.999999990Z",
		},
		{
			format:   time.RFC3339Nano,
			obtained: "2018-01-02T15:04:05.999999999-07:00",
			t:        "2018-01-02T15:04:05.999999999-06:00",
		},
		{
			format:   time.RFC822,
			obtained: "02 Jan 18 15:05 MST",
			t:        "02 Jan 18 15:04 MST",
		},
		{
			format:   time.RFC822Z,
			obtained: "02 Jan 18 15:04 +0700",
			t:        "02 Jan 18 15:04 +0800",
		},
		{
			format:   "2006-01-02",
			obtained: "2018-01-02",
			t:        "2018-01-02",
		},
	}
	for _, d := range after {
		obtained, err := time.Parse(d.format, d.obtained)
		c.Assert(err, check.IsNil)
		testCheck(c, IsBefore, false, "", obtained, d.t)
	}
}

func (s *TimeCheckerS) TestIsBeforeInvalids(c *check.C) {
	// Nils
	testCheck(c, IsBefore, false, "expected must be a Time struct, or parseable.", time.Now(), nil)
	testCheck(c, IsBefore, false, "obtained value is not a time.Time struct or parseable as a time.", nil, time.Now())

	// wrong type
	testCheck(c, IsBefore, false, "expected must be a Time struct, or parseable.", time.Now(), 0)
	testCheck(c, IsBefore, false, "obtained value is not a time.Time struct or parseable as a time.", 0, time.Now())
	testCheck(c, IsBefore, false, "expected must be a Time struct, or parseable.", time.Now(), randomStruct{})
	testCheck(c, IsBefore, false, "obtained value is not a time.Time struct or parseable as a time.", randomStruct{}, time.Now())

	// Invalid strings
	testCheck(c, IsBefore, false, "expected must be a Time struct, or parseable.", time.Now(), "this is not a date")
	testCheck(c, IsBefore, false, "obtained value is not a time.Time struct or parseable as a time.", "this is not a date", time.Now())

	// Invalids dates
	testCheck(c, IsBefore, false, "expected must be a Time struct, or parseable.", time.Now(), "2018-31-02")
	testCheck(c, IsBefore, false, "obtained value is not a time.Time struct or parseable as a time.", "2018-31-02", time.Now())

}

func (s *TimeCheckerS) TestIsAfterValid(c *check.C) {
	before := []struct {
		format   string
		obtained string
		t        interface{}
	}{
		{
			format:   "2006-01-02",
			obtained: "2018-01-02",
			t:        "2018-01-01",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:04:05Z",
			t:        "2018-01-01T15:04:05Z",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:04:05Z",
			t:        "2018-01-02T15:03:05Z",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:04:05+07:00",
			t:        "2018-01-01T15:04:05+07:00",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:04:05+07:00",
			t:        "2018-01-02T15:03:05+07:00",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:04:05+07:00",
			t:        "2018-01-02T15:04:05+08:00",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:04:05-07:00",
			t:        "2018-01-02T15:04:05-06:00",
		},
		{
			format:   time.RFC3339Nano,
			obtained: "2018-01-02T15:04:05.999999999Z",
			t:        "2018-01-01T15:04:05.999999999Z",
		},
		{
			format:   time.RFC3339Nano,
			obtained: "2018-01-02T15:04:05.999999999Z",
			t:        "2018-01-02T15:03:05.999999999Z",
		},
		{
			format:   time.RFC3339Nano,
			obtained: "2018-01-02T15:04:05.999999999+07:00",
			t:        "2018-01-01T15:04:05.999999999+07:00",
		},
		{
			format:   time.RFC3339Nano,
			obtained: "2018-01-02T15:04:05.999999999+07:00",
			t:        "2018-01-02T15:03:05.999999999+07:00",
		},
		{
			format:   time.RFC3339Nano,
			obtained: "2018-01-02T15:04:05.999999999+07:00",
			t:        "2018-01-02T15:04:05.999999999+08:00",
		},
		{
			format:   time.RFC3339Nano,
			obtained: "2018-01-02T15:04:05.999999999-07:00",
			t:        "2018-01-02T15:04:05.999999999-06:00",
		},
		{
			format:   time.RFC822,
			obtained: "02 Jan 18 15:04 MST",
			t:        "01 Jan 18 15:04 MST",
		},
		{
			format:   time.RFC822,
			obtained: "01 Jan 18 15:04 MST",
			t:        "01 Jan 18 15:03 MST",
		},
		{
			format:   time.RFC822Z,
			obtained: "02 Jan 18 15:04 +0700",
			t:        "01 Jan 18 15:04 +0700",
		},
		{
			format:   time.RFC822Z,
			obtained: "01 Jan 18 15:04 +0700",
			t:        "01 Jan 18 15:03 +0700",
		},
		{
			format:   time.RFC822Z,
			obtained: "01 Jan 18 15:04 +0700",
			t:        "01 Jan 18 15:04 +0800",
		},
		{
			format:   time.RFC822Z,
			obtained: "01 Jan 18 15:04 -0700",
			t:        "01 Jan 18 15:04 -0600",
		},
	}
	for _, d := range before {
		obtained, err := time.Parse(d.format, d.obtained)
		c.Assert(err, check.IsNil)
		testCheck(c, IsAfter, true, "", obtained, d.t)
	}
}

func (s *TimeCheckerS) TestIsAfterValidBefore(c *check.C) {
	after := []struct {
		format   string
		obtained string
		t        interface{}
	}{
		{
			format:   "2006-01-02",
			obtained: "2018-01-02",
			t:        "2018-01-12",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:04:05Z",
			t:        "2018-01-12T15:04:05Z",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:04:05Z",
			t:        "2018-01-02T15:05:05Z",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:04:05+07:00",
			t:        "2018-01-02T15:04:05+06:00",
		},
		{
			format:   time.RFC3339Nano,
			obtained: "2018-01-02T15:04:05.999999990Z",
			t:        "2018-01-02T15:04:05.999999995Z",
		},
		{
			format:   time.RFC3339Nano,
			obtained: "2018-01-02T15:04:05.999999999-06:00",
			t:        "2018-01-02T15:04:05.999999999-07:00",
		},
		{
			format:   time.RFC822,
			obtained: "02 Jan 18 15:04 MST",
			t:        "02 Jan 18 15:05 MST",
		},
		{
			format:   time.RFC822Z,
			obtained: "02 Jan 18 15:04 +0800",
			t:        "02 Jan 18 15:04 +0700",
		},
		{
			format:   "2006-01-02",
			obtained: "2018-01-02",
			t:        "2018-01-02",
		},
	}
	for _, d := range after {
		obtained, err := time.Parse(d.format, d.obtained)
		c.Assert(err, check.IsNil)
		testCheck(c, IsAfter, false, "", obtained, d.t)
	}
}

func (s *TimeCheckerS) TestIsAfterInvalids(c *check.C) {
	// Nils
	testCheck(c, IsAfter, false, "expected must be a Time struct, or parseable.", time.Now(), nil)
	testCheck(c, IsAfter, false, "obtained value is not a time.Time struct or parseable as a time.", nil, time.Now())

	// wrong type
	testCheck(c, IsAfter, false, "expected must be a Time struct, or parseable.", time.Now(), 0)
	testCheck(c, IsAfter, false, "obtained value is not a time.Time struct or parseable as a time.", 0, time.Now())
	testCheck(c, IsAfter, false, "expected must be a Time struct, or parseable.", time.Now(), randomStruct{})
	testCheck(c, IsAfter, false, "obtained value is not a time.Time struct or parseable as a time.", randomStruct{}, time.Now())

	// Invalid strings
	testCheck(c, IsAfter, false, "expected must be a Time struct, or parseable.", time.Now(), "this is not a date")
	testCheck(c, IsAfter, false, "obtained value is not a time.Time struct or parseable as a time.", "this is not a date", time.Now())

	// Invalids dates
	testCheck(c, IsAfter, false, "expected must be a Time struct, or parseable.", time.Now(), "2018-31-02")
	testCheck(c, IsAfter, false, "obtained value is not a time.Time struct or parseable as a time.", "2018-31-02", time.Now())

}

func (s *TimeCheckerS) TestIsBetweenValidBetween(c *check.C) {
	between := []struct {
		format   string
		obtained string
		start    interface{}
		end      interface{}
	}{
		{
			format:   "2006-01-02",
			obtained: "2018-01-02",
			start:    "2018-01-01",
			end:      "2018-01-03",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:04:05Z",
			start:    "2018-01-01T15:04:05Z",
			end:      "2018-01-03T15:04:05Z",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:04:05Z",
			start:    "2018-01-02T15:03:05Z",
			end:      "2018-01-02T15:05:05Z",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:04:05+07:00",
			start:    "2018-01-01T15:04:05+07:00",
			end:      "2018-01-03T15:04:05+07:00",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:04:05+07:00",
			start:    "2018-01-02T15:03:05+07:00",
			end:      "2018-01-02T15:05:05+07:00",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:04:05+07:00",
			start:    "2018-01-02T15:04:05+08:00",
			end:      "2018-01-02T15:04:05+06:00",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:04:05-07:00",
			start:    "2018-01-02T15:04:05-06:00",
			end:      "2018-01-02T15:04:05-08:00",
		},
		{
			format:   time.RFC3339Nano,
			obtained: "2018-01-02T15:04:05.999999999Z",
			start:    "2018-01-01T15:04:05.999999999Z",
			end:      "2018-01-03T15:04:05.999999999Z",
		},
		{
			format:   time.RFC3339Nano,
			obtained: "2018-01-02T15:04:05.999999999Z",
			start:    "2018-01-02T15:03:05.999999999Z",
			end:      "2018-01-02T15:05:05.999999999Z",
		},
		{
			format:   time.RFC3339Nano,
			obtained: "2018-01-02T15:04:05.999999999+07:00",
			start:    "2018-01-01T15:04:05.999999999+07:00",
			end:      "2018-01-03T15:04:05.999999999+07:00",
		},
		{
			format:   time.RFC3339Nano,
			obtained: "2018-01-02T15:04:05.999999999+07:00",
			start:    "2018-01-02T15:03:05.999999999+07:00",
			end:      "2018-01-02T15:05:05.999999999+07:00",
		},
		{
			format:   time.RFC3339Nano,
			obtained: "2018-01-02T15:04:05.999999999+07:00",
			start:    "2018-01-02T15:04:05.999999999+08:00",
			end:      "2018-01-02T15:04:05.999999999+06:00",
		},
		{
			format:   time.RFC3339Nano,
			obtained: "2018-01-02T15:04:05.999999999-07:00",
			start:    "2018-01-02T15:04:05.999999999-06:00",
			end:      "2018-01-02T15:04:05.999999999-08:00",
		},
		{
			format:   time.RFC822,
			obtained: "02 Jan 18 15:04 MST",
			start:    "01 Jan 18 15:04 MST",
			end:      "03 Jan 18 15:04 MST",
		},
		{
			format:   time.RFC822,
			obtained: "01 Jan 18 15:04 MST",
			start:    "01 Jan 18 15:03 MST",
			end:      "01 Jan 18 15:05 MST",
		},
		{
			format:   time.RFC822Z,
			obtained: "02 Jan 18 15:04 +0700",
			start:    "01 Jan 18 15:04 +0700",
			end:      "03 Jan 18 15:04 +0700",
		},
		{
			format:   time.RFC822Z,
			obtained: "01 Jan 18 15:04 +0700",
			start:    "01 Jan 18 15:03 +0700",
			end:      "01 Jan 18 15:05 +0700",
		},
		{
			format:   time.RFC822Z,
			obtained: "01 Jan 18 15:04 +0700",
			start:    "01 Jan 18 15:04 +0800",
			end:      "01 Jan 18 15:04 +0600",
		},
		{
			format:   time.RFC822Z,
			obtained: "01 Jan 18 15:04 -0700",
			start:    "01 Jan 18 15:04 -0600",
			end:      "01 Jan 18 15:04 -0800",
		},
	}
	for _, d := range between {
		obtained, err := time.Parse(d.format, d.obtained)
		c.Assert(err, check.IsNil)
		testCheck(c, IsBetween, true, "", obtained, d.start, d.end)
	}
}

func (s *TimeCheckerS) TestIsBetweenValidOutside(c *check.C) {
	outside := []struct {
		format   string
		obtained string
		start    interface{}
		end      interface{}
	}{
		{
			format:   "2006-01-02",
			obtained: "2018-01-02",
			start:    "2018-01-12",
			end:      "2018-01-22",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:04:05Z",
			start:    "2018-01-12T15:04:05Z",
			end:      "2018-01-22T15:04:05Z",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:04:05Z",
			start:    "2018-01-02T15:05:05Z",
			end:      "2018-01-02T15:06:05Z",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:04:05+07:00",
			start:    "2018-01-02T15:04:05+06:00",
			end:      "2018-01-02T15:04:05+05:00",
		},
		{
			format:   time.RFC3339Nano,
			obtained: "2018-01-02T15:04:05.999999990Z",
			start:    "2018-01-02T15:04:05.999999995Z",
			end:      "2018-01-02T15:04:05.999999996Z",
		},
		{
			format:   time.RFC3339Nano,
			obtained: "2018-01-02T15:04:05.999999999-06:00",
			start:    "2018-01-02T15:04:05.999999999-07:00",
			end:      "2018-01-02T15:04:05.999999999-08:00",
		},
		{
			format:   time.RFC822,
			obtained: "02 Jan 18 15:04 MST",
			start:    "02 Jan 18 15:05 MST",
			end:      "02 Jan 18 15:06 MST",
		},
		{
			format:   time.RFC822Z,
			obtained: "02 Jan 18 15:04 +0700",
			start:    "02 Jan 18 15:04 +0800",
			end:      "02 Jan 18 15:04 +0900",
		},
		{
			format:   "2006-01-02",
			obtained: "2018-01-02",
			start:    "2018-01-02",
			end:      "2018-01-10",
		},
		{
			format:   "2006-01-02",
			obtained: "2018-01-02",
			start:    "2018-01-01",
			end:      "2018-01-02",
		},
		{
			format:   "2006-01-02",
			obtained: "2018-01-02",
			start:    "2018-01-02",
			end:      "2018-01-02",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:04:05Z",
			start:    "2018-01-02T15:03:05Z",
			end:      "2018-01-02T15:04:05Z",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:04:05Z",
			start:    "2018-01-02T15:04:05Z",
			end:      "2018-01-02T15:05:05Z",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:04:05Z",
			start:    "2018-01-02T15:04:05Z",
			end:      "2018-01-02T15:04:05Z",
		},
	}
	for _, d := range outside {
		obtained, err := time.Parse(d.format, d.obtained)
		c.Assert(err, check.IsNil)
		testCheck(c, IsBetween, false, "", obtained, d.start, d.end)
	}
}

func (s *TimeCheckerS) TestIsBetweenInvalids(c *check.C) {
	// Nils
	testCheck(c, IsBetween, false, "start must be a Time struct, or parseable.", time.Now(), nil, time.Now())
	testCheck(c, IsBetween, false, "end must be a Time struct, or parseable.", time.Now(), time.Now(), nil)
	testCheck(c, IsBetween, false, "obtained value is not a time.Time struct or parseable as a time.", nil, time.Now(), time.Now())

	// wrong type
	testCheck(c, IsBetween, false, "start must be a Time struct, or parseable.", time.Now(), 0, time.Now())
	testCheck(c, IsBetween, false, "end must be a Time struct, or parseable.", time.Now(), time.Now(), 0)
	testCheck(c, IsBetween, false, "obtained value is not a time.Time struct or parseable as a time.", 0, time.Now(), time.Now())
	testCheck(c, IsBetween, false, "start must be a Time struct, or parseable.", time.Now(), randomStruct{}, time.Now())
	testCheck(c, IsBetween, false, "end must be a Time struct, or parseable.", time.Now(), time.Now(), randomStruct{})
	testCheck(c, IsBetween, false, "obtained value is not a time.Time struct or parseable as a time.", randomStruct{}, time.Now(), time.Now())

	// Invalid strings
	testCheck(c, IsBetween, false, "start must be a Time struct, or parseable.", time.Now(), "this is not a date", time.Now())
	testCheck(c, IsBetween, false, "end must be a Time struct, or parseable.", time.Now(), time.Now(), "this is not a date")
	testCheck(c, IsBetween, false, "obtained value is not a time.Time struct or parseable as a time.", "this is not a date", time.Now(), time.Now())

	// Invalids dates
	testCheck(c, IsBetween, false, "start must be a Time struct, or parseable.", time.Now(), "2018-31-02", time.Now())
	testCheck(c, IsBetween, false, "end must be a Time struct, or parseable.", time.Now(), time.Now(), "2018-31-02")
	testCheck(c, IsBetween, false, "obtained value is not a time.Time struct or parseable as a time.", "2018-31-02", time.Now(), time.Now())

}

func (s *TimeCheckerS) TestTimeEqualsValid(c *check.C) {
	before := []struct {
		format   string
		obtained string
		t        interface{}
	}{
		{
			format:   "2006-01-02",
			obtained: "2018-01-02",
			t:        "2018-01-02",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:04:05Z",
			t:        "2018-01-02T15:04:05Z",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:04:05Z",
			t:        "2018-01-02T15:04:05Z",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-01T15:04:05+07:00",
			t:        "2018-01-01T15:04:05+07:00",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:04:05-07:00",
			t:        "2018-01-02T15:04:05-07:00",
		},
		{
			format:   time.RFC3339Nano,
			obtained: "2018-01-01T15:04:05.999999999Z",
			t:        "2018-01-01T15:04:05.999999999Z",
		},
		{
			format:   time.RFC3339Nano,
			obtained: "2018-01-01T15:04:05.999999999+07:00",
			t:        "2018-01-01T15:04:05.999999999+07:00",
		},
		{
			format:   time.RFC3339Nano,
			obtained: "2018-01-02T15:04:05.999999999-07:00",
			t:        "2018-01-02T15:04:05.999999999-07:00",
		},
		{
			format:   time.RFC822,
			obtained: "01 Jan 18 15:04 MST",
			t:        "01 Jan 18 15:04 MST",
		},
		{
			format:   time.RFC822Z,
			obtained: "01 Jan 18 15:04 +0700",
			t:        "01 Jan 18 15:04 +0700",
		},
		{
			format:   time.RFC822Z,
			obtained: "01 Jan 18 15:04 -0700",
			t:        "01 Jan 18 15:04 -0700",
		},
	}
	for _, d := range before {
		obtained, err := time.Parse(d.format, d.obtained)
		c.Assert(err, check.IsNil)
		testCheck(c, TimeEquals, true, "", obtained, d.t)
	}
}

func (s *TimeCheckerS) TestTimeEqualsDifferent(c *check.C) {
	after := []struct {
		format   string
		obtained string
		t        interface{}
	}{
		{
			format:   "2006-01-02",
			obtained: "2018-01-02",
			t:        "2018-01-12",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:04:05Z",
			t:        "2018-01-12T15:04:05Z",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:04:05Z",
			t:        "2018-01-02T15:05:05Z",
		},
		{
			format:   time.RFC3339,
			obtained: "2018-01-02T15:04:05+07:00",
			t:        "2018-01-02T15:04:05+06:00",
		},
		{
			format:   time.RFC3339Nano,
			obtained: "2018-01-02T15:04:05.999999990Z",
			t:        "2018-01-02T15:04:05.999999995Z",
		},
		{
			format:   time.RFC3339Nano,
			obtained: "2018-01-02T15:04:05.999999999-06:00",
			t:        "2018-01-02T15:04:05.999999999-07:00",
		},
		{
			format:   time.RFC822,
			obtained: "02 Jan 18 15:04 MST",
			t:        "02 Jan 18 15:05 MST",
		},
		{
			format:   time.RFC822Z,
			obtained: "02 Jan 18 15:04 +0800",
			t:        "02 Jan 18 15:04 +0700",
		},
	}
	for _, d := range after {
		obtained, err := time.Parse(d.format, d.obtained)
		c.Assert(err, check.IsNil)
		testCheck(c, TimeEquals, false, "", obtained, d.t)
	}
}

func (s *TimeCheckerS) TestTimeEqualsInvalids(c *check.C) {
	// Nils
	testCheck(c, TimeEquals, false, "expected must be a Time struct, or parseable.", time.Now(), nil)
	testCheck(c, TimeEquals, false, "obtained value is not a time.Time struct or parseable as a time.", nil, time.Now())

	// wrong type
	testCheck(c, TimeEquals, false, "expected must be a Time struct, or parseable.", time.Now(), 0)
	testCheck(c, TimeEquals, false, "obtained value is not a time.Time struct or parseable as a time.", 0, time.Now())
	testCheck(c, TimeEquals, false, "expected must be a Time struct, or parseable.", time.Now(), randomStruct{})
	testCheck(c, TimeEquals, false, "obtained value is not a time.Time struct or parseable as a time.", randomStruct{}, time.Now())

	// Invalid strings
	testCheck(c, TimeEquals, false, "expected must be a Time struct, or parseable.", time.Now(), "this is not a date")
	testCheck(c, TimeEquals, false, "obtained value is not a time.Time struct or parseable as a time.", "this is not a date", time.Now())

	// Invalids dates
	testCheck(c, TimeEquals, false, "expected must be a Time struct, or parseable.", time.Now(), "2018-31-02")
	testCheck(c, TimeEquals, false, "obtained value is not a time.Time struct or parseable as a time.", "2018-31-02", time.Now())

}

func (s *TimeCheckerS) TestTimeIgnoreValids(c *check.C) {
	testCheck(c, TimeIgnore(TimeEquals, time.Second), true, "", "2018-01-02T15:04:05Z", "2018-01-02T15:04:07Z")
	testCheck(c, TimeIgnore(TimeEquals, time.Second), false, "", "2018-01-02T15:04:05Z", "2018-01-02T15:05:05Z")
	testCheck(c, TimeIgnore(TimeEquals, time.Minute), true, "", "2018-01-02T15:04:05Z", "2018-01-02T15:06:05Z")
	testCheck(c, TimeIgnore(TimeEquals, time.Minute), false, "", "2018-01-02T16:06:05Z", "2018-01-02T15:06:05Z")
	testCheck(c, TimeIgnore(TimeEquals, time.Hour), true, "", "2018-01-01T14:06:05Z", "2018-01-01T15:06:05Z")
	testCheck(c, TimeIgnore(TimeEquals, time.Hour), false, "", "2018-01-01T15:04:05Z", "2018-01-02T15:04:05Z")

	testCheck(c, TimeIgnore(IsBefore, time.Second), true, "", "2018-01-02T15:04:05Z", "2018-01-02T15:06:05Z")
	testCheck(c, TimeIgnore(IsBefore, time.Second), false, "", "2018-01-02T15:04:05Z", "2018-01-02T15:04:07Z")
	testCheck(c, TimeIgnore(IsBefore, time.Minute), true, "", "2018-01-02T14:04:05Z", "2018-01-02T15:06:05Z")
	testCheck(c, TimeIgnore(IsBefore, time.Minute), false, "", "2018-01-02T15:04:05Z", "2018-01-02T15:06:05Z")
	testCheck(c, TimeIgnore(IsBefore, time.Hour), true, "", "2018-01-01T14:04:05Z", "2018-01-02T15:06:05Z")
	testCheck(c, TimeIgnore(IsBefore, time.Hour), false, "", "2018-01-02T15:04:05Z", "2018-01-02T16:04:05Z")
}

func (s *TimeCheckerS) TestTimeIgnoreInvalids(c *check.C) {
	// Nils
	testCheck(c, TimeIgnore(TimeEquals, time.Hour), false, "expected must be a Time struct, or parseable.", time.Now(), nil)
	testCheck(c, TimeIgnore(TimeEquals, time.Hour), false, "obtained must be a Time struct, or parseable.", nil, time.Now())

	// wrong type
	testCheck(c, TimeIgnore(TimeEquals, time.Hour), false, "expected must be a Time struct, or parseable.", time.Now(), 0)
	testCheck(c, TimeIgnore(TimeEquals, time.Hour), false, "obtained must be a Time struct, or parseable.", 0, time.Now())
	testCheck(c, TimeIgnore(TimeEquals, time.Hour), false, "expected must be a Time struct, or parseable.", time.Now(), randomStruct{})
	testCheck(c, TimeIgnore(TimeEquals, time.Hour), false, "obtained must be a Time struct, or parseable.", randomStruct{}, time.Now())

	// Invalid strings
	testCheck(c, TimeIgnore(TimeEquals, time.Hour), false, "expected must be a Time struct, or parseable.", time.Now(), "this is not a date")
	testCheck(c, TimeIgnore(TimeEquals, time.Hour), false, "obtained must be a Time struct, or parseable.", "this is not a date", time.Now())

	// Invalids dates
	testCheck(c, TimeIgnore(TimeEquals, time.Hour), false, "expected must be a Time struct, or parseable.", time.Now(), "2018-31-02")
	testCheck(c, TimeIgnore(TimeEquals, time.Hour), false, "obtained must be a Time struct, or parseable.", "2018-31-02", time.Now())

	// Nils
	testCheck(c, TimeIgnore(IsBetween, time.Hour), false, "start must be a Time struct, or parseable.", time.Now(), nil, time.Now())
	testCheck(c, TimeIgnore(IsBetween, time.Hour), false, "end must be a Time struct, or parseable.", time.Now(), time.Now(), nil)
	testCheck(c, TimeIgnore(IsBetween, time.Hour), false, "obtained must be a Time struct, or parseable.", nil, time.Now(), time.Now())

	// wrong type
	testCheck(c, TimeIgnore(IsBetween, time.Hour), false, "start must be a Time struct, or parseable.", time.Now(), 0, time.Now())
	testCheck(c, TimeIgnore(IsBetween, time.Hour), false, "end must be a Time struct, or parseable.", time.Now(), time.Now(), 0)
	testCheck(c, TimeIgnore(IsBetween, time.Hour), false, "obtained must be a Time struct, or parseable.", 0, time.Now(), time.Now())
	testCheck(c, TimeIgnore(IsBetween, time.Hour), false, "start must be a Time struct, or parseable.", time.Now(), randomStruct{}, time.Now())
	testCheck(c, TimeIgnore(IsBetween, time.Hour), false, "end must be a Time struct, or parseable.", time.Now(), time.Now(), randomStruct{})
	testCheck(c, TimeIgnore(IsBetween, time.Hour), false, "obtained must be a Time struct, or parseable.", randomStruct{}, time.Now(), time.Now())

	// Invalid strings
	testCheck(c, TimeIgnore(IsBetween, time.Hour), false, "start must be a Time struct, or parseable.", time.Now(), "this is not a date", time.Now())
	testCheck(c, TimeIgnore(IsBetween, time.Hour), false, "end must be a Time struct, or parseable.", time.Now(), time.Now(), "this is not a date")
	testCheck(c, TimeIgnore(IsBetween, time.Hour), false, "obtained must be a Time struct, or parseable.", "this is not a date", time.Now(), time.Now())

	// Invalids dates
	testCheck(c, TimeIgnore(IsBetween, time.Hour), false, "start must be a Time struct, or parseable.", time.Now(), "2018-31-02", time.Now())
	testCheck(c, TimeIgnore(IsBetween, time.Hour), false, "end must be a Time struct, or parseable.", time.Now(), time.Now(), "2018-31-02")
	testCheck(c, TimeIgnore(IsBetween, time.Hour), false, "obtained must be a Time struct, or parseable.", "2018-31-02", time.Now(), time.Now())

}
