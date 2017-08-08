// +build !windows

package integration

import (
	"bufio"
	"io/ioutil"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/containous/traefik/integration/try"
	"github.com/go-check/check"
	checker "github.com/vdemeester/shakers"
)

// Log rotation integration test suite
type LogRotationSuite struct{ BaseSuite }

func (s *LogRotationSuite) TestAccessLogRotation(c *check.C) {
	// Start Traefik
	cmd, _ := s.cmdTraefik(withConfigFile("fixtures/access_log_config.toml"))
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()
	defer os.Remove(traefikTestAccessLogFile)
	defer os.Remove(traefikTestLogFile)

	// Verify Traefik started ok
	verifyEmptyErrorLog(c, "traefik.log")

	// Start test servers
	ts1 := startAccessLogServer(8081)
	defer ts1.Close()

	// Allow time to startup
	time.Sleep(1 * time.Second)

	// Make some requests
	err = try.GetRequest("http://127.0.0.1:8000/test1", 500*time.Millisecond)
	c.Assert(err, checker.IsNil)

	// Rename access log
	err = os.Rename(traefikTestAccessLogFile, traefikTestAccessLogFile+".rotated")
	c.Assert(err, checker.IsNil)

	// in the midst of the requests, issue SIGUSR1 signal to server process
	err = cmd.Process.Signal(syscall.SIGUSR1)
	c.Assert(err, checker.IsNil)

	// continue issuing requests
	_, err = http.Get("http://127.0.0.1:8000/test1")
	c.Assert(err, checker.IsNil)
	_, err = http.Get("http://127.0.0.1:8000/test1")
	c.Assert(err, checker.IsNil)

	// Verify access.log.rotated output as expected
	lineCount := verifyLogLines(c, traefikTestAccessLogFile+".rotated", 0, true)
	c.Assert(lineCount, checker.GreaterOrEqualThan, 1)

	// Verify access.log output as expected
	lineCount = verifyLogLines(c, traefikTestAccessLogFile, lineCount, true)
	c.Assert(lineCount, checker.Equals, 3)

	verifyEmptyErrorLog(c, traefikTestLogFile)
}

func (s *LogRotationSuite) TestTraefikLogRotation(c *check.C) {
	// Start Traefik
	cmd, _ := s.cmdTraefik(withConfigFile("fixtures/traefik_log_config.toml"))
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()
	defer os.Remove(traefikTestAccessLogFile)
	defer os.Remove(traefikTestLogFile)

	// Ensure Traefik has started
	err = try.GetRequest("http://127.0.0.1:8000/test1", 500*time.Millisecond)
	c.Assert(err, checker.IsNil)

	// Rename traefik log
	err = os.Rename(traefikTestLogFile, traefikTestLogFile+".rotated")
	c.Assert(err, checker.IsNil)

	// issue SIGUSR1 signal to server process
	err = cmd.Process.Signal(syscall.SIGUSR1)
	c.Assert(err, checker.IsNil)

	err = cmd.Process.Signal(syscall.SIGTERM)
	c.Assert(err, checker.IsNil)

	// Allow time for switch to be processed
	err = try.Do(3*time.Second, func() error {
		_, err = os.Stat(traefikTestLogFile)
		return err
	})
	c.Assert(err, checker.IsNil)

	// we have at least 6 lines in traefik.log.rotated
	lineCount := verifyLogLines(c, traefikTestLogFile+".rotated", 0, false)

	// GreaterOrEqualThan used to ensure test doesn't break
	// If more log entries are output on startup
	c.Assert(lineCount, checker.GreaterOrEqualThan, 5)

	//Verify traefik.log output as expected
	lineCount = verifyLogLines(c, traefikTestLogFile, lineCount, false)
	c.Assert(lineCount, checker.GreaterOrEqualThan, 7)
}

func verifyEmptyErrorLog(c *check.C, name string) {
	err := try.Do(5*time.Second, func() error {
		traefikLog, e2 := ioutil.ReadFile(name)
		if e2 != nil {
			return e2
		}
		c.Assert(traefikLog, checker.HasLen, 0)
		return nil
	})
	c.Assert(err, checker.IsNil)
}

func verifyLogLines(c *check.C, fileName string, countInit int, accessLog bool) int {
	rotated, err := os.Open(fileName)
	c.Assert(err, checker.IsNil)
	rotatedLog := bufio.NewScanner(rotated)
	count := countInit
	for rotatedLog.Scan() {
		line := rotatedLog.Text()
		c.Log(line)
		if accessLog {
			if len(line) > 0 {
				CheckAccessLogFormat(c, line, count)
			}
		}
		count++
	}

	return count
}
