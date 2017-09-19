package integration

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/go-check/check"
	"github.com/mattn/go-shellwords"
	checker "github.com/vdemeester/shakers"
)

const (
	traefikTestLogFile       = "traefik.log"
	traefikTestAccessLogFile = "access.log"
)

// AccessLogSuite
type AccessLogSuite struct{ BaseSuite }

func (s *AccessLogSuite) TestAccessLog(c *check.C) {
	// Ensure working directory is clean
	os.Remove(traefikTestAccessLogFile)
	os.Remove(traefikTestLogFile)

	// Start Traefik
	cmd, _ := s.cmdTraefik(withConfigFile("fixtures/access_log_config.toml"))
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()

	defer os.Remove(traefikTestAccessLogFile)
	defer os.Remove(traefikTestLogFile)

	err = try.Do(1*time.Second, func() error {
		if _, err := os.Stat(traefikTestLogFile); err != nil {
			return fmt.Errorf("could not get stats for log file: %s", err)
		}
		return nil
	})
	c.Assert(err, checker.IsNil)

	// Verify Traefik started OK
	traefikLog, err := ioutil.ReadFile(traefikTestLogFile)
	c.Assert(err, checker.IsNil)
	if len(traefikLog) > 0 {
		fmt.Printf("%s\n", string(traefikLog))
		c.Assert(traefikLog, checker.HasLen, 0)
	}

	// Start test servers
	ts1 := startAccessLogServer(8081)
	defer ts1.Close()
	ts2 := startAccessLogServer(8082)
	defer ts2.Close()
	ts3 := startAccessLogServer(8083)
	defer ts3.Close()

	// Make some requests
	err = try.GetRequest("http://127.0.0.1:8000/test1", 500*time.Millisecond)
	c.Assert(err, checker.IsNil)
	err = try.GetRequest("http://127.0.0.1:8000/test2", 500*time.Millisecond)
	c.Assert(err, checker.IsNil)
	err = try.GetRequest("http://127.0.0.1:8000/test2", 500*time.Millisecond)
	c.Assert(err, checker.IsNil)

	// Verify access.log output as expected
	accessLog, err := ioutil.ReadFile(traefikTestAccessLogFile)
	c.Assert(err, checker.IsNil)
	lines := strings.Split(string(accessLog), "\n")
	count := 0
	for i, line := range lines {
		if len(line) > 0 {
			count++
			CheckAccessLogFormat(c, line, i)
		}
	}
	c.Assert(count, checker.GreaterOrEqualThan, 3)

	// Verify no other Traefik problems
	traefikLog, err = ioutil.ReadFile(traefikTestLogFile)
	c.Assert(err, checker.IsNil)
	if len(traefikLog) > 0 {
		fmt.Printf("%s\n", string(traefikLog))
		c.Assert(traefikLog, checker.HasLen, 0)
	}
}

func CheckAccessLogFormat(c *check.C, line string, i int) {
	tokens, err := shellwords.Parse(line)
	c.Assert(err, checker.IsNil)
	c.Assert(tokens, checker.HasLen, 14)
	c.Assert(tokens[6], checker.Matches, `^\d{3}$`)
	c.Assert(tokens[10], checker.Equals, fmt.Sprintf("%d", i+1))
	c.Assert(tokens[11], checker.HasPrefix, "frontend")
	c.Assert(tokens[12], checker.HasPrefix, "http://127.0.0.1:808")
	c.Assert(tokens[13], checker.Matches, `^\d+ms$`)
}

func startAccessLogServer(port int) (ts *httptest.Server) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Received query %s!\n", r.URL.Path[1:])
	})
	if listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port)); err != nil {
		panic(err)
	} else {
		ts = &httptest.Server{
			Listener: listener,
			Config:   &http.Server{Handler: handler},
		}
		ts.Start()
	}
	return
}
