package docker

import (
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
)

// dockerData holds the need data to the provider.
type dockerData struct {
	ID              string
	ServiceName     string
	Name            string
	Status          string
	Labels          map[string]string // List of labels set to container or service
	NetworkSettings networkSettings
	Health          string
	NodeIP          string // Only filled in Swarm mode.
	ExtraConf       configuration
}

// NetworkSettings holds the networks data to the provider.
type networkSettings struct {
	NetworkMode containertypes.NetworkMode
	Ports       nat.PortMap
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
