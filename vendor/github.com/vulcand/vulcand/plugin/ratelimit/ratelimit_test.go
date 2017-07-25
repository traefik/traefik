package ratelimit

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/codegangsta/cli"
	"github.com/mailgun/timetools"
	"github.com/vulcand/oxy/testutils"
	"github.com/vulcand/vulcand/plugin"
	. "gopkg.in/check.v1"
)

func TestRL(t *testing.T) { TestingT(t) }

type RateLimitSuite struct {
	clock *timetools.FreezedTime
}

func (s *RateLimitSuite) SetUpSuite(c *C) {
	s.clock = &timetools.FreezedTime{
		CurrentTime: time.Date(2012, 3, 4, 5, 6, 7, 0, time.UTC),
	}
}

var _ = Suite(&RateLimitSuite{})

// One of the most important tests:
// Make sure the RateLimit spec is compatible and will be accepted by middleware registry
func (s *RateLimitSuite) TestSpecIsOK(c *C) {
	c.Assert(plugin.NewRegistry().AddSpec(GetSpec()), IsNil)
}

func (s *RateLimitSuite) TestFromOther(c *C) {
	rl, err := FromOther(
		RateLimit{
			PeriodSeconds: 1,
			Requests:      1,
			Burst:         10,
			Variable:      "client.ip",
			RateVar:       "request.header.X-Rates",
		})
	c.Assert(rl, NotNil)
	c.Assert(err, IsNil)
	c.Assert(fmt.Sprint(rl), Equals, "reqs/1s=1, burst=10, var=client.ip, rateVar=request.header.X-Rates")

	out, err := rl.NewHandler(nil)
	c.Assert(out, NotNil)
	c.Assert(err, IsNil)
}

func (s *RateLimitSuite) TestFromOtherNoConfigVar(c *C) {
	rl, err := FromOther(
		RateLimit{
			PeriodSeconds: 1,
			Requests:      1,
			Burst:         10,
			Variable:      "client.ip",
			RateVar:       "",
		})
	c.Assert(rl, NotNil)
	c.Assert(err, IsNil)

	out, err := rl.NewHandler(nil)
	c.Assert(out, NotNil)
	c.Assert(err, IsNil)
}

func (s *RateLimitSuite) TestFromOtherBadParams(c *C) {
	// Unknown variable
	_, err := FromOther(
		RateLimit{
			PeriodSeconds: 1,
			Requests:      1,
			Burst:         10,
			Variable:      "foo",
			RateVar:       "request.header.X-Rates",
		})
	c.Assert(err, NotNil)

	// Negative requests
	_, err = FromOther(
		RateLimit{
			PeriodSeconds: 1,
			Requests:      -1,
			Burst:         10,
			Variable:      "client.ip",
			RateVar:       "request.header.X-Rates",
		})
	c.Assert(err, NotNil)

	// Negative burst
	_, err = FromOther(
		RateLimit{
			PeriodSeconds: 1,
			Requests:      1,
			Burst:         -1,
			Variable:      "client.ip",
			RateVar:       "request.header.X-Rates",
		})
	c.Assert(err, NotNil)

	// Negative period
	_, err = FromOther(
		RateLimit{
			PeriodSeconds: -1,
			Requests:      1,
			Burst:         10,
			Variable:      "client.ip",
			RateVar:       "request.header.X-Rates",
		})
	c.Assert(err, NotNil)

	// Unknown config variable
	_, err = FromOther(
		RateLimit{
			PeriodSeconds: 1,
			Requests:      1,
			Burst:         10,
			Variable:      "client.ip",
			RateVar:       "foo",
		})
	c.Assert(err, NotNil)
}

func (s *RateLimitSuite) TestFromCli(c *C) {
	app := cli.NewApp()
	app.Name = "test"
	app.Flags = GetSpec().CliFlags
	executed := false
	app.Action = func(ctx *cli.Context) {
		executed = true
		out, err := FromCli(ctx)
		c.Assert(out, NotNil)
		c.Assert(err, IsNil)

		rl := out.(*RateLimit)
		m, err := rl.NewHandler(nil)
		c.Assert(m, NotNil)
		c.Assert(err, IsNil)
	}
	app.Run([]string{"test", "--var=client.ip", "--requests=10", "--burst=3", "--period=4"})
	c.Assert(executed, Equals, true)
}

