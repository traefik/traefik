package crd

import (
	"fmt"
	"os"
	"path/filepath"

	traefikv1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	"github.com/traefik/traefik/v2/pkg/provider/kubernetes/k8s"
	corev1 "k8s.io/api/core/v1"
	kscheme "k8s.io/client-go/kubernetes/scheme"
)

var _ Client = (*clientMock)(nil)

func init() {
	// required by k8s.MustParseYaml
	err := traefikv1alpha1.AddToScheme(kscheme.Scheme)
	if err != nil {
		panic(err)
	}
}

type clientMock struct {
	services  []*corev1.Service
	secrets   []*corev1.Secret
	endpoints []*corev1.Endpoints

	apiServiceError   error
	apiSecretError    error
	apiEndpointsError error

	ingressRoutes    []*traefikv1alpha1.IngressRoute
	ingressRouteTCPs []*traefikv1alpha1.IngressRouteTCP
	ingressRouteUDPs []*traefikv1alpha1.IngressRouteUDP
	middlewares      []*traefikv1alpha1.Middleware
	middlewareTCPs   []*traefikv1alpha1.MiddlewareTCP
	tlsOptions       []*traefikv1alpha1.TLSOption
	tlsStores        []*traefikv1alpha1.TLSStore
	traefikServices  []*traefikv1alpha1.TraefikService
	serversTransport []*traefikv1alpha1.ServersTransport

	watchChan chan interface{}
}

func newClientMock(paths ...string) clientMock {
	var c clientMock

	for _, path := range paths {
		yamlContent, err := os.ReadFile(filepath.FromSlash("./fixtures/" + path))
		if err != nil {
			panic(err)
		}

		k8sObjects := k8s.MustParseYaml(yamlContent)
		for _, obj := range k8sObjects {
			switch o := obj.(type) {
			case *corev1.Service:
				c.services = append(c.services, o)
			case *corev1.Endpoints:
				c.endpoints = append(c.endpoints, o)
			case *traefikv1alpha1.IngressRoute:
				c.ingressRoutes = append(c.ingressRoutes, o)
			case *traefikv1alpha1.IngressRouteTCP:
				c.ingressRouteTCPs = append(c.ingressRouteTCPs, o)
			case *traefikv1alpha1.IngressRouteUDP:
				c.ingressRouteUDPs = append(c.ingressRouteUDPs, o)
			case *traefikv1alpha1.Middleware:
				c.middlewares = append(c.middlewares, o)
			case *traefikv1alpha1.MiddlewareTCP:
				c.middlewareTCPs = append(c.middlewareTCPs, o)
			case *traefikv1alpha1.TraefikService:
				c.traefikServices = append(c.traefikServices, o)
			case *traefikv1alpha1.TLSOption:
				c.tlsOptions = append(c.tlsOptions, o)
			case *traefikv1alpha1.ServersTransport:
				c.serversTransport = append(c.serversTransport, o)
			case *traefikv1alpha1.TLSStore:
				c.tlsStores = append(c.tlsStores, o)
			case *corev1.Secret:
				c.secrets = append(c.secrets, o)
			default:
				panic(fmt.Sprintf("Unknown runtime object %+v %T", o, o))
			}
		}
	}

	return c
}

func (c clientMock) GetIngressRoutes() []*traefikv1alpha1.IngressRoute {
	return c.ingressRoutes
}

func (c clientMock) GetIngressRouteTCPs() []*traefikv1alpha1.IngressRouteTCP {
	return c.ingressRouteTCPs
}

func (c clientMock) GetIngressRouteUDPs() []*traefikv1alpha1.IngressRouteUDP {
	return c.ingressRouteUDPs
}

func (c clientMock) GetMiddlewares() []*traefikv1alpha1.Middleware {
	return c.middlewares
}

func (c clientMock) GetMiddlewareTCPs() []*traefikv1alpha1.MiddlewareTCP {
	return c.middlewareTCPs
}

func (c clientMock) GetTraefikService(namespace, name string) (*traefikv1alpha1.TraefikService, bool, error) {
	for _, svc := range c.traefikServices {
		if svc.Namespace == namespace && svc.Name == name {
			return svc, true, nil
		}
	}

	return nil, false, nil
}

func (c clientMock) GetTraefikServices() []*traefikv1alpha1.TraefikService {
	return c.traefikServices
}

func (c clientMock) GetTLSOptions() []*traefikv1alpha1.TLSOption {
	return c.tlsOptions
}

func (c clientMock) GetTLSStores() []*traefikv1alpha1.TLSStore {
	return c.tlsStores
}

func (c clientMock) GetServersTransports() []*traefikv1alpha1.ServersTransport {
	return c.serversTransport
}

func (c clientMock) GetTLSOption(namespace, name string) (*traefikv1alpha1.TLSOption, bool, error) {
	for _, option := range c.tlsOptions {
		if option.Namespace == namespace && option.Name == name {
			return option, true, nil
		}
	}

	return nil, false, nil
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

func (c clientMock) WatchAll(namespaces []string, stopCh <-chan struct{}) (<-chan interface{}, error) {
	return c.watchChan, nil
}
