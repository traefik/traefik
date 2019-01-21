# gRPC examples

## With HTTP (h2c)

This section explains how to use Traefik as reverse proxy for gRPC application.

### Traefik configuration

At last, we configure our Traefik instance to use both self-signed certificates.

```toml
defaultEntryPoints = ["https"]

[entryPoints]
  [entryPoints.http]
  address = ":80"

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

### Conclusion

We don't need specific configuration to use gRPC in Traefik, we just need to use `h2c` protocol, or use HTTPS communications to have HTTP2 with the backend.

## With HTTPS

This section explains how to use Traefik as reverse proxy for gRPC application with self-signed certificates.

![gRPC architecture](/img/grpc.svg)

### gRPC Server certificate

In order to secure the gRPC server, we generate a self-signed certificate for backend url:

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./backend.key -out ./backend.cert
```

That will prompt for information, the important answer is:

```
Common Name (e.g. server FQDN or YOUR name) []: backend.local
```

### gRPC Client certificate

Generate your self-signed certificate for frontend url:

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./frontend.key -out ./frontend.cert
```

with

```
Common Name (e.g. server FQDN or YOUR name) []: frontend.local
```

### Traefik configuration

At last, we configure our Traefik instance to use both self-signed certificates.

```toml
defaultEntryPoints = ["https"]

# For secure connection on backend.local
rootCAs = [ "./backend.cert" ]

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
    # Access on backend with HTTPS
    url = "https://backend.local:8080"


[frontends]
  [frontends.frontend1]
  backend = "backend1"
    [frontends.frontend1.routes.test_1]
    rule = "Host:frontend.local"
```

!!! warning
    With some backends, the server URLs use the IP, so you may need to configure `insecureSkipVerify` instead of the `rootCAS` to activate HTTPS without hostname verification.

### A gRPC example in go (modify for https)

We use the gRPC greeter example in [grpc-go](https://github.com/grpc/grpc-go/tree/master/examples/helloworld)

!!! warning
    In order to use this gRPC example, we need to modify it to use HTTPS

So we modify the "gRPC server example" to use our own self-signed certificate:

```go
// ...

// Read cert and key file
BackendCert, _ := ioutil.ReadFile("./backend.cert")
BackendKey, _ := ioutil.ReadFile("./backend.key")

// Generate Certificate struct
cert, err := tls.X509KeyPair(BackendCert, BackendKey)
if err != nil {
  log.Fatalf("failed to parse certificate: %v", err)
}

// Create credentials
creds := credentials.NewServerTLSFromCert(&cert)

// Use Credentials in gRPC server options
serverOption := grpc.Creds(creds)
var s *grpc.Server = grpc.NewServer(serverOption)
defer s.Stop()

pb.RegisterGreeterServer(s, &server{})
err := s.Serve(lis)

// ...
```

Next we will modify gRPC Client to use our Traefik self-signed certificate:

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
