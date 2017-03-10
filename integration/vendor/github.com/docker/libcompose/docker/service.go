package docker

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	eventtypes "github.com/docker/engine-api/types/events"
	"github.com/docker/engine-api/types/filters"
	"github.com/docker/go-connections/nat"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/docker/builder"
	composeclient "github.com/docker/libcompose/docker/client"
	"github.com/docker/libcompose/labels"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/project/events"
	"github.com/docker/libcompose/project/options"
	"github.com/docker/libcompose/utils"
	dockerevents "github.com/vdemeester/docker-events"
)

// Service is a project.Service implementations.
type Service struct {
	name          string
	project       *project.Project
	serviceConfig *config.ServiceConfig
	clientFactory composeclient.Factory
	authLookup    AuthLookup

	// FIXME(vdemeester) remove this at some point
	context *Context
}

// NewService creates a service
func NewService(name string, serviceConfig *config.ServiceConfig, context *Context) *Service {
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

	imageName, err := s.ensureImageExists(ctx, options.NoBuild)
	if err != nil {
		return err
	}

	if len(containers) != 0 {
		return s.eachContainer(ctx, containers, func(c *Container) error {
			return s.recreateIfNeeded(ctx, imageName, c, options.NoRecreate, options.ForceRecreate)
		})
	}

	_, err = s.createOne(ctx, imageName)
	return err
}

func (s *Service) collectContainers(ctx context.Context) ([]*Container, error) {
	client := s.clientFactory.Create(s)
	containers, err := GetContainersByFilter(ctx, client, labels.SERVICE.Eq(s.name), labels.PROJECT.Eq(s.project.Name))
	if err != nil {
		return nil, err
	}

	result := []*Container{}

	for _, container := range containers {
		containerNumber, err := strconv.Atoi(container.Labels[labels.NUMBER.Str()])
		if err != nil {
			return nil, err
		}
		// Compose add "/" before name, so Name[1] will store actaul name.
		name := strings.SplitAfter(container.Names[0], "/")
		result = append(result, NewContainer(client, name[1], containerNumber, s))
	}

	return result, nil
}

func (s *Service) createOne(ctx context.Context, imageName string) (*Container, error) {
	containers, err := s.constructContainers(ctx, imageName, 1)
	if err != nil {
		return nil, err
	}

	return containers[0], err
}

func (s *Service) ensureImageExists(ctx context.Context, noBuild bool) (string, error) {
	exists, err := s.ImageExists(ctx)
	if err != nil {
		return "", err
	}
	if exists {
		return s.imageName(), nil
	}

	if s.Config().Build.Context != "" {
		if noBuild {
			return "", fmt.Errorf("Service %q needs to be built, but no-build was specified", s.name)
		}
		return s.imageName(), s.build(ctx, options.Build{})
	}

	return s.imageName(), s.Pull(ctx)
}

// ImageExists returns whether or not the service image already exists
func (s *Service) ImageExists(ctx context.Context) (bool, error) {
	dockerClient := s.clientFactory.Create(s)

	_, _, err := dockerClient.ImageInspectWithRaw(ctx, s.imageName(), false)
	if err == nil {
		return true, nil
	}
	if err != nil && client.IsErrImageNotFound(err) {
		return false, nil
	}

	return false, err
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
	}
	return builder.Build(ctx, s.imageName())
}

func (s *Service) constructContainers(ctx context.Context, imageName string, count int) ([]*Container, error) {
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
		containerName, containerNumber := namer.Next()

		c := NewContainer(client, containerName, containerNumber, s)

		dockerContainer, err := c.Create(ctx, imageName)
		if err != nil {
			return nil, err
		}

		logrus.Debugf("Created container %s: %v", dockerContainer.ID, dockerContainer.Name)

		result = append(result, NewContainer(client, containerName, containerNumber, s))
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
		imageName, err = s.ensureImageExists(ctx, options.NoBuild)
		if err != nil {
			return err
		}
	}

	return s.up(ctx, imageName, true, options)
}

