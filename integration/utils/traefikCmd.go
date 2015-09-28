package utils

import (
	"os/exec"
)

var execCommand = exec.Command

// RunCommand runs the specified command with arguments and returns
// the output and the error if any.
func RunCommand(binary string, args ...string) (*exec.Cmd, string, error) {
	cmd := execCommand(binary, args...)
	out, err := cmd.CombinedOutput()
	output := string(out)
	return cmd, output, err
}
