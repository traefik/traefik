package route

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

func TestTrie(t *testing.T) { TestingT(t) }

type TrieSuite struct {
}

var _ = Suite(&TrieSuite{})

func (s *TrieSuite) TestParseTrieSuccess(c *C) {
	t, r := makeTrie(c, "/", &pathMapper{}, "val")
	c.Assert(t.match(makeReq(req{url: "http://google.com"})), DeepEquals, r)
}

func (s *TrieSuite) TestParseTrieFailures(c *C) {
	paths := []string{
		"",                       // empty path
		"/<uint8:hi>",            // unsupported matcher
		"/<string:hi:omg:hello>", // unsupported matcher parameters
	}
	for _, p := range paths {
		t, err := newTrieMatcher(p, &pathMapper{}, &match{val: "v1"})
		c.Assert(err, NotNil)
		c.Assert(t, IsNil)
	}
}

func (s *TrieSuite) testPathToTrie(c *C, p, trie string) {
	t, _ := makeTrie(c, p, &pathMapper{}, &match{val: "v"})
	c.Assert(printTrie(t), Equals, trie)
}

func (s *TrieSuite) TestPrintTries(c *C) {
	// Simple path
	s.testPathToTrie(c, "/a", `
root(0)
 node(0:/)
  match(0:a)
`)

	// Path wit default string parameter
	s.testPathToTrie(c, "/<param1>", `
root(0)
 node(0:/)
  match(0:<string:param1>)
`)

	// Path with trailing parameter
	s.testPathToTrie(c, "/m/<string:param1>", `
root(0)
 node(0:/)
  node(0:m)
   node(0:/)
    match(0:<string:param1>)
`)

	// Path with  parameter in the middle
	s.testPathToTrie(c, "/m/<string:param1>/a", `
root(0)
 node(0:/)
  node(0:m)
   node(0:/)
    node(0:<string:param1>)
     node(0:/)
      match(0:a)
`)

	// Path with two parameters
	s.testPathToTrie(c, "/m/<string:param1>/<string:param2>", `
root(0)
 node(0:/)
  node(0:m)
   node(0:/)
    node(0:<string:param1>)
     node(0:/)
      match(0:<string:param2>)
`)

}

func (s *TrieSuite) TestMergeTriesCommonPrefix(c *C) {
	t1, l1 := makeTrie(c, "/a", &pathMapper{}, &match{val: "v1"})
	t2, l2 := makeTrie(c, "/b", &pathMapper{}, &match{val: "v2"})

	t3, err := t1.merge(t2)
	c.Assert(err, IsNil)

	expected := `
root(0)
 node(0:/)
  match(0:a)
  match(0:b)
`
	c.Assert(printTrie(t3.(*trie)), Equals, expected)

	c.Assert(t3.match(makeReq(req{url: "http://google.com/a"})), Equals, l1)
	c.Assert(t3.match(makeReq(req{url: "http://google.com/b"})), Equals, l2)
}

func (s *TrieSuite) TestMergeTriesSubtree(c *C) {
	t1, l1 := makeTrie(c, "/aa", &pathMapper{}, &match{val: "v1"})
	t2, l2 := makeTrie(c, "/a", &pathMapper{}, &match{val: "v2"})

	t3, err := t1.merge(t2)
	c.Assert(err, IsNil)

	expected := `
root(0)
 node(0:/)
  match(0:a)
   match(0:a)
`
	c.Assert(printTrie(t3.(*trie)), Equals, expected)

	c.Assert(t3.match(makeReq(req{url: "http://google.com/aa"})), Equals, l1)
	c.Assert(t3.match(makeReq(req{url: "http://google.com/a"})), Equals, l2)
	c.Assert(t3.match(makeReq(req{url: "http://google.com/b"})), IsNil)
}

