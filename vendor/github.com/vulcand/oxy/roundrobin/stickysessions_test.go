package roundrobin

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/testutils"

	. "gopkg.in/check.v1"
)

func TestSS(t *testing.T) { TestingT(t) }

type SSSuite struct{}

var _ = Suite(&SSSuite{})

func (s *SSSuite) TestBasic(c *C) {
	a := testutils.NewResponder("a")
	b := testutils.NewResponder("b")

	defer a.Close()
	defer b.Close()

	fwd, err := forward.New()
	c.Assert(err, IsNil)

	sticky := NewStickySession("test")
	c.Assert(sticky, NotNil)

	lb, err := New(fwd, EnableStickySession(sticky))
	c.Assert(err, IsNil)

	lb.UpsertServer(testutils.ParseURI(a.URL))
	lb.UpsertServer(testutils.ParseURI(b.URL))

	proxy := httptest.NewServer(lb)
	defer proxy.Close()

	http_cli := &http.Client{}

	for i := 0; i < 10; i++ {
		req, err := http.NewRequest("GET", proxy.URL, nil)
		c.Assert(err, IsNil)
		req.AddCookie(&http.Cookie{Name: "test", Value: a.URL})

		resp, err := http_cli.Do(req)
		c.Assert(err, IsNil)

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		c.Assert(err, IsNil)
		c.Assert(string(body), Equals, "a")
	}
}

func (s *SSSuite) TestStickCookie(c *C) {
	a := testutils.NewResponder("a")
	b := testutils.NewResponder("b")

	defer a.Close()
	defer b.Close()

	fwd, err := forward.New()
	c.Assert(err, IsNil)

	sticky := NewStickySession("test")
	c.Assert(sticky, NotNil)

	lb, err := New(fwd, EnableStickySession(sticky))
	c.Assert(err, IsNil)

	lb.UpsertServer(testutils.ParseURI(a.URL))
	lb.UpsertServer(testutils.ParseURI(b.URL))

	proxy := httptest.NewServer(lb)
	defer proxy.Close()

	resp, err := http.Get(proxy.URL)
	c.Assert(err, IsNil)

	c_out := resp.Cookies()[0]
	c.Assert(c_out.Name, Equals, "test")
	c.Assert(c_out.Value, Equals, a.URL)
}

func (s *SSSuite) TestRemoveRespondingServer(c *C) {
	a := testutils.NewResponder("a")
	b := testutils.NewResponder("b")

	defer a.Close()
	defer b.Close()

	fwd, err := forward.New()
	c.Assert(err, IsNil)

	sticky := NewStickySession("test")
	c.Assert(sticky, NotNil)

	lb, err := New(fwd, EnableStickySession(sticky))
	c.Assert(err, IsNil)

	lb.UpsertServer(testutils.ParseURI(a.URL))
	lb.UpsertServer(testutils.ParseURI(b.URL))

	proxy := httptest.NewServer(lb)
	defer proxy.Close()

	http_cli := &http.Client{}

	for i := 0; i < 10; i++ {
		req, err := http.NewRequest("GET", proxy.URL, nil)
		c.Assert(err, IsNil)
		req.AddCookie(&http.Cookie{Name: "test", Value: a.URL})

		resp, err := http_cli.Do(req)
		c.Assert(err, IsNil)

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		c.Assert(err, IsNil)
		c.Assert(string(body), Equals, "a")
	}

	lb.RemoveServer(testutils.ParseURI(a.URL))

	// Now, use the organic cookie response in our next requests.
	req, err := http.NewRequest("GET", proxy.URL, nil)
	req.AddCookie(&http.Cookie{Name: "test", Value: a.URL})
	resp, err := http_cli.Do(req)
	c.Assert(err, IsNil)

	c.Assert(resp.Cookies()[0].Name, Equals, "test")
	c.Assert(resp.Cookies()[0].Value, Equals, b.URL)

	for i := 0; i < 10; i++ {
		req, err := http.NewRequest("GET", proxy.URL, nil)
		c.Assert(err, IsNil)

		resp, err := http_cli.Do(req)
		c.Assert(err, IsNil)

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		c.Assert(err, IsNil)
		c.Assert(string(body), Equals, "b")
	}
}

func (s *SSSuite) TestRemoveAllServers(c *C) {
	a := testutils.NewResponder("a")
	b := testutils.NewResponder("b")

	defer a.Close()
	defer b.Close()

	fwd, err := forward.New()
	c.Assert(err, IsNil)

	sticky := NewStickySession("test")
	c.Assert(sticky, NotNil)

	lb, err := New(fwd, EnableStickySession(sticky))
	c.Assert(err, IsNil)

	lb.UpsertServer(testutils.ParseURI(a.URL))
	lb.UpsertServer(testutils.ParseURI(b.URL))

	proxy := httptest.NewServer(lb)
	defer proxy.Close()

	http_cli := &http.Client{}

	for i := 0; i < 10; i++ {
		req, err := http.NewRequest("GET", proxy.URL, nil)
		c.Assert(err, IsNil)
		req.AddCookie(&http.Cookie{Name: "test", Value: a.URL})

		resp, err := http_cli.Do(req)
		c.Assert(err, IsNil)

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		c.Assert(err, IsNil)
		c.Assert(string(body), Equals, "a")
	}

	lb.RemoveServer(testutils.ParseURI(a.URL))
	lb.RemoveServer(testutils.ParseURI(b.URL))

	// Now, use the organic cookie response in our next requests.
	req, err := http.NewRequest("GET", proxy.URL, nil)
	req.AddCookie(&http.Cookie{Name: "test", Value: a.URL})
	resp, err := http_cli.Do(req)
	c.Assert(err, IsNil)
	c.Assert(resp.StatusCode, Equals, http.StatusInternalServerError)
}

func (s *SSSuite) TestBadCookieVal(c *C) {
	a := testutils.NewResponder("a")

	defer a.Close()

	fwd, err := forward.New()
	c.Assert(err, IsNil)

	sticky := NewStickySession("test")
	c.Assert(sticky, NotNil)

	lb, err := New(fwd, EnableStickySession(sticky))
	c.Assert(err, IsNil)

	lb.UpsertServer(testutils.ParseURI(a.URL))

	proxy := httptest.NewServer(lb)
	defer proxy.Close()

	http_cli := &http.Client{}

	req, err := http.NewRequest("GET", proxy.URL, nil)
	c.Assert(err, IsNil)
	req.AddCookie(&http.Cookie{Name: "test", Value: "This is a patently invalid url!  You can't parse it!  :-)"})

	resp, err := http_cli.Do(req)
	c.Assert(err, IsNil)

	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(string(body), Equals, "a")

	// Now, cycle off the good server to cause an error
	lb.RemoveServer(testutils.ParseURI(a.URL))

	http_cli = &http.Client{}

	req, err = http.NewRequest("GET", proxy.URL, nil)
	c.Assert(err, IsNil)
	req.AddCookie(&http.Cookie{Name: "test", Value: "This is a patently invalid url!  You can't parse it!  :-)"})

	resp, err = http_cli.Do(req)
	c.Assert(err, IsNil)

	body, err = ioutil.ReadAll(resp.Body)
	c.Assert(resp.StatusCode, Equals, http.StatusInternalServerError)
}
