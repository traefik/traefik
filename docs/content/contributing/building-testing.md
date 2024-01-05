---
title: "Traefik Building & Testing Documentation"
description: "Compile and test your own Traefik Proxy! Learn how to build your own Traefik binary from the sources, and read the technical documentation."
---

# Building and Testing

Compile and Test Your Own Traefik!
{: .subtitle }

You want to build your own Traefik binary from the sources?
Let's see how.

## Building

You need:
    - [Docker](https://github.com/docker/docker "Link to website of Docker") 
    - `make`
    - [Go](https://go.dev/ "Link to website of Go")

!!! tip "Source Directory"

    It is recommended that you clone Traefik into the `~/go/src/github.com/traefik/traefik` directory.
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

### Build Traefik

Once you've set up your go environment and cloned the source repository, you can build Traefik.

```bash
$ make binary
./script/make.sh generate binary
---> Making bundle: generate (in .)

---> Making bundle: binary (in .)

$ ls dist/
traefik*
```

You will find the Traefik executable (`traefik`) in the `./dist` directory.

## Testing

Run unit tests using the `test-unit` target.
Run integration tests using the `test-integration` target.
Run all tests (unit and integration) using the `test` target.

```bash
$ make test-unit
./script/make.sh generate test-unit
---> Making bundle: generate (in .)

---> Making bundle: test-unit (in .)
+ go test -cover -coverprofile=cover.out .
ok      github.com/traefik/traefik   0.005s  coverage: 4.1% of statements

Test success
```

For development purposes, you can specify which tests to run by using (only works the `test-integration` target):

```bash
# Run every tests in the MyTest suite
TESTFLAGS="-test.run TestAccessLogSuite" make test-integration

# Run the test "MyTest" in the MyTest suite
TESTFLAGS="-test.run TestAccessLogSuite -testify.m ^TestAccessLog$" make test-integration
```