func (s *TrieSuite) TestMergeTriesWithCommonParameter(c *C) {
	t1, l1 := makeTrie(c, "/a/<string:name>/b", &pathMapper{}, &match{val: "v1"})
	t2, l2 := makeTrie(c, "/a/<string:name>/c", &pathMapper{}, &match{val: "v2"})

	t3, err := t1.merge(t2)
	c.Assert(err, IsNil)

	expected := `
root(0)
 node(0:/)
  node(0:a)
   node(0:/)
    node(0:<string:name>)
     node(0:/)
      match(0:b)
      match(0:c)
`
	c.Assert(printTrie(t3.(*trie)), Equals, expected)

	c.Assert(t3.match(makeReq(req{url: "http://google.com/a/bla/b"})), Equals, l1)
	c.Assert(t3.match(makeReq(req{url: "http://google.com/a/bla/c"})), Equals, l2)
	c.Assert(t3.match(makeReq(req{url: "http://google.com/a/"})), IsNil)
}

func (s *TrieSuite) TestMergeTriesWithDivergedParameter(c *C) {
	t1, l1 := makeTrie(c, "/a/<string:name1>/b", &pathMapper{}, &match{val: "v1"})
	t2, l2 := makeTrie(c, "/a/<string:name2>/c", &pathMapper{}, &match{val: "v2"})

	t3, err := t1.merge(t2)
	c.Assert(err, IsNil)

	expected := `
root(0)
 node(0:/)
  node(0:a)
   node(0:/)
    node(0:<string:name1>)
     node(0:/)
      match(0:b)
    node(0:<string:name2>)
     node(0:/)
      match(0:c)
`
	c.Assert(printTrie(t3.(*trie)), Equals, expected)

	c.Assert(t3.match(makeReq(req{url: "http://google.com/a/bla/b"})), Equals, l1)
	c.Assert(t3.match(makeReq(req{url: "http://google.com/a/bla/c"})), Equals, l2)
	c.Assert(t3.match(makeReq(req{url: "http://google.com/a/"})), IsNil)
}

func (s *TrieSuite) TestMergeTriesWithSamePath(c *C) {
	t1, l1 := makeTrie(c, "/a", &pathMapper{}, &match{val: "v1"})
	t2, _ := makeTrie(c, "/a", &pathMapper{}, &match{val: "v2"})

	t3, err := t1.merge(t2)
	c.Assert(err, IsNil)

	expected := `
root(0)
 node(0:/)
  match(0:a)
`
	c.Assert(printTrie(t3.(*trie)), Equals, expected)
	// The first location will match as it will always go first
	c.Assert(t3.match(makeReq(req{url: "http://google.com/a"})), Equals, l1)
}

func (s *TrieSuite) TestMergeAndMatchCases(c *C) {
	testCases := []struct {
		trees    []string
		url      string
		expected string
	}{
		// Matching /
		{
			[]string{"/"},
			"http://google.com/",
			"/",
		},
		// Matching / when there's no trailing / in url
		{
			[]string{"/"},
			"http://google.com",
			"/",
		},
		// Choosing longest path
		{
			[]string{"/v2/domains/", "/v2/domains/domain1"},
			"http://google.com/v2/domains/domain1",
			"/v2/domains/domain1",
		},
		// Named parameters
		{
			[]string{"/v1/domains/<string:name>", "/v2/domains/<string:name>"},
			"http://google.com/v2/domains/domain1",
			"/v2/domains/<string:name>",
		},
		// Int matcher, match
		{
			[]string{"/v<int:version>/domains/<string:name>"},
			"http://google.com/v42/domains/domain1",
			"/v<int:version>/domains/<string:name>",
		},
		// Int matcher, no match
		{
			[]string{"/v<int:version>/domains/<string:name>", "/<string:version>/domains/<string:name>"},
			"http://google.com/v42abc/domains/domain1",
			"/<string:version>/domains/<string:name>",
		},
		// Different combinations of named parameters
		{
			[]string{"/v1/domains/<domain>", "/v2/users/<user>/mailboxes/<mbx>"},
			"http://google.com/v2/users/u1/mailboxes/mbx1",
			"/v2/users/<user>/mailboxes/<mbx>",
		},
		// Something that looks like a pattern, but it's not
		{
			[]string{"/v1/<hello"},
			"http://google.com/v1/<hello",
			"/v1/<hello",
		},
	}
	for _, tc := range testCases {
		t, _ := makeTrie(c, tc.trees[0], &pathMapper{}, tc.trees[0])
		for i, pattern := range tc.trees {
			if i == 0 {
				continue
			}
			t2, _ := makeTrie(c, pattern, &pathMapper{}, pattern)
			out, err := t.merge(t2)
			c.Assert(err, IsNil)
			t = out.(*trie)
		}
		out := t.match(makeReq(req{url: tc.url}))
		c.Assert(out.val, Equals, tc.expected)
	}
}

