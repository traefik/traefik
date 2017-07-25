package cbreaker

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/mailgun/timetools"
	"github.com/vulcand/oxy/memmetrics"
	"github.com/vulcand/oxy/testutils"

	. "gopkg.in/check.v1"
)

func TestCircuitBreaker(t *testing.T) { TestingT(t) }

type CBSuite struct {
	clock *timetools.FreezedTime
}

var _ = Suite(&CBSuite{
	clock: &timetools.FreezedTime{
		CurrentTime: time.Date(2012, 3, 4, 5, 6, 7, 0, time.UTC),
	},
})

const triggerNetRatio = `NetworkErrorRatio() > 0.5`

var fallbackResponse http.Handler
var fallbackRedirect http.Handler

func (s CBSuite) SetUpSuite(c *C) {
	f, err := NewResponseFallback(Response{StatusCode: 400, Body: []byte("Come back later")})
	c.Assert(err, IsNil)
	fallbackResponse = f

	rdr, err := NewRedirectFallback(Redirect{URL: "http://localhost:5000"})
	c.Assert(err, IsNil)
	fallbackRedirect = rdr
}

func (s *CBSuite) advanceTime(d time.Duration) {
	s.clock.CurrentTime = s.clock.CurrentTime.Add(d)
}

func (s *CBSuite) TestStandbyCycle(c *C) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello"))
	})

	cb, err := New(handler, triggerNetRatio)
	c.Assert(err, IsNil)

	srv := httptest.NewServer(cb)
	defer srv.Close()

	re, body, err := testutils.Get(srv.URL)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)
	c.Assert(string(body), Equals, "hello")
}

func (s *CBSuite) TestFullCycle(c *C) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello"))
	})

	cb, err := New(handler, triggerNetRatio, Clock(s.clock))
	c.Assert(err, IsNil)

	srv := httptest.NewServer(cb)
	defer srv.Close()

	re, _, err := testutils.Get(srv.URL)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)

	cb.metrics = statsNetErrors(0.6)
	s.advanceTime(defaultCheckPeriod + time.Millisecond)
	re, _, err = testutils.Get(srv.URL)
	c.Assert(err, IsNil)
	c.Assert(cb.state, Equals, cbState(stateTripped))

	// Some time has passed, but we are still in trpped state.
	s.advanceTime(9 * time.Second)
	re, _, err = testutils.Get(srv.URL)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusServiceUnavailable)
	c.Assert(cb.state, Equals, cbState(stateTripped))

	// We should be in recovering state by now
	s.advanceTime(time.Second*1 + time.Millisecond)
	re, _, err = testutils.Get(srv.URL)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusServiceUnavailable)
	c.Assert(cb.state, Equals, cbState(stateRecovering))

	// 5 seconds after we should be allowing some requests to pass
	s.advanceTime(5 * time.Second)
	allowed := 0
	for i := 0; i < 100; i++ {
		re, _, err = testutils.Get(srv.URL)
		if re.StatusCode == http.StatusOK && err == nil {
			allowed++
		}
	}
	c.Assert(allowed, Not(Equals), 0)

	// After some time, all is good and we should be in stand by mode again
	s.advanceTime(5*time.Second + time.Millisecond)
	re, _, err = testutils.Get(srv.URL)
	c.Assert(cb.state, Equals, cbState(stateStandby))
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusOK)
}

func (s *CBSuite) TestRedirect(c *C) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello"))
	})

	cb, err := New(handler, triggerNetRatio, Clock(s.clock), Fallback(fallbackRedirect))
	c.Assert(err, IsNil)

	srv := httptest.NewServer(cb)
	defer srv.Close()

	cb.metrics = statsNetErrors(0.6)
	re, _, err := testutils.Get(srv.URL)
	c.Assert(err, IsNil)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return fmt.Errorf("no redirects")
		},
	}

	re, err = client.Get(srv.URL)
	c.Assert(err, NotNil)
	c.Assert(re.StatusCode, Equals, http.StatusFound)
	c.Assert(re.Header.Get("Location"), Equals, "http://localhost:5000")
}

func (s *CBSuite) TestTriggerDuringRecovery(c *C) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello"))
	})

	cb, err := New(handler, triggerNetRatio, Clock(s.clock), CheckPeriod(time.Microsecond))
	c.Assert(err, IsNil)

	srv := httptest.NewServer(cb)
	defer srv.Close()

	cb.metrics = statsNetErrors(0.6)
	re, _, err := testutils.Get(srv.URL)
	c.Assert(err, IsNil)
	c.Assert(cb.state, Equals, cbState(stateTripped))

	// We should be in recovering state by now
	s.advanceTime(10*time.Second + time.Millisecond)
	re, _, err = testutils.Get(srv.URL)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusServiceUnavailable)
	c.Assert(cb.state, Equals, cbState(stateRecovering))

	// We have matched error condition during recovery state and are going back to tripped state
	s.advanceTime(5 * time.Second)
	cb.metrics = statsNetErrors(0.6)
	allowed := 0
	for i := 0; i < 100; i++ {
		re, _, err = testutils.Get(srv.URL)
		if re.StatusCode == http.StatusOK && err == nil {
			allowed++
		}
	}
	c.Assert(allowed, Not(Equals), 0)
	c.Assert(cb.state, Equals, cbState(stateTripped))
}

