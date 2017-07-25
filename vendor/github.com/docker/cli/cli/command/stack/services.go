package stack

import (
	"fmt"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/formatter"
	"github.com/docker/cli/cli/command/service"
	"github.com/docker/cli/opts"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

type servicesOptions struct {
	quiet     bool
	format    string
	filter    opts.FilterOpt
	namespace string
}

func newServicesCommand(dockerCli *command.DockerCli) *cobra.Command {
	options := servicesOptions{filter: opts.NewFilterOpt()}

	cmd := &cobra.Command{
		Use:   "services [OPTIONS] STACK",
		Short: "List the services in the stack",
		Args:  cli.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.namespace = args[0]
			return runServices(dockerCli, options)
		},
	}
	flags := cmd.Flags()
	flags.BoolVarP(&options.quiet, "quiet", "q", false, "Only display IDs")
	flags.StringVar(&options.format, "format", "", "Pretty-print services using a Go template")
	flags.VarP(&options.filter, "filter", "f", "Filter output based on conditions provided")

	return cmd
}

func runServices(dockerCli *command.DockerCli, options servicesOptions) error {
	ctx := context.Background()
	client := dockerCli.Client()

	filter := getStackFilterFromOpt(options.namespace, options.filter)
	services, err := client.ServiceList(ctx, types.ServiceListOptions{Filters: filter})
	if err != nil {
		return err
	}

	out := dockerCli.Out()

	// if no services in this stack, print message and exit 0
	if len(services) == 0 {
		fmt.Fprintf(out, "Nothing found in stack: %s\n", options.namespace)
		return nil
	}

	info := map[string]formatter.ServiceListInfo{}
	if !options.quiet {
		taskFilter := filters.NewArgs()
		for _, service := range services {
			taskFilter.Add("service", service.ID)
		}

		tasks, err := client.TaskList(ctx, types.TaskListOptions{Filters: taskFilter})
		if err != nil {
			return err
		}

		nodes, err := client.NodeList(ctx, types.NodeListOptions{})
		if err != nil {
			return err
		}

		info = service.GetServicesStatus(services, nodes, tasks)
	}

	format := options.format
	if len(format) == 0 {
		if len(dockerCli.ConfigFile().ServicesFormat) > 0 && !options.quiet {
			format = dockerCli.ConfigFile().ServicesFormat
		} else {
			format = formatter.TableFormatKey
		}
	}

	servicesCtx := formatter.Context{
		Output: dockerCli.Out(),
		Format: formatter.NewServiceListFormat(format, options.quiet),
	}
	return formatter.ServiceListWrite(servicesCtx, services, info)
}
