package kubernetes

import (
	"fmt"
	"io/ioutil"

	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	v1beta12 "k8s.io/api/extensions/v1beta1"
)

var _ Client = (*clientMock)(nil)

type clientMock struct {
	ingresses []*extensionsv1beta1.Ingress
	services  []*corev1.Service
	secrets   []*corev1.Secret
	endpoints []*corev1.Endpoints

	apiServiceError       error
	apiSecretError        error
	apiEndpointsError     error
	apiIngressStatusError error

	watchChan chan interface{}
}

func newClientMock(paths ...string) clientMock {
	var c clientMock

	for _, path := range paths {
		yamlContent, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}

		k8sObjects := MustDecodeYaml(yamlContent)
		for _, obj := range k8sObjects {
			switch o := obj.(type) {
			case *corev1.Service:
				c.services = append(c.services, o)
			case *corev1.Secret:
				c.secrets = append(c.secrets, o)
			case *corev1.Endpoints:
				c.endpoints = append(c.endpoints, o)
			case *v1beta12.Ingress:
				c.ingresses = append(c.ingresses, o)
			default:
				panic(fmt.Sprintf("Unknown runtime object %+v %T", o, o))
			}
		}
	}

	return c
}

func (c clientMock) GetIngresses() []*extensionsv1beta1.Ingress {
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

func (c clientMock) WatchAll(namespaces Namespaces, stopCh <-chan struct{}) (<-chan interface{}, error) {
	return c.watchChan, nil
}

func (c clientMock) UpdateIngressStatus(namespace, name, ip, hostname string) error {
	return c.apiIngressStatusError
}
