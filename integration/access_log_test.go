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

	"bufio"
	"github.com/containous/traefik/integration/utils"
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

	// Verify Traefik started OK
	verifyEmptyErrorLog(c, "traefik.log")

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
	file, err := os.Open("access.log")
	c.Assert(err, checker.IsNil)
	accessLog := bufio.NewScanner(file)
	count := 0
	for accessLog.Scan() {
		line := accessLog.Text()
		c.Log(line)
		count++
		if len(line) > 0 {
			tokens, err := shellwords.Parse(line)
			c.Assert(err, checker.IsNil)
			c.Assert(len(tokens), checker.Equals, 13) // not 14 because 'referer' is blank
			c.Assert(tokens[6], checker.Equals, "200")
			c.Assert(tokens[9], checker.Equals, fmt.Sprintf("%d", count))
			c.Assert(strings.HasPrefix(tokens[10], "frontend"), checker.True)
			c.Assert(strings.HasPrefix(tokens[11], "http://127.0.0.1:808"), checker.True)
			c.Assert(regexp.MustCompile("^\\d+ms$").MatchString(tokens[12]), checker.True)
		}
	}
	c.Assert(count, checker.Equals, 3)

	verifyEmptyErrorLog(c, "traefik.log")
}

func verifyEmptyErrorLog(c *check.C, name string) {
	err := utils.Try(30*time.Second, func() error {
		traefikLog, e2 := ioutil.ReadFile(name)
		if e2 != nil {
			return e2
		}
		if len(traefikLog) > 0 {
			fmt.Printf("%s\n", string(traefikLog))
			c.Assert(len(traefikLog), checker.Equals, 0)
		}
		return nil
	})
	c.Assert(err, checker.IsNil)
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
