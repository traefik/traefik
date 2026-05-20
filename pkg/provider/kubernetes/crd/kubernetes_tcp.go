package crd

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	traefikv1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	"github.com/traefik/traefik/v2/pkg/tls"
	corev1 "k8s.io/api/core/v1"
)

func (p *Provider) loadIngressRouteTCPConfiguration(ctx context.Context, client Client, tlsConfigs map[string]*tls.CertAndStores) *dynamic.TCPConfiguration {
	conf := &dynamic.TCPConfiguration{
		Routers:     map[string]*dynamic.TCPRouter{},
		Middlewares: map[string]*dynamic.TCPMiddleware{},
		Services:    map[string]*dynamic.TCPService{},
	}

	for _, ingressRouteTCP := range client.GetIngressRouteTCPs() {
		logger := log.FromContext(log.With(ctx, log.Str("ingress", ingressRouteTCP.Name), log.Str("namespace", ingressRouteTCP.Namespace)))

		if !shouldProcessIngress(p.IngressClass, ingressRouteTCP.Annotations[annotationKubernetesIngressClass]) {
			continue
		}

		if ingressRouteTCP.Spec.TLS != nil && !ingressRouteTCP.Spec.TLS.Passthrough {
			err := getTLSTCP(ctx, ingressRouteTCP, client, tlsConfigs)
			if err != nil {
				logger.Errorf("Error configuring TLS: %v", err)
			}
		}

		ingressName := ingressRouteTCP.Name
		if len(ingressName) == 0 {
			ingressName = ingressRouteTCP.GenerateName
		}

		for _, route := range ingressRouteTCP.Spec.Routes {
			if len(route.Match) == 0 {
				logger.Errorf("Empty match rule")
				continue
			}

			key, err := makeServiceKey(route.Match, ingressName)
			if err != nil {
				logger.Error(err)
				continue
			}

			mds, err := p.makeMiddlewareTCPKeys(ctx, ingressRouteTCP.Namespace, route.Middlewares)
			if err != nil {
				logger.Errorf("Failed to create middleware keys: %v", err)
				continue
			}

			serviceName := makeID(ingressRouteTCP.Namespace, key)

			for _, service := range route.Services {
				balancerServerTCP, err := p.createLoadBalancerServerTCP(client, ingressRouteTCP.Namespace, service)
				if err != nil {
					logger.
						WithField("serviceName", service.Name).
						WithField("servicePort", service.Port).
						Errorf("Cannot create service: %v", err)
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
				Service:     serviceName,
			}

			if ingressRouteTCP.Spec.TLS != nil {
				r.TLS = &dynamic.RouterTCPTLSConfig{
					Passthrough:  ingressRouteTCP.Spec.TLS.Passthrough,
					CertResolver: ingressRouteTCP.Spec.TLS.CertResolver,
					Domains:      ingressRouteTCP.Spec.TLS.Domains,
				}

				if ingressRouteTCP.Spec.TLS.Options != nil && len(ingressRouteTCP.Spec.TLS.Options.Name) > 0 {
					tlsOptions := ingressRouteTCP.Spec.TLS.Options
					ctxTLSOption := log.With(ctx, log.Str("TLSOption", tlsOptions.Name))

					r.TLS.Options, err = resolveReference(ctxTLSOption, ingressRouteTCP.Namespace, tlsOptions.Namespace, tlsOptions.Name, p.CrossProviderNamespaces, p.AllowCrossNamespace)
					if err != nil {
						logger.WithError(err).Errorf("Invalid reference to TLSOption %q", ingressRouteTCP.Spec.TLS.Options.Name)
						continue
					}
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
		ctxMid := log.With(ctx, log.Str(log.MiddlewareName, mi.Name))

		middlewareRef, err := resolveReference(ctxMid, ingRouteTCPNamespace, mi.Namespace, mi.Name, p.CrossProviderNamespaces, p.AllowCrossNamespace)
		if err != nil {
			return nil, fmt.Errorf("invalid reference to middleware %s: %w", mi.Name, err)
		}

		mds = append(mds, middlewareRef)
	}

	return mds, nil
}

func (p *Provider) createLoadBalancerServerTCP(client Client, parentNamespace string, service traefikv1alpha1.ServiceTCP) (*dynamic.TCPService, error) {
	ns := namespaceOrParentNamespace(service.Namespace, parentNamespace)

	if !isNamespaceAllowed(p.AllowCrossNamespace, parentNamespace, ns) {
		return nil, fmt.Errorf("tcp service %s/%s is not in the parent resource namespace %s", ns, service.Name, parentNamespace)
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

	if service.TerminationDelay != nil {
		tcpService.LoadBalancer.TerminationDelay = service.TerminationDelay
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

	if svc.NativeLB {
		address, err := getNativeServiceAddress(*service, *svcPort)
		if err != nil {
			return nil, fmt.Errorf("getting native Kubernetes Service address: %w", err)
		}

		return []dynamic.TCPServer{{Address: address}}, nil
	}

	var servers []dynamic.TCPServer
	if service.Spec.Type == corev1.ServiceTypeExternalName {
		servers = append(servers, dynamic.TCPServer{
			Address: net.JoinHostPort(service.Spec.ExternalName, strconv.Itoa(int(svcPort.Port))),
		})
	} else {
		endpoints, endpointsExists, endpointsErr := client.GetEndpoints(namespace, svc.Name)
		if endpointsErr != nil {
			return nil, endpointsErr
		}

		if !endpointsExists {
			return nil, errors.New("endpoints not found")
		}

		if len(endpoints.Subsets) == 0 && !p.AllowEmptyServices {
			return nil, errors.New("subset not found")
		}

		var port int32
		for _, subset := range endpoints.Subsets {
			for _, p := range subset.Ports {
				if svcPort.Name == p.Name {
					port = p.Port
					break
				}
			}

			if port == 0 {
				return nil, errors.New("cannot define a port")
			}

			for _, addr := range subset.Addresses {
				servers = append(servers, dynamic.TCPServer{
					Address: net.JoinHostPort(addr.IP, strconv.Itoa(int(port))),
				})
			}
		}
	}

	return servers, nil
}

// getTLSTCP mutates tlsConfigs.
func getTLSTCP(ctx context.Context, ingressRoute *traefikv1alpha1.IngressRouteTCP, k8sClient Client, tlsConfigs map[string]*tls.CertAndStores) error {
	if ingressRoute.Spec.TLS == nil {
		return nil
	}
	if ingressRoute.Spec.TLS.SecretName == "" {
		log.FromContext(ctx).Debugf("No secret name provided")
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
