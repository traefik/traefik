package crd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/rs/zerolog/log"
	traefikclientset "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/generated/clientset/versioned"
	traefikinformers "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/generated/informers/externalversions"
	traefikv1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1"
	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/k8s"
	"github.com/traefik/traefik/v3/pkg/version"
	corev1 "k8s.io/api/core/v1"
	kerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kinformers "k8s.io/client-go/informers"
	kclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const resyncPeriod = 10 * time.Minute

// Client is a client for the Provider master.
// WatchAll starts the watch of the Provider resources and updates the stores.
// The stores can then be accessed via the Get* functions.
type Client interface {
	WatchAll(namespaces []string, stopCh <-chan struct{}) (<-chan interface{}, error)
	GetIngressRoutes() []*traefikv1.IngressRoute
	GetIngressRouteTCPs() []*traefikv1.IngressRouteTCP
	GetIngressRouteUDPs() []*traefikv1.IngressRouteUDP
	GetMiddlewares() []*traefikv1.Middleware
	GetMiddlewareTCPs() []*traefikv1.MiddlewareTCP
	GetTraefikService(namespace, name string) (*traefikv1.TraefikService, bool, error)
	GetTraefikServices() []*traefikv1.TraefikService
	GetTLSOptions() []*traefikv1.TLSOption
	GetServersTransports() []*traefikv1.ServersTransport
	GetServersTransportTCPs() []*traefikv1.ServersTransportTCP
	GetTLSStores() []*traefikv1.TLSStore
	GetService(namespace, name string) (*corev1.Service, bool, error)
	GetSecret(namespace, name string) (*corev1.Secret, bool, error)
	GetEndpoints(namespace, name string) (*corev1.Endpoints, bool, error)
}

// TODO: add tests for the clientWrapper (and its methods) itself.
type clientWrapper struct {
	csCrd  traefikclientset.Interface
	csKube kclientset.Interface

	factoriesCrd    map[string]traefikinformers.SharedInformerFactory
	factoriesKube   map[string]kinformers.SharedInformerFactory
	factoriesSecret map[string]kinformers.SharedInformerFactory

	labelSelector string

	isNamespaceAll    bool
	watchedNamespaces []string
}

func createClientFromConfig(c *rest.Config) (*clientWrapper, error) {
	c.UserAgent = fmt.Sprintf(
		"%s/%s (%s/%s) kubernetes/crd",
		filepath.Base(os.Args[0]),
		version.Version,
		runtime.GOOS,
		runtime.GOARCH,
	)

	csCrd, err := traefikclientset.NewForConfig(c)
	if err != nil {
		return nil, err
	}

	csKube, err := kclientset.NewForConfig(c)
	if err != nil {
		return nil, err
	}

	return newClientImpl(csKube, csCrd), nil
}

func newClientImpl(csKube kclientset.Interface, csCrd traefikclientset.Interface) *clientWrapper {
	return &clientWrapper{
		csCrd:           csCrd,
		csKube:          csKube,
		factoriesCrd:    make(map[string]traefikinformers.SharedInformerFactory),
		factoriesKube:   make(map[string]kinformers.SharedInformerFactory),
		factoriesSecret: make(map[string]kinformers.SharedInformerFactory),
	}
}

// newInClusterClient returns a new Provider client that is expected to run
// inside the cluster.
func newInClusterClient(endpoint string) (*clientWrapper, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster configuration: %w", err)
	}

	if endpoint != "" {
		config.Host = endpoint
	}

	return createClientFromConfig(config)
}

func newExternalClusterClientFromFile(file string) (*clientWrapper, error) {
	configFromFlags, err := clientcmd.BuildConfigFromFlags("", file)
	if err != nil {
		return nil, err
	}
	return createClientFromConfig(configFromFlags)
}

// newExternalClusterClient returns a new Provider client that may run outside the cluster.
// The endpoint parameter must not be empty.
func newExternalClusterClient(endpoint, token, caFilePath string) (*clientWrapper, error) {
	if endpoint == "" {
		return nil, errors.New("endpoint missing for external cluster client")
	}

	config := &rest.Config{
		Host:        endpoint,
		BearerToken: token,
	}

	if caFilePath != "" {
		caData, err := os.ReadFile(caFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA file %s: %w", caFilePath, err)
		}

		config.TLSClientConfig = rest.TLSClientConfig{CAData: caData}
	}

	return createClientFromConfig(config)
}

