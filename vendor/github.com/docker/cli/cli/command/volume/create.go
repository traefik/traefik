package volume

import (
	"fmt"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/opts"
	volumetypes "github.com/docker/docker/api/types/volume"
	runconfigopts "github.com/docker/docker/runconfig/opts"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

type createOptions struct {
	name       string
	driver     string
	driverOpts opts.MapOpts
	labels     opts.ListOpts
}

func newCreateCommand(dockerCli command.Cli) *cobra.Command {
	options := createOptions{
		driverOpts: *opts.NewMapOpts(nil, nil),
		labels:     opts.NewListOpts(opts.ValidateEnv),
	}

	cmd := &cobra.Command{
		Use:   "create [OPTIONS] [VOLUME]",
		Short: "Create a volume",
		Args:  cli.RequiresMaxArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				if options.name != "" {
					return errors.Errorf("Conflicting options: either specify --name or provide positional arg, not both\n")
				}
				options.name = args[0]
			}
			return runCreate(dockerCli, options)
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&options.driver, "driver", "d", "local", "Specify volume driver name")
	flags.StringVar(&options.name, "name", "", "Specify volume name")
	flags.Lookup("name").Hidden = true
	flags.VarP(&options.driverOpts, "opt", "o", "Set driver specific options")
	flags.Var(&options.labels, "label", "Set metadata for a volume")

	return cmd
}

func runCreate(dockerCli command.Cli, options createOptions) error {
	client := dockerCli.Client()

	volReq := volumetypes.VolumesCreateBody{
		Driver:     options.driver,
		DriverOpts: options.driverOpts.GetAll(),
		Name:       options.name,
		Labels:     runconfigopts.ConvertKVStringsToMap(options.labels.GetAll()),
	}

	vol, err := client.VolumeCreate(context.Background(), volReq)
	if err != nil {
		return err
	}

	fmt.Fprintf(dockerCli.Out(), "%s\n", vol.Name)
	return nil
}
