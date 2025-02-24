package crd

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefikv1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	"github.com/traefik/traefik/v3/pkg/tls"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	knativenetworkingv1alpha1 "knative.dev/networking/pkg/apis/networking/v1alpha1"
)

func TestBuildKnativeService(t *testing.T) {
	// Create a mock client
	client := &MockClient{
		services: []*corev1.Service{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
				Spec: corev1.ServiceSpec{
					ExternalName: "traefik.default.svc.cluster.local",
					Ports: []corev1.ServicePort{
						{
							Name:       "http2",
							Port:       80,
							Protocol:   corev1.ProtocolTCP,
							TargetPort: intstr.FromInt(80),
						},
					},
					SessionAffinity: corev1.ServiceAffinityNone,
					Type:            corev1.ServiceTypeExternalName,
				},
			},
		},
	}

	// Create a sample Knative Ingress
	ingressRoute := &knativenetworkingv1alpha1.Ingress{
		Spec: knativenetworkingv1alpha1.IngressSpec{
			Rules: []knativenetworkingv1alpha1.IngressRule{
				{
					HTTP: &knativenetworkingv1alpha1.HTTPIngressRuleValue{
						Paths: []knativenetworkingv1alpha1.HTTPIngressPath{
							{
								Path: "/test",
								Splits: []knativenetworkingv1alpha1.IngressBackendSplit{
									{
										IngressBackend: knativenetworkingv1alpha1.IngressBackend{
											ServiceNamespace: "default",
											ServiceName:      "test-service",
											ServicePort:      intstr.FromInt(80),
										},
										Percent: 100,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Create a configBuilder instance
	cb := configBuilder{client: client, allowCrossNamespace: false}

	// Create maps for middleware and services
	middleware := make(map[string]*dynamic.Middleware)
	conf := make(map[string]*dynamic.Service)

	// Call the method
	results := cb.buildKnativeService(context.Background(), ingressRoute, middleware, conf, "test-service")

	// Assertions
	require.Len(t, results, 1)
	result := results[0]
	assert.Equal(t, "default-test-service-80", result.ServiceKey)
	assert.Equal(t, "/test", result.Path)
	assert.NoError(t, result.Err)
	assert.NotNil(t, conf["default-test-service-80"])
}

func TestLoadKnativeServers(t *testing.T) {
	mockClient := &MockClient{
		services: []*corev1.Service{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
				Spec: corev1.ServiceSpec{
					ClusterIP: "10.0.0.1",
					Ports: []corev1.ServicePort{
						{
							Name: "http",
							Port: 80,
						},
					},
				},
			},
		},
		serverlessService: []*knativenetworkingv1alpha1.ServerlessService{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
				Status: knativenetworkingv1alpha1.ServerlessServiceStatus{
					ServiceName: "test-service",
				},
			},
		},
	}

	cb := configBuilder{client: mockClient}

	t.Run("successful load of servers", func(t *testing.T) {
		svc := traefikv1alpha1.Service{
			LoadBalancerSpec: traefikv1alpha1.LoadBalancerSpec{
				Name:      "test-service",
				Namespace: "default",
				Port:      intstr.FromInt(80),
			},
		}

		servers, err := cb.loadKnativeServers("default", svc)
		require.NoError(t, err)
		require.Len(t, servers, 1)
		assert.Equal(t, "http://10.0.0.1:80", servers[0].URL)
	})

	t.Run("service not found", func(t *testing.T) {
		mockClient.apiServiceError = errors.New("service not found")
		svc := traefikv1alpha1.Service{
			LoadBalancerSpec: traefikv1alpha1.LoadBalancerSpec{
				Name: "non-existent-service",
				Port: intstr.FromInt(80),
			},
		}

		_, err := cb.loadKnativeServers("default", svc)
		require.Error(t, err)
		assert.Equal(t, "service not found", err.Error())
	})
}

func TestLoadKnativeIngressRouteConfiguration(t *testing.T) {
	mockClient := &MockClient{
		services: []*corev1.Service{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
				Spec: corev1.ServiceSpec{
					ClusterIP: "10.0.0.1",
					Ports: []corev1.ServicePort{
						{
							Name: "http",
							Port: 80,
						},
					},
				},
			},
		},
		serverlessService: []*knativenetworkingv1alpha1.ServerlessService{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
				Status: knativenetworkingv1alpha1.ServerlessServiceStatus{
					ServiceName: "test-service",
				},
			},
		},
		ingressRoute: []*knativenetworkingv1alpha1.Ingress{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "example-ingress",
					Namespace: "default",
					Annotations: map[string]string{
						"networking.knative.dev/ingress.class": "traefik.ingress.networking.knative.dev",
					},
				},
				Spec: knativenetworkingv1alpha1.IngressSpec{
					Rules: []knativenetworkingv1alpha1.IngressRule{
						{
							Hosts: []string{"example.com"},
							HTTP: &knativenetworkingv1alpha1.HTTPIngressRuleValue{
								Paths: []knativenetworkingv1alpha1.HTTPIngressPath{
									{
										Path: "/",
										Splits: []knativenetworkingv1alpha1.IngressBackendSplit{
											{
												IngressBackend: knativenetworkingv1alpha1.IngressBackend{
													ServiceNamespace: "default",
													ServiceName:      "test-service",
													ServicePort:      intstr.FromInt(80),
												},
												Percent: 100,
											},
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

	provider := &Provider{
		Entrypoints:         []string{"web"},
		EntrypointsInternal: []string{"web-internal"},
		AllowCrossNamespace: true,
	}

	tlsConfigs := make(map[string]*tls.CertAndStores)

	ctx := context.Background()
	conf := provider.loadKnativeIngressRouteConfiguration(ctx, mockClient, tlsConfigs)

	require.NotNil(t, conf)
	assert.NotEmpty(t, conf.Routers)
	assert.NotEmpty(t, conf.Services)

	router, ok := conf.Routers["default-test-service-80"]
	require.True(t, ok)
	assert.Equal(t, "web", router.EntryPoints[0])
	assert.Equal(t, "(Host(`example.com`)) && PathPrefix(`/`)", router.Rule)
	assert.Equal(t, "default-test-service-80", router.Service)
}
