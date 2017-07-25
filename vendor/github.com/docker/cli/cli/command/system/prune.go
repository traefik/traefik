package system

import (
	"fmt"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/prune"
	"github.com/docker/cli/opts"
	units "github.com/docker/go-units"
	"github.com/spf13/cobra"
)

type pruneOptions struct {
	force        bool
	all          bool
	pruneVolumes bool
	filter       opts.FilterOpt
}

// NewPruneCommand creates a new cobra.Command for `docker prune`
func NewPruneCommand(dockerCli command.Cli) *cobra.Command {
	options := pruneOptions{filter: opts.NewFilterOpt()}

	cmd := &cobra.Command{
		Use:   "prune [OPTIONS]",
		Short: "Remove unused data",
		Args:  cli.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPrune(dockerCli, options)
		},
		Tags: map[string]string{"version": "1.25"},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&options.force, "force", "f", false, "Do not prompt for confirmation")
	flags.BoolVarP(&options.all, "all", "a", false, "Remove all unused images not just dangling ones")
	flags.BoolVar(&options.pruneVolumes, "volumes", false, "Prune volumes")
	flags.Var(&options.filter, "filter", "Provide filter values (e.g. 'label=<key>=<value>')")
	// "filter" flag is available in 1.28 (docker 17.04) and up
	flags.SetAnnotation("filter", "version", []string{"1.28"})

	return cmd
}

const (
	warning = `WARNING! This will remove:
	- all stopped containers
	- all volumes not used by at least one container
	- all networks not used by at least one container
	%s
Are you sure you want to continue?`

	danglingImageDesc = "- all dangling images"
	allImageDesc      = `- all images without at least one container associated to them`
)

func runPrune(dockerCli command.Cli, options pruneOptions) error {
	var message string

	if options.all {
		message = fmt.Sprintf(warning, allImageDesc)
	} else {
		message = fmt.Sprintf(warning, danglingImageDesc)
	}

	if !options.force && !command.PromptForConfirmation(dockerCli.In(), dockerCli.Out(), message) {
		return nil
	}

	var spaceReclaimed uint64
	pruneFuncs := []func(dockerCli command.Cli, filter opts.FilterOpt) (uint64, string, error){
		prune.RunContainerPrune,
		prune.RunNetworkPrune,
	}
	if options.pruneVolumes {
		pruneFuncs = append(pruneFuncs, prune.RunVolumePrune)
	}

	for _, pruneFn := range pruneFuncs {
		spc, output, err := pruneFn(dockerCli, options.filter)
		if err != nil {
			return err
		}
		spaceReclaimed += spc
		if output != "" {
			fmt.Fprintln(dockerCli.Out(), output)
		}
	}

	spc, output, err := prune.RunImagePrune(dockerCli, options.all, options.filter)
	if err != nil {
		return err
	}
	if spc > 0 {
		spaceReclaimed += spc
		fmt.Fprintln(dockerCli.Out(), output)
	}

	fmt.Fprintln(dockerCli.Out(), "Total reclaimed space:", units.HumanSize(float64(spaceReclaimed)))

	return nil
}
