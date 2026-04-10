package ingressnginx

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"math"
	"net"
	"net/http"
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
	"github.com/traefik/traefik/v3/pkg/observability/logs"
	"github.com/traefik/traefik/v3/pkg/provider"
	"github.com/traefik/traefik/v3/pkg/safe"
	"github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/utils/ptr"
)

const (
	providerName = "kubernetesingressnginx"

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
	// The same regexp used in ingress-nginx:https://github.com/kubernetes/ingress-nginx/blob/main/internal/ingress/inspector/rules.go.
	strictPathTypeRegexp = regexp.MustCompile(`(?i)^/[[:alnum:]._\-/]*$`)
)

type backendAddress struct {
	Address string
	Fenced  bool
}

type namedServersTransport struct {
	*dynamic.ServersTransport

	Name string
}

type certBlocks struct {
	CA          *types.FileOrContent
	Certificate *tls.Certificate
}

type ingress struct {
	*netv1.Ingress

	IngressConfig IngressConfig
}

type ingressPath struct {
	netv1.HTTPIngressPath

	IngressConfig IngressConfig
}

type canaryBackend struct {
	*netv1.IngressBackend

	Cookie        string
	Header        string
	HeaderValue   string
	HeaderPattern string
	Weight        int
	WeightTotal   int
}

// RequiresCanaryRouter returns true if the canary backend requires a canary router configuration for Cookie or Header routing.
func (c canaryBackend) RequiresCanaryRouter() bool {
	return c.Cookie != "" || c.Header != ""
}

// RequiresNonCanaryRouter returns true if the canary backend requires a non-canary router configuration for Cookie or Header routing.
// This is the case when only the Header/Cookie options are configured, as a "never" value should forward the request to the production service.
// When the canary weight is 0, no canary router should be created as all the traffic will be handled by the non-canary router based on the weight configuration.
func (c canaryBackend) RequiresNonCanaryRouter() bool {
	return c.Weight > 0 && ((c.Header != "" && c.HeaderValue == "" && c.HeaderPattern == "") || c.Cookie != "")
}

// AppendCanaryRule appends the canary condition to the given rule based on the canary configuration for Cookie or Header routing.
func (c canaryBackend) AppendCanaryRule(rule string) string {
	var rules []string
	if c.Header != "" {
		switch {
		case c.HeaderValue == "" && c.HeaderPattern != "":
			rules = append(rules, fmt.Sprintf("HeaderRegexp(%q, %q)", c.Header, c.HeaderPattern))

		case c.HeaderValue != "":
			rules = append(rules, fmt.Sprintf("Header(%q, %q)", c.Header, c.HeaderValue))

		default:
			rules = append(rules, fmt.Sprintf(`Header(%q, "always")`, c.Header))
		}
	}

	if c.Cookie != "" {
		cookieRule := fmt.Sprintf(`HeaderRegexp("Cookie", %q)`, fmt.Sprintf("(^|;\\s*)%s=always(;|$)", c.Cookie))
		if c.Header != "" && c.HeaderValue == "" && c.HeaderPattern == "" {
			cookieRule = fmt.Sprintf("(%s && !%s)", cookieRule, fmt.Sprintf(`Header(%q, "never")`, c.Header))
		}

		rules = append(rules, cookieRule)
	}

	return fmt.Sprintf("(%s) && (%s)", rule, strings.Join(rules, " || "))
}

