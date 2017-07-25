package container

import (
	"fmt"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/opts"
	"github.com/docker/docker/api/types"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

type commitOptions struct {
	container string
	reference string

	pause   bool
	comment string
	author  string
	changes opts.ListOpts
}

// NewCommitCommand creates a new cobra.Command for `docker commit`
func NewCommitCommand(dockerCli *command.DockerCli) *cobra.Command {
	var options commitOptions

	cmd := &cobra.Command{
		Use:   "commit [OPTIONS] CONTAINER [REPOSITORY[:TAG]]",
		Short: "Create a new image from a container's changes",
		Args:  cli.RequiresRangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.container = args[0]
			if len(args) > 1 {
				options.reference = args[1]
			}
			return runCommit(dockerCli, &options)
		},
	}

	flags := cmd.Flags()
	flags.SetInterspersed(false)

	flags.BoolVarP(&options.pause, "pause", "p", true, "Pause container during commit")
	flags.StringVarP(&options.comment, "message", "m", "", "Commit message")
	flags.StringVarP(&options.author, "author", "a", "", "Author (e.g., \"John Hannibal Smith <hannibal@a-team.com>\")")

	options.changes = opts.NewListOpts(nil)
	flags.VarP(&options.changes, "change", "c", "Apply Dockerfile instruction to the created image")

	return cmd
}

func runCommit(dockerCli *command.DockerCli, options *commitOptions) error {
	ctx := context.Background()

	name := options.container
	reference := options.reference

	commitOptions := types.ContainerCommitOptions{
		Reference: reference,
		Comment:   options.comment,
		Author:    options.author,
		Changes:   options.changes.GetAll(),
		Pause:     options.pause,
	}

	response, err := dockerCli.Client().ContainerCommit(ctx, name, commitOptions)
	if err != nil {
		return err
	}

	fmt.Fprintln(dockerCli.Out(), response.ID)
	return nil
}
