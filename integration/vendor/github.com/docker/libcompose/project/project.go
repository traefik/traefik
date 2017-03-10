package project

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/logger"
	"github.com/docker/libcompose/project/events"
	"github.com/docker/libcompose/project/options"
	"github.com/docker/libcompose/utils"
	"github.com/docker/libcompose/yaml"
)

type wrapperAction func(*serviceWrapper, map[string]*serviceWrapper)
type serviceAction func(service Service) error

// Project holds libcompose project information.
type Project struct {
	Name           string
	ServiceConfigs *config.ServiceConfigs
	VolumeConfigs  map[string]*config.VolumeConfig
	NetworkConfigs map[string]*config.NetworkConfig
	Files          []string
	ReloadCallback func() error
	ParseOptions   *config.ParseOptions

	runtime       RuntimeProject
	networks      Networks
	configVersion string
	context       *Context
	reload        []string
	upCount       int
	listeners     []chan<- events.Event
	hasListeners  bool
}

// NewProject creates a new project with the specified context.
func NewProject(context *Context, runtime RuntimeProject, parseOptions *config.ParseOptions) *Project {
	p := &Project{
		context:        context,
		runtime:        runtime,
		ParseOptions:   parseOptions,
		ServiceConfigs: config.NewServiceConfigs(),
		VolumeConfigs:  make(map[string]*config.VolumeConfig),
		NetworkConfigs: make(map[string]*config.NetworkConfig),
	}

	if context.LoggerFactory == nil {
		context.LoggerFactory = &logger.NullLogger{}
	}

	context.Project = p

	p.listeners = []chan<- events.Event{NewDefaultListener(p)}

	return p
}

// Parse populates project information based on its context. It sets up the name,
// the composefile and the composebytes (the composefile content).
func (p *Project) Parse() error {
	err := p.context.open()
	if err != nil {
		return err
	}

	p.Name = p.context.ProjectName

	p.Files = p.context.ComposeFiles

	if len(p.Files) == 1 && p.Files[0] == "-" {
		p.Files = []string{"."}
	}

	if p.context.ComposeBytes != nil {
		for i, composeBytes := range p.context.ComposeBytes {
			file := ""
			if i < len(p.context.ComposeFiles) {
				file = p.Files[i]
			}
			if err := p.load(file, composeBytes); err != nil {
				return err
			}
		}
	}

	return nil
}

// CreateService creates a service with the specified name based. If there
// is no config in the project for this service, it will return an error.
func (p *Project) CreateService(name string) (Service, error) {
	existing, ok := p.ServiceConfigs.Get(name)
	if !ok {
		return nil, fmt.Errorf("Failed to find service: %s", name)
	}

	// Copy because we are about to modify the environment
	config := *existing

	if p.context.EnvironmentLookup != nil {
		parsedEnv := make([]string, 0, len(config.Environment))

		for _, env := range config.Environment {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) > 1 && parts[1] != "" {
				parsedEnv = append(parsedEnv, env)
				continue
			} else {
				env = parts[0]
			}

			for _, value := range p.context.EnvironmentLookup.Lookup(env, name, &config) {
				parsedEnv = append(parsedEnv, value)
			}
		}

		config.Environment = parsedEnv

		// check the environment for extra build Args that are set but not given a value in the compose file
		for arg, value := range config.Build.Args {
			if value == "\x00" {
				envValue := p.context.EnvironmentLookup.Lookup(arg, name, &config)
				// depending on what we get back we do different things
				switch l := len(envValue); l {
				case 0:
					delete(config.Build.Args, arg)
				case 1:
					parts := strings.SplitN(envValue[0], "=", 2)
					config.Build.Args[parts[0]] = parts[1]
				default:
					return nil, fmt.Errorf("Tried to set Build Arg %#v to multi-value %#v.", arg, envValue)
				}
			}
		}
	}

	return p.context.ServiceFactory.Create(p, name, &config)
}

// AddConfig adds the specified service config for the specified name.
func (p *Project) AddConfig(name string, config *config.ServiceConfig) error {
	p.Notify(events.ServiceAdd, name, nil)

	p.ServiceConfigs.Add(name, config)
	p.reload = append(p.reload, name)

	return nil
}

// AddVolumeConfig adds the specified volume config for the specified name.
func (p *Project) AddVolumeConfig(name string, config *config.VolumeConfig) error {
	p.Notify(events.VolumeAdd, name, nil)
	p.VolumeConfigs[name] = config
	return nil
}

