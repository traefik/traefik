package ingressnginx

import (
	"context"
	"fmt"
	"net"
	"os"
	"regexp"
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
	"github.com/traefik/traefik/v3/pkg/types"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
)

const (
	// ProviderName is the Kubernetes Ingress NGINX provider name.
	ProviderName = "kubernetesingressnginx"

	// unavailableServiceName is the name of a Traefik service returning a 503 Service Unavailable.
	unavailableServiceName = "unavailable-service"

	// NGINX default values.
	annotationIngressClass = "kubernetes.io/ingress.class"

	defaultControllerName  = "k8s.io/ingress-nginx"
	defaultAnnotationValue = "nginx"

	defaultBackendName    = "default-backend"
	defaultBackendTLSName = "default-backend-tls"

	defaultProxyConnectTimeout = 60
	defaultProxyReadTimeout    = 60
	defaultProxySendTimeout    = 60
	// https://nginx.org/en/docs/http/ngx_http_core_module.html#client_max_body_size
	defaultProxyBodySize = int64(1024 * 1024) // 1MB
	// https://nginx.org/en/docs/http/ngx_http_core_module.html#client_body_buffer_size
	defaultClientBodyBufferSize = int64(16 * 1024) // 16KB
	// https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_buffer_size
	defaultProxyBufferSize = int64(8 * 1024) // 8KB
	// https://github.com/kubernetes/ingress-nginx/blob/main/docs/user-guide/nginx-configuration/annotations.md#proxy-buffers-number
	defaultProxyBuffersNumber = 4
	// https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_max_temp_file_size
	defaultProxyMaxTempFileSize = int64(1024 * 1024 * 1024) // 1GB
	// https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/configmap/#proxy-next-upstream
	defaultProxyNextUpstream = "error timeout"
	// https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/configmap/#proxy-next-upstream-tries
	// ingress-nginx uses 3 as default value.
	defaultProxyNextUpstreamTries = 3

	// https://github.com/kubernetes/ingress-nginx/blob/main/docs/user-guide/nginx-configuration/annotations.md#rate-limiting
	defaultLimitBurstMultiplier = 5

	// https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/configmap/#upstream-keepalive-timeout
	defaultUpstreamKeepaliveTimeout = 60
)

var (
	nginxSizeRegexp = regexp.MustCompile(`^(?i)\s*([0-9]+)\s*([bkmg]?)\s*$`)

	headerRegexp      = regexp.MustCompile(`^[a-zA-Z\d\-_]+$`)
	headerValueRegexp = regexp.MustCompile(`^[a-zA-Z\d_ :;.,\\/"'?!(){}\[\]@<>=\-+*#$&\x60|~^%]+$`)
	// The same regexp used in ingress-nginx: https://github.com/kubernetes/ingress-nginx/blob/main/internal/ingress/inspector/rules.go.
	strictPathTypeRegexp = regexp.MustCompile(`(?i)^/[[:alnum:]._\-/]*$`)
	// The same regexp used in ingress-nginx: https://github.com/kubernetes/ingress-nginx/blob/main/internal/ingress/annotations/parser/validators.go#L77
	regexPathWithCapture = regexp.MustCompile(`^/?[-._~a-zA-Z0-9/$:]*$`)
)

