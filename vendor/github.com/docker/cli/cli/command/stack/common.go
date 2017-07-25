package stack

import (
	"github.com/docker/cli/cli/compose/convert"
	"github.com/docker/cli/opts"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

func getStackFilter(namespace string) filters.Args {
	filter := filters.NewArgs()
	filter.Add("label", convert.LabelNamespace+"="+namespace)
	return filter
}

func getServiceFilter(namespace string) filters.Args {
	return getStackFilter(namespace)
}

func getStackFilterFromOpt(namespace string, opt opts.FilterOpt) filters.Args {
	filter := opt.Value()
	filter.Add("label", convert.LabelNamespace+"="+namespace)
	return filter
}

func getAllStacksFilter() filters.Args {
	filter := filters.NewArgs()
	filter.Add("label", convert.LabelNamespace)
	return filter
}

func getServices(
	ctx context.Context,
	apiclient client.APIClient,
	namespace string,
) ([]swarm.Service, error) {
	return apiclient.ServiceList(
		ctx,
		types.ServiceListOptions{Filters: getServiceFilter(namespace)})
}

func getStackNetworks(
	ctx context.Context,
	apiclient client.APIClient,
	namespace string,
) ([]types.NetworkResource, error) {
	return apiclient.NetworkList(
		ctx,
		types.NetworkListOptions{Filters: getStackFilter(namespace)})
}

func getStackSecrets(
	ctx context.Context,
	apiclient client.APIClient,
	namespace string,
) ([]swarm.Secret, error) {
	return apiclient.SecretList(
		ctx,
		types.SecretListOptions{Filters: getStackFilter(namespace)})
}

func getStackConfigs(
	ctx context.Context,
	apiclient client.APIClient,
	namespace string,
) ([]swarm.Config, error) {
	return apiclient.ConfigList(
		ctx,
		types.ConfigListOptions{Filters: getStackFilter(namespace)})
}
