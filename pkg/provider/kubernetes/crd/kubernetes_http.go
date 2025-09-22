package crd

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/provider"
	traefikv1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/k8s"
	"github.com/traefik/traefik/v3/pkg/tls"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

const (
	httpsProtocol = "https"
	httpProtocol  = "http"
)

func (p *Provider) loadIngressRouteConfiguration(ctx context.Context, client Client, tlsConfigs map[string]*tls.CertAndStores) *dynamic.HTTPConfiguration {
	conf := &dynamic.HTTPConfiguration{
		Routers:           map[string]*dynamic.Router{},
		Middlewares:       map[string]*dynamic.Middleware{},
		Services:          map[string]*dynamic.Service{},
		ServersTransports: map[string]*dynamic.ServersTransport{},
	}

	for _, ingressRoute := range client.GetIngressRoutes() {
		logger := log.Ctx(ctx).With().Str("ingress", ingressRoute.Name).Str("namespace", ingressRoute.Namespace).Logger()

		// TODO keep the name ingressClass?
		if !shouldProcessIngress(p.IngressClass, ingressRoute.Annotations[annotationKubernetesIngressClass]) {
			continue
		}

		err := getTLSHTTP(ctx, ingressRoute, client, tlsConfigs)
		if err != nil {
			logger.Error().Err(err).Msg("Error configuring TLS")
		}

		ingressName := ingressRoute.Name
		if len(ingressName) == 0 {
			ingressName = ingressRoute.GenerateName
		}

		cb := configBuilder{
			client:                       client,
			allowCrossNamespace:          p.AllowCrossNamespace,
			allowExternalNameServices:    p.AllowExternalNameServices,
			allowEmptyServices:           p.AllowEmptyServices,
			nativeLBByDefault:            p.NativeLBByDefault,
			disableClusterScopeResources: p.DisableClusterScopeResources,
		}

		for _, route := range ingressRoute.Spec.Routes {
			if len(route.Kind) > 0 && route.Kind != "Rule" {
				logger.Error().Msgf("Unsupported match kind: %s. Only \"Rule\" is supported for now.", route.Kind)
				continue
			}

			if len(route.Match) == 0 {
				logger.Error().Msg("Empty match rule")
				continue
			}

			serviceKey, err := makeServiceKey(route.Match, ingressName)
			if err != nil {
				logger.Error().Err(err).Send()
				continue
			}

			mds, err := p.makeMiddlewareKeys(ctx, ingressRoute.Namespace, route.Middlewares)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to create middleware keys")
				continue
			}

			normalized := provider.Normalize(makeID(ingressRoute.Namespace, serviceKey))
			serviceName := normalized

			if len(route.Services) > 1 {
				spec := traefikv1alpha1.TraefikServiceSpec{
					Weighted: &traefikv1alpha1.WeightedRoundRobin{
						Services: route.Services,
					},
				}

				errBuild := cb.buildServicesLB(ctx, ingressRoute.Namespace, spec, serviceName, conf.Services)
				if errBuild != nil {
					logger.Error().Err(errBuild).Send()
					continue
				}
			} else if len(route.Services) == 1 {
				fullName, serversLB, err := cb.nameAndService(ctx, ingressRoute.Namespace, route.Services[0].LoadBalancerSpec)
				if err != nil {
					logger.Error().Err(err).Send()
					continue
				}

				if serversLB != nil {
					conf.Services[serviceName] = serversLB
				} else {
					serviceName = fullName
				}
			}

			r := &dynamic.Router{
				Middlewares:   mds,
				Priority:      route.Priority,
				RuleSyntax:    route.Syntax,
				EntryPoints:   ingressRoute.Spec.EntryPoints,
				Rule:          route.Match,
				Service:       serviceName,
				Observability: route.Observability,
			}

			if ingressRoute.Spec.TLS != nil {
				r.TLS = &dynamic.RouterTLSConfig{
					CertResolver: ingressRoute.Spec.TLS.CertResolver,
					Domains:      ingressRoute.Spec.TLS.Domains,
				}

				if ingressRoute.Spec.TLS.Options != nil && len(ingressRoute.Spec.TLS.Options.Name) > 0 {
					tlsOptionsName := ingressRoute.Spec.TLS.Options.Name
					// Is a Kubernetes CRD reference, (i.e. not a cross-provider reference)
					ns := ingressRoute.Spec.TLS.Options.Namespace
					if !strings.Contains(tlsOptionsName, providerNamespaceSeparator) {
						if len(ns) == 0 {
							ns = ingressRoute.Namespace
						}
						tlsOptionsName = makeID(ns, tlsOptionsName)
					} else if len(ns) > 0 {
						logger.
							Warn().Str("TLSOption", ingressRoute.Spec.TLS.Options.Name).
							Msgf("Namespace %q is ignored in cross-provider context", ns)
					}

					if !isNamespaceAllowed(p.AllowCrossNamespace, ingressRoute.Namespace, ns) {
						logger.Error().Msgf("TLSOption %s/%s is not in the IngressRoute namespace %s",
							ns, ingressRoute.Spec.TLS.Options.Name, ingressRoute.Namespace)
						continue
					}

					r.TLS.Options = tlsOptionsName
				}
			}

			p.applyRouterTransform(ctx, r, ingressRoute)

			conf.Routers[normalized] = r
		}
	}

	return conf
}

