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
						Status:             metav1.ConditionTrue,
						ObservedGeneration: route.Generation,
						LastTransitionTime: metav1.Now(),
						Reason:             string(gatev1.RouteReasonAccepted),
					},
				},
			}

			var attachedListeners bool
			notAcceptedReason := gatev1.RouteReasonNoMatchingParent
			for _, listener := range gatewayListeners {
				if !matchListener(listener, route.Namespace, parentRef) {
					continue
				}

				if !allowRoute(listener, route.Namespace, kindHTTPRoute) {
					notAcceptedReason = gatev1.RouteReasonNotAllowedByListeners
					continue
				}

				hostnames, ok := findMatchingHostnames(listener.Hostname, route.Spec.Hostnames)
				if !ok {
					notAcceptedReason = gatev1.RouteReasonNoMatchingListenerHostname
					continue
				}

				listener.Status.AttachedRoutes++

				// TODO should we build the conf if the listener is not attached
				// only consider the route attached if the listener is in an "attached" state.
				if listener.Attached {
					attachedListeners = true
				}
				resolveConditions := p.loadHTTPRoute(logger.WithContext(ctx), client, listener, route, hostnames, conf)

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
						Reason:             string(notAcceptedReason),
					},
					{
						Type:               string(gatev1.RouteConditionResolvedRefs),
						Status:             metav1.ConditionFalse,
						ObservedGeneration: route.Generation,
						LastTransitionTime: metav1.Now(),
						Reason:             string(gatev1.RouteReasonRefNotPermitted),
					},
				}
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

func (p *Provider) loadHTTPRoute(ctx context.Context, client Client, listener gatewayListener, route *gatev1.HTTPRoute, hostnames []gatev1.Hostname, conf *dynamic.Configuration) []metav1.Condition {
	routeConditions := []metav1.Condition{
		{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionTrue,
			ObservedGeneration: route.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteConditionResolvedRefs),
		},
	}

	hostRule := hostRule(hostnames)

	for _, routeRule := range route.Spec.Rules {
		router := dynamic.Router{
			RuleSyntax:  "v3",
			Rule:        routerRule(routeRule, hostRule),
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

			conf.HTTP.Services[wrrName] = &dynamic.Service{Weighted: &wrr}
			router.Service = wrrName
		} else {
			for name, middleware := range middlewares {
				// If the middleware config is nil in the return of the loadMiddlewares function,
				// it means that we just need a reference to that middleware.
				if middleware != nil {
					conf.HTTP.Middlewares[name] = middleware
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
						routeConditions = appendCondition(routeConditions, *errCondition)
						wrr.Services = append(wrr.Services, dynamic.WRRService{
							Name:   name,
							Status: ptr.To(500),
							Weight: weight,
						})
						continue
					}

					if svc != nil {
						conf.HTTP.Services[name] = svc
					}

					wrr.Services = append(wrr.Services, dynamic.WRRService{
						Name:   name,
						Weight: weight,
					})
				}

				conf.HTTP.Services[wrrName] = &dynamic.Service{Weighted: &wrr}
				router.Service = wrrName
			}
		}

		rt := &router
		p.applyRouterTransform(ctx, rt, route)

		routerKey = provider.Normalize(routerKey)
		conf.HTTP.Routers[routerKey] = rt
	}

	return routeConditions
}

// loadHTTPService returns a dynamic.Service config corresponding to the given gatev1.HTTPBackendRef.
// Note that the returned dynamic.Service config can be nil (for cross-provider, internal services, and backendFunc).
func (p *Provider) loadHTTPService(client Client, route *gatev1.HTTPRoute, backendRef gatev1.HTTPBackendRef) (string, *dynamic.Service, *metav1.Condition) {
	group := groupCore
	if backendRef.Group != nil && *backendRef.Group != "" {
		group = string(*backendRef.Group)
	}

	kind := ptr.Deref(backendRef.Kind, "Service")
	namespace := ptr.Deref(backendRef.Namespace, gatev1.Namespace(route.Namespace))
	namespaceStr := string(namespace)
	serviceName := provider.Normalize(makeID(namespaceStr, string(backendRef.Name)))

	// TODO support cross namespace through ReferenceGrant.
	if namespaceStr != route.Namespace {
		return serviceName, nil, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonRefNotPermitted),
			Message:            fmt.Sprintf("Cannot load HTTPBackendRef %s/%s/%s/%s namespace not allowed", group, kind, namespace, backendRef.Name),
		}
	}

	if group != groupCore || kind != "Service" {
		name, service, err := p.loadHTTPBackendRef(namespaceStr, backendRef)
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

	lb, err := loadHTTPServers(client, namespaceStr, backendRef)
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

func hostRule(hostnames []gatev1.Hostname) string {
	var rules []string

	for _, hostname := range hostnames {
		host := string(hostname)

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
		return ""
	case 1:
		return rules[0]
	default:
		return fmt.Sprintf("(%s)", strings.Join(rules, " || "))
	}
}

func routerRule(routeRule gatev1.HTTPRouteRule, hostRule string) string {
	var rule string
	var matchesRules []string

	for _, match := range routeRule.Matches {
		path := ptr.Deref(match.Path, gatev1.HTTPPathMatch{
			Type:  ptr.To(gatev1.PathMatchPathPrefix),
			Value: ptr.To("/"),
		})
		pathType := ptr.Deref(path.Type, gatev1.PathMatchPathPrefix)
		pathValue := ptr.Deref(path.Value, "/")

		var matchRules []string
		switch pathType {
		case gatev1.PathMatchExact:
			matchRules = append(matchRules, fmt.Sprintf("Path(`%s`)", pathValue))
		case gatev1.PathMatchPathPrefix:
			matchRules = append(matchRules, buildPathMatchPathPrefixRule(pathValue))
		case gatev1.PathMatchRegularExpression:
			matchRules = append(matchRules, fmt.Sprintf("PathRegexp(`%s`)", pathValue))
		}

		matchRules = append(matchRules, headerRules(match.Headers)...)
		matchesRules = append(matchesRules, strings.Join(matchRules, " && "))
	}

	// If no matches are specified, the default is a prefix
	// path match on "/", which has the effect of matching every
	// HTTP request.
	if len(routeRule.Matches) == 0 {
		matchesRules = append(matchesRules, "PathPrefix(`/`)")
	}

	if hostRule != "" {
		if len(matchesRules) == 0 {
			return hostRule
		}
		rule += hostRule + " && "
	}

	if len(matchesRules) == 1 {
		return rule + matchesRules[0]
	}

	if len(rule) == 0 {
		return strings.Join(matchesRules, " || ")
	}

	return rule + "(" + strings.Join(matchesRules, " || ") + ")"
}

func headerRules(headers []gatev1.HTTPHeaderMatch) []string {
	var headerRules []string
	for _, header := range headers {
		typ := ptr.Deref(header.Type, gatev1.HeaderMatchExact)
		switch typ {
		case gatev1.HeaderMatchExact:
			headerRules = append(headerRules, fmt.Sprintf("Header(`%s`,`%s`)", header.Name, header.Value))
		case gatev1.HeaderMatchRegularExpression:
			headerRules = append(headerRules, fmt.Sprintf("HeaderRegexp(`%s`,`%s`)", header.Name, header.Value))
		}
	}
	return headerRules
}

func buildPathMatchPathPrefixRule(path string) string {
	if path == "/" {
		return "PathPrefix(`/`)"
	}

	path = strings.TrimSuffix(path, "/")
	return fmt.Sprintf("(Path(`%[1]s`) || PathPrefix(`%[1]s/`))", path)
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
