package static

import (
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	legolog "github.com/go-acme/lego/v4/log"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/ping"
	acmeprovider "github.com/traefik/traefik/v3/pkg/provider/acme"
	"github.com/traefik/traefik/v3/pkg/provider/consulcatalog"
	"github.com/traefik/traefik/v3/pkg/provider/docker"
	"github.com/traefik/traefik/v3/pkg/provider/ecs"
	"github.com/traefik/traefik/v3/pkg/provider/file"
	"github.com/traefik/traefik/v3/pkg/provider/http"
	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd"
	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/gateway"
	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/ingress"
	ingressnginx "github.com/traefik/traefik/v3/pkg/provider/kubernetes/ingress-nginx"
	"github.com/traefik/traefik/v3/pkg/provider/kv/consul"
	"github.com/traefik/traefik/v3/pkg/provider/kv/etcd"
	"github.com/traefik/traefik/v3/pkg/provider/kv/redis"
	"github.com/traefik/traefik/v3/pkg/provider/kv/zk"
	"github.com/traefik/traefik/v3/pkg/provider/nomad"
	"github.com/traefik/traefik/v3/pkg/provider/rest"
	"github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
)

const (
	// DefaultInternalEntryPointName the name of the default internal entry point.
	DefaultInternalEntryPointName = "traefik"

	// DefaultGraceTimeout controls how long Traefik serves pending requests
	// prior to shutting down.
	DefaultGraceTimeout = 10 * time.Second

	// DefaultIdleTimeout before closing an idle connection.
	DefaultIdleTimeout = 180 * time.Second

	// DefaultReadTimeout defines the default maximum duration for reading the entire request, including the body.
	DefaultReadTimeout = 60 * time.Second

	// DefaultAcmeCAServer is the default ACME API endpoint.
	DefaultAcmeCAServer = "https://acme-v02.api.letsencrypt.org/directory"

	// DefaultUDPTimeout defines how long to wait by default on an idle session,
	// before releasing all resources related to that session.
	DefaultUDPTimeout = 3 * time.Second
)

