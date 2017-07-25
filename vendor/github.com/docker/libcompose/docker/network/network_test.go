package network

import (
	"fmt"
	"testing"

	"golang.org/x/net/context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/yaml"
	"github.com/pkg/errors"
)

type networkNotFound struct {
	network string
}

func (e networkNotFound) Error() string {
	return fmt.Sprintf("network %s not found", e.network)
}

func (e networkNotFound) NotFound() bool {
	return true
}

func TestNetworksFromServices(t *testing.T) {
	cases := []struct {
		networkConfigs   map[string]*config.NetworkConfig
		services         map[string]*config.ServiceConfig
		networkEnabled   bool
		expectedNetworks []*Network
		expectedError    bool
	}{
		{
			expectedNetworks: []*Network{
				{
					name:        "default",
					projectName: "prj",
				},
			},
		},
		{
			networkConfigs: map[string]*config.NetworkConfig{
				"net1": {},
			},
			services: map[string]*config.ServiceConfig{},
			expectedNetworks: []*Network{
				{
					name:        "net1",
					projectName: "prj",
				},
			},
			expectedError: true,
		},
		{
			networkConfigs: map[string]*config.NetworkConfig{
				"net1": {},
				"net2": {},
			},
			services: map[string]*config.ServiceConfig{
				"svc1": {
					Networks: &yaml.Networks{
						Networks: []*yaml.Network{
							{
								Name: "net1",
							},
						},
					},
				},
			},
			expectedNetworks: []*Network{
				{
					name:        "default",
					projectName: "prj",
				},
				{
					name:        "net1",
					projectName: "prj",
				},
				{
					name:        "net2",
					projectName: "prj",
				},
			},
			expectedError: true,
		},
		{
			networkConfigs: map[string]*config.NetworkConfig{
				"net1": {},
				"net2": {},
			},
			services: map[string]*config.ServiceConfig{
				"svc1": {
					Networks: &yaml.Networks{
						Networks: []*yaml.Network{
							{
								Name: "net1",
							},
						},
					},
				},
				"svc2": {
					Networks: &yaml.Networks{
						Networks: []*yaml.Network{
							{
								Name: "net1",
							},
							{
								Name: "net2",
							},
						},
					},
				},
			},
			expectedNetworks: []*Network{
				{
					name:        "net1",
					projectName: "prj",
				},
				{
					name:        "net2",
					projectName: "prj",
				},
			},
			expectedError: false,
		},
		{
			networkConfigs: map[string]*config.NetworkConfig{
				"net1": {},
				"net2": {},
			},
			services: map[string]*config.ServiceConfig{
				"svc1": {
					Networks: &yaml.Networks{
						Networks: []*yaml.Network{
							{
								Name: "net1",
							},
						},
					},
				},
				"svc2": {
					Networks: &yaml.Networks{
						Networks: []*yaml.Network{
							{
								Name: "net1",
							},
							{
								Name: "net2",
							},
						},
					},
				},
				"svc3": {
					NetworkMode: "host",
					Networks: &yaml.Networks{
						Networks: []*yaml.Network{
							{
								Name: "net3",
							},
						},
					},
				},
			},
			expectedNetworks: []*Network{
				{
					name:        "net1",
					projectName: "prj",
				},
				{
					name:        "net2",
					projectName: "prj",
				},
			},
			expectedError: false,
		},
	}
	for index, c := range cases {
		services := config.NewServiceConfigs()
		for name, service := range c.services {
			services.Add(name, service)
		}
		networks, err := NetworksFromServices(&networkClient{}, "prj", c.networkConfigs, services, c.networkEnabled)
		if c.expectedError {
			if err == nil {
				t.Fatalf("%d: expected an error, got nothing", index)
			}
		} else {
			if err != nil {
				t.Fatalf("%d: didn't expect an error, got one %s", index, err.Error())
			}
			if networks.networkEnabled != c.networkEnabled {
				t.Fatalf("%d: expected network enabled %v, got %v", index, c.networkEnabled, networks.networkEnabled)
			}
			if len(networks.networks) != len(c.expectedNetworks) {
				t.Fatalf("%d: expected %v, got %v", index, c.expectedNetworks, networks.networks)
			}
			for _, network := range networks.networks {
				testExpectedContainsNetwork(t, index, c.expectedNetworks, network)
			}
		}
	}
}

