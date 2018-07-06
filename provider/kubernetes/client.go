package kubernetes

import (
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/containous/traefik/log"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	kubeerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

const resyncPeriod = 10 * time.Minute

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

// Client is a client for the Provider master.
// WatchAll starts the watch of the Provider resources and updates the stores.
// The stores can then be accessed via the Get* functions.
type Client interface {
	WatchAll(namespaces Namespaces, stopCh <-chan struct{}, eventsChan chan<- interface{}) error
	WatchNamespaces(namespaces Namespaces, stopCh <-chan struct{}, namespaceChan chan<- interface{}) error
	GetIngresses() []*extensionsv1beta1.Ingress
	GetService(namespace, name string) (*corev1.Service, bool, error)
	GetSecret(namespace, name string) (*corev1.Secret, bool, error)
	GetEndpoints(namespace, name string) (*corev1.Endpoints, bool, error)
	UpdateIngressStatus(namespace, name, ip, hostname string) error
}

type clientImpl struct {
	clientset              *kubernetes.Clientset
	factories              map[string]informers.SharedInformerFactory
	namespaceFactory       informers.SharedInformerFactory
	ingressLabelSelector   labels.Selector
	namespaceLabelSelector labels.Selector
	isNamespaceAll         bool
}

func newClientImpl(clientset *kubernetes.Clientset) *clientImpl {
	return &clientImpl{
		clientset: clientset,
		factories: make(map[string]informers.SharedInformerFactory),
	}
}

// newInClusterClient returns a new Provider client that is expected to run
// inside the cluster.
func newInClusterClient(endpoint string) (*clientImpl, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster configuration: %v", err)
	}

	if endpoint != "" {
		config.Host = endpoint
	}

	return createClientFromConfig(config)
}

// newExternalClusterClient returns a new Provider client that may run outside
// of the cluster.
// The endpoint parameter must not be empty.
func newExternalClusterClient(endpoint, token, caFilePath string) (*clientImpl, error) {
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
			return nil, fmt.Errorf("failed to read CA file %s: %v", caFilePath, err)
		}

		config.TLSClientConfig = rest.TLSClientConfig{CAData: caData}
	}

	return createClientFromConfig(config)
}

func createClientFromConfig(c *rest.Config) (*clientImpl, error) {
	clientset, err := kubernetes.NewForConfig(c)
	if err != nil {
		return nil, err
	}

	return newClientImpl(clientset), nil
}

// WatchNamespaces starts a controller to watch for namespace events.
func (c *clientImpl) WatchNamespaces(namespaces Namespaces, stopCh <-chan struct{}, namespaceChan chan<- interface{}) error {
	eventHandler := c.newResourceEventHandler(namespaceChan)

	c.namespaceFactory = informers.NewSharedInformerFactory(c.clientset, resyncPeriod)
	c.namespaceFactory.Core().V1().Namespaces().Informer().AddEventHandler(eventHandler)

	// If no namespaces are specified
	if len(namespaces) == 0 {
		// If no namespacelabels are specified
		if c.namespaceLabelSelector.Empty() {
			namespaces = Namespaces{metav1.NamespaceAll}
			c.isNamespaceAll = true
		} else {
			// namespacelabels are being used: watch all namespaces for events and use them to get namespaces to watch.
			c.namespaceFactory.Start(stopCh)
			c.namespaceFactory.WaitForCacheSync(stopCh)
		}
	}
	return nil
}

// WatchAll starts namespace-specific controllers for all relevant kinds.
func (c *clientImpl) WatchAll(namespaces Namespaces, stopCh <-chan struct{}, eventsChan chan<- interface{}) error {
	eventHandler := c.newResourceEventHandler(eventsChan)

	var namespacesToWatch []string

	namespaceList, err := c.getNamespaces()
	if err != nil {
		return fmt.Errorf("could not list namespaces: %v", err)
	}

	for _, item := range namespaceList.Items {
		log.Debugf("Adding found namespace %q to namespace list", item.ObjectMeta.Name)
		namespacesToWatch = append(namespacesToWatch, item.ObjectMeta.Name)
	}

	for _, ns := range namespacesToWatch {
		factory := informers.NewFilteredSharedInformerFactory(c.clientset, resyncPeriod, ns, nil)
		factory.Extensions().V1beta1().Ingresses().Informer().AddEventHandler(eventHandler)
		factory.Core().V1().Services().Informer().AddEventHandler(eventHandler)
		factory.Core().V1().Endpoints().Informer().AddEventHandler(eventHandler)
		c.factories[ns] = factory
	}

	for _, ns := range namespacesToWatch {
		c.factories[ns].Start(stopCh)
	}

	for _, ns := range namespacesToWatch {
		for t, ok := range c.factories[ns].WaitForCacheSync(stopCh) {
			if !ok {
				return fmt.Errorf("timed out waiting for controller caches to sync %s in namespace %q", t.String(), ns)
			}
		}
	}

	// Do not wait for the Secrets store to get synced since we cannot rely on
	// users having granted RBAC permissions for this object.
	// https://github.com/containous/traefik/issues/1784 should improve the
	// situation here in the future.
	for _, ns := range namespacesToWatch {
		c.factories[ns].Core().V1().Secrets().Informer().AddEventHandler(eventHandler)
		c.factories[ns].Start(stopCh)
	}

	return nil
}

