package project

import (
	"golang.org/x/net/context"

	"github.com/docker/libcompose/config"
)

// Networks defines the methods a libcompose network aggregate should define.
type Networks interface {
	Initialize(ctx context.Context) error
	Remove(ctx context.Context) error
}

// NetworksFactory is an interface factory to create Networks object for the specified
// configurations (service, networks, …)
type NetworksFactory interface {
	Create(projectName string, networkConfigs map[string]*config.NetworkConfig, serviceConfigs *config.ServiceConfigs, networkEnabled bool) (Networks, error)
}