// Run implements Service.Run. It runs a one of command within the service container.
// It always create a new container.
func (s *Service) Run(ctx context.Context, commandParts []string, options options.Run) (int, error) {
	imageName, err := s.ensureImageExists(ctx, false)
	if err != nil {
		return -1, err
	}

	client := s.clientFactory.Create(s)

	namer, err := NewNamer(ctx, client, s.project.Name, s.name, true)
	if err != nil {
		return -1, err
	}

	containerName, containerNumber := namer.Next()

	c := NewOneOffContainer(client, containerName, containerNumber, s)

	configOverride := &config.ServiceConfig{Command: commandParts, Tty: true, StdinOpen: true}

	c.CreateWithOverride(ctx, imageName, configOverride)

	if err := s.connectContainerToNetworks(ctx, c); err != nil {
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
func (s *Service) Info(ctx context.Context, qFlag bool) (project.InfoSet, error) {
	result := project.InfoSet{}
	containers, err := s.collectContainers(ctx)
	if err != nil {
		return nil, err
	}

	for _, c := range containers {
		info, err := c.Info(ctx, qFlag)
		if err != nil {
			return nil, err
		}
		result = append(result, info)
	}

	return result, nil
}

// Start implements Service.Start. It tries to start a container without creating it.
func (s *Service) Start(ctx context.Context) error {
	return s.collectContainersAndDo(ctx, func(c *Container) error {
		if err := s.connectContainerToNetworks(ctx, c); err != nil {
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
		c, err := s.createOne(ctx, imageName)
		if err != nil {
			return err
		}
		containers = []*Container{c}
	}

	return s.eachContainer(ctx, containers, func(c *Container) error {
		if create {
			if err := s.recreateIfNeeded(ctx, imageName, c, options.NoRecreate, options.ForceRecreate); err != nil {
				return err
			}
		}

		if err := s.connectContainerToNetworks(ctx, c); err != nil {
			return err
		}
		return c.Start(ctx)
	})
}

func (s *Service) connectContainerToNetworks(ctx context.Context, c *Container) error {
	connectedNetworks, err := c.Networks(ctx)
	if err != nil {
		return nil
	}
	if s.serviceConfig.Networks != nil {
		for _, network := range s.serviceConfig.Networks.Networks {
			existingNetwork, ok := connectedNetworks[network.Name]
			if ok {
				// FIXME(vdemeester) implement alias checking (to not disconnect/reconnect for nothing)
				aliasPresent := false
				for _, alias := range existingNetwork.Aliases {
					// FIXME(vdemeester) use shortID instead of ID
					ID, _ := c.ID()
					if alias == ID {
						aliasPresent = true
					}
				}
				if aliasPresent {
					continue
				}
				if err := c.NetworkDisconnect(ctx, network); err != nil {
					return err
				}
			}
			if err := c.NetworkConnect(ctx, network); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) recreateIfNeeded(ctx context.Context, imageName string, c *Container, noRecreate, forceRecreate bool) error {
	if noRecreate {
		return nil
	}
	outOfSync, err := c.OutOfSync(ctx, imageName)
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"outOfSync":     outOfSync,
		"ForceRecreate": forceRecreate,
		"NoRecreate":    noRecreate}).Debug("Going to decide if recreate is needed")

	if forceRecreate || outOfSync {
		logrus.Infof("Recreating %s", s.name)
		if _, err := c.Recreate(ctx, imageName); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) collectContainersAndDo(ctx context.Context, action func(*Container) error) error {
	containers, err := s.collectContainers(ctx)
	if err != nil {
		return err
	}
	return s.eachContainer(ctx, containers, action)
}

func (s *Service) eachContainer(ctx context.Context, containers []*Container, action func(*Container) error) error {

	tasks := utils.InParallel{}
	for _, container := range containers {
		task := func(container *Container) func() error {
			return func() error {
				return action(container)
			}
		}(container)

		tasks.Add(task)
	}

	return tasks.Wait()
}

// Stop implements Service.Stop. It stops any containers related to the service.
func (s *Service) Stop(ctx context.Context, timeout int) error {
	return s.collectContainersAndDo(ctx, func(c *Container) error {
		return c.Stop(ctx, timeout)
	})
}

// Restart implements Service.Restart. It restarts any containers related to the service.
func (s *Service) Restart(ctx context.Context, timeout int) error {
	return s.collectContainersAndDo(ctx, func(c *Container) error {
		return c.Restart(ctx, timeout)
	})
}

// Kill implements Service.Kill. It kills any containers related to the service.
func (s *Service) Kill(ctx context.Context, signal string) error {
	return s.collectContainersAndDo(ctx, func(c *Container) error {
		return c.Kill(ctx, signal)
	})
}

// Delete implements Service.Delete. It removes any containers related to the service.
func (s *Service) Delete(ctx context.Context, options options.Delete) error {
	return s.collectContainersAndDo(ctx, func(c *Container) error {
		return c.Delete(ctx, options.RemoveVolume)
	})
}

// Log implements Service.Log. It returns the docker logs for each container related to the service.
func (s *Service) Log(ctx context.Context, follow bool) error {
	return s.collectContainersAndDo(ctx, func(c *Container) error {
		return c.Log(ctx, follow)
	})
}

// Scale implements Service.Scale. It creates or removes containers to have the specified number
// of related container to the service to run.
func (s *Service) Scale(ctx context.Context, scale int, timeout int) error {
	if s.specificiesHostPort() {
		logrus.Warnf("The \"%s\" service specifies a port on the host. If multiple containers for this service are created on a single host, the port will clash.", s.Name())
	}

	foundCount := 0
	err := s.collectContainersAndDo(ctx, func(c *Container) error {
		foundCount++
		if foundCount > scale {
			err := c.Stop(ctx, timeout)
			if err != nil {
				return err
			}
			// FIXME(vdemeester) remove volume in scale by default ?
			return c.Delete(ctx, false)
		}
		return nil
	})

	if err != nil {
		return err
	}

	if foundCount != scale {
		imageName, err := s.ensureImageExists(ctx, false)
		if err != nil {
			return err
		}

		if _, err = s.constructContainers(ctx, imageName, scale); err != nil {
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

	return pullImage(ctx, s.clientFactory.Create(s), s, s.Config().Image)
}

// Pause implements Service.Pause. It puts into pause the container(s) related
// to the service.
func (s *Service) Pause(ctx context.Context) error {
	return s.collectContainersAndDo(ctx, func(c *Container) error {
		return c.Pause(ctx)
	})
}

// Unpause implements Service.Pause. It brings back from pause the container(s)
// related to the service.
func (s *Service) Unpause(ctx context.Context) error {
	return s.collectContainersAndDo(ctx, func(c *Container) error {
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
		return removeImage(ctx, s.clientFactory.Create(s), s.imageName())
	case "all":
		return removeImage(ctx, s.clientFactory.Create(s), s.imageName())
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
	return <-dockerevents.Monitor(ctx, client, types.EventsOptions{
		Filters: filter,
	}, func(m eventtypes.Message) {
		service := m.Actor.Attributes[labels.SERVICE.Str()]
		attributes := map[string]string{}
		for _, attr := range eventAttributes {
			attributes[attr] = m.Actor.Attributes[attr]
		}
		e := events.ContainerEvent{
			Service:    service,
			Event:      m.Action,
			Type:       m.Type,
			ID:         m.Actor.ID,
			Time:       time.Unix(m.Time, 0),
			Attributes: attributes,
		}
		evts <- e
	})
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
