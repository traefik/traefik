package docker

import (
	containertypes "github.com/docker/docker/api/types/container"
	networktypes "github.com/docker/docker/api/types/network"
	swarmtypes "github.com/docker/docker/api/types/swarm"
	"github.com/docker/go-connections/nat"
)

func containerJSON(ops ...func(*containertypes.InspectResponse)) containertypes.InspectResponse {
	c := &containertypes.InspectResponse{
		ContainerJSONBase: &containertypes.ContainerJSONBase{
			Name:       "fake",
			HostConfig: &containertypes.HostConfig{},
			State:      &containertypes.State{},
		},
		Config: &containertypes.Config{},
		NetworkSettings: &containertypes.NetworkSettings{
			NetworkSettingsBase: containertypes.NetworkSettingsBase{},
		},
	}

	for _, op := range ops {
		op(c)
	}

	return *c
}

func name(name string) func(*containertypes.InspectResponse) {
	return func(c *containertypes.InspectResponse) {
		c.ContainerJSONBase.Name = name
	}
}

func networkMode(mode string) func(*containertypes.InspectResponse) {
	return func(c *containertypes.InspectResponse) {
		c.ContainerJSONBase.HostConfig.NetworkMode = containertypes.NetworkMode(mode)
	}
}

func ports(portMap nat.PortMap) func(*containertypes.InspectResponse) {
	return func(c *containertypes.InspectResponse) {
		c.NetworkSettings.NetworkSettingsBase.Ports = portMap
	}
}

func withNetwork(name string, ops ...func(*networktypes.EndpointSettings)) func(*containertypes.InspectResponse) {
	return func(c *containertypes.InspectResponse) {
		if c.NetworkSettings.Networks == nil {
			c.NetworkSettings.Networks = map[string]*networktypes.EndpointSettings{}
		}
		c.NetworkSettings.Networks[name] = &networktypes.EndpointSettings{}
		for _, op := range ops {
			op(c.NetworkSettings.Networks[name])
		}
	}
}

func ipv4(ip string) func(*networktypes.EndpointSettings) {
	return func(s *networktypes.EndpointSettings) {
		s.IPAddress = ip
	}
}

func ipv6(ip string) func(*networktypes.EndpointSettings) {
	return func(s *networktypes.EndpointSettings) {
		s.GlobalIPv6Address = ip
	}
}

func swarmTask(id string, ops ...func(*swarmtypes.Task)) swarmtypes.Task {
	task := &swarmtypes.Task{
		ID: id,
	}

	for _, op := range ops {
		op(task)
	}

	return *task
}

func taskSlot(slot int) func(*swarmtypes.Task) {
	return func(task *swarmtypes.Task) {
		task.Slot = slot
	}
}

func taskNodeID(id string) func(*swarmtypes.Task) {
	return func(task *swarmtypes.Task) {
		task.NodeID = id
	}
}

func taskNetworkAttachment(id, name, driver string, addresses []string) func(*swarmtypes.Task) {
	return func(task *swarmtypes.Task) {
		task.NetworksAttachments = append(task.NetworksAttachments, swarmtypes.NetworkAttachment{
			Network: swarmtypes.Network{
				ID: id,
				Spec: swarmtypes.NetworkSpec{
					Annotations: swarmtypes.Annotations{
						Name: name,
					},
					DriverConfiguration: &swarmtypes.Driver{
						Name: driver,
					},
				},
			},
			Addresses: addresses,
		})
	}
}

func taskStatus(ops ...func(*swarmtypes.TaskStatus)) func(*swarmtypes.Task) {
	return func(task *swarmtypes.Task) {
		status := &swarmtypes.TaskStatus{}

		for _, op := range ops {
			op(status)
		}

		task.Status = *status
	}
}

func taskState(state swarmtypes.TaskState) func(*swarmtypes.TaskStatus) {
	return func(status *swarmtypes.TaskStatus) {
		status.State = state
	}
}

func taskContainerStatus(id string) func(*swarmtypes.TaskStatus) {
	return func(status *swarmtypes.TaskStatus) {
		status.ContainerStatus = &swarmtypes.ContainerStatus{
			ContainerID: id,
		}
	}
}

func swarmService(ops ...func(*swarmtypes.Service)) swarmtypes.Service {
	service := &swarmtypes.Service{
		ID: "serviceID",
		Spec: swarmtypes.ServiceSpec{
			Annotations: swarmtypes.Annotations{
				Name: "defaultServiceName",
			},
		},
	}

	for _, op := range ops {
		op(service)
	}

	return *service
}

func serviceName(name string) func(service *swarmtypes.Service) {
	return func(service *swarmtypes.Service) {
		service.Spec.Annotations.Name = name
	}
}

func serviceLabels(labels map[string]string) func(service *swarmtypes.Service) {
	return func(service *swarmtypes.Service) {
		service.Spec.Annotations.Labels = labels
	}
}

func withEndpoint(ops ...func(*swarmtypes.Endpoint)) func(*swarmtypes.Service) {
	return func(service *swarmtypes.Service) {
		endpoint := &swarmtypes.Endpoint{}

		for _, op := range ops {
			op(endpoint)
		}

		service.Endpoint = *endpoint
	}
}

func virtualIP(networkID, addr string) func(*swarmtypes.Endpoint) {
	return func(endpoint *swarmtypes.Endpoint) {
		if endpoint.VirtualIPs == nil {
			endpoint.VirtualIPs = []swarmtypes.EndpointVirtualIP{}
		}
		endpoint.VirtualIPs = append(endpoint.VirtualIPs, swarmtypes.EndpointVirtualIP{
			NetworkID: networkID,
			Addr:      addr,
		})
	}
}

func withEndpointSpec(ops ...func(*swarmtypes.EndpointSpec)) func(*swarmtypes.Service) {
	return func(service *swarmtypes.Service) {
		endpointSpec := &swarmtypes.EndpointSpec{}

		for _, op := range ops {
			op(endpointSpec)
		}

		service.Spec.EndpointSpec = endpointSpec
	}
}

func modeDNSRR(spec *swarmtypes.EndpointSpec) {
	spec.Mode = swarmtypes.ResolutionModeDNSRR
}

func modeVIP(spec *swarmtypes.EndpointSpec) {
	spec.Mode = swarmtypes.ResolutionModeVIP
}
