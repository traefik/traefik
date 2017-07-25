package ratelimit

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/mailgun/timetools"
	"github.com/vulcand/oxy/testutils"
	"github.com/vulcand/oxy/utils"

	. "gopkg.in/check.v1"
)

type LimiterSuite struct {
	clock *timetools.FreezedTime
}

var _ = Suite(&LimiterSuite{})

func (s *LimiterSuite) SetUpSuite(c *C) {
	s.clock = &timetools.FreezedTime{
		CurrentTime: time.Date(2012, 3, 4, 5, 6, 7, 0, time.UTC),
	}
}

func (s *LimiterSuite) TestRateSetAdd(c *C) {
	rs := NewRateSet()

	// Invalid period
	err := rs.Add(0, 1, 1)
	c.Assert(err, NotNil)

	// Invalid Average
	err = rs.Add(time.Second, 0, 1)
	c.Assert(err, NotNil)

	// Invalid Burst
	err = rs.Add(time.Second, 1, 0)
	c.Assert(err, NotNil)

	err = rs.Add(time.Second, 1, 1)
	c.Assert(err, IsNil)
	c.Assert("map[1s:rate(1/1s, burst=1)]", Equals, fmt.Sprint(rs))
}

// We've hit the limit and were able to proceed on the next time run
func (s *LimiterSuite) TestHitLimit(c *C) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello"))
	})

	rates := NewRateSet()
	rates.Add(time.Second, 1, 1)

	l, err := New(handler, headerLimit, rates, Clock(s.clock))
	c.Assert(err, IsNil)

	srv := httptest.NewServer(l)
	defer srv.Close()

	re, _, err := testutils.Get(srv.URL, testutils.Header("Source", "a"))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)

	// Next request from the same source hits rate limit
	re, _, err = testutils.Get(srv.URL, testutils.Header("Source", "a"))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, 429)

	// Second later, the request from this ip will succeed
	s.clock.Sleep(time.Second)
	re, _, err = testutils.Get(srv.URL, testutils.Header("Source", "a"))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)
}

// We've failed to extract client ip
func (s *LimiterSuite) TestFailure(c *C) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello"))
	})

	rates := NewRateSet()
	rates.Add(time.Second, 1, 1)

	l, err := New(handler, faultyExtract, rates, Clock(s.clock))
	c.Assert(err, IsNil)

	srv := httptest.NewServer(l)
	defer srv.Close()

	re, _, err := testutils.Get(srv.URL, testutils.Header("Source", "a"))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusInternalServerError)
}

// Make sure rates from different ips are controlled separatedly
func (s *LimiterSuite) TestIsolation(c *C) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello"))
	})

	rates := NewRateSet()
	rates.Add(time.Second, 1, 1)

	l, err := New(handler, headerLimit, rates, Clock(s.clock))
	c.Assert(err, IsNil)

	srv := httptest.NewServer(l)
	defer srv.Close()

	re, _, err := testutils.Get(srv.URL, testutils.Header("Source", "a"))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)

	// Next request from the same source hits rate limit
	re, _, err = testutils.Get(srv.URL, testutils.Header("Source", "a"))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, 429)

	// The request from other source can proceed
	re, _, err = testutils.Get(srv.URL, testutils.Header("Source", "b"))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)
}

// Make sure that expiration works (Expiration is triggered after significant amount of time passes)
func (s *LimiterSuite) TestExpiration(c *C) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello"))
	})

	rates := NewRateSet()
	rates.Add(time.Second, 1, 1)

	l, err := New(handler, headerLimit, rates, Clock(s.clock))
	c.Assert(err, IsNil)

	srv := httptest.NewServer(l)
	defer srv.Close()

	re, _, err := testutils.Get(srv.URL, testutils.Header("Source", "a"))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)

	// Next request from the same source hits rate limit
	re, _, err = testutils.Get(srv.URL, testutils.Header("Source", "a"))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, 429)

	// 24 hours later, the request from this ip will succeed
	s.clock.Sleep(24 * time.Hour)
	re, _, err = testutils.Get(srv.URL, testutils.Header("Source", "a"))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)
}

