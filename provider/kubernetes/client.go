package kubernetes

import (
	"errors"
	"fmt"
	"io/ioutil"
	"time"

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

const resyncPeriod = time.Minute * 5

// Client is a client for the Provider master.
// WatchAll starts the watch of the Provider ressources and updates the stores.
// The stores can then be accessed via the Get* functions.
type Client interface {
	GetIngresses(namespaces Namespaces) []*v1beta1.Ingress
	GetService(namespace, name string) (*v1.Service, bool, error)
	GetEndpoints(namespace, name string) (*v1.Endpoints, bool, error)
	WatchAll(labelSelector string, stopCh <-chan struct{}) (<-chan interface{}, error)
}

type clientImpl struct {
	ingController *cache.Controller
	svcController *cache.Controller
	epController  *cache.Controller

	ingStore cache.Store
	svcStore cache.Store
	epStore  cache.Store

	clientset *kubernetes.Clientset
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

	return &clientImpl{
		clientset: clientset,
	}, nil
}

// GetIngresses returns all ingresses in the cluster
func (c *clientImpl) GetIngresses(namespaces Namespaces) []*v1beta1.Ingress {
	ingList := c.ingStore.List()
	result := make([]*v1beta1.Ingress, 0, len(ingList))

	for _, obj := range ingList {
		ingress := obj.(*v1beta1.Ingress)
		if HasNamespace(ingress, namespaces) {
			result = append(result, ingress)
		}
	}

	return result
}

// WatchIngresses starts the watch of Provider Ingresses resources and updates the corresponding store
func (c *clientImpl) WatchIngresses(labelSelector labels.Selector, watchCh chan<- interface{}, stopCh <-chan struct{}) {
	source := NewListWatchFromClient(
		c.clientset.ExtensionsV1beta1().RESTClient(),
		"ingresses",
		api.NamespaceAll,
		fields.Everything(),
		labelSelector)

	c.ingStore, c.ingController = cache.NewInformer(
		source,
		&v1beta1.Ingress{},
		resyncPeriod,
		newResourceEventHandlerFuncs(watchCh))
	go c.ingController.Run(stopCh)
}

// eventHandlerFunc will pass the obj on to the events channel or drop it
// This is so passing the events along won't block in the case of high volume
// The events are only used for signalling anyway so dropping a few is ok
func eventHandlerFunc(events chan<- interface{}, obj interface{}) {
	select {
	case events <- obj:
	default:
	}
}

func newResourceEventHandlerFuncs(events chan<- interface{}) cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { eventHandlerFunc(events, obj) },
		UpdateFunc: func(old, new interface{}) { eventHandlerFunc(events, new) },
		DeleteFunc: func(obj interface{}) { eventHandlerFunc(events, obj) },
	}
}

// GetService returns the named service from the named namespace
func (c *clientImpl) GetService(namespace, name string) (*v1.Service, bool, error) {
	var service *v1.Service
	item, exists, err := c.svcStore.GetByKey(namespace + "/" + name)
	if item != nil {
		service = item.(*v1.Service)
	}

	return service, exists, err
}

// WatchServices starts the watch of Provider Service resources and updates the corresponding store
func (c *clientImpl) WatchServices(watchCh chan<- interface{}, stopCh <-chan struct{}) {
	source := cache.NewListWatchFromClient(
		c.clientset.CoreV1().RESTClient(),
		"services",
		api.NamespaceAll,
		fields.Everything())

	c.svcStore, c.svcController = cache.NewInformer(
		source,
		&v1.Service{},
		resyncPeriod,
		newResourceEventHandlerFuncs(watchCh))
	go c.svcController.Run(stopCh)
}

// GetEndpoints returns the named Endpoints
// Endpoints have the same name as the coresponding service
func (c *clientImpl) GetEndpoints(namespace, name string) (*v1.Endpoints, bool, error) {
	var endpoint *v1.Endpoints
	item, exists, err := c.epStore.GetByKey(namespace + "/" + name)

	if item != nil {
		endpoint = item.(*v1.Endpoints)
	}

	return endpoint, exists, err
}

// WatchEndpoints starts the watch of Provider Endpoints resources and updates the corresponding store
func (c *clientImpl) WatchEndpoints(watchCh chan<- interface{}, stopCh <-chan struct{}) {
	source := cache.NewListWatchFromClient(
		c.clientset.CoreV1().RESTClient(),
		"endpoints",
		api.NamespaceAll,
		fields.Everything())

	c.epStore, c.epController = cache.NewInformer(
		source,
		&v1.Endpoints{},
		resyncPeriod,
		newResourceEventHandlerFuncs(watchCh))
	go c.epController.Run(stopCh)
}

// WatchAll returns events in the cluster and updates the stores via informer
// Filters ingresses by labelSelector
func (c *clientImpl) WatchAll(labelSelector string, stopCh <-chan struct{}) (<-chan interface{}, error) {
	watchCh := make(chan interface{}, 1)
	eventCh := make(chan interface{}, 1)

	kubeLabelSelector, err := labels.Parse(labelSelector)
	if err != nil {
		return nil, err
	}

	c.WatchIngresses(kubeLabelSelector, eventCh, stopCh)
	c.WatchServices(eventCh, stopCh)
	c.WatchEndpoints(eventCh, stopCh)

	go func() {
		defer close(watchCh)
		defer close(eventCh)

		for {
			select {
			case <-stopCh:
				return
			case event := <-eventCh:
				c.fireEvent(event, watchCh)
			}
		}
	}()

	return watchCh, nil
}

// fireEvent checks if all controllers have synced before firing
// Used after startup or a reconnect
func (c *clientImpl) fireEvent(event interface{}, eventCh chan interface{}) {
	if !c.ingController.HasSynced() || !c.svcController.HasSynced() || !c.epController.HasSynced() {
		return
	}
	eventHandlerFunc(eventCh, event)
}

// HasNamespace checks if the ingress is in one of the namespaces
func HasNamespace(ingress *v1beta1.Ingress, namespaces Namespaces) bool {
	if len(namespaces) == 0 {
		return true
	}
	for _, n := range namespaces {
		if ingress.ObjectMeta.Namespace == n {
			return true
		}
	}
	return false
}

// NewListWatchFromClient creates a new ListWatch from the specified client, resource, namespace, field selector and label selector.
// Extends cache.NewListWatchFromClient to support labelSelector
func NewListWatchFromClient(c cache.Getter, resource string, namespace string, fieldSelector fields.Selector, labelSelector labels.Selector) *cache.ListWatch {
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
