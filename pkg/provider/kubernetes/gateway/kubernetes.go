package gateway

import (
	"cmp"
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
	"github.com/rs/zerolog/log"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/job"
	"github.com/traefik/traefik/v3/pkg/observability/logs"
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
)

const (
	// ProviderName is the Kubernetes Gateway API provider name.
	ProviderName = "kubernetesgateway"

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
	kindListenerSet    = "ListenerSet"
	kindConfigMap      = "ConfigMap"
	kindSecret         = "Secret"

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
	Endpoint                string                `description:"Kubernetes server endpoint (required for external cluster client)." json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Token                   types.FileOrContent   `description:"Kubernetes bearer token (not needed for in-cluster client). It accepts either a token value or a file path to the token." json:"token,omitempty" toml:"token,omitempty" yaml:"token,omitempty" loggable:"false"`
	QPS                     int                   `description:"Defines the maximum QPS to the Kubernetes API server. Setting this to a negative value will disable client-side ratelimiting." json:"qps,omitempty" toml:"qps,omitempty" yaml:"qps,omitempty" export:"true"`
	Burst                   int                   `description:"Defines the maximum burst of requests to the Kubernetes API server." json:"burst,omitempty" toml:"burst,omitempty" yaml:"burst,omitempty" export:"true"`
	CertAuthFilePath        string                `description:"Kubernetes certificate authority file path (not needed for in-cluster client)." json:"certAuthFilePath,omitempty" toml:"certAuthFilePath,omitempty" yaml:"certAuthFilePath,omitempty"`
	Namespaces              []string              `description:"Kubernetes namespaces." json:"namespaces,omitempty" toml:"namespaces,omitempty" yaml:"namespaces,omitempty" export:"true"`
	LabelSelector           string                `description:"Kubernetes label selector to select specific GatewayClasses." json:"labelSelector,omitempty" toml:"labelSelector,omitempty" yaml:"labelSelector,omitempty" export:"true"`
	ThrottleDuration        ptypes.Duration       `description:"Kubernetes refresh throttle duration" json:"throttleDuration,omitempty" toml:"throttleDuration,omitempty" yaml:"throttleDuration,omitempty" export:"true"`
	ExperimentalChannel     bool                  `description:"Toggles Experimental Channel resources support. Requires the Experimental Channel CRDs." json:"experimentalChannel,omitempty" toml:"experimentalChannel,omitempty" yaml:"experimentalChannel,omitempty" export:"true"`
	StatusAddress           *StatusAddress        `description:"Defines the Kubernetes Gateway status address." json:"statusAddress,omitempty" toml:"statusAddress,omitempty" yaml:"statusAddress,omitempty" export:"true"`
	NativeLBByDefault       bool                  `description:"Defines whether to use Native Kubernetes load-balancing by default." json:"nativeLBByDefault,omitempty" toml:"nativeLBByDefault,omitempty" yaml:"nativeLBByDefault,omitempty" export:"true"`
	CrossProviderNamespaces []string              `description:"List of namespaces from which Gateway API routes are allowed to declare TraefikService backendRef references." json:"crossProviderNamespaces,omitempty" toml:"crossProviderNamespaces,omitempty" yaml:"crossProviderNamespaces,omitempty" export:"true"`
	EntryPoints             map[string]Entrypoint `json:"-" toml:"-" yaml:"-" label:"-" file:"-"`

	// groupKindFilterFuncs is the list of allowed Group and Kinds for the Filter ExtensionRef objects.
	groupKindFilterFuncs map[string]map[string]BuildFilterFunc
	// groupKindBackendFuncs is the list of allowed Group and Kinds for the Backend ExtensionRef objects.
	groupKindBackendFuncs map[string]map[string]BuildBackendFunc

	routerTransform k8s.RouterTransform
	client          *clientWrapper
}

func (p *Provider) SetDefaults() {
	p.QPS = 50    // the default value for the QPS is 10x the default Kubernetes client QPS value.
	p.Burst = 100 // the default value for the Burst is 10x the default Kubernetes client Burst value.
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
	TLS               *gatev1.ListenerTLSConfig
	Hostname          *gatev1.Hostname
	Status            *gatev1.ListenerStatus
	AllowedNamespaces []string
	AllowedRouteKinds []string

	Attached bool

	GWName      string
	GWNamespace string
	EPName      string

	// Source tracks whether this listener originated from a Gateway or a ListenerSet.
	// When empty, it is treated as kindGateway for backward compatibility.
	Source          string
	SourceName      string
	SourceNamespace string
}

// source returns the kind of the resource this listener originated from,
// defaulting to kindGateway when unset.
func (l gatewayListener) source() string {
	if l.Source == "" {
		return kindGateway
	}
	return l.Source
}

// policyAncestorRef identifies the resource this listener originated from, so that
// policy statuses reference the ListenerSet instead of a nonexistent Gateway section
// when the route attached through a ListenerSet listener.
func (l gatewayListener) policyAncestorRef(gatewayName, namespace string) gatev1.ParentReference {
	if l.source() == kindListenerSet {
		return gatev1.ParentReference{
			Group:       new(gatev1.Group(gatev1.GroupName)),
			Kind:        new(gatev1.Kind(kindListenerSet)),
			Namespace:   new(gatev1.Namespace(l.SourceNamespace)),
			Name:        gatev1.ObjectName(l.SourceName),
			SectionName: new(gatev1.SectionName(l.Name)),
		}
	}

	return gatev1.ParentReference{
		Group:       new(gatev1.Group(groupGateway)),
		Kind:        new(gatev1.Kind(kindGateway)),
		Namespace:   new(gatev1.Namespace(namespace)),
		Name:        gatev1.ObjectName(gatewayName),
		SectionName: new(gatev1.SectionName(l.Name)),
	}
}

// routeKeySegment returns the segment used in route keys to distinguish between
// Gateway-sourced and ListenerSet-sourced listeners.
// For Gateway-sourced listeners: "gw-<namespace>-<name>"
// For ListenerSet-sourced listeners: "ls-<namespace>-<name>".
func (l gatewayListener) routeKeySegment() string {
	if l.source() == kindListenerSet {
		return fmt.Sprintf("ls-%s-%s", l.SourceNamespace, l.SourceName)
	}
	return fmt.Sprintf("gw-%s-%s", l.GWNamespace, l.GWName)
}

type gatewayWithListeners struct {
	Name      string
	Namespace string

	listeners []gatewayListener

	// listenerSets are the ListenerSets referencing this Gateway, allowed by its
	// AllowedListeners policy or not. They drive route status reporting for
	// ListenerSet parentRefs that resolve to no listener.
	listenerSets []ktypes.NamespacedName
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

// Init the provider.
func (p *Provider) Init() error {
	logger := log.With().Str(logs.ProviderName, ProviderName).Logger()

	var err error
	p.client, err = p.newK8sClient(logger.WithContext(context.Background()))
	if err != nil {
		return fmt.Errorf("creating k8s client: %w", err)
	}

	return nil
}

// Provide allows the k8s provider to provide configurations to traefik using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	logger := log.With().Str(logs.ProviderName, ProviderName).Logger()
	ctxLog := logger.WithContext(context.Background())

	if p.CrossProviderNamespaces != nil {
		logger.Warn().Msgf("Cross-provider references are restricted to namespaces %v (see CrossProviderNamespaces option)", p.CrossProviderNamespaces)
	}

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
				case <-eventsChan:
					// Note that event is the *first* event that came in during this throttling interval -- if we're hitting our throttle, we may have dropped events.
					// This is fine, because we don't treat different event types differently.
					// But if we do in the future, we'll need to track more information about the dropped events.
					conf, statusReport, err := p.loadConfigurationFromGateways(ctxLog)
					if err != nil {
						logger.Error().Err(err).Msg("Unable to load configuration from Gateways")
					} else {
						configurationChan <- dynamic.Message{
							ProviderName:  ProviderName,
							Configuration: conf,
						}

						// Flush regardless of whether the dynamic configuration changed: the
						// statusReport is independent of confHash and may carry writes even
						// when the data plane has nothing new to consume (e.g. a GatewayClass
						// that's now Accepted but has no Gateway pointing at it yet).
						statusReport.Flush(ctxLog, p.client)
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
		client, err = newInClusterClient(p.Endpoint, p.QPS, p.Burst)
	case os.Getenv("KUBECONFIG") != "":
		logger.Info().Msgf("Creating cluster-external Provider client from KUBECONFIG %s", os.Getenv("KUBECONFIG"))
		client, err = newExternalClusterClientFromFile(os.Getenv("KUBECONFIG"), p.QPS, p.Burst)
	default:
		logger.Info().Str("endpoint", p.Endpoint).Msg("Creating cluster-external Provider client")
		client, err = newExternalClusterClient(p.Endpoint, p.CertAuthFilePath, p.Token, p.QPS, p.Burst)
	}

	if err != nil {
		return nil, err
	}

	client.labelSelector = p.LabelSelector

	return client, nil
}

// TODO Handle errors and update resources statuses (gatewayClass, gateway).
func (p *Provider) loadConfigurationFromGateways(ctx context.Context) (*dynamic.Configuration, *statusReport, error) {
	statusReport := newStatusReport()
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
		return nil, nil, fmt.Errorf("getting gateway addresses: %w", err)
	}

	gatewayClasses, err := p.client.ListGatewayClasses()
	if err != nil {
		return nil, nil, fmt.Errorf("listing gateway classes: %w", err)
	}

	var supportedFeatures []gatev1.SupportedFeature
	for _, feature := range SupportedFeatures() {
		supportedFeatures = append(supportedFeatures, gatev1.SupportedFeature{Name: gatev1.FeatureName(feature)})
	}
	slices.SortFunc(supportedFeatures, func(a, b gatev1.SupportedFeature) int {
		return strings.Compare(string(a.Name), string(b.Name))
	})

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

		statusReport.RecordGatewayClassStatus(gatewayClass.Name, status)
	}

	var gateways []*gatev1.Gateway
	for _, gateway := range p.client.ListGateways() {
		if _, ok := gatewayClassNames[string(gateway.Spec.GatewayClassName)]; !ok {
			continue
		}
		gateways = append(gateways, gateway)
	}

	var selectedGateways []gatewayWithListeners
	// listenerSetInfosByGateway tracks per-ListenerSet status info, keyed by gateway NamespacedName.
	listenerSetInfosByGateway := make(map[ktypes.NamespacedName]map[ktypes.NamespacedName]*listenerSetInfo)

	// ListenerSets are listed once and dispatched to their parent Gateway below.
	listenerSets := p.client.ListListenerSets()

	// parentAcceptedByGateway records the acceptance used to gate ListenerSet
	// programming, so that ListenerSet statuses are built from the same value.
	parentAcceptedByGateway := make(map[ktypes.NamespacedName]bool)

	for _, gateway := range gateways {
		logger := log.Ctx(ctx).With().
			Str("gateway", gateway.Name).
			Str("namespace", gateway.Namespace).
			Logger()

		gwListeners, allocatedListeners, allocatedPortProtocols := p.loadGatewayListeners(logger.WithContext(ctx), gateway, conf)

		// GEP-1713 forbids programming the listeners of a ListenerSet whose parent Gateway
		// is not accepted, so acceptance is computed before merging ListenerSet listeners.
		parentAccepted := isGatewayAccepted(gateway, gwListeners)

		var lsInfoMap map[ktypes.NamespacedName]*listenerSetInfo
		gwListeners, lsInfoMap = p.loadListenerSetListeners(logger.WithContext(ctx), gateway, listenerSets, parentAccepted, gwListeners, allocatedListeners, allocatedPortProtocols, conf)
		gwNSN := ktypes.NamespacedName{Name: gateway.Name, Namespace: gateway.Namespace}
		listenerSetInfosByGateway[gwNSN] = lsInfoMap
		parentAcceptedByGateway[gwNSN] = parentAccepted

		gwListenerSets := make([]ktypes.NamespacedName, 0, len(lsInfoMap))
		for nsn := range lsInfoMap {
			gwListenerSets = append(gwListenerSets, nsn)
		}

		selectedGateways = append(selectedGateways, gatewayWithListeners{
			Name:         gateway.Name,
			Namespace:    gateway.Namespace,
			listeners:    gwListeners,
			listenerSets: gwListenerSets,
		})
	}

	statusReport.gatewayListeners = selectedGateways

	p.loadHTTPRoutes(ctx, selectedGateways, conf, statusReport)

	p.loadGRPCRoutes(ctx, selectedGateways, conf, statusReport)

	p.loadTLSRoutes(ctx, selectedGateways, conf, statusReport)

	p.loadTCPRoutes(ctx, selectedGateways, conf, statusReport)

	for _, gateway := range gateways {
		logger := log.Ctx(ctx).With().
			Str("gateway", gateway.Name).
			Str("namespace", gateway.Namespace).
			Logger()

		gwNSN := ktypes.NamespacedName{Name: gateway.Name, Namespace: gateway.Namespace}

		// Collect this gateway's listeners (both Gateway- and ListenerSet-sourced).
		var listeners []gatewayListener
		for _, selectedGateway := range selectedGateways {
			if selectedGateway.Name == gateway.Name && selectedGateway.Namespace == gateway.Namespace {
				listeners = append(listeners, selectedGateway.listeners...)
			}
		}

		// Only Gateway-sourced listeners are passed to makeGatewayStatus.
		var gwOnlyListeners []gatewayListener
		for _, listener := range listeners {
			if listener.source() == kindGateway {
				gwOnlyListeners = append(gwOnlyListeners, listener)
			}
		}

		gatewayStatus, errConditions := p.makeGatewayStatus(gateway, gwOnlyListeners, addresses)
		if len(errConditions) > 0 {
			messages := map[string]struct{}{}
			for _, condition := range errConditions {
				messages[condition.Message] = struct{}{}
			}
			var conditionsErr error
			for message := range messages {
				conditionsErr = errors.Join(conditionsErr, errors.New(message))
			}
			logger.Debug().
				Err(conditionsErr).
				Msg("Gateway Not Accepted")
		}

		var attachedListenerSets int32
		for nsn, info := range listenerSetInfosByGateway[gwNSN] {
			lsStatus, lsAccepted := makeListenerSetStatus(info, listeners, parentAcceptedByGateway[gwNSN])
			statusReport.RecordListenerSetStatus(nsn, lsStatus)
			if lsAccepted {
				attachedListenerSets++
			}
		}
		gatewayStatus.AttachedListenerSets = &attachedListenerSets

		statusReport.RecordGatewayStatus(gwNSN, gatewayStatus)
	}

	return conf, statusReport, nil
}

func (p *Provider) loadGatewayListeners(ctx context.Context, gateway *gatev1.Gateway, conf *dynamic.Configuration) ([]gatewayListener, map[string]struct{}, map[gatev1.PortNumber]gatev1.ProtocolType) {
	tlsCerts := make(map[string]*tls.CertAndStores)
	allocatedListeners := make(map[string]struct{})
	// allocatedPortProtocols records the protocol claimed on each port, so that the
	// ListenerSet listeners merged afterwards can detect protocol conflicts with the
	// Gateway listeners.
	allocatedPortProtocols := make(map[gatev1.PortNumber]gatev1.ProtocolType)
	gatewayListeners := make([]gatewayListener, len(gateway.Spec.Listeners))

	for i, listener := range gateway.Spec.Listeners {
		gatewayListeners[i] = gatewayListener{
			Name:            string(listener.Name),
			GWName:          gateway.Name,
			GWNamespace:     gateway.Namespace,
			Port:            listener.Port,
			Protocol:        listener.Protocol,
			TLS:             listener.TLS,
			Hostname:        listener.Hostname,
			Source:          kindGateway,
			SourceName:      gateway.Name,
			SourceNamespace: gateway.Namespace,
			Status: &gatev1.ListenerStatus{
				Name:           listener.Name,
				SupportedKinds: []gatev1.RouteGroupKind{},
				Conditions:     []metav1.Condition{},
			},
		}

		// The listener protocol is validated first, so that an unsupported protocol
		// is reported as such instead of being masked by the entryPoint lookup,
		// which cannot succeed for a protocol Traefik does not know about.
		supportedKinds, conditions := supportedRouteKinds(gateway.Generation, listener.Protocol)
		if len(conditions) > 0 {
			gatewayListeners[i].Status.Conditions = append(gatewayListeners[i].Status.Conditions, conditions...)
			continue
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

		allowedRoutes := ptr.Deref(listener.AllowedRoutes, gatev1.AllowedRoutes{Namespaces: &gatev1.RouteNamespaces{From: new(gatev1.NamespacesFromSame)}})
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
		if _, ok := allocatedPortProtocols[listener.Port]; !ok {
			allocatedPortProtocols[listener.Port] = listener.Protocol
		}

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
			if listener.TLS == nil {
				gatewayListeners[i].Status.Conditions = append(gatewayListeners[i].Status.Conditions, metav1.Condition{
					Type:               string(gatev1.ListenerConditionAccepted),
					Status:             metav1.ConditionFalse,
					ObservedGeneration: gateway.Generation,
					LastTransitionTime: metav1.Now(),
					Reason:             "InvalidTLSConfiguration", // TODO check the spec if a proper reason is introduced at some point
					Message:            fmt.Sprintf("No TLS configuration for Gateway Listener %s:%d and protocol %q", listener.Name, listener.Port, listener.Protocol),
				})
				continue
			}

			tlsMode := ptr.Deref(listener.TLS.Mode, gatev1.TLSModeTerminate)
			isTLSPassthrough := tlsMode == gatev1.TLSModePassthrough

			if isTLSPassthrough && len(listener.TLS.CertificateRefs) > 0 {
				log.Ctx(ctx).Warn().Msg("In case of Passthrough TLS mode, no TLS settings take effect as the TLS session from the client is NOT terminated at the Gateway")
			}

			// Allowed configurations:
			// Protocol TLS -> Passthrough -> TLSRoute
			// Protocol TLS -> Terminate -> TLSRoute
			// Protocol HTTPS -> Terminate -> HTTPRoute
			if isTLSPassthrough && listener.Protocol == gatev1.HTTPSProtocolType {
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

				listenerTLSCerts, errCertConditions := p.resolveCertificateRefs(kindGateway, gateway.Namespace, listener.TLS.CertificateRefs, gateway.Generation)
				if len(errCertConditions) > 0 {
					gatewayListeners[i].Status.Conditions = append(gatewayListeners[i].Status.Conditions, errCertConditions...)
					gatewayListeners[i].Status.Conditions = append(gatewayListeners[i].Status.Conditions, metav1.Condition{
						Type:               string(gatev1.ListenerConditionProgrammed),
						Status:             metav1.ConditionFalse,
						ObservedGeneration: gateway.Generation,
						LastTransitionTime: metav1.Now(),
						Reason:             string(gatev1.ListenerReasonInvalid),
						Message:            "Invalid CertificateRefs",
					})
					continue
				}

				// Only copy if the certificate TLS config is not already known.
				for key, listenerTLSCert := range listenerTLSCerts {
					if _, ok := tlsCerts[key]; !ok {
						tlsCerts[key] = listenerTLSCert
					}
				}
			}
		}

		gatewayListeners[i].Attached = true
	}

	if len(tlsCerts) > 0 {
		conf.TLS.Certificates = append(conf.TLS.Certificates, getTLSConfig(tlsCerts)...)
	}

	return gatewayListeners, allocatedListeners, allocatedPortProtocols
}

// resolveCertificateRefs resolves the TLS certificateRefs of a listener and returns the
// corresponding certificates keyed by "<namespace>/<name>", along with the ResolvedRefs
// error conditions. fromKind and fromNamespace identify the resource holding the refs
// (Gateway or ListenerSet) for the ReferenceGrant check and the default certificate namespace.
func (p *Provider) resolveCertificateRefs(fromKind, fromNamespace string, certificateRefs []gatev1.SecretObjectReference, generation int64) (map[string]*tls.CertAndStores, []metav1.Condition) {
	var errCertConditions []metav1.Condition
	tlsCerts := make(map[string]*tls.CertAndStores)
	for _, certificateRef := range certificateRefs {
		if certificateRef.Kind == nil || *certificateRef.Kind != kindSecret || certificateRef.Group == nil || (*certificateRef.Group != "" && *certificateRef.Group != groupCore) {
			errCertConditions = append(errCertConditions, metav1.Condition{
				Type:               string(gatev1.ListenerConditionResolvedRefs),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.ListenerReasonInvalidCertificateRef),
				Message:            fmt.Sprintf("Unsupported TLS CertificateRef group/kind: %s/%s", groupToString(certificateRef.Group), kindToString(certificateRef.Kind)),
			})
			continue
		}

		certificateNamespace := string(ptr.Deref(certificateRef.Namespace, gatev1.Namespace(fromNamespace)))
		if err := p.isReferenceGranted(fromKind, fromNamespace, groupCore, kindSecret, string(certificateRef.Name), certificateNamespace); err != nil {
			errCertConditions = append(errCertConditions, metav1.Condition{
				Type:               string(gatev1.ListenerConditionResolvedRefs),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.ListenerReasonRefNotPermitted),
				Message:            fmt.Sprintf("Cannot reference CertificateRef %s/%s: %s", certificateNamespace, certificateRef.Name, err),
			})
			continue
		}

		configKey := certificateNamespace + "/" + string(certificateRef.Name)
		if _, tlsExists := tlsCerts[configKey]; !tlsExists {
			tlsCert, err := p.getTLSCert(certificateRef.Name, certificateNamespace)
			if err != nil {
				errCertConditions = append(errCertConditions, metav1.Condition{
					Type:               string(gatev1.ListenerConditionResolvedRefs),
					Status:             metav1.ConditionFalse,
					ObservedGeneration: generation,
					LastTransitionTime: metav1.Now(),
					Reason:             string(gatev1.ListenerReasonInvalidCertificateRef),
					Message:            fmt.Sprintf("Cannot load CertificateRef %s/%s: %s", certificateNamespace, certificateRef.Name, err),
				})
				continue
			}
			tlsCerts[configKey] = tlsCert
		}
	}

	return tlsCerts, errCertConditions
}

// isGatewayAccepted mirrors the acceptance rules of makeGatewayStatus: a Gateway is not
// accepted when its infrastructure parametersRef is set, or when it declares listeners and
// none of them is valid. It allows deciding, before the Gateway status is built, whether
// ListenerSet listeners can be programmed.
func isGatewayAccepted(gateway *gatev1.Gateway, listeners []gatewayListener) bool {
	if gateway.Spec.Infrastructure != nil && gateway.Spec.Infrastructure.ParametersRef != nil {
		return false
	}

	var validListeners, invalidListeners int
	for _, listener := range listeners {
		if len(listener.Status.Conditions) == 0 {
			validListeners++
		} else {
			invalidListeners++
		}
	}

	return invalidListeners == 0 || validListeners > 0
}

func (p *Provider) makeGatewayStatus(gateway *gatev1.Gateway, listeners []gatewayListener, addresses []gatev1.GatewayStatusAddress) (gatev1.GatewayStatus, []metav1.Condition) {
	gatewayStatus := gatev1.GatewayStatus{Addresses: addresses}

	var acceptedListeners int
	var errorConditions []metav1.Condition
	for _, listener := range listeners {
		if len(listener.Status.Conditions) == 0 {
			acceptedListeners++

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

	// Traefik supports no infrastructure parameters, and the specification requires
	// a parametersRef that cannot be resolved to be reported instead of ignored.
	if gateway.Spec.Infrastructure != nil && gateway.Spec.Infrastructure.ParametersRef != nil {
		condition := metav1.Condition{
			Type:               string(gatev1.GatewayConditionAccepted),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: gateway.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1.GatewayReasonInvalidParameters),
			Message:            "Gateway infrastructure parametersRef is not supported",
		}
		gatewayStatus.Conditions = append(gatewayStatus.Conditions, condition)

		return gatewayStatus, append(errorConditions, condition)
	}

	if len(errorConditions) > 0 && acceptedListeners == 0 {
		gatewayStatus.Conditions = append(gatewayStatus.Conditions,
			// update "Accepted" status with "Accepted" reason
			metav1.Condition{
				Type:               string(gatev1.GatewayConditionAccepted),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: gateway.Generation,
				Reason:             string(gatev1.GatewayReasonListenersNotValid),
				Message:            "At least one Listener must be valid",
				LastTransitionTime: metav1.Now(),
			},
			// update "Programmed" status with "Programmed" reason
			metav1.Condition{
				Type:               string(gatev1.GatewayConditionProgrammed),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: gateway.Generation,
				Reason:             string(gatev1.GatewayReasonInvalid),
				Message:            "No Listener is valid",
				LastTransitionTime: metav1.Now(),
			},
		)

		return gatewayStatus, errorConditions
	}

	acceptedConditionReason := gatev1.GatewayReasonAccepted
	acceptedConditionMessage := "Gateway successfully scheduled"
	programmedConditionMessage := "Gateway successfully programmed"
	if len(errorConditions) > 0 {
		acceptedConditionReason = gatev1.GatewayReasonListenersNotValid
		acceptedConditionMessage = "Gateway successfully scheduled, but some Listeners are not valid"
		programmedConditionMessage = "Gateway successfully programmed, but some Listeners are not valid"
	}

	gatewayStatus.Conditions = append(gatewayStatus.Conditions,
		// update "Accepted" status with "Accepted" reason
		metav1.Condition{
			Type:               string(gatev1.GatewayConditionAccepted),
			Status:             metav1.ConditionTrue,
			ObservedGeneration: gateway.Generation,
			Reason:             string(acceptedConditionReason),
			Message:            acceptedConditionMessage,
			LastTransitionTime: metav1.Now(),
		},
		// update "Programmed" status with "Programmed" reason
		metav1.Condition{
			Type:               string(gatev1.GatewayConditionProgrammed),
			Status:             metav1.ConditionTrue,
			ObservedGeneration: gateway.Generation,
			Reason:             string(gatev1.GatewayReasonProgrammed),
			Message:            programmedConditionMessage,
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
			Type:  new(gatev1.IPAddressType),
			Value: p.StatusAddress.IP,
		}}, nil
	}

	if p.StatusAddress.Hostname != "" {
		return []gatev1.GatewayStatusAddress{{
			Type:  new(gatev1.HostnameAddressType),
			Value: p.StatusAddress.Hostname,
		}}, nil
	}

	svcRef := p.StatusAddress.Service
	if svcRef.Name != "" && svcRef.Namespace != "" {
		svc, err := p.client.GetService(svcRef.Namespace, svcRef.Name)
		if err != nil {
			return nil, fmt.Errorf("getting service: %w", err)
		}

		var addresses []gatev1.GatewayStatusAddress
		for _, addr := range svc.Status.LoadBalancer.Ingress {
			switch {
			case addr.IP != "":
				addresses = append(addresses, gatev1.GatewayStatusAddress{
					Type:  new(gatev1.IPAddressType),
					Value: addr.IP,
				})

			case addr.Hostname != "":
				addresses = append(addresses, gatev1.GatewayStatusAddress{
					Type:  new(gatev1.HostnameAddressType),
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

func (p *Provider) getTLSCert(secretName gatev1.ObjectName, namespace string) (*tls.CertAndStores, error) {
	secret, err := p.client.GetSecret(namespace, string(secretName))
	if err != nil {
		return nil, fmt.Errorf("getting secret: %w", err)
	}

	cert, key, err := getCertificateBlocks(secret, namespace, string(secretName))
	if err != nil {
		return nil, fmt.Errorf("getting certificate blocks: %w", err)
	}

	certAndStore := &tls.CertAndStores{
		Certificate: tls.Certificate{
			CertFile: types.FileOrContent(cert),
			KeyFile:  types.FileOrContent(key),
		},
	}
	if _, err := certAndStore.GetCertificate(); err != nil {
		return nil, fmt.Errorf("validating certificate: %w", err)
	}

	return certAndStore, nil
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

	service, err := p.client.GetService(namespace, string(ref.Name))
	if err != nil {
		return nil, corev1.ServicePort{}, fmt.Errorf("getting service: %w", err)
	}
	if service.Spec.Type == corev1.ServiceTypeExternalName {
		return nil, corev1.ServicePort{}, errors.New("type ExternalName is not supported for Kubernetes Service reference")
	}

	var svcPort *corev1.ServicePort
	for _, p := range service.Spec.Ports {
		if p.Port == *ref.Port {
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
			if p.Name != nil && svcPort.Name == *p.Name {
				port = ptr.Deref(p.Port, 0)
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

func supportedRouteKinds(gatewayGeneration int64, protocol gatev1.ProtocolType) ([]gatev1.RouteGroupKind, []metav1.Condition) {
	group := gatev1.Group(gatev1.GroupName)

	switch protocol {
	case gatev1.TCPProtocolType:
		return []gatev1.RouteGroupKind{{Kind: kindTCPRoute, Group: &group}}, nil

	case gatev1.HTTPProtocolType, gatev1.HTTPSProtocolType:
		return []gatev1.RouteGroupKind{
			{Kind: kindHTTPRoute, Group: &group},
			{Kind: kindGRPCRoute, Group: &group},
		}, nil

	case gatev1.TLSProtocolType:
		return []gatev1.RouteGroupKind{
			{Kind: kindTLSRoute, Group: &group},
		}, nil
	}

	return nil, []metav1.Condition{{
		Type:               string(gatev1.ListenerConditionAccepted),
		Status:             metav1.ConditionFalse,
		ObservedGeneration: gatewayGeneration,
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

	if !strings.HasSuffix(string(h2), trimmedH1) {
		return ""
	}

	// h1 is a wildcard that encompasses h2, so h2 is always
	// the more specific hostname (the correct intersection).
	return h2
}

func allowRoute(listener gatewayListener, routeNamespace, routeKind string) bool {
	if !slices.Contains(listener.AllowedRouteKinds, routeKind) {
		return false
	}

	return slices.ContainsFunc(listener.AllowedNamespaces, func(allowedNamespace string) bool {
		return allowedNamespace == corev1.NamespaceAll || allowedNamespace == routeNamespace
	})
}

// gatewayListenersForParentRef associates a route parentRef with the listeners of
// the Gateway it refers to, among the Gateways managed by this controller.
type gatewayListenersForParentRef struct {
	parentRef gatev1.ParentReference

	gatewayName      string

	listeners []gatewayListener
}

// matchingGatewayListenersForParentRef returns, for each parentRef referring to a
// Gateway or a ListenerSet managed by this controller, the corresponding listeners.
// A Gateway parentRef yields the Gateway-sourced listeners of that Gateway, while a
// ListenerSet parentRef yields the listeners sourced from the referenced ListenerSet.
// parentRefs that do not refer to one of our Gateways or ListenerSets are omitted.
func matchingGatewayListenersForParentRef(gateways []gatewayWithListeners, routeNamespace string, parentRefs []gatev1.ParentReference) []gatewayListenersForParentRef {
	var matches []gatewayListenersForParentRef

	for _, parentRef := range parentRefs {
		if ptr.Deref(parentRef.Group, gatev1.GroupName) != gatev1.GroupName {
			continue
		}

		refKind := string(ptr.Deref(parentRef.Kind, kindGateway))
		parentRefNamespace := string(ptr.Deref(parentRef.Namespace, gatev1.Namespace(routeNamespace)))

		// Source discrimination: a Gateway parentRef targets only Gateway-sourced
		// listeners, and a ListenerSet parentRef targets only the listeners of the
		// referenced ListenerSet.
		switch refKind {
		case kindGateway:
			for _, gateway := range gateways {
				if parentRefNamespace != gateway.Namespace || string(parentRef.Name) != gateway.Name {
					continue
				}

				var listeners []gatewayListener
				for _, listener := range gateway.listeners {
					if listener.source() == kindGateway {
						listeners = append(listeners, listener)
					}
				}

				// All the Gateway listeners are kept: the parentRef is associated to its
				// Gateway here, and whether each listener is actually targeted (SectionName,
				// Port) is decided when loading the route, so that ResolvedRefs is reported
				// even for parentRefs that match no listener.
				matches = append(matches, gatewayListenersForParentRef{
					parentRef:   parentRef,
					gatewayName: gateway.Name,
					listeners:   listeners,
				})
				break
			}

		case kindListenerSet:
			// A ListenerSet parentRef targets the listeners sourced from the referenced
			// ListenerSet, regardless of which Gateway it is attached to.
			var listeners []gatewayListener
			var gatewayName string
			for _, gateway := range gateways {
				for _, listener := range gateway.listeners {
					if listener.Source != kindListenerSet {
						continue
					}
					if listener.SourceNamespace != parentRefNamespace || string(parentRef.Name) != listener.SourceName {
						continue
					}

					listeners = append(listeners, listener)
					gatewayName = gateway.Name
				}
			}

			if len(listeners) == 0 {
				// No listener was loaded from this ListenerSet (rejected by AllowedListeners,
				// or without any valid entry): the parentRef is still reported in the route
				// status when the ListenerSet references a managed Gateway.
				lsNSN := ktypes.NamespacedName{Namespace: parentRefNamespace, Name: string(parentRef.Name)}
				gwIndex := slices.IndexFunc(gateways, func(gw gatewayWithListeners) bool {
					return slices.Contains(gw.listenerSets, lsNSN)
				})
				if gwIndex < 0 {
					continue
				}

				gatewayName = gateways[gwIndex].Name
			}

			matches = append(matches, gatewayListenersForParentRef{
				parentRef:   parentRef,
				gatewayName: gatewayName,
				listeners:   listeners,
			})
		}
	}

	return matches
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

func makeRouterName(kind, rule, namespace, name string, listener gatewayListener, ruleIndex int) string {
	label := provider.Normalize(fmt.Sprintf("%s-%s-%s-%s-ep-%s-%d", kind, namespace, name, listener.routeKeySegment(), listener.EPName, ruleIndex))

	h := sha256.New()

	components := []string{namespace, name, listener.GWNamespace, listener.GWName, listener.EPName, strconv.Itoa(ruleIndex)}
	if listener.source() == kindListenerSet {
		components = []string{namespace, name, kindListenerSet, listener.SourceNamespace, listener.SourceName, listener.EPName, strconv.Itoa(ruleIndex)}
	}

	for _, c := range components {
		// Length-prefixing to avoid ambiguity between distinct components with embedded delimiter.
		fmt.Fprintf(h, "%d:%s", len(c), c)
	}

	// As explained in https://pkg.go.dev/hash#Hash,
	// Write never returns an error.
	h.Write([]byte(rule))

	return fmt.Sprintf("%s-%.10x", label, h.Sum(nil))
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

func throttleEvents(ctx context.Context, throttleDuration time.Duration, pool *safe.Pool, eventsChan <-chan any) chan any {
	if throttleDuration == 0 {
		return nil
	}
	// Create a buffered channel to hold the pending event (if we're delaying processing the event due to throttling)
	eventsChanBuffered := make(chan any, 1)

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

// isCrossProviderNamespaceAllowed reports whether the given namespace is allowed to use cross-provider references.
func isCrossProviderNamespaceAllowed(allowList []string, namespace string) bool {
	if allowList == nil {
		return true
	}

	return slices.Contains(allowList, namespace)
}

// makeListenerKey joins protocol, hostname, and port of a listener into a string key.
func makeListenerKey(l gatev1.Listener) string {
	var hostname gatev1.Hostname
	if l.Hostname != nil {
		hostname = *l.Hostname
	}

	return fmt.Sprintf("%s|%s|%d", l.Protocol, hostname, l.Port)
}

// makeListenerKeyFromEntry builds a conflict-detection key from a ListenerEntry.
func makeListenerKeyFromEntry(e gatev1.ListenerEntry) string {
	var hostname gatev1.Hostname
	if e.Hostname != nil {
		hostname = *e.Hostname
	}
	return fmt.Sprintf("%s|%s|%d", e.Protocol, hostname, e.Port)
}

// makeListenerConflictConditions returns the conditions the Gateway API spec requires on a
// conflicted listener: Accepted=False, Programmed=False and Conflicted=True, all sharing
// the conflict reason.
func makeListenerConflictConditions(generation int64, reason gatev1.ListenerConditionReason, message string) []metav1.Condition {
	return []metav1.Condition{
		{
			Type:               string(gatev1.ListenerConditionAccepted),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(reason),
			Message:            message,
		},
		{
			Type:               string(gatev1.ListenerConditionProgrammed),
			Status:             metav1.ConditionFalse,
			ObservedGeneration: generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(reason),
			Message:            message,
		},
		{
			Type:               string(gatev1.ListenerConditionConflicted),
			Status:             metav1.ConditionTrue,
			ObservedGeneration: generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(reason),
			Message:            message,
		},
	}
}

type listenerSetInfo struct {
	listenerSet *gatev1.ListenerSet

	// allowed reports whether the ListenerSet passed the Gateway's AllowedListeners policy.
	allowed bool
}

// loadListenerSetListeners loads listeners from the given ListenerSets that reference the
// given gateway, merging them into the existing gatewayListeners slice. It returns the
// updated slice plus per-ListenerSet status information. When parentAccepted is false the
// listeners are still validated and reported, but never programmed (GEP-1713).
func (p *Provider) loadListenerSetListeners(ctx context.Context, gateway *gatev1.Gateway, listenerSets []*gatev1.ListenerSet, parentAccepted bool, existingListeners []gatewayListener, allocatedListeners map[string]struct{}, allocatedPortProtocols map[gatev1.PortNumber]gatev1.ProtocolType, conf *dynamic.Configuration) ([]gatewayListener, map[ktypes.NamespacedName]*listenerSetInfo) {
	// Filter ListenerSets that reference this gateway and are in allowed namespaces.
	var matching []*gatev1.ListenerSet
	var disallowed []*gatev1.ListenerSet
	for _, ls := range listenerSets {
		if !listenerSetRefsGateway(ls, gateway) {
			continue
		}
		if !p.isListenerSetAllowed(ctx, gateway, ls) {
			log.Ctx(ctx).Debug().
				Str("listenerset", ls.Name).
				Str("namespace", ls.Namespace).
				Msg("ListenerSet not allowed by Gateway's AllowedListeners")
			disallowed = append(disallowed, ls)
			continue
		}
		matching = append(matching, ls)
	}

	// Sort by creation timestamp, then by namespace/name, so that the oldest ListenerSet
	// wins when two of them conflict (listener precedence, per GEP-1713).
	slices.SortStableFunc(matching, func(a, b *gatev1.ListenerSet) int {
		return cmp.Or(
			a.CreationTimestamp.Time.Compare(b.CreationTimestamp.Time),
			strings.Compare(a.Namespace, b.Namespace),
			strings.Compare(a.Name, b.Name),
		)
	})

	tlsConfigs := make(map[string]*tls.CertAndStores)
	infoMap := make(map[ktypes.NamespacedName]*listenerSetInfo, len(matching))

	for _, ls := range matching {
		nsn := ktypes.NamespacedName{Name: ls.Name, Namespace: ls.Namespace}
		info := &listenerSetInfo{listenerSet: ls, allowed: true}
		infoMap[nsn] = info

		for _, entry := range ls.Spec.Listeners {
			gl := gatewayListener{
				Name:            string(entry.Name),
				GWName:          gateway.Name,
				GWNamespace:     gateway.Namespace,
				Port:            entry.Port,
				Protocol:        entry.Protocol,
				TLS:             entry.TLS,
				Hostname:        entry.Hostname,
				Source:          kindListenerSet,
				SourceName:      ls.Name,
				SourceNamespace: ls.Namespace,
				Status: &gatev1.ListenerStatus{
					Name:           entry.Name,
					SupportedKinds: []gatev1.RouteGroupKind{},
					Conditions:     []metav1.Condition{},
				},
			}

			// The listener protocol is validated first, so that an unsupported protocol
			// is reported as such instead of being masked by the entryPoint lookup,
			// which cannot succeed for a protocol Traefik does not know about.
			supportedKinds, conditions := supportedRouteKinds(ls.Generation, entry.Protocol)
			if len(conditions) > 0 {
				gl.Status.Conditions = append(gl.Status.Conditions, conditions...)
				existingListeners = append(existingListeners, gl)
				continue
			}

			ep, err := p.entryPointName(entry.Port, entry.Protocol)
			if err != nil {
				gl.Status.Conditions = append(gl.Status.Conditions, metav1.Condition{
					Type:               string(gatev1.ListenerConditionAccepted),
					Status:             metav1.ConditionFalse,
					ObservedGeneration: ls.Generation,
					LastTransitionTime: metav1.Now(),
					Reason:             string(gatev1.ListenerReasonPortUnavailable),
					Message:            fmt.Sprintf("Cannot find entryPoint for ListenerSet listener: %v", err),
				})
				existingListeners = append(existingListeners, gl)
				continue
			}
			gl.EPName = ep

			allowedRoutes := ptr.Deref(entry.AllowedRoutes, gatev1.AllowedRoutes{Namespaces: &gatev1.RouteNamespaces{From: ptr.To(gatev1.NamespacesFromSame)}})
			// For ListenerSet listeners, "Same" means the ListenerSet's own namespace,
			// not the parent Gateway's namespace.
			gl.AllowedNamespaces, err = p.allowedNamespaces(ls.Namespace, allowedRoutes.Namespaces)
			if err != nil {
				gl.Status.Conditions = append(gl.Status.Conditions, metav1.Condition{
					Type:               string(gatev1.ListenerConditionResolvedRefs),
					Status:             metav1.ConditionFalse,
					ObservedGeneration: ls.Generation,
					LastTransitionTime: metav1.Now(),
					Reason:             "InvalidRouteNamespacesSelector",
					Message:            fmt.Sprintf("Invalid route namespaces selector: %v", err),
				})
				existingListeners = append(existingListeners, gl)
				continue
			}

			routeKinds, conditions := allowedRouteKindsFromListenerEntry(ls, entry, supportedKinds)
			for _, kind := range routeKinds {
				gl.AllowedRouteKinds = append(gl.AllowedRouteKinds, string(kind.Kind))
			}
			gl.Status.SupportedKinds = routeKinds
			if len(conditions) > 0 {
				gl.Status.Conditions = append(gl.Status.Conditions, conditions...)
				existingListeners = append(existingListeners, gl)
				continue
			}

			if protocol, ok := allocatedPortProtocols[entry.Port]; ok && protocol != entry.Protocol {
				gl.Status.Conditions = append(gl.Status.Conditions, makeListenerConflictConditions(ls.Generation, gatev1.ListenerReasonProtocolConflict,
					"A listener with a different protocol already uses this port")...)
				existingListeners = append(existingListeners, gl)
				continue
			}

			listenerKey := makeListenerKeyFromEntry(entry)
			if _, ok := allocatedListeners[listenerKey]; ok {
				gl.Status.Conditions = append(gl.Status.Conditions, makeListenerConflictConditions(ls.Generation, gatev1.ListenerReasonHostnameConflict,
					"A listener with the same protocol, port and hostname already exists")...)
				existingListeners = append(existingListeners, gl)
				continue
			}
			allocatedPortProtocols[entry.Port] = entry.Protocol
			allocatedListeners[listenerKey] = struct{}{}

			if (entry.Protocol == gatev1.HTTPProtocolType || entry.Protocol == gatev1.TCPProtocolType) && entry.TLS != nil {
				gl.Status.Conditions = append(gl.Status.Conditions, metav1.Condition{
					Type:               string(gatev1.ListenerConditionAccepted),
					Status:             metav1.ConditionFalse,
					ObservedGeneration: ls.Generation,
					LastTransitionTime: metav1.Now(),
					Reason:             "InvalidTLSConfiguration",
					Message:            "TLS configuration must not be defined when using HTTP or TCP protocol",
				})
				existingListeners = append(existingListeners, gl)
				continue
			}

			// TLS handling
			if entry.Protocol == gatev1.HTTPSProtocolType || entry.Protocol == gatev1.TLSProtocolType {
				if entry.TLS == nil {
					gl.Status.Conditions = append(gl.Status.Conditions, metav1.Condition{
						Type:               string(gatev1.ListenerConditionAccepted),
						Status:             metav1.ConditionFalse,
						ObservedGeneration: ls.Generation,
						LastTransitionTime: metav1.Now(),
						Reason:             "InvalidTLSConfiguration",
						Message: fmt.Sprintf("No TLS configuration for ListenerSet listener %s:%d and protocol %q",
							entry.Name, entry.Port, entry.Protocol),
					})
					existingListeners = append(existingListeners, gl)
					continue
				}

				tlsMode := ptr.Deref(entry.TLS.Mode, gatev1.TLSModeTerminate)
				isTLSPassthrough := tlsMode == gatev1.TLSModePassthrough

				if isTLSPassthrough && len(entry.TLS.CertificateRefs) > 0 {
					log.Ctx(ctx).Warn().Msg("In case of Passthrough TLS mode, no TLS settings take effect as the TLS session from the client is NOT terminated at the Gateway")
				}

				if entry.Protocol == gatev1.HTTPSProtocolType && isTLSPassthrough {
					gl.Status.Conditions = append(gl.Status.Conditions, metav1.Condition{
						Type:               string(gatev1.ListenerConditionAccepted),
						Status:             metav1.ConditionFalse,
						ObservedGeneration: ls.Generation,
						LastTransitionTime: metav1.Now(),
						Reason:             string(gatev1.ListenerReasonUnsupportedProtocol),
						Message:            "HTTPS protocol is not supported with TLS mode Passthrough",
					})
					existingListeners = append(existingListeners, gl)
					continue
				}

				if !isTLSPassthrough {
					if len(entry.TLS.CertificateRefs) == 0 {
						gl.Status.Conditions = append(gl.Status.Conditions, metav1.Condition{
							Type:               string(gatev1.ListenerConditionResolvedRefs),
							Status:             metav1.ConditionFalse,
							ObservedGeneration: ls.Generation,
							LastTransitionTime: metav1.Now(),
							Reason:             string(gatev1.ListenerReasonInvalidCertificateRef),
							Message:            "One TLS CertificateRef is required in Terminate mode",
						})
						existingListeners = append(existingListeners, gl)
						continue
					}

					// ReferenceGrants for ListenerSet use kindListenerSet as the from-kind.
					entryTLSCerts, errCertConditions := p.resolveCertificateRefs(kindListenerSet, ls.Namespace, entry.TLS.CertificateRefs, ls.Generation)
					if len(errCertConditions) > 0 {
						gl.Status.Conditions = append(gl.Status.Conditions, errCertConditions...)
						gl.Status.Conditions = append(gl.Status.Conditions, metav1.Condition{
							Type:               string(gatev1.ListenerConditionProgrammed),
							Status:             metav1.ConditionFalse,
							ObservedGeneration: ls.Generation,
							LastTransitionTime: metav1.Now(),
							Reason:             string(gatev1.ListenerReasonInvalid),
							Message:            "Invalid CertificateRefs",
						})
						existingListeners = append(existingListeners, gl)
						continue
					}

					// Only copy if the certificate TLS config is not already known.
					for key, entryTLSCert := range entryTLSCerts {
						if _, ok := tlsConfigs[key]; !ok {
							tlsConfigs[key] = entryTLSCert
						}
					}
				}
			}

			// GEP-1713: the listeners of a ListenerSet are only programmed when the
			// parent Gateway is accepted.
			gl.Attached = parentAccepted
			existingListeners = append(existingListeners, gl)
		}
	}

	// The certificates of listeners that cannot be programmed (parent Gateway not
	// accepted) must not enter the TLS store, as no router references them.
	if parentAccepted && len(tlsConfigs) > 0 {
		conf.TLS.Certificates = append(conf.TLS.Certificates, getTLSConfig(tlsConfigs)...)
	}

	// Track disallowed ListenerSets so they receive a NotAllowed status condition.
	for _, ls := range disallowed {
		nsn := ktypes.NamespacedName{Name: ls.Name, Namespace: ls.Namespace}
		infoMap[nsn] = &listenerSetInfo{listenerSet: ls, allowed: false}
	}

	return existingListeners, infoMap
}

// listenerSetRefsGateway returns true if the ListenerSet's ParentRef references the given gateway.
func listenerSetRefsGateway(ls *gatev1.ListenerSet, gw *gatev1.Gateway) bool {
	ref := ls.Spec.ParentRef

	if ref.Group != nil && string(*ref.Group) != gatev1.GroupName {
		return false
	}
	if ref.Kind != nil && string(*ref.Kind) != kindGateway {
		return false
	}
	if string(ref.Name) != gw.Name {
		return false
	}

	refNS := ls.Namespace
	if ref.Namespace != nil {
		refNS = string(*ref.Namespace)
	}
	return refNS == gw.Namespace
}

// isListenerSetAllowed checks whether the gateway's AllowedListeners permits the given ListenerSet.
func (p *Provider) isListenerSetAllowed(ctx context.Context, gw *gatev1.Gateway, ls *gatev1.ListenerSet) bool {
	if gw.Spec.AllowedListeners == nil {
		return false
	}

	ns := gw.Spec.AllowedListeners.Namespaces
	if ns == nil || ns.From == nil {
		return false
	}

	switch *ns.From {
	case gatev1.NamespacesFromNone:
		return false
	case gatev1.NamespacesFromSame:
		return ls.Namespace == gw.Namespace
	case gatev1.NamespacesFromAll:
		return true
	case gatev1.NamespacesFromSelector:
		if ns.Selector == nil {
			return false
		}
		selector, err := metav1.LabelSelectorAsSelector(ns.Selector)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("Invalid AllowedListeners namespace selector")
			return false
		}
		namespaces, err := p.client.ListNamespaces(selector)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("Unable to list namespaces for AllowedListeners selector")
			return false
		}
		return slices.Contains(namespaces, ls.Namespace)
	}

	return false
}

// allowedRouteKindsFromListenerEntry mirrors allowedRouteKinds but operates on a ListenerEntry.
func allowedRouteKindsFromListenerEntry(ls *gatev1.ListenerSet, entry gatev1.ListenerEntry, supportedKinds []gatev1.RouteGroupKind) ([]gatev1.RouteGroupKind, []metav1.Condition) {
	if entry.AllowedRoutes == nil || len(entry.AllowedRoutes.Kinds) == 0 {
		return supportedKinds, nil
	}

	var conditions []metav1.Condition
	routeKinds := []gatev1.RouteGroupKind{}
	uniqRouteKinds := map[gatev1.Kind]struct{}{}
	for _, routeKind := range entry.AllowedRoutes.Kinds {
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
				ObservedGeneration: ls.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.ListenerReasonInvalidRouteKinds),
				Message:            fmt.Sprintf("Listener protocol %q does not support RouteGroupKind %s/%s", entry.Protocol, groupToString(routeKind.Group), routeKind.Kind),
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

func makeListenerSetStatus(info *listenerSetInfo, allListeners []gatewayListener, parentAccepted bool) (gatev1.ListenerSetStatus, bool) {
	ls := info.listenerSet

	status := gatev1.ListenerSetStatus{}

	// If the ListenerSet was rejected by the Gateway's AllowedListeners policy,
	// return a NotAllowed status immediately without processing listeners.
	if !info.allowed {
		status.Conditions = append(status.Conditions,
			metav1.Condition{
				Type:               string(gatev1.ListenerSetConditionAccepted),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: ls.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.ListenerSetReasonNotAllowed),
				Message:            "ListenerSet is not allowed by the Gateway's AllowedListeners policy",
			},
			metav1.Condition{
				Type:               string(gatev1.ListenerSetConditionProgrammed),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: ls.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.ListenerSetReasonNotAllowed),
				Message:            "ListenerSet is not allowed by the Gateway's AllowedListeners policy",
			},
		)
		return status, false
	}

	// Build per-listener entry statuses from the gatewayListeners that came from this ListenerSet.
	for _, gl := range allListeners {
		if gl.Source != kindListenerSet || gl.SourceName != ls.Name || gl.SourceNamespace != ls.Namespace {
			continue
		}

		entryStatus := gatev1.ListenerEntryStatus{
			Name:           gatev1.SectionName(gl.Name),
			SupportedKinds: gl.Status.SupportedKinds,
			AttachedRoutes: gl.Status.AttachedRoutes,
		}

		if len(gl.Status.Conditions) == 0 {
			programmedCondition := metav1.Condition{
				Type:               string(gatev1.ListenerEntryConditionProgrammed),
				Status:             metav1.ConditionTrue,
				ObservedGeneration: ls.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.ListenerEntryReasonProgrammed),
				Message:            "No error found",
			}
			if !parentAccepted {
				// A valid listener entry is not programmed while the parent Gateway is not accepted.
				programmedCondition.Status = metav1.ConditionFalse
				programmedCondition.Reason = string(gatev1.ListenerEntryReasonPending)
				programmedCondition.Message = "Parent Gateway is not accepted"
			}

			entryStatus.Conditions = append(entryStatus.Conditions,
				metav1.Condition{
					Type:               string(gatev1.ListenerEntryConditionAccepted),
					Status:             metav1.ConditionTrue,
					ObservedGeneration: ls.Generation,
					LastTransitionTime: metav1.Now(),
					Reason:             string(gatev1.ListenerEntryReasonAccepted),
					Message:            "No error found",
				},
				metav1.Condition{
					Type:               string(gatev1.ListenerEntryConditionResolvedRefs),
					Status:             metav1.ConditionTrue,
					ObservedGeneration: ls.Generation,
					LastTransitionTime: metav1.Now(),
					Reason:             string(gatev1.ListenerEntryReasonResolvedRefs),
					Message:            "No error found",
				},
				programmedCondition,
			)
		} else {
			// The conditions field is a listType=map keyed by type: the apiserver rejects
			// duplicate entries, and several validation steps can report the same type.
			entryStatus.Conditions = dedupeConditionsByType(gl.Status.Conditions)
		}

		status.Listeners = append(status.Listeners, entryStatus)
	}

	// Determine top-level ListenerSet conditions from the per-listener validity.
	var validListeners int
	for _, entryStatus := range status.Listeners {
		valid := true
		for _, cond := range entryStatus.Conditions {
			if cond.Status == metav1.ConditionFalse {
				valid = false
				break
			}
		}
		if valid {
			validListeners++
		}
	}
	hasListenerErrors := validListeners != len(status.Listeners)

	if !parentAccepted {
		status.Conditions = append(status.Conditions,
			metav1.Condition{
				Type:               string(gatev1.ListenerSetConditionAccepted),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: ls.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.ListenerSetReasonParentNotAccepted),
				Message:            "Parent Gateway is not accepted",
			},
			metav1.Condition{
				Type:               string(gatev1.ListenerSetConditionProgrammed),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: ls.Generation,
				LastTransitionTime: metav1.Now(),
				// The Programmed condition documents this reason, but v1.6.1 does not define a constant for it.
				Reason:             "ParentNotProgrammed",
				Message:            "Parent Gateway is not accepted",
			},
		)
		return status, false
	}

	switch {
	case hasListenerErrors && validListeners == 0:
		// A ListenerSet with no valid listener at all is neither accepted nor programmed, per the spec.
		status.Conditions = append(status.Conditions,
			metav1.Condition{
				Type:               string(gatev1.ListenerSetConditionAccepted),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: ls.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.ListenerSetReasonListenersNotValid),
				Message:            "No valid listener",
			},
			metav1.Condition{
				Type:               string(gatev1.ListenerSetConditionProgrammed),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: ls.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.ListenerSetReasonListenersNotValid),
				Message:            "No valid listener",
			},
		)

		return status, false
	case hasListenerErrors:
		// The valid listeners are programmed, so the ListenerSet is accepted and programmed
		// even though some listeners have errors.
		status.Conditions = append(status.Conditions,
			metav1.Condition{
				Type:               string(gatev1.ListenerSetConditionAccepted),
				Status:             metav1.ConditionTrue,
				ObservedGeneration: ls.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.ListenerSetReasonListenersNotValid),
				Message:            "Some listeners have errors",
			},
			metav1.Condition{
				Type:               string(gatev1.ListenerSetConditionProgrammed),
				Status:             metav1.ConditionTrue,
				ObservedGeneration: ls.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.ListenerSetReasonProgrammed),
				Message:            "Valid listeners programmed",
			},
		)
	default:
		status.Conditions = append(status.Conditions,
			metav1.Condition{
				Type:               string(gatev1.ListenerSetConditionAccepted),
				Status:             metav1.ConditionTrue,
				ObservedGeneration: ls.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.ListenerSetReasonAccepted),
				Message:            "ListenerSet accepted",
			},
			metav1.Condition{
				Type:               string(gatev1.ListenerSetConditionProgrammed),
				Status:             metav1.ConditionTrue,
				ObservedGeneration: ls.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.ListenerSetReasonProgrammed),
				Message:            "ListenerSet programmed",
			},
		)
	}

	return status, true
}

// dedupeConditionsByType keeps the first condition of each type, as the status
// conditions field is a listType=map keyed by type and the apiserver rejects
// duplicate entries.
func dedupeConditionsByType(conditions []metav1.Condition) []metav1.Condition {
	seen := make(map[string]struct{}, len(conditions))
	deduped := make([]metav1.Condition, 0, len(conditions))
	for _, condition := range conditions {
		if _, ok := seen[condition.Type]; ok {
			continue
		}
		seen[condition.Type] = struct{}{}
		deduped = append(deduped, condition)
	}
	return deduped
}

func filterReferenceGrantsFrom(referenceGrants []*gatev1.ReferenceGrant, group, kind, namespace string) []*gatev1.ReferenceGrant {
	var matchingReferenceGrants []*gatev1.ReferenceGrant
	for _, referenceGrant := range referenceGrants {
		if referenceGrantMatchesFrom(referenceGrant, group, kind, namespace) {
			matchingReferenceGrants = append(matchingReferenceGrants, referenceGrant)
		}
	}
	return matchingReferenceGrants
}

func referenceGrantMatchesFrom(referenceGrant *gatev1.ReferenceGrant, group, kind, namespace string) bool {
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

func filterReferenceGrantsTo(referenceGrants []*gatev1.ReferenceGrant, group, kind, name string) []*gatev1.ReferenceGrant {
	var matchingReferenceGrants []*gatev1.ReferenceGrant
	for _, referenceGrant := range referenceGrants {
		if referenceGrantMatchesTo(referenceGrant, group, kind, name) {
			matchingReferenceGrants = append(matchingReferenceGrants, referenceGrant)
		}
	}
	return matchingReferenceGrants
}

func referenceGrantMatchesTo(referenceGrant *gatev1.ReferenceGrant, group, kind, name string) bool {
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
