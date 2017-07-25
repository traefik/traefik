package secret

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

func newSecretInspectCommand(dockerCli command.Cli) *cobra.Command {
	opts := inspectOptions{}
	cmd := &cobra.Command{
		Use:   "inspect [OPTIONS] SECRET [SECRET...]",
		Short: "Display detailed information on one or more secrets",
		Args:  cli.RequiresMinArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.names = args
			return runSecretInspect(dockerCli, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.format, "format", "f", "", "Format the output using the given Go template")
	cmd.Flags().BoolVar(&opts.pretty, "pretty", false, "Print the information in a human friendly format")
	return cmd
}

func runSecretInspect(dockerCli command.Cli, opts inspectOptions) error {
	client := dockerCli.Client()
	ctx := context.Background()

	if opts.pretty {
		opts.format = "pretty"
	}

	getRef := func(id string) (interface{}, []byte, error) {
		return client.SecretInspectWithRaw(ctx, id)
	}
	f := opts.format

	// check if the user is trying to apply a template to the pretty format, which
	// is not supported
	if strings.HasPrefix(f, "pretty") && f != "pretty" {
		return fmt.Errorf("Cannot supply extra formatting options to the pretty template")
	}

	secretCtx := formatter.Context{
		Output: dockerCli.Out(),
		Format: formatter.NewSecretFormat(f, false),
	}

	if err := formatter.SecretInspectWrite(secretCtx, opts.names, getRef); err != nil {
		return cli.StatusError{StatusCode: 1, Status: err.Error()}
	}
	return nil
}
