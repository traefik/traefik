
<p align="center">
    <picture>
      <source media="(prefers-color-scheme: dark)" srcset="docs/content/assets/img/baqup.logo-dark.png">
      <source media="(prefers-color-scheme: light)" srcset="docs/content/assets/img/baqup.logo.png">
      <img alt="Baqup" title="Baqup" src="docs/content/assets/img/baqup.logo.png">
    </picture>
</p>

[![Docs](https://img.shields.io/badge/docs-current-brightgreen.svg)](https://doc.baqup.io/baqup)
[![Go Report Card](https://goreportcard.com/badge/baqup/baqup)](https://goreportcard.com/report/baqup/baqup)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/baqupio/baqup/blob/master/LICENSE.md)
[![Join the community support forum at https://community.baqup.io/](https://img.shields.io/badge/style-register-green.svg?style=social&label=Discourse)](https://community.baqup.io/)
[![Twitter](https://img.shields.io/twitter/follow/baqup.svg?style=social)](https://twitter.com/intent/follow?screen_name=baqup)

Baqup (pronounced _backup_) is a modern HTTP reverse proxy and load balancer that makes deploying microservices easy.
Baqup integrates with your existing infrastructure components ([Docker](https://www.docker.com/), [Swarm mode](https://docs.docker.com/engine/swarm/), [Kubernetes](https://kubernetes.io), [Consul](https://www.consul.io/), [Etcd](https://coreos.com/etcd/), [Rancher v2](https://rancher.com), [Amazon ECS](https://aws.amazon.com/ecs), ...) and configures itself automatically and dynamically.
Pointing Baqup at your orchestrator should be the _only_ configuration step you need.

---

. **[Overview](#overview)** .
**[Features](#features)** .
**[Supported backends](#supported-backends)** .
**[Quickstart](#quickstart)** .
**[Web UI](#web-ui)** .
**[Documentation](#documentation)** .

. **[Support](#support)** .
**[Release cycle](#release-cycle)** .
**[Contributing](#contributing)** .
**[Maintainers](#maintainers)** .
**[Credits](#credits)** .

---

:warning: When migrating to a new major version of Baqup, please refer to the [migration guide](https://doc.baqup.io/baqup/migrate/v2-to-v3/) to ensure a smooth transition and to be aware of any breaking changes.


## Overview

Imagine that you have deployed a bunch of microservices with the help of an orchestrator (like Swarm or Kubernetes) or a service registry (like etcd or consul).
Now you want users to access these microservices, and you need a reverse proxy.

Traditional reverse-proxies require that you configure _each_ route that will connect paths and subdomains to _each_ microservice. 
In an environment where you add, remove, kill, upgrade, or scale your services _many_ times a day, the task of keeping the routes up to date becomes tedious. 

**This is when Baqup can help you!**

Baqup listens to your service registry/orchestrator API and instantly generates the routes so your microservices are connected to the outside world -- without further intervention from your part. 

**Run Baqup and let it do the work for you!** 
_(But if you'd rather configure some of your routes manually, Baqup supports that too!)_

![Architecture](docs/content/assets/img/baqup-architecture.png)

## Features

- Continuously updates its configuration (No restarts!)
- Supports multiple load balancing algorithms
- Provides HTTPS to your microservices by leveraging [Let's Encrypt](https://letsencrypt.org) (wildcard certificates support)
- Circuit breakers, retry
- See the magic through its clean web UI
- WebSocket, HTTP/2, gRPC ready
- Provides metrics (Rest, Prometheus, Datadog, Statsd, InfluxDB 2.X)
- Keeps access logs (JSON, CLF)
- Fast
- Exposes a Rest API
- Packaged as a single binary file (made with :heart: with go) and available as an [official](https://hub.docker.com/r/_/baqup/) docker image

## Supported Backends

- [Docker](https://doc.baqup.io/baqup/providers/docker/) / [Swarm mode](https://doc.baqup.io/baqup/providers/docker/)
- [Kubernetes](https://doc.baqup.io/baqup/providers/kubernetes-crd/)
- [ECS](https://doc.baqup.io/baqup/providers/ecs/)
- [File](https://doc.baqup.io/baqup/providers/file/)

## Quickstart

To get your hands on Baqup, you can use the [5-Minute Quickstart](https://doc.baqup.io/baqup/getting-started/quick-start/) in our documentation (you will need Docker).

## Web UI

You can access the simple HTML frontend of Baqup.

![Web UI Providers](docs/content/assets/img/webui-dashboard.png)

## Documentation

You can find the complete documentation of Baqup v3 at [https://doc.baqup.io/baqup/](https://doc.baqup.io/baqup/).

## Support

To get community support, you can:

- join the Baqup community forum: [![Join the chat at https://community.baqup.io/](https://img.shields.io/badge/style-register-green.svg?style=social&label=Discourse)](https://community.baqup.io/)

If you need commercial support, please contact [Baqup.io](https://baqup.io) by mail: <mailto:support@baqup.io>.

## Download

- Grab the latest binary from the [releases](https://github.com/baqupio/baqup/releases) page and run it with the [sample configuration file](https://raw.githubusercontent.com/baqup/baqup/master/baqup.sample.toml):

```shell
./baqup --configFile=baqup.toml
```

- Or use the official tiny Docker image and run it with the [sample configuration file](https://raw.githubusercontent.com/baqup/baqup/master/baqup.sample.toml):

```shell
docker run -d -p 8080:8080 -p 80:80 -v $PWD/baqup.toml:/etc/baqup/baqup.toml baqup
```

- Or get the sources:

```shell
git clone https://github.com/baqupio/baqup
```

## Introductory Videos

You can find high level and deep dive videos on [videos.baqup.io](https://videos.baqup.io).

## Maintainers

We are strongly promoting a philosophy of openness and sharing, and firmly standing against the elitist closed approach. Being part of the core team should be accessible to anyone who is motivated and want to be part of that journey!
This [document](docs/content/contributing/maintainers-guidelines.md) describes how to be part of the [maintainers' team](docs/content/contributing/maintainers.md) as well as various responsibilities and guidelines for Baqup maintainers.
You can also find more information on our process to review pull requests and manage issues [in this document](https://github.com/baqup/contributors-guide/blob/master/issue_triage.md).

## Contributing

If you'd like to contribute to the project, refer to the [contributing documentation](CONTRIBUTING.md).

Please note that this project is released with a [Contributor Code of Conduct](CODE_OF_CONDUCT.md).
By participating in this project, you agree to abide by its terms.

## Release Cycle

- We usually release 3/4 new versions (e.g. 1.1.0, 1.2.0, 1.3.0) per year.
- Release Candidates are available before the release (e.g. 1.1.0-rc1, 1.1.0-rc2, 1.1.0-rc3, 1.1.0-rc4, before 1.1.0).
- Bug-fixes (e.g. 1.1.1, 1.1.2, 1.2.1, 1.2.3) are released as needed (no additional features are delivered in those versions, bug-fixes only).

Each version is supported until the next one is released (e.g. 1.1.x will be supported until 1.2.0 is out).

We use [Semantic Versioning](https://semver.org/).

## Mailing Lists

- General announcements, new releases: mail at news+subscribe@baqup.io or on [the online viewer](https://groups.google.com/a/baqup.io/forum/#!forum/news).
- Security announcements: mail at security+subscribe@baqup.io or on [the online viewer](https://groups.google.com/a/baqup.io/forum/#!forum/security).

## Credits

Kudos to [Peka](https://www.instagram.com/pierroks/) for his awesome work on the gopher's logo!.

The gopher's logo of Baqup is licensed under the Creative Commons 3.0 Attributions license.

The gopher's logo of Baqup was inspired by the gopher stickers made by [Takuya Ueda](https://twitter.com/tenntenn).
The original Go gopher was designed by [Renee French](https://reneefrench.blogspot.com/).
