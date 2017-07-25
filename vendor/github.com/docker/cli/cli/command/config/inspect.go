package config

import (
	"fmt"
	"strings"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/formatter"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

type inspectOptions struct {
	names  []string
	format string
	pretty bool
}

func newConfigInspectCommand(dockerCli command.Cli) *cobra.Command {
	opts := inspectOptions{}
	cmd := &cobra.Command{
		Use:   "inspect [OPTIONS] CONFIG [CONFIG...]",
		Short: "Display detailed information on one or more configuration files",
		Args:  cli.RequiresMinArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.names = args
			return runConfigInspect(dockerCli, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.format, "format", "f", "", "Format the output using the given Go template")
	cmd.Flags().BoolVar(&opts.pretty, "pretty", false, "Print the information in a human friendly format")
	return cmd
}

func runConfigInspect(dockerCli command.Cli, opts inspectOptions) error {
	client := dockerCli.Client()
	ctx := context.Background()

	if opts.pretty {
		opts.format = "pretty"
	}

	getRef := func(id string) (interface{}, []byte, error) {
		return client.ConfigInspectWithRaw(ctx, id)
	}
	f := opts.format

	// check if the user is trying to apply a template to the pretty format, which
	// is not supported
	if strings.HasPrefix(f, "pretty") && f != "pretty" {
		return fmt.Errorf("Cannot supply extra formatting options to the pretty template")
	}

	configCtx := formatter.Context{
		Output: dockerCli.Out(),
		Format: formatter.NewConfigFormat(f, false),
	}

	if err := formatter.ConfigInspectWrite(configCtx, opts.names, getRef); err != nil {
		return cli.StatusError{StatusCode: 1, Status: err.Error()}
	}
	return nil

}