// AppendNonCanaryRule appends the non-canary condition to the given rule based on the canary configuration.
func (c canaryBackend) AppendNonCanaryRule(rule string) string {
	var rules []string
	if c.Header != "" && c.HeaderValue == "" && c.HeaderPattern == "" {
		rules = append(rules, fmt.Sprintf(`Header(%q, "never")`, c.Header))
	}
	if c.Cookie != "" {
		rules = append(rules, fmt.Sprintf(`HeaderRegexp("Cookie", %q)`, fmt.Sprintf("(^|;\\s*)%s=never(;|$)", c.Cookie)))
	}

	return fmt.Sprintf("(%s) && (%s)", rule, strings.Join(rules, " || "))
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

func (p *Provider) SetApplyMiddlewareFunc(fn func(routerKey string, router *dynamic.Router, config *dynamic.Configuration, ingressConfig IngressConfig) error) {
	p.applyMiddlewareFunc = fn
}

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
		return fmt.Errorf("validating kubernetesingressnginx provider configuration: %w", err)
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
		svc, err := p.buildService(p.defaultBackendServiceNamespace, ib, nil, IngressConfig{})
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("Cannot build default backend service")
			return conf
		}

		// Add the default backend service router to the configuration.
		conf.HTTP.Routers[defaultBackendName] = &dynamic.Router{
			EntryPoints: p.NonTLSEntryPoints,
			Rule:        `PathPrefix("/")`,
			// "default" stands for the default rule syntax in Traefik v3, i.e. the v3 syntax.
			RuleSyntax: "default",
			Priority:   math.MinInt32,
			Service:    defaultBackendName,
		}

		conf.HTTP.Routers[defaultBackendTLSName] = &dynamic.Router{
			EntryPoints: p.TLSEntryPoints,
			Rule:        `PathPrefix("/")`,
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

	var (
		ingresses       []ingress
		canaryIngresses []ingress
	)

	hosts := make(map[string]bool)
	hostsWithUseRegex := make(map[string]bool)
	serverSnippets := make(map[string]string)
	ingressPaths := make(map[string]ingressPath) // indexed by namespace+host+path+pathType.
	for _, ing := range p.k8sClient.ListIngresses() {
		if !p.shouldProcessIngress(ing, ingressClasses) {
			continue
		}

		logger := log.Ctx(ctx).With().
			Str("ingress", ing.Name).
			Str("namespace", ing.Namespace).
			Logger()

		i := ingress{
			Ingress:       ing,
			IngressConfig: parseIngressConfig(ing),
		}

		if err := p.isIngressValid(i); err != nil {
			logger.Error().Err(err).Msg("Invalid Ingress configuration")
			continue
		}

		// Canary ingresses should be processed separately after processing all the ingresses
		// to ensure that the canary rules are matching the ingress rules.
		if ptr.Deref(i.IngressConfig.Canary, false) {
			canaryIngresses = append(canaryIngresses, i)
			continue
		}

		for _, rule := range ing.Spec.Rules {
			hosts[rule.Host] = true

			// If any ingress in this host enable use-regex, all paths on that host must use regex matching.
			// Using rewrite-target annotation also implies that use-regex is true.
			if ptr.Deref(i.IngressConfig.UseRegex, false) || ptr.Deref(i.IngressConfig.RewriteTarget, "") != "" {
				hostsWithUseRegex[rule.Host] = true
			}

			if srvSnippet := ptr.Deref(i.IngressConfig.ServerSnippet, ""); srvSnippet != "" {
				if serverSnippets[rule.Host] != "" {
					logger.Debug().Msgf("Ignoring Server snippet because it is already defined for Host: %s", rule.Host)
				} else {
					serverSnippets[rule.Host] = srvSnippet
				}
			}

			if rule.HTTP != nil {
				for _, pa := range rule.HTTP.Paths {
					// We only consider paths with a defined backend service,
					// as those are the only ones that can serve requests.
					if pa.Backend.Service == nil {
						continue
					}

					key := ingressPathKey(ing.Namespace, rule.Host, pa)
					ingressPaths[key] = ingressPath{
						HTTPIngressPath: pa,
						IngressConfig:   i.IngressConfig,
					}
				}
			}
		}

		ingresses = append(ingresses, i)
	}

	// Now that we have all the ingresses and their paths,
	// we can process the canary ingresses and match them with the corresponding ingress rules,
	// to discover the canary backends.
	canaryBackends := make(map[string]*canaryBackend) // indexed by service namespace+name+port of the original ingress.
	matchedIngressPaths := make(map[string]struct{})  // indexed by namespace+host+path+pathType tracking which ingress paths have a matching canary rule.
	for _, canaryIngress := range canaryIngresses {
		for _, rule := range canaryIngress.Spec.Rules {
			if rule.HTTP == nil {
				continue
			}

			backends, matchedPaths, err := p.discoverCanaryBackends(canaryIngress.Namespace, rule, canaryIngress.IngressConfig, ingressPaths, matchedIngressPaths)
			if err != nil {
				log.Ctx(ctx).Error().
					Str("ingress", canaryIngress.Name).
					Str("namespace", canaryIngress.Namespace).
					Err(err).
					Msg("Error discovering canary backends for ingress")
				continue
			}

			maps.Insert(canaryBackends, maps.All(backends))
			maps.Insert(matchedIngressPaths, maps.All(matchedPaths))
		}
	}

	uniqCerts := make(map[string]*tls.CertAndStores)
	tlsOptions := make(map[string]tls.Options)
	for _, ingress := range ingresses {
		logger := log.Ctx(ctx).With().Str("ingress", ingress.Name).Str("namespace", ingress.Namespace).Logger()
		ctxIngress := logger.WithContext(ctx)

		if err := p.updateIngressStatus(ingress.Ingress); err != nil {
			logger.Error().Err(err).Msg("Error while updating ingress status")
		}

		if len(ingress.Spec.TLS) > 0 {
			if err := p.loadCertificates(ctxIngress, ingress.Ingress, uniqCerts); err != nil {
				logger.Warn().Err(err).Msg("Error loading TLS certificates defaulting to default certificate")
			}
		}

		var clientAuthTLSOptionName string
		if ingress.IngressConfig.AuthTLSSecret != nil {
			tlsOptName := provider.Normalize(ingress.Namespace + "-" + ingress.Name + "-" + *ingress.IngressConfig.AuthTLSSecret)

			if _, exists := tlsOptions[tlsOptName]; !exists {
				tlsOpt, err := p.buildClientAuthTLSOption(ingress.Namespace, ingress.IngressConfig)
				if err != nil {
					logger.Error().Err(err).Msg("Error configuring client auth TLS")
					continue
				}

				tlsOptions[tlsOptName] = tlsOpt
			}

			clientAuthTLSOptionName = tlsOptName
		}

		namedServersTransport, err := p.buildServersTransport(ctxIngress, ingress.Namespace, ingress.Name, ingress.IngressConfig)
		if err != nil {
			logger.Error().Err(err).Msg("Ignoring Ingress cannot create proxy SSL configuration")
			continue
		}

		var defaultBackendService *dynamic.Service
		if ingress.Spec.DefaultBackend != nil && ingress.Spec.DefaultBackend.Service != nil {
			var err error
			defaultBackendService, err = p.buildService(ingress.Namespace, *ingress.Spec.DefaultBackend, namedServersTransport, ingress.IngressConfig)
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
				EntryPoints: p.NonTLSEntryPoints,
				Rule:        `PathPrefix("/")`,
				// "default" stands for the default rule syntax in Traefik v3, i.e. the v3 syntax.
				RuleSyntax: "default",
				Priority:   math.MinInt32,
				Service:    defaultBackendName,
			}

			if err := p.applyMiddlewares(ingress, defaultBackendName, "", "", ingress.Spec.DefaultBackend, hosts, rt, conf, ""); err != nil {
				logger.Error().Err(err).Msg("Error applying middlewares")
			}

			conf.HTTP.Routers[defaultBackendName] = rt

			rtTLS := &dynamic.Router{
				EntryPoints: p.TLSEntryPoints,
				Rule:        `PathPrefix("/")`,
				// "default" stands for the default rule syntax in Traefik v3, i.e. the v3 syntax.
				RuleSyntax: "default",
				Priority:   math.MinInt32,
				Service:    defaultBackendName,
				TLS: &dynamic.RouterTLSConfig{
					Options: clientAuthTLSOptionName,
				},
			}

			if err := p.applyMiddlewares(ingress, defaultBackendTLSName, "", "", ingress.Spec.DefaultBackend, hosts, rtTLS, conf, ""); err != nil {
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
			if ptr.Deref(ingress.IngressConfig.SSLPassthrough, false) {
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

				service, err := p.buildPassthroughService(ingress.Namespace, *backend, ingress.IngressConfig)
				if err != nil {
					logger.Error().Err(err).Msgf("Cannot create passthrough service for %s", backend.Service.Name)
					continue
				}

				serviceName := provider.Normalize(ingress.Namespace + "-" + backend.Service.Name + "-" + portString(backend.Service.Port))
				conf.TCP.Services[serviceName] = service

				routerKey := strings.TrimPrefix(provider.Normalize(ingress.Namespace+"-"+ingress.Name+"-"+rule.Host), "-")
				conf.TCP.Routers[routerKey] = &dynamic.TCPRouter{
					EntryPoints: p.TLSEntryPoints,
					Rule:        fmt.Sprintf("HostSNI(%q)", rule.Host),
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
					EntryPoints: p.NonTLSEntryPoints,
					Rule:        buildHostRule(rule.Host),
					// "default" stands for the default rule syntax in Traefik v3, i.e. the v3 syntax.
					RuleSyntax: "default",
					Service:    key,
				}

				if err := p.applyMiddlewares(ingress, key, "", "", ingress.Spec.DefaultBackend, hosts, rt, conf, serverSnippets[rule.Host]); err != nil {
					logger.Error().Err(err).Msg("Error applying middlewares")
				}

				conf.HTTP.Routers[key] = rt

				rtTLS := &dynamic.Router{
					EntryPoints: p.TLSEntryPoints,
					Rule:        buildHostRule(rule.Host),
					// "default" stands for the default rule syntax in Traefik v3, i.e. the v3 syntax.
					RuleSyntax: "default",
					Service:    key,
					TLS: &dynamic.RouterTLSConfig{
						Options: clientAuthTLSOptionName,
					},
				}

				if err := p.applyMiddlewares(ingress, key+"-tls", "", "", ingress.Spec.DefaultBackend, hosts, rtTLS, conf, serverSnippets[rule.Host]); err != nil {
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
					logger.Error().
						Str("path", pa.Path).
						Err(err).
						Msg("Ignoring path with no service backend")

					continue
				}

				// TODO: if no service, do not add middlewares and 503.
				serviceName := provider.Normalize(ingress.Namespace + "-" + ingress.Name + "-" + pa.Backend.Service.Name + "-" + portString(pa.Backend.Service.Port))
				service, err := p.buildService(ingress.Namespace, pa.Backend, namedServersTransport, ingress.IngressConfig)
				if err != nil {
					logger.Error().
						Str("serviceName", pa.Backend.Service.Name).
						Str("servicePort", pa.Backend.Service.Port.String()).
						Err(err).
						Msg("Cannot create service")
					continue
				}

				// Retrieve the Canary backend corresponding to the service, and if one exists we are building a WRR,
				// corresponding to the canary configuration.
				var (
					wrrServiceName    string
					wrrService        *dynamic.Service
					canaryServiceName string
					canaryService     *dynamic.Service
				)
				canaryBackend, hasCanaryBackend := canaryBackends[canaryBackendKey(ingress.Namespace, *pa.Backend.Service)]
				if hasCanaryBackend {
					canaryServiceName = serviceName + "-canary"
					canaryService, err = p.buildService(ingress.Namespace, *canaryBackend.IngressBackend, namedServersTransport, ingress.IngressConfig)
					if err != nil {
						logger.Error().
							Str("serviceName", canaryBackend.IngressBackend.Service.Name).
							Str("servicePort", canaryBackend.IngressBackend.Service.Port.String()).
							Err(err).
							Msg("Cannot create canary service")
						continue
					}

					wrrServiceName = serviceName + "-wrr"
					wrrService = &dynamic.Service{
						Weighted: &dynamic.WeightedRoundRobin{
							Sticky: buildSticky(ingress.IngressConfig, "wrr"),
							Services: []dynamic.WRRService{
								{Name: serviceName, Weight: ptr.To(canaryBackend.WeightTotal - canaryBackend.Weight)},
								{Name: canaryServiceName, Weight: ptr.To(canaryBackend.Weight)},
							},
						},
					}
				}

				rt := &dynamic.Router{
					EntryPoints: p.NonTLSEntryPoints,
					Rule:        buildRule(ctxIngress, rule.Host, pa, ingress.IngressConfig, hosts, hostsWithUseRegex),
					// "default" stands for the default rule syntax in Traefik v3, i.e. the v3 syntax.
					RuleSyntax: "default",
					Service:    serviceName,
				}

				routerKey := provider.Normalize(fmt.Sprintf("%s-%s-rule-%d-path-%d", ingress.Namespace, ingress.Name, ri, pi))
				conf.HTTP.Routers[routerKey] = rt

				rtTLS := &dynamic.Router{
					EntryPoints: p.TLSEntryPoints,
					Rule:        rt.Rule,
					// "default" stands for the default rule syntax in Traefik v3, i.e. the v3 syntax.
					RuleSyntax: "default",
					Service:    rt.Service,
					TLS: &dynamic.RouterTLSConfig{
						Options: clientAuthTLSOptionName,
					},
				}

				routerKeyTLS := routerKey + "-tls"
				conf.HTTP.Routers[routerKeyTLS] = rtTLS

				conf.HTTP.Services[serviceName] = service
				if hasCanaryBackend {
					rt.Service = wrrServiceName
					rtTLS.Service = wrrServiceName
					conf.HTTP.Services[canaryServiceName] = canaryService
					conf.HTTP.Services[wrrServiceName] = wrrService
				}

				// Middlewares are applied after checking the canary backend to get the proper service.

				// HTTP Router middlewares.
				if err := p.applyMiddlewares(ingress, routerKey, pa.Path, rule.Host, &pa.Backend, hosts, rt, conf, serverSnippets[rule.Host]); err != nil {
					logger.Error().Err(err).Msg("Error applying middlewares")
				}
				// TLS Router middlewares.
				if err := p.applyMiddlewares(ingress, routerKeyTLS, pa.Path, rule.Host, &pa.Backend, hosts, rtTLS, conf, serverSnippets[rule.Host]); err != nil {
					logger.Error().Err(err).Msg("Error applying middlewares")
				}

				if hasCanaryBackend && canaryBackend.RequiresCanaryRouter() {
					canaryRouterKey := routerKey + "-canary"
					canaryRouter := &dynamic.Router{
						EntryPoints: rt.EntryPoints,
						Rule:        canaryBackend.AppendCanaryRule(rt.Rule),
						RuleSyntax:  rt.RuleSyntax,
						Service:     canaryServiceName,
						TLS:         rt.TLS,
					}
					conf.HTTP.Routers[canaryRouterKey] = canaryRouter

					if err := p.applyMiddlewares(ingress, canaryRouterKey, pa.Path, rule.Host, &pa.Backend, hosts, canaryRouter, conf, serverSnippets[rule.Host]); err != nil {
						logger.Error().Err(err).Msg("Error applying middlewares to canary router")
					}

					// default TLS router
					canaryRouterKeyTLS := canaryRouterKey + "-tls"
					canaryRouterTLS := &dynamic.Router{
						EntryPoints: rtTLS.EntryPoints,
						Rule:        canaryBackend.AppendCanaryRule(rtTLS.Rule),
						RuleSyntax:  rtTLS.RuleSyntax,
						Service:     canaryServiceName,
						TLS:         rtTLS.TLS,
					}
					conf.HTTP.Routers[canaryRouterKeyTLS] = canaryRouterTLS

					if err := p.applyMiddlewares(ingress, canaryRouterKeyTLS, pa.Path, rule.Host, &pa.Backend, hosts, canaryRouterTLS, conf, serverSnippets[rule.Host]); err != nil {
						logger.Error().Err(err).Msg("Error applying middlewares to canary router")
					}
				}

				if hasCanaryBackend && canaryBackend.RequiresNonCanaryRouter() {
					nonCanaryRouterKey := routerKey + "-non-canary"
					nonCanaryRouter := &dynamic.Router{
						EntryPoints: rt.EntryPoints,
						Rule:        canaryBackend.AppendNonCanaryRule(rt.Rule),
						RuleSyntax:  rt.RuleSyntax,
						Service:     serviceName,
						TLS:         rt.TLS,
					}
					conf.HTTP.Routers[nonCanaryRouterKey] = nonCanaryRouter

					if err := p.applyMiddlewares(ingress, nonCanaryRouterKey, pa.Path, rule.Host, &pa.Backend, hosts, nonCanaryRouter, conf, serverSnippets[rule.Host]); err != nil {
						logger.Error().Err(err).Msg("Error applying middlewares to non canary router")
					}

					// default TLS router
					nonCanaryRouterKeyTLS := nonCanaryRouterKey + "-tls"
					nonCanaryRouterTLS := &dynamic.Router{
						EntryPoints: rtTLS.EntryPoints,
						Rule:        canaryBackend.AppendNonCanaryRule(rtTLS.Rule),
						RuleSyntax:  rtTLS.RuleSyntax,
						Service:     serviceName,
						TLS:         rtTLS.TLS,
					}
					conf.HTTP.Routers[nonCanaryRouterKeyTLS] = nonCanaryRouterTLS

					if err := p.applyMiddlewares(ingress, nonCanaryRouterKeyTLS, pa.Path, rule.Host, &pa.Backend, hosts, nonCanaryRouterTLS, conf, serverSnippets[rule.Host]); err != nil {
						logger.Error().Err(err).Msg("Error applying middlewares to non canary router")
					}
				}

				if namedServersTransport != nil {
					conf.HTTP.ServersTransports[namedServersTransport.Name] = namedServersTransport.ServersTransport
				}
			}
		}
	}

	conf.TLS = &dynamic.TLSConfiguration{
		Certificates: slices.Collect(maps.Values(uniqCerts)),
	}

	if len(tlsOptions) > 0 {
		conf.TLS.Options = tlsOptions
	}

	return conf
}

