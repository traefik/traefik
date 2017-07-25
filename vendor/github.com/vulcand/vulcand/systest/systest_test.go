// package systest contains "Black box" tests that configure Vulcand using various methods and making sure
// Vulcand accepts the configuration and is capable of processing requests.
package systest

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/coreos/go-etcd/etcd"
	"github.com/vulcand/oxy/testutils"
	"github.com/vulcand/vulcand/engine"
	"github.com/vulcand/vulcand/secret"
	. "github.com/vulcand/vulcand/testutils"

	. "gopkg.in/check.v1"
)

func TestVulcandWithEtcd(t *testing.T) { TestingT(t) }

// VESuite performs "Black box" system test of Vulcan backed by Etcd by talking directly to Etcd and checking the side-effects
type VESuite struct {
	apiUrl     string
	serviceUrl string
	etcdNodes  []string
	etcdPrefix string
	sealKey    string
	box        *secret.Box
	client     *etcd.Client
}

var _ = Suite(&VESuite{})

func (s *VESuite) name(prefix string) string {
	return fmt.Sprintf("%s%d", prefix, time.Now().UnixNano())
}

func (s *VESuite) SetUpSuite(c *C) {
	etcdNodes := os.Getenv("VULCAND_TEST_ETCD_NODES")
	if etcdNodes == "" {
		c.Skip("This test requires running Etcd, please provide url via VULCAND_TEST_ETCD_NODES environment variable")
		return
	}
	s.etcdNodes = strings.Split(etcdNodes, ",")
	s.client = etcd.NewClient(s.etcdNodes)

	s.etcdPrefix = os.Getenv("VULCAND_TEST_ETCD_PREFIX")
	if s.etcdPrefix == "" {
		c.Skip("This test requires Etcd prefix, please provide url via VULCAND_TEST_ETCD_PREFIX environment variable")
		return
	}

	s.apiUrl = os.Getenv("VULCAND_TEST_API_URL")
	if s.apiUrl == "" {
		c.Skip("This test requires running vulcand daemon, provide API url via VULCAND_TEST_API_URL environment variable")
		return
	}

	s.serviceUrl = os.Getenv("VULCAND_TEST_SERVICE_URL")
	if s.serviceUrl == "" {
		c.Skip("This test requires running vulcand daemon, provide API url via VULCAND_TEST_SERVICE_URL environment variable")
		return
	}

	s.sealKey = os.Getenv("VULCAND_TEST_SEAL_KEY")
	if s.sealKey == "" {
		c.Skip("This test requires running vulcand daemon, provide API url via VULCAND_TEST_SEAL_KEY environment variable")
		return
	}

	key, err := secret.KeyFromString(s.sealKey)
	c.Assert(err, IsNil)

	box, err := secret.NewBox(key)
	c.Assert(err, IsNil)

	s.box = box
}

func (s VESuite) path(keys ...string) string {
	return strings.Join(append([]string{s.etcdPrefix}, keys...), "/")
}

func (s *VESuite) SetUpTest(c *C) {
	// Delete all values under the given prefix
	_, err := s.client.Get(s.etcdPrefix, false, false)
	if err != nil {
		e, ok := err.(*etcd.EtcdError)
		// We haven't expected this error, oops
		if !ok || e.ErrorCode != 100 {
			c.Assert(err, IsNil)
		}
	} else {
		_, err = s.client.Delete(s.etcdPrefix, true)
		c.Assert(err, IsNil)
	}

	// Restart vulcand
	exec.Command("killall", "vulcand").Output()

	args := []string{
		fmt.Sprintf("--etcdKey=%s", s.etcdPrefix),
		fmt.Sprintf("--sealKey=%s", s.sealKey),
		"--logSeverity=INFO",
		"--log=syslog",
	}
	for _, n := range s.etcdNodes {
		args = append(args, fmt.Sprintf("-etcd=%s", n))
	}
	cmd := exec.Command("vulcand", args...)
	c.Assert(cmd.Start(), IsNil)
	go func() {
		// Wait for process completion to avoid zombie process
		cmd.Wait()
	}()
}

