package stack

import (
	"fmt"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/formatter"
	"github.com/docker/cli/cli/command/idresolver"
	"github.com/docker/cli/cli/command/task"
	"github.com/docker/cli/opts"
	"github.com/docker/docker/api/types"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

type psOptions struct {
	filter    opts.FilterOpt
	noTrunc   bool
	namespace string
	noResolve bool
	quiet     bool
	format    string
}

func newPsCommand(dockerCli command.Cli) *cobra.Command {
	options := psOptions{filter: opts.NewFilterOpt()}

	cmd := &cobra.Command{
		Use:   "ps [OPTIONS] STACK",
		Short: "List the tasks in the stack",
		Args:  cli.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.namespace = args[0]
			return runPS(dockerCli, options)
		},
	}
	flags := cmd.Flags()
	flags.BoolVar(&options.noTrunc, "no-trunc", false, "Do not truncate output")
	flags.BoolVar(&options.noResolve, "no-resolve", false, "Do not map IDs to Names")
	flags.VarP(&options.filter, "filter", "f", "Filter output based on conditions provided")
	flags.BoolVarP(&options.quiet, "quiet", "q", false, "Only display task IDs")
	flags.StringVar(&options.format, "format", "", "Pretty-print tasks using a Go template")

	return cmd
}

func runPS(dockerCli command.Cli, options psOptions) error {
	namespace := options.namespace
	client := dockerCli.Client()
	ctx := context.Background()

	filter := getStackFilterFromOpt(options.namespace, options.filter)

	tasks, err := client.TaskList(ctx, types.TaskListOptions{Filters: filter})
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		fmt.Fprintf(dockerCli.Out(), "Nothing found in stack: %s\n", namespace)
		return nil
	}

	format := options.format
	if len(format) == 0 {
		if len(dockerCli.ConfigFile().TasksFormat) > 0 && !options.quiet {
			format = dockerCli.ConfigFile().TasksFormat
		} else {
			format = formatter.TableFormatKey
		}
	}

	return task.Print(ctx, dockerCli, tasks, idresolver.New(client, options.noResolve), !options.noTrunc, options.quiet, format)
}
