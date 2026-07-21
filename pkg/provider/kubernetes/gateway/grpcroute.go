package gateway

import (
	"context"
	"errors"
	"fmt"
	"net"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/provider"
	"google.golang.org/grpc/codes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	gatev1 "sigs.k8s.io/gateway-api/apis/v1"
)

// TODO: as described in the specification https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io%2fv1.GRPCRoute, we should check for hostname conflicts between HTTP and GRPC routes.
func (p *Provider) loadGRPCRoutes(ctx context.Context, gateways []gatewayWithListeners, conf *dynamic.Configuration, statusReport *statusReport) {
	routes, err := p.client.ListGRPCRoutes()
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Unable to list GRPCRoutes")
		return
	}

	for _, route := range routes {
		logger := log.Ctx(ctx).With().
			Str("grpc_route", route.Name).
			Str("namespace", route.Namespace).
			Logger()
		routeParentRefs := matchingGatewayListenersForParentRef(gateways, route.Namespace, route.Spec.ParentRefs)
		if len(routeParentRefs) == 0 {
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

				if accepted && !allowRoute(listener, route.Namespace, kindGRPCRoute) {
					if acceptedCondition.Status == metav1.ConditionFalse {
						acceptedCondition.Reason = string(gatev1.RouteReasonNotAllowedByListeners)
					}
					accepted = false
				}

				hostnames, ok := findMatchingHostnames(listener.Hostname, route.Spec.Hostnames)
				if accepted && !ok {
					if acceptedCondition.Status == metav1.ConditionFalse {
						acceptedCondition.Reason = string(gatev1.RouteReasonNoMatchingListenerHostname)
					}
					accepted = false
				}

				if accepted {
					// Gateway listener should have AttachedRoutes set even when Gateway has unresolved refs.
					listener.Status.AttachedRoutes++
				}

				// The ResolvedRefs condition must be reported for every parentRef,
				// even when the route does not attach to the listener.
				routeConf, condition := p.loadGRPCRoute(logger.WithContext(ctx), match.gatewayName, match.gatewayNamespace, listener, route, hostnames, statusReport)
				if resolvedRefCondition == nil || resolvedRefCondition.Status == metav1.ConditionTrue {
					resolvedRefCondition = new(condition)
				}

				if accepted && listener.Attached {
					mergeHTTPConfiguration(routeConf, conf)

					// Only consider the route attached if the listener is in an "attached" state.
					acceptedCondition.Reason = string(gatev1.RouteReasonAccepted)
					acceptedCondition.Status = metav1.ConditionTrue
				}
			}

			parentStatusConditions := []metav1.Condition{acceptedCondition}
			if resolvedRefCondition != nil {
				parentStatusConditions = append(parentStatusConditions, *resolvedRefCondition)
			}

			statusReport.RecordGRPCRouteStatus(ktypes.NamespacedName{Namespace: route.Namespace, Name: route.Name}, gatev1.RouteParentStatus{
				ParentRef:      match.parentRef,
				ControllerName: controllerName,
				Conditions:     parentStatusConditions,
			})
		}
	}
}

