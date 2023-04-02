package gateway

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/traefik/traefik/v2/pkg/log"
	corev1 "k8s.io/api/core/v1"
	kerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kinformers "k8s.io/client-go/informers"
	kclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	gatev1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gateclientset "sigs.k8s.io/gateway-api/pkg/client/clientset/gateway/versioned"
	gateinformers "sigs.k8s.io/gateway-api/pkg/client/informers/gateway/externalversions"
)

const resyncPeriod = 10 * time.Minute

type resourceEventHandler struct {
	ev chan<- interface{}
}

func (reh *resourceEventHandler) OnAdd(obj interface{}) {
	eventHandlerFunc(reh.ev, obj)
}

func (reh *resourceEventHandler) OnUpdate(oldObj, newObj interface{}) {
	switch oldObj.(type) {
	case *gatev1alpha2.GatewayClass:
		// Skip update for gateway classes. We only manage addition or deletion for this cluster-wide resource.
		return
	default:
		eventHandlerFunc(reh.ev, newObj)
	}
}

func (reh *resourceEventHandler) OnDelete(obj interface{}) {
	eventHandlerFunc(reh.ev, obj)
}

// Client is a client for the Provider master.
// WatchAll starts the watch of the Provider resources and updates the stores.
// The stores can then be accessed via the Get* functions.
type Client interface {
	WatchAll(namespaces []string, stopCh <-chan struct{}) (<-chan interface{}, error)
	GetGatewayClasses() ([]*gatev1alpha2.GatewayClass, error)
	UpdateGatewayStatus(gateway *gatev1alpha2.Gateway, gatewayStatus gatev1alpha2.GatewayStatus) error
	UpdateGatewayClassStatus(gatewayClass *gatev1alpha2.GatewayClass, condition metav1.Condition) error
	GetGateways() []*gatev1alpha2.Gateway
	GetHTTPRoutes(namespaces []string) ([]*gatev1alpha2.HTTPRoute, error)
	GetTCPRoutes(namespaces []string) ([]*gatev1alpha2.TCPRoute, error)
	GetTLSRoutes(namespaces []string) ([]*gatev1alpha2.TLSRoute, error)
	GetService(namespace, name string) (*corev1.Service, bool, error)
	GetSecret(namespace, name string) (*corev1.Secret, bool, error)
	GetEndpoints(namespace, name string) (*corev1.Endpoints, bool, error)
	GetNamespaces(selector labels.Selector) ([]string, error)
}

type clientWrapper struct {
	csGateway gateclientset.Interface
	csKube    kclientset.Interface

	factoryNamespace    kinformers.SharedInformerFactory
	factoryGatewayClass gateinformers.SharedInformerFactory
	factoriesGateway    map[string]gateinformers.SharedInformerFactory
	factoriesKube       map[string]kinformers.SharedInformerFactory
	factoriesSecret     map[string]kinformers.SharedInformerFactory

	isNamespaceAll    bool
	watchedNamespaces []string

	labelSelector string
}

func createClientFromConfig(c *rest.Config) (*clientWrapper, error) {
	csGateway, err := gateclientset.NewForConfig(c)
	if err != nil {
		return nil, err
	}

	csKube, err := kclientset.NewForConfig(c)
	if err != nil {
		return nil, err
	}

	return newClientImpl(csKube, csGateway), nil
}

