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
	kinformers "k8s.io/client-go/informers"
	kclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	resyncPeriod   = 10 * time.Minute
	defaultTimeout = 5 * time.Second

	podNamespacePath = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

type clientWrapper struct {
	clientset           kclientset.Interface
	clusterScopeFactory kinformers.SharedInformerFactory
	factoriesKube       map[string]kinformers.SharedInformerFactory
	factoriesSecret     map[string]kinformers.SharedInformerFactory
	factoriesConfigMap  map[string]kinformers.SharedInformerFactory
	factoriesIngress    map[string]kinformers.SharedInformerFactory
	factoryPod          kinformers.SharedInformerFactory
	isNamespaceAll      bool
	watchedNamespaces   []string

	ignoreIngressClasses bool

	// podNamespace and podSelector identify the controller pods whose node
	// addresses are published in the Ingress status. They are empty when the
	// controller pod cannot be discovered, which disables that fallback.
	podNamespace string
	podSelector  string
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
		clientset:          clientSet,
		factoriesSecret:    make(map[string]kinformers.SharedInformerFactory),
		factoriesConfigMap: make(map[string]kinformers.SharedInformerFactory),
		factoriesIngress:   make(map[string]kinformers.SharedInformerFactory),
		factoriesKube:      make(map[string]kinformers.SharedInformerFactory),
	}
}

// WatchAll starts namespace-specific controllers for all relevant kinds.
func (c *clientWrapper) WatchAll(ctx context.Context, namespace, namespaceSelector string) (<-chan any, error) {
	stopCh := ctx.Done()
	eventCh := make(chan any, 1)
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

	// The controller pods and the nodes they run on are the last resort source
	// for the Ingress status, as done by the ingress-nginx controller.
	// Discovering them requires running in-cluster with the pods and nodes list
	// permissions: when unavailable, only the publish options can set the status.
	if err := c.setupPodDiscovery(ctx); err != nil {
		log.Debug().Err(err).Msg("Skipping controller pods discovery for the Ingress status")
	}

	notOwnedByHelm := func(opts *metav1.ListOptions) {
		opts.LabelSelector = "owner!=helm"
	}

	for _, ns := range c.watchedNamespaces {
		factoryIngress := kinformers.NewSharedInformerFactoryWithOptions(c.clientset, resyncPeriod, kinformers.WithNamespace(ns), kinformers.WithTransform(k8s.StripManagedFields))

		_, err := factoryIngress.Networking().V1().Ingresses().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}

		c.factoriesIngress[ns] = factoryIngress

		factoryKube := kinformers.NewSharedInformerFactoryWithOptions(c.clientset, resyncPeriod, kinformers.WithNamespace(ns), kinformers.WithTransform(k8s.StripManagedFields))
		_, err = factoryKube.Core().V1().Services().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		endpointSliceInformer := factoryKube.Discovery().V1().EndpointSlices().Informer()
		if err = endpointSliceInformer.AddIndexers(k8s.EndpointSliceByServiceNameIndexers); err != nil {
			return nil, err
		}
		_, err = endpointSliceInformer.AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		c.factoriesKube[ns] = factoryKube

		factorySecret := kinformers.NewSharedInformerFactoryWithOptions(c.clientset, resyncPeriod, kinformers.WithNamespace(ns), kinformers.WithTweakListOptions(notOwnedByHelm), kinformers.WithTransform(k8s.StripManagedFields))
		_, err = factorySecret.Core().V1().Secrets().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		c.factoriesSecret[ns] = factorySecret

		factoryConfigMap := kinformers.NewSharedInformerFactoryWithOptions(c.clientset, resyncPeriod, kinformers.WithNamespace(ns), kinformers.WithTweakListOptions(notOwnedByHelm), kinformers.WithTransform(k8s.StripManagedFields))
		_, err = factoryConfigMap.Core().V1().ConfigMaps().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		c.factoriesConfigMap[ns] = factoryConfigMap
	}

	if c.podNamespace != "" {
		c.factoryPod = kinformers.NewSharedInformerFactoryWithOptions(c.clientset, resyncPeriod,
			kinformers.WithNamespace(c.podNamespace),
			kinformers.WithTweakListOptions(func(opts *metav1.ListOptions) { opts.LabelSelector = c.podSelector }),
			kinformers.WithTransform(k8s.StripManagedFields))

		if _, err = c.factoryPod.Core().V1().Pods().Informer().AddEventHandler(eventHandler); err != nil {
			return nil, err
		}

		c.factoryPod.Start(stopCh)

		for t, ok := range c.factoryPod.WaitForCacheSync(stopCh) {
			if !ok {
				return nil, fmt.Errorf("timed out waiting for controller caches to sync %s in namespace %q", t.String(), c.podNamespace)
			}
		}
	}

	for _, ns := range c.watchedNamespaces {
		c.factoriesIngress[ns].Start(stopCh)
		c.factoriesKube[ns].Start(stopCh)
		c.factoriesSecret[ns].Start(stopCh)
		c.factoriesConfigMap[ns].Start(stopCh)
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

		for t, ok := range c.factoriesConfigMap[ns].WaitForCacheSync(stopCh) {
			if !ok {
				return nil, fmt.Errorf("timed out waiting for controller caches to sync %s in namespace %q", t.String(), ns)
			}
		}
	}

	c.clusterScopeFactory = kinformers.NewSharedInformerFactoryWithOptions(c.clientset, resyncPeriod, kinformers.WithTransform(k8s.StripManagedFields))

	if !c.ignoreIngressClasses {
		_, err = c.clusterScopeFactory.Networking().V1().IngressClasses().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
	}

	if c.podNamespace != "" {
		_, err = c.clusterScopeFactory.Core().V1().Nodes().Informer().AddEventHandler(eventHandler)
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

// podIdentity returns the name and namespace of the pod this instance runs in.
// The downward API environment variables take precedence, as they are the way
// the ingress-nginx controller is configured, and the in-cluster hostname and
// service account namespace are used when they are not set.
func podIdentity() (string, string, error) {
	name := os.Getenv("POD_NAME")
	if name == "" {
		var err error
		if name, err = os.Hostname(); err != nil {
			return "", "", fmt.Errorf("getting pod name: %w", err)
		}
	}

	if namespace := os.Getenv("POD_NAMESPACE"); namespace != "" {
		return name, namespace, nil
	}

	namespace, err := os.ReadFile(podNamespacePath)
	if err != nil {
		return "", "", fmt.Errorf("getting pod namespace: %w", err)
	}

	return name, string(namespace), nil
}

// GetIngressPodNodeAddresses returns the addresses of the nodes running a ready
// controller pod. An empty result means the addresses cannot be determined,
// either because the controller pods are not discoverable or because none is ready.
func (c *clientWrapper) GetIngressPodNodeAddresses(internalIP bool) ([]string, error) {
	if c.factoryPod == nil {
		return nil, nil
	}

	pods, err := c.factoryPod.Core().V1().Pods().Lister().Pods(c.podNamespace).List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("listing controller pods: %w", err)
	}

	nodeLister := c.clusterScopeFactory.Core().V1().Nodes().Lister()

	var nodeNames []string
	for _, pod := range pods {
		// A pod being deleted still reports as ready while it terminates, and the
		// node it drains from must not be published as an Ingress address anymore.
		if pod.Spec.NodeName == "" || pod.DeletionTimestamp != nil || !isPodReady(pod) {
			continue
		}

		if !slices.Contains(nodeNames, pod.Spec.NodeName) {
			nodeNames = append(nodeNames, pod.Spec.NodeName)
		}
	}

	var addresses []string
	for _, nodeName := range nodeNames {
		node, err := nodeLister.Get(nodeName)
		if err != nil {
			return nil, fmt.Errorf("getting node %s: %w", nodeName, err)
		}

		var internal, external []string
		for _, address := range node.Status.Addresses {
			switch address.Type {
			case corev1.NodeInternalIP:
				internal = append(internal, address.Address)
			case corev1.NodeExternalIP:
				external = append(external, address.Address)
			}
		}

		// The internal addresses are also used as a fallback, as a node without
		// an external address would otherwise not be reported at all.
		if internalIP || len(external) == 0 {
			addresses = append(addresses, internal...)
			continue
		}

		addresses = append(addresses, external...)
	}

	return addresses, nil
}

