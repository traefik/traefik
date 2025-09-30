package knative

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	knativenetworkingv1alpha1 "knative.dev/networking/pkg/apis/networking/v1alpha1"
)

//  func TestBuildKnativeService(t *testing.T) {
//	client := &clientMock{
//		services: []*corev1.Service{
//			{
//				ObjectMeta: metav1.ObjectMeta{
//					Name:      "test-service",
//					Namespace: "default",
//				},
//				Spec: corev1.ServiceSpec{
//					ExternalName: "traefik.default.svc.cluster.local",
//					ClusterIP:    "1.2.3.4",
//					Ports: []corev1.ServicePort{
//						{
//							Name:       "http2",
//							Port:       80,
//							Protocol:   corev1.ProtocolTCP,
//							TargetPort: intstr.FromInt(80),
//						},
//					},
//					SessionAffinity: corev1.ServiceAffinityNone,
//					Type:            corev1.ServiceTypeExternalName,
//				},
//			},
//		},
//	}
//
//	// Create a sample Knative Ingress
//	ingressRoute := &knativenetworkingv1alpha1.Ingress{
//		Spec: knativenetworkingv1alpha1.IngressSpec{
//			Rules: []knativenetworkingv1alpha1.IngressRule{
//				{
//					HTTP: &knativenetworkingv1alpha1.HTTPIngressRuleValue{
//						Paths: []knativenetworkingv1alpha1.HTTPIngressPath{
//							{
//								Path: "/test",
//								Splits: []knativenetworkingv1alpha1.IngressBackendSplit{
//									{
//										IngressBackend: knativenetworkingv1alpha1.IngressBackend{
//											ServiceNamespace: "default",
//											ServiceName:      "test-service",
//											ServicePort:      intstr.FromInt(80),
//										},
//										Percent: 100,
//									},
//								},
//							},
//						},
//					},
//				},
//			},
//		},
//	}
//
//	// Create maps for middleware and services
//	middleware := make(map[string]*dynamic.Middleware)
//	conf := make(map[string]*dynamic.Service)
//
//	provider := &Provider{
//		ExternalEntrypoints: []string{"web"},
//		InternalEntrypoints: []string{"web-internal"},
//		k8sClient:           client,
//	}
//
//	// Call the method
//	results := provider.buildKService(t.Context(), ingressRoute, middleware, conf, "test-service")
//
//	// Assertions
//	require.Len(t, results, 1)
//	result := results[0]
//	assert.Equal(t, "default-test-service-80", result.ServiceKey)
//	assert.Equal(t, "/test", result.Path)
//	assert.NoError(t, result.Err)
//	assert.NotNil(t, conf["default-test-service-80"])
//}

