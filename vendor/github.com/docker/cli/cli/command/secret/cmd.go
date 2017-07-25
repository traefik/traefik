package secret

import (
	"github.com/spf13/cobra"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
)

// NewSecretCommand returns a cobra command for `secret` subcommands
// nolint: interfacer
func NewSecretCommand(dockerCli *command.DockerCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret",
		Short: "Manage Docker secrets",
		Args:  cli.NoArgs,
		RunE:  command.ShowHelp(dockerCli.Err()),
		Tags:  map[string]string{"version": "1.25"},
	}
	cmd.AddCommand(
		newSecretListCommand(dockerCli),
		newSecretCreateCommand(dockerCli),
		newSecretInspectCommand(dockerCli),
		newSecretRemoveCommand(dockerCli),
	)
	return cmd
}
