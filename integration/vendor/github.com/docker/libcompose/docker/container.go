package docker

import (
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/promise"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/engine-api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/labels"
	"github.com/docker/libcompose/logger"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/project/events"
	util "github.com/docker/libcompose/utils"
	"github.com/docker/libcompose/yaml"
)

// Container holds information about a docker container and the service it is tied on.
type Container struct {
	name            string
	serviceName     string
	projectName     string
	containerNumber int
	oneOff          bool
	eventNotifier   events.Notifier
	loggerFactory   logger.Factory
	client          client.APIClient

	// FIXME(vdemeester) Remove this dependency
	service *Service
}

// NewContainer creates a container struct with the specified docker client, name and service.
func NewContainer(client client.APIClient, name string, containerNumber int, service *Service) *Container {
	return &Container{
		client:          client,
		name:            name,
		containerNumber: containerNumber,

		// TODO(vdemeester) Move these to arguments
		serviceName:   service.name,
		projectName:   service.project.Name,
		eventNotifier: service.project,
		loggerFactory: service.context.LoggerFactory,

		// TODO(vdemeester) Remove this dependency
		service: service,
	}
}

// NewOneOffContainer creates a "oneoff" container struct with the specified docker client, name and service.
func NewOneOffContainer(client client.APIClient, name string, containerNumber int, service *Service) *Container {
	c := NewContainer(client, name, containerNumber, service)
	c.oneOff = true
	return c
}

func (c *Container) findExisting(ctx context.Context) (*types.ContainerJSON, error) {
	return GetContainer(ctx, c.client, c.name)
}

// Info returns info about the container, like name, command, state or ports.
func (c *Container) Info(ctx context.Context, qFlag bool) (project.Info, error) {
	container, err := c.findExisting(ctx)
	if err != nil || container == nil {
		return nil, err
	}

	infos, err := GetContainersByFilter(ctx, c.client, map[string][]string{
		"name": {container.Name},
	})
	if err != nil || len(infos) == 0 {
		return nil, err
	}
	info := infos[0]

	result := project.Info{}
	if qFlag {
		result = append(result, project.InfoPart{Key: "Id", Value: container.ID})
	} else {
		result = append(result, project.InfoPart{Key: "Name", Value: name(info.Names)})
		result = append(result, project.InfoPart{Key: "Command", Value: info.Command})
		result = append(result, project.InfoPart{Key: "State", Value: info.Status})
		result = append(result, project.InfoPart{Key: "Ports", Value: portString(info.Ports)})
	}

	return result, nil
}

func portString(ports []types.Port) string {
	result := []string{}

	for _, port := range ports {
		if port.PublicPort > 0 {
			result = append(result, fmt.Sprintf("%s:%d->%d/%s", port.IP, port.PublicPort, port.PrivatePort, port.Type))
		} else {
			result = append(result, fmt.Sprintf("%d/%s", port.PrivatePort, port.Type))
		}
	}

	return strings.Join(result, ", ")
}

func name(names []string) string {
	max := math.MaxInt32
	var current string

	for _, v := range names {
		if len(v) < max {
			max = len(v)
			current = v
		}
	}

	return current[1:]
}

// Recreate will not refresh the container by means of relaxation and enjoyment,
// just delete it and create a new one with the current configuration
func (c *Container) Recreate(ctx context.Context, imageName string) (*types.ContainerJSON, error) {
	container, err := c.findExisting(ctx)
	if err != nil || container == nil {
		return nil, err
	}

	hash := container.Config.Labels[labels.HASH.Str()]
	if hash == "" {
		return nil, fmt.Errorf("Failed to find hash on old container: %s", container.Name)
	}

	name := container.Name[1:]
	newName := fmt.Sprintf("%s_%s", name, container.ID[:12])
	logrus.Debugf("Renaming %s => %s", name, newName)
	if err := c.client.ContainerRename(ctx, container.ID, newName); err != nil {
		logrus.Errorf("Failed to rename old container %s", c.name)
		return nil, err
	}

	newContainer, err := c.createContainer(ctx, imageName, container.ID, nil)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("Created replacement container %s", newContainer.ID)

	if err := c.client.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{
		Force:         true,
		RemoveVolumes: false,
	}); err != nil {
		logrus.Errorf("Failed to remove old container %s", c.name)
		return nil, err
	}
	logrus.Debugf("Removed old container %s %s", c.name, container.ID)

	return newContainer, nil
}

