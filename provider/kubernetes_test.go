package provider

import (
	"encoding/json"
	"reflect"
	"testing"

	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/1.5/pkg/util/intstr"

	"github.com/containous/traefik/provider/k8s"
	"github.com/containous/traefik/types"
)

func TestLoadIngresses(t *testing.T) {
	ingresses := []*v1beta1.Ingress{{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "testing",
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{
					Host: "foo",
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Path: "/bar",
									Backend: v1beta1.IngressBackend{
										ServiceName: "service1",
										ServicePort: intstr.FromInt(80),
									},
								},
							},
						},
					},
				},
				{
					Host: "bar",
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Backend: v1beta1.IngressBackend{
										ServiceName: "service3",
										ServicePort: intstr.FromString("https"),
									},
								},
								{
									Backend: v1beta1.IngressBackend{
										ServiceName: "service2",
										ServicePort: intstr.FromInt(802),
									},
								},
							},
						},
					},
				},
			},
		},
	}}
	services := []*v1.Service{
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      "service1",
				UID:       "1",
				Namespace: "testing",
			},
			Spec: v1.ServiceSpec{
				ClusterIP: "10.0.0.1",
				Ports: []v1.ServicePort{
					{
						Port: 80,
					},
				},
			},
		},
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      "service2",
				UID:       "2",
				Namespace: "testing",
			},
			Spec: v1.ServiceSpec{
				ClusterIP: "10.0.0.2",
				Ports: []v1.ServicePort{
					{
						Port: 802,
					},
				},
			},
		},
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      "service3",
				UID:       "3",
				Namespace: "testing",
			},
			Spec: v1.ServiceSpec{
				ClusterIP: "10.0.0.3",
				Ports: []v1.ServicePort{
					{
						Name: "http",
						Port: 80,
					},
					{
						Name: "https",
						Port: 443,
					},
				},
			},
		},
	}
	endpoints := []*v1.Endpoints{
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      "service1",
				UID:       "1",
				Namespace: "testing",
			},
			Subsets: []v1.EndpointSubset{
				{
					Addresses: []v1.EndpointAddress{
						{
							IP: "10.10.0.1",
						},
					},
					Ports: []v1.EndpointPort{
						{
							Port: 8080,
						},
					},
				},
				{
					Addresses: []v1.EndpointAddress{
						{
							IP: "10.21.0.1",
						},
					},
					Ports: []v1.EndpointPort{
						{
							Port: 8080,
						},
					},
				},
			},
		},
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      "service3",
				UID:       "3",
				Namespace: "testing",
			},
			Subsets: []v1.EndpointSubset{
				{
					Addresses: []v1.EndpointAddress{
						{
							IP: "10.15.0.1",
						},
					},
					Ports: []v1.EndpointPort{
						{
							Name: "http",
							Port: 8080,
						},
						{
							Name: "https",
							Port: 8443,
						},
					},
				},
				{
					Addresses: []v1.EndpointAddress{
						{
							IP: "10.15.0.2",
						},
					},
					Ports: []v1.EndpointPort{
						{
							Name: "http",
							Port: 9080,
						},
						{
							Name: "https",
							Port: 9443,
						},
					},
				},
			},
		},
	}
	watchChan := make(chan interface{})
	client := clientMock{
		ingresses: ingresses,
		services:  services,
		endpoints: endpoints,
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
					"http://10.10.0.1:8080": {
						URL:    "http://10.10.0.1:8080",
						Weight: 0,
					},
					"http://10.21.0.1:8080": {
						URL:    "http://10.21.0.1:8080",
						Weight: 0,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer:   nil,
			},
			"bar": {
				Servers: map[string]types.Server{
					"2": {
						URL:    "http://10.0.0.2:802",
						Weight: 0,
					},
					"https://10.15.0.1:8443": {
						URL:    "https://10.15.0.1:8443",
						Weight: 0,
					},
					"https://10.15.0.2:9443": {
						URL:    "https://10.15.0.2:9443",
						Weight: 0,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer:   nil,
			},
		},
		Frontends: map[string]*types.Frontend{
			"foo/bar": {
				Backend:        "foo/bar",
				PassHostHeader: true,
				Priority:       len("/bar"),
				Routes: map[string]types.Route{
					"/bar": {
						Rule: "PathPrefix:/bar",
					},
					"foo": {
						Rule: "Host:foo",
					},
				},
			},
			"bar": {
				Backend:        "bar",
				PassHostHeader: true,
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

func TestRuleType(t *testing.T) {
	ingresses := []*v1beta1.Ingress{
		{
			ObjectMeta: v1.ObjectMeta{
				Annotations: map[string]string{"traefik.frontend.rule.type": "PathPrefixStrip"}, //camel case
			},
			Spec: v1beta1.IngressSpec{
				Rules: []v1beta1.IngressRule{
					{
						Host: "foo1",
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{
									{
										Path: "/bar1",
										Backend: v1beta1.IngressBackend{
											ServiceName: "service1",
											ServicePort: intstr.FromInt(801),
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
			ObjectMeta: v1.ObjectMeta{
				Annotations: map[string]string{"traefik.frontend.rule.type": "path"}, //lower case
			},
			Spec: v1beta1.IngressSpec{
				Rules: []v1beta1.IngressRule{
					{
						Host: "foo1",
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{
									{
										Path: "/bar2",
										Backend: v1beta1.IngressBackend{
											ServiceName: "service1",
											ServicePort: intstr.FromInt(801),
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
			ObjectMeta: v1.ObjectMeta{
				Annotations: map[string]string{"traefik.frontend.rule.type": "PathPrefix"}, //path prefix
			},
			Spec: v1beta1.IngressSpec{
				Rules: []v1beta1.IngressRule{
					{
						Host: "foo2",
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{
									{
										Path: "/bar1",
										Backend: v1beta1.IngressBackend{
											ServiceName: "service1",
											ServicePort: intstr.FromInt(801),
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
			ObjectMeta: v1.ObjectMeta{
				Annotations: map[string]string{"traefik.frontend.rule.type": "PathStrip"}, //path strip
			},
			Spec: v1beta1.IngressSpec{
				Rules: []v1beta1.IngressRule{
					{
						Host: "foo2",
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{
									{
										Path: "/bar2",
										Backend: v1beta1.IngressBackend{
											ServiceName: "service1",
											ServicePort: intstr.FromInt(801),
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
			ObjectMeta: v1.ObjectMeta{
				Annotations: map[string]string{"traefik.frontend.rule.type": "PathXXStrip"}, //wrong rule
			},
			Spec: v1beta1.IngressSpec{
				Rules: []v1beta1.IngressRule{
					{
						Host: "foo1",
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{
									{
										Path: "/bar3",
										Backend: v1beta1.IngressBackend{
											ServiceName: "service1",
											ServicePort: intstr.FromInt(801),
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
	services := []*v1.Service{
		{
			ObjectMeta: v1.ObjectMeta{
				Name: "service1",
				UID:  "1",
			},
			Spec: v1.ServiceSpec{
				ClusterIP: "10.0.0.1",
				Ports: []v1.ServicePort{
					{
						Name: "http",
						Port: 801,
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
	provider := Kubernetes{DisablePassHostHeaders: true}
	actualConfig, err := provider.loadIngresses(client)
	actual := actualConfig.Frontends
	if err != nil {
		t.Fatalf("error %+v", err)
	}

	expected := map[string]*types.Frontend{
		"foo1/bar1": {
			Backend:  "foo1/bar1",
			Priority: len("/bar1"),
			Routes: map[string]types.Route{
				"/bar1": {
					Rule: "PathPrefixStrip:/bar1",
				},
				"foo1": {
					Rule: "Host:foo1",
				},
			},
		},
		"foo1/bar2": {
			Backend:  "foo1/bar2",
			Priority: len("/bar2"),
			Routes: map[string]types.Route{
				"/bar2": {
					Rule: "Path:/bar2",
				},
				"foo1": {
					Rule: "Host:foo1",
				},
			},
		},
		"foo2/bar1": {
			Backend:  "foo2/bar1",
			Priority: len("/bar1"),
			Routes: map[string]types.Route{
				"/bar1": {
					Rule: "PathPrefix:/bar1",
				},
				"foo2": {
					Rule: "Host:foo2",
				},
			},
		},
		"foo2/bar2": {
			Backend:  "foo2/bar2",
			Priority: len("/bar2"),
			Routes: map[string]types.Route{
				"/bar2": {
					Rule: "PathStrip:/bar2",
				},
				"foo2": {
					Rule: "Host:foo2",
				},
			},
		},
		"foo1/bar3": {
			Backend:  "foo1/bar3",
			Priority: len("/bar3"),
			Routes: map[string]types.Route{
				"/bar3": {
					Rule: "PathPrefix:/bar3",
				},
				"foo1": {
					Rule: "Host:foo1",
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

func TestGetPassHostHeader(t *testing.T) {
	ingresses := []*v1beta1.Ingress{{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "awesome",
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{
					Host: "foo",
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Path: "/bar",
									Backend: v1beta1.IngressBackend{
										ServiceName: "service1",
										ServicePort: intstr.FromInt(801),
									},
								},
							},
						},
					},
				},
			},
		},
	}}
	services := []*v1.Service{
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      "service1",
				Namespace: "awesome",
				UID:       "1",
			},
			Spec: v1.ServiceSpec{
				ClusterIP: "10.0.0.1",
				Ports: []v1.ServicePort{
					{
						Name: "http",
						Port: 801,
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
	provider := Kubernetes{DisablePassHostHeaders: true}
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
						Weight: 0,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer:   nil,
			},
		},
		Frontends: map[string]*types.Frontend{
			"foo/bar": {
				Backend:  "foo/bar",
				Priority: len("/bar"),
				Routes: map[string]types.Route{
					"/bar": {
						Rule: "PathPrefix:/bar",
					},
					"foo": {
						Rule: "Host:foo",
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

func TestOnlyReferencesServicesFromOwnNamespace(t *testing.T) {
	ingresses := []*v1beta1.Ingress{
		{
			ObjectMeta: v1.ObjectMeta{
				Namespace: "awesome",
			},
			Spec: v1beta1.IngressSpec{
				Rules: []v1beta1.IngressRule{
					{
						Host: "foo",
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{
									{
										Backend: v1beta1.IngressBackend{
											ServiceName: "service",
											ServicePort: intstr.FromInt(80),
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
	services := []*v1.Service{
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      "service",
				UID:       "1",
				Namespace: "awesome",
			},
			Spec: v1.ServiceSpec{
				ClusterIP: "10.0.0.1",
				Ports: []v1.ServicePort{
					{
						Name: "http",
						Port: 80,
					},
				},
			},
		},
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      "service",
				UID:       "2",
				Namespace: "not-awesome",
			},
			Spec: v1.ServiceSpec{
				ClusterIP: "10.0.0.2",
				Ports: []v1.ServicePort{
					{
						Name: "http",
						Port: 80,
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
			"foo": {
				Servers: map[string]types.Server{
					"1": {
						URL:    "http://10.0.0.1:80",
						Weight: 0,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer:   nil,
			},
		},
		Frontends: map[string]*types.Frontend{
			"foo": {
				Backend:        "foo",
				PassHostHeader: true,
				Routes: map[string]types.Route{
					"foo": {
						Rule: "Host:foo",
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
	ingresses := []*v1beta1.Ingress{
		{
			ObjectMeta: v1.ObjectMeta{
				Namespace: "awesome",
			},
			Spec: v1beta1.IngressSpec{
				Rules: []v1beta1.IngressRule{
					{
						Host: "foo",
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{
									{
										Path: "/bar",
										Backend: v1beta1.IngressBackend{
											ServiceName: "service1",
											ServicePort: intstr.FromInt(801),
										},
									},
								},
							},
						},
					},
					{
						Host: "bar",
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{
									{
										Backend: v1beta1.IngressBackend{
											ServiceName: "service3",
											ServicePort: intstr.FromInt(443),
										},
									},
									{
										Backend: v1beta1.IngressBackend{
											ServiceName: "service2",
											ServicePort: intstr.FromInt(802),
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
			ObjectMeta: v1.ObjectMeta{
				Namespace: "not-awesome",
			},
			Spec: v1beta1.IngressSpec{
				Rules: []v1beta1.IngressRule{
					{
						Host: "baz",
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{
									{
										Path: "/baz",
										Backend: v1beta1.IngressBackend{
											ServiceName: "service1",
											ServicePort: intstr.FromInt(801),
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
	services := []*v1.Service{
		{
			ObjectMeta: v1.ObjectMeta{
				Namespace: "awesome",
				Name:      "service1",
				UID:       "1",
			},
			Spec: v1.ServiceSpec{
				ClusterIP: "10.0.0.1",
				Ports: []v1.ServicePort{
					{
						Name: "http",
						Port: 801,
					},
				},
			},
		},
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      "service1",
				Namespace: "not-awesome",
				UID:       "1",
			},
			Spec: v1.ServiceSpec{
				ClusterIP: "10.0.0.1",
				Ports: []v1.ServicePort{
					{
						Name: "http",
						Port: 801,
					},
				},
			},
		},
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      "service2",
				Namespace: "awesome",
				UID:       "2",
			},
			Spec: v1.ServiceSpec{
				ClusterIP: "10.0.0.2",
				Ports: []v1.ServicePort{
					{
						Port: 802,
					},
				},
			},
		},
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      "service3",
				Namespace: "awesome",
				UID:       "3",
			},
			Spec: v1.ServiceSpec{
				ClusterIP: "10.0.0.3",
				Ports: []v1.ServicePort{
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
						Weight: 0,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer:   nil,
			},
			"bar": {
				Servers: map[string]types.Server{
					"2": {
						URL:    "http://10.0.0.2:802",
						Weight: 0,
					},
					"3": {
						URL:    "https://10.0.0.3:443",
						Weight: 0,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer:   nil,
			},
		},
		Frontends: map[string]*types.Frontend{
			"foo/bar": {
				Backend:        "foo/bar",
				PassHostHeader: true,
				Priority:       len("/bar"),
				Routes: map[string]types.Route{
					"/bar": {
						Rule: "PathPrefix:/bar",
					},
					"foo": {
						Rule: "Host:foo",
					},
				},
			},
			"bar": {
				Backend:        "bar",
				PassHostHeader: true,
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
	ingresses := []*v1beta1.Ingress{
		{
			ObjectMeta: v1.ObjectMeta{
				Namespace: "awesome",
			},
			Spec: v1beta1.IngressSpec{
				Rules: []v1beta1.IngressRule{
					{
						Host: "foo",
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{
									{
										Path: "/bar",
										Backend: v1beta1.IngressBackend{
											ServiceName: "service1",
											ServicePort: intstr.FromInt(801),
										},
									},
								},
							},
						},
					},
					{
						Host: "bar",
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{
									{
										Backend: v1beta1.IngressBackend{
											ServiceName: "service3",
											ServicePort: intstr.FromInt(443),
										},
									},
									{
										Backend: v1beta1.IngressBackend{
											ServiceName: "service2",
											ServicePort: intstr.FromInt(802),
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
			ObjectMeta: v1.ObjectMeta{
				Namespace: "somewhat-awesome",
			},
			Spec: v1beta1.IngressSpec{
				Rules: []v1beta1.IngressRule{
					{
						Host: "awesome",
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{
									{
										Path: "/quix",
										Backend: v1beta1.IngressBackend{
											ServiceName: "service1",
											ServicePort: intstr.FromInt(801),
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
			ObjectMeta: v1.ObjectMeta{
				Namespace: "not-awesome",
			},
			Spec: v1beta1.IngressSpec{
				Rules: []v1beta1.IngressRule{
					{
						Host: "baz",
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{
									{
										Path: "/baz",
										Backend: v1beta1.IngressBackend{
											ServiceName: "service1",
											ServicePort: intstr.FromInt(801),
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
	services := []*v1.Service{
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      "service1",
				Namespace: "awesome",
				UID:       "1",
			},
			Spec: v1.ServiceSpec{
				ClusterIP: "10.0.0.1",
				Ports: []v1.ServicePort{
					{
						Name: "http",
						Port: 801,
					},
				},
			},
		},
		{
			ObjectMeta: v1.ObjectMeta{
				Namespace: "somewhat-awesome",
				Name:      "service1",
				UID:       "17",
			},
			Spec: v1.ServiceSpec{
				ClusterIP: "10.0.0.4",
				Ports: []v1.ServicePort{
					{
						Name: "http",
						Port: 801,
					},
				},
			},
		},
		{
			ObjectMeta: v1.ObjectMeta{
				Namespace: "awesome",
				Name:      "service2",
				UID:       "2",
			},
			Spec: v1.ServiceSpec{
				ClusterIP: "10.0.0.2",
				Ports: []v1.ServicePort{
					{
						Port: 802,
					},
				},
			},
		},
		{
			ObjectMeta: v1.ObjectMeta{
				Namespace: "awesome",
				Name:      "service3",
				UID:       "3",
			},
			Spec: v1.ServiceSpec{
				ClusterIP: "10.0.0.3",
				Ports: []v1.ServicePort{
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
						Weight: 0,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer:   nil,
			},
			"bar": {
				Servers: map[string]types.Server{
					"2": {
						URL:    "http://10.0.0.2:802",
						Weight: 0,
					},
					"3": {
						URL:    "https://10.0.0.3:443",
						Weight: 0,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer:   nil,
			},
			"awesome/quix": {
				Servers: map[string]types.Server{
					"17": {
						URL:    "http://10.0.0.4:801",
						Weight: 0,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer:   nil,
			},
		},
		Frontends: map[string]*types.Frontend{
			"foo/bar": {
				Backend:        "foo/bar",
				PassHostHeader: true,
				Priority:       len("/bar"),
				Routes: map[string]types.Route{
					"/bar": {
						Rule: "PathPrefix:/bar",
					},
					"foo": {
						Rule: "Host:foo",
					},
				},
			},
			"bar": {
				Backend:        "bar",
				PassHostHeader: true,
				Routes: map[string]types.Route{
					"bar": {
						Rule: "Host:bar",
					},
				},
			},
			"awesome/quix": {
				Backend:        "awesome/quix",
				PassHostHeader: true,
				Priority:       len("/quix"),
				Routes: map[string]types.Route{
					"/quix": {
						Rule: "PathPrefix:/quix",
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

func TestHostlessIngress(t *testing.T) {
	ingresses := []*v1beta1.Ingress{{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "awesome",
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Path: "/bar",
									Backend: v1beta1.IngressBackend{
										ServiceName: "service1",
										ServicePort: intstr.FromInt(801),
									},
								},
							},
						},
					},
				},
			},
		},
	}}
	services := []*v1.Service{
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      "service1",
				Namespace: "awesome",
				UID:       "1",
			},
			Spec: v1.ServiceSpec{
				ClusterIP: "10.0.0.1",
				Ports: []v1.ServicePort{
					{
						Name: "http",
						Port: 801,
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
	provider := Kubernetes{DisablePassHostHeaders: true}
	actual, err := provider.loadIngresses(client)
	if err != nil {
		t.Fatalf("error %+v", err)
	}

	expected := &types.Configuration{
		Backends: map[string]*types.Backend{
			"/bar": {
				Servers: map[string]types.Server{
					"1": {
						URL:    "http://10.0.0.1:801",
						Weight: 0,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer:   nil,
			},
		},
		Frontends: map[string]*types.Frontend{
			"/bar": {
				Backend:  "/bar",
				Priority: len("/bar"),
				Routes: map[string]types.Route{
					"/bar": {
						Rule: "PathPrefix:/bar",
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
	ingresses []*v1beta1.Ingress
	services  []*v1.Service
	endpoints []*v1.Endpoints
	watchChan chan interface{}
}

func (c clientMock) GetIngresses(namespaces k8s.Namespaces) []*v1beta1.Ingress {
	result := make([]*v1beta1.Ingress, 0, len(c.ingresses))

	for _, ingress := range c.ingresses {
		if k8s.HasNamespace(ingress, namespaces) {
			result = append(result, ingress)
		}
	}
	return result
}

func (c clientMock) WatchIngresses(labelSelector string, stopCh <-chan struct{}) chan interface{} {
	return c.watchChan
}

func (c clientMock) GetService(namespace, name string) (*v1.Service, bool, error) {
	for _, service := range c.services {
		if service.Namespace == namespace && service.Name == name {
			return service, true, nil
		}
	}
	return &v1.Service{}, true, nil
}

func (c clientMock) GetEndpoints(namespace, name string) (*v1.Endpoints, bool, error) {
	for _, endpoints := range c.endpoints {
		if endpoints.Namespace == namespace && endpoints.Name == name {
			return endpoints, true, nil
		}
	}
	return &v1.Endpoints{}, true, nil
}

func (c clientMock) WatchAll(labelString string, stopCh <-chan bool) (chan interface{}, error) {
	return c.watchChan, nil
}
