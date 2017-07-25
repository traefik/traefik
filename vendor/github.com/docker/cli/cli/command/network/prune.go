package network

import (
	"fmt"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/opts"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

type pruneOptions struct {
	force  bool
	filter opts.FilterOpt
}

// NewPruneCommand returns a new cobra prune command for networks
func NewPruneCommand(dockerCli command.Cli) *cobra.Command {
	options := pruneOptions{filter: opts.NewFilterOpt()}

	cmd := &cobra.Command{
		Use:   "prune [OPTIONS]",
		Short: "Remove all unused networks",
		Args:  cli.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := runPrune(dockerCli, options)
			if err != nil {
				return err
			}
			if output != "" {
				fmt.Fprintln(dockerCli.Out(), output)
			}
			return nil
		},
		Tags: map[string]string{"version": "1.25"},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&options.force, "force", "f", false, "Do not prompt for confirmation")
	flags.Var(&options.filter, "filter", "Provide filter values (e.g. 'until=<timestamp>')")

	return cmd
}

const warning = `WARNING! This will remove all networks not used by at least one container.
Are you sure you want to continue?`

func runPrune(dockerCli command.Cli, options pruneOptions) (output string, err error) {
	pruneFilters := command.PruneFilters(dockerCli, options.filter.Value())

	if !options.force && !command.PromptForConfirmation(dockerCli.In(), dockerCli.Out(), warning) {
		return
	}

	report, err := dockerCli.Client().NetworksPrune(context.Background(), pruneFilters)
	if err != nil {
		return
	}

	if len(report.NetworksDeleted) > 0 {
		output = "Deleted Networks:\n"
		for _, id := range report.NetworksDeleted {
			output += id + "\n"
		}
	}

	return
}

// RunPrune calls the Network Prune API
// This returns the amount of space reclaimed and a detailed output string
func RunPrune(dockerCli command.Cli, filter opts.FilterOpt) (uint64, string, error) {
	output, err := runPrune(dockerCli, pruneOptions{force: true, filter: filter})
	return 0, output, err
}
