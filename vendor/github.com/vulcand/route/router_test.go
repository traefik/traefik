package route

import (
	"fmt"
	"net/http"
	"regexp"

	. "gopkg.in/check.v1"
)

type RouteSuite struct {
}

var _ = Suite(&RouteSuite{})

func (s *RouteSuite) TestEmptyOperationsSucceed(c *C) {
	r := New()

	c.Assert(r.GetRoute("bla"), IsNil)
	c.Assert(r.RemoveRoute("bla"), IsNil)

	l, err := r.Route(makeReq(req{url: "http://google.com/blabla"}))
	c.Assert(err, IsNil)
	c.Assert(l, IsNil)
}

func (s *RouteSuite) TestCRUD(c *C) {
	r := New()

	match := "m"
	rt := `Path("/r1")`
	c.Assert(r.AddRoute(rt, match), IsNil)
	c.Assert(r.GetRoute(rt), Equals, match)
	c.Assert(r.RemoveRoute(rt), IsNil)
	c.Assert(r.GetRoute(rt), IsNil)
}

func (s *RouteSuite) TestAddTwiceFails(c *C) {
	r := New()

	match := "m"
	rt := `Path("/r1")`
	c.Assert(r.AddRoute(rt, match), IsNil)
	c.Assert(r.AddRoute(rt, match), NotNil)

	// Make sure that error did not have side effects
	out, err := r.Route(makeReq(req{url: "http://google.com/r1"}))
	c.Assert(err, IsNil)
	c.Assert(out, Equals, match)
}

func (s *RouteSuite) TestBadExpression(c *C) {
	r := New()

	m := "m"
	c.Assert(r.AddRoute(`Path("/r1")`, m), IsNil)
	c.Assert(r.AddRoute(`blabla`, "other"), NotNil)

	// Make sure that error did not have side effects
	out, err := r.Route(makeReq(req{url: "http://google.com/r1"}))
	c.Assert(err, IsNil)
	c.Assert(out, Equals, m)
}

func (s *RouteSuite) TestUpsert(c *C) {
	r := New()

	m1, m2 := "m1", "m2"
	c.Assert(r.UpsertRoute(`Path("/r1")`, m1), IsNil)
	c.Assert(r.UpsertRoute(`Path("/r1")`, m2), IsNil)
	c.Assert(r.UpsertRoute(`Path"/r1")`, m2), NotNil)

	out, err := r.Route(makeReq(req{url: "http://google.com/r1"}))
	c.Assert(err, IsNil)
	c.Assert(out, Equals, m2)
}

