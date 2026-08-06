package gateway

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"maps"
	"net"
	"slices"
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
)

func (p *Provider) loadUDPRoutes(ctx context.Context, gateways []gatewayWithListeners, conf *dynamic.Configuration, statusReport *statusReport) {
	logger := log.Ctx(ctx)
	routes, err := p.client.ListUDPRoutes()
	if err != nil {
		logger.Error().Err(err).Msgf("Unable to list UDPRoutes")
		return
	}

	// UDPRoutes attaching to the same listener all match every datagram,
	// and the specification gives precedence to the oldest one.
	// https://gateway-api.sigs.k8s.io/guides/api-design/#conflicts
	slices.SortStableFunc(routes, func(a, b *gatev1.UDPRoute) int {
		if cmpTime := a.CreationTimestamp.Time.Compare(b.CreationTimestamp.Time); cmpTime != 0 {
			return cmpTime
		}

		return cmp.Or(
			cmp.Compare(a.Namespace, b.Namespace),
			cmp.Compare(a.Name, b.Name),
		)
	})

	for _, route := range routes {
		routeName := ktypes.NamespacedName{Namespace: route.Namespace, Name: route.Name}
		routeParentRefs := matchingGatewayListenersForParentRef(gateways, route.Namespace, route.Spec.ParentRefs)
		if len(routeParentRefs) == 0 {
			if len(route.Status.Parents) > 0 {
				statusReport.RecordUDPRouteStatusReset(routeName)
			}
			continue
		}

		for _, match := range routeParentRefs {
			acceptedCondition := metav1.Condition{
				Type:               string(gatev1.RouteConditionAccepted),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: route.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.RouteReasonNoMatchingParent),
			}

			var resolvedRefCondition *metav1.Condition
			for _, listener := range match.listeners {
				// A parentRef can target specific listeners through its SectionName or Port.
				accepted := matchListener(listener, match.parentRef)

				if accepted && !allowRoute(listener, route.Namespace, kindUDPRoute) {
					if acceptedCondition.Status == metav1.ConditionFalse {
						acceptedCondition.Reason = string(gatev1.RouteReasonNotAllowedByListeners)
					}
					accepted = false
				}

				if accepted {
					listener.Status.AttachedRoutes++
				}

				// The ResolvedRefs condition must be reported for every parentRef,
				// even when the route does not attach to the listener.
				routeConf, condition := p.loadUDPRoute(match.gatewayName, match.gatewayNamespace, listener, route)
				if resolvedRefCondition == nil || resolvedRefCondition.Status == metav1.ConditionTrue {
					resolvedRefCondition = new(condition)
				}

				if accepted {
					// A conflicting route is still reported as accepted and counted in
					// the listener AttachedRoutes, only its configuration is dropped.
					// A UDP listener supports no route kind other than UDPRoute, so the
					// first route to attach to it is the one to bind.
					if listener.Attached && listener.Status.AttachedRoutes == 1 {
						mergeUDPConfiguration(routeConf, conf)
					}

					acceptedCondition.Reason = string(gatev1.RouteReasonAccepted)
					acceptedCondition.Status = metav1.ConditionTrue
				}
			}

			parentStatusConditions := []metav1.Condition{acceptedCondition}
			if resolvedRefCondition != nil {
				parentStatusConditions = append(parentStatusConditions, *resolvedRefCondition)
			}

			statusReport.RecordUDPRouteStatus(routeName, gatev1.RouteParentStatus{
				ParentRef:      match.parentRef,
				ControllerName: controllerName,
				Conditions:     parentStatusConditions,
			})
		}
	}
}

