package integration

import (
	"bufio"
	"github.com/pires/go-proxyproto"
	"net"
	"os"
	"time"

	"github.com/go-check/check"
	"github.com/traefik/traefik/v2/integration/try"
	checker "github.com/vdemeester/shakers"
)

type ProxyProtocolSuite struct {
	BaseSuite
	whoamiIP string
}

func (s *ProxyProtocolSuite) SetUpSuite(c *check.C) {
	s.BaseSuite.SetUpSuite(c)

	s.createComposeProject(c, "proxy-protocol")
	s.composeUp(c)

	s.whoamiIP = s.getComposeServiceIP(c, "whoami")
}

func (s *ProxyProtocolSuite) TearDownSuite(c *check.C) {
	s.BaseSuite.TearDownSuite(c)
}

func (s *ProxyProtocolSuite) TestProxyProtocolTrusted(c *check.C) {
	file := s.adaptFile(c, "fixtures/proxy-protocol/proxy-protocol.toml", struct {
		HaproxyIP string
		WhoamiIP  string
	}{WhoamiIP: s.whoamiIP})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:8000/whoami", 10*time.Second)
	c.Assert(err, checker.IsNil)

	content, err := proxyProtoRequest("127.0.0.1:8000", 1)
	c.Assert(err, checker.IsNil)
	c.Assert(content, checker.Contains, "X-Forwarded-For: 1.2.3.4")

	content, err = proxyProtoRequest("127.0.0.1:8000", 2)
	c.Assert(err, checker.IsNil)
	c.Assert(content, checker.Contains, "X-Forwarded-For: 1.2.3.4")
}

func (s *ProxyProtocolSuite) TestProxyProtocolNotTrusted(c *check.C) {
	file := s.adaptFile(c, "fixtures/proxy-protocol/proxy-protocol.toml", struct {
		HaproxyIP string
		WhoamiIP  string
	}{WhoamiIP: s.whoamiIP})
	defer os.Remove(file)

	cmd, display := s.traefikCmd(withConfigFile(file))
	defer display(c)
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer s.killCmd(cmd)

	err = try.GetRequest("http://127.0.0.1:9000/whoami", 10*time.Second)
	c.Assert(err, checker.IsNil)

	content, err := proxyProtoRequest("127.0.0.1:9000", 1)
	c.Assert(err, checker.IsNil)
	c.Assert(content, checker.Contains, "X-Forwarded-For: 127.0.0.1")

	content, err = proxyProtoRequest("127.0.0.1:9000", 2)
	c.Assert(err, checker.IsNil)
	c.Assert(content, checker.Contains, "X-Forwarded-For: 127.0.0.1")
}

func proxyProtoRequest(address string, version byte) (string, error) {
	// Open a TCP connection to the server
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	// Create a Proxy Protocol header with v1
	proxyHeader := &proxyproto.Header{
		Version:           version,
		Command:           proxyproto.PROXY,
		TransportProtocol: proxyproto.TCPv4,
		DestinationAddr: &net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 8000,
		},
		SourceAddr: &net.TCPAddr{
			IP:   net.ParseIP("1.2.3.4"),
			Port: 62541,
		},
	}

	// After the connection was created write the proxy headers first
	_, err = proxyHeader.WriteTo(conn)
	if err != nil {
		return "", err
	}

	// Create an HTTP request
	request := "GET /whoami HTTP/1.1\r\n" +
		"Host: 127.0.0.1\r\n" +
		"Connection: close\r\n" +
		"\r\n"

	// Write the HTTP request to the TCP connection
	writer := bufio.NewWriter(conn)
	_, err = writer.WriteString(request)
	if err != nil {
		return "", err
	}

	// Flush the buffer to ensure the request is sent
	err = writer.Flush()
	if err != nil {
		return "", err
	}

	// Read the response from the server
	var content string
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		content += scanner.Text() + "\n"
	}

	if scanner.Err() != nil {
		return "", err
	}

	return content, nil
}