func (p *Provider) isIngressValid(ingress ingress) error {
	// Discard the ingress if snippet annotations aren't allowed.
	if !p.AllowSnippetAnnotations && (ptr.Deref(ingress.IngressConfig.ServerSnippet, "") != "" ||
		ptr.Deref(ingress.IngressConfig.ConfigurationSnippet, "") != "" ||
		ptr.Deref(ingress.IngressConfig.AuthSnippet, "") != "") {
		return errors.New("snippet annotations aren't allowed when allowSnippetAnnotations is disabled")
	}

	// When strictValidatePathType is enabled, regex characters are not allowed if pathType is Prefix or Exact.
	// If one of the path is invalid, ignore the ingress.
	if p.StrictValidatePathType {
		for _, rule := range ingress.Spec.Rules {
			if rule.HTTP == nil {
				continue
			}

			for _, pa := range rule.HTTP.Paths {
				if len(pa.Path) > 0 {
					pathType := ptr.Deref(pa.PathType, netv1.PathTypePrefix)
					if pathType != netv1.PathTypeImplementationSpecific && !strictPathTypeRegexp.MatchString(pa.Path) {
						return fmt.Errorf("regex characters are not allowed for pathType %s when strictValidatePathType is enabled", pathType)
					}
				}
			}
		}
	}

	return nil
}