func (p *Provider) loadUDPRoute(gatewayName, gatewayNamespace string, listener gatewayListener, route *gatev1.UDPRoute) (*dynamic.Configuration, metav1.Condition) {
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
			// Should not happen due to CRD validation.
			continue
		}

		router := dynamic.UDPRouter{
			EntryPoints: []string{listener.EPName},
		}

		routerName := makeRouterName(strings.ToLower(kindUDPRoute), "", route.Namespace, route.Name, gatewayNamespace, gatewayName, listener.EPName, ri)

		// Keep zero-weight references behind the WRR so the weight is enforced.
		if len(rule.BackendRefs) == 1 && isInternalService(rule.BackendRefs[0]) &&
			ptr.Deref(rule.BackendRefs[0].Weight, 1) > 0 &&
			(rule.BackendRefs[0].Namespace == nil || *rule.BackendRefs[0].Namespace == "") {
			if !isCrossProviderNamespaceAllowed(p.CrossProviderNamespaces, route.Namespace) {
				condition = metav1.Condition{
					Type:               string(gatev1.RouteConditionResolvedRefs),
					Status:             metav1.ConditionFalse,
					ObservedGeneration: route.Generation,
					LastTransitionTime: metav1.Now(),
					Reason:             string(gatev1.RouteReasonRefNotPermitted),
					Message:            fmt.Sprintf("Cannot load UDPRoute BackendRef %s: internal service reference is not allowed: UDPRoute namespace %q is not in crossProviderNamespaces", rule.BackendRefs[0].Name, route.Namespace),
				}

				continue
			}

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

// loadUDPWRRService creates a WRR service even for a single backend.
func (p *Provider) loadUDPWRRService(conf *dynamic.Configuration, routerName string, backendRefs []gatev1.BackendRef, route *gatev1.UDPRoute) (string, *metav1.Condition) {
	name := routerName + "-wrr"
	if _, ok := conf.UDP.Services[name]; ok {
		return name, nil
	}

	var wrr dynamic.UDPWeightedRoundRobin
	var condition *metav1.Condition
	for bi, backendRef := range backendRefs {
		svcName, svc, errCondition := p.loadUDPService(routerName, route, bi, backendRef)
		weight := new(int(ptr.Deref(backendRef.Weight, 1)))

		if errCondition != nil {
			condition = errCondition

			errName := routerName + "-err-lb"
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

func (p *Provider) loadUDPService(routerName string, route *gatev1.UDPRoute, backendIndex int, backendRef gatev1.BackendRef) (string, *dynamic.UDPService, *metav1.Condition) {
	kind := ptr.Deref(backendRef.Kind, kindService)

	group := string(ptr.Deref(backendRef.Group, gatev1.Group("")))
	referenceGrantGroup := group
	if referenceGrantGroup == "" {
		referenceGrantGroup = groupCore
	}

	namespace := route.Namespace
	if backendRef.Namespace != nil && *backendRef.Namespace != "" {
		namespace = string(*backendRef.Namespace)
	}
	serviceName := fmt.Sprintf("%s-svc-%s-%s-%d", routerName, namespace, backendRef.Name, backendIndex)

	if (group != "" || kind != kindService) && !isTraefikService(backendRef) {
		return serviceName, nil, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonInvalidKind),
			Message:            fmt.Sprintf("Cannot load UDPRoute BackendRef %s/%s/%s/%s: unsupported BackendRef", referenceGrantGroup, kind, namespace, backendRef.Name),
		}
	}

	if backendRef.Namespace != nil && *backendRef.Namespace != "" {
		if strings.Contains(string(backendRef.Name), "@") {
			return provider.Normalize(serviceName), nil, &metav1.Condition{
				Type:               string(gatev1.RouteConditionResolvedRefs),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: route.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.RouteReasonRefNotPermitted),
				Message:            fmt.Sprintf("Cannot load UDPRoute BackendRef %s/%s/%s/%s: namespace is not allowed with a cross-provider reference", referenceGrantGroup, kind, namespace, backendRef.Name),
			}
		}
	}

	if err := p.isReferenceGranted(kindUDPRoute, route.Namespace, referenceGrantGroup, string(kind), string(backendRef.Name), namespace); err != nil {
		return serviceName, nil, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonRefNotPermitted),
			Message:            fmt.Sprintf("Cannot load UDPRoute BackendRef %s/%s/%s/%s: %s", referenceGrantGroup, kind, namespace, backendRef.Name, err),
		}
	}

	if group != "" || kind != kindService {
		name, err := p.loadUDPBackendRef(route.Namespace, group, kind, backendRef)
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
			Message:            fmt.Sprintf("Cannot load UDPRoute BackendRef %s/%s/%s/%s port is required", referenceGrantGroup, kind, namespace, backendRef.Name),
		}
	}

	lb, errCondition := p.loadUDPServers(namespace, route, backendRef)
	if errCondition != nil {
		return serviceName, nil, errCondition
	}

	return serviceName, &dynamic.UDPService{LoadBalancer: lb}, nil
}

func (p *Provider) loadUDPServers(namespace string, route *gatev1.UDPRoute, backendRef gatev1.BackendRef) (*dynamic.UDPServersLoadBalancer, *metav1.Condition) {
	backendAddresses, svcPort, err := p.getBackendAddresses(namespace, backendRef)
	if err != nil && !errors.Is(err, errNoEndpointSlices) {
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
	if errors.Is(err, errNoEndpointSlices) {
		return lb, nil
	}

	for _, ba := range backendAddresses {
		lb.Servers = append(lb.Servers, dynamic.UDPServer{
			Address: net.JoinHostPort(ba.IP, strconv.Itoa(int(ba.Port))),
		})
	}
	return lb, nil
}

func (p *Provider) loadUDPBackendRef(routeNamespace, group string, kind gatev1.Kind, backendRef gatev1.BackendRef) (string, error) {
	// Support for cross-provider references (e.g: api@internal).
	// This provides the same behavior as for IngressRoutes.
	if isTraefikService(backendRef) && strings.Contains(string(backendRef.Name), "@") {
		if !isCrossProviderNamespaceAllowed(p.CrossProviderNamespaces, routeNamespace) {
			return "", fmt.Errorf("TraefikService %q reference is not allowed: route namespace %q is not in crossProviderNamespaces", string(backendRef.Name), routeNamespace)
		}

		return string(backendRef.Name), nil
	}

	return "", fmt.Errorf("unsupported BackendRef %s/%s/%s", group, kind, backendRef.Name)
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
	maps.Copy(to.UDP.Routers, from.UDP.Routers)

	if to.UDP.Services == nil {
		to.UDP.Services = map[string]*dynamic.UDPService{}
	}
	maps.Copy(to.UDP.Services, from.UDP.Services)
}
