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
	"github.com/traefik/traefik/v3/pkg/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	gatev1 "sigs.k8s.io/gateway-api/apis/v1"
	gatev1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatev1alpha3 "sigs.k8s.io/gateway-api/apis/v1alpha3"
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
			logger.Warn().
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
				conf.HTTP.Services[errWrrName] = &dynamic.Service{
					Weighted: &dynamic.WeightedRoundRobin{
						Services: []dynamic.WRRService{
							{
								Name:   "invalid-httproute-filter",
								Status: ptr.To(500),
								Weight: ptr.To(1),
							},
						},
					},
				}
				router.Service = errWrrName

			case len(routeRule.BackendRefs) == 1 && isInternalService(routeRule.BackendRefs[0].BackendRef):
				router.Service = string(routeRule.BackendRefs[0].Name)

			default:
				var serviceCondition *metav1.Condition
				router.Service, serviceCondition = p.loadWRRService(ctx, listener, conf, routerName, routeRule, route)
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

func (p *Provider) loadWRRService(ctx context.Context, listener gatewayListener, conf *dynamic.Configuration, routeKey string, routeRule gatev1.HTTPRouteRule, route *gatev1.HTTPRoute) (string, *metav1.Condition) {
	name := routeKey + "-wrr"
	if _, ok := conf.HTTP.Services[name]; ok {
		return name, nil
	}

	var wrr dynamic.WeightedRoundRobin
	var condition *metav1.Condition
	for _, backendRef := range routeRule.BackendRefs {
		svcName, errCondition := p.loadService(ctx, listener, conf, route, backendRef)
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

		wrr.Services = append(wrr.Services, dynamic.WRRService{
			Name:   svcName,
			Weight: weight,
		})
	}

	conf.HTTP.Services[name] = &dynamic.Service{Weighted: &wrr}
	return name, condition
}

// loadService returns a dynamic.Service config corresponding to the given gatev1.HTTPBackendRef.
// Note that the returned dynamic.Service config can be nil (for cross-provider, internal services, and backendFunc).
func (p *Provider) loadService(ctx context.Context, listener gatewayListener, conf *dynamic.Configuration, route *gatev1.HTTPRoute, backendRef gatev1.HTTPBackendRef) (string, *metav1.Condition) {
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

	if err := p.isReferenceGranted(kindHTTPRoute, route.Namespace, group, string(kind), string(backendRef.Name), namespace); err != nil {
		return serviceName, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonRefNotPermitted),
			Message:            fmt.Sprintf("Cannot load HTTPBackendRef %s/%s/%s/%s: %s", group, kind, namespace, backendRef.Name, err),
		}
	}

	if group != groupCore || kind != kindService {
		name, service, err := p.loadHTTPBackendRef(namespace, backendRef)
		if err != nil {
			return serviceName, &metav1.Condition{
				Type:               string(gatev1.RouteConditionResolvedRefs),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: route.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.RouteReasonInvalidKind),
				Message:            fmt.Sprintf("Cannot load HTTPBackendRef %s/%s/%s/%s: %s", group, kind, namespace, backendRef.Name, err),
			}
		}

		if service != nil {
			conf.HTTP.Services[name] = service
		}

		return name, nil
	}

	port := ptr.Deref(backendRef.Port, gatev1.PortNumber(0))
	if port == 0 {
		return serviceName, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonUnsupportedProtocol),
			Message:            fmt.Sprintf("Cannot load HTTPBackendRef %s/%s/%s/%s: port is required", group, kind, namespace, backendRef.Name),
		}
	}

	portStr := strconv.FormatInt(int64(port), 10)
	serviceName = provider.Normalize(serviceName + "-" + portStr)

	lb, svcPort, errCondition := p.loadHTTPServers(namespace, route, backendRef)
	if errCondition != nil {
		return serviceName, errCondition
	}

	if !p.ExperimentalChannel {
		conf.HTTP.Services[serviceName] = &dynamic.Service{LoadBalancer: lb}

		return serviceName, nil
	}

	servicePolicies, err := p.client.ListBackendTLSPoliciesForService(namespace, string(backendRef.Name))
	if err != nil {
		return serviceName, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonRefNotPermitted),
			Message:            fmt.Sprintf("Cannot list BackendTLSPolicies for Service %s/%s: %s", namespace, string(backendRef.Name), err),
		}
	}

	var matchedPolicy *gatev1alpha3.BackendTLSPolicy
	for _, policy := range servicePolicies {
		matched := false
		for _, targetRef := range policy.Spec.TargetRefs {
			if targetRef.SectionName == nil || svcPort.Name == string(*targetRef.SectionName) {
				matchedPolicy = policy
				matched = true
				break
			}
		}

		// If the policy targets the service, but doesn't match any port.
		if !matched {
			// update policy status
			status := gatev1alpha2.PolicyStatus{
				Ancestors: []gatev1alpha2.PolicyAncestorStatus{{
					AncestorRef: gatev1alpha2.ParentReference{
						Group:       ptr.To(gatev1.Group(groupGateway)),
						Kind:        ptr.To(gatev1.Kind(kindGateway)),
						Namespace:   ptr.To(gatev1.Namespace(namespace)),
						Name:        gatev1.ObjectName(listener.GWName),
						SectionName: ptr.To(gatev1.SectionName(listener.Name)),
					},
					ControllerName: controllerName,
					Conditions: []metav1.Condition{{
						Type:               string(gatev1.RouteConditionResolvedRefs),
						Status:             metav1.ConditionFalse,
						ObservedGeneration: route.Generation,
						LastTransitionTime: metav1.Now(),
						Reason:             string(gatev1.RouteReasonBackendNotFound),
						Message:            fmt.Sprintf("BackendTLSPolicy has no valid TargetRef for Service %s/%s", namespace, string(backendRef.Name)),
					}},
				}},
			}

			if err := p.client.UpdateBackendTLSPolicyStatus(ctx, ktypes.NamespacedName{Namespace: policy.Namespace, Name: policy.Name}, status); err != nil {
				logger := log.Ctx(ctx).With().
					Str("http_route", route.Name).
					Str("namespace", route.Namespace).Logger()
				logger.Warn().
					Err(err).
					Msg("Unable to update TLSRoute status")
			}
		}
	}

	if matchedPolicy != nil {
		st, err := p.loadServersTransport(namespace, *matchedPolicy)
		if err != nil {
			return serviceName, &metav1.Condition{
				Type:               string(gatev1.RouteConditionResolvedRefs),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: route.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.RouteReasonRefNotPermitted),
				Message:            fmt.Sprintf("Cannot apply BackendTLSPolicy for Service %s/%s: %s", namespace, string(backendRef.Name), err),
			}
		}

		if st != nil {
			lb.ServersTransport = serviceName
			conf.HTTP.ServersTransports[serviceName] = st
		}
	}

	conf.HTTP.Services[serviceName] = &dynamic.Service{LoadBalancer: lb}

	return serviceName, nil
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

		case gatev1.HTTPRouteFilterResponseHeaderModifier:
			middlewares[name] = createResponseHeaderModifier(filter.ResponseHeaderModifier)

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