// If rate limiting configuration is valid, then it is applied.
func (s *LimiterSuite) TestExtractRates(c *C) {
	// Given
	extractRates := func(*http.Request) (*RateSet, error) {
		rates := NewRateSet()
		rates.Add(time.Second, 2, 2)
		rates.Add(60*time.Second, 10, 10)
		return rates, nil
	}
	rates := NewRateSet()
	rates.Add(time.Second, 1, 1)

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello"))
	})

	tl, err := New(handler, headerLimit, rates, Clock(s.clock), ExtractRates(RateExtractorFunc(extractRates)))
	c.Assert(err, IsNil)

	srv := httptest.NewServer(tl)
	defer srv.Close()

	// When/Then: The configured rate is applied, which 2 req/second
	re, _, err := testutils.Get(srv.URL, testutils.Header("Source", "a"))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)

	re, _, err = testutils.Get(srv.URL, testutils.Header("Source", "a"))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)

	re, _, err = testutils.Get(srv.URL, testutils.Header("Source", "a"))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, 429)

	s.clock.Sleep(time.Second)
	re, _, err = testutils.Get(srv.URL, testutils.Header("Source", "a"))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)
}

// If configMapper returns error, then the default rate is applied.
func (s *LimiterSuite) TestBadRateExtractor(c *C) {
	// Given
	extractor := func(*http.Request) (*RateSet, error) {
		return nil, fmt.Errorf("Boom!")
	}
	rates := NewRateSet()
	rates.Add(time.Second, 1, 1)

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello"))
	})

	l, err := New(handler, headerLimit, rates, Clock(s.clock), ExtractRates(RateExtractorFunc(extractor)))
	c.Assert(err, IsNil)

	srv := httptest.NewServer(l)
	defer srv.Close()

	// When/Then: The default rate is applied, which 1 req/second
	re, _, err := testutils.Get(srv.URL, testutils.Header("Source", "a"))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)

	re, _, err = testutils.Get(srv.URL, testutils.Header("Source", "a"))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, 429)

	s.clock.Sleep(time.Second)
	re, _, err = testutils.Get(srv.URL, testutils.Header("Source", "a"))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)
}

// If configMapper returns empty rates, then the default rate is applied.
func (s *LimiterSuite) TestExtractorEmpty(c *C) {
	// Given
	extractor := func(*http.Request) (*RateSet, error) {
		return NewRateSet(), nil
	}
	rates := NewRateSet()
	rates.Add(time.Second, 1, 1)

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello"))
	})

	l, err := New(handler, headerLimit, rates, Clock(s.clock), ExtractRates(RateExtractorFunc(extractor)))
	c.Assert(err, IsNil)

	srv := httptest.NewServer(l)
	defer srv.Close()

	// When/Then: The default rate is applied, which 1 req/second
	re, _, err := testutils.Get(srv.URL, testutils.Header("Source", "a"))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)

	re, _, err = testutils.Get(srv.URL, testutils.Header("Source", "a"))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, 429)

	s.clock.Sleep(time.Second)
	re, _, err = testutils.Get(srv.URL, testutils.Header("Source", "a"))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)
}

func (s *LimiterSuite) TestInvalidParams(c *C) {
	// Rates are missing
	rs := NewRateSet()
	rs.Add(time.Second, 1, 1)

	// Empty
	_, err := New(nil, nil, rs)
	c.Assert(err, NotNil)

	// Rates are empty
	_, err = New(nil, nil, NewRateSet())
	c.Assert(err, NotNil)

	// Bad capacity
	_, err = New(nil, headerLimit, rs, Capacity(-1))
	c.Assert(err, NotNil)
}

// We've hit the limit and were able to proceed on the next time run
func (s *LimiterSuite) TestOptions(c *C) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello"))
	})

	rates := NewRateSet()
	rates.Add(time.Second, 1, 1)

	errHandler := utils.ErrorHandlerFunc(func(w http.ResponseWriter, req *http.Request, err error) {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte(http.StatusText(http.StatusTeapot)))
	})

	buf := &bytes.Buffer{}
	log := utils.NewFileLogger(buf, utils.INFO)

	l, err := New(handler, headerLimit, rates, ErrorHandler(errHandler), Logger(log), Clock(s.clock))
	c.Assert(err, IsNil)

	srv := httptest.NewServer(l)
	defer srv.Close()

	re, _, err := testutils.Get(srv.URL, testutils.Header("Source", "a"))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)

	re, _, err = testutils.Get(srv.URL, testutils.Header("Source", "a"))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusTeapot)

	c.Assert(len(buf.String()), Not(Equals), 0)
}

func headerLimiter(req *http.Request) (string, int64, error) {
	return req.Header.Get("Source"), 1, nil
}

func faultyExtractor(req *http.Request) (string, int64, error) {
	return "", -1, fmt.Errorf("oops")
}

var headerLimit = utils.ExtractorFunc(headerLimiter)
var faultyExtract = utils.ExtractorFunc(faultyExtractor)