func isPodReady(pod *corev1.Pod) bool {
	if pod.Status.Phase != corev1.PodRunning {
		return false
	}

	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			return condition.Status == corev1.ConditionTrue
		}
	}

	return false
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

	return k8s.EndpointSlicesByServiceName(
		c.factoriesKube[c.lookupNamespace(namespace)].Discovery().V1().EndpointSlices().Informer().GetIndexer(),
		namespace,
		serviceName,
	)
}

// GetConfigMap returns the named configMap from the given namespace.
func (c *clientWrapper) GetConfigMap(namespace, name string) (*corev1.ConfigMap, error) {
	if !c.isWatchedNamespace(namespace) {
		return nil, fmt.Errorf("failed to get configmap %s/%s: namespace is not within watched namespaces", namespace, name)
	}

	return c.factoriesConfigMap[c.lookupNamespace(namespace)].Core().V1().ConfigMaps().Lister().ConfigMaps(namespace).Get(name)
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
// setupPodDiscovery identifies the controller pods, using the pod this instance
// runs in as the reference: its namespace and labels select the sibling pods.
func (c *clientWrapper) setupPodDiscovery(ctx context.Context) error {
	podName, namespace, err := podIdentity()
	if err != nil {
		return err
	}

	pod, err := c.clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("getting pod %s/%s: %w", namespace, podName, err)
	}

	if len(pod.Labels) == 0 {
		return fmt.Errorf("pod %s/%s has no label to select the controller pods with", pod.Namespace, pod.Name)
	}

	// The status addresses are read from the nodes running the controller pods.
	if _, err = c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{Limit: 1}); err != nil {
		return fmt.Errorf("listing nodes: %w", err)
	}

	c.podNamespace = pod.Namespace
	c.podSelector = labels.Set(pod.Labels).String()

	return nil
}

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
