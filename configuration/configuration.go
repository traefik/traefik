package configuration

import (
	"fmt"
	"strings"
	"time"

	"github.com/containous/flaeg"
	"github.com/containous/traefik-extra-service-fabric"
	"github.com/containous/traefik/acme"
	"github.com/containous/traefik/api"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/middlewares/tracing"
	"github.com/containous/traefik/middlewares/tracing/datadog"
	"github.com/containous/traefik/middlewares/tracing/jaeger"
	"github.com/containous/traefik/middlewares/tracing/zipkin"
	"github.com/containous/traefik/ping"
	acmeprovider "github.com/containous/traefik/provider/acme"
	"github.com/containous/traefik/provider/boltdb"
	"github.com/containous/traefik/provider/consul"
	"github.com/containous/traefik/provider/consulcatalog"
	"github.com/containous/traefik/provider/docker"
	"github.com/containous/traefik/provider/dynamodb"
	"github.com/containous/traefik/provider/ecs"
	"github.com/containous/traefik/provider/etcd"
	"github.com/containous/traefik/provider/eureka"
	"github.com/containous/traefik/provider/file"
	"github.com/containous/traefik/provider/kubernetes"
	"github.com/containous/traefik/provider/marathon"
	"github.com/containous/traefik/provider/mesos"
	"github.com/containous/traefik/provider/rancher"
	"github.com/containous/traefik/provider/rest"
	"github.com/containous/traefik/provider/zk"
	"github.com/containous/traefik/tls"
	"github.com/containous/traefik/types"
	"github.com/pkg/errors"
	"github.com/xenolf/lego/challenge/dns01"
)

const (
	// DefaultInternalEntryPointName the name of the default internal entry point
	DefaultInternalEntryPointName = "traefik"

	// DefaultHealthCheckInterval is the default health check interval.
	DefaultHealthCheckInterval = 30 * time.Second

	// DefaultDialTimeout when connecting to a backend server.
	DefaultDialTimeout = 30 * time.Second

	// DefaultIdleTimeout before closing an idle connection.
	DefaultIdleTimeout = 180 * time.Second

	// DefaultGraceTimeout controls how long Traefik serves pending requests
	// prior to shutting down.
	DefaultGraceTimeout = 10 * time.Second

	// DefaultAcmeCAServer is the default ACME API endpoint
	DefaultAcmeCAServer = "https://acme-v02.api.letsencrypt.org/directory"
)