// Provider holds configurations of the provider.
type Provider struct {
	Endpoint         string              `description:"Kubernetes server endpoint (required for external cluster client)." json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Token            types.FileOrContent `description:"Kubernetes bearer token (not needed for in-cluster client). It accepts either a token value or a file path to the token." json:"token,omitempty" toml:"token,omitempty" yaml:"token,omitempty" loggable:"false"`
	CertAuthFilePath string              `description:"Kubernetes certificate authority file path (not needed for in-cluster client)." json:"certAuthFilePath,omitempty" toml:"certAuthFilePath,omitempty" yaml:"certAuthFilePath,omitempty"`
	ThrottleDuration ptypes.Duration     `description:"Ingress refresh throttle duration." json:"throttleDuration,omitempty" toml:"throttleDuration,omitempty" yaml:"throttleDuration,omitempty" export:"true"`
	GlobalAuthURL    string              `description:"URL to the service that provides authentication for all the locations. Per ingress auth-url annotation has precedence over this option." json:"globalAuthURL,omitempty" toml:"globalAuthURL,omitempty" yaml:"globalAuthURL,omitempty" export:"true"`

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

	// Configuration options available within the NGINX Ingress Controller ConfigMap.
	ProxyRequestBuffering    bool     `description:"Defines whether to enable request buffering." json:"proxyRequestBuffering,omitempty" toml:"proxyRequestBuffering,omitempty" yaml:"proxyRequestBuffering,omitempty" export:"true"`
	ClientBodyBufferSize     int64    `description:"Default buffer size for reading client request body." json:"clientBodyBufferSize,omitempty" toml:"clientBodyBufferSize,omitempty" yaml:"clientBodyBufferSize,omitempty" export:"true"`
	ProxyBodySize            int64    `description:"Default maximum size of a client request body in bytes." json:"proxyBodySize,omitempty" toml:"proxyBodySize,omitempty" yaml:"proxyBodySize,omitempty" export:"true"`
	ProxyBuffering           bool     `description:"Defines whether to enable response buffering." json:"proxyBuffering,omitempty" toml:"proxyBuffering,omitempty" yaml:"proxyBuffering,omitempty" export:"true"`
	ProxyBufferSize          int64    `description:"Default buffer size for reading the response body." json:"proxyBufferSize,omitempty" toml:"proxyBufferSize,omitempty" yaml:"proxyBufferSize,omitempty" export:"true"`
	ProxyBuffersNumber       int      `description:"Default number of buffers for reading a response." json:"proxyBuffersNumber,omitempty" toml:"proxyBuffersNumber,omitempty" yaml:"proxyBuffersNumber,omitempty" export:"true"`
	ProxyConnectTimeout      int      `description:"Amount of time to wait until a connection to a server can be established. Timeout value is unitless and in seconds." json:"proxyConnectTimeout,omitempty" toml:"proxyConnectTimeout,omitempty" yaml:"proxyConnectTimeout,omitempty" export:"true"`
	ProxyReadTimeout         int      `description:"Amount of time between two successive read operations. Timeout value is unitless and in seconds." json:"proxyReadTimeout,omitempty" toml:"proxyReadTimeout,omitempty" yaml:"proxyReadTimeout,omitempty" export:"true"`
	ProxySendTimeout         int      `description:"Amount of time between two successive write operations. Timeout value is unitless and in seconds." json:"proxySendTimeout,omitempty" toml:"proxySendTimeout,omitempty" yaml:"proxySendTimeout,omitempty" export:"true"`
	ProxyNextUpstream        string   `description:"Defines in which cases a request should be retried." json:"proxyNextUpstream,omitempty" toml:"proxyNextUpstream,omitempty" yaml:"proxyNextUpstream,omitempty" export:"true"`
	ProxyNextUpstreamTries   int      `description:"Limits the number of possible tries if the backend server does not reply." json:"proxyNextUpstreamTries,omitempty" toml:"proxyNextUpstreamTries,omitempty" yaml:"proxyNextUpstreamTries,omitempty" export:"true"`
	ProxyNextUpstreamTimeout int      `description:"Limits the total elapsed time to retry the request if the backend server does not reply. Timeout value is unitless and in seconds." json:"proxyNextUpstreamTimeout,omitempty" toml:"proxyNextUpstreamTimeout,omitempty" yaml:"proxyNextUpstreamTimeout,omitempty" export:"true"`
	CustomHTTPErrors         []string `description:"Defines which status should result in calling the default backend to return an error page." json:"customHTTPErrors,omitempty" toml:"customHTTPErrors,omitempty" yaml:"customHTTPErrors,omitempty" export:"true"`
	UpstreamKeepaliveTimeout int      `description:"Defines the idle timeout for keep-alive connections to upstream servers. Timeout value is unitless and in seconds." json:"upstreamKeepaliveTimeout,omitempty" toml:"upstreamKeepaliveTimeout,omitempty" yaml:"upstreamKeepaliveTimeout,omitempty" export:"true"`

	AllowCrossNamespaceResources bool     `description:"Allow Ingress to reference resources (e.g. ConfigMaps, Secrets) in different namespaces." json:"allowCrossNamespaceResources,omitempty" toml:"allowCrossNamespaceResources,omitempty" yaml:"allowCrossNamespaceResources,omitempty" export:"true"`
	GlobalAllowedResponseHeaders []string `description:"List of allowed response headers inside the custom headers annotations." json:"globalAllowedResponseHeaders,omitempty" toml:"globalAllowedResponseHeaders,omitempty" yaml:"globalAllowedResponseHeaders,omitempty" export:"true"`

	AllowSnippetAnnotations bool `description:"Enables to parse and add -snippet annotations/directives." json:"allowSnippetAnnotations,omitempty" toml:"allowSnippetAnnotations,omitempty" yaml:"allowSnippetAnnotations,omitempty" export:"true"`

	HTTPEntryPoint  string `description:"Defines the EntryPoint to use for HTTP requests." json:"httpEntryPoint,omitempty" toml:"httpEntryPoint,omitempty" yaml:"httpEntryPoint,omitempty" export:"true"`
	HTTPSEntryPoint string `description:"Defines the EntryPoint to use for HTTPS requests." json:"httpsEntryPoint,omitempty" toml:"httpsEntryPoint,omitempty" yaml:"httpsEntryPoint,omitempty" export:"true"`
	// TLSEntryPoints is set to the HTTPSEntryPoint value if it is set, otherwise it is left empty.
	TLSEntryPoints []string `json:"-" toml:"-" yaml:"-" label:"-" file:"-"`
	// NonTLSEntryPoints contains the names of entrypoints that are configured without TLS.
	// Its value is set to the HTTPEntryPoint value if it is set, otherwise it is computed in SetEffectiveConfiguration.
	NonTLSEntryPoints []string `json:"-" toml:"-" yaml:"-" label:"-" file:"-"`

	StrictValidatePathType bool `description:"Defines whether to reject the entire ingress when any path contains regex characters and pathType is Prefix or Exact." json:"strictValidatePathType,omitempty" toml:"strictValidatePathType,omitempty" yaml:"strictValidatePathType,omitempty" export:"true"`

	allowedHeaders                 []string
	defaultBackendServiceNamespace string
	defaultBackendServiceName      string

	k8sClient         *clientWrapper
	lastConfiguration safe.Safe

	applyMiddlewareFunc func(routerKey string, router *dynamic.Router, config *dynamic.Configuration, ingressConfig IngressConfig) error
}

// SetApplyMiddlewareFunc sets an optional hook called for every router, allowing external code
// (e.g. the ModSecurity integration) to attach additional middlewares.
func (p *Provider) SetApplyMiddlewareFunc(fn func(routerKey string, router *dynamic.Router, config *dynamic.Configuration, ingressConfig IngressConfig) error) {
	p.applyMiddlewareFunc = fn
}

// SetDefaults sets the default values for the provider.
func (p *Provider) SetDefaults() {
	p.IngressClass = defaultAnnotationValue
	p.ControllerClass = defaultControllerName
	p.ProxyConnectTimeout = defaultProxyConnectTimeout
	p.ProxyReadTimeout = defaultProxyReadTimeout
	p.ProxySendTimeout = defaultProxySendTimeout
	p.ClientBodyBufferSize = defaultClientBodyBufferSize
	p.ProxyBodySize = defaultProxyBodySize
	p.ProxyBufferSize = defaultProxyBufferSize
	p.ProxyBuffersNumber = defaultProxyBuffersNumber
	p.ProxyNextUpstream = defaultProxyNextUpstream
	p.ProxyNextUpstreamTries = defaultProxyNextUpstreamTries
	p.UpstreamKeepaliveTimeout = defaultUpstreamKeepaliveTimeout
	p.StrictValidatePathType = true
}

// Init the provider.
func (p *Provider) Init() error {
	if err := p.validateConfiguration(); err != nil {
		return fmt.Errorf("validating %s provider configuration: %w", ProviderName, err)
	}

	if p.HTTPEntryPoint != "" {
		p.NonTLSEntryPoints = []string{p.HTTPEntryPoint}
	}
	if p.HTTPSEntryPoint != "" {
		p.TLSEntryPoints = []string{p.HTTPSEntryPoint}
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
	logger := log.With().Str(logs.ProviderName, ProviderName).Logger()
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
							ProviderName:  ProviderName,
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

func (p *Provider) loadConfiguration(ctx context.Context) *dynamic.Configuration {
	var ingressClasses []*netv1.IngressClass
	ics, err := p.k8sClient.ListIngressClasses()
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("Failed to list ingress classes")
	}
	ingressClasses = filterIngressClass(ics, p.IngressClassByName, p.IngressClass, p.ControllerClass)

	// Phase 1: build the metamodel from k8s resources.
	mc := p.build(ctx, ingressClasses)

	// Update ingress statuses (requires k8s access, must happen in Phase 1 context).
	for _, server := range mc.Servers {
		for _, loc := range server.Locations {
			if loc.IngressName == "" {
				continue
			}
			// Retrieve the original ingress to update its status.
			for _, ing := range p.k8sClient.ListIngresses() {
				if ing.Namespace == loc.Namespace && ing.Name == loc.IngressName {
					if err := p.updateIngressStatus(ing); err != nil {
						log.Ctx(ctx).Error().Err(err).
							Str("namespace", ing.Namespace).
							Str("ingress", ing.Name).
							Msg("Error while updating ingress status")
					}
					break
				}
			}
		}
	}

	// Phase 2: translate the metamodel into a Traefik dynamic.Configuration.
	return p.translate(ctx, mc)
}

func (p *Provider) validateConfiguration() error {
	// Validates and parses the default backend configuration.
	if p.DefaultBackendService != "" {
		parts := strings.Split(p.DefaultBackendService, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid default backend service format: %s, expected 'namespace/name'", p.DefaultBackendService)
		}
		p.defaultBackendServiceNamespace = parts[0]
		p.defaultBackendServiceName = parts[1]
	}

	var allowedHeaders []string
	for _, header := range p.GlobalAllowedResponseHeaders {
		if !headerRegexp.MatchString(header) {
			log.Warn().Msgf("GlobalAllowedResponseHeaders header value %q is invalid and will be ignored. Only alphanumeric characters, dashes and underscores are allowed.", header)
			continue
		}

		allowedHeaders = append(allowedHeaders, header)
	}

	p.allowedHeaders = allowedHeaders
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

func (p *Provider) updateIngressStatus(ing *netv1.Ingress) error {
	if p.PublishService == "" && len(p.PublishStatusAddress) == 0 {
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
					break
				}
			}
			if !found {
				ingressStatus = append(ingressStatus, netv1.IngressLoadBalancerIngress{IP: ip})
			}
		}
	}

	return p.k8sClient.UpdateIngressStatus(ing, ingressStatus)
}

func throttleEvents(ctx context.Context, throttleDuration time.Duration, pool *safe.Pool, eventsChan <-chan any) chan any {
	if throttleDuration == 0 {
		return nil
	}

	// Create a buffered channel to hold the pending event (if we're delaying processing the event due to throttling).
	eventsChanBuffered := make(chan any, 1)

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
