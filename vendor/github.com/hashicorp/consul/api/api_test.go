package api

import (
	crand "crypto/rand"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/hashicorp/consul/testutil"
)

type configCallback func(c *Config)

func makeClient(t *testing.T) (*Client, *testutil.TestServer) {
	return makeClientWithConfig(t, nil, nil)
}

func makeACLClient(t *testing.T) (*Client, *testutil.TestServer) {
	return makeClientWithConfig(t, func(clientConfig *Config) {
		clientConfig.Token = "root"
	}, func(serverConfig *testutil.TestServerConfig) {
		serverConfig.ACLMasterToken = "root"
		serverConfig.ACLDatacenter = "dc1"
		serverConfig.ACLDefaultPolicy = "deny"
	})
}

func makeClientWithConfig(
	t *testing.T,
	cb1 configCallback,
	cb2 testutil.ServerConfigCallback) (*Client, *testutil.TestServer) {

	// Make client config
	conf := DefaultConfig()
	if cb1 != nil {
		cb1(conf)
	}
	// Create server
	server, err := testutil.NewTestServerConfig(cb2)
	if err != nil {
		t.Fatal(err)
	}
	conf.Address = server.HTTPAddr

	// Create client
	client, err := NewClient(conf)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	return client, server
}

func testKey() string {
	buf := make([]byte, 16)
	if _, err := crand.Read(buf); err != nil {
		panic(fmt.Errorf("Failed to read random bytes: %v", err))
	}

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x",
		buf[0:4],
		buf[4:6],
		buf[6:8],
		buf[8:10],
		buf[10:16])
}

func TestDefaultConfig_env(t *testing.T) {
	t.Parallel()
	addr := "1.2.3.4:5678"
	token := "abcd1234"
	auth := "username:password"

	os.Setenv(HTTPAddrEnvName, addr)
	defer os.Setenv(HTTPAddrEnvName, "")
	os.Setenv(HTTPTokenEnvName, token)
	defer os.Setenv(HTTPTokenEnvName, "")
	os.Setenv(HTTPAuthEnvName, auth)
	defer os.Setenv(HTTPAuthEnvName, "")
	os.Setenv(HTTPSSLEnvName, "1")
	defer os.Setenv(HTTPSSLEnvName, "")
	os.Setenv(HTTPSSLVerifyEnvName, "0")
	defer os.Setenv(HTTPSSLVerifyEnvName, "")

	for i, config := range []*Config{DefaultConfig(), DefaultNonPooledConfig()} {
		if config.Address != addr {
			t.Errorf("expected %q to be %q", config.Address, addr)
		}
		if config.Token != token {
			t.Errorf("expected %q to be %q", config.Token, token)
		}
		if config.HttpAuth == nil {
			t.Fatalf("expected HttpAuth to be enabled")
		}
		if config.HttpAuth.Username != "username" {
			t.Errorf("expected %q to be %q", config.HttpAuth.Username, "username")
		}
		if config.HttpAuth.Password != "password" {
			t.Errorf("expected %q to be %q", config.HttpAuth.Password, "password")
		}
		if config.Scheme != "https" {
			t.Errorf("expected %q to be %q", config.Scheme, "https")
		}
		if !config.HttpClient.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify {
			t.Errorf("expected SSL verification to be off")
		}

		// Use keep alives as a check for whether pooling is on or off.
		if pooled := i == 0; pooled {
			if config.HttpClient.Transport.(*http.Transport).DisableKeepAlives != false {
				t.Errorf("expected keep alives to be enabled")
			}
		} else {
			if config.HttpClient.Transport.(*http.Transport).DisableKeepAlives != true {
				t.Errorf("expected keep alives to be disabled")
			}
		}
	}
}

func TestSetupTLSConfig(t *testing.T) {
	// A default config should result in a clean default client config.
	tlsConfig := &TLSConfig{}
	cc, err := SetupTLSConfig(tlsConfig)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	expected := &tls.Config{}
	if !reflect.DeepEqual(cc, expected) {
		t.Fatalf("bad: %v", cc)
	}

	// Try some address variations with and without ports.
	tlsConfig.Address = "127.0.0.1"
	cc, err = SetupTLSConfig(tlsConfig)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	expected.ServerName = "127.0.0.1"
	if !reflect.DeepEqual(cc, expected) {
		t.Fatalf("bad: %v", cc)
	}

	tlsConfig.Address = "127.0.0.1:80"
	cc, err = SetupTLSConfig(tlsConfig)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	expected.ServerName = "127.0.0.1"
	if !reflect.DeepEqual(cc, expected) {
		t.Fatalf("bad: %v", cc)
	}

	tlsConfig.Address = "demo.consul.io:80"
	cc, err = SetupTLSConfig(tlsConfig)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	expected.ServerName = "demo.consul.io"
	if !reflect.DeepEqual(cc, expected) {
		t.Fatalf("bad: %v", cc)
	}

	tlsConfig.Address = "[2001:db8:a0b:12f0::1]"
	cc, err = SetupTLSConfig(tlsConfig)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	expected.ServerName = "[2001:db8:a0b:12f0::1]"
	if !reflect.DeepEqual(cc, expected) {
		t.Fatalf("bad: %v", cc)
	}

	tlsConfig.Address = "[2001:db8:a0b:12f0::1]:80"
	cc, err = SetupTLSConfig(tlsConfig)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	expected.ServerName = "2001:db8:a0b:12f0::1"
	if !reflect.DeepEqual(cc, expected) {
		t.Fatalf("bad: %v", cc)
	}

	// Skip verification.
	tlsConfig.InsecureSkipVerify = true
	cc, err = SetupTLSConfig(tlsConfig)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	expected.InsecureSkipVerify = true
	if !reflect.DeepEqual(cc, expected) {
		t.Fatalf("bad: %v", cc)
	}

	// Make a new config that hits all the file parsers.
	tlsConfig = &TLSConfig{
		CertFile: "../test/hostname/Alice.crt",
		KeyFile:  "../test/hostname/Alice.key",
		CAFile:   "../test/hostname/CertAuth.crt",
	}
	cc, err = SetupTLSConfig(tlsConfig)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(cc.Certificates) != 1 {
		t.Fatalf("missing certificate: %v", cc.Certificates)
	}
	if cc.RootCAs == nil {
		t.Fatalf("didn't load root CAs")
	}
}