// WatchAll starts namespace-specific controllers for all relevant kinds.
func (c *clientWrapper) WatchAll(namespaces []string, stopCh <-chan struct{}) (<-chan interface{}, error) {
	eventCh := make(chan interface{}, 1)
	eventHandler := &k8s.ResourceEventHandler{Ev: eventCh}

	if len(namespaces) == 0 {
		namespaces = []string{metav1.NamespaceAll}
		c.isNamespaceAll = true
	}

	c.watchedNamespaces = namespaces

	notOwnedByHelm := func(opts *metav1.ListOptions) {
		opts.LabelSelector = "owner!=helm"
	}

	matchesLabelSelector := func(opts *metav1.ListOptions) {
		opts.LabelSelector = c.labelSelector
	}

	for _, ns := range namespaces {
		factoryCrd := traefikinformers.NewSharedInformerFactoryWithOptions(c.csCrd, resyncPeriod, traefikinformers.WithNamespace(ns), traefikinformers.WithTweakListOptions(matchesLabelSelector))
		_, err := factoryCrd.Traefik().V1().IngressRoutes().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		_, err = factoryCrd.Traefik().V1().Middlewares().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		_, err = factoryCrd.Traefik().V1().MiddlewareTCPs().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		_, err = factoryCrd.Traefik().V1().IngressRouteTCPs().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		_, err = factoryCrd.Traefik().V1().IngressRouteUDPs().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		_, err = factoryCrd.Traefik().V1().TLSOptions().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		_, err = factoryCrd.Traefik().V1().ServersTransports().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		_, err = factoryCrd.Traefik().V1().ServersTransportTCPs().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		_, err = factoryCrd.Traefik().V1().TLSStores().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		_, err = factoryCrd.Traefik().V1().TraefikServices().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}

		err = addV1alpha1Informers(factoryCrd, eventHandler)
		if err != nil {
			return nil, err
		}

		factoryKube := kinformers.NewSharedInformerFactoryWithOptions(c.csKube, resyncPeriod, kinformers.WithNamespace(ns))
		_, err = factoryKube.Core().V1().Services().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		_, err = factoryKube.Core().V1().Endpoints().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}

		factorySecret := kinformers.NewSharedInformerFactoryWithOptions(c.csKube, resyncPeriod, kinformers.WithNamespace(ns), kinformers.WithTweakListOptions(notOwnedByHelm))
		_, err = factorySecret.Core().V1().Secrets().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}

		c.factoriesCrd[ns] = factoryCrd
		c.factoriesKube[ns] = factoryKube
		c.factoriesSecret[ns] = factorySecret
	}

	for _, ns := range namespaces {
		c.factoriesCrd[ns].Start(stopCh)
		c.factoriesKube[ns].Start(stopCh)
		c.factoriesSecret[ns].Start(stopCh)
	}

	for _, ns := range namespaces {
		for t, ok := range c.factoriesCrd[ns].WaitForCacheSync(stopCh) {
			if !ok {
				return nil, fmt.Errorf("timed out waiting for controller caches to sync %s in namespace %q", t.String(), ns)
			}
		}

		for t, ok := range c.factoriesKube[ns].WaitForCacheSync(stopCh) {
			if !ok {
				return nil, fmt.Errorf("timed out waiting for controller caches to sync %s in namespace %q", t.String(), ns)
			}
		}

		for t, ok := range c.factoriesSecret[ns].WaitForCacheSync(stopCh) {
			if !ok {
				return nil, fmt.Errorf("timed out waiting for controller caches to sync %s in namespace %q", t.String(), ns)
			}
		}
	}

	return eventCh, nil
}

func (c *clientWrapper) GetIngressRoutes() []*traefikv1.IngressRoute {
	var result []*traefikv1.IngressRoute

	for ns, factory := range c.factoriesCrd {
		ings, err := factory.Traefik().V1().IngressRoutes().Lister().List(labels.Everything())
		if err != nil {
			log.Error().Err(err).Msgf("Failed to list ingress routes in namespace %s", ns)
		}
		result = append(result, ings...)
	}

	return c.appendV1alpha1IngressRoutes(result)
}

