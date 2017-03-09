package docker

import (
	"fmt"
	"strings"

	"golang.org/x/net/context"

	"github.com/docker/docker/runconfig/opts"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/engine-api/types/network"
	"github.com/docker/engine-api/types/strslice"
	"github.com/docker/go-connections/nat"
	"github.com/docker/go-units"
	"github.com/docker/libcompose/config"
	composeclient "github.com/docker/libcompose/docker/client"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/utils"
)

// ConfigWrapper wraps Config, HostConfig and NetworkingConfig for a container.
type ConfigWrapper struct {
	Config           *container.Config
	HostConfig       *container.HostConfig
	NetworkingConfig *network.NetworkingConfig
}

// Filter filters the specified string slice with the specified function.
func Filter(vs []string, f func(string) bool) []string {
	r := make([]string, 0, len(vs))
	for _, v := range vs {
		if f(v) {
			r = append(r, v)
		}
	}
	return r
}

func isBind(s string) bool {
	return strings.ContainsRune(s, ':')
}

func isVolume(s string) bool {
	return !isBind(s)
}

// ConvertToAPI converts a service configuration to a docker API container configuration.
func ConvertToAPI(s *Service) (*ConfigWrapper, error) {
	config, hostConfig, err := Convert(s.serviceConfig, s.context.Context, s.clientFactory)
	if err != nil {
		return nil, err
	}

	result := ConfigWrapper{
		Config:     config,
		HostConfig: hostConfig,
	}
	return &result, nil
}

func isNamedVolume(volume string) bool {
	return !strings.HasPrefix(volume, ".") && !strings.HasPrefix(volume, "/") && !strings.HasPrefix(volume, "~")
}

func volumes(c *config.ServiceConfig, ctx project.Context) map[string]struct{} {
	volumes := make(map[string]struct{}, len(c.Volumes))
	for k, v := range c.Volumes {
		if len(ctx.ComposeFiles) > 0 && !isNamedVolume(v) {
			v = ctx.ResourceLookup.ResolvePath(v, ctx.ComposeFiles[0])
		}

		c.Volumes[k] = v
		if isVolume(v) {
			volumes[v] = struct{}{}
		}
	}
	return volumes
}

func restartPolicy(c *config.ServiceConfig) (*container.RestartPolicy, error) {
	restart, err := opts.ParseRestartPolicy(c.Restart)
	if err != nil {
		return nil, err
	}
	return &container.RestartPolicy{Name: restart.Name, MaximumRetryCount: restart.MaximumRetryCount}, nil
}

func ports(c *config.ServiceConfig) (map[nat.Port]struct{}, nat.PortMap, error) {
	ports, binding, err := nat.ParsePortSpecs(c.Ports)
	if err != nil {
		return nil, nil, err
	}

	exPorts, _, err := nat.ParsePortSpecs(c.Expose)
	if err != nil {
		return nil, nil, err
	}

	for k, v := range exPorts {
		ports[k] = v
	}

	exposedPorts := map[nat.Port]struct{}{}
	for k, v := range ports {
		exposedPorts[nat.Port(k)] = v
	}

	portBindings := nat.PortMap{}
	for k, bv := range binding {
		dcbs := make([]nat.PortBinding, len(bv))
		for k, v := range bv {
			dcbs[k] = nat.PortBinding{HostIP: v.HostIP, HostPort: v.HostPort}
		}
		portBindings[nat.Port(k)] = dcbs
	}
	return exposedPorts, portBindings, nil
}