func (p *Provider) loadHTTPServers(namespace string, route *gatev1.HTTPRoute, backendRef gatev1.HTTPBackendRef) (*dynamic.ServersLoadBalancer, corev1.ServicePort, *metav1.Condition) {
	backendAddresses, svcPort, err := p.getBackendAddresses(namespace, backendRef.BackendRef)
	if err != nil {
		return nil, corev1.ServicePort{}, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonBackendNotFound),
			Message:            fmt.Sprintf("Cannot load HTTPBackendRef %s/%s: %s", namespace, backendRef.Name, err),
		}
	}

	protocol, err := getProtocol(svcPort)
	if err != nil {
		return nil, corev1.ServicePort{}, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonUnsupportedProtocol),
			Message:            fmt.Sprintf("Cannot load HTTPBackendRef %s/%s: %s", namespace, backendRef.Name, err),
		}
	}

	lb := &dynamic.ServersLoadBalancer{}
	lb.SetDefaults()

	for _, ba := range backendAddresses {
		lb.Servers = append(lb.Servers, dynamic.Server{
			URL: fmt.Sprintf("%s://%s", protocol, net.JoinHostPort(ba.Address, strconv.Itoa(int(ba.Port)))),
		})
	}
	return lb, svcPort, nil
}

func (p *Provider) loadServersTransport(namespace string, policy gatev1alpha3.BackendTLSPolicy) (*dynamic.ServersTransport, error) {
	st := &dynamic.ServersTransport{
		ServerName: string(policy.Spec.Validation.Hostname),
	}

	if policy.Spec.Validation.WellKnownCACertificates != nil {
		return st, nil
	}

	for _, caCertRef := range policy.Spec.Validation.CACertificateRefs {
		if caCertRef.Group != groupCore || caCertRef.Kind != "ConfigMap" {
			continue
		}

		configMap, exists, err := p.client.GetConfigMap(namespace, string(caCertRef.Name))
		if err != nil {
			return nil, fmt.Errorf("getting configmap: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("configmap %s/%s not found", namespace, string(caCertRef.Name))
		}

		caCRT, ok := configMap.Data["ca.crt"]
		if !ok {
			return nil, fmt.Errorf("configmap %s/%s does not have ca.crt", namespace, string(caCertRef.Name))
		}

		st.RootCAs = append(st.RootCAs, types.FileOrContent(caCRT))
	}

	return st, nil
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
// * "Exact" path match (+100000).
// * "Prefix" path match with largest number of characters (+10000 + nb_characters*100).
// * Method match (+1000).
// * Largest number of header matches (+100 each).
// * Largest number of query param matches (+10 each).
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

	if match.Method != nil {
		matchRules = append(matchRules, fmt.Sprintf("Method(`%s`)", *match.Method))
		priority += 1000
	}

	headerRules, headersPriority := buildHeaderRules(match.Headers)
	matchRules = append(matchRules, headerRules...)
	priority += headersPriority

	queryParamRules, queryParamsPriority := buildQueryParamRules(match.QueryParams)
	matchRules = append(matchRules, queryParamRules...)
	priority += queryParamsPriority

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
		return fmt.Sprintf("PathRegexp(`%s`)", pathValue), 10000 + len(pathValue)*100

	default:
		return "PathPrefix(`/`)", 1
	}
}

