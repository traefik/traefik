package gateway

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/traefik/traefik/v2/pkg/provider/kubernetes/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kscheme "k8s.io/client-go/kubernetes/scheme"
	gatev1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

var _ Client = (*clientMock)(nil)

func init() {
	// required by k8s.MustParseYaml
	err := gatev1alpha2.AddToScheme(kscheme.Scheme)
	if err != nil {
		panic(err)
	}
}

type clientMock struct {
	services   []*corev1.Service
	secrets    []*corev1.Secret
	endpoints  []*corev1.Endpoints
	namespaces []*corev1.Namespace

	apiServiceError   error
	apiSecretError    error
	apiEndpointsError error

	gatewayClasses []*gatev1alpha2.GatewayClass
	gateways       []*gatev1alpha2.Gateway
	httpRoutes     []*gatev1alpha2.HTTPRoute
	tcpRoutes      []*gatev1alpha2.TCPRoute
	tlsRoutes      []*gatev1alpha2.TLSRoute

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
			case *corev1.Secret:
				c.secrets = append(c.secrets, o)
			case *corev1.Namespace:
				c.namespaces = append(c.namespaces, o)
			case *corev1.Endpoints:
				c.endpoints = append(c.endpoints, o)
			case *gatev1alpha2.GatewayClass:
				c.gatewayClasses = append(c.gatewayClasses, o)
			case *gatev1alpha2.Gateway:
				c.gateways = append(c.gateways, o)
			case *gatev1alpha2.HTTPRoute:
				c.httpRoutes = append(c.httpRoutes, o)
			case *gatev1alpha2.TCPRoute:
				c.tcpRoutes = append(c.tcpRoutes, o)
			case *gatev1alpha2.TLSRoute:
				c.tlsRoutes = append(c.tlsRoutes, o)
			default:
				panic(fmt.Sprintf("Unknown runtime object %+v %T", o, o))
			}
		}
	}

	return c
}

func (c clientMock) UpdateGatewayStatus(gateway *gatev1alpha2.Gateway, gatewayStatus gatev1alpha2.GatewayStatus) error {
	for _, g := range c.gateways {
		if g.Name == gateway.Name {
			if !statusEquals(g.Status, gatewayStatus) {
				g.Status = gatewayStatus
				return nil
			}
			return fmt.Errorf("cannot update gateway %v", gateway.Name)
		}
	}
	return nil
}

func (c clientMock) UpdateGatewayClassStatus(gatewayClass *gatev1alpha2.GatewayClass, condition metav1.Condition) error {
	for _, gc := range c.gatewayClasses {
		if gc.Name == gatewayClass.Name {
			for _, c := range gc.Status.Conditions {
				if c.Type == condition.Type && c.Status != condition.Status {
					c.Status = condition.Status
					c.LastTransitionTime = condition.LastTransitionTime
					c.Message = condition.Message
					c.Reason = condition.Reason
				}
			}
		}
	}
	return nil
}

func (c clientMock) UpdateGatewayStatusConditions(gateway *gatev1alpha2.Gateway, condition metav1.Condition) error {
	for _, g := range c.gatewayClasses {
		if g.Name == gateway.Name {
			for _, c := range g.Status.Conditions {
				if c.Type == condition.Type && (c.Status != condition.Status || c.Reason != condition.Reason) {
					c.Status = condition.Status
					c.LastTransitionTime = condition.LastTransitionTime
					c.Message = condition.Message
					c.Reason = condition.Reason
				}
			}
		}
	}
	return nil
}

func (c clientMock) GetGatewayClasses() ([]*gatev1alpha2.GatewayClass, error) {
	return c.gatewayClasses, nil
}

func (c clientMock) GetGateways() []*gatev1alpha2.Gateway {
	return c.gateways
}

func inNamespace(m metav1.ObjectMeta, s string) bool {
	return s == metav1.NamespaceAll || m.Namespace == s
}

func (c clientMock) GetNamespaces(selector labels.Selector) ([]string, error) {
	var ns []string
	for _, namespace := range c.namespaces {
		if selector.Matches(labels.Set(namespace.Labels)) {
			ns = append(ns, namespace.Name)
		}
	}
	return ns, nil
}

func (c clientMock) GetHTTPRoutes(namespaces []string) ([]*gatev1alpha2.HTTPRoute, error) {
	var httpRoutes []*gatev1alpha2.HTTPRoute
	for _, namespace := range namespaces {
		for _, httpRoute := range c.httpRoutes {
			if inNamespace(httpRoute.ObjectMeta, namespace) {
				httpRoutes = append(httpRoutes, httpRoute)
			}
		}
	}
	return httpRoutes, nil
}

func (c clientMock) GetTCPRoutes(namespaces []string) ([]*gatev1alpha2.TCPRoute, error) {
	var tcpRoutes []*gatev1alpha2.TCPRoute
	for _, namespace := range namespaces {
		for _, tcpRoute := range c.tcpRoutes {
			if inNamespace(tcpRoute.ObjectMeta, namespace) {
				tcpRoutes = append(tcpRoutes, tcpRoute)
			}
		}
	}
	return tcpRoutes, nil
}

func (c clientMock) GetTLSRoutes(namespaces []string) ([]*gatev1alpha2.TLSRoute, error) {
	var tlsRoutes []*gatev1alpha2.TLSRoute
	for _, namespace := range namespaces {
		for _, tlsRoute := range c.tlsRoutes {
			if inNamespace(tlsRoute.ObjectMeta, namespace) {
				tlsRoutes = append(tlsRoutes, tlsRoute)
			}
		}
	}
	return tlsRoutes, nil
}

func (c clientMock) GetService(namespace, name string) (*corev1.Service, bool, error) {
	if c.apiServiceError != nil {
		return nil, false, c.apiServiceError
	}

	for _, service := range c.services {
		if inNamespace(service.ObjectMeta, namespace) && service.Name == name {
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
		if inNamespace(endpoints.ObjectMeta, namespace) && endpoints.Name == name {
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
		if inNamespace(secret.ObjectMeta, namespace) && secret.Name == name {
			return secret, true, nil
		}
	}
	return nil, false, nil
}

func (c clientMock) WatchAll(namespaces []string, stopCh <-chan struct{}) (<-chan interface{}, error) {
	return c.watchChan, nil
}