func (s *VESuite) TearDownTest(c *C) {
	exec.Command("killall", "vulcand").Output()
}

// Set up a frontend hit this frontend with request and make sure everything worked fine
func (s *VESuite) TestFrontendCRUD(c *C) {
	called := false
	server := testutils.NewHandler(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.Write([]byte("Hi, I'm fine, thanks!"))
	})
	defer server.Close()

	// Create a server
	b, srv, url := "bk1", "srv1", server.URL
	_, err := s.client.Set(s.path("backends", b, "backend"), `{"Type": "http"}`, 0)
	c.Assert(err, IsNil)

	_, err = s.client.Set(s.path("backends", b, "servers", srv), fmt.Sprintf(`{"URL": "%s"}`, url), 0)
	c.Assert(err, IsNil)

	// Add frontend
	fId := "fr1"
	_, err = s.client.Set(s.path("frontends", fId, "frontend"),
		`{"Type": "http", "BackendId": "bk1", "Route": "Path(\"/path\")"}`, 0)
	c.Assert(err, IsNil)

	time.Sleep(time.Second)
	response, _, err := testutils.Get(fmt.Sprintf("%s%s", s.serviceUrl, "/path"))
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)
	c.Assert(called, Equals, true)
}

func (s *VESuite) TestFrontendUpdateLimits(c *C) {
	var headers http.Header
	server := testutils.NewHandler(func(w http.ResponseWriter, r *http.Request) {
		headers = r.Header
		w.Write([]byte("Hello, I'm totally fine"))
	})
	defer server.Close()

	b, srv, url := "bk1", "srv1", server.URL
	_, err := s.client.Set(s.path("backends", b, "backend"), `{"Type": "http"}`, 0)
	c.Assert(err, IsNil)

	_, err = s.client.Set(s.path("backends", b, "servers", srv), fmt.Sprintf(`{"URL": "%s"}`, url), 0)
	c.Assert(err, IsNil)

	// Add frontend
	fId := "fr1"
	_, err = s.client.Set(s.path("frontends", fId, "frontend"), `{"Type": "http", "BackendId": "bk1", "Route": "Path(\"/path\")"}`, 0)
	c.Assert(err, IsNil)

	time.Sleep(time.Second)
	response, _, err := testutils.Get(fmt.Sprintf("%s%s", s.serviceUrl, "/path"))
	c.Assert(err, IsNil)

	c.Assert(response.StatusCode, Equals, http.StatusOK)
	c.Assert(response.Header.Get("X-Forwarded-For"), Not(Equals), "hello")

	_, err = s.client.Set(
		s.path("frontends", fId, "frontend"),
		`{"Type": "http", "BackendId": "bk1", "Route": "Path(\"/path\")", "Settings": {"Limits": {"MaxMemBodyBytes":2, "MaxBodyBytes":4}}}`, 0)
	c.Assert(err, IsNil)
	time.Sleep(time.Second)

	response, _, err = testutils.Get(fmt.Sprintf("%s%s", s.serviceUrl, "/path"), testutils.Body("This is longer than allowed 4 bytes"))
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusRequestEntityTooLarge)
}

