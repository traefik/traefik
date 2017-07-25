package checkpoint

import (
	"golang.org/x/net/context"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/formatter"
	"github.com/docker/docker/api/types"
	"github.com/spf13/cobra"
)

type listOptions struct {
	checkpointDir string
}

func newListCommand(dockerCli command.Cli) *cobra.Command {
	var opts listOptions

	cmd := &cobra.Command{
		Use:     "ls [OPTIONS] CONTAINER",
		Aliases: []string{"list"},
		Short:   "List checkpoints for a container",
		Args:    cli.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(dockerCli, args[0], opts)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.checkpointDir, "checkpoint-dir", "", "", "Use a custom checkpoint storage directory")

	return cmd

}

func runList(dockerCli command.Cli, container string, opts listOptions) error {
	client := dockerCli.Client()

	listOpts := types.CheckpointListOptions{
		CheckpointDir: opts.checkpointDir,
	}

	checkpoints, err := client.CheckpointList(context.Background(), container, listOpts)
	if err != nil {
		return err
	}

	cpCtx := formatter.Context{
		Output: dockerCli.Out(),
		Format: formatter.NewCheckpointFormat(formatter.TableFormatKey),
	}
	return formatter.CheckpointWrite(cpCtx, checkpoints)
}
