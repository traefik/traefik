package container

import (
	"fmt"
	"strings"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

type waitOptions struct {
	containers []string
}

// NewWaitCommand creates a new cobra.Command for `docker wait`
func NewWaitCommand(dockerCli *command.DockerCli) *cobra.Command {
	var opts waitOptions

	cmd := &cobra.Command{
		Use:   "wait CONTAINER [CONTAINER...]",
		Short: "Block until one or more containers stop, then print their exit codes",
		Args:  cli.RequiresMinArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.containers = args
			return runWait(dockerCli, &opts)
		},
	}

	return cmd
}

func runWait(dockerCli *command.DockerCli, opts *waitOptions) error {
	ctx := context.Background()

	var errs []string
	for _, container := range opts.containers {
		resultC, errC := dockerCli.Client().ContainerWait(ctx, container, "")

		select {
		case result := <-resultC:
			fmt.Fprintf(dockerCli.Out(), "%d\n", result.StatusCode)
		case err := <-errC:
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}
