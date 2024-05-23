package gateway

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/hashstructure"
	"github.com/rs/zerolog/log"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/job"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/provider"
	traefikv1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/k8s"
	"github.com/traefik/traefik/v3/pkg/safe"
	"github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	ktypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	gatev1 "sigs.k8s.io/gateway-api/apis/v1"
	gatev1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

const (
	providerName   = "kubernetesgateway"
	controllerName = "traefik.io/gateway-controller"

	groupCore = "core"

	kindGateway        = "Gateway"
	kindTraefikService = "TraefikService"
	kindHTTPRoute      = "HTTPRoute"
	kindTCPRoute       = "TCPRoute"
	kindTLSRoute       = "TLSRoute"
)

// Provider holds configurations of the provider.
type Provider struct {
	Endpoint            string              `description:"Kubernetes server endpoint (required for external cluster client)." json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Token               types.FileOrContent `description:"Kubernetes bearer token (not needed for in-cluster client). It accepts either a token value or a file path to the token." json:"token,omitempty" toml:"token,omitempty" yaml:"token,omitempty" loggable:"false"`
	CertAuthFilePath    string              `description:"Kubernetes certificate authority file path (not needed for in-cluster client)." json:"certAuthFilePath,omitempty" toml:"certAuthFilePath,omitempty" yaml:"certAuthFilePath,omitempty"`
	Namespaces          []string            `description:"Kubernetes namespaces." json:"namespaces,omitempty" toml:"namespaces,omitempty" yaml:"namespaces,omitempty" export:"true"`
	LabelSelector       string              `description:"Kubernetes label selector to select specific GatewayClasses." json:"labelSelector,omitempty" toml:"labelSelector,omitempty" yaml:"labelSelector,omitempty" export:"true"`
	ThrottleDuration    ptypes.Duration     `description:"Kubernetes refresh throttle duration" json:"throttleDuration,omitempty" toml:"throttleDuration,omitempty" yaml:"throttleDuration,omitempty" export:"true"`
	ExperimentalChannel bool                `description:"Toggles Experimental Channel resources support (TCPRoute, TLSRoute...)." json:"experimentalChannel,omitempty" toml:"experimentalChannel,omitempty" yaml:"experimentalChannel,omitempty" export:"true"`
	StatusAddress       *StatusAddress      `description:"Defines the Kubernetes Gateway status address." json:"statusAddress,omitempty" toml:"statusAddress,omitempty" yaml:"statusAddress,omitempty" export:"true"`

	EntryPoints map[string]Entrypoint `json:"-" toml:"-" yaml:"-" label:"-" file:"-"`

	// groupKindFilterFuncs is the list of allowed Group and Kinds for the Filter ExtensionRef objects.
	groupKindFilterFuncs map[string]map[string]BuildFilterFunc
	// groupKindBackendFuncs is the list of allowed Group and Kinds for the Backend ExtensionRef objects.
	groupKindBackendFuncs map[string]map[string]BuildBackendFunc

	lastConfiguration safe.Safe

	routerTransform k8s.RouterTransform
}

// Entrypoint defines the available entry points.
type Entrypoint struct {
	Address        string
	HasHTTPTLSConf bool
}

// StatusAddress holds the Gateway Status address configuration.
type StatusAddress struct {
	IP       string     `description:"IP used to set Kubernetes Gateway status address." json:"ip,omitempty" toml:"ip,omitempty" yaml:"ip,omitempty"`
	Hostname string     `description:"Hostname used for Kubernetes Gateway status address." json:"hostname,omitempty" toml:"hostname,omitempty" yaml:"hostname,omitempty"`
	Service  ServiceRef `description:"Published Kubernetes Service to copy status addresses from." json:"service,omitempty" toml:"service,omitempty" yaml:"service,omitempty"`
}

// ServiceRef holds a Kubernetes service reference.
type ServiceRef struct {
	Name      string `description:"Name of the Kubernetes service." json:"name,omitempty" toml:"name,omitempty" yaml:"name,omitempty"`
	Namespace string `description:"Namespace of the Kubernetes service." json:"namespace,omitempty" toml:"namespace,omitempty" yaml:"namespace,omitempty"`
}

// BuildFilterFunc returns the name of the filter and the related dynamic.Middleware if needed.
type BuildFilterFunc func(name, namespace string) (string, *dynamic.Middleware, error)

// BuildBackendFunc returns the name of the backend and the related dynamic.Service if needed.
type BuildBackendFunc func(name, namespace string) (string, *dynamic.Service, error)

type ExtensionBuilderRegistry interface {
	RegisterFilterFuncs(group, kind string, builderFunc BuildFilterFunc)
	RegisterBackendFuncs(group, kind string, builderFunc BuildBackendFunc)
}

// RegisterFilterFuncs registers an allowed Group, Kind, and builder for the Filter ExtensionRef objects.
func (p *Provider) RegisterFilterFuncs(group, kind string, builderFunc BuildFilterFunc) {
	if p.groupKindFilterFuncs == nil {
		p.groupKindFilterFuncs = map[string]map[string]BuildFilterFunc{}
	}

	if p.groupKindFilterFuncs[group] == nil {
		p.groupKindFilterFuncs[group] = map[string]BuildFilterFunc{}
	}

	p.groupKindFilterFuncs[group][kind] = builderFunc
}

// RegisterBackendFuncs registers an allowed Group, Kind, and builder for the Backend ExtensionRef objects.
func (p *Provider) RegisterBackendFuncs(group, kind string, builderFunc BuildBackendFunc) {
	if p.groupKindBackendFuncs == nil {
		p.groupKindBackendFuncs = map[string]map[string]BuildBackendFunc{}
	}

	if p.groupKindBackendFuncs[group] == nil {
		p.groupKindBackendFuncs[group] = map[string]BuildBackendFunc{}
	}

	p.groupKindBackendFuncs[group][kind] = builderFunc
}

func (p *Provider) SetRouterTransform(routerTransform k8s.RouterTransform) {
	p.routerTransform = routerTransform
}

func (p *Provider) applyRouterTransform(ctx context.Context, rt *dynamic.Router, route *gatev1.HTTPRoute) {
	if p.routerTransform == nil {
		return
	}

	err := p.routerTransform.Apply(ctx, rt, route)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Apply router transform")
	}
}

func (p *Provider) newK8sClient(ctx context.Context) (*clientWrapper, error) {
	// Label selector validation
	_, err := labels.Parse(p.LabelSelector)
	if err != nil {
		return nil, fmt.Errorf("invalid label selector: %q", p.LabelSelector)
	}

	logger := log.Ctx(ctx)
	logger.Info().Msgf("Label selector is: %q", p.LabelSelector)

	var client *clientWrapper
	switch {
	case os.Getenv("KUBERNETES_SERVICE_HOST") != "" && os.Getenv("KUBERNETES_SERVICE_PORT") != "":
		logger.Info().Str("endpoint", p.Endpoint).Msg("Creating in-cluster Provider client")
		client, err = newInClusterClient(p.Endpoint)
	case os.Getenv("KUBECONFIG") != "":
		logger.Info().Msgf("Creating cluster-external Provider client from KUBECONFIG %s", os.Getenv("KUBECONFIG"))
		client, err = newExternalClusterClientFromFile(os.Getenv("KUBECONFIG"))
	default:
		logger.Info().Str("endpoint", p.Endpoint).Msg("Creating cluster-external Provider client")
		client, err = newExternalClusterClient(p.Endpoint, p.CertAuthFilePath, p.Token)
	}

	if err != nil {
		return nil, err
	}

	client.labelSelector = p.LabelSelector
	client.experimentalChannel = p.ExperimentalChannel

	return client, nil
}

// Init the provider.
func (p *Provider) Init() error {
	return nil
}

// Provide allows the k8s provider to provide configurations to traefik using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	logger := log.With().Str(logs.ProviderName, providerName).Logger()
	ctxLog := logger.WithContext(context.Background())

	k8sClient, err := p.newK8sClient(ctxLog)
	if err != nil {
		return err
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
					conf := p.loadConfigurationFromGateway(ctxLog, k8sClient)

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

// TODO Handle errors and update resources statuses (gatewayClass, gateway).
func (p *Provider) loadConfigurationFromGateway(ctx context.Context, client Client) *dynamic.Configuration {
	logger := log.Ctx(ctx)

	gatewayClassNames := map[string]struct{}{}

	gatewayClasses, err := client.GetGatewayClasses()
	if err != nil {
		logger.Error().Err(err).Msg("Cannot find GatewayClasses")
		return &dynamic.Configuration{
			HTTP: &dynamic.HTTPConfiguration{
				Routers:           map[string]*dynamic.Router{},
				Middlewares:       map[string]*dynamic.Middleware{},
				Services:          map[string]*dynamic.Service{},
				ServersTransports: map[string]*dynamic.ServersTransport{},
			},
			TCP: &dynamic.TCPConfiguration{
				Routers:           map[string]*dynamic.TCPRouter{},
				Middlewares:       map[string]*dynamic.TCPMiddleware{},
				Services:          map[string]*dynamic.TCPService{},
				ServersTransports: map[string]*dynamic.TCPServersTransport{},
			},
			UDP: &dynamic.UDPConfiguration{
				Routers:  map[string]*dynamic.UDPRouter{},
				Services: map[string]*dynamic.UDPService{},
			},
			TLS: &dynamic.TLSConfiguration{},
		}
	}

	for _, gatewayClass := range gatewayClasses {
		if gatewayClass.Spec.ControllerName == controllerName {
			gatewayClassNames[gatewayClass.Name] = struct{}{}

			err := client.UpdateGatewayClassStatus(gatewayClass, metav1.Condition{
				Type:               string(gatev1.GatewayClassConditionStatusAccepted),
				Status:             metav1.ConditionTrue,
				ObservedGeneration: gatewayClass.Generation,
				Reason:             "Handled",
				Message:            "Handled by Traefik controller",
				LastTransitionTime: metav1.Now(),
			})
			if err != nil {
				logger.Error().Err(err).Msgf("Failed to update %s condition", gatev1.GatewayClassConditionStatusAccepted)
			}
		}
	}

	cfgs := map[string]*dynamic.Configuration{}

	// TODO check if we can only use the default filtering mechanism
	for _, gateway := range client.GetGateways() {
		logger := log.Ctx(ctx).With().Str("gateway", gateway.Name).Str("namespace", gateway.Namespace).Logger()
		ctxLog := logger.WithContext(ctx)

		if _, ok := gatewayClassNames[string(gateway.Spec.GatewayClassName)]; !ok {
			continue
		}

		cfg, err := p.createGatewayConf(ctxLog, client, gateway)
		if err != nil {
			logger.Error().Err(err).Send()
			continue
		}

		cfgs[gateway.Name+gateway.Namespace] = cfg
	}

	conf := provider.Merge(ctx, cfgs)

	conf.TLS = &dynamic.TLSConfiguration{}

	for _, cfg := range cfgs {
		if conf.TLS == nil {
			conf.TLS = &dynamic.TLSConfiguration{}
		}

		conf.TLS.Certificates = append(conf.TLS.Certificates, cfg.TLS.Certificates...)

		for name, options := range cfg.TLS.Options {
			if conf.TLS.Options == nil {
				conf.TLS.Options = map[string]tls.Options{}
			}

			conf.TLS.Options[name] = options
		}

		for name, store := range cfg.TLS.Stores {
			if conf.TLS.Stores == nil {
				conf.TLS.Stores = map[string]tls.Store{}
			}

			conf.TLS.Stores[name] = store
		}
	}

	return conf
}

func (p *Provider) createGatewayConf(ctx context.Context, client Client, gateway *gatev1.Gateway) (*dynamic.Configuration, error) {
	addresses, err := p.gatewayAddresses(client)
	if err != nil {
		return nil, fmt.Errorf("get Gateway status addresses: %w", err)
	}

	conf := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers:           map[string]*dynamic.Router{},
			Middlewares:       map[string]*dynamic.Middleware{},
			Services:          map[string]*dynamic.Service{},
			ServersTransports: map[string]*dynamic.ServersTransport{},
		},
		TCP: &dynamic.TCPConfiguration{
			Routers:           map[string]*dynamic.TCPRouter{},
			Middlewares:       map[string]*dynamic.TCPMiddleware{},
			Services:          map[string]*dynamic.TCPService{},
			ServersTransports: map[string]*dynamic.TCPServersTransport{},
		},
		UDP: &dynamic.UDPConfiguration{
			Routers:  map[string]*dynamic.UDPRouter{},
			Services: map[string]*dynamic.UDPService{},
		},
		TLS: &dynamic.TLSConfiguration{},
	}

	tlsConfigs := make(map[string]*tls.CertAndStores)

	// GatewayReasonListenersNotValid is used when one or more
	// Listeners have an invalid or unsupported configuration
	// and cannot be configured on the Gateway.
	listenerStatuses, httpRouteParentStatuses := p.fillGatewayConf(ctx, client, gateway, conf, tlsConfigs)

	if len(tlsConfigs) > 0 {
		conf.TLS.Certificates = append(conf.TLS.Certificates, getTLSConfig(tlsConfigs)...)
	}

	httpRouteStatuses := makeHTTPRouteStatuses(gateway.Namespace, httpRouteParentStatuses)
	for nsName, status := range httpRouteStatuses {
		if err := client.UpdateHTTPRouteStatus(ctx, gateway, nsName, status); err != nil {
			log.Error().
				Err(err).
				Str("namespace", nsName.Namespace).
				Str("name", nsName.Name).
				Msg("Unable to update HTTPRoute status")
		}
	}

	gatewayStatus, errG := p.makeGatewayStatus(gateway, listenerStatuses, addresses)
	if err = client.UpdateGatewayStatus(gateway, gatewayStatus); err != nil {
		log.Error().
			Err(err).
			Str("namespace", gateway.Namespace).
			Str("name", gateway.Name).
			Msg("Unable to update Gateway status")
	}
	if errG != nil {
		return nil, fmt.Errorf("creating gateway status: %w", errG)
	}

	return conf, nil
}

func (p *Provider) fillGatewayConf(ctx context.Context, client Client, gateway *gatev1.Gateway, conf *dynamic.Configuration, tlsConfigs map[string]*tls.CertAndStores) ([]gatev1.ListenerStatus, map[ktypes.NamespacedName][]gatev1.RouteParentStatus) {
	logger := log.Ctx(ctx)
	allocatedListeners := make(map[string]struct{})
	listenerStatuses := make([]gatev1.ListenerStatus, len(gateway.Spec.Listeners))
	httpRouteParentStatuses := make(map[ktypes.NamespacedName][]gatev1.RouteParentStatus)

	for i, listener := range gateway.Spec.Listeners {
		listenerStatuses[i] = gatev1.ListenerStatus{
			Name:           listener.Name,
			SupportedKinds: []gatev1.RouteGroupKind{},
			Conditions:     []metav1.Condition{},
			// AttachedRoutes: 0 TODO Set to number of Routes associated with a Listener regardless of Gateway or Route status
		}

		supportedKinds, conditions := supportedRouteKinds(listener.Protocol, p.ExperimentalChannel)
		if len(conditions) > 0 {
			listenerStatuses[i].Conditions = append(listenerStatuses[i].Conditions, conditions...)
			continue
		}

		routeKinds, conditions := getAllowedRouteKinds(gateway, listener, supportedKinds)
		listenerStatuses[i].SupportedKinds = routeKinds
		if len(conditions) > 0 {
			listenerStatuses[i].Conditions = append(listenerStatuses[i].Conditions, conditions...)
			continue
		}

		listenerKey := makeListenerKey(listener)

		if _, ok := allocatedListeners[listenerKey]; ok {
			listenerStatuses[i].Conditions = append(listenerStatuses[i].Conditions, metav1.Condition{
				Type:               string(gatev1.ListenerConditionConflicted),
				Status:             metav1.ConditionTrue,
				ObservedGeneration: gateway.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             "DuplicateListener",
				Message:            "A listener with same protocol, port and hostname already exists",
			})

			continue
		}

		allocatedListeners[listenerKey] = struct{}{}

		ep, err := p.entryPointName(listener.Port, listener.Protocol)
		if err != nil {
			// update "Detached" status with "PortUnavailable" reason
			listenerStatuses[i].Conditions = append(listenerStatuses[i].Conditions, metav1.Condition{
				Type:               string(gatev1.ListenerConditionAccepted),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: gateway.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.ListenerReasonPortUnavailable),
				Message:            fmt.Sprintf("Cannot find entryPoint for Gateway: %v", err),
			})

			continue
		}

		if (listener.Protocol == gatev1.HTTPProtocolType || listener.Protocol == gatev1.TCPProtocolType) && listener.TLS != nil {
			listenerStatuses[i].Conditions = append(listenerStatuses[i].Conditions, metav1.Condition{
				Type:               string(gatev1.ListenerConditionAccepted),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: gateway.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             "InvalidTLSConfiguration", // TODO check the spec if a proper reason is introduced at some point
				Message:            "TLS configuration must no be defined when using HTTP or TCP protocol",
			})

			continue
		}

		// TLS
		if listener.Protocol == gatev1.HTTPSProtocolType || listener.Protocol == gatev1.TLSProtocolType {
			if listener.TLS == nil || (len(listener.TLS.CertificateRefs) == 0 && listener.TLS.Mode != nil && *listener.TLS.Mode != gatev1.TLSModePassthrough) {
				// update "Detached" status with "UnsupportedProtocol" reason
				listenerStatuses[i].Conditions = append(listenerStatuses[i].Conditions, metav1.Condition{
					Type:               string(gatev1.ListenerConditionAccepted),
					Status:             metav1.ConditionFalse,
					ObservedGeneration: gateway.Generation,
					LastTransitionTime: metav1.Now(),
					Reason:             "InvalidTLSConfiguration", // TODO check the spec if a proper reason is introduced at some point
					Message: fmt.Sprintf("No TLS configuration for Gateway Listener %s:%d and protocol %q",
						listener.Name, listener.Port, listener.Protocol),
				})

				continue
			}

			var tlsModeType gatev1.TLSModeType
			if listener.TLS.Mode != nil {
				tlsModeType = *listener.TLS.Mode
			}

			isTLSPassthrough := tlsModeType == gatev1.TLSModePassthrough

			if isTLSPassthrough && len(listener.TLS.CertificateRefs) > 0 {
				// https://gateway-api.sigs.k8s.io/v1alpha2/references/spec/#gateway.networking.k8s.io/v1alpha2.GatewayTLSConfig
				logger.Warn().Msg("In case of Passthrough TLS mode, no TLS settings take effect as the TLS session from the client is NOT terminated at the Gateway")
			}

			// Allowed configurations:
			// Protocol TLS -> Passthrough -> TLSRoute/TCPRoute
			// Protocol TLS -> Terminate -> TLSRoute/TCPRoute
			// Protocol HTTPS -> Terminate -> HTTPRoute
			if listener.Protocol == gatev1.HTTPSProtocolType && isTLSPassthrough {
				listenerStatuses[i].Conditions = append(listenerStatuses[i].Conditions, metav1.Condition{
					Type:               string(gatev1.ListenerConditionAccepted),
					Status:             metav1.ConditionFalse,
					ObservedGeneration: gateway.Generation,
					LastTransitionTime: metav1.Now(),
					Reason:             string(gatev1.ListenerReasonUnsupportedProtocol),
					Message:            "HTTPS protocol is not supported with TLS mode Passthrough",
				})

				continue
			}

			if !isTLSPassthrough {
				if len(listener.TLS.CertificateRefs) == 0 {
					// update "ResolvedRefs" status true with "InvalidCertificateRef" reason
					listenerStatuses[i].Conditions = append(listenerStatuses[i].Conditions, metav1.Condition{
						Type:               string(gatev1.ListenerConditionResolvedRefs),
						Status:             metav1.ConditionFalse,
						ObservedGeneration: gateway.Generation,
						LastTransitionTime: metav1.Now(),
						Reason:             string(gatev1.ListenerReasonInvalidCertificateRef),
						Message:            "One TLS CertificateRef is required in Terminate mode",
					})

					continue
				}

				// TODO Should we support multiple certificates?
				certificateRef := listener.TLS.CertificateRefs[0]

				if certificateRef.Kind == nil || *certificateRef.Kind != "Secret" ||
					certificateRef.Group == nil || (*certificateRef.Group != "" && *certificateRef.Group != groupCore) {
					// update "ResolvedRefs" status true with "InvalidCertificateRef" reason
					listenerStatuses[i].Conditions = append(listenerStatuses[i].Conditions, metav1.Condition{
						Type:               string(gatev1.ListenerConditionResolvedRefs),
						Status:             metav1.ConditionFalse,
						ObservedGeneration: gateway.Generation,
						LastTransitionTime: metav1.Now(),
						Reason:             string(gatev1.ListenerReasonInvalidCertificateRef),
						Message:            fmt.Sprintf("Unsupported TLS CertificateRef group/kind: %s/%s", groupToString(certificateRef.Group), kindToString(certificateRef.Kind)),
					})

					continue
				}

				certificateNamespace := gateway.Namespace
				if certificateRef.Namespace != nil && string(*certificateRef.Namespace) != gateway.Namespace {
					certificateNamespace = string(*certificateRef.Namespace)
				}

				if certificateNamespace != gateway.Namespace {
					referenceGrants, err := client.GetReferenceGrants(certificateNamespace)
					if err != nil {
						listenerStatuses[i].Conditions = append(listenerStatuses[i].Conditions, metav1.Condition{
							Type:               string(gatev1.ListenerConditionResolvedRefs),
							Status:             metav1.ConditionFalse,
							ObservedGeneration: gateway.Generation,
							LastTransitionTime: metav1.Now(),
							Reason:             string(gatev1.ListenerReasonRefNotPermitted),
							Message:            fmt.Sprintf("Cannot find any ReferenceGrant: %v", err),
						})
						continue
					}

					referenceGrants = filterReferenceGrantsFrom(referenceGrants, "gateway.networking.k8s.io", "Gateway", gateway.Namespace)
					referenceGrants = filterReferenceGrantsTo(referenceGrants, groupCore, "Secret", string(certificateRef.Name))
					if len(referenceGrants) == 0 {
						listenerStatuses[i].Conditions = append(listenerStatuses[i].Conditions, metav1.Condition{
							Type:               string(gatev1.ListenerConditionResolvedRefs),
							Status:             metav1.ConditionFalse,
							ObservedGeneration: gateway.Generation,
							LastTransitionTime: metav1.Now(),
							Reason:             string(gatev1.ListenerReasonRefNotPermitted),
							Message:            "Required ReferenceGrant for cross namespace secret reference is missing",
						})

						continue
					}
				}

				configKey := certificateNamespace + "/" + string(certificateRef.Name)
				if _, tlsExists := tlsConfigs[configKey]; !tlsExists {
					tlsConf, err := getTLS(client, certificateRef.Name, certificateNamespace)
					if err != nil {
						// update "ResolvedRefs" status false with "InvalidCertificateRef" reason
						// update "Programmed" status false with "Invalid" reason
						listenerStatuses[i].Conditions = append(listenerStatuses[i].Conditions,
							metav1.Condition{
								Type:               string(gatev1.ListenerConditionResolvedRefs),
								Status:             metav1.ConditionFalse,
								ObservedGeneration: gateway.Generation,
								LastTransitionTime: metav1.Now(),
								Reason:             string(gatev1.ListenerReasonInvalidCertificateRef),
								Message:            fmt.Sprintf("Error while retrieving certificate: %v", err),
							},
							metav1.Condition{
								Type:               string(gatev1.ListenerConditionProgrammed),
								Status:             metav1.ConditionFalse,
								ObservedGeneration: gateway.Generation,
								LastTransitionTime: metav1.Now(),
								Reason:             string(gatev1.ListenerReasonInvalid),
								Message:            fmt.Sprintf("Error while retrieving certificate: %v", err),
							},
						)

						continue
					}
					tlsConfigs[configKey] = tlsConf
				}
			}
		}

		for _, routeKind := range routeKinds {
			switch routeKind.Kind {
			case kindHTTPRoute:
				listenerConditions, routeStatuses := p.gatewayHTTPRouteToHTTPConf(ctx, ep, listener, gateway, client, conf)
				listenerStatuses[i].Conditions = append(listenerStatuses[i].Conditions, listenerConditions...)
				for nsName, status := range routeStatuses {
					httpRouteParentStatuses[nsName] = append(httpRouteParentStatuses[nsName], status)
				}

			case kindTCPRoute:
				listenerStatuses[i].Conditions = append(listenerStatuses[i].Conditions, gatewayTCPRouteToTCPConf(ctx, ep, listener, gateway, client, conf)...)
			case kindTLSRoute:
				listenerStatuses[i].Conditions = append(listenerStatuses[i].Conditions, gatewayTLSRouteToTCPConf(ctx, ep, listener, gateway, client, conf)...)
			}
		}
	}

	return listenerStatuses, httpRouteParentStatuses
}

func (p *Provider) makeGatewayStatus(gateway *gatev1.Gateway, listenerStatuses []gatev1.ListenerStatus, addresses []gatev1.GatewayStatusAddress) (gatev1.GatewayStatus, error) {
	gatewayStatus := gatev1.GatewayStatus{Addresses: addresses}

	var result error
	for i, listener := range listenerStatuses {
		if len(listener.Conditions) == 0 {
			listenerStatuses[i].Conditions = append(listenerStatuses[i].Conditions,
				metav1.Condition{
					Type:               string(gatev1.ListenerConditionAccepted),
					Status:             metav1.ConditionTrue,
					ObservedGeneration: gateway.Generation,
					LastTransitionTime: metav1.Now(),
					Reason:             string(gatev1.ListenerReasonAccepted),
					Message:            "No error found",
				},
				metav1.Condition{
					Type:               string(gatev1.ListenerConditionResolvedRefs),
					Status:             metav1.ConditionTrue,
					ObservedGeneration: gateway.Generation,
					LastTransitionTime: metav1.Now(),
					Reason:             string(gatev1.ListenerReasonResolvedRefs),
					Message:            "No error found",
				},
				metav1.Condition{
					Type:               string(gatev1.ListenerConditionProgrammed),
					Status:             metav1.ConditionTrue,
					ObservedGeneration: gateway.Generation,
					LastTransitionTime: metav1.Now(),
					Reason:             string(gatev1.ListenerReasonProgrammed),
					Message:            "No error found",
				},
			)

			continue
		}

		for _, condition := range listener.Conditions {
			result = multierror.Append(result, errors.New(condition.Message))
		}
	}
	gatewayStatus.Listeners = listenerStatuses

	if result != nil {
		// GatewayConditionReady "Ready", GatewayConditionReason "ListenersNotValid"
		gatewayStatus.Conditions = append(gatewayStatus.Conditions, metav1.Condition{
			Type:               string(gatev1.GatewayConditionAccepted),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: gateway.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.GatewayReasonListenersNotValid),
			Message:            "All Listeners must be valid",
		})

		return gatewayStatus, result
	}

	gatewayStatus.Conditions = append(gatewayStatus.Conditions,
		// update "Accepted" status with "Accepted" reason
		metav1.Condition{
			Type:               string(gatev1.GatewayConditionAccepted),
			Status:             metav1.ConditionTrue,
			ObservedGeneration: gateway.Generation,
			Reason:             string(gatev1.GatewayReasonAccepted),
			Message:            "Gateway successfully scheduled",
			LastTransitionTime: metav1.Now(),
		},
		// update "Programmed" status with "Programmed" reason
		metav1.Condition{
			Type:               string(gatev1.GatewayConditionProgrammed),
			Status:             metav1.ConditionTrue,
			ObservedGeneration: gateway.Generation,
			Reason:             string(gatev1.GatewayReasonProgrammed),
			Message:            "Gateway successfully scheduled",
			LastTransitionTime: metav1.Now(),
		},
	)

	return gatewayStatus, nil
}

func (p *Provider) gatewayAddresses(client Client) ([]gatev1.GatewayStatusAddress, error) {
	if p.StatusAddress == nil {
		return nil, nil
	}

	if p.StatusAddress.IP != "" {
		return []gatev1.GatewayStatusAddress{{
			Type:  ptr.To(gatev1.IPAddressType),
			Value: p.StatusAddress.IP,
		}}, nil
	}

	if p.StatusAddress.Hostname != "" {
		return []gatev1.GatewayStatusAddress{{
			Type:  ptr.To(gatev1.HostnameAddressType),
			Value: p.StatusAddress.Hostname,
		}}, nil
	}

	svcRef := p.StatusAddress.Service
	if svcRef.Name != "" && svcRef.Namespace != "" {
		svc, exists, err := client.GetService(svcRef.Namespace, svcRef.Name)
		if err != nil {
			return nil, fmt.Errorf("unable to get service: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("could not find a service with name %s in namespace %s", svcRef.Name, svcRef.Namespace)
		}

		var addresses []gatev1.GatewayStatusAddress
		for _, addr := range svc.Status.LoadBalancer.Ingress {
			switch {
			case addr.IP != "":
				addresses = append(addresses, gatev1.GatewayStatusAddress{
					Type:  ptr.To(gatev1.IPAddressType),
					Value: addr.IP,
				})

			case addr.Hostname != "":
				addresses = append(addresses, gatev1.GatewayStatusAddress{
					Type:  ptr.To(gatev1.HostnameAddressType),
					Value: addr.Hostname,
				})
			}
		}
		return addresses, nil
	}

	return nil, errors.New("empty Gateway status address configuration")
}

func (p *Provider) entryPointName(port gatev1.PortNumber, protocol gatev1.ProtocolType) (string, error) {
	portStr := strconv.FormatInt(int64(port), 10)

	for name, entryPoint := range p.EntryPoints {
		if strings.HasSuffix(entryPoint.Address, ":"+portStr) {
			// If the protocol is HTTP the entryPoint must have no TLS conf
			// Not relevant for gatev1.TLSProtocolType && gatev1.TCPProtocolType
			if protocol == gatev1.HTTPProtocolType && entryPoint.HasHTTPTLSConf {
				continue
			}

			return name, nil
		}
	}

	return "", fmt.Errorf("no matching entryPoint for port %d and protocol %q", port, protocol)
}

func (p *Provider) gatewayHTTPRouteToHTTPConf(ctx context.Context, ep string, listener gatev1.Listener, gateway *gatev1.Gateway, client Client, conf *dynamic.Configuration) ([]metav1.Condition, map[ktypes.NamespacedName]gatev1.RouteParentStatus) {
	// Should not happen due to validation.
	if listener.AllowedRoutes == nil {
		return nil, nil
	}

	namespaces, err := getRouteBindingSelectorNamespace(client, gateway.Namespace, listener.AllowedRoutes.Namespaces)
	if err != nil {
		// update "ResolvedRefs" status true with "InvalidRoutesRef" reason
		return []metav1.Condition{{
			Type:               string(gatev1.ListenerConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: gateway.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             "InvalidRouteNamespacesSelector", // Should never happen as the selector is validated by kubernetes
			Message:            fmt.Sprintf("Invalid route namespaces selector: %v", err),
		}}, nil
	}

	routes, err := client.GetHTTPRoutes(namespaces)
	if err != nil {
		// update "ResolvedRefs" status true with "RefNotPermitted" reason
		return []metav1.Condition{{
			Type:               string(gatev1.ListenerConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: gateway.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.ListenerReasonRefNotPermitted),
			Message:            fmt.Sprintf("Cannot fetch HTTPRoutes: %v", err),
		}}, nil
	}

	if len(routes) == 0 {
		log.Ctx(ctx).Debug().Msg("No HTTPRoutes found")
		return nil, nil
	}

	routeStatuses := map[ktypes.NamespacedName]gatev1.RouteParentStatus{}
	for _, route := range routes {
		routeNsName := ktypes.NamespacedName{Namespace: route.Namespace, Name: route.Name}

		parentRef, ok := shouldAttach(gateway, listener, route.Namespace, route.Spec.CommonRouteSpec)
		if !ok {
			// TODO: to add an invalid HTTPRoute status when no parent is matching,
			//   we have to start the attachment evaluation from the route not from the listeners.
			//   This will fix the HTTPRouteInvalidParentRefNotMatchingSectionName test.
			continue
		}

		routeConditions := []metav1.Condition{
			{
				Type:               string(gatev1.RouteConditionAccepted),
				Status:             metav1.ConditionTrue,
				ObservedGeneration: route.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.RouteReasonAccepted),
			},
			{
				Type:               string(gatev1.RouteConditionResolvedRefs),
				Status:             metav1.ConditionTrue,
				ObservedGeneration: route.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.RouteConditionResolvedRefs),
			},
		}

		hostnames := matchingHostnames(listener, route.Spec.Hostnames)
		if len(hostnames) == 0 && listener.Hostname != nil && *listener.Hostname != "" && len(route.Spec.Hostnames) > 0 {
			// TODO update the corresponding route parent status.
			// https://gateway-api.sigs.k8s.io/v1alpha2/references/spec/#gateway.networking.k8s.io/v1alpha2.TLSRoute
			continue
		}

		hostRule, err := hostRule(hostnames)
		if err != nil {
			// TODO update the route status condition.
			continue
		}

		for _, routeRule := range route.Spec.Rules {
			rule, err := extractRule(routeRule, hostRule)
			if err != nil {
				// TODO update the route status condition.
				continue
			}

			router := dynamic.Router{
				Rule:        rule,
				RuleSyntax:  "v3",
				EntryPoints: []string{ep},
			}

			if listener.Protocol == gatev1.HTTPSProtocolType && listener.TLS != nil {
				// TODO support let's encrypt.
				router.TLS = &dynamic.RouterTLSConfig{}
			}

			// Adding the gateway desc and the entryPoint desc prevents overlapping of routers build from the same routes.
			routerName := route.Name + "-" + gateway.Name + "-" + ep
			routerKey, err := makeRouterKey(router.Rule, makeID(route.Namespace, routerName))
			if err != nil {
				// TODO update the route status condition.
				continue
			}

			middlewares, err := p.loadMiddlewares(listener, route.Namespace, routerKey, routeRule.Filters)
			if err != nil {
				// TODO update the route status condition.
				continue
			}

			for middlewareName, middleware := range middlewares {
				// If the middleware is not defined in the return of the loadMiddlewares function, it means we just need a reference to that middleware.
				if middleware != nil {
					conf.HTTP.Middlewares[middlewareName] = middleware
				}

				router.Middlewares = append(router.Middlewares, middlewareName)
			}

			if len(routeRule.BackendRefs) == 0 {
				continue
			}

			// Traefik internal service can be used only if there is only one BackendRef service reference.
			if len(routeRule.BackendRefs) == 1 && isInternalService(routeRule.BackendRefs[0].BackendRef) {
				router.Service = string(routeRule.BackendRefs[0].Name)
			} else {
				var wrr dynamic.WeightedRoundRobin
				for _, backendRef := range routeRule.BackendRefs {
					weight := ptr.To(int(ptr.Deref(backendRef.Weight, 1)))

					name, svc, errCondition := p.loadHTTPService(client, route, backendRef)
					if errCondition != nil {
						routeConditions = appendCondition(routeConditions, *errCondition)
						wrr.Services = append(wrr.Services, dynamic.WRRService{
							Name:   name,
							Weight: weight,
							Status: ptr.To(500),
						})
						continue
					}

					if svc != nil {
						conf.HTTP.Services[name] = svc
					}

					wrr.Services = append(wrr.Services, dynamic.WRRService{
						Name:   name,
						Weight: weight,
					})
				}

				wrrName := provider.Normalize(routerKey + "-wrr")
				conf.HTTP.Services[wrrName] = &dynamic.Service{Weighted: &wrr}

				router.Service = wrrName
			}

			rt := &router
			p.applyRouterTransform(ctx, rt, route)

			routerKey = provider.Normalize(routerKey)
			conf.HTTP.Routers[routerKey] = rt
		}

		routeStatuses[routeNsName] = gatev1.RouteParentStatus{
			ParentRef:      parentRef,
			ControllerName: controllerName,
			Conditions:     routeConditions,
		}
	}

	return nil, routeStatuses
}

// loadHTTPService returns a dynamic.Service config corresponding to the given gatev1.HTTPBackendRef.
// Note that the returned dynamic.Service config can be nil (for cross-provider, internal services, and backendFunc).
func (p *Provider) loadHTTPService(client Client, route *gatev1.HTTPRoute, backendRef gatev1.HTTPBackendRef) (string, *dynamic.Service, *metav1.Condition) {
	group := groupCore
	if backendRef.Group != nil && *backendRef.Group != "" {
		group = string(*backendRef.Group)
	}

	kind := ptr.Deref(backendRef.Kind, "Service")
	namespace := ptr.Deref(backendRef.Namespace, gatev1.Namespace(route.Namespace))
	namespaceStr := string(namespace)
	serviceName := provider.Normalize(makeID(namespaceStr, string(backendRef.Name)))

	if group != groupCore || kind != "Service" {
		// TODO support cross namespace through ReferencePolicy.
		if namespaceStr != route.Namespace {
			return serviceName, nil, &metav1.Condition{
				Type:               string(gatev1.RouteConditionResolvedRefs),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: route.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.RouteReasonRefNotPermitted),
				Message:            fmt.Sprintf("Cannot load HTTPBackendRef %s/%s/%s/%s namespace not allowed", group, kind, namespace, backendRef.Name),
			}
		}

		name, service, err := p.loadHTTPBackendRef(namespaceStr, backendRef)
		if err != nil {
			return serviceName, nil, &metav1.Condition{
				Type:               string(gatev1.RouteConditionResolvedRefs),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: route.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.RouteReasonInvalidKind),
				Message:            fmt.Sprintf("Cannot load HTTPBackendRef %s/%s/%s/%s: %s", group, kind, namespace, backendRef.Name, err),
			}
		}

		return name, service, nil
	}

	port := ptr.Deref(backendRef.Port, gatev1.PortNumber(0))
	if port == 0 {
		return serviceName, nil, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonUnsupportedProtocol),
			Message:            fmt.Sprintf("Cannot load HTTPBackendRef %s/%s/%s/%s port is required", group, kind, namespace, backendRef.Name),
		}
	}

	portStr := strconv.FormatInt(int64(port), 10)
	serviceName = provider.Normalize(serviceName + "-" + portStr)

	lb, err := loadHTTPServers(client, namespaceStr, backendRef)
	if err != nil {
		return serviceName, nil, &metav1.Condition{
			Type:               string(gatev1.RouteConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: route.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.RouteReasonBackendNotFound),
			Message:            fmt.Sprintf("Cannot load HTTPBackendRef %s/%s/%s/%s: %s", group, kind, namespace, backendRef.Name, err),
		}
	}

	return serviceName, &dynamic.Service{LoadBalancer: lb}, nil
}

func (p *Provider) loadHTTPBackendRef(namespace string, backendRef gatev1.HTTPBackendRef) (string, *dynamic.Service, error) {
	// Support for cross-provider references (e.g: api@internal).
	// This provides the same behavior as for IngressRoutes.
	if *backendRef.Kind == "TraefikService" && strings.Contains(string(backendRef.Name), "@") {
		return string(backendRef.Name), nil, nil
	}

	backendFunc, ok := p.groupKindBackendFuncs[string(*backendRef.Group)][string(*backendRef.Kind)]
	if !ok {
		return "", nil, fmt.Errorf("unsupported HTTPBackendRef %s/%s/%s", *backendRef.Group, *backendRef.Kind, backendRef.Name)
	}
	if backendFunc == nil {
		return "", nil, fmt.Errorf("undefined backendFunc for HTTPBackendRef %s/%s/%s", *backendRef.Group, *backendRef.Kind, backendRef.Name)
	}

	return backendFunc(string(backendRef.Name), namespace)
}

func (p *Provider) loadMiddlewares(listener gatev1.Listener, namespace string, prefix string, filters []gatev1.HTTPRouteFilter) (map[string]*dynamic.Middleware, error) {
	middlewares := make(map[string]*dynamic.Middleware)

	// The spec allows for an empty string in which case we should use the
	// scheme of the request which in this case is the listener scheme.
	var listenerScheme string
	switch listener.Protocol {
	case gatev1.HTTPProtocolType:
		listenerScheme = "http"
	case gatev1.HTTPSProtocolType:
		listenerScheme = "https"
	default:
		return nil, fmt.Errorf("invalid listener protocol %s", listener.Protocol)
	}

	for i, filter := range filters {
		var middleware *dynamic.Middleware
		switch filter.Type {
		case gatev1.HTTPRouteFilterRequestRedirect:
			var err error
			middleware, err = createRedirectRegexMiddleware(listenerScheme, filter.RequestRedirect)
			if err != nil {
				return nil, fmt.Errorf("creating RedirectRegex middleware: %w", err)
			}

			middlewareName := provider.Normalize(fmt.Sprintf("%s-%s-%d", prefix, strings.ToLower(string(filter.Type)), i))
			middlewares[middlewareName] = middleware
		case gatev1.HTTPRouteFilterExtensionRef:
			name, middleware, err := p.loadHTTPRouteFilterExtensionRef(namespace, filter.ExtensionRef)
			if err != nil {
				return nil, fmt.Errorf("unsupported filter %s: %w", filter.Type, err)
			}

			middlewares[name] = middleware

		case gatev1.HTTPRouteFilterRequestHeaderModifier:
			middlewareName := provider.Normalize(fmt.Sprintf("%s-%s-%d", prefix, strings.ToLower(string(filter.Type)), i))
			middlewares[middlewareName] = createRequestHeaderModifier(filter.RequestHeaderModifier)

		default:
			// As per the spec:
			// https://gateway-api.sigs.k8s.io/api-types/httproute/#filters-optional
			// In all cases where incompatible or unsupported filters are
			// specified, implementations MUST add a warning condition to
			// status.
			return nil, fmt.Errorf("unsupported filter %s", filter.Type)
		}
	}

	return middlewares, nil
}

func (p *Provider) loadHTTPRouteFilterExtensionRef(namespace string, extensionRef *gatev1.LocalObjectReference) (string, *dynamic.Middleware, error) {
	if extensionRef == nil {
		return "", nil, errors.New("filter extension ref undefined")
	}

	filterFunc, ok := p.groupKindFilterFuncs[string(extensionRef.Group)][string(extensionRef.Kind)]
	if !ok {
		return "", nil, fmt.Errorf("unsupported filter extension ref %s/%s/%s", extensionRef.Group, extensionRef.Kind, extensionRef.Name)
	}
	if filterFunc == nil {
		return "", nil, fmt.Errorf("undefined filterFunc for filter extension ref %s/%s/%s", extensionRef.Group, extensionRef.Kind, extensionRef.Name)
	}

	return filterFunc(string(extensionRef.Name), namespace)
}

// TODO support cross namespace through ReferencePolicy.
func loadHTTPServers(client Client, namespace string, backendRef gatev1.HTTPBackendRef) (*dynamic.ServersLoadBalancer, error) {
	service, exists, err := client.GetService(namespace, string(backendRef.Name))
	if err != nil {
		return nil, fmt.Errorf("getting service: %w", err)
	}
	if !exists {
		return nil, errors.New("service not found")
	}

	var portSpec corev1.ServicePort
	var match bool

	for _, p := range service.Spec.Ports {
		if backendRef.Port == nil || p.Port == int32(*backendRef.Port) {
			portSpec = p
			match = true
			break
		}
	}
	if !match {
		return nil, errors.New("service port not found")
	}

	endpoints, endpointsExists, err := client.GetEndpoints(namespace, string(backendRef.Name))
	if err != nil {
		return nil, fmt.Errorf("getting endpoints: %w", err)
	}
	if !endpointsExists {
		return nil, errors.New("endpoints not found")
	}

	if len(endpoints.Subsets) == 0 {
		return nil, errors.New("subset not found")
	}

	lb := &dynamic.ServersLoadBalancer{}
	lb.SetDefaults()

	var port int32
	var portStr string
	for _, subset := range endpoints.Subsets {
		for _, p := range subset.Ports {
			if portSpec.Name == p.Name {
				port = p.Port
				break
			}
		}

		if port == 0 {
			return nil, errors.New("cannot define a port")
		}

		protocol := getProtocol(portSpec)

		portStr = strconv.FormatInt(int64(port), 10)
		for _, addr := range subset.Addresses {
			lb.Servers = append(lb.Servers, dynamic.Server{
				URL: fmt.Sprintf("%s://%s", protocol, net.JoinHostPort(addr.IP, portStr)),
			})
		}
	}

	return lb, nil
}

// loadTCPServices is generating a WRR service, even when there is only one target.
func loadTCPServices(client Client, namespace string, backendRefs []gatev1.BackendRef) (*dynamic.TCPService, map[string]*dynamic.TCPService, error) {
	services := map[string]*dynamic.TCPService{}

	wrrSvc := &dynamic.TCPService{
		Weighted: &dynamic.TCPWeightedRoundRobin{
			Services: []dynamic.TCPWRRService{},
		},
	}

	for _, backendRef := range backendRefs {
		if backendRef.Group == nil || backendRef.Kind == nil {
			// Should not happen as this is validated by kubernetes
			continue
		}

		if isInternalService(backendRef) {
			return nil, nil, fmt.Errorf("traefik internal service %s is not allowed in a WRR loadbalancer", backendRef.Name)
		}

		weight := int(ptr.Deref(backendRef.Weight, 1))

		if isTraefikService(backendRef) {
			wrrSvc.Weighted.Services = append(wrrSvc.Weighted.Services, dynamic.TCPWRRService{Name: string(backendRef.Name), Weight: &weight})
			continue
		}

		if *backendRef.Group != "" && *backendRef.Group != groupCore && *backendRef.Kind != "Service" {
			return nil, nil, fmt.Errorf("unsupported BackendRef %s/%s/%s", *backendRef.Group, *backendRef.Kind, backendRef.Name)
		}

		svc := dynamic.TCPService{
			LoadBalancer: &dynamic.TCPServersLoadBalancer{},
		}

		service, exists, err := client.GetService(namespace, string(backendRef.Name))
		if err != nil {
			return nil, nil, err
		}

		if !exists {
			return nil, nil, errors.New("service not found")
		}

		if len(service.Spec.Ports) > 1 && backendRef.Port == nil {
			// If the port is unspecified and the backend is a Service
			// object consisting of multiple port definitions, the route
			// must be dropped from the Gateway. The controller should
			// raise the "ResolvedRefs" condition on the Gateway with the
			// "DroppedRoutes" reason. The gateway status for this route
			// should be updated with a condition that describes the error
			// more specifically.
			log.Error().Msg("A multiple ports Kubernetes Service cannot be used if unspecified backendRef.Port")
			continue
		}

		var portSpec corev1.ServicePort
		var match bool

		for _, p := range service.Spec.Ports {
			if backendRef.Port == nil || p.Port == int32(*backendRef.Port) {
				portSpec = p
				match = true
				break
			}
		}

		if !match {
			return nil, nil, errors.New("service port not found")
		}

		endpoints, endpointsExists, endpointsErr := client.GetEndpoints(namespace, string(backendRef.Name))
		if endpointsErr != nil {
			return nil, nil, endpointsErr
		}

		if !endpointsExists {
			return nil, nil, errors.New("endpoints not found")
		}

		if len(endpoints.Subsets) == 0 {
			return nil, nil, errors.New("subset not found")
		}

		var port int32
		var portStr string
		for _, subset := range endpoints.Subsets {
			for _, p := range subset.Ports {
				if portSpec.Name == p.Name {
					port = p.Port
					break
				}
			}

			if port == 0 {
				return nil, nil, errors.New("cannot define a port")
			}

			portStr = strconv.FormatInt(int64(port), 10)
			for _, addr := range subset.Addresses {
				svc.LoadBalancer.Servers = append(svc.LoadBalancer.Servers, dynamic.TCPServer{
					Address: net.JoinHostPort(addr.IP, portStr),
				})
			}
		}

		serviceName := provider.Normalize(makeID(service.Namespace, service.Name) + "-" + portStr)
		services[serviceName] = &svc

		wrrSvc.Weighted.Services = append(wrrSvc.Weighted.Services, dynamic.TCPWRRService{Name: serviceName, Weight: &weight})
	}

	if len(wrrSvc.Weighted.Services) == 0 {
		return nil, nil, errors.New("no service has been created")
	}

	return wrrSvc, services, nil
}

func supportedRouteKinds(protocol gatev1.ProtocolType, experimentalChannel bool) ([]gatev1.RouteGroupKind, []metav1.Condition) {
	group := gatev1.Group(gatev1.GroupName)

	switch protocol {
	case gatev1.TCPProtocolType:
		if experimentalChannel {
			return []gatev1.RouteGroupKind{{Kind: kindTCPRoute, Group: &group}}, nil
		}

		return nil, []metav1.Condition{{
			Type:               string(gatev1.ListenerConditionConflicted),
			Status:             metav1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.ListenerReasonInvalidRouteKinds),
			Message:            fmt.Sprintf("Protocol %q requires the experimental channel support to be enabled, please use the `experimentalChannel` option", protocol),
		}}

	case gatev1.HTTPProtocolType, gatev1.HTTPSProtocolType:
		return []gatev1.RouteGroupKind{{Kind: kindHTTPRoute, Group: &group}}, nil

	case gatev1.TLSProtocolType:
		if experimentalChannel {
			return []gatev1.RouteGroupKind{
				{Kind: kindTCPRoute, Group: &group},
				{Kind: kindTLSRoute, Group: &group},
			}, nil
		}

		return nil, []metav1.Condition{{
			Type:               string(gatev1.ListenerConditionConflicted),
			Status:             metav1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.ListenerReasonInvalidRouteKinds),
			Message:            fmt.Sprintf("Protocol %q requires the experimental channel support to be enabled, please use the `experimentalChannel` option", protocol),
		}}
	}

	return nil, []metav1.Condition{{
		Type:               string(gatev1.ListenerConditionAccepted),
		Status:             metav1.ConditionFalse,
		LastTransitionTime: metav1.Now(),
		Reason:             string(gatev1.ListenerReasonUnsupportedProtocol),
		Message:            fmt.Sprintf("Unsupported listener protocol %q", protocol),
	}}
}

func getAllowedRouteKinds(gateway *gatev1.Gateway, listener gatev1.Listener, supportedKinds []gatev1.RouteGroupKind) ([]gatev1.RouteGroupKind, []metav1.Condition) {
	if listener.AllowedRoutes == nil || len(listener.AllowedRoutes.Kinds) == 0 {
		return supportedKinds, nil
	}

	var (
		routeKinds = []gatev1.RouteGroupKind{}
		conditions []metav1.Condition
	)

	uniqRouteKinds := map[gatev1.Kind]struct{}{}
	for _, routeKind := range listener.AllowedRoutes.Kinds {
		var isSupported bool
		for _, kind := range supportedKinds {
			if routeKind.Kind == kind.Kind && routeKind.Group != nil && *routeKind.Group == *kind.Group {
				isSupported = true
				break
			}
		}

		if !isSupported {
			conditions = append(conditions, metav1.Condition{
				Type:               string(gatev1.ListenerConditionResolvedRefs),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: gateway.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.ListenerReasonInvalidRouteKinds),
				Message:            fmt.Sprintf("Listener protocol %q does not support RouteGroupKind %s/%s", listener.Protocol, groupToString(routeKind.Group), routeKind.Kind),
			})
			continue
		}

		if _, exists := uniqRouteKinds[routeKind.Kind]; !exists {
			routeKinds = append(routeKinds, routeKind)
			uniqRouteKinds[routeKind.Kind] = struct{}{}
		}
	}

	return routeKinds, conditions
}

func gatewayTCPRouteToTCPConf(ctx context.Context, ep string, listener gatev1.Listener, gateway *gatev1.Gateway, client Client, conf *dynamic.Configuration) []metav1.Condition {
	if listener.AllowedRoutes == nil {
		// Should not happen due to validation.
		return nil
	}

	namespaces, err := getRouteBindingSelectorNamespace(client, gateway.Namespace, listener.AllowedRoutes.Namespaces)
	if err != nil {
		// update "ResolvedRefs" status true with "InvalidRoutesRef" reason
		return []metav1.Condition{{
			Type:               string(gatev1.ListenerConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: gateway.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             "InvalidRouteNamespacesSelector", // TODO should never happen as the selector is validated by Kubernetes
			Message:            fmt.Sprintf("Invalid route namespaces selector: %v", err),
		}}
	}

	routes, err := client.GetTCPRoutes(namespaces)
	if err != nil {
		// update "ResolvedRefs" status true with "InvalidRoutesRef" reason
		return []metav1.Condition{{
			Type:               string(gatev1.ListenerConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: gateway.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.ListenerReasonRefNotPermitted),
			Message:            fmt.Sprintf("Cannot fetch TCPRoutes: %v", err),
		}}
	}

	if len(routes) == 0 {
		log.Ctx(ctx).Debug().Msg("No TCPRoutes found")
		return nil
	}

	var conditions []metav1.Condition
	for _, route := range routes {
		if _, ok := shouldAttach(gateway, listener, route.Namespace, route.Spec.CommonRouteSpec); !ok {
			continue
		}

		router := dynamic.TCPRouter{
			Rule:        "HostSNI(`*`)",
			EntryPoints: []string{ep},
			RuleSyntax:  "v3",
		}

		if listener.Protocol == gatev1.TLSProtocolType && listener.TLS != nil {
			// TODO support let's encrypt
			router.TLS = &dynamic.RouterTCPTLSConfig{
				Passthrough: listener.TLS.Mode != nil && *listener.TLS.Mode == gatev1.TLSModePassthrough,
			}
		}

		// Adding the gateway desc and the entryPoint desc prevents overlapping of routers build from the same routes.
		routerName := route.Name + "-" + gateway.Name + "-" + ep
		routerKey, err := makeRouterKey("", makeID(route.Namespace, routerName))
		if err != nil {
			// update "ResolvedRefs" status true with "DroppedRoutes" reason
			conditions = append(conditions, metav1.Condition{
				Type:               string(gatev1.ListenerConditionResolvedRefs),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: gateway.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             "InvalidRouterKey", // Should never happen
				Message:            fmt.Sprintf("Skipping TCPRoute %s: cannot make router's key with rule %s: %v", route.Name, router.Rule, err),
			})

			// TODO update the RouteStatus condition / deduplicate conditions on listener
			continue
		}

		routerKey = provider.Normalize(routerKey)

		var ruleServiceNames []string
		for i, rule := range route.Spec.Rules {
			if rule.BackendRefs == nil {
				// Should not happen due to validation.
				// https://github.com/kubernetes-sigs/gateway-api/blob/v0.4.0/apis/v1alpha2/tcproute_types.go#L76
				continue
			}

			wrrService, subServices, err := loadTCPServices(client, route.Namespace, rule.BackendRefs)
			if err != nil {
				// update "ResolvedRefs" status true with "DroppedRoutes" reason
				conditions = append(conditions, metav1.Condition{
					Type:               string(gatev1.ListenerConditionResolvedRefs),
					Status:             metav1.ConditionFalse,
					ObservedGeneration: gateway.Generation,
					LastTransitionTime: metav1.Now(),
					Reason:             "InvalidBackendRefs", // TODO check the spec if a proper reason is introduced at some point
					Message:            fmt.Sprintf("Cannot load TCPRoute service %s/%s: %v", route.Namespace, route.Name, err),
				})

				// TODO update the RouteStatus condition / deduplicate conditions on listener
				continue
			}

			for svcName, svc := range subServices {
				conf.TCP.Services[svcName] = svc
			}

			serviceName := fmt.Sprintf("%s-wrr-%d", routerKey, i)
			conf.TCP.Services[serviceName] = wrrService

			ruleServiceNames = append(ruleServiceNames, serviceName)
		}

		if len(ruleServiceNames) == 1 {
			router.Service = ruleServiceNames[0]
			conf.TCP.Routers[routerKey] = &router
			continue
		}

		routeServiceKey := routerKey + "-wrr"
		routeService := &dynamic.TCPService{Weighted: &dynamic.TCPWeightedRoundRobin{}}

		for _, name := range ruleServiceNames {
			service := dynamic.TCPWRRService{Name: name}
			service.SetDefaults()

			routeService.Weighted.Services = append(routeService.Weighted.Services, service)
		}

		conf.TCP.Services[routeServiceKey] = routeService

		router.Service = routeServiceKey
		conf.TCP.Routers[routerKey] = &router
	}

	return conditions
}

func gatewayTLSRouteToTCPConf(ctx context.Context, ep string, listener gatev1.Listener, gateway *gatev1.Gateway, client Client, conf *dynamic.Configuration) []metav1.Condition {
	if listener.AllowedRoutes == nil {
		// Should not happen due to validation.
		return nil
	}

	namespaces, err := getRouteBindingSelectorNamespace(client, gateway.Namespace, listener.AllowedRoutes.Namespaces)
	if err != nil {
		// update "ResolvedRefs" status true with "InvalidRoutesRef" reason
		return []metav1.Condition{{
			Type:               string(gatev1.ListenerConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: gateway.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             "InvalidRouteNamespacesSelector", // TODO should never happen as the selector is validated by Kubernetes
			Message:            fmt.Sprintf("Invalid route namespaces selector: %v", err),
		}}
	}

	routes, err := client.GetTLSRoutes(namespaces)
	if err != nil {
		// update "ResolvedRefs" status true with "InvalidRoutesRef" reason
		return []metav1.Condition{{
			Type:               string(gatev1.ListenerConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: gateway.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.ListenerReasonRefNotPermitted),
			Message:            fmt.Sprintf("Cannot fetch TLSRoutes: %v", err),
		}}
	}

	if len(routes) == 0 {
		log.Ctx(ctx).Debug().Msg("No TLSRoutes found")
		return nil
	}

	var conditions []metav1.Condition
	for _, route := range routes {
		if _, ok := shouldAttach(gateway, listener, route.Namespace, route.Spec.CommonRouteSpec); !ok {
			continue
		}

		hostnames := matchingHostnames(listener, route.Spec.Hostnames)
		if len(hostnames) == 0 && listener.Hostname != nil && *listener.Hostname != "" && len(route.Spec.Hostnames) > 0 {
			for _, parent := range route.Status.Parents {
				parent.Conditions = append(parent.Conditions, metav1.Condition{
					Type:               string(gatev1.GatewayClassConditionStatusAccepted),
					Status:             metav1.ConditionFalse,
					ObservedGeneration: gateway.Generation,
					Reason:             string(gatev1.ListenerReasonHostnameConflict),
					Message:            fmt.Sprintf("No hostname match between listener: %v and route: %v", listener.Hostname, route.Spec.Hostnames),
					LastTransitionTime: metav1.Now(),
				})
			}

			continue
		}

		rule, err := hostSNIRule(hostnames)
		if err != nil {
			// update "ResolvedRefs" status true with "InvalidHostnames" reason
			conditions = append(conditions, metav1.Condition{
				Type:               string(gatev1.ListenerConditionResolvedRefs),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: gateway.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             "InvalidHostnames", // TODO check the spec if a proper reason is introduced at some point
				Message:            fmt.Sprintf("Skipping TLSRoute %s: cannot make route's SNI match: %v", route.Name, err),
			})
			// TODO update the RouteStatus condition / deduplicate conditions on listener
			continue
		}

		router := dynamic.TCPRouter{
			Rule:        rule,
			RuleSyntax:  "v3",
			EntryPoints: []string{ep},
			TLS: &dynamic.RouterTCPTLSConfig{
				Passthrough: listener.TLS.Mode != nil && *listener.TLS.Mode == gatev1.TLSModePassthrough,
			},
		}

		// Adding the gateway desc and the entryPoint desc prevents overlapping of routers build from the same routes.
		routerName := route.Name + "-" + gateway.Name + "-" + ep
		routerKey, err := makeRouterKey(rule, makeID(route.Namespace, routerName))
		if err != nil {
			// update "ResolvedRefs" status true with "DroppedRoutes" reason
			conditions = append(conditions, metav1.Condition{
				Type:               string(gatev1.ListenerConditionResolvedRefs),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: gateway.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             "InvalidRouterKey", // Should never happen
				Message:            fmt.Sprintf("Skipping TLSRoute %s: cannot make router's key with rule %s: %v", route.Name, router.Rule, err),
			})

			// TODO update the RouteStatus condition / deduplicate conditions on listener
			continue
		}

		routerKey = provider.Normalize(routerKey)

		var ruleServiceNames []string
		for i, routeRule := range route.Spec.Rules {
			if len(routeRule.BackendRefs) == 0 {
				// Should not happen due to validation.
				// https://github.com/kubernetes-sigs/gateway-api/blob/v0.4.0/apis/v1alpha2/tlsroute_types.go#L120
				continue
			}

			wrrService, subServices, err := loadTCPServices(client, route.Namespace, routeRule.BackendRefs)
			if err != nil {
				// update "ResolvedRefs" status true with "InvalidBackendRefs" reason
				conditions = append(conditions, metav1.Condition{
					Type:               string(gatev1.ListenerConditionResolvedRefs),
					Status:             metav1.ConditionFalse,
					ObservedGeneration: gateway.Generation,
					LastTransitionTime: metav1.Now(),
					Reason:             "InvalidBackendRefs", // TODO check the spec if a proper reason is introduced at some point
					Message:            fmt.Sprintf("Cannot load TLSRoute service %s/%s: %v", route.Namespace, route.Name, err),
				})

				// TODO update the RouteStatus condition / deduplicate conditions on listener
				continue
			}

			for svcName, svc := range subServices {
				conf.TCP.Services[svcName] = svc
			}

			serviceName := fmt.Sprintf("%s-wrr-%d", routerKey, i)
			conf.TCP.Services[serviceName] = wrrService

			ruleServiceNames = append(ruleServiceNames, serviceName)
		}

		if len(ruleServiceNames) == 1 {
			router.Service = ruleServiceNames[0]
			conf.TCP.Routers[routerKey] = &router
			continue
		}

		routeServiceKey := routerKey + "-wrr"
		routeService := &dynamic.TCPService{Weighted: &dynamic.TCPWeightedRoundRobin{}}

		for _, name := range ruleServiceNames {
			service := dynamic.TCPWRRService{Name: name}
			service.SetDefaults()

			routeService.Weighted.Services = append(routeService.Weighted.Services, service)
		}

		conf.TCP.Services[routeServiceKey] = routeService

		router.Service = routeServiceKey
		conf.TCP.Routers[routerKey] = &router
	}

	return conditions
}

// Because of Kubernetes validation we admit that the given Hostnames are valid.
// https://github.com/kubernetes-sigs/gateway-api/blob/ff9883da4cad8554cd300394f725ab3a27502785/apis/v1alpha2/shared_types.go#L252
func matchingHostnames(listener gatev1.Listener, hostnames []gatev1.Hostname) []gatev1.Hostname {
	if listener.Hostname == nil || *listener.Hostname == "" {
		return hostnames
	}

	if len(hostnames) == 0 {
		return []gatev1.Hostname{*listener.Hostname}
	}

	listenerLabels := strings.Split(string(*listener.Hostname), ".")

	var matches []gatev1.Hostname

	for _, hostname := range hostnames {
		if hostname == *listener.Hostname {
			matches = append(matches, hostname)
			continue
		}

		hostnameLabels := strings.Split(string(hostname), ".")
		if len(listenerLabels) != len(hostnameLabels) {
			continue
		}

		if !slices.Equal(listenerLabels[1:], hostnameLabels[1:]) {
			continue
		}

		if listenerLabels[0] == "*" {
			matches = append(matches, hostname)
			continue
		}

		if hostnameLabels[0] == "*" {
			matches = append(matches, *listener.Hostname)
			continue
		}
	}

	return matches
}

func shouldAttach(gateway *gatev1.Gateway, listener gatev1.Listener, routeNamespace string, routeSpec gatev1.CommonRouteSpec) (gatev1.ParentReference, bool) {
	for _, parentRef := range routeSpec.ParentRefs {
		if parentRef.Group == nil || *parentRef.Group != gatev1.GroupName {
			continue
		}

		if parentRef.Kind == nil || *parentRef.Kind != kindGateway {
			continue
		}

		if parentRef.SectionName != nil && *parentRef.SectionName != listener.Name {
			continue
		}

		namespace := routeNamespace
		if parentRef.Namespace != nil {
			namespace = string(*parentRef.Namespace)
		}

		if namespace == gateway.Namespace && string(parentRef.Name) == gateway.Name {
			return parentRef, true
		}
	}

	return gatev1.ParentReference{}, false
}

func getRouteBindingSelectorNamespace(client Client, gatewayNamespace string, routeNamespaces *gatev1.RouteNamespaces) ([]string, error) {
	if routeNamespaces == nil || routeNamespaces.From == nil {
		return []string{gatewayNamespace}, nil
	}

	switch *routeNamespaces.From {
	case gatev1.NamespacesFromAll:
		return []string{metav1.NamespaceAll}, nil

	case gatev1.NamespacesFromSame:
		return []string{gatewayNamespace}, nil

	case gatev1.NamespacesFromSelector:
		selector, err := metav1.LabelSelectorAsSelector(routeNamespaces.Selector)
		if err != nil {
			return nil, fmt.Errorf("malformed selector: %w", err)
		}

		return client.GetNamespaces(selector)
	}

	return nil, fmt.Errorf("unsupported RouteSelectType: %q", *routeNamespaces.From)
}

func hostRule(hostnames []gatev1.Hostname) (string, error) {
	var rules []string

	for _, hostname := range hostnames {
		host := string(hostname)
		// When unspecified, "", or *, all hostnames are matched.
		// This field can be omitted for protocols that don't require hostname based matching.
		// TODO Refactor this when building support for TLS options.
		if host == "*" || host == "" {
			return "", nil
		}

		wildcard := strings.Count(host, "*")
		if wildcard == 0 {
			rules = append(rules, fmt.Sprintf("Host(`%s`)", host))
			continue
		}

		// https://gateway-api.sigs.k8s.io/v1alpha2/references/spec/#gateway.networking.k8s.io/v1alpha2.Hostname
		if !strings.HasPrefix(host, "*.") || wildcard > 1 {
			return "", fmt.Errorf("invalid rule: %q", host)
		}

		host = strings.Replace(regexp.QuoteMeta(host), `\*\.`, `[a-zA-Z0-9-]+\.`, 1)
		rules = append(rules, fmt.Sprintf("HostRegexp(`^%s$`)", host))
	}

	switch len(rules) {
	case 0:
		return "", nil
	case 1:
		return rules[0], nil
	default:
		return fmt.Sprintf("(%s)", strings.Join(rules, " || ")), nil
	}
}

func hostSNIRule(hostnames []gatev1.Hostname) (string, error) {
	rules := make([]string, 0, len(hostnames))
	uniqHostnames := map[gatev1.Hostname]struct{}{}

	for _, hostname := range hostnames {
		if len(hostname) == 0 {
			continue
		}

		if _, exists := uniqHostnames[hostname]; exists {
			continue
		}

		host := string(hostname)
		uniqHostnames[hostname] = struct{}{}

		wildcard := strings.Count(host, "*")
		if wildcard == 0 {
			rules = append(rules, fmt.Sprintf("HostSNI(`%s`)", host))
			continue
		}

		if !strings.HasPrefix(host, "*.") || wildcard > 1 {
			return "", fmt.Errorf("invalid rule: %q", host)
		}

		host = strings.Replace(regexp.QuoteMeta(host), `\*\.`, `[a-zA-Z0-9-]+\.`, 1)
		rules = append(rules, fmt.Sprintf("HostSNIRegexp(`^%s$`)", host))
	}

	if len(hostnames) == 0 || len(rules) == 0 {
		return "HostSNI(`*`)", nil
	}

	return strings.Join(rules, " || "), nil
}

func extractRule(routeRule gatev1.HTTPRouteRule, hostRule string) (string, error) {
	var rule string
	var matchesRules []string

	for _, match := range routeRule.Matches {
		if (match.Path == nil || match.Path.Type == nil) && match.Headers == nil {
			continue
		}

		var matchRules []string

		if match.Path != nil && match.Path.Type != nil && match.Path.Value != nil {
			switch *match.Path.Type {
			case gatev1.PathMatchExact:
				matchRules = append(matchRules, fmt.Sprintf("Path(`%s`)", *match.Path.Value))
			case gatev1.PathMatchPathPrefix:
				matchRules = append(matchRules, buildPathMatchPathPrefixRule(*match.Path.Value))
			default:
				return "", fmt.Errorf("unsupported path match type %s", *match.Path.Type)
			}
		}

		headerRules, err := extractHeaderRules(match.Headers)
		if err != nil {
			return "", err
		}

		matchRules = append(matchRules, headerRules...)
		matchesRules = append(matchesRules, strings.Join(matchRules, " && "))
	}

	// If no matches are specified, the default is a prefix
	// path match on "/", which has the effect of matching every
	// HTTP request.
	if len(routeRule.Matches) == 0 {
		matchesRules = append(matchesRules, "PathPrefix(`/`)")
	}

	if hostRule != "" {
		if len(matchesRules) == 0 {
			return hostRule, nil
		}
		rule += hostRule + " && "
	}

	if len(matchesRules) == 1 {
		return rule + matchesRules[0], nil
	}

	if len(rule) == 0 {
		return strings.Join(matchesRules, " || "), nil
	}

	return rule + "(" + strings.Join(matchesRules, " || ") + ")", nil
}

func extractHeaderRules(headers []gatev1.HTTPHeaderMatch) ([]string, error) {
	var headerRules []string

	// TODO handle other headers types
	for _, header := range headers {
		if header.Type == nil {
			// Should never happen due to kubernetes validation.
			continue
		}

		switch *header.Type {
		case gatev1.HeaderMatchExact:
			headerRules = append(headerRules, fmt.Sprintf("Header(`%s`,`%s`)", header.Name, header.Value))
		default:
			return nil, fmt.Errorf("unsupported header match type %s", *header.Type)
		}
	}

	return headerRules, nil
}

func buildPathMatchPathPrefixRule(path string) string {
	if path == "/" {
		return "PathPrefix(`/`)"
	}

	path = strings.TrimSuffix(path, "/")
	return fmt.Sprintf("(Path(`%[1]s`) || PathPrefix(`%[1]s/`))", path)
}

func makeRouterKey(rule, name string) (string, error) {
	h := sha256.New()
	if _, err := h.Write([]byte(rule)); err != nil {
		return "", err
	}

	key := fmt.Sprintf("%s-%.10x", name, h.Sum(nil))

	return key, nil
}

func makeID(namespace, name string) string {
	if namespace == "" {
		return name
	}

	return namespace + "-" + name
}

func getTLS(k8sClient Client, secretName gatev1.ObjectName, namespace string) (*tls.CertAndStores, error) {
	secret, exists, err := k8sClient.GetSecret(namespace, string(secretName))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch secret %s/%s: %w", namespace, secretName, err)
	}
	if !exists {
		return nil, fmt.Errorf("secret %s/%s does not exist", namespace, secretName)
	}

	cert, key, err := getCertificateBlocks(secret, namespace, string(secretName))
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

// createRequestHeaderModifier does not enforce/check the configuration,
// as the spec indicates that either the webhook or CEL (since v1.0 GA Release) should enforce that.
func createRequestHeaderModifier(filter *gatev1.HTTPHeaderFilter) *dynamic.Middleware {
	sets := map[string]string{}
	for _, header := range filter.Set {
		sets[string(header.Name)] = header.Value
	}

	adds := map[string]string{}
	for _, header := range filter.Add {
		adds[string(header.Name)] = header.Value
	}

	return &dynamic.Middleware{
		RequestHeaderModifier: &dynamic.RequestHeaderModifier{
			Set:    sets,
			Add:    adds,
			Remove: filter.Remove,
		},
	}
}

func createRedirectRegexMiddleware(scheme string, filter *gatev1.HTTPRequestRedirectFilter) (*dynamic.Middleware, error) {
	// Use the HTTPRequestRedirectFilter scheme if defined.
	filterScheme := scheme
	if filter.Scheme != nil {
		filterScheme = *filter.Scheme
	}

	if filterScheme != "http" && filterScheme != "https" {
		return nil, fmt.Errorf("invalid scheme %s", filterScheme)
	}

	statusCode := http.StatusFound
	if filter.StatusCode != nil {
		statusCode = *filter.StatusCode
	}

	if statusCode != http.StatusMovedPermanently && statusCode != http.StatusFound {
		return nil, fmt.Errorf("invalid status code %d", statusCode)
	}

	port := "${port}"
	if filter.Port != nil {
		port = fmt.Sprintf(":%d", *filter.Port)
	}

	hostname := "${hostname}"
	if filter.Hostname != nil && *filter.Hostname != "" {
		hostname = string(*filter.Hostname)
	}

	return &dynamic.Middleware{
		RedirectRegex: &dynamic.RedirectRegex{
			Regex:       `^[a-z]+:\/\/(?P<userInfo>.+@)?(?P<hostname>\[[\w:\.]+\]|[\w\._-]+)(?P<port>:\d+)?\/(?P<path>.*)`,
			Replacement: fmt.Sprintf("%s://${userinfo}%s%s/${path}", filterScheme, hostname, port),
			Permanent:   statusCode == http.StatusMovedPermanently,
		},
	}, nil
}

func getProtocol(portSpec corev1.ServicePort) string {
	protocol := "http"
	if portSpec.Port == 443 || strings.HasPrefix(portSpec.Name, "https") {
		protocol = "https"
	}

	return protocol
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

func isTraefikService(ref gatev1.BackendRef) bool {
	if ref.Kind == nil || ref.Group == nil {
		return false
	}

	return *ref.Group == traefikv1alpha1.GroupName && *ref.Kind == kindTraefikService
}

func isInternalService(ref gatev1.BackendRef) bool {
	return isTraefikService(ref) && strings.HasSuffix(string(ref.Name), "@internal")
}

// makeListenerKey joins protocol, hostname, and port of a listener into a string key.
func makeListenerKey(l gatev1.Listener) string {
	var hostname gatev1.Hostname
	if l.Hostname != nil {
		hostname = *l.Hostname
	}

	return fmt.Sprintf("%s|%s|%d", l.Protocol, hostname, l.Port)
}

func filterReferenceGrantsFrom(referenceGrants []*gatev1beta1.ReferenceGrant, group, kind, namespace string) []*gatev1beta1.ReferenceGrant {
	var matchingReferenceGrants []*gatev1beta1.ReferenceGrant
	for _, referenceGrant := range referenceGrants {
		if referenceGrantMatchesFrom(referenceGrant, group, kind, namespace) {
			matchingReferenceGrants = append(matchingReferenceGrants, referenceGrant)
		}
	}
	return matchingReferenceGrants
}

func referenceGrantMatchesFrom(referenceGrant *gatev1beta1.ReferenceGrant, group, kind, namespace string) bool {
	for _, from := range referenceGrant.Spec.From {
		sanitizedGroup := string(from.Group)
		if sanitizedGroup == "" {
			sanitizedGroup = groupCore
		}
		if string(from.Namespace) != namespace || string(from.Kind) != kind || sanitizedGroup != group {
			continue
		}
		return true
	}
	return false
}

func filterReferenceGrantsTo(referenceGrants []*gatev1beta1.ReferenceGrant, group, kind, name string) []*gatev1beta1.ReferenceGrant {
	var matchingReferenceGrants []*gatev1beta1.ReferenceGrant
	for _, referenceGrant := range referenceGrants {
		if referenceGrantMatchesTo(referenceGrant, group, kind, name) {
			matchingReferenceGrants = append(matchingReferenceGrants, referenceGrant)
		}
	}
	return matchingReferenceGrants
}

func referenceGrantMatchesTo(referenceGrant *gatev1beta1.ReferenceGrant, group, kind, name string) bool {
	for _, to := range referenceGrant.Spec.To {
		sanitizedGroup := string(to.Group)
		if sanitizedGroup == "" {
			sanitizedGroup = groupCore
		}
		if string(to.Kind) != kind || sanitizedGroup != group || (to.Name != nil && string(*to.Name) != name) {
			continue
		}
		return true
	}
	return false
}

func groupToString(p *gatev1.Group) string {
	if p == nil {
		return "<nil>"
	}
	return string(*p)
}

func kindToString(p *gatev1.Kind) string {
	if p == nil {
		return "<nil>"
	}
	return string(*p)
}

func makeHTTPRouteStatuses(gwNs string, routeParentStatuses map[ktypes.NamespacedName][]gatev1.RouteParentStatus) map[ktypes.NamespacedName]gatev1.HTTPRouteStatus {
	res := map[ktypes.NamespacedName]gatev1.HTTPRouteStatus{}

	for nsName, parentStatuses := range routeParentStatuses {
		var httpRouteStatus gatev1.HTTPRouteStatus
		for _, parentStatus := range parentStatuses {
			exists := slices.ContainsFunc(httpRouteStatus.Parents, func(status gatev1.RouteParentStatus) bool {
				return parentRefEquals(gwNs, parentStatus.ParentRef, status.ParentRef)
			})
			if !exists {
				httpRouteStatus.Parents = append(httpRouteStatus.Parents, parentStatus)
			}
		}

		res[nsName] = httpRouteStatus
	}

	return res
}

func parentRefEquals(gwNs string, p1, p2 gatev1.ParentReference) bool {
	if !pointerEquals(p1.Group, p2.Group) {
		return false
	}

	if !pointerEquals(p1.Kind, p2.Kind) {
		return false
	}

	if !pointerEquals(p1.SectionName, p2.SectionName) {
		return false
	}

	if p1.Name != p2.Name {
		return false
	}

	p1Ns := gwNs
	if p1.Namespace != nil {
		p1Ns = string(*p1.Namespace)
	}

	p2Ns := gwNs
	if p2.Namespace != nil {
		p2Ns = string(*p2.Namespace)
	}

	return p1Ns == p2Ns
}

func pointerEquals[T comparable](p1, p2 *T) bool {
	if p1 == nil && p2 == nil {
		return true
	}

	var val1 T
	if p1 != nil {
		val1 = *p1
	}

	var val2 T
	if p2 != nil {
		val2 = *p2
	}

	return val1 == val2
}

func appendCondition(conditions []metav1.Condition, condition metav1.Condition) []metav1.Condition {
	res := []metav1.Condition{condition}
	for _, c := range conditions {
		if c.Type != condition.Type {
			res = append(res, c)
		}
	}

	return res
}