func (c *clientWrapper) GetIngressRouteTCPs() []*traefikv1.IngressRouteTCP {
	var result []*traefikv1.IngressRouteTCP

	for ns, factory := range c.factoriesCrd {
		ings, err := factory.Traefik().V1().IngressRouteTCPs().Lister().List(labels.Everything())
		if err != nil {
			log.Error().Err(err).Msgf("Failed to list tcp ingress routes in namespace %s", ns)
		}
		result = append(result, ings...)
	}

	return c.appendV1alpha1IngressRouteTCPs(result)
}

func (c *clientWrapper) GetIngressRouteUDPs() []*traefikv1.IngressRouteUDP {
	var result []*traefikv1.IngressRouteUDP

	for ns, factory := range c.factoriesCrd {
		ings, err := factory.Traefik().V1().IngressRouteUDPs().Lister().List(labels.Everything())
		if err != nil {
			log.Error().Err(err).Msgf("Failed to list udp ingress routes in namespace %s", ns)
		}
		result = append(result, ings...)
	}

	return c.appendV1alpha1IngressRouteUDPs(result)
}

func (c *clientWrapper) GetMiddlewares() []*traefikv1.Middleware {
	var result []*traefikv1.Middleware

	for ns, factory := range c.factoriesCrd {
		middlewares, err := factory.Traefik().V1().Middlewares().Lister().List(labels.Everything())
		if err != nil {
			log.Error().Err(err).Msgf("Failed to list middlewares in namespace %s", ns)
		}
		result = append(result, middlewares...)
	}

	return c.appendV1alpha1Middlewares(result)
}

func (c *clientWrapper) GetMiddlewareTCPs() []*traefikv1.MiddlewareTCP {
	var result []*traefikv1.MiddlewareTCP

	for ns, factory := range c.factoriesCrd {
		middlewares, err := factory.Traefik().V1().MiddlewareTCPs().Lister().List(labels.Everything())
		if err != nil {
			log.Error().Err(err).Msgf("Failed to list TCP middlewares in namespace %s", ns)
		}
		result = append(result, middlewares...)
	}

	return c.appendV1alpha1MiddlewareTCPs(result)
}

// GetTraefikService returns the named service from the given namespace.
func (c *clientWrapper) GetTraefikService(namespace, name string) (*traefikv1.TraefikService, bool, error) {
	if !c.isWatchedNamespace(namespace) {
		return nil, false, fmt.Errorf("failed to get service %s/%s: namespace is not within watched namespaces", namespace, name)
	}

	service, err := c.factoriesCrd[c.lookupNamespace(namespace)].Traefik().V1().TraefikServices().Lister().TraefikServices(namespace).Get(name)
	exist, err := translateNotFoundError(err)

	if !exist {
		return c.getV1alpha1TraefikService(namespace, name)
	}

	return service, exist, err
}

func (c *clientWrapper) GetTraefikServices() []*traefikv1.TraefikService {
	var result []*traefikv1.TraefikService

	for ns, factory := range c.factoriesCrd {
		traefikServices, err := factory.Traefik().V1().TraefikServices().Lister().List(labels.Everything())
		if err != nil {
			log.Error().Err(err).Msgf("Failed to list Traefik services in namespace %s", ns)
		}
		result = append(result, traefikServices...)
	}

	return c.appendV1alpha1TraefikServices(result)
}

// GetServersTransports returns all ServersTransport.
func (c *clientWrapper) GetServersTransports() []*traefikv1.ServersTransport {
	var result []*traefikv1.ServersTransport

	for ns, factory := range c.factoriesCrd {
		serversTransports, err := factory.Traefik().V1().ServersTransports().Lister().List(labels.Everything())
		if err != nil {
			log.Error().Err(err).Str("namespace", ns).Msg("Failed to list servers transport in namespace")
		}
		result = append(result, serversTransports...)
	}

	return c.appendV1alpha1ServersTransport(result)
}