// GlobalConfiguration holds global configuration (with providers, etc.).
// It's populated from the traefik configuration file passed as an argument to the binary.
type GlobalConfiguration struct {
	LifeCycle                 *LifeCycle        `description:"Timeouts influencing the server life cycle" export:"true"`
	GraceTimeOut              flaeg.Duration    `short:"g" description:"(Deprecated) Duration to give active requests a chance to finish before Traefik stops" export:"true"` // Deprecated
	Debug                     bool              `short:"d" description:"Enable debug mode" export:"true"`
	CheckNewVersion           bool              `description:"Periodically check if a new version has been released" export:"true"`
	SendAnonymousUsage        bool              `description:"send periodically anonymous usage statistics" export:"true"`
	AccessLogsFile            string            `description:"(Deprecated) Access logs file" export:"true"` // Deprecated
	AccessLog                 *types.AccessLog  `description:"Access log settings" export:"true"`
	TraefikLogsFile           string            `description:"(Deprecated) Traefik logs file. Stdout is used when omitted or empty" export:"true"` // Deprecated
	TraefikLog                *types.TraefikLog `description:"Traefik log settings" export:"true"`
	Tracing                   *tracing.Tracing  `description:"OpenTracing configuration" export:"true"`
	LogLevel                  string            `short:"l" description:"Log level" export:"true"`
	EntryPoints               EntryPoints       `description:"Entrypoints definition using format: --entryPoints='Name:http Address::8000 Redirect.EntryPoint:https' --entryPoints='Name:https Address::4442 TLS:tests/traefik.crt,tests/traefik.key;prod/traefik.crt,prod/traefik.key'" export:"true"`
	Cluster                   *types.Cluster
	Constraints               types.Constraints       `description:"Filter services by constraint, matching with service tags" export:"true"`
	ACME                      *acme.ACME              `description:"Enable ACME (Let's Encrypt): automatic SSL" export:"true"`
	DefaultEntryPoints        DefaultEntryPoints      `description:"Entrypoints to be used by frontends that do not specify any entrypoint" export:"true"`
	ProvidersThrottleDuration flaeg.Duration          `description:"Backends throttle duration: minimum duration between 2 events from providers before applying a new configuration. It avoids unnecessary reloads if multiples events are sent in a short amount of time." export:"true"`
	MaxIdleConnsPerHost       int                     `description:"If non-zero, controls the maximum idle (keep-alive) to keep per-host.  If zero, DefaultMaxIdleConnsPerHost is used" export:"true"`
	IdleTimeout               flaeg.Duration          `description:"(Deprecated) maximum amount of time an idle (keep-alive) connection will remain idle before closing itself." export:"true"` // Deprecated
	InsecureSkipVerify        bool                    `description:"Disable SSL certificate verification" export:"true"`
	RootCAs                   tls.FilesOrContents     `description:"Add cert file for self-signed certificate"`
	Retry                     *Retry                  `description:"Enable retry sending request if network error" export:"true"`
	HealthCheck               *HealthCheckConfig      `description:"Health check parameters" export:"true"`
	RespondingTimeouts        *RespondingTimeouts     `description:"Timeouts for incoming requests to the Traefik instance" export:"true"`
	ForwardingTimeouts        *ForwardingTimeouts     `description:"Timeouts for requests forwarded to the backend servers" export:"true"`
	AllowMinWeightZero        bool                    `description:"Allow weight to take 0 as minimum real value." export:"true"`         // Deprecated
	KeepTrailingSlash         bool                    `description:"Do not remove trailing slash." export:"true"`                         // Deprecated
	Web                       *WebCompatibility       `description:"(Deprecated) Enable Web backend with default settings" export:"true"` // Deprecated
	Docker                    *docker.Provider        `description:"Enable Docker backend with default settings" export:"true"`
	File                      *file.Provider          `description:"Enable File backend with default settings" export:"true"`
	Marathon                  *marathon.Provider      `description:"Enable Marathon backend with default settings" export:"true"`
	Consul                    *consul.Provider        `description:"Enable Consul backend with default settings" export:"true"`
	ConsulCatalog             *consulcatalog.Provider `description:"Enable Consul catalog backend with default settings" export:"true"`
	Etcd                      *etcd.Provider          `description:"Enable Etcd backend with default settings" export:"true"`
	Zookeeper                 *zk.Provider            `description:"Enable Zookeeper backend with default settings" export:"true"`
	Boltdb                    *boltdb.Provider        `description:"Enable Boltdb backend with default settings" export:"true"`
	Kubernetes                *kubernetes.Provider    `description:"Enable Kubernetes backend with default settings" export:"true"`
	Mesos                     *mesos.Provider         `description:"Enable Mesos backend with default settings" export:"true"`
	Eureka                    *eureka.Provider        `description:"Enable Eureka backend with default settings" export:"true"`
	ECS                       *ecs.Provider           `description:"Enable ECS backend with default settings" export:"true"`
	Rancher                   *rancher.Provider       `description:"Enable Rancher backend with default settings" export:"true"`
	DynamoDB                  *dynamodb.Provider      `description:"Enable DynamoDB backend with default settings" export:"true"`
	ServiceFabric             *servicefabric.Provider `description:"Enable Service Fabric backend with default settings" export:"true"`
	Rest                      *rest.Provider          `description:"Enable Rest backend with default settings" export:"true"`
	API                       *api.Handler            `description:"Enable api/dashboard" export:"true"`
	Metrics                   *types.Metrics          `description:"Enable a metrics exporter" export:"true"`
	Ping                      *ping.Handler           `description:"Enable ping" export:"true"`
	HostResolver              *HostResolverConfig     `description:"Enable CNAME Flattening" export:"true"`
}

