package kubernetes

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/util/intstr"
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
								{
									Path: "/namedthing",
									Backend: v1beta1.IngressBackend{
										ServiceName: "service4",
										ServicePort: intstr.FromString("https"),
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
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      "service4",
				UID:       "4",
				Namespace: "testing",
			},
			Spec: v1.ServiceSpec{
				ClusterIP:    "10.0.0.4",
				Type:         "ExternalName",
				ExternalName: "example.com",
				Ports: []v1.ServicePort{
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
	provider := Provider{}
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
						Weight: 1,
					},
					"http://10.21.0.1:8080": {
						URL:    "http://10.21.0.1:8080",
						Weight: 1,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer: &types.LoadBalancer{
					Sticky: false,
					Method: "wrr",
				},
			},
			"foo/namedthing": {
				Servers: map[string]types.Server{
					"https://example.com": {
						URL:    "https://example.com",
						Weight: 1,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer: &types.LoadBalancer{
					Sticky: false,
					Method: "wrr",
				},
			},
			"bar": {
				Servers: map[string]types.Server{
					"https://10.15.0.1:8443": {
						URL:    "https://10.15.0.1:8443",
						Weight: 1,
					},
					"https://10.15.0.2:9443": {
						URL:    "https://10.15.0.2:9443",
						Weight: 1,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer: &types.LoadBalancer{
					Sticky: false,
					Method: "wrr",
				},
			},
		},
		Frontends: map[string]*types.Frontend{
			"foo/bar": {
				Backend:        "foo/bar",
				PassHostHeader: true,
				Routes: map[string]types.Route{
					"/bar": {
						Rule: "PathPrefix:/bar",
					},
					"foo": {
						Rule: "Host:foo",
					},
				},
			},
			"foo/namedthing": {
				Backend:        "foo/namedthing",
				PassHostHeader: true,
				Routes: map[string]types.Route{
					"/namedthing": {
						Rule: "PathPrefix:/namedthing",
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

	assert.Equal(t, expected, actual)
}

func TestRuleType(t *testing.T) {
	tests := []struct {
		desc             string
		ingressRuleType  string
		frontendRuleType string
	}{
		{
			desc:             "rule type annotation missing",
			ingressRuleType:  "",
			frontendRuleType: ruleTypePathPrefix,
		},
		{
			desc:             "Path rule type annotation set",
			ingressRuleType:  "Path",
			frontendRuleType: "Path",
		},
		{
			desc:             "PathStrip rule type annotation set",
			ingressRuleType:  "PathStrip",
			frontendRuleType: "PathStrip",
		},
		{
			desc:             "PathStripPrefix rule type annotation set",
			ingressRuleType:  "PathStripPrefix",
			frontendRuleType: "PathStripPrefix",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			ingress := &v1beta1.Ingress{
				Spec: v1beta1.IngressSpec{
					Rules: []v1beta1.IngressRule{
						{
							Host: "host",
							IngressRuleValue: v1beta1.IngressRuleValue{
								HTTP: &v1beta1.HTTPIngressRuleValue{
									Paths: []v1beta1.HTTPIngressPath{
										{
											Path: "/path",
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
			}

			if test.ingressRuleType != "" {
				ingress.ObjectMeta.Annotations = map[string]string{
					types.LabelFrontendRuleType: test.ingressRuleType,
				}
			}

			service := &v1.Service{
				ObjectMeta: v1.ObjectMeta{
					Name: "service",
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
			}

			watchChan := make(chan interface{})
			client := clientMock{
				ingresses: []*v1beta1.Ingress{ingress},
				services:  []*v1.Service{service},
				watchChan: watchChan,
			}
			provider := Provider{DisablePassHostHeaders: true}
			actualConfig, err := provider.loadIngresses(client)
			if err != nil {
				t.Fatalf("error loading ingresses: %+v", err)
			}

			actual := actualConfig.Frontends
			expected := map[string]*types.Frontend{
				"host/path": {
					Backend: "host/path",
					Routes: map[string]types.Route{
						"/path": {
							Rule: fmt.Sprintf("%s:/path", test.frontendRuleType),
						},
						"host": {
							Rule: "Host:host",
						},
					},
				},
			}

			assert.Equal(t, expected, actual)
		})
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
	provider := Provider{DisablePassHostHeaders: true}
	actual, err := provider.loadIngresses(client)
	if err != nil {
		t.Fatalf("error %+v", err)
	}

	expected := &types.Configuration{
		Backends: map[string]*types.Backend{
			"foo/bar": {
				Servers:        map[string]types.Server{},
				CircuitBreaker: nil,
				LoadBalancer: &types.LoadBalancer{
					Sticky: false,
					Method: "wrr",
				},
			},
		},
		Frontends: map[string]*types.Frontend{
			"foo/bar": {
				Backend: "foo/bar",
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

	assert.Equal(t, expected, actual)
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
	provider := Provider{}
	actual, err := provider.loadIngresses(client)
	if err != nil {
		t.Fatalf("error %+v", err)
	}

	expected := &types.Configuration{
		Backends: map[string]*types.Backend{
			"foo": {
				Servers:        map[string]types.Server{},
				CircuitBreaker: nil,
				LoadBalancer: &types.LoadBalancer{
					Sticky: false,
					Method: "wrr",
				},
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

	assert.Equal(t, expected, actual)
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
	provider := Provider{
		Namespaces: []string{"awesome"},
	}
	actual, err := provider.loadIngresses(client)
	if err != nil {
		t.Fatalf("error %+v", err)
	}

	expected := &types.Configuration{
		Backends: map[string]*types.Backend{
			"foo/bar": {
				Servers:        map[string]types.Server{},
				CircuitBreaker: nil,
				LoadBalancer: &types.LoadBalancer{
					Sticky: false,
					Method: "wrr",
				},
			},
			"bar": {
				Servers:        map[string]types.Server{},
				CircuitBreaker: nil,
				LoadBalancer: &types.LoadBalancer{
					Sticky: false,
					Method: "wrr",
				},
			},
		},
		Frontends: map[string]*types.Frontend{
			"foo/bar": {
				Backend:        "foo/bar",
				PassHostHeader: true,
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

	assert.Equal(t, expected, actual)
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
	provider := Provider{
		Namespaces: []string{"awesome", "somewhat-awesome"},
	}
	actual, err := provider.loadIngresses(client)
	if err != nil {
		t.Fatalf("error %+v", err)
	}

	expected := &types.Configuration{
		Backends: map[string]*types.Backend{
			"foo/bar": {
				Servers:        map[string]types.Server{},
				CircuitBreaker: nil,
				LoadBalancer: &types.LoadBalancer{
					Sticky: false,
					Method: "wrr",
				},
			},
			"bar": {
				Servers:        map[string]types.Server{},
				CircuitBreaker: nil,
				LoadBalancer: &types.LoadBalancer{
					Sticky: false,
					Method: "wrr",
				},
			},
			"awesome/quix": {
				Servers:        map[string]types.Server{},
				CircuitBreaker: nil,
				LoadBalancer: &types.LoadBalancer{
					Sticky: false,
					Method: "wrr",
				},
			},
		},
		Frontends: map[string]*types.Frontend{
			"foo/bar": {
				Backend:        "foo/bar",
				PassHostHeader: true,
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

	assert.Equal(t, expected, actual)
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
	provider := Provider{DisablePassHostHeaders: true}
	actual, err := provider.loadIngresses(client)
	if err != nil {
		t.Fatalf("error %+v", err)
	}

	expected := &types.Configuration{
		Backends: map[string]*types.Backend{
			"/bar": {
				Servers:        map[string]types.Server{},
				CircuitBreaker: nil,
				LoadBalancer: &types.LoadBalancer{
					Sticky: false,
					Method: "wrr",
				},
			},
		},
		Frontends: map[string]*types.Frontend{
			"/bar": {
				Backend: "/bar",
				Routes: map[string]types.Route{
					"/bar": {
						Rule: "PathPrefix:/bar",
					},
				},
			},
		},
	}

	assert.Equal(t, expected, actual)
}

func TestServiceAnnotations(t *testing.T) {
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
				Annotations: map[string]string{
					types.LabelTraefikBackendCircuitbreaker: "NetworkErrorRatio() > 0.5",
					types.LabelBackendLoadbalancerMethod:    "drr",
				},
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
				Annotations: map[string]string{
					types.LabelTraefikBackendCircuitbreaker: "",
					types.LabelBackendLoadbalancerSticky:    "true",
				},
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
				Name:      "service2",
				UID:       "2",
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
							Port: 8080,
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
	provider := Provider{}
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
						Weight: 1,
					},
					"http://10.21.0.1:8080": {
						URL:    "http://10.21.0.1:8080",
						Weight: 1,
					},
				},
				CircuitBreaker: &types.CircuitBreaker{
					Expression: "NetworkErrorRatio() > 0.5",
				},
				LoadBalancer: &types.LoadBalancer{
					Method: "drr",
					Sticky: false,
				},
			},
			"bar": {
				Servers: map[string]types.Server{
					"http://10.15.0.1:8080": {
						URL:    "http://10.15.0.1:8080",
						Weight: 1,
					},
					"http://10.15.0.2:8080": {
						URL:    "http://10.15.0.2:8080",
						Weight: 1,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer: &types.LoadBalancer{
					Method: "wrr",
					Sticky: true,
				},
			},
		},
		Frontends: map[string]*types.Frontend{
			"foo/bar": {
				Backend:        "foo/bar",
				PassHostHeader: true,
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

	assert.Equal(t, expected, actual)
}

func TestIngressAnnotations(t *testing.T) {
	ingresses := []*v1beta1.Ingress{
		{
			ObjectMeta: v1.ObjectMeta{
				Namespace: "testing",
				Annotations: map[string]string{
					types.LabelFrontendPassHostHeader: "false",
				},
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
				},
			},
		},
		{
			ObjectMeta: v1.ObjectMeta{
				Namespace: "testing",
				Annotations: map[string]string{
					"kubernetes.io/ingress.class":     "traefik",
					types.LabelFrontendPassHostHeader: "true",
				},
			},
			Spec: v1beta1.IngressSpec{
				Rules: []v1beta1.IngressRule{
					{
						Host: "other",
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{
									{
										Path: "/stuff",
										Backend: v1beta1.IngressBackend{
											ServiceName: "service1",
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
		{
			ObjectMeta: v1.ObjectMeta{
				Namespace: "testing",
				Annotations: map[string]string{
					"ingress.kubernetes.io/auth-type":   "basic",
					"ingress.kubernetes.io/auth-secret": "mySecret",
				},
			},
			Spec: v1beta1.IngressSpec{
				Rules: []v1beta1.IngressRule{
					{
						Host: "basic",
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{
									{
										Path: "/auth",
										Backend: v1beta1.IngressBackend{
											ServiceName: "service1",
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
		{
			ObjectMeta: v1.ObjectMeta{
				Namespace: "testing",
				Annotations: map[string]string{
					"kubernetes.io/ingress.class": "somethingOtherThanTraefik",
				},
			},
			Spec: v1beta1.IngressSpec{
				Rules: []v1beta1.IngressRule{
					{
						Host: "herp",
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{
									{
										Path: "/derp",
										Backend: v1beta1.IngressBackend{
											ServiceName: "service2",
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
		{
			ObjectMeta: v1.ObjectMeta{
				Namespace: "testing",
				Annotations: map[string]string{
					"kubernetes.io/ingress.class":                  "traefik",
					"ingress.kubernetes.io/whitelist-source-range": "1.1.1.1/24, 1234:abcd::42/32",
				},
			},
			Spec: v1beta1.IngressSpec{
				Rules: []v1beta1.IngressRule{
					{
						Host: "test",
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{
									{
										Path: "/whitelist-source-range",
										Backend: v1beta1.IngressBackend{
											ServiceName: "service1",
											ServicePort: intstr.FromInt(80),
										},
									},
								},
							},
						},
					},
				},
			},
		}, {
			ObjectMeta: v1.ObjectMeta{
				Namespace: "testing",
				Annotations: map[string]string{
					"ingress.kubernetes.io/rewrite-target": "/",
				},
			},
			Spec: v1beta1.IngressSpec{
				Rules: []v1beta1.IngressRule{
					{
						Host: "rewrite",
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{
									{
										Path: "/api",
										Backend: v1beta1.IngressBackend{
											ServiceName: "service1",
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
				Name:      "service1",
				UID:       "1",
				Namespace: "testing",
			},
			Spec: v1.ServiceSpec{
				ClusterIP:    "10.0.0.1",
				Type:         "ExternalName",
				ExternalName: "example.com",
				Ports: []v1.ServicePort{
					{
						Name: "http",
						Port: 80,
					},
				},
			},
		},
	}
	secrets := []*v1.Secret{
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      "mySecret",
				UID:       "1",
				Namespace: "testing",
			},
			Data: map[string][]byte{
				"auth": []byte("myUser:myEncodedPW"),
			},
		},
	}

	endpoints := []*v1.Endpoints{}
	watchChan := make(chan interface{})
	client := clientMock{
		ingresses: ingresses,
		services:  services,
		secrets:   secrets,
		endpoints: endpoints,
		watchChan: watchChan,
	}
	provider := Provider{}
	actual, err := provider.loadIngresses(client)
	if err != nil {
		t.Fatalf("error %+v", err)
	}

	expected := &types.Configuration{
		Backends: map[string]*types.Backend{
			"foo/bar": {
				Servers: map[string]types.Server{
					"http://example.com": {
						URL:    "http://example.com",
						Weight: 1,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer: &types.LoadBalancer{
					Sticky: false,
					Method: "wrr",
				},
			},
			"other/stuff": {
				Servers: map[string]types.Server{
					"http://example.com": {
						URL:    "http://example.com",
						Weight: 1,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer: &types.LoadBalancer{
					Sticky: false,
					Method: "wrr",
				},
			},
			"basic/auth": {
				Servers: map[string]types.Server{
					"http://example.com": {
						URL:    "http://example.com",
						Weight: 1,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer: &types.LoadBalancer{
					Sticky: false,
					Method: "wrr",
				},
			},
			"test/whitelist-source-range": {
				Servers: map[string]types.Server{
					"http://example.com": {
						URL:    "http://example.com",
						Weight: 1,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer: &types.LoadBalancer{
					Sticky: false,
					Method: "wrr",
				},
			},
			"rewrite/api": {
				Servers: map[string]types.Server{
					"http://example.com": {
						URL:    "http://example.com",
						Weight: 1,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer: &types.LoadBalancer{
					Sticky: false,
					Method: "wrr",
				},
			},
		},
		Frontends: map[string]*types.Frontend{
			"foo/bar": {
				Backend:        "foo/bar",
				PassHostHeader: false,
				Routes: map[string]types.Route{
					"/bar": {
						Rule: "PathPrefix:/bar",
					},
					"foo": {
						Rule: "Host:foo",
					},
				},
			},
			"other/stuff": {
				Backend:        "other/stuff",
				PassHostHeader: true,
				Routes: map[string]types.Route{
					"/stuff": {
						Rule: "PathPrefix:/stuff",
					},
					"other": {
						Rule: "Host:other",
					},
				},
			},
			"basic/auth": {
				Backend:        "basic/auth",
				PassHostHeader: true,
				Routes: map[string]types.Route{
					"/auth": {
						Rule: "PathPrefix:/auth",
					},
					"basic": {
						Rule: "Host:basic",
					},
				},
				BasicAuth: []string{"myUser:myEncodedPW"},
			},
			"test/whitelist-source-range": {
				Backend:        "test/whitelist-source-range",
				PassHostHeader: true,
				WhitelistSourceRange: []string{
					"1.1.1.1/24",
					"1234:abcd::42/32",
				},
				Routes: map[string]types.Route{
					"/whitelist-source-range": {
						Rule: "PathPrefix:/whitelist-source-range",
					},
					"test": {
						Rule: "Host:test",
					},
				},
			},
			"rewrite/api": {
				Backend:        "rewrite/api",
				PassHostHeader: true,
				Routes: map[string]types.Route{
					"/api": {
						Rule: "ReplacePath:/",
					},
					"rewrite": {
						Rule: "Host:rewrite",
					},
				},
			},
		},
	}

	assert.Equal(t, expected, actual)
}

func TestPriorityHeaderValue(t *testing.T) {
	ingresses := []*v1beta1.Ingress{
		{
			ObjectMeta: v1.ObjectMeta{
				Namespace: "testing",
				Annotations: map[string]string{
					types.LabelFrontendPriority: "1337",
				},
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
				},
			},
		},
	}
	services := []*v1.Service{
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      "service1",
				UID:       "1",
				Namespace: "testing",
			},
			Spec: v1.ServiceSpec{
				ClusterIP:    "10.0.0.1",
				Type:         "ExternalName",
				ExternalName: "example.com",
				Ports: []v1.ServicePort{
					{
						Name: "http",
						Port: 80,
					},
				},
			},
		},
	}

	endpoints := []*v1.Endpoints{}
	watchChan := make(chan interface{})
	client := clientMock{
		ingresses: ingresses,
		services:  services,
		endpoints: endpoints,
		watchChan: watchChan,
	}
	provider := Provider{}
	actual, err := provider.loadIngresses(client)
	if err != nil {
		t.Fatalf("error %+v", err)
	}

	expected := &types.Configuration{
		Backends: map[string]*types.Backend{
			"foo/bar": {
				Servers: map[string]types.Server{
					"http://example.com": {
						URL:    "http://example.com",
						Weight: 1,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer: &types.LoadBalancer{
					Sticky: false,
					Method: "wrr",
				},
			},
		},
		Frontends: map[string]*types.Frontend{
			"foo/bar": {
				Backend:        "foo/bar",
				PassHostHeader: true,
				Priority:       1337,
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

	assert.Equal(t, expected, actual)
}

func TestInvalidPassHostHeaderValue(t *testing.T) {
	ingresses := []*v1beta1.Ingress{
		{
			ObjectMeta: v1.ObjectMeta{
				Namespace: "testing",
				Annotations: map[string]string{
					types.LabelFrontendPassHostHeader: "herpderp",
				},
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
				},
			},
		},
	}
	services := []*v1.Service{
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      "service1",
				UID:       "1",
				Namespace: "testing",
			},
			Spec: v1.ServiceSpec{
				ClusterIP:    "10.0.0.1",
				Type:         "ExternalName",
				ExternalName: "example.com",
				Ports: []v1.ServicePort{
					{
						Name: "http",
						Port: 80,
					},
				},
			},
		},
	}

	endpoints := []*v1.Endpoints{}
	watchChan := make(chan interface{})
	client := clientMock{
		ingresses: ingresses,
		services:  services,
		endpoints: endpoints,
		watchChan: watchChan,
	}
	provider := Provider{}
	actual, err := provider.loadIngresses(client)
	if err != nil {
		t.Fatalf("error %+v", err)
	}

	expected := &types.Configuration{
		Backends: map[string]*types.Backend{
			"foo/bar": {
				Servers: map[string]types.Server{
					"http://example.com": {
						URL:    "http://example.com",
						Weight: 1,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer: &types.LoadBalancer{
					Sticky: false,
					Method: "wrr",
				},
			},
		},
		Frontends: map[string]*types.Frontend{
			"foo/bar": {
				Backend:        "foo/bar",
				PassHostHeader: true,
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

	assert.Equal(t, expected, actual)
}

func TestKubeAPIErrors(t *testing.T) {
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
			},
		},
	}}

	services := []*v1.Service{{
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
	}}

	endpoints := []*v1.Endpoints{}
	watchChan := make(chan interface{})
	apiErr := errors.New("failed kube api call")

	testCases := []struct {
		desc            string
		apiServiceErr   error
		apiEndpointsErr error
	}{
		{
			desc:          "failed service call",
			apiServiceErr: apiErr,
		},
		{
			desc:            "failed endpoints call",
			apiEndpointsErr: apiErr,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			client := clientMock{
				ingresses:         ingresses,
				services:          services,
				endpoints:         endpoints,
				watchChan:         watchChan,
				apiServiceError:   tc.apiServiceErr,
				apiEndpointsError: tc.apiEndpointsErr,
			}

			provider := Provider{}
			if _, err := provider.loadIngresses(client); err != apiErr {
				t.Errorf("Got error %v, wanted error %v", err, apiErr)
			}
		})
	}
}

func TestMissingResources(t *testing.T) {
	ingresses := []*v1beta1.Ingress{{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "testing",
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{
					Host: "fully_working",
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Backend: v1beta1.IngressBackend{
										ServiceName: "fully_working_service",
										ServicePort: intstr.FromInt(80),
									},
								},
							},
						},
					},
				},
				{
					Host: "missing_service",
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Backend: v1beta1.IngressBackend{
										ServiceName: "missing_service_service",
										ServicePort: intstr.FromInt(80),
									},
								},
							},
						},
					},
				},
				{
					Host: "missing_endpoints",
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Backend: v1beta1.IngressBackend{
										ServiceName: "missing_endpoints_service",
										ServicePort: intstr.FromInt(80),
									},
								},
							},
						},
					},
				},
				{
					Host: "missing_endpoint_subsets",
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Backend: v1beta1.IngressBackend{
										ServiceName: "missing_endpoint_subsets_service",
										ServicePort: intstr.FromInt(80),
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
				Name:      "fully_working_service",
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
				Name:      "missing_endpoints_service",
				UID:       "3",
				Namespace: "testing",
			},
			Spec: v1.ServiceSpec{
				ClusterIP: "10.0.0.3",
				Ports: []v1.ServicePort{
					{
						Port: 80,
					},
				},
			},
		},
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      "missing_endpoint_subsets_service",
				UID:       "4",
				Namespace: "testing",
			},
			Spec: v1.ServiceSpec{
				ClusterIP: "10.0.0.4",
				Ports: []v1.ServicePort{
					{
						Port: 80,
					},
				},
			},
		},
	}
	endpoints := []*v1.Endpoints{
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      "fully_working_service",
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
			},
		},
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      "missing_endpoint_subsets_service",
				UID:       "4",
				Namespace: "testing",
			},
			Subsets: []v1.EndpointSubset{},
		},
	}

	watchChan := make(chan interface{})
	client := clientMock{
		ingresses: ingresses,
		services:  services,
		endpoints: endpoints,
		watchChan: watchChan,
	}

	provider := Provider{}
	actual, err := provider.loadIngresses(client)
	if err != nil {
		t.Fatalf("error %+v", err)
	}

	expected := &types.Configuration{
		Backends: map[string]*types.Backend{
			"fully_working": {
				Servers: map[string]types.Server{
					"http://10.10.0.1:8080": {
						URL:    "http://10.10.0.1:8080",
						Weight: 1,
					},
				},
				CircuitBreaker: nil,
				LoadBalancer: &types.LoadBalancer{
					Method: "wrr",
					Sticky: false,
				},
			},
			"missing_service": {
				Servers: map[string]types.Server{},
				LoadBalancer: &types.LoadBalancer{
					Method: "wrr",
					Sticky: false,
				},
			},
			"missing_endpoints": {
				Servers:        map[string]types.Server{},
				CircuitBreaker: nil,
				LoadBalancer: &types.LoadBalancer{
					Method: "wrr",
					Sticky: false,
				},
			},
			"missing_endpoint_subsets": {
				Servers:        map[string]types.Server{},
				CircuitBreaker: nil,
				LoadBalancer: &types.LoadBalancer{
					Method: "wrr",
					Sticky: false,
				},
			},
		},
		Frontends: map[string]*types.Frontend{
			"fully_working": {
				Backend:        "fully_working",
				PassHostHeader: true,
				Routes: map[string]types.Route{
					"fully_working": {
						Rule: "Host:fully_working",
					},
				},
			},
			"missing_endpoints": {
				Backend:        "missing_endpoints",
				PassHostHeader: true,
				Routes: map[string]types.Route{
					"missing_endpoints": {
						Rule: "Host:missing_endpoints",
					},
				},
			},
			"missing_endpoint_subsets": {
				Backend:        "missing_endpoint_subsets",
				PassHostHeader: true,
				Routes: map[string]types.Route{
					"missing_endpoint_subsets": {
						Rule: "Host:missing_endpoint_subsets",
					},
				},
			},
		},
	}

	assert.Equal(t, expected, actual)
}

func TestBasicAuthInTemplate(t *testing.T) {
	ingresses := []*v1beta1.Ingress{
		{
			ObjectMeta: v1.ObjectMeta{
				Namespace: "testing",
				Annotations: map[string]string{
					"ingress.kubernetes.io/auth-type":   "basic",
					"ingress.kubernetes.io/auth-secret": "mySecret",
				},
			},
			Spec: v1beta1.IngressSpec{
				Rules: []v1beta1.IngressRule{
					{
						Host: "basic",
						IngressRuleValue: v1beta1.IngressRuleValue{
							HTTP: &v1beta1.HTTPIngressRuleValue{
								Paths: []v1beta1.HTTPIngressPath{
									{
										Path: "/auth",
										Backend: v1beta1.IngressBackend{
											ServiceName: "service1",
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
				Name:      "service1",
				UID:       "1",
				Namespace: "testing",
			},
			Spec: v1.ServiceSpec{
				ClusterIP:    "10.0.0.1",
				Type:         "ExternalName",
				ExternalName: "example.com",
				Ports: []v1.ServicePort{
					{
						Name: "http",
						Port: 80,
					},
				},
			},
		},
	}
	secrets := []*v1.Secret{
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      "mySecret",
				UID:       "1",
				Namespace: "testing",
			},
			Data: map[string][]byte{
				"auth": []byte("myUser:myEncodedPW"),
			},
		},
	}

	endpoints := []*v1.Endpoints{}
	watchChan := make(chan interface{})
	client := clientMock{
		ingresses: ingresses,
		services:  services,
		secrets:   secrets,
		endpoints: endpoints,
		watchChan: watchChan,
	}
	provider := Provider{}
	actual, err := provider.loadIngresses(client)
	if err != nil {
		t.Fatalf("error %+v", err)
	}

	actual = provider.loadConfig(*actual)
	got := actual.Frontends["basic/auth"].BasicAuth
	if !reflect.DeepEqual(got, []string{"myUser:myEncodedPW"}) {
		t.Fatalf("unexpected credentials: %+v", got)
	}
}

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

func (c clientMock) GetIngresses(namespaces Namespaces) []*v1beta1.Ingress {
	result := make([]*v1beta1.Ingress, 0, len(c.ingresses))

	for _, ingress := range c.ingresses {
		if HasNamespace(ingress, namespaces) {
			result = append(result, ingress)
		}
	}
	return result
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

func (c clientMock) WatchAll(labelString string, stopCh <-chan struct{}) (<-chan interface{}, error) {
	return c.watchChan, nil
}
