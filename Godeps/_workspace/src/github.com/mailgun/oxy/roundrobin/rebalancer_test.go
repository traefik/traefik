package roundrobin

import (
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/mailgun/oxy/forward"
	"github.com/mailgun/oxy/testutils"
	"github.com/mailgun/oxy/utils"
	"github.com/mailgun/timetools"

	. "gopkg.in/check.v1"
)

type RBSuite struct {
	clock *timetools.FreezedTime
	log   utils.Logger
}

var _ = Suite(&RBSuite{
	clock: &timetools.FreezedTime{
		CurrentTime: time.Date(2012, 3, 4, 5, 6, 7, 0, time.UTC),
	},
})

func (s *RBSuite) SetUpSuite(c *C) {
	s.log = utils.NewFileLogger(os.Stdout, utils.INFO)
}

func (s *RBSuite) TestRebalancerNormalOperation(c *C) {
	a, b := testutils.NewResponder("a"), testutils.NewResponder("b")
	defer a.Close()
	defer b.Close()

	fwd, err := forward.New()
	c.Assert(err, IsNil)

	lb, err := New(fwd)
	c.Assert(err, IsNil)

	rb, err := NewRebalancer(lb)
	c.Assert(err, IsNil)

	rb.UpsertServer(testutils.ParseURI(a.URL))
	c.Assert(rb.Servers()[0].String(), Equals, a.URL)

	proxy := httptest.NewServer(rb)
	defer proxy.Close()

	c.Assert(seq(c, proxy.URL, 3), DeepEquals, []string{"a", "a", "a"})
}

func (s *RBSuite) TestRebalancerNoServers(c *C) {
	fwd, err := forward.New()
	c.Assert(err, IsNil)

	lb, err := New(fwd)
	c.Assert(err, IsNil)

	rb, err := NewRebalancer(lb)
	c.Assert(err, IsNil)

	proxy := httptest.NewServer(rb)
	defer proxy.Close()

	re, _, err := testutils.Get(proxy.URL)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusInternalServerError)
}

func (s *RBSuite) TestRebalancerRemoveServer(c *C) {
	a, b := testutils.NewResponder("a"), testutils.NewResponder("b")
	defer a.Close()
	defer b.Close()

	fwd, err := forward.New()
	c.Assert(err, IsNil)

	lb, err := New(fwd)
	c.Assert(err, IsNil)

	rb, err := NewRebalancer(lb)
	c.Assert(err, IsNil)

	rb.UpsertServer(testutils.ParseURI(a.URL))
	rb.UpsertServer(testutils.ParseURI(b.URL))

	proxy := httptest.NewServer(rb)
	defer proxy.Close()

	c.Assert(seq(c, proxy.URL, 3), DeepEquals, []string{"a", "b", "a"})

	c.Assert(rb.RemoveServer(testutils.ParseURI(a.URL)), IsNil)
	c.Assert(seq(c, proxy.URL, 3), DeepEquals, []string{"b", "b", "b"})
}

// Test scenario when one server goes down after what it recovers
func (s *RBSuite) TestRebalancerRecovery(c *C) {
	a, b := testutils.NewResponder("a"), testutils.NewResponder("b")
	defer a.Close()
	defer b.Close()

	fwd, err := forward.New()
	c.Assert(err, IsNil)

	lb, err := New(fwd)
	c.Assert(err, IsNil)

	newMeter := func() (Meter, error) {
		return &testMeter{}, nil
	}
	rb, err := NewRebalancer(lb, RebalancerMeter(newMeter), RebalancerClock(s.clock), RebalancerLogger(s.log))
	c.Assert(err, IsNil)

	rb.UpsertServer(testutils.ParseURI(a.URL))
	rb.UpsertServer(testutils.ParseURI(b.URL))
	rb.servers[0].meter.(*testMeter).rating = 0.3

	proxy := httptest.NewServer(rb)
	defer proxy.Close()

	for i := 0; i < 6; i += 1 {
		testutils.Get(proxy.URL)
		testutils.Get(proxy.URL)
		s.clock.CurrentTime = s.clock.CurrentTime.Add(rb.backoffDuration + time.Second)
	}

	c.Assert(rb.servers[0].curWeight, Equals, 1)
	c.Assert(rb.servers[1].curWeight, Equals, FSMMaxWeight)

	c.Assert(lb.servers[0].weight, Equals, 1)
	c.Assert(lb.servers[1].weight, Equals, FSMMaxWeight)

	// server a is now recovering, the weights should go back to the original state
	rb.servers[0].meter.(*testMeter).rating = 0

	for i := 0; i < 6; i += 1 {
		testutils.Get(proxy.URL)
		testutils.Get(proxy.URL)
		s.clock.CurrentTime = s.clock.CurrentTime.Add(rb.backoffDuration + time.Second)
	}

	c.Assert(rb.servers[0].curWeight, Equals, 1)
	c.Assert(rb.servers[1].curWeight, Equals, 1)

	// Make sure we have applied the weights to the inner load balancer
	c.Assert(lb.servers[0].weight, Equals, 1)
	c.Assert(lb.servers[1].weight, Equals, 1)
}