func buildHeaderRules(headers []gatev1.HTTPHeaderMatch) ([]string, int) {
	var (
		rules    []string
		priority int
	)
	for _, header := range headers {
		typ := ptr.Deref(header.Type, gatev1.HeaderMatchExact)
		switch typ {
		case gatev1.HeaderMatchExact:
			rules = append(rules, fmt.Sprintf("Header(`%s`,`%s`)", header.Name, header.Value))
		case gatev1.HeaderMatchRegularExpression:
			rules = append(rules, fmt.Sprintf("HeaderRegexp(`%s`,`%s`)", header.Name, header.Value))
		}
		priority += 100
	}

	return rules, priority
}

func buildQueryParamRules(queryParams []gatev1.HTTPQueryParamMatch) ([]string, int) {
	var (
		rules    []string
		priority int
	)
	for _, qp := range queryParams {
		typ := ptr.Deref(qp.Type, gatev1.QueryParamMatchExact)
		switch typ {
		case gatev1.QueryParamMatchExact:
			rules = append(rules, fmt.Sprintf("Query(`%s`,`%s`)", qp.Name, qp.Value))
		case gatev1.QueryParamMatchRegularExpression:
			rules = append(rules, fmt.Sprintf("QueryRegexp(`%s`,`%s`)", qp.Name, qp.Value))
		}
		priority += 10
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
		RequestHeaderModifier: &dynamic.HeaderModifier{
			Set:    sets,
			Add:    adds,
			Remove: filter.Remove,
		},
	}
}

// createResponseHeaderModifier does not enforce/check the configuration,
// as the spec indicates that either the webhook or CEL (since v1.0 GA Release) should enforce that.
func createResponseHeaderModifier(filter *gatev1.HTTPHeaderFilter) *dynamic.Middleware {
	sets := map[string]string{}
	for _, header := range filter.Set {
		sets[string(header.Name)] = header.Value
	}

	adds := map[string]string{}
	for _, header := range filter.Add {
		adds[string(header.Name)] = header.Value
	}

	return &dynamic.Middleware{
		ResponseHeaderModifier: &dynamic.HeaderModifier{
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

func getProtocol(portSpec corev1.ServicePort) (string, error) {
	if portSpec.Protocol != corev1.ProtocolTCP {
		return "", errors.New("only TCP protocol is supported")
	}

	if portSpec.AppProtocol == nil {
		protocol := "http"
		if portSpec.Port == 443 || strings.HasPrefix(portSpec.Name, "https") {
			protocol = "https"
		}
		return protocol, nil
	}

	switch ap := *portSpec.AppProtocol; ap {
	case appProtocolH2C:
		return "h2c", nil
	case appProtocolWS:
		return "http", nil
	case appProtocolWSS:
		return "https", nil
	default:
		return "", fmt.Errorf("unsupported application protocol %s", ap)
	}
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

	if to.HTTP.ServersTransports == nil {
		to.HTTP.ServersTransports = map[string]*dynamic.ServersTransport{}
	}
	for name, serversTransport := range from.HTTP.ServersTransports {
		to.HTTP.ServersTransports[name] = serversTransport
	}
}
