package service

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/docker/auth"
	"github.com/docker/libcompose/docker/builder"
	composeclient "github.com/docker/libcompose/docker/client"
	"github.com/docker/libcompose/docker/container"
	"github.com/docker/libcompose/docker/ctx"
	"github.com/docker/libcompose/docker/image"
	"github.com/docker/libcompose/labels"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/project/events"
	"github.com/docker/libcompose/project/options"
	"github.com/docker/libcompose/utils"
	"github.com/docker/libcompose/yaml"
	"github.com/sirupsen/logrus"
)

// Service is a project.Service implementations.
type Service struct {
	name          string
	project       *project.Project
	serviceConfig *config.ServiceConfig
	clientFactory composeclient.Factory
	authLookup    auth.Lookup

	// FIXME(vdemeester) remove this at some point
	context *ctx.Context
}

// NewService creates a service
func NewService(name string, serviceConfig *config.ServiceConfig, context *ctx.Context) *Service {
	return &Service{
		name:          name,
		project:       context.Project,
		serviceConfig: serviceConfig,
		clientFactory: context.ClientFactory,
		authLookup:    context.AuthLookup,
		context:       context,
	}
}

// Name returns the service name.
func (s *Service) Name() string {
	return s.name
}

// Config returns the configuration of the service (config.ServiceConfig).
func (s *Service) Config() *config.ServiceConfig {
	return s.serviceConfig
}

// DependentServices returns the dependent services (as an array of ServiceRelationship) of the service.
func (s *Service) DependentServices() []project.ServiceRelationship {
	return DefaultDependentServices(s.project, s)
}

// Create implements Service.Create. It ensures the image exists or build it
// if it can and then create a container.
func (s *Service) Create(ctx context.Context, options options.Create) error {
	containers, err := s.collectContainers(ctx)
	if err != nil {
		return err
	}

	if err := s.ensureImageExists(ctx, options.NoBuild, options.ForceBuild); err != nil {
		return err
	}

	if len(containers) != 0 {
		return s.eachContainer(ctx, containers, func(c *container.Container) error {
			_, err := s.recreateIfNeeded(ctx, c, options.NoRecreate, options.ForceRecreate)
			return err
		})
	}

	namer, err := s.namer(ctx, 1)
	if err != nil {
		return err
	}

	_, err = s.createContainer(ctx, namer, "", nil, false)
	return err
}

func (s *Service) namer(ctx context.Context, count int) (Namer, error) {
	var namer Namer
	var err error

	if s.serviceConfig.ContainerName != "" {
		if count > 1 {
			logrus.Warnf(`The "%s" service is using the custom container name "%s". Docker requires each container to have a unique name. Remove the custom name to scale the service.`, s.name, s.serviceConfig.ContainerName)
		}
		namer = NewSingleNamer(s.serviceConfig.ContainerName)
	} else {
		client := s.clientFactory.Create(s)
		namer, err = NewNamer(ctx, client, s.project.Name, s.name, false)
		if err != nil {
			return nil, err
		}
	}
	return namer, nil
}

func (s *Service) collectContainers(ctx context.Context) ([]*container.Container, error) {
	client := s.clientFactory.Create(s)
	containers, err := container.ListByFilter(ctx, client, labels.SERVICE.Eq(s.name), labels.PROJECT.Eq(s.project.Name))
	if err != nil {
		return nil, err
	}

	result := []*container.Container{}

	for _, cont := range containers {
		c, err := container.New(ctx, client, cont.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, c)
	}

	return result, nil
}

func (s *Service) ensureImageExists(ctx context.Context, noBuild bool, forceBuild bool) error {
	if forceBuild {
		return s.build(ctx, options.Build{})
	}

	exists, err := image.Exists(ctx, s.clientFactory.Create(s), s.imageName())
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	if s.Config().Build.Context != "" {
		if noBuild {
			return fmt.Errorf("Service %q needs to be built, but no-build was specified", s.name)
		}
		return s.build(ctx, options.Build{})
	}

	return s.Pull(ctx)
}

func (s *Service) imageName() string {
	if s.Config().Image != "" {
		return s.Config().Image
	}
	return fmt.Sprintf("%s_%s", s.project.Name, s.Name())
}

// Build implements Service.Build. It will try to build the image and returns an error if any.
func (s *Service) Build(ctx context.Context, buildOptions options.Build) error {
	return s.build(ctx, buildOptions)
}

