package service

import (
	"strings"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/formatter"
	"github.com/docker/cli/cli/command/idresolver"
	"github.com/docker/cli/cli/command/node"
	"github.com/docker/cli/cli/command/task"
	"github.com/docker/cli/opts"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

type psOptions struct {
	services  []string
	quiet     bool
	noResolve bool
	noTrunc   bool
	format    string
	filter    opts.FilterOpt
}

func newPsCommand(dockerCli command.Cli) *cobra.Command {
	options := psOptions{filter: opts.NewFilterOpt()}

	cmd := &cobra.Command{
		Use:   "ps [OPTIONS] SERVICE [SERVICE...]",
		Short: "List the tasks of one or more services",
		Args:  cli.RequiresMinArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.services = args
			return runPS(dockerCli, options)
		},
	}
	flags := cmd.Flags()
	flags.BoolVarP(&options.quiet, "quiet", "q", false, "Only display task IDs")
	flags.BoolVar(&options.noTrunc, "no-trunc", false, "Do not truncate output")
	flags.BoolVar(&options.noResolve, "no-resolve", false, "Do not map IDs to Names")
	flags.StringVar(&options.format, "format", "", "Pretty-print tasks using a Go template")
	flags.VarP(&options.filter, "filter", "f", "Filter output based on conditions provided")

	return cmd
}

func runPS(dockerCli command.Cli, options psOptions) error {
	client := dockerCli.Client()
	ctx := context.Background()

	filter := options.filter.Value()

	serviceIDFilter := filters.NewArgs()
	serviceNameFilter := filters.NewArgs()
	for _, service := range options.services {
		serviceIDFilter.Add("id", service)
		serviceNameFilter.Add("name", service)
	}
	serviceByIDList, err := client.ServiceList(ctx, types.ServiceListOptions{Filters: serviceIDFilter})
	if err != nil {
		return err
	}
	serviceByNameList, err := client.ServiceList(ctx, types.ServiceListOptions{Filters: serviceNameFilter})
	if err != nil {
		return err
	}

	for _, service := range options.services {
		serviceCount := 0
		// Lookup by ID/Prefix
		for _, serviceEntry := range serviceByIDList {
			if strings.HasPrefix(serviceEntry.ID, service) {
				filter.Add("service", serviceEntry.ID)
				serviceCount++
			}
		}

		// Lookup by Name/Prefix
		for _, serviceEntry := range serviceByNameList {
			if strings.HasPrefix(serviceEntry.Spec.Annotations.Name, service) {
				filter.Add("service", serviceEntry.ID)
				serviceCount++
			}
		}
		// If nothing has been found, return immediately.
		if serviceCount == 0 {
			return errors.Errorf("no such services: %s", service)
		}
	}

	if filter.Include("node") {
		nodeFilters := filter.Get("node")
		for _, nodeFilter := range nodeFilters {
			nodeReference, err := node.Reference(ctx, client, nodeFilter)
			if err != nil {
				return err
			}
			filter.Del("node", nodeFilter)
			filter.Add("node", nodeReference)
		}
	}

	tasks, err := client.TaskList(ctx, types.TaskListOptions{Filters: filter})
	if err != nil {
		return err
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
