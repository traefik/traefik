package gateway

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
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
)

func (p *Provider) loadHTTPRoutes(ctx context.Context, client Client, gatewayListeners []gatewayListener, conf *dynamic.Configuration) {
	routes, err := client.ListHTTPRoutes()
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Unable to list HTTPRoutes")
		return
	}

	for _, route := range routes {
		logger := log.Ctx(ctx).With().
			Str("http_route", route.Name).
			Str("namespace", route.Namespace).
			Logger()

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

			for _, listener := range gatewayListeners {
				if !matchListener(listener, route.Namespace, parentRef) {
					continue
				}

				accepted := true
				if !allowRoute(listener, route.Namespace, kindHTTPRoute) {
					parentStatus.Conditions = updateRouteConditionAccepted(parentStatus.Conditions, string(gatev1.RouteReasonNotAllowedByListeners))
					accepted = false
				}
				hostnames, ok := findMatchingHostnames(listener.Hostname, route.Spec.Hostnames)
				if !ok {
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

				routeConf, resolveRefCondition := p.loadHTTPRoute(logger.WithContext(ctx), client, listener, route, hostnames)
				if accepted && listener.Attached {
					mergeHTTPConfiguration(routeConf, conf)
				}

				parentStatus.Conditions = upsertRouteConditionResolvedRefs(parentStatus.Conditions, resolveRefCondition)
			}

			parentStatuses = append(parentStatuses, *parentStatus)
		}

		status := gatev1.HTTPRouteStatus{
			RouteStatus: gatev1.RouteStatus{
				Parents: parentStatuses,
			},
		}
		if err := client.UpdateHTTPRouteStatus(ctx, ktypes.NamespacedName{Namespace: route.Namespace, Name: route.Name}, status); err != nil {
			logger.Error().
				Err(err).
				Msg("Unable to update HTTPRoute status")
		}
	}
}

func (p *Provider) loadHTTPRoute(ctx context.Context, client Client, listener gatewayListener, route *gatev1.HTTPRoute, hostnames []gatev1.Hostname) (*dynamic.Configuration, metav1.Condition) {
	routeConf := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers:           make(map[string]*dynamic.Router),
			Middlewares:       make(map[string]*dynamic.Middleware),
			Services:          make(map[string]*dynamic.Service),
			ServersTransports: make(map[string]*dynamic.ServersTransport),
		},
	}

	routeCondition := metav1.Condition{
		Type:               string(gatev1.RouteConditionResolvedRefs),
		Status:             metav1.ConditionTrue,
		ObservedGeneration: route.Generation,
		LastTransitionTime: metav1.Now(),
		Reason:             string(gatev1.RouteConditionResolvedRefs),
	}

	for _, routeRule := range route.Spec.Rules {
		rule, priority := buildRouterRule(hostnames, routeRule.Matches)
		router := dynamic.Router{
			RuleSyntax:  "v3",
			Rule:        rule,
			Priority:    priority,
			EntryPoints: []string{listener.EPName},
		}
		if listener.Protocol == gatev1.HTTPSProtocolType {
			router.TLS = &dynamic.RouterTLSConfig{}
		}

		// Adding the gateway desc and the entryPoint desc prevents overlapping of routers build from the same routes.
		routerName := route.Name + "-" + listener.GWName + "-" + listener.EPName
		routerKey := makeRouterKey(router.Rule, makeID(route.Namespace, routerName))

		var wrr dynamic.WeightedRoundRobin
		wrrName := provider.Normalize(routerKey + "-wrr")

		middlewares, err := p.loadMiddlewares(listener.Protocol, route.Namespace, routerKey, routeRule.Filters)
		if err != nil {
			log.Ctx(ctx).Error().
				Err(err).
				Msg("Unable to load HTTPRoute filters")

			wrr.Services = append(wrr.Services, dynamic.WRRService{
				Name:   "invalid-httproute-filter",
				Status: ptr.To(500),
				Weight: ptr.To(1),
			})

			routeConf.HTTP.Services[wrrName] = &dynamic.Service{Weighted: &wrr}
			router.Service = wrrName
		} else {
			for name, middleware := range middlewares {
				// If the middleware config is nil in the return of the loadMiddlewares function,
				// it means that we just need a reference to that middleware.
				if middleware != nil {
					routeConf.HTTP.Middlewares[name] = middleware
				}

				router.Middlewares = append(router.Middlewares, name)
			}

			// Traefik internal service can be used only if there is only one BackendRef service reference.
			if len(routeRule.BackendRefs) == 1 && isInternalService(routeRule.BackendRefs[0].BackendRef) {
				router.Service = string(routeRule.BackendRefs[0].Name)
			} else {
				for _, backendRef := range routeRule.BackendRefs {
					name, svc, errCondition := p.loadHTTPService(client, route, backendRef)
					weight := ptr.To(int(ptr.Deref(backendRef.Weight, 1)))
					if errCondition != nil {
						routeCondition = *errCondition
						wrr.Services = append(wrr.Services, dynamic.WRRService{
							Name:   name,
							Status: ptr.To(500),
							Weight: weight,
						})
						continue
					}

					if svc != nil {
						routeConf.HTTP.Services[name] = svc
					}

					wrr.Services = append(wrr.Services, dynamic.WRRService{
						Name:   name,
						Weight: weight,
					})
				}

				routeConf.HTTP.Services[wrrName] = &dynamic.Service{Weighted: &wrr}
				router.Service = wrrName
			}
		}

		rt := &router
		p.applyRouterTransform(ctx, rt, route)

		routerKey = provider.Normalize(routerKey)
		routeConf.HTTP.Routers[routerKey] = rt
	}

	return routeConf, routeCondition
}

