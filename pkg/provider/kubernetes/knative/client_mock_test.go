package knative

import (
	corev1 "k8s.io/api/core/v1"
	knativenetworkingv1alpha1 "knative.dev/networking/pkg/apis/networking/v1alpha1"
)

// clientMock is a mock implementation of the client interface.
type clientMock struct {
	services           []*corev1.Service
	serverlessServices []*knativenetworkingv1alpha1.ServerlessService
	ingresses          []*knativenetworkingv1alpha1.Ingress

	apiServiceError error
}

func (m *clientMock) WatchAll(namespaces []string, stopCh <-chan struct{}) (<-chan interface{}, error) {
	panic("implement me")
}

func (m *clientMock) ListIngresses() []*knativenetworkingv1alpha1.Ingress {
	return m.ingresses
}

func (m *clientMock) GetService(namespace, name string) (*corev1.Service, error) {
	for _, service := range m.services {
		if service.Namespace == namespace && service.Name == name {
			return service, nil
		}
	}
	return nil, m.apiServiceError
}

func (m *clientMock) GetSecret(namespace, name string) (*corev1.Secret, error) {
	// TODO implement me
	panic("implement me")
}

func (m *clientMock) UpdateIngressStatus(ingress *knativenetworkingv1alpha1.Ingress) error {
	return nil
}