func (s *RouteSuite) TestMatchCases(c *C) {
	tc := []struct {
		name     string
		routes   []route // routes to add
		tries    []try   // various requests and outcomes
		expected int     // expected compiled matchers
	}{
		{
			name: "Simple Trie Path Matching",
			routes: []route{
				route{`Path("/r1")`, "m1"},
				route{`Path("/r2")`, "m2"},
			},
			expected: 1,
			tries: []try{
				try{
					r:     req{url: "http://google.com/r1"},
					match: "m1",
				},
				try{
					r:     req{url: "http://google.com/r2"},
					match: "m2",
				},
				try{
					r: req{url: "http://google.com/r3"},
				},
			},
		},
		{
			name: "Simple Trie Path Matching",
			routes: []route{
				route{`Path("/r1")`, "m1"},
			},
			expected: 1,
			tries: []try{
				try{
					r: req{url: "http://google.com/r3"},
				},
			},
		},
		{
			name: "Regexp path matching",
			routes: []route{
				route{`PathRegexp("/r1")`, "m1"},
				route{`PathRegexp("/r2")`, "m2"},
			},
			expected: 2, // Note that router does not compress regular expressions
			tries: []try{
				try{
					r:     req{url: "http://google.com/r1"},
					match: "m1",
				},
				try{
					r:     req{url: "http://google.com/r2"},
					match: "m2",
				},
				try{
					r: req{url: "http://google.com/r3"},
				},
			},
		},
		{
			name: "Mixed matching with trie and regexp",
			routes: []route{
				route{`PathRegexp("/r1")`, "m1"},
				route{`Path("/r2")`, "m2"},
			},
			expected: 2, // Note that router does not compress regular expressions
			tries: []try{
				try{
					r:     req{url: "http://google.com/r1"},
					match: "m1",
				},
				try{
					r:     req{url: "http://google.com/r2"},
					match: "m2",
				},
				try{
					r: req{url: "http://google.com/r3"},
				},
			},
		},
		{
			name: "Make sure longest path matches",
			routes: []route{
				route{`Path("/r")`, "m1"},
				route{`Path("/r/hello")`, "m2"},
			},
			expected: 1,
			tries: []try{
				try{
					r:     req{url: "http://google.com/r/hello"},
					match: "m2",
				},
			},
		},
		{
			name: "Match by method and path",
			routes: []route{
				route{`Method("POST") && Path("/r1")`, "m1"},
				route{`Method("GET") && Path("/r1")`, "m2"},
			},
			expected: 1,
			tries: []try{
				try{
					r:     req{url: "http://google.com/r1", method: "POST"},
					match: "m1",
				},
				try{
					r:     req{url: "http://google.com/r1", method: "GET"},
					match: "m2",
				},
				try{
					r: req{url: "http://google.com/r1", method: "PUT"},
				},
			},
		},
		{
			name: "Match by method and path",
			routes: []route{
				route{`Method("GET") && Path("/v1")`, "m1"},
				route{`Method("GET") && Path("/v2")`, "m2"},
				route{`Method("GET") && Path("/v3")`, "m3"},
			},
			expected: 1,
			tries: []try{
				try{
					r:     req{url: "http://google.com/v1", method: "GET"},
					match: "m1",
				},
				try{
					r:     req{url: "http://google.com/v2", method: "GET"},
					match: "m2",
				},
				try{
					r:     req{url: "http://google.com/v3", method: "GET"},
					match: "m3",
				},
			},
		},
		{
			name: "Match by method, path and hostname, same method and path",
			routes: []route{
				route{`Host("h1") && Method("POST") && Path("/r1")`, "m1"},
				route{`Host("h2") && Method("POST") && Path("/r1")`, "m2"},
			},
			expected: 1,
			tries: []try{
				try{
					r:     req{url: "http://h1/r1", method: "POST", host: "h1"},
					match: "m1",
				},
				try{
					r:     req{url: "http://h2/r1", method: "POST", host: "h2"},
					match: "m2",
				},
				try{
					r: req{url: "http://h2/r1", method: "GET", host: "h2"},
				},
				try{
					r: req{url: "http://h2/r1", method: "GET"},
				},
			},
		},
		{
			name: "Match by method, path and hostname, same method and path",
			routes: []route{
				route{`Host("h1") && Method("POST") && Path("/r1")`, "m1"},
				route{`Host("h2") && Method("GET") && Path("/r1")`, "m2"},
			},
			expected: 1,
			tries: []try{
				try{
					r:     req{url: "http://h1/r1", method: "POST", host: "h1"},
					match: "m1",
				},
				try{
					r:     req{url: "http://h2/r1", method: "GET", host: "h2"},
					match: "m2",
				},
				try{
					r: req{url: "http://h2/r1", method: "GET"},
				},
			},
		},
		{
			name: "Mixed match by method, path and hostname, same method and path",
			routes: []route{
				route{`Host("h1") && Method("POST") && Path("/r1")`, "m1"},
				route{`HostRegexp("h2") && Method("POST") && Path("/r1")`, "m2"},
			},
			expected: 2,
			tries: []try{
				try{
					r:     req{url: "http://h1/r1", method: "POST", host: "h1"},
					match: "m1",
				},
				try{
					r:     req{url: "http://h2/r1", method: "POST", host: "h2"},
					match: "m2",
				},
			},
		},
		{
			name: "Match by regexp method",
			routes: []route{
				route{`MethodRegexp("POST|PUT") && Path("/r1")`, "m1"},
				route{`MethodRegexp("GET") && Path("/r1")`, "m2"},
			},
			expected: 2,
			tries: []try{
				try{
					r:     req{url: "http://h1/r1", method: "POST"},
					match: "m1",
				},
				try{
					r:     req{url: "http://h1/r1", method: "PUT"},
					match: "m1",
				},
				try{
					r:     req{url: "http://h2/r1", method: "GET"},
					match: "m2",
				},
			},
		},
		{
			name: "Match by method, path and hostname and header",
			routes: []route{
				route{`Host("h1") && Method("POST") && Path("/r1")`, "m1"},
				route{`Host("h2") && Method("POST") && Path("/r1") && Header("Content-Type", "application/json")`, "m2"},
			},
			expected: 1,
			tries: []try{
				try{
					r:     req{url: "http://h1/r1", method: "POST", host: "h1"},
					match: "m1",
				},
				try{
					r:     req{url: "http://h2/r1", method: "POST", host: "h2", headers: http.Header{"Content-Type": []string{"application/json"}}},
					match: "m2",
				},
			},
		},
		{
			name: "Match by method, path and hostname and header for same hosts",
			routes: []route{
				route{`Host("h1") && Method("POST") && Path("/r1")`, "m1"},
				route{`Host("h1") && Method("POST") && Path("/r1") && Header("Content-Type", "application/json")`, "m2"},
			},
			expected: 1,
			tries: []try{
				try{
					r:     req{url: "http://h1/r1", method: "POST", host: "h1"},
					match: "m1",
				},
				try{
					r:     req{url: "http://h1/r1", method: "POST", host: "h1", headers: http.Header{"Content-Type": []string{"application/json"}}},
					match: "m2",
				},
				try{
					r:     req{url: "http://h1/r1", method: "POST", host: "h1", headers: http.Header{"Content-Type": []string{"text/plain"}}},
					match: "m1",
				},
			},
		},
		{
			name: "Catch all match for content-type",
			routes: []route{
				route{`Host("h1") && Method("POST") && Path("/r1") && Header("Content-Type", "<string>/<string>")`, "m1"},
				route{`Host("h1") && Method("POST") && Path("/r1") && Header("Content-Type", "application/json")`, "m2"},
			},
			expected: 1,
			tries: []try{
				try{
					r:     req{url: "http://h1/r1", method: "POST", host: "h1", headers: http.Header{"Content-Type": []string{"text/plain"}}},
					match: "m1",
				},
				try{
					r:     req{url: "http://h1/r1", method: "POST", host: "h1", headers: http.Header{"Content-Type": []string{"application/json"}}},
					match: "m2",
				},
			},
		},
		{
			name: "Match by method, path and hostname and header regexp",
			routes: []route{
				route{`Host("h1") && Method("POST") && Path("/r1")`, "m1"},
				route{`Host("h2") && Method("POST") && Path("/r1") && HeaderRegexp("Content-Type", "application/.*")`, "m2"},
			},
			expected: 2,
			tries: []try{
				try{
					r:     req{url: "http://h1/r1", method: "POST", host: "h1"},
					match: "m1",
				},
				try{
					r:     req{url: "http://h2/r1", method: "POST", host: "h2", headers: http.Header{"Content-Type": []string{"application/json"}}},
					match: "m2",
				},
			},
		},
		{
			name: "Make sure there is no match overlap",
			routes: []route{
				route{`Host("h1") && Method("POST") && Path("/r1")`, "m1"},
			},
			expected: 1,
			tries: []try{
				try{
					r: req{url: "http://h/r1", method: "1POST", host: "h"},
				},
			},
		},
	}
	for _, t := range tc {
		comment := Commentf("%v", t.name)
		r := New().(*router)
		for _, rt := range t.routes {
			c.Assert(r.AddRoute(rt.expr, rt.match), IsNil, comment)
		}
		if t.expected != 0 {
			c.Assert(len(r.matchers), Equals, t.expected, comment)
		}

		for _, a := range t.tries {
			req := makeReq(a.r)

			out, err := r.Route(req)
			c.Assert(err, IsNil)
			if a.match != "" {
				c.Assert(out, Equals, a.match, comment)
			} else {
				c.Assert(out, IsNil, comment)
			}
		}
	}
}