// loadHTTPService returns a dynamic.Service config corresponding to the given gatev1.HTTPBackendRef.
// Note that the returned dynamic.Service config can be nil (for cross-provider, internal services, and backendFunc).
func (p *Provider) loadHTTPService(client Client, route *gatev1.HTTPRoute, backendRef gatev1.HTTPBackendRef) (string, *dynamic.Service, *metav1.Condition) {
	kind := ptr.Deref(backendRef.Kind, "Service")

	group := groupCore
	if backendRef.Group != nil && *backendRef.Group != "" {
		group = string(*backendRef.Group)
	}

	namespace := route.Namespace
	if backendRef.Namespace != nil && *backendRef.Namespace != "" {
		namespace = string(*backendRef.Namespace)
	}

	serviceName := provider.Normalize(makeID(namespace, string(backendRef.Name)))

	if err := isReferenceGranted(client, groupGateway, kindHTTPRoute, route.Namespace, group, string(kind), string(backendRef.Name), namespace); err != nil {
		return serviceName, nil, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonRefNotPermitted),
			Message:            fmt.Sprintf("Cannot load HTTPBackendRef %s/%s/%s/%s: %s", group, kind, namespace, backendRef.Name, err),
		}
	}

	if group != groupCore || kind != "Service" {
		name, service, err := p.loadHTTPBackendRef(namespace, backendRef)
		if err != nil {
			return serviceName, nil, &metav1.Condition{
				Type:               string(gatev1.RouteConditionResolvedRefs),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: route.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.RouteReasonInvalidKind),
				Message:            fmt.Sprintf("Cannot load HTTPBackendRef %s/%s/%s/%s: %s", group, kind, namespace, backendRef.Name, err),
			}
		}

		return name, service, nil
	}

	port := ptr.Deref(backendRef.Port, gatev1.PortNumber(0))
	if port == 0 {
		return serviceName, nil, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonUnsupportedProtocol),
			Message:            fmt.Sprintf("Cannot load HTTPBackendRef %s/%s/%s/%s port is required", group, kind, namespace, backendRef.Name),
		}
	}

	portStr := strconv.FormatInt(int64(port), 10)
	serviceName = provider.Normalize(serviceName + "-" + portStr)

	lb, err := loadHTTPServers(client, namespace, backendRef)
	if err != nil {
		return serviceName, nil, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonBackendNotFound),
			Message:            fmt.Sprintf("Cannot load HTTPBackendRef %s/%s/%s/%s: %s", group, kind, namespace, backendRef.Name, err),
		}
	}

	return serviceName, &dynamic.Service{LoadBalancer: lb}, nil
}