// GetServersTransportTCPs returns all ServersTransportTCP.
func (c *clientWrapper) GetServersTransportTCPs() []*traefikv1.ServersTransportTCP {
	var result []*traefikv1.ServersTransportTCP

	for ns, factory := range c.factoriesCrd {
		serversTransports, err := factory.Traefik().V1().ServersTransportTCPs().Lister().List(labels.Everything())
		if err != nil {
			log.Error().Err(err).Str("namespace", ns).Msg("Failed to list servers transport TCP in namespace")
		}
		result = append(result, serversTransports...)
	}

	return result
}

// GetTLSOptions returns all TLS options.
func (c *clientWrapper) GetTLSOptions() []*traefikv1.TLSOption {
	var result []*traefikv1.TLSOption

	for ns, factory := range c.factoriesCrd {
		options, err := factory.Traefik().V1().TLSOptions().Lister().List(labels.Everything())
		if err != nil {
			log.Error().Err(err).Msgf("Failed to list tls options in namespace %s", ns)
		}
		result = append(result, options...)
	}

	return c.appendV1alpha1TLSOptions(result)
}

// GetTLSStores returns all TLS stores.
func (c *clientWrapper) GetTLSStores() []*traefikv1.TLSStore {
	var result []*traefikv1.TLSStore

	for ns, factory := range c.factoriesCrd {
		stores, err := factory.Traefik().V1().TLSStores().Lister().List(labels.Everything())
		if err != nil {
			log.Error().Err(err).Msgf("Failed to list tls stores in namespace %s", ns)
		}
		result = append(result, stores...)
	}

	return c.appendV1alpha1TLSStores(result)
}

// GetService returns the named service from the given namespace.
func (c *clientWrapper) GetService(namespace, name string) (*corev1.Service, bool, error) {
	if !c.isWatchedNamespace(namespace) {
		return nil, false, fmt.Errorf("failed to get service %s/%s: namespace is not within watched namespaces", namespace, name)
	}

	service, err := c.factoriesKube[c.lookupNamespace(namespace)].Core().V1().Services().Lister().Services(namespace).Get(name)
	exist, err := translateNotFoundError(err)
	return service, exist, err
}

// GetEndpoints returns the named endpoints from the given namespace.
func (c *clientWrapper) GetEndpoints(namespace, name string) (*corev1.Endpoints, bool, error) {
	if !c.isWatchedNamespace(namespace) {
		return nil, false, fmt.Errorf("failed to get endpoints %s/%s: namespace is not within watched namespaces", namespace, name)
	}

	endpoint, err := c.factoriesKube[c.lookupNamespace(namespace)].Core().V1().Endpoints().Lister().Endpoints(namespace).Get(name)
	exist, err := translateNotFoundError(err)
	return endpoint, exist, err
}

// GetSecret returns the named secret from the given namespace.
func (c *clientWrapper) GetSecret(namespace, name string) (*corev1.Secret, bool, error) {
	if !c.isWatchedNamespace(namespace) {
		return nil, false, fmt.Errorf("failed to get secret %s/%s: namespace is not within watched namespaces", namespace, name)
	}

	secret, err := c.factoriesSecret[c.lookupNamespace(namespace)].Core().V1().Secrets().Lister().Secrets(namespace).Get(name)
	exist, err := translateNotFoundError(err)
	return secret, exist, err
}

// lookupNamespace returns the lookup namespace key for the given namespace.
// When listening on all namespaces, it returns the client-go identifier ("")
// for all-namespaces. Otherwise, it returns the given namespace.
// The distinction is necessary because we index all informers on the special
// identifier iff all-namespaces are requested but receive specific namespace
// identifiers from the Kubernetes API, so we have to bridge this gap.
func (c *clientWrapper) lookupNamespace(ns string) string {
	if c.isNamespaceAll {
		return metav1.NamespaceAll
	}
	return ns
}

// isWatchedNamespace checks to ensure that the namespace is being watched before we request
// it to ensure we don't panic by requesting an out-of-watch object.
func (c *clientWrapper) isWatchedNamespace(ns string) bool {
	if c.isNamespaceAll {
		return true
	}
	for _, watchedNamespace := range c.watchedNamespaces {
		if watchedNamespace == ns {
			return true
		}
	}
	return false
}

// translateNotFoundError will translate a "not found" error to a boolean return
// value which indicates if the resource exists and a nil error.
func translateNotFoundError(err error) (bool, error) {
	if kerror.IsNotFound(err) {
		return false, nil
	}
	return err == nil, err
}
