# gRPC example

This section explains how to use Traefik as reverse proxy for gRPC application with self-signed certificates.

!!! warning
    As gRPC needs HTTP2, we need HTTPS certificates on Træfik.
    For exchanges with the backend, we will use h2c (HTTP2 on HTTP without TLS)

<p align="center">
<img src="/img/grpc.svg" alt="gRPC architecture" title="gRPC architecture" />
</p>

## gRPC Client certificate

Generate your self-signed certificate for frontend url:

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./frontend.key -out ./frontend.cert
```

with

```
Common Name (e.g. server FQDN or YOUR name) []: frontend.local
```

## Træfik configuration

At last, we configure our Træfik instance to use both self-signed certificates.

```toml
defaultEntryPoints = ["https"]

[entryPoints]
  [entryPoints.https]
  address = ":4443"
    [entryPoints.https.tls]
     # For secure connection on frontend.local
     [[entryPoints.https.tls.certificates]]
     certFile = "./frontend.cert"
     keyFile  = "./frontend.key"


[api]

[file]

[backends]
  [backends.backend1]
    [backends.backend1.servers.server1]
    # Access on backend with h2c
    url = "h2c://backend.local:8080"


[frontends]
  [frontends.frontend1]
  backend = "backend1"
    [frontends.frontend1.routes.test_1]
    rule = "Host:frontend.local"
```

!!! warning
    For provider with label, you will have to specify the `traefik.protocol=h2c`

## Conclusion

We don't need specific configuration to use gRPC in Træfik, we just need to be careful that exchanges between client and Træfik are HTTPS communications.
For exchanges between Træfik and backend, you need to use `h2c` protocol, or use HTTPS communications to have HTTP2.

## A gRPC example in go

We will use the gRPC greeter example in [grpc-go](https://github.com/grpc/grpc-go/tree/master/examples/helloworld)

We can keep the Server example as is with the h2c protocol
```go
// ...
lis, err := net.Listen("tcp", port)
if err != nil {
    log.Fatalf("failed to listen: %v", err)
}
var s *grpc.Server = grpc.NewServer()
defer s.Stop()

pb.RegisterGreeterServer(s, &server{})
err := s.Serve(lis)

// ...
```

!!! warning
    In order to use this gRPC example, we need to modify it to use HTTPS


Next we will modify gRPC Client to use our Træfik self-signed certificate:

```go
// ...

// Read cert file
FrontendCert, _ := ioutil.ReadFile("./frontend.cert")

// Create CertPool
roots := x509.NewCertPool()
roots.AppendCertsFromPEM(FrontendCert)

// Create credentials
credsClient := credentials.NewClientTLSFromCert(roots, "")

// Dial with specific Transport (with credentials)
conn, err := grpc.Dial("frontend.local:4443", grpc.WithTransportCredentials(credsClient))
if err != nil {
    log.Fatalf("did not connect: %v", err)
}

defer conn.Close()
client := pb.NewGreeterClient(conn)

name := "World"
r, err := client.SayHello(context.Background(), &pb.HelloRequest{Name: name})

// ...
```
