package gateway

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

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

func (p *Provider) loadUDPRoutes(ctx context.Context, gatewayListeners []gatewayListener, conf *dynamic.Configuration) {
	logger := log.Ctx(ctx)
	routes, err := p.client.ListUDPRoutes()
	if err != nil {
		logger.Error().Err(err).Msgf("Unable to list UDPRoute")
		return
	}

	for _, route := range routes {
		logger := log.Ctx(ctx).With().
			Str("udp_route", route.Name).
			Str("namespace", route.Namespace).
			Logger()

		routeListeners := matchingGatewayListeners(gatewayListeners, route.Namespace, route.Spec.ParentRefs)
		if len(routeListeners) == 0 {
			continue
		}

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

			for _, listener := range routeListeners {
				accepted := matchListener(listener, parentRef)

				if accepted && !allowRoute(listener, route.Namespace, kindUDPRoute) {
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

				routeConf, resolveRefCondition := p.loadUDPRoute(listener, route)
				if accepted && listener.Attached {
					mergeUDPConfiguration(routeConf, conf)
				}
				parentStatus.Conditions = upsertRouteConditionResolvedRefs(parentStatus.Conditions, resolveRefCondition)
			}

			parentStatuses = append(parentStatuses, *parentStatus)
		}

		routeStatus := gatev1alpha2.UDPRouteStatus{
			RouteStatus: gatev1alpha2.RouteStatus{
				Parents: parentStatuses,
			},
		}
		if err := p.client.UpdateUDPRouteStatus(ctx, ktypes.NamespacedName{Namespace: route.Namespace, Name: route.Name}, routeStatus); err != nil {
			logger.Warn().
				Err(err).
				Msg("Unable to update UDPRoute status")
		}
	}
}

func (p *Provider) loadUDPRoute(listener gatewayListener, route *gatev1alpha2.UDPRoute) (*dynamic.Configuration, metav1.Condition) {
	conf := &dynamic.Configuration{
		UDP: &dynamic.UDPConfiguration{
			Routers:  make(map[string]*dynamic.UDPRouter),
			Services: make(map[string]*dynamic.UDPService),
		},
	}

	condition := metav1.Condition{
		Type:               string(gatev1.RouteConditionResolvedRefs),
		Status:             metav1.ConditionTrue,
		ObservedGeneration: route.Generation,
		LastTransitionTime: metav1.Now(),
		Reason:             string(gatev1.RouteConditionResolvedRefs),
	}

	for ri, rule := range route.Spec.Rules {
		if rule.BackendRefs == nil {
			// Should not happen due to validation.
			// https://github.com/kubernetes-sigs/gateway-api/blob/v0.4.0/apis/v1alpha2/udproute_types.go#L76
			continue
		}

		router := dynamic.UDPRouter{
			EntryPoints: []string{listener.EPName},
		}

		// Adding the gateway desc and the entryPoint desc prevents overlapping of routers build from the same routes.
		routeKey := provider.Normalize(fmt.Sprintf("%s-%s-%s-gw-%s-%s-ep-%s-%d", strings.ToLower(kindUDPRoute), route.Namespace, route.Name, listener.GWNamespace, listener.GWName, listener.EPName, ri))
		// Routing criteria should be introduced at some point.
		routerName := makeRouterName("", routeKey)

		if len(rule.BackendRefs) == 1 && isInternalService(rule.BackendRefs[0]) {
			router.Service = string(rule.BackendRefs[0].Name)
			conf.UDP.Routers[routerName] = &router
			continue
		}

		var serviceCondition *metav1.Condition
		router.Service, serviceCondition = p.loadUDPWRRService(conf, routerName, rule.BackendRefs, route)
		if serviceCondition != nil {
			condition = *serviceCondition
		}

		conf.UDP.Routers[routerName] = &router
	}

	return conf, condition
}

// loadUDPWRRService is generating a WRR service, even when there is only one target.
func (p *Provider) loadUDPWRRService(conf *dynamic.Configuration, routeKey string, backendRefs []gatev1.BackendRef, route *gatev1alpha2.UDPRoute) (string, *metav1.Condition) {
	name := routeKey + "-wrr"
	if _, ok := conf.UDP.Services[name]; ok {
		return name, nil
	}

	var wrr dynamic.UDPWeightedRoundRobin
	var condition *metav1.Condition
	for _, backendRef := range backendRefs {
		svcName, svc, errCondition := p.loadUDPService(route, backendRef)
		weight := ptr.To(int(ptr.Deref(backendRef.Weight, 1)))

		if errCondition != nil {
			condition = errCondition

			errName := routeKey + "-err-lb"
			conf.UDP.Services[errName] = &dynamic.UDPService{
				LoadBalancer: &dynamic.UDPServersLoadBalancer{
					Servers: []dynamic.UDPServer{},
				},
			}

			wrr.Services = append(wrr.Services, dynamic.UDPWRRService{
				Name:   errName,
				Weight: weight,
			})
			continue
		}

		if svc != nil {
			conf.UDP.Services[svcName] = svc
		}

		wrr.Services = append(wrr.Services, dynamic.UDPWRRService{
			Name:   svcName,
			Weight: weight,
		})
	}

	conf.UDP.Services[name] = &dynamic.UDPService{Weighted: &wrr}
	return name, condition
}