func (s *Service) build(ctx context.Context, buildOptions options.Build) error {
	if s.Config().Build.Context == "" {
		return fmt.Errorf("Specified service does not have a build section")
	}
	builder := &builder.DaemonBuilder{
		Client:           s.clientFactory.Create(s),
		ContextDirectory: s.Config().Build.Context,
		Dockerfile:       s.Config().Build.Dockerfile,
		BuildArgs:        s.Config().Build.Args,
		AuthConfigs:      s.authLookup.All(),
		NoCache:          buildOptions.NoCache,
		ForceRemove:      buildOptions.ForceRemove,
		Pull:             buildOptions.Pull,
		LoggerFactory:    s.context.LoggerFactory,
	}
	return builder.Build(ctx, s.imageName())
}

func (s *Service) constructContainers(ctx context.Context, count int) ([]*container.Container, error) {
	result, err := s.collectContainers(ctx)
	if err != nil {
		return nil, err
	}

	client := s.clientFactory.Create(s)

	var namer Namer

	if s.serviceConfig.ContainerName != "" {
		if count > 1 {
			logrus.Warnf(`The "%s" service is using the custom container name "%s". Docker requires each container to have a unique name. Remove the custom name to scale the service.`, s.name, s.serviceConfig.ContainerName)
		}
		namer = NewSingleNamer(s.serviceConfig.ContainerName)
	} else {
		namer, err = NewNamer(ctx, client, s.project.Name, s.name, false)
		if err != nil {
			return nil, err
		}
	}

	for i := len(result); i < count; i++ {
		c, err := s.createContainer(ctx, namer, "", nil, false)
		if err != nil {
			return nil, err
		}

		id := c.ID()
		logrus.Debugf("Created container %s: %v", id, c.Name())

		result = append(result, c)
	}

	return result, nil
}

// Up implements Service.Up. It builds the image if needed, creates a container
// and start it.
func (s *Service) Up(ctx context.Context, options options.Up) error {
	containers, err := s.collectContainers(ctx)
	if err != nil {
		return err
	}

	var imageName = s.imageName()
	if len(containers) == 0 || !options.NoRecreate {
		if err = s.ensureImageExists(ctx, options.NoBuild, options.ForceBuild); err != nil {
			return err
		}
	}

	return s.up(ctx, imageName, true, options)
}

// Run implements Service.Run. It runs a one of command within the service container.
// It always create a new container.
func (s *Service) Run(ctx context.Context, commandParts []string, options options.Run) (int, error) {
	err := s.ensureImageExists(ctx, false, false)
	if err != nil {
		return -1, err
	}

	client := s.clientFactory.Create(s)

	namer, err := NewNamer(ctx, client, s.project.Name, s.name, true)
	if err != nil {
		return -1, err
	}

	configOverride := &config.ServiceConfig{Command: commandParts, Tty: !options.DisableTty, StdinOpen: !options.DisableTty}

	c, err := s.createContainer(ctx, namer, "", configOverride, true)
	if err != nil {
		return -1, err
	}

	if err := s.connectContainerToNetworks(ctx, c, true); err != nil {
		return -1, err
	}

	if options.Detached {
		logrus.Infof("%s", c.Name())
		return 0, c.Start(ctx)
	}
	return c.Run(ctx, configOverride)
}

// Info implements Service.Info. It returns an project.InfoSet with the containers
// related to this service (can be multiple if using the scale command).
func (s *Service) Info(ctx context.Context) (project.InfoSet, error) {
	result := project.InfoSet{}
	containers, err := s.collectContainers(ctx)
	if err != nil {
		return nil, err
	}

	for _, c := range containers {
		info, err := c.Info(ctx)
		if err != nil {
			return nil, err
		}
		result = append(result, info)
	}

	return result, nil
}

// Start implements Service.Start. It tries to start a container without creating it.
func (s *Service) Start(ctx context.Context) error {
	return s.collectContainersAndDo(ctx, func(c *container.Container) error {
		if err := s.connectContainerToNetworks(ctx, c, false); err != nil {
			return err
		}
		return c.Start(ctx)
	})
}