func (p *Provider) loadHTTPBackendRef(namespace string, backendRef gatev1.HTTPBackendRef) (string, *dynamic.Service, error) {
	// Support for cross-provider references (e.g: api@internal).
	// This provides the same behavior as for IngressRoutes.
	if *backendRef.Kind == "TraefikService" && strings.Contains(string(backendRef.Name), "@") {
		return string(backendRef.Name), nil, nil
	}

	backendFunc, ok := p.groupKindBackendFuncs[string(*backendRef.Group)][string(*backendRef.Kind)]
	if !ok {
		return "", nil, fmt.Errorf("unsupported HTTPBackendRef %s/%s/%s", *backendRef.Group, *backendRef.Kind, backendRef.Name)
	}
	if backendFunc == nil {
		return "", nil, fmt.Errorf("undefined backendFunc for HTTPBackendRef %s/%s/%s", *backendRef.Group, *backendRef.Kind, backendRef.Name)
	}

	return backendFunc(string(backendRef.Name), namespace)
}

func (p *Provider) loadMiddlewares(listenerProtocol gatev1.ProtocolType, namespace, prefix string, filters []gatev1.HTTPRouteFilter) (map[string]*dynamic.Middleware, error) {
	middlewares := make(map[string]*dynamic.Middleware)

	for i, filter := range filters {
		switch filter.Type {
		case gatev1.HTTPRouteFilterRequestRedirect:
			middlewareName := provider.Normalize(fmt.Sprintf("%s-%s-%d", prefix, strings.ToLower(string(filter.Type)), i))
			middlewares[middlewareName] = createRedirectRegexMiddleware(listenerProtocol, filter.RequestRedirect)

		case gatev1.HTTPRouteFilterRequestHeaderModifier:
			middlewareName := provider.Normalize(fmt.Sprintf("%s-%s-%d", prefix, strings.ToLower(string(filter.Type)), i))
			middlewares[middlewareName] = createRequestHeaderModifier(filter.RequestHeaderModifier)

		case gatev1.HTTPRouteFilterExtensionRef:
			name, middleware, err := p.loadHTTPRouteFilterExtensionRef(namespace, filter.ExtensionRef)
			if err != nil {
				return nil, fmt.Errorf("loading ExtensionRef filter %s: %w", filter.Type, err)
			}

			middlewares[name] = middleware

		default:
			// As per the spec: https://gateway-api.sigs.k8s.io/api-types/httproute/#filters-optional
			// In all cases where incompatible or unsupported filters are
			// specified, implementations MUST add a warning condition to
			// status.
			return nil, fmt.Errorf("unsupported filter %s", filter.Type)
		}
	}

	return middlewares, nil
}

func (p *Provider) loadHTTPRouteFilterExtensionRef(namespace string, extensionRef *gatev1.LocalObjectReference) (string, *dynamic.Middleware, error) {
	if extensionRef == nil {
		return "", nil, errors.New("filter extension ref undefined")
	}

	filterFunc, ok := p.groupKindFilterFuncs[string(extensionRef.Group)][string(extensionRef.Kind)]
	if !ok {
		return "", nil, fmt.Errorf("unsupported filter extension ref %s/%s/%s", extensionRef.Group, extensionRef.Kind, extensionRef.Name)
	}
	if filterFunc == nil {
		return "", nil, fmt.Errorf("undefined filterFunc for filter extension ref %s/%s/%s", extensionRef.Group, extensionRef.Kind, extensionRef.Name)
	}

	return filterFunc(string(extensionRef.Name), namespace)
}

// TODO support cross namespace through ReferencePolicy.
func loadHTTPServers(client Client, namespace string, backendRef gatev1.HTTPBackendRef) (*dynamic.ServersLoadBalancer, error) {
	service, exists, err := client.GetService(namespace, string(backendRef.Name))
	if err != nil {
		return nil, fmt.Errorf("getting service: %w", err)
	}
	if !exists {
		return nil, errors.New("service not found")
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
		return nil, errors.New("service port not found")
	}

	endpoints, endpointsExists, err := client.GetEndpoints(namespace, string(backendRef.Name))
	if err != nil {
		return nil, fmt.Errorf("getting endpoints: %w", err)
	}
	if !endpointsExists {
		return nil, errors.New("endpoints not found")
	}

	if len(endpoints.Subsets) == 0 {
		return nil, errors.New("subset not found")
	}

	lb := &dynamic.ServersLoadBalancer{}
	lb.SetDefaults()

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
			return nil, errors.New("cannot define a port")
		}

		protocol := getProtocol(portSpec)

		portStr = strconv.FormatInt(int64(port), 10)
		for _, addr := range subset.Addresses {
			lb.Servers = append(lb.Servers, dynamic.Server{
				URL: fmt.Sprintf("%s://%s", protocol, net.JoinHostPort(addr.IP, portStr)),
			})
		}
	}

	return lb, nil
}

