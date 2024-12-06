package crd

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/provider"
	traefikv1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	"github.com/traefik/traefik/v3/pkg/tls"
	corev1 "k8s.io/api/core/v1"
)

func (p *Provider) loadIngressRouteTCPConfiguration(ctx context.Context, client Client, tlsConfigs map[string]*tls.CertAndStores) *dynamic.TCPConfiguration {
	conf := &dynamic.TCPConfiguration{
		Routers:           map[string]*dynamic.TCPRouter{},
		Middlewares:       map[string]*dynamic.TCPMiddleware{},
		Services:          map[string]*dynamic.TCPService{},
		ServersTransports: map[string]*dynamic.TCPServersTransport{},
	}

	for _, ingressRouteTCP := range client.GetIngressRouteTCPs() {
		logger := log.Ctx(ctx).With().Str("ingress", ingressRouteTCP.Name).Str("namespace", ingressRouteTCP.Namespace).Logger()

		if !shouldProcessIngress(p.IngressClass, ingressRouteTCP.Annotations[annotationKubernetesIngressClass]) {
			continue
		}

		if ingressRouteTCP.Spec.TLS != nil && !ingressRouteTCP.Spec.TLS.Passthrough {
			err := getTLSTCP(ctx, ingressRouteTCP, client, tlsConfigs)
			if err != nil {
				logger.Error().Err(err).Msg("Error configuring TLS")
			}
		}

		ingressName := ingressRouteTCP.Name
		if len(ingressName) == 0 {
			ingressName = ingressRouteTCP.GenerateName
		}

		for _, route := range ingressRouteTCP.Spec.Routes {
			if len(route.Match) == 0 {
				logger.Error().Msg("Empty match rule")
				continue
			}

			key, err := makeServiceKey(route.Match, ingressName)
			if err != nil {
				logger.Error().Err(err).Send()
				continue
			}

			mds, err := p.makeMiddlewareTCPKeys(ctx, ingressRouteTCP.Namespace, route.Middlewares)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to create middleware keys")
				continue
			}

			serviceName := makeID(ingressRouteTCP.Namespace, key)

			for _, service := range route.Services {
				balancerServerTCP, err := p.createLoadBalancerServerTCP(client, ingressRouteTCP.Namespace, service)
				if err != nil {
					logger.Error().
						Str("serviceName", service.Name).
						Stringer("servicePort", &service.Port).
						Err(err).
						Msg("Cannot create service")
					continue
				}

				// If there is only one service defined, we skip the creation of the load balancer of services,
				// i.e. the service on top is directly a load balancer of servers.
				if len(route.Services) == 1 {
					conf.Services[serviceName] = balancerServerTCP
					break
				}

				serviceKey := fmt.Sprintf("%s-%s-%s", serviceName, service.Name, &service.Port)
				conf.Services[serviceKey] = balancerServerTCP

				srv := dynamic.TCPWRRService{Name: serviceKey}
				srv.SetDefaults()
				if service.Weight != nil {
					srv.Weight = service.Weight
				}

				if conf.Services[serviceName] == nil {
					conf.Services[serviceName] = &dynamic.TCPService{Weighted: &dynamic.TCPWeightedRoundRobin{}}
				}
				conf.Services[serviceName].Weighted.Services = append(conf.Services[serviceName].Weighted.Services, srv)
			}

			r := &dynamic.TCPRouter{
				EntryPoints: ingressRouteTCP.Spec.EntryPoints,
				Middlewares: mds,
				Rule:        route.Match,
				Priority:    route.Priority,
				RuleSyntax:  route.Syntax,
				Service:     serviceName,
			}

			if ingressRouteTCP.Spec.TLS != nil {
				r.TLS = &dynamic.RouterTCPTLSConfig{
					Passthrough:  ingressRouteTCP.Spec.TLS.Passthrough,
					CertResolver: ingressRouteTCP.Spec.TLS.CertResolver,
					Domains:      ingressRouteTCP.Spec.TLS.Domains,
				}

				if ingressRouteTCP.Spec.TLS.Options != nil && len(ingressRouteTCP.Spec.TLS.Options.Name) > 0 {
					tlsOptionsName := ingressRouteTCP.Spec.TLS.Options.Name
					// Is a Kubernetes CRD reference (i.e. not a cross-provider reference)
					ns := ingressRouteTCP.Spec.TLS.Options.Namespace
					if !strings.Contains(tlsOptionsName, providerNamespaceSeparator) {
						if len(ns) == 0 {
							ns = ingressRouteTCP.Namespace
						}
						tlsOptionsName = makeID(ns, tlsOptionsName)
					} else if len(ns) > 0 {
						logger.Warn().
							Str("TLSOption", ingressRouteTCP.Spec.TLS.Options.Name).
							Msgf("Namespace %q is ignored in cross-provider context", ns)
					}

					if !isNamespaceAllowed(p.AllowCrossNamespace, ingressRouteTCP.Namespace, ns) {
						logger.Error().Msgf("TLSOption %s/%s is not in the IngressRouteTCP namespace %s",
							ns, ingressRouteTCP.Spec.TLS.Options.Name, ingressRouteTCP.Namespace)
						continue
					}

					r.TLS.Options = tlsOptionsName
				}
			}

			conf.Routers[serviceName] = r
		}
	}

	return conf
}

