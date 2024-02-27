package ingress

import (
	"fmt"
	"os"

	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/k8s"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
)

var _ Client = (*clientMock)(nil)

type clientMock struct {
	ingresses      []*netv1.Ingress
	services       []*corev1.Service
	secrets        []*corev1.Secret
	endpoints      []*corev1.Endpoints
	nodes          []*corev1.Node
	ingressClasses []*netv1.IngressClass

	apiServiceError       error
	apiSecretError        error
	apiEndpointsError     error
	apiNodesError         error
	apiIngressStatusError error

	watchChan chan interface{}
}

func newClientMock(path string) clientMock {
	c := clientMock{}

	yamlContent, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("unable to read file %q: %w", path, err))
	}

	k8sObjects := k8s.MustParseYaml(yamlContent)
	for _, obj := range k8sObjects {
		switch o := obj.(type) {
		case *corev1.Service:
			c.services = append(c.services, o)
		case *corev1.Secret:
			c.secrets = append(c.secrets, o)
		case *corev1.Endpoints:
			c.endpoints = append(c.endpoints, o)
		case *corev1.Node:
			c.nodes = append(c.nodes, o)
		case *netv1.Ingress:
			c.ingresses = append(c.ingresses, o)
		case *netv1.IngressClass:
			c.ingressClasses = append(c.ingressClasses, o)
		default:
			panic(fmt.Sprintf("Unknown runtime object %+v %T", o, o))
		}
	}

	return c
}

func (c clientMock) GetIngresses() []*netv1.Ingress {
	return c.ingresses
}

func (c clientMock) GetService(namespace, name string) (*corev1.Service, bool, error) {
	if c.apiServiceError != nil {
		return nil, false, c.apiServiceError
	}

	for _, service := range c.services {
		if service.Namespace == namespace && service.Name == name {
			return service, true, nil
		}
	}
	return nil, false, c.apiServiceError
}

func (c clientMock) GetEndpoints(namespace, name string) (*corev1.Endpoints, bool, error) {
	if c.apiEndpointsError != nil {
		return nil, false, c.apiEndpointsError
	}

	for _, endpoints := range c.endpoints {
		if endpoints.Namespace == namespace && endpoints.Name == name {
			return endpoints, true, nil
		}
	}

	return &corev1.Endpoints{}, false, nil
}

func (c clientMock) GetNodes() ([]*corev1.Node, bool, error) {
	if c.apiNodesError != nil {
		return nil, false, c.apiNodesError
	}

	return c.nodes, true, nil
}

func (c clientMock) GetSecret(namespace, name string) (*corev1.Secret, bool, error) {
	if c.apiSecretError != nil {
		return nil, false, c.apiSecretError
	}

	for _, secret := range c.secrets {
		if secret.Namespace == namespace && secret.Name == name {
			return secret, true, nil
		}
	}
	return nil, false, nil
}

func (c clientMock) GetIngressClasses() ([]*netv1.IngressClass, error) {
	return c.ingressClasses, nil
}

func (c clientMock) WatchAll(namespaces []string, stopCh <-chan struct{}) (<-chan interface{}, error) {
	return c.watchChan, nil
}

func (c clientMock) UpdateIngressStatus(_ *netv1.Ingress, _ []netv1.IngressLoadBalancerIngress) error {
	return c.apiIngressStatusError
}
