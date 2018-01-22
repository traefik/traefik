package client

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Program is an interface to execute external programs.
type Program interface {
	Output() ([]byte, error)
	Input(in io.Reader)
}

// ProgramFunc is a type of function that initializes programs based on arguments.
type ProgramFunc func(args ...string) Program

// NewShellProgramFunc creates programs that are executed in a Shell.
func NewShellProgramFunc(name string) ProgramFunc {
	return NewShellProgramFuncWithEnv(name, nil)
}

// NewShellProgramFuncWithEnv creates programs that are executed in a Shell with environment variables
func NewShellProgramFuncWithEnv(name string, env *map[string]string) ProgramFunc {
	return func(args ...string) Program {
		return &Shell{cmd: createProgramCmdRedirectErr(name, args, env)}
	}
}

func createProgramCmdRedirectErr(commandName string, args []string, env *map[string]string) *exec.Cmd {
	programCmd := exec.Command(commandName, args...)
	programCmd.Env = os.Environ()
	if env != nil {
		for k, v := range *env {
			programCmd.Env = append(programCmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}
	programCmd.Stderr = os.Stderr
	return programCmd
}

// Shell invokes shell commands to talk with a remote credentials helper.
type Shell struct {
	cmd *exec.Cmd
}

// Output returns responses from the remote credentials helper.
func (s *Shell) Output() ([]byte, error) {
	return s.cmd.Output()
}

// Input sets the input to send to a remote credentials helper.
func (s *Shell) Input(in io.Reader) {
	s.cmd.Stdin = in
}
