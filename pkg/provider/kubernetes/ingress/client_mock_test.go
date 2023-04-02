package ingress

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-version"
	"github.com/traefik/traefik/v2/pkg/provider/kubernetes/k8s"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	netv1beta1 "k8s.io/api/networking/v1beta1"
)

var _ Client = (*clientMock)(nil)

type clientMock struct {
	ingresses      []*netv1.Ingress
	services       []*corev1.Service
	secrets        []*corev1.Secret
	endpoints      []*corev1.Endpoints
	ingressClasses []*netv1.IngressClass

	serverVersion *version.Version

	apiServiceError       error
	apiSecretError        error
	apiEndpointsError     error
	apiIngressStatusError error

	watchChan chan interface{}
}

func newClientMock(serverVersion string, paths ...string) clientMock {
	c := clientMock{}

	c.serverVersion = version.Must(version.NewVersion(serverVersion))

	for _, path := range paths {
		yamlContent, err := os.ReadFile(path)
		if err != nil {
			panic(err)
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
			case *netv1beta1.Ingress:
				ing, err := convert[netv1.Ingress](o)
				if err != nil {
					panic(err)
				}
				addServiceFromV1Beta1(ing, *o)
				c.ingresses = append(c.ingresses, ing)
			case *netv1.Ingress:
				c.ingresses = append(c.ingresses, o)
			case *netv1beta1.IngressClass:
				ic, err := convert[netv1.IngressClass](o)
				if err != nil {
					panic(err)
				}
				c.ingressClasses = append(c.ingressClasses, ic)
			case *netv1.IngressClass:
				c.ingressClasses = append(c.ingressClasses, o)
			default:
				panic(fmt.Sprintf("Unknown runtime object %+v %T", o, o))
			}
		}
	}

	return c
}

func (c clientMock) GetIngresses() []*netv1.Ingress {
	return c.ingresses
}

func (c clientMock) GetServerVersion() *version.Version {
	return c.serverVersion
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

func (c clientMock) GetIngressClasses() ([]*netv1.IngressClass, error) {
	return c.ingressClasses, nil
}

func (c clientMock) WatchAll(namespaces []string, stopCh <-chan struct{}) (<-chan interface{}, error) {
	return c.watchChan, nil
}

func (c clientMock) UpdateIngressStatus(_ *netv1.Ingress, _ []netv1.IngressLoadBalancerIngress) error {
	return c.apiIngressStatusError
}
