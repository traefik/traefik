# grpc

[gRPC](http://www.grpc.io/) is an excellent, modern IDL and transport for microservices.
If you're starting a greenfield project, Go kit strongly recommends gRPC as your default transport.
And using gRPC and Go kit together is very simple.

First, define your service using protobuf3.
This is explained [in gRPC documentation](http://www.grpc.io/docs/#defining-a-service).
See [add.proto](https://github.com/go-kit/kit/blob/ec8b02591ee873433565a1ae9d317353412d1d27/examples/addsvc/pb/add.proto) for an example.
Make sure the proto definition matches your service's Go kit (interface) definition.

Next, get the protoc compiler.
Unfortunately, this needs to be done from source.
Fortunately, it's pretty straightforward.

```
brew install autoconf automake libtool
git clone https://github.com/google/protobuf
cd protobuf
./autogen.sh ; ./configure ; make ; make install
```

Then, compile your service definition, from .proto to .go.

```
protoc add.proto --go_out=plugins=grpc:.
```

Finally, write a tiny binding from your service definition to the gRPC definition.
It's a simple conversion from one domain to another.
See [grpc_binding.go](https://github.com/go-kit/kit/blob/ec8b02591ee873433565a1ae9d317353412d1d27/examples/addsvc/grpc_binding.go) for an example.

That's it!
The gRPC binding can be bound to a listener and serve normal gRPC requests.
And within your service, you can use standard Go kit components and idioms.
See [addsvc](https://github.com/go-kit/kit/tree/master/examples/addsvc) for a complete working example with gRPC support.
And remember: Go kit services can support multiple transports simultaneously.
