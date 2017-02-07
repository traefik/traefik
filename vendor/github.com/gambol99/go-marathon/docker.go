/*
Copyright 2014 Rohith All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package marathon

import (
	"errors"
	"fmt"
)

// Container is the definition for a container type in marathon
type Container struct {
	Type    string    `json:"type,omitempty"`
	Docker  *Docker   `json:"docker,omitempty"`
	Volumes *[]Volume `json:"volumes,omitempty"`
}

// PortMapping is the portmapping structure between container and mesos
type PortMapping struct {
	ContainerPort int                `json:"containerPort,omitempty"`
	HostPort      int                `json:"hostPort"`
	Labels        *map[string]string `json:"labels,omitempty"`
	Name          string             `json:"name,omitempty"`
	ServicePort   int                `json:"servicePort,omitempty"`
	Protocol      string             `json:"protocol,omitempty"`
}

// Parameters is the parameters to pass to the docker client when creating the container
type Parameters struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

// Volume is the docker volume details associated to the container
type Volume struct {
	ContainerPath string          `json:"containerPath,omitempty"`
	HostPath      string          `json:"hostPath,omitempty"`
	External      *ExternalVolume `json:"external,omitempty"`
	Mode          string          `json:"mode,omitempty"`
}

// ExternalVolume is an external volume definition
type ExternalVolume struct {
	Name     string             `json:"name,omitempty"`
	Provider string             `json:"provider,omitempty"`
	Options  *map[string]string `json:"options,omitempty"`
}

// Docker is the docker definition from a marathon application
type Docker struct {
	ForcePullImage *bool          `json:"forcePullImage,omitempty"`
	Image          string         `json:"image,omitempty"`
	Network        string         `json:"network,omitempty"`
	Parameters     *[]Parameters  `json:"parameters,omitempty"`
	PortMappings   *[]PortMapping `json:"portMappings,omitempty"`
	Privileged     *bool          `json:"privileged,omitempty"`
}

// Volume attachs a volume to the container
//		host_path:			the path on the docker host to map
//		container_path:		the path inside the container to map the host volume
//		mode:				the mode to map the container
func (container *Container) Volume(hostPath, containerPath, mode string) *Container {
	if container.Volumes == nil {
		container.EmptyVolumes()
	}

	volumes := *container.Volumes
	volumes = append(volumes, Volume{
		ContainerPath: containerPath,
		HostPath:      hostPath,
		Mode:          mode,
	})

	container.Volumes = &volumes

	return container
}

// EmptyVolumes explicitly empties the volumes -- use this if you need to empty
// volumes of an application that already has volumes set (setting volumes to nil will
// keep the current value)
func (container *Container) EmptyVolumes() *Container {
	container.Volumes = &[]Volume{}
	return container
}

// SetExternalVolume define external elements for a volume
//      name: the name of the volume
//      provider: the provider of the volume (e.g. dvdi)
func (v *Volume) SetExternalVolume(name, provider string) *ExternalVolume {
	ev := &ExternalVolume{
		Name:     name,
		Provider: provider,
	}
	v.External = ev
	return ev
}

// EmptyExternalVolume emptys the external volume definition
func (v *Volume) EmptyExternalVolume() *Volume {
	v.External = &ExternalVolume{}
	return v
}

// AddOption adds an option to an ExternalVolume
//		name:  the name of the option
//		value: value for the option
func (ev *ExternalVolume) AddOption(name, value string) *ExternalVolume {
	if ev.Options == nil {
		ev.EmptyOptions()
	}
	(*ev.Options)[name] = value

	return ev
}

// EmptyOptions explicitly empties the options
func (ev *ExternalVolume) EmptyOptions() *ExternalVolume {
	ev.Options = &map[string]string{}

	return ev
}

// NewDockerContainer creates a default docker container for you
func NewDockerContainer() *Container {
	container := &Container{}
	container.Type = "DOCKER"
	container.Docker = &Docker{}

	return container
}

// SetForcePullImage sets whether the docker image should always be force pulled before
// starting an instance
//		forcePull:			true / false
func (docker *Docker) SetForcePullImage(forcePull bool) *Docker {
	docker.ForcePullImage = &forcePull

	return docker
}

// SetPrivileged sets whether the docker image should be started
// with privilege turned on
//		priv:			true / false
func (docker *Docker) SetPrivileged(priv bool) *Docker {
	docker.Privileged = &priv

	return docker
}

// Container sets the image of the container
//		image:			the image name you are using
func (docker *Docker) Container(image string) *Docker {
	docker.Image = image
	return docker
}

// Bridged sets the networking mode to bridged
func (docker *Docker) Bridged() *Docker {
	docker.Network = "BRIDGE"
	return docker
}

// Host sets the networking mode to host
func (docker *Docker) Host() *Docker {
	docker.Network = "HOST"
	return docker
}

// Expose sets the container to expose the following TCP ports
//		ports:			the TCP ports the container is exposing
func (docker *Docker) Expose(ports ...int) *Docker {
	for _, port := range ports {
		docker.ExposePort(PortMapping{
			ContainerPort: port,
			HostPort:      0,
			ServicePort:   0,
			Protocol:      "tcp"})
	}
	return docker
}

// ExposeUDP sets the container to expose the following UDP ports
//		ports:			the UDP ports the container is exposing
func (docker *Docker) ExposeUDP(ports ...int) *Docker {
	for _, port := range ports {
		docker.ExposePort(PortMapping{
			ContainerPort: port,
			HostPort:      0,
			ServicePort:   0,
			Protocol:      "udp"})
	}
	return docker
}

// ExposePort exposes an port in the container
func (docker *Docker) ExposePort(portMapping PortMapping) *Docker {
	if docker.PortMappings == nil {
		docker.EmptyPortMappings()
	}

	portMappings := *docker.PortMappings
	portMappings = append(portMappings, portMapping)
	docker.PortMappings = &portMappings

	return docker
}

// EmptyPortMappings explicitly empties the port mappings -- use this if you need to empty
// port mappings of an application that already has port mappings set (setting port mappings to nil will
// keep the current value)
func (docker *Docker) EmptyPortMappings() *Docker {
	docker.PortMappings = &[]PortMapping{}
	return docker
}

// AddLabel adds a label to a PortMapping
//		name:	the name of the label
//		value: value for this label
func (p *PortMapping) AddLabel(name, value string) *PortMapping {
	if p.Labels == nil {
		p.EmptyLabels()
	}
	(*p.Labels)[name] = value

	return p
}

// EmptyLabels explicitly empties the labels -- use this if you need to empty
// the labels of a port mapping that already has labels set (setting labels to
// nil will keep the current value)
func (p *PortMapping) EmptyLabels() *PortMapping {
	p.Labels = &map[string]string{}

	return p
}

// AddParameter adds a parameter to the docker execution line when creating the container
//		key:			the name of the option to add
//		value:		the value of the option
func (docker *Docker) AddParameter(key string, value string) *Docker {
	if docker.Parameters == nil {
		docker.EmptyParameters()
	}

	parameters := *docker.Parameters
	parameters = append(parameters, Parameters{
		Key:   key,
		Value: value})

	docker.Parameters = &parameters

	return docker
}

// EmptyParameters explicitly empties the parameters -- use this if you need to empty
// parameters of an application that already has parameters set (setting parameters to nil will
// keep the current value)
func (docker *Docker) EmptyParameters() *Docker {
	docker.Parameters = &[]Parameters{}
	return docker
}

// ServicePortIndex finds the service port index of the exposed port
//		port:			the port you are looking for
func (docker *Docker) ServicePortIndex(port int) (int, error) {
	if docker.PortMappings == nil || len(*docker.PortMappings) == 0 {
		return 0, errors.New("The docker does not contain any port mappings to search")
	}

	// step: iterate and find the port
	for index, containerPort := range *docker.PortMappings {
		if containerPort.ContainerPort == port {
			return index, nil
		}
	}

	// step: we didn't find the port in the mappings
	return 0, fmt.Errorf("The container port required was not found in the container port mappings")
}