// AddNetworkConfig adds the specified network config for the specified name.
func (p *Project) AddNetworkConfig(name string, config *config.NetworkConfig) error {
	p.Notify(events.NetworkAdd, name, nil)
	p.NetworkConfigs[name] = config
	return nil
}

// Load loads the specified byte array (the composefile content) and adds the
// service configuration to the project.
// FIXME is it needed ?
func (p *Project) Load(bytes []byte) error {
	return p.load("", bytes)
}

func (p *Project) load(file string, bytes []byte) error {
	version, serviceConfigs, volumeConfigs, networkConfigs, err := config.Merge(p.ServiceConfigs, p.context.EnvironmentLookup, p.context.ResourceLookup, file, bytes, p.ParseOptions)
	if err != nil {
		log.Errorf("Could not parse config for project %s : %v", p.Name, err)
		return err
	}

	p.configVersion = version

	for name, config := range volumeConfigs {
		err := p.AddVolumeConfig(name, config)
		if err != nil {
			return err
		}
	}

	for name, config := range networkConfigs {
		err := p.AddNetworkConfig(name, config)
		if err != nil {
			return err
		}
	}

	for name, config := range serviceConfigs {
		err := p.AddConfig(name, config)
		if err != nil {
			return err
		}
	}

	// Update network configuration a little bit
	if p.isNetworkEnabled() {
		for _, serviceName := range p.ServiceConfigs.Keys() {
			serviceConfig, _ := p.ServiceConfigs.Get(serviceName)
			if serviceConfig.NetworkMode != "" {
				continue
			}
			if serviceConfig.Networks == nil || len(serviceConfig.Networks.Networks) == 0 {
				// Add default as network
				serviceConfig.Networks = &yaml.Networks{
					Networks: []*yaml.Network{
						{
							Name: "default",
						},
					},
				}
			}
			// Consolidate the name of the network
			// FIXME(vdemeester) probably shouldn't be there, maybe move that to interface/factory
			for _, network := range serviceConfig.Networks.Networks {
				if net, ok := p.NetworkConfigs[network.Name]; ok {
					if net.External.External {
						network.RealName = network.Name
						if net.External.Name != "" {
							network.RealName = net.External.Name
						}
					} else {
						network.RealName = p.Name + "_" + network.Name
					}
				}
				// Ignoring if we don't find the network, it will be catched later
			}
		}
	}

	// FIXME(vdemeester) Not sure about this..
	if p.context.NetworksFactory != nil {
		networks, err := p.context.NetworksFactory.Create(p.Name, p.NetworkConfigs, p.ServiceConfigs, p.isNetworkEnabled())
		if err != nil {
			return err
		}

		p.networks = networks
	}

	return nil
}

func (p *Project) isNetworkEnabled() bool {
	return p.configVersion == "2"
}

// initialize sets up required element for project before any action (on project and service).
// This means it's not needed to be called on Config for example.
func (p *Project) initialize(ctx context.Context) error {
	if err := p.networks.Initialize(ctx); err != nil {
		return err
	}
	// TODO Initialize volumes
	return nil
}

func (p *Project) loadWrappers(wrappers map[string]*serviceWrapper, servicesToConstruct []string) error {
	for _, name := range servicesToConstruct {
		wrapper, err := newServiceWrapper(name, p)
		if err != nil {
			return err
		}
		wrappers[name] = wrapper
	}

	return nil
}

// Build builds the specified services (like docker build).
func (p *Project) Build(ctx context.Context, buildOptions options.Build, services ...string) error {
	return p.perform(events.ProjectBuildStart, events.ProjectBuildDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(wrappers, events.ServiceBuildStart, events.ServiceBuild, func(service Service) error {
			return service.Build(ctx, buildOptions)
		})
	}), nil)
}

// Create creates the specified services (like docker create).
func (p *Project) Create(ctx context.Context, options options.Create, services ...string) error {
	if options.NoRecreate && options.ForceRecreate {
		return fmt.Errorf("no-recreate and force-recreate cannot be combined")
	}
	return p.perform(events.ProjectCreateStart, events.ProjectCreateDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(wrappers, events.ServiceCreateStart, events.ServiceCreate, func(service Service) error {
			return service.Create(ctx, options)
		})
	}), nil)
}

// Stop stops the specified services (like docker stop).
func (p *Project) Stop(ctx context.Context, timeout int, services ...string) error {
	return p.perform(events.ProjectStopStart, events.ProjectStopDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(nil, events.ServiceStopStart, events.ServiceStop, func(service Service) error {
			return service.Stop(ctx, timeout)
		})
	}), nil)
}