// Configuration is the static configuration.
type Configuration struct {
	Global *Global `description:"Global configuration options" json:"global,omitempty" toml:"global,omitempty" yaml:"global,omitempty" export:"true"`

	ServersTransport    *ServersTransport    `description:"Servers default transport." json:"serversTransport,omitempty" toml:"serversTransport,omitempty" yaml:"serversTransport,omitempty" export:"true"`
	TCPServersTransport *TCPServersTransport `description:"TCP servers default transport." json:"tcpServersTransport,omitempty" toml:"tcpServersTransport,omitempty" yaml:"tcpServersTransport,omitempty" export:"true"`
	EntryPoints         EntryPoints          `description:"Entry points definition." json:"entryPoints,omitempty" toml:"entryPoints,omitempty" yaml:"entryPoints,omitempty" export:"true"`
	Providers           *Providers           `description:"Providers configuration." json:"providers,omitempty" toml:"providers,omitempty" yaml:"providers,omitempty" export:"true"`

	API     *API           `description:"Enable api/dashboard." json:"api,omitempty" toml:"api,omitempty" yaml:"api,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	Metrics *types.Metrics `description:"Enable a metrics exporter." json:"metrics,omitempty" toml:"metrics,omitempty" yaml:"metrics,omitempty" export:"true"`
	Ping    *ping.Handler  `description:"Enable ping." json:"ping,omitempty" toml:"ping,omitempty" yaml:"ping,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`

	Log       *types.TraefikLog `description:"Traefik log settings." json:"log,omitempty" toml:"log,omitempty" yaml:"log,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	AccessLog *types.AccessLog  `description:"Access log settings." json:"accessLog,omitempty" toml:"accessLog,omitempty" yaml:"accessLog,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	Tracing   *Tracing          `description:"Tracing configuration." json:"tracing,omitempty" toml:"tracing,omitempty" yaml:"tracing,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`

	HostResolver *types.HostResolverConfig `description:"Enable CNAME Flattening." json:"hostResolver,omitempty" toml:"hostResolver,omitempty" yaml:"hostResolver,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`

	CertificatesResolvers map[string]CertificateResolver `description:"Certificates resolvers configuration." json:"certificatesResolvers,omitempty" toml:"certificatesResolvers,omitempty" yaml:"certificatesResolvers,omitempty" export:"true"`

	Experimental *Experimental `description:"Experimental features." json:"experimental,omitempty" toml:"experimental,omitempty" yaml:"experimental,omitempty" export:"true"`

	// Deprecated: Please do not use this field.
	Core *Core `description:"Core controls." json:"core,omitempty" toml:"core,omitempty" yaml:"core,omitempty" export:"true"`

	Spiffe *SpiffeClientConfig `description:"SPIFFE integration configuration." json:"spiffe,omitempty" toml:"spiffe,omitempty" yaml:"spiffe,omitempty" export:"true"`

	OCSP *tls.OCSPConfig `description:"OCSP configuration." json:"ocsp,omitempty" toml:"ocsp,omitempty" yaml:"ocsp,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
}

// Core configures Traefik core behavior.
type Core struct {
	// Deprecated: Please do not use this field and rewrite the router rules to use the v3 syntax.
	DefaultRuleSyntax string `description:"Defines the rule parser default syntax (v2 or v3)" json:"defaultRuleSyntax,omitempty" toml:"defaultRuleSyntax,omitempty" yaml:"defaultRuleSyntax,omitempty"`
}

// SetDefaults sets the default values.
func (c *Core) SetDefaults() {
	c.DefaultRuleSyntax = "v3"
}

// SpiffeClientConfig defines the SPIFFE client configuration.
type SpiffeClientConfig struct {
	WorkloadAPIAddr string `description:"Defines the workload API address." json:"workloadAPIAddr,omitempty" toml:"workloadAPIAddr,omitempty" yaml:"workloadAPIAddr,omitempty"`
}

// CertificateResolver contains the configuration for the different types of certificates resolver.
type CertificateResolver struct {
	ACME      *acmeprovider.Configuration `description:"Enables ACME (Let's Encrypt) automatic SSL." json:"acme,omitempty" toml:"acme,omitempty" yaml:"acme,omitempty" export:"true"`
	Tailscale *struct{}                   `description:"Enables Tailscale certificate resolution." json:"tailscale,omitempty" toml:"tailscale,omitempty" yaml:"tailscale,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
}

// Global holds the global configuration.
type Global struct {
	CheckNewVersion    bool `description:"Periodically check if a new version has been released." json:"checkNewVersion,omitempty" toml:"checkNewVersion,omitempty" yaml:"checkNewVersion,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	SendAnonymousUsage bool `description:"Periodically send anonymous usage statistics. If the option is not specified, it will be disabled by default." json:"sendAnonymousUsage,omitempty" toml:"sendAnonymousUsage,omitempty" yaml:"sendAnonymousUsage,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
}