func newClientImpl(csKube kclientset.Interface, csGateway gateclientset.Interface) *clientWrapper {
	return &clientWrapper{
		csGateway:        csGateway,
		csKube:           csKube,
		factoriesGateway: make(map[string]gateinformers.SharedInformerFactory),
		factoriesKube:    make(map[string]kinformers.SharedInformerFactory),
		factoriesSecret:  make(map[string]kinformers.SharedInformerFactory),
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

// newExternalClusterClient returns a new Provider client that may run outside of the cluster.
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
	eventHandler := &resourceEventHandler{ev: eventCh}

	if len(namespaces) == 0 {
		namespaces = []string{metav1.NamespaceAll}
		c.isNamespaceAll = true
	}

	c.watchedNamespaces = namespaces

	notOwnedByHelm := func(opts *metav1.ListOptions) {
		opts.LabelSelector = "owner!=helm"
	}

	labelSelectorOptions := func(options *metav1.ListOptions) {
		options.LabelSelector = c.labelSelector
	}

	c.factoryNamespace = kinformers.NewSharedInformerFactory(c.csKube, resyncPeriod)
	_, err := c.factoryNamespace.Core().V1().Namespaces().Informer().AddEventHandler(eventHandler)
	if err != nil {
		return nil, err
	}

	c.factoryGatewayClass = gateinformers.NewSharedInformerFactoryWithOptions(c.csGateway, resyncPeriod, gateinformers.WithTweakListOptions(labelSelectorOptions))
	_, err = c.factoryGatewayClass.Gateway().V1alpha2().GatewayClasses().Informer().AddEventHandler(eventHandler)
	if err != nil {
		return nil, err
	}

	// TODO manage Reference Policy
	// https://gateway-api.sigs.k8s.io/v1alpha2/references/spec/#gateway.networking.k8s.io/v1alpha2.ReferencePolicy

	for _, ns := range namespaces {
		factoryGateway := gateinformers.NewSharedInformerFactoryWithOptions(c.csGateway, resyncPeriod, gateinformers.WithNamespace(ns))
		_, err = factoryGateway.Gateway().V1alpha2().Gateways().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		_, err = factoryGateway.Gateway().V1alpha2().HTTPRoutes().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		_, err = factoryGateway.Gateway().V1alpha2().TCPRoutes().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		_, err = factoryGateway.Gateway().V1alpha2().TLSRoutes().Informer().AddEventHandler(eventHandler)
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

		c.factoriesGateway[ns] = factoryGateway
		c.factoriesKube[ns] = factoryKube
		c.factoriesSecret[ns] = factorySecret
	}

	c.factoryNamespace.Start(stopCh)
	c.factoryGatewayClass.Start(stopCh)

	for _, ns := range namespaces {
		c.factoriesGateway[ns].Start(stopCh)
		c.factoriesKube[ns].Start(stopCh)
		c.factoriesSecret[ns].Start(stopCh)
	}

	for t, ok := range c.factoryNamespace.WaitForCacheSync(stopCh) {
		if !ok {
			return nil, fmt.Errorf("timed out waiting for controller caches to sync %s", t.String())
		}
	}

	for t, ok := range c.factoryGatewayClass.WaitForCacheSync(stopCh) {
		if !ok {
			return nil, fmt.Errorf("timed out waiting for controller caches to sync %s", t.String())
		}
	}

	for _, ns := range namespaces {
		for t, ok := range c.factoriesGateway[ns].WaitForCacheSync(stopCh) {
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

func (c *clientWrapper) GetNamespaces(selector labels.Selector) ([]string, error) {
	ns, err := c.factoryNamespace.Core().V1().Namespaces().Lister().List(selector)
	if err != nil {
		return nil, err
	}

	var namespaces []string
	for _, namespace := range ns {
		if !c.isWatchedNamespace(namespace.Name) {
			log.WithoutContext().Warnf("Namespace %q is not within  watched namespaces", selector, namespace)
			continue
		}
		namespaces = append(namespaces, namespace.Name)
	}
	return namespaces, nil
}

func (c *clientWrapper) GetHTTPRoutes(namespaces []string) ([]*gatev1alpha2.HTTPRoute, error) {
	var httpRoutes []*gatev1alpha2.HTTPRoute
	for _, namespace := range namespaces {
		if !c.isWatchedNamespace(namespace) {
			log.WithoutContext().Warnf("Failed to get HTTPRoutes: %q is not within watched namespaces", namespace)
			continue
		}

		routes, err := c.factoriesGateway[c.lookupNamespace(namespace)].Gateway().V1alpha2().HTTPRoutes().Lister().HTTPRoutes(namespace).List(labels.Everything())
		if err != nil {
			return nil, err
		}

		if len(routes) == 0 {
			log.WithoutContext().Debugf("No HTTPRoutes found in namespace %q", namespace)
			continue
		}

		httpRoutes = append(httpRoutes, routes...)
	}

	return httpRoutes, nil
}

func (c *clientWrapper) GetTCPRoutes(namespaces []string) ([]*gatev1alpha2.TCPRoute, error) {
	var tcpRoutes []*gatev1alpha2.TCPRoute
	for _, namespace := range namespaces {
		if !c.isWatchedNamespace(namespace) {
			log.WithoutContext().Warnf("Failed to get TCPRoutes: %q is not within watched namespaces", namespace)
			continue
		}

		routes, err := c.factoriesGateway[c.lookupNamespace(namespace)].Gateway().V1alpha2().TCPRoutes().Lister().TCPRoutes(namespace).List(labels.Everything())
		if err != nil {
			return nil, err
		}

		if len(routes) == 0 {
			log.WithoutContext().Debugf("No TCPRoutes found in namespace %q", namespace)
			continue
		}

		tcpRoutes = append(tcpRoutes, routes...)
	}
	return tcpRoutes, nil
}

func (c *clientWrapper) GetTLSRoutes(namespaces []string) ([]*gatev1alpha2.TLSRoute, error) {
	var tlsRoutes []*gatev1alpha2.TLSRoute
	for _, namespace := range namespaces {
		if !c.isWatchedNamespace(namespace) {
			log.WithoutContext().Warnf("Failed to get TLSRoutes: %q is not within watched namespaces", namespace)
			continue
		}

		routes, err := c.factoriesGateway[c.lookupNamespace(namespace)].Gateway().V1alpha2().TLSRoutes().Lister().TLSRoutes(namespace).List(labels.Everything())
		if err != nil {
			return nil, err
		}

		if len(routes) == 0 {
			log.WithoutContext().Debugf("No TLSRoutes found in namespace %q", namespace)
			continue
		}

		tlsRoutes = append(tlsRoutes, routes...)
	}
	return tlsRoutes, nil
}

func (c *clientWrapper) GetGateways() []*gatev1alpha2.Gateway {
	var result []*gatev1alpha2.Gateway

	for ns, factory := range c.factoriesGateway {
		gateways, err := factory.Gateway().V1alpha2().Gateways().Lister().List(labels.Everything())
		if err != nil {
			log.WithoutContext().Errorf("Failed to list Gateways in namespace %s: %v", ns, err)
			continue
		}
		result = append(result, gateways...)
	}

	return result
}

func (c *clientWrapper) GetGatewayClasses() ([]*gatev1alpha2.GatewayClass, error) {
	return c.factoryGatewayClass.Gateway().V1alpha2().GatewayClasses().Lister().List(labels.Everything())
}

func (c *clientWrapper) UpdateGatewayClassStatus(gatewayClass *gatev1alpha2.GatewayClass, condition metav1.Condition) error {
	gc := gatewayClass.DeepCopy()

	var newConditions []metav1.Condition
	for _, cond := range gc.Status.Conditions {
		// No update for identical condition.
		if cond.Type == condition.Type && cond.Status == condition.Status {
			return nil
		}

		// Keep other condition types.
		if cond.Type != condition.Type {
			newConditions = append(newConditions, cond)
		}
	}

	// Append the condition to update.
	newConditions = append(newConditions, condition)
	gc.Status.Conditions = newConditions

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.csGateway.GatewayV1alpha2().GatewayClasses().UpdateStatus(ctx, gc, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update GatewayClass %q status: %w", gatewayClass.Name, err)
	}

	return nil
}

func (c *clientWrapper) UpdateGatewayStatus(gateway *gatev1alpha2.Gateway, gatewayStatus gatev1alpha2.GatewayStatus) error {
	if !c.isWatchedNamespace(gateway.Namespace) {
		return fmt.Errorf("cannot update Gateway status %s/%s: namespace is not within watched namespaces", gateway.Namespace, gateway.Name)
	}

	if statusEquals(gateway.Status, gatewayStatus) {
		return nil
	}

	g := gateway.DeepCopy()
	g.Status = gatewayStatus

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.csGateway.GatewayV1alpha2().Gateways(gateway.Namespace).UpdateStatus(ctx, g, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update Gateway %q status: %w", gateway.Name, err)
	}

	return nil
}

func statusEquals(oldStatus, newStatus gatev1alpha2.GatewayStatus) bool {
	if len(oldStatus.Listeners) != len(newStatus.Listeners) {
		return false
	}

	if !conditionsEquals(oldStatus.Conditions, newStatus.Conditions) {
		return false
	}

	listenerMatches := 0
	for _, newListener := range newStatus.Listeners {
		for _, oldListener := range oldStatus.Listeners {
			if newListener.Name == oldListener.Name {
				if !conditionsEquals(newListener.Conditions, oldListener.Conditions) {
					return false
				}

				listenerMatches++
			}
		}
	}

	return listenerMatches == len(oldStatus.Listeners)
}

func conditionsEquals(conditionsA, conditionsB []metav1.Condition) bool {
	if len(conditionsA) != len(conditionsB) {
		return false
	}

	conditionMatches := 0
	for _, conditionA := range conditionsA {
		for _, conditionB := range conditionsB {
			if conditionA.Type == conditionB.Type {
				if conditionA.Reason != conditionB.Reason || conditionA.Status != conditionB.Status || conditionA.Message != conditionB.Message {
					return false
				}
				conditionMatches++
			}
		}
	}

	return conditionMatches == len(conditionsA)
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
