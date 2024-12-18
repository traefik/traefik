package crd

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/provider"
	traefikknativev1alpha1 "github.com/traefik/traefik/v3/pkg/provider/knative/crd/traefikio/v1alpha1"
	traefikv1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	"github.com/traefik/traefik/v3/pkg/tls"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	knativenetworkingv1alpha1 "knative.dev/networking/pkg/apis/networking/v1alpha1"
)

const revisionHeaderName = "revision-tag"

func (p *Provider) loadKnativeIngressRouteConfiguration(ctx context.Context, client Client, tlsConfigs map[string]*tls.CertAndStores) *dynamic.HTTPConfiguration {
	conf := &dynamic.HTTPConfiguration{
		Routers:     map[string]*dynamic.Router{},
		Middlewares: map[string]*dynamic.Middleware{},
		Services:    map[string]*dynamic.Service{},
	}

	// address ingress routes configuration for knatives
	for _, ingressRoute := range client.GetIngressRoutes() {
		logger := log.Ctx(ctx).With().Str("ingressknative", ingressRoute.Name).Str("namespace",
			ingressRoute.Namespace).Logger()

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

		for _, route := range ingressRoute.Spec.Routes {
			if route.Kind != "Rule" {
				logger.Error().Msgf("Unsupported match kind: %s. Only \"Rule\" is supported for now.", route.Kind)
				continue
			}

			if len(route.Match) == 0 {
				logger.Error().Msg("Empty match rule")
				continue
			}

			mds, err := p.makeMiddlewareKeys(ctx, ingressRoute.Namespace, route.Middlewares)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to create middleware keys")
				continue
			}

			serviceKey, err := makeServiceKey(route.Match, ingressName)
			if err != nil {
				logger.Error().Err(err).Send()
				continue
			}

			serviceName := provider.Normalize(makeID(ingressRoute.Namespace, serviceKey))

			knativeIngressRoute, _, err := client.GetKnativeIngressRoute(ingressRoute.Namespace, route.Services[0].Name)
			if err != nil {
				logger.Error().Err(err).Msgf("Cannot get IngressRoute %q", route.Services[0].Name)
				continue
			}
			knativeResult := buildKnativeService(ctx, client, knativeIngressRoute, conf.Middlewares, conf.Services, serviceName)
			for _, result := range knativeResult {
				if result.Err != nil {
					logger.Error().Err(result.Err).Send()
					continue
				}

				match := route.Match
				if result.Tag != "" {
					match = fmt.Sprintf("(%s) && Header(`%s`, `%s`)", match, revisionHeaderName, result.Tag)
				}
				if result.Middleware != "" {
					mds = append(mds, result.Middleware)
				}

				r := &dynamic.Router{
					Middlewares: mds,
					Priority:    route.Priority,
					RuleSyntax:  route.Syntax,
					EntryPoints: ingressRoute.Spec.EntryPoints,
					Rule:        match,
					Service:     result.ServiceKey,
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
							tlsOptionsName = provider.Normalize(makeID(ns, tlsOptionsName))
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
				conf.Routers[provider.Normalize(makeID(result.Tag, result.ServiceKey))] = r
			}

			err = p.updateKnativeIngressStatus(client, knativeIngressRoute)
			if err != nil {
				logger.Error().Err(err).Msgf("error %v", err)
				return nil
			}
		}
	}
	return conf
}