// WebCompatibility is a configuration to handle compatibility with deprecated web provider options
type WebCompatibility struct {
	Address    string            `description:"(Deprecated) Web administration port" export:"true"`
	CertFile   string            `description:"(Deprecated) SSL certificate" export:"true"`
	KeyFile    string            `description:"(Deprecated) SSL certificate" export:"true"`
	ReadOnly   bool              `description:"(Deprecated) Enable read only API" export:"true"`
	Statistics *types.Statistics `description:"(Deprecated) Enable more detailed statistics" export:"true"`
	Metrics    *types.Metrics    `description:"(Deprecated) Enable a metrics exporter" export:"true"`
	Path       string            `description:"(Deprecated) Root path for dashboard and API" export:"true"`
	Auth       *types.Auth       `export:"true"`
	Debug      bool              `export:"true"`
}

func (gc *GlobalConfiguration) handleWebDeprecation() {
	if gc.Web != nil {
		log.Warn("web provider configuration is deprecated, you should use these options : api, rest provider, ping and metrics")

		if gc.API != nil || gc.Metrics != nil || gc.Ping != nil || gc.Rest != nil {
			log.Warn("web option is ignored if you use it with one of these options : api, rest provider, ping or metrics")
			return
		}
		gc.EntryPoints[DefaultInternalEntryPointName] = &EntryPoint{
			Address: gc.Web.Address,
			Auth:    gc.Web.Auth,
		}
		if gc.Web.CertFile != "" {
			gc.EntryPoints[DefaultInternalEntryPointName].TLS = &tls.TLS{
				Certificates: []tls.Certificate{
					{
						CertFile: tls.FileOrContent(gc.Web.CertFile),
						KeyFile:  tls.FileOrContent(gc.Web.KeyFile),
					},
				},
			}
		}

		if gc.API == nil {
			gc.API = &api.Handler{
				EntryPoint: DefaultInternalEntryPointName,
				Statistics: gc.Web.Statistics,
				Dashboard:  true,
			}
		}

		if gc.Ping == nil {
			gc.Ping = &ping.Handler{
				EntryPoint: DefaultInternalEntryPointName,
			}
		}

		if gc.Metrics == nil {
			gc.Metrics = gc.Web.Metrics
		}

		if !gc.Debug {
			gc.Debug = gc.Web.Debug
		}
	}
}

