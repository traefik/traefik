package gateway

import (
	"context"
	"errors"
	"fmt"
	"net"
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
func (p *Provider) loadGRPCRoutes(ctx context.Context, gatewayListeners []gatewayListener, conf *dynamic.Configuration) {
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

		routeListeners := matchingGatewayListeners(gatewayListeners, route.Namespace, route.Spec.ParentRefs)
		if len(routeListeners) == 0 {
			continue
		}

		var parentStatuses []gatev1.RouteParentStatus
		for _, parentRef := range route.Spec.ParentRefs {
			parentStatus := &gatev1.RouteParentStatus{
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

				if accepted && !allowRoute(listener, route.Namespace, kindGRPCRoute) {
					parentStatus.Conditions = updateRouteConditionAccepted(parentStatus.Conditions, string(gatev1.RouteReasonNotAllowedByListeners))
					accepted = false
				}
				hostnames, ok := findMatchingHostnames(listener.Hostname, route.Spec.Hostnames)
				if accepted && !ok {
					parentStatus.Conditions = updateRouteConditionAccepted(parentStatus.Conditions, string(gatev1.RouteReasonNoMatchingListenerHostname))
					accepted = false
				}

				if accepted {
					// Gateway listener should have AttachedRoutes set even when Gateway has unresolved refs.
					listener.Status.AttachedRoutes++
					// Only consider the route attached if the listener is in an "attached" state.
					if listener.Attached {
						parentStatus.Conditions = updateRouteConditionAccepted(parentStatus.Conditions, string(gatev1.RouteReasonAccepted))
					}
				}

				routeConf, resolveRefCondition := p.loadGRPCRoute(logger.WithContext(ctx), listener, route, hostnames)
				if accepted && listener.Attached {
					mergeHTTPConfiguration(routeConf, conf)
				}

				parentStatus.Conditions = upsertRouteConditionResolvedRefs(parentStatus.Conditions, resolveRefCondition)
			}

			parentStatuses = append(parentStatuses, *parentStatus)
		}

		status := gatev1.GRPCRouteStatus{
			RouteStatus: gatev1.RouteStatus{
				Parents: parentStatuses,
			},
		}
		if err := p.client.UpdateGRPCRouteStatus(ctx, ktypes.NamespacedName{Namespace: route.Namespace, Name: route.Name}, status); err != nil {
			logger.Warn().
				Err(err).
				Msg("Unable to update GRPCRoute status")
		}
	}
}

func (p *Provider) loadGRPCRoute(ctx context.Context, listener gatewayListener, route *gatev1.GRPCRoute, hostnames []gatev1.Hostname) (*dynamic.Configuration, metav1.Condition) {
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
		routeKey := provider.Normalize(fmt.Sprintf("%s-%s-%s-gw-%s-%s-ep-%s-%d", strings.ToLower(kindGRPCRoute), route.Namespace, route.Name, listener.GWNamespace, listener.GWName, listener.EPName, ri))

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
								Weight: ptr.To(1),
							},
						},
					},
				}
				router.Service = errWrrName

			default:
				var serviceCondition *metav1.Condition
				router.Service, serviceCondition = p.loadGRPCService(conf, routerName, routeRule, route)
				if serviceCondition != nil {
					condition = *serviceCondition
				}
			}

			conf.HTTP.Routers[routerName] = &router
		}
	}

	return conf, condition
}

func (p *Provider) loadGRPCService(conf *dynamic.Configuration, routeKey string, routeRule gatev1.GRPCRouteRule, route *gatev1.GRPCRoute) (string, *metav1.Condition) {
	name := routeKey + "-wrr"
	if _, ok := conf.HTTP.Services[name]; ok {
		return name, nil
	}

	var wrr dynamic.WeightedRoundRobin
	var condition *metav1.Condition
	for _, backendRef := range routeRule.BackendRefs {
		svcName, svc, errCondition := p.loadGRPCBackendRef(route, backendRef)
		weight := ptr.To(int(ptr.Deref(backendRef.Weight, 1)))
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

func (p *Provider) loadGRPCBackendRef(route *gatev1.GRPCRoute, backendRef gatev1.GRPCBackendRef) (string, *dynamic.Service, *metav1.Condition) {
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

	portStr := strconv.FormatInt(int64(port), 10)
	serviceName = provider.Normalize(serviceName + "-" + portStr + "-grpc")

	lb, errCondition := p.loadGRPCServers(namespace, route, backendRef)
	if errCondition != nil {
		return serviceName, nil, errCondition
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

func (p *Provider) loadGRPCServers(namespace string, route *gatev1.GRPCRoute, backendRef gatev1.GRPCBackendRef) (*dynamic.ServersLoadBalancer, *metav1.Condition) {
	backendAddresses, svcPort, err := p.getBackendAddresses(namespace, backendRef.BackendRef)
	if err != nil {
		return nil, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonBackendNotFound),
			Message:            fmt.Sprintf("Cannot load GRPCBackendRef %s/%s: %s", namespace, backendRef.Name, err),
		}
	}

	if svcPort.Protocol != corev1.ProtocolTCP {
		return nil, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonUnsupportedProtocol),
			Message:            fmt.Sprintf("Cannot load GRPCBackendRef %s/%s: only TCP protocol is supported", namespace, backendRef.Name),
		}
	}

	protocol, err := getGRPCServiceProtocol(svcPort)
	if err != nil {
		return nil, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonUnsupportedProtocol),
			Message:            fmt.Sprintf("Cannot load GRPCBackendRef %s/%s: only \"kubernetes.io/h2c\" and \"https\" appProtocol is supported", namespace, backendRef.Name),
		}
	}

	lb := &dynamic.ServersLoadBalancer{}
	lb.SetDefaults()

	for _, ba := range backendAddresses {
		lb.Servers = append(lb.Servers, dynamic.Server{
			URL: fmt.Sprintf("%s://%s", protocol, net.JoinHostPort(ba.IP, strconv.Itoa(int(ba.Port)))),
		})
	}
	return lb, nil
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
		return "PathPrefix(`/`)"
	}

	sExpr := "[^/]+"
	if s := ptr.Deref(method.Service, ""); s != "" {
		sExpr = s
	}

	mExpr := "[^/]+"
	if m := ptr.Deref(method.Method, ""); m != "" {
		mExpr = m
	}

	return fmt.Sprintf("PathRegexp(`/%s/%s`)", sExpr, mExpr)
}

func buildGRPCHeaderRules(headers []gatev1.GRPCHeaderMatch) []string {
	var rules []string
	for _, header := range headers {
		switch ptr.Deref(header.Type, gatev1.GRPCHeaderMatchExact) {
		case gatev1.GRPCHeaderMatchExact:
			rules = append(rules, fmt.Sprintf("Header(`%s`,`%s`)", header.Name, header.Value))
		case gatev1.GRPCHeaderMatchRegularExpression:
			rules = append(rules, fmt.Sprintf("HeaderRegexp(`%s`,`%s`)", header.Name, header.Value))
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

	switch ap := *portSpec.AppProtocol; ap {
	case appProtocolH2C:
		return schemeH2C, nil
	case appProtocolHTTPS:
		return schemeHTTPS, nil
	default:
		return "", fmt.Errorf("unsupported application protocol %s", ap)
	}
}