// ServersTransport options to configure communication between Traefik and the servers.
type ServersTransport struct {
	InsecureSkipVerify  bool                  `description:"Disable SSL certificate verification." json:"insecureSkipVerify,omitempty" toml:"insecureSkipVerify,omitempty" yaml:"insecureSkipVerify,omitempty" export:"true"`
	RootCAs             []types.FileOrContent `description:"Add cert file for self-signed certificate." json:"rootCAs,omitempty" toml:"rootCAs,omitempty" yaml:"rootCAs,omitempty"`
	MaxIdleConnsPerHost int                   `description:"If non-zero, controls the maximum idle (keep-alive) to keep per-host. If zero, DefaultMaxIdleConnsPerHost is used" json:"maxIdleConnsPerHost,omitempty" toml:"maxIdleConnsPerHost,omitempty" yaml:"maxIdleConnsPerHost,omitempty" export:"true"`
	ForwardingTimeouts  *ForwardingTimeouts   `description:"Timeouts for requests forwarded to the backend servers." json:"forwardingTimeouts,omitempty" toml:"forwardingTimeouts,omitempty" yaml:"forwardingTimeouts,omitempty" export:"true"`
	Spiffe              *Spiffe               `description:"Defines the SPIFFE configuration." json:"spiffe,omitempty" toml:"spiffe,omitempty" yaml:"spiffe,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
}

// Spiffe holds the SPIFFE configuration.
type Spiffe struct {
	IDs         []string `description:"Defines the allowed SPIFFE IDs (takes precedence over the SPIFFE TrustDomain)." json:"ids,omitempty" toml:"ids,omitempty" yaml:"ids,omitempty"`
	TrustDomain string   `description:"Defines the allowed SPIFFE trust domain." json:"trustDomain,omitempty" toml:"trustDomain,omitempty" yaml:"trustDomain,omitempty"`
}

// TCPServersTransport options to configure communication between Traefik and the servers.
type TCPServersTransport struct {
	DialKeepAlive ptypes.Duration `description:"Defines the interval between keep-alive probes for an active network connection. If zero, keep-alive probes are sent with a default value (currently 15 seconds), if supported by the protocol and operating system. Network protocols or operating systems that do not support keep-alives ignore this field. If negative, keep-alive probes are disabled" json:"dialKeepAlive,omitempty" toml:"dialKeepAlive,omitempty" yaml:"dialKeepAlive,omitempty" export:"true"`
	DialTimeout   ptypes.Duration `description:"Defines the amount of time to wait until a connection to a backend server can be established. If zero, no timeout exists." json:"dialTimeout,omitempty" toml:"dialTimeout,omitempty" yaml:"dialTimeout,omitempty" export:"true"`
	// TerminationDelay, corresponds to the deadline that the proxy sets, after one
	// of its connected peers indicates it has closed the writing capability of its
	// connection, to close the reading capability as well, hence fully terminating the
	// connection. It is a duration in milliseconds, defaulting to 100. A negative value
	// means an infinite deadline (i.e. the reading capability is never closed).
	TerminationDelay ptypes.Duration  `description:"Defines the delay to wait before fully terminating the connection, after one connected peer has closed its writing capability." json:"terminationDelay,omitempty" toml:"terminationDelay,omitempty" yaml:"terminationDelay,omitempty" export:"true"`
	TLS              *TLSClientConfig `description:"Defines the TLS configuration." json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" label:"allowEmpty" file:"allowEmpty" kv:"allowEmpty" export:"true"`
}

// TLSClientConfig options to configure TLS communication between Traefik and the servers.
type TLSClientConfig struct {
	InsecureSkipVerify bool                  `description:"Disables SSL certificate verification." json:"insecureSkipVerify,omitempty" toml:"insecureSkipVerify,omitempty" yaml:"insecureSkipVerify,omitempty" export:"true"`
	RootCAs            []types.FileOrContent `description:"Defines a list of CA secret used to validate self-signed certificate" json:"rootCAs,omitempty" toml:"rootCAs,omitempty" yaml:"rootCAs,omitempty"`
	Spiffe             *Spiffe               `description:"Defines the SPIFFE TLS configuration." json:"spiffe,omitempty" toml:"spiffe,omitempty" yaml:"spiffe,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
}

// API holds the API configuration.
type API struct {
	BasePath           string `description:"Defines the base path where the API and Dashboard will be exposed." json:"basePath,omitempty" toml:"basePath,omitempty" yaml:"basePath,omitempty" export:"true"`
	Insecure           bool   `description:"Activate API directly on the entryPoint named traefik." json:"insecure,omitempty" toml:"insecure,omitempty" yaml:"insecure,omitempty" export:"true"`
	Dashboard          bool   `description:"Activate dashboard." json:"dashboard,omitempty" toml:"dashboard,omitempty" yaml:"dashboard,omitempty" export:"true"`
	Debug              bool   `description:"Enable additional endpoints for debugging and profiling." json:"debug,omitempty" toml:"debug,omitempty" yaml:"debug,omitempty" export:"true"`
	DisableDashboardAd bool   `description:"Disable ad in the dashboard." json:"disableDashboardAd,omitempty" toml:"disableDashboardAd,omitempty" yaml:"disableDashboardAd,omitempty" export:"true"`
	// TODO: Re-enable statistics
	// Statistics      *types.Statistics `description:"Enable more detailed statistics." json:"statistics,omitempty" toml:"statistics,omitempty" yaml:"statistics,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
}

