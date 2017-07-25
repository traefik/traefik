# Mesos-DNS [![Circle CI](https://circleci.com/gh/mesosphere/mesos-dns.svg?style=svg)](https://circleci.com/gh/mesosphere/mesos-dns) [![velocity](http://velocity.mesosphere.com/service/velocity/buildStatus/icon?job=public-mesos-dns-master)](http://velocity.mesosphere.com/service/velocity/job/public-mesos-dns-master/) [![Coverage Status](https://coveralls.io/repos/mesosphere/mesos-dns/badge.svg?branch=master&service=github)](https://coveralls.io/github/mesosphere/mesos-dns?branch=master) [![GoDoc](https://godoc.org/github.com/mesosphere/mesos-dns?status.svg)](https://godoc.org/github.com/mesosphere/mesos-dns) [![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/mesosphere/mesos-dns?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)
Mesos-DNS enables [DNS](http://en.wikipedia.org/wiki/Domain_Name_System) based service discovery in [Apache Mesos](http://mesos.apache.org/) clusters.

![Architecture
Diagram](http://mesosphere.github.io/mesos-dns/img/architecture.png)

## Compatibility
`mesos-N` tags mark the start of support for a specific Mesos version while
maintaining backwards compatibility with the previous major version.

## Installing
The official distribution and installation channel is pre-compiled binaries available in [Github releases](https://github.com/mesosphere/mesos-dns/releases).

## Building
Building the **master** branch from source should always succeed but doesn't provide
the same stability and compatibility guarantees as releases.

All branches and pull requests are tested by [Circle-CI](https://circleci.com/gh/mesosphere/mesos-dns), which also
outputs artifacts for Mac OS X, Windows, and Linux via cross-compilation.

You will need [Go](https://golang.org/) *1.5* or later to build the project.
All dependencies are vendored using `Godeps`. You must first install it in order to build from source.

```shell
$ go get github.com/tools/godep
$ godep go build ./...
```

### Building for release
#### To do a build:
1. Cut a branch
2. Tag it with the relevant version, and push the tags along with the branch
3. If the build doesn't trigger automatically, go here: https://circleci.com/gh/mesosphere/mesos-dns, find your branch, and trigger the build.

#### If you choose to do a private build:
1. Fork the repo on Github to a private repo
2. Customize that repo
3. Add it to Circle-CI

    Circle-CI allows for private repositories to be kept, and built in private
4. Go to the build steps.

#### Releasing
1. Download the artifacts from the Circle-CI builds
2. Cut a release based on the tag on Github
3. Upload the artifacts back to Github. Ensure you upload all the artifacts, including the .asc files.

#### Code signing
This repo using code signing. There is an armored, encrypted gpg key in the repo at build/private.key. This file includes the Mesos-DNS gpg signing key. The passphrase for the key is stored in Circle-CI's environment. This makes it fairly difficult to leak both components without detectable maliciousness.

There are only very few users with access to the private key, and they also have access to a revocation certificate in case the private key leaks.


## Testing
```shell
$ godep go test -race ./...
```

## Documentation
Detailed documentation on how to configure, operate and use Mesos-DNS
under different scenarios and environments is available in http://mesosphere.github.io/mesos-dns/.

## Contributing
Contributions are welcome. Please refer to [CONTRIBUTING.md](CONTRIBUTING.md) for
guidelines.

## Contact
For any discussion that isn't well suited for Github [issues](https://github.com/mesosphere/mesos-dns/issues),
please use our [mailing list](https://groups.google.com/forum/#!forum/mesos-dns) or our public [chat room](https://gitter.im/mesosphere/mesos-dns).

## License
This project is [Apache License 2.0](LICENSE).
