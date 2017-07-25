go-github tests
===============

This directory contains additional test suites beyond the unit tests already in
[../github](../github). Whereas the unit tests run very quickly (since they
don't make any network calls) and are run by Travis on every commit, the tests
in this directory are only run manually.

The test packages are:

integration
-----------

This will exercise the entire go-github library (or at least as much as is
practical) against the live GitHub API. These tests will verify that the
library is properly coded against the actual behavior of the API, and will
(hopefully) fail upon any incompatible change in the API.

Because these tests are running using live data, there is a much higher
probability of false positives in test failures due to network issues, test
data having been changed, etc.

These tests send real network traffic to the GitHub API and will exhaust the
default unregistered rate limit (60 requests per hour) very quickly.
Additionally, in order to test the methods that modify data, a real OAuth token
will need to be present. While the tests will try to be well-behaved in terms
of what data they modify, it is **strongly** recommended that these tests only
be run using a dedicated test account.

Run tests using:

    GITHUB_AUTH_TOKEN=XXX go test -v -tags=integration ./integration

Additionally there are a set of integration tests for the Authorizations API.
These tests require a GitHub user (username and password), and also that a
[GitHub Application](https://github.com/settings/applications/new) (with
attendant Client ID and Client Secret) be available. Then, to execute just the
Authorization tests:

    GITHUB_USERNAME='<GH_USERNAME>' GITHUB_PASSWORD='<GH_PASSWORD>' GITHUB_CLIENT_ID='<CLIENT_ID>' GITHUB_CLIENT_SECRET='<CLIENT_SECRET>' go test -v -tags=integration --run=Authorizations ./integration

If some or all of these environment variables are not available, certain of the
Authorization integration tests will be skipped.

fields
------

This will identify the fields being returned by the live GitHub API that are
not currently being mapped into the relevant Go data type. Sometimes fields
are deliberately not mapped, so the results of this tool should just be taken
as a hint.

This test sends real network traffic to the GitHub API and will exhaust the
default unregistered rate limit (60 requests per hour) very quickly.
Additionally, some data is only returned for authenticated API calls. Unlike
the integration tests above, these tests only read data, so it's less
imperitive that these be run using a dedicated test account (though you still
really should).

Run the fields tool using:

    GITHUB_AUTH_TOKEN=XXX go run ./fields/fields.go
