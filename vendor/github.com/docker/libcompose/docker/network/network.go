package network

import (
	"fmt"
	"reflect"
	"strings"

	"golang.org/x/net/context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/yaml"
)

// Network holds attributes and method for a network definition in compose
type Network struct {
	client        client.NetworkAPIClient
	name          string
	projectName   string
	driver        string
	driverOptions map[string]string
	ipam          config.Ipam
	external      bool
}

func (n *Network) fullName() string {
	name := n.projectName + "_" + n.name
	if n.external {
		name = n.name
	}
	return name
}

// Inspect inspect the current network
func (n *Network) Inspect(ctx context.Context) (types.NetworkResource, error) {
	return n.client.NetworkInspect(ctx, n.fullName(), types.NetworkInspectOptions{
		Verbose: true,
	})
}

// Remove removes the current network (from docker engine)
func (n *Network) Remove(ctx context.Context) error {
	if n.external {
		fmt.Printf("Network %s is external, skipping", n.fullName())
		return nil
	}
	fmt.Printf("Removing network %q\n", n.fullName())
	return n.client.NetworkRemove(ctx, n.fullName())
}

// EnsureItExists make sure the network exists and return an error if it does not exists
// and cannot be created.
func (n *Network) EnsureItExists(ctx context.Context) error {
	networkResource, err := n.Inspect(ctx)
	if n.external {
		if client.IsErrNotFound(err) {
			// FIXME(vdemeester) introduce some libcompose error type
			return fmt.Errorf("Network %s declared as external, but could not be found. Please create the network manually using docker network create %s and try again", n.fullName(), n.fullName())
		}
		return err
	}
	if err != nil && client.IsErrNotFound(err) {
		return n.create(ctx)
	}
	if n.driver != "" && networkResource.Driver != n.driver {
		return fmt.Errorf("Network %q needs to be recreated - driver has changed", n.fullName())
	}
	if len(n.driverOptions) != 0 && !reflect.DeepEqual(networkResource.Options, n.driverOptions) {
		return fmt.Errorf("Network %q needs to be recreated - options have changed", n.fullName())
	}
	return err
}

func (n *Network) create(ctx context.Context) error {
	fmt.Printf("Creating network %q with driver %q\n", n.fullName(), n.driver)
	_, err := n.client.NetworkCreate(ctx, n.fullName(), types.NetworkCreate{
		Driver:  n.driver,
		Options: n.driverOptions,
		IPAM:    convertToAPIIpam(n.ipam),
	})
	return err
}

func convertToAPIIpam(ipam config.Ipam) *network.IPAM {
	ipamConfigs := []network.IPAMConfig{}
	for _, config := range ipam.Config {
		ipamConfigs = append(ipamConfigs, network.IPAMConfig{
			Subnet:     config.Subnet,
			IPRange:    config.IPRange,
			Gateway:    config.Gateway,
			AuxAddress: config.AuxAddress,
		})
	}
	return &network.IPAM{
		Driver: ipam.Driver,
		Config: ipamConfigs,
	}
}

// NewNetwork creates a new network from the specified name and config.
func NewNetwork(projectName, name string, config *config.NetworkConfig, client client.NetworkAPIClient) *Network {
	networkName := name
	if config.External.External {
		networkName = config.External.Name
	}
	return &Network{
		client:        client,
		name:          networkName,
		projectName:   projectName,
		driver:        config.Driver,
		driverOptions: config.DriverOpts,
		external:      config.External.External,
		ipam:          config.Ipam,
	}
}

// Networks holds a list of network
type Networks struct {
	networks       []*Network
	networkEnabled bool
}

// Initialize make sure network exists if network is enabled
func (n *Networks) Initialize(ctx context.Context) error {
	if !n.networkEnabled {
		return nil
	}
	for _, network := range n.networks {
		err := network.EnsureItExists(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// Remove removes networks (clean-up)
func (n *Networks) Remove(ctx context.Context) error {
	if !n.networkEnabled {
		return nil
	}
	for _, network := range n.networks {
		err := network.Remove(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// NetworksFromServices creates a new Networks struct based on networks configurations and
// services configuration. If a network is defined but not used by any service, it will return
// an error along the Networks.
func NetworksFromServices(cli client.NetworkAPIClient, projectName string, networkConfigs map[string]*config.NetworkConfig, services *config.ServiceConfigs, networkEnabled bool) (*Networks, error) {
	var err error
	networks := make([]*Network, 0, len(networkConfigs))
	networkNames := map[string]*yaml.Network{}
	for _, serviceName := range services.Keys() {
		serviceConfig, _ := services.Get(serviceName)
		if serviceConfig.NetworkMode != "" || serviceConfig.Networks == nil || len(serviceConfig.Networks.Networks) == 0 {
			continue
		}
		for _, network := range serviceConfig.Networks.Networks {
			if network.Name != "default" {
				if _, ok := networkConfigs[network.Name]; !ok {
					return nil, fmt.Errorf(`Service "%s" uses an undefined network "%s"`, serviceName, network.Name)
				}
			}
			networkNames[network.Name] = network
		}
	}
	if len(networkConfigs) == 0 {
		network := NewNetwork(projectName, "default", &config.NetworkConfig{
			Driver: "bridge",
		}, cli)
		networks = append(networks, network)
	}
	for name, config := range networkConfigs {
		network := NewNetwork(projectName, name, config, cli)
		networks = append(networks, network)
	}
	if len(networkNames) != len(networks) {
		unused := []string{}
		for name := range networkConfigs {
			if name == "default" {
				continue
			}
			if _, ok := networkNames[name]; !ok {
				unused = append(unused, name)
			}
		}
		if len(unused) != 0 {
			err = fmt.Errorf("Some networks were defined but are not used by any service: %v", strings.Join(unused, " "))
		}
	}
	return &Networks{
		networks:       networks,
		networkEnabled: networkEnabled,
	}, err
}
