package crd

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/provider"
	traefikv1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	"github.com/traefik/traefik/v3/pkg/tls"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	knativenetworkingv1alpha1 "knative.dev/networking/pkg/apis/networking/v1alpha1"
)

const (
	httpsProtocol = "https"
	httpProtocol  = "http"
	h2cProtocol   = "h2c"
	http2Protocol = "http2"
)

func (p *Provider) loadKnativeIngressRouteConfiguration(ctx context.Context, client Client,
	tlsConfigs map[string]*tls.CertAndStores,
) *dynamic.HTTPConfiguration {
	conf := &dynamic.HTTPConfiguration{
		Routers:     map[string]*dynamic.Router{},
		Middlewares: map[string]*dynamic.Middleware{},
		Services:    map[string]*dynamic.Service{},
	}

	for _, ingressRoute := range client.GetKnativeIngressRoutes() {
		logger := log.Ctx(ctx).With().Str("KNativeIngress", ingressRoute.Name).Str("namespace",
			ingressRoute.Namespace).Logger()

		err := getTLSHTTP(ctx, ingressRoute, client, tlsConfigs)
		if err != nil {
			logger.Error().Err(err).Send()
			continue
		}

		if !(traefikDefaultIngressClass == ingressRoute.Annotations[annotationKubernetesIngressClass]) {
			logger.Debug().Msgf("Skipping Ingress %s/%s", ingressRoute.Namespace, ingressRoute.Name)
			continue
		}

		ingressName := getIngressName(ingressRoute)
		cb := configBuilder{client: client, allowCrossNamespace: p.AllowCrossNamespace}

		serviceKey, err := makeServiceKey(ingressRoute.Namespace, ingressName)
		if err != nil {
			logger.Error().Err(err).Send()
			continue
		}

		serviceName := provider.Normalize(makeID(ingressRoute.Namespace, serviceKey))
		knativeResult := cb.buildKnativeService(ctx, ingressRoute, conf.Middlewares, conf.Services, serviceName)

		for _, result := range knativeResult {
			var entrypoints []string

			if result.Visibility == knativenetworkingv1alpha1.IngressVisibilityClusterLocal {
				if p.EntrypointsInternal == nil {
					continue // skip route creation as no internal entrypoints are defined for cluster local visibility
				}
				entrypoints = p.EntrypointsInternal
			} else {
				entrypoints = p.Entrypoints
			}

			if result.Err != nil {
				logger.Error().Err(result.Err).Send()
				continue
			}

			match := buildMatchRule(result.Hosts, result.Path)
			mds := append([]string{}, result.Middleware...)

			r := &dynamic.Router{
				Middlewares: mds,
				Rule:        match,
				Service:     result.ServiceKey,
			}

			if entrypoints != nil {
				r.EntryPoints = entrypoints
			}
			conf.Routers[provider.Normalize(result.ServiceKey)] = r
		}
		if err := p.updateKnativeIngressStatus(client, ingressRoute); err != nil {
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
	logger := log.With().Logger()
	strategy := ""
	if strategy == "" {
		strategy = "RoundRobin"
	}
	if strategy != "RoundRobin" {
		return nil, fmt.Errorf("load balancing strategy %v is not supported", strategy)
	}

	serverlessservice, exists, err := c.client.GetServerlessService(namespace, svc.Name)
	if err != nil {
		logger.Info().Msgf("Unable to find serverlessservice, trying to find service %s/%s", namespace, svc.Name)
	}

	serviceName := svc.Name
	if exists {
		serviceName = serverlessservice.Status.ServiceName
	}

	service, exists, err := c.client.GetService(namespace, serviceName)
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
		protocol, err := parseServiceProtocol(portSpec.Name, portSpec.Port)
		if err != nil {
			return nil, err
		}

		hostPort := net.JoinHostPort(service.Spec.ClusterIP, strconv.Itoa(int(portSpec.Port)))
		servers = append(servers, dynamic.Server{
			URL: fmt.Sprintf("%s://%s", protocol, hostPort),
		})
	}
	return servers, nil
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
	Middleware []string
	Path       string
	Visibility knativenetworkingv1alpha1.IngressVisibility
	Err        error
}

func (c configBuilder) buildKnativeService(ctx context.Context, ingressRoute *knativenetworkingv1alpha1.Ingress,
	middleware map[string]*dynamic.Middleware, conf map[string]*dynamic.Service, serviceName string,
) []*ServiceResult {
	logger := log.Ctx(ctx).With().Str("ingressknative", ingressRoute.Name).Str("service", serviceName).
		Str("namespace", ingressRoute.Namespace).Logger()
	var results []*ServiceResult

	for ruleIndex, route := range ingressRoute.Spec.Rules {
		if route.HTTP == nil {
			logger.Warn().Msgf("No HTTP rule defined for Knative service %s", ingressRoute.Name)
			continue
		}

		for pathIndex, pathroute := range route.HTTP.Paths {
			var tagServiceName string
			headers := c.buildHeaders(middleware, serviceName, ruleIndex, pathIndex, pathroute.AppendHeaders)
			path := pathroute.Path

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
				if len(pathroute.Splits) == 1 {
					if len(service.AppendHeaders) > 0 {
						headers = append(headers, provider.Normalize(makeID(serviceKey, "KnativeHeader")))
						middleware[headers[len(headers)-1]] = &dynamic.Middleware{
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
			}
			results = append(results, &ServiceResult{tagServiceName, route.Hosts, headers, path, route.Visibility, nil})
		}
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
	log.Ctx(context.Background()).Debug().Msgf("Updating status for Ingress %s/%s", ingressRoute.Namespace, ingressRoute.Name)
	log.Ctx(context.Background()).Debug().Msgf("ingressRoute.GetStatus() %v", ingressRoute.GetStatus())
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

// parseServiceProtocol parses the scheme, port name, and number to determine the correct protocol.
// an error is returned if the scheme provided is invalid.
func parseServiceProtocol(portName string, portNumber int32) (string, error) {
	switch portName {
	case httpProtocol, httpsProtocol:
		return portName, nil
	case http2Protocol, h2cProtocol:
		return h2cProtocol, nil
	case "":
		if portNumber == 443 || strings.HasPrefix(portName, httpsProtocol) {
			return httpsProtocol, nil
		}
		return httpProtocol, nil
	}

	return "", fmt.Errorf("invalid scheme %q specified", portName)
}

func getIngressName(ingressRoute *knativenetworkingv1alpha1.Ingress) string {
	if len(ingressRoute.Name) == 0 {
		return ingressRoute.GenerateName
	}
	return ingressRoute.Name
}

func buildMatchRule(hosts []string, path string) string {
	var hostRules []string
	for _, host := range hosts {
		hostRules = append(hostRules, fmt.Sprintf("Host(`%v`)", host))
	}
	match := fmt.Sprintf("(%v)", strings.Join(hostRules, " || "))
	if len(path) > 0 {
		match += fmt.Sprintf(" && PathPrefix(`%v`)", path)
	}
	return match
}

func (c configBuilder) buildHeaders(middleware map[string]*dynamic.Middleware, serviceName string, ruleIndex, pathIndex int, appendHeaders map[string]string) []string {
	if appendHeaders == nil {
		return nil
	}

	headerID := provider.Normalize(makeID(serviceName, fmt.Sprintf("PreHeader-%d-%d", ruleIndex, pathIndex)))
	middleware[headerID] = &dynamic.Middleware{
		Headers: &dynamic.Headers{
			CustomRequestHeaders: appendHeaders,
		},
	}

	return []string{headerID}
}

// getTLSHTTP mutates tlsConfigs.
func getTLSHTTP(ctx context.Context, ingressRoute *knativenetworkingv1alpha1.Ingress, k8sClient Client, tlsConfigs map[string]*tls.CertAndStores) error {
	if ingressRoute.Spec.TLS != nil {
		for _, tls := range ingressRoute.Spec.TLS {
			if tls.SecretName == "" {
				log.Ctx(ctx).Debug().Msg("No secret name provided")
				continue
			}

			configKey := ingressRoute.Namespace + "/" + tls.SecretName
			if _, tlsExists := tlsConfigs[configKey]; !tlsExists {
				tlsConf, err := getTLS(k8sClient, tls.SecretName, ingressRoute.Namespace)
				if err != nil {
					return err
				}

				tlsConfigs[configKey] = tlsConf
			}
		}
	}
	return nil
}
