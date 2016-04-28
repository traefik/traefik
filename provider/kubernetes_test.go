package provider

import (
	"encoding/json"
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
										ServicePort: k8s.FromInt(443),
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
						Port: 443,
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
						URL:    "https://10.0.0.3:443",
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
	actualJSON, _ := json.Marshal(actual)
	expectedJSON, _ := json.Marshal(expected)

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected %+v, got %+v", string(expectedJSON), string(actualJSON))
	}
}

func TestLoadNamespacedIngresses(t *testing.T) {
	ingresses := []k8s.Ingress{
		{
			ObjectMeta: k8s.ObjectMeta{
				Namespace: "awesome",
			},
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
											ServicePort: k8s.FromInt(443),
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
		},
		{
			ObjectMeta: k8s.ObjectMeta{
				Namespace: "not-awesome",
			},
			Spec: k8s.IngressSpec{
				Rules: []k8s.IngressRule{
					{
						Host: "baz",
						IngressRuleValue: k8s.IngressRuleValue{
							HTTP: &k8s.HTTPIngressRuleValue{
								Paths: []k8s.HTTPIngressPath{
									{
										Path: "/baz",
										Backend: k8s.IngressBackend{
											ServiceName: "service1",
											ServicePort: k8s.FromInt(801),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
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
						Port: 443,
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
	provider := Kubernetes{
		Namespaces: []string{"awesome"},
	}
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
						URL:    "https://10.0.0.3:443",
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
	actualJSON, _ := json.Marshal(actual)
	expectedJSON, _ := json.Marshal(expected)

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected %+v, got %+v", string(expectedJSON), string(actualJSON))
	}
}

func TestLoadMultipleNamespacedIngresses(t *testing.T) {
	ingresses := []k8s.Ingress{
		{
			ObjectMeta: k8s.ObjectMeta{
				Namespace: "awesome",
			},
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
											ServicePort: k8s.FromInt(443),
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
		},
		{
			ObjectMeta: k8s.ObjectMeta{
				Namespace: "somewhat-awesome",
			},
			Spec: k8s.IngressSpec{
				Rules: []k8s.IngressRule{
					{
						Host: "awesome",
						IngressRuleValue: k8s.IngressRuleValue{
							HTTP: &k8s.HTTPIngressRuleValue{
								Paths: []k8s.HTTPIngressPath{
									{
										Path: "/quix",
										Backend: k8s.IngressBackend{
											ServiceName: "service1",
											ServicePort: k8s.FromInt(801),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			ObjectMeta: k8s.ObjectMeta{
				Namespace: "not-awesome",
			},
			Spec: k8s.IngressSpec{
				Rules: []k8s.IngressRule{
					{
						Host: "baz",
						IngressRuleValue: k8s.IngressRuleValue{
							HTTP: &k8s.HTTPIngressRuleValue{
								Paths: []k8s.HTTPIngressPath{
									{
										Path: "/baz",
										Backend: k8s.IngressBackend{
											ServiceName: "service1",
											ServicePort: k8s.FromInt(801),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
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
						Port: 443,
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
	provider := Kubernetes{
		Namespaces: []string{"awesome", "somewhat-awesome"},
	}
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
						URL:    "https://10.0.0.3:443",
						Weight: 1,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer:   nil,
			},
			"awesome/quix": {
				Servers: map[string]types.Server{
					"1": {
						URL:    "http://10.0.0.1:801",
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
			"awesome/quix": {
				Backend: "awesome/quix",
				Routes: map[string]types.Route{
					"/quix": {
						Rule: "PathPrefixStrip:/quix",
					},
					"awesome": {
						Rule: "Host:awesome",
					},
				},
			},
		},
	}
	actualJSON, _ := json.Marshal(actual)
	expectedJSON, _ := json.Marshal(expected)

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected %+v, got %+v", string(expectedJSON), string(actualJSON))
	}
}

type clientMock struct {
	ingresses []k8s.Ingress
	services  []k8s.Service
	watchChan chan interface{}
}

func (c clientMock) GetIngresses(predicate func(k8s.Ingress) bool) ([]k8s.Ingress, error) {
	var ingresses []k8s.Ingress
	for _, ingress := range c.ingresses {
		if predicate(ingress) {
			ingresses = append(ingresses, ingress)
		}
	}
	return ingresses, nil
}
func (c clientMock) WatchIngresses(predicate func(k8s.Ingress) bool, stopCh <-chan bool) (chan interface{}, chan error, error) {
	return c.watchChan, make(chan error), nil
}
func (c clientMock) GetServices(predicate func(k8s.Service) bool) ([]k8s.Service, error) {
	return c.services, nil
}
func (c clientMock) WatchAll(stopCh <-chan bool) (chan interface{}, chan error, error) {
	return c.watchChan, make(chan error), nil
}