func (p *Provider) loadUDPService(route *gatev1alpha2.UDPRoute, backendRef gatev1.BackendRef) (string, *dynamic.UDPService, *metav1.Condition) {
	kind := ptr.Deref(backendRef.Kind, kindService)

	group := groupCore
	if backendRef.Group != nil && *backendRef.Group != "" {
		group = string(*backendRef.Group)
	}

	namespace := route.Namespace
	if backendRef.Namespace != nil && *backendRef.Namespace != "" {
		namespace = string(*backendRef.Namespace)
	}

	serviceName := provider.Normalize(namespace + "-" + string(backendRef.Name))

	if err := p.isReferenceGranted(kindUDPRoute, route.Namespace, group, string(kind), string(backendRef.Name), namespace); err != nil {
		return serviceName, nil, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonRefNotPermitted),
			Message:            fmt.Sprintf("Cannot load UDPRoute BackendRef %s/%s/%s/%s: %s", group, kind, namespace, backendRef.Name, err),
		}
	}

	if group != groupCore || kind != kindService {
		name, err := p.loadUDPBackendRef(backendRef)
		if err != nil {
			return serviceName, nil, &metav1.Condition{
				Type:               string(gatev1.RouteConditionResolvedRefs),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: route.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.RouteReasonInvalidKind),
				Message:            fmt.Sprintf("Cannot load UDPRoute BackendRef %s/%s/%s/%s: %s", group, kind, namespace, backendRef.Name, err),
			}
		}

		return name, nil, nil
	}

	port := ptr.Deref(backendRef.Port, gatev1.PortNumber(0))
	if port == 0 {
		return serviceName, nil, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonUnsupportedProtocol),
			Message:            fmt.Sprintf("Cannot load UDPRoute BackendRef %s/%s/%s/%s port is required", group, kind, namespace, backendRef.Name),
		}
	}

	portStr := strconv.FormatInt(int64(port), 10)
	serviceName = provider.Normalize(serviceName + "-" + portStr)

	lb, errCondition := p.loadUDPServers(namespace, route, backendRef)
	if errCondition != nil {
		return serviceName, nil, errCondition
	}

	return serviceName, &dynamic.UDPService{LoadBalancer: lb}, nil
}

func (p *Provider) loadUDPServers(namespace string, route *gatev1alpha2.UDPRoute, backendRef gatev1.BackendRef) (*dynamic.UDPServersLoadBalancer, *metav1.Condition) {
	backendAddresses, svcPort, err := p.getBackendAddresses(namespace, backendRef)
	if err != nil {
		return nil, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.GetGeneration(),
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonBackendNotFound),
			Message:            fmt.Sprintf("Cannot load UDPRoute BackendRef %s/%s: %s", namespace, backendRef.Name, err),
		}
	}

	if svcPort.Protocol != corev1.ProtocolUDP {
		return nil, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.GetGeneration(),
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonUnsupportedProtocol),
			Message:            fmt.Sprintf("Cannot load UDPRoute BackendRef %s/%s: only UDP protocol is supported", namespace, backendRef.Name),
		}
	}

	lb := &dynamic.UDPServersLoadBalancer{}

	for _, ba := range backendAddresses {
		lb.Servers = append(lb.Servers, dynamic.UDPServer{
			Address: net.JoinHostPort(ba.IP, strconv.Itoa(int(ba.Port))),
		})
	}
	return lb, nil
}

func (p *Provider) loadUDPBackendRef(backendRef gatev1.BackendRef) (string, error) {
	// Support for cross-provider references (e.g: api@internal).
	// This provides the same behavior as for IngressRoutes.
	if *backendRef.Kind == "TraefikService" && strings.Contains(string(backendRef.Name), "@") {
		return string(backendRef.Name), nil
	}

	return "", fmt.Errorf("unsupported BackendRef %s/%s/%s", *backendRef.Group, *backendRef.Kind, backendRef.Name)
}

func mergeUDPConfiguration(from, to *dynamic.Configuration) {
	if from == nil || from.UDP == nil || to == nil {
		return
	}

	if to.UDP == nil {
		to.UDP = from.UDP
		return
	}

	if to.UDP.Routers == nil {
		to.UDP.Routers = map[string]*dynamic.UDPRouter{}
	}
	for routerName, router := range from.UDP.Routers {
		to.UDP.Routers[routerName] = router
	}

	if to.UDP.Services == nil {
		to.UDP.Services = map[string]*dynamic.UDPService{}
	}
	for serviceName, service := range from.UDP.Services {
		to.UDP.Services[serviceName] = service
	}
}