func (p *Provider) loadGRPCRoute(ctx context.Context, gatewayName, gatewayNamespace string, listener gatewayListener, route *gatev1.GRPCRoute, hostnames []gatev1.Hostname, statusReport *statusReport) (*dynamic.Configuration, metav1.Condition) {
	conf := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers:           make(map[string]*dynamic.Router),
			Middlewares:       make(map[string]*dynamic.Middleware),
			Services:          make(map[string]*dynamic.Service),
			ServersTransports: make(map[string]*dynamic.ServersTransport),
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
		// Adding the gateway desc and the entryPoint desc prevents overlapping of routers build from the same routes.
		routeKey := provider.Normalize(fmt.Sprintf("%s-%s-%s-gw-%s-%s-ep-%s-%d", strings.ToLower(kindGRPCRoute), route.Namespace, route.Name, gatewayNamespace, gatewayName, listener.EPName, ri))

		matches := routeRule.Matches
		if len(matches) == 0 {
			matches = []gatev1.GRPCRouteMatch{{}}
		}

		for _, match := range matches {
			rule, priority := buildGRPCMatchRule(hostnames, match)

			router := dynamic.Router{
				// "default" stands for the default rule syntax in Traefik v3, i.e. the v3 syntax.
				RuleSyntax:  "default",
				Rule:        rule,
				Priority:    priority,
				EntryPoints: []string{listener.EPName},
			}
			if listener.Protocol == gatev1.HTTPSProtocolType {
				router.TLS = &dynamic.RouterTLSConfig{}
			}

			var err error
			routerName := makeRouterName(rule, routeKey)
			router.Middlewares, err = p.loadGRPCMiddlewares(conf, route.Namespace, routerName, routeRule.Filters)
			switch {
			case err != nil:
				log.Ctx(ctx).Error().Err(err).Msg("Unable to load GRPC route filters")

				errWrrName := routerName + "-err-wrr"
				conf.HTTP.Services[errWrrName] = &dynamic.Service{
					Weighted: &dynamic.WeightedRoundRobin{
						Services: []dynamic.WRRService{
							{
								Name: "invalid-grpcroute-filter",
								GRPCStatus: &dynamic.GRPCStatus{
									Code: codes.Unavailable,
									Msg:  "Service Unavailable",
								},
								Weight: new(1),
							},
						},
					},
				}
				router.Service = errWrrName

			default:
				var serviceCondition *metav1.Condition
				router.Service, serviceCondition = p.loadGRPCService(gatewayName, listener, conf, routerName, routeRule, route, statusReport)
				if serviceCondition != nil {
					condition = *serviceCondition
				}
			}

			conf.HTTP.Routers[routerName] = &router
		}
	}

	return conf, condition
}

func (p *Provider) loadGRPCService(gatewayName string, listener gatewayListener, conf *dynamic.Configuration, routeKey string, routeRule gatev1.GRPCRouteRule, route *gatev1.GRPCRoute, statusReport *statusReport) (string, *metav1.Condition) {
	name := routeKey + "-wrr"
	if _, ok := conf.HTTP.Services[name]; ok {
		return name, nil
	}

	var wrr dynamic.WeightedRoundRobin
	var condition *metav1.Condition
	for bi, backendRef := range routeRule.BackendRefs {
		svcName, svc, errCondition := p.loadGRPCBackendRef(gatewayName, listener, conf, routeKey, route, bi, backendRef, statusReport)
		weight := new(int(ptr.Deref(backendRef.Weight, 1)))
		if errCondition != nil {
			condition = errCondition
			wrr.Services = append(wrr.Services, dynamic.WRRService{
				Name: svcName,
				GRPCStatus: &dynamic.GRPCStatus{
					Code: codes.Unavailable,
					Msg:  "Service Unavailable",
				},
				Weight: weight,
			})
			continue
		}

		if svc != nil {
			conf.HTTP.Services[svcName] = svc
		}

		wrr.Services = append(wrr.Services, dynamic.WRRService{
			Name:   svcName,
			Weight: weight,
		})
	}

	conf.HTTP.Services[name] = &dynamic.Service{Weighted: &wrr}
	return name, condition
}