func (p *Provider) makeMiddlewareKeys(ctx context.Context, ingRouteNamespace string, middlewares []traefikv1alpha1.MiddlewareRef) ([]string, error) {
	var mds []string

	for _, mi := range middlewares {
		name := mi.Name

		if !p.AllowCrossNamespace && strings.HasSuffix(mi.Name, providerNamespaceSeparator+providerName) {
			// Since we are not able to know if another namespace is in the name (namespace-name@kubernetescrd),
			// if the provider namespace kubernetescrd is used,
			// we don't allow this format to avoid cross namespace references.
			return nil, fmt.Errorf("invalid reference to middleware %s: with crossnamespace disallowed, the namespace field needs to be explicitly specified", mi.Name)
		}

		if strings.Contains(name, providerNamespaceSeparator) {
			if len(mi.Namespace) > 0 {
				log.Ctx(ctx).
					Warn().Str(logs.MiddlewareName, mi.Name).
					Msgf("namespace %q is ignored in cross-provider context", mi.Namespace)
			}

			mds = append(mds, name)
			continue
		}

		ns := ingRouteNamespace
		if len(mi.Namespace) > 0 {
			if !isNamespaceAllowed(p.AllowCrossNamespace, ingRouteNamespace, mi.Namespace) {
				return nil, fmt.Errorf("middleware %s/%s is not in the IngressRoute namespace %s", mi.Namespace, mi.Name, ingRouteNamespace)
			}

			ns = mi.Namespace
		}

		mds = append(mds, provider.Normalize(makeID(ns, name)))
	}

	return mds, nil
}

type configBuilder struct {
	client                       Client
	allowCrossNamespace          bool
	allowExternalNameServices    bool
	allowEmptyServices           bool
	nativeLBByDefault            bool
	disableClusterScopeResources bool
}

// buildTraefikService creates the configuration for the traefik service defined in tService,
// and adds it to the given conf map.
func (c configBuilder) buildTraefikService(ctx context.Context, tService *traefikv1alpha1.TraefikService, conf map[string]*dynamic.Service) error {
	id := provider.Normalize(makeID(tService.Namespace, tService.Name))

	if tService.Spec.Weighted != nil {
		return c.buildServicesLB(ctx, tService.Namespace, tService.Spec, id, conf)
	} else if tService.Spec.Mirroring != nil {
		return c.buildMirroring(ctx, tService, id, conf)
	}

	return errors.New("unspecified service type")
}

// buildServicesLB creates the configuration for the load-balancer of services named id, and defined in tService.
// It adds it to the given conf map.
func (c configBuilder) buildServicesLB(ctx context.Context, namespace string, tService traefikv1alpha1.TraefikServiceSpec, id string, conf map[string]*dynamic.Service) error {
	var wrrServices []dynamic.WRRService

	for _, service := range tService.Weighted.Services {
		fullName, k8sService, err := c.nameAndService(ctx, namespace, service.LoadBalancerSpec)
		if err != nil {
			return err
		}

		if k8sService != nil {
			conf[fullName] = k8sService
		}

		weight := service.Weight
		if weight == nil {
			weight = func(i int) *int { return &i }(1)
		}

		wrrServices = append(wrrServices, dynamic.WRRService{
			Name:   fullName,
			Weight: weight,
		})
	}

	var sticky *dynamic.Sticky
	if tService.Weighted.Sticky != nil && tService.Weighted.Sticky.Cookie != nil {
		sticky = &dynamic.Sticky{
			Cookie: &dynamic.Cookie{
				Name:     tService.Weighted.Sticky.Cookie.Name,
				Secure:   tService.Weighted.Sticky.Cookie.Secure,
				HTTPOnly: tService.Weighted.Sticky.Cookie.HTTPOnly,
				SameSite: tService.Weighted.Sticky.Cookie.SameSite,
				MaxAge:   tService.Weighted.Sticky.Cookie.MaxAge,
				Domain:   tService.Weighted.Sticky.Cookie.Domain,
			},
		}
		sticky.Cookie.SetDefaults()

		if tService.Weighted.Sticky.Cookie.Path != nil {
			sticky.Cookie.Path = tService.Weighted.Sticky.Cookie.Path
		}
	}

	conf[id] = &dynamic.Service{
		Weighted: &dynamic.WeightedRoundRobin{
			Services: wrrServices,
			Sticky:   sticky,
		},
	}
	return nil
}

