package knative

import (
	"errors"

	corev1 "k8s.io/api/core/v1"
	knativenetworkingv1alpha1 "knative.dev/networking/pkg/apis/networking/v1alpha1"
)

// clientMock is a mock implementation of the client interface.
type clientMock struct {
	services          []*corev1.Service
	serverlessService []*knativenetworkingv1alpha1.ServerlessService
	ingressRoute      []*knativenetworkingv1alpha1.Ingress

	apiServiceError        error
	serverlessServiceError error
}

func (m *clientMock) WatchAll(namespaces []string, stopCh <-chan struct{}) (<-chan interface{}, error) {
	panic("implement me")
}

func (m *clientMock) ListIngresses() []*knativenetworkingv1alpha1.Ingress {
	return m.ingressRoute
}

func (m *clientMock) GetIngress(namespace, name string) (*knativenetworkingv1alpha1.Ingress, bool, error) {
	panic("implement me")
}

func (m *clientMock) GetServerlessService(namespace, name string) (*knativenetworkingv1alpha1.ServerlessService, bool, error) {
	if len(m.services) == 0 {
		return nil, false, errors.New("no services found")
	}

	for _, service := range m.serverlessService {
		if service.Namespace == namespace && service.Name == name {
			return service, true, nil
		}
	}
	return nil, false, m.serverlessServiceError
}

func (m *clientMock) GetService(namespace, name string) (*corev1.Service, bool, error) {
	for _, service := range m.services {
		if service.Namespace == namespace && service.Name == name {
			return service, true, nil
		}
	}
	return nil, false, m.apiServiceError
}

func (m *clientMock) GetSecret(namespace, name string) (*corev1.Secret, bool, error) {
	// TODO implement me
	panic("implement me")
}

func (m *clientMock) GetEndpoints(namespace, name string) (*corev1.Endpoints, bool, error) {
	// TODO implement me
	panic("implement me")
}

func (m *clientMock) UpdateIngressStatus(ingress *knativenetworkingv1alpha1.Ingress) error {
	return nil
}
