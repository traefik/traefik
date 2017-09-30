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
    url = "https://backend.local:50051"


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
/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

//go:generate protoc -I ../helloworld --go_out=plugins=grpc:../helloworld ../helloworld/helloworld.proto

package main

import (
	"log"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/reflection"

	"crypto/tls"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
)

const (
	port = ":50051"
)

// server is used to implement helloworld.GreeterServer.
type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

func main() {

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

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// s := grpc.NewServer()
	// pb.RegisterGreeterServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
```

Next we will modify gRPC Client to use our Træfik self-signed certificate:

```go
/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package main

import (
	"log"
	"os"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"

	"crypto/x509"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
)

const (
	address     = "localhost:4443"
	defaultName = "world"
)

func main() {

	// Read cert file
	FrontendCert, _ := ioutil.ReadFile("./frontend.cert")

	// Create CertPool
	roots := x509.NewCertPool()
	roots.AppendCertsFromPEM(FrontendCert)

	// Create credentials
	credsClient := credentials.NewClientTLSFromCert(roots, "")

	// Set up a connection to the server.
	conn, err := grpc.Dial("https://frontend.local:4443", grpc.WithTransportCredentials(credsClient))
	// conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	r, err := c.SayHello(context.Background(), &pb.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.Message)
}

```

