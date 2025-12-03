package crd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"time"

	baqupv1alpha1 "github.com/baqupio/baqup/v3/pkg/provider/kubernetes/crd/baqupio/v1alpha1"
	baqupclientset "github.com/baqupio/baqup/v3/pkg/provider/kubernetes/crd/generated/clientset/versioned"
	baqupinformers "github.com/baqupio/baqup/v3/pkg/provider/kubernetes/crd/generated/informers/externalversions"
	"github.com/baqupio/baqup/v3/pkg/provider/kubernetes/k8s"
	"github.com/baqupio/baqup/v3/pkg/types"
	"github.com/baqupio/baqup/v3/pkg/version"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	kerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
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
	GetIngressRoutes() []*baqupv1alpha1.IngressRoute
	GetIngressRouteTCPs() []*baqupv1alpha1.IngressRouteTCP
	GetIngressRouteUDPs() []*baqupv1alpha1.IngressRouteUDP
	GetMiddlewares() []*baqupv1alpha1.Middleware
	GetMiddlewareTCPs() []*baqupv1alpha1.MiddlewareTCP
	GetBaqupService(namespace, name string) (*baqupv1alpha1.BaqupService, bool, error)
	GetBaqupServices() []*baqupv1alpha1.BaqupService
	GetTLSOptions() []*baqupv1alpha1.TLSOption
	GetServersTransports() []*baqupv1alpha1.ServersTransport
	GetServersTransportTCPs() []*baqupv1alpha1.ServersTransportTCP
	GetTLSStores() []*baqupv1alpha1.TLSStore
	GetService(namespace, name string) (*corev1.Service, bool, error)
	GetSecret(namespace, name string) (*corev1.Secret, bool, error)
	GetEndpointSlicesForService(namespace, serviceName string) ([]*discoveryv1.EndpointSlice, error)
	GetNodes() ([]*corev1.Node, bool, error)
	GetConfigMap(namespace, name string) (*corev1.ConfigMap, bool, error)
}

// TODO: add tests for the clientWrapper (and its methods) itself.
type clientWrapper struct {
	csCrd  baqupclientset.Interface
	csKube kclientset.Interface

	clusterScopeFactory         kinformers.SharedInformerFactory
	disableClusterScopeInformer bool

	factoriesCrd    map[string]baqupinformers.SharedInformerFactory
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

	csCrd, err := baqupclientset.NewForConfig(c)
	if err != nil {
		return nil, err
	}

	csKube, err := kclientset.NewForConfig(c)
	if err != nil {
		return nil, err
	}

	return newClientImpl(csKube, csCrd), nil
}

