# Protocol Buffers for Go with Gadgets

[![Build Status](https://travis-ci.org/gogo/protobuf.svg?branch=master)](https://travis-ci.org/gogo/protobuf)

gogoprotobuf is a fork of <a href="https://github.com/golang/protobuf">golang/protobuf</a> with extra code generation features.

This code generation is used to achieve:
  
  - fast marshalling and unmarshalling
  - more canonical Go structures
  - goprotobuf compatibility
  - less typing by optionally generating extra helper code
  - peace of mind by optionally generating test and benchmark code
  - other serialization formats

Keeping track of how up to date gogoprotobuf is relative to golang/protobuf is done in this
<a href="https://github.com/gogo/protobuf/issues/191">issue</a>

## Users

These projects use gogoprotobuf:

  - <a href="http://godoc.org/github.com/coreos/etcd">etcd</a> - <a href="https://blog.gopheracademy.com/advent-2015/etcd-distributed-key-value-store-with-grpc-http2/">blog</a>
  - <a href="https://www.spacemonkey.com/">spacemonkey</a> - <a href="https://www.spacemonkey.com/blog/posts/go-space-monkey">blog</a>
  - <a href="http://bazil.org">bazil</a>
  - <a href="http://badoo.com">badoo</a>
  - <a href="https://github.com/mesos/mesos-go">mesos-go</a>
  - <a href="https://github.com/mozilla-services/heka">heka</a>
  - <a href="https://github.com/cockroachdb/cockroach">cockroachdb</a>
  - <a href="https://github.com/jbenet/go-ipfs">go-ipfs</a>
  - <a href="https://github.com/philhofer/rkive">rkive-go</a>
  - <a href="https://www.dropbox.com">dropbox</a>
  - <a href="https://srclib.org/">srclib</a> - <a href="https://sourcegraph.com/sourcegraph.com/sourcegraph/srclib@97f54fed4f9a4bff0a28edf6eb7c0e013afc7bcd/.tree/graph/def.proto">sample proto file</a>
  - <a href="http://www.adyoulike.com/">adyoulike</a>
  - <a href="http://www.cloudfoundry.org/">cloudfoundry</a>
  - <a href="http://kubernetes.io/">kubernetes</a>
  - <a href="https://dgraph.io/">dgraph</a> - <a href="https://github.com/dgraph-io/dgraph/releases/tag/v0.4.3">release notes</a> - <a href="https://discuss.dgraph.io/t/gogoprotobuf-is-extremely-fast/639">benchmarks</a></a>
  - <a href="https://github.com/centrifugal/centrifugo">centrifugo</a> - <a href="https://forum.golangbridge.org/t/centrifugo-real-time-messaging-websocket-or-sockjs-server-v1-5-0-released/2861">release notes</a> - <a href="https://medium.com/@fzambia/centrifugo-protobuf-inside-json-outside-21d39bdabd68#.o3icmgjqd">blog</a>

Please lets us know if you are using gogoprotobuf by posting on our <a href="https://groups.google.com/forum/#!topic/gogoprotobuf/Brw76BxmFpQ">GoogleGroup</a>.

### Mentioned

  - <a href="http://www.slideshare.net/albertstrasheim/serialization-in-go">Cloudflare - go serialization talk - Albert Strasheim</a>
  - <a href="http://gophercon.sourcegraph.com/post/83747547505/writing-a-high-performance-database-in-go">gophercon</a>
  - <a href="https://github.com/alecthomas/go_serialization_benchmarks">alecthomas' go serialization benchmarks</a>

## Getting Started 

There are several ways to use gogoprotobuf, but for all you need to install go and protoc.
After that you can choose:
 
  - Speed
  - More Speed and more generated code
  - Most Speed and most customization

### Installation

To install it, you must first have Go (at least version 1.3.3) installed (see [http://golang.org/doc/install](http://golang.org/doc/install)).  Go 1.4.2, 1.5.4, 1.6.3 and 1.7 are continuously tested.

Next, install the standard protocol buffer implementation from [https://github.com/google/protobuf](https://github.com/google/protobuf).
Most versions from 2.3.1 should not give any problems, but 2.5.0, 2.6.1 and 3 are continuously tested.

### Speed

Install the protoc-gen-gofast binary

    go get github.com/gogo/protobuf/protoc-gen-gofast

Use it to generate faster marshaling and unmarshaling go code for your protocol buffers.

    protoc --gofast_out=. myproto.proto

This does not allow you to use any of the other gogoprotobuf [extensions](https://github.com/gogo/protobuf/blob/master/extensions.md).

### More Speed and more generated code

Fields without pointers cause less time in the garbage collector.
More code generation results in more convenient methods.

Other binaries are also included:

    protoc-gen-gogofast (same as gofast, but imports gogoprotobuf)
    protoc-gen-gogofaster (same as gogofast, without XXX_unrecognized, less pointer fields)
    protoc-gen-gogoslick (same as gogofaster, but with generated string, gostring and equal methods)

Installing any of these binaries is easy.  Simply run:

    go get github.com/gogo/protobuf/proto
    go get github.com/gogo/protobuf/{binary}
    go get github.com/gogo/protobuf/gogoproto

These binaries allow you to using gogoprotobuf [extensions](https://github.com/gogo/protobuf/blob/master/extensions.md).

### Most Speed and most customization

Customizing the fields of the messages to be the fields that you actually want to use removes the need to copy between the structs you use and structs you use to serialize.
gogoprotobuf also offers more serialization formats and generation of tests and even more methods.

Please visit the [extensions](https://github.com/gogo/protobuf/blob/master/extensions.md) page for more documentation.

Install protoc-gen-gogo:

    go get github.com/gogo/protobuf/proto
    go get github.com/gogo/protobuf/jsonpb
    go get github.com/gogo/protobuf/protoc-gen-gogo
    go get github.com/gogo/protobuf/gogoproto

## Proto3

Proto3 is supported, but the new well known types are not supported yet.
[See Proto3 Issue](https://github.com/gogo/protobuf/issues/57) for more details.

## GRPC

It works the same as golang/protobuf, simply specify the plugin.
Here is an example using gofast:

    protoc --gofast_out=plugins=grpc:. my.proto
