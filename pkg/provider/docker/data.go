package docker

import (
	containertypes "github.com/moby/moby/api/types/container"
	networktypes "github.com/moby/moby/api/types/network"
)

// dockerData holds the need data to the provider.
type dockerData struct {
	ID              string
	ServiceName     string
	Name            string
	Status          containertypes.ContainerState
	Labels          map[string]string // List of labels set to container or service
	NetworkSettings networkSettings
	Health          containertypes.HealthStatus
	NodeIP          string // Only filled in Swarm mode.
	ExtraConf       configuration
}

// NetworkSettings holds the networks data to the provider.
type networkSettings struct {
	NetworkMode containertypes.NetworkMode
	Ports       networktypes.PortMap
	Networks    map[string]*networkData
}

// networkData holds the network data to the provider.
type networkData struct {
	Name     string
	Addr     string
	Port     int
	Protocol string
	ID       string
}
