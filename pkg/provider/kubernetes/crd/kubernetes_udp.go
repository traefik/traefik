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
	corev1 "k8s.io/api/core/v1"
)

func (p *Provider) loadIngressRouteUDPConfiguration(ctx context.Context, client Client) *dynamic.UDPConfiguration {
	conf := &dynamic.UDPConfiguration{
		Routers:  map[string]*dynamic.UDPRouter{},
		Services: map[string]*dynamic.UDPService{},
	}

	for _, ingressRouteUDP := range client.GetIngressRouteUDPs() {
		logger := log.FromContext(log.With(ctx, log.Str("ingress", ingressRouteUDP.Name), log.Str("namespace", ingressRouteUDP.Namespace)))

		if !shouldProcessIngress(p.IngressClass, ingressRouteUDP.Annotations[annotationKubernetesIngressClass]) {
			continue
		}

		ingressName := ingressRouteUDP.Name
		if len(ingressName) == 0 {
			ingressName = ingressRouteUDP.GenerateName
		}

		for i, route := range ingressRouteUDP.Spec.Routes {
			key := fmt.Sprintf("%s-%d", ingressName, i)
			serviceName := makeID(ingressRouteUDP.Namespace, key)

			for _, service := range route.Services {
				balancerServerUDP, err := p.createLoadBalancerServerUDP(client, ingressRouteUDP.Namespace, service)
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
					conf.Services[serviceName] = balancerServerUDP
					break
				}

				serviceKey := fmt.Sprintf("%s-%s-%s", serviceName, service.Name, &service.Port)
				conf.Services[serviceKey] = balancerServerUDP

				srv := dynamic.UDPWRRService{Name: serviceKey}
				srv.SetDefaults()
				if service.Weight != nil {
					srv.Weight = service.Weight
				}

				if conf.Services[serviceName] == nil {
					conf.Services[serviceName] = &dynamic.UDPService{Weighted: &dynamic.UDPWeightedRoundRobin{}}
				}
				conf.Services[serviceName].Weighted.Services = append(conf.Services[serviceName].Weighted.Services, srv)
			}

			conf.Routers[serviceName] = &dynamic.UDPRouter{
				EntryPoints: ingressRouteUDP.Spec.EntryPoints,
				Service:     serviceName,
			}
		}
	}

	return conf
}

func (p *Provider) createLoadBalancerServerUDP(client Client, parentNamespace string, service traefikv1alpha1.ServiceUDP) (*dynamic.UDPService, error) {
	ns := parentNamespace
	if len(service.Namespace) > 0 {
		if !isNamespaceAllowed(p.AllowCrossNamespace, parentNamespace, service.Namespace) {
			return nil, fmt.Errorf("udp service %s/%s is not in the parent resource namespace %s", service.Namespace, service.Name, ns)
		}

		ns = service.Namespace
	}

	servers, err := p.loadUDPServers(client, ns, service)
	if err != nil {
		return nil, err
	}

	udpService := &dynamic.UDPService{
		LoadBalancer: &dynamic.UDPServersLoadBalancer{
			Servers: servers,
		},
	}

	return udpService, nil
}

func (p *Provider) loadUDPServers(client Client, namespace string, svc traefikv1alpha1.ServiceUDP) ([]dynamic.UDPServer, error) {
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

		return []dynamic.UDPServer{{Address: address}}, nil
	}

	var servers []dynamic.UDPServer
	if service.Spec.Type == corev1.ServiceTypeExternalName {
		servers = append(servers, dynamic.UDPServer{
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
				servers = append(servers, dynamic.UDPServer{
					Address: net.JoinHostPort(addr.IP, strconv.Itoa(int(port))),
				})
			}
		}
	}

	return servers, nil
}