// Down stops the specified services and clean related containers (like docker stop + docker rm).
func (p *Project) Down(ctx context.Context, opts options.Down, services ...string) error {
	if !opts.RemoveImages.Valid() {
		return fmt.Errorf("--rmi flag must be local, all or empty")
	}
	if err := p.Stop(ctx, 10, services...); err != nil {
		return err
	}
	if opts.RemoveOrphans {
		if err := p.runtime.RemoveOrphans(ctx, p.Name, p.ServiceConfigs); err != nil {
			return err
		}
	}
	if err := p.Delete(ctx, options.Delete{
		RemoveVolume: opts.RemoveVolume,
	}, services...); err != nil {
		return err
	}

	networks, err := p.context.NetworksFactory.Create(p.Name, p.NetworkConfigs, p.ServiceConfigs, p.isNetworkEnabled())
	if err != nil {
		return err
	}
	if err := networks.Remove(ctx); err != nil {
		return err
	}

	return p.forEach([]string{}, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(wrappers, events.NoEvent, events.NoEvent, func(service Service) error {
			return service.RemoveImage(ctx, opts.RemoveImages)
		})
	}), func(service Service) error {
		return service.Create(ctx, options.Create{})
	})
}

// RemoveOrphans implements project.RuntimeProject.RemoveOrphans.
// It does nothing by default as it is supposed to be overriden by specific implementation.
func (p *Project) RemoveOrphans(ctx context.Context) error {
	return nil
}

// Restart restarts the specified services (like docker restart).
func (p *Project) Restart(ctx context.Context, timeout int, services ...string) error {
	return p.perform(events.ProjectRestartStart, events.ProjectRestartDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(wrappers, events.ServiceRestartStart, events.ServiceRestart, func(service Service) error {
			return service.Restart(ctx, timeout)
		})
	}), nil)
}

// Port returns the public port for a port binding of the specified service.
func (p *Project) Port(ctx context.Context, index int, protocol, serviceName, privatePort string) (string, error) {
	service, err := p.CreateService(serviceName)
	if err != nil {
		return "", err
	}

	containers, err := service.Containers(ctx)
	if err != nil {
		return "", err
	}

	if index < 1 || index > len(containers) {
		return "", fmt.Errorf("Invalid index %d", index)
	}

	return containers[index-1].Port(ctx, fmt.Sprintf("%s/%s", privatePort, protocol))
}

// Ps list containers for the specified services.
func (p *Project) Ps(ctx context.Context, onlyID bool, services ...string) (InfoSet, error) {
	allInfo := InfoSet{}
	for _, name := range p.ServiceConfigs.Keys() {
		service, err := p.CreateService(name)
		if err != nil {
			return nil, err
		}

		info, err := service.Info(ctx, onlyID)
		if err != nil {
			return nil, err
		}

		allInfo = append(allInfo, info...)
	}
	return allInfo, nil
}

// Start starts the specified services (like docker start).
func (p *Project) Start(ctx context.Context, services ...string) error {
	return p.perform(events.ProjectStartStart, events.ProjectStartDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(wrappers, events.ServiceStartStart, events.ServiceStart, func(service Service) error {
			return service.Start(ctx)
		})
	}), nil)
}

// Run executes a one off command (like `docker run image command`).
func (p *Project) Run(ctx context.Context, serviceName string, commandParts []string, opts options.Run) (int, error) {
	if !p.ServiceConfigs.Has(serviceName) {
		return 1, fmt.Errorf("%s is not defined in the template", serviceName)
	}

	if err := p.initialize(ctx); err != nil {
		return 1, err
	}
	var exitCode int
	err := p.forEach([]string{}, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(wrappers, events.ServiceRunStart, events.ServiceRun, func(service Service) error {
			if service.Name() == serviceName {
				code, err := service.Run(ctx, commandParts, opts)
				exitCode = code
				return err
			}
			return nil
		})
	}), func(service Service) error {
		return service.Create(ctx, options.Create{})
	})
	return exitCode, err
}

// Up creates and starts the specified services (kinda like docker run).
func (p *Project) Up(ctx context.Context, options options.Up, services ...string) error {
	if err := p.initialize(ctx); err != nil {
		return err
	}
	return p.perform(events.ProjectUpStart, events.ProjectUpDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(wrappers, events.ServiceUpStart, events.ServiceUp, func(service Service) error {
			return service.Up(ctx, options)
		})
	}), func(service Service) error {
		return service.Create(ctx, options.Create)
	})
}

