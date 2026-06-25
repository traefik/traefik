package gateway

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/provider"
	"github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	gatev1 "sigs.k8s.io/gateway-api/apis/v1"
)

func (p *Provider) loadTLSRoutes(ctx context.Context, gatewayListeners []gatewayListener, conf *dynamic.Configuration, statusReport *statusReport) {
	logger := log.Ctx(ctx)
	routes, err := p.client.ListTLSRoutes()
	if err != nil {
		logger.Error().Err(err).Msgf("Unable to list TLSRoute")
		return
	}

	for _, route := range routes {
		routeListeners := matchingGatewayListeners(gatewayListeners, route.Namespace, route.Spec.ParentRefs)
		if len(routeListeners) == 0 {
			continue
		}

		for _, parentRef := range route.Spec.ParentRefs {
			parentStatus := gatev1.RouteParentStatus{
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

				routeConf, resolveRefCondition := p.loadTLSRoute(listener, route, hostnames, statusReport)
				if accepted && listener.Attached {
					mergeTCPConfiguration(routeConf, conf)
				}
				parentStatus.Conditions = upsertRouteConditionResolvedRefs(parentStatus.Conditions, resolveRefCondition)
			}

			statusReport.RecordTLSRouteStatus(ktypes.NamespacedName{Namespace: route.Namespace, Name: route.Name}, parentStatus)
		}

		// When there is at least one TLS listener, we add a default deny-all route to avoid accepting traffic for undefined hosts.
		// Note that when there is HTTPS listeners this will predate the traffic and reject the connection to undefined hosts instead of returning a 404.
		if len(conf.TCP.Routers) > 0 {
			conf.TCP.Routers["deny-unknown-host"] = &dynamic.TCPRouter{
				Rule:     "HostSNI(`*`) && !ALPN(`h2`) && !ALPN(`http/1.1`)",
				Priority: 1,
				Service:  "deny-unknown-host",
				TLS:      &dynamic.RouterTCPTLSConfig{},
			}
			conf.TCP.Services["deny-unknown-host"] = &dynamic.TCPService{
				LoadBalancer: &dynamic.TCPServersLoadBalancer{},
			}
		}
	}
}

func (p *Provider) loadTLSRoute(listener gatewayListener, route *gatev1.TLSRoute, hostnames []gatev1.Hostname, statusReport *statusReport) (*dynamic.Configuration, metav1.Condition) {
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
			if !isCrossProviderNamespaceAllowed(p.CrossProviderNamespaces, route.Namespace) {
				condition = metav1.Condition{
					Type:               string(gatev1.RouteConditionResolvedRefs),
					Status:             metav1.ConditionFalse,
					ObservedGeneration: route.Generation,
					LastTransitionTime: metav1.Now(),
					Reason:             string(gatev1.RouteReasonRefNotPermitted),
					Message:            fmt.Sprintf("Cannot load TLSRoute BackendRef %s: internal service reference is not allowed: TLSRoute namespace %q is not in crossProviderNamespaces", routeRule.BackendRefs[0].Name, route.Namespace),
				}

				continue
			}

			router.Service = string(routeRule.BackendRefs[0].Name)
			conf.TCP.Routers[routerName] = &router
			continue
		}

		var serviceCondition *metav1.Condition
		router.Service, serviceCondition = p.loadTLSWRRService(listener, conf, routerName, routeRule.BackendRefs, route, statusReport)
		if serviceCondition != nil {
			condition = *serviceCondition
		}

		conf.TCP.Routers[routerName] = &router
	}

	return conf, condition
}