// buildMirroring creates the configuration for the mirroring service named id, and defined by tService.
// It adds it to the given conf map.
func (c configBuilder) buildMirroring(ctx context.Context, tService *traefikv1alpha1.TraefikService, id string, conf map[string]*dynamic.Service) error {
	fullNameMain, k8sService, err := c.nameAndService(ctx, tService.Namespace, tService.Spec.Mirroring.LoadBalancerSpec)
	if err != nil {
		return err
	}

	if k8sService != nil {
		conf[fullNameMain] = k8sService
	}

	var mirrorServices []dynamic.MirrorService
	for _, mirror := range tService.Spec.Mirroring.Mirrors {
		mirroredName, k8sService, err := c.nameAndService(ctx, tService.Namespace, mirror.LoadBalancerSpec)
		if err != nil {
			return err
		}

		if k8sService != nil {
			conf[mirroredName] = k8sService
		}

		mirrorServices = append(mirrorServices, dynamic.MirrorService{
			Name:    mirroredName,
			Percent: mirror.Percent,
		})
	}

	conf[id] = &dynamic.Service{
		Mirroring: &dynamic.Mirroring{
			Service:     fullNameMain,
			Mirrors:     mirrorServices,
			MirrorBody:  tService.Spec.Mirroring.MirrorBody,
			MaxBodySize: tService.Spec.Mirroring.MaxBodySize,
		},
	}

	return nil
}

