package ingress

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"regexp"
	"slices"
	"sort"
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
	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/k8s"
	"github.com/traefik/traefik/v3/pkg/safe"
	"github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	annotationKubernetesIngressClass     = "kubernetes.io/ingress.class"
	traefikDefaultIngressClass           = "traefik"
	traefikDefaultIngressClassController = "traefik.io/ingress-controller"
	defaultPathMatcher                   = "PathPrefix"
)

// Provider holds configurations of the provider.
type Provider struct {
	Endpoint                  string              `description:"Kubernetes server endpoint (required for external cluster client)." json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Token                     types.FileOrContent `description:"Kubernetes bearer token (not needed for in-cluster client). It accepts either a token value or a file path to the token." json:"token,omitempty" toml:"token,omitempty" yaml:"token,omitempty" loggable:"false"`
	CertAuthFilePath          string              `description:"Kubernetes certificate authority file path (not needed for in-cluster client)." json:"certAuthFilePath,omitempty" toml:"certAuthFilePath,omitempty" yaml:"certAuthFilePath,omitempty"`
	Namespaces                []string            `description:"Kubernetes namespaces." json:"namespaces,omitempty" toml:"namespaces,omitempty" yaml:"namespaces,omitempty" export:"true"`
	LabelSelector             string              `description:"Kubernetes Ingress label selector to use." json:"labelSelector,omitempty" toml:"labelSelector,omitempty" yaml:"labelSelector,omitempty" export:"true"`
	IngressClass              string              `description:"Value of kubernetes.io/ingress.class annotation or IngressClass name to watch for." json:"ingressClass,omitempty" toml:"ingressClass,omitempty" yaml:"ingressClass,omitempty" export:"true"`
	IngressEndpoint           *EndpointIngress    `description:"Kubernetes Ingress Endpoint." json:"ingressEndpoint,omitempty" toml:"ingressEndpoint,omitempty" yaml:"ingressEndpoint,omitempty" export:"true"`
	ThrottleDuration          ptypes.Duration     `description:"Ingress refresh throttle duration" json:"throttleDuration,omitempty" toml:"throttleDuration,omitempty" yaml:"throttleDuration,omitempty" export:"true"`
	AllowEmptyServices        bool                `description:"Allow creation of services without endpoints." json:"allowEmptyServices,omitempty" toml:"allowEmptyServices,omitempty" yaml:"allowEmptyServices,omitempty" export:"true"`
	AllowExternalNameServices bool                `description:"Allow ExternalName services." json:"allowExternalNameServices,omitempty" toml:"allowExternalNameServices,omitempty" yaml:"allowExternalNameServices,omitempty" export:"true"`
	DisableIngressClassLookup bool                `description:"Disables the lookup of IngressClasses." json:"disableIngressClassLookup,omitempty" toml:"disableIngressClassLookup,omitempty" yaml:"disableIngressClassLookup,omitempty" export:"true"`
	NativeLBByDefault         bool                `description:"Defines whether to use Native Kubernetes load-balancing mode by default." json:"nativeLBByDefault,omitempty" toml:"nativeLBByDefault,omitempty" yaml:"nativeLBByDefault,omitempty" export:"true"`

	lastConfiguration safe.Safe

	routerTransform k8s.RouterTransform
}

func (p *Provider) SetRouterTransform(routerTransform k8s.RouterTransform) {
	p.routerTransform = routerTransform
}

func (p *Provider) applyRouterTransform(ctx context.Context, rt *dynamic.Router, ingress *netv1.Ingress) {
	if p.routerTransform == nil {
		return
	}

	err := p.routerTransform.Apply(ctx, rt, ingress)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Apply router transform")
	}
}

// EndpointIngress holds the endpoint information for the Kubernetes provider.
type EndpointIngress struct {
	IP               string `description:"IP used for Kubernetes Ingress endpoints." json:"ip,omitempty" toml:"ip,omitempty" yaml:"ip,omitempty"`
	Hostname         string `description:"Hostname used for Kubernetes Ingress endpoints." json:"hostname,omitempty" toml:"hostname,omitempty" yaml:"hostname,omitempty"`
	PublishedService string `description:"Published Kubernetes Service to copy status from." json:"publishedService,omitempty" toml:"publishedService,omitempty" yaml:"publishedService,omitempty"`
}

func (p *Provider) newK8sClient(ctx context.Context) (*clientWrapper, error) {
	_, err := labels.Parse(p.LabelSelector)
	if err != nil {
		return nil, fmt.Errorf("invalid ingress label selector: %q", p.LabelSelector)
	}

	logger := log.Ctx(ctx)

	logger.Info().Msgf("ingress label selector is: %q", p.LabelSelector)

	withEndpoint := ""
	if p.Endpoint != "" {
		withEndpoint = fmt.Sprintf(" with endpoint %v", p.Endpoint)
	}

	var cl *clientWrapper
	switch {
	case os.Getenv("KUBERNETES_SERVICE_HOST") != "" && os.Getenv("KUBERNETES_SERVICE_PORT") != "":
		logger.Info().Msgf("Creating in-cluster Provider client%s", withEndpoint)
		cl, err = newInClusterClient(p.Endpoint)
	case os.Getenv("KUBECONFIG") != "":
		logger.Info().Msgf("Creating cluster-external Provider client from KUBECONFIG %s", os.Getenv("KUBECONFIG"))
		cl, err = newExternalClusterClientFromFile(os.Getenv("KUBECONFIG"))
	default:
		logger.Info().Msgf("Creating cluster-external Provider client%s", withEndpoint)
		cl, err = newExternalClusterClient(p.Endpoint, p.CertAuthFilePath, p.Token)
	}

	if err != nil {
		return nil, err
	}

	cl.ingressLabelSelector = p.LabelSelector
	cl.disableIngressClassInformer = p.DisableIngressClassLookup
	return cl, nil
}

// Init the provider.
func (p *Provider) Init() error {
	return nil
}

// Provide allows the k8s provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	logger := log.With().Str(logs.ProviderName, "kubernetes").Logger()
	ctxLog := logger.WithContext(context.Background())

	k8sClient, err := p.newK8sClient(ctxLog)
	if err != nil {
		return err
	}

	if p.AllowExternalNameServices {
		logger.Warn().Msg("ExternalName service loading is enabled, please ensure that this is expected (see AllowExternalNameServices option)")
	}

	pool.GoCtx(func(ctxPool context.Context) {
		operation := func() error {
			eventsChan, err := k8sClient.WatchAll(p.Namespaces, ctxPool.Done())
			if err != nil {
				logger.Error().Err(err).Msg("Error watching kubernetes events")
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
					// Note that event is the *first* event that came in during this
					// throttling interval -- if we're hitting our throttle, we may have
					// dropped events. This is fine, because we don't treat different
					// event types differently. But if we do in the future, we'll need to
					// track more information about the dropped events.
					conf := p.loadConfigurationFromIngresses(ctxLog, k8sClient)

					confHash, err := hashstructure.Hash(conf, nil)
					switch {
					case err != nil:
						logger.Error().Msg("Unable to hash the configuration")
					case p.lastConfiguration.Get() == confHash:
						logger.Debug().Msgf("Skipping Kubernetes event kind %T", event)
					default:
						p.lastConfiguration.Set(confHash)
						configurationChan <- dynamic.Message{
							ProviderName:  "kubernetes",
							Configuration: conf,
						}
					}

					// If we're throttling, we sleep here for the throttle duration to
					// enforce that we don't refresh faster than our throttle. time.Sleep
					// returns immediately if p.ThrottleDuration is 0 (no throttle).
					time.Sleep(throttleDuration)
				}
			}
		}

		notify := func(err error, time time.Duration) {
			logger.Error().Err(err).Msgf("Provider error, retrying in %s", time)
		}

		err := backoff.RetryNotify(safe.OperationWithRecover(operation), backoff.WithContext(job.NewBackOff(backoff.NewExponentialBackOff()), ctxPool), notify)
		if err != nil {
			logger.Error().Err(err).Msg("Cannot retrieve data")
		}
	})

	return nil
}

func (p *Provider) loadConfigurationFromIngresses(ctx context.Context, client Client) *dynamic.Configuration {
	conf := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers:     map[string]*dynamic.Router{},
			Middlewares: map[string]*dynamic.Middleware{},
			Services:    map[string]*dynamic.Service{},
		},
		TCP: &dynamic.TCPConfiguration{},
	}

	var ingressClasses []*netv1.IngressClass

	if !p.DisableIngressClassLookup {
		ics, err := client.GetIngressClasses()
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("Failed to list ingress classes")
		}

		if p.IngressClass != "" {
			ingressClasses = filterIngressClassByName(p.IngressClass, ics)
		} else {
			ingressClasses = ics
		}
	}

	ingresses := client.GetIngresses()

	certConfigs := make(map[string]*tls.CertAndStores)
	for _, ingress := range ingresses {
		logger := log.Ctx(ctx).With().Str("ingress", ingress.Name).Str("namespace", ingress.Namespace).Logger()
		ctx = logger.WithContext(ctx)

		if !p.shouldProcessIngress(ingress, ingressClasses) {
			continue
		}

		rtConfig, err := parseRouterConfig(ingress.Annotations)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to parse annotations")
			continue
		}

		err = getCertificates(ctx, ingress, client, certConfigs)
		if err != nil {
			logger.Error().Err(err).Msg("Error configuring TLS")
		}

		if len(ingress.Spec.Rules) == 0 && ingress.Spec.DefaultBackend != nil {
			if _, ok := conf.HTTP.Services["default-backend"]; ok {
				logger.Error().Msg("The default backend already exists.")
				continue
			}

			service, err := p.loadService(client, ingress.Namespace, *ingress.Spec.DefaultBackend)
			if err != nil {
				logger.Error().
					Str("serviceName", ingress.Spec.DefaultBackend.Service.Name).
					Str("servicePort", ingress.Spec.DefaultBackend.Service.Port.String()).
					Err(err).
					Msg("Cannot create service")
				continue
			}

			if len(service.LoadBalancer.Servers) == 0 && !p.AllowEmptyServices {
				logger.Error().
					Str("serviceName", ingress.Spec.DefaultBackend.Service.Name).
					Str("servicePort", ingress.Spec.DefaultBackend.Service.Port.String()).
					Msg("Skipping service: no endpoints found")
				continue
			}

			rt := &dynamic.Router{
				Rule:       "PathPrefix(`/`)",
				RuleSyntax: "v3",
				Priority:   math.MinInt32,
				Service:    "default-backend",
			}

			if rtConfig != nil && rtConfig.Router != nil {
				rt.EntryPoints = rtConfig.Router.EntryPoints
				rt.Middlewares = rtConfig.Router.Middlewares
				rt.TLS = rtConfig.Router.TLS
			}

			p.applyRouterTransform(ctx, rt, ingress)

			conf.HTTP.Routers["default-router"] = rt
			conf.HTTP.Services["default-backend"] = service
		}

		routers := map[string][]*dynamic.Router{}

		for _, rule := range ingress.Spec.Rules {
			if err := p.updateIngressStatus(ingress, client); err != nil {
				logger.Error().Err(err).Msg("Error while updating ingress status")
			}

			if rule.HTTP == nil {
				continue
			}

			for _, pa := range rule.HTTP.Paths {
				service, err := p.loadService(client, ingress.Namespace, pa.Backend)
				if err != nil {
					logger.Error().
						Str("serviceName", pa.Backend.Service.Name).
						Str("servicePort", pa.Backend.Service.Port.String()).
						Err(err).
						Msg("Cannot create service")
					continue
				}

				if len(service.LoadBalancer.Servers) == 0 && !p.AllowEmptyServices {
					logger.Error().
						Str("serviceName", pa.Backend.Service.Name).
						Str("servicePort", pa.Backend.Service.Port.String()).
						Msg("Skipping service: no endpoints found")
					continue
				}

				portString := pa.Backend.Service.Port.Name

				if len(pa.Backend.Service.Port.Name) == 0 {
					portString = strconv.Itoa(int(pa.Backend.Service.Port.Number))
				}

				serviceName := provider.Normalize(ingress.Namespace + "-" + pa.Backend.Service.Name + "-" + portString)
				conf.HTTP.Services[serviceName] = service

				rt := loadRouter(rule, pa, rtConfig, serviceName)

				p.applyRouterTransform(ctx, rt, ingress)

				routerKey := strings.TrimPrefix(provider.Normalize(ingress.Namespace+"-"+ingress.Name+"-"+rule.Host+pa.Path), "-")

				routers[routerKey] = append(routers[routerKey], rt)
			}
		}

		for routerKey, conflictingRouters := range routers {
			if len(conflictingRouters) == 1 {
				conf.HTTP.Routers[routerKey] = conflictingRouters[0]
				continue
			}

			logger.Debug().Msgf("Multiple routers are defined with the same key %q, generating hashes to avoid conflicts", routerKey)

			for _, router := range conflictingRouters {
				key, err := makeRouterKeyWithHash(routerKey, router.Rule)
				if err != nil {
					logger.Error().Err(err).Send()
					continue
				}

				conf.HTTP.Routers[key] = router
			}
		}
	}

	certs := getTLSConfig(certConfigs)
	if len(certs) > 0 {
		conf.TLS = &dynamic.TLSConfiguration{
			Certificates: certs,
		}
	}

	return conf
}

func (p *Provider) updateIngressStatus(ing *netv1.Ingress, k8sClient Client) error {
	// Only process if an EndpointIngress has been configured.
	if p.IngressEndpoint == nil {
		return nil
	}

	if len(p.IngressEndpoint.PublishedService) == 0 {
		if len(p.IngressEndpoint.IP) == 0 && len(p.IngressEndpoint.Hostname) == 0 {
			return errors.New("publishedService or ip or hostname must be defined")
		}

		return k8sClient.UpdateIngressStatus(ing, []netv1.IngressLoadBalancerIngress{{IP: p.IngressEndpoint.IP, Hostname: p.IngressEndpoint.Hostname}})
	}

	serviceInfo := strings.Split(p.IngressEndpoint.PublishedService, "/")
	if len(serviceInfo) != 2 {
		return fmt.Errorf("invalid publishedService format (expected 'namespace/service' format): %s", p.IngressEndpoint.PublishedService)
	}

	serviceNamespace, serviceName := serviceInfo[0], serviceInfo[1]

	service, exists, err := k8sClient.GetService(serviceNamespace, serviceName)
	if err != nil {
		return fmt.Errorf("cannot get service %s, received error: %w", p.IngressEndpoint.PublishedService, err)
	}

	if exists && service.Status.LoadBalancer.Ingress == nil {
		// service exists, but has no Load Balancer status
		log.Debug().Msgf("Skipping updating Ingress %s/%s due to service %s having no status set", ing.Namespace, ing.Name, p.IngressEndpoint.PublishedService)
		return nil
	}

	if !exists {
		return fmt.Errorf("missing service: %s", p.IngressEndpoint.PublishedService)
	}

	ingresses, err := convertSlice[netv1.IngressLoadBalancerIngress](service.Status.LoadBalancer.Ingress)
	if err != nil {
		return err
	}

	return k8sClient.UpdateIngressStatus(ing, ingresses)
}

func (p *Provider) shouldProcessIngress(ingress *netv1.Ingress, ingressClasses []*netv1.IngressClass) bool {
	// configuration through the new kubernetes ingressClass
	if ingress.Spec.IngressClassName != nil {
		return slices.ContainsFunc(ingressClasses, func(ic *netv1.IngressClass) bool {
			return *ingress.Spec.IngressClassName == ic.ObjectMeta.Name
		})
	}

	return p.IngressClass == ingress.Annotations[annotationKubernetesIngressClass] ||
		len(p.IngressClass) == 0 && ingress.Annotations[annotationKubernetesIngressClass] == traefikDefaultIngressClass
}

func buildHostRule(host string) string {
	if strings.HasPrefix(host, "*.") {
		host = strings.Replace(regexp.QuoteMeta(host), `\*\.`, `[a-zA-Z0-9-]+\.`, 1)
		return fmt.Sprintf("HostRegexp(`^%s$`)", host)
	}

	return fmt.Sprintf("Host(`%s`)", host)
}

func getCertificates(ctx context.Context, ingress *netv1.Ingress, k8sClient Client, tlsConfigs map[string]*tls.CertAndStores) error {
	for _, t := range ingress.Spec.TLS {
		if t.SecretName == "" {
			log.Ctx(ctx).Debug().Msg("Skipping TLS sub-section: No secret name provided")
			continue
		}

		configKey := ingress.Namespace + "-" + t.SecretName
		if _, tlsExists := tlsConfigs[configKey]; !tlsExists {
			secret, exists, err := k8sClient.GetSecret(ingress.Namespace, t.SecretName)
			if err != nil {
				return fmt.Errorf("failed to fetch secret %s/%s: %w", ingress.Namespace, t.SecretName, err)
			}
			if !exists {
				return fmt.Errorf("secret %s/%s does not exist", ingress.Namespace, t.SecretName)
			}

			cert, key, err := getCertificateBlocks(secret, ingress.Namespace, t.SecretName)
			if err != nil {
				return err
			}

			tlsConfigs[configKey] = &tls.CertAndStores{
				Certificate: tls.Certificate{
					CertFile: types.FileOrContent(cert),
					KeyFile:  types.FileOrContent(key),
				},
			}
		}
	}

	return nil
}

func getCertificateBlocks(secret *corev1.Secret, namespace, secretName string) (string, string, error) {
	var missingEntries []string

	tlsCrtData, tlsCrtExists := secret.Data["tls.crt"]
	if !tlsCrtExists {
		missingEntries = append(missingEntries, "tls.crt")
	}

	tlsKeyData, tlsKeyExists := secret.Data["tls.key"]
	if !tlsKeyExists {
		missingEntries = append(missingEntries, "tls.key")
	}

	if len(missingEntries) > 0 {
		return "", "", fmt.Errorf("secret %s/%s is missing the following TLS data entries: %s",
			namespace, secretName, strings.Join(missingEntries, ", "))
	}

	cert := string(tlsCrtData)
	if cert == "" {
		missingEntries = append(missingEntries, "tls.crt")
	}

	key := string(tlsKeyData)
	if key == "" {
		missingEntries = append(missingEntries, "tls.key")
	}

	if len(missingEntries) > 0 {
		return "", "", fmt.Errorf("secret %s/%s contains the following empty TLS data entries: %s",
			namespace, secretName, strings.Join(missingEntries, ", "))
	}

	return cert, key, nil
}

func getTLSConfig(tlsConfigs map[string]*tls.CertAndStores) []*tls.CertAndStores {
	var secretNames []string
	for secretName := range tlsConfigs {
		secretNames = append(secretNames, secretName)
	}
	sort.Strings(secretNames)

	var configs []*tls.CertAndStores
	for _, secretName := range secretNames {
		configs = append(configs, tlsConfigs[secretName])
	}

	return configs
}

func (p *Provider) loadService(client Client, namespace string, backend netv1.IngressBackend) (*dynamic.Service, error) {
	if backend.Resource != nil {
		// https://kubernetes.io/docs/concepts/services-networking/ingress/#resource-backend
		return nil, errors.New("resource backends are not supported")
	}

	if backend.Service == nil {
		return nil, errors.New("missing service definition")
	}

	service, exists, err := client.GetService(namespace, backend.Service.Name)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errors.New("service not found")
	}

	if !p.AllowExternalNameServices && service.Spec.Type == corev1.ServiceTypeExternalName {
		return nil, fmt.Errorf("externalName services not allowed: %s/%s", namespace, backend.Service.Name)
	}

	var portName string
	var portSpec corev1.ServicePort
	var match bool
	for _, p := range service.Spec.Ports {
		if backend.Service.Port.Number == p.Port || (backend.Service.Port.Name == p.Name && len(p.Name) > 0) {
			portName = p.Name
			portSpec = p
			match = true
			break
		}
	}

	if !match {
		return nil, errors.New("service port not found")
	}

	lb := &dynamic.ServersLoadBalancer{}
	lb.SetDefaults()

	svc := &dynamic.Service{LoadBalancer: lb}

	svcConfig, err := parseServiceConfig(service.Annotations)
	if err != nil {
		return nil, err
	}

	nativeLB := p.NativeLBByDefault

	if svcConfig != nil && svcConfig.Service != nil {
		svc.LoadBalancer.Sticky = svcConfig.Service.Sticky

		if svcConfig.Service.PassHostHeader != nil {
			svc.LoadBalancer.PassHostHeader = svcConfig.Service.PassHostHeader
		}

		if svcConfig.Service.ServersTransport != "" {
			svc.LoadBalancer.ServersTransport = svcConfig.Service.ServersTransport
		}

		if svcConfig.Service.NativeLB != nil {
			nativeLB = *svcConfig.Service.NativeLB
		}
	}

	if service.Spec.Type == corev1.ServiceTypeExternalName {
		protocol := getProtocol(portSpec, portSpec.Name, svcConfig)
		hostPort := net.JoinHostPort(service.Spec.ExternalName, strconv.Itoa(int(portSpec.Port)))

		svc.LoadBalancer.Servers = []dynamic.Server{
			{URL: fmt.Sprintf("%s://%s", protocol, hostPort)},
		}

		return svc, nil
	}

	if nativeLB {
		address, err := getNativeServiceAddress(*service, portSpec)
		if err != nil {
			return nil, fmt.Errorf("getting native Kubernetes Service address: %w", err)
		}

		protocol := getProtocol(portSpec, portSpec.Name, svcConfig)
		svc.LoadBalancer.Servers = []dynamic.Server{
			{URL: fmt.Sprintf("%s://%s", protocol, address)},
		}

		return svc, nil
	}

	endpoints, endpointsExists, endpointsErr := client.GetEndpoints(namespace, backend.Service.Name)
	if endpointsErr != nil {
		return nil, endpointsErr
	}

	if !endpointsExists {
		return nil, errors.New("endpoints not found")
	}

	for _, subset := range endpoints.Subsets {
		var port int32
		for _, p := range subset.Ports {
			if portName == p.Name {
				port = p.Port
				break
			}
		}

		if port == 0 {
			continue
		}

		protocol := getProtocol(portSpec, portName, svcConfig)

		for _, addr := range subset.Addresses {
			hostPort := net.JoinHostPort(addr.IP, strconv.Itoa(int(port)))

			svc.LoadBalancer.Servers = append(svc.LoadBalancer.Servers, dynamic.Server{
				URL: fmt.Sprintf("%s://%s", protocol, hostPort),
			})
		}
	}

	return svc, nil
}

func getNativeServiceAddress(service corev1.Service, svcPort corev1.ServicePort) (string, error) {
	if service.Spec.ClusterIP == "None" {
		return "", fmt.Errorf("no clusterIP on headless service: %s/%s", service.Namespace, service.Name)
	}

	if service.Spec.ClusterIP == "" {
		return "", fmt.Errorf("no clusterIP found for service: %s/%s", service.Namespace, service.Name)
	}

	return net.JoinHostPort(service.Spec.ClusterIP, strconv.Itoa(int(svcPort.Port))), nil
}

func getProtocol(portSpec corev1.ServicePort, portName string, svcConfig *ServiceConfig) string {
	if svcConfig != nil && svcConfig.Service != nil && svcConfig.Service.ServersScheme != "" {
		return svcConfig.Service.ServersScheme
	}

	protocol := "http"
	if portSpec.Port == 443 || strings.HasPrefix(portName, "https") {
		protocol = "https"
	}

	return protocol
}

func makeRouterKeyWithHash(key, rule string) (string, error) {
	h := sha256.New()
	if _, err := h.Write([]byte(rule)); err != nil {
		return "", err
	}

	dupKey := fmt.Sprintf("%s-%.10x", key, h.Sum(nil))

	return dupKey, nil
}

func loadRouter(rule netv1.IngressRule, pa netv1.HTTPIngressPath, rtConfig *RouterConfig, serviceName string) *dynamic.Router {
	var rules []string
	if len(rule.Host) > 0 {
		rules = []string{buildHostRule(rule.Host)}
	}

	if len(pa.Path) > 0 {
		matcher := defaultPathMatcher

		if pa.PathType == nil || *pa.PathType == "" || *pa.PathType == netv1.PathTypeImplementationSpecific {
			if rtConfig != nil && rtConfig.Router != nil && rtConfig.Router.PathMatcher != "" {
				matcher = rtConfig.Router.PathMatcher
			}
		} else if *pa.PathType == netv1.PathTypeExact {
			matcher = "Path"
		}

		rules = append(rules, fmt.Sprintf("%s(`%s`)", matcher, pa.Path))
	}

	rt := &dynamic.Router{
		Rule:    strings.Join(rules, " && "),
		Service: serviceName,
	}

	if rtConfig != nil && rtConfig.Router != nil {
		rt.Priority = rtConfig.Router.Priority
		rt.EntryPoints = rtConfig.Router.EntryPoints
		rt.Middlewares = rtConfig.Router.Middlewares

		if rtConfig.Router.TLS != nil {
			rt.TLS = rtConfig.Router.TLS
		}
	}

	return rt
}

func throttleEvents(ctx context.Context, throttleDuration time.Duration, pool *safe.Pool, eventsChan <-chan interface{}) chan interface{} {
	if throttleDuration == 0 {
		return nil
	}

	// Create a buffered channel to hold the pending event (if we're delaying processing the event due to throttling).
	eventsChanBuffered := make(chan interface{}, 1)

	// Run a goroutine that reads events from eventChan and does a
	// non-blocking write to pendingEvent. This guarantees that writing to
	// eventChan will never block, and that pendingEvent will have
	// something in it if there's been an event since we read from that channel.
	pool.GoCtx(func(ctxPool context.Context) {
		for {
			select {
			case <-ctxPool.Done():
				return
			case nextEvent := <-eventsChan:
				select {
				case eventsChanBuffered <- nextEvent:
				default:
					// We already have an event in eventsChanBuffered, so we'll
					// do a refresh as soon as our throttle allows us to. It's fine
					// to drop the event and keep whatever's in the buffer -- we
					// don't do different things for different events.
					log.Ctx(ctx).Debug().Msgf("Dropping event kind %T due to throttling", nextEvent)
				}
			}
		}
	})

	return eventsChanBuffered
}
