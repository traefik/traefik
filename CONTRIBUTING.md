# Contributing

## Building

You need either [Docker](https://github.com/docker/docker) and `make` (Method 1), or `go` (Method 2) in order to build traefik. For changes to its dependencies, the `glide` dependency management tool and `glide-vc` plugin are required.

### Method 1: Using `Docker` and `Makefile`

You need to run the `binary` target. This will create binaries for Linux platform in the `dist` folder.

```bash
$ make binary
docker build -t "traefik-dev:no-more-godep-ever" -f build.Dockerfile .
Sending build context to Docker daemon 295.3 MB
Step 0 : FROM golang:1.9-alpine
 ---> 8c6473912976
Step 1 : RUN go get github.com/Masterminds/glide
[...]
docker run --rm  -v "/var/run/docker.sock:/var/run/docker.sock" -it -e OS_ARCH_ARG -e OS_PLATFORM_ARG -e TESTFLAGS -v "/home/emile/dev/go/src/github.com/containous/traefik/"dist":/go/src/github.com/containous/traefik/"dist"" "traefik-dev:no-more-godep-ever" ./script/make.sh generate binary
---> Making bundle: generate (in .)
removed 'gen.go'

---> Making bundle: binary (in .)

$ ls dist/
traefik*
```

### Method 2: Using `go`

##### Setting up your `go` environment

- You need `go` v1.9+
- It is recommended you clone Træfik into a directory like `~/go/src/github.com/containous/traefik` (This is the official golang workspace hierarchy, and will allow dependencies to resolve properly)
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

##### Build Træfik

Once your environment is set up and the Træfik repository cloned you can build Træfik. You need get `go-bindata` once to be able to use `go generate` command as part of the build.  The steps to build are:

```bash
cd ~/go/src/github.com/containous/traefik

# Get go-bindata. Please note, the ellipses are required
go get github.com/jteeuwen/go-bindata/... 

# Start build
go generate

# Standard go build
go build ./cmd/traefik
# run other commands like tests
```

You will find the Træfik executable in the `~/go/src/github.com/containous/traefik` folder as `traefik`. 

### Setting up `glide` and `glide-vc` for dependency management

- Glide is not required for building; however, it is necessary to modify dependencies (i.e., add, update, or remove third-party packages)
- Glide can be installed either via homebrew: `$ brew install glide` or via the official glide script: `$ curl https://glide.sh/get | sh`
- The glide plugin `glide-vc` must be installed from source: `go get github.com/sgotti/glide-vc`

If you want to add a dependency, use `$ glide get` to have glide put it into the vendor folder and update the glide manifest/lock files (`glide.yaml` and `glide.lock`, respectively). A following `glide-vc` run should be triggered to trim down the size of the vendor folder. The final result must be committed into VCS.

Care must be taken to choose the right arguments to `glide` when dealing with dependencies, or otherwise risk ending up with a broken build. For that reason, the helper script `script/glide.sh` encapsulates the gory details and conveniently calls `glide-vc` as well. Call it without parameters for basic usage instructions.

Here's a full example using glide to add a new dependency:

```bash
# install the new main dependency github.com/foo/bar and minimize vendor size
$ ./script/glide.sh get github.com/foo/bar
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
docker run --rm -it -e OS_ARCH_ARG -e OS_PLATFORM_ARG -e TESTFLAGS -v "/home/vincent/src/github/vdemeester/traefik/dist:/go/src/github.com/containous/traefik/dist" "traefik-dev:your-feature-branch" ./script/make.sh generate test-unit
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

- Tests can be run from the cloned directory, by `$ go test ./...` which should return `ok` similar to:
```
ok      _/home/vincent/src/github/vdemeester/traefik    0.004s
```

## Documentation

The [documentation site](http://docs.traefik.io/) is built with [mkdocs](http://mkdocs.org/)

First make sure you have python and pip installed

```shell
$ python --version
Python 2.7.2
$ pip --version
pip 1.5.2
```

Then install mkdocs with pip

```shell
$ pip install mkdocs
```

To test documentation locally run `mkdocs serve` in the root directory, this should start a server locally to preview your changes.

```shell
$ mkdocs serve
INFO    -  Building documentation...
WARNING -  Config value: 'theme'. Warning: The theme 'united' will be removed in an upcoming MkDocs release. See http://www.mkdocs.org/about/release-notes/ for more details
INFO    -  Cleaning site directory
[I 160505 22:31:24 server:281] Serving on http://127.0.0.1:8000
[I 160505 22:31:24 handlers:59] Start watching changes
[I 160505 22:31:24 handlers:61] Start detecting changes
```


## How to Write a Good Issue

Please keep in mind that the GitHub issue tracker is not intended as a general support forum, but for reporting bugs and feature requests.

For end-user related support questions, refer to one of the following:
- the Traefik community Slack channel: [![Join the chat at https://traefik.herokuapp.com](https://img.shields.io/badge/style-register-green.svg?style=social&label=Slack)](https://traefik.herokuapp.com) 
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
