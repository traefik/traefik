package knative

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kinformers "k8s.io/client-go/informers"
	kclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	knativenetworkingv1alpha1 "knative.dev/networking/pkg/apis/networking/v1alpha1"
	knativenetworkingclientset "knative.dev/networking/pkg/client/clientset/versioned"
	knativenetworkinginformers "knative.dev/networking/pkg/client/informers/externalversions"
)

const resyncPeriod = 10 * time.Minute

type clientWrapper struct {
	csKnativeNetworking knativenetworkingclientset.Interface
	csKube              kclientset.Interface

	factoriesKnativeNetworking map[string]knativenetworkinginformers.SharedInformerFactory
	factoriesKube              map[string]kinformers.SharedInformerFactory

	labelSelector string

	isNamespaceAll    bool
	watchedNamespaces []string
}

func createClientFromConfig(c *rest.Config) (*clientWrapper, error) {
	csKnativeNetworking, err := knativenetworkingclientset.NewForConfig(c)
	if err != nil {
		return nil, err
	}

	csKube, err := kclientset.NewForConfig(c)
	if err != nil {
		return nil, err
	}

	return newClientImpl(csKnativeNetworking, csKube), nil
}

func newClientImpl(csKnativeNetworking knativenetworkingclientset.Interface, csKube kclientset.Interface) *clientWrapper {
	return &clientWrapper{
		csKnativeNetworking:        csKnativeNetworking,
		csKube:                     csKube,
		factoriesKnativeNetworking: make(map[string]knativenetworkinginformers.SharedInformerFactory),
		factoriesKube:              make(map[string]kinformers.SharedInformerFactory),
	}
}

// newInClusterClient returns a new Provider client that is expected to run
// inside the cluster.
func newInClusterClient(endpoint string) (*clientWrapper, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("creating in-cluster configuration: %w", err)
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
		caData, err := os.ReadFile(caFilePath)
		if err != nil {
			return nil, fmt.Errorf("reading CA file %s: %w", caFilePath, err)
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

	for _, ns := range namespaces {
		factory := knativenetworkinginformers.NewSharedInformerFactoryWithOptions(c.csKnativeNetworking, resyncPeriod, knativenetworkinginformers.WithNamespace(ns), knativenetworkinginformers.WithTweakListOptions(func(opts *metav1.ListOptions) {
			opts.LabelSelector = c.labelSelector
		}))
		_, err := factory.Networking().V1alpha1().Ingresses().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}

		factoryKube := kinformers.NewSharedInformerFactoryWithOptions(c.csKube, resyncPeriod, kinformers.WithNamespace(ns))
		_, err = factoryKube.Core().V1().Services().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}
		_, err = factoryKube.Core().V1().Secrets().Informer().AddEventHandler(eventHandler)
		if err != nil {
			return nil, err
		}

		c.factoriesKube[ns] = factoryKube
		c.factoriesKnativeNetworking[ns] = factory
	}

	for _, ns := range namespaces {
		c.factoriesKnativeNetworking[ns].Start(stopCh)
		c.factoriesKube[ns].Start(stopCh)
	}

	for _, ns := range namespaces {
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

func (c *clientWrapper) ListIngresses() []*knativenetworkingv1alpha1.Ingress {
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

func (c *clientWrapper) UpdateIngressStatus(ingress *knativenetworkingv1alpha1.Ingress) error {
	_, err := c.csKnativeNetworking.NetworkingV1alpha1().Ingresses(ingress.Namespace).UpdateStatus(context.TODO(), ingress, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("updating knative ingress status %s/%s: %w", ingress.Namespace, ingress.Name, err)
	}

	log.Info().Msgf("Updated status on knative ingress %s/%s", ingress.Namespace, ingress.Name)
	return nil
}

// GetService returns the named service from the given namespace.
func (c *clientWrapper) GetService(namespace, name string) (*corev1.Service, error) {
	if !c.isWatchedNamespace(namespace) {
		return nil, fmt.Errorf("getting service %s/%s: namespace is not within watched namespaces", namespace, name)
	}

	return c.factoriesKube[c.lookupNamespace(namespace)].Core().V1().Services().Lister().Services(namespace).Get(name)
}

// GetSecret returns the named secret from the given namespace.
func (c *clientWrapper) GetSecret(namespace, name string) (*corev1.Secret, error) {
	if !c.isWatchedNamespace(namespace) {
		return nil, fmt.Errorf("getting secret %s/%s: namespace is not within watched namespaces", namespace, name)
	}

	return c.factoriesKube[c.lookupNamespace(namespace)].Core().V1().Secrets().Lister().Secrets(namespace).Get(name)
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