func TestSetQueryOptions(t *testing.T) {
	t.Parallel()
	c, s := makeClient(t)
	defer s.Stop()

	r := c.newRequest("GET", "/v1/kv/foo")
	q := &QueryOptions{
		Datacenter:        "foo",
		AllowStale:        true,
		RequireConsistent: true,
		WaitIndex:         1000,
		WaitTime:          100 * time.Second,
		Token:             "12345",
		Near:              "nodex",
	}
	r.setQueryOptions(q)

	if r.params.Get("dc") != "foo" {
		t.Fatalf("bad: %v", r.params)
	}
	if _, ok := r.params["stale"]; !ok {
		t.Fatalf("bad: %v", r.params)
	}
	if _, ok := r.params["consistent"]; !ok {
		t.Fatalf("bad: %v", r.params)
	}
	if r.params.Get("index") != "1000" {
		t.Fatalf("bad: %v", r.params)
	}
	if r.params.Get("wait") != "100000ms" {
		t.Fatalf("bad: %v", r.params)
	}
	if r.header.Get("X-Consul-Token") != "12345" {
		t.Fatalf("bad: %v", r.header)
	}
	if r.params.Get("near") != "nodex" {
		t.Fatalf("bad: %v", r.params)
	}
}

func TestSetWriteOptions(t *testing.T) {
	t.Parallel()
	c, s := makeClient(t)
	defer s.Stop()

	r := c.newRequest("GET", "/v1/kv/foo")
	q := &WriteOptions{
		Datacenter: "foo",
		Token:      "23456",
	}
	r.setWriteOptions(q)

	if r.params.Get("dc") != "foo" {
		t.Fatalf("bad: %v", r.params)
	}
	if r.header.Get("X-Consul-Token") != "23456" {
		t.Fatalf("bad: %v", r.header)
	}
}

func TestRequestToHTTP(t *testing.T) {
	t.Parallel()
	c, s := makeClient(t)
	defer s.Stop()

	r := c.newRequest("DELETE", "/v1/kv/foo")
	q := &QueryOptions{
		Datacenter: "foo",
	}
	r.setQueryOptions(q)
	req, err := r.toHTTP()
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if req.Method != "DELETE" {
		t.Fatalf("bad: %v", req)
	}
	if req.URL.RequestURI() != "/v1/kv/foo?dc=foo" {
		t.Fatalf("bad: %v", req)
	}
}

func TestParseQueryMeta(t *testing.T) {
	t.Parallel()
	resp := &http.Response{
		Header: make(map[string][]string),
	}
	resp.Header.Set("X-Consul-Index", "12345")
	resp.Header.Set("X-Consul-LastContact", "80")
	resp.Header.Set("X-Consul-KnownLeader", "true")
	resp.Header.Set("X-Consul-Translate-Addresses", "true")

	qm := &QueryMeta{}
	if err := parseQueryMeta(resp, qm); err != nil {
		t.Fatalf("err: %v", err)
	}

	if qm.LastIndex != 12345 {
		t.Fatalf("Bad: %v", qm)
	}
	if qm.LastContact != 80*time.Millisecond {
		t.Fatalf("Bad: %v", qm)
	}
	if !qm.KnownLeader {
		t.Fatalf("Bad: %v", qm)
	}
	if !qm.AddressTranslationEnabled {
		t.Fatalf("Bad: %v", qm)
	}
}

func TestAPI_UnixSocket(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.SkipNow()
	}

	tempDir, err := ioutil.TempDir("", "consul")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(tempDir)
	socket := filepath.Join(tempDir, "test.sock")

	c, s := makeClientWithConfig(t, func(c *Config) {
		c.Address = "unix://" + socket
	}, func(c *testutil.TestServerConfig) {
		c.Addresses = &testutil.TestAddressConfig{
			HTTP: "unix://" + socket,
		}
	})
	defer s.Stop()

	agent := c.Agent()

	info, err := agent.Self()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if info["Config"]["NodeName"] == "" {
		t.Fatalf("bad: %v", info)
	}
}

func TestAPI_durToMsec(t *testing.T) {
	if ms := durToMsec(0); ms != "0ms" {
		t.Fatalf("bad: %s", ms)
	}

	if ms := durToMsec(time.Millisecond); ms != "1ms" {
		t.Fatalf("bad: %s", ms)
	}

	if ms := durToMsec(time.Microsecond); ms != "1ms" {
		t.Fatalf("bad: %s", ms)
	}

	if ms := durToMsec(5 * time.Millisecond); ms != "5ms" {
		t.Fatalf("bad: %s", ms)
	}
}

func TestAPI_IsServerError(t *testing.T) {
	if IsServerError(nil) {
		t.Fatalf("should not be a server error")
	}

	if IsServerError(fmt.Errorf("not the error you are looking for")) {
		t.Fatalf("should not be a server error")
	}

	if !IsServerError(fmt.Errorf(serverError)) {
		t.Fatalf("should be a server error")
	}
}
