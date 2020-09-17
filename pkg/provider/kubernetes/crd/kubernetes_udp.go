package crd

import (
	"context"
	"errors"
	"fmt"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
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
				balancerServerUDP, err := createLoadBalancerServerUDP(client, ingressRouteUDP.Namespace, service)
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

				serviceKey := fmt.Sprintf("%s-%s-%d", serviceName, service.Name, service.Port)
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

func createLoadBalancerServerUDP(client Client, namespace string, service v1alpha1.ServiceUDP) (*dynamic.UDPService, error) {
	ns := namespace
	if len(service.Namespace) > 0 {
		ns = service.Namespace
	}

	servers, err := loadUDPServers(client, ns, service)
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

func loadUDPServers(client Client, namespace string, svc v1alpha1.ServiceUDP) ([]dynamic.UDPServer, error) {
	service, exists, err := client.GetService(namespace, svc.Name)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errors.New("service not found")
	}

	var portSpec *corev1.ServicePort
	for _, p := range service.Spec.Ports {
		p := p
		if svc.Port == p.Port {
			portSpec = &p
			break
		}
	}

	if portSpec == nil {
		return nil, errors.New("service port not found")
	}

	var servers []dynamic.UDPServer
	if service.Spec.Type == corev1.ServiceTypeExternalName {
		servers = append(servers, dynamic.UDPServer{
			Address: fmt.Sprintf("%s:%d", service.Spec.ExternalName, portSpec.Port),
		})
	} else {
		endpoints, endpointsExists, endpointsErr := client.GetEndpoints(namespace, svc.Name)
		if endpointsErr != nil {
			return nil, endpointsErr
		}

		if !endpointsExists {
			return nil, errors.New("endpoints not found")
		}

		if len(endpoints.Subsets) == 0 {
			return nil, errors.New("subset not found")
		}

		var port int32
		for _, subset := range endpoints.Subsets {
			for _, p := range subset.Ports {
				if portSpec.Name == p.Name {
					port = p.Port
					break
				}
			}

			if port == 0 {
				return nil, errors.New("cannot define a port")
			}

			for _, addr := range subset.Addresses {
				servers = append(servers, dynamic.UDPServer{
					Address: fmt.Sprintf("%s:%d", addr.IP, port),
				})
			}
		}
	}

	return servers, nil
}
