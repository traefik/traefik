package crd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/provider"
	"github.com/containous/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	"github.com/containous/traefik/v2/pkg/tls"
	corev1 "k8s.io/api/core/v1"
)

const (
	roundRobinStrategy = "RoundRobin"
	https              = "https"
	http               = "http"
)

func (p *Provider) loadIngressRouteConfiguration(ctx context.Context, client Client, tlsConfigs map[string]*tls.CertAndStores) *dynamic.HTTPConfiguration {
	conf := &dynamic.HTTPConfiguration{
		Routers:     map[string]*dynamic.Router{},
		Middlewares: map[string]*dynamic.Middleware{},
		Services:    map[string]*dynamic.Service{},
	}

	for _, ingressRoute := range client.GetIngressRoutes() {
		ctxRt := log.With(ctx, log.Str("ingress", ingressRoute.Name), log.Str("namespace", ingressRoute.Namespace))
		logger := log.FromContext(ctxRt)

		// TODO keep the name ingressClass?
		if !shouldProcessIngress(p.IngressClass, ingressRoute.Annotations[annotationKubernetesIngressClass]) {
			continue
		}

		err := getTLSHTTP(ctx, ingressRoute, client, tlsConfigs)
		if err != nil {
			logger.Errorf("Error configuring TLS: %v", err)
		}

		ingressName := ingressRoute.Name
		if len(ingressName) == 0 {
			ingressName = ingressRoute.GenerateName
		}

		for _, route := range ingressRoute.Spec.Routes {
			if route.Kind != "Rule" {
				logger.Errorf("Unsupported match kind: %s. Only \"Rule\" is supported for now.", route.Kind)
				continue
			}

			if len(route.Match) == 0 {
				logger.Errorf("Empty match rule")
				continue
			}

			if err := checkStringQuoteValidity(route.Match); err != nil {
				logger.Errorf("Invalid syntax for match rule: %s", route.Match)
				continue
			}

			serviceKey, err := makeServiceKey(route.Match, ingressName)
			if err != nil {
				logger.Error(err)
				continue
			}

			serviceName := makeID(ingressRoute.Namespace, serviceKey)

			r := configBuilder{
				conf:     conf,
				client:   client,
				toplevel: true,
				seen:     make(map[string]struct{}),
			}

			tsvc := &v1alpha1.TraefikService{
				Spec: v1alpha1.ServiceSpec{
					Weighted: &v1alpha1.WeightedRoundRobin{
						Services: route.Services,
					},
				},
			}
			tsvc.Namespace = ingressRoute.Namespace

			r.buildServicesLB(ctxRt, serviceName, tsvc)

			var mds []string
			for _, mi := range route.Middlewares {
				if strings.Contains(mi.Name, "@") {
					if len(mi.Namespace) > 0 {
						logger.
							WithField(log.MiddlewareName, mi.Name).
							Warnf("namespace %q is ignored in cross-provider context", mi.Namespace)
					}
					mds = append(mds, mi.Name)
					continue
				}

				ns := mi.Namespace
				if len(ns) == 0 {
					ns = ingressRoute.Namespace
				}
				mds = append(mds, makeID(ns, mi.Name))
			}

			normalized := provider.Normalize(serviceName)
			conf.Routers[normalized] = &dynamic.Router{
				Middlewares: mds,
				Priority:    route.Priority,
				EntryPoints: ingressRoute.Spec.EntryPoints,
				Rule:        route.Match,
				Service:     normalized,
			}

			if ingressRoute.Spec.TLS != nil {
				tlsConf := &dynamic.RouterTLSConfig{
					CertResolver: ingressRoute.Spec.TLS.CertResolver,
					Domains:      ingressRoute.Spec.TLS.Domains,
				}

				if ingressRoute.Spec.TLS.Options != nil && len(ingressRoute.Spec.TLS.Options.Name) > 0 {
					tlsOptionsName := ingressRoute.Spec.TLS.Options.Name
					// Is a Kubernetes CRD reference, (i.e. not a cross-provider reference)
					ns := ingressRoute.Spec.TLS.Options.Namespace
					if !strings.Contains(tlsOptionsName, "@") {
						if len(ns) == 0 {
							ns = ingressRoute.Namespace
						}
						tlsOptionsName = makeID(ns, tlsOptionsName)
					} else if len(ns) > 0 {
						logger.
							WithField("TLSoptions", ingressRoute.Spec.TLS.Options.Name).
							Warnf("namespace %q is ignored in cross-provider context", ns)
					}

					tlsConf.Options = tlsOptionsName
				}
				conf.Routers[normalized].TLS = tlsConf
			}
		}
	}

	return conf
}

