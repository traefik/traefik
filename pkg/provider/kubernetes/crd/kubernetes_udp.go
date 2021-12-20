package crd

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/provider"
	"github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func (p *Provider) loadIngressRouteUDPConfiguration(ctx context.Context, client Client) *dynamic.UDPConfiguration {
	conf := &dynamic.UDPConfiguration{
		Routers:     map[string]*dynamic.UDPRouter{},
		Middlewares: map[string]*dynamic.UDPMiddleware{},
		Services:    map[string]*dynamic.UDPService{},
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

			mds, err := p.makeMiddlewareUDPKeys(ctx, ingressRouteUDP.Namespace, route.Middlewares)
			if err != nil {
				logger.Errorf("Failed to create middleware keys: %v", err)
				continue
			}

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
				Middlewares: mds,
				Service:     serviceName,
			}
		}
	}

	return conf
}

func (p *Provider) makeMiddlewareUDPKeys(ctx context.Context, ingRouteUDPNamespace string, middlewares []v1alpha1.ObjectReference) ([]string, error) {
	var mds []string

	for _, mi := range middlewares {
		if strings.Contains(mi.Name, providerNamespaceSeparator) {
			if len(mi.Namespace) > 0 {
				log.FromContext(ctx).
					WithField(log.MiddlewareName, mi.Name).
					Warnf("namespace %q is ignored in cross-provider context", mi.Namespace)
			}
			mds = append(mds, mi.Name)
			continue
		}

		ns := ingRouteUDPNamespace
		if len(mi.Namespace) > 0 {
			if !isNamespaceAllowed(p.AllowCrossNamespace, ingRouteUDPNamespace, mi.Namespace) {
				return nil, fmt.Errorf("middleware %s/%s is not in the IngressRouteUDP namespace %s", mi.Namespace, mi.Name, ingRouteUDPNamespace)
			}

			ns = mi.Namespace
		}

		mds = append(mds, provider.Normalize(makeID(ns, mi.Name)))
	}

	return mds, nil
}

func (p *Provider) createLoadBalancerServerUDP(client Client, parentNamespace string, service v1alpha1.ServiceUDP) (*dynamic.UDPService, error) {
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

func (p *Provider) loadUDPServers(client Client, namespace string, svc v1alpha1.ServiceUDP) ([]dynamic.UDPServer, error) {
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

		if len(endpoints.Subsets) == 0 {
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
