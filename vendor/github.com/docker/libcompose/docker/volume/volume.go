package volume

import (
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/libcompose/config"
	composeclient "github.com/docker/libcompose/docker/client"
	"github.com/docker/libcompose/project"
	"golang.org/x/net/context"
)

// Volume holds attributes and method for a volume definition in compose
type Volume struct {
	client        client.VolumeAPIClient
	projectName   string
	name          string
	driver        string
	driverOptions map[string]string
	external      bool
	// TODO (shouze) missing labels
}

func (v *Volume) fullName() string {
	name := v.projectName + "_" + v.name
	if v.external {
		name = v.name
	}
	return name
}

// Inspect inspect the current volume
func (v *Volume) Inspect(ctx context.Context) (types.Volume, error) {
	return v.client.VolumeInspect(ctx, v.fullName())
}

// Remove removes the current volume (from docker engine)
func (v *Volume) Remove(ctx context.Context) error {
	if v.external {
		fmt.Printf("Volume %s is external, skipping", v.fullName())
		return nil
	}
	fmt.Printf("Removing volume %q\n", v.fullName())
	return v.client.VolumeRemove(ctx, v.fullName(), true)
}

// EnsureItExists make sure the volume exists and return an error if it does not exists
// and cannot be created.
func (v *Volume) EnsureItExists(ctx context.Context) error {
	volumeResource, err := v.Inspect(ctx)
	if v.external {
		if client.IsErrNotFound(err) {
			// FIXME(shouze) introduce some libcompose error type
			return fmt.Errorf("Volume %s declared as external, but could not be found. Please create the volume manually using docker volume create %s and try again", v.name, v.name)
		}
		return err
	}
	if err != nil && client.IsErrNotFound(err) {
		return v.create(ctx)
	}
	if volumeResource.Driver != v.driver {
		return fmt.Errorf("Volume %q needs to be recreated - driver has changed", v.name)
	}
	return err
}

func (v *Volume) create(ctx context.Context) error {
	fmt.Printf("Creating volume %q with driver %q\n", v.fullName(), v.driver)
	_, err := v.client.VolumeCreate(ctx, volume.VolumesCreateBody{
		Name:       v.fullName(),
		Driver:     v.driver,
		DriverOpts: v.driverOptions,
		// TODO (shouze) missing labels
	})

	return err
}

// NewVolume creates a new volume from the specified name and config.
func NewVolume(projectName, name string, config *config.VolumeConfig, client client.VolumeAPIClient) *Volume {
	vol := &Volume{
		client:      client,
		projectName: projectName,
		name:        name,
	}
	if config != nil {
		vol.driver = config.Driver
		vol.driverOptions = config.DriverOpts
		vol.external = config.External.External

	}
	return vol
}

// Volumes holds a list of volume
type Volumes struct {
	volumes       []*Volume
	volumeEnabled bool
}

// Initialize make sure volume exists if volume is enabled
func (v *Volumes) Initialize(ctx context.Context) error {
	if !v.volumeEnabled {
		return nil
	}
	for _, volume := range v.volumes {
		err := volume.EnsureItExists(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// Remove removes volumes (clean-up)
func (v *Volumes) Remove(ctx context.Context) error {
	if !v.volumeEnabled {
		return nil
	}
	for _, volume := range v.volumes {
		err := volume.Remove(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// VolumesFromServices creates a new Volumes struct based on volumes configurations and
// services configuration. If a volume is defined but not used by any service, it will return
// an error along the Volumes.
func VolumesFromServices(cli client.VolumeAPIClient, projectName string, volumeConfigs map[string]*config.VolumeConfig, services *config.ServiceConfigs, volumeEnabled bool) (*Volumes, error) {
	var err error
	volumes := make([]*Volume, 0, len(volumeConfigs))
	for name, config := range volumeConfigs {
		volume := NewVolume(projectName, name, config, cli)
		volumes = append(volumes, volume)
	}
	return &Volumes{
		volumes:       volumes,
		volumeEnabled: volumeEnabled,
	}, err
}

// DockerFactory implements project.VolumesFactory
type DockerFactory struct {
	ClientFactory composeclient.Factory
}

// Create implements project.VolumesFactory Create method.
// It creates a Volumes (that implements project.Volumes) from specified configurations.
func (f *DockerFactory) Create(projectName string, volumeConfigs map[string]*config.VolumeConfig, serviceConfigs *config.ServiceConfigs, volumeEnabled bool) (project.Volumes, error) {
	cli := f.ClientFactory.Create(nil)
	return VolumesFromServices(cli, projectName, volumeConfigs, serviceConfigs, volumeEnabled)
}