// configBuilder holds parameters to help recursively build a dynamic config from a CRD.
type configBuilder struct {
	toplevel bool   // whether we're in the first buildServicesLB call.
	parent   string // to help with infinite recursion detection.

	conf   *dynamic.HTTPConfiguration // the configuration we're building.
	client Client
	// seen keeps track of the parent->child relations we've already seen, to detect
	// infinite recursions. it is keyed by "parent:child".
	seen map[string]struct{}
}

func splitSvcNameProvider(name string) (string, string) {
	parts := strings.Split(name, "@")
	provider := parts[len(parts)-1]
	svc := strings.Join(parts[:len(parts)-1], "@")
	return svc, provider
}

func fullServiceName(ctx context.Context, namespace, serviceName string, port int32) string {
	if port != 0 {
		return provider.Normalize(fmt.Sprintf("%s-%s-%d", namespace, serviceName, port))
	}

	if !strings.Contains(serviceName, "@") {
		return provider.Normalize(fmt.Sprintf("%s-%s", namespace, serviceName))
	}

	name, providerName := splitSvcNameProvider(serviceName)
	if providerName == "kubernetescrd" {
		return provider.Normalize(fmt.Sprintf("%s-%s", namespace, name))
	}

	log.FromContext(ctx).
		WithField(log.ServiceName, serviceName).
		Warnf("namespace %q is ignored in cross-provider context", namespace)

	return provider.Normalize(name) + "@" + providerName
}

func isCrossProvider(serviceName string) bool {
	return strings.Contains(serviceName, "@") && !strings.HasSuffix(serviceName, "@kubernetescrd")
}

// buildServicesLB creates the configuration for the load-balancer of services
// named serviceName, and defined in tsvc.
func (c configBuilder) buildServicesLB(ctx context.Context, serviceName string, tsvc *v1alpha1.TraefikService) {
	toplevel := c.toplevel
	c.toplevel = false
	services := tsvc.Spec.Weighted.Services
	var wrrsvcs []dynamic.WRRService

	for _, service := range services {
		seen := false
		namespace := namespaceOrFallback(service.LoadBalancer(), tsvc.Namespace)
		var fullName string
		switch {
		case service.Kind == "" || service.Kind == "Service":
			fullName = fullServiceName(ctx, namespace, service.Name, service.Port)
		case service.Kind == "TraefikService":
			fullName = fullServiceName(ctx, namespace, service.Name, 0)
			tuple := serviceName + ":" + fullName
			if _, exists := c.seen[tuple]; exists {
				seen = true
				log.FromContext(ctx).
					WithField(log.ServiceName, serviceName).
					WithField("serviceNamespace", namespace).
					Warnf("Infinite recursion detected: %v -> %v", serviceName, fullName)
			}
			c.seen[tuple] = struct{}{}
		default:
			log.FromContext(ctx).
				WithField(log.ServiceName, serviceName).
				WithField("serviceNamespace", namespace).
				Warnf("Unsupported service kind %v", service.Kind)
			continue
		}

		var svc *dynamic.Service
		if !seen {
			var err error
			svc, err = c.buildService(ctx, namespace, service)
			if err != nil {
				log.FromContext(ctx).
					WithField(log.ServiceName, serviceName).
					Errorf("failed to create child service of Weighted: %v", err)
				continue
			}
		}

		// sentinel for when we are a servers load-balancer
		if svc != nil {
			// shortcut for when there's only one child, that is a servers loadbalancer.
			// In that case we don't wrap it in a services loadbalancer.
			if toplevel && len(services) == 1 {
				c.conf.Services[provider.Normalize(serviceName)] = svc
				return
			}
			c.conf.Services[fullServiceName(ctx, namespace, service.Name, service.Port)] = svc
		}

		weight := service.Weight
		if weight == nil {
			weight = func(i int) *int { return &i }(1)
		}

		wrrsvcs = append(wrrsvcs, dynamic.WRRService{
			Name:   fullName,
			Weight: weight,
		})
	}

	namespace := tsvc.Namespace
	if toplevel {
		namespace = ""
	}
	normalized := fullServiceName(ctx, namespace, serviceName, 0)

	if isCrossProvider(normalized) {
		return
	}

	c.conf.Services[normalized] = &dynamic.Service{
		Weighted: &dynamic.WeightedRoundRobin{
			Services: wrrsvcs,
			Sticky:   tsvc.Spec.Weighted.Sticky,
		},
	}
}