// SetEffectiveConfiguration adds missing configuration parameters derived from existing ones.
// It also takes care of maintaining backwards compatibility.
func (gc *GlobalConfiguration) SetEffectiveConfiguration(configFile string) {
	if len(gc.EntryPoints) == 0 {
		gc.EntryPoints = map[string]*EntryPoint{"http": {
			Address:          ":80",
			ForwardedHeaders: &ForwardedHeaders{Insecure: true},
		}}
		gc.DefaultEntryPoints = []string{"http"}
	}

	gc.handleWebDeprecation()

	if (gc.API != nil && gc.API.EntryPoint == DefaultInternalEntryPointName) ||
		(gc.Ping != nil && gc.Ping.EntryPoint == DefaultInternalEntryPointName) ||
		(gc.Metrics != nil && gc.Metrics.Prometheus != nil && gc.Metrics.Prometheus.EntryPoint == DefaultInternalEntryPointName) ||
		(gc.Rest != nil && gc.Rest.EntryPoint == DefaultInternalEntryPointName) {
		if _, ok := gc.EntryPoints[DefaultInternalEntryPointName]; !ok {
			gc.EntryPoints[DefaultInternalEntryPointName] = &EntryPoint{Address: ":8080"}
		}
	}

	for entryPointName := range gc.EntryPoints {
		entryPoint := gc.EntryPoints[entryPointName]
		// ForwardedHeaders must be remove in the next breaking version
		if entryPoint.ForwardedHeaders == nil {
			entryPoint.ForwardedHeaders = &ForwardedHeaders{Insecure: true}
		}

		if len(entryPoint.WhitelistSourceRange) > 0 {
			log.Warnf("Deprecated configuration found: %s. Please use %s.", "whiteListSourceRange", "whiteList.sourceRange")

			if entryPoint.WhiteList == nil {
				entryPoint.WhiteList = &types.WhiteList{
					SourceRange: entryPoint.WhitelistSourceRange,
				}
				entryPoint.WhitelistSourceRange = nil
			}
		}

		if entryPoint.TLS != nil && entryPoint.TLS.DefaultCertificate == nil && len(entryPoint.TLS.Certificates) > 0 {
			log.Infof("No tls.defaultCertificate given for %s: using the first item in tls.certificates as a fallback.", entryPointName)
			entryPoint.TLS.DefaultCertificate = &entryPoint.TLS.Certificates[0]
		}
	}

	// Make sure LifeCycle isn't nil to spare nil checks elsewhere.
	if gc.LifeCycle == nil {
		gc.LifeCycle = &LifeCycle{}
	}

	// Prefer legacy grace timeout parameter for backwards compatibility reasons.
	if gc.GraceTimeOut > 0 {
		log.Warn("top-level grace period configuration has been deprecated -- please use lifecycle grace period")
		gc.LifeCycle.GraceTimeOut = gc.GraceTimeOut
	}

	if gc.Docker != nil {
		if len(gc.Docker.Filename) != 0 && gc.Docker.TemplateVersion != 2 {
			log.Warn("Template version 1 is deprecated, please use version 2, see TemplateVersion.")
			gc.Docker.TemplateVersion = 1
		} else {
			gc.Docker.TemplateVersion = 2
		}

		if gc.Docker.SwarmModeRefreshSeconds <= 0 {
			gc.Docker.SwarmModeRefreshSeconds = 15
		}
	}

	if gc.Marathon != nil {
		if len(gc.Marathon.Filename) != 0 && gc.Marathon.TemplateVersion != 2 {
			log.Warn("Template version 1 is deprecated, please use version 2, see TemplateVersion.")
			gc.Marathon.TemplateVersion = 1
		} else {
			gc.Marathon.TemplateVersion = 2
		}
	}

	if gc.Mesos != nil {
		if len(gc.Mesos.Filename) != 0 && gc.Mesos.TemplateVersion != 2 {
			log.Warn("Template version 1 is deprecated, please use version 2, see TemplateVersion.")
			gc.Mesos.TemplateVersion = 1
		} else {
			gc.Mesos.TemplateVersion = 2
		}
	}

	if gc.Eureka != nil {
		if gc.Eureka.Delay != 0 {
			log.Warn("Delay has been deprecated -- please use RefreshSeconds")
			gc.Eureka.RefreshSeconds = gc.Eureka.Delay
		}
	}

	if gc.ECS != nil {
		if len(gc.ECS.Filename) != 0 && gc.ECS.TemplateVersion != 2 {
			log.Warn("Template version 1 is deprecated, please use version 2, see TemplateVersion.")
			gc.ECS.TemplateVersion = 1
		} else {
			gc.ECS.TemplateVersion = 2
		}
	}

	if gc.ConsulCatalog != nil {
		if len(gc.ConsulCatalog.Filename) != 0 && gc.ConsulCatalog.TemplateVersion != 2 {
			log.Warn("Template version 1 is deprecated, please use version 2, see TemplateVersion.")
			gc.ConsulCatalog.TemplateVersion = 1
		} else {
			gc.ConsulCatalog.TemplateVersion = 2
		}
	}

	if gc.Rancher != nil {
		if len(gc.Rancher.Filename) != 0 && gc.Rancher.TemplateVersion != 2 {
			log.Warn("Template version 1 is deprecated, please use version 2, see TemplateVersion.")
			gc.Rancher.TemplateVersion = 1
		} else {
			gc.Rancher.TemplateVersion = 2
		}

		// Ensure backwards compatibility for now
		if len(gc.Rancher.AccessKey) > 0 ||
			len(gc.Rancher.Endpoint) > 0 ||
			len(gc.Rancher.SecretKey) > 0 {

			if gc.Rancher.API == nil {
				gc.Rancher.API = &rancher.APIConfiguration{
					AccessKey: gc.Rancher.AccessKey,
					SecretKey: gc.Rancher.SecretKey,
					Endpoint:  gc.Rancher.Endpoint,
				}
			}
			log.Warn("Deprecated configuration found: rancher.[accesskey|secretkey|endpoint]. " +
				"Please use rancher.api.[accesskey|secretkey|endpoint] instead.")
		}

		if gc.Rancher.Metadata != nil && len(gc.Rancher.Metadata.Prefix) == 0 {
			gc.Rancher.Metadata.Prefix = "latest"
		}
	}

	if gc.API != nil {
		gc.API.Debug = gc.Debug
	}

	if gc.Web != nil && (gc.Web.Path == "" || !strings.HasSuffix(gc.Web.Path, "/")) {
		gc.Web.Path += "/"
	}

	if gc.File != nil {
		gc.File.TraefikFile = configFile
	}

	gc.initACMEProvider()
	gc.initTracing()
}