func (s *Service) up(ctx context.Context, imageName string, create bool, options options.Up) error {
	containers, err := s.collectContainers(ctx)
	if err != nil {
		return err
	}

	logrus.Debugf("Found %d existing containers for service %s", len(containers), s.name)

	if len(containers) == 0 && create {
		namer, err := s.namer(ctx, 1)
		if err != nil {
			return err
		}
		c, err := s.createContainer(ctx, namer, "", nil, false)
		if err != nil {
			return err
		}
		containers = []*container.Container{c}
	}

	return s.eachContainer(ctx, containers, func(c *container.Container) error {
		var err error
		if create {
			c, err = s.recreateIfNeeded(ctx, c, options.NoRecreate, options.ForceRecreate)
			if err != nil {
				return err
			}
		}

		if err := s.connectContainerToNetworks(ctx, c, false); err != nil {
			return err
		}

		err = c.Start(ctx)

		if err == nil {
			s.project.Notify(events.ContainerStarted, s.name, map[string]string{
				"name": c.Name(),
			})
		}

		return err
	})
}

func (s *Service) connectContainerToNetworks(ctx context.Context, c *container.Container, oneOff bool) error {
	connectedNetworks, err := c.Networks()
	if err != nil {
		return nil
	}
	if s.serviceConfig.Networks != nil {
		for _, network := range s.serviceConfig.Networks.Networks {
			existingNetwork, ok := connectedNetworks[network.RealName]
			if ok {
				// FIXME(vdemeester) implement alias checking (to not disconnect/reconnect for nothing)
				aliasPresent := false
				for _, alias := range existingNetwork.Aliases {
					ID := c.ShortID()
					if alias == ID {
						aliasPresent = true
					}
				}
				if aliasPresent {
					continue
				}
				if err := s.NetworkDisconnect(ctx, c, network, oneOff); err != nil {
					return err
				}
			}
			if err := s.NetworkConnect(ctx, c, network, oneOff); err != nil {
				return err
			}
		}
	}
	return nil
}

// NetworkDisconnect disconnects the container from the specified network
func (s *Service) NetworkDisconnect(ctx context.Context, c *container.Container, net *yaml.Network, oneOff bool) error {
	containerID := c.ID()
	client := s.clientFactory.Create(s)
	return client.NetworkDisconnect(ctx, net.RealName, containerID, true)
}

// NetworkConnect connects the container to the specified network
// FIXME(vdemeester) will be refactor with Container refactoring
func (s *Service) NetworkConnect(ctx context.Context, c *container.Container, net *yaml.Network, oneOff bool) error {
	containerID := c.ID()
	client := s.clientFactory.Create(s)
	internalLinks, err := s.getLinks()
	if err != nil {
		return err
	}
	links := []string{}
	// TODO(vdemeester) handle link to self (?)
	for k, v := range internalLinks {
		links = append(links, strings.Join([]string{v, k}, ":"))
	}
	for _, v := range s.serviceConfig.ExternalLinks {
		links = append(links, v)
	}
	aliases := []string{}
	if !oneOff {
		aliases = []string{s.Name()}
	}
	aliases = append(aliases, net.Aliases...)
	return client.NetworkConnect(ctx, net.RealName, containerID, &network.EndpointSettings{
		Aliases:   aliases,
		Links:     links,
		IPAddress: net.IPv4Address,
		IPAMConfig: &network.EndpointIPAMConfig{
			IPv4Address: net.IPv4Address,
			IPv6Address: net.IPv6Address,
		},
	})
}

func (s *Service) recreateIfNeeded(ctx context.Context, c *container.Container, noRecreate, forceRecreate bool) (*container.Container, error) {
	if noRecreate {
		return c, nil
	}
	outOfSync, err := s.OutOfSync(ctx, c)
	if err != nil {
		return c, err
	}

	logrus.WithFields(logrus.Fields{
		"outOfSync":     outOfSync,
		"ForceRecreate": forceRecreate,
		"NoRecreate":    noRecreate}).Debug("Going to decide if recreate is needed")

	if forceRecreate || outOfSync {
		logrus.Infof("Recreating %s", s.name)
		newContainer, err := s.recreate(ctx, c)
		if err != nil {
			return c, err
		}
		return newContainer, nil
	}

	return c, err
}

func (s *Service) recreate(ctx context.Context, c *container.Container) (*container.Container, error) {
	name := c.Name()
	id := c.ID()
	newName := fmt.Sprintf("%s_%s", name, id[:12])
	logrus.Debugf("Renaming %s => %s", name, newName)
	if err := c.Rename(ctx, newName); err != nil {
		logrus.Errorf("Failed to rename old container %s", c.Name())
		return nil, err
	}
	namer := NewSingleNamer(name)
	newContainer, err := s.createContainer(ctx, namer, id, nil, false)
	if err != nil {
		return nil, err
	}
	newID := newContainer.ID()
	logrus.Debugf("Created replacement container %s", newID)
	if err := c.Remove(ctx, false); err != nil {
		logrus.Errorf("Failed to remove old container %s", c.Name())
		return nil, err
	}
	logrus.Debugf("Removed old container %s %s", c.Name(), id)
	return newContainer, nil
}

