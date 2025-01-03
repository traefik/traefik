package crd

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/rs/zerolog/log"
	traefikclientset "github.com/traefik/traefik/v3/pkg/provider/knative/crd/generated/clientset/versioned"
	"github.com/traefik/traefik/v3/pkg/provider/knative/crd/generated/informers/externalversions"
	traefikinformers "github.com/traefik/traefik/v3/pkg/provider/knative/crd/generated/informers/externalversions"
	traefikv1alpha1 "github.com/traefik/traefik/v3/pkg/provider/knative/crd/traefikio/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	kerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kinformers "k8s.io/client-go/informers"
	kclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	knativenetworkingv1alpha1 "knative.dev/networking/pkg/apis/networking/v1alpha1"
	knativenetworkingclientset "knative.dev/networking/pkg/client/clientset/versioned"
	knativenetworkinginformers "knative.dev/networking/pkg/client/informers/externalversions"
)

const resyncPeriod = 10 * time.Minute

type resourceEventHandler struct {
	ev chan<- interface{}
}

func (reh *resourceEventHandler) OnAdd(obj interface{}, isInInitialList bool) {
	eventHandlerFunc(reh.ev, obj)
}

func (reh *resourceEventHandler) OnUpdate(oldObj, newObj interface{}) {
	eventHandlerFunc(reh.ev, newObj)
}

func (reh *resourceEventHandler) OnDelete(obj interface{}) {
	eventHandlerFunc(reh.ev, obj)
}

// Client is a client for the Provider master.
// WatchAll starts the watch of the Provider resources and updates the stores.
// The stores can then be accessed via the Get* functions.
type Client interface {
	WatchAll(namespaces []string, stopCh <-chan struct{}) (<-chan interface{}, error)

	GetKnativeIngressRoute(namespace, name string) (*knativenetworkingv1alpha1.Ingress, bool, error)
	UpdateKnativeIngressStatus(ingress *knativenetworkingv1alpha1.Ingress) error
	GetKnativeIngressRoutes() []*knativenetworkingv1alpha1.Ingress
	GetServerlessService(namespace, name string) (*knativenetworkingv1alpha1.ServerlessService, bool, error)
	GetService(namespace, name string) (*corev1.Service, bool, error)
	GetSecret(namespace, name string) (*corev1.Secret, bool, error)
	GetEndpoints(namespace, name string) (*corev1.Endpoints, bool, error)

	GetIngressRoutes() []*traefikv1alpha1.IngressRouteKnative
}

// TODO: add tests for the clientWrapper (and its methods) itself.
type clientWrapper struct {
	csKnativeNetworking *knativenetworkingclientset.Clientset
	csKube              *kclientset.Clientset
	csCrd               traefikclientset.Interface

	factoriesKnativeNetworking map[string]knativenetworkinginformers.SharedInformerFactory
	factoriesKube              map[string]kinformers.SharedInformerFactory
	factoriesCrd               map[string]externalversions.SharedInformerFactory

	labelSelector labels.Selector

	isNamespaceAll    bool
	watchedNamespaces []string
}

func createClientFromConfig(c *rest.Config) (*clientWrapper, error) {
	csKnativeNetworking, err := knativenetworkingclientset.NewForConfig(c)
	if err != nil {
		return nil, err
	}

	csCrd, err := traefikclientset.NewForConfig(c)
	if err != nil {
		return nil, err
	}

	csKube, err := kclientset.NewForConfig(c)
	if err != nil {
		return nil, err
	}

	return newClientImpl(csKnativeNetworking, csKube, csCrd), nil
}