// buildService creates the configuration for the service defined in svc.
// It can be either a Kubernetes service (a load-balancer of servers),
// or a traefik service.
func (c configBuilder) buildService(ctx context.Context, namespace string, svc v1alpha1.Service) (*dynamic.Service, error) {
	isServersLB, err := svc.IsServersLB()
	if err != nil {
		return nil, err
	}
	if isServersLB {
		return c.buildServersLB(ctx, namespace, svc)
	}

	return nil, c.buildTraefikService(ctx, namespace, svc.Name)
}

// buildTraefikService creates the configuration for the traefik service referenced as name.
func (c configBuilder) buildTraefikService(ctx context.Context, namespace, name string) error {
	if isCrossProvider(name) {
		return nil
	}

	// If the service uses explicitly the @kubernetescrd provider suffix
	sanitizedName := strings.TrimSuffix(name, "@kubernetescrd")

	tsvc, exists, err := c.client.GetTraefikService(namespace, sanitizedName)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("service not found: %s/%s", namespace, sanitizedName)
	}

	stsvc := tsvc.Spec
	if stsvc.Weighted != nil {
		c.buildServicesLB(ctx, sanitizedName, tsvc)
		return nil
	} else if stsvc.Mirroring != nil {
		return c.buildMirroring(ctx, sanitizedName, tsvc)
	}

	return errors.New("unspecified service type")
}

// buildMirror creates the configuration for one of the parts of a mirroring
// service, defined in svc.
// i.e. it is called either for the main service of a mirroring service,
// or for one of the mirrors.
// fallbackNamespace is the namespace of the parent mirroring service, which is
// used as the fallback when no namespace is defined for the part currently being
// built.
func (c configBuilder) buildMirror(ctx context.Context, fallbackNamespace string, svc v1alpha1.HasBalancer) (string, error) {
	lb := svc.LoadBalancer()
	isServersLB, err := lb.IsServersLB()
	if err != nil {
		return "", err
	}
	namespace := namespaceOrFallback(lb, fallbackNamespace)
	if !isServersLB {
		fullName := fullServiceName(ctx, namespace, lb.Name, 0)
		tuple := c.parent + ":" + fullName
		if _, exists := c.seen[tuple]; exists {
			log.FromContext(ctx).
				WithField(log.ServiceName, lb.Name).
				WithField("serviceNamespace", namespace).
				Warnf("Infinite recursion detected: %v -> %v", c.parent, fullName)
			return fullName, nil
		}
		c.seen[tuple] = struct{}{}

		if err := c.buildTraefikService(ctx, namespace, lb.Name); err != nil {
			return "", err
		}
		return fullServiceName(ctx, namespace, lb.Name, 0), nil
	}

	fullName := fullServiceName(ctx, namespace, lb.Name, lb.Port)
	service, err := c.buildServersLB(ctx, namespace, svc)
	if err != nil {
		return "", err
	}
	if isCrossProvider(fullName) {
		return fullName, nil
	}
	c.conf.Services[fullName] = service

	return fullName, nil
}

// buildMirroring creates the configuration for the mirroring service named serviceName,
// and defined by tsvc.
func (c configBuilder) buildMirroring(ctx context.Context, serviceName string, tsvc *v1alpha1.TraefikService) error {
	mirroring := tsvc.Spec.Mirroring
	namespace := tsvc.Namespace

	fullName := fullServiceName(ctx, namespace, serviceName, 0)

	// Deal with the main service first
	c.parent = fullName
	main, err := c.buildMirror(ctx, namespace, mirroring)
	if err != nil {
		return fmt.Errorf("in mirroring TraefikService: %v", err)
	}

	// Then with the "children" mirrors
	var mirrorServices []dynamic.MirrorService
	// TODO: do we return an error if no valid mirror at all was created?
	for _, mirror := range mirroring.Mirrors {
		mirroredName, err := c.buildMirror(ctx, namespace, mirror)
		if err != nil {
			log.FromContext(ctx).
				WithField(log.ServiceName, serviceName).
				Errorf("Failed to create child %v of mirror service: %v", mirror.Name, err)
			continue
		}
		mirrorServices = append(mirrorServices, dynamic.MirrorService{
			Name:    mirroredName,
			Percent: mirror.Percent,
		})
	}

	if isCrossProvider(fullName) {
		return nil
	}
	c.conf.Services[fullName] = &dynamic.Service{
		Mirroring: &dynamic.Mirroring{
			Service: main,
			Mirrors: mirrorServices,
		},
	}

	return nil
}

