package main

import (
	"bufio"
	"github.com/go-check/check"
	checker "github.com/vdemeester/shakers"
	"net/http"
	"os"
	"os/exec"
)

// AccessLogSuite
type AccessLogJSONSuite struct{ BaseSuite }

func (s *AccessLogJSONSuite) TestAccessLogJSON(c *check.C) {
	// Ensure working directory is clean
	os.Remove("access-log.json")
	os.Remove("traefik.log")

	// Start Traefik
	cmd := exec.Command(traefikBinary, "--configFile=fixtures/access_log_json_config.toml")
	err := cmd.Start()
	c.Assert(err, checker.IsNil)
	defer cmd.Process.Kill()
	defer os.Remove("access-log.json")
	defer os.Remove("traefik.log")

	// Verify Traefik started OK
	verifyEmptyErrorLog(c, "traefik.log")

	// Start test servers
	ts1 := startAccessLogServer(8081)
	defer ts1.Close()
	ts2 := startAccessLogServer(8092)
	defer ts2.Close()
	ts3 := startAccessLogServer(8093)
	defer ts3.Close()

	// Make some requests
	_, err = http.Get("http://127.0.0.1:8000/test1/a")
	c.Assert(err, checker.IsNil)
	_, err = http.Get("http://127.0.0.1:8000/test2/b/c")
	c.Assert(err, checker.IsNil)
	_, err = http.Get("http://127.0.0.1:8000/test2/d/e/f")
	c.Assert(err, checker.IsNil)

	// Verify access.log output as expected
	file, err := os.Open("access-log.json")
	c.Assert(err, checker.IsNil)
	accessLog := bufio.NewScanner(file)
	count := 0
	for accessLog.Scan() {
		line := accessLog.Text()
		c.Log(line)
		if len(line) > 0 {
			count++
			c.Assert(line, checker.Not(checker.Contains), `"request_body_size":"`)
			c.Assert(line, checker.Contains, `"time_utc":"`)
			c.Assert(line, checker.Contains, `"time_local":"`)
			c.Assert(line, checker.Contains, `"total_duration":`)
			c.Assert(line, checker.Contains, `"origin_duration":`)
			c.Assert(line, checker.Contains, `"overhead":`)
			c.Assert(line, checker.Contains, `"addr":"127.0.0.1:8000"`)
			c.Assert(line, checker.Contains, `"host":"127.0.0.1"`)
			c.Assert(line, checker.Contains, `"port":"8000"`)
			c.Assert(line, checker.Contains, `"method":"GET"`)
			c.Assert(line, checker.Contains, `"protocol":"HTTP/1.1"`)
			c.Assert(line, checker.Contains, `"remote_addr":"127.0.0.1:`)
			c.Assert(line, checker.Contains, `"remote_user":"-"`)
			c.Assert(line, checker.Contains, `"remote_ip":"127.0.0.1"`)
			c.Assert(line, checker.Contains, `"remote_port":"`)
			c.Assert(line, checker.Contains, `"http_user_agent":"`)
			c.Assert(line, checker.Contains, `"status":200`)
			c.Assert(line, checker.Contains, `"status_line":"200 OK"`)
			c.Assert(line, checker.Contains, `"origin_status":200`)
			c.Assert(line, checker.Contains, `"origin_status_line":"200 OK"`)
			switch count {
			case 1:
				c.Assert(line, checker.Contains, `"request_path":"/test1/a"`)
				c.Assert(line, checker.Contains, `"request_line":"GET /test1/a HTTP/1.1"`)
				c.Assert(line, checker.Contains, `"frontend":"frontend1"`)
				c.Assert(line, checker.Contains, `"proxy_host":"backend1"`)
				c.Assert(line, checker.Contains, `"proxy_url":"http://127.0.0.1:8081"`)
				c.Assert(line, checker.Contains, `"origin_addr":"127.0.0.1:8081"`)
				c.Assert(line, checker.Contains, `"origin_response_length":24`)
				c.Assert(line, checker.Contains, `"body_bytes_sent":48`)
				c.Assert(line, checker.Contains, `"gzip_ratio":0.5`)
				c.Assert(line, checker.Contains, `"request_count":1`)
			case 2:
				c.Assert(line, checker.Contains, `"request_path":"/test2/b/c"`)
				c.Assert(line, checker.Contains, `"request_line":"GET /test2/b/c HTTP/1.1"`)
				c.Assert(line, checker.Contains, `"frontend":"frontend2"`)
				c.Assert(line, checker.Contains, `"proxy_host":"backend2"`)
				c.Assert(line, checker.Contains, `"proxy_url":"http://127.0.0.1:809`)
				c.Assert(line, checker.Contains, `"origin_addr":"127.0.0.1:809`)
				c.Assert(line, checker.Contains, `"origin_response_length":26`)
				c.Assert(line, checker.Contains, `"body_bytes_sent":50`)
				c.Assert(line, checker.Contains, `"gzip_ratio":0.52`)
				c.Assert(line, checker.Contains, `"request_count":2`)
			case 3:
				c.Assert(line, checker.Contains, `"request_path":"/test2/d/e/f"`)
				c.Assert(line, checker.Contains, `"request_line":"GET /test2/d/e/f HTTP/1.1"`)
				c.Assert(line, checker.Contains, `"frontend":"frontend2"`)
				c.Assert(line, checker.Contains, `"proxy_host":"backend2"`)
				c.Assert(line, checker.Contains, `"proxy_url":"http://127.0.0.1:809`)
				c.Assert(line, checker.Contains, `"origin_addr":"127.0.0.1:809`)
				c.Assert(line, checker.Contains, `"origin_response_length":28`)
				c.Assert(line, checker.Contains, `"body_bytes_sent":52`)
				c.Assert(line, checker.Contains, `"gzip_ratio":0.5385`)
				c.Assert(line, checker.Contains, `"request_count":3`)
			}
		}
	}
	c.Assert(accessLog.Err(), checker.IsNil)
	c.Assert(count, checker.Equals, 3)

	verifyEmptyErrorLog(c, "traefik.log")
}