// buildServersLB creates the configuration for the load-balancer of servers defined by svc.
func (c configBuilder) buildServersLB(namespace string, svc traefikv1alpha1.LoadBalancerSpec) (*dynamic.Service, error) {
	lb := &dynamic.ServersLoadBalancer{}
	lb.SetDefaults()

	// This is required by the tests as the fake client does not apply default values.
	// TODO: remove this when the fake client apply default values.
	if svc.Strategy != "" {
		switch svc.Strategy {
		case dynamic.BalancerStrategyWRR, dynamic.BalancerStrategyP2C:
			lb.Strategy = svc.Strategy

		// Here we are just logging a warning as the default value is already applied.
		case "RoundRobin":
			log.Warn().
				Str("namespace", namespace).
				Str("service", svc.Name).
				Msgf("RoundRobin strategy value is deprecated, please use %s value instead", dynamic.BalancerStrategyWRR)

		default:
			return nil, fmt.Errorf("load-balancer strategy %s is not supported", svc.Strategy)
		}
	}

	servers, err := c.loadServers(namespace, svc)
	if err != nil {
		return nil, err
	}

	lb.Servers = servers

	if svc.HealthCheck != nil {
		lb.HealthCheck = &dynamic.ServerHealthCheck{
			Scheme:   svc.HealthCheck.Scheme,
			Path:     svc.HealthCheck.Path,
			Method:   svc.HealthCheck.Method,
			Status:   svc.HealthCheck.Status,
			Port:     svc.HealthCheck.Port,
			Hostname: svc.HealthCheck.Hostname,
			Headers:  svc.HealthCheck.Headers,
		}
		lb.HealthCheck.SetDefaults()

		if svc.HealthCheck.FollowRedirects != nil {
			lb.HealthCheck.FollowRedirects = svc.HealthCheck.FollowRedirects
		}
		if svc.HealthCheck.Mode != "http" {
			lb.HealthCheck.Mode = svc.HealthCheck.Mode
		}
		if svc.HealthCheck.Interval != nil {
			if err := lb.HealthCheck.Interval.Set(svc.HealthCheck.Interval.String()); err != nil {
				return nil, err
			}
		}
		// If the UnhealthyInterval option is not set, we use the Interval option value,
		// to check the unhealthy targets as often as the healthy ones.
		if svc.HealthCheck.UnhealthyInterval == nil {
			lb.HealthCheck.UnhealthyInterval = &lb.HealthCheck.Interval
		} else {
			var unhealthyInterval ptypes.Duration
			if err := unhealthyInterval.Set(svc.HealthCheck.UnhealthyInterval.String()); err != nil {
				return nil, err
			}
			lb.HealthCheck.UnhealthyInterval = &unhealthyInterval
		}
		if svc.HealthCheck.Timeout != nil {
			if err := lb.HealthCheck.Timeout.Set(svc.HealthCheck.Timeout.String()); err != nil {
				return nil, err
			}
		}
	}

	if svc.PassiveHealthCheck != nil {
		lb.PassiveHealthCheck = &dynamic.PassiveServerHealthCheck{}
		lb.PassiveHealthCheck.SetDefaults()

		if svc.PassiveHealthCheck.MaxFailedAttempts != nil {
			lb.PassiveHealthCheck.MaxFailedAttempts = *svc.PassiveHealthCheck.MaxFailedAttempts
		}

		if svc.PassiveHealthCheck.FailureWindow != nil {
			if err := lb.PassiveHealthCheck.FailureWindow.Set(svc.PassiveHealthCheck.FailureWindow.String()); err != nil {
				return nil, err
			}
		}
	}

	conf := svc
	lb.PassHostHeader = conf.PassHostHeader
	if lb.PassHostHeader == nil {
		passHostHeader := true
		lb.PassHostHeader = &passHostHeader
	}

	if conf.ResponseForwarding != nil && conf.ResponseForwarding.FlushInterval != "" {
		err := lb.ResponseForwarding.FlushInterval.Set(conf.ResponseForwarding.FlushInterval)
		if err != nil {
			return nil, fmt.Errorf("unable to parse flushInterval: %w", err)
		}
	}

	if svc.Sticky != nil && svc.Sticky.Cookie != nil {
		lb.Sticky = &dynamic.Sticky{
			Cookie: &dynamic.Cookie{
				Name:     svc.Sticky.Cookie.Name,
				Secure:   svc.Sticky.Cookie.Secure,
				HTTPOnly: svc.Sticky.Cookie.HTTPOnly,
				SameSite: svc.Sticky.Cookie.SameSite,
				MaxAge:   svc.Sticky.Cookie.MaxAge,
				Domain:   svc.Sticky.Cookie.Domain,
			},
		}
		lb.Sticky.Cookie.SetDefaults()

		if svc.Sticky.Cookie.Path != nil {
			lb.Sticky.Cookie.Path = svc.Sticky.Cookie.Path
		}
	}

	lb.ServersTransport, err = c.makeServersTransportKey(namespace, svc.ServersTransport)
	if err != nil {
		return nil, err
	}

	return &dynamic.Service{LoadBalancer: lb}, nil
}

func (c configBuilder) makeServersTransportKey(parentNamespace string, serversTransportName string) (string, error) {
	if serversTransportName == "" {
		return "", nil
	}

	if !c.allowCrossNamespace && strings.HasSuffix(serversTransportName, providerNamespaceSeparator+providerName) {
		// Since we are not able to know if another namespace is in the name (namespace-name@kubernetescrd),
		// if the provider namespace kubernetescrd is used,
		// we don't allow this format to avoid cross namespace references.
		return "", fmt.Errorf("invalid reference to serversTransport %s: namespace-name@kubernetescrd format is not allowed when crossnamespace is disallowed", serversTransportName)
	}

	if strings.Contains(serversTransportName, providerNamespaceSeparator) {
		return serversTransportName, nil
	}

	return provider.Normalize(makeID(parentNamespace, serversTransportName)), nil
}