// Test scenario when increaing the weight on good endpoints made it worse
func (s *RBSuite) TestRebalancerCascading(c *C) {
	a, b, d := testutils.NewResponder("a"), testutils.NewResponder("b"), testutils.NewResponder("d")
	defer a.Close()
	defer b.Close()
	defer d.Close()

	fwd, err := forward.New()
	c.Assert(err, IsNil)

	lb, err := New(fwd)
	c.Assert(err, IsNil)

	newMeter := func() (Meter, error) {
		return &testMeter{}, nil
	}
	rb, err := NewRebalancer(lb, RebalancerMeter(newMeter), RebalancerClock(s.clock), RebalancerLogger(s.log))
	c.Assert(err, IsNil)

	rb.UpsertServer(testutils.ParseURI(a.URL))
	rb.UpsertServer(testutils.ParseURI(b.URL))
	rb.UpsertServer(testutils.ParseURI(d.URL))
	rb.servers[0].meter.(*testMeter).rating = 0.3

	proxy := httptest.NewServer(rb)
	defer proxy.Close()

	for i := 0; i < 6; i += 1 {
		testutils.Get(proxy.URL)
		testutils.Get(proxy.URL)
		s.clock.CurrentTime = s.clock.CurrentTime.Add(rb.backoffDuration + time.Second)
	}

	// We have increased the load, and the situation became worse as the other servers started failing
	c.Assert(rb.servers[0].curWeight, Equals, 1)
	c.Assert(rb.servers[1].curWeight, Equals, FSMMaxWeight)
	c.Assert(rb.servers[2].curWeight, Equals, FSMMaxWeight)

	// server a is now recovering, the weights should go back to the original state
	rb.servers[0].meter.(*testMeter).rating = 0.3
	rb.servers[1].meter.(*testMeter).rating = 0.2
	rb.servers[2].meter.(*testMeter).rating = 0.2

	for i := 0; i < 6; i += 1 {
		testutils.Get(proxy.URL)
		testutils.Get(proxy.URL)
		s.clock.CurrentTime = s.clock.CurrentTime.Add(rb.backoffDuration + time.Second)
	}

	// the algo reverted it back
	c.Assert(rb.servers[0].curWeight, Equals, 1)
	c.Assert(rb.servers[1].curWeight, Equals, 1)
	c.Assert(rb.servers[2].curWeight, Equals, 1)
}

// Test scenario when all servers started failing
func (s *RBSuite) TestRebalancerAllBad(c *C) {
	a, b, d := testutils.NewResponder("a"), testutils.NewResponder("b"), testutils.NewResponder("d")
	defer a.Close()
	defer b.Close()
	defer d.Close()

	fwd, err := forward.New()
	c.Assert(err, IsNil)

	lb, err := New(fwd)
	c.Assert(err, IsNil)

	newMeter := func() (Meter, error) {
		return &testMeter{}, nil
	}
	rb, err := NewRebalancer(lb, RebalancerMeter(newMeter), RebalancerClock(s.clock), RebalancerLogger(s.log))
	c.Assert(err, IsNil)

	rb.UpsertServer(testutils.ParseURI(a.URL))
	rb.UpsertServer(testutils.ParseURI(b.URL))
	rb.UpsertServer(testutils.ParseURI(d.URL))
	rb.servers[0].meter.(*testMeter).rating = 0.12
	rb.servers[1].meter.(*testMeter).rating = 0.13
	rb.servers[2].meter.(*testMeter).rating = 0.11

	proxy := httptest.NewServer(rb)
	defer proxy.Close()

	for i := 0; i < 6; i += 1 {
		testutils.Get(proxy.URL)
		testutils.Get(proxy.URL)
		s.clock.CurrentTime = s.clock.CurrentTime.Add(rb.backoffDuration + time.Second)
	}

	// load balancer does nothing
	c.Assert(rb.servers[0].curWeight, Equals, 1)
	c.Assert(rb.servers[1].curWeight, Equals, 1)
	c.Assert(rb.servers[2].curWeight, Equals, 1)
}

