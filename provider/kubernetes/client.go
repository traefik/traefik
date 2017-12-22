package kubernetes

import (
	"errors"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"github.com/containous/traefik/safe"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/fields"
	"k8s.io/client-go/pkg/labels"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/watch"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

const resyncPeriod = 10 * time.Minute

const (
	kindIngresses = "ingresses"
	kindServices  = "services"
	kindEndpoints = "endpoints"
	kindSecrets   = "secrets"
)

type resourceEventHandler struct {
	ev chan<- interface{}
}

func (reh *resourceEventHandler) OnAdd(obj interface{}) {
	eventHandlerFunc(reh.ev, obj)
}

func (reh *resourceEventHandler) OnUpdate(oldObj, newObj interface{}) {
	eventHandlerFunc(reh.ev, newObj)
}

func (reh *resourceEventHandler) OnDelete(obj interface{}) {
	eventHandlerFunc(reh.ev, obj)
}

type informerManager struct {
	informers []cache.SharedInformer
	syncFuncs []cache.InformerSynced
}

func (im *informerManager) extend(informer cache.SharedInformer, withSyncFunc bool) {
	im.informers = append(im.informers, informer)
	if withSyncFunc {
		im.syncFuncs = append(im.syncFuncs, informer.HasSynced)
	}
}

// Client is a client for the Provider master.
// WatchAll starts the watch of the Provider resources and updates the stores.
// The stores can then be accessed via the Get* functions.
type Client interface {
	WatchAll(namespaces Namespaces, labelSelector string, stopCh <-chan struct{}) (<-chan interface{}, error)
	GetIngresses() []*v1beta1.Ingress
	GetService(namespace, name string) (*v1.Service, bool, error)
	GetSecret(namespace, name string) (*v1.Secret, bool, error)
	GetEndpoints(namespace, name string) (*v1.Endpoints, bool, error)
}

type clientImpl struct {
	clientset      *kubernetes.Clientset
	ingStores      []cache.Store
	svcStores      map[string]cache.Store
	epStores       map[string]cache.Store
	secStores      map[string]cache.Store
	isNamespaceAll bool
}

func newClientImpl(clientset *kubernetes.Clientset) Client {
	return &clientImpl{
		clientset: clientset,
		ingStores: []cache.Store{},
		svcStores: map[string]cache.Store{},
		epStores:  map[string]cache.Store{},
		secStores: map[string]cache.Store{},
	}
}

// NewInClusterClient returns a new Provider client that is expected to run
// inside the cluster.
func NewInClusterClient(endpoint string) (Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster configuration: %s", err)
	}

	if endpoint != "" {
		config.Host = endpoint
	}

	return createClientFromConfig(config)
}

// NewExternalClusterClient returns a new Provider client that may run outside
// of the cluster.
// The endpoint parameter must not be empty.
func NewExternalClusterClient(endpoint, token, caFilePath string) (Client, error) {
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
			return nil, fmt.Errorf("failed to read CA file %s: %s", caFilePath, err)
		}

		config.TLSClientConfig = rest.TLSClientConfig{CAData: caData}
	}

	return createClientFromConfig(config)
}

func createClientFromConfig(c *rest.Config) (Client, error) {
	clientset, err := kubernetes.NewForConfig(c)
	if err != nil {
		return nil, err
	}

	return newClientImpl(clientset), nil
}

// WatchAll starts namespace-specific controllers for all relevant kinds.
func (c *clientImpl) WatchAll(namespaces Namespaces, labelSelector string, stopCh <-chan struct{}) (<-chan interface{}, error) {
	eventCh := make(chan interface{}, 1)

	kubeLabelSelector, err := labels.Parse(labelSelector)
	if err != nil {
		return nil, err
	}

	if len(namespaces) == 0 {
		namespaces = Namespaces{api.NamespaceAll}
		c.isNamespaceAll = true
	}

	var informManager informerManager
	for _, ns := range namespaces {
		ns := ns
		informManager.extend(c.WatchIngresses(ns, kubeLabelSelector, eventCh), true)
		informManager.extend(c.WatchObjects(ns, kindServices, &v1.Service{}, c.svcStores, eventCh), true)
		informManager.extend(c.WatchObjects(ns, kindEndpoints, &v1.Endpoints{}, c.epStores, eventCh), true)
		// Do not wait for the Secrets store to get synced since we cannot rely on
		// users having granted RBAC permissions for this object.
		// https://github.com/containous/traefik/issues/1784 should improve the
		// situation here in the future.
		informManager.extend(c.WatchObjects(ns, kindSecrets, &v1.Secret{}, c.secStores, eventCh), false)
	}

	var wg sync.WaitGroup
	for _, informer := range informManager.informers {
		informer := informer
		safe.Go(func() {
			wg.Add(1)
			informer.Run(stopCh)
			wg.Done()
		})
	}

	if !cache.WaitForCacheSync(stopCh, informManager.syncFuncs...) {
		return nil, fmt.Errorf("timed out waiting for controller caches to sync")
	}

	safe.Go(func() {
		<-stopCh
		wg.Wait()
		close(eventCh)
	})

	return eventCh, nil
}

