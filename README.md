![Træfɪk](http://traefik.github.io/traefik.logo.svg  "Træfɪk")
___

[![Circle CI](https://circleci.com/gh/emilevauge/traefik/tree/master.png?circle-token)](https://circleci.com/gh/emilevauge/traefik)
[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](https://github.com/EmileVauge/traefik/blob/master/LICENSE.md)
[![Join the chat at https://traefik.herokuapp.com](https://img.shields.io/badge/style-register-green.svg?style=social&label=Slack)](https://traefik.herokuapp.com)


Træfɪk is a modern HTTP reverse proxy and load balancer made to deploy microservices with ease.
It supports several backends ([Docker :whale:](https://www.docker.com/), [Mesos/Marathon](https://mesosphere.github.io/marathon/), [Consul](https://www.consul.io/), [Etcd](https://coreos.com/etcd/), [Zookeeper](https://zookeeper.apache.org), [BoltDB](https://github.com/boltdb/bolt), Rest API, file...) to manage its configuration automatically and dynamically.


## Features

- No dependency hell, single binary made with go
- Simple json Rest API
- Simple TOML file configuration
- Multiple backends supported: Docker, Mesos/Marathon, Consul, Etcd, and more to come
- Watchers for backends, can listen change in backends to apply a new configuration automatically
- Hot-reloading of configuration. No need to restart the process
- Graceful shutdown http connections during hot-reloads
- Circuit breakers on backends
- Round Robin, rebalancer load-balancers
- Rest Metrics
- Tiny docker image included [![Image Layers](https://badge.imagelayers.io/emilevauge/traefik:latest.svg)](https://imagelayers.io/?images=emilevauge/traefik:latest 'Image Layers')
- SSL backends support
- SSL frontend support
- Clean AngularJS Web UI
- Websocket support

## Demo

Here is a demo of Træfɪk using Docker backend, showing a load-balancing between two servers, hot reloading of configuration, and graceful shutdown.

[![asciicast](https://asciinema.org/a/4tcyde7riou5vxulo6my3mtko.png)](https://asciinema.org/a/4tcyde7riou5vxulo6my3mtko)

## Web UI

You can access to a simple HTML frontend of Træfik.

![Web UI Providers](docs/img/web.frontend.png)
![Web UI Health](docs/img/traefik-health.png)

## Plumbing

- [Oxy](https://github.com/vulcand/oxy): an awsome proxy library made by Mailgun guys
- [Gorilla mux](https://github.com/gorilla/mux): famous request router
- [Negroni](https://github.com/codegangsta/negroni): web middlewares made simple
- [Manners](https://github.com/mailgun/manners): graceful shutdown of http.Handler servers

## Quick start

- The simple way: grab the latest binary from the [releases](https://github.com/emilevauge/traefik/releases) page and just run it with the [sample configuration file](https://raw.githubusercontent.com/EmileVauge/traefik/master/traefik.sample.toml):

```shell
./traefik traefik.toml
```

- Use the tiny Docker image:

```shell
docker run -d -p 8080:8080 -p 80:80 -v $PWD/traefik.toml:/traefik.toml emilevauge/traefik
```

- From sources:

```shell
git clone https://github.com/EmileVauge/traefik
```

## Documentation

You can find the complete documentation [here](docs/index.md).

## Benchmarks

Refer to the [benchmarks section](docs/index.md#benchmarks) in the documentation.

## Træfɪk here and there

These projects use Træfɪk internally. If your company uses Træfɪk, we would be glad to get your feedback :) Contact us on [![Join the chat at https://traefik.herokuapp.com](https://img.shields.io/badge/style-register-green.svg?style=social&label=Slack)](https://traefik.herokuapp.com)

- Project [Mantl](https://mantl.io/) from Cisco

![Web UI Providers](docs/img/mantl-logo.png)
> Mantl is a modern platform for rapidly deploying globally distributed services. A container orchestrator, docker, a network stack, something to pool your logs, something to monitor health, a sprinkle of service discovery and some automation.

- Project [Apollo](http://capgemini.github.io/devops/apollo/) from Cap Gemini

![Web UI Providers](docs/img/apollo-logo.png)
> Apollo is an open source project to aid with building and deploying IAAS and PAAS services. It is particularly geared towards managing containerized applications across multiple hosts, and big data type workloads. Apollo leverages other open source components to provide basic mechanisms for deployment, maintenance, and scaling of infrastructure and applications.

## Contributing

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
