package volume

import (
	"sort"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/formatter"
	"github.com/docker/cli/opts"
	"github.com/docker/docker/api/types"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

type byVolumeName []*types.Volume

func (r byVolumeName) Len() int      { return len(r) }
func (r byVolumeName) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r byVolumeName) Less(i, j int) bool {
	return r[i].Name < r[j].Name
}

type listOptions struct {
	quiet  bool
	format string
	filter opts.FilterOpt
}

func newListCommand(dockerCli command.Cli) *cobra.Command {
	options := listOptions{filter: opts.NewFilterOpt()}

	cmd := &cobra.Command{
		Use:     "ls [OPTIONS]",
		Aliases: []string{"list"},
		Short:   "List volumes",
		Args:    cli.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(dockerCli, options)
		},
	}

	flags := cmd.Flags()
	flags.BoolVarP(&options.quiet, "quiet", "q", false, "Only display volume names")
	flags.StringVar(&options.format, "format", "", "Pretty-print volumes using a Go template")
	flags.VarP(&options.filter, "filter", "f", "Provide filter values (e.g. 'dangling=true')")

	return cmd
}

func runList(dockerCli command.Cli, options listOptions) error {
	client := dockerCli.Client()
	volumes, err := client.VolumeList(context.Background(), options.filter.Value())
	if err != nil {
		return err
	}

	format := options.format
	if len(format) == 0 {
		if len(dockerCli.ConfigFile().VolumesFormat) > 0 && !options.quiet {
			format = dockerCli.ConfigFile().VolumesFormat
		} else {
			format = formatter.TableFormatKey
		}
	}

	sort.Sort(byVolumeName(volumes.Volumes))

	volumeCtx := formatter.Context{
		Output: dockerCli.Out(),
		Format: formatter.NewVolumeFormat(format, options.quiet),
	}
	return formatter.VolumeWrite(volumeCtx, volumes.Volumes)
}