// SetDefaults sets the default values.
func (a *API) SetDefaults() {
	a.BasePath = "/"
	a.Dashboard = true
}

// RespondingTimeouts contains timeout configurations for incoming requests to the Traefik instance.
type RespondingTimeouts struct {
	ReadTimeout  ptypes.Duration `description:"ReadTimeout is the maximum duration for reading the entire request, including the body. If zero, no timeout is set." json:"readTimeout,omitempty" toml:"readTimeout,omitempty" yaml:"readTimeout,omitempty" export:"true"`
	WriteTimeout ptypes.Duration `description:"WriteTimeout is the maximum duration before timing out writes of the response. If zero, no timeout is set." json:"writeTimeout,omitempty" toml:"writeTimeout,omitempty" yaml:"writeTimeout,omitempty" export:"true"`
	IdleTimeout  ptypes.Duration `description:"IdleTimeout is the maximum amount duration an idle (keep-alive) connection will remain idle before closing itself. If zero, no timeout is set." json:"idleTimeout,omitempty" toml:"idleTimeout,omitempty" yaml:"idleTimeout,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (a *RespondingTimeouts) SetDefaults() {
	a.ReadTimeout = ptypes.Duration(DefaultReadTimeout)
	a.IdleTimeout = ptypes.Duration(DefaultIdleTimeout)
}

// ForwardingTimeouts contains timeout configurations for forwarding requests to the backend servers.
type ForwardingTimeouts struct {
	DialTimeout           ptypes.Duration `description:"The amount of time to wait until a connection to a backend server can be established. If zero, no timeout exists." json:"dialTimeout,omitempty" toml:"dialTimeout,omitempty" yaml:"dialTimeout,omitempty" export:"true"`
	ResponseHeaderTimeout ptypes.Duration `description:"The amount of time to wait for a server's response headers after fully writing the request (including its body, if any). If zero, no timeout exists." json:"responseHeaderTimeout,omitempty" toml:"responseHeaderTimeout,omitempty" yaml:"responseHeaderTimeout,omitempty" export:"true"`
	IdleConnTimeout       ptypes.Duration `description:"The maximum period for which an idle HTTP keep-alive connection will remain open before closing itself" json:"idleConnTimeout,omitempty" toml:"idleConnTimeout,omitempty" yaml:"idleConnTimeout,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (f *ForwardingTimeouts) SetDefaults() {
	f.DialTimeout = ptypes.Duration(30 * time.Second)
	f.IdleConnTimeout = ptypes.Duration(90 * time.Second)
}