func (s *CBSuite) TestSideEffects(c *C) {
	srv1Chan := make(chan *http.Request, 1)
	var srv1Body []byte
	srv1 := testutils.NewHandler(func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		c.Assert(err, IsNil)
		srv1Body = b
		w.Write([]byte("srv1"))
		srv1Chan <- r
	})
	defer srv1.Close()

	srv2Chan := make(chan *http.Request, 1)
	srv2 := testutils.NewHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("srv2"))
		r.ParseForm()
		srv2Chan <- r
	})
	defer srv2.Close()

	onTripped, err := NewWebhookSideEffect(
		Webhook{
			URL:     fmt.Sprintf("%s/post.json", srv1.URL),
			Method:  "POST",
			Headers: map[string][]string{"Content-Type": []string{"application/json"}},
			Body:    []byte(`{"Key": ["val1", "val2"]}`),
		})
	c.Assert(err, IsNil)

	onStandby, err := NewWebhookSideEffect(
		Webhook{
			URL:    fmt.Sprintf("%s/post", srv2.URL),
			Method: "POST",
			Form:   map[string][]string{"key": []string{"val1", "val2"}},
		})
	c.Assert(err, IsNil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello"))
	})

	cb, err := New(handler, triggerNetRatio, Clock(s.clock), CheckPeriod(time.Microsecond), OnTripped(onTripped), OnStandby(onStandby))
	c.Assert(err, IsNil)

	srv := httptest.NewServer(cb)
	defer srv.Close()

	cb.metrics = statsNetErrors(0.6)

	_, _, err = testutils.Get(srv.URL)
	c.Assert(err, IsNil)
	c.Assert(cb.state, Equals, cbState(stateTripped))

	select {
	case req := <-srv1Chan:
		c.Assert(req.Method, Equals, "POST")
		c.Assert(req.URL.Path, Equals, "/post.json")
		c.Assert(string(srv1Body), Equals, `{"Key": ["val1", "val2"]}`)
		c.Assert(req.Header.Get("Content-Type"), Equals, "application/json")
	case <-time.After(time.Second):
		c.Error("timeout waiting for side effect to kick off")
	}

	// Transition to recovering state
	s.advanceTime(10*time.Second + time.Millisecond)
	cb.metrics = statsOK()
	testutils.Get(srv.URL)
	c.Assert(cb.state, Equals, cbState(stateRecovering))

	// Going back to standby
	s.advanceTime(10*time.Second + time.Millisecond)
	testutils.Get(srv.URL)
	c.Assert(cb.state, Equals, cbState(stateStandby))

	select {
	case req := <-srv2Chan:
		c.Assert(req.Method, Equals, "POST")
		c.Assert(req.URL.Path, Equals, "/post")
		c.Assert(req.Form, DeepEquals, url.Values{"key": []string{"val1", "val2"}})
	case <-time.After(time.Second):
		c.Error("timeout waiting for side effect to kick off")
	}
}

func statsOK() *memmetrics.RTMetrics {
	m, err := memmetrics.NewRTMetrics()
	if err != nil {
		panic(err)
	}
	return m
}

func statsNetErrors(threshold float64) *memmetrics.RTMetrics {
	m, err := memmetrics.NewRTMetrics()
	if err != nil {
		panic(err)
	}
	for i := 0; i < 100; i++ {
		if i < int(threshold*100) {
			m.Record(http.StatusGatewayTimeout, 0)
		} else {
			m.Record(http.StatusOK, 0)
		}
	}
	return m
}

func statsLatencyAtQuantile(quantile float64, value time.Duration) *memmetrics.RTMetrics {
	m, err := memmetrics.NewRTMetrics()
	if err != nil {
		panic(err)
	}
	m.Record(http.StatusOK, value)
	return m
}

func statsResponseCodes(codes ...statusCode) *memmetrics.RTMetrics {
	m, err := memmetrics.NewRTMetrics()
	if err != nil {
		panic(err)
	}
	for _, c := range codes {
		for i := int64(0); i < c.Count; i++ {
			m.Record(c.Code, 0)
		}
	}
	return m
}

type statusCode struct {
	Code  int
	Count int64
}
