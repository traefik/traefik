/*
Copyright 2017 The go-marathon Authors All rights reserved.

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

// PodNetworkMode is the mode of a network descriptor
type PodNetworkMode string

const (
	ContainerNetworkMode PodNetworkMode = "container"
	BridgeNetworkMode    PodNetworkMode = "container/bridge"
	HostNetworkMode      PodNetworkMode = "host"
)

// PodNetwork contains network descriptors for a pod
type PodNetwork struct {
	Name   string            `json:"name,omitempty"`
	Mode   PodNetworkMode    `json:"mode,omitempty"`
	Labels map[string]string `json:"labels,omitempty"`
}

// PodEndpoint describes an endpoint for a pod's container
type PodEndpoint struct {
	Name          string            `json:"name,omitempty"`
	ContainerPort int               `json:"containerPort,omitempty"`
	HostPort      int               `json:"hostPort,omitempty"`
	Protocol      []string          `json:"protocol,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
}

// NewPodNetwork creates an empty PodNetwork
func NewPodNetwork(name string) *PodNetwork {
	return &PodNetwork{
		Name:   name,
		Labels: map[string]string{},
	}
}

// NewPodEndpoint creates an empty PodEndpoint
func NewPodEndpoint() *PodEndpoint {
	return &PodEndpoint{
		Protocol: []string{},
		Labels:   map[string]string{},
	}
}

// NewBridgePodNetwork creates a PodNetwork for a container in bridge mode
func NewBridgePodNetwork() *PodNetwork {
	pn := NewPodNetwork("")
	return pn.SetMode(BridgeNetworkMode)
}

// NewContainerPodNetwork creates a PodNetwork for a container
func NewContainerPodNetwork(name string) *PodNetwork {
	pn := NewPodNetwork(name)
	return pn.SetMode(ContainerNetworkMode)
}

// NewHostPodNetwork creates a PodNetwork for a container in host mode
func NewHostPodNetwork() *PodNetwork {
	pn := NewPodNetwork("")
	return pn.SetMode(HostNetworkMode)
}

// SetName sets the name of a PodNetwork
func (n *PodNetwork) SetName(name string) *PodNetwork {
	n.Name = name
	return n
}

// SetMode sets the mode of a PodNetwork
func (n *PodNetwork) SetMode(mode PodNetworkMode) *PodNetwork {
	n.Mode = mode
	return n
}

// Label sets a label of a PodNetwork
func (n *PodNetwork) Label(key, value string) *PodNetwork {
	n.Labels[key] = value
	return n
}

// SetName sets the name for a PodEndpoint
func (e *PodEndpoint) SetName(name string) *PodEndpoint {
	e.Name = name
	return e
}

// SetContainerPort sets the container port for a PodEndpoint
func (e *PodEndpoint) SetContainerPort(port int) *PodEndpoint {
	e.ContainerPort = port
	return e
}

// SetHostPort sets the host port for a PodEndpoint
func (e *PodEndpoint) SetHostPort(port int) *PodEndpoint {
	e.HostPort = port
	return e
}

// AddProtocol appends a protocol for a PodEndpoint
func (e *PodEndpoint) AddProtocol(protocol string) *PodEndpoint {
	e.Protocol = append(e.Protocol, protocol)
	return e
}

// Label sets a label for a PodEndpoint
func (e *PodEndpoint) Label(key, value string) *PodEndpoint {
	e.Labels[key] = value
	return e
}
