package stapler

import (
	"testing"
	"time"

	"github.com/mailgun/timetools"
	"github.com/vulcand/vulcand/engine"
	"github.com/vulcand/vulcand/testutils"
	"golang.org/x/crypto/ocsp"
	. "gopkg.in/check.v1"
)

func TestStapler(t *testing.T) { TestingT(t) }

var _ = Suite(&StaplerSuite{})

type StaplerSuite struct {
	st    *stapler
	clock *timetools.FreezedTime
	re    *ocsp.Response
}

func (s *StaplerSuite) SetUpSuite(c *C) {
	s.re = testutils.OCSPResponse
}

func (s *StaplerSuite) SetUpTest(c *C) {
	s.clock = &timetools.FreezedTime{CurrentTime: s.re.ThisUpdate.Add(time.Hour)}
	s.st = New(Clock(s.clock)).(*stapler)
}

func (s *StaplerSuite) TearDownTest(c *C) {
	s.st.Close()
}

func (s *StaplerSuite) TestCRUD(c *C) {
	srv := testutils.NewOCSPResponder()
	defer srv.Close()

	h, err := engine.NewHost("localhost",
		engine.HostSettings{
			KeyPair: &engine.KeyPair{Key: testutils.LocalhostKey, Cert: testutils.LocalhostCertChain},
			OCSP:    engine.OCSPSettings{Enabled: true, Period: "1h", Responders: []string{srv.URL}, SkipSignatureCheck: true},
		})
	c.Assert(err, IsNil)

	re, err := s.st.StapleHost(h)
	c.Assert(err, IsNil)
	c.Assert(re, NotNil)

	c.Assert(re.Response.Status, Equals, ocsp.Good)

	// subsequent call will return cached response
	other, err := s.st.StapleHost(h)
	c.Assert(err, IsNil)
	c.Assert(re, NotNil)
	c.Assert(other, Equals, re)

	// delete host
	hk := engine.HostKey{Name: h.Name}
	s.st.DeleteHost(hk)
	c.Assert(len(s.st.v), Equals, 0)

	// second call succeeds
	s.st.DeleteHost(hk)
}

// Update of the settings re-initializes staple
func (s *StaplerSuite) TestUpdateSettings(c *C) {
	srv := testutils.NewOCSPResponder()
	defer srv.Close()

	h, err := engine.NewHost("localhost",
		engine.HostSettings{
			KeyPair: &engine.KeyPair{Key: testutils.LocalhostKey, Cert: testutils.LocalhostCertChain},
			OCSP:    engine.OCSPSettings{Enabled: true, Period: "1h", Responders: []string{srv.URL}, SkipSignatureCheck: true},
		})
	c.Assert(err, IsNil)

	re, err := s.st.StapleHost(h)
	c.Assert(err, IsNil)
	c.Assert(re, NotNil)

	c.Assert(re.Response.Status, Equals, ocsp.Good)

	id := s.st.v[h.Name].id

	h2, err := engine.NewHost("localhost",
		engine.HostSettings{
			KeyPair: &engine.KeyPair{Key: testutils.LocalhostKey, Cert: testutils.LocalhostCertChain},
			OCSP:    engine.OCSPSettings{Enabled: true, Period: "2h", Responders: []string{srv.URL}, SkipSignatureCheck: true},
		})
	c.Assert(err, IsNil)

	re2, err := s.st.StapleHost(h2)
	c.Assert(err, IsNil)
	c.Assert(re2, NotNil)
	c.Assert(re2.Response.Status, Equals, ocsp.Good)

	// the host stapler has been updated
	id2 := s.st.v[h.Name].id

	c.Assert(re2, Not(Equals), re)
	c.Assert(id2, Not(Equals), id)
}