// GetIngresses returns all Ingresses for observed namespaces in the cluster.
func (c *clientImpl) GetIngresses() []*extensionsv1beta1.Ingress {
	namespaceList, err := c.getNamespaces()
	if err != nil {
		log.Errorf("could not list namespaces: %v", err)
		return nil
	}

	var result []*extensionsv1beta1.Ingress
	for _, item := range namespaceList.Items {
		ns := item.ObjectMeta.Name
		ings, err := c.factories[ns].Extensions().V1beta1().Ingresses().Lister().List(c.ingressLabelSelector)
		if err != nil {
			log.Errorf("Failed to list ingresses in namespace %s: %v", ns, err)
			continue
		}

		for _, ing := range ings {
			result = append(result, ing)
		}
	}
	return result
}

// UpdateIngressStatus updates an Ingress with a provided status.
func (c *clientImpl) UpdateIngressStatus(namespace, name, ip, hostname string) error {
	ing, err := c.factories[namespace].Extensions().V1beta1().Ingresses().Lister().Ingresses(namespace).Get(name)
	if err != nil {
		return fmt.Errorf("failed to get ingress %s/%s: %v", namespace, name, err)
	}

	if len(ing.Status.LoadBalancer.Ingress) > 0 {
		if ing.Status.LoadBalancer.Ingress[0].Hostname == hostname && ing.Status.LoadBalancer.Ingress[0].IP == ip {
			// If status is already set, skip update
			log.Debugf("Skipping status update on ingress %s/%s", ing.Namespace, ing.Name)
			return nil
		}
	}
	ingCopy := ing.DeepCopy()
	ingCopy.Status = extensionsv1beta1.IngressStatus{LoadBalancer: corev1.LoadBalancerStatus{Ingress: []corev1.LoadBalancerIngress{{IP: ip, Hostname: hostname}}}}

	_, err = c.clientset.ExtensionsV1beta1().Ingresses(ingCopy.Namespace).UpdateStatus(ingCopy)
	if err != nil {
		return fmt.Errorf("failed to update ingress status %s/%s: %v", namespace, name, err)
	}
	log.Infof("Updated status on ingress %s/%s", namespace, name)
	return nil
}

// GetNamespaces returns namespaces with the configured labelselector.
func (c *clientImpl) getNamespaces() (*corev1.NamespaceList, error) {
	return c.clientset.CoreV1().Namespaces().List(metav1.ListOptions{LabelSelector: c.namespaceLabelSelector.String()})
}

// GetService returns the named service from the configured namespace.
func (c *clientImpl) GetService(namespace, name string) (*corev1.Service, bool, error) {
	service, err := c.factories[namespace].Core().V1().Services().Lister().Services(namespace).Get(name)
	exist, err := translateNotFoundError(err)
	return service, exist, err
}

// GetEndpoints returns the named endpoints from the configured namespace.
func (c *clientImpl) GetEndpoints(namespace, name string) (*corev1.Endpoints, bool, error) {
	endpoint, err := c.factories[namespace].Core().V1().Endpoints().Lister().Endpoints(namespace).Get(name)
	exist, err := translateNotFoundError(err)
	return endpoint, exist, err
}

// GetSecret returns the named secret from the configured namespace.
func (c *clientImpl) GetSecret(namespace, name string) (*corev1.Secret, bool, error) {
	secret, err := c.factories[namespace].Core().V1().Secrets().Lister().Secrets(namespace).Get(name)
	exist, err := translateNotFoundError(err)
	return secret, exist, err
}

func (c *clientImpl) newResourceEventHandler(events chan<- interface{}) cache.ResourceEventHandler {
	return &cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			// Ignore Ingresses that do not match our custom label selector.
			if ing, ok := obj.(*extensionsv1beta1.Ingress); ok {
				lbls := labels.Set(ing.GetLabels())
				return c.ingressLabelSelector.Matches(lbls)
			}
			return true
		},
		Handler: &resourceEventHandler{events},
	}
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

// translateNotFoundError will translate a "not found" error to a boolean return
// value which indicates if the resource exists and a nil error.
func translateNotFoundError(err error) (bool, error) {
	if kubeerror.IsNotFound(err) {
		return false, nil
	}
	return err == nil, err
}