func (s *TrieSuite) TestChainAndMatchCases(c *C) {
	tcs := []chainTc{
		chainTc{
			name: "Chain method and path",
			tries: []*trie{
				newTrie(c, "GET", &methodMapper{}, "v1"),
				newTrie(c, "/v1", &pathMapper{}, "v1"),
			},
			req:      makeReq(req{url: "http://localhost/v1", method: "GET"}),
			expected: "v1",
		},
		chainTc{
			name: "Chain hostname, method and path",
			tries: []*trie{
				newTrie(c, "h1", &hostMapper{}, "v0"),
				newTrie(c, "GET", &methodMapper{}, "v1"),
				newTrie(c, "/v1", &pathMapper{}, "v2"),
			},
			req:      makeReq(req{url: "http://localhost/v1", method: "GET", host: "h1"}),
			expected: "v2",
		},
	}
	for _, tc := range tcs {
		comment := Commentf("%v", tc.name)
		var out *trie
		for _, t := range tc.tries {
			if out == nil {
				out = t
				continue
			}
			m, err := out.chain(t)
			c.Assert(err, IsNil, comment)
			out = m.(*trie)
		}
		result := out.match(tc.req)
		c.Assert(result, NotNil, comment)
		c.Assert(result.val, Equals, tc.expected, comment)
	}
}

type chainTc struct {
	name     string
	tries    []*trie
	req      *http.Request
	expected string
}

func (s *TrieSuite) BenchmarkMatching(c *C) {
	rndString := NewRndString()

	t, _ := makeTrie(c, rndString.MakePath(20, 10), &pathMapper{}, "v")

	for i := 0; i < 10000; i++ {
		t2, _ := makeTrie(c, rndString.MakePath(20, 10), &pathMapper{}, "v")
		out, err := t.merge(t2)
		if err != nil {
			c.Assert(err, IsNil)
		}
		t = out.(*trie)
	}
	req := makeReq(req{url: fmt.Sprintf("http://google.com/%s", rndString.MakePath(20, 10))})
	for i := 0; i < c.N; i++ {
		t.match(req)
	}
}

func cutTrie(index int, expressions []string) []string {
	v := make([]string, 0, len(expressions)-1)
	v = append(v, expressions[:index]...)
	v = append(v, expressions[index+1:]...)
	return v
}

func makeTrie(c *C, expr string, mp requestMapper, val interface{}) (*trie, *match) {
	l := &match{
		val: val,
	}
	t, err := newTrieMatcher(expr, mp, l)
	c.Assert(err, IsNil)
	c.Assert(t, NotNil)
	return t, l
}

func newTrie(c *C, expr string, mp requestMapper, val interface{}) *trie {
	t, _ := makeTrie(c, expr, mp, val)
	return t
}

type req struct {
	url     string
	host    string
	headers http.Header
	method  string
}

func makeReq(rq req) *http.Request {
	ur, err := url.ParseRequestURI(rq.url)
	if err != nil {
		panic(err)
	}
	r := &http.Request{
		URL:        ur,
		RequestURI: rq.url,
		Host:       rq.host,
		Header:     rq.headers,
		Method:     rq.method,
	}
	return r
}

type RndString struct {
	src rand.Source
}

func NewRndString() *RndString {
	return &RndString{rand.NewSource(time.Now().UTC().UnixNano())}
}

func (r *RndString) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = byte(r.src.Int63()%26 + 97)
	}
	return len(p), nil
}

func (r *RndString) MakeString(n int) string {
	buffer := &bytes.Buffer{}
	io.CopyN(buffer, r, int64(n))
	return buffer.String()
}

func (s *RndString) MakePath(varlen, minlen int) string {
	return fmt.Sprintf("/%s", s.MakeString(rand.Intn(varlen)+minlen))
}
