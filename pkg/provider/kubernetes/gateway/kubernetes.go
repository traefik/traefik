package gateway

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
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
	providerName = "kubernetesgateway"

	controllerName = "traefik.io/gateway-controller"

	groupCore    = "core"
	groupGateway = "gateway.networking.k8s.io"

	kindGateway        = "Gateway"
	kindTraefikService = "TraefikService"
	kindHTTPRoute      = "HTTPRoute"
	kindGRPCRoute      = "GRPCRoute"
	kindTCPRoute       = "TCPRoute"
	kindTLSRoute       = "TLSRoute"
	kindService        = "Service"

	appProtocolHTTP  = "http"
	appProtocolHTTPS = "https"
	appProtocolH2C   = "kubernetes.io/h2c"
	appProtocolWS    = "kubernetes.io/ws"
	appProtocolWSS   = "kubernetes.io/wss"

	schemeHTTP  = "http"
	schemeHTTPS = "https"
	schemeH2C   = "h2c"
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
	NativeLBByDefault   bool                `description:"Defines whether to use Native Kubernetes load-balancing by default." json:"nativeLBByDefault,omitempty" toml:"nativeLBByDefault,omitempty" yaml:"nativeLBByDefault,omitempty" export:"true"`

	EntryPoints map[string]Entrypoint `json:"-" toml:"-" yaml:"-" label:"-" file:"-"`

	// groupKindFilterFuncs is the list of allowed Group and Kinds for the Filter ExtensionRef objects.
	groupKindFilterFuncs map[string]map[string]BuildFilterFunc
	// groupKindBackendFuncs is the list of allowed Group and Kinds for the Backend ExtensionRef objects.
	groupKindBackendFuncs map[string]map[string]BuildBackendFunc

	lastConfiguration safe.Safe

	routerTransform k8s.RouterTransform
	client          *clientWrapper
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