// Create creates the container based on the specified image name and send an event
// to notify the container has been created. If the container already exists, does
// nothing.
func (c *Container) Create(ctx context.Context, imageName string) (*types.ContainerJSON, error) {
	return c.CreateWithOverride(ctx, imageName, nil)
}

// CreateWithOverride create container and override parts of the config to
// allow special situations to override the config generated from the compose
// file
func (c *Container) CreateWithOverride(ctx context.Context, imageName string, configOverride *config.ServiceConfig) (*types.ContainerJSON, error) {
	container, err := c.findExisting(ctx)
	if err != nil {
		return nil, err
	}

	if container == nil {
		container, err = c.createContainer(ctx, imageName, "", configOverride)
		if err != nil {
			return nil, err
		}
		c.eventNotifier.Notify(events.ContainerCreated, c.serviceName, map[string]string{
			"name": c.Name(),
		})
	}

	return container, err
}

// Stop stops the container.
func (c *Container) Stop(ctx context.Context, timeout int) error {
	return c.withContainer(ctx, func(container *types.ContainerJSON) error {
		timeoutDuration := time.Duration(timeout) * time.Second
		return c.client.ContainerStop(ctx, container.ID, &timeoutDuration)
	})
}

// Pause pauses the container. If the containers are already paused, don't fail.
func (c *Container) Pause(ctx context.Context) error {
	return c.withContainer(ctx, func(container *types.ContainerJSON) error {
		if !container.State.Paused {
			return c.client.ContainerPause(ctx, container.ID)
		}
		return nil
	})
}

// Unpause unpauses the container. If the containers are not paused, don't fail.
func (c *Container) Unpause(ctx context.Context) error {
	return c.withContainer(ctx, func(container *types.ContainerJSON) error {
		if container.State.Paused {
			return c.client.ContainerUnpause(ctx, container.ID)
		}
		return nil
	})
}

// Kill kill the container.
func (c *Container) Kill(ctx context.Context, signal string) error {
	return c.withContainer(ctx, func(container *types.ContainerJSON) error {
		return c.client.ContainerKill(ctx, container.ID, signal)
	})
}

// Delete removes the container if existing. If the container is running, it tries
// to stop it first.
func (c *Container) Delete(ctx context.Context, removeVolume bool) error {
	container, err := c.findExisting(ctx)
	if err != nil || container == nil {
		return err
	}

	info, err := c.client.ContainerInspect(ctx, container.ID)
	if err != nil {
		return err
	}

	if !info.State.Running {
		return c.client.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{
			Force:         true,
			RemoveVolumes: removeVolume,
		})
	}

	return nil
}

// IsRunning returns the running state of the container.
func (c *Container) IsRunning(ctx context.Context) (bool, error) {
	container, err := c.findExisting(ctx)
	if err != nil || container == nil {
		return false, err
	}

	info, err := c.client.ContainerInspect(ctx, container.ID)
	if err != nil {
		return false, err
	}

	return info.State.Running, nil
}

// Run creates, start and attach to the container based on the image name,
// the specified configuration.
// It will always create a new container.
func (c *Container) Run(ctx context.Context, configOverride *config.ServiceConfig) (int, error) {
	var (
		errCh       chan error
		out, stderr io.Writer
		in          io.ReadCloser
	)

	container, err := c.findExisting(ctx)
	if err != nil || container == nil {
		return -1, err
	}

	if configOverride.StdinOpen {
		in = os.Stdin
	}
	if configOverride.Tty {
		out = os.Stdout
	}
	if configOverride.Tty {
		stderr = os.Stderr
	}

	options := types.ContainerAttachOptions{
		Stream: true,
		Stdin:  configOverride.StdinOpen,
		Stdout: configOverride.Tty,
		Stderr: configOverride.Tty,
	}

	resp, err := c.client.ContainerAttach(ctx, container.ID, options)
	if err != nil {
		return -1, err
	}

	// set raw terminal
	inFd, _ := term.GetFdInfo(in)
	state, err := term.SetRawTerminal(inFd)
	if err != nil {
		return -1, err
	}
	// restore raw terminal
	defer term.RestoreTerminal(inFd, state)
	// holdHijackedConnection (in goroutine)
	errCh = promise.Go(func() error {
		return holdHijackedConnection(configOverride.Tty, in, out, stderr, resp)
	})

	if err := c.client.ContainerStart(ctx, container.ID, types.ContainerStartOptions{}); err != nil {
		return -1, err
	}

	if err := <-errCh; err != nil {
		logrus.Debugf("Error hijack: %s", err)
		return -1, err
	}

	exitedContainer, err := c.client.ContainerInspect(ctx, container.ID)
	if err != nil {
		return -1, err
	}

	return exitedContainer.State.ExitCode, nil
}

