package service

import (
	"fmt"
	"strings"

	"github.com/docker/cli/opts"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/go-connections/nat"
	"github.com/docker/go-units"
	"github.com/docker/libcompose/config"
	composeclient "github.com/docker/libcompose/docker/client"
	composecontainer "github.com/docker/libcompose/docker/container"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/utils"
	"golang.org/x/net/context"
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

func toMap(vs []string) map[string]struct{} {
	m := map[string]struct{}{}
	for _, v := range vs {
		if v != "" {
			m[v] = struct{}{}
		}
	}
	return m
}

func isBind(s string) bool {
	return strings.ContainsRune(s, ':')
}

func isVolume(s string) bool {
	return !isBind(s)
}

// ConvertToAPI converts a service configuration to a docker API container configuration.
func ConvertToAPI(serviceConfig *config.ServiceConfig, ctx project.Context, clientFactory composeclient.Factory) (*ConfigWrapper, error) {
	config, hostConfig, err := Convert(serviceConfig, ctx, clientFactory)
	if err != nil {
		return nil, err
	}

	result := ConfigWrapper{
		Config:     config,
		HostConfig: hostConfig,
	}
	return &result, nil
}

func volumes(c *config.ServiceConfig, ctx project.Context) []string {
	if c.Volumes == nil {
		return []string{}
	}
	volumes := make([]string, len(c.Volumes.Volumes))
	for _, v := range c.Volumes.Volumes {
		vol := v
		if len(ctx.ComposeFiles) > 0 && !project.IsNamedVolume(v.Source) {
			sourceVol := ctx.ResourceLookup.ResolvePath(v.String(), ctx.ComposeFiles[0])
			vol.Source = strings.SplitN(sourceVol, ":", 2)[0]
		}
		volumes = append(volumes, vol.String())
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

	vols := volumes(c, ctx)

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
		Volumes:      toMap(Filter(vols, isVolume)),
		MacAddress:   c.MacAddress,
		StopSignal:   c.StopSignal,
		StopTimeout:  utils.DurationStrToSecondsInt(c.StopGracePeriod),
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

	memorySwappiness := int64(c.MemSwappiness)

	resources := container.Resources{
		CgroupParent:      c.CgroupParent,
		Memory:            int64(c.MemLimit),
		MemoryReservation: int64(c.MemReservation),
		MemorySwap:        int64(c.MemSwapLimit),
		MemorySwappiness:  &memorySwappiness,
		CPUShares:         int64(c.CPUShares),
		CPUQuota:          int64(c.CPUQuota),
		CpusetCpus:        c.CPUSet,
		Ulimits:           ulimits,
		Devices:           deviceMappings,
		OomKillDisable:    &c.OomKillDisable,
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
					containerID := container.ID()
					networkMode = "container:" + containerID
				}
				// FIXME(vdemeester) log/warn in case of len(containers) == 0
			}
		case strings.HasPrefix(c.NetworkMode, "container:"):
			containerName := c.NetworkMode[10:]
			client := clientFactory.Create(nil)
			container, err := composecontainer.Get(context.Background(), client, containerName)
			if err != nil {
				return nil, nil, err
			}
			networkMode = "container:" + container.ID
		default:
			// do nothing :)
		}
	}

	tmpfs := map[string]string{}
	for _, path := range c.Tmpfs {
		split := strings.SplitN(path, ":", 2)
		if len(split) == 1 {
			tmpfs[split[0]] = ""
		} else if len(split) == 2 {
			tmpfs[split[0]] = split[1]
		}
	}

	hostConfig := &container.HostConfig{
		VolumesFrom: volumesFrom,
		CapAdd:      strslice.StrSlice(utils.CopySlice(c.CapAdd)),
		CapDrop:     strslice.StrSlice(utils.CopySlice(c.CapDrop)),
		GroupAdd:    c.GroupAdd,
		ExtraHosts:  utils.CopySlice(c.ExtraHosts),
		Privileged:  c.Privileged,
		Binds:       Filter(vols, isBind),
		DNS:         utils.CopySlice(c.DNS),
		DNSOptions:  utils.CopySlice(c.DNSOpts),
		DNSSearch:   utils.CopySlice(c.DNSSearch),
		Isolation:   container.Isolation(c.Isolation),
		LogConfig: container.LogConfig{
			Type:   c.Logging.Driver,
			Config: utils.CopyMap(c.Logging.Options),
		},
		NetworkMode:    container.NetworkMode(networkMode),
		ReadonlyRootfs: c.ReadOnly,
		OomScoreAdj:    int(c.OomScoreAdj),
		PidMode:        container.PidMode(c.Pid),
		UTSMode:        container.UTSMode(c.Uts),
		IpcMode:        container.IpcMode(c.Ipc),
		PortBindings:   portBindings,
		RestartPolicy:  *restartPolicy,
		ShmSize:        int64(c.ShmSize),
		SecurityOpt:    utils.CopySlice(c.SecurityOpt),
		Tmpfs:          tmpfs,
		VolumeDriver:   c.VolumeDriver,
		Resources:      resources,
	}

	if config.Labels == nil {
		config.Labels = map[string]string{}
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
		v, err := parseDevice(device)
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

// parseDevice parses a device mapping string to a container.DeviceMapping struct
// FIXME(vdemeester) de-duplicate this by re-exporting it in docker/docker
func parseDevice(device string) (container.DeviceMapping, error) {
	src := ""
	dst := ""
	permissions := "rwm"
	arr := strings.Split(device, ":")
	switch len(arr) {
	case 3:
		permissions = arr[2]
		fallthrough
	case 2:
		if validDeviceMode(arr[1]) {
			permissions = arr[1]
		} else {
			dst = arr[1]
		}
		fallthrough
	case 1:
		src = arr[0]
	default:
		return container.DeviceMapping{}, fmt.Errorf("invalid device specification: %s", device)
	}

	if dst == "" {
		dst = src
	}

	deviceMapping := container.DeviceMapping{
		PathOnHost:        src,
		PathInContainer:   dst,
		CgroupPermissions: permissions,
	}
	return deviceMapping, nil
}

// validDeviceMode checks if the mode for device is valid or not.
// Valid mode is a composition of r (read), w (write), and m (mknod).
func validDeviceMode(mode string) bool {
	var legalDeviceMode = map[rune]bool{
		'r': true,
		'w': true,
		'm': true,
	}
	if mode == "" {
		return false
	}
	for _, c := range mode {
		if !legalDeviceMode[c] {
			return false
		}
		legalDeviceMode[c] = false
	}
	return true
}