// Middleware instance created by the factory is using rates configuration
// from the respective request header.
func (s *RateLimitSuite) TestRequestProcessing(c *C) {
	// Given
	rl, _ := FromOther(
		RateLimit{
			PeriodSeconds: 1,
			Requests:      1,
			Burst:         1,
			Variable:      "client.ip",
			RateVar:       "request.header.X-Rates",
			clock:         s.clock,
		})

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello"))
	})

	rli, err := rl.NewHandler(handler)
	c.Assert(err, IsNil)

	srv := httptest.NewServer(rli)
	defer srv.Close()

	// When/Then: The configured rate is applied, which 2 request/second, note
	// that the default rate is 1 request/second.
	hdr := testutils.Header("X-Rates", `[{"PeriodSeconds": 1, "Requests": 2}]`)

	re, _, err := testutils.Get(srv.URL, hdr)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)

	re, _, err = testutils.Get(srv.URL, hdr)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)

	re, _, err = testutils.Get(srv.URL, hdr)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, 429)

	s.clock.Sleep(time.Second)
	re, _, err = testutils.Get(srv.URL, hdr)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)

}

func (s *RateLimitSuite) TestRequestProcessingEmptyConfig(c *C) {
	// Given
	rl, _ := FromOther(
		RateLimit{
			PeriodSeconds: 1,
			Requests:      1,
			Burst:         1,
			Variable:      "client.ip",
			RateVar:       "request.header.X-Rates",
			clock:         s.clock,
		})

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello"))
	})

	rli, err := rl.NewHandler(handler)
	c.Assert(err, IsNil)

	srv := httptest.NewServer(rli)
	defer srv.Close()

	// When/Then: The default rate of 1 request/second is used.
	hdr := testutils.Header("X-Rates", `[]`)

	re, _, err := testutils.Get(srv.URL, hdr)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)

	re, _, err = testutils.Get(srv.URL, hdr)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, 429)

	s.clock.Sleep(time.Second)
	re, _, err = testutils.Get(srv.URL, hdr)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)
}

func (s *RateLimitSuite) TestRequestProcessingNoHeader(c *C) {
	// Given
	rl, _ := FromOther(
		RateLimit{
			PeriodSeconds: 1,
			Requests:      1,
			Burst:         1,
			Variable:      "client.ip",
			RateVar:       "request.header.X-Rates",
			clock:         s.clock,
		})

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello"))
	})

	rli, err := rl.NewHandler(handler)
	c.Assert(err, IsNil)

	srv := httptest.NewServer(rli)
	defer srv.Close()

	// When/Then: The default rate of 1 request/second is used.
	re, _, err := testutils.Get(srv.URL)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)

	re, _, err = testutils.Get(srv.URL)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, 429)

	s.clock.Sleep(time.Second)
	re, _, err = testutils.Get(srv.URL)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)
}

// If the rate set from the HTTP header has more then one rate for the same
// time period defined, then the one mentioned in the list last is used.
func (s *RateLimitSuite) TestRequestInvalidConfig(c *C) {
	// Given
	rl, _ := FromOther(
		RateLimit{
			PeriodSeconds: 1,
			Requests:      1,
			Burst:         1,
			Variable:      "client.ip",
			RateVar:       "request.header.X-Rates",
			clock:         s.clock,
		})

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello"))
	})

	rli, err := rl.NewHandler(handler)
	c.Assert(err, IsNil)

	srv := httptest.NewServer(rli)
	defer srv.Close()

	// When/Then: The default rate of 1 request/second is used.
	hdr := testutils.Header("X-Rates", `[{"PeriodSeconds": -1, "Requests": 10}]`)

	re, _, err := testutils.Get(srv.URL, hdr)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)

	re, _, err = testutils.Get(srv.URL, hdr)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, 429)

	s.clock.Sleep(time.Second)
	re, _, err = testutils.Get(srv.URL, hdr)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)
}

// If the rate set from the HTTP header has more then one rate for the same
// time period defined, then the one mentioned in the list last is used.
func (s *RateLimitSuite) TestRequestProcessingAmbiguousConfig(c *C) {
	// Given
	rl, _ := FromOther(
		RateLimit{
			PeriodSeconds: 1,
			Requests:      1,
			Burst:         1,
			Variable:      "client.ip",
			RateVar:       "request.header.X-Rates",
			clock:         s.clock,
		})

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello"))
	})

	rli, err := rl.NewHandler(handler)
	c.Assert(err, IsNil)

	srv := httptest.NewServer(rli)
	defer srv.Close()

	// When/Then: The last of configured rates with the same period is applied,
	// which 2 request/second, note that the default rate is 1 request/second.
	hdr := testutils.Header("X-Rates", `[{"PeriodSeconds": 1, "Requests": 10},
					                  {"PeriodSeconds": 1, "Requests": 2}]`)

	re, _, err := testutils.Get(srv.URL, hdr)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)

	re, _, err = testutils.Get(srv.URL, hdr)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)

	re, _, err = testutils.Get(srv.URL, hdr)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, 429)

	s.clock.Sleep(time.Second)
	re, _, err = testutils.Get(srv.URL, hdr)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)
}
