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

func (p *Provider) loadTCPRoutes(ctx context.Context, gatewayListeners []gatewayListener, conf *dynamic.Configuration) {
	logger := log.Ctx(ctx)
	routes, err := p.client.ListTCPRoutes()
	if err != nil {
		logger.Error().Err(err).Msgf("Unable to list TCPRoutes")
		return
	}

	for _, route := range routes {
		logger := log.Ctx(ctx).With().
			Str("tcp_route", route.Name).
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

				if accepted && !allowRoute(listener, route.Namespace, kindTCPRoute) {
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
			logger.Warn().
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

	for ri, rule := range route.Spec.Rules {
		if rule.BackendRefs == nil {
			// Should not happen due to validation.
			// https://github.com/kubernetes-sigs/gateway-api/blob/v0.4.0/apis/v1alpha2/tcproute_types.go#L76
			continue
		}

		router := dynamic.TCPRouter{
			Rule:        "HostSNI(`*`)",
			EntryPoints: []string{listener.EPName},
			RuleSyntax:  "default",
		}

		if listener.Protocol == gatev1.TLSProtocolType && listener.TLS != nil {
			router.TLS = &dynamic.RouterTCPTLSConfig{
				Passthrough: listener.TLS.Mode != nil && *listener.TLS.Mode == gatev1.TLSModePassthrough,
			}
		}

		// Adding the gateway desc and the entryPoint desc prevents overlapping of routers build from the same routes.
		routeKey := provider.Normalize(fmt.Sprintf("%s-%s-%s-gw-%s-%s-ep-%s-%d", strings.ToLower(kindTCPRoute), route.Namespace, route.Name, listener.GWNamespace, listener.GWName, listener.EPName, ri))
		// Routing criteria should be introduced at some point.
		routerName := makeRouterName("", routeKey)

		if len(rule.BackendRefs) == 1 && isInternalService(rule.BackendRefs[0]) {
			router.Service = string(rule.BackendRefs[0].Name)
			conf.TCP.Routers[routerName] = &router
			continue
		}

		var serviceCondition *metav1.Condition
		router.Service, serviceCondition = p.loadTCPWRRService(conf, routerName, rule.BackendRefs, route)
		if serviceCondition != nil {
			condition = *serviceCondition
		}

		conf.TCP.Routers[routerName] = &router
	}

	return conf, condition
}

// loadTCPWRRService is generating a WRR service, even when there is only one target.
func (p *Provider) loadTCPWRRService(conf *dynamic.Configuration, routeKey string, backendRefs []gatev1.BackendRef, route *gatev1alpha2.TCPRoute) (string, *metav1.Condition) {
	name := routeKey + "-wrr"
	if _, ok := conf.TCP.Services[name]; ok {
		return name, nil
	}

	var wrr dynamic.TCPWeightedRoundRobin
	var condition *metav1.Condition
	for _, backendRef := range backendRefs {
		svcName, svc, errCondition := p.loadTCPService(route, backendRef)
		weight := ptr.To(int(ptr.Deref(backendRef.Weight, 1)))

		if errCondition != nil {
			condition = errCondition

			errName := routeKey + "-err-lb"
			conf.TCP.Services[errName] = &dynamic.TCPService{
				LoadBalancer: &dynamic.TCPServersLoadBalancer{
					Servers: []dynamic.TCPServer{},
				},
			}

			wrr.Services = append(wrr.Services, dynamic.TCPWRRService{
				Name:   errName,
				Weight: weight,
			})
			continue
		}

		if svc != nil {
			conf.TCP.Services[svcName] = svc
		}

		wrr.Services = append(wrr.Services, dynamic.TCPWRRService{
			Name:   svcName,
			Weight: weight,
		})
	}

	conf.TCP.Services[name] = &dynamic.TCPService{Weighted: &wrr}
	return name, condition
}

func (p *Provider) loadTCPService(route *gatev1alpha2.TCPRoute, backendRef gatev1.BackendRef) (string, *dynamic.TCPService, *metav1.Condition) {
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

	if err := p.isReferenceGranted(kindTCPRoute, route.Namespace, group, string(kind), string(backendRef.Name), namespace); err != nil {
		return serviceName, nil, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonRefNotPermitted),
			Message:            fmt.Sprintf("Cannot load TCPRoute BackendRef %s/%s/%s/%s: %s", group, kind, namespace, backendRef.Name, err),
		}
	}

	if group != groupCore || kind != kindService {
		name, err := p.loadTCPBackendRef(backendRef)
		if err != nil {
			return serviceName, nil, &metav1.Condition{
				Type:               string(gatev1.RouteConditionResolvedRefs),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: route.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.RouteReasonInvalidKind),
				Message:            fmt.Sprintf("Cannot load TCPRoute BackendRef %s/%s/%s/%s: %s", group, kind, namespace, backendRef.Name, err),
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
			Message:            fmt.Sprintf("Cannot load TCPRoute BackendRef %s/%s/%s/%s port is required", group, kind, namespace, backendRef.Name),
		}
	}

	portStr := strconv.FormatInt(int64(port), 10)
	serviceName = provider.Normalize(serviceName + "-" + portStr)

	lb, errCondition := p.loadTCPServers(namespace, route, backendRef)
	if errCondition != nil {
		return serviceName, nil, errCondition
	}

	return serviceName, &dynamic.TCPService{LoadBalancer: lb}, nil
}

func (p *Provider) loadTCPServers(namespace string, route *gatev1alpha2.TCPRoute, backendRef gatev1.BackendRef) (*dynamic.TCPServersLoadBalancer, *metav1.Condition) {
	backendAddresses, svcPort, err := p.getBackendAddresses(namespace, backendRef)
	if err != nil {
		return nil, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.GetGeneration(),
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonBackendNotFound),
			Message:            fmt.Sprintf("Cannot load TCPRoute BackendRef %s/%s: %s", namespace, backendRef.Name, err),
		}
	}

	if svcPort.Protocol != corev1.ProtocolTCP {
		return nil, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.GetGeneration(),
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonUnsupportedProtocol),
			Message:            fmt.Sprintf("Cannot load TCPRoute BackendRef %s/%s: only TCP protocol is supported", namespace, backendRef.Name),
		}
	}

	lb := &dynamic.TCPServersLoadBalancer{}

	for _, ba := range backendAddresses {
		lb.Servers = append(lb.Servers, dynamic.TCPServer{
			Address: net.JoinHostPort(ba.IP, strconv.Itoa(int(ba.Port))),
		})
	}
	return lb, nil
}

func (p *Provider) loadTCPBackendRef(backendRef gatev1.BackendRef) (string, error) {
	// Support for cross-provider references (e.g: api@internal).
	// This provides the same behavior as for IngressRoutes.
	if *backendRef.Kind == "TraefikService" && strings.Contains(string(backendRef.Name), "@") {
		return string(backendRef.Name), nil
	}

	return "", fmt.Errorf("unsupported BackendRef %s/%s/%s", *backendRef.Group, *backendRef.Kind, backendRef.Name)
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
