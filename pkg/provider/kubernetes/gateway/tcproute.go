package gateway

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/provider"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	gatev1 "sigs.k8s.io/gateway-api/apis/v1"
	gatev1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func (p *Provider) loadTCPRoutes(ctx context.Context, gatewayListeners []gatewayListener, conf *dynamic.Configuration) {
	logger := log.Ctx(ctx)
	routes, err := p.client.ListTCPRoutes()
	if err != nil {
		logger.Error().Err(err).Msgf("Unable to list TCPRoutes")
		return
	}

	for _, route := range routes {
		logger := log.Ctx(ctx).With().Str("tcproute", route.Name).Str("namespace", route.Namespace).Logger()

		var parentStatuses []gatev1alpha2.RouteParentStatus
		for _, parentRef := range route.Spec.ParentRefs {
			parentStatus := &gatev1alpha2.RouteParentStatus{
				ParentRef:      parentRef,
				ControllerName: controllerName,
				Conditions: []metav1.Condition{
					{
						Type:               string(gatev1.RouteConditionAccepted),
						Status:             metav1.ConditionFalse,
						ObservedGeneration: route.Generation,
						LastTransitionTime: metav1.Now(),
						Reason:             string(gatev1.RouteReasonNoMatchingParent),
					},
				},
			}

			for _, listener := range gatewayListeners {
				if !matchListener(listener, route.Namespace, parentRef) {
					continue
				}

				accepted := true
				if !allowRoute(listener, route.Namespace, kindTCPRoute) {
					parentStatus.Conditions = updateRouteConditionAccepted(parentStatus.Conditions, string(gatev1.RouteReasonNotAllowedByListeners))
					accepted = false
				}

				if accepted {
					listener.Status.AttachedRoutes++
					// only consider the route attached if the listener is in an "attached" state.
					if listener.Attached {
						parentStatus.Conditions = updateRouteConditionAccepted(parentStatus.Conditions, string(gatev1.RouteReasonAccepted))
					}
				}

				routeConf, resolveRefCondition := p.loadTCPRoute(listener, route)
				if accepted && listener.Attached {
					mergeTCPConfiguration(routeConf, conf)
				}
				parentStatus.Conditions = upsertRouteConditionResolvedRefs(parentStatus.Conditions, resolveRefCondition)
			}

			parentStatuses = append(parentStatuses, *parentStatus)
		}

		routeStatus := gatev1alpha2.TCPRouteStatus{
			RouteStatus: gatev1alpha2.RouteStatus{
				Parents: parentStatuses,
			},
		}
		if err := p.client.UpdateTCPRouteStatus(ctx, ktypes.NamespacedName{Namespace: route.Namespace, Name: route.Name}, routeStatus); err != nil {
			logger.Error().
				Err(err).
				Msg("Unable to update TCPRoute status")
		}
	}
}

func (p *Provider) loadTCPRoute(listener gatewayListener, route *gatev1alpha2.TCPRoute) (*dynamic.Configuration, metav1.Condition) {
	conf := &dynamic.Configuration{
		TCP: &dynamic.TCPConfiguration{
			Routers:           make(map[string]*dynamic.TCPRouter),
			Middlewares:       make(map[string]*dynamic.TCPMiddleware),
			Services:          make(map[string]*dynamic.TCPService),
			ServersTransports: make(map[string]*dynamic.TCPServersTransport),
		},
	}

	condition := metav1.Condition{
		Type:               string(gatev1.RouteConditionResolvedRefs),
		Status:             metav1.ConditionTrue,
		ObservedGeneration: route.Generation,
		LastTransitionTime: metav1.Now(),
		Reason:             string(gatev1.RouteConditionResolvedRefs),
	}

	router := dynamic.TCPRouter{
		Rule:        "HostSNI(`*`)",
		EntryPoints: []string{listener.EPName},
		RuleSyntax:  "v3",
	}

	if listener.Protocol == gatev1.TLSProtocolType && listener.TLS != nil {
		// TODO support let's encrypt
		router.TLS = &dynamic.RouterTCPTLSConfig{
			Passthrough: listener.TLS.Mode != nil && *listener.TLS.Mode == gatev1.TLSModePassthrough,
		}
	}

	// Adding the gateway desc and the entryPoint desc prevents overlapping of routers build from the same routes.
	routerName := provider.Normalize(route.Namespace + "-" + route.Name + "-" + listener.GWName + "-" + listener.EPName)

	var ruleServiceNames []string
	for i, rule := range route.Spec.Rules {
		if rule.BackendRefs == nil {
			// Should not happen due to validation.
			// https://github.com/kubernetes-sigs/gateway-api/blob/v0.4.0/apis/v1alpha2/tcproute_types.go#L76
			continue
		}

		wrrService, subServices, err := p.loadTCPServices(route.Namespace, rule.BackendRefs)
		if err != nil {
			return conf, metav1.Condition{
				Type:               string(gatev1.RouteConditionResolvedRefs),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: route.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.RouteReasonBackendNotFound),
				Message:            fmt.Sprintf("Cannot load TCPRoute service %s/%s: %v", route.Namespace, route.Name, err),
			}
		}

		for svcName, svc := range subServices {
			conf.TCP.Services[svcName] = svc
		}

		serviceName := fmt.Sprintf("%s-wrr-%d", routerName, i)
		conf.TCP.Services[serviceName] = wrrService

		ruleServiceNames = append(ruleServiceNames, serviceName)
	}

	if len(ruleServiceNames) == 1 {
		router.Service = ruleServiceNames[0]
		conf.TCP.Routers[routerName] = &router
		return conf, condition
	}

	routeServiceKey := routerName + "-wrr"
	routeService := &dynamic.TCPService{Weighted: &dynamic.TCPWeightedRoundRobin{}}

	for _, name := range ruleServiceNames {
		service := dynamic.TCPWRRService{Name: name}
		service.SetDefaults()

		routeService.Weighted.Services = append(routeService.Weighted.Services, service)
	}

	conf.TCP.Services[routeServiceKey] = routeService

	router.Service = routeServiceKey
	conf.TCP.Routers[routerName] = &router

	return conf, condition
}