// LifeCycle contains configurations relevant to the lifecycle (such as the shutdown phase) of Traefik.
type LifeCycle struct {
	RequestAcceptGraceTimeout ptypes.Duration `description:"Duration to keep accepting requests before Traefik initiates the graceful shutdown procedure." json:"requestAcceptGraceTimeout,omitempty" toml:"requestAcceptGraceTimeout,omitempty" yaml:"requestAcceptGraceTimeout,omitempty" export:"true"`
	GraceTimeOut              ptypes.Duration `description:"Duration to give active requests a chance to finish before Traefik stops." json:"graceTimeOut,omitempty" toml:"graceTimeOut,omitempty" yaml:"graceTimeOut,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (a *LifeCycle) SetDefaults() {
	a.GraceTimeOut = ptypes.Duration(DefaultGraceTimeout)
}

// Tracing holds the tracing configuration.
type Tracing struct {
	ServiceName             string             `description:"Defines the service name resource attribute." json:"serviceName,omitempty" toml:"serviceName,omitempty" yaml:"serviceName,omitempty" export:"true"`
	ResourceAttributes      map[string]string  `description:"Defines additional resource attributes (key:value)." json:"resourceAttributes,omitempty" toml:"resourceAttributes,omitempty" yaml:"resourceAttributes,omitempty" export:"true"`
	CapturedRequestHeaders  []string           `description:"Request headers to add as attributes for server and client spans." json:"capturedRequestHeaders,omitempty" toml:"capturedRequestHeaders,omitempty" yaml:"capturedRequestHeaders,omitempty" export:"true"`
	CapturedResponseHeaders []string           `description:"Response headers to add as attributes for server and client spans." json:"capturedResponseHeaders,omitempty" toml:"capturedResponseHeaders,omitempty" yaml:"capturedResponseHeaders,omitempty" export:"true"`
	SafeQueryParams         []string           `description:"Query params to not redact." json:"safeQueryParams,omitempty" toml:"safeQueryParams,omitempty" yaml:"safeQueryParams,omitempty" export:"true"`
	SampleRate              float64            `description:"Sets the rate between 0.0 and 1.0 of requests to trace." json:"sampleRate,omitempty" toml:"sampleRate,omitempty" yaml:"sampleRate,omitempty" export:"true"`
	AddInternals            bool               `description:"Enables tracing for internal services (ping, dashboard, etc...)." json:"addInternals,omitempty" toml:"addInternals,omitempty" yaml:"addInternals,omitempty" export:"true"`
	OTLP                    *types.OTelTracing `description:"Settings for OpenTelemetry." json:"otlp,omitempty" toml:"otlp,omitempty" yaml:"otlp,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`

	// Deprecated: please use ResourceAttributes instead.
	GlobalAttributes map[string]string `description:"(Deprecated) Defines additional resource attributes (key:value)." json:"globalAttributes,omitempty" toml:"globalAttributes,omitempty" yaml:"globalAttributes,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (t *Tracing) SetDefaults() {
	t.ServiceName = "traefik"
	t.SampleRate = 1.0

	t.OTLP = &types.OTelTracing{}
	t.OTLP.SetDefaults()
}

// Providers contains providers configuration.
type Providers struct {
	ProvidersThrottleDuration ptypes.Duration `description:"Backends throttle duration: minimum duration between 2 events from providers before applying a new configuration. It avoids unnecessary reloads if multiples events are sent in a short amount of time." json:"providersThrottleDuration,omitempty" toml:"providersThrottleDuration,omitempty" yaml:"providersThrottleDuration,omitempty" export:"true"`

	Docker *docker.Provider      `description:"Enable Docker backend with default settings." json:"docker,omitempty" toml:"docker,omitempty" yaml:"docker,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	Swarm  *docker.SwarmProvider `description:"Enable Docker Swarm backend with default settings." json:"swarm,omitempty" toml:"swarm,omitempty" yaml:"swarm,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`

	File                   *file.Provider                 `description:"Enable File backend with default settings." json:"file,omitempty" toml:"file,omitempty" yaml:"file,omitempty" export:"true"`
	KubernetesIngress      *ingress.Provider              `description:"Enable Kubernetes backend with default settings." json:"kubernetesIngress,omitempty" toml:"kubernetesIngress,omitempty" yaml:"kubernetesIngress,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	KubernetesIngressNGINX *ingressnginx.Provider         `description:"Enable Kubernetes Ingress NGINX provider." json:"kubernetesIngressNGINX,omitempty" toml:"kubernetesIngressNGINX,omitempty" yaml:"kubernetesIngressNGINX,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	KubernetesCRD          *crd.Provider                  `description:"Enable Kubernetes backend with default settings." json:"kubernetesCRD,omitempty" toml:"kubernetesCRD,omitempty" yaml:"kubernetesCRD,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	KubernetesGateway      *gateway.Provider              `description:"Enable Kubernetes gateway api provider with default settings." json:"kubernetesGateway,omitempty" toml:"kubernetesGateway,omitempty" yaml:"kubernetesGateway,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	Rest                   *rest.Provider                 `description:"Enable Rest backend with default settings." json:"rest,omitempty" toml:"rest,omitempty" yaml:"rest,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	ConsulCatalog          *consulcatalog.ProviderBuilder `description:"Enable ConsulCatalog backend with default settings." json:"consulCatalog,omitempty" toml:"consulCatalog,omitempty" yaml:"consulCatalog,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	Nomad                  *nomad.ProviderBuilder         `description:"Enable Nomad backend with default settings." json:"nomad,omitempty" toml:"nomad,omitempty" yaml:"nomad,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	Ecs                    *ecs.Provider                  `description:"Enable AWS ECS backend with default settings." json:"ecs,omitempty" toml:"ecs,omitempty" yaml:"ecs,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	Consul                 *consul.ProviderBuilder        `description:"Enable Consul backend with default settings." json:"consul,omitempty" toml:"consul,omitempty" yaml:"consul,omitempty" label:"allowEmpty" file:"allowEmpty"  export:"true"`
	Etcd                   *etcd.Provider                 `description:"Enable Etcd backend with default settings." json:"etcd,omitempty" toml:"etcd,omitempty" yaml:"etcd,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	ZooKeeper              *zk.Provider                   `description:"Enable ZooKeeper backend with default settings." json:"zooKeeper,omitempty" toml:"zooKeeper,omitempty" yaml:"zooKeeper,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	Redis                  *redis.Provider                `description:"Enable Redis backend with default settings." json:"redis,omitempty" toml:"redis,omitempty" yaml:"redis,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	HTTP                   *http.Provider                 `description:"Enable HTTP backend with default settings." json:"http,omitempty" toml:"http,omitempty" yaml:"http,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`

	Plugin map[string]PluginConf `description:"Plugins configuration." json:"plugin,omitempty" toml:"plugin,omitempty" yaml:"plugin,omitempty"`
}

