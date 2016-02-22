# Contributing

### Building

You need either [Docker](https://github.com/docker/docker) and `make`, or `go` and `glide` in order to build traefik.

#### Setting up your `go` environment

- You need `go` v1.5
- You need to set `export GO15VENDOREXPERIMENT=1` environment variable
- You need `go-bindata` to be able to use `go generate` command (needed to build) : `go get github.com/jteeuwen/go-bindata/...`.
- If you clone Træfɪk into something like `~/go/src/github.com/traefik`, your `GOPATH` variable will have to be set to `~/go`: export `GOPATH=~/go`.

#### Using `Docker` and `Makefile`

You need to run the `binary` target. This will create binaries for Linux platform in the `dist` folder.

```bash
$ make binary
docker build -t "traefik-dev:no-more-godep-ever" -f build.Dockerfile .
Sending build context to Docker daemon 295.3 MB
Step 0 : FROM golang:1.5
 ---> 8c6473912976
Step 1 : RUN go get github.com/Masterminds/glide
[...]
docker run --rm  -v "/var/run/docker.sock:/var/run/docker.sock" -it -e OS_ARCH_ARG -e OS_PLATFORM_ARG -e TESTFLAGS -v "/home/emile/dev/go/src/github.com/emilevauge/traefik/"dist":/go/src/github.com/emilevauge/traefik/"dist"" "traefik-dev:no-more-godep-ever" ./script/make.sh generate binary
---> Making bundle: generate (in .)
removed 'gen.go'

---> Making bundle: binary (in .)

$ ls dist/
traefik*
```

#### Using `glide`

The idea behind `glide` is the following :

- when checkout(ing) a project, **run `glide up --quick`** to install
  (`go get …`) the dependencies in the `GOPATH`.
- if you need another dependency, import and use it in
  the source, and **run `glide get github.com/Masterminds/cookoo`** to save it in
  `vendor` and add it to your `glide.yaml`.

```bash
$ glide up --quick
# generate
$ go generate
# Simple go build
$ go build
# Using gox to build multiple platform
$ gox "linux darwin" "386 amd64 arm" \
    -output="dist/traefik_{{.OS}}-{{.Arch}}"
# run other commands like tests
$ go test ./...
ok      _/home/vincent/src/github/vdemeester/traefik    0.004s
```

### Tests

You can run unit tests using the `test-unit` target and the
integration test using the `test-integration` target.

```bash
$ make test-unit
docker build -t "traefik-dev:your-feature-branch" -f build.Dockerfile .
# […]
docker run --rm -it -e OS_ARCH_ARG -e OS_PLATFORM_ARG -e TESTFLAGS -v "/home/vincent/src/github/vdemeester/traefik/dist:/go/src/github.com/emilevauge/traefik/dist" "traefik-dev:your-feature-branch" ./script/make.sh generate test-unit
---> Making bundle: generate (in .)
removed 'gen.go'

---> Making bundle: test-unit (in .)
+ go test -cover -coverprofile=cover.out .
ok      github.com/emilevauge/traefik   0.005s  coverage: 4.1% of statements

Test success
```