func buildHostRule(hostnames []gatev1.Hostname) (string, int) {
	var rules []string
	var priority int

	for _, hostname := range hostnames {
		host := string(hostname)

		if priority < len(host) {
			priority = len(host)
		}

		wildcard := strings.Count(host, "*")
		if wildcard == 0 {
			rules = append(rules, fmt.Sprintf("Host(`%s`)", host))
			continue
		}

		host = strings.Replace(regexp.QuoteMeta(host), `\*\.`, `[a-z0-9-\.]+\.`, 1)
		rules = append(rules, fmt.Sprintf("HostRegexp(`^%s$`)", host))
	}

	switch len(rules) {
	case 0:
		return "", 0
	case 1:
		return rules[0], priority
	default:
		return fmt.Sprintf("(%s)", strings.Join(rules, " || ")), priority
	}
}

// buildRouterRule builds the route rule and computes its priority.
// The current priority computing is rather naive but aims to fulfill Conformance tests suite requirement.
// The priority is computed to match the following precedence order:
//
// * "Exact" path match. (+100000)
// * "Prefix" path match with largest number of characters. (+10000) PathRegex (+1000)
// * Method match. (not implemented)
// * Largest number of header matches. (+100 each) or with PathRegex (+10 each)
// * Largest number of query param matches. (not implemented)
//
// In case of multiple matches for a route, the maximum priority among all matches is retain.
func buildRouterRule(hostnames []gatev1.Hostname, routeMatches []gatev1.HTTPRouteMatch) (string, int) {
	var matchesRules []string
	var maxPriority int

	for _, match := range routeMatches {
		path := ptr.Deref(match.Path, gatev1.HTTPPathMatch{
			Type:  ptr.To(gatev1.PathMatchPathPrefix),
			Value: ptr.To("/"),
		})

		var priority int
		var matchRules []string

		pathRule, pathPriority := buildPathRule(path)
		matchRules = append(matchRules, pathRule)
		priority += pathPriority

		headerRules, headersPriority := buildHeaderRules(match.Headers)
		matchRules = append(matchRules, headerRules...)
		priority += headersPriority

		matchesRules = append(matchesRules, strings.Join(matchRules, " && "))

		if priority > maxPriority {
			maxPriority = priority
		}
	}

	hostRule, hostPriority := buildHostRule(hostnames)

	matchesRulesStr := strings.Join(matchesRules, " || ")

	if hostRule == "" && matchesRulesStr == "" {
		return "PathPrefix(`/`)", 1
	}

	if hostRule != "" && matchesRulesStr == "" {
		return hostRule, hostPriority
	}

	// Enforce that, at the same priority,
	// the route with fewer matches (more specific) matches first.
	maxPriority -= len(matchesRules) * 10
	if maxPriority < 1 {
		maxPriority = 1
	}

	if hostRule == "" {
		return matchesRulesStr, maxPriority
	}

	// A route with a host should match over the same route with no host.
	maxPriority += hostPriority
	return hostRule + " && " + "(" + matchesRulesStr + ")", maxPriority
}

func buildPathRule(pathMatch gatev1.HTTPPathMatch) (string, int) {
	pathType := ptr.Deref(pathMatch.Type, gatev1.PathMatchPathPrefix)
	pathValue := ptr.Deref(pathMatch.Value, "/")

	switch pathType {
	case gatev1.PathMatchExact:
		return fmt.Sprintf("Path(`%s`)", pathValue), 100000

	case gatev1.PathMatchPathPrefix:
		// PathPrefix(`/`) rule is a catch-all,
		// here we ensure it would be evaluated last.
		if pathValue == "/" {
			return "PathPrefix(`/`)", 1
		}

		pv := strings.TrimSuffix(pathValue, "/")
		return fmt.Sprintf("(Path(`%[1]s`) || PathPrefix(`%[1]s/`))", pv), 10000 + len(pathValue)*100

	case gatev1.PathMatchRegularExpression:
		return fmt.Sprintf("PathRegexp(`%s`)", pathValue), 1000 + len(pathValue)*100

	default:
		return "PathPrefix(`/`)", 1
	}
}