// SetEffectiveConfiguration adds missing configuration parameters derived from existing ones.
// It also takes care of maintaining backwards compatibility.
func (c *Configuration) SetEffectiveConfiguration() {
	// Creates the default entry point if needed
	if !c.hasUserDefinedEntrypoint() {
		ep := &EntryPoint{Address: ":80"}
		ep.SetDefaults()
		// TODO: double check this tomorrow
		if c.EntryPoints == nil {
			c.EntryPoints = make(EntryPoints)
		}
		c.EntryPoints["http"] = ep
	}

	// Creates the internal traefik entry point if needed
	if (c.API != nil && c.API.Insecure) ||
		(c.Ping != nil && !c.Ping.ManualRouting && c.Ping.EntryPoint == DefaultInternalEntryPointName) ||
		(c.Metrics != nil && c.Metrics.Prometheus != nil && !c.Metrics.Prometheus.ManualRouting && c.Metrics.Prometheus.EntryPoint == DefaultInternalEntryPointName) ||
		(c.Providers != nil && c.Providers.Rest != nil && c.Providers.Rest.Insecure) {
		if _, ok := c.EntryPoints[DefaultInternalEntryPointName]; !ok {
			ep := &EntryPoint{Address: ":8080"}
			ep.SetDefaults()
			c.EntryPoints[DefaultInternalEntryPointName] = ep
		}
	}

	if c.Tracing != nil && c.Tracing.GlobalAttributes != nil && c.Tracing.ResourceAttributes == nil {
		c.Tracing.ResourceAttributes = c.Tracing.GlobalAttributes
	}

	if c.Providers.Docker != nil {
		if c.Providers.Docker.HTTPClientTimeout < 0 {
			c.Providers.Docker.HTTPClientTimeout = 0
		}
	}

	if c.Providers.Swarm != nil {
		if c.Providers.Swarm.RefreshSeconds <= 0 {
			c.Providers.Swarm.RefreshSeconds = ptypes.Duration(15 * time.Second)
		}

		if c.Providers.Swarm.HTTPClientTimeout < 0 {
			c.Providers.Swarm.HTTPClientTimeout = 0
		}
	}

	// Configure Gateway API provider
	if c.Providers.KubernetesGateway != nil {
		entryPoints := make(map[string]gateway.Entrypoint)
		for epName, entryPoint := range c.EntryPoints {
			entryPoints[epName] = gateway.Entrypoint{Address: entryPoint.GetAddress(), HasHTTPTLSConf: entryPoint.HTTP.TLS != nil}
		}

		if c.Providers.KubernetesCRD != nil {
			c.Providers.KubernetesCRD.FillExtensionBuilderRegistry(c.Providers.KubernetesGateway)
		}

		c.Providers.KubernetesGateway.EntryPoints = entryPoints
	}

	// Defines the default rule syntax for the Kubernetes Ingress Provider.
	// This allows the provider to adapt the matcher syntax to the desired rule syntax version.
	if c.Core != nil && c.Providers.KubernetesIngress != nil {
		c.Providers.KubernetesIngress.DefaultRuleSyntax = c.Core.DefaultRuleSyntax
	}

	for _, resolver := range c.CertificatesResolvers {
		if resolver.ACME == nil {
			continue
		}

		if resolver.ACME.DNSChallenge == nil {
			continue
		}

		switch resolver.ACME.DNSChallenge.Provider {
		case "googledomains", "cloudxns", "brandit":
			log.Warn().Msgf("%s DNS provider is deprecated.", resolver.ACME.DNSChallenge.Provider)
		case "dnspod":
			log.Warn().Msgf("%s provider is deprecated, please use 'tencentcloud' provider instead.", resolver.ACME.DNSChallenge.Provider)
		case "azure":
			log.Warn().Msgf("%s provider is deprecated, please use 'azuredns' provider instead.", resolver.ACME.DNSChallenge.Provider)
		}

		if resolver.ACME.DNSChallenge.DisablePropagationCheck {
			log.Warn().Msgf("disablePropagationCheck is now deprecated, please use propagation.disableChecks instead.")

			if resolver.ACME.DNSChallenge.Propagation == nil {
				resolver.ACME.DNSChallenge.Propagation = &acmeprovider.Propagation{}
			}

			resolver.ACME.DNSChallenge.Propagation.DisableChecks = true
		}

		if resolver.ACME.DNSChallenge.DelayBeforeCheck > 0 {
			log.Warn().Msgf("delayBeforeCheck is now deprecated, please use propagation.delayBeforeChecks instead.")

			if resolver.ACME.DNSChallenge.Propagation == nil {
				resolver.ACME.DNSChallenge.Propagation = &acmeprovider.Propagation{}
			}

			resolver.ACME.DNSChallenge.Propagation.DelayBeforeChecks = resolver.ACME.DNSChallenge.DelayBeforeCheck
		}
	}

	c.initACMEProvider()
}

