# valkeyrie

[![GoDoc](https://godoc.org/github.com/abronan/valkeyrie?status.png)](https://godoc.org/github.com/abronan/valkeyrie)
[![Build Status](https://travis-ci.org/abronan/valkeyrie.svg?branch=master)](https://travis-ci.org/abronan/valkeyrie)
[![Coverage Status](https://coveralls.io/repos/abronan/valkeyrie/badge.svg)](https://coveralls.io/r/abronan/valkeyrie)
[![Go Report Card](https://goreportcard.com/badge/github.com/abronan/valkeyrie)](https://goreportcard.com/report/github.com/abronan/valkeyrie)

`valkeyrie` provides a `Go` native library to store metadata using Distributed Key/Value stores (or common databases).

The goal of `valkeyrie` is to abstract common store operations (Get/Put/List/etc.) for multiple distributed and/or local Key/Value store backends thus using the same self-contained codebase to manage them all.

This repository is a fork of the [docker/libkv](https://github.com/docker/libkv) project which includes many fixes/additional features and is maintained by an original project maintainer. This project is notably used by [containous/traefik](https://github.com/containous/traefik), [docker/swarm](https://github.com/docker/swarm) and [docker/libnetwork](https://github.com/docker/libnetwork).

As of now, `valkeyrie` offers support for `Consul`, `Etcd`, `Zookeeper`, `Redis` (**Distributed** store) and `BoltDB` (**Local** store).

## Usage

`valkeyrie` is meant to be used as an abstraction layer over existing distributed Key/Value stores. It is especially useful if you plan to support `consul`, `etcd` and `zookeeper` using the same codebase.

It is ideal if you plan for something written in Go that should support:

- A simple metadata storage, distributed or local
- A lightweight discovery service for your nodes
- A distributed lock mechanism

You can also easily implement a generic *Leader Election* algorithm on top of it (see the [docker/leadership](https://github.com/docker/leadership) repository).

You can find examples of usage for `valkeyrie` under in [docs/examples.go](https://github.com/abronan/valkeyrie/blob/master/docs/examples.md). Optionally you can also take a look at the `docker/swarm`, `docker/libnetwork` or `containous/traefik` repositories which are using `valkeyrie` for all the use cases listed above.

## Supported versions

`valkeyrie` supports:
- **Consul** versions >= `0.5.1` because it uses Sessions with `Delete` behavior for the use of `TTLs` (mimics zookeeper's Ephemeral node support), If you don't plan to use `TTLs`: you can use Consul version `0.4.0+`.
- **Etcd** versions >= `2.0` with **APIv2** (*deprecated*) and **APIv3** (*recommended*).
- **Zookeeper** versions >= `3.4.5`. Although this might work with previous version but this remains untested as of now.
- **Boltdb**, which shouldn't be subject to any version dependencies.
- **Redis** versions >= `3.2.6`. Although this might work with previous version but this remains untested as of now.

## Interface

A **storage backend** in `valkeyrie` should implement (fully or partially) these interfaces:

```go
type Store interface {
	Put(key string, value []byte, options *WriteOptions) error
	Get(key string, options *ReadOptions) (*KVPair, error)
	Delete(key string) error
	Exists(key string, options *ReadOptions) (bool, error)
	Watch(key string, stopCh <-chan struct{}, options *ReadOptions) (<-chan *KVPair, error)
	WatchTree(directory string, stopCh <-chan struct{}, options *ReadOptions) (<-chan []*KVPair, error)
	NewLock(key string, options *LockOptions) (Locker, error)
	List(directory string, options *ReadOptions) ([]*KVPair, error)
	DeleteTree(directory string) error
	AtomicPut(key string, value []byte, previous *KVPair, options *WriteOptions) (bool, *KVPair, error)
	AtomicDelete(key string, previous *KVPair) (bool, error)
	Close()
}

type Locker interface {
	Lock(stopChan chan struct{}) (<-chan struct{}, error)
	Unlock() error
}
```

## Compatibility matrix

Backend drivers in `valkeyrie` are generally divided between **local drivers** and **distributed drivers**. Distributed backends offer enhanced capabilities like `Watches` and/or distributed `Locks`.

Local drivers are usually used in complement to the distributed drivers to store informations that only needs to be available locally.

| Calls                 |   Consul   |  Etcd  |  Zookeeper  |    Redis   |  BoltDB  |
|-----------------------|:----------:|:------:|:-----------:|:----------:|:--------:|
| Put                   |     X      |   X    |      X      |      X     |    X     |
| Get                   |     X      |   X    |      X      |      X     |    X     |
| Delete                |     X      |   X    |      X      |      X     |    X     |
| Exists                |     X      |   X    |      X      |      X     |    X     |
| Watch                 |     X      |   X    |      X      |      X     |          |
| WatchTree             |     X      |   X    |      X      |      X     |          |
| NewLock (Lock/Unlock) |     X      |   X    |      X      |      X     |          |
| List                  |     X      |   X    |      X      |      X     |    X     |
| DeleteTree            |     X      |   X    |      X      |      X     |    X     |
| AtomicPut             |     X      |   X    |      X      |      X     |    X     |
| AtomicDelete          |     X      |   X    |      X      |      X     |    X     |
| Close                 |     X      |   X    |      X      |      X     |    X     |

## Limitations

Distributed Key/Value stores often have different concepts for managing and formatting keys and their associated values. Even though `valkeyrie` tries to abstract those stores aiming for some consistency, in some cases it can't be applied easily.

Please refer to the `docs/compatibility.md` to see what are the special cases for cross-backend compatibility.

Other than those special cases, you should expect the same experience for basic operations like `Get`/`Put`, etc.

Calls like `WatchTree` may return different events (or number of events) depending on the backend (for now, `Etcd` and `Consul` will likely return more events than `Zookeeper` that you should triage properly). Although you should be able to use it successfully to watch on events in an interchangeable way (see the **docker/leadership** repository or the **pkg/discovery/kv** package in **docker/docker**).

For `Redis` backend, it relies on [key space notification](https://redis.io/topics/notifications) to perform WatchXXX/Lock related features. Please read the doc before using this feature.

## TLS

Only `Consul` and `etcd` have support for TLS and you should build and provide your own `config.TLS` object to feed the client. Support is planned for `zookeeper` and `redis`.

## Contributing

Want to contribute to valkeyrie? Take a look at the [Contribution Guidelines](https://github.com/abronan/valkeyrie/blob/master/CONTRIBUTING.md).

## Maintainers

**Alexandre Beslic**

- [abronan.com](https://abronan.com)
- [@abronan](https://twitter.com/abronan)

## Copyright and license

Copyright © 2014-2016 Docker, Inc. All rights reserved, except as follows. Code is released under the Apache 2.0 license. The README.md file, and files in the "docs" folder are licensed under the Creative Commons Attribution 4.0 International License under the terms and conditions set forth in the file "LICENSE.docs". You may obtain a duplicate copy of the same license, titled CC-BY-SA-4.0, at http://creativecommons.org/licenses/by/4.0/.
