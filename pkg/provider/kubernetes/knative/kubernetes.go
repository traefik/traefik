package knative

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"maps"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/mitchellh/hashstructure"
	"github.com/rs/zerolog/log"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/job"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/provider"
	traefikv1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	"github.com/traefik/traefik/v3/pkg/safe"
	"github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	knativenetworking "knative.dev/networking/pkg/apis/networking"
	knativenetworkingv1alpha1 "knative.dev/networking/pkg/apis/networking/v1alpha1"
)

const (
	providerName            = "knative"
	traefikIngressClassName = "traefik.ingress.networking.knative.dev"
)

const (
	httpsProtocol = "https"
	httpProtocol  = "http"
	h2cProtocol   = "h2c"
	http2Protocol = "http2"
)

// Provider holds configurations of the provider.
type Provider struct {
	Endpoint                   string          `description:"Kubernetes server endpoint (required for external cluster client)." json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Token                      string          `description:"Kubernetes bearer token (not needed for in-cluster client)." json:"token,omitempty" toml:"token,omitempty" yaml:"token,omitempty"`
	CertAuthFilePath           string          `description:"Kubernetes certificate authority file path (not needed for in-cluster client)." json:"certAuthFilePath,omitempty" toml:"certAuthFilePath,omitempty" yaml:"certAuthFilePath,omitempty"`
	Namespaces                 []string        `description:"Kubernetes namespaces." json:"namespaces,omitempty" toml:"namespaces,omitempty" yaml:"namespaces,omitempty" export:"true"`
	LabelSelector              string          `description:"Kubernetes label selector to use." json:"labelSelector,omitempty" toml:"labelSelector,omitempty" yaml:"labelSelector,omitempty" export:"true"`
	LoadBalancerIP             string          `description:"set for load-balancer ingress points that are IP based." json:"loadBalancerIP,omitempty" toml:"loadBalancerIP,omitempty" yaml:"loadBalancerIP,omitempty"`
	LoadBalancerDomain         string          `description:"set for load-balancer ingress points that are DNS based." json:"loadBalancerDomain,omitempty" toml:"loadBalancerDomain,omitempty" yaml:"loadBalancerDomain,omitempty"`
	LoadBalancerDomainInternal string          `description:"set if there is a cluster-local DNS name to access the Ingress." json:"loadBalancerDomainInternal,omitempty" toml:"loadBalancerDomainInternal,omitempty" yaml:"loadBalancerDomainInternal,omitempty"`
	Entrypoints                []string        `description:"Entry points for Knative. (default: [\"traefik\"])" json:"entrypoints,omitempty" toml:"entrypoints,omitempty" yaml:"entrypoints,omitempty" export:"true"`
	EntrypointsInternal        []string        `description:"Entry points for Knative." json:"entrypointsInternal,omitempty" toml:"entrypointsInternal,omitempty" yaml:"entrypointsInternal,omitempty" export:"true"`
	ThrottleDuration           ptypes.Duration `description:"Ingress refresh throttle duration" json:"throttleDuration,omitempty" toml:"throttleDuration,omitempty" yaml:"throttleDuration,omitempty"`

	k8sClient         Client
	lastConfiguration safe.Safe
}

// Init the provider.
func (p *Provider) Init() error {
	logger := log.With().Str(logs.ProviderName, providerName).Logger()

	// Initializes Kubernetes client.
	var err error
	p.k8sClient, err = p.newK8sClient(logger.WithContext(context.Background()))
	if err != nil {
		return fmt.Errorf("creating kubernetes client: %w", err)
	}

	return nil
}

// Provide allows the k8s provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	logger := log.With().Str(logs.ProviderName, providerName).Logger()
	ctxLog := logger.WithContext(context.Background())

	pool.GoCtx(func(ctxPool context.Context) {
		operation := func() error {
			eventsChan, err := p.k8sClient.WatchAll(p.Namespaces, ctxPool.Done())
			if err != nil {
				logger.Error().Msgf("Error watching kubernetes events: %v", err)
				timer := time.NewTimer(1 * time.Second)
				select {
				case <-timer.C:
					return err
				case <-ctxPool.Done():
					return nil
				}
			}

			throttleDuration := time.Duration(p.ThrottleDuration)
			throttledChan := throttleEvents(ctxLog, throttleDuration, pool, eventsChan)
			if throttledChan != nil {
				eventsChan = throttledChan
			}

			for {
				select {
				case <-ctxPool.Done():
					return nil
				case event := <-eventsChan:
					// Note that event is the *first* event that came in during this throttling interval -- if we're hitting our throttle, we may have dropped events.
					// This is fine, because we don't treat different event types differently.
					// But if we do in the future, we'll need to track more information about the dropped events.
					conf, ingressStatuses := p.loadConfiguration(ctxLog)

					confHash, err := hashstructure.Hash(conf, nil)
					switch {
					case err != nil:
						logger.Error().Msg("Unable to hash the configuration")
					case p.lastConfiguration.Get() == confHash:
						logger.Debug().Msgf("Skipping Kubernetes event kind %T", event)
					default:
						p.lastConfiguration.Set(confHash)
						configurationChan <- dynamic.Message{
							ProviderName:  providerName,
							Configuration: conf,
						}
						time.Sleep(5 * time.Second) // Wait for the routes to be updated before updating ingress
						// status. Not having this can lead to conformance tests failing intermittently as the routes
						// are queried as soon as the status is set to ready.
						for _, ingress := range ingressStatuses {
							if err := p.updateKnativeIngressStatus(ingress); err != nil {
								logger.Error().Err(err).Msgf("Error updating status for Ingress %s/%s", ingress.Namespace, ingress.Name)
							}
						}
					}

					// If we're throttling,
					// we sleep here for the throttle duration to enforce that we don't refresh faster than our throttle.
					// time.Sleep returns immediately if p.ThrottleDuration is 0 (no throttle).
					time.Sleep(throttleDuration)
				}
			}
		}

		notify := func(err error, time time.Duration) {
			logger.Error().Msgf("Provider connection error: %v; retrying in %s", err, time)
		}
		err := backoff.RetryNotify(safe.OperationWithRecover(operation), backoff.WithContext(job.NewBackOff(backoff.NewExponentialBackOff()), ctxPool), notify)
		if err != nil {
			logger.Error().Msgf("Cannot connect to Provider: %v", err)
		}
	})

	return nil
}

func (p *Provider) newK8sClient(ctx context.Context) (*clientWrapper, error) {
	logger := log.Ctx(ctx).With().Logger()

	_, err := labels.Parse(p.LabelSelector)
	if err != nil {
		return nil, fmt.Errorf("parsing label selector: %q", p.LabelSelector)
	}
	logger.Info().Msgf("Label selector is: %q", p.LabelSelector)

	withEndpoint := ""
	if p.Endpoint != "" {
		withEndpoint = fmt.Sprintf(" with endpoint %s", p.Endpoint)
	}

	var client *clientWrapper
	switch {
	case os.Getenv("KUBERNETES_SERVICE_HOST") != "" && os.Getenv("KUBERNETES_SERVICE_PORT") != "":
		logger.Info().Msgf("Creating in-cluster Provider client%s", withEndpoint)
		client, err = newInClusterClient(p.Endpoint)
	case os.Getenv("KUBECONFIG") != "":
		logger.Info().Msgf("Creating cluster-external Provider client from KUBECONFIG %s", os.Getenv("KUBECONFIG"))
		client, err = newExternalClusterClientFromFile(os.Getenv("KUBECONFIG"))
	default:
		logger.Info().Msgf("Creating cluster-external Provider client%s", withEndpoint)
		client, err = newExternalClusterClient(p.Endpoint, p.Token, p.CertAuthFilePath)
	}
	if err != nil {
		return nil, err
	}

	client.labelSelector = p.LabelSelector
	return client, nil
}

func (p *Provider) loadConfiguration(ctx context.Context) (*dynamic.Configuration, []*knativenetworkingv1alpha1.Ingress) {
	conf := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers:           map[string]*dynamic.Router{},
			Middlewares:       map[string]*dynamic.Middleware{},
			Services:          map[string]*dynamic.Service{},
			ServersTransports: map[string]*dynamic.ServersTransport{},
		},
	}

	var ingressStatuses []*knativenetworkingv1alpha1.Ingress

	uniqCerts := make(map[string]*tls.CertAndStores)
	for _, ingress := range p.k8sClient.ListIngresses() {
		logger := log.Ctx(ctx).With().
			Str("ingress", ingress.Name).
			Str("namespace", ingress.Namespace).
			Logger()

		if ingress.Annotations[knativenetworking.IngressClassAnnotationKey] != traefikIngressClassName {
			logger.Debug().Msgf("Skipping Ingress %s/%s", ingress.Namespace, ingress.Name)
			continue
		}

		if err := p.loadCertificates(ctx, ingress, uniqCerts); err != nil {
			logger.Error().Err(err).Msg("Error loading TLS certificates")
			continue
		}

		ingressName := getIngressName(ingress)

		serviceKey, err := makeServiceKey(ingress.Namespace, ingressName)
		if err != nil {
			logger.Error().Err(err).Send()
			continue
		}

		serviceName := provider.Normalize(makeID(ingress.Namespace, serviceKey))
		knativeResult := p.buildKnativeService(ctx, ingress, conf.HTTP.Middlewares, conf.HTTP.Services, serviceName)

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
			if ingress.Spec.TLS != nil {
				r.TLS = &dynamic.RouterTLSConfig{
					CertResolver: "default", // setting to default as we will only have secretName for KNative's.
				}
			}
			conf.HTTP.Routers[provider.Normalize(result.ServiceKey)] = r
			ingressStatuses = append(ingressStatuses, ingress)
		}
	}

	if len(uniqCerts) > 0 {
		conf.TLS = &dynamic.TLSConfiguration{
			Certificates: slices.Collect(maps.Values(uniqCerts)),
		}
	}

	return conf, ingressStatuses
}

// loadCertificates loads the TLS certificates for the given Knative Ingress.
// This method mutates the uniqCerts map to add the loaded certificates.
func (p *Provider) loadCertificates(ctx context.Context, ingress *knativenetworkingv1alpha1.Ingress, uniqCerts map[string]*tls.CertAndStores) error {
	for _, t := range ingress.Spec.TLS {
		// TODO: maybe this could be allowed with an allowCrossNamespace option in the future.
		if t.SecretNamespace != ingress.Namespace {
			log.Ctx(ctx).Debug().Msg("TLS secret namespace has to be the same as the Ingress one")
			continue
		}

		key := ingress.Namespace + "-" + t.SecretName

		// TODO: as specified in the GoDoc we should validate that the certificates contain the configured Hosts.
		if _, exists := uniqCerts[key]; !exists {
			cert, err := p.loadCertificate(ingress.Namespace, t.SecretName)
			if err != nil {
				return fmt.Errorf("getting certificate: %w", err)
			}
			uniqCerts[key] = &tls.CertAndStores{Certificate: cert}
		}
	}

	return nil
}

func (p *Provider) loadCertificate(namespace, secretName string) (tls.Certificate, error) {
	secret, err := p.k8sClient.GetSecret(namespace, secretName)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("getting secret %s/%s: %w", namespace, secretName, err)
	}

	certBytes, hasCert := secret.Data[corev1.TLSCertKey]
	keyBytes, hasKey := secret.Data[corev1.TLSPrivateKeyKey]

	if (!hasCert || len(certBytes) == 0) || (!hasKey || len(keyBytes) == 0) {
		return tls.Certificate{}, errors.New("secret does not contain a keypair")
	}

	return tls.Certificate{
		CertFile: types.FileOrContent(certBytes),
		KeyFile:  types.FileOrContent(keyBytes),
	}, nil
}

func (p *Provider) createKnativeLoadBalancerServerHTTP(namespace string, service traefikv1alpha1.Service) (*dynamic.Service, error) {
	servers, err := p.loadKnativeServers(namespace, service)
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

func (p *Provider) loadKnativeServers(namespace string, svc traefikv1alpha1.Service) ([]dynamic.Server, error) {
	logger := log.With().Logger()

	serverlessservice, exists, err := p.k8sClient.GetServerlessService(namespace, svc.Name)
	if err != nil {
		logger.Info().Msgf("Unable to find serverlessservice, trying to find service %s/%s", namespace, svc.Name)
	}

	serviceName := svc.Name
	if exists {
		serviceName = serverlessservice.Status.ServiceName
	}

	service, exists, err := p.k8sClient.GetService(namespace, serviceName)
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

type serviceResult struct {
	ServiceKey string
	Hosts      []string
	Middleware []string
	Path       string
	Visibility knativenetworkingv1alpha1.IngressVisibility
	Err        error
}

func (p *Provider) buildKnativeService(ctx context.Context, ingress *knativenetworkingv1alpha1.Ingress, middleware map[string]*dynamic.Middleware, conf map[string]*dynamic.Service, serviceName string) []*serviceResult {
	logger := log.Ctx(ctx).With().
		Str("ingressknative", ingress.Name).
		Str("service", serviceName).
		Str("namespace", ingress.Namespace).
		Logger()

	var results []*serviceResult
	for ruleIndex, route := range ingress.Spec.Rules {
		if route.HTTP == nil {
			logger.Warn().Msgf("No HTTP rule defined for Knative service %s", ingress.Name)
			continue
		}

		for pathIndex, pathroute := range route.HTTP.Paths {
			var tagServiceName string
			headers := p.buildHeaders(middleware, serviceName, ruleIndex, pathIndex, pathroute.AppendHeaders)
			path := pathroute.Path

			for _, service := range pathroute.Splits {
				balancerServerHTTP, err := p.createKnativeLoadBalancerServerHTTP(service.ServiceNamespace, traefikv1alpha1.Service{
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
			results = append(results, &serviceResult{tagServiceName, route.Hosts, headers, path, route.Visibility, nil})
		}
	}
	return results
}

func (p *Provider) buildHeaders(middleware map[string]*dynamic.Middleware, serviceName string, ruleIndex, pathIndex int, appendHeaders map[string]string) []string {
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

func (p *Provider) updateKnativeIngressStatus(ingress *knativenetworkingv1alpha1.Ingress) error {
	log.Ctx(context.Background()).Debug().Msgf("Updating status for Ingress %s/%s", ingress.Namespace, ingress.Name)
	if ingress.GetStatus() == nil ||
		!ingress.GetStatus().GetCondition(knativenetworkingv1alpha1.IngressConditionNetworkConfigured).IsTrue() ||
		ingress.GetGeneration() != ingress.GetStatus().ObservedGeneration {
		ingress.Status.MarkLoadBalancerReady(
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

		ingress.Status.MarkNetworkConfigured()
		ingress.Status.ObservedGeneration = ingress.GetGeneration()
		return p.k8sClient.UpdateIngressStatus(ingress)
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

func getIngressName(ingress *knativenetworkingv1alpha1.Ingress) string {
	if len(ingress.Name) == 0 {
		return ingress.GenerateName
	}
	return ingress.Name
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

func makeServiceKey(rule, ingressName string) (string, error) {
	h := sha256.New()
	if _, err := h.Write([]byte(rule)); err != nil {
		return "", err
	}

	key := fmt.Sprintf("%s-%.10x", ingressName, h.Sum(nil))
	return key, nil
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

func throttleEvents(ctx context.Context, throttleDuration time.Duration, pool *safe.Pool, eventsChan <-chan interface{}) chan interface{} {
	logger := log.Ctx(ctx).With().Logger()
	if throttleDuration == 0 {
		return nil
	}
	// Create a buffered channel to hold the pending event (if we're delaying processing the event due to throttling)
	eventsChanBuffered := make(chan interface{}, 1)

	// Run a goroutine that reads events from eventChan and does a non-blocking write to pendingEvent.
	// This guarantees that writing to eventChan will never block,
	// and that pendingEvent will have something in it if there's been an event since we read from that channel.
	pool.GoCtx(func(ctxPool context.Context) {
		for {
			select {
			case <-ctxPool.Done():
				return
			case nextEvent := <-eventsChan:
				select {
				case eventsChanBuffered <- nextEvent:
				default:
					// We already have an event in eventsChanBuffered, so we'll do a refresh as soon as our throttle allows us to.
					// It's fine to drop the event and keep whatever's in the buffer -- we don't do different things for different events
					logger.Debug().Msgf("Dropping event kind %T due to throttling", nextEvent)
				}
			}
		}
	})

	return eventsChanBuffered
}
