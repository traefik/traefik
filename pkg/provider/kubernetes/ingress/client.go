package ingress

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/go-version"
	"github.com/traefik/traefik/v2/pkg/log"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	kubeerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	resyncPeriod   = 10 * time.Minute
	defaultTimeout = 5 * time.Second
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

// Client is a client for the Provider master.
// WatchAll starts the watch of the Provider resources and updates the stores.
// The stores can then be accessed via the Get* functions.
type Client interface {
	WatchAll(namespaces []string, stopCh <-chan struct{}) (<-chan interface{}, error)
	GetIngresses() []*networkingv1beta1.Ingress
	GetIngressClass() (*networkingv1beta1.IngressClass, error)
	GetService(namespace, name string) (*corev1.Service, bool, error)
	GetSecret(namespace, name string) (*corev1.Secret, bool, error)
	GetEndpoints(namespace, name string) (*corev1.Endpoints, bool, error)
	UpdateIngressStatus(ing *networkingv1beta1.Ingress, ip, hostname string) error
	GetServerVersion() (*version.Version, error)
}

type clientWrapper struct {
	clientset            *kubernetes.Clientset
	factories            map[string]informers.SharedInformerFactory
	clusterFactory       informers.SharedInformerFactory
	ingressLabelSelector labels.Selector
	isNamespaceAll       bool
	watchedNamespaces    []string
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

func createClientFromConfig(c *rest.Config) (*clientWrapper, error) {
	clientset, err := kubernetes.NewForConfig(c)
	if err != nil {
		return nil, err
	}

	return newClientImpl(clientset), nil
}

func newClientImpl(clientset *kubernetes.Clientset) *clientWrapper {
	return &clientWrapper{
		clientset: clientset,
		factories: make(map[string]informers.SharedInformerFactory),
	}
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

	for _, ns := range namespaces {
		factory := informers.NewSharedInformerFactoryWithOptions(c.clientset, resyncPeriod, informers.WithNamespace(ns))
		factory.Extensions().V1beta1().Ingresses().Informer().AddEventHandler(eventHandler)
		factory.Core().V1().Services().Informer().AddEventHandler(eventHandler)
		factory.Core().V1().Endpoints().Informer().AddEventHandler(eventHandler)
		factory.Core().V1().Secrets().Informer().AddEventHandler(eventHandler)
		c.factories[ns] = factory
	}

	for _, ns := range namespaces {
		c.factories[ns].Start(stopCh)
	}

	for _, ns := range namespaces {
		for typ, ok := range c.factories[ns].WaitForCacheSync(stopCh) {
			if !ok {
				return nil, fmt.Errorf("timed out waiting for controller caches to sync %s in namespace %q", typ, ns)
			}
		}
	}

	serverVersion, err := c.GetServerVersion()
	if err != nil {
		log.WithoutContext().Errorf("Failed to get server version: %v", err)
		return eventCh, nil
	}

	if supportsIngressClass(serverVersion) {
		c.clusterFactory = informers.NewSharedInformerFactoryWithOptions(c.clientset, resyncPeriod)
		c.clusterFactory.Networking().V1beta1().IngressClasses().Informer().AddEventHandler(eventHandler)
		c.clusterFactory.Start(stopCh)

		for typ, ok := range c.clusterFactory.WaitForCacheSync(stopCh) {
			if !ok {
				return nil, fmt.Errorf("timed out waiting for controller caches to sync %s", typ)
			}
		}
	}

	return eventCh, nil
}

// GetIngresses returns all Ingresses for observed namespaces in the cluster.
func (c *clientWrapper) GetIngresses() []*networkingv1beta1.Ingress {
	var results []*networkingv1beta1.Ingress

	for ns, factory := range c.factories {
		// extensions
		ings, err := factory.Extensions().V1beta1().Ingresses().Lister().List(c.ingressLabelSelector)
		if err != nil {
			log.Errorf("Failed to list ingresses in namespace %s: %v", ns, err)
		}

		for _, ing := range ings {
			n, err := extensionsToNetworking(ing)
			if err != nil {
				log.Errorf("Failed to convert ingress %s from extensions/v1beta1 to networking/v1beta1: %v", ns, err)
				continue
			}
			results = append(results, n)
		}

		// networking
		list, err := factory.Networking().V1beta1().Ingresses().Lister().List(c.ingressLabelSelector)
		if err != nil {
			log.Errorf("Failed to list ingresses in namespace %s: %v", ns, err)
		}
		results = append(results, list...)
	}
	return results
}

func extensionsToNetworking(ing proto.Marshaler) (*networkingv1beta1.Ingress, error) {
	data, err := ing.Marshal()
	if err != nil {
		return nil, err
	}

	ni := &networkingv1beta1.Ingress{}
	err = ni.Unmarshal(data)
	if err != nil {
		return nil, err
	}

	return ni, nil
}

// UpdateIngressStatus updates an Ingress with a provided status.
func (c *clientWrapper) UpdateIngressStatus(src *networkingv1beta1.Ingress, ip, hostname string) error {
	if !c.isWatchedNamespace(src.Namespace) {
		return fmt.Errorf("failed to get ingress %s/%s: namespace is not within watched namespaces", src.Namespace, src.Name)
	}

	if src.GetObjectKind().GroupVersionKind().Group != "networking.k8s.io" {
		return c.updateIngressStatusOld(src, ip, hostname)
	}

	ing, err := c.factories[c.lookupNamespace(src.Namespace)].Networking().V1beta1().Ingresses().Lister().Ingresses(src.Namespace).Get(src.Name)
	if err != nil {
		return fmt.Errorf("failed to get ingress %s/%s: %w", src.Namespace, src.Name, err)
	}

	if len(ing.Status.LoadBalancer.Ingress) > 0 {
		if ing.Status.LoadBalancer.Ingress[0].Hostname == hostname && ing.Status.LoadBalancer.Ingress[0].IP == ip {
			// If status is already set, skip update
			log.Debugf("Skipping status update on ingress %s/%s", ing.Namespace, ing.Name)
			return nil
		}
	}

	ingCopy := ing.DeepCopy()
	ingCopy.Status = networkingv1beta1.IngressStatus{LoadBalancer: corev1.LoadBalancerStatus{Ingress: []corev1.LoadBalancerIngress{{IP: ip, Hostname: hostname}}}}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	_, err = c.clientset.NetworkingV1beta1().Ingresses(ingCopy.Namespace).UpdateStatus(ctx, ingCopy, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update ingress status %s/%s: %w", src.Namespace, src.Name, err)
	}

	log.Infof("Updated status on ingress %s/%s", src.Namespace, src.Name)
	return nil
}

func (c *clientWrapper) updateIngressStatusOld(src *networkingv1beta1.Ingress, ip, hostname string) error {
	ing, err := c.factories[c.lookupNamespace(src.Namespace)].Extensions().V1beta1().Ingresses().Lister().Ingresses(src.Namespace).Get(src.Name)
	if err != nil {
		return fmt.Errorf("failed to get ingress %s/%s: %w", src.Namespace, src.Name, err)
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

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	_, err = c.clientset.ExtensionsV1beta1().Ingresses(ingCopy.Namespace).UpdateStatus(ctx, ingCopy, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update ingress status %s/%s: %w", src.Namespace, src.Name, err)
	}

	log.Infof("Updated status on ingress %s/%s", src.Namespace, src.Name)
	return nil
}

// GetService returns the named service from the given namespace.
func (c *clientWrapper) GetService(namespace, name string) (*corev1.Service, bool, error) {
	if !c.isWatchedNamespace(namespace) {
		return nil, false, fmt.Errorf("failed to get service %s/%s: namespace is not within watched namespaces", namespace, name)
	}

	service, err := c.factories[c.lookupNamespace(namespace)].Core().V1().Services().Lister().Services(namespace).Get(name)
	exist, err := translateNotFoundError(err)
	return service, exist, err
}

// GetEndpoints returns the named endpoints from the given namespace.
func (c *clientWrapper) GetEndpoints(namespace, name string) (*corev1.Endpoints, bool, error) {
	if !c.isWatchedNamespace(namespace) {
		return nil, false, fmt.Errorf("failed to get endpoints %s/%s: namespace is not within watched namespaces", namespace, name)
	}

	endpoint, err := c.factories[c.lookupNamespace(namespace)].Core().V1().Endpoints().Lister().Endpoints(namespace).Get(name)
	exist, err := translateNotFoundError(err)
	return endpoint, exist, err
}

// GetSecret returns the named secret from the given namespace.
func (c *clientWrapper) GetSecret(namespace, name string) (*corev1.Secret, bool, error) {
	if !c.isWatchedNamespace(namespace) {
		return nil, false, fmt.Errorf("failed to get secret %s/%s: namespace is not within watched namespaces", namespace, name)
	}

	secret, err := c.factories[c.lookupNamespace(namespace)].Core().V1().Secrets().Lister().Secrets(namespace).Get(name)
	exist, err := translateNotFoundError(err)
	return secret, exist, err
}

func (c *clientWrapper) GetIngressClass() (*networkingv1beta1.IngressClass, error) {
	if c.clusterFactory == nil {
		return nil, errors.New("failed to find ingressClass: factory not loaded")
	}

	ingressClasses, err := c.clusterFactory.Networking().V1beta1().IngressClasses().Lister().List(labels.Everything())
	if err != nil {
		return nil, err
	}

	for _, ic := range ingressClasses {
		if ic.Spec.Controller == traefikDefaultIngressClassController {
			return ic, nil
		}
	}

	return nil, nil
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
			switch v := obj.(type) {
			case *extensionsv1beta1.Ingress:
				lbls := labels.Set(v.GetLabels())
				return c.ingressLabelSelector.Matches(lbls)
			case *networkingv1beta1.Ingress:
				lbls := labels.Set(v.GetLabels())
				return c.ingressLabelSelector.Matches(lbls)
			default:
				return true
			}
		},
		Handler: &resourceEventHandler{ev: events},
	}
}

// GetServerVersion returns the cluster server version, or an error.
func (c *clientWrapper) GetServerVersion() (*version.Version, error) {
	serverVersion, err := c.clientset.Discovery().ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("could not retrieve server version: %w", err)
	}

	return version.NewVersion(serverVersion.GitVersion)
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
	if kubeerror.IsNotFound(err) {
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

// IngressClass objects are supported since Kubernetes v1.18.
// See https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-class
func supportsIngressClass(serverVersion *version.Version) bool {
	ingressClassVersion := version.Must(version.NewVersion("1.18"))

	return ingressClassVersion.LessThanOrEqual(serverVersion)
}
