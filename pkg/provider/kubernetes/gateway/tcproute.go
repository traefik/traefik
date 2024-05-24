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

func (p *Provider) loadTCPRoutes(ctx context.Context, client Client, gatewayListeners []gatewayListener, conf *dynamic.Configuration) {
	logger := log.Ctx(ctx)
	routes, err := client.ListTCPRoutes()
	if err != nil {
		logger.Error().Err(err).Msgf("Get TCPRoutes: %s", err)
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
						Status:             metav1.ConditionTrue,
						ObservedGeneration: route.Generation,
						LastTransitionTime: metav1.Now(),
						Reason:             string(gatev1.RouteReasonAccepted),
					},
				},
			}

			var attachedListeners bool
			for _, listener := range gatewayListeners {
				if !matchListener(listener, route.Namespace, parentRef) {
					continue
				}

				if !allowRoute(listener, route.Namespace, kindTCPRoute) {
					continue
				}

				listener.Status.AttachedRoutes++
				attachedListeners = true

				resolveConditions := p.loadTCPRoute(client, listener, route, conf)

				// TODO: handle more accurately route conditions (in case of multiple listener matching).
				for _, condition := range resolveConditions {
					parentStatus.Conditions = appendCondition(parentStatus.Conditions, condition)
				}
			}

			if !attachedListeners {
				parentStatus.Conditions = []metav1.Condition{
					{
						Type:               string(gatev1.RouteConditionAccepted),
						Status:             metav1.ConditionFalse,
						ObservedGeneration: route.Generation,
						LastTransitionTime: metav1.Now(),
						Reason:             string(gatev1.RouteReasonNoMatchingParent),
					},
				}
			}

			parentStatuses = append(parentStatuses, *parentStatus)
		}

		routeStatus := gatev1alpha2.TCPRouteStatus{
			RouteStatus: gatev1alpha2.RouteStatus{
				Parents: parentStatuses,
			},
		}
		if err := client.UpdateTCPRouteStatus(ctx, ktypes.NamespacedName{Namespace: route.Namespace, Name: route.Name}, routeStatus); err != nil {
			logger.Error().
				Err(err).
				Msg("Unable to update TCPRoute status")
		}
	}
}

func (p *Provider) loadTCPRoute(client Client, listener gatewayListener, route *gatev1alpha2.TCPRoute, conf *dynamic.Configuration) []metav1.Condition {
	routeConditions := []metav1.Condition{
		{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionTrue,
			ObservedGeneration: route.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteConditionResolvedRefs),
		},
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
	routerName := route.Name + "-" + listener.GWName + "-" + listener.EPName
	routerKey := provider.Normalize(makeRouterKey("", makeID(route.Namespace, routerName)))

	var ruleServiceNames []string
	for i, rule := range route.Spec.Rules {
		if rule.BackendRefs == nil {
			// Should not happen due to validation.
			// https://github.com/kubernetes-sigs/gateway-api/blob/v0.4.0/apis/v1alpha2/tcproute_types.go#L76
			continue
		}

		wrrService, subServices, err := loadTCPServices(client, route.Namespace, rule.BackendRefs)
		if err != nil {
			routeConditions = appendCondition(routeConditions, metav1.Condition{
				Type:               string(gatev1.RouteConditionResolvedRefs),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: route.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.RouteReasonBackendNotFound),
				Message:            fmt.Sprintf("Cannot load TCPRoute service %s/%s: %v", route.Namespace, route.Name, err),
			})
			return routeConditions
		}

		for svcName, svc := range subServices {
			conf.TCP.Services[svcName] = svc
		}

		serviceName := fmt.Sprintf("%s-wrr-%d", routerKey, i)
		conf.TCP.Services[serviceName] = wrrService

		ruleServiceNames = append(ruleServiceNames, serviceName)
	}

	if len(ruleServiceNames) == 1 {
		router.Service = ruleServiceNames[0]
		conf.TCP.Routers[routerKey] = &router
		return routeConditions
	}

	routeServiceKey := routerKey + "-wrr"
	routeService := &dynamic.TCPService{Weighted: &dynamic.TCPWeightedRoundRobin{}}

	for _, name := range ruleServiceNames {
		service := dynamic.TCPWRRService{Name: name}
		service.SetDefaults()

		routeService.Weighted.Services = append(routeService.Weighted.Services, service)
	}

	conf.TCP.Services[routeServiceKey] = routeService

	router.Service = routeServiceKey
	conf.TCP.Routers[routerKey] = &router

	return routeConditions
}

// loadTCPServices is generating a WRR service, even when there is only one target.
func loadTCPServices(client Client, namespace string, backendRefs []gatev1.BackendRef) (*dynamic.TCPService, map[string]*dynamic.TCPService, error) {
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

		svc := dynamic.TCPService{
			LoadBalancer: &dynamic.TCPServersLoadBalancer{},
		}

		service, exists, err := client.GetService(namespace, string(backendRef.Name))
		if err != nil {
			return nil, nil, err
		}

		if !exists {
			return nil, nil, errors.New("service not found")
		}

		if len(service.Spec.Ports) > 1 && backendRef.Port == nil {
			// If the port is unspecified and the backend is a Service
			// object consisting of multiple port definitions, the route
			// must be dropped from the Gateway. The controller should
			// raise the "ResolvedRefs" condition on the Gateway with the
			// "DroppedRoutes" reason. The gateway status for this route
			// should be updated with a condition that describes the error
			// more specifically.
			log.Error().Msg("A multiple ports Kubernetes Service cannot be used if unspecified backendRef.Port")
			continue
		}

		var portSpec corev1.ServicePort
		var match bool

		for _, p := range service.Spec.Ports {
			if backendRef.Port == nil || p.Port == int32(*backendRef.Port) {
				portSpec = p
				match = true
				break
			}
		}

		if !match {
			return nil, nil, errors.New("service port not found")
		}

		endpoints, endpointsExists, endpointsErr := client.GetEndpoints(namespace, string(backendRef.Name))
		if endpointsErr != nil {
			return nil, nil, endpointsErr
		}

		if !endpointsExists {
			return nil, nil, errors.New("endpoints not found")
		}

		if len(endpoints.Subsets) == 0 {
			return nil, nil, errors.New("subset not found")
		}

		var port int32
		var portStr string
		for _, subset := range endpoints.Subsets {
			for _, p := range subset.Ports {
				if portSpec.Name == p.Name {
					port = p.Port
					break
				}
			}

			if port == 0 {
				return nil, nil, errors.New("cannot define a port")
			}

			portStr = strconv.FormatInt(int64(port), 10)
			for _, addr := range subset.Addresses {
				svc.LoadBalancer.Servers = append(svc.LoadBalancer.Servers, dynamic.TCPServer{
					Address: net.JoinHostPort(addr.IP, portStr),
				})
			}
		}

		serviceName := provider.Normalize(makeID(service.Namespace, service.Name) + "-" + portStr)
		services[serviceName] = &svc

		wrrSvc.Weighted.Services = append(wrrSvc.Weighted.Services, dynamic.TCPWRRService{Name: serviceName, Weight: &weight})
	}

	if len(wrrSvc.Weighted.Services) == 0 {
		return nil, nil, errors.New("no service has been created")
	}

	return wrrSvc, services, nil
}
