package knative

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	knativenetworkingv1alpha1 "knative.dev/networking/pkg/apis/networking/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MockClient is a mock implementation of the client interface
type MockClient struct {
	services          []*corev1.Service
	serverlessService []*knativenetworkingv1alpha1.ServerlessService
	ingressRoute      []*knativenetworkingv1alpha1.Ingress

	apiServiceError        error
	serverlessServiceError error
}

func (m *MockClient) WatchAll(namespaces []string, stopCh <-chan struct{}) (<-chan interface{}, error) {
	panic("implement me")
}

func (m *MockClient) GetKnativeIngressRoute(namespace, name string) (*knativenetworkingv1alpha1.Ingress, bool, error) {
	panic("implement me")
}

func (m *MockClient) UpdateKnativeIngressStatus(ingress *knativenetworkingv1alpha1.Ingress) error {
	return nil
}

func (m *MockClient) GetKnativeIngressRoutes() []*knativenetworkingv1alpha1.Ingress {
	return m.ingressRoute
}

func (m *MockClient) GetServerlessService(namespace, name string) (*knativenetworkingv1alpha1.ServerlessService, bool, error) {
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

func (m *MockClient) GetService(namespace, name string) (*corev1.Service, bool, error) {
	for _, service := range m.services {
		if service.Namespace == namespace && service.Name == name {
			return service, true, nil
		}
	}
	return nil, false, m.apiServiceError
}

func (m *MockClient) GetSecret(namespace, name string) (*corev1.Secret, bool, error) {
	// TODO implement me
	panic("implement me")
}

func (m *MockClient) GetEndpoints(namespace, name string) (*corev1.Endpoints, bool, error) {
	// TODO implement me
	panic("implement me")
}

// Implement the necessary methods for MockClient
// For example, if the client interface has a Get method, you can mock it like this:
func (m *MockClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	// Mock implementation
	return nil
}