func (p *Provider) buildServersTransport(ctx context.Context, namespace, name string, cfg IngressConfig) (*namedServersTransport, error) {
	proxyConnectTimeout := ptr.Deref(cfg.ProxyConnectTimeout, p.ProxyConnectTimeout)
	proxyReadTimeout := ptr.Deref(cfg.ProxyReadTimeout, p.ProxyReadTimeout)
	proxySendTimeout := ptr.Deref(cfg.ProxySendTimeout, p.ProxySendTimeout)
	nst := &namedServersTransport{
		Name: provider.Normalize(namespace + "-" + name),
		ServersTransport: &dynamic.ServersTransport{
			ForwardingTimeouts: &dynamic.ForwardingTimeouts{
				DialTimeout:     ptypes.Duration(time.Duration(proxyConnectTimeout) * time.Second),
				ReadTimeout:     ptypes.Duration(time.Duration(proxyReadTimeout) * time.Second),
				WriteTimeout:    ptypes.Duration(time.Duration(proxySendTimeout) * time.Second),
				IdleConnTimeout: ptypes.Duration(time.Duration(p.UpstreamKeepaliveTimeout) * time.Second),
			},
		},
	}

	if proxyHTTPVersion := ptr.Deref(cfg.ProxyHTTPVersion, ""); proxyHTTPVersion != "" {
		switch proxyHTTPVersion {
		case "1.1":
			nst.ServersTransport.DisableHTTP2 = true
		case "1.0":
			log.Ctx(ctx).Warn().Msg("Value '1.0' is not supported with proxy-http-version, ignoring annotation")
		default:
			log.Ctx(ctx).Warn().Msgf("Invalid proxy-http-version value: %q, ignoring annotation", proxyHTTPVersion)
		}
	}

	if scheme := parseBackendProtocol(ptr.Deref(cfg.BackendProtocol, "HTTP")); scheme != "https" {
		return nst, nil
	}

	nst.ServersTransport.ServerName = ptr.Deref(cfg.ProxySSLName, ptr.Deref(cfg.ProxySSLServerName, ""))
	nst.ServersTransport.InsecureSkipVerify = strings.ToLower(ptr.Deref(cfg.ProxySSLVerify, "off")) != "on"

	if sslSecret := ptr.Deref(cfg.ProxySSLSecret, ""); sslSecret != "" {
		parts := strings.Split(sslSecret, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("malformed proxy SSL secret: %s, expected namespace/name", sslSecret)
		}

		secretNamespace, secretName := parts[0], parts[1]
		if !p.AllowCrossNamespaceResources && secretNamespace != namespace {
			return nil, fmt.Errorf("cross-namespace proxy ssl secret is not allowed: secret %s/%s is not from ingress namespace %q", secretName, secretNamespace, namespace)
		}

		blocks, err := p.certificateBlocks(secretNamespace, secretName)
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

func (p *Provider) buildService(namespace string, backend netv1.IngressBackend, nst *namedServersTransport, cfg IngressConfig) (*dynamic.Service, error) {
	backendAddresses, err := p.getBackendAddresses(namespace, backend, cfg)
	if err != nil {
		return nil, fmt.Errorf("getting backend addresses: %w", err)
	}

	lb := &dynamic.ServersLoadBalancer{}
	lb.SetDefaults()

	lb.Sticky = buildSticky(cfg, "")

	if nst != nil {
		lb.ServersTransport = nst.Name
	}

	if upstreamHashBy := ptr.Deref(cfg.UpstreamHashBy, ""); upstreamHashBy != "" {
		lb.Strategy = dynamic.BalancerStrategyHRW
		lb.NginxUpstreamHashBy = upstreamHashBy
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

func (p *Provider) buildPassthroughService(namespace string, backend netv1.IngressBackend, cfg IngressConfig) (*dynamic.TCPService, error) {
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

func getPort(service *corev1.Service, backend netv1.IngressBackend) (string, corev1.ServicePort, bool) {
	for _, p := range service.Spec.Ports {
		// A port with number 0 or an empty name is not allowed, this case is there for the default backend service.
		if (backend.Service.Port.Number == 0 && backend.Service.Port.Name == "") ||
			(backend.Service.Port.Number == p.Port || (backend.Service.Port.Name == p.Name && len(p.Name) > 0)) {
			return p.Name, p, true
		}
	}

	return "", corev1.ServicePort{}, false
}

func (p *Provider) getBackendAddresses(namespace string, backend netv1.IngressBackend, cfg IngressConfig) ([]backendAddress, error) {
	service, err := p.k8sClient.GetService(namespace, backend.Service.Name)
	if err != nil {
		return nil, fmt.Errorf("getting service: %w", err)
	}

	if p.DisableSvcExternalName && service.Spec.Type == corev1.ServiceTypeExternalName {
		return nil, errors.New("externalName services not allowed")
	}

	portName, portSpec, match := getPort(service, backend)
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

	addresses, err := p.getBackendAddressesFromEndpointSlices(namespace, backend.Service.Name, portName)
	if err != nil {
		return nil, fmt.Errorf("getting backend addresses: %w", err)
	}

	defaultBackend := ptr.Deref(cfg.DefaultBackend, "")
	if defaultBackend == "" || defaultBackend == backend.Service.Name || len(addresses) > 0 {
		return addresses, nil
	}

	serviceDefaultBackend, err := p.k8sClient.GetService(namespace, defaultBackend)
	if err != nil {
		return nil, fmt.Errorf("getting service: %w", err)
	}

	if p.DisableSvcExternalName && serviceDefaultBackend.Spec.Type == corev1.ServiceTypeExternalName {
		return nil, errors.New("externalName services not allowed")
	}

	portName, _, match = getPort(serviceDefaultBackend, netv1.IngressBackend{Service: &netv1.IngressServiceBackend{Name: defaultBackend}})
	if !match {
		return nil, errors.New("service port not found")
	}

	// If the default backend has no endpoints,
	// and if there is no default-backend-service configured,
	// the fallback with Ingress NGINX is to serve a 404,
	// but here, we will later build an empty server load-balancer which serves a 503.
	// TODO: make the built service return a 404.
	return p.getBackendAddressesFromEndpointSlices(namespace, defaultBackend, portName)
}

func (p *Provider) getBackendAddressesFromEndpointSlices(namespace, name, portName string) ([]backendAddress, error) {
	endpointSlices, err := p.k8sClient.GetEndpointSlicesForService(namespace, name)
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
			// The Serving condition allows to track if the Pod can receive traffic.
			// It is set to true when the Pod is Ready or Terminating.
			// From the go documentation, a nil value should be interpreted as "true".
			if !ptr.Deref(endpoint.Conditions.Serving, true) {
				continue
			}

			for _, address := range endpoint.Addresses {
				if _, ok := uniqAddresses[address]; ok {
					continue
				}

				uniqAddresses[address] = struct{}{}
				addresses = append(addresses, backendAddress{
					Address: net.JoinHostPort(address, strconv.Itoa(int(port))),
					Fenced:  ptr.Deref(endpoint.Conditions.Terminating, false),
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

func (p *Provider) applyMiddlewares(ingress ingress, routerKey, rulePath, ruleHost string, backend *netv1.IngressBackend, hosts map[string]bool, rt *dynamic.Router, conf *dynamic.Configuration, serverSnippet string) error {
	if p.applySSLRedirectConfiguration(ingress, routerKey, rt, conf) {
		return nil
	}

	if err := p.applyCustomHTTPErrors(ingress.Namespace, ingress.Name, routerKey, backend, ingress.IngressConfig, rt, conf); err != nil {
		return fmt.Errorf("applying custom HTTP errors: %w", err)
	}
	applyAppRootConfiguration(routerKey, ingress.IngressConfig, rt, conf)
	applyFromToWwwRedirect(hosts, ruleHost, routerKey, ingress.IngressConfig, rt, conf)
	applyRedirect(routerKey, ingress.IngressConfig, rt, conf)

	if err := p.applyBasicAuthConfiguration(ingress.Namespace, routerKey, ingress.IngressConfig, rt, conf); err != nil {
		return fmt.Errorf("applying basic auth: %w", err)
	}

	if err := p.applyBufferingConfiguration(routerKey, ingress.IngressConfig, rt, conf); err != nil {
		return fmt.Errorf("applying buffering: %w", err)
	}

	applyAllowedSourceRangeConfiguration(routerKey, ingress.IngressConfig, rt, conf)

	applyCORSConfiguration(routerKey, ingress.IngressConfig, rt, conf)

	applyRewriteTargetConfiguration(rulePath, routerKey, ingress.IngressConfig, rt, conf)

	applyUpstreamVhost(routerKey, ingress.IngressConfig, rt, conf)

	applyLimitRPMConfiguration(routerKey, ingress.IngressConfig, rt, conf)

	applyLimitRPSConfiguration(routerKey, ingress.IngressConfig, rt, conf)

	if err := p.applyAuthTLSPassCertificateToUpstream(ingress.Namespace, routerKey, ingress.IngressConfig, rt, conf); err != nil {
		return fmt.Errorf("applying auth tls pass certificate to upstream: %w", err)
	}

	if err := p.applyCustomHeaders(routerKey, ingress.IngressConfig, rt, conf); err != nil {
		return fmt.Errorf("applying custom headers: %w", err)
	}

	p.applySnippetsAndAuth(routerKey, serverSnippet, ingress.IngressConfig, rt, conf)

	p.applyRetry(routerKey, ingress.IngressConfig, rt, conf)

	if p.applyMiddlewareFunc == nil &&
		(ptr.Deref(ingress.IngressConfig.EnableModSecurity, false) ||
			ptr.Deref(ingress.IngressConfig.EnableOWASPCoreRules, false) ||
			ptr.Deref(ingress.IngressConfig.ModSecuritySnippet, "") != "" ||
			ptr.Deref(ingress.IngressConfig.ModSecurityTransactionID, "") != "") {
		return errors.New("mod-security annotations are not supported")
	}

	if p.applyMiddlewareFunc != nil {
		err := p.applyMiddlewareFunc(routerKey, rt, conf, ingress.IngressConfig)
		if err != nil {
			return fmt.Errorf("applying middleware: %w", err)
		}
	}

	return nil
}

func (p *Provider) applySnippetsAndAuth(routerName, serverSnippet string, ingressConfig IngressConfig, rt *dynamic.Router, conf *dynamic.Configuration) {
	configurationSnippet := ptr.Deref(ingressConfig.ConfigurationSnippet, "")
	authURL := ptr.Deref(ingressConfig.AuthURL, "")
	if serverSnippet == "" && configurationSnippet == "" && authURL == "" {
		return
	}

	snippetMiddlewareName := routerName + "-snippet"
	conf.HTTP.Middlewares[snippetMiddlewareName] = &dynamic.Middleware{
		Snippet: &dynamic.Snippet{
			ServerSnippet:        serverSnippet,
			ConfigurationSnippet: configurationSnippet,
		},
	}

	if authURL != "" {
		var authResponseHeaders []string
		if raw := ptr.Deref(ingressConfig.AuthResponseHeaders, ""); raw != "" {
			for h := range strings.SplitSeq(raw, ",") {
				if trimmed := strings.TrimSpace(h); trimmed != "" {
					authResponseHeaders = append(authResponseHeaders, trimmed)
				}
			}
		}

		conf.HTTP.Middlewares[snippetMiddlewareName].Snippet.Auth = &dynamic.Auth{
			Address:             authURL,
			AuthResponseHeaders: authResponseHeaders,
			AuthSigninURL:       ptr.Deref(ingressConfig.AuthSignin, ""),
			Method:              ptr.Deref(ingressConfig.AuthMethod, ""),
			Snippet:             ptr.Deref(ingressConfig.AuthSnippet, ""),
		}
	}

	rt.Middlewares = append(rt.Middlewares, snippetMiddlewareName)
}

func (p *Provider) applyCustomHTTPErrors(namespace, ingressName, routerName string, targetedService *netv1.IngressBackend, config IngressConfig, rt *dynamic.Router, conf *dynamic.Configuration) error {
	customHTTPErrors := ptr.Deref(config.CustomHTTPErrors, p.CustomHTTPErrors)
	if len(customHTTPErrors) == 0 {
		return nil
	}

	if targetedService == nil {
		return errors.New("targeted ingress backend is nil")
	}

	if targetedService.Service == nil {
		return errors.New("targeted ingress backend has no service")
	}

	serviceName := defaultBackendName
	if defaultBackend := ptr.Deref(config.DefaultBackend, ""); defaultBackend != "" {
		backend := netv1.IngressBackend{Service: &netv1.IngressServiceBackend{Name: defaultBackend}}
		service, err := p.buildService(namespace, backend, nil, config)
		if err != nil {
			return err
		}

		serviceName = fmt.Sprintf("default-backend-%s", routerName)
		conf.HTTP.Services[serviceName] = service
	} else if _, ok := conf.HTTP.Services[defaultBackendName]; !ok {
		// No default backend available (no annotation and no global default).
		// Skip the middleware — matches nginx behavior where errors pass through.
		return nil
	}

	k8sServiceName := targetedService.Service.Name
	serviceK8s, err := p.k8sClient.GetService(namespace, k8sServiceName)
	if err != nil {
		return fmt.Errorf("getting service: %w", err)
	}

	_, portSpec, ok := getPort(serviceK8s, *targetedService)
	if !ok {
		return fmt.Errorf("port not found for service %s", k8sServiceName)
	}

	customErrorMiddlewareName := routerName + "-custom-http-errors"
	headers := http.Header(map[string][]string{
		"X-Namespaces":   {namespace},
		"X-Ingress-Name": {ingressName},
		"X-Service-Name": {k8sServiceName},
		"X-Service-Port": {strconv.Itoa(int(portSpec.Port))},
	})

	conf.HTTP.Middlewares[customErrorMiddlewareName] = &dynamic.Middleware{
		Errors: &dynamic.ErrorPage{
			Status:       customHTTPErrors,
			Service:      serviceName,
			NginxHeaders: &headers,
		},
	}

	rt.Middlewares = append(rt.Middlewares, customErrorMiddlewareName)

	return nil
}

func applyLimitRPMConfiguration(routerName string, ingressConfig IngressConfig, rt *dynamic.Router, conf *dynamic.Configuration) {
	limitRPM := ptr.Deref(ingressConfig.LimitRPM, 0)
	if limitRPM <= 0 {
		return
	}

	rateLimitMiddlewareName := routerName + "-limit-rpm"
	conf.HTTP.Middlewares[rateLimitMiddlewareName] = &dynamic.Middleware{
		RateLimit: &dynamic.RateLimit{
			Average: int64(limitRPM),
			Period:  ptypes.Duration(time.Minute),
			Burst:   int64(limitRPM) * defaultLimitBurstMultiplier,
		},
	}

	rt.Middlewares = append(rt.Middlewares, rateLimitMiddlewareName)
}

func applyLimitRPSConfiguration(routerName string, ingressConfig IngressConfig, rt *dynamic.Router, conf *dynamic.Configuration) {
	limitRPS := ptr.Deref(ingressConfig.LimitRPS, 0)
	if limitRPS <= 0 {
		return
	}

	rateLimitMiddlewareName := routerName + "-limit-rps"
	conf.HTTP.Middlewares[rateLimitMiddlewareName] = &dynamic.Middleware{
		RateLimit: &dynamic.RateLimit{
			Average: int64(limitRPS),
			Period:  ptypes.Duration(time.Second),
			Burst:   int64(limitRPS) * defaultLimitBurstMultiplier,
		},
	}

	rt.Middlewares = append(rt.Middlewares, rateLimitMiddlewareName)
}

func applyRedirect(routerName string, ingressConfig IngressConfig, rt *dynamic.Router, conf *dynamic.Configuration) {
	if ingressConfig.PermanentRedirect == nil && ingressConfig.TemporalRedirect == nil {
		return
	}

	var (
		redirectURL string
		code        int
	)

	if ingressConfig.PermanentRedirect != nil {
		redirectURL = *ingressConfig.PermanentRedirect
		code = ptr.Deref(ingressConfig.PermanentRedirectCode, http.StatusMovedPermanently)

		// NGINX only accepts valid redirect codes and defaults to 301.
		if code < 300 || code > 308 {
			code = http.StatusMovedPermanently
		}
	}

	// TemporalRedirect takes precedence over the PermanentRedirect.
	if ingressConfig.TemporalRedirect != nil {
		redirectURL = *ingressConfig.TemporalRedirect
		code = ptr.Deref(ingressConfig.TemporalRedirectCode, http.StatusFound)

		// NGINX only accepts valid redirect codes and defaults to 302.
		if code < 300 || code > 308 {
			code = http.StatusFound
		}
	}

	redirectMiddlewareName := routerName + "-redirect"
	conf.HTTP.Middlewares[redirectMiddlewareName] = &dynamic.Middleware{
		RedirectRegex: &dynamic.RedirectRegex{
			Regex:       ".*",
			Replacement: redirectURL,
			StatusCode:  &code,
		},
	}
	rt.Middlewares = append(rt.Middlewares, redirectMiddlewareName)
}

func (p *Provider) applyCustomHeaders(routerName string, ingressConfig IngressConfig, rt *dynamic.Router, conf *dynamic.Configuration) error {
	customHeaders := ptr.Deref(ingressConfig.CustomHeaders, "")
	if customHeaders == "" {
		return nil
	}

	customHeadersParts := strings.Split(customHeaders, "/")
	if len(customHeadersParts) != 2 {
		return fmt.Errorf("invalid custom headers config map %q", customHeaders)
	}

	// We purposely allow cross-namespace for custom headers config maps,
	// because Ingress-Nginx does not have this limitation,
	// even if allowCrossNamespaceResources is supposed to have the same behavior for all cross-namespace resources.
	configMapNamespace := customHeadersParts[0]
	configMapName := customHeadersParts[1]

	configMap, err := p.k8sClient.GetConfigMap(configMapNamespace, configMapName)
	if err != nil {
		return fmt.Errorf("getting configMap %s: %w", customHeaders, err)
	}

	customResponseHeaders := make(map[string]string)
	for key, value := range configMap.Data {
		if !slices.Contains(p.allowedHeaders, key) {
			return fmt.Errorf("header %q is not allowed in the GlobalAllowedResponseHeaders list", key)
		}

		if !headerValueRegexp.MatchString(value) {
			return fmt.Errorf("invalid value for custom header %q: %q", key, value)
		}

		customResponseHeaders[key] = value
	}

	customHeadersMiddlewareName := routerName + "-custom-headers"
	conf.HTTP.Middlewares[customHeadersMiddlewareName] = &dynamic.Middleware{
		Headers: &dynamic.Headers{
			CustomResponseHeaders: customResponseHeaders,
		},
	}
	rt.Middlewares = append(rt.Middlewares, customHeadersMiddlewareName)

	return nil
}

// Validation identical to ingress-nginx.
var regexPathWithCapture = regexp.MustCompile(`^/?[-._~a-zA-Z0-9/$:]*$`)

func applyRewriteTargetConfiguration(rulePath, routerName string, ingressConfig IngressConfig, rt *dynamic.Router, conf *dynamic.Configuration) {
	rewrite := ptr.Deref(ingressConfig.RewriteTarget, "")
	if rewrite == "" {
		return
	}

	// Skip rewrite if the path is equal to the target.
	if rewrite == rulePath {
		return
	}

	rewriteTargetMiddlewareName := routerName + "-rewrite-target"

	// The usage of rewrite-target annotation implies the usage of regex.
	rewriteTarget := &dynamic.RewriteTarget{
		// Location modifier regex on ingress-nginx is case-insensitive.
		Regex:       "(?i)" + rulePath,
		Replacement: rewrite,
	}

	if ingressConfig.XForwardedPrefix != nil {
		if !regexPathWithCapture.MatchString(*ingressConfig.XForwardedPrefix) {
			log.Error().Msgf("Invalid x-forwarded-prefix value %q for router %q, skipping x-forwarded-prefix configuration", *ingressConfig.XForwardedPrefix, routerName)
		} else {
			rewriteTarget.XForwardedPrefix = *ingressConfig.XForwardedPrefix
		}
	}

	conf.HTTP.Middlewares[rewriteTargetMiddlewareName] = &dynamic.Middleware{
		RewriteTarget: rewriteTarget,
	}

	rt.Middlewares = append(rt.Middlewares, rewriteTargetMiddlewareName)
}

func applyAppRootConfiguration(routerName string, ingressConfig IngressConfig, rt *dynamic.Router, conf *dynamic.Configuration) {
	if ingressConfig.AppRoot == nil || !strings.HasPrefix(*ingressConfig.AppRoot, "/") {
		return
	}

	appRootMiddlewareName := routerName + "-app-root"
	conf.HTTP.Middlewares[appRootMiddlewareName] = &dynamic.Middleware{
		RedirectRegex: &dynamic.RedirectRegex{
			Regex:       `^(https?://[^/]+)/$`,
			Replacement: "$1" + *ingressConfig.AppRoot,
		},
	}

	rt.Middlewares = append(rt.Middlewares, appRootMiddlewareName)
}

func applyFromToWwwRedirect(hosts map[string]bool, ruleHost, routerName string, ingressConfig IngressConfig, rt *dynamic.Router, conf *dynamic.Configuration) {
	if ingressConfig.FromToWwwRedirect == nil || !*ingressConfig.FromToWwwRedirect {
		return
	}

	wwwType := strings.HasPrefix(ruleHost, "www.")
	wildcardType := strings.HasPrefix(ruleHost, "*.")
	bypass := wwwType && hosts[strings.TrimPrefix(ruleHost, "www.")] || !wwwType && hosts["www."+ruleHost] || wildcardType

	if bypass {
		// Wildcard host not compatible with this annotation. (limitation)
		// hosts already configured for www. and normal hosts.
		return
	}

	newRule := fmt.Sprintf("Host(%q)", fmt.Sprintf("www.%s", ruleHost))
	if wwwType {
		// if current ingress host is www.example.com, redirect from example.com => www.example.com
		host := strings.TrimPrefix(ruleHost, "www.")
		newRule = fmt.Sprintf("Host(%q)", host)
	}

	fromToWwwRedirectMiddlewareName := routerName + "-from-to-www-redirect"
	conf.HTTP.Middlewares[fromToWwwRedirectMiddlewareName] = &dynamic.Middleware{
		RedirectRegex: &dynamic.RedirectRegex{
			Regex:       `(https?)://[^/:]+(:[0-9]+)?/(.*)`,
			Replacement: fmt.Sprintf("$1://%s$2/$3", ruleHost),
			StatusCode:  ptr.To(http.StatusPermanentRedirect),
		},
	}

	wwwRedirectRouter := &dynamic.Router{
		EntryPoints: rt.EntryPoints,
		Rule:        newRule,
		Priority:    rt.Priority,
		// "default" stands for the default rule syntax in Traefik v3, i.e. the v3 syntax.
		RuleSyntax:  "default",
		Middlewares: []string{fromToWwwRedirectMiddlewareName},
		Service:     rt.Service,
		TLS:         rt.TLS,
	}
	conf.HTTP.Routers[routerName+"-from-to-www-redirect"] = wwwRedirectRouter
}

func (p *Provider) applyBasicAuthConfiguration(namespace, routerName string, ingressConfig IngressConfig, rt *dynamic.Router, conf *dynamic.Configuration) error {
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

	if !p.AllowCrossNamespaceResources && secretNamespace != namespace {
		return fmt.Errorf("cross-namespace auth secret is not allowed: secret %s/%s is not from ingress namespace %q", secretName, secretNamespace, namespace)
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

func applyCORSConfiguration(routerName string, ingressConfig IngressConfig, rt *dynamic.Router, conf *dynamic.Configuration) {
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

func applyUpstreamVhost(routerName string, ingressConfig IngressConfig, rt *dynamic.Router, conf *dynamic.Configuration) {
	if ingressConfig.UpstreamVhost == nil {
		return
	}

	vHostMiddlewareName := routerName + "-vhost"
	conf.HTTP.Middlewares[vHostMiddlewareName] = &dynamic.Middleware{
		Headers: &dynamic.Headers{
			CustomRequestHeaders: map[string]string{"Host": *ingressConfig.UpstreamVhost},
		},
	}

	rt.Middlewares = append(rt.Middlewares, vHostMiddlewareName)
}

func applyAllowedSourceRangeConfiguration(routerName string, ingressConfig IngressConfig, rt *dynamic.Router, conf *dynamic.Configuration) {
	allowedSourceRange := ptr.Deref(ingressConfig.AllowlistSourceRange, ptr.Deref(ingressConfig.WhitelistSourceRange, ""))
	if allowedSourceRange == "" {
		return
	}

	sourceRanges := strings.Split(allowedSourceRange, ",")
	for i := range sourceRanges {
		sourceRanges[i] = strings.TrimSpace(sourceRanges[i])
	}

	allowedSourceRangeMiddlewareName := routerName + "-allowed-source-range"
	conf.HTTP.Middlewares[allowedSourceRangeMiddlewareName] = &dynamic.Middleware{
		IPAllowList: &dynamic.IPAllowList{
			SourceRange: sourceRanges,
		},
	}

	rt.Middlewares = append(rt.Middlewares, allowedSourceRangeMiddlewareName)
}

func (p *Provider) applyBufferingConfiguration(routerName string, ingressConfig IngressConfig, rt *dynamic.Router, conf *dynamic.Configuration) error {
	disableRequestBuffering := !p.ProxyRequestBuffering
	if ingressConfig.ProxyRequestBuffering != nil {
		// Without value validation, lean on disabling by checking for "on", which is more likely to satisfy user input.
		disableRequestBuffering = *ingressConfig.ProxyRequestBuffering != "on"
	}

	disableResponseBuffering := !p.ProxyBuffering
	if ingressConfig.ProxyBuffering != nil {
		// Without value validation, lean on disabling by checking for "on", which is more likely to satisfy user input.
		disableResponseBuffering = *ingressConfig.ProxyBuffering != "on"
	}

	if disableRequestBuffering && disableResponseBuffering {
		return nil
	}

	buffering := &dynamic.Buffering{
		DisableRequestBuffer:  disableRequestBuffering,
		DisableResponseBuffer: disableResponseBuffering,
		MemRequestBodyBytes:   p.ClientBodyBufferSize,
		MaxRequestBodyBytes:   p.ProxyBodySize,
		MemResponseBodyBytes:  p.ProxyBufferSize * int64(p.ProxyBuffersNumber),
	}

	if !disableRequestBuffering {
		if clientBodyBufferSize := ptr.Deref(ingressConfig.ClientBodyBufferSize, ""); clientBodyBufferSize != "" {
			memRequestBodySize, err := nginxSizeToBytes(clientBodyBufferSize)
			if err != nil {
				return fmt.Errorf("client-body-buffer-size annotation has invalid value: %w", err)
			}
			buffering.MemRequestBodyBytes = memRequestBodySize
		}

		if proxyBodySize := ptr.Deref(ingressConfig.ProxyBodySize, ""); proxyBodySize != "" {
			maxRequestBody, err := nginxSizeToBytes(proxyBodySize)
			if err != nil {
				return fmt.Errorf("proxy-body-size annotation has invalid value: %w", err)
			}

			buffering.MaxRequestBodyBytes = maxRequestBody
		}
	}

	if !disableResponseBuffering {
		if ingressConfig.ProxyBufferSize != nil || ingressConfig.ProxyBuffersNumber != nil {
			bufferSize := p.ProxyBufferSize
			if proxyBufferSize := ptr.Deref(ingressConfig.ProxyBufferSize, ""); proxyBufferSize != "" {
				var err error
				if bufferSize, err = nginxSizeToBytes(proxyBufferSize); err != nil {
					return fmt.Errorf("proxy-buffer-size annotation has invalid value: %w", err)
				}
			}

			buffering.MemResponseBodyBytes = bufferSize * int64(ptr.Deref(ingressConfig.ProxyBuffersNumber, p.ProxyBuffersNumber))
		}

		proxyMaxTempFileSize := defaultProxyMaxTempFileSize
		if ingressConfig.ProxyMaxTempFileSize != nil {
			var err error
			if proxyMaxTempFileSize, err = nginxSizeToBytes(*ingressConfig.ProxyMaxTempFileSize); err != nil {
				return fmt.Errorf("proxy-max-temp-file-size annotation has invalid value: %w", err)
			}
		}

		buffering.MaxResponseBodyBytes = buffering.MemResponseBodyBytes + proxyMaxTempFileSize
	}

	bufferingMiddlewareName := routerName + "-buffering"
	conf.HTTP.Middlewares[bufferingMiddlewareName] = &dynamic.Middleware{
		Buffering: buffering,
	}
	rt.Middlewares = append(rt.Middlewares, bufferingMiddlewareName)

	return nil
}

func (p *Provider) applySSLRedirectConfiguration(ingress ingress, routerName string, rt *dynamic.Router, conf *dynamic.Configuration) bool {
	// Only apply SSL redirect on HTTP routers when the ingress has a TLS section.
	if rt.TLS != nil || ingress.Spec.TLS == nil {
		return false
	}

	sslRedirect := ptr.Deref(ingress.IngressConfig.SSLRedirect, false)
	forceSSLRedirect := ptr.Deref(ingress.IngressConfig.ForceSSLRedirect, false)

	// If either forceSSLRedirect or sslRedirect are enabled,
	// the HTTP router needs to redirect to HTTPS.
	if forceSSLRedirect || sslRedirect {
		redirectMiddlewareName := routerName + "-redirect-scheme"
		conf.HTTP.Middlewares[redirectMiddlewareName] = &dynamic.Middleware{
			RedirectScheme: &dynamic.RedirectScheme{
				Scheme:                 "https",
				ForcePermanentRedirect: true,
			},
		}
		rt.Middlewares = []string{redirectMiddlewareName}
		rt.Service = "noop@internal"
		return true
	}

	// An Ingress that is not forcing sslRedirect and has no TLS configuration does not redirect,
	// even if sslRedirect is enabled.
	return false
}

// discoverCanaryBackends checks if the canary ingress is matching any of the existing ingress rules,
// and if so returns the canary backends to be applied, and the paths of the matched ingress rules.
func (p *Provider) discoverCanaryBackends(namespace string, canaryIngressRule netv1.IngressRule, canaryIngressConfig IngressConfig, ingressPaths map[string]ingressPath, matchedIngressPaths map[string]struct{}) (map[string]*canaryBackend, map[string]struct{}, error) {
	canaryPaths := make(map[string]struct{})          // indexed by namespace+host+path+pathType.
	canaryBackends := make(map[string]*canaryBackend) // indexed by service namespace+name+port.

	for _, pa := range canaryIngressRule.HTTP.Paths {
		// The canary ingress is not matching an existing ingress rule,
		// we cannot apply the whole canary configuration.
		pathKey := ingressPathKey(namespace, canaryIngressRule.Host, pa)
		pathType := ptr.Deref(pa.PathType, netv1.PathTypePrefix)
		ingressPath, ok := ingressPaths[pathKey]
		if !ok {
			return nil, nil, fmt.Errorf("canary ingress does not match Ingress rule host=%s, path=%s, pathType=%s", canaryIngressRule.Host, pa.Path, pathType)
		}

		// A canary ingress is already matching this ingress rule,
		// we cannot apply the whole canary configuration.
		if _, ok := matchedIngressPaths[pathKey]; ok {
			return nil, nil, fmt.Errorf("a canary ingress is already matching Ingress rule host=%s, path=%s, pathType=%s", canaryIngressRule.Host, pa.Path, pathType)
		}

		if pa.Backend.Service == nil {
			continue
		}

		// In case the service cannot be retrieved, or it has no endpoints we should ignore this matching canary rule.
		// Here we are using the original Ingress configuration as a canary ingress inherit the original Ingress configuration.
		if addresses, err := p.getBackendAddresses(namespace, pa.Backend, ingressPath.IngressConfig); err != nil || len(addresses) == 0 {
			continue
		}

		canaryPaths[pathKey] = struct{}{}

		weightTotal := max(ptr.Deref(canaryIngressConfig.CanaryWeightTotal, 0), 100)       // the minimum value in NGINX is 100.
		weight := min(max(ptr.Deref(canaryIngressConfig.CanaryWeight, 0), 0), weightTotal) // weight cannot be negative, and cannot be greater than weightTotal.
		canaryBackends[canaryBackendKey(namespace, *ingressPath.Backend.Service)] = &canaryBackend{
			IngressBackend: &pa.Backend,
			Weight:         weight,
			WeightTotal:    weightTotal,
			Header:         ptr.Deref(canaryIngressConfig.CanaryHeader, ""),
			HeaderValue:    ptr.Deref(canaryIngressConfig.CanaryHeaderValue, ""),
			HeaderPattern:  ptr.Deref(canaryIngressConfig.CanaryHeaderPattern, ""),
			Cookie:         ptr.Deref(canaryIngressConfig.CanaryCookie, ""),
		}
	}

	return canaryBackends, canaryPaths, nil
}

func (p *Provider) applyAuthTLSPassCertificateToUpstream(ingressNamespace string, routerName string, ingressConfig IngressConfig, rt *dynamic.Router, conf *dynamic.Configuration) error {
	if !ptr.Deref(ingressConfig.AuthTLSPassCertificateToUpstream, false) {
		return nil
	}
	// Passing TLS client certificates upstream only applies to TLS routers.
	if rt.TLS == nil {
		return nil
	}
	if ingressConfig.AuthTLSSecret == nil {
		return errors.New("auth-tls-pass-certificate-to-upstream requires auth-tls-secret to be configured")
	}

	verifyClient := clientAuthTypeFromString(ingressConfig.AuthTLSVerifyClient)

	var caFiles []types.FileOrContent
	if verifyClient == tls.RequestClientCert {
		blocks, err := p.loadCertBlock(ingressNamespace, ingressConfig)
		if err != nil {
			return fmt.Errorf("reading client certificate: %w", err)
		}
		caFiles = []types.FileOrContent{*blocks.CA}
	}

	passCertificateToUpstreamMiddlewareName := routerName + "-pass-certificate-to-upstream"
	conf.HTTP.Middlewares[passCertificateToUpstreamMiddlewareName] = &dynamic.Middleware{
		AuthTLSPassCertificateToUpstream: &dynamic.AuthTLSPassCertificateToUpstream{
			ClientAuthType: verifyClient,
			CAFiles:        caFiles,
		},
	}
	rt.Middlewares = append(rt.Middlewares, passCertificateToUpstreamMiddlewareName)

	return nil
}

func (p *Provider) applyRetry(routerName string, ingressConfig IngressConfig, rt *dynamic.Router, conf *dynamic.Configuration) {
	attempts := ptr.Deref(ingressConfig.ProxyNextUpstreamTries, p.ProxyNextUpstreamTries)
	// Safeguard to deactivate retry when the value is less than 0.
	if attempts < 0 {
		return
	}

	proxyNextUpstream := ptr.Deref(ingressConfig.ProxyNextUpstream, p.ProxyNextUpstream)
	if proxyNextUpstream == "" {
		return
	}

	retryConditions := strings.Fields(proxyNextUpstream)
	// "off" disables the retry entirely.
	if slices.Contains(retryConditions, "off") {
		return
	}

	// proxy-next-upstream-tries = 0 on NGINX means unlimited tries, which maps to try every available server.
	// To avoid infinite retries, put the number of servers as the attempts limit.
	if attempts == 0 {
		svc, ok := conf.HTTP.Services[rt.Service]
		if !ok || svc.LoadBalancer == nil {
			return
		}
		serverCount := len(svc.LoadBalancer.Servers)
		attempts = serverCount
	}

	retryConfig := &dynamic.Retry{
		Attempts: attempts,
	}

	// Disable network error retry if no error nor timeout present on the configuration.
	hasError := slices.Contains(retryConditions, "error")
	hasTimeout := slices.Contains(retryConditions, "timeout")
	if !hasError && !hasTimeout {
		retryConfig.DisableRetryOnNetworkError = true
	}

	// HTTP status codes condition.
	var statusCodes []string
	for _, statusCode := range retryConditions {
		if code, ok := strings.CutPrefix(statusCode, "http_"); ok {
			statusCodes = append(statusCodes, code)
		}
	}
	if len(statusCodes) > 0 {
		retryConfig.Status = statusCodes
	}

	// Non-idempotent configuration.
	if slices.Contains(retryConditions, "non_idempotent") {
		retryConfig.RetryNonIdempotentMethod = true
	}

	// Timeout configuration.
	timeout := ptr.Deref(ingressConfig.ProxyNextUpstreamTimeout, p.ProxyNextUpstreamTimeout)
	if timeout > 0 {
		retryConfig.Timeout = ptypes.Duration(time.Duration(timeout) * time.Second)
	}

	retryMiddlewareName := routerName + "-retry"
	conf.HTTP.Middlewares[retryMiddlewareName] = &dynamic.Middleware{
		Retry: retryConfig,
	}
	rt.Middlewares = append(rt.Middlewares, retryMiddlewareName)
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
	for rawLine := range strings.SplitSeq(string(authFileContent), "\n") {
		line := strings.TrimSpace(rawLine)
		if line != "" && !strings.HasPrefix(line, "#") {
			users = append(users, line)
		}
	}

	return users, nil
}

func buildRule(ctx context.Context, host string, pa netv1.HTTPIngressPath, config IngressConfig, allHosts map[string]bool, hostsWithUseRegex map[string]bool) string {
	var rules []string
	if host != "" {
		hosts := []string{host}
		if config.ServerAlias != nil {
			for _, alias := range *config.ServerAlias {
				if _, ok := allHosts[strings.ToLower(alias)]; ok {
					log.Ctx(ctx).Debug().
						Str("alias", alias).
						Msg("Skipping server-alias because it is already defined as a host in another Ingress")
					continue
				}
				hosts = append(hosts, alias)
			}
		}

		var hostRules []string
		for _, h := range hosts {
			hostRules = append(hostRules, buildHostRule(h))
		}

		if len(hostRules) > 1 {
			rules = append(rules, "("+strings.Join(hostRules, " || ")+")")
		} else {
			rules = append(rules, hostRules[0])
		}
	}

	if len(pa.Path) > 0 {
		pathType := ptr.Deref(pa.PathType, netv1.PathTypePrefix)
		if pathType == netv1.PathTypeImplementationSpecific {
			pathType = netv1.PathTypePrefix
		}

		useRegex := hostsWithUseRegex[host]

		switch pathType {
		case netv1.PathTypeExact:
			rules = append(rules, fmt.Sprintf("Path(%q)", pa.Path))
		case netv1.PathTypePrefix:
			if useRegex {
				rules = append(rules, fmt.Sprintf("PathRegexp(%q)", fmt.Sprintf("(?i)^%s", pa.Path)))
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
		return fmt.Sprintf("HostRegexp(%q)", fmt.Sprintf("^%s$", host))
	}

	return fmt.Sprintf("Host(%q)", host)
}

// buildPrefixRule is a helper function to build a path prefix rule that matches path prefix split by `/`.
// For example, the paths `/abc`, `/abc/`, and `/abc/def` would all match the prefix `/abc`,
// but the path `/abcd` would not. See TestStrictPrefixMatchingRule() for more examples.
//
// "PathPrefix" in Kubernetes Gateway API is semantically equivalent to the "Prefix" path type in the
// Kubernetes Ingress API.
func buildPrefixRule(path string) string {
	if path == "/" {
		return `PathPrefix("/")`
	}

	path = strings.TrimSuffix(path, "/")
	return fmt.Sprintf("(Path(%q) || PathPrefix(%q))", path, fmt.Sprintf("%s/", path))
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

func (p *Provider) loadCertBlock(ingressNamespace string, config IngressConfig) (*certBlocks, error) {
	secretParts := strings.SplitN(*config.AuthTLSSecret, "/", 2)
	if len(secretParts) != 2 {
		return nil, errors.New("auth-tls-secret is not in a correct namespace/name format")
	}

	// Expected format: namespace/name.
	secretNamespace := secretParts[0]
	secretName := secretParts[1]

	if secretNamespace == "" {
		return nil, errors.New("auth-tls-secret has empty namespace")
	}
	if secretName == "" {
		return nil, errors.New("auth-tls-secret has empty name")
	}

	// Verify when cross-namespace secrets are not allowed.
	if !p.AllowCrossNamespaceResources && secretNamespace != ingressNamespace {
		return nil, fmt.Errorf("cross-namespace auth-tls-secret is not supported: secret namespace %q does not match ingress namespace %q", secretNamespace, ingressNamespace)
	}

	blocks, err := p.certificateBlocks(secretNamespace, secretName)
	if err != nil {
		return nil, fmt.Errorf("reading client certificate: %w", err)
	}

	if blocks.CA == nil {
		return nil, errors.New("secret does not contain a CA certificate")
	}

	return blocks, nil
}

// clientAuthTypeFromString maps an ingress-nginx auth-tls-verify-client value to the corresponding ClientAuthType.
// Default is "on" (RequireAndVerifyClientCert) when verifyClient is nil.
func clientAuthTypeFromString(verifyClient *string) string {
	if verifyClient == nil {
		return tls.RequireAndVerifyClientCert
	}
	switch *verifyClient {
	// off means that client certificate is not requested and no verification will be passed.
	case "off":
		return tls.NoClientCert
	// optional means that the client certificate is requested, but not required.
	// If the certificate is present, it needs to be verified.
	case "optional":
		return tls.VerifyClientCertIfGiven
	// optional_no_ca means that the client certificate is requested, but does not require it to be signed by a trusted CA certificate.
	case "optional_no_ca":
		return tls.RequestClientCert
	default:
		return tls.RequireAndVerifyClientCert
	}
}

func (p *Provider) buildClientAuthTLSOption(ingressNamespace string, config IngressConfig) (tls.Options, error) {
	blocks, err := p.loadCertBlock(ingressNamespace, config)
	if err != nil {
		return tls.Options{}, fmt.Errorf("reading client certificate: %w", err)
	}

	clientAuthType := clientAuthTypeFromString(config.AuthTLSVerifyClient)

	tlsOpt := tls.Options{}
	tlsOpt.SetDefaults()
	tlsOpt.ClientAuth = tls.ClientAuth{
		CAFiles:        []types.FileOrContent{*blocks.CA},
		ClientAuthType: clientAuthType,
	}

	return tlsOpt, nil
}

// nginxSizeToBytes convert nginx size to memory bytes as defined in https://nginx.org/en/docs/syntax.html.
func nginxSizeToBytes(nginxSize string) (int64, error) {
	units := map[string]int64{
		"g": 1024 * 1024 * 1024,
		"m": 1024 * 1024,
		"k": 1024,
		"b": 1,
		"":  1,
	}

	if !nginxSizeRegexp.MatchString(nginxSize) {
		return 0, fmt.Errorf("unable to parse number %s", nginxSize)
	}
	size := nginxSizeRegexp.FindStringSubmatch(nginxSize)
	bytes, err := strconv.ParseInt(size[1], 10, 64)
	if err != nil {
		return 0, err
	}
	return bytes * units[strings.ToLower(size[2])], nil
}

// buildSticky returns a Sticky configuration if the affinity configuration is set to "cookie" and nil otherwise.
// It also appends the given nameSuffix to the cookie name if not empty.
func buildSticky(cfg IngressConfig, nameSuffix string) *dynamic.Sticky {
	if ptr.Deref(cfg.Affinity, "") != "cookie" {
		return nil
	}

	name := ptr.Deref(cfg.SessionCookieName, "INGRESSCOOKIE")
	if nameSuffix != "" {
		name += "-" + nameSuffix
	}

	return &dynamic.Sticky{
		Cookie: &dynamic.Cookie{
			Name:     name,
			Secure:   ptr.Deref(cfg.SessionCookieSecure, false),
			HTTPOnly: true, // Default value in Nginx.
			SameSite: strings.ToLower(ptr.Deref(cfg.SessionCookieSameSite, "")),
			MaxAge:   ptr.Deref(cfg.SessionCookieMaxAge, 0),
			Expires:  ptr.Deref(cfg.SessionCookieExpires, 0),
			Path:     ptr.To(ptr.Deref(cfg.SessionCookiePath, "/")),
			Domain:   ptr.Deref(cfg.SessionCookieDomain, ""),
		},
	}
}

func ingressPathKey(namespace, host string, pa netv1.HTTPIngressPath) string {
	pathType := ptr.Deref(pa.PathType, netv1.PathTypePrefix)
	return namespace + "/" + host + pa.Path + "/" + string(pathType)
}

func canaryBackendKey(namespace string, backend netv1.IngressServiceBackend) string {
	return namespace + "/" + backend.Name + "/" + portString(backend.Port)
}

func portString(port netv1.ServiceBackendPort) string {
	if port.Name == "" {
		return strconv.Itoa(int(port.Number))
	}
	return port.Name
}