func TestLoadKnativeServers(t *testing.T) {
	mockClient := &clientMock{
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
		serverlessServices: []*knativenetworkingv1alpha1.ServerlessService{
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
		ingresses: []*knativenetworkingv1alpha1.Ingress{
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
		ExternalEntrypoints: []string{"web"},
		InternalEntrypoints: []string{"web-internal"},
		k8sClient:           mockClient,
	}

	t.Run("successful load of servers", func(t *testing.T) {
		servers, err := provider.buildServers("default", "test-service", intstr.FromInt32(80))
		require.NoError(t, err)
		require.Len(t, servers, 1)
		assert.Equal(t, "http://10.0.0.1:80", servers[0].URL)
	})

	t.Run("service not found", func(t *testing.T) {
		mockClient.apiServiceError = errors.New("service not found")

		_, err := provider.buildServers("default", "non-existent-service", intstr.FromInt32(80))
		require.Error(t, err)
		assert.Equal(t, "getting service default/non-existent-service: service not found", err.Error())
	})
}

//  func TestLoadKnativeIngressRouteConfiguration(t *testing.T) {
//	provider := &Provider{
//		ExternalEntrypoints: []string{"web"},
//		InternalEntrypoints: []string{"web-internal"},
//		k8sClient: &clientMock{
//			services: []*corev1.Service{
//				{
//					ObjectMeta: metav1.ObjectMeta{
//						Name:      "test-service",
//						Namespace: "default",
//					},
//					Spec: corev1.ServiceSpec{
//						ClusterIP: "10.0.0.1",
//						Ports: []corev1.ServicePort{
//							{
//								Name: "http",
//								Port: 80,
//							},
//						},
//					},
//				},
//			},
//			serverlessServices: []*knativenetworkingv1alpha1.ServerlessService{
//				{
//					ObjectMeta: metav1.ObjectMeta{
//						Name:      "test-service",
//						Namespace: "default",
//					},
//					Status: knativenetworkingv1alpha1.ServerlessServiceStatus{
//						ServiceName: "test-service",
//					},
//				},
//			},
//			ingresses: []*knativenetworkingv1alpha1.Ingress{
//				{
//					ObjectMeta: metav1.ObjectMeta{
//						Name:      "example-ingress",
//						Namespace: "default",
//						Annotations: map[string]string{
//							"networking.knative.dev/ingress.class": "traefik.ingress.networking.knative.dev",
//						},
//					},
//					Spec: knativenetworkingv1alpha1.IngressSpec{
//						Rules: []knativenetworkingv1alpha1.IngressRule{
//							{
//								Hosts: []string{"example.com"},
//								HTTP: &knativenetworkingv1alpha1.HTTPIngressRuleValue{
//									Paths: []knativenetworkingv1alpha1.HTTPIngressPath{
//										{
//											Path: "/",
//											Splits: []knativenetworkingv1alpha1.IngressBackendSplit{
//												{
//													IngressBackend: knativenetworkingv1alpha1.IngressBackend{
//														ServiceNamespace: "default",
//														ServiceName:      "test-service",
//														ServicePort:      intstr.FromInt(80),
//													},
//													Percent: 100,
//												},
//											},
//										},
//									},
//								},
//							},
//						},
//					},
//				},
//			},
//		},
//	}
//
//	ctx := t.Context()
//	conf, ingressStatusList := provider.loadConfiguration(ctx)
//
//	require.NotNil(t, conf)
//	assert.NotEmpty(t, conf.HTTP.Routers)
//	assert.NotEmpty(t, conf.HTTP.Services)
//
//	router, ok := conf.HTTP.Routers["default-test-service-80"]
//	require.True(t, ok)
//	assert.Equal(t, "web", router.EntryPoints[0])
//	assert.Equal(t, "(Host(`example.com`)) && PathPrefix(`/`)", router.Rule)
//	assert.Equal(t, "default-test-service-80", router.Service)
//
//	require.NotNil(t, ingressStatusList)
//	assert.NotEmpty(t, ingressStatusList)
//	assert.Equal(t, "example-ingress", ingressStatusList[0].Name)
//	assert.Equal(t, "default", ingressStatusList[0].Namespace)
//}

func Test_buildRule(t *testing.T) {
	tests := []struct {
		name    string
		hosts   []string
		headers map[string]knativenetworkingv1alpha1.HeaderMatch
		path    string
		want    string
	}{
		{
			name:  "single host, no headers, no path",
			hosts: []string{"example.com"},
			want:  "(Host(`example.com`))",
		},
		{
			name:  "multiple hosts, no headers, no path",
			hosts: []string{"example.com", "foo.com"},
			want:  "(Host(`example.com`) || Host(`foo.com`))",
		},
		{
			name:  "single host, single header, no path",
			hosts: []string{"example.com"},
			headers: map[string]knativenetworkingv1alpha1.HeaderMatch{
				"X-Header": {Exact: "value"},
			},
			want: "(Host(`example.com`)) && (Header(`X-Header`,`value`))",
		},
		{
			name:  "single host, multiple headers, no path",
			hosts: []string{"example.com"},
			headers: map[string]knativenetworkingv1alpha1.HeaderMatch{
				"X-Header":  {Exact: "value"},
				"X-Header2": {Exact: "value2"},
			},
			want: "(Host(`example.com`)) && (Header(`X-Header`,`value`) && Header(`X-Header2`,`value2`))",
		},
		{
			name:  "single host, multiple headers, with path",
			hosts: []string{"example.com"},
			headers: map[string]knativenetworkingv1alpha1.HeaderMatch{
				"X-Header":  {Exact: "value"},
				"X-Header2": {Exact: "value2"},
			},
			path: "/foo",
			want: "(Host(`example.com`)) && (Header(`X-Header`,`value`) && Header(`X-Header2`,`value2`)) && PathPrefix(`/foo`)",
		},
		{
			name:  "single host, no headers, with path",
			hosts: []string{"example.com"},
			path:  "/foo",
			want:  "(Host(`example.com`)) && PathPrefix(`/foo`)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := buildRule(test.hosts, test.headers, test.path)
			assert.Equal(t, test.want, got)
		})
	}
}
