package integration

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"math/rand"
	"net"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/traefik/traefik/v3/integration/helloworld"
	"github.com/traefik/traefik/v3/integration/try"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	LocalhostCert []byte
	LocalhostKey  []byte
)

const randCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

// GRPCSuite tests suite.
type GRPCSuite struct{ BaseSuite }

func TestGRPCSuite(t *testing.T) {
	suite.Run(t, new(GRPCSuite))
}

type myserver struct {
	stopStreamExample chan bool
}

func (s *GRPCSuite) SetupSuite() {
	var err error
	LocalhostCert, err = os.ReadFile("./resources/tls/local.cert")
	assert.NoError(s.T(), err)
	LocalhostKey, err = os.ReadFile("./resources/tls/local.key")
	assert.NoError(s.T(), err)
}

func (s *myserver) SayHello(ctx context.Context, in *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	return &helloworld.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func (s *myserver) StreamExample(in *helloworld.StreamExampleRequest, server helloworld.Greeter_StreamExampleServer) error {
	data := make([]byte, 512)
	for i := range data {
		data[i] = randCharset[rand.Intn(len(randCharset))]
	}

	if err := server.Send(&helloworld.StreamExampleReply{Data: string(data)}); err != nil {
		log.Error().Err(err).Send()
	}

	<-s.stopStreamExample
	return nil
}

func startGRPCServer(lis net.Listener, server *myserver) error {
	cert, err := tls.X509KeyPair(LocalhostCert, LocalhostKey)
	if err != nil {
		return err
	}

	creds := credentials.NewServerTLSFromCert(&cert)
	serverOption := grpc.Creds(creds)

	s := grpc.NewServer(serverOption)
	defer s.Stop()

	helloworld.RegisterGreeterServer(s, server)
	return s.Serve(lis)
}

func starth2cGRPCServer(lis net.Listener, server *myserver) error {
	s := grpc.NewServer()
	defer s.Stop()

	helloworld.RegisterGreeterServer(s, server)
	return s.Serve(lis)
}

func getHelloClientGRPC() (helloworld.GreeterClient, func() error, error) {
	roots := x509.NewCertPool()
	roots.AppendCertsFromPEM(LocalhostCert)
	credsClient := credentials.NewClientTLSFromCert(roots, "")
	conn, err := grpc.NewClient("127.0.0.1:4443", grpc.WithTransportCredentials(credsClient))
	if err != nil {
		return nil, func() error { return nil }, err
	}
	return helloworld.NewGreeterClient(conn), conn.Close, nil
}

func getHelloClientGRPCh2c() (helloworld.GreeterClient, func() error, error) {
	conn, err := grpc.NewClient("127.0.0.1:8081", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, func() error { return nil }, err
	}
	return helloworld.NewGreeterClient(conn), conn.Close, nil
}

func callHelloClientGRPC(t *testing.T, name string, secure bool) (string, error) {
	t.Helper()

	var client helloworld.GreeterClient
	var closer func() error
	var err error

	if secure {
		client, closer, err = getHelloClientGRPC()
	} else {
		client, closer, err = getHelloClientGRPCh2c()
	}
	defer func() { _ = closer() }()

	if err != nil {
		return "", err
	}
	r, err := client.SayHello(t.Context(), &helloworld.HelloRequest{Name: name})
	if err != nil {
		return "", err
	}
	return r.GetMessage(), nil
}

func callStreamExampleClientGRPC(t *testing.T) (helloworld.Greeter_StreamExampleClient, func() error, error) {
	t.Helper()

	client, closer, err := getHelloClientGRPC()
	if err != nil {
		return nil, closer, err
	}
	s, err := client.StreamExample(t.Context(), &helloworld.StreamExampleRequest{})
	if err != nil {
		return nil, closer, err
	}

	return s, closer, nil
}

func (s *GRPCSuite) TestGRPC() {
	lis, err := net.Listen("tcp", ":0")
	assert.NoError(s.T(), err)
	_, port, err := net.SplitHostPort(lis.Addr().String())
	assert.NoError(s.T(), err)

	go func() {
		err := startGRPCServer(lis, &myserver{})
		assert.NoError(s.T(), err)
	}()

	file := s.adaptFile("fixtures/grpc/config.toml", struct {
		CertContent    string
		KeyContent     string
		GRPCServerPort string
	}{
		CertContent:    string(LocalhostCert),
		KeyContent:     string(LocalhostKey),
		GRPCServerPort: port,
	})

	s.traefikCmd(withConfigFile(file))

	// wait for Traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("Host(`127.0.0.1`)"))
	assert.NoError(s.T(), err)

	var response string
	err = try.Do(1*time.Second, func() error {
		response, err = callHelloClientGRPC(s.T(), "World", true)
		return err
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "Hello World", response)
}

func (s *GRPCSuite) TestGRPCh2c() {
	lis, err := net.Listen("tcp", ":0")
	assert.NoError(s.T(), err)
	_, port, err := net.SplitHostPort(lis.Addr().String())
	assert.NoError(s.T(), err)

	go func() {
		err := starth2cGRPCServer(lis, &myserver{})
		assert.NoError(s.T(), err)
	}()

	file := s.adaptFile("fixtures/grpc/config_h2c.toml", struct {
		GRPCServerPort string
	}{
		GRPCServerPort: port,
	})

	s.traefikCmd(withConfigFile(file))

	// wait for Traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("Host(`127.0.0.1`)"))
	assert.NoError(s.T(), err)

	var response string
	err = try.Do(1*time.Second, func() error {
		response, err = callHelloClientGRPC(s.T(), "World", false)
		return err
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "Hello World", response)
}

func (s *GRPCSuite) TestGRPCh2cTermination() {
	lis, err := net.Listen("tcp", ":0")
	assert.NoError(s.T(), err)
	_, port, err := net.SplitHostPort(lis.Addr().String())
	assert.NoError(s.T(), err)

	go func() {
		err := starth2cGRPCServer(lis, &myserver{})
		assert.NoError(s.T(), err)
	}()

	file := s.adaptFile("fixtures/grpc/config_h2c_termination.toml", struct {
		CertContent    string
		KeyContent     string
		GRPCServerPort string
	}{
		CertContent:    string(LocalhostCert),
		KeyContent:     string(LocalhostKey),
		GRPCServerPort: port,
	})

	s.traefikCmd(withConfigFile(file))

	// wait for Traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("Host(`127.0.0.1`)"))
	assert.NoError(s.T(), err)

	var response string
	err = try.Do(1*time.Second, func() error {
		response, err = callHelloClientGRPC(s.T(), "World", true)
		return err
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "Hello World", response)
}

func (s *GRPCSuite) TestGRPCInsecure() {
	lis, err := net.Listen("tcp", ":0")
	assert.NoError(s.T(), err)
	_, port, err := net.SplitHostPort(lis.Addr().String())
	assert.NoError(s.T(), err)

	go func() {
		err := startGRPCServer(lis, &myserver{})
		assert.NoError(s.T(), err)
	}()

	file := s.adaptFile("fixtures/grpc/config_insecure.toml", struct {
		CertContent    string
		KeyContent     string
		GRPCServerPort string
	}{
		CertContent:    string(LocalhostCert),
		KeyContent:     string(LocalhostKey),
		GRPCServerPort: port,
	})

	s.traefikCmd(withConfigFile(file))

	// wait for Traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("Host(`127.0.0.1`)"))
	assert.NoError(s.T(), err)

	var response string
	err = try.Do(1*time.Second, func() error {
		response, err = callHelloClientGRPC(s.T(), "World", true)
		return err
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "Hello World", response)
}

func (s *GRPCSuite) TestGRPCBuffer() {
	stopStreamExample := make(chan bool)
	defer func() { stopStreamExample <- true }()
	lis, err := net.Listen("tcp", ":0")
	assert.NoError(s.T(), err)
	_, port, err := net.SplitHostPort(lis.Addr().String())
	assert.NoError(s.T(), err)

	go func() {
		err := startGRPCServer(lis, &myserver{
			stopStreamExample: stopStreamExample,
		})
		assert.NoError(s.T(), err)
	}()

	file := s.adaptFile("fixtures/grpc/config.toml", struct {
		CertContent    string
		KeyContent     string
		GRPCServerPort string
	}{
		CertContent:    string(LocalhostCert),
		KeyContent:     string(LocalhostKey),
		GRPCServerPort: port,
	})

	s.traefikCmd(withConfigFile(file))

	// wait for Traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("Host(`127.0.0.1`)"))
	assert.NoError(s.T(), err)
	var client helloworld.Greeter_StreamExampleClient
	client, closer, err := callStreamExampleClientGRPC(s.T())
	defer func() { _ = closer() }()
	assert.NoError(s.T(), err)

	received := make(chan bool)
	go func() {
		tr, err := client.Recv()
		assert.NoError(s.T(), err)
		assert.Len(s.T(), tr.GetData(), 512)
		received <- true
	}()

	err = try.Do(10*time.Second, func() error {
		select {
		case <-received:
			return nil
		default:
			return errors.New("failed to receive stream data")
		}
	})
	assert.NoError(s.T(), err)
}

func (s *GRPCSuite) TestGRPCBufferWithFlushInterval() {
	stopStreamExample := make(chan bool)
	lis, err := net.Listen("tcp", ":0")
	assert.NoError(s.T(), err)
	_, port, err := net.SplitHostPort(lis.Addr().String())
	assert.NoError(s.T(), err)

	go func() {
		err := startGRPCServer(lis, &myserver{
			stopStreamExample: stopStreamExample,
		})
		assert.NoError(s.T(), err)
	}()

	file := s.adaptFile("fixtures/grpc/config.toml", struct {
		CertContent    string
		KeyContent     string
		GRPCServerPort string
	}{
		CertContent:    string(LocalhostCert),
		KeyContent:     string(LocalhostKey),
		GRPCServerPort: port,
	})

	s.traefikCmd(withConfigFile(file))
	// wait for Traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("Host(`127.0.0.1`)"))
	assert.NoError(s.T(), err)

	var client helloworld.Greeter_StreamExampleClient
	client, closer, err := callStreamExampleClientGRPC(s.T())
	defer func() {
		_ = closer()
		stopStreamExample <- true
	}()
	assert.NoError(s.T(), err)

	received := make(chan bool)
	go func() {
		tr, err := client.Recv()
		assert.NoError(s.T(), err)
		assert.Len(s.T(), tr.GetData(), 512)
		received <- true
	}()

	err = try.Do(100*time.Millisecond, func() error {
		select {
		case <-received:
			return nil
		default:
			return errors.New("failed to receive stream data")
		}
	})
	assert.NoError(s.T(), err)
}

func (s *GRPCSuite) TestGRPCWithRetry() {
	lis, err := net.Listen("tcp", ":0")
	assert.NoError(s.T(), err)
	_, port, err := net.SplitHostPort(lis.Addr().String())
	assert.NoError(s.T(), err)

	go func() {
		err := startGRPCServer(lis, &myserver{})
		assert.NoError(s.T(), err)
	}()

	file := s.adaptFile("fixtures/grpc/config_retry.toml", struct {
		CertContent    string
		KeyContent     string
		GRPCServerPort string
	}{
		CertContent:    string(LocalhostCert),
		KeyContent:     string(LocalhostKey),
		GRPCServerPort: port,
	})

	s.traefikCmd(withConfigFile(file))

	// wait for Traefik
	err = try.GetRequest("http://127.0.0.1:8080/api/rawdata", 1*time.Second, try.BodyContains("Host(`127.0.0.1`)"))
	assert.NoError(s.T(), err)

	var response string
	err = try.Do(1*time.Second, func() error {
		response, err = callHelloClientGRPC(s.T(), "World", true)
		return err
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "Hello World", response)
}