func holdHijackedConnection(tty bool, inputStream io.ReadCloser, outputStream, errorStream io.Writer, resp types.HijackedResponse) error {
	var err error
	receiveStdout := make(chan error, 1)
	if outputStream != nil || errorStream != nil {
		go func() {
			// When TTY is ON, use regular copy
			if tty && outputStream != nil {
				_, err = io.Copy(outputStream, resp.Reader)
			} else {
				_, err = stdcopy.StdCopy(outputStream, errorStream, resp.Reader)
			}
			logrus.Debugf("[hijack] End of stdout")
			receiveStdout <- err
		}()
	}

	stdinDone := make(chan struct{})
	go func() {
		if inputStream != nil {
			io.Copy(resp.Conn, inputStream)
			logrus.Debugf("[hijack] End of stdin")
		}

		if err := resp.CloseWrite(); err != nil {
			logrus.Debugf("Couldn't send EOF: %s", err)
		}
		close(stdinDone)
	}()

	select {
	case err := <-receiveStdout:
		if err != nil {
			logrus.Debugf("Error receiveStdout: %s", err)
			return err
		}
	case <-stdinDone:
		if outputStream != nil || errorStream != nil {
			if err := <-receiveStdout; err != nil {
				logrus.Debugf("Error receiveStdout: %s", err)
				return err
			}
		}
	}

	return nil
}

// Start the specified container with the specified host config
func (c *Container) Start(ctx context.Context) error {
	container, err := c.findExisting(ctx)
	if err != nil || container == nil {
		return err
	}
	logrus.WithFields(logrus.Fields{"container.ID": container.ID, "c.name": c.name}).Debug("Starting container")
	if err := c.client.ContainerStart(context.Background(), container.ID, types.ContainerStartOptions{}); err != nil {
		logrus.WithFields(logrus.Fields{"container.ID": container.ID, "c.name": c.name}).Debug("Failed to start container")
		return err
	}
	c.eventNotifier.Notify(events.ContainerStarted, c.serviceName, map[string]string{
		"name": c.Name(),
	})
	return nil
}

// OutOfSync checks if the container is out of sync with the service definition.
// It looks if the the service hash container label is the same as the computed one.
func (c *Container) OutOfSync(ctx context.Context, imageName string) (bool, error) {
	container, err := c.findExisting(ctx)
	if err != nil || container == nil {
		return false, err
	}

	if container.Config.Image != imageName {
		logrus.Debugf("Images for %s do not match %s!=%s", c.name, container.Config.Image, imageName)
		return true, nil
	}

	if container.Config.Labels[labels.HASH.Str()] != c.getHash() {
		logrus.Debugf("Hashes for %s do not match %s!=%s", c.name, container.Config.Labels[labels.HASH.Str()], c.getHash())
		return true, nil
	}

	image, _, err := c.client.ImageInspectWithRaw(ctx, container.Config.Image, false)
	if err != nil {
		if client.IsErrImageNotFound(err) {
			logrus.Debugf("Image %s do not exist, do not know if it's out of sync", container.Config.Image)
			return false, nil
		}
		return false, err
	}

	logrus.Debugf("Checking existing image name vs id: %s == %s", image.ID, container.Image)
	return image.ID != container.Image, err
}

func (c *Container) getHash() string {
	return config.GetServiceHash(c.serviceName, c.service.Config())
}

func volumeBinds(volumes map[string]struct{}, container *types.ContainerJSON) []string {
	result := make([]string, 0, len(container.Mounts))
	for _, mount := range container.Mounts {
		if _, ok := volumes[mount.Destination]; ok {
			result = append(result, fmt.Sprint(mount.Source, ":", mount.Destination))
		}
	}
	return result
}

