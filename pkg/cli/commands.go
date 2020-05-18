// Package cli provides tools to create commands that support advanced configuration features,
// sub-commands, and allowing configuration from command-line flags, configuration files, and environment variables.
package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Command structure contains program/command information (command name and description).
type Command struct {
	Name           string
	Description    string
	Configuration  interface{}
	Resources      []ResourceLoader
	Run            func([]string) error
	CustomHelpFunc func(io.Writer, *Command) error
	Hidden         bool
	// AllowArg if not set, disallows any argument that is not a known command or a sub-command.
	AllowArg    bool
	subCommands []*Command
}

// AddCommand Adds a sub command.
func (c *Command) AddCommand(cmd *Command) error {
	if c == nil || cmd == nil {
		return nil
	}

	if c.Name == cmd.Name {
		return fmt.Errorf("child command cannot have the same name as their parent: %s", cmd.Name)
	}

	c.subCommands = append(c.subCommands, cmd)
	return nil
}

// PrintHelp calls the custom help function of the command if it's set.
// Otherwise, it calls the default help function.
func (c *Command) PrintHelp(w io.Writer) error {
	if c.CustomHelpFunc != nil {
		return c.CustomHelpFunc(w, c)
	}
	return PrintHelp(w, c)
}

// Execute Executes a command.
func Execute(cmd *Command) error {
	return execute(cmd, os.Args, true)
}

func execute(cmd *Command, args []string, root bool) error {
	// Calls command without args.
	if len(args) == 1 {
		if err := run(cmd, args[1:]); err != nil {
			return fmt.Errorf("command %s error: %w", args[0], err)
		}
		return nil
	}

	// Special case: if the command is the top level one,
	// and the first arg (`args[1]`) is not the command name or a known sub-command,
	// then we run the top level command itself.
	if root && cmd.Name != args[1] && !contains(cmd.subCommands, args[1]) {
		if err := run(cmd, args[1:]); err != nil {
			return fmt.Errorf("command %s error: %w", filepath.Base(args[0]), err)
		}
		return nil
	}

	// Calls command by its name.
	if len(args) >= 2 && cmd.Name == args[1] {
		if len(args) < 3 || !contains(cmd.subCommands, args[2]) {
			if err := run(cmd, args[2:]); err != nil {
				return fmt.Errorf("command %s error: %w", cmd.Name, err)
			}
			return nil
		}
	}

	// No sub-command, calls the current command.
	if len(cmd.subCommands) == 0 {
		if err := run(cmd, args[1:]); err != nil {
			return fmt.Errorf("command %s error: %w", cmd.Name, err)
		}
		return nil
	}

	// Trying to find the sub-command.
	for _, subCmd := range cmd.subCommands {
		if len(args) >= 2 && subCmd.Name == args[1] {
			return execute(subCmd, args, false)
		}
		if len(args) >= 3 && subCmd.Name == args[2] {
			return execute(subCmd, args[1:], false)
		}
	}

	return fmt.Errorf("command not found: %v", args)
}

func run(cmd *Command, args []string) error {
	if len(args) > 0 && !isFlag(args[0]) && !cmd.AllowArg {
		_ = cmd.PrintHelp(os.Stdout)
		return fmt.Errorf("command not found: %s", args[0])
	}

	if isHelp(args) {
		return cmd.PrintHelp(os.Stdout)
	}

	if cmd.Run == nil {
		_ = cmd.PrintHelp(os.Stdout)
		return fmt.Errorf("command %s is not runnable", cmd.Name)
	}

	if cmd.Configuration == nil {
		return cmd.Run(args)
	}

	for _, resource := range cmd.Resources {
		done, err := resource.Load(args, cmd)
		if err != nil {
			return err
		}
		if done {
			break
		}
	}

	return cmd.Run(args)
}

func contains(cmds []*Command, name string) bool {
	for _, cmd := range cmds {
		if cmd.Name == name {
			return true
		}
	}

	return false
}

func isFlag(arg string) bool {
	return len(arg) > 0 && arg[0] == '-'
}