func (p *Provider) loadGRPCBackendRef(gatewayName string, listener gatewayListener, conf *dynamic.Configuration, routeKey string, route *gatev1.GRPCRoute, backendIndex int, backendRef gatev1.GRPCBackendRef, statusReport *statusReport) (string, *dynamic.Service, *metav1.Condition) {
	kind := ptr.Deref(backendRef.Kind, kindService)

	group := groupCore
	if backendRef.Group != nil && *backendRef.Group != "" {
		group = string(*backendRef.Group)
	}

	namespace := route.Namespace
	if backendRef.Namespace != nil && *backendRef.Namespace != "" {
		namespace = string(*backendRef.Namespace)
	}

	serviceName := fmt.Sprintf("%s-svc-%s-%s-%d", routeKey, namespace, string(backendRef.Name), backendIndex)

	if group != groupCore || kind != kindService {
		return serviceName, nil, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonInvalidKind),
			Message:            fmt.Sprintf("Cannot load GRPCBackendRef %s/%s/%s/%s: only Kubernetes services are supported", group, kind, namespace, backendRef.Name),
		}
	}

	if err := p.isReferenceGranted(kindGRPCRoute, route.Namespace, group, string(kind), string(backendRef.Name), namespace); err != nil {
		return serviceName, nil, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonRefNotPermitted),
			Message:            fmt.Sprintf("Cannot load GRPCBackendRef %s/%s/%s/%s: %s", group, kind, namespace, backendRef.Name, err),
		}
	}

	port := ptr.Deref(backendRef.Port, gatev1.PortNumber(0))
	if port == 0 {
		return serviceName, nil, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonUnsupportedProtocol),
			Message:            fmt.Sprintf("Cannot load GRPCBackendRef %s/%s/%s/%s: port is required", group, kind, namespace, backendRef.Name),
		}
	}

	lb, st, errCondition := p.loadGRPCServers(gatewayName, namespace, route, backendRef, listener, statusReport)
	if errCondition != nil {
		return serviceName, nil, errCondition
	}

	if st != nil {
		lb.ServersTransport = serviceName
		conf.HTTP.ServersTransports[serviceName] = st
	}

	return serviceName, &dynamic.Service{LoadBalancer: lb}, nil
}

func (p *Provider) loadGRPCMiddlewares(conf *dynamic.Configuration, namespace, routerName string, filters []gatev1.GRPCRouteFilter) ([]string, error) {
	type namedMiddleware struct {
		Name   string
		Config *dynamic.Middleware
	}

	var middlewares []namedMiddleware
	for i, filter := range filters {
		name := fmt.Sprintf("%s-%s-%d", routerName, strings.ToLower(string(filter.Type)), i)
		switch filter.Type {
		case gatev1.GRPCRouteFilterRequestHeaderModifier:
			middlewares = append(middlewares, namedMiddleware{
				name,
				createRequestHeaderModifier(filter.RequestHeaderModifier),
			})

		case gatev1.GRPCRouteFilterExtensionRef:
			name, middleware, err := p.loadHTTPRouteFilterExtensionRef(namespace, filter.ExtensionRef)
			if err != nil {
				return nil, fmt.Errorf("loading ExtensionRef filter %s: %w", filter.Type, err)
			}
			middlewares = append(middlewares, namedMiddleware{
				name,
				middleware,
			})

		default:
			// As per the spec: https://gateway-api.sigs.k8s.io/api-types/httproute/#filters-optional
			// In all cases where incompatible or unsupported filters are
			// specified, implementations MUST add a warning condition to
			// status.
			return nil, fmt.Errorf("unsupported filter %s", filter.Type)
		}
	}

	var middlewareNames []string
	for _, m := range middlewares {
		if m.Config != nil {
			conf.HTTP.Middlewares[m.Name] = m.Config
		}
		middlewareNames = append(middlewareNames, m.Name)
	}

	return middlewareNames, nil
}