// loadTLSWRRService is generating a WRR service, even when there is only one target.
func (p *Provider) loadTLSWRRService(listener gatewayListener, conf *dynamic.Configuration, routeKey string, backendRefs []gatev1.BackendRef, route *gatev1.TLSRoute, statusReport *statusReport) (string, *metav1.Condition) {
	name := routeKey + "-wrr"
	if _, ok := conf.TCP.Services[name]; ok {
		return name, nil
	}

	var wrr dynamic.TCPWeightedRoundRobin
	var condition *metav1.Condition
	for _, backendRef := range backendRefs {
		svcName, svc, errCondition := p.loadTLSService(listener, conf, route, backendRef, statusReport)
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

func (p *Provider) loadTLSService(listener gatewayListener, conf *dynamic.Configuration, route *gatev1.TLSRoute, backendRef gatev1.BackendRef, statusReport *statusReport) (string, *dynamic.TCPService, *metav1.Condition) {
	kind := ptr.Deref(backendRef.Kind, kindService)

	group := groupCore
	if backendRef.Group != nil && *backendRef.Group != "" {
		group = string(*backendRef.Group)
	}

	namespace := route.Namespace
	if backendRef.Namespace != nil && *backendRef.Namespace != "" {
		namespace = string(*backendRef.Namespace)

		if strings.Contains(string(backendRef.Name), "@") {
			return provider.Normalize(namespace + "-" + string(backendRef.Name)), nil, &metav1.Condition{
				Type:               string(gatev1.RouteConditionResolvedRefs),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: route.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.RouteReasonRefNotPermitted),
				Message:            fmt.Sprintf("Cannot load TLSRoute BackendRef %s/%s/%s/%s: namespace is not allowed with a cross-provider reference", group, kind, namespace, backendRef.Name),
			}
		}
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
		name, err := p.loadTCPBackendRef(route.Namespace, backendRef)
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

	lb, st, errCondition := p.loadTLSServers(namespace, route, backendRef, listener, statusReport)
	if errCondition != nil {
		return serviceName, nil, errCondition
	}

	if st != nil {
		lb.ServersTransport = serviceName
		conf.TCP.ServersTransports[serviceName] = st
	}

	return serviceName, &dynamic.TCPService{LoadBalancer: lb}, nil
}

func (p *Provider) loadTLSServers(namespace string, route *gatev1.TLSRoute, backendRef gatev1.BackendRef, listener gatewayListener, statusReport *statusReport) (*dynamic.TCPServersLoadBalancer, *dynamic.TCPServersTransport, *metav1.Condition) {
	backendAddresses, svcPort, err := p.getBackendAddresses(namespace, backendRef)
	if err != nil {
		return nil, nil, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.GetGeneration(),
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonBackendNotFound),
			Message:            fmt.Sprintf("Cannot load TLSRoute BackendRef %s/%s: %s", namespace, backendRef.Name, err),
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

	var serversTransport *dynamic.TCPServersTransport
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
					Group:       ptr.To(gatev1.Group(groupGateway)),
					Kind:        ptr.To(gatev1.Kind(kindGateway)),
					Namespace:   ptr.To(gatev1.Namespace(namespace)),
					Name:        gatev1.ObjectName(listener.GWName),
					SectionName: ptr.To(gatev1.SectionName(listener.Name)),
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
			serversTransport, resolvedRefCondition = p.loadTCPServersTransport(namespace, policy)

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

	if svcPort.Protocol != corev1.ProtocolTCP {
		return nil, nil, &metav1.Condition{
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
	return lb, serversTransport, nil
}

func (p *Provider) loadTCPServersTransport(namespace string, policy *gatev1.BackendTLSPolicy) (*dynamic.TCPServersTransport, metav1.Condition) {
	st := &dynamic.TCPServersTransport{
		TLS: &dynamic.TLSClientConfig{
			ServerName: string(policy.Spec.Validation.Hostname),
		},
	}

	if len(policy.Spec.Validation.SubjectAltNames) > 0 {
		// Per the Gateway API specification the Hostname should only be used for authentication
		// and not for certificate validation. Thus, if SubjectAltNames is specified, we ignore
		// the Hostname validation by setting the InsecureSkipVerify option to true.
		st.TLS.InsecureSkipVerify = true

		for _, san := range policy.Spec.Validation.SubjectAltNames {
			switch san.Type {
			case gatev1.URISubjectAltNameType:
				st.TLS.PeerCertSANs = append(st.TLS.PeerCertSANs, tls.SAN{
					Type:  tls.SANURIType,
					Value: string(san.URI),
				})
			case gatev1.HostnameSubjectAltNameType:
				st.TLS.PeerCertSANs = append(st.TLS.PeerCertSANs, tls.SAN{
					Type:  tls.SANDNSNameType,
					Value: string(san.Hostname),
				})
			default:
				return nil, metav1.Condition{
					Type:               string(gatev1.BackendTLSPolicyConditionResolvedRefs),
					Status:             metav1.ConditionFalse,
					ObservedGeneration: policy.Generation,
					LastTransitionTime: metav1.Now(),
					Reason:             string(gatev1.BackendTLSPolicyReasonInvalidKind),
					Message:            fmt.Sprintf("Unsupported SubjectAltName type %q; only URI and Hostname types are supported", san.Type),
				}
			}
		}
	}

	if policy.Spec.Validation.WellKnownCACertificates != nil {
		return st, metav1.Condition{
			Type:               string(gatev1.BackendTLSPolicyConditionResolvedRefs),
			Status:             metav1.ConditionTrue,
			ObservedGeneration: policy.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.BackendTLSPolicyReasonResolvedRefs),
		}
	}

	for _, caCertRef := range policy.Spec.Validation.CACertificateRefs {
		if (caCertRef.Group != "" && caCertRef.Group != groupCore) || (caCertRef.Kind != kindConfigMap && caCertRef.Kind != kindSecret) {
			return nil, metav1.Condition{
				Type:               string(gatev1.BackendTLSPolicyConditionResolvedRefs),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: policy.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.BackendTLSPolicyReasonInvalidKind),
				Message:            "Only ConfigMaps and Secrets are supported",
			}
		}

		var caCRT string
		switch caCertRef.Kind {
		case kindConfigMap:
			configmap, err := p.client.GetConfigMap(namespace, string(caCertRef.Name))
			if err != nil {
				return nil, metav1.Condition{
					Type:               string(gatev1.BackendTLSPolicyConditionResolvedRefs),
					Status:             metav1.ConditionFalse,
					ObservedGeneration: policy.Generation,
					LastTransitionTime: metav1.Now(),
					Reason:             string(gatev1.BackendTLSPolicyReasonInvalidCACertificateRef),
					Message:            fmt.Sprintf("getting configmap %s/%s: %s", namespace, string(caCertRef.Name), err),
				}
			}
			caCRT = configmap.Data["ca.crt"]
		case kindSecret:
			secret, err := p.client.GetSecret(namespace, string(caCertRef.Name))
			if err != nil {
				return nil, metav1.Condition{
					Type:               string(gatev1.BackendTLSPolicyConditionResolvedRefs),
					Status:             metav1.ConditionFalse,
					ObservedGeneration: policy.Generation,
					LastTransitionTime: metav1.Now(),
					Reason:             string(gatev1.BackendTLSPolicyReasonInvalidCACertificateRef),
					Message:            fmt.Sprintf("getting secret %s/%s: %s", namespace, string(caCertRef.Name), err),
				}
			}
			caCRT = string(secret.Data["ca.crt"])
		}

		if caCRT == "" {
			return nil, metav1.Condition{
				Type:               string(gatev1.BackendTLSPolicyConditionResolvedRefs),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: policy.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.BackendTLSPolicyReasonInvalidCACertificateRef),
				Message:            fmt.Sprintf("%s %s/%s does not have a ca.crt", caCertRef.Kind, namespace, string(caCertRef.Name)),
			}
		}

		st.TLS.RootCAs = append(st.TLS.RootCAs, types.FileOrContent(caCRT))
	}

	return st, metav1.Condition{
		Type:               string(gatev1.BackendTLSPolicyConditionResolvedRefs),
		Status:             metav1.ConditionTrue,
		ObservedGeneration: policy.Generation,
		LastTransitionTime: metav1.Now(),
		Reason:             string(gatev1.BackendTLSPolicyReasonResolvedRefs),
	}
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
			rules = append(rules, fmt.Sprintf("HostSNI(%q)", host))
			continue
		}

		host = strings.Replace(regexp.QuoteMeta(host), `\*\.`, `[a-z0-9-]+\.`, 1)
		rules = append(rules, fmt.Sprintf("HostSNIRegexp(%q)", fmt.Sprintf("^%s$", host)))
	}

	if len(hostnames) == 0 || len(rules) == 0 {
		return `HostSNI("*")`, 0
	}

	return strings.Join(rules, " || "), priority
}
