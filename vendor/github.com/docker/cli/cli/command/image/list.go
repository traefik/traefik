package image

import (
	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/formatter"
	"github.com/docker/cli/opts"
	"github.com/docker/docker/api/types"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

type imagesOptions struct {
	matchName string

	quiet       bool
	all         bool
	noTrunc     bool
	showDigests bool
	format      string
	filter      opts.FilterOpt
}

// NewImagesCommand creates a new `docker images` command
func NewImagesCommand(dockerCli command.Cli) *cobra.Command {
	options := imagesOptions{filter: opts.NewFilterOpt()}

	cmd := &cobra.Command{
		Use:   "images [OPTIONS] [REPOSITORY[:TAG]]",
		Short: "List images",
		Args:  cli.RequiresMaxArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				options.matchName = args[0]
			}
			return runImages(dockerCli, options)
		},
	}

	flags := cmd.Flags()

	flags.BoolVarP(&options.quiet, "quiet", "q", false, "Only show numeric IDs")
	flags.BoolVarP(&options.all, "all", "a", false, "Show all images (default hides intermediate images)")
	flags.BoolVar(&options.noTrunc, "no-trunc", false, "Don't truncate output")
	flags.BoolVar(&options.showDigests, "digests", false, "Show digests")
	flags.StringVar(&options.format, "format", "", "Pretty-print images using a Go template")
	flags.VarP(&options.filter, "filter", "f", "Filter output based on conditions provided")

	return cmd
}

func newListCommand(dockerCli command.Cli) *cobra.Command {
	cmd := *NewImagesCommand(dockerCli)
	cmd.Aliases = []string{"images", "list"}
	cmd.Use = "ls [OPTIONS] [REPOSITORY[:TAG]]"
	return &cmd
}

func runImages(dockerCli command.Cli, options imagesOptions) error {
	ctx := context.Background()

	filters := options.filter.Value()
	if options.matchName != "" {
		filters.Add("reference", options.matchName)
	}

	listOptions := types.ImageListOptions{
		All:     options.all,
		Filters: filters,
	}

	images, err := dockerCli.Client().ImageList(ctx, listOptions)
	if err != nil {
		return err
	}

	format := options.format
	if len(format) == 0 {
		if len(dockerCli.ConfigFile().ImagesFormat) > 0 && !options.quiet {
			format = dockerCli.ConfigFile().ImagesFormat
		} else {
			format = formatter.TableFormatKey
		}
	}

	imageCtx := formatter.ImageContext{
		Context: formatter.Context{
			Output: dockerCli.Out(),
			Format: formatter.NewImageFormat(format, options.quiet, options.showDigests),
			Trunc:  !options.noTrunc,
		},
		Digest: options.showDigests,
	}
	return formatter.ImageWrite(imageCtx, images)
}
