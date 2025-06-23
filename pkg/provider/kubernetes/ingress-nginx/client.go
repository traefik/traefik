package ingressnginx

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/k8s"
	"github.com/traefik/traefik/v3/pkg/types"
	traefikversion "github.com/traefik/traefik/v3/pkg/version"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	netv1 "k8s.io/api/networking/v1"
	kerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	kinformers "k8s.io/client-go/informers"
	kclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	resyncPeriod   = 10 * time.Minute
	defaultTimeout = 5 * time.Second
)

type clientWrapper struct {
	clientset           kclientset.Interface
	clusterScopeFactory kinformers.SharedInformerFactory
	factoriesKube       map[string]kinformers.SharedInformerFactory
	factoriesSecret     map[string]kinformers.SharedInformerFactory
	factoriesIngress    map[string]kinformers.SharedInformerFactory
	isNamespaceAll      bool
	watchedNamespaces   []string

	ignoreIngressClasses bool
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

func createClientFromConfig(c *rest.Config) (*clientWrapper, error) {
	c.UserAgent = fmt.Sprintf(
		"%s/%s (%s/%s) kubernetes/ingress",
		filepath.Base(os.Args[0]),
		traefikversion.Version,
		runtime.GOOS,
		runtime.GOARCH,
	)

	clientset, err := kclientset.NewForConfig(c)
	if err != nil {
		return nil, err
	}

	return newClient(clientset), nil
}

func newClient(clientSet kclientset.Interface) *clientWrapper {
	return &clientWrapper{
		clientset:        clientSet,
		factoriesSecret:  make(map[string]kinformers.SharedInformerFactory),
		factoriesIngress: make(map[string]kinformers.SharedInformerFactory),
		factoriesKube:    make(map[string]kinformers.SharedInformerFactory),
	}
}

// WatchAll starts namespace-specific controllers for all relevant kinds.
func (c *clientWrapper) WatchAll(ctx context.Context, namespace, namespaceSelector string) (<-chan interface{}, error) {
	stopCh := ctx.Done()
	eventCh := make(chan interface{}, 1)
	eventHandler := &k8s.ResourceEventHandler{Ev: eventCh}

	c.ignoreIngressClasses = false
	_, err := c.clientset.NetworkingV1().IngressClasses().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		if !kerror.IsNotFound(err) {
			if kerror.IsForbidden(err) {
				c.ignoreIngressClasses = true
			}
		}
	}

	if namespaceSelector != "" {
		ns, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{LabelSelector: namespaceSelector})
		if err != nil {
			return nil, fmt.Errorf("listing namespaces: %w", err)
		}
		for _, item := range ns.Items {
			c.watchedNamespaces = append(c.watchedNamespaces, item.Name)
		}
	} else {
		c.isNamespaceAll = namespace == metav1.NamespaceAll
		c.watchedNamespaces = []string{namespace}
	}

	notOwnedByHelm := func(opts *metav1.ListOptions) {
		opts.LabelSelector = "owner!=helm"
	}

	for _, ns := range c.watchedNamespaces {
		factoryIngress := kinformers.NewSharedInformerFactoryWithOptions(c.clientset, resyncPeriod, kinformers.WithNamespace(ns))

		_, err := factoryIngress.Networking().V1().Ingresses().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}

		c.factoriesIngress[ns] = factoryIngress

		factoryKube := kinformers.NewSharedInformerFactoryWithOptions(c.clientset, resyncPeriod, kinformers.WithNamespace(ns))
		_, err = factoryKube.Core().V1().Services().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		_, err = factoryKube.Discovery().V1().EndpointSlices().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		c.factoriesKube[ns] = factoryKube

		factorySecret := kinformers.NewSharedInformerFactoryWithOptions(c.clientset, resyncPeriod, kinformers.WithNamespace(ns), kinformers.WithTweakListOptions(notOwnedByHelm))
		_, err = factorySecret.Core().V1().Secrets().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		c.factoriesSecret[ns] = factorySecret
	}

	for _, ns := range c.watchedNamespaces {
		c.factoriesIngress[ns].Start(stopCh)
		c.factoriesKube[ns].Start(stopCh)
		c.factoriesSecret[ns].Start(stopCh)
	}

	for _, ns := range c.watchedNamespaces {
		for t, ok := range c.factoriesIngress[ns].WaitForCacheSync(stopCh) {
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

	c.clusterScopeFactory = kinformers.NewSharedInformerFactory(c.clientset, resyncPeriod)

	if !c.ignoreIngressClasses {
		_, err = c.clusterScopeFactory.Networking().V1().IngressClasses().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
	}

	c.clusterScopeFactory.Start(stopCh)

	for t, ok := range c.clusterScopeFactory.WaitForCacheSync(stopCh) {
		if !ok {
			return nil, fmt.Errorf("timed out waiting for controller caches to sync %s", t.String())
		}
	}

	return eventCh, nil
}

func (c *clientWrapper) ListIngressClasses() ([]*netv1.IngressClass, error) {
	if c.ignoreIngressClasses {
		return []*netv1.IngressClass{}, nil
	}

	return c.clusterScopeFactory.Networking().V1().IngressClasses().Lister().List(labels.Everything())
}

// ListIngresses returns all Ingresses for observed namespaces in the cluster.
func (c *clientWrapper) ListIngresses() []*netv1.Ingress {
	var results []*netv1.Ingress

	for ns, factory := range c.factoriesIngress {
		// networking
		listNew, err := factory.Networking().V1().Ingresses().Lister().List(labels.Everything())
		if err != nil {
			log.Error().Err(err).Msgf("Failed to list ingresses in namespace %s", ns)
			continue
		}

		results = append(results, listNew...)
	}

	return results
}

// UpdateIngressStatus updates an Ingress with a provided status.
func (c *clientWrapper) UpdateIngressStatus(src *netv1.Ingress, ingStatus []netv1.IngressLoadBalancerIngress) error {
	if !c.isWatchedNamespace(src.Namespace) {
		return fmt.Errorf("failed to get ingress %s/%s: namespace is not within watched namespaces", src.Namespace, src.Name)
	}

	ing, err := c.factoriesIngress[c.lookupNamespace(src.Namespace)].Networking().V1().Ingresses().Lister().Ingresses(src.Namespace).Get(src.Name)
	if err != nil {
		return fmt.Errorf("failed to get ingress %s/%s: %w", src.Namespace, src.Name, err)
	}

	logger := log.With().Str("namespace", ing.Namespace).Str("ingress", ing.Name).Logger()

	if isLoadBalancerIngressEquals(ing.Status.LoadBalancer.Ingress, ingStatus) {
		logger.Debug().Msg("Skipping ingress status update")
		return nil
	}

	ingCopy := ing.DeepCopy()
	ingCopy.Status = netv1.IngressStatus{LoadBalancer: netv1.IngressLoadBalancerStatus{Ingress: ingStatus}}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	_, err = c.clientset.NetworkingV1().Ingresses(ingCopy.Namespace).UpdateStatus(ctx, ingCopy, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update ingress status %s/%s: %w", src.Namespace, src.Name, err)
	}

	logger.Info().Msg("Updated ingress status")
	return nil
}

// GetService returns the named service from the given namespace.
func (c *clientWrapper) GetService(namespace, name string) (*corev1.Service, error) {
	if !c.isWatchedNamespace(namespace) {
		return nil, fmt.Errorf("failed to get service %s/%s: namespace is not within watched namespaces", namespace, name)
	}

	return c.factoriesKube[c.lookupNamespace(namespace)].Core().V1().Services().Lister().Services(namespace).Get(name)
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
func (c *clientWrapper) GetSecret(namespace, name string) (*corev1.Secret, error) {
	if !c.isWatchedNamespace(namespace) {
		return nil, fmt.Errorf("failed to get secret %s/%s: namespace is not within watched namespaces", namespace, name)
	}

	return c.factoriesSecret[c.lookupNamespace(namespace)].Core().V1().Secrets().Lister().Secrets(namespace).Get(name)
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

// isLoadBalancerIngressEquals returns true if the given slices are equal, false otherwise.
func isLoadBalancerIngressEquals(aSlice, bSlice []netv1.IngressLoadBalancerIngress) bool {
	if len(aSlice) != len(bSlice) {
		return false
	}

	aMap := make(map[string]struct{})
	for _, aIngress := range aSlice {
		aMap[aIngress.Hostname+aIngress.IP] = struct{}{}
	}

	for _, bIngress := range bSlice {
		if _, exists := aMap[bIngress.Hostname+bIngress.IP]; !exists {
			return false
		}
	}

	return true
}

// filterIngressClass return a slice containing IngressClass matching either the annotation name or the controller.
func filterIngressClass(ingressClasses []*netv1.IngressClass, ingressClassByName bool, ingressClass, controllerClass string) []*netv1.IngressClass {
	var filteredIngressClasses []*netv1.IngressClass
	for _, ic := range ingressClasses {
		if ingressClassByName && ic.Name == ingressClass {
			return append(filteredIngressClasses, ic)
		}

		if ic.Spec.Controller == controllerClass {
			filteredIngressClasses = append(filteredIngressClasses, ic)
			continue
		}
	}

	return filteredIngressClasses
}
