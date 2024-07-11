package gateway

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/provider"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktypes "k8s.io/apimachinery/pkg/types"
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
		logger := log.Ctx(ctx).With().Str("tlsroute", route.Name).Str("namespace", route.Namespace).Logger()

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
				if !allowRoute(listener, route.Namespace, kindTLSRoute) {
					parentStatus.Conditions = updateRouteConditionAccepted(parentStatus.Conditions, string(gatev1.RouteReasonNotAllowedByListeners))
					accepted = false
				}
				hostnames, ok := findMatchingHostnames(listener.Hostname, route.Spec.Hostnames)
				if !ok {
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

	router := dynamic.TCPRouter{
		RuleSyntax:  "v3",
		Rule:        hostSNIRule(hostnames),
		EntryPoints: []string{listener.EPName},
		TLS: &dynamic.RouterTCPTLSConfig{
			Passthrough: listener.TLS != nil && listener.TLS.Mode != nil && *listener.TLS.Mode == gatev1.TLSModePassthrough,
		},
	}

	// Adding the gateway desc and the entryPoint desc prevents overlapping of routers build from the same routes.
	routeKey := provider.Normalize(route.Namespace + "-" + route.Name + "-" + listener.GWName + "-" + listener.EPName)
	routerName := makeRouterName(router.Rule, routeKey)

	var ruleServiceNames []string
	for i, routeRule := range route.Spec.Rules {
		if len(routeRule.BackendRefs) == 0 {
			// Should not happen due to validation.
			// https://github.com/kubernetes-sigs/gateway-api/blob/v0.4.0/apis/v1alpha2/tlsroute_types.go#L120
			continue
		}

		wrrService, subServices, err := p.loadTCPServices(route.Namespace, routeRule.BackendRefs)
		if err != nil {
			// update "ResolvedRefs" status true with "InvalidBackendRefs" reason
			condition = metav1.Condition{
				Type:               string(gatev1.RouteConditionResolvedRefs),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: route.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.RouteReasonBackendNotFound),
				Message:            fmt.Sprintf("Cannot load TLSRoute service %s/%s: %v", route.Namespace, route.Name, err),
			}
			continue
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

func hostSNIRule(hostnames []gatev1.Hostname) string {
	rules := make([]string, 0, len(hostnames))
	uniqHostnames := map[gatev1.Hostname]struct{}{}

	for _, hostname := range hostnames {
		if len(hostname) == 0 {
			continue
		}

		if _, exists := uniqHostnames[hostname]; exists {
			continue
		}

		host := string(hostname)
		uniqHostnames[hostname] = struct{}{}

		wildcard := strings.Count(host, "*")
		if wildcard == 0 {
			rules = append(rules, fmt.Sprintf("HostSNI(`%s`)", host))
			continue
		}

		host = strings.Replace(regexp.QuoteMeta(host), `\*\.`, `[a-z0-9-\.]+\.`, 1)
		rules = append(rules, fmt.Sprintf("HostSNIRegexp(`^%s$`)", host))
	}

	if len(hostnames) == 0 || len(rules) == 0 {
		return "HostSNI(`*`)"
	}

	return strings.Join(rules, " || ")
}
