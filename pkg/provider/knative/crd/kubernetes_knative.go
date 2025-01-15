package crd

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/provider"
	traefikv1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	knativenetworkingv1alpha1 "knative.dev/networking/pkg/apis/networking/v1alpha1"
)

func (p *Provider) loadKnativeIngressRouteConfiguration(ctx context.Context, client Client) *dynamic.HTTPConfiguration {
	conf := &dynamic.HTTPConfiguration{
		Routers:     map[string]*dynamic.Router{},
		Middlewares: map[string]*dynamic.Middleware{},
		Services:    map[string]*dynamic.Service{},
	}

	for _, ingressRoute := range client.GetKnativeIngressRoutes() {
		logger := log.Ctx(ctx).With().Str("KNativeIngress", ingressRoute.Name).Str("namespace",
			ingressRoute.Namespace).Logger()

		if !shouldProcessIngress(p.IngressClass, ingressRoute.Annotations[annotationKubernetesIngressClass]) {
			continue
		}

		ingressName := ingressRoute.Name
		if len(ingressName) == 0 {
			ingressName = ingressRoute.GenerateName
		}

		cb := configBuilder{
			client:              client,
			allowCrossNamespace: p.AllowCrossNamespace,
		}

		serviceKey, err := makeServiceKey(ingressRoute.Namespace, ingressName)
		if err != nil {
			logger.Error().Err(err).Send()
			continue
		}

		serviceName := provider.Normalize(makeID(ingressRoute.Namespace, serviceKey))

		knativeResult := cb.buildKnativeService(ctx, ingressRoute, conf.Middlewares, conf.Services, serviceName)

		for _, result := range knativeResult {
			if result.Err != nil {
				logger.Error().Err(result.Err).Send()
				continue
			}

			var hosts []string
			var mds []string
			for _, host := range result.Hosts {
				hosts = append(hosts, fmt.Sprintf("Host(`%v`)", host))
			}
			match := fmt.Sprintf("(%v)", strings.Join(hosts, " || "))

			if result.Middleware != "" {
				mds = append(mds, result.Middleware)
			}

			r := &dynamic.Router{
				Middlewares: mds,
				Rule:        match,
				Service:     result.ServiceKey,
			}
			conf.Routers[provider.Normalize(result.ServiceKey)] = r
		}
		err = p.updateKnativeIngressStatus(client, ingressRoute)
		if err != nil {
			logger.Error().Err(err).Msgf("error %v", err)
			return nil
		}
	}
	return conf
}

type configBuilder struct {
	client              Client
	allowCrossNamespace bool
}

func (c configBuilder) createKnativeLoadBalancerServerHTTP(namespace string,
	service traefikv1alpha1.Service,
) (*dynamic.Service, error) {
	servers, err := c.loadKnativeServers(namespace, service)
	if err != nil {
		return nil, err
	}

	// TODO: support other strategies.
	lb := &dynamic.ServersLoadBalancer{}
	lb.SetDefaults()

	lb.Servers = servers

	conf := service
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

	return &dynamic.Service{
		LoadBalancer: lb,
	}, nil
}

func (c configBuilder) loadKnativeServers(namespace string,
	svc traefikv1alpha1.Service,
) ([]dynamic.Server, error) {
	strategy := ""
	if strategy == "" {
		strategy = "RoundRobin"
	}
	if strategy != "RoundRobin" {
		return nil, fmt.Errorf("load balancing strategy %v is not supported", strategy)
	}

	serverlessservice, exists, err := c.client.GetServerlessService(namespace, svc.Name)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("serverless service not found %s/%s", namespace, svc.Name)
	}

	service, exists, err := c.client.GetService(namespace, serverlessservice.Status.ServiceName)
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

type ServiceResult struct {
	ServiceKey string
	Hosts      []string
	Middleware string
	Err        error
}

func (c configBuilder) buildKnativeService(ctx context.Context, ingressRoute *knativenetworkingv1alpha1.Ingress,
	middleware map[string]*dynamic.Middleware, conf map[string]*dynamic.Service, serviceName string,
) []*ServiceResult {
	logger := log.Ctx(ctx).With().Str("ingressknative", ingressRoute.Name).Str("service", serviceName).
		Str("namespace", ingressRoute.Namespace).Logger()
	var results []*ServiceResult

	for _, route := range ingressRoute.Spec.Rules {
		var (
			tagServiceName string
			headers        string
		)

		if route.HTTP == nil {
			logger.Warn().Msgf("No HTTP rule defined for Knative service %s", ingressRoute.Name)
			continue
		}

		for _, pathroute := range route.HTTP.Paths {
			for _, service := range pathroute.Splits {
				balancerServerHTTP, err := c.createKnativeLoadBalancerServerHTTP(service.ServiceNamespace, traefikv1alpha1.Service{
					LoadBalancerSpec: traefikv1alpha1.LoadBalancerSpec{
						Name: service.ServiceName,
						Port: service.ServicePort,
					},
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
				tagServiceName = serviceName
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
		results = append(results, &ServiceResult{tagServiceName, route.Hosts, headers, nil})
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