func (p *Provider) makeMiddlewareTCPKeys(ctx context.Context, ingRouteTCPNamespace string, middlewares []traefikv1alpha1.ObjectReference) ([]string, error) {
	var mds []string

	for _, mi := range middlewares {
		if strings.Contains(mi.Name, providerNamespaceSeparator) {
			if len(mi.Namespace) > 0 {
				log.Ctx(ctx).Warn().
					Str(logs.MiddlewareName, mi.Name).
					Msgf("Namespace %q is ignored in cross-provider context", mi.Namespace)
			}
			mds = append(mds, mi.Name)
			continue
		}

		ns := ingRouteTCPNamespace
		if len(mi.Namespace) > 0 {
			if !isNamespaceAllowed(p.AllowCrossNamespace, ingRouteTCPNamespace, mi.Namespace) {
				return nil, fmt.Errorf("middleware %s/%s is not in the IngressRouteTCP namespace %s", mi.Namespace, mi.Name, ingRouteTCPNamespace)
			}

			ns = mi.Namespace
		}

		mds = append(mds, provider.Normalize(makeID(ns, mi.Name)))
	}

	return mds, nil
}

func (p *Provider) createLoadBalancerServerTCP(client Client, parentNamespace string, service traefikv1alpha1.ServiceTCP) (*dynamic.TCPService, error) {
	ns := parentNamespace
	if len(service.Namespace) > 0 {
		if !isNamespaceAllowed(p.AllowCrossNamespace, parentNamespace, service.Namespace) {
			return nil, fmt.Errorf("tcp service %s/%s is not in the parent resource namespace %s", service.Namespace, service.Name, parentNamespace)
		}

		ns = service.Namespace
	}

	servers, err := p.loadTCPServers(client, ns, service)
	if err != nil {
		return nil, err
	}

	tcpService := &dynamic.TCPService{
		LoadBalancer: &dynamic.TCPServersLoadBalancer{
			Servers: servers,
		},
	}

	if service.ProxyProtocol != nil {
		tcpService.LoadBalancer.ProxyProtocol = &dynamic.ProxyProtocol{}
		tcpService.LoadBalancer.ProxyProtocol.SetDefaults()

		if service.ProxyProtocol.Version != 0 {
			tcpService.LoadBalancer.ProxyProtocol.Version = service.ProxyProtocol.Version
		}
	}

	if service.ServersTransport == "" && service.TerminationDelay != nil {
		tcpService.LoadBalancer.TerminationDelay = service.TerminationDelay
	}

	if service.ServersTransport != "" {
		tcpService.LoadBalancer.ServersTransport, err = p.makeTCPServersTransportKey(parentNamespace, service.ServersTransport)
		if err != nil {
			return nil, err
		}
	}

	return tcpService, nil
}

