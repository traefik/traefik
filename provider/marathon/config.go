package marathon

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"

	"github.com/containous/traefik/config"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/label"
	"github.com/gambol99/go-marathon"
)

func (p *Provider) buildConfiguration(ctx context.Context, applications *marathon.Applications) *config.Configuration {
	configurations := make(map[string]*config.Configuration)

	for _, app := range applications.Apps {
		ctxApp := log.With(ctx, log.Str("applicationID", app.ID))
		logger := log.FromContext(ctxApp)

		extraConf, err := p.getConfiguration(app)
		if err != nil {
			logger.Errorf("Skip application: %v", err)
			continue
		}

		if !p.keepApplication(ctxApp, extraConf) {
			continue
		}

		confFromLabel, err := label.DecodeConfiguration(stringValueMap(app.Labels))
		if err != nil {
			logger.Error(err)
			continue
		}

		err = p.buildServiceConfiguration(ctxApp, app, extraConf, confFromLabel)
		if err != nil {
			logger.Error(err)
			continue
		}

		model := struct {
			Name   string
			Labels map[string]string
		}{
			Name:   app.ID,
			Labels: stringValueMap(app.Labels),
		}

		serviceName := getServiceName(app)

		provider.BuildRouterConfiguration(ctxApp, confFromLabel, serviceName, p.defaultRuleTpl, model)

		configurations[app.ID] = confFromLabel
	}

	return provider.Merge(ctx, configurations)
}

func getServiceName(app marathon.Application) string {
	return strings.Replace(strings.TrimPrefix(app.ID, "/"), "/", "_", -1)
}

func (p *Provider) buildServiceConfiguration(ctx context.Context, app marathon.Application, extraConf configuration, conf *config.Configuration) error {
	appName := getServiceName(app)
	appCtx := log.With(ctx, log.Str("ApplicationID", appName))

	if len(conf.Services) == 0 {
		conf.Services = make(map[string]*config.Service)
		lb := &config.LoadBalancerService{}
		lb.SetDefaults()
		conf.Services[appName] = &config.Service{
			LoadBalancer: lb,
		}
	}

	for serviceName, service := range conf.Services {
		var servers []config.Server

		defaultServer := config.Server{}
		defaultServer.SetDefaults()

		// Server dans labels?
		if len(service.LoadBalancer.Servers) > 0 {
			defaultServer = service.LoadBalancer.Servers[0]
		}

		for _, task := range app.Tasks {
			if p.taskFilter(ctx, *task, app) {
				server, err := p.getServer(app, *task, extraConf, defaultServer)
				if err != nil {
					log.FromContext(appCtx).Errorf("Skip task: %v", err)
					continue
				}
				servers = append(servers, server)
			}
		}
		if len(servers) == 0 {
			return fmt.Errorf("no server for the service %s", serviceName)
		}
		service.LoadBalancer.Servers = servers
	}

	return nil
}

func (p *Provider) keepApplication(ctx context.Context, extraConf configuration) bool {
	logger := log.FromContext(ctx)

	// Filter disabled application.
	if !extraConf.Enable {
		logger.Debug("Filtering disabled Marathon application")
		return false
	}

	// Filter by constraints.
	if ok, failingConstraint := p.MatchConstraints(extraConf.Tags); !ok {
		if failingConstraint != nil {
			logger.Debugf("Filtering Marathon application, pruned by %q constraint", failingConstraint.String())
		}
		return false
	}

	return true
}

func (p *Provider) taskFilter(ctx context.Context, task marathon.Task, application marathon.Application) bool {
	if task.State != string(taskStateRunning) {
		return false
	}

	if ready := p.readyChecker.Do(task, application); !ready {
		log.FromContext(ctx).Infof("Filtering unready task %s from application %s", task.ID, application.ID)
		return false
	}

	return true
}

func (p *Provider) getServer(app marathon.Application, task marathon.Task, extraConf configuration, defaultServer config.Server) (config.Server, error) {
	host, err := p.getServerHost(task, app, extraConf)
	if len(host) == 0 {
		return config.Server{}, err
	}

	port, err := getPort(task, app, defaultServer.Port)
	if err != nil {
		return config.Server{}, err
	}

	server := config.Server{
		URL:    fmt.Sprintf("%s://%s", defaultServer.Scheme, net.JoinHostPort(host, port)),
		Weight: 1,
	}

	return server, nil
}