// Convert converts a service configuration to an docker API structures (Config and HostConfig)
func Convert(c *config.ServiceConfig, ctx project.Context, clientFactory composeclient.Factory) (*container.Config, *container.HostConfig, error) {
	restartPolicy, err := restartPolicy(c)
	if err != nil {
		return nil, nil, err
	}

	exposedPorts, portBindings, err := ports(c)
	if err != nil {
		return nil, nil, err
	}

	deviceMappings, err := parseDevices(c.Devices)
	if err != nil {
		return nil, nil, err
	}

	var volumesFrom []string
	if c.VolumesFrom != nil {
		volumesFrom, err = getVolumesFrom(c.VolumesFrom, ctx.Project.ServiceConfigs, ctx.ProjectName)
		if err != nil {
			return nil, nil, err
		}
	}

	config := &container.Config{
		Entrypoint:   strslice.StrSlice(utils.CopySlice(c.Entrypoint)),
		Hostname:     c.Hostname,
		Domainname:   c.DomainName,
		User:         c.User,
		Env:          utils.CopySlice(c.Environment),
		Cmd:          strslice.StrSlice(utils.CopySlice(c.Command)),
		Image:        c.Image,
		Labels:       utils.CopyMap(c.Labels),
		ExposedPorts: exposedPorts,
		Tty:          c.Tty,
		OpenStdin:    c.StdinOpen,
		WorkingDir:   c.WorkingDir,
		Volumes:      volumes(c, ctx),
		MacAddress:   c.MacAddress,
	}

	ulimits := []*units.Ulimit{}
	if c.Ulimits.Elements != nil {
		for _, ulimit := range c.Ulimits.Elements {
			ulimits = append(ulimits, &units.Ulimit{
				Name: ulimit.Name,
				Soft: ulimit.Soft,
				Hard: ulimit.Hard,
			})
		}
	}

	resources := container.Resources{
		CgroupParent: c.CgroupParent,
		Memory:       c.MemLimit,
		MemorySwap:   c.MemSwapLimit,
		CPUShares:    c.CPUShares,
		CPUQuota:     c.CPUQuota,
		CpusetCpus:   c.CPUSet,
		Ulimits:      ulimits,
		Devices:      deviceMappings,
	}

	networkMode := c.NetworkMode
	if c.NetworkMode == "" {
		if c.Networks != nil && len(c.Networks.Networks) > 0 {
			networkMode = c.Networks.Networks[0].RealName
		}
	} else {
		switch {
		case strings.HasPrefix(c.NetworkMode, "service:"):
			serviceName := c.NetworkMode[8:]
			if serviceConfig, ok := ctx.Project.ServiceConfigs.Get(serviceName); ok {
				// FIXME(vdemeester) this is actually not right, should be fixed but not there
				service, err := ctx.ServiceFactory.Create(ctx.Project, serviceName, serviceConfig)
				if err != nil {
					return nil, nil, err
				}
				containers, err := service.Containers(context.Background())
				if err != nil {
					return nil, nil, err
				}
				if len(containers) != 0 {
					container := containers[0]
					containerID, err := container.ID()
					if err != nil {
						return nil, nil, err
					}
					networkMode = "container:" + containerID
				}
				// FIXME(vdemeester) log/warn in case of len(containers) == 0
			}
		case strings.HasPrefix(c.NetworkMode, "container:"):
			containerName := c.NetworkMode[10:]
			client := clientFactory.Create(nil)
			container, err := GetContainer(context.Background(), client, containerName)
			if err != nil {
				return nil, nil, err
			}
			networkMode = "container:" + container.ID
		default:
			// do nothing :)
		}
	}

	hostConfig := &container.HostConfig{
		VolumesFrom: volumesFrom,
		CapAdd:      strslice.StrSlice(utils.CopySlice(c.CapAdd)),
		CapDrop:     strslice.StrSlice(utils.CopySlice(c.CapDrop)),
		ExtraHosts:  utils.CopySlice(c.ExtraHosts),
		Privileged:  c.Privileged,
		Binds:       Filter(c.Volumes, isBind),
		DNS:         utils.CopySlice(c.DNS),
		DNSSearch:   utils.CopySlice(c.DNSSearch),
		LogConfig: container.LogConfig{
			Type:   c.Logging.Driver,
			Config: utils.CopyMap(c.Logging.Options),
		},
		NetworkMode:    container.NetworkMode(networkMode),
		ReadonlyRootfs: c.ReadOnly,
		PidMode:        container.PidMode(c.Pid),
		UTSMode:        container.UTSMode(c.Uts),
		IpcMode:        container.IpcMode(c.Ipc),
		PortBindings:   portBindings,
		RestartPolicy:  *restartPolicy,
		ShmSize:        c.ShmSize,
		SecurityOpt:    utils.CopySlice(c.SecurityOpt),
		VolumeDriver:   c.VolumeDriver,
		Resources:      resources,
	}

	return config, hostConfig, nil
}

func getVolumesFrom(volumesFrom []string, serviceConfigs *config.ServiceConfigs, projectName string) ([]string, error) {
	volumes := []string{}
	for _, volumeFrom := range volumesFrom {
		if serviceConfig, ok := serviceConfigs.Get(volumeFrom); ok {
			// It's a service - Use the first one
			name := fmt.Sprintf("%s_%s_1", projectName, volumeFrom)
			// If a container name is specified, use that instead
			if serviceConfig.ContainerName != "" {
				name = serviceConfig.ContainerName
			}
			volumes = append(volumes, name)
		} else {
			volumes = append(volumes, volumeFrom)
		}
	}
	return volumes, nil
}

func parseDevices(devices []string) ([]container.DeviceMapping, error) {
	// parse device mappings
	deviceMappings := []container.DeviceMapping{}
	for _, device := range devices {
		v, err := opts.ParseDevice(device)
		if err != nil {
			return nil, err
		}
		deviceMappings = append(deviceMappings, container.DeviceMapping{
			PathOnHost:        v.PathOnHost,
			PathInContainer:   v.PathInContainer,
			CgroupPermissions: v.CgroupPermissions,
		})
	}

	return deviceMappings, nil
}