// getTLSHTTP mutates tlsConfigs.
func getTLSHTTP(ctx context.Context, ingressRoute *traefikknativev1alpha1.IngressRouteKnative, k8sClient Client, tlsConfigs map[string]*tls.CertAndStores) error {
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

func createKnativeLoadBalancerServerHTTP(client Client, namespace string, service traefikknativev1alpha1.ServiceKnativeSpec) (*dynamic.Service, error) {
	servers, err := loadKnativeServers(client, namespace, service)
	if err != nil {
		return nil, err
	}

	// TODO: support other strategies.
	lb := &dynamic.ServersLoadBalancer{}
	lb.SetDefaults()

	lb.Servers = servers

	lb.PassHostHeader = service.PassHostHeader
	if lb.PassHostHeader == nil {
		passHostHeader := true
		lb.PassHostHeader = &passHostHeader
	}
	lb.ResponseForwarding, err = convertResponseForwarding(service.ResponseForwarding)
	if err != nil {
		return nil, err
	}

	return &dynamic.Service{
		LoadBalancer: lb,
	}, nil
}

func convertResponseForwarding(rf *traefikknativev1alpha1.ResponseForwarding) (*dynamic.ResponseForwarding, error) {
	if rf == nil {
		return nil, nil
	}

	flushInterval, err := time.ParseDuration(rf.FlushInterval)
	if err != nil {
		return nil, err
	}

	flushIntervalProto := ptypes.Duration(flushInterval)

	return &dynamic.ResponseForwarding{
		FlushInterval: flushIntervalProto,
	}, nil
}

func loadKnativeServers(client Client, namespace string, svc traefikknativev1alpha1.ServiceKnativeSpec) ([]dynamic.Server, error) {
	strategy := ""
	if strategy == "" {
		strategy = "RoundRobin"
	}
	if strategy != "RoundRobin" {
		return nil, fmt.Errorf("load balancing strategy %v is not supported", strategy)
	}

	serverlessservice, exists, err := client.GetServerlessService(namespace, svc.Name)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("serverless service not found %s/%s", namespace, svc.Name)
	}

	service, exists, err := client.GetService(namespace, serverlessservice.Status.ServiceName)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("service not found %s/%s", namespace, svc.Name)
	}

	var portSpec *corev1.ServicePort
	for _, p := range service.Spec.Ports {
		if svc.Port == intstr.FromInt32(p.Port) {
			portSpec = p.DeepCopy()
			break
		}
	}

	if portSpec == nil {
		return nil, errors.New("service port not found")
	}

	var servers []dynamic.Server
	if service.Spec.ClusterIP != "" {
		if svc.Port == intstr.FromInt32(80) {
			servers = append(servers, dynamic.Server{
				URL: fmt.Sprintf("%s://%s:%d", "http", service.Spec.ClusterIP, portSpec.Port),
			})
		} else if svc.Port == intstr.FromInt32(443) {
			servers = append(servers, dynamic.Server{
				URL: fmt.Sprintf("%s://%s:%d", "https", service.Spec.ClusterIP, portSpec.Port),
			})
		}
	}
	return servers, nil
}

func shouldProcessIngress(ingressClass, ingressClassAnnotation string) bool {
	return ingressClass == ingressClassAnnotation ||
		(len(ingressClass) == 0 && ingressClassAnnotation == traefikDefaultIngressClass)
}

func makeServiceKey(rule, ingressName string) (string, error) {
	h := sha256.New()
	if _, err := h.Write([]byte(rule)); err != nil {
		return "", err
	}

	key := fmt.Sprintf("%s-%.10x", ingressName, h.Sum(nil))
	return key, nil
}