// buildServersLB creates the configuration for the load-balancer of servers defined by svc.
func (c configBuilder) buildServersLB(ctx context.Context, namespace string, svc v1alpha1.HasBalancer) (*dynamic.Service, error) {
	if isCrossProvider(svc.LoadBalancer().Name) {
		svcName, providerPart := splitSvcNameProvider(svc.LoadBalancer().Name)
		return nil, fmt.Errorf(`kubernetes Service %q can only be handled by "kubernetescrd" provider, and not by %q`, svcName, providerPart)
	}

	servers, err := c.loadServers(namespace, svc)
	if err != nil {
		return nil, err
	}

	lb := &dynamic.ServersLoadBalancer{}
	lb.SetDefaults()
	lb.Servers = servers

	conf := svc.LoadBalancer()
	lb.PassHostHeader = conf.PassHostHeader
	if lb.PassHostHeader == nil {
		passHostHeader := true
		lb.PassHostHeader = &passHostHeader
	}
	lb.ResponseForwarding = conf.ResponseForwarding

	ssvc, ok := svc.(v1alpha1.Service)
	if ok {
		lb.Sticky = ssvc.Sticky
	}

	return &dynamic.Service{LoadBalancer: lb}, nil
}

func namespaceOrFallback(lb v1alpha1.LoadBalancerSpec, fallback string) string {
	if lb.Namespace != "" {
		return lb.Namespace
	}
	return fallback
}

func (c configBuilder) loadServers(fallbackNamespace string, svc v1alpha1.HasBalancer) ([]dynamic.Server, error) {
	client := c.client
	conf := svc.LoadBalancer()

	strategy := conf.Strategy

	if strategy == "" {
		strategy = roundRobinStrategy
	}
	if strategy != roundRobinStrategy {
		return nil, fmt.Errorf("load balancing strategy %v is not supported", strategy)
	}

	name := conf.Name
	namespace := namespaceOrFallback(conf, fallbackNamespace)

	// If the service uses explicitly the @kubernetescrd provider suffix
	sanitizedName := strings.TrimSuffix(name, "@kubernetescrd")

	service, exists, err := client.GetService(namespace, sanitizedName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("kubernetes service not found: %s/%s", namespace, sanitizedName)
	}

	confPort := conf.Port
	var portSpec *corev1.ServicePort
	for _, p := range service.Spec.Ports {
		if confPort == p.Port {
			portSpec = &p
			break
		}
	}
	if portSpec == nil {
		return nil, errors.New("service port not found")
	}

	var servers []dynamic.Server
	if service.Spec.Type == corev1.ServiceTypeExternalName {
		return append(servers, dynamic.Server{
			URL: fmt.Sprintf("http://%s:%d", service.Spec.ExternalName, portSpec.Port),
		}), nil
	}

	endpoints, endpointsExists, endpointsErr := client.GetEndpoints(namespace, sanitizedName)
	if endpointsErr != nil {
		return nil, endpointsErr
	}
	if !endpointsExists {
		return nil, fmt.Errorf("endpoints not found for %v/%v", namespace, sanitizedName)
	}
	if len(endpoints.Subsets) == 0 {
		return nil, fmt.Errorf("subset not found for %v/%v", namespace, sanitizedName)
	}

	var port int32
	for _, subset := range endpoints.Subsets {
		for _, p := range subset.Ports {
			if portSpec.Name == p.Name {
				port = p.Port
				break
			}
		}

		if port == 0 {
			return nil, fmt.Errorf("cannot define a port for %v/%v", namespace, sanitizedName)
		}

		protocol := http
		scheme := conf.Scheme
		switch scheme {
		case http, https, "h2c":
			protocol = scheme
		case "":
			if portSpec.Port == 443 || strings.HasPrefix(portSpec.Name, https) {
				protocol = https
			}
		default:
			return nil, fmt.Errorf("invalid scheme %q specified", scheme)
		}

		for _, addr := range subset.Addresses {
			servers = append(servers, dynamic.Server{
				URL: fmt.Sprintf("%s://%s:%d", protocol, addr.IP, port),
			})
		}
	}

	return servers, nil
}

func getTLSHTTP(ctx context.Context, ingressRoute *v1alpha1.IngressRoute, k8sClient Client, tlsConfigs map[string]*tls.CertAndStores) error {
	if ingressRoute.Spec.TLS == nil {
		return nil
	}
	if ingressRoute.Spec.TLS.SecretName == "" {
		log.FromContext(ctx).Debugf("No secret name provided")
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
