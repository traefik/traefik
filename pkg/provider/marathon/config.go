package marathon

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"

	"github.com/gambol99/go-marathon"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/config/label"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/provider"
	"github.com/traefik/traefik/v2/pkg/provider/constraints"
)

func (p *Provider) buildConfiguration(ctx context.Context, applications *marathon.Applications) *dynamic.Configuration {
	configurations := make(map[string]*dynamic.Configuration)

	for _, app := range applications.Apps {
		ctxApp := log.With(ctx, log.Str("applicationID", app.ID))
		logger := log.FromContext(ctxApp)

		extraConf, err := p.getConfiguration(app)
		if err != nil {
			logger.Errorf("Skip application: %v", err)
			continue
		}

		labels := stringValueMap(app.Labels)

		if app.Constraints != nil {
			for i, constraintParts := range *app.Constraints {
				key := constraints.MarathonConstraintPrefix + "-" + strconv.Itoa(i)
				labels[key] = strings.Join(constraintParts, ":")
			}
		}

		if !p.keepApplication(ctxApp, extraConf, labels) {
			continue
		}

		confFromLabel, err := label.DecodeConfiguration(labels)
		if err != nil {
			logger.Error(err)
			continue
		}

		var tcpOrUDP bool
		if len(confFromLabel.TCP.Routers) > 0 || len(confFromLabel.TCP.Services) > 0 {
			tcpOrUDP = true

			err := p.buildTCPServiceConfiguration(ctxApp, app, extraConf, confFromLabel.TCP)
			if err != nil {
				logger.Error(err)
				continue
			}
			provider.BuildTCPRouterConfiguration(ctxApp, confFromLabel.TCP)
		}

		if len(confFromLabel.UDP.Routers) > 0 || len(confFromLabel.UDP.Services) > 0 {
			tcpOrUDP = true

			err := p.buildUDPServiceConfiguration(ctxApp, app, extraConf, confFromLabel.UDP)
			if err != nil {
				logger.Error(err)
			} else {
				provider.BuildUDPRouterConfiguration(ctxApp, confFromLabel.UDP)
			}
		}

		if tcpOrUDP && len(confFromLabel.HTTP.Routers) == 0 &&
			len(confFromLabel.HTTP.Middlewares) == 0 &&
			len(confFromLabel.HTTP.Services) == 0 {
			configurations[app.ID] = confFromLabel
			continue
		}

		err = p.buildServiceConfiguration(ctxApp, app, extraConf, confFromLabel.HTTP)
		if err != nil {
			logger.Error(err)
			continue
		}

		model := struct {
			Name   string
			Labels map[string]string
		}{
			Name:   app.ID,
			Labels: labels,
		}

		serviceName := getServiceName(app)

		provider.BuildRouterConfiguration(ctxApp, confFromLabel.HTTP, serviceName, p.defaultRuleTpl, model)

		configurations[app.ID] = confFromLabel
	}

	return provider.Merge(ctx, configurations)
}

func getServiceName(app marathon.Application) string {
	return strings.ReplaceAll(strings.TrimPrefix(app.ID, "/"), "/", "_")
}

func (p *Provider) buildServiceConfiguration(ctx context.Context, app marathon.Application, extraConf configuration, conf *dynamic.HTTPConfiguration) error {
	appName := getServiceName(app)
	appCtx := log.With(ctx, log.Str("ApplicationID", appName))

	if len(conf.Services) == 0 {
		conf.Services = make(map[string]*dynamic.Service)
		lb := &dynamic.ServersLoadBalancer{}
		lb.SetDefaults()
		conf.Services[appName] = &dynamic.Service{
			LoadBalancer: lb,
		}
	}

	for serviceName, service := range conf.Services {
		var servers []dynamic.Server

		defaultServer := dynamic.Server{}
		defaultServer.SetDefaults()

		if len(service.LoadBalancer.Servers) > 0 {
			defaultServer = service.LoadBalancer.Servers[0]
		}

		for _, task := range app.Tasks {
			if !p.taskFilter(ctx, *task, app) {
				continue
			}
			server, err := p.getServer(app, *task, extraConf, defaultServer)
			if err != nil {
				log.FromContext(appCtx).Errorf("Skip task: %v", err)
				continue
			}
			servers = append(servers, server)
		}
		if len(servers) == 0 {
			return fmt.Errorf("no server for the service %s", serviceName)
		}
		service.LoadBalancer.Servers = servers
	}

	return nil
}