func newClientImpl(csKube kclientset.Interface, csCrd baqupclientset.Interface) *clientWrapper {
	return &clientWrapper{
		csCrd:           csCrd,
		csKube:          csKube,
		factoriesCrd:    make(map[string]baqupinformers.SharedInformerFactory),
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

// newExternalClusterClient returns a new Provider client that may run outside
// of the cluster.
// The endpoint parameter must not be empty.
func newExternalClusterClient(endpoint, caFilePath string, token types.FileOrContent) (*clientWrapper, error) {
	if endpoint == "" {
		return nil, errors.New("endpoint missing for external cluster client")
	}

	tokenData, err := token.Read()
	if err != nil {
		return nil, fmt.Errorf("read token: %w", err)
	}

	config := &rest.Config{
		Host:        endpoint,
		BearerToken: string(tokenData),
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
		factoryCrd := baqupinformers.NewSharedInformerFactoryWithOptions(c.csCrd, resyncPeriod, baqupinformers.WithNamespace(ns), baqupinformers.WithTweakListOptions(matchesLabelSelector))
		_, err := factoryCrd.Baqup().V1alpha1().IngressRoutes().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		_, err = factoryCrd.Baqup().V1alpha1().Middlewares().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		_, err = factoryCrd.Baqup().V1alpha1().MiddlewareTCPs().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		_, err = factoryCrd.Baqup().V1alpha1().IngressRouteTCPs().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		_, err = factoryCrd.Baqup().V1alpha1().IngressRouteUDPs().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		_, err = factoryCrd.Baqup().V1alpha1().TLSOptions().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		_, err = factoryCrd.Baqup().V1alpha1().ServersTransports().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		_, err = factoryCrd.Baqup().V1alpha1().ServersTransportTCPs().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		_, err = factoryCrd.Baqup().V1alpha1().TLSStores().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		_, err = factoryCrd.Baqup().V1alpha1().BaqupServices().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}

		factoryKube := kinformers.NewSharedInformerFactoryWithOptions(c.csKube, resyncPeriod, kinformers.WithNamespace(ns))
		_, err = factoryKube.Core().V1().Services().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		_, err = factoryKube.Discovery().V1().EndpointSlices().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		_, err = factoryKube.Core().V1().ConfigMaps().Informer().AddEventHandler(eventHandler)
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

	if !c.disableClusterScopeInformer {
		c.clusterScopeFactory = kinformers.NewSharedInformerFactory(c.csKube, resyncPeriod)
		_, err := c.clusterScopeFactory.Core().V1().Nodes().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}

		c.clusterScopeFactory.Start(stopCh)

		for t, ok := range c.clusterScopeFactory.WaitForCacheSync(stopCh) {
			if !ok {
				return nil, fmt.Errorf("timed out waiting for controller caches to sync %s", t.String())
			}
		}
	}

	return eventCh, nil
}

func (c *clientWrapper) GetIngressRoutes() []*baqupv1alpha1.IngressRoute {
	var result []*baqupv1alpha1.IngressRoute

	for ns, factory := range c.factoriesCrd {
		ings, err := factory.Baqup().V1alpha1().IngressRoutes().Lister().List(labels.Everything())
		if err != nil {
			log.Error().Err(err).Msgf("Failed to list ingress routes in namespace %s", ns)
		}
		result = append(result, ings...)
	}

	return result
}

func (c *clientWrapper) GetIngressRouteTCPs() []*baqupv1alpha1.IngressRouteTCP {
	var result []*baqupv1alpha1.IngressRouteTCP

	for ns, factory := range c.factoriesCrd {
		ings, err := factory.Baqup().V1alpha1().IngressRouteTCPs().Lister().List(labels.Everything())
		if err != nil {
			log.Error().Err(err).Msgf("Failed to list tcp ingress routes in namespace %s", ns)
		}
		result = append(result, ings...)
	}

	return result
}

func (c *clientWrapper) GetIngressRouteUDPs() []*baqupv1alpha1.IngressRouteUDP {
	var result []*baqupv1alpha1.IngressRouteUDP

	for ns, factory := range c.factoriesCrd {
		ings, err := factory.Baqup().V1alpha1().IngressRouteUDPs().Lister().List(labels.Everything())
		if err != nil {
			log.Error().Err(err).Msgf("Failed to list udp ingress routes in namespace %s", ns)
		}
		result = append(result, ings...)
	}

	return result
}

func (c *clientWrapper) GetMiddlewares() []*baqupv1alpha1.Middleware {
	var result []*baqupv1alpha1.Middleware

	for ns, factory := range c.factoriesCrd {
		middlewares, err := factory.Baqup().V1alpha1().Middlewares().Lister().List(labels.Everything())
		if err != nil {
			log.Error().Err(err).Msgf("Failed to list middlewares in namespace %s", ns)
		}
		result = append(result, middlewares...)
	}

	return result
}

func (c *clientWrapper) GetMiddlewareTCPs() []*baqupv1alpha1.MiddlewareTCP {
	var result []*baqupv1alpha1.MiddlewareTCP

	for ns, factory := range c.factoriesCrd {
		middlewares, err := factory.Baqup().V1alpha1().MiddlewareTCPs().Lister().List(labels.Everything())
		if err != nil {
			log.Error().Err(err).Msgf("Failed to list TCP middlewares in namespace %s", ns)
		}
		result = append(result, middlewares...)
	}

	return result
}

// GetBaqupService returns the named service from the given namespace.
func (c *clientWrapper) GetBaqupService(namespace, name string) (*baqupv1alpha1.BaqupService, bool, error) {
	if !c.isWatchedNamespace(namespace) {
		return nil, false, fmt.Errorf("failed to get service %s/%s: namespace is not within watched namespaces", namespace, name)
	}

	service, err := c.factoriesCrd[c.lookupNamespace(namespace)].Baqup().V1alpha1().BaqupServices().Lister().BaqupServices(namespace).Get(name)
	exist, err := translateNotFoundError(err)

	return service, exist, err
}

func (c *clientWrapper) GetBaqupServices() []*baqupv1alpha1.BaqupService {
	var result []*baqupv1alpha1.BaqupService

	for ns, factory := range c.factoriesCrd {
		baqupServices, err := factory.Baqup().V1alpha1().BaqupServices().Lister().List(labels.Everything())
		if err != nil {
			log.Error().Err(err).Msgf("Failed to list Baqup services in namespace %s", ns)
		}
		result = append(result, baqupServices...)
	}

	return result
}

// GetServersTransports returns all ServersTransport.
func (c *clientWrapper) GetServersTransports() []*baqupv1alpha1.ServersTransport {
	var result []*baqupv1alpha1.ServersTransport

	for ns, factory := range c.factoriesCrd {
		serversTransports, err := factory.Baqup().V1alpha1().ServersTransports().Lister().List(labels.Everything())
		if err != nil {
			log.Error().Err(err).Str("namespace", ns).Msg("Failed to list servers transport in namespace")
		}
		result = append(result, serversTransports...)
	}

	return result
}

// GetServersTransportTCPs returns all ServersTransportTCP.
func (c *clientWrapper) GetServersTransportTCPs() []*baqupv1alpha1.ServersTransportTCP {
	var result []*baqupv1alpha1.ServersTransportTCP

	for ns, factory := range c.factoriesCrd {
		serversTransports, err := factory.Baqup().V1alpha1().ServersTransportTCPs().Lister().List(labels.Everything())
		if err != nil {
			log.Error().Err(err).Str("namespace", ns).Msg("Failed to list servers transport TCP in namespace")
		}
		result = append(result, serversTransports...)
	}

	return result
}

// GetTLSOptions returns all TLS options.
func (c *clientWrapper) GetTLSOptions() []*baqupv1alpha1.TLSOption {
	var result []*baqupv1alpha1.TLSOption

	for ns, factory := range c.factoriesCrd {
		options, err := factory.Baqup().V1alpha1().TLSOptions().Lister().List(labels.Everything())
		if err != nil {
			log.Error().Err(err).Msgf("Failed to list tls options in namespace %s", ns)
		}
		result = append(result, options...)
	}

	return result
}

// GetTLSStores returns all TLS stores.
func (c *clientWrapper) GetTLSStores() []*baqupv1alpha1.TLSStore {
	var result []*baqupv1alpha1.TLSStore

	for ns, factory := range c.factoriesCrd {
		stores, err := factory.Baqup().V1alpha1().TLSStores().Lister().List(labels.Everything())
		if err != nil {
			log.Error().Err(err).Msgf("Failed to list tls stores in namespace %s", ns)
		}
		result = append(result, stores...)
	}

	return result
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

// GetEndpointSlicesForService returns the EndpointSlices for the given service name in the given namespace.
func (c *clientWrapper) GetEndpointSlicesForService(namespace, serviceName string) ([]*discoveryv1.EndpointSlice, error) {
	if !c.isWatchedNamespace(namespace) {
		return nil, fmt.Errorf("failed to get endpointslices for service %s/%s: namespace is not within watched namespaces", namespace, serviceName)
	}

	serviceLabelRequirement, err := labels.NewRequirement(discoveryv1.LabelServiceName, selection.Equals, []string{serviceName})
	if err != nil {
		return nil, fmt.Errorf("failed to create service label selector requirement: %w", err)
	}
	serviceSelector := labels.NewSelector()
	serviceSelector = serviceSelector.Add(*serviceLabelRequirement)

	return c.factoriesKube[c.lookupNamespace(namespace)].Discovery().V1().EndpointSlices().Lister().EndpointSlices(namespace).List(serviceSelector)
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

// GetConfigMap returns the named config map from the given namespace.
func (c *clientWrapper) GetConfigMap(namespace, name string) (*corev1.ConfigMap, bool, error) {
	if !c.isWatchedNamespace(namespace) {
		return nil, false, fmt.Errorf("failed to get config map %s/%s: namespace is not within watched namespaces", namespace, name)
	}

	configMap, err := c.factoriesKube[c.lookupNamespace(namespace)].Core().V1().ConfigMaps().Lister().ConfigMaps(namespace).Get(name)
	exist, err := translateNotFoundError(err)
	return configMap, exist, err
}

func (c *clientWrapper) GetNodes() ([]*corev1.Node, bool, error) {
	nodes, err := c.clusterScopeFactory.Core().V1().Nodes().Lister().List(labels.Everything())
	exist, err := translateNotFoundError(err)
	return nodes, exist, err
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

	return slices.Contains(c.watchedNamespaces, ns)
}

// translateNotFoundError will translate a "not found" error to a boolean return
// value which indicates if the resource exists and a nil error.
func translateNotFoundError(err error) (bool, error) {
	if kerror.IsNotFound(err) {
		return false, nil
	}
	return err == nil, err
}
