package project

import (
	"golang.org/x/net/context"

	"github.com/docker/libcompose/config"
)

// Volumes defines the methods a libcompose volume aggregate should define.
type Volumes interface {
	Initialize(ctx context.Context) error
	Remove(ctx context.Context) error
}

// VolumesFactory is an interface factory to create Volumes object for the specified
// configurations (service, volumes, …)
type VolumesFactory interface {
	Create(projectName string, volumeConfigs map[string]*config.VolumeConfig, serviceConfigs *config.ServiceConfigs, volumeEnabled bool) (Volumes, error)
}
