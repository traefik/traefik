package gateway

import (
	"context"
	"fmt"
	"net"
	"regexp"
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

func (p *Provider) loadTLSRoutes(ctx context.Context, gatewayListeners []gatewayListener, conf *dynamic.Configuration) {
	logger := log.Ctx(ctx)
	routes, err := p.client.ListTLSRoutes()
	if err != nil {
		logger.Error().Err(err).Msgf("Unable to list TLSRoute")
		return
	}

	for _, route := range routes {
		logger := log.Ctx(ctx).With().
			Str("tls_route", route.Name).
			Str("namespace", route.Namespace).Logger()

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

				if accepted && !allowRoute(listener, route.Namespace, kindTLSRoute) {
					parentStatus.Conditions = updateRouteConditionAccepted(parentStatus.Conditions, string(gatev1.RouteReasonNotAllowedByListeners))
					accepted = false
				}
				hostnames, ok := findMatchingHostnames(listener.Hostname, route.Spec.Hostnames)
				if accepted && !ok {
					parentStatus.Conditions = updateRouteConditionAccepted(parentStatus.Conditions, string(gatev1.RouteReasonNoMatchingListenerHostname))
					accepted = false
				}

				if accepted {
					listener.Status.AttachedRoutes++
					// only consider the route attached if the listener is in an "attached" state.
					if listener.Attached {
						parentStatus.Conditions = updateRouteConditionAccepted(parentStatus.Conditions, string(gatev1.RouteReasonAccepted))
					}
				}

				routeConf, resolveRefCondition := p.loadTLSRoute(listener, route, hostnames)
				if accepted && listener.Attached {
					mergeTCPConfiguration(routeConf, conf)
				}
				parentStatus.Conditions = upsertRouteConditionResolvedRefs(parentStatus.Conditions, resolveRefCondition)
			}

			parentStatuses = append(parentStatuses, *parentStatus)
		}

		routeStatus := gatev1alpha2.TLSRouteStatus{
			RouteStatus: gatev1alpha2.RouteStatus{
				Parents: parentStatuses,
			},
		}
		if err := p.client.UpdateTLSRouteStatus(ctx, ktypes.NamespacedName{Namespace: route.Namespace, Name: route.Name}, routeStatus); err != nil {
			logger.Warn().
				Err(err).
				Msg("Unable to update TLSRoute status")
		}
	}
}

func (p *Provider) loadTLSRoute(listener gatewayListener, route *gatev1alpha2.TLSRoute, hostnames []gatev1.Hostname) (*dynamic.Configuration, metav1.Condition) {
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

	for ri, routeRule := range route.Spec.Rules {
		if len(routeRule.BackendRefs) == 0 {
			// Should not happen due to validation.
			// https://github.com/kubernetes-sigs/gateway-api/blob/v0.4.0/apis/v1alpha2/tlsroute_types.go#L120
			continue
		}

		rule, priority := hostSNIRule(hostnames)
		router := dynamic.TCPRouter{
			RuleSyntax:  "default",
			Rule:        rule,
			Priority:    priority,
			EntryPoints: []string{listener.EPName},
			TLS: &dynamic.RouterTCPTLSConfig{
				Passthrough: listener.TLS != nil && listener.TLS.Mode != nil && *listener.TLS.Mode == gatev1.TLSModePassthrough,
			},
		}

		// Adding the gateway desc and the entryPoint desc prevents overlapping of routers build from the same routes.
		routeKey := provider.Normalize(fmt.Sprintf("%s-%s-%s-gw-%s-%s-ep-%s-%d", strings.ToLower(kindTLSRoute), route.Namespace, route.Name, listener.GWNamespace, listener.GWName, listener.EPName, ri))
		// Routing criteria should be introduced at some point.
		routerName := makeRouterName("", routeKey)

		if len(routeRule.BackendRefs) == 1 && isInternalService(routeRule.BackendRefs[0]) {
			router.Service = string(routeRule.BackendRefs[0].Name)
			conf.TCP.Routers[routerName] = &router
			continue
		}

		var serviceCondition *metav1.Condition
		router.Service, serviceCondition = p.loadTLSWRRService(conf, routerName, routeRule.BackendRefs, route)
		if serviceCondition != nil {
			condition = *serviceCondition
		}

		conf.TCP.Routers[routerName] = &router
	}

	return conf, condition
}