func (p *Provider) makeMiddlewareKeys(ctx context.Context, ingRouteNamespace string,
	middlewares []traefikv1alpha1.MiddlewareRef,
) ([]string, error) {
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

type ServiceResult struct {
	ServiceKey string
	Tag        string
	Middleware string
	Err        error
}

func buildKnativeService(ctx context.Context, client Client, ingressRoute *knativenetworkingv1alpha1.Ingress,
	middleware map[string]*dynamic.Middleware, conf map[string]*dynamic.Service, serviceName string,
) []ServiceResult {
	logger := log.Ctx(ctx).With().Str("ingressknative", ingressRoute.Name).Str("service", serviceName).
		Str("namespace", ingressRoute.Namespace).Logger()
	var results []ServiceResult

	for index, route := range ingressRoute.Spec.Rules {
		var (
			tag            string
			tagServiceName string
			headers        string
		)

		// If the tag is defined as part of knative traffic split, it will be appended as prefix to hosts.
		// This is only for tag based routing,
		// hosts with tags as prefixes are populated from second element in the rules.
		if len(route.Hosts) > 0 && index != 0 {
			index := strings.Index(route.Hosts[0], ingressRoute.Name)
			if index != -1 {
				tag = route.Hosts[0][:index-1] // -1 to remove the trailing - character
			}
		}
		if route.HTTP == nil {
			logger.Warn().Msgf("No HTTP rule defined for Knative service %s", ingressRoute.Name)
			continue
		}

		for _, pathroute := range route.HTTP.Paths {
			for _, service := range pathroute.Splits {
				balancerServerHTTP, err := createKnativeLoadBalancerServerHTTP(client, service.ServiceNamespace,
					traefikknativev1alpha1.ServiceKnativeSpec{
						Name: service.ServiceName,
						Port: service.ServicePort,
					})
				if err != nil {
					logger.Err(err).Str("serviceName", service.ServiceName).Str("servicePort",
						service.ServicePort.String()).Msgf("Cannot create service: %v", err)
					continue
				}

				serviceKey := fmt.Sprintf("%s-%s-%d", service.ServiceNamespace, service.ServiceName,
					int32(service.ServicePort.IntValue()))
				conf[serviceKey] = balancerServerHTTP
				if len(route.HTTP.Paths) == 1 && len(pathroute.Splits) == 1 {
					if len(service.AppendHeaders) > 0 {
						headers = provider.Normalize(makeID(serviceKey, "AppendHeader"))
						middleware[headers] = &dynamic.Middleware{
							Headers: &dynamic.Headers{
								CustomRequestHeaders: service.AppendHeaders,
							},
						}
					}
					tagServiceName = serviceKey
					continue
				}
				tagServiceName = provider.Normalize(makeID(tag, serviceName))
				srv := dynamic.WRRService{Name: serviceKey}
				srv.SetDefaults()
				if service.Percent != 0 {
					val := service.Percent
					srv.Weight = &val
					srv.Headers = service.AppendHeaders
				}

				if conf[tagServiceName] == nil {
					conf[tagServiceName] = &dynamic.Service{Weighted: &dynamic.WeightedRoundRobin{}}
				}
				conf[tagServiceName].Weighted.Services = append(conf[tagServiceName].Weighted.Services, srv)
				// results = append(results, ServiceResult{serviceName, tag, nil})
			}
		}
		results = append(results, ServiceResult{tagServiceName, tag, headers, nil})
	}
	return results
}

func makeID(s1, s2 string) string {
	if s1 == "" {
		return s2
	}
	if s2 == "" {
		return s1
	}
	return fmt.Sprintf("%s-%s", s1, s2)
}

func isNamespaceAllowed(allowCrossNamespace bool, parentNamespace, namespace string) bool {
	// If allowCrossNamespace option is not defined the default behavior is to allow cross namespace references.
	return allowCrossNamespace || parentNamespace == namespace
}

func (p *Provider) updateKnativeIngressStatus(client Client, ingressRoute *knativenetworkingv1alpha1.Ingress) error {
	if ingressRoute.GetStatus() == nil ||
		!ingressRoute.GetStatus().GetCondition(knativenetworkingv1alpha1.IngressConditionNetworkConfigured).IsTrue() ||
		ingressRoute.GetGeneration() != ingressRoute.GetStatus().ObservedGeneration {
		ingressRoute.Status.MarkLoadBalancerReady(
			// public lbs
			[]knativenetworkingv1alpha1.LoadBalancerIngressStatus{{
				Domain:         p.LoadBalancerDomain,
				DomainInternal: p.LoadBalancerDomainInternal,
				IP:             p.LoadBalancerIP,
			}},
			// private lbs
			[]knativenetworkingv1alpha1.LoadBalancerIngressStatus{{
				Domain:         p.LoadBalancerDomain,
				DomainInternal: p.LoadBalancerDomainInternal,
				IP:             p.LoadBalancerIP,
			}},
		)

		ingressRoute.Status.MarkNetworkConfigured()
		ingressRoute.Status.ObservedGeneration = ingressRoute.GetGeneration()

		return client.UpdateKnativeIngressStatus(ingressRoute)
	}
	return nil
}
