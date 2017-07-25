package route

import (
	"net/http"

	. "gopkg.in/check.v1"
)

type ParseSuite struct {
}

var _ = Suite(&ParseSuite{})

func (s *ParseSuite) TestParseAndMatchSuccess(c *C) {
	testCases := []struct {
		Expression string
		Url        string
		Method     string
		Host       string
		Headers    http.Header
	}{
		// Trie cases
		{
			`Path("/helloworld")`,
			`http://google.com/helloworld`,
			"GET",
			"localhost",
			nil,
		},
		{
			`Method("GET") && Path("/helloworld")`,
			`http://google.com/helloworld`,
			"GET",
			"localhost",
			nil,
		},
		{
			`Path("/hello/<world>")`,
			`http://google.com/hello/world`,
			"GET",
			"localhost",
			nil,
		},
		{
			`Method("POST") &&  Path("/helloworld%2F")`,
			`http://google.com/helloworld%2F`,
			"POST",
			"localhost",
			nil,
		},
		{
			`Method("POST") && Path("/helloworld%2F")`,
			`http://google.com/helloworld%2F?q=b`,
			"POST",
			"localhost",
			nil,
		},
		{
			`Method("POST") && Path("/helloworld/<name>")`,
			`http://google.com/helloworld/%2F`,
			"POST",
			"localhost",
			nil,
		},
		{
			`Path("/helloworld")`,
			`http://google.com/helloworld`,
			"GET",
			"localhost",
			nil,
		},
		{
			`Method("POST") && Path("/helloworld")`,
			`http://google.com/helloworld`,
			"POST",
			"localhost",
			nil,
		},
		{
			`Host("localhost") && Method("POST") && Path("/helloworld")`,
			`http://google.com/helloworld`,
			"POST",
			"localhost",
			nil,
		},
		{
			`Host("<subdomain>.localhost") && Method("POST") && Path("/helloworld")`,
			`http://google.com/helloworld`,
			"POST",
			"a.localhost",
			nil,
		},
		{
			`Host("<sub1>.<sub2>.localhost") && Method("POST") && Path("/helloworld")`,
			`http://google.com/helloworld`,
			"POST",
			"a.b.localhost",
			nil,
		},
		{
			`Host("<sub1>.<sub2>.localhost") && Method("POST") && Path("/helloworld")`,
			`http://google.com/helloworld`,
			"POST",
			"a.b.localhost",
			nil,
		},
		{
			`Header("Content-Type", "application/json")`,
			`http://google.com/helloworld`,
			"POST",
			"",
			map[string][]string{"Content-Type": []string{"application/json"}},
		},
		{
			`Header("Content-Type", "application/<string>")`,
			`http://google.com/helloworld`,
			"POST",
			"",
			map[string][]string{"Content-Type": []string{"application/json"}},
		},
		{
			`Host("<sub1>.<sub2>.localhost") && Method("POST") && Path("/helloworld") && Header("Content-Type", "application/<string>")`,
			`http://google.com/helloworld`,
			"POST",
			"a.b.localhost",
			map[string][]string{"Content-Type": []string{"application/json"}},
		},
		// Regexp cases
		{
			`PathRegexp("/helloworld")`,
			`http://google.com/helloworld`,
			"GET",
			"localhost",
			nil,
		},
		{
			`HostRegexp("[^\\.]+\\.localhost") && Method("POST") && PathRegexp("/hello.*")`,
			`http://google.com/helloworld`,
			"POST",
			"a.localhost",
			nil,
		},
		{
			`HostRegexp("[^\\.]+\\.localhost") && Method("POST") && PathRegexp("/hello.*") && HeaderRegexp("Content-Type", "application/.+")`,
			`http://google.com/helloworld`,
			"POST",
			"a.b.localhost",
			map[string][]string{"Content-Type": []string{"application/json"}},
		},
	}
	for _, tc := range testCases {
		comment := Commentf("%v", tc.Expression)
		result := &match{val: "ok"}
		p, err := parse(tc.Expression, result)
		c.Assert(err, IsNil)
		c.Assert(p, NotNil)

		req := makeReq(req{url: tc.Url})
		req.Method = tc.Method
		req.Host = tc.Host

		req.Header = tc.Headers

		out := p.match(req)
		c.Assert(out, NotNil, comment)
		c.Assert(out, Equals, result, comment)
	}
}

func (s *ParseSuite) TestParseFailures(c *C) {
	testCases := []string{
		`bad`,                             // unsupported identifier
		`bad expression`,                  // not a valid go expression
		`Path("/path") || Path("/path2")`, // unsupported operator
		`1 && 2`,                // unsupported statements
		`"standalone literal"`,  // standalone literal
		`UnknownFunction("hi")`, // unknown functin
		`Path(1)`,               // bad argument type
		`RegexpRoute(1)`,        // bad argument type
		`Path()`,                // no arguments
		`PathRegexp()`,          // no arguments
		`Path(Path("hello"))`,   // nested calls
		`Path("")`,              // bad trie expression
		`PathRegexp("[[[[")`,    // bad regular expression
	}

	for _, expr := range testCases {
		m, err := parse(expr, &match{val: "ok"})
		c.Assert(err, NotNil)
		c.Assert(m, IsNil)
	}
}