func (c configBuilder) loadServers(parentNamespace string, svc traefikv1alpha1.LoadBalancerSpec) ([]dynamic.Server, error) {
	namespace := namespaceOrFallback(svc, parentNamespace)

	if !isNamespaceAllowed(c.allowCrossNamespace, parentNamespace, namespace) {
		return nil, fmt.Errorf("load balancer service %s/%s is not in the parent resource namespace %s", svc.Namespace, svc.Name, parentNamespace)
	}

	// If the service uses explicitly the provider suffix
	sanitizedName := strings.TrimSuffix(svc.Name, providerNamespaceSeparator+providerName)
	service, exists, err := c.client.GetService(namespace, sanitizedName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("kubernetes service not found: %s/%s", namespace, sanitizedName)
	}

	svcPort, err := getServicePort(service, svc.Port)
	if err != nil {
		return nil, err
	}

	if service.Spec.Type != corev1.ServiceTypeExternalName && svc.HealthCheck != nil {
		return nil, fmt.Errorf("healthCheck allowed only for ExternalName services: %s/%s", namespace, sanitizedName)
	}

	if service.Spec.Type == corev1.ServiceTypeExternalName {
		if !c.allowExternalNameServices {
			return nil, fmt.Errorf("externalName services not allowed: %s/%s", namespace, sanitizedName)
		}

		protocol, err := parseServiceProtocol(svc.Scheme, svcPort.Name, svcPort.Port)
		if err != nil {
			return nil, err
		}

		hostPort := net.JoinHostPort(service.Spec.ExternalName, strconv.Itoa(int(svcPort.Port)))

		return []dynamic.Server{{URL: fmt.Sprintf("%s://%s", protocol, hostPort)}}, nil
	}

	nativeLB := c.nativeLBByDefault
	if svc.NativeLB != nil {
		nativeLB = *svc.NativeLB
	}
	if nativeLB {
		address, err := getNativeServiceAddress(*service, *svcPort)
		if err != nil {
			return nil, fmt.Errorf("getting native Kubernetes Service address: %w", err)
		}

		protocol, err := parseServiceProtocol(svc.Scheme, svcPort.Name, svcPort.Port)
		if err != nil {
			return nil, err
		}

		return []dynamic.Server{{URL: fmt.Sprintf("%s://%s", protocol, address)}}, nil
	}

	var servers []dynamic.Server
	if service.Spec.Type == corev1.ServiceTypeNodePort && svc.NodePortLB {
		if c.disableClusterScopeResources {
			return nil, errors.New("nodes lookup is disabled")
		}

		nodes, nodesExists, nodesErr := c.client.GetNodes()
		if nodesErr != nil {
			return nil, nodesErr
		}
		if !nodesExists || len(nodes) == 0 {
			return nil, fmt.Errorf("nodes not found for NodePort service %s/%s", namespace, sanitizedName)
		}

		protocol, err := parseServiceProtocol(svc.Scheme, svcPort.Name, svcPort.Port)
		if err != nil {
			return nil, err
		}

		for _, node := range nodes {
			for _, addr := range node.Status.Addresses {
				if addr.Type == corev1.NodeInternalIP {
					hostPort := net.JoinHostPort(addr.Address, strconv.Itoa(int(svcPort.NodePort)))

					servers = append(servers, dynamic.Server{
						URL: fmt.Sprintf("%s://%s", protocol, hostPort),
					})
				}
			}
		}

		if len(servers) == 0 {
			return nil, fmt.Errorf("no servers were generated for service %s in namespace", sanitizedName)
		}

		return servers, nil
	}

	endpointSlices, err := c.client.GetEndpointSlicesForService(namespace, sanitizedName)
	if err != nil {
		return nil, fmt.Errorf("getting endpointslices: %w", err)
	}

	addresses := map[string]struct{}{}
	for _, endpointSlice := range endpointSlices {
		var port int32
		for _, p := range endpointSlice.Ports {
			if svcPort.Name == *p.Name {
				port = *p.Port
				break
			}
		}
		if port == 0 {
			continue
		}

		protocol, err := parseServiceProtocol(svc.Scheme, svcPort.Name, svcPort.Port)
		if err != nil {
			return nil, err
		}

		for _, endpoint := range endpointSlice.Endpoints {
			if !k8s.EndpointServing(endpoint) {
				continue
			}

			for _, address := range endpoint.Addresses {
				if _, ok := addresses[address]; ok {
					continue
				}

				addresses[address] = struct{}{}
				servers = append(servers, dynamic.Server{
					URL:    fmt.Sprintf("%s://%s", protocol, net.JoinHostPort(address, strconv.Itoa(int(port)))),
					Fenced: ptr.Deref(endpoint.Conditions.Terminating, false) && ptr.Deref(endpoint.Conditions.Serving, false),
				})
			}
		}
	}

	if len(servers) == 0 && !c.allowEmptyServices {
		return nil, fmt.Errorf("no servers found for %s/%s", namespace, sanitizedName)
	}

	return servers, nil
}

