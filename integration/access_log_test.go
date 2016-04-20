package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/go-check/check"
	shellwords "github.com/mattn/go-shellwords"

	checker "github.com/vdemeester/shakers"
)

// AccessLogSuite
type AccessLogSuite struct{ BaseSuite }

func (s *AccessLogSuite) TestAccessLog(c *check.C) {
	// Ensure working directory is clean
	os.Remove("access.log")
	os.Remove("traefik.log")

	// Start Traefik
	cmd := exec.Command(traefikBinary, "--configFile=fixtures/access_log_config.toml")
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()
	defer os.Remove("access.log")
	defer os.Remove("traefik.log")

	time.Sleep(500 * time.Millisecond)

	// Verify Traefik started OK
	if traefikLog, err := ioutil.ReadFile("traefik.log"); err != nil {
		c.Assert(err.Error(), checker.Equals, "")
	} else if len(traefikLog) > 0 {
		fmt.Printf("%s\n", string(traefikLog))
		c.Assert(len(traefikLog), checker.Equals, 0)
	}

	// Start test servers
	ts1 := startAccessLogServer(8081)
	defer ts1.Close()
	ts2 := startAccessLogServer(8082)
	defer ts2.Close()
	ts3 := startAccessLogServer(8083)
	defer ts3.Close()

	// Make some requests
	_, err = http.Get("http://127.0.0.1:8000/test1")
	c.Assert(err, checker.IsNil)
	_, err = http.Get("http://127.0.0.1:8000/test2")
	c.Assert(err, checker.IsNil)
	_, err = http.Get("http://127.0.0.1:8000/test2")
	c.Assert(err, checker.IsNil)

	// Verify access.log output as expected
	if accessLog, err := ioutil.ReadFile("access.log"); err != nil {
		c.Assert(err.Error(), checker.Equals, "")
	} else {
		lines := strings.Split(string(accessLog), "\n")
		count := 0
		for i, line := range lines {
			if len(line) > 0 {
				count++
				if tokens, err := shellwords.Parse(line); err != nil {
					c.Assert(err.Error(), checker.Equals, "")
				} else {
					c.Assert(len(tokens), checker.Equals, 13)
					c.Assert(tokens[6], checker.Equals, "200")
					c.Assert(tokens[9], checker.Equals, fmt.Sprintf("%d", i+1))
					c.Assert(strings.HasPrefix(tokens[10], "frontend"), checker.True)
					c.Assert(strings.HasPrefix(tokens[11], "http://127.0.0.1:808"), checker.True)
					c.Assert(regexp.MustCompile("^\\d+\\.\\d+.*s$").MatchString(tokens[12]), checker.True)
				}
			}
		}
		c.Assert(count, checker.Equals, 3)
	}

	// Verify no other Trarfik problems
	if traefikLog, err := ioutil.ReadFile("traefik.log"); err != nil {
		c.Assert(err.Error(), checker.Equals, "")
	} else if len(traefikLog) > 0 {
		fmt.Printf("%s\n", string(traefikLog))
		c.Assert(len(traefikLog), checker.Equals, 0)
	}
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