// Log aggregates and prints out the logs for the specified services.
func (p *Project) Log(ctx context.Context, follow bool, services ...string) error {
	return p.forEach(services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(nil, events.NoEvent, events.NoEvent, func(service Service) error {
			return service.Log(ctx, follow)
		})
	}), nil)
}

// Scale scales the specified services.
func (p *Project) Scale(ctx context.Context, timeout int, servicesScale map[string]int) error {
	// This code is a bit verbose but I wanted to parse everything up front
	order := make([]string, 0, 0)
	services := make(map[string]Service)

	for name := range servicesScale {
		if !p.ServiceConfigs.Has(name) {
			return fmt.Errorf("%s is not defined in the template", name)
		}

		service, err := p.CreateService(name)
		if err != nil {
			return fmt.Errorf("Failed to lookup service: %s: %v", service, err)
		}

		order = append(order, name)
		services[name] = service
	}

	for _, name := range order {
		scale := servicesScale[name]
		log.Infof("Setting scale %s=%d...", name, scale)
		err := services[name].Scale(ctx, scale, timeout)
		if err != nil {
			return fmt.Errorf("Failed to set the scale %s=%d: %v", name, scale, err)
		}
	}
	return nil
}

// Pull pulls the specified services (like docker pull).
func (p *Project) Pull(ctx context.Context, services ...string) error {
	return p.forEach(services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(nil, events.ServicePullStart, events.ServicePull, func(service Service) error {
			return service.Pull(ctx)
		})
	}), nil)
}

// listStoppedContainers lists the stopped containers for the specified services.
func (p *Project) listStoppedContainers(ctx context.Context, services ...string) ([]string, error) {
	stoppedContainers := []string{}
	err := p.forEach(services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(nil, events.NoEvent, events.NoEvent, func(service Service) error {
			containers, innerErr := service.Containers(ctx)
			if innerErr != nil {
				return innerErr
			}

			for _, container := range containers {
				running, innerErr := container.IsRunning(ctx)
				if innerErr != nil {
					log.Error(innerErr)
				}
				if !running {
					containerID, innerErr := container.ID()
					if innerErr != nil {
						log.Error(innerErr)
					}
					stoppedContainers = append(stoppedContainers, containerID)
				}
			}

			return nil
		})
	}), nil)
	if err != nil {
		return nil, err
	}
	return stoppedContainers, nil
}

// Delete removes the specified services (like docker rm).
func (p *Project) Delete(ctx context.Context, options options.Delete, services ...string) error {
	stoppedContainers, err := p.listStoppedContainers(ctx, services...)
	if err != nil {
		return err
	}
	if len(stoppedContainers) == 0 {
		p.Notify(events.ProjectDeleteDone, "", nil)
		fmt.Println("No stopped containers")
		return nil
	}
	if options.BeforeDeleteCallback != nil && !options.BeforeDeleteCallback(stoppedContainers) {
		return nil
	}
	return p.perform(events.ProjectDeleteStart, events.ProjectDeleteDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(nil, events.ServiceDeleteStart, events.ServiceDelete, func(service Service) error {
			return service.Delete(ctx, options)
		})
	}), nil)
}

// Kill kills the specified services (like docker kill).
func (p *Project) Kill(ctx context.Context, signal string, services ...string) error {
	return p.perform(events.ProjectKillStart, events.ProjectKillDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(nil, events.ServiceKillStart, events.ServiceKill, func(service Service) error {
			return service.Kill(ctx, signal)
		})
	}), nil)
}

// Pause pauses the specified services containers (like docker pause).
func (p *Project) Pause(ctx context.Context, services ...string) error {
	return p.perform(events.ProjectPauseStart, events.ProjectPauseDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(nil, events.ServicePauseStart, events.ServicePause, func(service Service) error {
			return service.Pause(ctx)
		})
	}), nil)
}

// Unpause pauses the specified services containers (like docker pause).
func (p *Project) Unpause(ctx context.Context, services ...string) error {
	return p.perform(events.ProjectUnpauseStart, events.ProjectUnpauseDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(nil, events.ServiceUnpauseStart, events.ServiceUnpause, func(service Service) error {
			return service.Unpause(ctx)
		})
	}), nil)
}

