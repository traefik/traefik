package knative

import (
	"context"
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
	"github.com/traefik/traefik/v3/pkg/observability/logs"
	"github.com/traefik/traefik/v3/pkg/safe"
	"github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	knativenetworking "knative.dev/networking/pkg/apis/networking"
	knativenetworkingv1alpha1 "knative.dev/networking/pkg/apis/networking/v1alpha1"
	"knative.dev/pkg/network"
)

const (
	providerName            = "knative"
	traefikIngressClassName = "traefik.ingress.networking.knative.dev"
)

// ServiceRef holds a Kubernetes service reference.
type ServiceRef struct {
	Name      string `description:"Name of the Kubernetes service." json:"desc,omitempty" toml:"desc,omitempty" yaml:"desc,omitempty"`
	Namespace string `description:"Namespace of the Kubernetes service." json:"namespace,omitempty" toml:"namespace,omitempty" yaml:"namespace,omitempty"`
}

// Provider holds configurations of the provider.
type Provider struct {
	Endpoint           string          `description:"Kubernetes server endpoint (required for external cluster client)." json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Token              string          `description:"Kubernetes bearer token (not needed for in-cluster client)." json:"token,omitempty" toml:"token,omitempty" yaml:"token,omitempty"`
	CertAuthFilePath   string          `description:"Kubernetes certificate authority file path (not needed for in-cluster client)." json:"certAuthFilePath,omitempty" toml:"certAuthFilePath,omitempty" yaml:"certAuthFilePath,omitempty"`
	Namespaces         []string        `description:"Kubernetes namespaces." json:"namespaces,omitempty" toml:"namespaces,omitempty" yaml:"namespaces,omitempty" export:"true"`
	LabelSelector      string          `description:"Kubernetes label selector to use." json:"labelSelector,omitempty" toml:"labelSelector,omitempty" yaml:"labelSelector,omitempty" export:"true"`
	PublicEntrypoints  []string        `description:"Entrypoint names used to expose the Ingress publicly. If empty an Ingress is exposed on all entrypoints." json:"publicEntrypoints,omitempty" toml:"publicEntrypoints,omitempty" yaml:"publicEntrypoints,omitempty" export:"true"`
	PublicService      ServiceRef      `description:"Kubernetes service used to expose the networking controller publicly." json:"publicService,omitempty" toml:"publicService,omitempty" yaml:"publicService,omitempty" export:"true"`
	PrivateEntrypoints []string        `description:"Entrypoint names used to expose the Ingress privately. If empty local Ingresses are skipped." json:"privateEntrypoints,omitempty" toml:"privateEntrypoints,omitempty" yaml:"privateEntrypoints,omitempty" export:"true"`
	PrivateService     ServiceRef      `description:"Kubernetes service used to expose the networking controller privately." json:"privateService,omitempty" toml:"privateService,omitempty" yaml:"privateService,omitempty" export:"true"`
	ThrottleDuration   ptypes.Duration `description:"Ingress refresh throttle duration" json:"throttleDuration,omitempty" toml:"throttleDuration,omitempty" yaml:"throttleDuration,omitempty"`

	client            *clientWrapper
	lastConfiguration safe.Safe
}

// Init the provider.
func (p *Provider) Init() error {
	logger := log.With().Str(logs.ProviderName, providerName).Logger()

	// Initializes Kubernetes client.
	var err error
	p.client, err = p.newK8sClient(logger.WithContext(context.Background()))
	if err != nil {
		return fmt.Errorf("creating kubernetes client: %w", err)
	}

	return nil
}

// Provide allows the knative provider to provide configurations to traefik using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	logger := log.With().Str(logs.ProviderName, providerName).Logger()
	ctxLog := logger.WithContext(context.Background())

	pool.GoCtx(func(ctxPool context.Context) {
		operation := func() error {
			eventsChan, err := p.client.WatchAll(p.Namespaces, ctxPool.Done())
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
					}

					// If we're throttling,
					// we sleep here for the throttle duration to enforce that we don't refresh faster than our throttle.
					// time.Sleep returns immediately if p.ThrottleDuration is 0 (no throttle).
					time.Sleep(throttleDuration)

					// Updating the ingress status after the throttleDuration allows to wait to make sure that the dynamic conf is updated before updating the status.
					// This is needed for the conformance tests to pass, for example.
					for _, ingress := range ingressStatuses {
						if err := p.updateKnativeIngressStatus(ctxLog, ingress); err != nil {
							logger.Error().Err(err).Msgf("Error updating status for Ingress %s/%s", ingress.Namespace, ingress.Name)
						}
					}
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
			Routers:     make(map[string]*dynamic.Router),
			Middlewares: make(map[string]*dynamic.Middleware),
			Services:    make(map[string]*dynamic.Service),
		},
	}

	var ingressStatuses []*knativenetworkingv1alpha1.Ingress

	uniqCerts := make(map[string]*tls.CertAndStores)
	for _, ingress := range p.client.ListIngresses() {
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

		conf.HTTP = mergeHTTPConfigs(conf.HTTP, p.buildRouters(ctx, ingress))

		// TODO: should we handle configuration errors?
		ingressStatuses = append(ingressStatuses, ingress)
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
	secret, err := p.client.GetSecret(namespace, secretName)
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

func (p *Provider) buildRouters(ctx context.Context, ingress *knativenetworkingv1alpha1.Ingress) *dynamic.HTTPConfiguration {
	logger := log.Ctx(ctx).With().Logger()

	conf := &dynamic.HTTPConfiguration{
		Routers:     make(map[string]*dynamic.Router),
		Middlewares: make(map[string]*dynamic.Middleware),
		Services:    make(map[string]*dynamic.Service),
	}

	for ri, rule := range ingress.Spec.Rules {
		if rule.HTTP == nil {
			logger.Debug().Msgf("No HTTP rule defined for rule %d in Ingress %s", ri, ingress.Name)
			continue
		}

		entrypoints := p.PublicEntrypoints
		if rule.Visibility == knativenetworkingv1alpha1.IngressVisibilityClusterLocal {
			if p.PrivateEntrypoints == nil {
				// Skip route creation as no internal entrypoints are defined for cluster local visibility.
				continue
			}
			entrypoints = p.PrivateEntrypoints
		}

		// TODO: support rewrite host
		for pi, path := range rule.HTTP.Paths {
			routerKey := fmt.Sprintf("%s-%s-rule-%d-path-%d", ingress.Namespace, ingress.Name, ri, pi)
			router := &dynamic.Router{
				EntryPoints: entrypoints,
				Rule:        buildRule(rule.Hosts, path.Headers, path.Path),
				Middlewares: make([]string, 0),
				Service:     routerKey + "-wrr",
			}

			if len(path.AppendHeaders) > 0 {
				midKey := fmt.Sprintf("%s-append-headers", routerKey)

				router.Middlewares = append(router.Middlewares, midKey)
				conf.Middlewares[midKey] = &dynamic.Middleware{
					Headers: &dynamic.Headers{
						CustomRequestHeaders: path.AppendHeaders,
					},
				}
			}

			wrr, services, err := p.buildWeightedRoundRobin(routerKey, path.Splits)
			if err != nil {
				logger.Error().Err(err).Msg("Error building weighted round robin")
				continue
			}

			// TODO: support Ingress#HTTPOption to check if HTTP router should redirect to the HTTPS one.
			conf.Routers[routerKey] = router

			// TODO: at some point we should allow to define a default TLS secret at the provider level to enable TLS with a custom cert when external-domain-tls is disabled.
			//       see https://knative.dev/docs/serving/encryption/external-domain-tls/#manually-obtain-and-renew-certificates
			if len(ingress.Spec.TLS) > 0 {
				conf.Routers[routerKey+"-tls"] = &dynamic.Router{
					EntryPoints: router.EntryPoints,
					Rule:        router.Rule, // TODO: maybe the rule should be a new one containing the TLS hosts injected by Knative.
					Middlewares: router.Middlewares,
					Service:     router.Service,
					TLS:         &dynamic.RouterTLSConfig{},
				}
			}

			conf.Services[routerKey+"-wrr"] = &dynamic.Service{Weighted: wrr}
			for k, v := range services {
				conf.Services[k] = v
			}
		}
	}

	return conf
}

func (p *Provider) buildWeightedRoundRobin(routerKey string, splits []knativenetworkingv1alpha1.IngressBackendSplit) (*dynamic.WeightedRoundRobin, map[string]*dynamic.Service, error) {
	wrr := &dynamic.WeightedRoundRobin{
		Services: make([]dynamic.WRRService, 0),
	}

	services := make(map[string]*dynamic.Service)
	for si, split := range splits {
		serviceKey := fmt.Sprintf("%s-split-%d", routerKey, si)

		var err error
		services[serviceKey], err = p.buildService(split.ServiceNamespace, split.ServiceName, split.ServicePort)
		if err != nil {
			return nil, nil, fmt.Errorf("building service: %w", err)
		}

		// As described in the spec if there is only one split it defaults to 100.
		percent := split.Percent
		if len(splits) == 1 {
			percent = 100
		}

		wrr.Services = append(wrr.Services, dynamic.WRRService{
			Name:    serviceKey,
			Weight:  ptr.To(percent),
			Headers: split.AppendHeaders,
		})
	}

	return wrr, services, nil
}

func (p *Provider) buildService(namespace, serviceName string, port intstr.IntOrString) (*dynamic.Service, error) {
	servers, err := p.buildServers(namespace, serviceName, port)
	if err != nil {
		return nil, fmt.Errorf("building servers: %w", err)
	}

	var lb dynamic.ServersLoadBalancer
	lb.SetDefaults()
	lb.Servers = servers

	return &dynamic.Service{LoadBalancer: &lb}, nil
}

func (p *Provider) buildServers(namespace, serviceName string, port intstr.IntOrString) ([]dynamic.Server, error) {
	service, err := p.client.GetService(namespace, serviceName)
	if err != nil {
		return nil, fmt.Errorf("getting service %s/%s: %w", namespace, serviceName, err)
	}

	var svcPort *corev1.ServicePort
	for _, p := range service.Spec.Ports {
		if p.Name == port.String() || strconv.Itoa(int(p.Port)) == port.String() {
			svcPort = &p
			break
		}
	}
	if svcPort == nil {
		return nil, errors.New("service port not found")
	}

	if service.Spec.ClusterIP == "" {
		return nil, errors.New("service does not have a ClusterIP")
	}

	scheme := "http"
	if svcPort.AppProtocol != nil && *svcPort.AppProtocol == knativenetworking.AppProtocolH2C {
		scheme = "h2c"
	}

	hostPort := net.JoinHostPort(service.Spec.ClusterIP, strconv.Itoa(int(svcPort.Port)))
	return []dynamic.Server{{URL: fmt.Sprintf("%s://%s", scheme, hostPort)}}, nil
}

func (p *Provider) updateKnativeIngressStatus(ctx context.Context, ingress *knativenetworkingv1alpha1.Ingress) error {
	log.Ctx(ctx).Debug().Msgf("Updating status for Ingress %s/%s", ingress.Namespace, ingress.Name)

	var publicLbs []knativenetworkingv1alpha1.LoadBalancerIngressStatus
	if p.PublicService.Name != "" && p.PublicService.Namespace != "" {
		publicLbs = append(publicLbs, knativenetworkingv1alpha1.LoadBalancerIngressStatus{
			DomainInternal: network.GetServiceHostname(p.PublicService.Name, p.PublicService.Namespace),
		})
	}

	var privateLbs []knativenetworkingv1alpha1.LoadBalancerIngressStatus
	if p.PrivateService.Name != "" && p.PrivateService.Namespace != "" {
		privateLbs = append(privateLbs, knativenetworkingv1alpha1.LoadBalancerIngressStatus{
			DomainInternal: network.GetServiceHostname(p.PrivateService.Name, p.PrivateService.Namespace),
		})
	}

	if ingress.GetStatus() == nil || !ingress.GetStatus().GetCondition(knativenetworkingv1alpha1.IngressConditionNetworkConfigured).IsTrue() || ingress.GetGeneration() != ingress.GetStatus().ObservedGeneration {
		ingress.Status.MarkNetworkConfigured()
		ingress.Status.MarkLoadBalancerReady(publicLbs, privateLbs)
		ingress.Status.ObservedGeneration = ingress.GetGeneration()

		return p.client.UpdateIngressStatus(ingress)
	}
	return nil
}

func buildRule(hosts []string, headers map[string]knativenetworkingv1alpha1.HeaderMatch, path string) string {
	var operands []string

	if len(hosts) > 0 {
		var hostRules []string
		for _, host := range hosts {
			hostRules = append(hostRules, fmt.Sprintf("Host(`%v`)", host))
		}
		operands = append(operands, fmt.Sprintf("(%s)", strings.Join(hostRules, " || ")))
	}

	if len(headers) > 0 {
		headerKeys := slices.Collect(maps.Keys(headers))
		slices.Sort(headerKeys)

		var headerRules []string
		for _, key := range headerKeys {
			headerRules = append(headerRules, fmt.Sprintf("Header(`%s`,`%s`)", key, headers[key].Exact))
		}
		operands = append(operands, fmt.Sprintf("(%s)", strings.Join(headerRules, " && ")))
	}

	if len(path) > 0 {
		operands = append(operands, fmt.Sprintf("PathPrefix(`%s`)", path))
	}

	return strings.Join(operands, " && ")
}

func mergeHTTPConfigs(confs ...*dynamic.HTTPConfiguration) *dynamic.HTTPConfiguration {
	conf := &dynamic.HTTPConfiguration{
		Routers:     map[string]*dynamic.Router{},
		Middlewares: map[string]*dynamic.Middleware{},
		Services:    map[string]*dynamic.Service{},
	}

	for _, c := range confs {
		for k, v := range c.Routers {
			conf.Routers[k] = v
		}
		for k, v := range c.Middlewares {
			conf.Middlewares[k] = v
		}
		for k, v := range c.Services {
			conf.Services[k] = v
		}
	}

	return conf
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
