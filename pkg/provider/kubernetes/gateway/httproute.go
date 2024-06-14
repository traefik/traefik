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

func (p *Provider) loadHTTPRoutes(ctx context.Context, gatewayListeners []gatewayListener, conf *dynamic.Configuration) {
	routes, err := p.client.ListHTTPRoutes()
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

				routeConf, resolveRefCondition := p.loadHTTPRoute(logger.WithContext(ctx), listener, route, hostnames)
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
		if err := p.client.UpdateHTTPRouteStatus(ctx, ktypes.NamespacedName{Namespace: route.Namespace, Name: route.Name}, status); err != nil {
			logger.Error().
				Err(err).
				Msg("Unable to update HTTPRoute status")
		}
	}
}

func (p *Provider) loadHTTPRoute(ctx context.Context, listener gatewayListener, route *gatev1.HTTPRoute, hostnames []gatev1.Hostname) (*dynamic.Configuration, metav1.Condition) {
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

	errWrr := dynamic.WeightedRoundRobin{
		Services: []dynamic.WRRService{
			{
				Name:   "invalid-httproute-filter",
				Status: ptr.To(500),
				Weight: ptr.To(1),
			},
		},
	}

	for ri, routeRule := range route.Spec.Rules {
		// Adding the gateway desc and the entryPoint desc prevents overlapping of routers build from the same routes.
		routeKey := provider.Normalize(fmt.Sprintf("%s-%s-%s-%s-%d", route.Namespace, route.Name, listener.GWName, listener.EPName, ri))

		for _, match := range routeRule.Matches {
			rule, priority := buildMatchRule(hostnames, match)
			router := dynamic.Router{
				RuleSyntax:  "v3",
				Rule:        rule,
				Priority:    priority + len(route.Spec.Rules) - ri,
				EntryPoints: []string{listener.EPName},
			}
			if listener.Protocol == gatev1.HTTPSProtocolType {
				router.TLS = &dynamic.RouterTLSConfig{}
			}

			var err error
			routerName := makeRouterName(rule, routeKey)
			router.Middlewares, err = p.loadMiddlewares(conf, route.Namespace, routerName, routeRule.Filters, match.Path)
			switch {
			case err != nil:
				log.Ctx(ctx).Error().Err(err).Msg("Unable to load HTTPRoute filters")

				errWrrName := routerName + "-err-wrr"
				conf.HTTP.Services[errWrrName] = &dynamic.Service{Weighted: &errWrr}
				router.Service = errWrrName

			case len(routeRule.BackendRefs) == 1 && isInternalService(routeRule.BackendRefs[0].BackendRef):
				router.Service = string(routeRule.BackendRefs[0].Name)

			default:
				var serviceCondition *metav1.Condition
				router.Service, serviceCondition = p.loadService(conf, routeKey, routeRule, route)
				if serviceCondition != nil {
					condition = *serviceCondition
				}
			}

			p.applyRouterTransform(ctx, &router, route)

			conf.HTTP.Routers[routerName] = &router
		}
	}

	return conf, condition
}