func (p *Project) perform(start, done events.EventType, services []string, action wrapperAction, cycleAction serviceAction) error {
	p.Notify(start, "", nil)

	err := p.forEach(services, action, cycleAction)

	p.Notify(done, "", nil)
	return err
}

func isSelected(wrapper *serviceWrapper, selected map[string]bool) bool {
	return len(selected) == 0 || selected[wrapper.name]
}

func (p *Project) forEach(services []string, action wrapperAction, cycleAction serviceAction) error {
	selected := make(map[string]bool)
	wrappers := make(map[string]*serviceWrapper)

	for _, s := range services {
		selected[s] = true
	}

	return p.traverse(true, selected, wrappers, action, cycleAction)
}

func (p *Project) startService(wrappers map[string]*serviceWrapper, history []string, selected, launched map[string]bool, wrapper *serviceWrapper, action wrapperAction, cycleAction serviceAction) error {
	if launched[wrapper.name] {
		return nil
	}

	launched[wrapper.name] = true
	history = append(history, wrapper.name)

	for _, dep := range wrapper.service.DependentServices() {
		target := wrappers[dep.Target]
		if target == nil {
			log.Debugf("Failed to find %s", dep.Target)
			return fmt.Errorf("Service '%s' has a link to service '%s' which is undefined", wrapper.name, dep.Target)
		}

		if utils.Contains(history, dep.Target) {
			cycle := strings.Join(append(history, dep.Target), "->")
			if dep.Optional {
				log.Debugf("Ignoring cycle for %s", cycle)
				wrapper.IgnoreDep(dep.Target)
				if cycleAction != nil {
					var err error
					log.Debugf("Running cycle action for %s", cycle)
					err = cycleAction(target.service)
					if err != nil {
						return err
					}
				}
			} else {
				return fmt.Errorf("Cycle detected in path %s", cycle)
			}

			continue
		}

		err := p.startService(wrappers, history, selected, launched, target, action, cycleAction)
		if err != nil {
			return err
		}
	}

	if isSelected(wrapper, selected) {
		log.Debugf("Launching action for %s", wrapper.name)
		go action(wrapper, wrappers)
	} else {
		wrapper.Ignore()
	}

	return nil
}

func (p *Project) traverse(start bool, selected map[string]bool, wrappers map[string]*serviceWrapper, action wrapperAction, cycleAction serviceAction) error {
	restart := false
	wrapperList := []string{}

	if start {
		for _, name := range p.ServiceConfigs.Keys() {
			wrapperList = append(wrapperList, name)
		}
	} else {
		for _, wrapper := range wrappers {
			if err := wrapper.Reset(); err != nil {
				return err
			}
		}
		wrapperList = p.reload
	}

	p.loadWrappers(wrappers, wrapperList)
	p.reload = []string{}

	// check service name
	for s := range selected {
		if wrappers[s] == nil {
			return errors.New("No such service: " + s)
		}
	}

	launched := map[string]bool{}

	for _, wrapper := range wrappers {
		if err := p.startService(wrappers, []string{}, selected, launched, wrapper, action, cycleAction); err != nil {
			return err
		}
	}

	var firstError error

	for _, wrapper := range wrappers {
		if !isSelected(wrapper, selected) {
			continue
		}
		if err := wrapper.Wait(); err == ErrRestart {
			restart = true
		} else if err != nil {
			log.Errorf("Failed to start: %s : %v", wrapper.name, err)
			if firstError == nil {
				firstError = err
			}
		}
	}

	if restart {
		if p.ReloadCallback != nil {
			if err := p.ReloadCallback(); err != nil {
				log.Errorf("Failed calling callback: %v", err)
			}
		}
		return p.traverse(false, selected, wrappers, action, cycleAction)
	}
	return firstError
}

// AddListener adds the specified listener to the project.
// This implements implicitly events.Emitter.
func (p *Project) AddListener(c chan<- events.Event) {
	if !p.hasListeners {
		for _, l := range p.listeners {
			close(l)
		}
		p.hasListeners = true
		p.listeners = []chan<- events.Event{c}
	} else {
		p.listeners = append(p.listeners, c)
	}
}

// Notify notifies all project listener with the specified eventType, service name and datas.
// This implements implicitly events.Notifier interface.
func (p *Project) Notify(eventType events.EventType, serviceName string, data map[string]string) {
	if eventType == events.NoEvent {
		return
	}

	event := events.Event{
		EventType:   eventType,
		ServiceName: serviceName,
		Data:        data,
	}

	for _, l := range p.listeners {
		l <- event
	}
}