// Removing the server resets the state
func (s *RBSuite) TestRebalancerReset(c *C) {
	a, b, d := testutils.NewResponder("a"), testutils.NewResponder("b"), testutils.NewResponder("d")
	defer a.Close()
	defer b.Close()
	defer d.Close()

	fwd, err := forward.New()
	c.Assert(err, IsNil)

	lb, err := New(fwd)
	c.Assert(err, IsNil)

	newMeter := func() (Meter, error) {
		return &testMeter{}, nil
	}
	rb, err := NewRebalancer(lb, RebalancerMeter(newMeter), RebalancerClock(s.clock), RebalancerLogger(s.log))
	c.Assert(err, IsNil)

	rb.UpsertServer(testutils.ParseURI(a.URL))
	rb.UpsertServer(testutils.ParseURI(b.URL))
	rb.UpsertServer(testutils.ParseURI(d.URL))
	rb.servers[0].meter.(*testMeter).rating = 0.3
	rb.servers[1].meter.(*testMeter).rating = 0
	rb.servers[2].meter.(*testMeter).rating = 0

	proxy := httptest.NewServer(rb)
	defer proxy.Close()

	for i := 0; i < 6; i += 1 {
		testutils.Get(proxy.URL)
		testutils.Get(proxy.URL)
		s.clock.CurrentTime = s.clock.CurrentTime.Add(rb.backoffDuration + time.Second)
	}

	// load balancer changed weights
	c.Assert(rb.servers[0].curWeight, Equals, 1)
	c.Assert(rb.servers[1].curWeight, Equals, FSMMaxWeight)
	c.Assert(rb.servers[2].curWeight, Equals, FSMMaxWeight)

	// Removing servers has reset the state
	rb.RemoveServer(testutils.ParseURI(d.URL))

	c.Assert(rb.servers[0].curWeight, Equals, 1)
	c.Assert(rb.servers[1].curWeight, Equals, 1)
}

func (s *RBSuite) TestRebalancerLive(c *C) {
	a, b := testutils.NewResponder("a"), testutils.NewResponder("b")
	defer a.Close()
	defer b.Close()

	fwd, err := forward.New()
	c.Assert(err, IsNil)

	lb, err := New(fwd)
	c.Assert(err, IsNil)

	rb, err := NewRebalancer(lb, RebalancerBackoff(time.Millisecond), RebalancerClock(s.clock))
	c.Assert(err, IsNil)

	rb.UpsertServer(testutils.ParseURI(a.URL))
	rb.UpsertServer(testutils.ParseURI(b.URL))
	rb.UpsertServer(testutils.ParseURI("http://localhost:62345"))

	proxy := httptest.NewServer(rb)
	defer proxy.Close()

	for i := 0; i < 1000; i += 1 {
		testutils.Get(proxy.URL)
		if i%10 == 0 {
			s.clock.CurrentTime = s.clock.CurrentTime.Add(rb.backoffDuration + time.Second)
		}
	}

	// load balancer changed weights
	c.Assert(rb.servers[0].curWeight, Equals, FSMMaxWeight)
	c.Assert(rb.servers[1].curWeight, Equals, FSMMaxWeight)
	c.Assert(rb.servers[2].curWeight, Equals, 1)
}

type testMeter struct {
	rating   float64
	notReady bool
}

func (tm *testMeter) Rating() float64 {
	return tm.rating
}

func (tm *testMeter) Record(int, time.Duration) {
}

func (tm *testMeter) IsReady() bool {
	return !tm.notReady
}