func (c *Container) createContainer(ctx context.Context, imageName, oldContainer string, configOverride *config.ServiceConfig) (*types.ContainerJSON, error) {
	serviceConfig := c.service.serviceConfig
	if configOverride != nil {
		serviceConfig.Command = configOverride.Command
		serviceConfig.Tty = configOverride.Tty
		serviceConfig.StdinOpen = configOverride.StdinOpen
	}
	configWrapper, err := ConvertToAPI(c.service)
	if err != nil {
		return nil, err
	}

	configWrapper.Config.Image = imageName

	if configWrapper.Config.Labels == nil {
		configWrapper.Config.Labels = map[string]string{}
	}

	oneOffString := "False"
	if c.oneOff {
		oneOffString = "True"
	}

	configWrapper.Config.Labels[labels.SERVICE.Str()] = c.serviceName
	configWrapper.Config.Labels[labels.PROJECT.Str()] = c.projectName
	configWrapper.Config.Labels[labels.HASH.Str()] = c.getHash()
	configWrapper.Config.Labels[labels.ONEOFF.Str()] = oneOffString
	configWrapper.Config.Labels[labels.NUMBER.Str()] = fmt.Sprint(c.containerNumber)
	configWrapper.Config.Labels[labels.VERSION.Str()] = ComposeVersion

	err = c.populateAdditionalHostConfig(configWrapper.HostConfig)
	if err != nil {
		return nil, err
	}

	if oldContainer != "" {
		info, err := c.client.ContainerInspect(ctx, oldContainer)
		if err != nil {
			return nil, err
		}
		configWrapper.HostConfig.Binds = util.Merge(configWrapper.HostConfig.Binds, volumeBinds(configWrapper.Config.Volumes, &info))
	}

	logrus.Debugf("Creating container %s %#v", c.name, configWrapper)

	container, err := c.client.ContainerCreate(ctx, configWrapper.Config, configWrapper.HostConfig, configWrapper.NetworkingConfig, c.name)
	if err != nil {
		logrus.Debugf("Failed to create container %s: %v", c.name, err)
		return nil, err
	}

	return GetContainer(ctx, c.client, container.ID)
}

func (c *Container) populateAdditionalHostConfig(hostConfig *container.HostConfig) error {
	links, err := c.getLinks()
	if err != nil {
		return err
	}

	for _, link := range c.service.DependentServices() {
		if !c.service.project.ServiceConfigs.Has(link.Target) {
			continue
		}

		service, err := c.service.project.CreateService(link.Target)
		if err != nil {
			return err
		}

		// FIXME(vdemeester) container should not know service
		containers, err := service.Containers(context.Background())
		if err != nil {
			return err
		}

		if link.Type == project.RelTypeIpcNamespace {
			hostConfig, err = c.addIpc(hostConfig, service, containers)
		} else if link.Type == project.RelTypeNetNamespace {
			hostConfig, err = c.addNetNs(hostConfig, service, containers)
		}

		if err != nil {
			return err
		}
	}

	hostConfig.Links = []string{}
	for k, v := range links {
		hostConfig.Links = append(hostConfig.Links, strings.Join([]string{v, k}, ":"))
	}
	for _, v := range c.service.Config().ExternalLinks {
		hostConfig.Links = append(hostConfig.Links, v)
	}

	return nil
}

// FIXME(vdemeester) this is temporary
func (c *Container) getLinks() (map[string]string, error) {
	links := map[string]string{}
	for _, link := range c.service.DependentServices() {
		if !c.service.project.ServiceConfigs.Has(link.Target) {
			continue
		}

		service, err := c.service.project.CreateService(link.Target)
		if err != nil {
			return nil, err
		}

		// FIXME(vdemeester) container should not know service
		containers, err := service.Containers(context.Background())
		if err != nil {
			return nil, err
		}

		if link.Type == project.RelTypeLink {
			c.addLinks(links, service, link, containers)
		}

		if err != nil {
			return nil, err
		}
	}
	return links, nil
}

func (c *Container) addLinks(links map[string]string, service project.Service, rel project.ServiceRelationship, containers []project.Container) {
	for _, container := range containers {
		if _, ok := links[rel.Alias]; !ok {
			links[rel.Alias] = container.Name()
		}

		links[container.Name()] = container.Name()
	}
}

func (c *Container) addIpc(config *container.HostConfig, service project.Service, containers []project.Container) (*container.HostConfig, error) {
	if len(containers) == 0 {
		return nil, fmt.Errorf("Failed to find container for IPC %v", c.service.Config().Ipc)
	}

	id, err := containers[0].ID()
	if err != nil {
		return nil, err
	}

	config.IpcMode = container.IpcMode("container:" + id)
	return config, nil
}

func (c *Container) addNetNs(config *container.HostConfig, service project.Service, containers []project.Container) (*container.HostConfig, error) {
	if len(containers) == 0 {
		return nil, fmt.Errorf("Failed to find container for networks ns %v", c.service.Config().NetworkMode)
	}

	id, err := containers[0].ID()
	if err != nil {
		return nil, err
	}

	config.NetworkMode = container.NetworkMode("container:" + id)
	return config, nil
}

