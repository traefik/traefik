package roundrobin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/testutils"
	"github.com/vulcand/oxy/utils"

	. "gopkg.in/check.v1"
)

func TestRR(t *testing.T) { TestingT(t) }

type RRSuite struct{}

var _ = Suite(&RRSuite{})

func (s *RRSuite) TestNoServers(c *C) {
	fwd, err := forward.New()
	c.Assert(err, IsNil)

	lb, err := New(fwd)
	c.Assert(err, IsNil)

	proxy := httptest.NewServer(lb)
	defer proxy.Close()

	re, _, err := testutils.Get(proxy.URL)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusInternalServerError)
}

func (s *RRSuite) TestRemoveBadServer(c *C) {
	lb, err := New(nil)
	c.Assert(err, IsNil)

	c.Assert(lb.RemoveServer(testutils.ParseURI("http://google.com")), NotNil)
}

func (s *RRSuite) TestCustomErrHandler(c *C) {
	errHandler := utils.ErrorHandlerFunc(func(w http.ResponseWriter, req *http.Request, err error) {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte(http.StatusText(http.StatusTeapot)))
	})

	fwd, err := forward.New()
	c.Assert(err, IsNil)

	lb, err := New(fwd, ErrorHandler(errHandler))
	c.Assert(err, IsNil)

	proxy := httptest.NewServer(lb)
	defer proxy.Close()

	re, _, err := testutils.Get(proxy.URL)
	c.Assert(err, IsNil)
	c.Assert(re.StatusCode, Equals, http.StatusTeapot)
}

func (s *RRSuite) TestOneServer(c *C) {
	a := testutils.NewResponder("a")
	defer a.Close()

	fwd, err := forward.New()
	c.Assert(err, IsNil)

	lb, err := New(fwd)
	c.Assert(err, IsNil)

	lb.UpsertServer(testutils.ParseURI(a.URL))

	proxy := httptest.NewServer(lb)
	defer proxy.Close()

	c.Assert(seq(c, proxy.URL, 3), DeepEquals, []string{"a", "a", "a"})
}

func (s *RRSuite) TestSimple(c *C) {
	a := testutils.NewResponder("a")
	defer a.Close()

	b := testutils.NewResponder("b")
	defer b.Close()

	fwd, err := forward.New()
	c.Assert(err, IsNil)

	lb, err := New(fwd)
	c.Assert(err, IsNil)

	lb.UpsertServer(testutils.ParseURI(a.URL))
	lb.UpsertServer(testutils.ParseURI(b.URL))

	proxy := httptest.NewServer(lb)
	defer proxy.Close()

	c.Assert(seq(c, proxy.URL, 3), DeepEquals, []string{"a", "b", "a"})
}

func (s *RRSuite) TestRemoveServer(c *C) {
	a := testutils.NewResponder("a")
	defer a.Close()

	b := testutils.NewResponder("b")
	defer b.Close()

	fwd, err := forward.New()
	c.Assert(err, IsNil)

	lb, err := New(fwd)
	c.Assert(err, IsNil)

	lb.UpsertServer(testutils.ParseURI(a.URL))
	lb.UpsertServer(testutils.ParseURI(b.URL))

	proxy := httptest.NewServer(lb)
	defer proxy.Close()

	c.Assert(seq(c, proxy.URL, 3), DeepEquals, []string{"a", "b", "a"})

	lb.RemoveServer(testutils.ParseURI(a.URL))

	c.Assert(seq(c, proxy.URL, 3), DeepEquals, []string{"b", "b", "b"})
}

func (s *RRSuite) TestUpsertSame(c *C) {
	a := testutils.NewResponder("a")
	defer a.Close()

	fwd, err := forward.New()
	c.Assert(err, IsNil)

	lb, err := New(fwd)
	c.Assert(err, IsNil)

	c.Assert(lb.UpsertServer(testutils.ParseURI(a.URL)), IsNil)
	c.Assert(lb.UpsertServer(testutils.ParseURI(a.URL)), IsNil)

	proxy := httptest.NewServer(lb)
	defer proxy.Close()

	c.Assert(seq(c, proxy.URL, 3), DeepEquals, []string{"a", "a", "a"})
}

func (s *RRSuite) TestUpsertWeight(c *C) {
	a := testutils.NewResponder("a")
	defer a.Close()

	b := testutils.NewResponder("b")
	defer b.Close()

	fwd, err := forward.New()
	c.Assert(err, IsNil)

	lb, err := New(fwd)
	c.Assert(err, IsNil)

	c.Assert(lb.UpsertServer(testutils.ParseURI(a.URL)), IsNil)
	c.Assert(lb.UpsertServer(testutils.ParseURI(b.URL)), IsNil)

	proxy := httptest.NewServer(lb)
	defer proxy.Close()

	c.Assert(seq(c, proxy.URL, 3), DeepEquals, []string{"a", "b", "a"})

	c.Assert(lb.UpsertServer(testutils.ParseURI(b.URL), Weight(3)), IsNil)

	c.Assert(seq(c, proxy.URL, 4), DeepEquals, []string{"b", "b", "a", "b"})
}

func (s *RRSuite) TestWeighted(c *C) {
	a := testutils.NewResponder("a")
	defer a.Close()

	b := testutils.NewResponder("b")
	defer b.Close()

	fwd, err := forward.New()
	c.Assert(err, IsNil)

	lb, err := New(fwd)
	c.Assert(err, IsNil)

	lb.UpsertServer(testutils.ParseURI(a.URL), Weight(3))
	lb.UpsertServer(testutils.ParseURI(b.URL), Weight(2))

	proxy := httptest.NewServer(lb)
	defer proxy.Close()

	c.Assert(seq(c, proxy.URL, 6), DeepEquals, []string{"a", "a", "b", "a", "b", "a"})

	w, ok := lb.ServerWeight(testutils.ParseURI(a.URL))
	c.Assert(w, Equals, 3)
	c.Assert(ok, Equals, true)

	w, ok = lb.ServerWeight(testutils.ParseURI(b.URL))
	c.Assert(w, Equals, 2)
	c.Assert(ok, Equals, true)

	w, ok = lb.ServerWeight(testutils.ParseURI("http://caramba:4000"))
	c.Assert(w, Equals, -1)
	c.Assert(ok, Equals, false)
}

func seq(c *C, url string, repeat int) []string {
	out := []string{}
	for i := 0; i < repeat; i++ {
		_, body, err := testutils.Get(url)
		c.Assert(err, IsNil)
		out = append(out, string(body))
	}
	return out
}