func (p *Provider) loadGRPCServers(gatewayName, namespace string, route *gatev1.GRPCRoute, backendRef gatev1.GRPCBackendRef, listener gatewayListener, statusReport *statusReport) (*dynamic.ServersLoadBalancer, *dynamic.ServersTransport, *metav1.Condition) {
	backendAddresses, svcPort, err := p.getBackendAddresses(namespace, backendRef.BackendRef)
	if err != nil {
		return nil, nil, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonBackendNotFound),
			Message:            fmt.Sprintf("Cannot load GRPCBackendRef %s/%s: %s", namespace, backendRef.Name, err),
		}
	}

	backendTLSPolicies, err := p.client.ListBackendTLSPoliciesForService(namespace, string(backendRef.Name))
	if err != nil {
		return nil, nil, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonRefNotPermitted),
			Message:            fmt.Sprintf("Cannot list BackendTLSPolicies for Service %s/%s: %s", namespace, string(backendRef.Name), err),
		}
	}

	// Sort BackendTLSPolicies by creation timestamp, then by name to match the BackendTLSPolicy requirements.
	slices.SortStableFunc(backendTLSPolicies, func(a, b *gatev1.BackendTLSPolicy) int {
		cmpTime := a.CreationTimestamp.Time.Compare(b.CreationTimestamp.Time)
		if cmpTime == 0 {
			return strings.Compare(a.Name, b.Name)
		}
		return cmpTime
	})

	var serversTransport *dynamic.ServersTransport
	for _, policy := range backendTLSPolicies {
		for _, targetRef := range policy.Spec.TargetRefs {
			// Skip targetRefs that doesn't match the backendRef,
			// since a BackendTLSPolicy can select multiple services.
			if targetRef.Name != backendRef.Name {
				continue
			}
			// Skip the targetRef if the sectionName doesn't match the backendRef port.
			if targetRef.SectionName != nil && svcPort.Name != string(*targetRef.SectionName) {
				continue
			}

			policyAncestorStatus := gatev1.PolicyAncestorStatus{
				AncestorRef: gatev1.ParentReference{
					Group:       new(gatev1.Group(groupGateway)),
					Kind:        new(gatev1.Kind(kindGateway)),
					Namespace:   new(gatev1.Namespace(namespace)),
					Name:        gatev1.ObjectName(gatewayName),
					SectionName: new(gatev1.SectionName(listener.Name)),
				},
				ControllerName: controllerName,
			}

			// Multiple BackendTLSPolicies can match the same service port, meaning that there is a conflict.
			if serversTransport != nil {
				policyAncestorStatus.Conditions = append(policyAncestorStatus.Conditions,
					metav1.Condition{
						Type:               string(gatev1.BackendTLSPolicyConditionResolvedRefs),
						Status:             metav1.ConditionFalse,
						ObservedGeneration: policy.Generation,
						LastTransitionTime: metav1.Now(),
						Reason:             string(gatev1.BackendTLSPolicyReasonResolvedRefs),
					},
					metav1.Condition{
						Type:               string(gatev1.PolicyConditionAccepted),
						Status:             metav1.ConditionFalse,
						ObservedGeneration: policy.Generation,
						LastTransitionTime: metav1.Now(),
						Reason:             string(gatev1.PolicyReasonConflicted),
					},
				)

				statusReport.RecordBackendTLSPolicyStatus(ktypes.NamespacedName{Namespace: policy.Namespace, Name: policy.Name}, policyAncestorStatus)

				continue
			}

			var resolvedRefCondition metav1.Condition
			serversTransport, resolvedRefCondition = p.loadServersTransport(namespace, policy)

			policyAncestorStatus.Conditions = append(policyAncestorStatus.Conditions, resolvedRefCondition)
			if resolvedRefCondition.Status == metav1.ConditionFalse {
				policyAncestorStatus.Conditions = append(policyAncestorStatus.Conditions, metav1.Condition{
					Type:               string(gatev1.PolicyConditionAccepted),
					Status:             metav1.ConditionFalse,
					ObservedGeneration: policy.Generation,
					LastTransitionTime: metav1.Now(),
					Reason:             string(gatev1.BackendTLSPolicyReasonNoValidCACertificate),
				})
			} else {
				policyAncestorStatus.Conditions = append(policyAncestorStatus.Conditions, metav1.Condition{
					Type:               string(gatev1.PolicyConditionAccepted),
					Status:             metav1.ConditionTrue,
					ObservedGeneration: policy.Generation,
					LastTransitionTime: metav1.Now(),
					Reason:             string(gatev1.PolicyReasonAccepted),
				})
			}

			statusReport.RecordBackendTLSPolicyStatus(ktypes.NamespacedName{Namespace: policy.Namespace, Name: policy.Name}, policyAncestorStatus)

			// When something went wrong during the loading of a ServersTransport,
			// we stop here and return a route condition error.
			if resolvedRefCondition.Status == metav1.ConditionFalse {
				return nil, nil, &metav1.Condition{
					Type:               string(gatev1.RouteConditionResolvedRefs),
					Status:             metav1.ConditionFalse,
					ObservedGeneration: route.Generation,
					LastTransitionTime: metav1.Now(),
					Reason:             string(gatev1.RouteReasonRefNotPermitted),
					Message:            fmt.Sprintf("Cannot apply BackendTLSPolicy for Service %s/%s: %s", namespace, string(backendRef.Name), resolvedRefCondition.Message),
				}
			}
		}
	}

	// If a ServersTransport is set, it means a BackendTLSPolicy matched the service port, and we can safely assume the protocol is HTTPS.
	// When no ServersTransport is set, we need to determine the protocol based on the service port.
	protocol := "https"
	if serversTransport == nil {
		protocol, err = getGRPCServiceProtocol(svcPort)
		if err != nil {
			return nil, nil, &metav1.Condition{
				Type:               string(gatev1.RouteConditionResolvedRefs),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: route.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.RouteReasonUnsupportedProtocol),
				Message:            fmt.Sprintf("Cannot load GRPCBackendRef %s/%s: %s", namespace, backendRef.Name, err),
			}
		}
	}

	lb := &dynamic.ServersLoadBalancer{}
	lb.SetDefaults()

	for _, ba := range backendAddresses {
		lb.Servers = append(lb.Servers, dynamic.Server{
			URL: fmt.Sprintf("%s://%s", protocol, net.JoinHostPort(ba.IP, strconv.Itoa(int(ba.Port)))),
		})
	}

	return lb, serversTransport, nil
}

