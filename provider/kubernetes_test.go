package provider

import (
	"github.com/containous/traefik/provider/k8s"
	"github.com/containous/traefik/types"
	"reflect"
	"testing"
)

func TestLoadIngresses(t *testing.T) {
	ingresses := []k8s.Ingress{{
		Spec: k8s.IngressSpec{
			Rules: []k8s.IngressRule{
				{
					Host: "foo",
					IngressRuleValue: k8s.IngressRuleValue{
						HTTP: &k8s.HTTPIngressRuleValue{
							Paths: []k8s.HTTPIngressPath{
								{
									Path: "/bar",
									Backend: k8s.IngressBackend{
										ServiceName: "service1",
										ServicePort: k8s.FromInt(801),
									},
								},
							},
						},
					},
				},
				{
					Host: "bar",
					IngressRuleValue: k8s.IngressRuleValue{
						HTTP: &k8s.HTTPIngressRuleValue{
							Paths: []k8s.HTTPIngressPath{
								{
									Backend: k8s.IngressBackend{
										ServiceName: "service3",
										ServicePort: k8s.FromInt(803),
									},
								},
								{
									Backend: k8s.IngressBackend{
										ServiceName: "service2",
										ServicePort: k8s.FromInt(802),
									},
								},
							},
						},
					},
				},
			},
		},
	}}
	services := []k8s.Service{
		{
			ObjectMeta: k8s.ObjectMeta{
				Name: "service1",
				UID:  "1",
			},
			Spec: k8s.ServiceSpec{
				ClusterIP: "10.0.0.1",
				Ports: []k8s.ServicePort{
					{
						Name: "http",
						Port: 801,
					},
				},
			},
		},
		{
			ObjectMeta: k8s.ObjectMeta{
				Name: "service2",
				UID:  "2",
			},
			Spec: k8s.ServiceSpec{
				ClusterIP: "10.0.0.2",
				Ports: []k8s.ServicePort{
					{
						Name: "http",
						Port: 802,
					},
				},
			},
		},
		{
			ObjectMeta: k8s.ObjectMeta{
				Name: "service3",
				UID:  "3",
			},
			Spec: k8s.ServiceSpec{
				ClusterIP: "10.0.0.3",
				Ports: []k8s.ServicePort{
					{
						Name: "http",
						Port: 803,
					},
				},
			},
		},
	}
	watchChan := make(chan interface{})
	client := clientMock{
		ingresses: ingresses,
		services:  services,
		watchChan: watchChan,
	}
	provider := Kubernetes{}
	actual, err := provider.loadIngresses(client)
	if err != nil {
		t.Fatalf("error %+v", err)
	}

	expected := &types.Configuration{
		Backends: map[string]*types.Backend{
			"foo/bar": {
				Servers: map[string]types.Server{
					"1": {
						URL:    "http://10.0.0.1:801",
						Weight: 1,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer:   nil,
			},
			"bar": {
				Servers: map[string]types.Server{
					"2": {
						URL:    "http://10.0.0.2:802",
						Weight: 1,
					},
					"3": {
						URL:    "http://10.0.0.3:803",
						Weight: 1,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer:   nil,
			},
		},
		Frontends: map[string]*types.Frontend{
			"foo/bar": {
				Backend: "foo/bar",
				Routes: map[string]types.Route{
					"/bar": {
						Rule: "PathPrefixStrip:/bar",
					},
					"foo": {
						Rule: "Host:foo",
					},
				},
			},
			"bar": {
				Backend: "bar",
				Routes: map[string]types.Route{
					"bar": {
						Rule: "Host:bar",
					},
				},
			},
		},
	}
	if !reflect.DeepEqual(actual.Backends, expected.Backends) {
		t.Fatalf("expected %+v, got %+v", expected.Backends, actual.Backends)
	}
	if !reflect.DeepEqual(actual.Frontends, expected.Frontends) {
		t.Fatalf("expected %+v, got %+v", expected.Frontends, actual.Frontends)
	}
}

type clientMock struct {
	ingresses []k8s.Ingress
	services  []k8s.Service
	watchChan chan interface{}
}

func (c clientMock) GetIngresses(predicate func(k8s.Ingress) bool) ([]k8s.Ingress, error) {
	return c.ingresses, nil
}
func (c clientMock) WatchIngresses(predicate func(k8s.Ingress) bool, stopCh <-chan bool) (chan interface{}, chan error, error) {
	return c.watchChan, make(chan error), nil
}
func (c clientMock) GetServices(namespace string, predicate func(k8s.Service) bool) ([]k8s.Service, error) {
	return c.services, nil
}
