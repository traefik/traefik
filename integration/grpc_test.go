package integration

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"os"
	"time"

	"github.com/containous/traefik/integration/helloworld"
	"github.com/containous/traefik/integration/try"
	"github.com/go-check/check"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var LocalhostCert []byte
var LocalhostKey []byte

// GRPCSuite
type GRPCSuite struct{ BaseSuite }

type myserver struct{}

func (suite *GRPCSuite) SetUpSuite(c *check.C) {
	var err error
	LocalhostCert, err = ioutil.ReadFile("./resources/tls/local.cert")
	c.Assert(err, check.IsNil)
	LocalhostKey, err = ioutil.ReadFile("./resources/tls/local.key")
	c.Assert(err, check.IsNil)
}

func (s *myserver) SayHello(ctx context.Context, in *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	return &helloworld.HelloReply{Message: "Hello " + in.Name}, nil
}

func startGRPCServer(lis net.Listener) error {
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

func callHelloClientGRPC() (string, error) {
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

func (suite *GRPCSuite) TestGRPC(c *check.C) {
	lis, err := net.Listen("tcp", ":0")
	_, port, err := net.SplitHostPort(lis.Addr().String())
	c.Assert(err, check.IsNil)

	go func() {
		err := startGRPCServer(lis)
		c.Log(err)
		c.Assert(err, check.IsNil)
	}()

	file := suite.adaptFile(c, "fixtures/grpc/config.toml", struct {
		CertContent    string
		KeyContent     string
		GRPCServerPort string
	}{
		CertContent:    string(LocalhostCert),
		KeyContent:     string(LocalhostKey),
		GRPCServerPort: port,
	})

	defer os.Remove(file)
	cmd, _ := suite.cmdTraefik(withConfigFile(file))

	err = cmd.Start()
	c.Assert(err, check.IsNil)
	defer cmd.Process.Kill()

	// wait for Traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/providers", 1*time.Second, try.BodyContains("Host:127.0.0.1"))
	c.Assert(err, check.IsNil)
	var response string
	err = try.Do(1*time.Second, func() error {
		response, err = callHelloClientGRPC()
		return err
	})

	c.Assert(err, check.IsNil)
	c.Assert(response, check.Equals, "Hello World")
}