func (p *Provider) loadService(conf *dynamic.Configuration, routeKey string, routeRule gatev1.HTTPRouteRule, route *gatev1.HTTPRoute) (string, *metav1.Condition) {
	name := routeKey + "-wrr"
	if _, ok := conf.HTTP.Services[name]; ok {
		return name, nil
	}

	var wrr dynamic.WeightedRoundRobin
	var condition *metav1.Condition
	for _, backendRef := range routeRule.BackendRefs {
		svcName, svc, errCondition := p.loadHTTPService(route, backendRef)
		weight := ptr.To(int(ptr.Deref(backendRef.Weight, 1)))
		if errCondition != nil {
			condition = errCondition
			wrr.Services = append(wrr.Services, dynamic.WRRService{
				Name:   svcName,
				Status: ptr.To(500),
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

// loadHTTPService returns a dynamic.Service config corresponding to the given gatev1.HTTPBackendRef.
// Note that the returned dynamic.Service config can be nil (for cross-provider, internal services, and backendFunc).
func (p *Provider) loadHTTPService(route *gatev1.HTTPRoute, backendRef gatev1.HTTPBackendRef) (string, *dynamic.Service, *metav1.Condition) {
	kind := ptr.Deref(backendRef.Kind, "Service")

	group := groupCore
	if backendRef.Group != nil && *backendRef.Group != "" {
		group = string(*backendRef.Group)
	}

	namespace := route.Namespace
	if backendRef.Namespace != nil && *backendRef.Namespace != "" {
		namespace = string(*backendRef.Namespace)
	}

	serviceName := provider.Normalize(namespace + "-" + string(backendRef.Name))

	if err := p.isReferenceGranted(groupGateway, kindHTTPRoute, route.Namespace, group, string(kind), string(backendRef.Name), namespace); err != nil {
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

	lb, err := p.loadHTTPServers(namespace, backendRef)
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

func (p *Provider) loadMiddlewares(conf *dynamic.Configuration, namespace, routerName string, filters []gatev1.HTTPRouteFilter, pathMatch *gatev1.HTTPPathMatch) ([]string, error) {
	pm := ptr.Deref(pathMatch, gatev1.HTTPPathMatch{
		Type:  ptr.To(gatev1.PathMatchPathPrefix),
		Value: ptr.To("/"),
	})

	middlewares := make(map[string]*dynamic.Middleware)
	for i, filter := range filters {
		name := fmt.Sprintf("%s-%s-%d", routerName, strings.ToLower(string(filter.Type)), i)
		switch filter.Type {
		case gatev1.HTTPRouteFilterRequestRedirect:
			middlewares[name] = createRequestRedirect(filter.RequestRedirect, pm)

		case gatev1.HTTPRouteFilterRequestHeaderModifier:
			middlewares[name] = createRequestHeaderModifier(filter.RequestHeaderModifier)

		case gatev1.HTTPRouteFilterExtensionRef:
			name, middleware, err := p.loadHTTPRouteFilterExtensionRef(namespace, filter.ExtensionRef)
			if err != nil {
				return nil, fmt.Errorf("loading ExtensionRef filter %s: %w", filter.Type, err)
			}

			middlewares[name] = middleware

		case gatev1.HTTPRouteFilterURLRewrite:
			var err error
			middleware, err := createURLRewrite(filter.URLRewrite, pm)
			if err != nil {
				return nil, fmt.Errorf("invalid filter %s: %w", filter.Type, err)
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

	var middlewareNames []string
	for name, middleware := range middlewares {
		if middleware != nil {
			conf.HTTP.Middlewares[name] = middleware
		}

		middlewareNames = append(middlewareNames, name)
	}

	return middlewareNames, nil
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

func (p *Provider) loadHTTPServers(namespace string, backendRef gatev1.HTTPBackendRef) (*dynamic.ServersLoadBalancer, error) {
	service, exists, err := p.client.GetService(namespace, string(backendRef.Name))
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

	endpoints, endpointsExists, err := p.client.GetEndpoints(namespace, string(backendRef.Name))
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

// buildMatchRule builds the route rule and computes its priority.
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
func buildMatchRule(hostnames []gatev1.Hostname, match gatev1.HTTPRouteMatch) (string, int) {
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

	matchRulesStr := strings.Join(matchRules, " && ")

	hostRule, hostPriority := buildHostRule(hostnames)

	if hostRule == "" {
		return matchRulesStr, priority
	}

	// A route with a host should match over the same route with no host.
	priority += hostPriority
	return hostRule + " && " + matchRulesStr, priority
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

func createRequestRedirect(filter *gatev1.HTTPRequestRedirectFilter, pathMatch gatev1.HTTPPathMatch) *dynamic.Middleware {
	var hostname *string
	if filter.Hostname != nil {
		hostname = ptr.To(string(*filter.Hostname))
	}

	var port *string
	filterScheme := ptr.Deref(filter.Scheme, "")
	if filterScheme == "http" || filterScheme == "https" {
		port = ptr.To("")
	}
	if filter.Port != nil {
		port = ptr.To(fmt.Sprintf("%d", *filter.Port))
	}

	var path *string
	var pathPrefix *string
	if filter.Path != nil {
		switch filter.Path.Type {
		case gatev1.FullPathHTTPPathModifier:
			path = filter.Path.ReplaceFullPath
		case gatev1.PrefixMatchHTTPPathModifier:
			path = filter.Path.ReplacePrefixMatch
			pathPrefix = pathMatch.Value
		}
	}

	return &dynamic.Middleware{
		RequestRedirect: &dynamic.RequestRedirect{
			Scheme:     filter.Scheme,
			Hostname:   hostname,
			Port:       port,
			Path:       path,
			PathPrefix: pathPrefix,
			StatusCode: ptr.Deref(filter.StatusCode, http.StatusFound),
		},
	}
}

func createURLRewrite(filter *gatev1.HTTPURLRewriteFilter, pathMatch gatev1.HTTPPathMatch) (*dynamic.Middleware, error) {
	if filter.Path == nil && filter.Hostname == nil {
		return nil, errors.New("empty configuration")
	}

	var host *string
	if filter.Hostname != nil {
		host = ptr.To(string(*filter.Hostname))
	}

	var path *string
	var pathPrefix *string
	if filter.Path != nil {
		switch filter.Path.Type {
		case gatev1.FullPathHTTPPathModifier:
			path = filter.Path.ReplaceFullPath
		case gatev1.PrefixMatchHTTPPathModifier:
			path = filter.Path.ReplacePrefixMatch
			pathPrefix = pathMatch.Value
		}
	}

	return &dynamic.Middleware{
		URLRewrite: &dynamic.URLRewrite{
			Hostname:   host,
			Path:       path,
			PathPrefix: pathPrefix,
		},
	}, nil
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
