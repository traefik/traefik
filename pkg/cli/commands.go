// Package cli provides tools to create commands that support advanced configuration features,
// sub-commands, and allowing configuration from command-line flags, configuration files, and environment variables.
package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Command structure contains program/command information (command name and description).
type Command struct {
	Name          string
	Description   string
	Configuration interface{}
	Resources     []ResourceLoader
	Run           func([]string) error
	Hidden        bool
	subCommands   []*Command
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

// Execute Executes a command.
func Execute(cmd *Command) error {
	return execute(cmd, os.Args, true)
}

func execute(cmd *Command, args []string, root bool) error {
	if len(args) == 1 {
		if err := run(cmd, args); err != nil {
			return fmt.Errorf("command %s error: %v", args[0], err)
		}
		return nil
	}

	if root && cmd.Name != args[1] && !contains(cmd.subCommands, args[1]) {
		if err := run(cmd, args[1:]); err != nil {
			return fmt.Errorf("command %s error: %v", filepath.Base(args[0]), err)
		}
		return nil
	}

	if len(args) >= 2 && cmd.Name == args[1] {
		if err := run(cmd, args[2:]); err != nil {
			return fmt.Errorf("command %s error: %v", cmd.Name, err)
		}
		return nil
	}

	if len(cmd.subCommands) == 0 {
		if err := run(cmd, args[1:]); err != nil {
			return fmt.Errorf("command %s error: %v", cmd.Name, err)
		}
		return nil
	}

	for _, subCmd := range cmd.subCommands {
		if len(args) >= 2 && subCmd.Name == args[1] {
			return execute(subCmd, args[1:], false)
		}
	}

	return fmt.Errorf("command not found: %v", args)
}

func run(cmd *Command, args []string) error {
	if isHelp(args) {
		return PrintHelp(os.Stdout, cmd)
	}

	if cmd.Run == nil {
		_ = PrintHelp(os.Stdout, cmd)
		return errors.New("command not found")
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