type gatewayListener struct {
	Name string

	Port              gatev1.PortNumber
	Protocol          gatev1.ProtocolType
	TLS               *gatev1.GatewayTLSConfig
	Hostname          *gatev1.Hostname
	Status            *gatev1.ListenerStatus
	AllowedNamespaces []string
	AllowedRouteKinds []string

	Attached bool

	GWName       string
	GWNamespace  string
	GWGeneration int64
	EPName       string
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

	if err := p.routerTransform.Apply(ctx, rt, route); err != nil {
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
	logger := log.With().Str(logs.ProviderName, providerName).Logger()

	var err error
	p.client, err = p.newK8sClient(logger.WithContext(context.Background()))
	if err != nil {
		return fmt.Errorf("creating k8s client: %w", err)
	}

	return nil
}

// Provide allows the k8s provider to provide configurations to traefik using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	logger := log.With().Str(logs.ProviderName, providerName).Logger()
	ctxLog := logger.WithContext(context.Background())

	pool.GoCtx(func(ctxPool context.Context) {
		operation := func() error {
			eventsChan, err := p.client.WatchAll(p.Namespaces, ctxPool.Done())
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
					conf := p.loadConfigurationFromGateways(ctxLog)

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
func (p *Provider) loadConfigurationFromGateways(ctx context.Context) *dynamic.Configuration {
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

	addresses, err := p.gatewayAddresses()
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Unable to get Gateway status addresses")
		return nil
	}

	gatewayClasses, err := p.client.ListGatewayClasses()
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Unable to list GatewayClasses")
		return nil
	}

	var supportedFeatures []gatev1.SupportedFeature
	if p.ExperimentalChannel {
		for _, feature := range SupportedFeatures() {
			supportedFeatures = append(supportedFeatures, gatev1.SupportedFeature{Name: gatev1.FeatureName(feature)})
		}
		slices.SortFunc(supportedFeatures, func(a, b gatev1.SupportedFeature) int {
			return strings.Compare(string(a.Name), string(b.Name))
		})
	}

	gatewayClassNames := map[string]struct{}{}
	for _, gatewayClass := range gatewayClasses {
		if gatewayClass.Spec.ControllerName != controllerName {
			continue
		}

		gatewayClassNames[gatewayClass.Name] = struct{}{}

		status := gatev1.GatewayClassStatus{
			Conditions: upsertGatewayClassConditionAccepted(gatewayClass.Status.Conditions, metav1.Condition{
				Type:               string(gatev1.GatewayClassConditionStatusAccepted),
				Status:             metav1.ConditionTrue,
				ObservedGeneration: gatewayClass.Generation,
				Reason:             "Handled",
				Message:            "Handled by Traefik controller",
				LastTransitionTime: metav1.Now(),
			}),
			SupportedFeatures: supportedFeatures,
		}

		if err := p.client.UpdateGatewayClassStatus(ctx, gatewayClass.Name, status); err != nil {
			log.Ctx(ctx).
				Warn().
				Err(err).
				Str("gateway_class", gatewayClass.Name).
				Msg("Unable to update GatewayClass status")
		}
	}

	var gateways []*gatev1.Gateway
	for _, gateway := range p.client.ListGateways() {
		if _, ok := gatewayClassNames[string(gateway.Spec.GatewayClassName)]; !ok {
			continue
		}
		gateways = append(gateways, gateway)
	}

	var gatewayListeners []gatewayListener
	for _, gateway := range gateways {
		logger := log.Ctx(ctx).With().
			Str("gateway", gateway.Name).
			Str("namespace", gateway.Namespace).
			Logger()

		gatewayListeners = append(gatewayListeners, p.loadGatewayListeners(logger.WithContext(ctx), gateway, conf)...)
	}

	p.loadHTTPRoutes(ctx, gatewayListeners, conf)

	p.loadGRPCRoutes(ctx, gatewayListeners, conf)

	if p.ExperimentalChannel {
		p.loadTCPRoutes(ctx, gatewayListeners, conf)
		p.loadTLSRoutes(ctx, gatewayListeners, conf)
	}

	for _, gateway := range gateways {
		logger := log.Ctx(ctx).With().
			Str("gateway", gateway.Name).
			Str("namespace", gateway.Namespace).
			Logger()

		var listeners []gatewayListener
		for _, listener := range gatewayListeners {
			if listener.GWName == gateway.Name && listener.GWNamespace == gateway.Namespace {
				listeners = append(listeners, listener)
			}
		}

		gatewayStatus, errConditions := p.makeGatewayStatus(gateway, listeners, addresses)
		if len(errConditions) > 0 {
			messages := map[string]struct{}{}
			for _, condition := range errConditions {
				messages[condition.Message] = struct{}{}
			}
			var conditionsErr error
			for message := range messages {
				conditionsErr = multierror.Append(conditionsErr, errors.New(message))
			}
			logger.Error().
				Err(conditionsErr).
				Msg("Gateway Not Accepted")
		}

		if err = p.client.UpdateGatewayStatus(ctx, ktypes.NamespacedName{Name: gateway.Name, Namespace: gateway.Namespace}, gatewayStatus); err != nil {
			logger.Warn().
				Err(err).
				Msg("Unable to update Gateway status")
		}
	}

	return conf
}

func (p *Provider) loadGatewayListeners(ctx context.Context, gateway *gatev1.Gateway, conf *dynamic.Configuration) []gatewayListener {
	tlsConfigs := make(map[string]*tls.CertAndStores)
	allocatedListeners := make(map[string]struct{})
	gatewayListeners := make([]gatewayListener, len(gateway.Spec.Listeners))

	for i, listener := range gateway.Spec.Listeners {
		gatewayListeners[i] = gatewayListener{
			Name:         string(listener.Name),
			GWName:       gateway.Name,
			GWNamespace:  gateway.Namespace,
			GWGeneration: gateway.Generation,
			Port:         listener.Port,
			Protocol:     listener.Protocol,
			TLS:          listener.TLS,
			Hostname:     listener.Hostname,
			Status: &gatev1.ListenerStatus{
				Name:           listener.Name,
				SupportedKinds: []gatev1.RouteGroupKind{},
				Conditions:     []metav1.Condition{},
			},
		}

		ep, err := p.entryPointName(listener.Port, listener.Protocol)
		if err != nil {
			// update "Detached" status with "PortUnavailable" reason
			gatewayListeners[i].Status.Conditions = append(gatewayListeners[i].Status.Conditions, metav1.Condition{
				Type:               string(gatev1.ListenerConditionAccepted),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: gateway.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.ListenerReasonPortUnavailable),
				Message:            fmt.Sprintf("Cannot find entryPoint for Gateway: %v", err),
			})

			continue
		}
		gatewayListeners[i].EPName = ep

		allowedRoutes := ptr.Deref(listener.AllowedRoutes, gatev1.AllowedRoutes{Namespaces: &gatev1.RouteNamespaces{From: ptr.To(gatev1.NamespacesFromSame)}})
		gatewayListeners[i].AllowedNamespaces, err = p.allowedNamespaces(gateway.Namespace, allowedRoutes.Namespaces)
		if err != nil {
			// update "ResolvedRefs" status true with "InvalidRoutesRef" reason
			gatewayListeners[i].Status.Conditions = append(gatewayListeners[i].Status.Conditions, metav1.Condition{
				Type:               string(gatev1.ListenerConditionResolvedRefs),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: gateway.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             "InvalidRouteNamespacesSelector", // Should never happen as the selector is validated by kubernetes
				Message:            fmt.Sprintf("Invalid route namespaces selector: %v", err),
			})

			continue
		}

		supportedKinds, conditions := supportedRouteKinds(listener.Protocol, p.ExperimentalChannel)
		if len(conditions) > 0 {
			gatewayListeners[i].Status.Conditions = append(gatewayListeners[i].Status.Conditions, conditions...)
			continue
		}

		routeKinds, conditions := allowedRouteKinds(gateway, listener, supportedKinds)
		for _, kind := range routeKinds {
			gatewayListeners[i].AllowedRouteKinds = append(gatewayListeners[i].AllowedRouteKinds, string(kind.Kind))
		}
		gatewayListeners[i].Status.SupportedKinds = routeKinds
		if len(conditions) > 0 {
			gatewayListeners[i].Status.Conditions = append(gatewayListeners[i].Status.Conditions, conditions...)
			continue
		}

		listenerKey := makeListenerKey(listener)

		if _, ok := allocatedListeners[listenerKey]; ok {
			gatewayListeners[i].Status.Conditions = append(gatewayListeners[i].Status.Conditions, metav1.Condition{
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

		if (listener.Protocol == gatev1.HTTPProtocolType || listener.Protocol == gatev1.TCPProtocolType) && listener.TLS != nil {
			gatewayListeners[i].Status.Conditions = append(gatewayListeners[i].Status.Conditions, metav1.Condition{
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
				gatewayListeners[i].Status.Conditions = append(gatewayListeners[i].Status.Conditions, metav1.Condition{
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
				log.Ctx(ctx).Warn().Msg("In case of Passthrough TLS mode, no TLS settings take effect as the TLS session from the client is NOT terminated at the Gateway")
			}

			// Allowed configurations:
			// Protocol TLS -> Passthrough -> TLSRoute/TCPRoute
			// Protocol TLS -> Terminate -> TLSRoute/TCPRoute
			// Protocol HTTPS -> Terminate -> HTTPRoute
			if listener.Protocol == gatev1.HTTPSProtocolType && isTLSPassthrough {
				gatewayListeners[i].Status.Conditions = append(gatewayListeners[i].Status.Conditions, metav1.Condition{
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
					gatewayListeners[i].Status.Conditions = append(gatewayListeners[i].Status.Conditions, metav1.Condition{
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
					gatewayListeners[i].Status.Conditions = append(gatewayListeners[i].Status.Conditions, metav1.Condition{
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

				if err := p.isReferenceGranted(kindGateway, gateway.Namespace, groupCore, "Secret", string(certificateRef.Name), certificateNamespace); err != nil {
					gatewayListeners[i].Status.Conditions = append(gatewayListeners[i].Status.Conditions, metav1.Condition{
						Type:               string(gatev1.ListenerConditionResolvedRefs),
						Status:             metav1.ConditionFalse,
						ObservedGeneration: gateway.Generation,
						LastTransitionTime: metav1.Now(),
						Reason:             string(gatev1.ListenerReasonRefNotPermitted),
						Message:            fmt.Sprintf("Cannot load CertificateRef %s/%s: %s", certificateNamespace, certificateRef.Name, err),
					})

					continue
				}

				configKey := certificateNamespace + "/" + string(certificateRef.Name)
				if _, tlsExists := tlsConfigs[configKey]; !tlsExists {
					tlsConf, err := p.getTLS(certificateRef.Name, certificateNamespace)
					if err != nil {
						// update "ResolvedRefs" status false with "InvalidCertificateRef" reason
						// update "Programmed" status false with "Invalid" reason
						gatewayListeners[i].Status.Conditions = append(gatewayListeners[i].Status.Conditions,
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

		gatewayListeners[i].Attached = true
	}

	if len(tlsConfigs) > 0 {
		conf.TLS.Certificates = append(conf.TLS.Certificates, getTLSConfig(tlsConfigs)...)
	}

	return gatewayListeners
}

func (p *Provider) makeGatewayStatus(gateway *gatev1.Gateway, listeners []gatewayListener, addresses []gatev1.GatewayStatusAddress) (gatev1.GatewayStatus, []metav1.Condition) {
	gatewayStatus := gatev1.GatewayStatus{Addresses: addresses}

	var errorConditions []metav1.Condition
	for _, listener := range listeners {
		if len(listener.Status.Conditions) == 0 {
			listener.Status.Conditions = append(listener.Status.Conditions,
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

			// TODO: refactor
			gatewayStatus.Listeners = append(gatewayStatus.Listeners, *listener.Status)
			continue
		}

		errorConditions = append(errorConditions, listener.Status.Conditions...)
		gatewayStatus.Listeners = append(gatewayStatus.Listeners, *listener.Status)
	}

	if len(errorConditions) > 0 {
		// GatewayConditionReady "Ready", GatewayConditionReason "ListenersNotValid"
		gatewayStatus.Conditions = append(gatewayStatus.Conditions, metav1.Condition{
			Type:               string(gatev1.GatewayConditionAccepted),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: gateway.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.GatewayReasonListenersNotValid),
			Message:            "All Listeners must be valid",
		})

		return gatewayStatus, errorConditions
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

func (p *Provider) gatewayAddresses() ([]gatev1.GatewayStatusAddress, error) {
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
		svc, exists, err := p.client.GetService(svcRef.Namespace, svcRef.Name)
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

func (p *Provider) isReferenceGranted(fromKind, fromNamespace, toGroup, toKind, toName, toNamespace string) error {
	if toNamespace == fromNamespace {
		return nil
	}

	refGrants, err := p.client.ListReferenceGrants(toNamespace)
	if err != nil {
		return fmt.Errorf("listing ReferenceGrant: %w", err)
	}

	refGrants = filterReferenceGrantsFrom(refGrants, groupGateway, fromKind, fromNamespace)
	refGrants = filterReferenceGrantsTo(refGrants, toGroup, toKind, toName)
	if len(refGrants) == 0 {
		return errors.New("missing ReferenceGrant")
	}

	return nil
}

func (p *Provider) getTLS(secretName gatev1.ObjectName, namespace string) (*tls.CertAndStores, error) {
	secret, exists, err := p.client.GetSecret(namespace, string(secretName))
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

func (p *Provider) allowedNamespaces(gatewayNamespace string, routeNamespaces *gatev1.RouteNamespaces) ([]string, error) {
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

		return p.client.ListNamespaces(selector)
	}

	return nil, fmt.Errorf("unsupported RouteSelectType: %q", *routeNamespaces.From)
}

type backendAddress struct {
	IP   string
	Port int32
}

func (p *Provider) getBackendAddresses(namespace string, ref gatev1.BackendRef) ([]backendAddress, corev1.ServicePort, error) {
	if ref.Port == nil {
		return nil, corev1.ServicePort{}, errors.New("port is required for Kubernetes Service reference")
	}

	service, exists, err := p.client.GetService(namespace, string(ref.Name))
	if err != nil {
		return nil, corev1.ServicePort{}, fmt.Errorf("getting service: %w", err)
	}
	if !exists {
		return nil, corev1.ServicePort{}, errors.New("service not found")
	}
	if service.Spec.Type == corev1.ServiceTypeExternalName {
		return nil, corev1.ServicePort{}, errors.New("type ExternalName is not supported for Kubernetes Service reference")
	}

	var svcPort *corev1.ServicePort
	for _, p := range service.Spec.Ports {
		if p.Port == int32(*ref.Port) {
			svcPort = &p
			break
		}
	}
	if svcPort == nil {
		return nil, corev1.ServicePort{}, fmt.Errorf("service port %d not found", *ref.Port)
	}

	annotationsConfig, err := parseServiceAnnotations(service.Annotations)
	if err != nil {
		return nil, corev1.ServicePort{}, fmt.Errorf("parsing service annotations config: %w", err)
	}

	nativeLB := p.NativeLBByDefault
	if annotationsConfig.Service.NativeLB != nil {
		nativeLB = *annotationsConfig.Service.NativeLB
	}

	if nativeLB {
		if service.Spec.ClusterIP == "" || service.Spec.ClusterIP == "None" {
			return nil, corev1.ServicePort{}, fmt.Errorf("no clusterIP found for service: %s/%s", service.Namespace, service.Name)
		}

		return []backendAddress{{
			IP:   service.Spec.ClusterIP,
			Port: svcPort.Port,
		}}, *svcPort, nil
	}

	endpointSlices, err := p.client.ListEndpointSlicesForService(namespace, string(ref.Name))
	if err != nil {
		return nil, corev1.ServicePort{}, fmt.Errorf("getting endpointslices: %w", err)
	}
	if len(endpointSlices) == 0 {
		return nil, corev1.ServicePort{}, errors.New("endpointslices not found")
	}

	uniqAddresses := map[string]struct{}{}
	backendServers := make([]backendAddress, 0)
	for _, endpointSlice := range endpointSlices {
		var port int32
		for _, p := range endpointSlice.Ports {
			if svcPort.Name == *p.Name {
				port = *p.Port
				break
			}
		}
		if port == 0 {
			continue
		}

		for _, endpoint := range endpointSlice.Endpoints {
			if endpoint.Conditions.Ready == nil || !*endpoint.Conditions.Ready {
				continue
			}

			for _, address := range endpoint.Addresses {
				if _, ok := uniqAddresses[address]; ok {
					continue
				}

				uniqAddresses[address] = struct{}{}
				backendServers = append(backendServers, backendAddress{
					IP:   address,
					Port: port,
				})
			}
		}
	}

	return backendServers, *svcPort, nil
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
			Status:             metav1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.ListenerReasonProtocolConflict),
			Message:            fmt.Sprintf("Protocol %q requires the experimental channel support to be enabled, please use the `experimentalChannel` option", protocol),
		}}

	case gatev1.HTTPProtocolType, gatev1.HTTPSProtocolType:
		return []gatev1.RouteGroupKind{
			{Kind: kindHTTPRoute, Group: &group},
			{Kind: kindGRPCRoute, Group: &group},
		}, nil

	case gatev1.TLSProtocolType:
		if experimentalChannel {
			return []gatev1.RouteGroupKind{
				{Kind: kindTCPRoute, Group: &group},
				{Kind: kindTLSRoute, Group: &group},
			}, nil
		}

		return nil, []metav1.Condition{{
			Type:               string(gatev1.ListenerConditionConflicted),
			Status:             metav1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.ListenerReasonInvalidRouteKinds),
			Message:            fmt.Sprintf("Protocol %q requires the experimental channel support to be enabled, please use the `experimentalChannel` option", protocol),
		}}
	}

	return nil, []metav1.Condition{{
		Type:               string(gatev1.ListenerConditionConflicted),
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             string(gatev1.ListenerReasonUnsupportedProtocol),
		Message:            fmt.Sprintf("Unsupported listener protocol %q", protocol),
	}}
}

func allowedRouteKinds(gateway *gatev1.Gateway, listener gatev1.Listener, supportedKinds []gatev1.RouteGroupKind) ([]gatev1.RouteGroupKind, []metav1.Condition) {
	if listener.AllowedRoutes == nil || len(listener.AllowedRoutes.Kinds) == 0 {
		return supportedKinds, nil
	}

	var conditions []metav1.Condition
	routeKinds := []gatev1.RouteGroupKind{}
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

func findMatchingHostnames(listenerHostname *gatev1.Hostname, routeHostnames []gatev1.Hostname) ([]gatev1.Hostname, bool) {
	if listenerHostname == nil {
		return routeHostnames, true
	}

	if len(routeHostnames) == 0 {
		return []gatev1.Hostname{*listenerHostname}, true
	}

	var matches []gatev1.Hostname
	for _, routeHostname := range routeHostnames {
		if match := findMatchingHostname(*listenerHostname, routeHostname); match != "" {
			matches = append(matches, match)
			continue
		}

		if match := findMatchingHostname(routeHostname, *listenerHostname); match != "" {
			matches = append(matches, match)
			continue
		}
	}

	return matches, len(matches) > 0
}

func findMatchingHostname(h1, h2 gatev1.Hostname) gatev1.Hostname {
	if h1 == h2 {
		return h1
	}

	if !strings.HasPrefix(string(h1), "*.") {
		return ""
	}

	trimmedH1 := strings.TrimPrefix(string(h1), "*")
	// root domain doesn't match subdomain wildcard.
	if trimmedH1 == string(h2) {
		return ""
	}

	if !strings.HasSuffix(string(h2), trimmedH1) {
		return ""
	}

	return lessWildcards(h1, h2)
}

func lessWildcards(h1, h2 gatev1.Hostname) gatev1.Hostname {
	if strings.Count(string(h1), "*") > strings.Count(string(h2), "*") {
		return h2
	}

	return h1
}

func allowRoute(listener gatewayListener, routeNamespace, routeKind string) bool {
	if !slices.Contains(listener.AllowedRouteKinds, routeKind) {
		return false
	}

	return slices.ContainsFunc(listener.AllowedNamespaces, func(allowedNamespace string) bool {
		return allowedNamespace == corev1.NamespaceAll || allowedNamespace == routeNamespace
	})
}

func matchingGatewayListeners(gatewayListeners []gatewayListener, routeNamespace string, parentRefs []gatev1.ParentReference) []gatewayListener {
	var listeners []gatewayListener

	for _, listener := range gatewayListeners {
		for _, parentRef := range parentRefs {
			if ptr.Deref(parentRef.Group, gatev1.GroupName) != gatev1.GroupName {
				continue
			}

			if ptr.Deref(parentRef.Kind, kindGateway) != kindGateway {
				continue
			}

			parentRefNamespace := string(ptr.Deref(parentRef.Namespace, gatev1.Namespace(routeNamespace)))
			if listener.GWNamespace != parentRefNamespace {
				continue
			}

			if string(parentRef.Name) != listener.GWName {
				continue
			}

			listeners = append(listeners, listener)
		}
	}

	return listeners
}

func matchListener(listener gatewayListener, parentRef gatev1.ParentReference) bool {
	sectionName := string(ptr.Deref(parentRef.SectionName, ""))
	if sectionName != "" && sectionName != listener.Name {
		return false
	}

	if parentRef.Port != nil && *parentRef.Port != listener.Port {
		return false
	}

	return true
}

func makeRouterName(rule, name string) string {
	h := sha256.New()

	// As explained in https://pkg.go.dev/hash#Hash,
	// Write never returns an error.
	h.Write([]byte(rule))

	return fmt.Sprintf("%s-%.10x", name, h.Sum(nil))
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

func updateRouteConditionAccepted(conditions []metav1.Condition, reason string) []metav1.Condition {
	var conds []metav1.Condition
	for _, c := range conditions {
		if c.Type == string(gatev1.RouteConditionAccepted) && c.Status != metav1.ConditionTrue {
			c.Reason = reason
			c.LastTransitionTime = metav1.Now()

			if reason == string(gatev1.RouteReasonAccepted) {
				c.Status = metav1.ConditionTrue
			}
		}

		conds = append(conds, c)
	}

	return conds
}

func upsertRouteConditionResolvedRefs(conditions []metav1.Condition, condition metav1.Condition) []metav1.Condition {
	var (
		curr  *metav1.Condition
		conds []metav1.Condition
	)
	for _, c := range conditions {
		if c.Type == string(gatev1.RouteConditionResolvedRefs) {
			curr = &c
			continue
		}
		conds = append(conds, c)
	}
	if curr != nil && curr.Status == metav1.ConditionFalse && condition.Status == metav1.ConditionTrue {
		return append(conds, *curr)
	}
	return append(conds, condition)
}

func upsertGatewayClassConditionAccepted(conditions []metav1.Condition, condition metav1.Condition) []metav1.Condition {
	var conds []metav1.Condition
	for _, c := range conditions {
		if c.Type == string(gatev1.GatewayClassConditionStatusAccepted) {
			continue
		}
		conds = append(conds, c)
	}
	return append(conds, condition)
}