// loadTCPServices is generating a WRR service, even when there is only one target.
func (p *Provider) loadTCPServices(namespace string, backendRefs []gatev1.BackendRef) (*dynamic.TCPService, map[string]*dynamic.TCPService, error) {
	services := map[string]*dynamic.TCPService{}

	wrrSvc := &dynamic.TCPService{
		Weighted: &dynamic.TCPWeightedRoundRobin{
			Services: []dynamic.TCPWRRService{},
		},
	}

	for _, backendRef := range backendRefs {
		if backendRef.Group == nil || backendRef.Kind == nil {
			// Should not happen as this is validated by kubernetes
			continue
		}

		if isInternalService(backendRef) {
			return nil, nil, fmt.Errorf("traefik internal service %s is not allowed in a WRR loadbalancer", backendRef.Name)
		}

		weight := int(ptr.Deref(backendRef.Weight, 1))

		if isTraefikService(backendRef) {
			wrrSvc.Weighted.Services = append(wrrSvc.Weighted.Services, dynamic.TCPWRRService{Name: string(backendRef.Name), Weight: &weight})
			continue
		}

		if *backendRef.Group != "" && *backendRef.Group != groupCore && *backendRef.Kind != "Service" {
			return nil, nil, fmt.Errorf("unsupported BackendRef %s/%s/%s", *backendRef.Group, *backendRef.Kind, backendRef.Name)
		}

		if backendRef.Port == nil {
			return nil, nil, errors.New("port is required for Kubernetes Service reference")
		}

		service, exists, err := p.client.GetService(namespace, string(backendRef.Name))
		if err != nil {
			return nil, nil, fmt.Errorf("getting service: %w", err)
		}
		if !exists {
			return nil, nil, errors.New("service not found")
		}

		var svcPort *corev1.ServicePort
		for _, p := range service.Spec.Ports {
			if p.Port == int32(*backendRef.Port) {
				svcPort = &p
				break
			}
		}
		if svcPort == nil {
			return nil, nil, fmt.Errorf("service port %d not found", *backendRef.Port)
		}

		endpointSlices, err := p.client.ListEndpointSlicesForService(namespace, string(backendRef.Name))
		if err != nil {
			return nil, nil, fmt.Errorf("getting endpointslices: %w", err)
		}
		if len(endpointSlices) == 0 {
			return nil, nil, errors.New("endpointslices not found")
		}

		svc := dynamic.TCPService{LoadBalancer: &dynamic.TCPServersLoadBalancer{}}

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
					svc.LoadBalancer.Servers = append(svc.LoadBalancer.Servers, dynamic.TCPServer{
						Address: net.JoinHostPort(address, strconv.Itoa(int(port))),
					})
				}
			}
		}

		serviceName := provider.Normalize(service.Namespace + "-" + service.Name + "-" + strconv.Itoa(int(svcPort.Port)))
		services[serviceName] = &svc

		wrrSvc.Weighted.Services = append(wrrSvc.Weighted.Services, dynamic.TCPWRRService{Name: serviceName, Weight: &weight})
	}

	if len(wrrSvc.Weighted.Services) == 0 {
		return nil, nil, errors.New("no service has been created")
	}

	return wrrSvc, services, nil
}

func mergeTCPConfiguration(from, to *dynamic.Configuration) {
	if from == nil || from.TCP == nil || to == nil {
		return
	}

	if to.TCP == nil {
		to.TCP = from.TCP
		return
	}

	if to.TCP.Routers == nil {
		to.TCP.Routers = map[string]*dynamic.TCPRouter{}
	}
	for routerName, router := range from.TCP.Routers {
		to.TCP.Routers[routerName] = router
	}

	if to.TCP.Middlewares == nil {
		to.TCP.Middlewares = map[string]*dynamic.TCPMiddleware{}
	}
	for middlewareName, middleware := range from.TCP.Middlewares {
		to.TCP.Middlewares[middlewareName] = middleware
	}

	if to.TCP.Services == nil {
		to.TCP.Services = map[string]*dynamic.TCPService{}
	}
	for serviceName, service := range from.TCP.Services {
		to.TCP.Services[serviceName] = service
	}
}