// OutOfSync checks if the container is out of sync with the service definition.
// It looks if the the service hash container label is the same as the computed one.
func (s *Service) OutOfSync(ctx context.Context, c *container.Container) (bool, error) {
	if c.ImageConfig() != s.serviceConfig.Image {
		logrus.Debugf("Images for %s do not match %s!=%s", c.Name(), c.ImageConfig(), s.serviceConfig.Image)
		return true, nil
	}

	expectedHash := config.GetServiceHash(s.name, s.Config())
	if c.Hash() != expectedHash {
		logrus.Debugf("Hashes for %s do not match %s!=%s", c.Name(), c.Hash(), expectedHash)
		return true, nil
	}

	image, err := image.InspectImage(ctx, s.clientFactory.Create(s), c.ImageConfig())
	if err != nil {
		if client.IsErrNotFound(err) {
			logrus.Debugf("Image %s do not exist, do not know if it's out of sync", c.Image())
			return false, nil
		}
		return false, err
	}

	logrus.Debugf("Checking existing image name vs id: %s == %s", image.ID, c.Image())
	return image.ID != c.Image(), err
}

func (s *Service) collectContainersAndDo(ctx context.Context, action func(*container.Container) error) error {
	containers, err := s.collectContainers(ctx)
	if err != nil {
		return err
	}
	return s.eachContainer(ctx, containers, action)
}

func (s *Service) eachContainer(ctx context.Context, containers []*container.Container, action func(*container.Container) error) error {

	tasks := utils.InParallel{}
	for _, cont := range containers {
		task := func(cont *container.Container) func() error {
			return func() error {
				return action(cont)
			}
		}(cont)

		tasks.Add(task)
	}

	return tasks.Wait()
}

// Stop implements Service.Stop. It stops any containers related to the service.
func (s *Service) Stop(ctx context.Context, timeout int) error {
	timeout = s.stopTimeout(timeout)
	return s.collectContainersAndDo(ctx, func(c *container.Container) error {
		return c.Stop(ctx, timeout)
	})
}

// Restart implements Service.Restart. It restarts any containers related to the service.
func (s *Service) Restart(ctx context.Context, timeout int) error {
	timeout = s.stopTimeout(timeout)
	return s.collectContainersAndDo(ctx, func(c *container.Container) error {
		return c.Restart(ctx, timeout)
	})
}

// Kill implements Service.Kill. It kills any containers related to the service.
func (s *Service) Kill(ctx context.Context, signal string) error {
	return s.collectContainersAndDo(ctx, func(c *container.Container) error {
		return c.Kill(ctx, signal)
	})
}

// Delete implements Service.Delete. It removes any containers related to the service.
func (s *Service) Delete(ctx context.Context, options options.Delete) error {
	return s.collectContainersAndDo(ctx, func(c *container.Container) error {
		running := c.IsRunning(ctx)
		if !running || options.RemoveRunning {
			return c.Remove(ctx, options.RemoveVolume)
		}
		return nil
	})
}

// Log implements Service.Log. It returns the docker logs for each container related to the service.
func (s *Service) Log(ctx context.Context, follow bool) error {
	return s.collectContainersAndDo(ctx, func(c *container.Container) error {
		containerNumber, err := c.Number()
		if err != nil {
			return err
		}
		name := fmt.Sprintf("%s_%d", s.name, containerNumber)
		if s.Config().ContainerName != "" {
			name = s.Config().ContainerName
		}
		l := s.context.LoggerFactory.CreateContainerLogger(name)
		return c.Log(ctx, l, follow)
	})
}