func (gc *GlobalConfiguration) initTracing() {
	if gc.Tracing != nil {
		switch gc.Tracing.Backend {
		case jaeger.Name:
			if gc.Tracing.Jaeger == nil {
				gc.Tracing.Jaeger = &jaeger.Config{
					SamplingServerURL:  "http://localhost:5778/sampling",
					SamplingType:       "const",
					SamplingParam:      1.0,
					LocalAgentHostPort: "127.0.0.1:6831",
					TraceContextHeaderName: "uber-trace-id",
				}
			}
			if gc.Tracing.Zipkin != nil {
				log.Warn("Zipkin configuration will be ignored")
				gc.Tracing.Zipkin = nil
			}
			if gc.Tracing.DataDog != nil {
				log.Warn("DataDog configuration will be ignored")
				gc.Tracing.DataDog = nil
			}
		case zipkin.Name:
			if gc.Tracing.Zipkin == nil {
				gc.Tracing.Zipkin = &zipkin.Config{
					HTTPEndpoint: "http://localhost:9411/api/v1/spans",
					SameSpan:     false,
					ID128Bit:     true,
					Debug:        false,
				}
			}
			if gc.Tracing.Jaeger != nil {
				log.Warn("Jaeger configuration will be ignored")
				gc.Tracing.Jaeger = nil
			}
			if gc.Tracing.DataDog != nil {
				log.Warn("DataDog configuration will be ignored")
				gc.Tracing.DataDog = nil
			}
		case datadog.Name:
			if gc.Tracing.DataDog == nil {
				gc.Tracing.DataDog = &datadog.Config{
					LocalAgentHostPort: "localhost:8126",
					GlobalTag:          "",
					Debug:              false,
					PrioritySampling:   false,
				}
			}
			if gc.Tracing.Zipkin != nil {
				log.Warn("Zipkin configuration will be ignored")
				gc.Tracing.Zipkin = nil
			}
			if gc.Tracing.Jaeger != nil {
				log.Warn("Jaeger configuration will be ignored")
				gc.Tracing.Jaeger = nil
			}
		default:
			log.Warnf("Unknown tracer %q", gc.Tracing.Backend)
			return
		}
	}
}

