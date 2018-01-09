package image

import (
	"io"
	"os"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	dockeropts "github.com/docker/cli/opts"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/urlutil"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

type importOptions struct {
	source    string
	reference string
	changes   dockeropts.ListOpts
	message   string
}

// NewImportCommand creates a new `docker import` command
func NewImportCommand(dockerCli command.Cli) *cobra.Command {
	var options importOptions

	cmd := &cobra.Command{
		Use:   "import [OPTIONS] file|URL|- [REPOSITORY[:TAG]]",
		Short: "Import the contents from a tarball to create a filesystem image",
		Args:  cli.RequiresMinArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.source = args[0]
			if len(args) > 1 {
				options.reference = args[1]
			}
			return runImport(dockerCli, options)
		},
	}

	flags := cmd.Flags()

	options.changes = dockeropts.NewListOpts(nil)
	flags.VarP(&options.changes, "change", "c", "Apply Dockerfile instruction to the created image")
	flags.StringVarP(&options.message, "message", "m", "", "Set commit message for imported image")

	return cmd
}

func runImport(dockerCli command.Cli, options importOptions) error {
	var (
		in      io.Reader
		srcName = options.source
	)

	if options.source == "-" {
		in = dockerCli.In()
	} else if !urlutil.IsURL(options.source) {
		srcName = "-"
		file, err := os.Open(options.source)
		if err != nil {
			return err
		}
		defer file.Close()
		in = file
	}

	source := types.ImageImportSource{
		Source:     in,
		SourceName: srcName,
	}

	importOptions := types.ImageImportOptions{
		Message: options.message,
		Changes: options.changes.GetAll(),
	}

	clnt := dockerCli.Client()

	responseBody, err := clnt.ImageImport(context.Background(), source, options.reference, importOptions)
	if err != nil {
		return err
	}
	defer responseBody.Close()

	return jsonmessage.DisplayJSONMessagesToStream(responseBody, dockerCli.Out(), nil)
}