func (s *RouteSuite) TestGithubAPI(c *C) {
	r := New()

	re := regexp.MustCompile(":([^/]*)")
	for _, sp := range githubAPI {
		path := re.ReplaceAllString(sp.path, "<$1>")
		expr := fmt.Sprintf(`Method("%s") && Path("%s")`, sp.method, path)
		c.Assert(r.AddRoute(expr, expr), IsNil)
	}

	for _, sp := range githubAPI {
		path := re.ReplaceAllString(sp.path, "<$1>")
		expr := fmt.Sprintf(`Method("%s") && Path("%s")`, sp.method, path)
		out, err := r.Route(makeReq(req{method: sp.method, url: sp.path}))
		c.Assert(err, IsNil)
		c.Assert(out, Equals, expr)
	}
}

type route struct {
	expr  string
	match string
}

type try struct {
	r     req
	match string
}

type spec struct {
	method string
	path   string
}

var githubAPI = []spec{
	// OAuth Authorizations
	{"GET", "/authorizations"},
	{"GET", "/authorizations/:id"},
	{"POST", "/authorizations"},
	//{"PUT", "/authorizations/clients/:client_id"},
	//{"PATCH", "/authorizations/:id"},
	{"DELETE", "/authorizations/:id"},
	{"GET", "/applications/:client_id/tokens/:access_token"},
	{"DELETE", "/applications/:client_id/tokens"},
	{"DELETE", "/applications/:client_id/tokens/:access_token"},

	// Activity
	{"GET", "/events"},
	{"GET", "/repos/:owner/:repo/events"},
	{"GET", "/networks/:owner/:repo/events"},
	{"GET", "/orgs/:org/events"},
	{"GET", "/users/:user/received_events"},
	{"GET", "/users/:user/received_events/public"},
	{"GET", "/users/:user/events"},
	{"GET", "/users/:user/events/public"},
	{"GET", "/users/:user/events/orgs/:org"},
	{"GET", "/feeds"},
	{"GET", "/notifications"},
	{"GET", "/repos/:owner/:repo/notifications"},
	{"PUT", "/notifications"},
	{"PUT", "/repos/:owner/:repo/notifications"},
	{"GET", "/notifications/threads/:id"},
	//{"PATCH", "/notifications/threads/:id"},
	{"GET", "/notifications/threads/:id/subscription"},
	{"PUT", "/notifications/threads/:id/subscription"},
	{"DELETE", "/notifications/threads/:id/subscription"},
	{"GET", "/repos/:owner/:repo/stargazers"},
	{"GET", "/users/:user/starred"},
	{"GET", "/user/starred"},
	{"GET", "/user/starred/:owner/:repo"},
	{"PUT", "/user/starred/:owner/:repo"},
	{"DELETE", "/user/starred/:owner/:repo"},
	{"GET", "/repos/:owner/:repo/subscribers"},
	{"GET", "/users/:user/subscriptions"},
	{"GET", "/user/subscriptions"},
	{"GET", "/repos/:owner/:repo/subscription"},
	{"PUT", "/repos/:owner/:repo/subscription"},
	{"DELETE", "/repos/:owner/:repo/subscription"},
	{"GET", "/user/subscriptions/:owner/:repo"},
	{"PUT", "/user/subscriptions/:owner/:repo"},
	{"DELETE", "/user/subscriptions/:owner/:repo"},

	// Gists
	{"GET", "/users/:user/gists"},
	{"GET", "/gists"},
	//{"GET", "/gists/public"},
	//{"GET", "/gists/starred"},
	{"GET", "/gists/:id"},
	{"POST", "/gists"},
	//{"PATCH", "/gists/:id"},
	{"PUT", "/gists/:id/star"},
	{"DELETE", "/gists/:id/star"},
	{"GET", "/gists/:id/star"},
	{"POST", "/gists/:id/forks"},
	{"DELETE", "/gists/:id"},

	// Git Data
	{"GET", "/repos/:owner/:repo/git/blobs/:sha"},
	{"POST", "/repos/:owner/:repo/git/blobs"},
	{"GET", "/repos/:owner/:repo/git/commits/:sha"},
	{"POST", "/repos/:owner/:repo/git/commits"},
	//{"GET", "/repos/:owner/:repo/git/refs/*ref"},
	{"GET", "/repos/:owner/:repo/git/refs"},
	{"POST", "/repos/:owner/:repo/git/refs"},
	//{"PATCH", "/repos/:owner/:repo/git/refs/*ref"},
	//{"DELETE", "/repos/:owner/:repo/git/refs/*ref"},
	{"GET", "/repos/:owner/:repo/git/tags/:sha"},
	{"POST", "/repos/:owner/:repo/git/tags"},
	{"GET", "/repos/:owner/:repo/git/trees/:sha"},
	{"POST", "/repos/:owner/:repo/git/trees"},

	// Issues
	{"GET", "/issues"},
	{"GET", "/user/issues"},
	{"GET", "/orgs/:org/issues"},
	{"GET", "/repos/:owner/:repo/issues"},
	{"GET", "/repos/:owner/:repo/issues/:number"},
	{"POST", "/repos/:owner/:repo/issues"},
	//{"PATCH", "/repos/:owner/:repo/issues/:number"},
	{"GET", "/repos/:owner/:repo/assignees"},
	{"GET", "/repos/:owner/:repo/assignees/:assignee"},
	{"GET", "/repos/:owner/:repo/issues/:number/comments"},
	//{"GET", "/repos/:owner/:repo/issues/comments"},
	//{"GET", "/repos/:owner/:repo/issues/comments/:id"},
	{"POST", "/repos/:owner/:repo/issues/:number/comments"},
	//{"PATCH", "/repos/:owner/:repo/issues/comments/:id"},
	//{"DELETE", "/repos/:owner/:repo/issues/comments/:id"},
	{"GET", "/repos/:owner/:repo/issues/:number/events"},
	//{"GET", "/repos/:owner/:repo/issues/events"},
	//{"GET", "/repos/:owner/:repo/issues/events/:id"},
	{"GET", "/repos/:owner/:repo/labels"},
	{"GET", "/repos/:owner/:repo/labels/:name"},
	{"POST", "/repos/:owner/:repo/labels"},
	//{"PATCH", "/repos/:owner/:repo/labels/:name"},
	{"DELETE", "/repos/:owner/:repo/labels/:name"},
	{"GET", "/repos/:owner/:repo/issues/:number/labels"},
	{"POST", "/repos/:owner/:repo/issues/:number/labels"},
	{"DELETE", "/repos/:owner/:repo/issues/:number/labels/:name"},
	{"PUT", "/repos/:owner/:repo/issues/:number/labels"},
	{"DELETE", "/repos/:owner/:repo/issues/:number/labels"},
	{"GET", "/repos/:owner/:repo/milestones/:number/labels"},
	{"GET", "/repos/:owner/:repo/milestones"},
	{"GET", "/repos/:owner/:repo/milestones/:number"},
	{"POST", "/repos/:owner/:repo/milestones"},
	//{"PATCH", "/repos/:owner/:repo/milestones/:number"},
	{"DELETE", "/repos/:owner/:repo/milestones/:number"},

	// Miscellaneous
	{"GET", "/emojis"},
	{"GET", "/gitignore/templates"},
	{"GET", "/gitignore/templates/:name"},
	{"POST", "/markdown"},
	{"POST", "/markdown/raw"},
	{"GET", "/meta"},
	{"GET", "/rate_limit"},

	// Organizations
	{"GET", "/users/:user/orgs"},
	{"GET", "/user/orgs"},
	{"GET", "/orgs/:org"},
	//{"PATCH", "/orgs/:org"},
	{"GET", "/orgs/:org/members"},
	{"GET", "/orgs/:org/members/:user"},
	{"DELETE", "/orgs/:org/members/:user"},
	{"GET", "/orgs/:org/public_members"},
	{"GET", "/orgs/:org/public_members/:user"},
	{"PUT", "/orgs/:org/public_members/:user"},
	{"DELETE", "/orgs/:org/public_members/:user"},
	{"GET", "/orgs/:org/teams"},
	{"GET", "/teams/:id"},
	{"POST", "/orgs/:org/teams"},
	//{"PATCH", "/teams/:id"},
	{"DELETE", "/teams/:id"},
	{"GET", "/teams/:id/members"},
	{"GET", "/teams/:id/members/:user"},
	{"PUT", "/teams/:id/members/:user"},
	{"DELETE", "/teams/:id/members/:user"},
	{"GET", "/teams/:id/repos"},
	{"GET", "/teams/:id/repos/:owner/:repo"},
	{"PUT", "/teams/:id/repos/:owner/:repo"},
	{"DELETE", "/teams/:id/repos/:owner/:repo"},
	{"GET", "/user/teams"},

	// Pull Requests
	{"GET", "/repos/:owner/:repo/pulls"},
	{"GET", "/repos/:owner/:repo/pulls/:number"},
	{"POST", "/repos/:owner/:repo/pulls"},
	//{"PATCH", "/repos/:owner/:repo/pulls/:number"},
	{"GET", "/repos/:owner/:repo/pulls/:number/commits"},
	{"GET", "/repos/:owner/:repo/pulls/:number/files"},
	{"GET", "/repos/:owner/:repo/pulls/:number/merge"},
	{"PUT", "/repos/:owner/:repo/pulls/:number/merge"},
	{"GET", "/repos/:owner/:repo/pulls/:number/comments"},
	//{"GET", "/repos/:owner/:repo/pulls/comments"},
	//{"GET", "/repos/:owner/:repo/pulls/comments/:number"},
	{"PUT", "/repos/:owner/:repo/pulls/:number/comments"},
	//{"PATCH", "/repos/:owner/:repo/pulls/comments/:number"},
	//{"DELETE", "/repos/:owner/:repo/pulls/comments/:number"},

	// Repositories
	{"GET", "/user/repos"},
	{"GET", "/users/:user/repos"},
	{"GET", "/orgs/:org/repos"},
	{"GET", "/repositories"},
	{"POST", "/user/repos"},
	{"POST", "/orgs/:org/repos"},
	{"GET", "/repos/:owner/:repo"},
	//{"PATCH", "/repos/:owner/:repo"},
	{"GET", "/repos/:owner/:repo/contributors"},
	{"GET", "/repos/:owner/:repo/languages"},
	{"GET", "/repos/:owner/:repo/teams"},
	{"GET", "/repos/:owner/:repo/tags"},
	{"GET", "/repos/:owner/:repo/branches"},
	{"GET", "/repos/:owner/:repo/branches/:branch"},
	{"DELETE", "/repos/:owner/:repo"},
	{"GET", "/repos/:owner/:repo/collaborators"},
	{"GET", "/repos/:owner/:repo/collaborators/:user"},
	{"PUT", "/repos/:owner/:repo/collaborators/:user"},
	{"DELETE", "/repos/:owner/:repo/collaborators/:user"},
	{"GET", "/repos/:owner/:repo/comments"},
	{"GET", "/repos/:owner/:repo/commits/:sha/comments"},
	{"POST", "/repos/:owner/:repo/commits/:sha/comments"},
	{"GET", "/repos/:owner/:repo/comments/:id"},
	//{"PATCH", "/repos/:owner/:repo/comments/:id"},
	{"DELETE", "/repos/:owner/:repo/comments/:id"},
	{"GET", "/repos/:owner/:repo/commits"},
	{"GET", "/repos/:owner/:repo/commits/:sha"},
	{"GET", "/repos/:owner/:repo/readme"},
	//{"GET", "/repos/:owner/:repo/contents/*path"},
	//{"PUT", "/repos/:owner/:repo/contents/*path"},
	//{"DELETE", "/repos/:owner/:repo/contents/*path"},
	//{"GET", "/repos/:owner/:repo/:archive_format/:ref"},
	{"GET", "/repos/:owner/:repo/keys"},
	{"GET", "/repos/:owner/:repo/keys/:id"},
	{"POST", "/repos/:owner/:repo/keys"},
	//{"PATCH", "/repos/:owner/:repo/keys/:id"},
	{"DELETE", "/repos/:owner/:repo/keys/:id"},
	{"GET", "/repos/:owner/:repo/downloads"},
	{"GET", "/repos/:owner/:repo/downloads/:id"},
	{"DELETE", "/repos/:owner/:repo/downloads/:id"},
	{"GET", "/repos/:owner/:repo/forks"},
	{"POST", "/repos/:owner/:repo/forks"},
	{"GET", "/repos/:owner/:repo/hooks"},
	{"GET", "/repos/:owner/:repo/hooks/:id"},
	{"POST", "/repos/:owner/:repo/hooks"},
	//{"PATCH", "/repos/:owner/:repo/hooks/:id"},
	{"POST", "/repos/:owner/:repo/hooks/:id/tests"},
	{"DELETE", "/repos/:owner/:repo/hooks/:id"},
	{"POST", "/repos/:owner/:repo/merges"},
	{"GET", "/repos/:owner/:repo/releases"},
	{"GET", "/repos/:owner/:repo/releases/:id"},
	{"POST", "/repos/:owner/:repo/releases"},
	//{"PATCH", "/repos/:owner/:repo/releases/:id"},
	{"DELETE", "/repos/:owner/:repo/releases/:id"},
	{"GET", "/repos/:owner/:repo/releases/:id/assets"},
	{"GET", "/repos/:owner/:repo/stats/contributors"},
	{"GET", "/repos/:owner/:repo/stats/commit_activity"},
	{"GET", "/repos/:owner/:repo/stats/code_frequency"},
	{"GET", "/repos/:owner/:repo/stats/participation"},
	{"GET", "/repos/:owner/:repo/stats/punch_card"},
	{"GET", "/repos/:owner/:repo/statuses/:ref"},
	{"POST", "/repos/:owner/:repo/statuses/:ref"},

	// Search
	{"GET", "/search/repositories"},
	{"GET", "/search/code"},
	{"GET", "/search/issues"},
	{"GET", "/search/users"},
	{"GET", "/legacy/issues/search/:owner/:repository/:state/:keyword"},
	{"GET", "/legacy/repos/search/:keyword"},
	{"GET", "/legacy/user/search/:keyword"},
	{"GET", "/legacy/user/email/:email"},

	// Users
	{"GET", "/users/:user"},
	{"GET", "/user"},
	//{"PATCH", "/user"},
	{"GET", "/users"},
	{"GET", "/user/emails"},
	{"POST", "/user/emails"},
	{"DELETE", "/user/emails"},
	{"GET", "/users/:user/followers"},
	{"GET", "/user/followers"},
	{"GET", "/users/:user/following"},
	{"GET", "/user/following"},
	{"GET", "/user/following/:user"},
	{"GET", "/users/:user/following/:target_user"},
	{"PUT", "/user/following/:user"},
	{"DELETE", "/user/following/:user"},
	{"GET", "/users/:user/keys"},
	{"GET", "/user/keys"},
	{"GET", "/user/keys/:id"},
	{"POST", "/user/keys"},
	//{"PATCH", "/user/keys/:id"},
	{"DELETE", "/user/keys/:id"},
}
