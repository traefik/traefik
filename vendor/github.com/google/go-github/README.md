# go-github #

go-github is a Go client library for accessing the [GitHub API][].

**Documentation:** [![GoDoc](https://godoc.org/github.com/google/go-github/github?status.svg)](https://godoc.org/github.com/google/go-github/github)  
**Mailing List:** [go-github@googlegroups.com](https://groups.google.com/group/go-github)  
**Build Status:** [![Build Status](https://travis-ci.org/google/go-github.svg?branch=master)](https://travis-ci.org/google/go-github)  
**Test Coverage:** [![Test Coverage](https://coveralls.io/repos/google/go-github/badge.svg?branch=master)](https://coveralls.io/r/google/go-github?branch=master) ([gocov report](https://drone.io/github.com/google/go-github/files/coverage.html))

go-github requires Go version 1.7 or greater.

## Usage ##

```go
import "github.com/google/go-github/github"
```

Construct a new GitHub client, then use the various services on the client to
access different parts of the GitHub API. For example:

```go
client := github.NewClient(nil)

// list all organizations for user "willnorris"
orgs, _, err := client.Organizations.List(ctx, "willnorris", nil)
```

Some API methods have optional parameters that can be passed. For example:

```go
client := github.NewClient(nil)

// list public repositories for org "github"
opt := &github.RepositoryListByOrgOptions{Type: "public"}
repos, _, err := client.Repositories.ListByOrg(ctx, "github", opt)
```

The services of a client divide the API into logical chunks and correspond to
the structure of the GitHub API documentation at
https://developer.github.com/v3/.

### Authentication ###

The go-github library does not directly handle authentication. Instead, when
creating a new client, pass an `http.Client` that can handle authentication for
you. The easiest and recommended way to do this is using the [oauth2][]
library, but you can always use any other library that provides an
`http.Client`. If you have an OAuth2 access token (for example, a [personal
API token][]), you can use it with the oauth2 library using:

```go
import "golang.org/x/oauth2"

func main() {
  ctx := context.Background()
  ts := oauth2.StaticTokenSource(
    &oauth2.Token{AccessToken: "... your access token ..."},
  )
  tc := oauth2.NewClient(ctx, ts)

  client := github.NewClient(tc)

  // list all repositories for the authenticated user
  repos, _, err := client.Repositories.List(ctx, "", nil)
}
```

Note that when using an authenticated Client, all calls made by the client will
include the specified OAuth token. Therefore, authenticated clients should
almost never be shared between different users.

See the [oauth2 docs][] for complete instructions on using that library.

For API methods that require HTTP Basic Authentication, use the
[`BasicAuthTransport`](https://godoc.org/github.com/google/go-github/github#BasicAuthTransport).

GitHub Apps authentication can be provided by the [ghinstallation](https://github.com/bradleyfalzon/ghinstallation)
package.

```go
import "github.com/bradleyfalzon/ghinstallation"

func main() {
  // Wrap the shared transport for use with the integration ID 1 authenticating with installation ID 99.
  itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, 1, 99, "2016-10-19.private-key.pem")
  if err != nil {
    // Handle error.
  }

  // Use installation transport with client.
  client := github.NewClient(&http.Client{Transport: itr})

  // Use client...
}
```

### Rate Limiting ###

GitHub imposes a rate limit on all API clients. Unauthenticated clients are
limited to 60 requests per hour, while authenticated clients can make up to
5,000 requests per hour. To receive the higher rate limit when making calls
that are not issued on behalf of a user, use the
`UnauthenticatedRateLimitedTransport`.

The returned `Response.Rate` value contains the rate limit information
from the most recent API call. If a recent enough response isn't
available, you can use `RateLimits` to fetch the most up-to-date rate
limit data for the client.

To detect an API rate limit error, you can check if its type is `*github.RateLimitError`:

```go
repos, _, err := client.Repositories.List(ctx, "", nil)
if _, ok := err.(*github.RateLimitError); ok {
	log.Println("hit rate limit")
}
```

Learn more about GitHub rate limiting at
https://developer.github.com/v3/#rate-limiting.

### Accepted Status ###

Some endpoints may return a 202 Accepted status code, meaning that the
information required is not yet ready and was scheduled to be gathered on
the GitHub side. Methods known to behave like this are documented specifying
this behavior.

To detect this condition of error, you can check if its type is
`*github.AcceptedError`:

```go
stats, _, err := client.Repositories.ListContributorsStats(ctx, org, repo)
if _, ok := err.(*github.AcceptedError); ok {
	log.Println("scheduled on GitHub side")
}
```

### Conditional Requests ###

The GitHub API has good support for conditional requests which will help
prevent you from burning through your rate limit, as well as help speed up your
application. `go-github` does not handle conditional requests directly, but is
instead designed to work with a caching `http.Transport`. We recommend using
https://github.com/gregjones/httpcache for that.

Learn more about GitHub conditional requests at
https://developer.github.com/v3/#conditional-requests.

### Creating and Updating Resources ###

All structs for GitHub resources use pointer values for all non-repeated fields.
This allows distinguishing between unset fields and those set to a zero-value.
Helper functions have been provided to easily create these pointers for string,
bool, and int values. For example:

```go
// create a new private repository named "foo"
repo := &github.Repository{
	Name:    github.String("foo"),
	Private: github.Bool(true),
}
client.Repositories.Create(ctx, "", repo)
```

Users who have worked with protocol buffers should find this pattern familiar.

### Pagination ###

All requests for resource collections (repos, pull requests, issues, etc.)
support pagination. Pagination options are described in the
`github.ListOptions` struct and passed to the list methods directly or as an
embedded type of a more specific list options struct (for example
`github.PullRequestListOptions`). Pages information is available via the
`github.Response` struct.

```go
client := github.NewClient(nil)

opt := &github.RepositoryListByOrgOptions{
	ListOptions: github.ListOptions{PerPage: 10},
}
// get all pages of results
var allRepos []*github.Repository
for {
	repos, resp, err := client.Repositories.ListByOrg(ctx, "github", opt)
	if err != nil {
		return err
	}
	allRepos = append(allRepos, repos...)
	if resp.NextPage == 0 {
		break
	}
	opt.Page = resp.NextPage
}
```

For complete usage of go-github, see the full [package docs][].

[GitHub API]: https://developer.github.com/v3/
[oauth2]: https://github.com/golang/oauth2
[oauth2 docs]: https://godoc.org/golang.org/x/oauth2
[personal API token]: https://github.com/blog/1509-personal-api-tokens
[package docs]: https://godoc.org/github.com/google/go-github/github

### Integration Tests ###

You can run integration tests from the `tests` directory. See the integration tests [README](tests/README.md).
## Roadmap ##

This library is being initially developed for an internal application at
Google, so API methods will likely be implemented in the order that they are
needed by that application. You can track the status of implementation in
[this Google spreadsheet][roadmap]. Eventually, I would like to cover the entire
GitHub API, so contributions are of course [always welcome][contributing]. The
calling pattern is pretty well established, so adding new methods is relatively
straightforward.

[roadmap]: https://docs.google.com/spreadsheet/ccc?key=0ApoVX4GOiXr-dGNKN1pObFh6ek1DR2FKUjBNZ1FmaEE&usp=sharing
[contributing]: CONTRIBUTING.md


## Google App Engine ##

Go on App Engine Classic (which as of this writing uses Go 1.6) can not use
the `"context"` import and still relies on `"golang.org/x/net/context"`.
As a result, if you wish to continue to use `go-github` on App Engine Classic,
you will need to rewrite all the `"context"` imports using the following command:

    `gofmt -w -r '"context" -> "golang.org/x/net/context"' *.go`

See `with_appengine.go` for more details.

## License ##

This library is distributed under the BSD-style license found in the [LICENSE](./LICENSE)
file.
