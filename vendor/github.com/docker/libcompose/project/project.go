package project

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/logger"
	"github.com/docker/libcompose/lookup"
	"github.com/docker/libcompose/project/events"
	"github.com/docker/libcompose/utils"
	"github.com/docker/libcompose/yaml"
)

// ComposeVersion is name of docker-compose.yml file syntax supported version
const ComposeVersion = "1.5.0"

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
	volumes       Volumes
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

	if context.ResourceLookup == nil {
		context.ResourceLookup = &lookup.FileResourceLookup{}
	}

	if context.EnvironmentLookup == nil {
		var envPath, absPath, cwd string
		var err error
		if len(context.ComposeFiles) > 0 {
			absPath, err = filepath.Abs(context.ComposeFiles[0])
			dir, _ := path.Split(absPath)
			envPath = filepath.Join(dir, ".env")
		} else {
			cwd, err = os.Getwd()
			envPath = filepath.Join(cwd, ".env")
		}

		if err != nil {
			log.Errorf("Could not get the rooted path name to the current directory: %v", err)
			return nil
		}
		context.EnvironmentLookup = &lookup.ComposableEnvLookup{
			Lookups: []config.EnvironmentLookup{
				&lookup.EnvfileLookup{
					Path: envPath,
				},
				&lookup.OsEnvLookup{},
			},
		}
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
	existing, ok := p.GetServiceConfig(name)
	if !ok {
		return nil, fmt.Errorf("Failed to find service: %s", name)
	}

	// Copy because we are about to modify the environment
	config := *existing

	if p.context.EnvironmentLookup != nil {
		parsedEnv := make([]string, 0, len(config.Environment))

		for _, env := range config.Environment {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) > 1 {
				parsedEnv = append(parsedEnv, env)
				continue
			} else {
				env = parts[0]
			}

			for _, value := range p.context.EnvironmentLookup.Lookup(env, &config) {
				parsedEnv = append(parsedEnv, value)
			}
		}

		config.Environment = parsedEnv

		// check the environment for extra build Args that are set but not given a value in the compose file
		for arg, value := range config.Build.Args {
			if *value == "\x00" {
				envValue := p.context.EnvironmentLookup.Lookup(arg, &config)
				// depending on what we get back we do different things
				switch l := len(envValue); l {
				case 0:
					delete(config.Build.Args, arg)
				case 1:
					parts := strings.SplitN(envValue[0], "=", 2)
					config.Build.Args[parts[0]] = &parts[1]
				default:
					return nil, fmt.Errorf("tried to set Build Arg %#v to multi-value %#v", arg, envValue)
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
	p.handleNetworkConfig()
	p.handleVolumeConfig()

	if p.context.NetworksFactory != nil {
		networks, err := p.context.NetworksFactory.Create(p.Name, p.NetworkConfigs, p.ServiceConfigs, p.isNetworkEnabled())
		if err != nil {
			return err
		}

		p.networks = networks
	}

	if p.context.VolumesFactory != nil {
		volumes, err := p.context.VolumesFactory.Create(p.Name, p.VolumeConfigs, p.ServiceConfigs, p.isVolumeEnabled())
		if err != nil {
			return err
		}

		p.volumes = volumes
	}

	return nil
}

func (p *Project) handleNetworkConfig() {
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
							Name:     "default",
							RealName: fmt.Sprintf("%s_%s", p.Name, "default"),
						},
					},
				}
				p.AddNetworkConfig("default", &config.NetworkConfig{})
			}
			// Consolidate the name of the network
			// FIXME(vdemeester) probably shouldn't be there, maybe move that to interface/factory
			for _, network := range serviceConfig.Networks.Networks {
				net, ok := p.NetworkConfigs[network.Name]
				if ok && net != nil {
					if net.External.External {
						network.RealName = network.Name
						if net.External.Name != "" {
							network.RealName = net.External.Name
						}
					} else {
						network.RealName = p.Name + "_" + network.Name
					}
				} else {
					network.RealName = p.Name + "_" + network.Name

					p.NetworkConfigs[network.Name] = &config.NetworkConfig{
						External: yaml.External{External: false},
					}
				}
				// Ignoring if we don't find the network, it will be catched later
			}
		}
	}
}

func (p *Project) isNetworkEnabled() bool {
	return p.configVersion == "2"
}

func (p *Project) handleVolumeConfig() {
	if p.isVolumeEnabled() {
		for _, serviceName := range p.ServiceConfigs.Keys() {
			serviceConfig, _ := p.ServiceConfigs.Get(serviceName)
			// Consolidate the name of the volume
			// FIXME(vdemeester) probably shouldn't be there, maybe move that to interface/factory
			if serviceConfig.Volumes == nil {
				continue
			}
			for _, volume := range serviceConfig.Volumes.Volumes {
				if !IsNamedVolume(volume.Source) {
					continue
				}

				vol, ok := p.VolumeConfigs[volume.Source]
				if !ok || vol == nil {
					continue
				}

				if vol.External.External {
					if vol.External.Name != "" {
						volume.Source = vol.External.Name
					}
				} else {
					volume.Source = p.Name + "_" + volume.Source
				}
			}
		}
	}
}

func (p *Project) isVolumeEnabled() bool {
	return p.configVersion == "2"
}

// initialize sets up required element for project before any action (on project and service).
// This means it's not needed to be called on Config for example.
func (p *Project) initialize(ctx context.Context) error {
	if p.networks != nil {
		if err := p.networks.Initialize(ctx); err != nil {
			return err
		}
	}
	if p.volumes != nil {
		if err := p.volumes.Initialize(ctx); err != nil {
			return err
		}
	}
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

// GetServiceConfig looks up a service config for a given service name, returning the ServiceConfig
// object and a bool flag indicating whether it was found
func (p *Project) GetServiceConfig(name string) (*config.ServiceConfig, bool) {
	return p.ServiceConfigs.Get(name)
}

// IsNamedVolume returns whether the specified volume (string) is a named volume or not.
func IsNamedVolume(volume string) bool {
	return !strings.HasPrefix(volume, ".") && !strings.HasPrefix(volume, "/") && !strings.HasPrefix(volume, "~")
}
