package consulcatalog

import (
	"context"
	gtls "crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"

	"github.com/hashicorp/consul/agent/connect"
	"github.com/hashicorp/consul/api"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/config/label"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/provider"
	"github.com/traefik/traefik/v2/pkg/provider/constraints"
	"github.com/traefik/traefik/v2/pkg/tls"
)

func (p *Provider) buildConfiguration(ctx context.Context, items []itemData, certInfo *connectCert) *dynamic.Configuration {
	configurations := make(map[string]*dynamic.Configuration)

	transports := make(map[string]*dynamic.ServersTransport)

	for _, item := range items {
		svcName := provider.Normalize(item.Node + "-" + item.Name + "-" + item.ID)
		ctxSvc := log.With(ctx, log.Str(log.ServiceName, svcName))

		if !p.keepContainer(ctxSvc, item) {
			continue
		}

		logger := log.FromContext(ctxSvc)

		confFromLabel, err := label.DecodeConfiguration(item.Labels)
		if err != nil {
			logger.Error(err)
			continue
		}

		var tcpOrUDP bool
		if len(confFromLabel.TCP.Routers) > 0 || len(confFromLabel.TCP.Services) > 0 {
			tcpOrUDP = true

			err := p.buildTCPServiceConfiguration(ctxSvc, item, confFromLabel.TCP)
			if err != nil {
				logger.Error(err)
				continue
			}
			provider.BuildTCPRouterConfiguration(ctxSvc, confFromLabel.TCP)
		}

		if len(confFromLabel.UDP.Routers) > 0 || len(confFromLabel.UDP.Services) > 0 {
			tcpOrUDP = true

			err := p.buildUDPServiceConfiguration(ctxSvc, item, confFromLabel.UDP)
			if err != nil {
				logger.Error(err)
				continue
			}
			provider.BuildUDPRouterConfiguration(ctxSvc, confFromLabel.UDP)
		}

		if tcpOrUDP && len(confFromLabel.HTTP.Routers) == 0 &&
			len(confFromLabel.HTTP.Middlewares) == 0 &&
			len(confFromLabel.HTTP.Services) == 0 {
			configurations[svcName] = confFromLabel
			continue
		}

		if item.ConnectEnabled {
			transportName := connectTransportName(item.Name)
			if transports[transportName] == nil {
				transports[transportName] = certInfo.serverTransport(item)
			}
			if len(confFromLabel.HTTP.ServersTransports) == 0 {
				confFromLabel.HTTP.ServersTransports = make(map[string]*dynamic.ServersTransport)
			}
			confFromLabel.HTTP.ServersTransports[transportName] = transports[transportName]
		}

		err = p.buildServiceConfiguration(ctxSvc, item, confFromLabel.HTTP)
		if err != nil {
			logger.Error(err)
			continue
		}

		model := struct {
			Name   string
			Labels map[string]string
		}{
			Name:   item.Name,
			Labels: item.Labels,
		}

		provider.BuildRouterConfiguration(ctx, confFromLabel.HTTP, provider.Normalize(item.Name), p.defaultRuleTpl, model)

		configurations[svcName] = confFromLabel
	}

	return provider.Merge(ctx, configurations)
}

func connectTransportName(n string) string {
	return "connect-tls-" + n
}

type connectCert struct {
	root []string
	leaf keyPair
	err  error
}

func (c *connectCert) getRoot() []tls.FileOrContent {
	var result []tls.FileOrContent
	for _, r := range c.root {
		result = append(result, tls.FileOrContent(r))
	}
	return result
}

func (c *connectCert) getLeaf() tls.Certificate {
	return tls.Certificate{
		CertFile: tls.FileOrContent(c.leaf.cert),
		KeyFile:  tls.FileOrContent(c.leaf.key),
	}
}

func (c *connectCert) isReady() bool {
	return len(c.root) > 0 && c.leaf.cert != "" && c.leaf.key != ""
}

func (c *connectCert) equals(other *connectCert) bool {
	if c == nil && other == nil {
		return true
	}
	if c == nil || other == nil {
		return false
	}
	if len(c.root) != len(other.root) {
		return false
	}
	for i, v := range c.root {
		if v != other.root[i] {
			return false
		}
	}
	return c.leaf == other.leaf
}

func (c *connectCert) serverTransport(item itemData) *dynamic.ServersTransport {
	return &dynamic.ServersTransport{
		// InsecureSkipVerify is needed because Go wants to verify a hostname otherwise
		InsecureSkipVerify: true,
		RootCAs:            c.getRoot(),
		Certificates: tls.Certificates{
			c.getLeaf(),
		},
		VerifyPeerCertificate: func(cfg *gtls.Config, rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			// We should use RootCAs here, but consul expect that in ClientCAs (don't ask)
			// https://github.com/hashicorp/consul/blob/cd428060f6547afddd9e0060c07b2a2c862da801/connect/tls.go#L279-L282
			// called via https://github.com/hashicorp/consul/blob/cd428060f6547afddd9e0060c07b2a2c862da801/connect/tls.go#L258
			cfg.ClientCAs = cfg.RootCAs
			cert, err := verifyChain(cfg, rawCerts, true)
			if err != nil {
				return err
			}
			certs := []*x509.Certificate{0: cert}
			uri := &connect.SpiffeIDService{
				Namespace:  item.Namespace,
				Datacenter: item.Datacenter,
				Service:    item.Name,
			}
			return verifyServerCertMatchesURI(certs, uri)
		},
	}
}