func newClientImpl(csKnativeNetworking *knativenetworkingclientset.Clientset, csKube *kclientset.Clientset,
	csCrd traefikclientset.Interface,
) *clientWrapper {
	return &clientWrapper{
		csKnativeNetworking:        csKnativeNetworking,
		csKube:                     csKube,
		csCrd:                      csCrd,
		factoriesKnativeNetworking: make(map[string]knativenetworkinginformers.SharedInformerFactory),
		factoriesKube:              make(map[string]kinformers.SharedInformerFactory),
		factoriesCrd:               make(map[string]traefikinformers.SharedInformerFactory),
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
func newExternalClusterClient(endpoint, token, caFilePath string) (*clientWrapper, error) {
	if endpoint == "" {
		return nil, errors.New("endpoint missing for external cluster client")
	}

	config := &rest.Config{
		Host:        endpoint,
		BearerToken: token,
	}

	if caFilePath != "" {
		caData, err := ioutil.ReadFile(caFilePath)
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
	eventHandler := c.newResourceEventHandler(eventCh)

	if len(namespaces) == 0 {
		namespaces = []string{metav1.NamespaceAll}
		c.isNamespaceAll = true
	}
	c.watchedNamespaces = namespaces

	matchesLabelSelector := func(opts *metav1.ListOptions) {
		opts.LabelSelector = c.labelSelector.String()
	}
	for _, ns := range namespaces {
		factory := knativenetworkinginformers.NewSharedInformerFactoryWithOptions(c.csKnativeNetworking, resyncPeriod, knativenetworkinginformers.WithNamespace(ns), knativenetworkinginformers.WithTweakListOptions(matchesLabelSelector))
		_, err := factory.Networking().V1alpha1().Ingresses().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		_, err = factory.Networking().V1alpha1().ServerlessServices().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}

		factoryCrd := traefikinformers.NewSharedInformerFactoryWithOptions(c.csCrd, resyncPeriod,
			traefikinformers.WithNamespace(ns), traefikinformers.WithTweakListOptions(matchesLabelSelector))
		_, err = factoryCrd.Traefik().V1alpha1().IngressRouteKnatives().Informer().AddEventHandler(eventHandler)
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
		_, err = factoryKube.Core().V1().Secrets().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}

		c.factoriesKube[ns] = factoryKube
		c.factoriesCrd[ns] = factoryCrd
		c.factoriesKnativeNetworking[ns] = factory
	}

	for _, ns := range namespaces {
		c.factoriesKnativeNetworking[ns].Start(stopCh)
		c.factoriesKube[ns].Start(stopCh)
		c.factoriesCrd[ns].Start(stopCh)
	}

	for _, ns := range namespaces {
		for t, ok := range c.factoriesCrd[ns].WaitForCacheSync(stopCh) {
			if !ok {
				return nil, fmt.Errorf("timed out waiting for controller caches to sync %s in namespace %q", t.String(), ns)
			}
		}
		for t, ok := range c.factoriesKnativeNetworking[ns].WaitForCacheSync(stopCh) {
			if !ok {
				return nil, fmt.Errorf("timed out waiting for controller caches to sync %s in namespace %q", t.String(), ns)
			}
		}
		for t, ok := range c.factoriesKube[ns].WaitForCacheSync(stopCh) {
			if !ok {
				return nil, fmt.Errorf("timed out waiting for controller caches to sync %s in namespace %q", t.String(), ns)
			}
		}
	}

	return eventCh, nil
}

func (c *clientWrapper) GetKnativeIngressRoutes() []*knativenetworkingv1alpha1.Ingress {
	var result []*knativenetworkingv1alpha1.Ingress

	for ns, factory := range c.factoriesKnativeNetworking {
		ings, err := factory.Networking().V1alpha1().Ingresses().Lister().List(labels.Everything()) // todo: label selector
		if err != nil {
			log.Error().Msgf("Failed to list ingresses in namespace %s: %s", ns, err)
		}
		result = append(result, ings...)
	}

	return result
}

func (c *clientWrapper) UpdateKnativeIngressStatus(ingressRoute *knativenetworkingv1alpha1.Ingress) error {
	_, err := c.csKnativeNetworking.NetworkingV1alpha1().Ingresses(ingressRoute.Namespace).UpdateStatus(context.TODO(), ingressRoute, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update knative ingress status %s/%s: %w", ingressRoute.Namespace,
			ingressRoute.Name, err)
	}
	log.Info().Msgf("Updated status on knative ingress %s/%s", ingressRoute.Namespace, ingressRoute.Name)
	return err
}

func (c *clientWrapper) GetServerlessService(namespace, name string) (*knativenetworkingv1alpha1.ServerlessService, bool, error) {
	if !c.isWatchedNamespace(namespace) {
		return nil, false, fmt.Errorf("failed to get service %s/%s: namespace is not within watched namespaces", namespace, name)
	}

	service, err := c.factoriesKnativeNetworking[c.lookupNamespace(namespace)].Networking().V1alpha1().ServerlessServices().Lister().ServerlessServices(namespace).Get(name)
	exist, err := translateNotFoundError(err)
	return service, exist, err
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

	secret, err := c.factoriesKube[c.lookupNamespace(namespace)].Core().V1().Secrets().Lister().Secrets(namespace).Get(name)
	exist, err := translateNotFoundError(err)
	return secret, exist, err
}

func (c *clientWrapper) GetKnativeIngressRoute(namespace, name string) (*knativenetworkingv1alpha1.Ingress, bool,
	error,
) {
	if !c.isWatchedNamespace(namespace) {
		return nil, false, fmt.Errorf("failed to get ingress %s/%s: namespace is not within watched namespaces", namespace, name)
	}
	ingress, err := c.factoriesKnativeNetworking[c.lookupNamespace(namespace)].Networking().V1alpha1().Ingresses().Lister().Ingresses(namespace).Get(name)
	exist, err := translateNotFoundError(err)
	return ingress, exist, err
}

func (c *clientWrapper) GetIngressRoutes() []*traefikv1alpha1.IngressRouteKnative {
	var result []*traefikv1alpha1.IngressRouteKnative

	for ns, factory := range c.factoriesCrd {
		ings, err := factory.Traefik().V1alpha1().IngressRouteKnatives().Lister().List(labels.Everything())
		if err != nil {
			log.Error().Err(err).Msgf("Failed to list ingress routes in namespace %s", ns)
		}
		result = append(result, ings...)
	}

	return result
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

func (c *clientWrapper) newResourceEventHandler(events chan<- interface{}) cache.ResourceEventHandler {
	return &cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			// Ignore Ingresses that do not match our custom label selector.
			// switch v := obj.(type) {
			// default:
			// 	return true
			// }
			return true
		},
		Handler: &resourceEventHandler{ev: events},
	}
}

// eventHandlerFunc will pass the obj on to the events channel or drop it.
// This is so passing the events along won't block in the case of high volume.
// The events are only used for signaling anyway so dropping a few is ok.
func eventHandlerFunc(events chan<- interface{}, obj interface{}) {
	select {
	case events <- obj:
	default:
	}
}

// translateNotFoundError will translate a "not found" error to a boolean return
// value which indicates if the resource exists and a nil error.
func translateNotFoundError(err error) (bool, error) {
	if kerror.IsNotFound(err) {
		return false, nil
	}
	return err == nil, err
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