// Scale implements Service.Scale. It creates or removes containers to have the specified number
// of related container to the service to run.
func (s *Service) Scale(ctx context.Context, scale int, timeout int) error {
	if s.specificiesHostPort() {
		logrus.Warnf("The \"%s\" service specifies a port on the host. If multiple containers for this service are created on a single host, the port will clash.", s.Name())
	}

	containers, err := s.collectContainers(ctx)
	if err != nil {
		return err
	}
	if len(containers) > scale {
		foundCount := 0
		for _, c := range containers {
			foundCount++
			if foundCount > scale {
				timeout = s.stopTimeout(timeout)
				if err := c.Stop(ctx, timeout); err != nil {
					return err
				}
				// FIXME(vdemeester) remove volume in scale by default ?
				if err := c.Remove(ctx, false); err != nil {
					return err
				}
			}
		}
	}

	if err != nil {
		return err
	}

	if len(containers) < scale {
		err := s.ensureImageExists(ctx, false, false)
		if err != nil {
			return err
		}

		if _, err = s.constructContainers(ctx, scale); err != nil {
			return err
		}
	}

	return s.up(ctx, "", false, options.Up{})
}

// Pull implements Service.Pull. It pulls the image of the service and skip the service that
// would need to be built.
func (s *Service) Pull(ctx context.Context) error {
	if s.Config().Image == "" {
		return nil
	}

	return image.PullImage(ctx, s.clientFactory.Create(s), s.name, s.authLookup, s.Config().Image)
}

// Pause implements Service.Pause. It puts into pause the container(s) related
// to the service.
func (s *Service) Pause(ctx context.Context) error {
	return s.collectContainersAndDo(ctx, func(c *container.Container) error {
		return c.Pause(ctx)
	})
}

// Unpause implements Service.Pause. It brings back from pause the container(s)
// related to the service.
func (s *Service) Unpause(ctx context.Context) error {
	return s.collectContainersAndDo(ctx, func(c *container.Container) error {
		return c.Unpause(ctx)
	})
}

// RemoveImage implements Service.RemoveImage. It removes images used for the service
// depending on the specified type.
func (s *Service) RemoveImage(ctx context.Context, imageType options.ImageType) error {
	switch imageType {
	case "local":
		if s.Config().Image != "" {
			return nil
		}
		return image.RemoveImage(ctx, s.clientFactory.Create(s), s.imageName())
	case "all":
		return image.RemoveImage(ctx, s.clientFactory.Create(s), s.imageName())
	default:
		// Don't do a thing, should be validated up-front
		return nil
	}
}

var eventAttributes = []string{"image", "name"}

// Events implements Service.Events. It listen to all real-time events happening
// for the service, and put them into the specified chan.
func (s *Service) Events(ctx context.Context, evts chan events.ContainerEvent) error {
	filter := filters.NewArgs()
	filter.Add("label", fmt.Sprintf("%s=%s", labels.PROJECT, s.project.Name))
	filter.Add("label", fmt.Sprintf("%s=%s", labels.SERVICE, s.name))
	client := s.clientFactory.Create(s)
	eventq, errq := client.Events(ctx, types.EventsOptions{
		Filters: filter,
	})
	go func() {
		for {
			select {
			case event := <-eventq:
				service := event.Actor.Attributes[labels.SERVICE.Str()]
				attributes := map[string]string{}
				for _, attr := range eventAttributes {
					attributes[attr] = event.Actor.Attributes[attr]
				}
				e := events.ContainerEvent{
					Service:    service,
					Event:      event.Action,
					Type:       event.Type,
					ID:         event.Actor.ID,
					Time:       time.Unix(event.Time, 0),
					Attributes: attributes,
				}
				evts <- e
			}
		}
	}()
	return <-errq
}

// Containers implements Service.Containers. It returns the list of containers
// that are related to the service.
func (s *Service) Containers(ctx context.Context) ([]project.Container, error) {
	result := []project.Container{}
	containers, err := s.collectContainers(ctx)
	if err != nil {
		return nil, err
	}

	for _, c := range containers {
		result = append(result, c)
	}

	return result, nil
}

func (s *Service) specificiesHostPort() bool {
	_, bindings, err := nat.ParsePortSpecs(s.Config().Ports)

	if err != nil {
		fmt.Println(err)
	}

	for _, portBindings := range bindings {
		for _, portBinding := range portBindings {
			if portBinding.HostPort != "" {
				return true
			}
		}
	}

	return false
}

//take in timeout flag from cli as parameter
//return timeout if it is set,
//else return stop_grace_period if it is set,
//else return default 10s
func (s *Service) stopTimeout(timeout int) int {
	DEFAULTTIMEOUT := 10
	if timeout != 0 {
		return timeout
	}
	configTimeout := utils.DurationStrToSecondsInt(s.Config().StopGracePeriod)
	if configTimeout != nil {
		return *configTimeout
	}
	return DEFAULTTIMEOUT
}
