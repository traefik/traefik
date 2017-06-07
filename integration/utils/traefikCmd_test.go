package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

var traefikBinary = "traefik"

func TestRunCommand(t *testing.T) {
	// Override exec.Command :D
	execCommand = fakeExecCommand
	_, output, err := RunCommand(traefikBinary, "it", "works")
	if err != nil {
		t.Fatal(err)
	}
	if output != "it works" {
		t.Fatalf("Expected 'it works' as output, got : %q", output)
	}
}

func TestRunCommandError(t *testing.T) {
	// Override exec.Command :D
	execCommand = fakeExecCommand
	_, output, err := RunCommand(traefikBinary, "an", "error")
	if err == nil {
		t.Fatalf("Expected an error, got %q", output)
	}
}

// Helpers :)

// Type implementing the io.Writer interface for analyzing output.
type String struct {
	value string
}

// The only function required by the io.Writer interface.  Will append
// written data to the String.value string.
func (s *String) Write(p []byte) (n int, err error) {
	s.value += string(p)
	return len(p), nil
}

// Helper function that mock the exec.Command call (and call the test binary)
func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	args := os.Args

	// Previous arguments are tests stuff, that looks like :
	// /tmp/go-build970079519/…/_test/integration.test -test.run=TestHelperProcess --
	cmd, args := args[3], args[4:]
	// Handle the case where args[0] is dir:...

	switch cmd {
	case traefikBinary:
		argsStr := strings.Join(args, " ")
		switch argsStr {
		case "an exitCode 127":
			fmt.Fprint(os.Stderr, "an error has occurred with exitCode 127")
			os.Exit(127)
		case "an error":
			fmt.Fprint(os.Stderr, "an error has occurred")
			os.Exit(1)
		case "it works":
			fmt.Fprint(os.Stdout, "it works")
		default:
			fmt.Fprint(os.Stdout, "no arguments")
		}
	default:
		fmt.Fprintf(os.Stderr, "Command %s not found.", cmd)
		os.Exit(1)
	}
	// some code here to check arguments perhaps?
	os.Exit(0)
}
