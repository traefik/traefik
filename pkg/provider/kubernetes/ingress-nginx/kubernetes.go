package ingressnginx

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"math"
	"net"
	"os"
	"regexp"
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
	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/k8s"
	"github.com/traefik/traefik/v3/pkg/safe"
	"github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/utils/ptr"
)

const (
	providerName = "kubernetesingressnginx"

	annotationIngressClass = "kubernetes.io/ingress.class"

	defaultControllerName  = "k8s.io/ingress-nginx"
	defaultAnnotationValue = "nginx"

	defaultBackendName    = "default-backend"
	defaultBackendTLSName = "default-backend-tls"
)

type backendAddress struct {
	Address string
	Fenced  bool
}

type namedServersTransport struct {
	Name             string
	ServersTransport *dynamic.ServersTransport
}

type certBlocks struct {
	CA          *types.FileOrContent
	Certificate *tls.Certificate
}

// Provider holds configurations of the provider.
type Provider struct {
	Endpoint         string              `description:"Kubernetes server endpoint (required for external cluster client)." json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Token            types.FileOrContent `description:"Kubernetes bearer token (not needed for in-cluster client). It accepts either a token value or a file path to the token." json:"token,omitempty" toml:"token,omitempty" yaml:"token,omitempty" loggable:"false"`
	CertAuthFilePath string              `description:"Kubernetes certificate authority file path (not needed for in-cluster client)." json:"certAuthFilePath,omitempty" toml:"certAuthFilePath,omitempty" yaml:"certAuthFilePath,omitempty"`
	ThrottleDuration ptypes.Duration     `description:"Ingress refresh throttle duration." json:"throttleDuration,omitempty" toml:"throttleDuration,omitempty" yaml:"throttleDuration,omitempty" export:"true"`

	WatchNamespace         string `description:"Namespace the controller watches for updates to Kubernetes objects. All namespaces are watched if this parameter is left empty." json:"watchNamespace,omitempty" toml:"watchNamespace,omitempty" yaml:"watchNamespace,omitempty" export:"true"`
	WatchNamespaceSelector string `description:"Selector selects namespaces the controller watches for updates to Kubernetes objects." json:"watchNamespaceSelector,omitempty" toml:"watchNamespaceSelector,omitempty" yaml:"watchNamespaceSelector,omitempty" export:"true"`

	IngressClass             string `description:"Name of the ingress class this controller satisfies." json:"ingressClass,omitempty" toml:"ingressClass,omitempty" yaml:"ingressClass,omitempty" export:"true"`
	ControllerClass          string `description:"Ingress Class Controller value this controller satisfies." json:"controllerClass,omitempty" toml:"controllerClass,omitempty" yaml:"controllerClass,omitempty" export:"true"`
	WatchIngressWithoutClass bool   `description:"Define if Ingress Controller should also watch for Ingresses without an IngressClass or the annotation specified." json:"watchIngressWithoutClass,omitempty" toml:"watchIngressWithoutClass,omitempty" yaml:"watchIngressWithoutClass,omitempty" export:"true"`
	IngressClassByName       bool   `description:"Define if Ingress Controller should watch for Ingress Class by Name together with Controller Class." json:"ingressClassByName,omitempty" toml:"ingressClassByName,omitempty" yaml:"ingressClassByName,omitempty" export:"true"`

	// TODO: support report-node-internal-ip-address and update-status.
	PublishService       string   `description:"Service fronting the Ingress controller. Takes the form 'namespace/name'." json:"publishService,omitempty" toml:"publishService,omitempty" yaml:"publishService,omitempty" export:"true"`
	PublishStatusAddress []string `description:"Customized address (or addresses, separated by comma) to set as the load-balancer status of Ingress objects this controller satisfies." json:"publishStatusAddress,omitempty" toml:"publishStatusAddress,omitempty" yaml:"publishStatusAddress,omitempty"`

	DefaultBackendService  string `description:"Service used to serve HTTP requests not matching any known server name (catch-all). Takes the form 'namespace/name'." json:"defaultBackendService,omitempty" toml:"defaultBackendService,omitempty" yaml:"defaultBackendService,omitempty" export:"true"`
	DisableSvcExternalName bool   `description:"Disable support for Services of type ExternalName." json:"disableSvcExternalName,omitempty" toml:"disableSvcExternalName,omitempty" yaml:"disableSvcExternalName,omitempty" export:"true"`

	defaultBackendServiceNamespace string
	defaultBackendServiceName      string

	k8sClient         *clientWrapper
	lastConfiguration safe.Safe
}

func (p *Provider) SetDefaults() {
	p.IngressClass = defaultAnnotationValue
	p.ControllerClass = defaultControllerName
}

// Init the provider.
func (p *Provider) Init() error {
	// Validates and parses the default backend configuration.
	if p.DefaultBackendService != "" {
		parts := strings.Split(p.DefaultBackendService, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid default backend service format: %s, expected 'namespace/name'", p.DefaultBackendService)
		}
		p.defaultBackendServiceNamespace = parts[0]
		p.defaultBackendServiceName = parts[1]
	}

	// Initializes Kubernetes client.
	var err error
	p.k8sClient, err = p.newK8sClient()
	if err != nil {
		return fmt.Errorf("creating kubernetes client: %w", err)
	}

	return nil
}

// Provide allows the k8s provider to provide configurations to traefik using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	logger := log.With().Str(logs.ProviderName, providerName).Logger()
	ctxLog := logger.WithContext(context.Background())

	pool.GoCtx(func(ctxPool context.Context) {
		operation := func() error {
			eventsChan, err := p.k8sClient.WatchAll(ctxPool, p.WatchNamespace, p.WatchNamespaceSelector)
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
					conf := p.loadConfiguration(ctxLog)

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

func (p *Provider) newK8sClient() (*clientWrapper, error) {
	withEndpoint := ""
	if p.Endpoint != "" {
		withEndpoint = fmt.Sprintf(" with endpoint %v", p.Endpoint)
	}

	switch {
	case os.Getenv("KUBERNETES_SERVICE_HOST") != "" && os.Getenv("KUBERNETES_SERVICE_PORT") != "":
		log.Info().Msgf("Creating in-cluster Provider client%s", withEndpoint)
		return newInClusterClient(p.Endpoint)
	case os.Getenv("KUBECONFIG") != "":
		log.Info().Msgf("Creating cluster-external Provider client from KUBECONFIG %s", os.Getenv("KUBECONFIG"))
		return newExternalClusterClientFromFile(os.Getenv("KUBECONFIG"))
	default:
		log.Info().Msgf("Creating cluster-external Provider client%s", withEndpoint)
		return newExternalClusterClient(p.Endpoint, p.CertAuthFilePath, p.Token)
	}
}

func (p *Provider) loadConfiguration(ctx context.Context) *dynamic.Configuration {
	conf := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers:           map[string]*dynamic.Router{},
			Middlewares:       map[string]*dynamic.Middleware{},
			Services:          map[string]*dynamic.Service{},
			ServersTransports: map[string]*dynamic.ServersTransport{},
		},
		TCP: &dynamic.TCPConfiguration{
			Routers:  map[string]*dynamic.TCPRouter{},
			Services: map[string]*dynamic.TCPService{},
		},
	}

	// We configure the default backend when it is configured at the provider level.
	if p.defaultBackendServiceNamespace != "" && p.defaultBackendServiceName != "" {
		ib := netv1.IngressBackend{Service: &netv1.IngressServiceBackend{Name: p.defaultBackendServiceName}}
		svc, err := p.buildService(p.defaultBackendServiceNamespace, ib, ingressConfig{})
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("Cannot build default backend service")
			return conf
		}

		// Add the default backend service router to the configuration.
		conf.HTTP.Routers[defaultBackendName] = &dynamic.Router{
			Rule: "PathPrefix(`/`)",
			// "default" stands for the default rule syntax in Traefik v3, i.e. the v3 syntax.
			RuleSyntax: "default",
			Priority:   math.MinInt32,
			Service:    defaultBackendName,
		}

		conf.HTTP.Routers[defaultBackendTLSName] = &dynamic.Router{
			Rule: "PathPrefix(`/`)",
			// "default" stands for the default rule syntax in Traefik v3, i.e. the v3 syntax.
			RuleSyntax: "default",
			Priority:   math.MinInt32,
			Service:    defaultBackendName,
			TLS:        &dynamic.RouterTLSConfig{},
		}

		conf.HTTP.Services[defaultBackendName] = svc
	}

	var ingressClasses []*netv1.IngressClass
	ics, err := p.k8sClient.ListIngressClasses()
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("Failed to list ingress classes")
	}
	ingressClasses = filterIngressClass(ics, p.IngressClassByName, p.IngressClass, p.ControllerClass)

	ingresses := p.k8sClient.ListIngresses()

	uniqCerts := make(map[string]*tls.CertAndStores)
	for _, ingress := range ingresses {
		logger := log.Ctx(ctx).With().Str("ingress", ingress.Name).Str("namespace", ingress.Namespace).Logger()
		ctxIngress := logger.WithContext(ctx)

		if !p.shouldProcessIngress(ingress, ingressClasses) {
			continue
		}

		ingressConfig, err := parseIngressConfig(ingress)
		if err != nil {
			logger.Error().Err(err).Msg("Error parsing ingress configuration")
			continue
		}

		if err := p.updateIngressStatus(ingress); err != nil {
			logger.Error().Err(err).Msg("Error while updating ingress status")
		}

		var hasTLS bool
		if len(ingress.Spec.TLS) > 0 {
			hasTLS = true
			if err := p.loadCertificates(ctxIngress, ingress, uniqCerts); err != nil {
				logger.Error().Err(err).Msg("Error configuring TLS")
				continue
			}
		}

		namedServersTransport, err := p.buildServersTransport(ingress.Namespace, ingress.Name, ingressConfig)
		if err != nil {
			logger.Error().Err(err).Msg("Ignoring Ingress cannot create proxy SSL configuration")
			continue
		}

		var defaultBackendService *dynamic.Service
		if ingress.Spec.DefaultBackend != nil && ingress.Spec.DefaultBackend.Service != nil {
			var err error
			defaultBackendService, err = p.buildService(ingress.Namespace, *ingress.Spec.DefaultBackend, ingressConfig)
			if err != nil {
				logger.Error().
					Str("serviceName", ingress.Spec.DefaultBackend.Service.Name).
					Str("servicePort", ingress.Spec.DefaultBackend.Service.Port.String()).
					Err(err).
					Msg("Cannot create default backend service")
			}
		}

		if defaultBackendService != nil && len(ingress.Spec.Rules) == 0 {
			rt := &dynamic.Router{
				Rule: "PathPrefix(`/`)",
				// "default" stands for the default rule syntax in Traefik v3, i.e. the v3 syntax.
				RuleSyntax: "default",
				Priority:   math.MinInt32,
				Service:    defaultBackendName,
			}

			if err := p.applyMiddlewares(ingress.Namespace, defaultBackendName, ingressConfig, hasTLS, rt, conf); err != nil {
				logger.Error().Err(err).Msg("Error applying middlewares")
			}

			conf.HTTP.Routers[defaultBackendName] = rt

			rtTLS := &dynamic.Router{
				Rule: "PathPrefix(`/`)",
				// "default" stands for the default rule syntax in Traefik v3, i.e. the v3 syntax.
				RuleSyntax: "default",
				Priority:   math.MinInt32,
				Service:    defaultBackendName,
				TLS:        &dynamic.RouterTLSConfig{},
			}

			if err := p.applyMiddlewares(ingress.Namespace, defaultBackendTLSName, ingressConfig, false, rtTLS, conf); err != nil {
				logger.Error().Err(err).Msg("Error applying middlewares")
			}

			conf.HTTP.Routers[defaultBackendTLSName] = rtTLS

			if namedServersTransport != nil && defaultBackendService.LoadBalancer != nil {
				defaultBackendService.LoadBalancer.ServersTransport = namedServersTransport.Name
				conf.HTTP.ServersTransports[namedServersTransport.Name] = namedServersTransport.ServersTransport
			}
			conf.HTTP.Services[defaultBackendName] = defaultBackendService
		}

		for ri, rule := range ingress.Spec.Rules {
			if ptr.Deref(ingressConfig.SSLPassthrough, false) {
				if rule.Host == "" {
					logger.Error().Err(err).Msg("Cannot process ssl-passthrough for rule without host")
					continue
				}

				var backend *netv1.IngressBackend
				if rule.HTTP != nil {
					for _, path := range rule.HTTP.Paths {
						if path.Path == "/" {
							backend = &path.Backend
							break
						}
					}
				} else if ingress.Spec.DefaultBackend != nil {
					// Passthrough with the default backend if no HTTP section.
					backend = ingress.Spec.DefaultBackend
				}

				if backend == nil {
					logger.Error().Msgf("No backend found for ssl-passthrough for rule with host %q", rule.Host)
					continue
				}

				service, err := p.buildPassthroughService(ingress.Namespace, *backend, ingressConfig)
				if err != nil {
					logger.Error().Err(err).Msgf("Cannot create passthrough service for %s", backend.Service.Name)
					continue
				}

				port := backend.Service.Port.Name
				if len(backend.Service.Port.Name) == 0 {
					port = strconv.Itoa(int(backend.Service.Port.Number))
				}

				serviceName := provider.Normalize(ingress.Namespace + "-" + backend.Service.Name + "-" + port)
				conf.TCP.Services[serviceName] = service

				routerKey := strings.TrimPrefix(provider.Normalize(ingress.Namespace+"-"+ingress.Name+"-"+rule.Host), "-")

				conf.TCP.Routers[routerKey] = &dynamic.TCPRouter{
					Rule: fmt.Sprintf("HostSNI(`%s`)", rule.Host),
					// "default" stands for the default rule syntax in Traefik v3, i.e. the v3 syntax.
					RuleSyntax: "default",
					Service:    serviceName,
					TLS: &dynamic.RouterTCPTLSConfig{
						Passthrough: true,
					},
				}

				continue
			}

			if defaultBackendService != nil && rule.Host != "" {
				key := provider.Normalize(ingress.Namespace + "-" + ingress.Name + "-default-backend")

				rt := &dynamic.Router{
					Rule: buildHostRule(rule.Host),
					// "default" stands for the default rule syntax in Traefik v3, i.e. the v3 syntax.
					RuleSyntax: "default",
					Service:    key,
				}

				if err := p.applyMiddlewares(ingress.Namespace, key, ingressConfig, hasTLS, rt, conf); err != nil {
					logger.Error().Err(err).Msg("Error applying middlewares")
				}

				conf.HTTP.Routers[key] = rt

				rtTLS := &dynamic.Router{
					Rule: buildHostRule(rule.Host),
					// "default" stands for the default rule syntax in Traefik v3, i.e. the v3 syntax.
					RuleSyntax: "default",
					Service:    key,
					TLS:        &dynamic.RouterTLSConfig{},
				}

				if err := p.applyMiddlewares(ingress.Namespace, key+"-tls", ingressConfig, false, rtTLS, conf); err != nil {
					logger.Error().Err(err).Msg("Error applying middlewares")
				}

				conf.HTTP.Routers[key+"-tls"] = rtTLS

				if namedServersTransport != nil && defaultBackendService.LoadBalancer != nil {
					defaultBackendService.LoadBalancer.ServersTransport = namedServersTransport.Name
					conf.HTTP.ServersTransports[namedServersTransport.Name] = namedServersTransport.ServersTransport
				}

				conf.HTTP.Services[key] = defaultBackendService
			}

			if rule.HTTP == nil {
				continue
			}

			for pi, pa := range rule.HTTP.Paths {
				// As NGINX we are ignoring resource backend.
				// An Ingress backend must have se service or a resource definition.
				if pa.Backend.Service == nil {
					logger.Error().Str("path", pa.Path).
						Err(err).Msg("Ignoring path with no service backend")
					continue
				}

				portString := pa.Backend.Service.Port.Name
				if len(pa.Backend.Service.Port.Name) == 0 {
					portString = strconv.Itoa(int(pa.Backend.Service.Port.Number))
				}

				// TODO: if no service, do not add middlewares and 503.
				serviceName := provider.Normalize(ingress.Namespace + "-" + pa.Backend.Service.Name + "-" + portString)

				service, err := p.buildService(ingress.Namespace, pa.Backend, ingressConfig)
				if err != nil {
					logger.Error().
						Str("serviceName", pa.Backend.Service.Name).
						Str("servicePort", pa.Backend.Service.Port.String()).
						Err(err).
						Msg("Cannot create service")
					continue
				}

				rt := &dynamic.Router{
					Rule: buildRule(rule.Host, pa, ingressConfig),
					// "default" stands for the default rule syntax in Traefik v3, i.e. the v3 syntax.
					RuleSyntax: "default",
					Service:    serviceName,
				}
				if hasTLS {
					rt.TLS = &dynamic.RouterTLSConfig{}
				}

				routerKey := provider.Normalize(fmt.Sprintf("%s-%s-rule-%d-path-%d", ingress.Namespace, ingress.Name, ri, pi))

				conf.HTTP.Routers[routerKey] = rt
				conf.HTTP.Services[serviceName] = service

				if namedServersTransport != nil && service.LoadBalancer != nil {
					service.LoadBalancer.ServersTransport = namedServersTransport.Name
					conf.HTTP.ServersTransports[namedServersTransport.Name] = namedServersTransport.ServersTransport
				}

				if err := p.applyMiddlewares(ingress.Namespace, routerKey, ingressConfig, hasTLS, rt, conf); err != nil {
					logger.Error().Err(err).Msg("Error applying middlewares")
				}
			}
		}
	}

	conf.TLS = &dynamic.TLSConfiguration{
		Certificates: slices.Collect(maps.Values(uniqCerts)),
	}

	return conf
}

func (p *Provider) buildServersTransport(namespace, name string, cfg ingressConfig) (*namedServersTransport, error) {
	scheme := parseBackendProtocol(ptr.Deref(cfg.BackendProtocol, "HTTP"))
	if scheme != "https" {
		return nil, nil
	}

	nst := &namedServersTransport{
		Name: provider.Normalize(namespace + "-" + name),
		ServersTransport: &dynamic.ServersTransport{
			ServerName:         ptr.Deref(cfg.ProxySSLName, ptr.Deref(cfg.ProxySSLServerName, "")),
			InsecureSkipVerify: strings.ToLower(ptr.Deref(cfg.ProxySSLVerify, "off")) == "on",
		},
	}

	if sslSecret := ptr.Deref(cfg.ProxySSLSecret, ""); sslSecret != "" {
		parts := strings.Split(sslSecret, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("malformed proxy SSL secret: %s, expected namespace/name", sslSecret)
		}

		blocks, err := p.certificateBlocks(parts[0], parts[1])
		if err != nil {
			return nil, fmt.Errorf("getting certificate blocks: %w", err)
		}

		if blocks.CA != nil {
			nst.ServersTransport.RootCAs = []types.FileOrContent{*blocks.CA}
		}

		if blocks.Certificate != nil {
			nst.ServersTransport.Certificates = []tls.Certificate{*blocks.Certificate}
		}
	}

	return nst, nil
}

func (p *Provider) buildService(namespace string, backend netv1.IngressBackend, cfg ingressConfig) (*dynamic.Service, error) {
	backendAddresses, err := p.getBackendAddresses(namespace, backend, cfg)
	if err != nil {
		return nil, fmt.Errorf("getting backend addresses: %w", err)
	}

	lb := &dynamic.ServersLoadBalancer{}
	lb.SetDefaults()

	if ptr.Deref(cfg.Affinity, "") != "" {
		lb.Sticky = &dynamic.Sticky{
			Cookie: &dynamic.Cookie{
				Name:     ptr.Deref(cfg.SessionCookieName, "INGRESSCOOKIE"),
				Secure:   ptr.Deref(cfg.SessionCookieSecure, false),
				HTTPOnly: true, // Default value in Nginx.
				SameSite: strings.ToLower(ptr.Deref(cfg.SessionCookieSameSite, "")),
				MaxAge:   ptr.Deref(cfg.SessionCookieMaxAge, 0),
				Path:     ptr.To(ptr.Deref(cfg.SessionCookiePath, "/")),
				Domain:   ptr.Deref(cfg.SessionCookieDomain, ""),
			},
		}
	}

	scheme := parseBackendProtocol(ptr.Deref(cfg.BackendProtocol, "HTTP"))

	svc := &dynamic.Service{LoadBalancer: lb}
	for _, addr := range backendAddresses {
		svc.LoadBalancer.Servers = append(svc.LoadBalancer.Servers, dynamic.Server{
			URL: fmt.Sprintf("%s://%s", scheme, addr.Address),
		})
	}

	return svc, nil
}

func (p *Provider) buildPassthroughService(namespace string, backend netv1.IngressBackend, cfg ingressConfig) (*dynamic.TCPService, error) {
	backendAddresses, err := p.getBackendAddresses(namespace, backend, cfg)
	if err != nil {
		return nil, fmt.Errorf("getting backend addresses: %w", err)
	}

	lb := &dynamic.TCPServersLoadBalancer{}
	for _, addr := range backendAddresses {
		lb.Servers = append(lb.Servers, dynamic.TCPServer{
			Address: addr.Address,
		})
	}

	return &dynamic.TCPService{LoadBalancer: lb}, nil
}

func (p *Provider) getBackendAddresses(namespace string, backend netv1.IngressBackend, cfg ingressConfig) ([]backendAddress, error) {
	service, err := p.k8sClient.GetService(namespace, backend.Service.Name)
	if err != nil {
		return nil, fmt.Errorf("getting service: %w", err)
	}

	if p.DisableSvcExternalName && service.Spec.Type == corev1.ServiceTypeExternalName {
		return nil, errors.New("externalName services not allowed")
	}

	var portName string
	var portSpec corev1.ServicePort
	var match bool
	for _, p := range service.Spec.Ports {
		// A port with number 0 or an empty name is not allowed, this case is there for the default backend service.
		if (backend.Service.Port.Number == 0 && backend.Service.Port.Name == "") ||
			(backend.Service.Port.Number == p.Port || (backend.Service.Port.Name == p.Name && len(p.Name) > 0)) {
			portName = p.Name
			portSpec = p
			match = true
			break
		}
	}
	if !match {
		return nil, errors.New("service port not found")
	}

	if service.Spec.Type == corev1.ServiceTypeExternalName {
		return []backendAddress{{Address: net.JoinHostPort(service.Spec.ExternalName, strconv.Itoa(int(portSpec.Port)))}}, nil
	}

	// When service upstream is set to true we return the service ClusterIP as the backend address.
	if ptr.Deref(cfg.ServiceUpstream, false) {
		return []backendAddress{{Address: net.JoinHostPort(service.Spec.ClusterIP, strconv.Itoa(int(portSpec.Port)))}}, nil
	}

	endpointSlices, err := p.k8sClient.GetEndpointSlicesForService(namespace, backend.Service.Name)
	if err != nil {
		return nil, fmt.Errorf("getting endpointslices: %w", err)
	}

	var addresses []backendAddress
	uniqAddresses := map[string]struct{}{}
	for _, endpointSlice := range endpointSlices {
		var port int32
		for _, p := range endpointSlice.Ports {
			if portName == *p.Name {
				port = *p.Port
				break
			}
		}
		if port == 0 {
			continue
		}

		for _, endpoint := range endpointSlice.Endpoints {
			if !k8s.EndpointServing(endpoint) {
				continue
			}

			for _, address := range endpoint.Addresses {
				if _, ok := uniqAddresses[address]; ok {
					continue
				}

				uniqAddresses[address] = struct{}{}
				addresses = append(addresses, backendAddress{
					Address: net.JoinHostPort(address, strconv.Itoa(int(port))),
					Fenced:  ptr.Deref(endpoint.Conditions.Terminating, false) && ptr.Deref(endpoint.Conditions.Serving, false),
				})
			}
		}
	}

	return addresses, nil
}

func (p *Provider) updateIngressStatus(ing *netv1.Ingress) error {
	if p.PublishService == "" && len(p.PublishStatusAddress) == 0 {
		// Nothing to do, no PublishService or PublishStatusAddress defined.
		return nil
	}

	if len(p.PublishStatusAddress) > 0 {
		ingStatus := make([]netv1.IngressLoadBalancerIngress, 0, len(p.PublishStatusAddress))
		for _, nameOrIP := range p.PublishStatusAddress {
			if net.ParseIP(nameOrIP) != nil {
				ingStatus = append(ingStatus, netv1.IngressLoadBalancerIngress{IP: nameOrIP})
				continue
			}

			ingStatus = append(ingStatus, netv1.IngressLoadBalancerIngress{Hostname: nameOrIP})
		}

		return p.k8sClient.UpdateIngressStatus(ing, ingStatus)
	}

	serviceInfo := strings.Split(p.PublishService, "/")
	if len(serviceInfo) != 2 {
		return fmt.Errorf("parsing publishService, 'namespace/service' format expected: %s", p.PublishService)
	}

	serviceNamespace, serviceName := serviceInfo[0], serviceInfo[1]

	service, err := p.k8sClient.GetService(serviceNamespace, serviceName)
	if err != nil {
		return fmt.Errorf("getting service: %w", err)
	}

	var ingressStatus []netv1.IngressLoadBalancerIngress

	switch service.Spec.Type {
	case corev1.ServiceTypeExternalName:
		ingressStatus = []netv1.IngressLoadBalancerIngress{{
			Hostname: service.Spec.ExternalName,
		}}

	case corev1.ServiceTypeClusterIP:
		ingressStatus = []netv1.IngressLoadBalancerIngress{{
			IP: service.Spec.ClusterIP,
		}}

	case corev1.ServiceTypeNodePort:
		if service.Spec.ExternalIPs == nil {
			ingressStatus = []netv1.IngressLoadBalancerIngress{{
				IP: service.Spec.ClusterIP,
			}}
		} else {
			ingressStatus = make([]netv1.IngressLoadBalancerIngress, 0, len(service.Spec.ExternalIPs))
			for _, ip := range service.Spec.ExternalIPs {
				ingressStatus = append(ingressStatus, netv1.IngressLoadBalancerIngress{IP: ip})
			}
		}

	case corev1.ServiceTypeLoadBalancer:
		ingressStatus, err = convertSlice[netv1.IngressLoadBalancerIngress](service.Status.LoadBalancer.Ingress)
		if err != nil {
			return fmt.Errorf("converting ingress loadbalancer status: %w", err)
		}
		for _, ip := range service.Spec.ExternalIPs {
			// Avoid duplicates in the ingress status.
			var found bool
			for _, status := range ingressStatus {
				if status.IP == ip || status.Hostname == ip {
					found = true
					continue
				}
			}
			if !found {
				ingressStatus = append(ingressStatus, netv1.IngressLoadBalancerIngress{IP: ip})
			}
		}
	}

	return p.k8sClient.UpdateIngressStatus(ing, ingressStatus)
}

func (p *Provider) shouldProcessIngress(ingress *netv1.Ingress, ingressClasses []*netv1.IngressClass) bool {
	if len(ingressClasses) > 0 && ingress.Spec.IngressClassName != nil {
		return slices.ContainsFunc(ingressClasses, func(ic *netv1.IngressClass) bool {
			return *ingress.Spec.IngressClassName == ic.ObjectMeta.Name
		})
	}

	if class, ok := ingress.Annotations[annotationIngressClass]; ok {
		return class == p.IngressClass
	}

	return p.WatchIngressWithoutClass
}

func (p *Provider) loadCertificates(ctx context.Context, ingress *netv1.Ingress, uniqCerts map[string]*tls.CertAndStores) error {
	for _, t := range ingress.Spec.TLS {
		if t.SecretName == "" {
			log.Ctx(ctx).Debug().Msg("Skipping TLS sub-section: No secret name provided")
			continue
		}

		certKey := ingress.Namespace + "-" + t.SecretName
		if _, certExists := uniqCerts[certKey]; !certExists {
			blocks, err := p.certificateBlocks(ingress.Namespace, t.SecretName)
			if err != nil {
				return fmt.Errorf("getting certificate blocks: %w", err)
			}

			if blocks.Certificate == nil {
				return fmt.Errorf("no keypair found in secret %s/%s", ingress.Namespace, t.SecretName)
			}

			uniqCerts[certKey] = &tls.CertAndStores{
				Certificate: *blocks.Certificate,
			}
		}
	}

	return nil
}

func (p *Provider) applyMiddlewares(namespace, routerKey string, ingressConfig ingressConfig, hasTLS bool, rt *dynamic.Router, conf *dynamic.Configuration) error {
	if err := p.applyBasicAuthConfiguration(namespace, routerKey, ingressConfig, rt, conf); err != nil {
		return fmt.Errorf("applying basic auth configuration: %w", err)
	}

	if err := applyForwardAuthConfiguration(routerKey, ingressConfig, rt, conf); err != nil {
		return fmt.Errorf("applying forward auth configuration: %w", err)
	}

	applyCORSConfiguration(routerKey, ingressConfig, rt, conf)

	// Apply SSL redirect is mandatory to be applied after all other middlewares.
	// TODO: check how to remove this, and create the HTTP router elsewhere.
	applySSLRedirectConfiguration(routerKey, ingressConfig, hasTLS, rt, conf)

	return nil
}

func (p *Provider) applyBasicAuthConfiguration(namespace, routerName string, ingressConfig ingressConfig, rt *dynamic.Router, conf *dynamic.Configuration) error {
	if ingressConfig.AuthType == nil {
		return nil
	}

	authType := ptr.Deref(ingressConfig.AuthType, "")
	if authType != "basic" && authType != "digest" {
		return fmt.Errorf("invalid auth-type %q, must be 'basic' or 'digest'", authType)
	}

	authSecret := ptr.Deref(ingressConfig.AuthSecret, "")
	if authSecret == "" {
		return fmt.Errorf("invalid auth-secret %q, must not be empty", authSecret)
	}

	authSecretParts := strings.Split(authSecret, "/")
	if len(authSecretParts) > 2 {
		return fmt.Errorf("invalid auth secret %q", authSecret)
	}

	secretName := authSecretParts[0]
	secretNamespace := namespace
	if len(authSecretParts) == 2 {
		secretNamespace = authSecretParts[0]
		secretName = authSecretParts[1]
	}

	secret, err := p.k8sClient.GetSecret(secretNamespace, secretName)
	if err != nil {
		return fmt.Errorf("getting secret %s: %w", authSecret, err)
	}

	authSecretType := ptr.Deref(ingressConfig.AuthSecretType, "auth-file")
	if authSecretType != "auth-file" && authSecretType != "auth-map" {
		return fmt.Errorf("invalid auth-secret-type %q, must be 'auth-file' or 'auth-map'", authSecretType)
	}

	users, err := basicAuthUsers(secret, authSecretType)
	if err != nil {
		return fmt.Errorf("getting users from secret %s: %w", authSecret, err)
	}

	realm := ptr.Deref(ingressConfig.AuthRealm, "")

	switch authType {
	case "basic":
		basicMiddlewareName := routerName + "-basic-auth"
		conf.HTTP.Middlewares[basicMiddlewareName] = &dynamic.Middleware{
			BasicAuth: &dynamic.BasicAuth{
				Users:        users,
				Realm:        realm,
				RemoveHeader: false,
			},
		}
		rt.Middlewares = append(rt.Middlewares, basicMiddlewareName)

	case "digest":
		digestMiddlewareName := routerName + "-digest-auth"
		conf.HTTP.Middlewares[digestMiddlewareName] = &dynamic.Middleware{
			DigestAuth: &dynamic.DigestAuth{
				Users:        users,
				Realm:        realm,
				RemoveHeader: false,
			},
		}
		rt.Middlewares = append(rt.Middlewares, digestMiddlewareName)
	}

	return nil
}

func (p *Provider) certificateBlocks(namespace, name string) (*certBlocks, error) {
	secret, err := p.k8sClient.GetSecret(namespace, name)
	if err != nil {
		return nil, fmt.Errorf("fetching secret %s/%s: %w", namespace, name, err)
	}

	certBytes, hasCert := secret.Data[corev1.TLSCertKey]
	keyBytes, hasKey := secret.Data[corev1.TLSPrivateKeyKey]
	caBytes, hasCA := secret.Data[corev1.ServiceAccountRootCAKey]

	if !hasCert && !hasKey && !hasCA {
		return nil, errors.New("secret does not contain a keypair or CA certificate")
	}

	var blocks certBlocks
	if hasCA {
		if len(caBytes) == 0 {
			return nil, errors.New("secret contains an empty CA certificate")
		}

		ca := types.FileOrContent(caBytes)
		blocks.CA = &ca
	}

	if hasKey && hasCert {
		if len(certBytes) == 0 {
			return nil, errors.New("secret contains an empty certificate")
		}
		if len(keyBytes) == 0 {
			return nil, errors.New("secret contains an empty key")
		}
		blocks.Certificate = &tls.Certificate{
			CertFile: types.FileOrContent(certBytes),
			KeyFile:  types.FileOrContent(keyBytes),
		}
	}

	return &blocks, nil
}

func applyCORSConfiguration(routerName string, ingressConfig ingressConfig, rt *dynamic.Router, conf *dynamic.Configuration) {
	if !ptr.Deref(ingressConfig.EnableCORS, false) {
		return
	}

	corsMiddlewareName := routerName + "-cors"
	conf.HTTP.Middlewares[corsMiddlewareName] = &dynamic.Middleware{
		Headers: &dynamic.Headers{
			AccessControlAllowCredentials: ptr.Deref(ingressConfig.EnableCORSAllowCredentials, true),
			AccessControlExposeHeaders:    ptr.Deref(ingressConfig.CORSExposeHeaders, []string{}),
			AccessControlAllowHeaders:     ptr.Deref(ingressConfig.CORSAllowHeaders, []string{"DNT", "Keep-Alive", "User-Agent", "X-Requested-With", "If-Modified-Since", "Cache-Control", "Content-Type", "Range,Authorization"}),
			AccessControlAllowMethods:     ptr.Deref(ingressConfig.CORSAllowMethods, []string{"GET", "PUT", "POST", "DELETE", "PATCH", "OPTIONS"}),
			AccessControlAllowOriginList:  ptr.Deref(ingressConfig.CORSAllowOrigin, []string{"*"}),
			AccessControlMaxAge:           int64(ptr.Deref(ingressConfig.CORSMaxAge, 1728000)),
		},
	}

	rt.Middlewares = append(rt.Middlewares, corsMiddlewareName)
}

func applySSLRedirectConfiguration(routerName string, ingressConfig ingressConfig, hasTLS bool, rt *dynamic.Router, conf *dynamic.Configuration) {
	var forceSSLRedirect bool
	if ingressConfig.ForceSSLRedirect != nil {
		forceSSLRedirect = *ingressConfig.ForceSSLRedirect
	}

	sslRedirect := ptr.Deref(ingressConfig.SSLRedirect, hasTLS)

	if !forceSSLRedirect && !sslRedirect {
		if hasTLS {
			httpRouter := &dynamic.Router{
				Rule: rt.Rule,
				// "default" stands for the default rule syntax in Traefik v3, i.e. the v3 syntax.
				RuleSyntax:  "default",
				Middlewares: rt.Middlewares,
				Service:     rt.Service,
			}

			conf.HTTP.Routers[routerName+"-http"] = httpRouter
		}

		return
	}

	redirectRouter := &dynamic.Router{
		Rule: rt.Rule,
		// "default" stands for the default rule syntax in Traefik v3, i.e. the v3 syntax.
		RuleSyntax: "default",
		Service:    "noop@internal",
	}

	redirectMiddlewareName := routerName + "-redirect-scheme"
	conf.HTTP.Middlewares[redirectMiddlewareName] = &dynamic.Middleware{
		RedirectScheme: &dynamic.RedirectScheme{
			Scheme:    "https",
			Permanent: true,
		},
	}
	redirectRouter.Middlewares = append(redirectRouter.Middlewares, redirectMiddlewareName)

	conf.HTTP.Routers[routerName+"-redirect"] = redirectRouter
}

func applyForwardAuthConfiguration(routerName string, ingressConfig ingressConfig, rt *dynamic.Router, conf *dynamic.Configuration) error {
	if ingressConfig.AuthURL == nil {
		return nil
	}

	if *ingressConfig.AuthURL == "" {
		return errors.New("empty auth-url found in ingress annotations")
	}

	authResponseHeaders := strings.Split(ptr.Deref(ingressConfig.AuthResponseHeaders, ""), ",")

	forwardMiddlewareName := routerName + "-forward-auth"
	conf.HTTP.Middlewares[forwardMiddlewareName] = &dynamic.Middleware{
		ForwardAuth: &dynamic.ForwardAuth{
			Address:             *ingressConfig.AuthURL,
			AuthResponseHeaders: authResponseHeaders,
		},
	}
	rt.Middlewares = append(rt.Middlewares, forwardMiddlewareName)

	return nil
}

func basicAuthUsers(secret *corev1.Secret, authSecretType string) (dynamic.Users, error) {
	var users dynamic.Users
	if authSecretType == "auth-map" {
		if len(secret.Data) == 0 {
			return nil, fmt.Errorf("secret %s/%s does not contain any user credentials", secret.Namespace, secret.Name)
		}

		for user, pass := range secret.Data {
			users = append(users, user+":"+string(pass))
		}

		return users, nil
	}

	// Default to auth-file type.
	authFileContent, ok := secret.Data["auth"]
	if !ok {
		return nil, fmt.Errorf("secret %s/%s does not contain auth-file content key `auth`", secret.Namespace, secret.Name)
	}

	// Trim lines and filter out blanks
	rawLines := strings.Split(string(authFileContent), "\n")
	for _, rawLine := range rawLines {
		line := strings.TrimSpace(rawLine)
		if line != "" && !strings.HasPrefix(line, "#") {
			users = append(users, line)
		}
	}

	return users, nil
}

func buildRule(host string, pa netv1.HTTPIngressPath, config ingressConfig) string {
	var rules []string
	if len(host) > 0 {
		rules = append(rules, buildHostRule(host))
	}

	if len(pa.Path) > 0 {
		pathType := ptr.Deref(pa.PathType, netv1.PathTypePrefix)
		if pathType == netv1.PathTypeImplementationSpecific {
			pathType = netv1.PathTypePrefix
		}

		switch pathType {
		case netv1.PathTypeExact:
			rules = append(rules, fmt.Sprintf("Path(`%s`)", pa.Path))
		case netv1.PathTypePrefix:
			if ptr.Deref(config.UseRegex, false) {
				rules = append(rules, fmt.Sprintf("PathRegexp(`^%s`)", regexp.QuoteMeta(pa.Path)))
			} else {
				rules = append(rules, buildPrefixRule(pa.Path))
			}
		}
	}

	return strings.Join(rules, " && ")
}

func buildHostRule(host string) string {
	if strings.HasPrefix(host, "*.") {
		host = strings.Replace(regexp.QuoteMeta(host), `\*\.`, `[a-zA-Z0-9-]+\.`, 1)
		return fmt.Sprintf("HostRegexp(`^%s$`)", host)
	}

	return fmt.Sprintf("Host(`%s`)", host)
}

// buildPrefixRule is a helper function to build a path prefix rule that matches path prefix split by `/`.
// For example, the paths `/abc`, `/abc/`, and `/abc/def` would all match the prefix `/abc`,
// but the path `/abcd` would not. See TestStrictPrefixMatchingRule() for more examples.
//
// "PathPrefix" in Kubernetes Gateway API is semantically equivalent to the "Prefix" path type in the
// Kubernetes Ingress API.
func buildPrefixRule(path string) string {
	if path == "/" {
		return "PathPrefix(`/`)"
	}

	path = strings.TrimSuffix(path, "/")
	return fmt.Sprintf("(Path(`%[1]s`) || PathPrefix(`%[1]s/`))", path)
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