// loadTLSWRRService is generating a WRR service, even when there is only one target.
func (p *Provider) loadTLSWRRService(conf *dynamic.Configuration, routeKey string, backendRefs []gatev1.BackendRef, route *gatev1alpha2.TLSRoute) (string, *metav1.Condition) {
	name := routeKey + "-wrr"
	if _, ok := conf.TCP.Services[name]; ok {
		return name, nil
	}

	var wrr dynamic.TCPWeightedRoundRobin
	var condition *metav1.Condition
	for _, backendRef := range backendRefs {
		svcName, svc, errCondition := p.loadTLSService(route, backendRef)
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

func (p *Provider) loadTLSService(route *gatev1alpha2.TLSRoute, backendRef gatev1.BackendRef) (string, *dynamic.TCPService, *metav1.Condition) {
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

	if err := p.isReferenceGranted(kindTLSRoute, route.Namespace, group, string(kind), string(backendRef.Name), namespace); err != nil {
		return serviceName, nil, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonRefNotPermitted),
			Message:            fmt.Sprintf("Cannot load TLSRoute BackendRef %s/%s/%s/%s: %s", group, kind, namespace, backendRef.Name, err),
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
				Message:            fmt.Sprintf("Cannot load TLSRoute BackendRef %s/%s/%s/%s: %s", group, kind, namespace, backendRef.Name, err),
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
			Message:            fmt.Sprintf("Cannot load TLSRoute BackendRef %s/%s/%s/%s port is required", group, kind, namespace, backendRef.Name),
		}
	}

	portStr := strconv.FormatInt(int64(port), 10)
	serviceName = provider.Normalize(serviceName + "-" + portStr)

	lb, errCondition := p.loadTLSServers(namespace, route, backendRef)
	if errCondition != nil {
		return serviceName, nil, errCondition
	}

	return serviceName, &dynamic.TCPService{LoadBalancer: lb}, nil
}

func (p *Provider) loadTLSServers(namespace string, route *gatev1alpha2.TLSRoute, backendRef gatev1.BackendRef) (*dynamic.TCPServersLoadBalancer, *metav1.Condition) {
	backendAddresses, svcPort, err := p.getBackendAddresses(namespace, backendRef)
	if err != nil {
		return nil, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.GetGeneration(),
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonBackendNotFound),
			Message:            fmt.Sprintf("Cannot load TLSRoute BackendRef %s/%s: %s", namespace, backendRef.Name, err),
		}
	}

	if svcPort.Protocol != corev1.ProtocolTCP {
		return nil, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.GetGeneration(),
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonUnsupportedProtocol),
			Message:            fmt.Sprintf("Cannot load TLSRoute BackendRef %s/%s: only TCP protocol is supported", namespace, backendRef.Name),
		}
	}

	lb := &dynamic.TCPServersLoadBalancer{}

	for _, ba := range backendAddresses {
		lb.Servers = append(lb.Servers, dynamic.TCPServer{
			// TODO determine whether the servers needs TLS, from the port?
			Address: net.JoinHostPort(ba.IP, strconv.Itoa(int(ba.Port))),
		})
	}
	return lb, nil
}

func hostSNIRule(hostnames []gatev1.Hostname) (string, int) {
	var priority int

	rules := make([]string, 0, len(hostnames))
	uniqHostnames := map[gatev1.Hostname]struct{}{}

	for _, hostname := range hostnames {
		if len(hostname) == 0 {
			continue
		}

		host := string(hostname)
		wildcard := strings.Count(host, "*")

		thisPriority := len(hostname) - wildcard

		if priority < thisPriority {
			priority = thisPriority
		}

		if _, exists := uniqHostnames[hostname]; exists {
			continue
		}

		uniqHostnames[hostname] = struct{}{}

		if wildcard == 0 {
			rules = append(rules, fmt.Sprintf("HostSNI(`%s`)", host))
			continue
		}

		host = strings.Replace(regexp.QuoteMeta(host), `\*\.`, `[a-z0-9-\.]+\.`, 1)
		rules = append(rules, fmt.Sprintf("HostSNIRegexp(`^%s$`)", host))
	}

	if len(hostnames) == 0 || len(rules) == 0 {
		return "HostSNI(`*`)", 0
	}

	return strings.Join(rules, " || "), priority
}