func buildHeaderRules(headers []gatev1.HTTPHeaderMatch) ([]string, int) {
	var rules []string
	var priority int
	for _, header := range headers {
		typ := ptr.Deref(header.Type, gatev1.HeaderMatchExact)
		switch typ {
		case gatev1.HeaderMatchExact:
			rules = append(rules, fmt.Sprintf("Header(`%s`,`%s`)", header.Name, header.Value))
			priority += 100
		case gatev1.HeaderMatchRegularExpression:
			rules = append(rules, fmt.Sprintf("HeaderRegexp(`%s`,`%s`)", header.Name, header.Value))
			priority += 10
		}
	}

	return rules, priority
}

// createRequestHeaderModifier does not enforce/check the configuration,
// as the spec indicates that either the webhook or CEL (since v1.0 GA Release) should enforce that.
func createRequestHeaderModifier(filter *gatev1.HTTPHeaderFilter) *dynamic.Middleware {
	sets := map[string]string{}
	for _, header := range filter.Set {
		sets[string(header.Name)] = header.Value
	}

	adds := map[string]string{}
	for _, header := range filter.Add {
		adds[string(header.Name)] = header.Value
	}

	return &dynamic.Middleware{
		RequestHeaderModifier: &dynamic.RequestHeaderModifier{
			Set:    sets,
			Add:    adds,
			Remove: filter.Remove,
		},
	}
}

func createRedirectRegexMiddleware(listenerProtocol gatev1.ProtocolType, filter *gatev1.HTTPRequestRedirectFilter) *dynamic.Middleware {
	// The spec allows for an empty string in which case we should use the
	// scheme of the request which in this case is the listener scheme.
	filterScheme := ptr.Deref(filter.Scheme, strings.ToLower(string(listenerProtocol)))
	statusCode := ptr.Deref(filter.StatusCode, http.StatusFound)

	port := "${port}"
	if filter.Port != nil {
		port = fmt.Sprintf(":%d", *filter.Port)
	}

	hostname := "${hostname}"
	if filter.Hostname != nil && *filter.Hostname != "" {
		hostname = string(*filter.Hostname)
	}

	return &dynamic.Middleware{
		RedirectRegex: &dynamic.RedirectRegex{
			Regex:       `^[a-z]+:\/\/(?P<userInfo>.+@)?(?P<hostname>\[[\w:\.]+\]|[\w\._-]+)(?P<port>:\d+)?\/(?P<path>.*)`,
			Replacement: fmt.Sprintf("%s://${userinfo}%s%s/${path}", filterScheme, hostname, port),
			Permanent:   statusCode == http.StatusMovedPermanently,
		},
	}
}

func getProtocol(portSpec corev1.ServicePort) string {
	protocol := "http"
	if portSpec.Port == 443 || strings.HasPrefix(portSpec.Name, "https") {
		protocol = "https"
	}

	return protocol
}

func mergeHTTPConfiguration(from, to *dynamic.Configuration) {
	if from == nil || from.HTTP == nil || to == nil {
		return
	}

	if to.HTTP == nil {
		to.HTTP = from.HTTP
		return
	}

	if to.HTTP.Routers == nil {
		to.HTTP.Routers = map[string]*dynamic.Router{}
	}
	for routerName, router := range from.HTTP.Routers {
		to.HTTP.Routers[routerName] = router
	}

	if to.HTTP.Middlewares == nil {
		to.HTTP.Middlewares = map[string]*dynamic.Middleware{}
	}
	for middlewareName, middleware := range from.HTTP.Middlewares {
		to.HTTP.Middlewares[middlewareName] = middleware
	}

	if to.HTTP.Services == nil {
		to.HTTP.Services = map[string]*dynamic.Service{}
	}
	for serviceName, service := range from.HTTP.Services {
		to.HTTP.Services[serviceName] = service
	}
}