func (p *Provider) loadTCPServers(client Client, namespace string, svc traefikv1alpha1.ServiceTCP) ([]dynamic.TCPServer, error) {
	service, exists, err := client.GetService(namespace, svc.Name)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errors.New("service not found")
	}

	if service.Spec.Type == corev1.ServiceTypeExternalName && !p.AllowExternalNameServices {
		return nil, fmt.Errorf("externalName services not allowed: %s/%s", namespace, svc.Name)
	}

	svcPort, err := getServicePort(service, svc.Port)
	if err != nil {
		return nil, err
	}

	var servers []dynamic.TCPServer
	if service.Spec.Type == corev1.ServiceTypeNodePort && svc.NodePortLB {
		if p.DisableClusterScopeResources {
			return nil, errors.New("nodes lookup is disabled")
		}

		nodes, nodesExists, nodesErr := client.GetNodes()
		if nodesErr != nil {
			return nil, nodesErr
		}

		if !nodesExists || len(nodes) == 0 {
			return nil, fmt.Errorf("nodes not found for NodePort service %s/%s", svc.Namespace, svc.Name)
		}

		for _, node := range nodes {
			for _, addr := range node.Status.Addresses {
				if addr.Type == corev1.NodeInternalIP {
					servers = append(servers, dynamic.TCPServer{
						Address: net.JoinHostPort(addr.Address, strconv.Itoa(int(svcPort.NodePort))),
						TLS:     svc.TLS,
					})
				}
			}
		}

		if len(servers) == 0 {
			return nil, fmt.Errorf("no servers were generated for service %s/%s", svc.Namespace, svc.Name)
		}

		return servers, nil
	}

	if service.Spec.Type == corev1.ServiceTypeExternalName {
		servers = append(servers, dynamic.TCPServer{
			Address: net.JoinHostPort(service.Spec.ExternalName, strconv.Itoa(int(svcPort.Port))),
			TLS:     svc.TLS,
		})
	} else {
		nativeLB := p.NativeLBByDefault
		if svc.NativeLB != nil {
			nativeLB = *svc.NativeLB
		}
		if nativeLB {
			address, err := getNativeServiceAddress(*service, *svcPort)
			if err != nil {
				return nil, fmt.Errorf("getting native Kubernetes Service address: %w", err)
			}

			return []dynamic.TCPServer{{Address: address, TLS: svc.TLS}}, nil
		}

		endpointSlices, err := client.GetEndpointSlicesForService(namespace, svc.Name)
		if err != nil {
			return nil, fmt.Errorf("getting endpointslices: %w", err)
		}

		addresses := map[string]struct{}{}
		for _, endpointSlice := range endpointSlices {
			var port int32
			for _, p := range endpointSlice.Ports {
				if svcPort.Name == *p.Name {
					port = *p.Port
					break
				}
			}
			if port == 0 {
				continue
			}

			for _, endpoint := range endpointSlice.Endpoints {
				if endpoint.Conditions.Ready == nil || !*endpoint.Conditions.Ready {
					continue
				}

				for _, address := range endpoint.Addresses {
					if _, ok := addresses[address]; ok {
						continue
					}

					addresses[address] = struct{}{}
					servers = append(servers, dynamic.TCPServer{
						Address: net.JoinHostPort(address, strconv.Itoa(int(port))),
						TLS:     svc.TLS,
					})
				}
			}
		}
	}

	if len(servers) == 0 && !p.AllowEmptyServices {
		return nil, fmt.Errorf("no servers found for %s/%s", namespace, svc.Name)
	}

	return servers, nil
}

func (p *Provider) makeTCPServersTransportKey(parentNamespace string, serversTransportName string) (string, error) {
	if serversTransportName == "" {
		return "", nil
	}

	if !p.AllowCrossNamespace && strings.HasSuffix(serversTransportName, providerNamespaceSeparator+providerName) {
		// Since we are not able to know if another namespace is in the name (namespace-name@kubernetescrd),
		// if the provider namespace kubernetescrd is used,
		// we don't allow this format to avoid cross namespace references.
		return "", fmt.Errorf("invalid reference to serversTransport %s: namespace-name@kubernetescrd format is not allowed when crossnamespace is disallowed", serversTransportName)
	}

	if strings.Contains(serversTransportName, providerNamespaceSeparator) {
		return serversTransportName, nil
	}

	return provider.Normalize(makeID(parentNamespace, serversTransportName)), nil
}

// getTLSTCP mutates tlsConfigs.
func getTLSTCP(ctx context.Context, ingressRoute *traefikv1alpha1.IngressRouteTCP, k8sClient Client, tlsConfigs map[string]*tls.CertAndStores) error {
	if ingressRoute.Spec.TLS == nil {
		return nil
	}
	if ingressRoute.Spec.TLS.SecretName == "" {
		log.Ctx(ctx).Debug().Msg("No secret name provided")
		return nil
	}

	configKey := ingressRoute.Namespace + "/" + ingressRoute.Spec.TLS.SecretName
	if _, tlsExists := tlsConfigs[configKey]; !tlsExists {
		tlsConf, err := getTLS(k8sClient, ingressRoute.Spec.TLS.SecretName, ingressRoute.Namespace)
		if err != nil {
			return err
		}

		tlsConfigs[configKey] = tlsConf
	}

	return nil
}