// ID returns the container Id.
func (c *Container) ID() (string, error) {
	// FIXME(vdemeester) container should not ask for his ID..
	container, err := c.findExisting(context.Background())
	if container == nil {
		return "", err
	}
	return container.ID, err
}

// Name returns the container name.
func (c *Container) Name() string {
	return c.name
}

// Restart restarts the container if existing, does nothing otherwise.
func (c *Container) Restart(ctx context.Context, timeout int) error {
	container, err := c.findExisting(ctx)
	if err != nil || container == nil {
		return err
	}

	timeoutDuration := time.Duration(timeout) * time.Second
	return c.client.ContainerRestart(ctx, container.ID, &timeoutDuration)
}

// Log forwards container logs to the project configured logger.
func (c *Container) Log(ctx context.Context, follow bool) error {
	container, err := c.findExisting(ctx)
	if container == nil || err != nil {
		return err
	}

	info, err := c.client.ContainerInspect(ctx, container.ID)
	if err != nil {
		return err
	}

	// FIXME(vdemeester) update container struct to do less API calls
	name := fmt.Sprintf("%s_%d", c.service.name, c.containerNumber)
	l := c.loggerFactory.Create(name)

	options := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Tail:       "all",
	}
	responseBody, err := c.client.ContainerLogs(ctx, c.name, options)
	if err != nil {
		return err
	}
	defer responseBody.Close()

	if info.Config.Tty {
		_, err = io.Copy(&logger.Wrapper{Logger: l}, responseBody)
	} else {
		_, err = stdcopy.StdCopy(&logger.Wrapper{Logger: l}, &logger.Wrapper{Logger: l, Err: true}, responseBody)
	}
	logrus.WithFields(logrus.Fields{"Logger": l, "err": err}).Debug("c.client.Logs() returned error")

	return err
}

func (c *Container) withContainer(ctx context.Context, action func(*types.ContainerJSON) error) error {
	container, err := c.findExisting(ctx)
	if err != nil {
		return err
	}

	if container != nil {
		return action(container)
	}

	return nil
}

// Port returns the host port the specified port is mapped on.
func (c *Container) Port(ctx context.Context, port string) (string, error) {
	container, err := c.findExisting(ctx)
	if err != nil {
		return "", err
	}

	if bindings, ok := container.NetworkSettings.Ports[nat.Port(port)]; ok {
		result := []string{}
		for _, binding := range bindings {
			result = append(result, binding.HostIP+":"+binding.HostPort)
		}

		return strings.Join(result, "\n"), nil
	}
	return "", nil
}

// Networks returns the containers network
// FIXME(vdemeester) should not need ctx or calling the API, will take care of it
// when refactoring Container.
func (c *Container) Networks(ctx context.Context) (map[string]*network.EndpointSettings, error) {
	container, err := c.findExisting(ctx)
	if err != nil {
		return nil, err
	}
	if container == nil {
		return map[string]*network.EndpointSettings{}, nil
	}
	return container.NetworkSettings.Networks, nil
}

// NetworkDisconnect disconnects the container from the specified network
// FIXME(vdemeester) will be refactor with Container refactoring
func (c *Container) NetworkDisconnect(ctx context.Context, net *yaml.Network) error {
	container, err := c.findExisting(ctx)
	if err != nil || container == nil {
		return err
	}
	return c.client.NetworkDisconnect(ctx, net.RealName, container.ID, true)
}

// NetworkConnect connects the container to the specified network
// FIXME(vdemeester) will be refactor with Container refactoring
func (c *Container) NetworkConnect(ctx context.Context, net *yaml.Network) error {
	container, err := c.findExisting(ctx)
	if err != nil || container == nil {
		return err
	}
	internalLinks, err := c.getLinks()
	if err != nil {
		return err
	}
	links := []string{}
	// TODO(vdemeester) handle link to self (?)
	for k, v := range internalLinks {
		links = append(links, strings.Join([]string{v, k}, ":"))
	}
	for _, v := range c.service.Config().ExternalLinks {
		links = append(links, v)
	}
	aliases := []string{}
	if !c.oneOff {
		aliases = []string{c.serviceName}
	}
	aliases = append(aliases, net.Aliases...)
	return c.client.NetworkConnect(ctx, net.RealName, container.ID, &network.EndpointSettings{
		Aliases:   aliases,
		Links:     links,
		IPAddress: net.IPv4Address,
		IPAMConfig: &network.EndpointIPAMConfig{
			IPv4Address: net.IPv4Address,
			IPv6Address: net.IPv6Address,
		},
	})
}