// Periodic update updated the staple value, we got the notification
func (s *StaplerSuite) TestUpdateStapleResult(c *C) {
	srv := testutils.NewOCSPResponder()
	defer srv.Close()

	h, err := engine.NewHost("localhost",
		engine.HostSettings{
			KeyPair: &engine.KeyPair{Key: testutils.LocalhostKey, Cert: testutils.LocalhostCertChain},
			OCSP:    engine.OCSPSettings{Enabled: true, Period: "1h", Responders: []string{srv.URL}, SkipSignatureCheck: true},
		})
	c.Assert(err, IsNil)

	events := make(chan *StapleUpdated, 1)
	close := make(chan struct{})
	s.st.Subscribe(events, close)

	re, err := s.st.StapleHost(h)
	c.Assert(err, IsNil)
	c.Assert(re, NotNil)
	c.Assert(re.Response.Status, Equals, ocsp.Good)

	s.st.kickC <- true

	var update *StapleUpdated
	select {
	case update = <-events:
		c.Assert(update, NotNil)
		c.Assert(update.Staple.IsValid(), Equals, true)
	case <-time.After(100 * time.Millisecond):
		c.Fatalf("timeout waiting for update")
	}
}

// Make sure the staple host cleaned
func (s *StaplerSuite) TestStapleReplaced(c *C) {
	srv := testutils.NewOCSPResponder()
	defer srv.Close()

	s.st.beforeUpdateC = make(chan bool)

	h, err := engine.NewHost("localhost",
		engine.HostSettings{
			KeyPair: &engine.KeyPair{Key: testutils.LocalhostKey, Cert: testutils.LocalhostCertChain},
			OCSP:    engine.OCSPSettings{Enabled: true, Period: "1h", Responders: []string{srv.URL}, SkipSignatureCheck: true},
		})
	c.Assert(err, IsNil)

	re, err := s.st.StapleHost(h)
	c.Assert(err, IsNil)
	c.Assert(re, NotNil)
	c.Assert(re.Response.Status, Equals, ocsp.Good)

	beforeId := s.st.v[h.Name].id

	s.st.kickC <- true
	// first make sure the fan out channel is at the right place
	s.st.beforeUpdateC <- true
	// update settings so the staple will be replaced
	h2, err := engine.NewHost("localhost",
		engine.HostSettings{
			KeyPair: &engine.KeyPair{Key: testutils.LocalhostKey, Cert: testutils.LocalhostCertChain},
			OCSP:    engine.OCSPSettings{Enabled: true, Period: "10h", Responders: []string{srv.URL}, SkipSignatureCheck: true},
		})
	s.st.StapleHost(h2)
	afterId := s.st.v[h.Name].id
	// now allow it to proceed
	s.st.beforeUpdateC <- true
	c.Assert(afterId, Not(Equals), beforeId)
}

func (s *StaplerSuite) TestFanOutUnsubscribe(c *C) {
	srv := testutils.NewOCSPResponder()
	defer srv.Close()

	h, err := engine.NewHost("localhost",
		engine.HostSettings{
			KeyPair: &engine.KeyPair{Key: testutils.LocalhostKey, Cert: testutils.LocalhostCertChain},
			OCSP:    engine.OCSPSettings{Enabled: true, Period: "1h", Responders: []string{srv.URL}, SkipSignatureCheck: true},
		})
	c.Assert(err, IsNil)

	events := make(chan *StapleUpdated, 1)
	closeC := make(chan struct{})
	s.st.Subscribe(events, closeC)

	events2 := make(chan *StapleUpdated, 1)
	closeC2 := make(chan struct{})
	s.st.Subscribe(events2, closeC2)

	re, err := s.st.StapleHost(h)
	c.Assert(err, IsNil)
	c.Assert(re, NotNil)
	c.Assert(re.Response.Status, Equals, ocsp.Good)

	s.st.kickC <- true

	// both channels got the update
	for _, ch := range []chan *StapleUpdated{events, events2} {
		select {
		case update := <-ch:
			c.Assert(update, NotNil)
		case <-time.After(100 * time.Millisecond):
			c.Fatalf("timeout waiting for update")
		}
	}

	// unsubscribe first channel, second will still get the notification
	close(closeC)

	s.st.kickC <- true
	select {
	case update := <-events2:
		c.Assert(update, NotNil)
	case <-time.After(100 * time.Millisecond):
		c.Fatalf("timeout waiting for update")
	}
}