func (p *Provider) buildTCPServiceConfiguration(ctx context.Context, app marathon.Application, extraConf configuration, conf *dynamic.TCPConfiguration) error {
	appName := getServiceName(app)
	appCtx := log.With(ctx, log.Str("ApplicationID", appName))

	if len(conf.Services) == 0 {
		conf.Services = make(map[string]*dynamic.TCPService)
		lb := &dynamic.TCPServersLoadBalancer{}
		lb.SetDefaults()
		conf.Services[appName] = &dynamic.TCPService{
			LoadBalancer: lb,
		}
	}

	for serviceName, service := range conf.Services {
		var servers []dynamic.TCPServer

		defaultServer := dynamic.TCPServer{}

		if len(service.LoadBalancer.Servers) > 0 {
			defaultServer = service.LoadBalancer.Servers[0]
		}

		for _, task := range app.Tasks {
			if p.taskFilter(ctx, *task, app) {
				server, err := p.getTCPServer(app, *task, extraConf, defaultServer)
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

func (p *Provider) buildUDPServiceConfiguration(ctx context.Context, app marathon.Application, extraConf configuration, conf *dynamic.UDPConfiguration) error {
	appName := getServiceName(app)
	appCtx := log.With(ctx, log.Str("ApplicationID", appName))

	if len(conf.Services) == 0 {
		conf.Services = make(map[string]*dynamic.UDPService)
		lb := &dynamic.UDPServersLoadBalancer{}

		conf.Services[appName] = &dynamic.UDPService{
			LoadBalancer: lb,
		}
	}

	for serviceName, service := range conf.Services {
		var servers []dynamic.UDPServer

		defaultServer := dynamic.UDPServer{}

		if len(service.LoadBalancer.Servers) > 0 {
			defaultServer = service.LoadBalancer.Servers[0]
		}

		for _, task := range app.Tasks {
			if p.taskFilter(ctx, *task, app) {
				server, err := p.getUDPServer(app, *task, extraConf, defaultServer)
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

func (p *Provider) keepApplication(ctx context.Context, extraConf configuration, labels map[string]string) bool {
	logger := log.FromContext(ctx)

	// Filter disabled application.
	if !extraConf.Enable {
		logger.Debug("Filtering disabled Marathon application")
		return false
	}

	// Filter by constraints.
	matches, err := constraints.MatchLabels(labels, p.Constraints)
	if err != nil {
		logger.Errorf("Error matching constraints expression: %v", err)
		return false
	}
	if !matches {
		logger.Debugf("Marathon application filtered by constraint expression: %q", p.Constraints)
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

func (p *Provider) getTCPServer(app marathon.Application, task marathon.Task, extraConf configuration, defaultServer dynamic.TCPServer) (dynamic.TCPServer, error) {
	host, err := p.getServerHost(task, app, extraConf)
	if len(host) == 0 {
		return dynamic.TCPServer{}, err
	}

	port, err := getPort(task, app, defaultServer.Port)
	if err != nil {
		return dynamic.TCPServer{}, err
	}

	server := dynamic.TCPServer{
		Address: net.JoinHostPort(host, port),
	}

	return server, nil
}

func (p *Provider) getUDPServer(app marathon.Application, task marathon.Task, extraConf configuration, defaultServer dynamic.UDPServer) (dynamic.UDPServer, error) {
	host, err := p.getServerHost(task, app, extraConf)
	if len(host) == 0 {
		return dynamic.UDPServer{}, err
	}

	port, err := getPort(task, app, defaultServer.Port)
	if err != nil {
		return dynamic.UDPServer{}, err
	}

	server := dynamic.UDPServer{
		Address: net.JoinHostPort(host, port),
	}

	return server, nil
}

func (p *Provider) getServer(app marathon.Application, task marathon.Task, extraConf configuration, defaultServer dynamic.Server) (dynamic.Server, error) {
	host, err := p.getServerHost(task, app, extraConf)
	if len(host) == 0 {
		return dynamic.Server{}, err
	}

	port, err := getPort(task, app, defaultServer.Port)
	if err != nil {
		return dynamic.Server{}, err
	}

	server := dynamic.Server{
		URL: fmt.Sprintf("%s://%s", defaultServer.Scheme, net.JoinHostPort(host, port)),
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
		return "", fmt.Errorf("unable to process ports for %s %s: %w", app.ID, task.ID, err)
	}

	return strconv.Itoa(port), nil
}

// processPorts returns the configured port.
// An explicitly specified port is preferred. If none is specified, it selects
// one of the available port. The first such found port is returned unless an
// optional index is provided.
func processPorts(app marathon.Application, task marathon.Task, serverPort string) (int, error) {
	if len(serverPort) > 0 && !(strings.HasPrefix(serverPort, "index:") || strings.HasPrefix(serverPort, "name:")) {
		port, err := strconv.Atoi(serverPort)
		if err != nil {
			return 0, err
		}

		if port <= 0 {
			return 0, fmt.Errorf("explicitly specified port %d must be greater than zero", port)
		} else if port > 0 {
			return port, nil
		}
	}

	if strings.HasPrefix(serverPort, "name:") {
		name := strings.TrimPrefix(serverPort, "name:")
		port := retrieveNamedPort(app, name)

		if port == 0 {
			return 0, fmt.Errorf("no port with name %s", name)
		}

		return port, nil
	}

	ports := retrieveAvailablePorts(app, task)
	if len(ports) == 0 {
		return 0, errors.New("no port found")
	}

	portIndex := 0
	if strings.HasPrefix(serverPort, "index:") {
		indexString := strings.TrimPrefix(serverPort, "index:")
		index, err := strconv.Atoi(indexString)
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

func retrieveNamedPort(app marathon.Application, name string) int {
	// Using port definition if available
	if app.PortDefinitions != nil && len(*app.PortDefinitions) > 0 {
		for _, def := range *app.PortDefinitions {
			if def.Port != nil && *def.Port > 0 && def.Name == name {
				return *def.Port
			}
		}
	}

	// If using IP-per-task using this port definition
	if app.IPAddressPerTask != nil && app.IPAddressPerTask.Discovery != nil && len(*(app.IPAddressPerTask.Discovery.Ports)) > 0 {
		for _, def := range *(app.IPAddressPerTask.Discovery.Ports) {
			if def.Number > 0 && def.Name == name {
				return def.Number
			}
		}
	}

	return 0
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
