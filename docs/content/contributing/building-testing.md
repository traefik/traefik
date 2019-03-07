# Building and Testing

Compile and Test Your Own Traefik!
{: .subtitle }

So you want to build your own Traefik binary from the sources?
Let's see how.

## Building

You need either [Docker](https://github.com/docker/docker) and `make` (Method 1), or `go` (Method 2) in order to build Traefik.
For changes to its dependencies, the `dep` dependency management tool is required.

### Method 1: Using `Docker` and `Makefile`

Run make with the `binary` target.
This will create binaries for the Linux platform in the `dist` folder.

```bash
$ make binary
docker build -t traefik-webui -f webui/Dockerfile webui
Sending build context to Docker daemon  2.686MB
Step 1/11 : FROM node:8.15.0
 ---> 1f6c34f7921c
[...]
Successfully built ce4ff439c06a
Successfully tagged traefik-webui:latest
[...]
docker build  -t "traefik-dev:4475--feature-documentation" -f build.Dockerfile .
Sending build context to Docker daemon    279MB
Step 1/10 : FROM golang:1.12-alpine
 ---> f4bfb3d22bda
[...]
Successfully built 5c3c1a911277
Successfully tagged traefik-dev:4475--feature-documentation
docker run  -e "TEST_CONTAINER=1" -v "/var/run/docker.sock:/var/run/docker.sock" -it -e OS_ARCH_ARG -e OS_PLATFORM_ARG -e TESTFLAGS -e VERBOSE -e VERSION -e CODENAME -e TESTDIRS -e CI -e CONTAINER=DOCKER		 -v "/home/ldez/sources/go/src/github.com/containous/traefik/"dist":/go/src/github.com/containous/traefik/"dist"" "traefik-dev:4475--feature-documentation" ./script/make.sh generate binary
---> Making bundle: generate (in .)
removed 'autogen/gentemplates/gen.go'
removed 'autogen/genstatic/gen.go'

---> Making bundle: binary (in .)

$ ls dist/
traefik*
```

### Method 2: Using `go`

You need `go` v1.9+.

!!! tip "Source Directory"
 
    It is recommended that you clone Traefik into the `~/go/src/github.com/containous/traefik` directory.
    This is the official golang workspace hierarchy that will allow dependencies to be properly resolved.

!!! note "Environment"

    Set your `GOPATH` and `PATH` variable to be set to `~/go` via:
    
    ```bash
    export GOPATH=~/go
    export PATH=$PATH:$GOPATH/bin
    ```
 
    For convenience, add `GOPATH` and `PATH` to your `.bashrc` or `.bash_profile`
    
    Verify your environment is setup properly by running `$ go env`.
    Depending on your OS and environment, you should see an output similar to:
    
    ```bash
    GOARCH="amd64"
    GOBIN=""
    GOEXE=""
    GOHOSTARCH="amd64"
    GOHOSTOS="linux"
    GOOS="linux"
    GOPATH="/home/<yourusername>/go"
    GORACE=""
    ## ... and the list goes on
    ```

#### Build Traefik

Once you've set up your go environment and cloned the source repository, you can build Traefik.
Beforehand, you need to get `go-bindata` (the first time) in order to be able to use the `go generate` command (which is part of the build process).

```bash
cd ~/go/src/github.com/containous/traefik

# Get go-bindata. (Important: the ellipses are required.)
go get github.com/containous/go-bindata/...

# Let's build

# generate
# (required to merge non-code components into the final binary, such as the web dashboard and the provider's templates)
go generate

# Standard go build
go build ./cmd/traefik
```

You will find the Traefik executable (`traefik`) in the `~/go/src/github.com/containous/traefik` directory.

### Updating the templates

If you happen to update the provider's templates (located in `/templates`), you must run `go generate` to update the `autogen` package.

### Setting up dependency management

The [dep](https://github.com/golang/dep) command is not required for building;
however, it is necessary if you need to update the dependencies (i.e., add, update, or remove third-party packages).

You need [dep](https://github.com/golang/dep) >= 0.5.0.

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
```

## Testing

### Method 1: `Docker` and `make`

Run unit tests using the `test-unit` target.
Run integration tests using the `test-integration` target.
Run all tests (unit and integration) using the `test` target.

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

For development purposes, you can specify which tests to run by using (only works the `test-integration` target):

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

### Method 2: `go`

Unit tests can be run from the cloned directory using `$ go test ./...` which should return `ok`, similar to:

```test
ok      _/home/user/go/src/github/containous/traefik    0.004s
```

Integration tests must be run from the `integration/` directory and require the `-integration` switch: `$ cd integration && go test -integration ./...`.