// WatchIngresses sets up a watch on Ingress objects and returns a corresponding shared informer.
func (c *clientImpl) WatchIngresses(namespace string, labelSelector labels.Selector, watchCh chan<- interface{}) cache.SharedInformer {
	listWatch := newListWatchFromClientWithLabelSelector(
		c.clientset.ExtensionsV1beta1().RESTClient(),
		kindIngresses,
		namespace,
		fields.Everything(),
		labelSelector)

	informer := loadInformer(listWatch, &v1beta1.Ingress{}, watchCh)
	c.ingStores = append(c.ingStores, informer.GetStore())
	return informer
}

// WatchObjects sets up a watch on objects and returns a corresponding shared informer.
func (c *clientImpl) WatchObjects(namespace, kind string, object runtime.Object, storeMap map[string]cache.Store, watchCh chan<- interface{}) cache.SharedInformer {
	listWatch := cache.NewListWatchFromClient(
		c.clientset.CoreV1().RESTClient(),
		kind,
		namespace,
		fields.Everything())

	informer := loadInformer(listWatch, object, watchCh)
	storeMap[namespace] = informer.GetStore()
	return informer
}

func loadInformer(listWatch cache.ListerWatcher, object runtime.Object, watchCh chan<- interface{}) cache.SharedInformer {
	informer := cache.NewSharedInformer(
		listWatch,
		object,
		resyncPeriod,
	)

	if err := informer.AddEventHandler(newResourceEventHandler(watchCh)); err != nil {
		// This should only ever fail if we add an event handler after the
		// informer has been started already, which would be a programming bug.
		panic(err)
	}

	return informer
}

// GetIngresses returns all Ingresses for observed namespaces in the cluster.
func (c *clientImpl) GetIngresses() []*v1beta1.Ingress {
	var result []*v1beta1.Ingress

	for _, store := range c.ingStores {
		for _, obj := range store.List() {
			ing := obj.(*v1beta1.Ingress)
			result = append(result, ing)
		}
	}

	return result
}

// GetService returns the named service from the given namespace.
func (c *clientImpl) GetService(namespace, name string) (*v1.Service, bool, error) {
	var service *v1.Service
	item, exists, err := c.svcStores[c.lookupNamespace(namespace)].GetByKey(namespace + "/" + name)
	if item != nil {
		service = item.(*v1.Service)
	}

	return service, exists, err
}

// GetEndpoints returns the named endpoints from the given namespace.
func (c *clientImpl) GetEndpoints(namespace, name string) (*v1.Endpoints, bool, error) {
	var endpoint *v1.Endpoints
	item, exists, err := c.epStores[c.lookupNamespace(namespace)].GetByKey(namespace + "/" + name)

	if item != nil {
		endpoint = item.(*v1.Endpoints)
	}

	return endpoint, exists, err
}

// GetSecret returns the named secret from the given namespace.
func (c *clientImpl) GetSecret(namespace, name string) (*v1.Secret, bool, error) {
	var secret *v1.Secret
	item, exists, err := c.secStores[c.lookupNamespace(namespace)].GetByKey(namespace + "/" + name)
	if err == nil && item != nil {
		secret = item.(*v1.Secret)
	}

	return secret, exists, err
}

// lookupNamespace returns the lookup namespace key for the given namespace.
// When listening on all namespaces, it returns the client-go identifier ("")
// for all-namespaces. Otherwise, it returns the given namespace.
// The distinction is necessary because we index all informers on the special
// identifier iff all-namespaces are requested but receive specific namespace
// identifiers from the Kubernetes API, so we have to bridge this gap.
func (c *clientImpl) lookupNamespace(ns string) string {
	if c.isNamespaceAll {
		return api.NamespaceAll
	}
	return ns
}

// newListWatchFromClientWithLabelSelector creates a new ListWatch from the given parameters.
// It extends cache.NewListWatchFromClient to support label selectors.
func newListWatchFromClientWithLabelSelector(c cache.Getter, resource string, namespace string, fieldSelector fields.Selector, labelSelector labels.Selector) *cache.ListWatch {
	listFunc := func(options api.ListOptions) (runtime.Object, error) {
		return c.Get().
			Namespace(namespace).
			Resource(resource).
			VersionedParams(&options, api.ParameterCodec).
			FieldsSelectorParam(fieldSelector).
			LabelsSelectorParam(labelSelector).
			Do().
			Get()
	}
	watchFunc := func(options api.ListOptions) (watch.Interface, error) {
		return c.Get().
			Prefix("watch").
			Namespace(namespace).
			Resource(resource).
			VersionedParams(&options, api.ParameterCodec).
			FieldsSelectorParam(fieldSelector).
			LabelsSelectorParam(labelSelector).
			Watch()
	}
	return &cache.ListWatch{ListFunc: listFunc, WatchFunc: watchFunc}
}

func newResourceEventHandler(events chan<- interface{}) cache.ResourceEventHandler {
	return &resourceEventHandler{events}
}

// eventHandlerFunc will pass the obj on to the events channel or drop it.
// This is so passing the events along won't block in the case of high volume.
// The events are only used for signalling anyway so dropping a few is ok.
func eventHandlerFunc(events chan<- interface{}, obj interface{}) {
	select {
	case events <- obj:
	default:
	}
}
