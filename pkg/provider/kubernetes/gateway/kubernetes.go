package gateway

import (
	"cmp"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"maps"
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

	// routeReasonHostnameConflict is raised when an HTTP and a GRPC route are
	// attached to the same listener with intersecting hostnames.
	routeReasonHostnameConflict gatev1.RouteConditionReason = "HostnameConflict"

	conditionNoErrorMessage = "No error found"
)

// NamespacedName holds a Kubernetes resource reference with namespace and name.
type NamespacedName struct {
	Namespace string `description:"Defines the resource namespace." json:"namespace,omitempty" toml:"namespace,omitempty" yaml:"namespace,omitempty"`
	Name      string `description:"Defines the resource name." json:"name,omitempty" toml:"name,omitempty" yaml:"name,omitempty"`
}

// Provider holds configurations of the provider.
type Provider struct {
	Endpoint                string              `description:"Kubernetes server endpoint (required for external cluster client)." json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Token                   types.FileOrContent `description:"Kubernetes bearer token (not needed for in-cluster client). It accepts either a token value or a file path to the token." json:"token,omitempty" toml:"token,omitempty" yaml:"token,omitempty" loggable:"false"`
	QPS                     int                 `description:"Defines the maximum QPS to the Kubernetes API server. Setting this to a negative value will disable client-side ratelimiting." json:"qps,omitempty" toml:"qps,omitempty" yaml:"qps,omitempty" export:"true"`
	Burst                   int                 `description:"Defines the maximum burst of requests to the Kubernetes API server." json:"burst,omitempty" toml:"burst,omitempty" yaml:"burst,omitempty" export:"true"`
	CertAuthFilePath        string              `description:"Kubernetes certificate authority file path (not needed for in-cluster client)." json:"certAuthFilePath,omitempty" toml:"certAuthFilePath,omitempty" yaml:"certAuthFilePath,omitempty"`
	Namespaces              []string            `description:"Kubernetes namespaces." json:"namespaces,omitempty" toml:"namespaces,omitempty" yaml:"namespaces,omitempty" export:"true"`
	LabelSelector           string              `description:"Kubernetes label selector to select specific GatewayClasses." json:"labelSelector,omitempty" toml:"labelSelector,omitempty" yaml:"labelSelector,omitempty" export:"true"`
	Gateways                []NamespacedName    `description:"Scopes the provider to specific Gateways." json:"gateways,omitempty" toml:"gateways,omitempty" yaml:"gateways,omitempty" export:"true"`
	ThrottleDuration        ptypes.Duration     `description:"Kubernetes refresh throttle duration" json:"throttleDuration,omitempty" toml:"throttleDuration,omitempty" yaml:"throttleDuration,omitempty" export:"true"`
	ExperimentalChannel     bool                `description:"Toggles Experimental Channel resources support. Requires the Experimental Channel CRDs." json:"experimentalChannel,omitempty" toml:"experimentalChannel,omitempty" yaml:"experimentalChannel,omitempty" export:"true"`
	StatusAddress           *StatusAddress      `description:"Defines the Kubernetes Gateway status address." json:"statusAddress,omitempty" toml:"statusAddress,omitempty" yaml:"statusAddress,omitempty" export:"true"`
	NativeLBByDefault       bool                `description:"Defines whether to use Native Kubernetes load-balancing by default." json:"nativeLBByDefault,omitempty" toml:"nativeLBByDefault,omitempty" yaml:"nativeLBByDefault,omitempty" export:"true"`
	CrossProviderNamespaces []string            `description:"List of namespaces from which Gateway API routes are allowed to declare TraefikService backendRef references." json:"crossProviderNamespaces,omitempty" toml:"crossProviderNamespaces,omitempty" yaml:"crossProviderNamespaces,omitempty" export:"true"`

	EntryPoints map[string]Entrypoint `json:"-" toml:"-" yaml:"-" label:"-" file:"-"`

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
	Service  ServiceRef `description:"Published Kubernetes Service to copy status addresses from." json:"service" toml:"service,omitempty" yaml:"service,omitempty"`
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

// listenerOwner identifies the resource declaring a listener, a Gateway or a ListenerSet.
// The owner drives the ReferenceGrant checks, the namespace "Same" resolves to in allowedRoutes,
// and the status the listener is reported in.
type listenerOwner struct {
	Kind      string
	Namespace string
	Name      string
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

	EPName string

	// RouterNames holds one parent router per entry point hostname
	// the listener is the most specific match for.
	RouterNames []string

	// Gateway is the Gateway serving this listener:
	// the Owner itself for a Gateway listener, and the parent Gateway for a ListenerSet one.
	Gateway ktypes.NamespacedName

	Owner listenerOwner
}

// fromListenerSet reports whether the listener is declared by a ListenerSet rather than by the Gateway itself.
func (l gatewayListener) fromListenerSet() bool {
	return l.Owner.Kind == kindListenerSet
}

type gatewayWithListeners struct {
	Name      string
	Namespace string

	listeners []gatewayListener

	// listenerSets are the ListenerSets referencing this Gateway, allowed by its AllowedListeners policy or not.
	// They drive route status reporting for ListenerSet parentRefs that resolve to no listener.
	listenerSets map[ktypes.NamespacedName]*listenerSetInfo

	// accepted reports whether the Gateway is accepted, counting the listeners of its ListenerSets.
	accepted bool
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

// managesGateway reports whether the provider should reconcile the given Gateways.
// When the Gateways scoping option is set, only the referenced namespace/name entries are managed.
func (p *Provider) managesGateway(gateway *gatev1.Gateway) bool {
	if len(p.Gateways) == 0 {
		return true
	}

	for _, g := range p.Gateways {
		if gateway.Namespace == g.Namespace && gateway.Name == g.Name {
			return true
		}
	}

	return false
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
		TLS: &dynamic.TLSConfiguration{
			Options: map[string]tls.Options{},
		},
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
		if !p.managesGateway(gateway) {
			continue
		}
		gateways = append(gateways, gateway)
	}

	slices.SortStableFunc(gateways, func(a, b *gatev1.Gateway) int {
		return cmp.Or(a.GetCreationTimestamp().Time.Compare(b.GetCreationTimestamp().Time),
			strings.Compare(a.GetNamespace(), b.GetNamespace()),
			strings.Compare(a.GetName(), b.GetName()))
	})

	// ListenerSets are listed once and dispatched to their parent Gateway below,
	// oldest first, as the oldest ListenerSet wins a listener conflict with its siblings (GEP-1713).
	listenerSets := p.client.ListListenerSets()
	slices.SortStableFunc(listenerSets, func(a, b *gatev1.ListenerSet) int {
		return cmp.Or(a.GetCreationTimestamp().Time.Compare(b.GetCreationTimestamp().Time),
			strings.Compare(a.GetNamespace(), b.GetNamespace()),
			strings.Compare(a.GetName(), b.GetName()))
	})

	// selectedGateways is built in the order of gateways, so that both slices share indexes.
	selectedGateways := make([]gatewayWithListeners, 0, len(gateways))
	for _, gateway := range gateways {
		logger := log.Ctx(ctx).With().
			Str("gateway", gateway.Name).
			Str("namespace", gateway.Namespace).
			Logger()

		gwNSN := ktypes.NamespacedName{Namespace: gateway.Namespace, Name: gateway.Name}
		owner := listenerOwner{Kind: kindGateway, Namespace: gateway.Namespace, Name: gateway.Name}
		allocatedListeners := make(map[string]struct{})

		listeners := p.loadGatewayListeners(logger.WithContext(ctx), gwNSN, owner, gateway.Generation, gateway.Spec.Listeners, allocatedListeners, conf)

		listenerSetListeners, listenerSetInfos := p.loadListenerSetListeners(logger.WithContext(ctx), gateway, listenerSets, allocatedListeners, conf)
		listeners = append(listeners, listenerSetListeners...)

		// A Gateway is accepted as soon as one of the listeners serving it is valid, whoever declares it.
		// GEP-1713 forbids programming the listeners of a ListenerSet whose parent Gateway is not accepted,
		// which holds on its own here:
		// only a valid listener is programmed, and a valid one makes the Gateway accepted.
		accepted := len(listeners) == 0 || slices.ContainsFunc(listeners, func(listener gatewayListener) bool {
			return len(listener.Status.Conditions) == 0
		})

		selectedGateways = append(selectedGateways, gatewayWithListeners{
			Name:         gateway.Name,
			Namespace:    gateway.Namespace,
			listeners:    listeners,
			listenerSets: listenerSetInfos,
			accepted:     accepted,
		})
	}

	statusReport.gatewayListeners = selectedGateways

	// The isolation of a listener depends on the other listeners of its entry point.
	listenerRouters := p.buildListenerRouters(selectedGateways, conf)

	p.loadHTTPAndGRPCRoutes(ctx, selectedGateways, conf, statusReport)

	p.loadTLSRoutes(ctx, selectedGateways, conf, statusReport)

	p.loadTCPRoutes(ctx, selectedGateways, conf, statusReport)

	// A listener with no route attached gives a parent router with no child,
	// which the router manager reports in error as it has no service either.
	dropChildlessListenerRouters(conf, listenerRouters)

	for i, gateway := range gateways {
		logger := log.Ctx(ctx).With().
			Str("gateway", gateway.Name).
			Str("namespace", gateway.Namespace).
			Logger()

		selectedGateway := selectedGateways[i]

		// The Gateway status only reports the listeners the Gateway declares itself,
		// the ListenerSet ones being reported in their own ListenerSet status.
		var gatewayListeners []gatewayListener
		for _, listener := range selectedGateway.listeners {
			if !listener.fromListenerSet() {
				gatewayListeners = append(gatewayListeners, listener)
			}
		}

		gatewayStatus, errConditions := p.makeGatewayStatus(gateway, gatewayListeners, addresses, selectedGateway.accepted)
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
		for nsn, info := range selectedGateway.listenerSets {
			listenerSetStatus, listenerSetAccepted := makeListenerSetStatus(info, selectedGateway.listeners, selectedGateway.accepted)
			statusReport.RecordListenerSetStatus(nsn, listenerSetStatus)
			if listenerSetAccepted {
				attachedListenerSets++
			}
		}
		gatewayStatus.AttachedListenerSets = &attachedListenerSets

		statusReport.RecordGatewayStatus(ktypes.NamespacedName{Name: gateway.Name, Namespace: gateway.Namespace}, gatewayStatus)
	}

	return conf, statusReport, nil
}

// loadHTTPAndGRPCRoutes loads the HTTPRoutes and the GRPCRoutes together, ordered by
// creation timestamp then by "{namespace}/{name}".
// As stated in the specification, when an HTTPRoute and a GRPCRoute attached to the same
// listener have intersecting hostnames, only the first one in that order is accepted.
func (p *Provider) loadHTTPAndGRPCRoutes(ctx context.Context, gateways []gatewayWithListeners, conf *dynamic.Configuration, statusReport *statusReport) {
	routes := make([]metav1.Object, 0)

	httpRoutes, err := p.client.ListHTTPRoutes()
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Unable to list HTTPRoutes")
	}
	for _, route := range httpRoutes {
		routes = append(routes, route)
	}

	grpcRoutes, err := p.client.ListGRPCRoutes()
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Unable to list GRPCRoutes")
	}
	for _, route := range grpcRoutes {
		routes = append(routes, route)
	}

	slices.SortStableFunc(routes, compareRoutes)

	attached := make(attachedRoutes)
	served := make(servedRules)
	for _, route := range routes {
		switch route := route.(type) {
		case *gatev1.HTTPRoute:
			p.loadHTTPRoute(ctx, gateways, route, conf, attached, served, statusReport)
		case *gatev1.GRPCRoute:
			p.loadGRPCRoute(ctx, gateways, route, conf, attached, served, statusReport)
		}
	}
}

// loadGatewayListeners loads the given listeners, declared by owner for the given Gateway,
// and claims the valid ones in allocatedListeners.
func (p *Provider) loadGatewayListeners(ctx context.Context, gateway ktypes.NamespacedName, owner listenerOwner, generation int64, listeners []gatev1.Listener, allocatedListeners map[string]struct{}, conf *dynamic.Configuration) []gatewayListener {
	tlsCerts := make(map[string]*tls.CertAndStores)
	gatewayListeners := make([]gatewayListener, len(listeners))

	for i, listener := range listeners {
		gatewayListeners[i] = gatewayListener{
			Name:     string(listener.Name),
			Port:     listener.Port,
			Protocol: listener.Protocol,
			TLS:      listener.TLS,
			Hostname: listener.Hostname,
			Gateway:  gateway,
			Owner:    owner,
			Status: &gatev1.ListenerStatus{
				Name:           listener.Name,
				SupportedKinds: []gatev1.RouteGroupKind{},
				Conditions:     []metav1.Condition{},
			},
		}

		// The listener protocol is validated first, so that an unsupported protocol
		// is reported as such instead of being masked by the entryPoint lookup,
		// which cannot succeed for a protocol Traefik does not know about.
		supportedKinds, conditions := supportedRouteKinds(generation, listener.Protocol)
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
				ObservedGeneration: generation,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1.ListenerReasonPortUnavailable),
				Message:            fmt.Sprintf("Cannot find entryPoint for %s: %v", owner.Kind, err),
			})

			continue
		}
		gatewayListeners[i].EPName = ep

		allowedRoutes := ptr.Deref(listener.AllowedRoutes, gatev1.AllowedRoutes{Namespaces: &gatev1.RouteNamespaces{From: new(gatev1.NamespacesFromSame)}})
		gatewayListeners[i].AllowedNamespaces, err = p.allowedNamespaces(owner.Namespace, allowedRoutes.Namespaces)
		if err != nil {
			// update "ResolvedRefs" status true with "InvalidRoutesRef" reason
			gatewayListeners[i].Status.Conditions = append(gatewayListeners[i].Status.Conditions, metav1.Condition{
				Type:               string(gatev1.ListenerConditionResolvedRefs),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: generation,
				LastTransitionTime: metav1.Now(),
				Reason:             "InvalidRouteNamespacesSelector", // Should never happen as the selector is validated by kubernetes
				Message:            fmt.Sprintf("Invalid route namespaces selector: %v", err),
			})

			continue
		}

		routeKinds, conditions := allowedRouteKinds(generation, listener, supportedKinds)
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
			const message = "A listener with the same protocol, port and hostname already exists"
			gatewayListeners[i].Status.Conditions = append(gatewayListeners[i].Status.Conditions,
				metav1.Condition{
					Type:               string(gatev1.ListenerConditionAccepted),
					Status:             metav1.ConditionFalse,
					ObservedGeneration: generation,
					LastTransitionTime: metav1.Now(),
					Reason:             string(gatev1.ListenerReasonHostnameConflict),
					Message:            message,
				},
				metav1.Condition{
					Type:               string(gatev1.ListenerConditionProgrammed),
					Status:             metav1.ConditionFalse,
					ObservedGeneration: generation,
					LastTransitionTime: metav1.Now(),
					Reason:             string(gatev1.ListenerReasonHostnameConflict),
					Message:            message,
				},
				metav1.Condition{
					Type:               string(gatev1.ListenerConditionConflicted),
					Status:             metav1.ConditionTrue,
					ObservedGeneration: generation,
					LastTransitionTime: metav1.Now(),
					Reason:             string(gatev1.ListenerReasonHostnameConflict),
					Message:            message,
				},
			)

			continue
		}

		allocatedListeners[listenerKey] = struct{}{}

		if (listener.Protocol == gatev1.HTTPProtocolType || listener.Protocol == gatev1.TCPProtocolType) && listener.TLS != nil {
			gatewayListeners[i].Status.Conditions = append(gatewayListeners[i].Status.Conditions, metav1.Condition{
				Type:               string(gatev1.ListenerConditionAccepted),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: generation,
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
					ObservedGeneration: generation,
					LastTransitionTime: metav1.Now(),
					Reason:             "InvalidTLSConfiguration", // TODO check the spec if a proper reason is introduced at some point
					Message:            fmt.Sprintf("No TLS configuration for %s Listener %s:%d and protocol %q", owner.Kind, listener.Name, listener.Port, listener.Protocol),
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
					ObservedGeneration: generation,
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
						ObservedGeneration: generation,
						LastTransitionTime: metav1.Now(),
						Reason:             string(gatev1.ListenerReasonInvalidCertificateRef),
						Message:            "One TLS CertificateRef is required in Terminate mode",
					})
					continue
				}

				var errCertConditions []metav1.Condition
				listenerTLSCerts := make(map[string]*tls.CertAndStores)
				for _, certificateRef := range listener.TLS.CertificateRefs {
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

					certificateNamespace := string(ptr.Deref(certificateRef.Namespace, gatev1.Namespace(owner.Namespace)))
					if err := p.isReferenceGranted(owner.Kind, owner.Namespace, groupCore, kindSecret, string(certificateRef.Name), certificateNamespace); err != nil {
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
					if _, tlsExists := listenerTLSCerts[configKey]; !tlsExists {
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
						listenerTLSCerts[configKey] = tlsCert
					}
				}

				if len(errCertConditions) > 0 {
					gatewayListeners[i].Status.Conditions = append(gatewayListeners[i].Status.Conditions, errCertConditions...)
					gatewayListeners[i].Status.Conditions = append(gatewayListeners[i].Status.Conditions, metav1.Condition{
						Type:               string(gatev1.ListenerConditionProgrammed),
						Status:             metav1.ConditionFalse,
						ObservedGeneration: generation,
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

	return gatewayListeners
}

// uniqListener identifies a unique listener configuration.
// The protocol is part of it because a TLS parent router and a plain one are built on the two
// distinct handlers of an entry point, and because merging them would put the routes of an HTTP
// listener behind the TLS configuration of an HTTPS one.
//
// TODO: The Gateway frontend TLS configuration (client certificate validation) has to be part of it once supported.
// TODO: The listener TLS options (listener.TLS.Options) have to be part of it once supported.
type uniqListener struct {
	epName   string
	protocol gatev1.ProtocolType
}

// buildListenerRouters builds a parent router per entry point hostname, scoped to the requests it
// is the most specific match for. Electing the listener per Gateway isolates the listeners of a
// Gateway without hiding the routes of the other Gateways sharing the entry point.
func (p *Provider) buildListenerRouters(gateways []gatewayWithListeners, conf *dynamic.Configuration) []string {
	var listenerRouterNames []string

	hostnamesByListener := map[uniqListener][]string{}
	for _, gateway := range gateways {
		for _, listener := range gateway.listeners {
			if !listener.Attached ||
				listener.Protocol != gatev1.HTTPProtocolType && listener.Protocol != gatev1.HTTPSProtocolType {
				continue
			}

			uniq := uniqListener{epName: listener.EPName, protocol: listener.Protocol}

			hostname := string(ptr.Deref(listener.Hostname, ""))
			if !slices.Contains(hostnamesByListener[uniq], hostname) {
				hostnamesByListener[uniq] = append(hostnamesByListener[uniq], hostname)
			}
		}
	}

	uniqListeners := slices.SortedFunc(maps.Keys(hostnamesByListener), func(a, b uniqListener) int {
		return cmp.Or(cmp.Compare(a.epName, b.epName), cmp.Compare(a.protocol, b.protocol))
	})

	for _, uniq := range uniqListeners {
		hostnames := hostnamesByListener[uniq]
		slices.Sort(hostnames)

		for _, hostname := range hostnames {
			listenerRouterName := makeListenerRouterName(uniq, hostname)

			listenerRouter := &dynamic.Router{
				Rule:        buildListenerRule(hostname, hostnames),
				EntryPoints: []string{uniq.epName},
			}

			for _, gateway := range gateways {
				listener := mostSpecificListener(gateway.listeners, uniq, hostname)
				if listener == nil {
					continue
				}

				listener.RouterNames = append(listener.RouterNames, listenerRouterName)
			}

			if uniq.protocol == gatev1.HTTPSProtocolType {
				listenerTLSOptions := tls.Options{}
				listenerTLSOptions.SetDefaults()

				conf.TLS.Options[listenerRouterName] = listenerTLSOptions

				listenerRouter.TLS = &dynamic.RouterTLSConfig{
					Options: listenerRouterName,
				}
			}

			conf.HTTP.Routers[listenerRouterName] = listenerRouter
			listenerRouterNames = append(listenerRouterNames, listenerRouterName)
		}
	}

	return listenerRouterNames
}

// dropChildlessListenerRouters removes the parent routers no route is attached to.
func dropChildlessListenerRouters(conf *dynamic.Configuration, listenerRouterNames []string) {
	parents := map[string]struct{}{}
	for _, router := range conf.HTTP.Routers {
		for _, parent := range router.ParentRefs {
			parents[parent] = struct{}{}
		}
	}

	for _, name := range listenerRouterNames {
		if _, ok := parents[name]; ok {
			continue
		}

		delete(conf.HTTP.Routers, name)
		delete(conf.TLS.Options, name)
	}
}

func mostSpecificListener(listeners []gatewayListener, uniq uniqListener, hostname string) *gatewayListener {
	var elected *gatewayListener
	for i, listener := range listeners {
		if listener.EPName != uniq.epName || listener.Protocol != uniq.protocol || !listener.Attached {
			continue
		}

		listenerHostname := string(ptr.Deref(listener.Hostname, ""))
		if !hostnameCovers(listenerHostname, hostname) {
			continue
		}

		// A hostname covered by the elected one is a more specific match.
		if elected == nil || hostnameCovers(string(ptr.Deref(elected.Hostname, "")), listenerHostname) {
			elected = &listeners[i]
		}
	}

	return elected
}

// hostnameCovers reports whether every request matching hostname also matches listenerHostname.
func hostnameCovers(listenerHostname, hostname string) bool {
	if listenerHostname == "" {
		return true
	}

	return findMatchingHostname(gatev1.Hostname(listenerHostname), gatev1.Hostname(hostname)) != ""
}

func buildListenerRule(hostname string, entryPointHostnames []string) string {
	rule := `Host("*")`
	if hostname != "" {
		rule = fmt.Sprintf("Host(%q)", hostnameMatcherValue(hostname))
	}

	var exclusions []string
	for _, entryPointHostname := range entryPointHostnames {
		if entryPointHostname == hostname || !hostnameCovers(hostname, entryPointHostname) {
			continue
		}

		exclusions = append(exclusions, fmt.Sprintf("Host(%q)", hostnameMatcherValue(entryPointHostname)))
	}

	if len(exclusions) == 0 {
		return rule
	}

	return fmt.Sprintf("(%s) && !(%s)", rule, strings.Join(exclusions, " || "))
}

// hostnameMatcherValue returns the Host matcher value for a Gateway API hostname,
// whose wildcard spans one or more labels, unlike the Traefik single one.
func hostnameMatcherValue(hostname string) string {
	if suffix, ok := strings.CutPrefix(hostname, "*."); ok {
		return "**." + suffix
	}

	return hostname
}

// loadListenerSetListeners loads the listeners of the ListenerSets referencing the given Gateway.
func (p *Provider) loadListenerSetListeners(ctx context.Context, gateway *gatev1.Gateway, listenerSets []*gatev1.ListenerSet, allocatedListeners map[string]struct{}, conf *dynamic.Configuration) ([]gatewayListener, map[ktypes.NamespacedName]*listenerSetInfo) {
	infos := make(map[ktypes.NamespacedName]*listenerSetInfo)

	var allowed []*gatev1.ListenerSet
	for _, listenerSet := range listenerSets {
		if !listenerSetRefsGateway(listenerSet, gateway) {
			continue
		}

		info := &listenerSetInfo{listenerSet: listenerSet, allowed: p.isListenerSetAllowed(ctx, gateway, listenerSet)}
		infos[ktypes.NamespacedName{Namespace: listenerSet.Namespace, Name: listenerSet.Name}] = info

		if !info.allowed {
			log.Ctx(ctx).Debug().
				Str("listenerset", listenerSet.Name).
				Str("namespace", listenerSet.Namespace).
				Msg("ListenerSet not allowed by Gateway's AllowedListeners")
			continue
		}

		allowed = append(allowed, listenerSet)
	}

	gwNSN := ktypes.NamespacedName{Namespace: gateway.Namespace, Name: gateway.Name}

	var listeners []gatewayListener
	for _, listenerSet := range allowed {
		owner := listenerOwner{Kind: kindListenerSet, Namespace: listenerSet.Namespace, Name: listenerSet.Name}

		// A ListenerEntry mirrors a Gateway Listener field for field.
		entries := make([]gatev1.Listener, 0, len(listenerSet.Spec.Listeners))
		for _, entry := range listenerSet.Spec.Listeners {
			entries = append(entries, gatev1.Listener(entry))
		}

		listeners = append(listeners, p.loadGatewayListeners(ctx, gwNSN, owner, listenerSet.Generation, entries, allocatedListeners, conf)...)
	}

	return listeners, infos
}

func (p *Provider) makeGatewayStatus(gateway *gatev1.Gateway, listeners []gatewayListener, addresses []gatev1.GatewayStatusAddress, accepted bool) (gatev1.GatewayStatus, []metav1.Condition) {
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
					Message:            conditionNoErrorMessage,
				},
				metav1.Condition{
					Type:               string(gatev1.ListenerConditionResolvedRefs),
					Status:             metav1.ConditionTrue,
					ObservedGeneration: gateway.Generation,
					LastTransitionTime: metav1.Now(),
					Reason:             string(gatev1.ListenerReasonResolvedRefs),
					Message:            conditionNoErrorMessage,
				},
				metav1.Condition{
					Type:               string(gatev1.ListenerConditionProgrammed),
					Status:             metav1.ConditionTrue,
					ObservedGeneration: gateway.Generation,
					LastTransitionTime: metav1.Now(),
					Reason:             string(gatev1.ListenerReasonProgrammed),
					Message:            conditionNoErrorMessage,
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

	if !accepted {
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

func allowedRouteKinds(generation int64, listener gatev1.Listener, supportedKinds []gatev1.RouteGroupKind) ([]gatev1.RouteGroupKind, []metav1.Condition) {
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
				ObservedGeneration: generation,
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

	matches := map[gatev1.Hostname]struct{}{}
	for _, routeHostname := range routeHostnames {
		if match := findMatchingHostname(*listenerHostname, routeHostname); match != "" {
			matches[match] = struct{}{}
			continue
		}

		if match := findMatchingHostname(routeHostname, *listenerHostname); match != "" {
			matches[match] = struct{}{}
			continue
		}
	}

	return slices.Sorted(maps.Keys(matches)), len(matches) > 0
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

// compareRoutes orders the routes as the specification asks for the ones matching
// a request equally well to be discriminated.
func compareRoutes(a, b metav1.Object) int {
	return cmp.Or(a.GetCreationTimestamp().Time.Compare(b.GetCreationTimestamp().Time),
		strings.Compare(a.GetNamespace(), b.GetNamespace()),
		strings.Compare(a.GetName(), b.GetName()))
}

// servedRuleKey identifies a rule within the scope its router competes in.
// A TCP router competes in the muxer of its entry points,
// and an HTTP router in the muxer of its listener parent routers, as it has no entry points of its own.
type servedRuleKey struct {
	Scope string
	Rule  string
}

// servedRules keeps the router that serves each rule.
// Two routers that compete and have the same rule match the same requests.
// Thus only the first router that a muxer evaluates can serve a request.
type servedRules map[servedRuleKey]string

// register keeps the given router as the router that serves its rule within the given scope.
// If a different router serves that rule already, register gives its name, and true.
func (sr servedRules) register(name string, scope []string, rule string) (string, bool) {
	key := servedRuleKey{Scope: strings.Join(scope, ","), Rule: rule}

	if servedBy, served := sr[key]; served {
		return servedBy, true
	}

	sr[key] = name

	return "", false
}

// listenerRef identifies a listener of a Gateway.
type listenerRef struct {
	Name             string
	GatewayNamespace string
	GatewayName      string
}

// attachedRoutes holds, for each listener, the kind of the route already attached to a hostname.
// It allows detecting the hostname Conflicts between the HTTPRoutes and the GRPCRoutes attached to the same listener.
type attachedRoutes map[listenerRef]map[gatev1.Hostname]string

func (ar attachedRoutes) Record(ref listenerRef, kind string, hostnames []gatev1.Hostname) {
	if ar[ref] == nil {
		ar[ref] = make(map[gatev1.Hostname]string)
	}

	// A route without hostname matches all of them.
	if len(hostnames) == 0 {
		ar[ref][""] = kind
		return
	}

	for _, hostname := range hostnames {
		ar[ref][hostname] = kind
	}
}

// Conflicts returns whether one of the given hostnames is already attached to the listener by a route of another kind.
func (ar attachedRoutes) Conflicts(ref listenerRef, kind string, hostnames []gatev1.Hostname) bool {
	for attachedHostname, attachedKind := range ar[ref] {
		if attachedKind == kind {
			continue
		}

		// The empty hostname stands for all the hostnames,
		// so an empty hostname on either side means they intersect.
		if attachedHostname == "" || len(hostnames) == 0 {
			return true
		}

		for _, hostname := range hostnames {
			if findMatchingHostname(attachedHostname, hostname) != "" || findMatchingHostname(hostname, attachedHostname) != "" {
				return true
			}
		}
	}

	return false
}

// gatewayListenersForParentRef associates a route ParentRef with the Listeners of
// the Gateway it refers to, among the Gateways managed by this controller.
type gatewayListenersForParentRef struct {
	ParentRef gatev1.ParentReference

	GatewayName      string
	GatewayNamespace string

	Listeners []gatewayListener
}

// matchingGatewayListenersForParentRef returns, for each parentRef referring to a Gateway or a ListenerSet managed by this controller,
// the listeners this parent declares.
// parentRefs that do not refer to one of our Gateways or ListenerSets are omitted.
func matchingGatewayListenersForParentRef(gateways []gatewayWithListeners, routeNamespace string, parentRefs []gatev1.ParentReference) []gatewayListenersForParentRef {
	var matches []gatewayListenersForParentRef

	for _, parentRef := range parentRefs {
		if ptr.Deref(parentRef.Group, gatev1.GroupName) != gatev1.GroupName {
			continue
		}

		parent := listenerOwner{
			Kind:      string(ptr.Deref(parentRef.Kind, kindGateway)),
			Namespace: string(ptr.Deref(parentRef.Namespace, gatev1.Namespace(routeNamespace))),
			Name:      string(parentRef.Name),
		}
		if parent.Kind != kindGateway && parent.Kind != kindListenerSet {
			continue
		}

		gateway := gatewayForParent(gateways, parent)
		if gateway == nil {
			continue
		}

		// All the parent listeners are kept:
		// whether each listener is actually targeted (SectionName, Port) is decided when loading the route,
		// so that ResolvedRefs is reported even for parentRefs that match no listener.
		// A ListenerSet exposing none, rejected by AllowedListeners or without any valid entry, is still reported in the route status.
		var listeners []gatewayListener
		for _, listener := range gateway.listeners {
			// A Gateway parent targets the listeners the Gateway declares itself,
			// and a ListenerSet parent only the listeners of that ListenerSet.
			switch {
			case parent.Kind == kindGateway && !listener.fromListenerSet(),
				parent.Kind == kindListenerSet && listener.Owner == parent:
				listeners = append(listeners, listener)
			}
		}

		matches = append(matches, gatewayListenersForParentRef{
			ParentRef:        parentRef,
			GatewayName:      gateway.Name,
			GatewayNamespace: gateway.Namespace,
			Listeners:        listeners,
		})
	}

	return matches
}

// gatewayForParent returns the managed Gateway serving the given route parent, a Gateway or a ListenerSet,
// or nil when this controller does not manage that parent.
// A ListenerSet is served by the Gateway it references, allowed by its AllowedListeners policy or not.
func gatewayForParent(gateways []gatewayWithListeners, parent listenerOwner) *gatewayWithListeners {
	for i, gateway := range gateways {
		if parent.Kind == kindListenerSet {
			if _, ok := gateway.listenerSets[ktypes.NamespacedName{Namespace: parent.Namespace, Name: parent.Name}]; ok {
				return &gateways[i]
			}
			continue
		}

		if gateway.Namespace == parent.Namespace && gateway.Name == parent.Name {
			return &gateways[i]
		}
	}

	return nil
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

func makeRouterName(kind, rule, namespace, name, gatewayNamespace, gatewayName string, listener gatewayListener, ruleIndex int) string {
	label := provider.Normalize(fmt.Sprintf("%s-%s-%s-gw-%s-%s-ep-%s-%d", kind, namespace, name, gatewayNamespace, gatewayName, listener.EPName, ruleIndex))
	components := []string{namespace, name, gatewayNamespace, gatewayName, listener.EPName, strconv.Itoa(ruleIndex)}
	if listener.fromListenerSet() {
		// The routers attached through a ListenerSet are named after it,
		// and the kind keeps them apart from the ones attached to a Gateway of the same name.
		label = provider.Normalize(fmt.Sprintf("%s-%s-%s-ls-%s-%s-ep-%s-%d", kind, namespace, name, listener.Owner.Namespace, listener.Owner.Name, listener.EPName, ruleIndex))
		components = []string{namespace, name, kindListenerSet, listener.Owner.Namespace, listener.Owner.Name, listener.EPName, strconv.Itoa(ruleIndex)}
	}

	h := sha256.New()

	for _, c := range components {
		// Length-prefixing to avoid ambiguity between distinct components with embedded delimiter.
		fmt.Fprintf(h, "%d:%s", len(c), c)
	}

	// As explained in https://pkg.go.dev/hash#Hash,
	// Write never returns an error.
	h.Write([]byte(rule))

	return fmt.Sprintf("%s-%.10x", label, h.Sum(nil))
}

// makeListenerRouterName hashes the hostname, as provider.Normalize drops the characters
// telling two of them apart: the "*.example.com" and "example.com" hostnames of an entry
// point both normalize to the "listener-web-http-example-com" label.
func makeListenerRouterName(uniq uniqListener, hostname string) string {
	protocol := strings.ToLower(string(uniq.protocol))

	label := provider.Normalize(fmt.Sprintf("listener-%s-%s-%s", uniq.epName, protocol, hostname))

	h := sha256.New()

	for _, c := range []string{uniq.epName, protocol, hostname} {
		// Length-prefixing to avoid ambiguity between distinct components with embedded delimiter.
		fmt.Fprintf(h, "%d:%s", len(c), c)
	}

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

type listenerSetInfo struct {
	listenerSet *gatev1.ListenerSet

	// allowed reports whether the ListenerSet passed the Gateway's AllowedListeners policy.
	allowed bool
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

func makeListenerSetStatus(info *listenerSetInfo, listeners []gatewayListener, parentAccepted bool) (gatev1.ListenerSetStatus, bool) {
	listenerSet := info.listenerSet
	generation := listenerSet.Generation

	var status gatev1.ListenerSetStatus

	if !info.allowed {
		const message = "ListenerSet is not allowed by the Gateway's AllowedListeners policy"
		status.Conditions = []metav1.Condition{
			makeListenerSetCondition(gatev1.ListenerSetConditionAccepted, metav1.ConditionFalse, generation, string(gatev1.ListenerSetReasonNotAllowed), message),
			makeListenerSetCondition(gatev1.ListenerSetConditionProgrammed, metav1.ConditionFalse, generation, string(gatev1.ListenerSetReasonNotAllowed), message),
		}

		return status, false
	}

	var validListeners int
	for _, listener := range listeners {
		if listener.Owner != (listenerOwner{Kind: kindListenerSet, Namespace: listenerSet.Namespace, Name: listenerSet.Name}) {
			continue
		}

		// A ListenerEntryStatus mirrors a ListenerStatus field for field.
		entryStatus := gatev1.ListenerEntryStatus(*listener.Status)
		if len(entryStatus.Conditions) == 0 {
			validListeners++

			entryStatus.Conditions = []metav1.Condition{
				{
					Type:               string(gatev1.ListenerEntryConditionAccepted),
					Status:             metav1.ConditionTrue,
					ObservedGeneration: generation,
					LastTransitionTime: metav1.Now(),
					Reason:             string(gatev1.ListenerEntryReasonAccepted),
					Message:            conditionNoErrorMessage,
				},
				{
					Type:               string(gatev1.ListenerEntryConditionResolvedRefs),
					Status:             metav1.ConditionTrue,
					ObservedGeneration: generation,
					LastTransitionTime: metav1.Now(),
					Reason:             string(gatev1.ListenerEntryReasonResolvedRefs),
					Message:            conditionNoErrorMessage,
				},
				{
					Type:               string(gatev1.ListenerEntryConditionProgrammed),
					Status:             metav1.ConditionTrue,
					ObservedGeneration: generation,
					LastTransitionTime: metav1.Now(),
					Reason:             string(gatev1.ListenerEntryReasonProgrammed),
					Message:            conditionNoErrorMessage,
				},
			}
		}

		status.Listeners = append(status.Listeners, entryStatus)
	}

	switch {
	case !parentAccepted:
		const message = "Parent Gateway is not accepted"
		status.Conditions = []metav1.Condition{
			makeListenerSetCondition(gatev1.ListenerSetConditionAccepted, metav1.ConditionFalse, generation, string(gatev1.ListenerSetReasonParentNotAccepted), message),
			// The Programmed condition documents this reason, but v1.6.1 defines no constant for it.
			makeListenerSetCondition(gatev1.ListenerSetConditionProgrammed, metav1.ConditionFalse, generation, "ParentNotProgrammed", message),
		}

		return status, false

	case len(status.Listeners) > 0 && validListeners == 0:
		// A ListenerSet with no valid listener at all is neither accepted nor programmed, per the spec.
		const message = "No valid listener"
		status.Conditions = []metav1.Condition{
			makeListenerSetCondition(gatev1.ListenerSetConditionAccepted, metav1.ConditionFalse, generation, string(gatev1.ListenerSetReasonListenersNotValid), message),
			makeListenerSetCondition(gatev1.ListenerSetConditionProgrammed, metav1.ConditionFalse, generation, string(gatev1.ListenerSetReasonListenersNotValid), message),
		}

		return status, false

	case validListeners < len(status.Listeners):
		// The valid listeners are programmed, so the ListenerSet is accepted and programmed even though some listeners have errors.
		status.Conditions = []metav1.Condition{
			makeListenerSetCondition(gatev1.ListenerSetConditionAccepted, metav1.ConditionTrue, generation, string(gatev1.ListenerSetReasonListenersNotValid), "Some listeners have errors"),
			makeListenerSetCondition(gatev1.ListenerSetConditionProgrammed, metav1.ConditionTrue, generation, string(gatev1.ListenerSetReasonProgrammed), "Valid listeners programmed"),
		}

	default:
		status.Conditions = []metav1.Condition{
			makeListenerSetCondition(gatev1.ListenerSetConditionAccepted, metav1.ConditionTrue, generation, string(gatev1.ListenerSetReasonAccepted), "ListenerSet accepted"),
			makeListenerSetCondition(gatev1.ListenerSetConditionProgrammed, metav1.ConditionTrue, generation, string(gatev1.ListenerSetReasonProgrammed), "ListenerSet programmed"),
		}
	}

	return status, true
}

func makeListenerSetCondition(conditionType gatev1.ListenerSetConditionType, status metav1.ConditionStatus, generation int64, reason, message string) metav1.Condition {
	return metav1.Condition{
		Type:               string(conditionType),
		Status:             status,
		ObservedGeneration: generation,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	}
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