// Responder became unavailable after series of retries
func (s *StaplerSuite) TestResponderUnavailable(c *C) {
	// we are setting up discard channel that will be used by Stapler to notify about discard events
	s.st.discardC = make(chan bool)

	srv := testutils.NewOCSPResponder()

	h, err := engine.NewHost("localhost",
		engine.HostSettings{
			KeyPair: &engine.KeyPair{Key: testutils.LocalhostKey, Cert: testutils.LocalhostCertChain},
			OCSP:    engine.OCSPSettings{Enabled: true, Period: "1h", Responders: []string{srv.URL}, SkipSignatureCheck: true},
		})
	c.Assert(err, IsNil)

	events := make(chan *StapleUpdated, 1)
	close := make(chan struct{})
	s.st.Subscribe(events, close)

	re, err := s.st.StapleHost(h)
	c.Assert(err, IsNil)
	c.Assert(re, NotNil)
	c.Assert(re.Response.Status, Equals, ocsp.Good)

	srv.Close()

	// The first update did not generate any events, as the time for the next update has not passed
	s.clock.CurrentTime = s.re.NextUpdate.Add(-1 * time.Hour)
	s.st.kickC <- true

	// The server discarded the event because the server is unreachable
	select {
	case <-s.st.discardC:
	case <-time.After(100 * time.Millisecond):
		c.Fatalf("timeout waiting for discard")
	}

	s.clock.CurrentTime = s.re.NextUpdate.Add(time.Hour)
	s.st.kickC <- true

	var update *StapleUpdated
	select {
	case update = <-events:
		c.Assert(update, NotNil)
		c.Assert(update.Err, NotNil)
	case <-time.After(100 * time.Millisecond):
		c.Fatalf("timeout waiting for update")
	}
}

func (s *StaplerSuite) TestStopInFlightTimers(c *C) {
	srv := testutils.NewOCSPResponder()
	defer srv.Close()

	h, err := engine.NewHost("localhost",
		engine.HostSettings{
			KeyPair: &engine.KeyPair{Key: testutils.LocalhostKey, Cert: testutils.LocalhostCertChain},
			OCSP:    engine.OCSPSettings{Enabled: true, Period: "1h", Responders: []string{srv.URL}, SkipSignatureCheck: true},
		})
	c.Assert(err, IsNil)

	events := make(chan *StapleUpdated, 1)
	close := make(chan struct{})
	s.st.Subscribe(events, close)

	re, err := s.st.StapleHost(h)
	c.Assert(err, IsNil)
	c.Assert(re, NotNil)
	c.Assert(re.Response.Status, Equals, ocsp.Good)
}

func (s *StaplerSuite) TestStapleFailed(c *C) {
	srv := testutils.NewOCSPResponder()
	srv.Close()

	h, err := engine.NewHost("localhost",
		engine.HostSettings{
			KeyPair: &engine.KeyPair{Key: testutils.LocalhostKey, Cert: testutils.LocalhostCertChain},
			OCSP:    engine.OCSPSettings{Enabled: true, Period: "1h", Responders: []string{srv.URL}, SkipSignatureCheck: true},
		})
	c.Assert(err, IsNil)

	re, err := s.st.StapleHost(h)
	c.Assert(err, NotNil)
	c.Assert(re, IsNil)
}

func (s *StaplerSuite) TestBadArguments(c *C) {
	h, err := engine.NewHost("localhost", engine.HostSettings{})
	c.Assert(err, IsNil)

	re, err := s.st.StapleHost(h)
	c.Assert(err, NotNil)
	c.Assert(re, IsNil)
}
