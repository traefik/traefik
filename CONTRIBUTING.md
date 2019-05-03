# Contributing

## Building

You need either [Docker](https://github.com/docker/docker) and `make` (Method 1), or `go` (Method 2) in order to build Traefik.
For changes to its dependencies, the `dep` dependency management tool is required.

### Method 1: Using `Docker` and `Makefile`

You need to run the `binary` target. This will create binaries for Linux platform in the `dist` folder.

```bash
$ make binary
docker build -t "traefik-dev:no-more-godep-ever" -f build.Dockerfile .
Sending build context to Docker daemon 295.3 MB
Step 0 : FROM golang:1.11-alpine
 ---> 8c6473912976
Step 1 : RUN go get github.com/golang/dep/cmd/dep
[...]
docker run --rm  -v "/var/run/docker.sock:/var/run/docker.sock" -it -e OS_ARCH_ARG -e OS_PLATFORM_ARG -e TESTFLAGS -v "/home/user/go/src/github.com/containous/traefik/"dist":/go/src/github.com/containous/traefik/"dist"" "traefik-dev:no-more-godep-ever" ./script/make.sh generate binary
---> Making bundle: generate (in .)
removed 'gen.go'

---> Making bundle: binary (in .)

$ ls dist/
traefik*
```

### Method 2: Using `go`

##### Setting up your `go` environment

- You need `go` v1.9+
- It is recommended you clone Traefik into a directory like `~/go/src/github.com/containous/traefik` (This is the official golang workspace hierarchy, and will allow dependencies to resolve properly)
- Set your `GOPATH` and `PATH` variable to be set to `~/go` via:

```bash
export GOPATH=~/go
export PATH=$PATH:$GOPATH/bin
```

> Note: You will want to add those 2 export lines to your `.bashrc` or `.bash_profile`

- Verify your environment is setup properly by running `$ go env`.  Depending on your OS and environment you should see output similar to:

```bash
GOARCH="amd64"
GOBIN=""
GOEXE=""
GOHOSTARCH="amd64"
GOHOSTOS="linux"
GOOS="linux"
GOPATH="/home/<yourusername>/go"
GORACE=""
## more go env's will be listed
```

##### Build Traefik

Once your environment is set up and the Traefik repository cloned you can build Traefik. You need get `go-bindata` once to be able to use `go generate` command as part of the build.  The steps to build are:

```bash
cd ~/go/src/github.com/containous/traefik

# Get go-bindata. Please note, the ellipses are required
go get github.com/containous/go-bindata/...

# Start build

# generate
# (required to merge non-code components into the final binary, such as the web dashboard and provider's Go templates)
go generate

# Standard go build
go build ./cmd/traefik
# run other commands like tests
```

You will find the Traefik executable in the `~/go/src/github.com/containous/traefik` folder as `traefik`.

### Updating the templates

If you happen to update the provider templates (in `/templates`), you need to run `go generate` to update the `autogen` package.

### Setting up dependency management

[dep](https://github.com/golang/dep) is not required for building; however, it is necessary to modify dependencies (i.e., add, update, or remove third-party packages)

You need to use [dep](https://github.com/golang/dep) >= 0.5.0.

If you want to add a dependency, use `dep ensure -add` to have [dep](https://github.com/golang/dep) put it into the vendor folder and update the dep manifest/lock files (`Gopkg.toml` and `Gopkg.lock`, respectively).

A following `make dep-prune` run should be triggered to trim down the size of the vendor folder.
The final result must be committed into VCS.

Here's a full example using dep to add a new dependency:

```bash
# install the new main dependency github.com/foo/bar and minimize vendor size
$ dep ensure -add github.com/foo/bar
# generate (Only required to integrate other components such as web dashboard)
$ go generate
# Standard go build
$ go build ./cmd/traefik
# run other commands like tests
```

### Tests

#### Method 1: `Docker` and `make`

You can run unit tests using the `test-unit` target and the
integration test using the `test-integration` target.

```bash
$ make test-unit
docker build -t "traefik-dev:your-feature-branch" -f build.Dockerfile .
# […]
docker run --rm -it -e OS_ARCH_ARG -e OS_PLATFORM_ARG -e TESTFLAGS -v "/home/user/go/src/github/containous/traefik/dist:/go/src/github.com/containous/traefik/dist" "traefik-dev:your-feature-branch" ./script/make.sh generate test-unit
---> Making bundle: generate (in .)
removed 'gen.go'

---> Making bundle: test-unit (in .)
+ go test -cover -coverprofile=cover.out .
ok      github.com/containous/traefik   0.005s  coverage: 4.1% of statements

Test success
```

For development purposes, you can specify which tests to run by using:

```bash
# Run every tests in the MyTest suite
TESTFLAGS="-check.f MyTestSuite" make test-integration

# Run the test "MyTest" in the MyTest suite
TESTFLAGS="-check.f MyTestSuite.MyTest" make test-integration

# Run every tests starting with "My", in the MyTest suite
TESTFLAGS="-check.f MyTestSuite.My" make test-integration

# Run every tests ending with "Test", in the MyTest suite
TESTFLAGS="-check.f MyTestSuite.*Test" make test-integration
```

More: https://labix.org/gocheck

#### Method 2: `go`

Unit tests can be run from the cloned directory by `$ go test ./...` which should return `ok` similar to:

```
ok      _/home/user/go/src/github/containous/traefik    0.004s
```

Integration tests must be run from the `integration/` directory and require the `-integration` switch to be passed like this: `$ cd integration && go test -integration ./...`.

## Documentation

The [documentation site](https://docs.traefik.io/) is built with [mkdocs](https://mkdocs.org/)

### Building Documentation

#### Method 1: `Docker` and `make`

You can build the documentation and serve it locally with livereloading, using the `docs` target:

```bash
$ make docs
docker build -t traefik-docs -f docs.Dockerfile .
# […]
docker run  --rm -v /home/user/go/github/containous/traefik:/mkdocs -p 8000:8000 traefik-docs mkdocs serve
# […]
[I 170828 20:47:48 server:283] Serving on http://0.0.0.0:8000
[I 170828 20:47:48 handlers:60] Start watching changes
[I 170828 20:47:48 handlers:62] Start detecting changes
```

And go to [http://127.0.0.1:8000](http://127.0.0.1:8000).

If you only want to build the documentation without serving it locally, you can use the following command:

```bash
$ make docs-build
...
```

#### Method 2: `mkdocs`

First make sure you have python and pip installed

```bash
$ python --version
Python 2.7.2
$ pip --version
pip 1.5.2
```

Then install mkdocs with pip

```bash
pip install --user -r requirements.txt
```

To build documentation locally and serve it locally,
run `mkdocs serve` in the root directory,
this should start a server locally to preview your changes.

```bash
$ mkdocs serve
INFO    -  Building documentation...
INFO    -  Cleaning site directory
[I 160505 22:31:24 server:281] Serving on http://127.0.0.1:8000
[I 160505 22:31:24 handlers:59] Start watching changes
[I 160505 22:31:24 handlers:61] Start detecting changes
```

### Verify Documentation

You can verify that the documentation meets some expectations, as checking for dead links, html markup validity.

```bash
$ make docs-verify
docker build -t traefik-docs-verify ./script/docs-verify-docker-image ## Build Validator image
...
docker run --rm -v /home/travis/build/containous/traefik:/app traefik-docs-verify ## Check for dead links and w3c compliance
=== Checking HTML content...
Running ["HtmlCheck", "ImageCheck", "ScriptCheck", "LinkCheck"] on /app/site/basics/index.html on *.html...
```

If you recently changed the documentation, do not forget to clean it to have it rebuilt:

```bash
$ make docs-clean docs-verify
...
```

Please note that verification can be disabled by setting the environment variable `DOCS_VERIFY_SKIP` to `true`:

```shell
DOCS_VERIFY_SKIP=true make docs-verify
...
DOCS_LINT_SKIP is true: no linting done.
```

## How to Write a Good Issue

Please keep in mind that the GitHub issue tracker is not intended as a general support forum, but for reporting bugs and feature requests.

For end-user related support questions, refer to one of the following:
- the Traefik community Slack channel: [![Join the chat at https://slack.traefik.io](https://img.shields.io/badge/style-register-green.svg?style=social&label=Slack)](https://slack.traefik.io)
- [Stack Overflow](https://stackoverflow.com/questions/tagged/traefik) (using the `traefik` tag)

### Title

The title must be short and descriptive. (~60 characters)

### Description

- Respect the issue template as much as possible. [template](.github/ISSUE_TEMPLATE.md)
- If it's possible use the command `traefik bug`. See https://www.youtube.com/watch?v=Lyz62L8m93I.
- Explain the conditions which led you to write this issue: the context.
- The context should lead to something, an idea or a problem that you’re facing.
- Remain clear and concise.
- Format your messages to help the reader focus on what matters and understand the structure of your message, use [Markdown syntax](https://help.github.com/articles/github-flavored-markdown)


## How to Write a Good Pull Request

### Title

The title must be short and descriptive. (~60 characters)

### Description

- Respect the pull request template as much as possible. [template](.github/PULL_REQUEST_TEMPLATE.md)
- Explain the conditions which led you to write this PR: the context.
- The context should lead to something, an idea or a problem that you’re facing.
- Remain clear and concise.
- Format your messages to help the reader focus on what matters and understand the structure of your message, use [Markdown syntax](https://help.github.com/articles/github-flavored-markdown)

### Content

- Make it small.
- Do only one thing.
- Write useful descriptions and titles.
- Avoid re-formatting.
- Make sure the code builds.
- Make sure all tests pass.
- Add tests.
- Address review comments in terms of additional commits.
- Do not amend/squash existing ones unless the PR is trivial.
- If a PR involves changes to third-party dependencies, the commits pertaining to the vendor folder and the manifest/lock file(s) should be committed separated.


Read [10 tips for better pull requests](http://blog.ploeh.dk/2015/01/15/10-tips-for-better-pull-requests/).
