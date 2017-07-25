package integration

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	. "gopkg.in/check.v1"

	"github.com/kr/pty"
)

// FIXME find out why it fails with "inappropriate ioctl for device"
func (s *CliSuite) TestRun(c *C) {
	p := s.RandomProject()
	cmd := exec.Command(s.command, "-f", "./assets/run/docker-compose.yml", "-p", p, "run", "hello", "echo", "test")
	var b bytes.Buffer
	wbuf := bufio.NewWriter(&b)

	tty, err := pty.Start(cmd)

	_, err = io.Copy(wbuf, tty)
	if e, ok := err.(*os.PathError); ok && e.Err == syscall.EIO {
		// We can safely ignore this error, because it's just
		// the PTY telling us that it closed successfully.
		// See:
		// https://github.com/buildkite/agent/pull/34#issuecomment-46080419
		err = nil
	}
	c.Assert(cmd.Wait(), IsNil)
	output := string(b.Bytes())

	c.Assert(err, IsNil, Commentf("%s", output))

	name := fmt.Sprintf("%s_%s_run_1", p, "hello")
	cn := s.GetContainerByName(c, name)
	c.Assert(cn, NotNil)

	lines := strings.Split(output, "\r\n")
	lastLine := lines[len(lines)-2 : len(lines)-1][0]

	c.Assert(cn.State.Running, Equals, false)
	c.Assert(lastLine, Equals, "test\r")
}