func (p *Provider) getServerHost(task marathon.Task, app marathon.Application, extraConf configuration) (string, error) {
	networks := app.Networks
	var hostFlag bool

	if networks == nil {
		hostFlag = app.IPAddressPerTask == nil
	} else {
		hostFlag = (*networks)[0].Mode != marathon.ContainerNetworkMode
	}

	if hostFlag || p.ForceTaskHostname {
		if len(task.Host) == 0 {
			return "", fmt.Errorf("host is undefined for task %q app %q", task.ID, app.ID)
		}
		return task.Host, nil
	}

	numTaskIPAddresses := len(task.IPAddresses)
	switch numTaskIPAddresses {
	case 0:
		return "", fmt.Errorf("missing IP address for Marathon application %s on task %s", app.ID, task.ID)
	case 1:
		return task.IPAddresses[0].IPAddress, nil
	default:
		if extraConf.Marathon.IPAddressIdx == math.MinInt32 {
			return "", fmt.Errorf("found %d task IP addresses but missing IP address index for Marathon application %s on task %s",
				numTaskIPAddresses, app.ID, task.ID)
		}
		if extraConf.Marathon.IPAddressIdx < 0 || extraConf.Marathon.IPAddressIdx > numTaskIPAddresses {
			return "", fmt.Errorf("cannot use IP address index to select from %d task IP addresses for Marathon application %s on task %s",
				numTaskIPAddresses, app.ID, task.ID)
		}

		return task.IPAddresses[extraConf.Marathon.IPAddressIdx].IPAddress, nil
	}
}

func getPort(task marathon.Task, app marathon.Application, serverPort string) (string, error) {
	port, err := processPorts(app, task, serverPort)
	if err != nil {
		return "", fmt.Errorf("unable to process ports for %s %s: %v", app.ID, task.ID, err)
	}

	return strconv.Itoa(port), nil
}

// processPorts returns the configured port.
// An explicitly specified port is preferred. If none is specified, it selects
// one of the available port. The first such found port is returned unless an
// optional index is provided.
func processPorts(app marathon.Application, task marathon.Task, serverPort string) (int, error) {
	if len(serverPort) > 0 && !strings.HasPrefix(serverPort, "index:") {
		port, err := strconv.Atoi(serverPort)
		if err != nil {
			return 0, err
		}

		if port <= 0 {
			return 0, fmt.Errorf("explicitly specified port %d must be larger than zero", port)
		} else if port > 0 {
			return port, nil
		}
	}

	ports := retrieveAvailablePorts(app, task)
	if len(ports) == 0 {
		return 0, errors.New("no port found")
	}

	portIndex := 0
	if strings.HasPrefix(serverPort, "index:") {
		split := strings.SplitN(serverPort, ":", 2)
		index, err := strconv.Atoi(split[1])
		if err != nil {
			return 0, err
		}

		if index < 0 || index > len(ports)-1 {
			return 0, fmt.Errorf("index %d must be within range (0, %d)", index, len(ports)-1)
		}
		portIndex = index
	}
	return ports[portIndex], nil
}

func retrieveAvailablePorts(app marathon.Application, task marathon.Task) []int {
	// Using default port configuration
	if len(task.Ports) > 0 {
		return task.Ports
	}

	// Using port definition if available
	if app.PortDefinitions != nil && len(*app.PortDefinitions) > 0 {
		var ports []int
		for _, def := range *app.PortDefinitions {
			if def.Port != nil {
				ports = append(ports, *def.Port)
			}
		}
		return ports
	}

	// If using IP-per-task using this port definition
	if app.IPAddressPerTask != nil && app.IPAddressPerTask.Discovery != nil && len(*(app.IPAddressPerTask.Discovery.Ports)) > 0 {
		var ports []int
		for _, def := range *(app.IPAddressPerTask.Discovery.Ports) {
			ports = append(ports, def.Number)
		}
		return ports
	}

	return []int{}
}
