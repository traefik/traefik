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

import "errors"

type Container struct {
	Type    string    `json:"type,omitempty"`
	Docker  *Docker   `json:"docker,omitempty"`
	Volumes []*Volume `json:"volumes,omitempty"`
}
type PortMapping struct {
	ContainerPort int    `json:"containerPort,omitempty"`
	HostPort      int    `json:"hostPort"`
	ServicePort   int    `json:"servicePort,omitempty"`
	Protocol      string `json:"protocol"`
}

type Parameters struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type Volume struct {
	ContainerPath string `json:"containerPath,omitempty"`
	HostPath      string `json:"hostPath,omitempty"`
	Mode          string `json:"mode,omitempty"`
}

type Docker struct {
	ForcePullImage bool           `json:"forcePullImage,omitempty"`
	Image          string         `json:"image,omitempty"`
	Network        string         `json:"network,omitempty"`
	Parameters     []*Parameters  `json:"parameters,omitempty"`
	PortMappings   []*PortMapping `json:"portMappings,omitempty"`
	Privileged     bool           `json:"privileged,omitempty"`
}

func (container *Container) Volume(host_path, container_path, mode string) *Container {
	if container.Volumes == nil {
		container.Volumes = make([]*Volume, 0)
	}
	container.Volumes = append(container.Volumes, &Volume{
		ContainerPath: container_path,
		HostPath:      host_path,
		Mode:          mode,
	})
	return container
}

func NewDockerContainer() *Container {
	container := new(Container)
	container.Type = "DOCKER"
	container.Docker = &Docker{
		Image:        "",
		Network:      "BRIDGE",
		PortMappings: make([]*PortMapping, 0),
		Parameters:   make([]*Parameters, 0),
	}
	container.Volumes = make([]*Volume, 0)
	return container
}

func (docker *Docker) Container(image string) *Docker {
	docker.Image = image
	return docker
}

func (docker *Docker) Bridged() *Docker {
	docker.Network = "BRIDGE"
	return docker
}

func (docker *Docker) Expose(port int) *Docker {
	docker.ExposePort(port, 0, 0, "tcp")
	return docker
}

func (docker *Docker) ExposeUDP(port int) *Docker {
	docker.ExposePort(port, 0, 0, "udp")
	return docker
}

func (docker *Docker) ExposePort(container_port, host_port, service_port int, protocol string) *Docker {
	if docker.PortMappings == nil {
		docker.PortMappings = make([]*PortMapping, 0)
	}
	docker.PortMappings = append(docker.PortMappings, &PortMapping{
		ContainerPort: container_port,
		HostPort:      host_port,
		ServicePort:   service_port,
		Protocol:      protocol})
	return docker
}

func (docker *Docker) Parameter(key string, value string) *Docker {
	if docker.Parameters == nil {
		docker.Parameters = make([]*Parameters, 0)
	}
	docker.Parameters = append(docker.Parameters, &Parameters{
		Key:   key,
		Value: value})

	return docker
}

func (docker *Docker) ServicePortIndex(port int) (int, error) {
	if docker.PortMappings == nil || len(docker.PortMappings) <= 0 {
		return 0, errors.New("The docker does not contain any port mappings to search")
	}
	/* step: iterate and find the port */
	for index, container_port := range docker.PortMappings {
		if container_port.ContainerPort == port {
			return index, nil
		}
	}
	/* step: we didn't find the port in the mappings */
	return 0, errors.New("The container port required was not found in the container port mappings")
}
