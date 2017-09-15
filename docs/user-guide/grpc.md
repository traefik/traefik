# gRPC example

This section explains how to use Traefik as reverse proxy for gRPC application with self-signed certificates.

!!! warning
    As gRPC needs HTTP2, we need valid HTTPS certificates on both gRPC Server and Træfik.

<p align="center">
<img src="/img/grpc.svg" alt="gRPC architecture" title="gRPC architecture" />
</p>

## gRPC Server certificate

In order to secure the gRPC server, we generate a self-signed certificate for backend url:

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./backend.key -out ./backend.crt
```

That will prompt for information, the important answer is:

```
Common Name (e.g. server FQDN or YOUR name) []: backend.local
```

## gRPC Client certificate

Generate your self-signed certificate for frontend url:

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./frontend.key -out ./frontend.crt
```

with

```
Common Name (e.g. server FQDN or YOUR name) []: frontend.local
```

## Træfik configuration

At last, we configure our Træfik instance to use both self-signed certificates.

```toml
defaultEntryPoints = ["https"]

# For secure connection on backend.local
RootCAs = [ "./backend.cert" ]

[entryPoints]
  [entryPoints.https]
  address = ":4443"
    [entryPoints.https.tls]
     # For secure connection on frontend.local
     [[entryPoints.https.tls.certificates]]
     certFile = "./frontend.cert"
     keyFile  = "./frontend.key"


[web]
  address = ":8080"

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

## Conclusion

We don't need specific configuration to use gRPC in Træfik, we just need to be careful that all the exchanges (between client and Træfik, and between Træfik and backend) are valid HTTPS communications (without `InsecureSkipVerify` enabled) because gRPC use HTTP2.

## A gRPC example in go

We will use the gRPC greeter example in [grpc-go](https://github.com/grpc/grpc-go/tree/master/examples/helloworld)

!!! warning
    In order to use this gRPC example, we need to modify it to use HTTPS

So we modify the "gRPC server example" to use our own self-signed certificate:

```go
// ...

// Read cert and key file
BackendCert := ioutil.ReadFile("./backend.cert")
BackendKey := ioutil.ReadFile("./backend.key")

// Generate Certificate struct
cert, err := tls.X509KeyPair(BackendCert, BackendKey)
if err != nil {
  return err
}

// Create credentials
creds := credentials.NewServerTLSFromCert(&cert)

// Use Credentials in gRPC server options
serverOption := grpc.Creds(creds)
var s *grpc.Server = grpc.NewServer(serverOption)
defer s.Stop()

helloworld.RegisterGreeterServer(s, &myserver{})
err := s.Serve(lis)

// ...
```

Next we will modify gRPC Client to use our Træfik self-signed certificate:

```go
// ...

// Read cert file
FrontendCert := ioutil.ReadFile("./frontend.cert")

// Create CertPool
roots := x509.NewCertPool()
roots.AppendCertsFromPEM(FrontendCert)

// Create credentials
credsClient := credentials.NewClientTLSFromCert(roots, "")

// Dial with specific Transport (with credentials)
conn, err := grpc.Dial("https://frontend:4443", grpc.WithTransportCredentials(credsClient))
if err != nil {
  return err
}

defer conn.Close()
client := helloworld.NewGreeterClient(conn)

name := "World"
r, err := client.SayHello(context.Background(), &helloworld.HelloRequest{Name: name})

// ...
```

