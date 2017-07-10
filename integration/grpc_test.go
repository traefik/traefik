package integration

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/containous/traefik/integration/helloworld"
	"github.com/containous/traefik/integration/try"
	"github.com/go-check/check"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var LocalhostCert = []byte(`-----BEGIN CERTIFICATE-----
MIICEzCCAXygAwIBAgIQMIMChMLGrR+QvmQvpwAU6zANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQKEwdBY21lIENvMCAXDTcwMDEwMTAwMDAwMFoYDzIwODQwMTI5MTYw
MDAwWjASMRAwDgYDVQQKEwdBY21lIENvMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCB
iQKBgQDuLnQAI3mDgey3VBzWnB2L39JUU4txjeVE6myuDqkM/uGlfjb9SjY1bIw4
iA5sBBZzHi3z0h1YV8QPuxEbi4nW91IJm2gsvvZhIrCHS3l6afab4pZBl2+XsDul
rKBxKKtD1rGxlG4LjncdabFn9gvLZad2bSysqz/qTAUStTvqJQIDAQABo2gwZjAO
BgNVHQ8BAf8EBAMCAqQwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDwYDVR0TAQH/BAUw
AwEB/zAuBgNVHREEJzAlggtleGFtcGxlLmNvbYcEfwAAAYcQAAAAAAAAAAAAAAAA
AAAAATANBgkqhkiG9w0BAQsFAAOBgQCEcetwO59EWk7WiJsG4x8SY+UIAA+flUI9
tyC4lNhbcF2Idq9greZwbYCqTTTr2XiRNSMLCOjKyI7ukPoPjo16ocHj+P3vZGfs
h1fIw3cSS2OolhloGw/XM6RWPWtPAlGykKLciQrBru5NAPvCMsb/I1DAceTiotQM
fblo6RBxUQ==
-----END CERTIFICATE-----`)

// LocalhostKey is the private key for localhostCert.
var LocalhostKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIICXgIBAAKBgQDuLnQAI3mDgey3VBzWnB2L39JUU4txjeVE6myuDqkM/uGlfjb9
SjY1bIw4iA5sBBZzHi3z0h1YV8QPuxEbi4nW91IJm2gsvvZhIrCHS3l6afab4pZB
l2+XsDulrKBxKKtD1rGxlG4LjncdabFn9gvLZad2bSysqz/qTAUStTvqJQIDAQAB
AoGAGRzwwir7XvBOAy5tM/uV6e+Zf6anZzus1s1Y1ClbjbE6HXbnWWF/wbZGOpet
3Zm4vD6MXc7jpTLryzTQIvVdfQbRc6+MUVeLKwZatTXtdZrhu+Jk7hx0nTPy8Jcb
uJqFk541aEw+mMogY/xEcfbWd6IOkp+4xqjlFLBEDytgbIECQQDvH/E6nk+hgN4H
qzzVtxxr397vWrjrIgPbJpQvBsafG7b0dA4AFjwVbFLmQcj2PprIMmPcQrooz8vp
jy4SHEg1AkEA/v13/5M47K9vCxmb8QeD/asydfsgS5TeuNi8DoUBEmiSJwma7FXY
fFUtxuvL7XvjwjN5B30pNEbc6Iuyt7y4MQJBAIt21su4b3sjXNueLKH85Q+phy2U
fQtuUE9txblTu14q3N7gHRZB4ZMhFYyDy8CKrN2cPg/Fvyt0Xlp/DoCzjA0CQQDU
y2ptGsuSmgUtWj3NM9xuwYPm+Z/F84K6+ARYiZ6PYj013sovGKUFfYAqVXVlxtIX
qyUBnu3X9ps8ZfjLZO7BAkEAlT4R5Yl6cGhaJQYZHOde3JEMhNRcVFMO8dJDaFeo
f9Oeos0UUothgiDktdQHxdNEwLjQf7lJJBzV+5OtwswCWA==
-----END RSA PRIVATE KEY-----`)

// GrpcSuite
type GrpcSuite struct{ BaseSuite }

type myserver struct{}

func (s *myserver) SayHello(ctx context.Context, in *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	return &helloworld.HelloReply{Message: "Hello " + in.Name}, nil
}

func startGrpcServer(lis net.Listener) error {
	cert, err := tls.X509KeyPair(LocalhostCert, LocalhostKey)
	if err != nil {
		return err
	}

	creds := credentials.NewServerTLSFromCert(&cert)
	serverOption := grpc.Creds(creds)

	var s *grpc.Server = grpc.NewServer(serverOption)
	defer s.Stop()

	helloworld.RegisterGreeterServer(s, &myserver{})
	return s.Serve(lis)
}

func callHelloClientGrpc() (string, error) {
	roots := x509.NewCertPool()
	roots.AppendCertsFromPEM(LocalhostCert)
	credsClient := credentials.NewClientTLSFromCert(roots, "")
	conn, err := grpc.Dial("127.0.0.1:4443", grpc.WithTransportCredentials(credsClient))
	if err != nil {
		return "", err
	}

	defer conn.Close()
	client := helloworld.NewGreeterClient(conn)

	name := "World"
	r, err := client.SayHello(context.Background(), &helloworld.HelloRequest{Name: name})
	if err != nil {
		return "", err
	}
	return r.Message, nil
}

func (suite *GrpcSuite) TestGrpc(c *check.C) {
	lis, err := net.Listen("tcp", ":0")
	_, port, err := net.SplitHostPort(lis.Addr().String())
	c.Assert(err, check.IsNil)

	go func() {
		err := startGrpcServer(lis)
		c.Assert(err, check.IsNil)
	}()

	file := suite.adaptFile(c, "fixtures/grpc/config.toml", struct {
		CertContent    string
		KeyContent     string
		GrpcServerPort string
	}{
		CertContent:    string(LocalhostCert),
		KeyContent:     string(LocalhostKey),
		GrpcServerPort: port,
	})

	defer os.Remove(file)
	cmd := exec.Command(traefikBinary, "--configFile="+file)

	err = cmd.Start()
	c.Assert(err, check.IsNil)
	defer cmd.Process.Kill()

	// wait for Traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/providers", 1*time.Second, try.BodyContains("Host:127.0.0.1"))
	c.Assert(err, check.IsNil)

	var response string
	err = try.Do(1*time.Second, func() error {
		response, err = callHelloClientGrpc()
		return err
	})

	c.Assert(err, check.IsNil)
	c.Assert(response, check.Equals, "Hello World")
}
