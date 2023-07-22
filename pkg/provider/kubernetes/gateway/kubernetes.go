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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/utils/pointer"
	"k8s.io/utils/strings/slices"
	gatev1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

const (
	providerName = "kubernetesgateway"

	controllerName = "traefik.io/gateway-controller"

	kindGateway        = "Gateway"
	kindTraefikService = "TraefikService"
	kindHTTPRoute      = "HTTPRoute"
	kindTCPRoute       = "TCPRoute"
	kindTLSRoute       = "TLSRoute"
)

// timeNow is a mockable version of time.Now.
var timeNow = time.Now

// Provider holds configurations of the provider.
type Provider struct {
	Endpoint         string                `description:"Kubernetes server endpoint (required for external cluster client)." json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Token            string                `description:"Kubernetes bearer token (not needed for in-cluster client)." json:"token,omitempty" toml:"token,omitempty" yaml:"token,omitempty" loggable:"false"`
	CertAuthFilePath string                `description:"Kubernetes certificate authority file path (not needed for in-cluster client)." json:"certAuthFilePath,omitempty" toml:"certAuthFilePath,omitempty" yaml:"certAuthFilePath,omitempty"`
	Namespaces       []string              `description:"Kubernetes namespaces." json:"namespaces,omitempty" toml:"namespaces,omitempty" yaml:"namespaces,omitempty" export:"true"`
	LabelSelector    string                `description:"Kubernetes label selector to select specific GatewayClasses." json:"labelSelector,omitempty" toml:"labelSelector,omitempty" yaml:"labelSelector,omitempty" export:"true"`
	ThrottleDuration ptypes.Duration       `description:"Kubernetes refresh throttle duration" json:"throttleDuration,omitempty" toml:"throttleDuration,omitempty" yaml:"throttleDuration,omitempty" export:"true"`
	EntryPoints      map[string]Entrypoint `json:"-" toml:"-" yaml:"-" label:"-" file:"-"`

	lastConfiguration safe.Safe

	routerTransform k8s.RouterTransform
}

func (p *Provider) SetRouterTransform(routerTransform k8s.RouterTransform) {
	p.routerTransform = routerTransform
}

func (p *Provider) applyRouterTransform(ctx context.Context, rt *dynamic.Router, route *gatev1alpha2.HTTPRoute) {
	if p.routerTransform == nil {
		return
	}

	err := p.routerTransform.Apply(ctx, rt, route.Annotations)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Apply router transform")
	}
}

// Entrypoint defines the available entry points.
type Entrypoint struct {
	Address        string
	HasHTTPTLSConf bool
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
		client, err = newExternalClusterClient(p.Endpoint, p.Token, p.CertAuthFilePath)
	}

	if err != nil {
		return nil, err
	}

	client.labelSelector = p.LabelSelector

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
				Type:               string(gatev1alpha2.GatewayClassConditionStatusAccepted),
				Status:             metav1.ConditionTrue,
				Reason:             "Handled",
				Message:            "Handled by Traefik controller",
				LastTransitionTime: metav1.Now(),
			})
			if err != nil {
				logger.Error().Err(err).Msgf("Failed to update %s condition", gatev1alpha2.GatewayClassConditionStatusAccepted)
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

func (p *Provider) createGatewayConf(ctx context.Context, client Client, gateway *gatev1alpha2.Gateway) (*dynamic.Configuration, error) {
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
	listenerStatuses := p.fillGatewayConf(ctx, client, gateway, conf, tlsConfigs)

	gatewayStatus, err := p.makeGatewayStatus(listenerStatuses)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while creating gateway status: %w", err)
	}

	err = client.UpdateGatewayStatus(gateway, gatewayStatus)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while updating gateway status: %w", err)
	}

	if len(tlsConfigs) > 0 {
		conf.TLS.Certificates = append(conf.TLS.Certificates, getTLSConfig(tlsConfigs)...)
	}

	return conf, nil
}

func (p *Provider) fillGatewayConf(ctx context.Context, client Client, gateway *gatev1alpha2.Gateway, conf *dynamic.Configuration, tlsConfigs map[string]*tls.CertAndStores) []gatev1alpha2.ListenerStatus {
	logger := log.Ctx(ctx)
	listenerStatuses := make([]gatev1alpha2.ListenerStatus, len(gateway.Spec.Listeners))
	allocatedListeners := make(map[string]struct{})

	for i, listener := range gateway.Spec.Listeners {
		listenerStatuses[i] = gatev1alpha2.ListenerStatus{
			Name:           listener.Name,
			SupportedKinds: []gatev1alpha2.RouteGroupKind{},
			Conditions:     []metav1.Condition{},
			AttachedRoutes: 0,
		}

		supportedKinds, conditions := supportedRouteKinds(listener.Protocol)
		if len(conditions) > 0 {
			listenerStatuses[i].Conditions = append(listenerStatuses[i].Conditions, conditions...)
			continue
		}

		listenerStatuses[i].SupportedKinds = supportedKinds

		routeKinds, conditions := getAllowedRouteKinds(listener, supportedKinds)
		if len(conditions) > 0 {
			listenerStatuses[i].Conditions = append(listenerStatuses[i].Conditions, conditions...)
			continue
		}

		listenerKey := makeListenerKey(listener)

		if _, ok := allocatedListeners[listenerKey]; ok {
			listenerStatuses[i].Conditions = append(listenerStatuses[i].Conditions, metav1.Condition{
				Type:               string(gatev1alpha2.ListenerConditionConflicted),
				Status:             metav1.ConditionTrue,
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
				Type:               string(gatev1alpha2.ListenerConditionDetached),
				Status:             metav1.ConditionTrue,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1alpha2.ListenerReasonPortUnavailable),
				Message:            fmt.Sprintf("Cannot find entryPoint for Gateway: %v", err),
			})

			continue
		}

		if (listener.Protocol == gatev1alpha2.HTTPProtocolType || listener.Protocol == gatev1alpha2.TCPProtocolType) && listener.TLS != nil {
			listenerStatuses[i].Conditions = append(listenerStatuses[i].Conditions, metav1.Condition{
				Type:               string(gatev1alpha2.ListenerConditionDetached),
				Status:             metav1.ConditionTrue,
				LastTransitionTime: metav1.Now(),
				Reason:             "InvalidTLSConfiguration", // TODO check the spec if a proper reason is introduced at some point
				Message:            "TLS configuration must no be defined when using HTTP or TCP protocol",
			})

			continue
		}

		// TLS
		if listener.Protocol == gatev1alpha2.HTTPSProtocolType || listener.Protocol == gatev1alpha2.TLSProtocolType {
			if listener.TLS == nil || (len(listener.TLS.CertificateRefs) == 0 && listener.TLS.Mode != nil && *listener.TLS.Mode != gatev1alpha2.TLSModePassthrough) {
				// update "Detached" status with "UnsupportedProtocol" reason
				listenerStatuses[i].Conditions = append(listenerStatuses[i].Conditions, metav1.Condition{
					Type:               string(gatev1alpha2.ListenerConditionDetached),
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.Now(),
					Reason:             "InvalidTLSConfiguration", // TODO check the spec if a proper reason is introduced at some point
					Message: fmt.Sprintf("No TLS configuration for Gateway Listener %s:%d and protocol %q",
						listener.Name, listener.Port, listener.Protocol),
				})

				continue
			}

			var tlsModeType gatev1alpha2.TLSModeType
			if listener.TLS.Mode != nil {
				tlsModeType = *listener.TLS.Mode
			}

			isTLSPassthrough := tlsModeType == gatev1alpha2.TLSModePassthrough

			if isTLSPassthrough && len(listener.TLS.CertificateRefs) > 0 {
				// https://gateway-api.sigs.k8s.io/v1alpha2/references/spec/#gateway.networking.k8s.io/v1alpha2.GatewayTLSConfig
				logger.Warn().Msg("In case of Passthrough TLS mode, no TLS settings take effect as the TLS session from the client is NOT terminated at the Gateway")
			}

			// Allowed configurations:
			// Protocol TLS -> Passthrough -> TLSRoute/TCPRoute
			// Protocol TLS -> Terminate -> TLSRoute/TCPRoute
			// Protocol HTTPS -> Terminate -> HTTPRoute
			if listener.Protocol == gatev1alpha2.HTTPSProtocolType && isTLSPassthrough {
				listenerStatuses[i].Conditions = append(listenerStatuses[i].Conditions, metav1.Condition{
					Type:               string(gatev1alpha2.ListenerConditionDetached),
					Status:             metav1.ConditionTrue,
					LastTransitionTime: metav1.Now(),
					Reason:             string(gatev1alpha2.ListenerReasonUnsupportedProtocol),
					Message:            "HTTPS protocol is not supported with TLS mode Passthrough",
				})

				continue
			}

			if !isTLSPassthrough {
				if len(listener.TLS.CertificateRefs) == 0 {
					// update "ResolvedRefs" status true with "InvalidCertificateRef" reason
					listenerStatuses[i].Conditions = append(listenerStatuses[i].Conditions, metav1.Condition{
						Type:               string(gatev1alpha2.ListenerConditionResolvedRefs),
						Status:             metav1.ConditionFalse,
						LastTransitionTime: metav1.Now(),
						Reason:             string(gatev1alpha2.ListenerReasonInvalidCertificateRef),
						Message:            "One TLS CertificateRef is required in Terminate mode",
					})

					continue
				}

				// TODO Should we support multiple certificates?
				certificateRef := listener.TLS.CertificateRefs[0]

				if certificateRef.Kind == nil || *certificateRef.Kind != "Secret" ||
					certificateRef.Group == nil || (*certificateRef.Group != "" && *certificateRef.Group != "core") {
					// update "ResolvedRefs" status true with "InvalidCertificateRef" reason
					listenerStatuses[i].Conditions = append(listenerStatuses[i].Conditions, metav1.Condition{
						Type:               string(gatev1alpha2.ListenerConditionResolvedRefs),
						Status:             metav1.ConditionFalse,
						LastTransitionTime: metav1.Now(),
						Reason:             string(gatev1alpha2.ListenerReasonInvalidCertificateRef),
						Message:            fmt.Sprintf("Unsupported TLS CertificateRef group/kind: %v/%v", certificateRef.Group, certificateRef.Kind),
					})

					continue
				}

				// TODO Support ReferencePolicy to support cross namespace references.
				if certificateRef.Namespace != nil && string(*certificateRef.Namespace) != gateway.Namespace {
					listenerStatuses[i].Conditions = append(listenerStatuses[i].Conditions, metav1.Condition{
						Type:               string(gatev1alpha2.ListenerConditionResolvedRefs),
						Status:             metav1.ConditionFalse,
						LastTransitionTime: metav1.Now(),
						Reason:             string(gatev1alpha2.ListenerReasonInvalidCertificateRef),
						Message:            "Cross namespace secrets are not supported",
					})

					continue
				}

				configKey := gateway.Namespace + "/" + string(certificateRef.Name)
				if _, tlsExists := tlsConfigs[configKey]; !tlsExists {
					tlsConf, err := getTLS(client, certificateRef.Name, gateway.Namespace)
					if err != nil {
						// update "ResolvedRefs" status true with "InvalidCertificateRef" reason
						listenerStatuses[i].Conditions = append(listenerStatuses[i].Conditions, metav1.Condition{
							Type:               string(gatev1alpha2.ListenerConditionResolvedRefs),
							Status:             metav1.ConditionFalse,
							LastTransitionTime: metav1.Now(),
							Reason:             string(gatev1alpha2.ListenerReasonInvalidCertificateRef),
							Message:            fmt.Sprintf("Error while retrieving certificate: %v", err),
						})

						continue
					}

					tlsConfigs[configKey] = tlsConf
				}
			}
		}

		routesAttached := int32(0)
		for _, routeKind := range routeKinds {
			var (
				numRoutesAttached int32
				conditions        []metav1.Condition
			)

			switch routeKind.Kind {
			case kindHTTPRoute:
				numRoutesAttached, conditions = p.gatewayHTTPRouteToHTTPConf(ctx, ep, listener, gateway, client, conf)
			case kindTCPRoute:
				numRoutesAttached, conditions = gatewayTCPRouteToTCPConf(ctx, ep, listener, gateway, client, conf)
			case kindTLSRoute:
				numRoutesAttached, conditions = gatewayTLSRouteToTCPConf(ctx, ep, listener, gateway, client, conf)
			}

			routesAttached += numRoutesAttached
			listenerStatuses[i].Conditions = append(listenerStatuses[i].Conditions, conditions...)
		}
		listenerStatuses[i].AttachedRoutes = routesAttached
	}

	return listenerStatuses
}

func (p *Provider) makeGatewayStatus(listenerStatuses []gatev1alpha2.ListenerStatus) (gatev1alpha2.GatewayStatus, error) {
	// As Status.Addresses are not implemented yet, we initialize an empty array to follow the API expectations.
	gatewayStatus := gatev1alpha2.GatewayStatus{
		Addresses: []gatev1alpha2.GatewayAddress{},
	}

	var result error
	for i, listener := range listenerStatuses {
		if len(listener.Conditions) == 0 {
			// GatewayConditionReady "Ready", GatewayConditionReason "ListenerReady"
			listenerStatuses[i].Conditions = append(listenerStatuses[i].Conditions, metav1.Condition{
				Type:               string(gatev1alpha2.ListenerConditionReady),
				Status:             metav1.ConditionTrue,
				LastTransitionTime: metav1.Now(),
				Reason:             "ListenerReady",
				Message:            "No error found",
			})

			continue
		}

		for _, condition := range listener.Conditions {
			result = multierror.Append(result, errors.New(condition.Message))
		}
	}

	if result != nil {
		// GatewayConditionReady "Ready", GatewayConditionReason "ListenersNotValid"
		gatewayStatus.Conditions = append(gatewayStatus.Conditions, metav1.Condition{
			Type:               string(gatev1alpha2.GatewayConditionReady),
			Status:             metav1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1alpha2.GatewayReasonListenersNotValid),
			Message:            "All Listeners must be valid",
		})

		return gatewayStatus, result
	}

	gatewayStatus.Listeners = listenerStatuses

	gatewayStatus.Conditions = append(gatewayStatus.Conditions,
		// update "Scheduled" status with "ResourcesAvailable" reason
		metav1.Condition{
			Type:               string(gatev1alpha2.GatewayConditionScheduled),
			Status:             metav1.ConditionTrue,
			Reason:             "ResourcesAvailable",
			Message:            "Resources available",
			LastTransitionTime: metav1.Now(),
		},
		// update "Ready" status with "ListenersValid" reason
		metav1.Condition{
			Type:               string(gatev1alpha2.GatewayConditionReady),
			Status:             metav1.ConditionTrue,
			Reason:             "ListenersValid",
			Message:            "Listeners valid",
			LastTransitionTime: metav1.Now(),
		},
	)

	return gatewayStatus, nil
}

func (p *Provider) entryPointName(port gatev1alpha2.PortNumber, protocol gatev1alpha2.ProtocolType) (string, error) {
	portStr := strconv.FormatInt(int64(port), 10)

	for name, entryPoint := range p.EntryPoints {
		if strings.HasSuffix(entryPoint.Address, ":"+portStr) {
			// If the protocol is HTTP the entryPoint must have no TLS conf
			// Not relevant for gatev1alpha2.TLSProtocolType && gatev1alpha2.TCPProtocolType
			if protocol == gatev1alpha2.HTTPProtocolType && entryPoint.HasHTTPTLSConf {
				continue
			}

			return name, nil
		}
	}

	return "", fmt.Errorf("no matching entryPoint for port %d and protocol %q", port, protocol)
}

func supportedRouteKinds(protocol gatev1alpha2.ProtocolType) ([]gatev1alpha2.RouteGroupKind, []metav1.Condition) {
	group := gatev1alpha2.Group(gatev1alpha2.GroupName)

	switch protocol {
	case gatev1alpha2.TCPProtocolType:
		return []gatev1alpha2.RouteGroupKind{{Kind: kindTCPRoute, Group: &group}}, nil

	case gatev1alpha2.HTTPProtocolType, gatev1alpha2.HTTPSProtocolType:
		return []gatev1alpha2.RouteGroupKind{{Kind: kindHTTPRoute, Group: &group}}, nil

	case gatev1alpha2.TLSProtocolType:
		return []gatev1alpha2.RouteGroupKind{
			{Kind: kindTCPRoute, Group: &group},
			{Kind: kindTLSRoute, Group: &group},
		}, nil
	}

	return nil, []metav1.Condition{{
		Type:               string(gatev1alpha2.ListenerConditionDetached),
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             string(gatev1alpha2.ListenerReasonUnsupportedProtocol),
		Message:            fmt.Sprintf("Unsupported listener protocol %q", protocol),
	}}
}

func getAllowedRouteKinds(listener gatev1alpha2.Listener, supportedKinds []gatev1alpha2.RouteGroupKind) ([]gatev1alpha2.RouteGroupKind, []metav1.Condition) {
	if listener.AllowedRoutes == nil || len(listener.AllowedRoutes.Kinds) == 0 {
		return supportedKinds, nil
	}

	var (
		routeKinds []gatev1alpha2.RouteGroupKind
		conditions []metav1.Condition
	)

	uniqRouteKinds := map[gatev1alpha2.Kind]struct{}{}
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
				Type:               string(gatev1alpha2.ListenerConditionDetached),
				Status:             metav1.ConditionTrue,
				LastTransitionTime: metav1.Now(),
				Reason:             string(gatev1alpha2.ListenerReasonInvalidRouteKinds),
				Message:            fmt.Sprintf("Listener protocol %q does not support RouteGroupKind %v/%s", listener.Protocol, routeKind.Group, routeKind.Kind),
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

func (p *Provider) gatewayHTTPRouteToHTTPConf(ctx context.Context, ep string, listener gatev1alpha2.Listener, gateway *gatev1alpha2.Gateway, client Client, conf *dynamic.Configuration) (int32, []metav1.Condition) {
	if listener.AllowedRoutes == nil {
		// Should not happen due to validation.
		return 0, nil
	}

	namespaces, err := getRouteBindingSelectorNamespace(client, gateway.Namespace, listener.AllowedRoutes.Namespaces)
	if err != nil {
		// update "ResolvedRefs" status true with "InvalidRoutesRef" reason
		return 0, []metav1.Condition{{
			Type:               string(gatev1alpha2.ListenerConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             "InvalidRouteNamespacesSelector", // Should never happen as the selector is validated by kubernetes
			Message:            fmt.Sprintf("Invalid route namespaces selector: %v", err),
		}}
	}

	routes, err := client.GetHTTPRoutes(namespaces)
	if err != nil {
		// update "ResolvedRefs" status true with "InvalidRoutesRef" reason
		return 0, []metav1.Condition{{
			Type:               string(gatev1alpha2.ListenerConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1alpha2.ListenerReasonRefNotPermitted),
			Message:            fmt.Sprintf("Cannot fetch HTTPRoutes: %v", err),
		}}
	}

	if len(routes) == 0 {
		log.Ctx(ctx).Debug().Msg("No HTTPRoutes found")
		return 0, nil
	}

	numRoutesAttached := int32(0)
	var listenerConditions []metav1.Condition
	for _, route := range routes {
		if !shouldAttach(gateway, listener, route.Namespace, route.Spec.CommonRouteSpec) {
			continue
		}

		hostnames := matchingHostnames(listener, route.Spec.Hostnames)
		if len(hostnames) == 0 && listener.Hostname != nil && *listener.Hostname != "" && len(route.Spec.Hostnames) > 0 {
			// TODO update the corresponding route parent status
			// https://gateway-api.sigs.k8s.io/v1alpha2/references/spec/#gateway.networking.k8s.io/v1alpha2.TLSRoute
			continue
		}

		var routeConditions []metav1.Condition

		hostRule, err := hostRule(hostnames)
		if err != nil {
			routeConditions = append(routeConditions, metav1.Condition{
				Type:               string(gatev1alpha2.ConditionRouteResolvedRefs),
				Status:             metav1.ConditionFalse,
				LastTransitionTime: metav1.Now(),
				Reason:             "InvalidRouteHostname", // TODO check the spec if a proper reason is introduced at some point
				Message:            fmt.Sprintf("Skipping HTTPRoute %s: invalid hostname: %v", route.Name, err),
			})
			listenerConditions = append(listenerConditions, metav1.Condition{
				Type:               string(gatev1alpha2.ListenerConditionResolvedRefs),
				Status:             metav1.ConditionFalse,
				LastTransitionTime: metav1.Now(),
				Reason:             "InvalidRouteHostname", // TODO check the spec if a proper reason is introduced at some point
				Message:            fmt.Sprintf("Skipping HTTPRoute %s: invalid hostname: %v", route.Name, err),
			})
			continue
		}

		atLeastOneRuleMatched := false
		for _, routeRule := range route.Spec.Rules {
			rule, err := extractRule(routeRule, hostRule)
			if err != nil {
				routeConditions = append(routeConditions, metav1.Condition{
					Type:               string(gatev1alpha2.ConditionRouteResolvedRefs),
					Status:             metav1.ConditionFalse,
					LastTransitionTime: metav1.Now(),
					Reason:             "UnsupportedPathOrHeaderType", // TODO check the spec if a proper reason is introduced at some point
					Message:            fmt.Sprintf("Skipping HTTPRoute %s: cannot generate rule: %v", route.Name, err),
				})
				// update "ResolvedRefs" status true with "DroppedRoutes" reason
				listenerConditions = append(listenerConditions, metav1.Condition{
					Type:               string(gatev1alpha2.ListenerConditionResolvedRefs),
					Status:             metav1.ConditionFalse,
					LastTransitionTime: metav1.Now(),
					Reason:             "UnsupportedPathOrHeaderType", // TODO check the spec if a proper reason is introduced at some point
					Message:            fmt.Sprintf("Skipping HTTPRoute %s: cannot generate rule: %v", route.Name, err),
				})
			}

			router := dynamic.Router{
				Rule:        rule,
				EntryPoints: []string{ep},
			}

			if listener.Protocol == gatev1alpha2.HTTPSProtocolType && listener.TLS != nil {
				// TODO support let's encrypt
				router.TLS = &dynamic.RouterTLSConfig{}
			}

			// Adding the gateway desc and the entryPoint desc prevents overlapping of routers build from the same routes.
			routerName := route.Name + "-" + gateway.Name + "-" + ep
			routerKey, err := makeRouterKey(router.Rule, makeID(route.Namespace, routerName))
			if err != nil {
				routeConditions = append(routeConditions, metav1.Condition{
					Type:               string(gatev1alpha2.ConditionRouteResolvedRefs),
					Status:             metav1.ConditionFalse,
					LastTransitionTime: metav1.Now(),
					Reason:             "InvalidRouterKey", // Should never happen
					Message:            fmt.Sprintf("Skipping HTTPRoute %s: cannot make router's key with rule %s: %v", route.Name, router.Rule, err),
				})

				// update "ResolvedRefs" status true with "DroppedRoutes" reason
				listenerConditions = append(listenerConditions, metav1.Condition{
					Type:               string(gatev1alpha2.ListenerConditionResolvedRefs),
					Status:             metav1.ConditionFalse,
					LastTransitionTime: metav1.Now(),
					Reason:             "InvalidRouterKey", // Should never happen
					Message:            fmt.Sprintf("Skipping HTTPRoute %s: cannot make router's key with rule %s: %v", route.Name, router.Rule, err),
				})

				// TODO update the RouteStatus condition / deduplicate conditions on listener
				continue
			}

			middlewares, err := loadMiddlewares(listener, routerKey, routeRule.Filters)
			if err != nil {
				routeConditions = append(routeConditions, metav1.Condition{
					Type:               string(gatev1alpha2.ConditionRouteResolvedRefs),
					Status:             metav1.ConditionFalse,
					LastTransitionTime: metav1.Now(),
					Reason:             "InvalidFilters", // TODO check the spec if a proper reason is introduced at some point
					Message:            fmt.Sprintf("Cannot load HTTPRoute filter %s/%s: %v", route.Namespace, route.Name, err),
				})

				// update "ResolvedRefs" status true with "InvalidFilters" reason
				listenerConditions = append(listenerConditions, metav1.Condition{
					Type:               string(gatev1alpha2.ListenerConditionResolvedRefs),
					Status:             metav1.ConditionFalse,
					LastTransitionTime: metav1.Now(),
					Reason:             "InvalidFilters", // TODO check the spec if a proper reason is introduced at some point
					Message:            fmt.Sprintf("Cannot load HTTPRoute filter %s/%s: %v", route.Namespace, route.Name, err),
				})

				// TODO update the RouteStatus condition / deduplicate conditions on listener
				continue
			}

			for middlewareName, middleware := range middlewares {
				conf.HTTP.Middlewares[middlewareName] = middleware
				router.Middlewares = append(router.Middlewares, middlewareName)
			}

			if len(routeRule.BackendRefs) == 0 {
				continue
			}

			// Traefik internal service can be used only if there is only one BackendRef service reference.
			if len(routeRule.BackendRefs) == 1 && isInternalService(routeRule.BackendRefs[0].BackendRef) {
				router.Service = string(routeRule.BackendRefs[0].Name)
			} else {
				wrrService, subServices, err := loadServices(client, route.Namespace, routeRule.BackendRefs)
				if err != nil {
					routeConditions = append(routeConditions, metav1.Condition{
						Type:               string(gatev1alpha2.ConditionRouteResolvedRefs),
						Status:             metav1.ConditionFalse,
						LastTransitionTime: metav1.Now(),
						Reason:             "InvalidBackendRefs", // TODO check the spec if a proper reason is introduced at some point
						Message:            fmt.Sprintf("Cannot load HTTPRoute service %s/%s: %v", route.Namespace, route.Name, err),
					})

					// update "ResolvedRefs" status true with "DroppedRoutes" reason
					listenerConditions = append(listenerConditions, metav1.Condition{
						Type:               string(gatev1alpha2.ListenerConditionResolvedRefs),
						Status:             metav1.ConditionFalse,
						LastTransitionTime: metav1.Now(),
						Reason:             "InvalidBackendRefs", // TODO check the spec if a proper reason is introduced at some point
						Message:            fmt.Sprintf("Cannot load HTTPRoute service %s/%s: %v", route.Namespace, route.Name, err),
					})

					// TODO update the RouteStatus condition / deduplicate conditions on listener
					continue
				}

				for svcName, svc := range subServices {
					conf.HTTP.Services[svcName] = svc
				}

				serviceName := provider.Normalize(routerKey + "-wrr")
				conf.HTTP.Services[serviceName] = wrrService

				router.Service = serviceName
			}

			rt := &router
			p.applyRouterTransform(ctx, rt, route)

			routerKey = provider.Normalize(routerKey)
			conf.HTTP.Routers[routerKey] = rt

			atLeastOneRuleMatched = true
		}

		if atLeastOneRuleMatched {
			numRoutesAttached++
		}

		if err := updateHTTPRouteStatus(ctx, client, gateway, listener, route, routeConditions); err != nil {
			log.Ctx(ctx).Err(err).Send()
		}
	}

	return numRoutesAttached, listenerConditions
}

func gatewayTCPRouteToTCPConf(ctx context.Context, ep string, listener gatev1alpha2.Listener, gateway *gatev1alpha2.Gateway, client Client, conf *dynamic.Configuration) (int32, []metav1.Condition) {
	if listener.AllowedRoutes == nil {
		// Should not happen due to validation.
		return 0, nil
	}

	namespaces, err := getRouteBindingSelectorNamespace(client, gateway.Namespace, listener.AllowedRoutes.Namespaces)
	if err != nil {
		// update "ResolvedRefs" status true with "InvalidRoutesRef" reason
		return 0, []metav1.Condition{{
			Type:               string(gatev1alpha2.ListenerConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             "InvalidRouteNamespacesSelector", // TODO should never happen as the selector is validated by Kubernetes
			Message:            fmt.Sprintf("Invalid route namespaces selector: %v", err),
		}}
	}

	routes, err := client.GetTCPRoutes(namespaces)
	if err != nil {
		// update "ResolvedRefs" status true with "InvalidRoutesRef" reason
		return 0, []metav1.Condition{{
			Type:               string(gatev1alpha2.ListenerConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1alpha2.ListenerReasonRefNotPermitted),
			Message:            fmt.Sprintf("Cannot fetch TCPRoutes: %v", err),
		}}
	}

	if len(routes) == 0 {
		log.Ctx(ctx).Debug().Msg("No TCPRoutes found")
		return 0, nil
	}

	numRoutesAttached := int32(0)
	var listenerConditions []metav1.Condition
	for _, route := range routes {
		if !shouldAttach(gateway, listener, route.Namespace, route.Spec.CommonRouteSpec) {
			continue
		}

		router := dynamic.TCPRouter{
			Rule:        "HostSNI(`*`)",
			EntryPoints: []string{ep},
		}

		if listener.Protocol == gatev1alpha2.TLSProtocolType && listener.TLS != nil {
			// TODO support let's encrypt
			router.TLS = &dynamic.RouterTCPTLSConfig{
				Passthrough: listener.TLS.Mode != nil && *listener.TLS.Mode == gatev1alpha2.TLSModePassthrough,
			}
		}

		var routeConditions []metav1.Condition

		// Adding the gateway desc and the entryPoint desc prevents overlapping of routers build from the same routes.
		routerName := route.Name + "-" + gateway.Name + "-" + ep
		routerKey, err := makeRouterKey("", makeID(route.Namespace, routerName))
		if err != nil {
			routeConditions = append(routeConditions, metav1.Condition{
				Type:               string(gatev1alpha2.ConditionRouteResolvedRefs),
				Status:             metav1.ConditionFalse,
				LastTransitionTime: metav1.Now(),
				Reason:             "InvalidRouterKey", // Should never happen
				Message:            fmt.Sprintf("Skipping TCPRoute %s: cannot make router's key with rule %s: %v", route.Name, router.Rule, err),
			})

			// update "ResolvedRefs" status true with "DroppedRoutes" reason
			listenerConditions = append(listenerConditions, metav1.Condition{
				Type:               string(gatev1alpha2.ListenerConditionResolvedRefs),
				Status:             metav1.ConditionFalse,
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
				routeConditions = append(routeConditions, metav1.Condition{
					Type:               string(gatev1alpha2.ConditionRouteResolvedRefs),
					Status:             metav1.ConditionFalse,
					LastTransitionTime: metav1.Now(),
					Reason:             "InvalidBackendRefs", // TODO check the spec if a proper reason is introduced at some point
					Message:            fmt.Sprintf("Cannot load TCPRoute service %s/%s: %v", route.Namespace, route.Name, err),
				})

				// update "ResolvedRefs" status true with "DroppedRoutes" reason
				listenerConditions = append(listenerConditions, metav1.Condition{
					Type:               string(gatev1alpha2.ListenerConditionResolvedRefs),
					Status:             metav1.ConditionFalse,
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
			if err := updateTCPRouteStatus(ctx, client, gateway, listener, route, routeConditions); err != nil {
				log.Ctx(ctx).Err(err).Send()
				continue
			}
			numRoutesAttached++
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

		if len(ruleServiceNames) > 0 {
			if err := updateTCPRouteStatus(ctx, client, gateway, listener, route, routeConditions); err != nil {
				log.Ctx(ctx).Err(err).Send()
				continue
			}
			numRoutesAttached++
		}
	}

	return numRoutesAttached, listenerConditions
}

func gatewayTLSRouteToTCPConf(ctx context.Context, ep string, listener gatev1alpha2.Listener, gateway *gatev1alpha2.Gateway, client Client, conf *dynamic.Configuration) (int32, []metav1.Condition) {
	if listener.AllowedRoutes == nil {
		// Should not happen due to validation.
		return 0, nil
	}

	namespaces, err := getRouteBindingSelectorNamespace(client, gateway.Namespace, listener.AllowedRoutes.Namespaces)
	if err != nil {
		// update "ResolvedRefs" status true with "InvalidRoutesRef" reason
		return 0, []metav1.Condition{{
			Type:               string(gatev1alpha2.ListenerConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             "InvalidRouteNamespacesSelector", // TODO should never happen as the selector is validated by Kubernetes
			Message:            fmt.Sprintf("Invalid route namespaces selector: %v", err),
		}}
	}

	routes, err := client.GetTLSRoutes(namespaces)
	if err != nil {
		// update "ResolvedRefs" status true with "InvalidRoutesRef" reason
		return 0, []metav1.Condition{{
			Type:               string(gatev1alpha2.ListenerConditionResolvedRefs),
			Status:             metav1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatev1alpha2.ListenerReasonRefNotPermitted),
			Message:            fmt.Sprintf("Cannot fetch TLSRoutes: %v", err),
		}}
	}

	if len(routes) == 0 {
		log.Ctx(ctx).Debug().Msg("No TLSRoutes found")
		return 0, nil
	}

	numRoutesAttached := int32(0)
	var listenerConditions []metav1.Condition
	for _, route := range routes {
		if !shouldAttach(gateway, listener, route.Namespace, route.Spec.CommonRouteSpec) {
			continue
		}

		hostnames := matchingHostnames(listener, route.Spec.Hostnames)
		if len(hostnames) == 0 && listener.Hostname != nil && *listener.Hostname != "" && len(route.Spec.Hostnames) > 0 {
			for _, parent := range route.Status.Parents {
				parent.Conditions = append(parent.Conditions, metav1.Condition{
					Type:               string(gatev1alpha2.GatewayClassConditionStatusAccepted),
					Status:             metav1.ConditionFalse,
					Reason:             string(gatev1alpha2.ListenerReasonRouteConflict),
					Message:            fmt.Sprintf("No hostname match between listener: %v and route: %v", listener.Hostname, route.Spec.Hostnames),
					LastTransitionTime: metav1.Now(),
				})
			}

			continue
		}

		var routeConditions []metav1.Condition

		rule, err := hostSNIRule(hostnames)
		if err != nil {
			routeConditions = append(routeConditions, metav1.Condition{
				Type:               string(gatev1alpha2.ConditionRouteResolvedRefs),
				Status:             metav1.ConditionFalse,
				LastTransitionTime: metav1.Now(),
				Reason:             "InvalidHostnames", // TODO check the spec if a proper reason is introduced at some point
				Message:            fmt.Sprintf("Skipping TLSRoute %s: cannot make route's SNI match: %v", route.Name, err),
			})

			// update "ResolvedRefs" status true with "DroppedRoutes" reason
			listenerConditions = append(listenerConditions, metav1.Condition{
				Type:               string(gatev1alpha2.ListenerConditionResolvedRefs),
				Status:             metav1.ConditionFalse,
				LastTransitionTime: metav1.Now(),
				Reason:             "InvalidHostnames", // TODO check the spec if a proper reason is introduced at some point
				Message:            fmt.Sprintf("Skipping TLSRoute %s: cannot make route's SNI match: %v", route.Name, err),
			})
			// TODO update the RouteStatus condition / deduplicate conditions on listener
			continue
		}

		router := dynamic.TCPRouter{
			Rule:        rule,
			EntryPoints: []string{ep},
			TLS: &dynamic.RouterTCPTLSConfig{
				Passthrough: listener.TLS.Mode != nil && *listener.TLS.Mode == gatev1alpha2.TLSModePassthrough,
			},
		}

		// Adding the gateway desc and the entryPoint desc prevents overlapping of routers build from the same routes.
		routerName := route.Name + "-" + gateway.Name + "-" + ep
		routerKey, err := makeRouterKey(rule, makeID(route.Namespace, routerName))
		if err != nil {
			routeConditions = append(routeConditions, metav1.Condition{
				Type:               string(gatev1alpha2.ConditionRouteResolvedRefs),
				Status:             metav1.ConditionFalse,
				LastTransitionTime: metav1.Now(),
				Reason:             "InvalidRouterKey", // Should never happen
				Message:            fmt.Sprintf("Skipping TLSRoute %s: cannot make router's key with rule %s: %v", route.Name, router.Rule, err),
			})

			// update "ResolvedRefs" status true with "DroppedRoutes" reason
			listenerConditions = append(listenerConditions, metav1.Condition{
				Type:               string(gatev1alpha2.ListenerConditionResolvedRefs),
				Status:             metav1.ConditionFalse,
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
				routeConditions = append(routeConditions, metav1.Condition{
					Type:               string(gatev1alpha2.ConditionRouteResolvedRefs),
					Status:             metav1.ConditionFalse,
					LastTransitionTime: metav1.Now(),
					Reason:             "InvalidBackendRefs", // TODO check the spec if a proper reason is introduced at some point
					Message:            fmt.Sprintf("Cannot load TLSRoute service %s/%s: %v", route.Namespace, route.Name, err),
				})

				// update "ResolvedRefs" status true with "DroppedRoutes" reason
				listenerConditions = append(listenerConditions, metav1.Condition{
					Type:               string(gatev1alpha2.ListenerConditionResolvedRefs),
					Status:             metav1.ConditionFalse,
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
			if err := updateTLSRouteStatus(ctx, client, gateway, listener, route, routeConditions); err != nil {
				log.Ctx(ctx).Err(err).Send()
				continue
			}
			numRoutesAttached++
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

		if len(ruleServiceNames) > 0 {
			if err := updateTLSRouteStatus(ctx, client, gateway, listener, route, routeConditions); err != nil {
				log.Ctx(ctx).Err(err).Send()
				continue
			}
			numRoutesAttached++
		}
	}

	return numRoutesAttached, listenerConditions
}

// Because of Kubernetes validation we admit that the given Hostnames are valid.
// https://github.com/kubernetes-sigs/gateway-api/blob/ff9883da4cad8554cd300394f725ab3a27502785/apis/v1alpha2/shared_types.go#L252
func matchingHostnames(listener gatev1alpha2.Listener, hostnames []gatev1alpha2.Hostname) []gatev1alpha2.Hostname {
	if listener.Hostname == nil || *listener.Hostname == "" {
		return hostnames
	}

	if len(hostnames) == 0 {
		return []gatev1alpha2.Hostname{*listener.Hostname}
	}

	listenerLabels := strings.Split(string(*listener.Hostname), ".")

	var matches []gatev1alpha2.Hostname

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

func shouldAttach(gateway *gatev1alpha2.Gateway, listener gatev1alpha2.Listener, routeNamespace string, routeSpec gatev1alpha2.CommonRouteSpec) bool {
	for _, parentRef := range routeSpec.ParentRefs {
		if parentRef.Group == nil || *parentRef.Group != gatev1alpha2.GroupName {
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
			return true
		}
	}

	return false
}

func getRouteBindingSelectorNamespace(client Client, gatewayNamespace string, routeNamespaces *gatev1alpha2.RouteNamespaces) ([]string, error) {
	if routeNamespaces == nil || routeNamespaces.From == nil {
		return []string{gatewayNamespace}, nil
	}

	switch *routeNamespaces.From {
	case gatev1alpha2.NamespacesFromAll:
		return []string{metav1.NamespaceAll}, nil

	case gatev1alpha2.NamespacesFromSame:
		return []string{gatewayNamespace}, nil

	case gatev1alpha2.NamespacesFromSelector:
		selector, err := metav1.LabelSelectorAsSelector(routeNamespaces.Selector)
		if err != nil {
			return nil, fmt.Errorf("malformed selector: %w", err)
		}

		return client.GetNamespaces(selector)
	}

	return nil, fmt.Errorf("unsupported RouteSelectType: %q", *routeNamespaces.From)
}

func hostRule(hostnames []gatev1alpha2.Hostname) (string, error) {
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

func hostSNIRule(hostnames []gatev1alpha2.Hostname) (string, error) {
	rules := make([]string, 0, len(hostnames))
	uniqHostnames := map[gatev1alpha2.Hostname]struct{}{}

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

func extractRule(routeRule gatev1alpha2.HTTPRouteRule, hostRule string) (string, error) {
	var rule string
	var matchesRules []string

	for _, match := range routeRule.Matches {
		if (match.Path == nil || match.Path.Type == nil) && match.Headers == nil {
			continue
		}

		var matchRules []string

		if match.Path != nil && match.Path.Type != nil && match.Path.Value != nil {
			// TODO handle other path types
			switch *match.Path.Type {
			case gatev1alpha2.PathMatchExact:
				matchRules = append(matchRules, fmt.Sprintf("Path(`%s`)", *match.Path.Value))
			case gatev1alpha2.PathMatchPathPrefix:
				matchRules = append(matchRules, fmt.Sprintf("PathPrefix(`%s`)", *match.Path.Value))
			default:
				return "", fmt.Errorf("unsupported path match %s", *match.Path.Type)
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

func extractHeaderRules(headers []gatev1alpha2.HTTPHeaderMatch) ([]string, error) {
	var headerRules []string

	// TODO handle other headers types
	for _, header := range headers {
		if header.Type == nil {
			// Should never happen due to kubernetes validation.
			continue
		}

		switch *header.Type {
		case gatev1alpha2.HeaderMatchExact:
			headerRules = append(headerRules, fmt.Sprintf("Headers(`%s`,`%s`)", header.Name, header.Value))
		default:
			return nil, fmt.Errorf("unsupported header match type %s", *header.Type)
		}
	}

	return headerRules, nil
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

func getTLS(k8sClient Client, secretName gatev1alpha2.ObjectName, namespace string) (*tls.CertAndStores, error) {
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
			CertFile: tls.FileOrContent(cert),
			KeyFile:  tls.FileOrContent(key),
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

// loadServices is generating a WRR service, even when there is only one target.
func loadServices(client Client, namespace string, backendRefs []gatev1alpha2.HTTPBackendRef) (*dynamic.Service, map[string]*dynamic.Service, error) {
	services := map[string]*dynamic.Service{}

	wrrSvc := &dynamic.Service{
		Weighted: &dynamic.WeightedRoundRobin{
			Services: []dynamic.WRRService{},
		},
	}

	for _, backendRef := range backendRefs {
		if backendRef.Group == nil || backendRef.Kind == nil {
			// Should not happen as this is validated by kubernetes
			continue
		}

		if isInternalService(backendRef.BackendRef) {
			return nil, nil, fmt.Errorf("traefik internal service %s is not allowed in a WRR loadbalancer", backendRef.BackendRef.Name)
		}

		weight := int(pointer.Int32Deref(backendRef.Weight, 1))

		if isTraefikService(backendRef.BackendRef) {
			wrrSvc.Weighted.Services = append(wrrSvc.Weighted.Services, dynamic.WRRService{Name: string(backendRef.Name), Weight: &weight})
			continue
		}

		if *backendRef.Group != "" && *backendRef.Group != "core" && *backendRef.Kind != "Service" {
			return nil, nil, fmt.Errorf("unsupported HTTPBackendRef %s/%s/%s", *backendRef.Group, *backendRef.Kind, backendRef.Name)
		}

		lb := &dynamic.ServersLoadBalancer{}
		lb.SetDefaults()

		svc := dynamic.Service{LoadBalancer: lb}

		// TODO support cross namespace through ReferencePolicy
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

			protocol := getProtocol(portSpec)

			portStr = strconv.FormatInt(int64(port), 10)
			for _, addr := range subset.Addresses {
				svc.LoadBalancer.Servers = append(svc.LoadBalancer.Servers, dynamic.Server{
					URL: fmt.Sprintf("%s://%s", protocol, net.JoinHostPort(addr.IP, portStr)),
				})
			}
		}

		serviceName := provider.Normalize(makeID(service.Namespace, service.Name) + "-" + portStr)
		services[serviceName] = &svc

		wrrSvc.Weighted.Services = append(wrrSvc.Weighted.Services, dynamic.WRRService{Name: serviceName, Weight: &weight})
	}

	if len(wrrSvc.Weighted.Services) == 0 {
		return nil, nil, errors.New("no service has been created")
	}

	return wrrSvc, services, nil
}

// loadTCPServices is generating a WRR service, even when there is only one target.
func loadTCPServices(client Client, namespace string, backendRefs []gatev1alpha2.BackendRef) (*dynamic.TCPService, map[string]*dynamic.TCPService, error) {
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

		weight := int(pointer.Int32Deref(backendRef.Weight, 1))

		if isTraefikService(backendRef) {
			wrrSvc.Weighted.Services = append(wrrSvc.Weighted.Services, dynamic.TCPWRRService{Name: string(backendRef.Name), Weight: &weight})
			continue
		}

		if *backendRef.Group != "" && *backendRef.Group != "core" && *backendRef.Kind != "Service" {
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

func loadMiddlewares(listener gatev1alpha2.Listener, prefix string, filters []gatev1alpha2.HTTPRouteFilter) (map[string]*dynamic.Middleware, error) {
	middlewares := make(map[string]*dynamic.Middleware)

	// The spec allows for an empty string in which case we should use the
	// scheme of the request which in this case is the listener scheme.
	var listenerScheme string
	switch listener.Protocol {
	case gatev1alpha2.HTTPProtocolType:
		listenerScheme = "http"
	case gatev1alpha2.HTTPSProtocolType:
		listenerScheme = "https"
	default:
		return nil, fmt.Errorf("invalid listener protocol %s", listener.Protocol)
	}

	for i, filter := range filters {
		var middleware *dynamic.Middleware
		switch filter.Type {
		case gatev1alpha2.HTTPRouteFilterRequestRedirect:
			var err error
			middleware, err = createRedirectRegexMiddleware(listenerScheme, filter.RequestRedirect)
			if err != nil {
				return nil, fmt.Errorf("creating RedirectRegex middleware: %w", err)
			}
		default:
			// As per the spec:
			// https://gateway-api.sigs.k8s.io/api-types/httproute/#filters-optional
			// In all cases where incompatible or unsupported filters are
			// specified, implementations MUST add a warning condition to
			// status.
			return nil, fmt.Errorf("unsupported filter %s", filter.Type)
		}

		middlewareName := provider.Normalize(fmt.Sprintf("%s-%s-%d", prefix, strings.ToLower(string(filter.Type)), i))
		middlewares[middlewareName] = middleware
	}

	return middlewares, nil
}

func createRedirectRegexMiddleware(scheme string, filter *gatev1alpha2.HTTPRequestRedirectFilter) (*dynamic.Middleware, error) {
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

func isTraefikService(ref gatev1alpha2.BackendRef) bool {
	if ref.Kind == nil || ref.Group == nil {
		return false
	}

	return *ref.Group == traefikv1alpha1.GroupName && *ref.Kind == kindTraefikService
}

func isInternalService(ref gatev1alpha2.BackendRef) bool {
	return isTraefikService(ref) && strings.HasSuffix(string(ref.Name), "@internal")
}

// makeListenerKey joins protocol, hostname, and port of a listener into a string key.
func makeListenerKey(l gatev1alpha2.Listener) string {
	var hostname gatev1alpha2.Hostname
	if l.Hostname != nil {
		hostname = *l.Hostname
	}

	return fmt.Sprintf("%s|%s|%d", l.Protocol, hostname, l.Port)
}

func updateHTTPRouteStatus(ctx context.Context, client Client, gateway *gatev1alpha2.Gateway, listener gatev1alpha2.Listener, route *gatev1alpha2.HTTPRoute, conditions []metav1.Condition) error {
	status := route.Status.RouteStatus.DeepCopy()
	spec := route.Spec.CommonRouteSpec.DeepCopy()
	return client.UpdateHTTPRouteStatus(route, makeRouteStatus(ctx, *status, *spec, route.ObjectMeta, gateway, listener, conditions))
}

func updateTCPRouteStatus(ctx context.Context, client Client, gateway *gatev1alpha2.Gateway, listener gatev1alpha2.Listener, route *gatev1alpha2.TCPRoute, conditions []metav1.Condition) error {
	status := route.Status.RouteStatus.DeepCopy()
	spec := route.Spec.CommonRouteSpec.DeepCopy()
	return client.UpdateTCPRouteStatus(route, makeRouteStatus(ctx, *status, *spec, route.ObjectMeta, gateway, listener, conditions))
}

func updateTLSRouteStatus(ctx context.Context, client Client, gateway *gatev1alpha2.Gateway, listener gatev1alpha2.Listener, route *gatev1alpha2.TLSRoute, conditions []metav1.Condition) error {
	status := route.Status.RouteStatus.DeepCopy()
	spec := route.Spec.CommonRouteSpec.DeepCopy()
	return client.UpdateTLSRouteStatus(route, makeRouteStatus(ctx, *status, *spec, route.ObjectMeta, gateway, listener, conditions))
}

func makeRouteStatus(ctx context.Context, routeStatus gatev1alpha2.RouteStatus, routeSpec gatev1alpha2.CommonRouteSpec, routeMeta metav1.ObjectMeta, gateway *gatev1alpha2.Gateway, listener gatev1alpha2.Listener, conditions []metav1.Condition) gatev1alpha2.RouteStatus {
	var routeParentStatus *gatev1alpha2.RouteParentStatus

	// Check if we need to update an existing parent reference
	for i, parent := range routeStatus.Parents {
		if parent.ControllerName != controllerName {
			continue
		}

		parentRef := parent.ParentRef

		if parentRef.Group == nil || *parentRef.Group != gatev1alpha2.GroupName {
			continue
		}
		if parentRef.Kind == nil || *parentRef.Kind != kindGateway {
			continue
		}
		if parentRef.SectionName != nil && *parentRef.SectionName != listener.Name {
			continue
		}

		namespace := routeMeta.Namespace
		if parentRef.Namespace != nil {
			namespace = string(*parentRef.Namespace)
		}

		if namespace == gateway.Namespace && string(parentRef.Name) == gateway.Name {
			log.Ctx(ctx).Debug().Msgf("Found existing parent reference for route %s/%s", routeMeta.Namespace, routeMeta.Name)
			routeParentStatus = &routeStatus.Parents[i]
		}
	}

	// No existing parent reference was found - create a new one
	if routeParentStatus == nil {
		log.Ctx(ctx).Debug().Msgf("Adding parent reference for route %s/%s", routeMeta.Namespace, routeMeta.Name)

		group := gatev1alpha2.Group(gatev1alpha2.GroupName)
		kind := gatev1alpha2.Kind(kindGateway)
		namespace := gatev1alpha2.Namespace(gateway.Namespace)

		// Only set the section name if the route requests it
		var sectionName *gatev1alpha2.SectionName
		for _, p := range routeSpec.ParentRefs {
			if p.SectionName != nil && *p.SectionName == listener.Name {
				sectionName = p.SectionName
			}
		}

		routeStatus.Parents = append(routeStatus.Parents, gatev1alpha2.RouteParentStatus{
			ParentRef: gatev1alpha2.ParentRef{
				Group:       &group,
				Kind:        &kind,
				Namespace:   &namespace,
				Name:        gatev1alpha2.ObjectName(gateway.Name),
				SectionName: sectionName,
			},
			ControllerName: controllerName,
		})

		routeParentStatus = &routeStatus.Parents[len(routeStatus.Parents)-1]
	}

	routeStatusCondition := metav1.ConditionTrue
	routeStatusReason := string(gatev1alpha2.ConditionRouteAccepted)
	if len(conditions) > 0 {
		routeStatusCondition = metav1.ConditionFalse
		routeStatusReason = conditions[0].Reason
	}

	// Don't write duplicate conditions
	shouldUpdateConditions := true
	if len(routeParentStatus.Conditions) > 0 {
		lastCondition := routeParentStatus.Conditions[len(routeParentStatus.Conditions)-1]

		if lastCondition.Type == string(gatev1alpha2.ConditionRouteAccepted) && lastCondition.Status == routeStatusCondition {
			shouldUpdateConditions = false
		}
	}

	if shouldUpdateConditions {
		routeStatusString := "attached to"
		if routeStatusCondition == metav1.ConditionFalse {
			routeStatusString = "rejected by"
		}

		log.Ctx(ctx).Debug().Msgf("Route %s/%s was %s gateway", routeMeta.Namespace, routeMeta.Name, routeStatusString)

		// TODO dedupe and add passed in conditions

		routeParentStatus.Conditions = append(routeParentStatus.Conditions, metav1.Condition{
			Type:               string(gatev1alpha2.ConditionRouteAccepted),
			Status:             routeStatusCondition,
			ObservedGeneration: routeMeta.Generation,
			LastTransitionTime: metav1.NewTime(timeNow()),
			Reason:             routeStatusReason,
			Message:            fmt.Sprintf("The route was %s the Gateway", routeStatusString),
		})
	}

	return routeStatus
}