func (c *Configuration) hasUserDefinedEntrypoint() bool {
	return len(c.EntryPoints) != 0
}

func (c *Configuration) initACMEProvider() {
	for _, resolver := range c.CertificatesResolvers {
		if resolver.ACME != nil {
			resolver.ACME.CAServer = getSafeACMECAServer(resolver.ACME.CAServer)
		}
	}

	logger := logs.NoLevel(log.Logger, zerolog.DebugLevel).With().Str("lib", "lego").Logger()
	legolog.Logger = logs.NewLogrusWrapper(logger)
}

// ValidateConfiguration validate that configuration is coherent.
func (c *Configuration) ValidateConfiguration() error {
	for name, resolver := range c.CertificatesResolvers {
		if resolver.ACME != nil && resolver.Tailscale != nil {
			return fmt.Errorf("unable to initialize certificates resolver %q, as ACME and Tailscale providers are mutually exclusive", name)
		}

		if resolver.ACME == nil {
			continue
		}

		if len(resolver.ACME.Storage) == 0 {
			return fmt.Errorf("unable to initialize certificates resolver %q with no storage location for the certificates", name)
		}
	}

	if c.Core != nil {
		switch c.Core.DefaultRuleSyntax {
		case "v3": // NOOP
		case "v2":
			// TODO: point to migration guide.
			log.Warn().Msgf("v2 rules syntax is now deprecated, please use v3 instead...")
		default:
			return fmt.Errorf("unsupported default rule syntax configuration: %q", c.Core.DefaultRuleSyntax)
		}
	}

	if c.Providers != nil && c.Providers.KubernetesIngressNGINX != nil {
		if c.Experimental == nil || !c.Experimental.KubernetesIngressNGINX {
			return errors.New("the experimental KubernetesIngressNGINX feature must be enabled to use the KubernetesIngressNGINX provider")
		}

		if c.Providers.KubernetesIngressNGINX.WatchNamespace != "" && c.Providers.KubernetesIngressNGINX.WatchNamespaceSelector != "" {
			return errors.New("watchNamespace and watchNamespaceSelector options are mutually exclusive")
		}
	}

	if c.AccessLog != nil && c.AccessLog.OTLP != nil {
		if c.Experimental == nil || !c.Experimental.OTLPLogs {
			return errors.New("the experimental OTLPLogs feature must be enabled to use OTLP access logging")
		}

		if c.AccessLog.OTLP.GRPC != nil && c.AccessLog.OTLP.GRPC.TLS != nil && c.AccessLog.OTLP.GRPC.Insecure {
			return errors.New("access logs OTLP GRPC: TLS and Insecure options are mutually exclusive")
		}
	}

	if c.Log != nil && c.Log.OTLP != nil {
		if c.Experimental == nil || !c.Experimental.OTLPLogs {
			return errors.New("the experimental OTLPLogs feature must be enabled to use OTLP logging")
		}

		if c.Log.OTLP.GRPC != nil && c.Log.OTLP.GRPC.TLS != nil && c.Log.OTLP.GRPC.Insecure {
			return errors.New("logs OTLP GRPC: TLS and Insecure options are mutually exclusive")
		}
	}

	if c.Tracing != nil && c.Tracing.OTLP != nil {
		if c.Tracing.OTLP.GRPC != nil && c.Tracing.OTLP.GRPC.TLS != nil && c.Tracing.OTLP.GRPC.Insecure {
			return errors.New("tracing OTLP GRPC: TLS and Insecure options are mutually exclusive")
		}
	}

	if c.Metrics != nil && c.Metrics.OTLP != nil {
		if c.Metrics.OTLP.GRPC != nil && c.Metrics.OTLP.GRPC.TLS != nil && c.Metrics.OTLP.GRPC.Insecure {
			return errors.New("metrics OTLP GRPC: TLS and Insecure options are mutually exclusive")
		}
	}

	if c.API != nil && !path.IsAbs(c.API.BasePath) {
		return errors.New("API basePath must be a valid absolute path")
	}

	if c.OCSP != nil {
		for responderURL, url := range c.OCSP.ResponderOverrides {
			if url == "" {
				return fmt.Errorf("OCSP responder override value for %s cannot be empty", responderURL)
			}
		}
	}

	return nil
}

func getSafeACMECAServer(caServerSrc string) string {
	if len(caServerSrc) == 0 {
		return DefaultAcmeCAServer
	}

	if strings.HasPrefix(caServerSrc, "https://acme-v01.api.letsencrypt.org") {
		caServer := strings.Replace(caServerSrc, "v01", "v02", 1)
		log.Warn().Msgf("The CA server %[1]q refers to a v01 endpoint of the ACME API, please change to %[2]q. Fallback to %[2]q.", caServerSrc, caServer)
		return caServer
	}

	if strings.HasPrefix(caServerSrc, "https://acme-staging.api.letsencrypt.org") {
		caServer := strings.Replace(caServerSrc, "https://acme-staging.api.letsencrypt.org", "https://acme-staging-v02.api.letsencrypt.org", 1)
		log.Warn().Msgf("The CA server %[1]q refers to a v01 endpoint of the ACME API, please change to %[2]q. Fallback to %[2]q.", caServerSrc, caServer)
		return caServer
	}

	return caServerSrc
}
