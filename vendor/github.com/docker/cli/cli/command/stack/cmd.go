package stack

import (
	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/spf13/cobra"
)

// NewStackCommand returns a cobra command for `stack` subcommands
// nolint: interfacer
func NewStackCommand(dockerCli *command.DockerCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stack",
		Short: "Manage Docker stacks",
		Args:  cli.NoArgs,
		RunE:  command.ShowHelp(dockerCli.Err()),
		Tags:  map[string]string{"version": "1.25"},
	}
	cmd.AddCommand(
		newDeployCommand(dockerCli),
		newListCommand(dockerCli),
		newRemoveCommand(dockerCli),
		newServicesCommand(dockerCli),
		newPsCommand(dockerCli),
	)
	return cmd
}

// NewTopLevelDeployCommand returns a command for `docker deploy`
func NewTopLevelDeployCommand(dockerCli command.Cli) *cobra.Command {
	cmd := newDeployCommand(dockerCli)
	// Remove the aliases at the top level
	cmd.Aliases = []string{}
	cmd.Tags = map[string]string{"experimental": "", "version": "1.25"}
	return cmd
}
