package kubernetes

import (
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

type clientMock struct {
	ingresses []*v1beta1.Ingress
	services  []*v1.Service
	secrets   []*v1.Secret
	endpoints []*v1.Endpoints
	watchChan chan interface{}

	apiServiceError   error
	apiSecretError    error
	apiEndpointsError error
}

func (c clientMock) GetIngresses() []*v1beta1.Ingress {
	return c.ingresses
}

func (c clientMock) GetService(namespace, name string) (*v1.Service, bool, error) {
	if c.apiServiceError != nil {
		return nil, false, c.apiServiceError
	}

	for _, service := range c.services {
		if service.Namespace == namespace && service.Name == name {
			return service, true, nil
		}
	}
	return nil, false, nil
}

func (c clientMock) GetEndpoints(namespace, name string) (*v1.Endpoints, bool, error) {
	if c.apiEndpointsError != nil {
		return nil, false, c.apiEndpointsError
	}

	for _, endpoints := range c.endpoints {
		if endpoints.Namespace == namespace && endpoints.Name == name {
			return endpoints, true, nil
		}
	}

	return &v1.Endpoints{}, false, nil
}

func (c clientMock) GetSecret(namespace, name string) (*v1.Secret, bool, error) {
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

func (c clientMock) WatchAll(namespaces Namespaces, labelString string, stopCh <-chan struct{}) (<-chan interface{}, error) {
	return c.watchChan, nil
}