func (p *Provider) keepContainer(ctx context.Context, item itemData) bool {
	logger := log.FromContext(ctx)

	if !item.ExtraConf.Enable {
		logger.Debug("Filtering disabled item")
		return false
	}

	matches, err := constraints.MatchTags(item.Tags, p.Constraints)
	if err != nil {
		logger.Errorf("Error matching constraints expression: %v", err)
		return false
	}
	if !matches {
		logger.Debugf("Container pruned by constraint expression: %q", p.Constraints)
		return false
	}

	if item.Status != api.HealthPassing && item.Status != api.HealthWarning {
		logger.Debug("Filtering unhealthy or starting item")
		return false
	}

	return true
}

func (p *Provider) buildTCPServiceConfiguration(ctx context.Context, item itemData, configuration *dynamic.TCPConfiguration) error {
	if len(configuration.Services) == 0 {
		configuration.Services = make(map[string]*dynamic.TCPService)

		lb := &dynamic.TCPServersLoadBalancer{}
		lb.SetDefaults()

		configuration.Services[provider.Normalize(item.Name)] = &dynamic.TCPService{
			LoadBalancer: lb,
		}
	}

	for name, service := range configuration.Services {
		ctxSvc := log.With(ctx, log.Str(log.ServiceName, name))
		err := p.addServerTCP(ctxSvc, item, service.LoadBalancer)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) buildUDPServiceConfiguration(ctx context.Context, item itemData, configuration *dynamic.UDPConfiguration) error {
	if len(configuration.Services) == 0 {
		configuration.Services = make(map[string]*dynamic.UDPService)

		lb := &dynamic.UDPServersLoadBalancer{}

		configuration.Services[provider.Normalize(item.Name)] = &dynamic.UDPService{
			LoadBalancer: lb,
		}
	}

	for name, service := range configuration.Services {
		ctxSvc := log.With(ctx, log.Str(log.ServiceName, name))
		err := p.addServerUDP(ctxSvc, item, service.LoadBalancer)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) buildServiceConfiguration(ctx context.Context, item itemData, configuration *dynamic.HTTPConfiguration) error {
	if len(configuration.Services) == 0 {
		configuration.Services = make(map[string]*dynamic.Service)

		lb := &dynamic.ServersLoadBalancer{}
		lb.SetDefaults()

		configuration.Services[provider.Normalize(item.Name)] = &dynamic.Service{
			LoadBalancer: lb,
		}
	}

	for name, service := range configuration.Services {
		ctxSvc := log.With(ctx, log.Str(log.ServiceName, name))
		err := p.addServer(ctxSvc, item, service.LoadBalancer)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) addServerTCP(ctx context.Context, item itemData, loadBalancer *dynamic.TCPServersLoadBalancer) error {
	if loadBalancer == nil {
		return errors.New("load-balancer is not defined")
	}

	var port string
	if len(loadBalancer.Servers) > 0 {
		port = loadBalancer.Servers[0].Port
	}

	if len(loadBalancer.Servers) == 0 {
		loadBalancer.Servers = []dynamic.TCPServer{{}}
	}

	if item.Port != "" && port == "" {
		port = item.Port
	}
	loadBalancer.Servers[0].Port = ""

	if port == "" {
		return errors.New("port is missing")
	}

	if item.Address == "" {
		return errors.New("address is missing")
	}

	loadBalancer.Servers[0].Address = net.JoinHostPort(item.Address, port)
	return nil
}

func (p *Provider) addServerUDP(ctx context.Context, item itemData, loadBalancer *dynamic.UDPServersLoadBalancer) error {
	if loadBalancer == nil {
		return errors.New("load-balancer is not defined")
	}

	if len(loadBalancer.Servers) == 0 {
		loadBalancer.Servers = []dynamic.UDPServer{{}}
	}

	var port string
	if item.Port != "" {
		port = item.Port
		loadBalancer.Servers[0].Port = ""
	}

	if port == "" {
		return errors.New("port is missing")
	}

	if item.Address == "" {
		return errors.New("address is missing")
	}

	loadBalancer.Servers[0].Address = net.JoinHostPort(item.Address, port)
	return nil
}

func (p *Provider) addServer(ctx context.Context, item itemData, loadBalancer *dynamic.ServersLoadBalancer) error {
	if loadBalancer == nil {
		return errors.New("load-balancer is not defined")
	}

	var port string
	if len(loadBalancer.Servers) > 0 {
		port = loadBalancer.Servers[0].Port
	}

	if len(loadBalancer.Servers) == 0 {
		server := dynamic.Server{}
		server.SetDefaults()

		loadBalancer.Servers = []dynamic.Server{server}
	}

	if item.Port != "" && port == "" {
		port = item.Port
	}
	loadBalancer.Servers[0].Port = ""

	if port == "" {
		return errors.New("port is missing")
	}

	if item.Address == "" {
		return errors.New("address is missing")
	}

	if item.ConnectEnabled {
		loadBalancer.ServersTransport = connectTransportName(item.Name)
		loadBalancer.Servers[0].Scheme = "https"
	}

	loadBalancer.Servers[0].URL = fmt.Sprintf("%s://%s", loadBalancer.Servers[0].Scheme, net.JoinHostPort(item.Address, port))
	loadBalancer.Servers[0].Scheme = ""

	return nil
}