func testExpectedContainsNetwork(t *testing.T, index int, expected []*Network, network *Network) {
	found := false
	for _, e := range expected {
		if e.name == network.name && e.projectName == network.projectName {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("%d: network %v not found in %v", index, network, expected)
	}
}

type networkClient struct {
	client.Client
	expectedNetworkCreate   types.NetworkCreate
	expectedRemoveNetworkID string
	expectedName            string
	inspectError            error
	inspectNetworkDriver    string
	inspectNetworkOptions   map[string]string
	removeError             error
}

func (c *networkClient) NetworkInspect(ctx context.Context, networkID string, verbose bool) (types.NetworkResource, error) {
	if c.inspectError != nil {
		return types.NetworkResource{}, c.inspectError
	}
	return types.NetworkResource{
		ID:      "network_id",
		Driver:  c.inspectNetworkDriver,
		Options: c.inspectNetworkOptions,
	}, nil
}

func (c *networkClient) NetworkCreate(ctx context.Context, name string, options types.NetworkCreate) (types.NetworkCreateResponse, error) {
	if c.expectedName != "" {
		if options.Driver != c.expectedNetworkCreate.Driver {
			return types.NetworkCreateResponse{}, fmt.Errorf("Invalid network create, expected driver %q, got %q", c.expectedNetworkCreate.Driver, options.Driver)
		}
		if c.expectedNetworkCreate.IPAM != nil && options.IPAM.Driver != c.expectedNetworkCreate.IPAM.Driver {
			return types.NetworkCreateResponse{}, fmt.Errorf("Invalid network create, expected ipam %q, got %q", c.expectedNetworkCreate.IPAM, options.IPAM)
		}
		if c.expectedNetworkCreate.IPAM != nil && len(options.IPAM.Config) != len(c.expectedNetworkCreate.IPAM.Config) {
			return types.NetworkCreateResponse{}, fmt.Errorf("Invalid network create, expected ipam %q, got %q", c.expectedNetworkCreate.Driver, options.Driver)
		}
		return types.NetworkCreateResponse{
			ID: c.expectedName,
		}, nil
	}
	return types.NetworkCreateResponse{}, errors.New("Engine no longer exists")
}

func (c *networkClient) NetworkRemove(ctx context.Context, networkID string) error {
	if c.expectedRemoveNetworkID != "" {
		if networkID != c.expectedRemoveNetworkID {
			return fmt.Errorf("Invalid network id for removing, expected %q, got %q", c.expectedRemoveNetworkID, networkID)
		}
		return nil
	}
	return errors.New("Engine no longer exists")
}

func TestNetworksInitialize(t *testing.T) {
	errorCases := []struct {
		networkEnabled        bool
		network               *Network
		inspectError          error
		inspectNetworkDriver  string
		inspectNetworkOptions map[string]string
		expectedNetworkCreate types.NetworkCreate
		expectedName          string
	}{
		// NetworkNotEnabled, never an error
		{
			networkEnabled: false,
			network: &Network{
				name:   "net1",
				driver: "driver1",
			},
			inspectError: networkNotFound{
				network: "net1",
			},
		},
		// External
		{
			networkEnabled: true,
			network: &Network{
				name:     "net1",
				external: true,
			},
		},
		// NotFound, will create a new one
		{
			networkEnabled: true,
			network: &Network{
				name:   "net1",
				driver: "driver1",
			},
			inspectError: networkNotFound{
				network: "net1",
			},
			expectedName: "net1",
			expectedNetworkCreate: types.NetworkCreate{
				Driver: "driver1",
			},
		},
		// NotFound, will create a new one
		// with IPAM
		{
			networkEnabled: true,
			network: &Network{
				name:   "net1",
				driver: "driver1",
				ipam: config.Ipam{
					Driver: "ipamDriver",
					Config: []config.IpamConfig{
						{
							Subnet:  "subnet",
							IPRange: "iprange",
						},
					},
				},
			},
			inspectError: networkNotFound{
				network: "net1",
			},
			expectedName: "net1",
			expectedNetworkCreate: types.NetworkCreate{
				Driver: "driver1",
				IPAM: &network.IPAM{
					Driver: "ipamDriver",
					Config: []network.IPAMConfig{
						{
							Subnet:  "subnet",
							IPRange: "iprange",
						},
					},
				},
			},
		},
		{
			networkEnabled: true,
			network: &Network{
				name:   "net1",
				driver: "driver1",
			},
			inspectNetworkDriver: "driver1",
		},
		{
			networkEnabled: true,
			network: &Network{
				name:   "net1",
				driver: "driver1",
				driverOptions: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
			inspectNetworkDriver: "driver1",
			inspectNetworkOptions: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
	}
	for _, e := range errorCases {
		cli := &networkClient{
			expectedName:          e.expectedName,
			expectedNetworkCreate: e.expectedNetworkCreate,
			inspectError:          e.inspectError,
			inspectNetworkDriver:  e.inspectNetworkDriver,
			inspectNetworkOptions: e.inspectNetworkOptions,
		}
		e.network.client = cli
		networks := &Networks{
			networkEnabled: e.networkEnabled,
			networks: []*Network{
				e.network,
			},
		}
		err := networks.Initialize(context.Background())
		if err != nil {
			t.Error(err)
		}
	}
}

func TestNetworksInitializeErrors(t *testing.T) {
	errorCases := []struct {
		network               *Network
		inspectError          error
		inspectNetworkDriver  string
		inspectNetworkOptions map[string]string
		expectedNetworkCreate types.NetworkCreate
		expectedName          string
		expectedError         string
	}{
		{
			network: &Network{
				projectName: "prj",
				name:        "net1",
			},
			inspectError:  fmt.Errorf("any error"),
			expectedError: "any error",
		},
		{
			network: &Network{
				projectName: "prj",
				name:        "net1",
				external:    true,
			},
			inspectError: networkNotFound{
				network: "net1",
			},
			expectedError: "Network net1 declared as external, but could not be found. Please create the network manually using docker network create net1 and try again",
		},
		{
			network: &Network{
				projectName: "prj",
				name:        "net1",
			},
			inspectError: networkNotFound{
				network: "net1",
			},
			expectedError: "Engine no longer exists", // default error
		},
		{
			network: &Network{
				projectName: "prj",
				name:        "net1",
				driver:      "driver1",
			},
			inspectNetworkDriver: "driver2",
			expectedError:        `Network "prj_net1" needs to be recreated - driver has changed`,
		},
		{
			network: &Network{
				projectName: "prj",
				name:        "net1",
				driver:      "driver1",
				driverOptions: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
			inspectNetworkDriver: "driver1",
			inspectNetworkOptions: map[string]string{
				"key1": "value1",
				"key2": "anothervalue",
			},
			expectedError: `Network "prj_net1" needs to be recreated - options have changed`,
		},
	}
	for index, e := range errorCases {
		cli := &networkClient{
			expectedName:          e.expectedName,
			expectedNetworkCreate: e.expectedNetworkCreate,
			inspectError:          e.inspectError,
			inspectNetworkDriver:  e.inspectNetworkDriver,
			inspectNetworkOptions: e.inspectNetworkOptions,
		}
		e.network.client = cli
		networks := &Networks{
			networkEnabled: true,
			networks: []*Network{
				e.network,
			},
		}
		err := networks.Initialize(context.Background())
		if err == nil || err.Error() != e.expectedError {
			t.Errorf("%d: expected an error %v, got %v", index, e.expectedError, err)
		}
	}
}

func TestNetworksRemove(t *testing.T) {
	removeCases := []struct {
		networkEnabled          bool
		expectedRemoveNetworkID string
		network                 *Network
	}{
		// Network not enabled, always no error
		{
			networkEnabled: false,
			network: &Network{
				projectName: "prj",
				name:        "net1",
				driver:      "driver1",
			},
		},
		// Network enabled
		{
			networkEnabled:          true,
			expectedRemoveNetworkID: "prj_net1",
			network: &Network{
				projectName: "prj",
				name:        "net1",
				driver:      "driver1",
			},
		},
	}
	for _, c := range removeCases {
		cli := &networkClient{
			expectedRemoveNetworkID: c.expectedRemoveNetworkID,
		}
		c.network.client = cli
		networks := &Networks{
			networkEnabled: c.networkEnabled,
			networks: []*Network{
				c.network,
			},
		}
		err := networks.Remove(context.Background())
		if err != nil {
			t.Error(err)
		}
	}
}

func TestNetworksRemoveErrors(t *testing.T) {
	cli := &networkClient{}
	networks := &Networks{
		networkEnabled: true,
		networks: []*Network{
			{
				client:      cli,
				projectName: "prj",
				name:        "net1",
			},
		},
	}
	err := networks.Remove(context.Background())
	if err == nil {
		t.Errorf("Expected a error, got nothing.")
	}
}