func buildGRPCMatchRule(hostnames []gatev1.Hostname, match gatev1.GRPCRouteMatch) (string, int) {
	var matchRules []string

	methodRule := buildGRPCMethodRule(match.Method)
	matchRules = append(matchRules, methodRule)

	headerRules := buildGRPCHeaderRules(match.Headers)
	matchRules = append(matchRules, headerRules...)

	matchRulesStr := strings.Join(matchRules, " && ")

	hostRule, priority := buildHostRule(hostnames)
	if hostRule == "" {
		return matchRulesStr, len(matchRulesStr)
	}
	return hostRule + " && " + matchRulesStr, priority + len(matchRulesStr)
}

func buildGRPCMethodRule(method *gatev1.GRPCMethodMatch) string {
	if method == nil {
		return `PathPrefix("/")`
	}

	isExact := method.Type == nil || *method.Type == gatev1.GRPCMethodMatchExact

	sExpr := "[^/]+"
	if s := ptr.Deref(method.Service, ""); s != "" {
		if isExact {
			sExpr = regexp.QuoteMeta(s)
		} else {
			sExpr = s
		}
	}

	mExpr := "[^/]+"
	if m := ptr.Deref(method.Method, ""); m != "" {
		if isExact {
			mExpr = regexp.QuoteMeta(m)
		} else {
			mExpr = m
		}
	}

	return fmt.Sprintf("PathRegexp(%q)", fmt.Sprintf("/%s/%s", sExpr, mExpr))
}

func buildGRPCHeaderRules(headers []gatev1.GRPCHeaderMatch) []string {
	var rules []string
	for _, header := range headers {
		switch ptr.Deref(header.Type, gatev1.GRPCHeaderMatchExact) {
		case gatev1.GRPCHeaderMatchExact:
			rules = append(rules, fmt.Sprintf("Header(%q,%q)", header.Name, header.Value))
		case gatev1.GRPCHeaderMatchRegularExpression:
			rules = append(rules, fmt.Sprintf("HeaderRegexp(%q,%q)", header.Name, header.Value))
		}
	}

	return rules
}

func getGRPCServiceProtocol(portSpec corev1.ServicePort) (string, error) {
	if portSpec.Protocol != corev1.ProtocolTCP {
		return "", errors.New("only TCP protocol is supported")
	}

	if portSpec.AppProtocol == nil {
		return schemeH2C, nil
	}

	switch ap := strings.ToLower(*portSpec.AppProtocol); ap {
	case appProtocolH2C:
		return schemeH2C, nil
	case appProtocolHTTPS:
		return schemeHTTPS, nil
	default:
		return "", fmt.Errorf("unsupported application protocol %s", ap)
	}
}
