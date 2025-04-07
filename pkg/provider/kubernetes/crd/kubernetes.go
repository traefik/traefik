package crd

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
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
	traefikv1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/gateway"
	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/k8s"
	"github.com/traefik/traefik/v3/pkg/safe"
	"github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
	corev1 "k8s.io/api/core/v1"
	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	annotationKubernetesIngressClass = "kubernetes.io/ingress.class"
	traefikDefaultIngressClass       = "traefik"
)

const (
	providerName               = "kubernetescrd"
	providerNamespaceSeparator = "@"
)

// Provider holds configurations of the provider.
type Provider struct {
	Endpoint                     string              `description:"Kubernetes server endpoint (required for external cluster client)." json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Token                        types.FileOrContent `description:"Kubernetes bearer token (not needed for in-cluster client). It accepts either a token value or a file path to the token." json:"token,omitempty" toml:"token,omitempty" yaml:"token,omitempty" loggable:"false"`
	CertAuthFilePath             string              `description:"Kubernetes certificate authority file path (not needed for in-cluster client)." json:"certAuthFilePath,omitempty" toml:"certAuthFilePath,omitempty" yaml:"certAuthFilePath,omitempty"`
	Namespaces                   []string            `description:"Kubernetes namespaces." json:"namespaces,omitempty" toml:"namespaces,omitempty" yaml:"namespaces,omitempty" export:"true"`
	AllowCrossNamespace          bool                `description:"Allow cross namespace resource reference." json:"allowCrossNamespace,omitempty" toml:"allowCrossNamespace,omitempty" yaml:"allowCrossNamespace,omitempty" export:"true"`
	AllowExternalNameServices    bool                `description:"Allow ExternalName services." json:"allowExternalNameServices,omitempty" toml:"allowExternalNameServices,omitempty" yaml:"allowExternalNameServices,omitempty" export:"true"`
	LabelSelector                string              `description:"Kubernetes label selector to use." json:"labelSelector,omitempty" toml:"labelSelector,omitempty" yaml:"labelSelector,omitempty" export:"true"`
	IngressClass                 string              `description:"Value of kubernetes.io/ingress.class annotation to watch for." json:"ingressClass,omitempty" toml:"ingressClass,omitempty" yaml:"ingressClass,omitempty" export:"true"`
	ThrottleDuration             ptypes.Duration     `description:"Ingress refresh throttle duration" json:"throttleDuration,omitempty" toml:"throttleDuration,omitempty" yaml:"throttleDuration,omitempty" export:"true"`
	AllowEmptyServices           bool                `description:"Allow the creation of services without endpoints." json:"allowEmptyServices,omitempty" toml:"allowEmptyServices,omitempty" yaml:"allowEmptyServices,omitempty" export:"true"`
	NativeLBByDefault            bool                `description:"Defines whether to use Native Kubernetes load-balancing mode by default." json:"nativeLBByDefault,omitempty" toml:"nativeLBByDefault,omitempty" yaml:"nativeLBByDefault,omitempty" export:"true"`
	DisableClusterScopeResources bool                `description:"Disables the lookup of cluster scope resources (incompatible with IngressClasses and NodePortLB enabled services)." json:"disableClusterScopeResources,omitempty" toml:"disableClusterScopeResources,omitempty" yaml:"disableClusterScopeResources,omitempty" export:"true"`

	lastConfiguration safe.Safe

	routerTransform k8s.RouterTransform
}

func (p *Provider) SetRouterTransform(routerTransform k8s.RouterTransform) {
	p.routerTransform = routerTransform
}

func (p *Provider) applyRouterTransform(ctx context.Context, rt *dynamic.Router, ingress *traefikv1alpha1.IngressRoute) {
	if p.routerTransform == nil {
		return
	}

	err := p.routerTransform.Apply(ctx, rt, ingress)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Apply router transform")
	}
}

func (p *Provider) newK8sClient(ctx context.Context) (*clientWrapper, error) {
	_, err := labels.Parse(p.LabelSelector)
	if err != nil {
		return nil, fmt.Errorf("invalid label selector: %q", p.LabelSelector)
	}
	log.Ctx(ctx).Info().Msgf("label selector is: %q", p.LabelSelector)

	withEndpoint := ""
	if p.Endpoint != "" {
		withEndpoint = fmt.Sprintf(" with endpoint %s", p.Endpoint)
	}

	var client *clientWrapper
	switch {
	case os.Getenv("KUBERNETES_SERVICE_HOST") != "" && os.Getenv("KUBERNETES_SERVICE_PORT") != "":
		log.Ctx(ctx).Info().Msgf("Creating in-cluster Provider client%s", withEndpoint)
		client, err = newInClusterClient(p.Endpoint)
	case os.Getenv("KUBECONFIG") != "":
		log.Ctx(ctx).Info().Msgf("Creating cluster-external Provider client from KUBECONFIG %s", os.Getenv("KUBECONFIG"))
		client, err = newExternalClusterClientFromFile(os.Getenv("KUBECONFIG"))
	default:
		log.Ctx(ctx).Info().Msgf("Creating cluster-external Provider client%s", withEndpoint)
		client, err = newExternalClusterClient(p.Endpoint, p.CertAuthFilePath, p.Token)
	}

	if err != nil {
		return nil, err
	}

	client.labelSelector = p.LabelSelector
	client.disableClusterScopeInformer = p.DisableClusterScopeResources
	return client, nil
}

// Init the provider.
func (p *Provider) Init() error {
	return nil
}

// Provide allows the k8s provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	logger := log.With().Str(logs.ProviderName, providerName).Logger()
	ctxLog := logger.WithContext(context.Background())

	k8sClient, err := p.newK8sClient(ctxLog)
	if err != nil {
		return err
	}

	if p.AllowCrossNamespace {
		logger.Warn().Msg("Cross-namespace reference between IngressRoutes and resources is enabled, please ensure that this is expected (see AllowCrossNamespace option)")
	}

	if p.AllowExternalNameServices {
		logger.Info().Msg("ExternalName service loading is enabled, please ensure that this is expected (see AllowExternalNameServices option)")
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
					// Note that event is the *first* event that came in during this throttling interval -- if we're hitting our throttle, we may have dropped events.
					// This is fine, because we don't treat different event types differently.
					// But if we do in the future, we'll need to track more information about the dropped events.
					conf := p.loadConfigurationFromCRD(ctxLog, k8sClient)

					confHash, err := hashstructure.Hash(conf, nil)
					switch {
					case err != nil:
						logger.Error().Err(err).Msg("Unable to hash the configuration")
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

func (p *Provider) loadConfigurationFromCRD(ctx context.Context, client Client) *dynamic.Configuration {
	stores, tlsConfigs := buildTLSStores(ctx, client)
	if tlsConfigs == nil {
		tlsConfigs = make(map[string]*tls.CertAndStores)
	}

	conf := &dynamic.Configuration{
		// TODO: choose between mutating and returning tlsConfigs
		HTTP: p.loadIngressRouteConfiguration(ctx, client, tlsConfigs),
		TCP:  p.loadIngressRouteTCPConfiguration(ctx, client, tlsConfigs),
		UDP:  p.loadIngressRouteUDPConfiguration(ctx, client),
		TLS: &dynamic.TLSConfiguration{
			Options: buildTLSOptions(ctx, client),
			Stores:  stores,
		},
	}

	// Done after because tlsConfigs is mutated by the others above.
	conf.TLS.Certificates = getTLSConfig(tlsConfigs)

	for _, middleware := range client.GetMiddlewares() {
		id := provider.Normalize(makeID(middleware.Namespace, middleware.Name))
		logger := log.Ctx(ctx).With().Str(logs.MiddlewareName, id).Logger()
		ctxMid := logger.WithContext(ctx)

		basicAuth, err := createBasicAuthMiddleware(client, middleware.Namespace, middleware.Spec.BasicAuth)
		if err != nil {
			logger.Error().Err(err).Msg("Error while reading basic auth middleware")
			continue
		}

		digestAuth, err := createDigestAuthMiddleware(client, middleware.Namespace, middleware.Spec.DigestAuth)
		if err != nil {
			logger.Error().Err(err).Msg("Error while reading digest auth middleware")
			continue
		}

		forwardAuth, err := createForwardAuthMiddleware(client, middleware.Namespace, middleware.Spec.ForwardAuth)
		if err != nil {
			logger.Error().Err(err).Msg("Error while reading forward auth middleware")
			continue
		}

		errorPage, errorPageService, err := p.createErrorPageMiddleware(client, middleware.Namespace, middleware.Spec.Errors)
		if err != nil {
			logger.Error().Err(err).Msg("Error while reading error page middleware")
			continue
		}

		if errorPage != nil && errorPageService != nil {
			serviceName := id + "-errorpage-service"
			errorPage.Service = serviceName
			conf.HTTP.Services[serviceName] = errorPageService
		}

		plugin, err := createPluginMiddleware(client, middleware.Namespace, middleware.Spec.Plugin)
		if err != nil {
			logger.Error().Err(err).Msg("Error while reading plugins middleware")
			continue
		}

		rateLimit, err := createRateLimitMiddleware(client, middleware.Namespace, middleware.Spec.RateLimit)
		if err != nil {
			logger.Error().Err(err).Msg("Error while reading rateLimit middleware")
			continue
		}

		retry, err := createRetryMiddleware(middleware.Spec.Retry)
		if err != nil {
			logger.Error().Err(err).Msg("Error while reading retry middleware")
			continue
		}

		circuitBreaker, err := createCircuitBreakerMiddleware(middleware.Spec.CircuitBreaker)
		if err != nil {
			logger.Error().Err(err).Msg("Error while reading circuit breaker middleware")
			continue
		}

		conf.HTTP.Middlewares[id] = &dynamic.Middleware{
			AddPrefix:         middleware.Spec.AddPrefix,
			StripPrefix:       middleware.Spec.StripPrefix,
			StripPrefixRegex:  middleware.Spec.StripPrefixRegex,
			ReplacePath:       middleware.Spec.ReplacePath,
			ReplacePathRegex:  middleware.Spec.ReplacePathRegex,
			Chain:             createChainMiddleware(ctxMid, middleware.Namespace, middleware.Spec.Chain),
			IPWhiteList:       middleware.Spec.IPWhiteList,
			IPAllowList:       middleware.Spec.IPAllowList,
			Headers:           middleware.Spec.Headers,
			Errors:            errorPage,
			RateLimit:         rateLimit,
			RedirectRegex:     middleware.Spec.RedirectRegex,
			RedirectScheme:    middleware.Spec.RedirectScheme,
			BasicAuth:         basicAuth,
			DigestAuth:        digestAuth,
			ForwardAuth:       forwardAuth,
			InFlightReq:       middleware.Spec.InFlightReq,
			Buffering:         middleware.Spec.Buffering,
			CircuitBreaker:    circuitBreaker,
			Compress:          createCompressMiddleware(middleware.Spec.Compress),
			PassTLSClientCert: middleware.Spec.PassTLSClientCert,
			Retry:             retry,
			ContentType:       middleware.Spec.ContentType,
			GrpcWeb:           middleware.Spec.GrpcWeb,
			Plugin:            plugin,
		}
	}

	for _, middlewareTCP := range client.GetMiddlewareTCPs() {
		id := provider.Normalize(makeID(middlewareTCP.Namespace, middlewareTCP.Name))

		conf.TCP.Middlewares[id] = &dynamic.TCPMiddleware{
			InFlightConn: middlewareTCP.Spec.InFlightConn,
			IPWhiteList:  middlewareTCP.Spec.IPWhiteList,
			IPAllowList:  middlewareTCP.Spec.IPAllowList,
		}
	}

	cb := configBuilder{
		client:                    client,
		allowCrossNamespace:       p.AllowCrossNamespace,
		allowExternalNameServices: p.AllowExternalNameServices,
		allowEmptyServices:        p.AllowEmptyServices,
	}

	for _, service := range client.GetTraefikServices() {
		err := cb.buildTraefikService(ctx, service, conf.HTTP.Services)
		if err != nil {
			log.Ctx(ctx).Error().Str(logs.ServiceName, service.Name).Err(err).
				Msg("Error while building TraefikService")
			continue
		}
	}

	for _, serversTransport := range client.GetServersTransports() {
		logger := log.Ctx(ctx).With().
			Str(logs.ServersTransportName, serversTransport.Name).
			Str("namespace", serversTransport.Namespace).
			Logger()

		if len(serversTransport.Spec.RootCAsSecrets) > 0 {
			logger.Warn().Msg("RootCAsSecrets option is deprecated, please use the RootCA option instead.")
		}

		var rootCAs []types.FileOrContent
		for _, secret := range serversTransport.Spec.RootCAsSecrets {
			caSecret, err := loadCASecret(serversTransport.Namespace, secret, client)
			if err != nil {
				logger.Error().
					Err(err).
					Str("secret", secret).
					Msg("Error while loading CA Secret")
				continue
			}

			rootCAs = append(rootCAs, types.FileOrContent(caSecret))
		}

		for _, rootCA := range serversTransport.Spec.RootCAs {
			if rootCA.Secret != nil && rootCA.ConfigMap != nil {
				logger.Error().Msg("Error while loading CA: both Secret and ConfigMap are defined")
				continue
			}

			if rootCA.Secret != nil {
				ca, err := loadCASecret(serversTransport.Namespace, *rootCA.Secret, client)
				if err != nil {
					logger.Error().
						Err(err).
						Str("secret", *rootCA.Secret).
						Msg("Error while loading CA Secret")
					continue
				}

				rootCAs = append(rootCAs, types.FileOrContent(ca))
				continue
			}

			ca, err := loadCAConfigMap(serversTransport.Namespace, *rootCA.ConfigMap, client)
			if err != nil {
				logger.Error().
					Err(err).
					Str("configMap", *rootCA.ConfigMap).
					Msg("Error while loading CA ConfigMap")
				continue
			}

			rootCAs = append(rootCAs, types.FileOrContent(ca))
		}

		var certs tls.Certificates
		for _, secret := range serversTransport.Spec.CertificatesSecrets {
			tlsSecret, tlsKey, err := loadAuthTLSSecret(serversTransport.Namespace, secret, client)
			if err != nil {
				logger.Error().Err(err).Msgf("Error while loading certificates %s", secret)
				continue
			}

			certs = append(certs, tls.Certificate{
				CertFile: types.FileOrContent(tlsSecret),
				KeyFile:  types.FileOrContent(tlsKey),
			})
		}

		forwardingTimeout := &dynamic.ForwardingTimeouts{}
		forwardingTimeout.SetDefaults()

		if serversTransport.Spec.ForwardingTimeouts != nil {
			if serversTransport.Spec.ForwardingTimeouts.DialTimeout != nil {
				err := forwardingTimeout.DialTimeout.Set(serversTransport.Spec.ForwardingTimeouts.DialTimeout.String())
				if err != nil {
					logger.Error().Err(err).Msg("Error while reading DialTimeout")
				}
			}

			if serversTransport.Spec.ForwardingTimeouts.ResponseHeaderTimeout != nil {
				err := forwardingTimeout.ResponseHeaderTimeout.Set(serversTransport.Spec.ForwardingTimeouts.ResponseHeaderTimeout.String())
				if err != nil {
					logger.Error().Err(err).Msg("Error while reading ResponseHeaderTimeout")
				}
			}

			if serversTransport.Spec.ForwardingTimeouts.IdleConnTimeout != nil {
				err := forwardingTimeout.IdleConnTimeout.Set(serversTransport.Spec.ForwardingTimeouts.IdleConnTimeout.String())
				if err != nil {
					logger.Error().Err(err).Msg("Error while reading IdleConnTimeout")
				}
			}

			if serversTransport.Spec.ForwardingTimeouts.ReadIdleTimeout != nil {
				err := forwardingTimeout.ReadIdleTimeout.Set(serversTransport.Spec.ForwardingTimeouts.ReadIdleTimeout.String())
				if err != nil {
					logger.Error().Err(err).Msg("Error while reading ReadIdleTimeout")
				}
			}

			if serversTransport.Spec.ForwardingTimeouts.PingTimeout != nil {
				err := forwardingTimeout.PingTimeout.Set(serversTransport.Spec.ForwardingTimeouts.PingTimeout.String())
				if err != nil {
					logger.Error().Err(err).Msg("Error while reading PingTimeout")
				}
			}
		}

		id := provider.Normalize(makeID(serversTransport.Namespace, serversTransport.Name))
		conf.HTTP.ServersTransports[id] = &dynamic.ServersTransport{
			ServerName:          serversTransport.Spec.ServerName,
			InsecureSkipVerify:  serversTransport.Spec.InsecureSkipVerify,
			RootCAs:             rootCAs,
			Certificates:        certs,
			DisableHTTP2:        serversTransport.Spec.DisableHTTP2,
			MaxIdleConnsPerHost: serversTransport.Spec.MaxIdleConnsPerHost,
			ForwardingTimeouts:  forwardingTimeout,
			PeerCertURI:         serversTransport.Spec.PeerCertURI,
			Spiffe:              serversTransport.Spec.Spiffe,
		}
	}

	for _, serversTransportTCP := range client.GetServersTransportTCPs() {
		logger := log.Ctx(ctx).With().Str(logs.ServersTransportName, serversTransportTCP.Name).Logger()

		var tcpServerTransport dynamic.TCPServersTransport
		tcpServerTransport.SetDefaults()

		if serversTransportTCP.Spec.DialTimeout != nil {
			err := tcpServerTransport.DialTimeout.Set(serversTransportTCP.Spec.DialTimeout.String())
			if err != nil {
				logger.Error().Err(err).Msg("Error while reading DialTimeout")
			}
		}

		if serversTransportTCP.Spec.DialKeepAlive != nil {
			err := tcpServerTransport.DialKeepAlive.Set(serversTransportTCP.Spec.DialKeepAlive.String())
			if err != nil {
				logger.Error().Err(err).Msg("Error while reading DialKeepAlive")
			}
		}

		if serversTransportTCP.Spec.TerminationDelay != nil {
			err := tcpServerTransport.TerminationDelay.Set(serversTransportTCP.Spec.TerminationDelay.String())
			if err != nil {
				logger.Error().Err(err).Msg("Error while reading TerminationDelay")
			}
		}

		if serversTransportTCP.Spec.TLS != nil {
			if len(serversTransportTCP.Spec.TLS.RootCAsSecrets) > 0 {
				logger.Warn().Msg("RootCAsSecrets option is deprecated, please use the RootCA option instead.")
			}

			var rootCAs []types.FileOrContent
			for _, secret := range serversTransportTCP.Spec.TLS.RootCAsSecrets {
				caSecret, err := loadCASecret(serversTransportTCP.Namespace, secret, client)
				if err != nil {
					logger.Error().
						Err(err).
						Str("secret", secret).
						Msg("Error while loading CA Secret")
					continue
				}

				rootCAs = append(rootCAs, types.FileOrContent(caSecret))
			}

			for _, rootCA := range serversTransportTCP.Spec.TLS.RootCAs {
				if rootCA.Secret != nil && rootCA.ConfigMap != nil {
					logger.Error().Msg("Error while loading CA: both Secret and ConfigMap are defined")
					continue
				}

				if rootCA.Secret != nil {
					ca, err := loadCASecret(serversTransportTCP.Namespace, *rootCA.Secret, client)
					if err != nil {
						logger.Error().
							Err(err).
							Str("secret", *rootCA.Secret).
							Msg("Error while loading CA Secret")
						continue
					}

					rootCAs = append(rootCAs, types.FileOrContent(ca))
					continue
				}

				ca, err := loadCAConfigMap(serversTransportTCP.Namespace, *rootCA.ConfigMap, client)
				if err != nil {
					logger.Error().
						Err(err).
						Str("configMap", *rootCA.ConfigMap).
						Msg("Error while loading CA ConfigMap")
					continue
				}

				rootCAs = append(rootCAs, types.FileOrContent(ca))
			}

			var certs tls.Certificates
			for _, secret := range serversTransportTCP.Spec.TLS.CertificatesSecrets {
				tlsCert, tlsKey, err := loadAuthTLSSecret(serversTransportTCP.Namespace, secret, client)
				if err != nil {
					logger.Error().
						Err(err).
						Str("certificates", secret).
						Msg("Error while loading certificates")
					continue
				}

				certs = append(certs, tls.Certificate{
					CertFile: types.FileOrContent(tlsCert),
					KeyFile:  types.FileOrContent(tlsKey),
				})
			}

			tcpServerTransport.TLS = &dynamic.TLSClientConfig{
				ServerName:         serversTransportTCP.Spec.TLS.ServerName,
				InsecureSkipVerify: serversTransportTCP.Spec.TLS.InsecureSkipVerify,
				RootCAs:            rootCAs,
				Certificates:       certs,
				PeerCertURI:        serversTransportTCP.Spec.TLS.PeerCertURI,
			}

			tcpServerTransport.TLS.Spiffe = serversTransportTCP.Spec.TLS.Spiffe
		}

		id := provider.Normalize(makeID(serversTransportTCP.Namespace, serversTransportTCP.Name))
		conf.TCP.ServersTransports[id] = &tcpServerTransport
	}

	return conf
}

// getServicePort always returns a valid port, an error otherwise.
func getServicePort(svc *corev1.Service, port intstr.IntOrString) (*corev1.ServicePort, error) {
	if svc == nil {
		return nil, errors.New("service is not defined")
	}

	if (port.Type == intstr.Int && port.IntVal == 0) || (port.Type == intstr.String && port.StrVal == "") {
		return nil, errors.New("ingressRoute service port not defined")
	}

	hasValidPort := false
	for _, p := range svc.Spec.Ports {
		if (port.Type == intstr.Int && port.IntVal == p.Port) || (port.Type == intstr.String && port.StrVal == p.Name) {
			return &p, nil
		}

		if p.Port != 0 {
			hasValidPort = true
		}
	}

	if svc.Spec.Type != corev1.ServiceTypeExternalName || port.Type == intstr.String {
		return nil, fmt.Errorf("service port not found: %s", &port)
	}

	if hasValidPort {
		log.Warn().Msgf("The port %s from IngressRoute doesn't match with ports defined in the ExternalName service %s/%s.",
			&port, svc.Namespace, svc.Name)
	}

	return &corev1.ServicePort{Port: port.IntVal}, nil
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

func createPluginMiddleware(k8sClient Client, ns string, plugins map[string]apiextensionv1.JSON) (map[string]dynamic.PluginConf, error) {
	if plugins == nil {
		return nil, nil
	}

	data, err := json.Marshal(plugins)
	if err != nil {
		return nil, err
	}

	pcMap := map[string]dynamic.PluginConf{}
	if err = json.Unmarshal(data, &pcMap); err != nil {
		return nil, err
	}

	for _, pc := range pcMap {
		for key := range pc {
			if pc[key], err = loadSecretKeys(k8sClient, ns, pc[key]); err != nil {
				return nil, err
			}
		}
	}

	return pcMap, nil
}

func loadSecretKeys(k8sClient Client, ns string, i interface{}) (interface{}, error) {
	var err error
	switch iv := i.(type) {
	case string:
		if !strings.HasPrefix(iv, "urn:k8s:secret:") {
			return iv, nil
		}

		return getSecretValue(k8sClient, ns, iv)

	case []interface{}:
		for i := range iv {
			if iv[i], err = loadSecretKeys(k8sClient, ns, iv[i]); err != nil {
				return nil, err
			}
		}

	case map[string]interface{}:
		for k := range iv {
			if iv[k], err = loadSecretKeys(k8sClient, ns, iv[k]); err != nil {
				return nil, err
			}
		}
	}

	return i, nil
}

func getSecretValue(c Client, ns, urn string) (string, error) {
	parts := strings.Split(urn, ":")
	if len(parts) != 5 {
		return "", fmt.Errorf("malformed secret URN %q", urn)
	}

	secretName := parts[3]
	secret, ok, err := c.GetSecret(ns, secretName)
	if err != nil {
		return "", err
	}

	if !ok {
		return "", fmt.Errorf("secret %s/%s is not found", ns, secretName)
	}

	secretKey := parts[4]
	secretValue, ok := secret.Data[secretKey]
	if !ok {
		return "", fmt.Errorf("key %q not found in secret %s/%s", secretKey, ns, secretName)
	}

	return string(secretValue), nil
}

func createCircuitBreakerMiddleware(circuitBreaker *traefikv1alpha1.CircuitBreaker) (*dynamic.CircuitBreaker, error) {
	if circuitBreaker == nil {
		return nil, nil
	}

	cb := &dynamic.CircuitBreaker{Expression: circuitBreaker.Expression}
	cb.SetDefaults()

	if circuitBreaker.CheckPeriod != nil {
		if err := cb.CheckPeriod.Set(circuitBreaker.CheckPeriod.String()); err != nil {
			return nil, err
		}
	}

	if circuitBreaker.FallbackDuration != nil {
		if err := cb.FallbackDuration.Set(circuitBreaker.FallbackDuration.String()); err != nil {
			return nil, err
		}
	}

	if circuitBreaker.RecoveryDuration != nil {
		if err := cb.RecoveryDuration.Set(circuitBreaker.RecoveryDuration.String()); err != nil {
			return nil, err
		}
	}

	if circuitBreaker.ResponseCode != 0 {
		cb.ResponseCode = circuitBreaker.ResponseCode
	}

	return cb, nil
}

func createCompressMiddleware(compress *traefikv1alpha1.Compress) *dynamic.Compress {
	if compress == nil {
		return nil
	}

	c := &dynamic.Compress{}
	c.SetDefaults()

	if compress.ExcludedContentTypes != nil {
		c.ExcludedContentTypes = compress.ExcludedContentTypes
	}

	if compress.IncludedContentTypes != nil {
		c.IncludedContentTypes = compress.IncludedContentTypes
	}

	if compress.MinResponseBodyBytes != nil {
		c.MinResponseBodyBytes = *compress.MinResponseBodyBytes
	}

	if compress.Encodings != nil {
		c.Encodings = compress.Encodings
	}

	if compress.DefaultEncoding != nil {
		c.DefaultEncoding = *compress.DefaultEncoding
	}

	return c
}

func createRateLimitMiddleware(client Client, namespace string, rateLimit *traefikv1alpha1.RateLimit) (*dynamic.RateLimit, error) {
	if rateLimit == nil {
		return nil, nil
	}

	rl := &dynamic.RateLimit{}
	rl.SetDefaults()

	if rateLimit.Average != nil {
		rl.Average = *rateLimit.Average
	}

	if rateLimit.Burst != nil {
		rl.Burst = *rateLimit.Burst
	}

	if rateLimit.Period != nil {
		err := rl.Period.Set(rateLimit.Period.String())
		if err != nil {
			return nil, err
		}
	}

	if rateLimit.SourceCriterion != nil {
		rl.SourceCriterion = rateLimit.SourceCriterion
	}

	if rateLimit.Redis != nil {
		rl.Redis = &dynamic.Redis{
			DB:             rateLimit.Redis.DB,
			PoolSize:       rateLimit.Redis.PoolSize,
			MinIdleConns:   rateLimit.Redis.MinIdleConns,
			MaxActiveConns: rateLimit.Redis.MaxActiveConns,
		}
		rl.Redis.SetDefaults()

		if len(rateLimit.Redis.Endpoints) > 0 {
			rl.Redis.Endpoints = rateLimit.Redis.Endpoints
		}

		if rateLimit.Redis.TLS != nil {
			rl.Redis.TLS = &types.ClientTLS{
				InsecureSkipVerify: rateLimit.Redis.TLS.InsecureSkipVerify,
			}

			if len(rateLimit.Redis.TLS.CASecret) > 0 {
				caSecret, err := loadCASecret(namespace, rateLimit.Redis.TLS.CASecret, client)
				if err != nil {
					return nil, fmt.Errorf("failed to load auth ca secret: %w", err)
				}
				rl.Redis.TLS.CA = caSecret
			}

			if len(rateLimit.Redis.TLS.CertSecret) > 0 {
				authSecretCert, authSecretKey, err := loadAuthTLSSecret(namespace, rateLimit.Redis.TLS.CertSecret, client)
				if err != nil {
					return nil, fmt.Errorf("failed to load auth secret: %w", err)
				}
				rl.Redis.TLS.Cert = authSecretCert
				rl.Redis.TLS.Key = authSecretKey
			}
		}

		if rateLimit.Redis.DialTimeout != nil {
			err := rl.Redis.DialTimeout.Set(rateLimit.Redis.DialTimeout.String())
			if err != nil {
				return nil, err
			}
		}

		if rateLimit.Redis.ReadTimeout != nil {
			err := rl.Redis.ReadTimeout.Set(rateLimit.Redis.ReadTimeout.String())
			if err != nil {
				return nil, err
			}
		}

		if rateLimit.Redis.WriteTimeout != nil {
			err := rl.Redis.WriteTimeout.Set(rateLimit.Redis.WriteTimeout.String())
			if err != nil {
				return nil, err
			}
		}

		if rateLimit.Redis.Secret != "" {
			var err error
			rl.Redis.Username, rl.Redis.Password, err = loadRedisCredentials(namespace, rateLimit.Redis.Secret, client)
			if err != nil {
				return nil, err
			}
		}
	}

	return rl, nil
}

func loadRedisCredentials(namespace, secretName string, k8sClient Client) (string, string, error) {
	secret, exists, err := k8sClient.GetSecret(namespace, secretName)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch secret '%s/%s': %w", namespace, secretName, err)
	}

	if !exists {
		return "", "", fmt.Errorf("secret '%s/%s' not found", namespace, secretName)
	}

	if secret == nil {
		return "", "", fmt.Errorf("data for secret '%s/%s' must not be nil", namespace, secretName)
	}

	username, usernameExists := secret.Data["username"]
	password, passwordExists := secret.Data["password"]
	if !usernameExists || !passwordExists {
		return "", "", fmt.Errorf("secret '%s/%s' must contain both username and password keys", secret.Namespace, secret.Name)
	}
	return string(username), string(password), nil
}

func createRetryMiddleware(retry *traefikv1alpha1.Retry) (*dynamic.Retry, error) {
	if retry == nil {
		return nil, nil
	}

	r := &dynamic.Retry{Attempts: retry.Attempts}

	err := r.InitialInterval.Set(retry.InitialInterval.String())
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (p *Provider) createErrorPageMiddleware(client Client, namespace string, errorPage *traefikv1alpha1.ErrorPage) (*dynamic.ErrorPage, *dynamic.Service, error) {
	if errorPage == nil {
		return nil, nil, nil
	}

	errorPageMiddleware := &dynamic.ErrorPage{
		Status:         errorPage.Status,
		StatusRewrites: errorPage.StatusRewrites,
		Query:          errorPage.Query,
	}

	cb := configBuilder{
		client:                    client,
		allowCrossNamespace:       p.AllowCrossNamespace,
		allowExternalNameServices: p.AllowExternalNameServices,
		allowEmptyServices:        p.AllowEmptyServices,
	}

	balancerServerHTTP, err := cb.buildServersLB(namespace, errorPage.Service.LoadBalancerSpec)
	if err != nil {
		return nil, nil, err
	}

	return errorPageMiddleware, balancerServerHTTP, nil
}

func (p *Provider) FillExtensionBuilderRegistry(registry gateway.ExtensionBuilderRegistry) {
	registry.RegisterFilterFuncs(traefikv1alpha1.GroupName, "Middleware", func(name, namespace string) (string, *dynamic.Middleware, error) {
		if len(p.Namespaces) > 0 && !slices.Contains(p.Namespaces, namespace) {
			return "", nil, fmt.Errorf("namespace %q is not allowed", namespace)
		}

		return makeID(namespace, name) + providerNamespaceSeparator + providerName, nil, nil
	})

	registry.RegisterBackendFuncs(traefikv1alpha1.GroupName, "TraefikService", func(name, namespace string) (string, *dynamic.Service, error) {
		if len(p.Namespaces) > 0 && !slices.Contains(p.Namespaces, namespace) {
			return "", nil, fmt.Errorf("namespace %q is not allowed", namespace)
		}

		return makeID(namespace, name) + providerNamespaceSeparator + providerName, nil, nil
	})
}

func createForwardAuthMiddleware(k8sClient Client, namespace string, auth *traefikv1alpha1.ForwardAuth) (*dynamic.ForwardAuth, error) {
	if auth == nil {
		return nil, nil
	}
	if len(auth.Address) == 0 {
		return nil, errors.New("forward authentication requires an address")
	}

	forwardAuth := &dynamic.ForwardAuth{
		Address:                  auth.Address,
		TrustForwardHeader:       auth.TrustForwardHeader,
		AuthResponseHeaders:      auth.AuthResponseHeaders,
		AuthResponseHeadersRegex: auth.AuthResponseHeadersRegex,
		AuthRequestHeaders:       auth.AuthRequestHeaders,
		AddAuthCookiesToResponse: auth.AddAuthCookiesToResponse,
		HeaderField:              auth.HeaderField,
		ForwardBody:              auth.ForwardBody,
		PreserveLocationHeader:   auth.PreserveLocationHeader,
		PreserveRequestMethod:    auth.PreserveRequestMethod,
	}
	forwardAuth.SetDefaults()

	if auth.MaxBodySize != nil {
		forwardAuth.MaxBodySize = auth.MaxBodySize
	}

	if auth.TLS != nil {
		forwardAuth.TLS = &dynamic.ClientTLS{
			InsecureSkipVerify: auth.TLS.InsecureSkipVerify,
		}

		if len(auth.TLS.CASecret) > 0 {
			caSecret, err := loadCASecret(namespace, auth.TLS.CASecret, k8sClient)
			if err != nil {
				return nil, fmt.Errorf("failed to load auth ca secret: %w", err)
			}
			forwardAuth.TLS.CA = caSecret
		}

		if len(auth.TLS.CertSecret) > 0 {
			authSecretCert, authSecretKey, err := loadAuthTLSSecret(namespace, auth.TLS.CertSecret, k8sClient)
			if err != nil {
				return nil, fmt.Errorf("failed to load auth secret: %w", err)
			}
			forwardAuth.TLS.Cert = authSecretCert
			forwardAuth.TLS.Key = authSecretKey
		}

		forwardAuth.TLS.CAOptional = auth.TLS.CAOptional
	}

	return forwardAuth, nil
}

func loadCASecret(namespace, secretName string, k8sClient Client) (string, error) {
	secret, ok, err := k8sClient.GetSecret(namespace, secretName)
	if err != nil {
		return "", fmt.Errorf("failed to fetch secret '%s/%s': %w", namespace, secretName, err)
	}

	if !ok {
		return "", fmt.Errorf("secret '%s/%s' not found", namespace, secretName)
	}

	if secret == nil {
		return "", fmt.Errorf("data for secret '%s/%s' must not be nil", namespace, secretName)
	}

	tlsCAData, err := getCABlocks(secret, namespace, secretName)
	if err == nil {
		return tlsCAData, nil
	}

	// TODO: remove this behavior in the next major version (v4)
	if len(secret.Data) == 1 {
		// For backwards compatibility, use the only available secret data as CA if both 'ca.crt' and 'tls.ca' are missing.
		for _, v := range secret.Data {
			return string(v), nil
		}
	}

	return "", fmt.Errorf("secret '%s/%s' has no CA block: %w", namespace, secretName, err)
}

func loadCAConfigMap(namespace, name string, k8sClient Client) (string, error) {
	configMap, ok, err := k8sClient.GetConfigMap(namespace, name)
	if err != nil {
		return "", fmt.Errorf("failed to fetch configMap '%s/%s': %w", namespace, name, err)
	}

	if !ok {
		return "", fmt.Errorf("configMap '%s/%s' not found", namespace, name)
	}

	if configMap == nil {
		return "", fmt.Errorf("data for configMap '%s/%s' must not be nil", namespace, name)
	}

	tlsCAData, err := getCABlocksFromConfigMap(configMap, namespace, name)
	if err == nil {
		return tlsCAData, nil
	}

	return "", fmt.Errorf("configMap '%s/%s' has no CA block: %w", namespace, name, err)
}

func loadAuthTLSSecret(namespace, secretName string, k8sClient Client) (string, string, error) {
	secret, exists, err := k8sClient.GetSecret(namespace, secretName)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch secret '%s/%s': %w", namespace, secretName, err)
	}

	if !exists {
		return "", "", fmt.Errorf("secret '%s/%s' does not exist", namespace, secretName)
	}

	if secret == nil {
		return "", "", fmt.Errorf("data for secret '%s/%s' must not be nil", namespace, secretName)
	}

	return getCertificateBlocks(secret, namespace, secretName)
}

func createBasicAuthMiddleware(client Client, namespace string, basicAuth *traefikv1alpha1.BasicAuth) (*dynamic.BasicAuth, error) {
	if basicAuth == nil {
		return nil, nil
	}

	if basicAuth.Secret == "" {
		return nil, errors.New("auth secret must be set")
	}

	secret, ok, err := client.GetSecret(namespace, basicAuth.Secret)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch secret '%s/%s': %w", namespace, basicAuth.Secret, err)
	}
	if !ok {
		return nil, fmt.Errorf("secret '%s/%s' not found", namespace, basicAuth.Secret)
	}
	if secret == nil {
		return nil, fmt.Errorf("data for secret '%s/%s' must not be nil", namespace, basicAuth.Secret)
	}

	if secret.Type == corev1.SecretTypeBasicAuth {
		credentials, err := loadBasicAuthCredentials(secret)
		if err != nil {
			return nil, fmt.Errorf("failed to load basic auth credentials: %w", err)
		}

		return &dynamic.BasicAuth{
			Users:        credentials,
			Realm:        basicAuth.Realm,
			RemoveHeader: basicAuth.RemoveHeader,
			HeaderField:  basicAuth.HeaderField,
		}, nil
	}

	credentials, err := loadAuthCredentials(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to load basic auth credentials: %w", err)
	}

	return &dynamic.BasicAuth{
		Users:        credentials,
		Realm:        basicAuth.Realm,
		RemoveHeader: basicAuth.RemoveHeader,
		HeaderField:  basicAuth.HeaderField,
	}, nil
}

func createDigestAuthMiddleware(client Client, namespace string, digestAuth *traefikv1alpha1.DigestAuth) (*dynamic.DigestAuth, error) {
	if digestAuth == nil {
		return nil, nil
	}

	if digestAuth.Secret == "" {
		return nil, errors.New("auth secret must be set")
	}

	secret, ok, err := client.GetSecret(namespace, digestAuth.Secret)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch secret '%s/%s': %w", namespace, digestAuth.Secret, err)
	}
	if !ok {
		return nil, fmt.Errorf("secret '%s/%s' not found", namespace, digestAuth.Secret)
	}
	if secret == nil {
		return nil, fmt.Errorf("data for secret '%s/%s' must not be nil", namespace, digestAuth.Secret)
	}

	credentials, err := loadAuthCredentials(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to load digest auth credentials: %w", err)
	}

	return &dynamic.DigestAuth{
		Users:        credentials,
		Realm:        digestAuth.Realm,
		RemoveHeader: digestAuth.RemoveHeader,
		HeaderField:  digestAuth.HeaderField,
	}, nil
}

func loadBasicAuthCredentials(secret *corev1.Secret) ([]string, error) {
	username, usernameExists := secret.Data["username"]
	password, passwordExists := secret.Data["password"]
	if !(usernameExists && passwordExists) {
		return nil, fmt.Errorf("secret '%s/%s' must contain both username and password keys", secret.Namespace, secret.Name)
	}

	hash := sha1.New()
	hash.Write(password)
	passwordSum := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	return []string{fmt.Sprintf("%s:{SHA}%s", username, passwordSum)}, nil
}

func loadAuthCredentials(secret *corev1.Secret) ([]string, error) {
	if len(secret.Data) != 1 {
		return nil, fmt.Errorf("found %d elements for secret '%s/%s', must be single element exactly", len(secret.Data), secret.Namespace, secret.Name)
	}

	var firstSecret []byte
	for _, v := range secret.Data {
		firstSecret = v
		break
	}

	var credentials []string
	scanner := bufio.NewScanner(bytes.NewReader(firstSecret))
	for scanner.Scan() {
		if cred := scanner.Text(); len(cred) > 0 {
			credentials = append(credentials, cred)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading secret for %s/%s: %w", secret.Namespace, secret.Name, err)
	}
	if len(credentials) == 0 {
		return nil, fmt.Errorf("secret '%s/%s' does not contain any credentials", secret.Namespace, secret.Name)
	}

	return credentials, nil
}

func createChainMiddleware(ctx context.Context, namespace string, chain *traefikv1alpha1.Chain) *dynamic.Chain {
	if chain == nil {
		return nil
	}

	var mds []string
	for _, mi := range chain.Middlewares {
		if strings.Contains(mi.Name, providerNamespaceSeparator) {
			if len(mi.Namespace) > 0 {
				log.Ctx(ctx).Warn().Msgf("namespace %q is ignored in cross-provider context", mi.Namespace)
			}
			mds = append(mds, mi.Name)
			continue
		}

		ns := mi.Namespace
		if len(ns) == 0 {
			ns = namespace
		}
		mds = append(mds, makeID(ns, mi.Name))
	}
	return &dynamic.Chain{Middlewares: mds}
}

func buildTLSOptions(ctx context.Context, client Client) map[string]tls.Options {
	tlsOptionsCRDs := client.GetTLSOptions()
	var tlsOptions map[string]tls.Options

	if len(tlsOptionsCRDs) == 0 {
		return tlsOptions
	}
	tlsOptions = make(map[string]tls.Options)
	var nsDefault []string

	for _, tlsOptionsCRD := range tlsOptionsCRDs {
		logger := log.Ctx(ctx).With().Str("tlsOption", tlsOptionsCRD.Name).Str("namespace", tlsOptionsCRD.Namespace).Logger()
		var clientCAs []types.FileOrContent

		for _, secretName := range tlsOptionsCRD.Spec.ClientAuth.SecretNames {
			secret, exists, err := client.GetSecret(tlsOptionsCRD.Namespace, secretName)
			if err != nil {
				logger.Error().Err(err).Msgf("Failed to fetch secret %s/%s", tlsOptionsCRD.Namespace, secretName)
				continue
			}

			if !exists {
				logger.Warn().Msgf("Secret %s/%s does not exist", tlsOptionsCRD.Namespace, secretName)
				continue
			}

			cert, err := getCABlocks(secret, tlsOptionsCRD.Namespace, secretName)
			if err != nil {
				logger.Error().Err(err).Msgf("Failed to extract CA from secret %s/%s", tlsOptionsCRD.Namespace, secretName)
				continue
			}

			clientCAs = append(clientCAs, types.FileOrContent(cert))
		}

		id := makeID(tlsOptionsCRD.Namespace, tlsOptionsCRD.Name)
		// If the name is default, we override the default config.
		if tlsOptionsCRD.Name == tls.DefaultTLSConfigName {
			id = tlsOptionsCRD.Name
			nsDefault = append(nsDefault, tlsOptionsCRD.Namespace)
		}

		tlsOption := tls.Options{}
		tlsOption.SetDefaults()

		tlsOption.MinVersion = tlsOptionsCRD.Spec.MinVersion
		tlsOption.MaxVersion = tlsOptionsCRD.Spec.MaxVersion

		if tlsOptionsCRD.Spec.CipherSuites != nil {
			tlsOption.CipherSuites = tlsOptionsCRD.Spec.CipherSuites
		}

		tlsOption.CurvePreferences = tlsOptionsCRD.Spec.CurvePreferences
		tlsOption.ClientAuth = tls.ClientAuth{
			CAFiles:        clientCAs,
			ClientAuthType: tlsOptionsCRD.Spec.ClientAuth.ClientAuthType,
		}
		tlsOption.SniStrict = tlsOptionsCRD.Spec.SniStrict

		if tlsOptionsCRD.Spec.ALPNProtocols != nil {
			tlsOption.ALPNProtocols = tlsOptionsCRD.Spec.ALPNProtocols
		}

		tlsOption.DisableSessionTickets = tlsOptionsCRD.Spec.DisableSessionTickets

		tlsOptions[id] = tlsOption
	}

	if len(nsDefault) > 1 {
		delete(tlsOptions, tls.DefaultTLSConfigName)
		log.Ctx(ctx).Error().Msgf("Default TLS Options defined in multiple namespaces: %v", nsDefault)
	}

	return tlsOptions
}

func buildTLSStores(ctx context.Context, client Client) (map[string]tls.Store, map[string]*tls.CertAndStores) {
	tlsStoreCRD := client.GetTLSStores()
	if len(tlsStoreCRD) == 0 {
		return nil, nil
	}

	var nsDefault []string
	tlsStores := make(map[string]tls.Store)
	tlsConfigs := make(map[string]*tls.CertAndStores)

	for _, t := range tlsStoreCRD {
		logger := log.Ctx(ctx).With().Str("TLSStore", t.Name).Str("namespace", t.Namespace).Logger()

		id := makeID(t.Namespace, t.Name)

		// If the name is default, we override the default config.
		if t.Name == tls.DefaultTLSStoreName {
			id = t.Name
			nsDefault = append(nsDefault, t.Namespace)
		}

		var tlsStore tls.Store

		if t.Spec.DefaultCertificate != nil {
			secretName := t.Spec.DefaultCertificate.SecretName

			secret, exists, err := client.GetSecret(t.Namespace, secretName)
			if err != nil {
				logger.Error().Err(err).Msgf("Failed to fetch secret %s/%s", t.Namespace, secretName)
				continue
			}
			if !exists {
				logger.Error().Msgf("Secret %s/%s does not exist", t.Namespace, secretName)
				continue
			}

			cert, key, err := getCertificateBlocks(secret, t.Namespace, secretName)
			if err != nil {
				logger.Error().Err(err).Msg("Could not get certificate blocks")
				continue
			}

			tlsStore.DefaultCertificate = &tls.Certificate{
				CertFile: types.FileOrContent(cert),
				KeyFile:  types.FileOrContent(key),
			}
		}

		if t.Spec.DefaultGeneratedCert != nil {
			tlsStore.DefaultGeneratedCert = &tls.GeneratedCert{
				Resolver: t.Spec.DefaultGeneratedCert.Resolver,
				Domain:   t.Spec.DefaultGeneratedCert.Domain,
			}
		}

		if err := buildCertificates(client, id, t.Namespace, t.Spec.Certificates, tlsConfigs); err != nil {
			logger.Error().Err(err).Msg("Failed to load certificates")
			continue
		}

		tlsStores[id] = tlsStore
	}

	if len(nsDefault) > 1 {
		delete(tlsStores, tls.DefaultTLSStoreName)
		log.Ctx(ctx).Error().Msgf("Default TLS Stores defined in multiple namespaces: %v", nsDefault)
	}

	return tlsStores, tlsConfigs
}

// buildCertificates loads TLSStore certificates from secrets and sets them into tlsConfigs.
func buildCertificates(client Client, tlsStore, namespace string, certificates []traefikv1alpha1.Certificate, tlsConfigs map[string]*tls.CertAndStores) error {
	for _, c := range certificates {
		configKey := namespace + "/" + c.SecretName
		if _, tlsExists := tlsConfigs[configKey]; !tlsExists {
			certAndStores, err := getTLS(client, c.SecretName, namespace)
			if err != nil {
				return fmt.Errorf("unable to read secret %s: %w", configKey, err)
			}

			certAndStores.Stores = []string{tlsStore}
			tlsConfigs[configKey] = certAndStores
		}
	}

	return nil
}

func makeServiceKey(rule, ingressName string) (string, error) {
	h := sha256.New()
	if _, err := h.Write([]byte(rule)); err != nil {
		return "", err
	}

	key := fmt.Sprintf("%s-%.10x", ingressName, h.Sum(nil))

	return key, nil
}

func makeID(namespace, name string) string {
	if namespace == "" {
		return name
	}

	return namespace + "-" + name
}

func shouldProcessIngress(ingressClass, ingressClassAnnotation string) bool {
	return ingressClass == ingressClassAnnotation ||
		(len(ingressClass) == 0 && ingressClassAnnotation == traefikDefaultIngressClass)
}

func getTLS(k8sClient Client, secretName, namespace string) (*tls.CertAndStores, error) {
	secret, exists, err := k8sClient.GetSecret(namespace, secretName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch secret %s/%s: %w", namespace, secretName, err)
	}
	if !exists {
		return nil, fmt.Errorf("secret %s/%s does not exist", namespace, secretName)
	}

	cert, key, err := getCertificateBlocks(secret, namespace, secretName)
	if err != nil {
		return nil, err
	}

	return &tls.CertAndStores{
		Certificate: tls.Certificate{
			CertFile: types.FileOrContent(cert),
			KeyFile:  types.FileOrContent(key),
		},
	}, nil
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

func getCABlocks(secret *corev1.Secret, namespace, secretName string) (string, error) {
	tlsCrtData, tlsCrtExists := secret.Data["tls.ca"]
	if tlsCrtExists {
		return string(tlsCrtData), nil
	}

	tlsCrtData, tlsCrtExists = secret.Data["ca.crt"]
	if tlsCrtExists {
		return string(tlsCrtData), nil
	}

	return "", fmt.Errorf("secret %s/%s contains neither tls.ca nor ca.crt", namespace, secretName)
}

func getCABlocksFromConfigMap(configMap *corev1.ConfigMap, namespace, name string) (string, error) {
	tlsCrtData, tlsCrtExists := configMap.Data["tls.ca"]
	if tlsCrtExists {
		return tlsCrtData, nil
	}

	tlsCrtData, tlsCrtExists = configMap.Data["ca.crt"]
	if tlsCrtExists {
		return tlsCrtData, nil
	}

	return "", fmt.Errorf("config map %s/%s contains neither tls.ca nor ca.crt", namespace, name)
}

func throttleEvents(ctx context.Context, throttleDuration time.Duration, pool *safe.Pool, eventsChan <-chan interface{}) chan interface{} {
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
					log.Ctx(ctx).Debug().Msgf("Dropping event kind %T due to throttling", nextEvent)
				}
			}
		}
	})

	return eventsChanBuffered
}

func isNamespaceAllowed(allowCrossNamespace bool, parentNamespace, namespace string) bool {
	// If allowCrossNamespace option is not defined the default behavior is to allow cross namespace references.
	return allowCrossNamespace || parentNamespace == namespace
}