func (gc *GlobalConfiguration) initACMEProvider() {
	if gc.ACME != nil {
		gc.ACME.CAServer = getSafeACMECAServer(gc.ACME.CAServer)

		if gc.ACME.DNSChallenge != nil && gc.ACME.HTTPChallenge != nil {
			log.Warn("Unable to use DNS challenge and HTTP challenge at the same time. Fallback to DNS challenge.")
			gc.ACME.HTTPChallenge = nil
		}

		if gc.ACME.DNSChallenge != nil && gc.ACME.TLSChallenge != nil {
			log.Warn("Unable to use DNS challenge and TLS challenge at the same time. Fallback to DNS challenge.")
			gc.ACME.TLSChallenge = nil
		}

		if gc.ACME.HTTPChallenge != nil && gc.ACME.TLSChallenge != nil {
			log.Warn("Unable to use HTTP challenge and TLS challenge at the same time. Fallback to TLS challenge.")
			gc.ACME.HTTPChallenge = nil
		}

		for _, domain := range gc.ACME.Domains {
			if domain.Main != dns01.UnFqdn(domain.Main) {
				log.Warnf("FQDN detected, please remove the trailing dot: %s", domain.Main)
			}
			for _, san := range domain.SANs {
				if san != dns01.UnFqdn(san) {
					log.Warnf("FQDN detected, please remove the trailing dot: %s", san)
				}
			}
		}

		// TODO: to remove in the future
		if len(gc.ACME.StorageFile) > 0 && len(gc.ACME.Storage) == 0 {
			log.Warn("ACME.StorageFile is deprecated, use ACME.Storage instead")
			gc.ACME.Storage = gc.ACME.StorageFile
		}

		if len(gc.ACME.DNSProvider) > 0 {
			log.Warn("ACME.DNSProvider is deprecated, use ACME.DNSChallenge instead")
			gc.ACME.DNSChallenge = &acmeprovider.DNSChallenge{Provider: gc.ACME.DNSProvider, DelayBeforeCheck: gc.ACME.DelayDontCheckDNS}
		}

		if gc.ACME.OnDemand {
			log.Warn("ACME.OnDemand is deprecated")
		}
	}
}

// InitACMEProvider create an acme provider from the ACME part of globalConfiguration
func (gc *GlobalConfiguration) InitACMEProvider() (*acmeprovider.Provider, error) {
	if gc.ACME != nil {
		if len(gc.ACME.Storage) == 0 {
			// Delete the ACME configuration to avoid starting ACME in cluster mode
			gc.ACME = nil
			return nil, errors.New("unable to initialize ACME provider with no storage location for the certificates")
		}
		// TODO: Remove when Provider ACME will replace totally ACME
		// If provider file, use Provider ACME instead of ACME
		if gc.Cluster == nil {
			provider := &acmeprovider.Provider{}
			provider.Configuration = &acmeprovider.Configuration{
				KeyType:       gc.ACME.KeyType,
				OnHostRule:    gc.ACME.OnHostRule,
				OnDemand:      gc.ACME.OnDemand,
				Email:         gc.ACME.Email,
				Storage:       gc.ACME.Storage,
				HTTPChallenge: gc.ACME.HTTPChallenge,
				DNSChallenge:  gc.ACME.DNSChallenge,
				TLSChallenge:  gc.ACME.TLSChallenge,
				Domains:       gc.ACME.Domains,
				ACMELogging:   gc.ACME.ACMELogging,
				CAServer:      gc.ACME.CAServer,
				EntryPoint:    gc.ACME.EntryPoint,
			}

			store := acmeprovider.NewLocalStore(provider.Storage)
			provider.Store = store
			acme.ConvertToNewFormat(provider.Storage)
			gc.ACME = nil
			return provider, nil
		}
	}
	return nil, nil
}

func getSafeACMECAServer(caServerSrc string) string {
	if len(caServerSrc) == 0 {
		return DefaultAcmeCAServer
	}

	if strings.HasPrefix(caServerSrc, "https://acme-v01.api.letsencrypt.org") {
		caServer := strings.Replace(caServerSrc, "v01", "v02", 1)
		log.Warnf("The CA server %[1]q refers to a v01 endpoint of the ACME API, please change to %[2]q. Fallback to %[2]q.", caServerSrc, caServer)
		return caServer
	}

	if strings.HasPrefix(caServerSrc, "https://acme-staging.api.letsencrypt.org") {
		caServer := strings.Replace(caServerSrc, "https://acme-staging.api.letsencrypt.org", "https://acme-staging-v02.api.letsencrypt.org", 1)
		log.Warnf("The CA server %[1]q refers to a v01 endpoint of the ACME API, please change to %[2]q. Fallback to %[2]q.", caServerSrc, caServer)
		return caServer
	}

	return caServerSrc
}