func (s *VESuite) TestFrontendUpdateBackend(c *C) {
	server1 := testutils.NewHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("1"))
	})
	defer server1.Close()

	server2 := testutils.NewHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("2"))
	})
	defer server2.Close()

	// Create two different backends
	b1, srv1, url1 := "bk1", "srv1", server1.URL

	_, err := s.client.Set(s.path("backends", b1, "backend"), `{"Type": "http"}`, 0)
	c.Assert(err, IsNil)

	_, err = s.client.Set(s.path("backends", b1, "servers", srv1), fmt.Sprintf(`{"URL": "%s"}`, url1), 0)
	c.Assert(err, IsNil)

	b2, srv2, url2 := "bk2", "srv2", server2.URL
	_, err = s.client.Set(s.path("backends", b2, "backend"), `{"Type": "http"}`, 0)
	c.Assert(err, IsNil)

	_, err = s.client.Set(s.path("backends", b2, "servers", srv2), fmt.Sprintf(`{"URL": "%s"}`, url2), 0)
	c.Assert(err, IsNil)

	// Add frontend inititally pointing to the first backend
	fId := "fr1"
	_, err = s.client.Set(s.path("frontends", fId, "frontend"), `{"Type": "http", "BackendId": "bk1", "Route": "Path(\"/path\")"}`, 0)
	c.Assert(err, IsNil)

	time.Sleep(time.Second)
	url := fmt.Sprintf("%s%s", s.serviceUrl, "/path")
	response, body, err := testutils.Get(url)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)
	c.Assert(string(body), Equals, "1")

	// Update the backend
	_, err = s.client.Set(s.path("frontends", fId, "frontend"), `{"Type": "http", "BackendId": "bk2", "Route": "Path(\"/path\")"}`, 0)
	c.Assert(err, IsNil)

	time.Sleep(time.Second)
	response, body, err = testutils.Get(url)
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)
	c.Assert(string(body), Equals, "2")
}

func (s *VESuite) TestHTTPListenerCRUD(c *C) {
	called := false
	server := testutils.NewHandler(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.Write([]byte("Hi, I'm fine, thanks!"))
	})
	defer server.Close()

	b, srv, url := "bk1", "srv1", server.URL
	_, err := s.client.Set(s.path("backends", b, "backend"), `{"Type": "http"}`, 0)
	c.Assert(err, IsNil)

	_, err = s.client.Set(s.path("backends", b, "servers", srv), fmt.Sprintf(`{"URL": "%s"}`, url), 0)
	c.Assert(err, IsNil)

	// Add frontend
	fId := "fr1"
	_, err = s.client.Set(s.path("frontends", fId, "frontend"), `{"Type": "http", "BackendId": "bk1", "Route": "Path(\"/path\")"}`, 0)
	c.Assert(err, IsNil)

	time.Sleep(time.Second)
	response, _, err := testutils.Get(fmt.Sprintf("%s%s", s.serviceUrl, "/path"))
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)

	// Add HTTP listener
	l1 := "l1"
	listener, err := engine.NewListener(l1, "http", "tcp", "localhost:31000", "", nil)
	c.Assert(err, IsNil)
	bytes, err := json.Marshal(listener)
	c.Assert(err, IsNil)
	s.client.Set(s.path("listeners", l1), string(bytes), 0)

	time.Sleep(time.Second)
	_, _, err = testutils.Get(fmt.Sprintf("%s%s", "http://localhost:31000", "/path"))
	c.Assert(err, IsNil)
	c.Assert(called, Equals, true)

	_, err = s.client.Delete(s.path("listeners", l1), true)
	c.Assert(err, IsNil)

	time.Sleep(time.Second)

	_, _, err = testutils.Get(fmt.Sprintf("%s%s", "http://localhost:31000", "/path"))
	c.Assert(err, NotNil)
}