// nameAndService returns the name that should be used for the svc service in the generated config.
// In addition, if the service is a Kubernetes one,
// it generates and returns the configuration part for such a service,
// so that the caller can add it to the global config map.
func (c configBuilder) nameAndService(ctx context.Context, parentNamespace string, service traefikv1alpha1.LoadBalancerSpec) (string, *dynamic.Service, error) {
	svcCtx := log.Ctx(ctx).With().Str(logs.ServiceName, service.Name).Logger().WithContext(ctx)

	namespace := namespaceOrFallback(service, parentNamespace)

	if !isNamespaceAllowed(c.allowCrossNamespace, parentNamespace, namespace) {
		return "", nil, fmt.Errorf("service %s/%s not in the parent resource namespace %s", service.Namespace, service.Name, parentNamespace)
	}

	switch service.Kind {
	case "", "Service":
		serversLB, err := c.buildServersLB(namespace, service)
		if err != nil {
			return "", nil, err
		}

		fullName := fullServiceName(svcCtx, namespace, service, service.Port)

		return fullName, serversLB, nil

	case "TraefikService":
		return fullServiceName(svcCtx, namespace, service, intstr.FromInt(0)), nil, nil

	default:
		return "", nil, fmt.Errorf("unsupported service kind %s", service.Kind)
	}
}

func splitSvcNameProvider(name string) (string, string) {
	parts := strings.Split(name, providerNamespaceSeparator)

	svc := strings.Join(parts[:len(parts)-1], providerNamespaceSeparator)
	pvd := parts[len(parts)-1]

	return svc, pvd
}

func fullServiceName(ctx context.Context, namespace string, service traefikv1alpha1.LoadBalancerSpec, port intstr.IntOrString) string {
	if (port.Type == intstr.Int && port.IntVal != 0) || (port.Type == intstr.String && port.StrVal != "") {
		return provider.Normalize(fmt.Sprintf("%s-%s-%s", namespace, service.Name, &port))
	}

	if !strings.Contains(service.Name, providerNamespaceSeparator) {
		return provider.Normalize(fmt.Sprintf("%s-%s", namespace, service.Name))
	}

	name, pName := splitSvcNameProvider(service.Name)
	if pName == providerName {
		return provider.Normalize(fmt.Sprintf("%s-%s", namespace, name))
	}

	if service.Namespace != "" {
		log.Ctx(ctx).Warn().Msgf("namespace %q is ignored in cross-provider context", service.Namespace)
	}

	return provider.Normalize(name) + providerNamespaceSeparator + pName
}

func namespaceOrFallback(lb traefikv1alpha1.LoadBalancerSpec, fallback string) string {
	if lb.Namespace != "" {
		return lb.Namespace
	}
	return fallback
}

// getTLSHTTP mutates tlsConfigs.
func getTLSHTTP(ctx context.Context, ingressRoute *traefikv1alpha1.IngressRoute, k8sClient Client, tlsConfigs map[string]*tls.CertAndStores) error {
	if ingressRoute.Spec.TLS == nil {
		return nil
	}
	if ingressRoute.Spec.TLS.SecretName == "" {
		log.Ctx(ctx).Debug().Msg("No secret name provided")
		return nil
	}

	configKey := ingressRoute.Namespace + "/" + ingressRoute.Spec.TLS.SecretName
	if _, tlsExists := tlsConfigs[configKey]; !tlsExists {
		tlsConf, err := getTLS(k8sClient, ingressRoute.Spec.TLS.SecretName, ingressRoute.Namespace)
		if err != nil {
			return err
		}

		tlsConfigs[configKey] = tlsConf
	}

	return nil
}

// parseServiceProtocol parses the scheme, port name, and number to determine the correct protocol.
// an error is returned if the scheme provided is invalid.
func parseServiceProtocol(providedScheme, portName string, portNumber int32) (string, error) {
	switch providedScheme {
	case httpProtocol, httpsProtocol, "h2c":
		return providedScheme, nil
	case "":
		if portNumber == 443 || strings.HasPrefix(portName, httpsProtocol) {
			return httpsProtocol, nil
		}
		return httpProtocol, nil
	}

	return "", fmt.Errorf("invalid scheme %q specified", providedScheme)
}
