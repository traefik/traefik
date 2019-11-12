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

		cb := configBuilder{client}
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

			normalized := provider.Normalize(makeID(ingressRoute.Namespace, serviceKey))
			serviceName := normalized

			if len(route.Services) > 1 {
				spec := v1alpha1.ServiceSpec{
					Weighted: &v1alpha1.WeightedRoundRobin{
						Services: route.Services,
					},
				}

				errBuild := cb.buildServicesLB(ctx, ingressRoute.Namespace, spec, serviceName, conf.Services)
				if errBuild != nil {
					logger.Error(err)
					continue
				}
			} else if len(route.Services) == 1 {
				fullName, serversLB, err := cb.foo(ctx, ingressRoute.Namespace, route.Services[0])
				if err != nil {
					logger.Error(err)
					continue
				}

				if serversLB != nil {
					conf.Services[serviceName] = serversLB
				} else {
					serviceName = fullName
				}
			}

			conf.Routers[normalized] = &dynamic.Router{
				Middlewares: mds,
				Priority:    route.Priority,
				EntryPoints: ingressRoute.Spec.EntryPoints,
				Rule:        route.Match,
				Service:     serviceName,
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

type configBuilder struct {
	client Client
}

func (c configBuilder) buildTraefikService(ctx context.Context, tsvc *v1alpha1.TraefikService, conf map[string]*dynamic.Service) error {
	stsvc := tsvc.Spec
	id := provider.Normalize(makeID(tsvc.Namespace, tsvc.Name))
	if stsvc.Weighted != nil {
		return c.buildServicesLB(ctx, tsvc.Namespace, stsvc, id, conf)
	} else if stsvc.Mirroring != nil {
		return c.buildMirroring(ctx, tsvc, id, conf)
	}

	return errors.New("unspecified service type")
}

// buildServicesLB creates the configuration for the load-balancer of services
// named serviceName, and defined in tsvc.
func (c configBuilder) buildServicesLB(ctx context.Context, namespace string, tsvc v1alpha1.ServiceSpec, id string, conf map[string]*dynamic.Service) error {
	services := tsvc.Weighted.Services
	var wrrsvcs []dynamic.WRRService

	for _, service := range services {
		fullName, serviceGenerated, err := c.foo(ctx, namespace, service)
		if err != nil {
			return err
		}

		if serviceGenerated != nil {
			conf[fullName] = serviceGenerated
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

	conf[id] = &dynamic.Service{
		Weighted: &dynamic.WeightedRoundRobin{
			Services: wrrsvcs,
			Sticky:   tsvc.Weighted.Sticky,
		},
	}
	return nil
}

// buildMirroring creates the configuration for the mirroring service named serviceName,
// and defined by tsvc.
func (c configBuilder) buildMirroring(ctx context.Context, tsvc *v1alpha1.TraefikService, id string, conf map[string]*dynamic.Service) error {
	mirroring := tsvc.Spec.Mirroring
	namespace := tsvc.Namespace

	fullNameMain, serviceGenerated, err := c.foo(ctx, tsvc.Namespace, tsvc.Spec.Mirroring)
	if err != nil {
		return err
	}

	if serviceGenerated != nil {
		conf[fullNameMain] = serviceGenerated
	}

	var mirrorServices []dynamic.MirrorService
	for _, mirror := range mirroring.Mirrors {
		mirroredName, serviceGenerated, err := c.foo(ctx, namespace, mirror)
		if err != nil {
			return err
		}

		if serviceGenerated != nil {
			conf[mirroredName] = serviceGenerated
		}

		mirrorServices = append(mirrorServices, dynamic.MirrorService{
			Name:    mirroredName,
			Percent: mirror.Percent,
		})
	}

	conf[id] = &dynamic.Service{
		Mirroring: &dynamic.Mirroring{
			Service: fullNameMain,
			Mirrors: mirrorServices,
		},
	}

	return nil
}

func (c configBuilder) buildServersLB(ctx context.Context, namespace string, svc v1alpha1.HasBalancer) (*dynamic.Service, error) {
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

func (c configBuilder) loadServers(fallbackNamespace string, svc v1alpha1.HasBalancer) ([]dynamic.Server, error) {
	conf := svc.LoadBalancer()
	client := c.client

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

func (c configBuilder) foo(ctx context.Context, namespaceService string, b v1alpha1.HasBalancer) (string, *dynamic.Service, error) {
	service := b.LoadBalancer()
	namespace := namespaceOrFallback(service, namespaceService)
	var fullName string
	switch {
	case service.Kind == "" || service.Kind == "Service":
		fullName = fullServiceName(ctx, namespace, service.Name, service.Port)
		serversLB, err := c.buildServersLB(ctx, namespace, b)
		if err != nil {
			return "", nil, err
		}
		return fullName, serversLB, nil
		// TODO Service creation
	case service.Kind == "TraefikService":
		fullName = fullServiceName(ctx, namespace, service.Name, 0)
	default:
		return "", nil, fmt.Errorf("unsupported service kind %v", service.Kind)
	}
	return fullName, nil, nil
	// ref service
	// 1. TraefikService @kubernetescrd (kind: TraefikService)
	// 2. TraefikService @other (kind: TraefikService)
	// 3. ServiceLoadBalancers (kind: Service)
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

	// At this point, if namespace == "default", we do not know whether it had been
	// intentionnally set as such, or if we're simply hitting the value set by default.
	// But as it is most likely very much the latter, and we do not want to systematically log spam
	// users in that case, we skip logging whenever the namespace is "default".
	if namespace != "default" {
		log.FromContext(ctx).
			WithField(log.ServiceName, serviceName).
			Warnf("namespace %q is ignored in cross-provider context", namespace)
	}

	return provider.Normalize(name) + "@" + providerName
}

func namespaceOrFallback(lb v1alpha1.LoadBalancerSpec, fallback string) string {
	if lb.Namespace != "" {
		return lb.Namespace
	}
	return fallback
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