func (s *VESuite) TestHTTPSListenerCRUD(c *C) {
	called := false
	server := testutils.NewHandler(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.Write([]byte("Hi, I'm fine, thanks!"))
	})
	defer server.Close()

	b, srv, url := "bk1", "srv1", server.URL
	_, err := s.client.Set(s.path("backends", b, "backend"), `{"Type": "http"}`, 0)
	c.Assert(err, IsNil)

	_, err = s.client.Set(s.path("backends", b, "servers", srv), fmt.Sprintf(`{"URL": "%s"}`, url), 0)
	c.Assert(err, IsNil)

	// Add frontend
	fId := "fr1"
	_, err = s.client.Set(s.path("frontends", fId, "frontend"), `{"Type": "http", "BackendId": "bk1", "Route": "Path(\"/path\")"}`, 0)
	c.Assert(err, IsNil)

	keyPair := NewTestKeyPair()

	bytes, err := secret.SealKeyPairToJSON(s.box, keyPair)
	c.Assert(err, IsNil)
	sealed := base64.StdEncoding.EncodeToString(bytes)
	host := "localhost"

	_, err = s.client.Set(s.path("hosts", host, "host"), fmt.Sprintf(`{"Name": "localhost", "Settings": {"KeyPair": "%v"}}`, sealed), 0)
	c.Assert(err, IsNil)

	// Add HTTPS listener
	l2 := "ls2"
	listener, err := engine.NewListener(l2, "https", "tcp", "localhost:32000", "", nil)
	c.Assert(err, IsNil)
	bytes, err = json.Marshal(listener)
	c.Assert(err, IsNil)
	s.client.Set(s.path("listeners", l2), string(bytes), 0)

	time.Sleep(time.Second)
	_, _, err = testutils.Get(fmt.Sprintf("%s%s", "https://localhost:32000", "/path"))
	c.Assert(err, IsNil)
	c.Assert(called, Equals, true)

	_, err = s.client.Delete(s.path("listeners", l2), true)
	c.Assert(err, IsNil)

	time.Sleep(time.Second)

	_, _, err = testutils.Get(fmt.Sprintf("%s%s", "https://localhost:32000", "/path"))
	c.Assert(err, NotNil)
}

func (s *VESuite) TestExpiringServer(c *C) {
	server := testutils.NewResponder("e1")
	defer server.Close()

	server2 := testutils.NewResponder("e2")
	defer server2.Close()

	// Create backend and servers
	b, url, url2 := "bk1", server.URL, server2.URL
	srv, srv2 := "s1", "s2"

	_, err := s.client.Set(s.path("backends", b, "backend"), `{"Type": "http"}`, 0)
	c.Assert(err, IsNil)

	// This one will stay
	_, err = s.client.Set(s.path("backends", b, "servers", srv), fmt.Sprintf(`{"URL": "%s"}`, url), 0)
	c.Assert(err, IsNil)

	// This one will expire
	_, err = s.client.Set(s.path("backends", b, "servers", srv2), fmt.Sprintf(`{"URL": "%s"}`, url2), 2)
	c.Assert(err, IsNil)

	// Add frontend
	fId := "fr1"
	_, err = s.client.Set(s.path("frontends", fId, "frontend"), `{"Type": "http", "BackendId": "bk1", "Route": "Path(\"/path\")"}`, 0)
	c.Assert(err, IsNil)

	time.Sleep(time.Second)
	responses1 := make(map[string]bool)
	for i := 0; i < 3; i += 1 {
		response, body, err := testutils.Get(fmt.Sprintf("%s%s", s.serviceUrl, "/path"))
		c.Assert(err, IsNil)
		c.Assert(response.StatusCode, Equals, http.StatusOK)
		responses1[string(body)] = true
	}
	c.Assert(responses1, DeepEquals, map[string]bool{"e1": true, "e2": true})

	// Now the second endpoint should expire
	time.Sleep(time.Second * 2)

	responses2 := make(map[string]bool)
	for i := 0; i < 3; i += 1 {
		response, body, err := testutils.Get(fmt.Sprintf("%s%s", s.serviceUrl, "/path"))
		c.Assert(err, IsNil)
		c.Assert(response.StatusCode, Equals, http.StatusOK)
		responses2[string(body)] = true
	}
	c.Assert(responses2, DeepEquals, map[string]bool{"e1": true})
}

