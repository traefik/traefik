package network

import (
	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/opts"
	"github.com/docker/docker/api/types/network"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

type connectOptions struct {
	network      string
	container    string
	ipaddress    string
	ipv6address  string
	links        opts.ListOpts
	aliases      []string
	linklocalips []string
}

func newConnectCommand(dockerCli *command.DockerCli) *cobra.Command {
	options := connectOptions{
		links: opts.NewListOpts(opts.ValidateLink),
	}

	cmd := &cobra.Command{
		Use:   "connect [OPTIONS] NETWORK CONTAINER",
		Short: "Connect a container to a network",
		Args:  cli.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.network = args[0]
			options.container = args[1]
			return runConnect(dockerCli, options)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&options.ipaddress, "ip", "", "IPv4 address (e.g., 172.30.100.104)")
	flags.StringVar(&options.ipv6address, "ip6", "", "IPv6 address (e.g., 2001:db8::33)")
	flags.Var(&options.links, "link", "Add link to another container")
	flags.StringSliceVar(&options.aliases, "alias", []string{}, "Add network-scoped alias for the container")
	flags.StringSliceVar(&options.linklocalips, "link-local-ip", []string{}, "Add a link-local address for the container")

	return cmd
}

func runConnect(dockerCli *command.DockerCli, options connectOptions) error {
	client := dockerCli.Client()

	epConfig := &network.EndpointSettings{
		IPAMConfig: &network.EndpointIPAMConfig{
			IPv4Address:  options.ipaddress,
			IPv6Address:  options.ipv6address,
			LinkLocalIPs: options.linklocalips,
		},
		Links:   options.links.GetAll(),
		Aliases: options.aliases,
	}

	return client.NetworkConnect(context.Background(), options.network, options.container, epConfig)
}