// ValidateConfiguration validate that configuration is coherent
func (gc *GlobalConfiguration) ValidateConfiguration() {
	if gc.ACME != nil {
		if _, ok := gc.EntryPoints[gc.ACME.EntryPoint]; !ok {
			log.Fatalf("Unknown entrypoint %q for ACME configuration", gc.ACME.EntryPoint)
		} else {
			if gc.EntryPoints[gc.ACME.EntryPoint].TLS == nil {
				log.Fatalf("Entrypoint %q has no TLS configuration for ACME configuration", gc.ACME.EntryPoint)
			}
		}
	}
}

// DefaultEntryPoints holds default entry points
type DefaultEntryPoints []string

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (dep *DefaultEntryPoints) String() string {
	return strings.Join(*dep, ",")
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (dep *DefaultEntryPoints) Set(value string) error {
	entrypoints := strings.Split(value, ",")
	if len(entrypoints) == 0 {
		return fmt.Errorf("bad DefaultEntryPoints format: %s", value)
	}
	for _, entrypoint := range entrypoints {
		*dep = append(*dep, entrypoint)
	}
	return nil
}

// Get return the EntryPoints map
func (dep *DefaultEntryPoints) Get() interface{} {
	return *dep
}

// SetValue sets the EntryPoints map with val
func (dep *DefaultEntryPoints) SetValue(val interface{}) {
	*dep = val.(DefaultEntryPoints)
}

// Type is type of the struct
func (dep *DefaultEntryPoints) Type() string {
	return "defaultentrypoints"
}

// Retry contains request retry config
type Retry struct {
	Attempts int `description:"Number of attempts" export:"true"`
}

// HealthCheckConfig contains health check configuration parameters.
type HealthCheckConfig struct {
	Interval flaeg.Duration `description:"Default periodicity of enabled health checks" export:"true"`
}

// RespondingTimeouts contains timeout configurations for incoming requests to the Traefik instance.
type RespondingTimeouts struct {
	ReadTimeout  flaeg.Duration `description:"ReadTimeout is the maximum duration for reading the entire request, including the body. If zero, no timeout is set" export:"true"`
	WriteTimeout flaeg.Duration `description:"WriteTimeout is the maximum duration before timing out writes of the response. If zero, no timeout is set" export:"true"`
	IdleTimeout  flaeg.Duration `description:"IdleTimeout is the maximum amount duration an idle (keep-alive) connection will remain idle before closing itself. Defaults to 180 seconds. If zero, no timeout is set" export:"true"`
}

// ForwardingTimeouts contains timeout configurations for forwarding requests to the backend servers.
type ForwardingTimeouts struct {
	DialTimeout           flaeg.Duration `description:"The amount of time to wait until a connection to a backend server can be established. Defaults to 30 seconds. If zero, no timeout exists" export:"true"`
	ResponseHeaderTimeout flaeg.Duration `description:"The amount of time to wait for a server's response headers after fully writing the request (including its body, if any). If zero, no timeout exists" export:"true"`
}

// LifeCycle contains configurations relevant to the lifecycle (such as the
// shutdown phase) of Traefik.
type LifeCycle struct {
	RequestAcceptGraceTimeout flaeg.Duration `description:"Duration to keep accepting requests before Traefik initiates the graceful shutdown procedure"`
	GraceTimeOut              flaeg.Duration `description:"Duration to give active requests a chance to finish before Traefik stops"`
}

// HostResolverConfig contain configuration for CNAME Flattening
type HostResolverConfig struct {
	CnameFlattening bool   `description:"A flag to enable/disable CNAME flattening" export:"true"`
	ResolvConfig    string `description:"resolv.conf used for DNS resolving" export:"true"`
	ResolvDepth     int    `description:"The maximal depth of DNS recursive resolving" export:"true"`
}