func (s *VESuite) TestBackendUpdateSettings(c *C) {
	server := testutils.NewHandler(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.Write([]byte("tc: update upstream options"))
	})
	defer server.Close()

	b, srv, url := "bk1", "srv1", server.URL
	_, err := s.client.Set(s.path("backends", b, "backend"), `{"Type": "http", "Settings": {"Timeouts": {"Read":"10ms"}}}`, 0)
	c.Assert(err, IsNil)

	_, err = s.client.Set(s.path("backends", b, "servers", srv), fmt.Sprintf(`{"URL": "%s"}`, url), 0)
	c.Assert(err, IsNil)

	// Add frontend
	fId := "fr1"
	_, err = s.client.Set(s.path("frontends", fId, "frontend"), `{"Type": "http", "BackendId": "bk1", "Route": "Path(\"/path\")"}`, 0)
	c.Assert(err, IsNil)

	// Wait for the changes to take effect
	time.Sleep(time.Second)

	// Make sure request times out
	response, _, err := testutils.Get(fmt.Sprintf("%s%s", s.serviceUrl, "/path"))
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusGatewayTimeout)

	// Update backend timeout
	_, err = s.client.Set(s.path("backends", b, "backend"), `{"Type": "http", "Settings": {"Timeouts": {"Read":"100ms"}}}`, 0)
	c.Assert(err, IsNil)

	// Wait for the changes to take effect
	time.Sleep(time.Second)

	response, body, err := testutils.Get(fmt.Sprintf("%s%s", s.serviceUrl, "/path"))
	c.Assert(err, IsNil)
	c.Assert(response.StatusCode, Equals, http.StatusOK)
	c.Assert(string(body), Equals, "tc: update upstream options")
}

func (s *VESuite) TestLiveBinaryUpgrade(c *C) {
	server := testutils.NewHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello 1"))
	})
	defer server.Close()

	b, srv, url := "bk1", "srv1", server.URL
	_, err := s.client.Set(s.path("backends", b, "backend"), `{"Type": "http"}`, 0)
	c.Assert(err, IsNil)

	_, err = s.client.Set(s.path("backends", b, "servers", srv), fmt.Sprintf(`{"URL": "%s"}`, url), 0)
	c.Assert(err, IsNil)

	// Add frontend
	fId := "fr1"
	_, err = s.client.Set(s.path("frontends", fId, "frontend"), `{"Type": "http", "BackendId": "bk1", "Route": "Path(\"/path\")"}`, 0)
	c.Assert(err, IsNil)

	keyPair := NewTestKeyPair()

	bytes, err := secret.SealKeyPairToJSON(s.box, keyPair)
	c.Assert(err, IsNil)
	sealed := base64.StdEncoding.EncodeToString(bytes)
	host := "localhost"

	_, err = s.client.Set(s.path("hosts", host, "host"), fmt.Sprintf(`{"Name": "localhost", "Settings": {"KeyPair": "%v"}}`, sealed), 0)
	c.Assert(err, IsNil)

	// Add HTTPS listener
	l2 := "ls2"
	listener, err := engine.NewListener(l2, "https", "tcp", "localhost:32000", "", nil)
	c.Assert(err, IsNil)
	bytes, err = json.Marshal(listener)
	c.Assert(err, IsNil)
	s.client.Set(s.path("listeners", l2), string(bytes), 0)

	time.Sleep(time.Second)
	_, body, err := testutils.Get(fmt.Sprintf("%s%s", "https://localhost:32000", "/path"))
	c.Assert(err, IsNil)
	c.Assert(string(body), Equals, "Hello 1")

	pidS, err := exec.Command("pidof", "vulcand").Output()
	c.Assert(err, IsNil)

	// Find a running vulcand
	pid, err := strconv.Atoi(strings.TrimSpace(string(pidS)))
	c.Assert(err, IsNil)

	vulcand, err := os.FindProcess(pid)
	c.Assert(err, IsNil)

	// Ask vulcand to fork a child
	vulcand.Signal(syscall.SIGUSR2)
	time.Sleep(time.Second)

	// Ask parent process to stop
	vulcand.Signal(syscall.SIGTERM)

	// Make sure the child is running
	pid2S, err := exec.Command("pidof", "vulcand").Output()
	c.Assert(err, IsNil)
	c.Assert(string(pid2S), Not(Equals), "")
	c.Assert(string(pid2S), Not(Equals), string(pidS))

	time.Sleep(time.Second)

	// Make sure we are still running and responding
	_, body, err = testutils.Get(fmt.Sprintf("%s%s", "https://localhost:32000", "/path"))
	c.Assert(err, IsNil)
	c.Assert(string(body), Equals, "Hello 1")
}
