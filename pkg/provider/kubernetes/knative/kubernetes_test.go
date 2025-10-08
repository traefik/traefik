package knative

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/k8s"
	"k8s.io/apimachinery/pkg/runtime"
	kubefake "k8s.io/client-go/kubernetes/fake"
	kscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	knativenetworkingv1alpha1 "knative.dev/networking/pkg/apis/networking/v1alpha1"
	knfake "knative.dev/networking/pkg/client/clientset/versioned/fake"
)

func init() {
	// required by k8s.MustParseYaml
	if err := knativenetworkingv1alpha1.AddToScheme(kscheme.Scheme); err != nil {
		panic(err)
	}
}

func Test_loadConfiguration(t *testing.T) {
	testCases := []struct {
		desc    string
		paths   []string
		want    *dynamic.Configuration
		wantLen int
	}{
		{
			desc:    "Wrong ingress class",
			paths:   []string{"wrong_ingress_class.yaml"},
			wantLen: 0,
			want: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Services:    map[string]*dynamic.Service{},
					Middlewares: map[string]*dynamic.Middleware{},
				},
			},
		},
		{
			desc:    "Cluster Local",
			paths:   []string{"cluster_local.yaml", "services.yaml"},
			wantLen: 1,
			want: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-helloworld-go-rule-0-path-0": {
							EntryPoints: []string{"priv-http", "priv-https"},
							Service:     "default-helloworld-go-rule-0-path-0-wrr",
							Rule:        "(Host(`helloworld-go.default`) || Host(`helloworld-go.default.svc`) || Host(`helloworld-go.default.svc.cluster.local`))",
							Middlewares: []string{},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-helloworld-go-rule-0-path-0-split-0": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: types.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.43.38.208:80",
									},
								},
							},
						},
						"default-helloworld-go-rule-0-path-0-split-1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: types.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.43.44.18:80",
									},
								},
							},
						},
						"default-helloworld-go-rule-0-path-0-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-helloworld-go-rule-0-path-0-split-0",
										Weight: ptr.To(50),
										Headers: map[string]string{
											"Knative-Serving-Namespace": "default",
											"Knative-Serving-Revision":  "helloworld-go-00001",
										},
									},
									{
										Name:   "default-helloworld-go-rule-0-path-0-split-1",
										Weight: ptr.To(50),
										Headers: map[string]string{
											"Knative-Serving-Namespace": "default",
											"Knative-Serving-Revision":  "helloworld-go-00002",
										},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
				},
			},
		},
		{
			desc:    "External IP",
			paths:   []string{"external_ip.yaml", "services.yaml"},
			wantLen: 1,
			want: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-helloworld-go-rule-0-path-0": {
							EntryPoints: []string{"http", "https"},
							Service:     "default-helloworld-go-rule-0-path-0-wrr",
							Rule:        "(Host(`helloworld-go.default`) || Host(`helloworld-go.default.svc`) || Host(`helloworld-go.default.svc.cluster.local`))",
							Middlewares: []string{},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-helloworld-go-rule-0-path-0-split-0": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: types.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.43.38.208:80",
									},
								},
							},
						},
						"default-helloworld-go-rule-0-path-0-split-1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: types.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.43.44.18:80",
									},
								},
							},
						},
						"default-helloworld-go-rule-0-path-0-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-helloworld-go-rule-0-path-0-split-0",
										Weight: ptr.To(50),
										Headers: map[string]string{
											"Knative-Serving-Namespace": "default",
											"Knative-Serving-Revision":  "helloworld-go-00001",
										},
									},
									{
										Name:   "default-helloworld-go-rule-0-path-0-split-1",
										Weight: ptr.To(50),
										Headers: map[string]string{
											"Knative-Serving-Namespace": "default",
											"Knative-Serving-Revision":  "helloworld-go-00002",
										},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
				},
			},
		},
		{
			desc:    "TLS",
			paths:   []string{"tls.yaml", "services.yaml"},
			wantLen: 1,
			want: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-helloworld-go-rule-0-path-0": {
							EntryPoints: []string{"http", "https"},
							Service:     "default-helloworld-go-rule-0-path-0-wrr",
							Rule:        "(Host(`helloworld-go.default`) || Host(`helloworld-go.default.svc`) || Host(`helloworld-go.default.svc.cluster.local`))",
							Middlewares: []string{},
						},
						"default-helloworld-go-rule-0-path-0-tls": {
							EntryPoints: []string{"http", "https"},
							Service:     "default-helloworld-go-rule-0-path-0-wrr",
							Rule:        "(Host(`helloworld-go.default`) || Host(`helloworld-go.default.svc`) || Host(`helloworld-go.default.svc.cluster.local`))",
							Middlewares: []string{},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-helloworld-go-rule-0-path-0-split-0": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: types.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.43.38.208:80",
									},
								},
							},
						},
						"default-helloworld-go-rule-0-path-0-split-1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: types.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.43.44.18:80",
									},
								},
							},
						},
						"default-helloworld-go-rule-0-path-0-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-helloworld-go-rule-0-path-0-split-0",
										Weight: ptr.To(50),
										Headers: map[string]string{
											"Knative-Serving-Namespace": "default",
											"Knative-Serving-Revision":  "helloworld-go-00001",
										},
									},
									{
										Name:   "default-helloworld-go-rule-0-path-0-split-1",
										Weight: ptr.To(50),
										Headers: map[string]string{
											"Knative-Serving-Namespace": "default",
											"Knative-Serving-Revision":  "helloworld-go-00002",
										},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			t.Parallel()

			k8sObjects, knObjects := readResources(t, testCase.paths)

			k8sClient := kubefake.NewClientset(k8sObjects...)
			knClient := knfake.NewSimpleClientset(knObjects...)

			client := newClientImpl(knClient, k8sClient)

			eventCh, err := client.WatchAll(nil, make(chan struct{}))
			require.NoError(t, err)

			if len(k8sObjects) > 0 || len(knObjects) > 0 {
				// just wait for the first event
				<-eventCh
			}

			p := Provider{
				PublicEntrypoints:  []string{"http", "https"},
				PrivateEntrypoints: []string{"priv-http", "priv-https"},
				client:             client,
			}

			got, gotIngresses := p.loadConfiguration(t.Context())
			assert.Len(t, gotIngresses, testCase.wantLen)
			assert.Equal(t, testCase.want, got)
		})
	}
}

func Test_buildRule(t *testing.T) {
	testCases := []struct {
		desc    string
		hosts   []string
		headers map[string]knativenetworkingv1alpha1.HeaderMatch
		path    string
		want    string
	}{
		{
			desc:  "single host, no headers, no path",
			hosts: []string{"example.com"},
			want:  "(Host(`example.com`))",
		},
		{
			desc:  "multiple hosts, no headers, no path",
			hosts: []string{"example.com", "foo.com"},
			want:  "(Host(`example.com`) || Host(`foo.com`))",
		},
		{
			desc:  "single host, single header, no path",
			hosts: []string{"example.com"},
			headers: map[string]knativenetworkingv1alpha1.HeaderMatch{
				"X-Header": {Exact: "value"},
			},
			want: "(Host(`example.com`)) && (Header(`X-Header`,`value`))",
		},
		{
			desc:  "single host, multiple headers, no path",
			hosts: []string{"example.com"},
			headers: map[string]knativenetworkingv1alpha1.HeaderMatch{
				"X-Header":  {Exact: "value"},
				"X-Header2": {Exact: "value2"},
			},
			want: "(Host(`example.com`)) && (Header(`X-Header`,`value`) && Header(`X-Header2`,`value2`))",
		},
		{
			desc:  "single host, multiple headers, with path",
			hosts: []string{"example.com"},
			headers: map[string]knativenetworkingv1alpha1.HeaderMatch{
				"X-Header":  {Exact: "value"},
				"X-Header2": {Exact: "value2"},
			},
			path: "/foo",
			want: "(Host(`example.com`)) && (Header(`X-Header`,`value`) && Header(`X-Header2`,`value2`)) && PathPrefix(`/foo`)",
		},
		{
			desc:  "single host, no headers, with path",
			hosts: []string{"example.com"},
			path:  "/foo",
			want:  "(Host(`example.com`)) && PathPrefix(`/foo`)",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got := buildRule(test.hosts, test.headers, test.path)
			assert.Equal(t, test.want, got)
		})
	}
}

func Test_mergeHTTPConfigs(t *testing.T) {
	testCases := []struct {
		desc    string
		configs []*dynamic.HTTPConfiguration
		want    *dynamic.HTTPConfiguration
	}{
		{
			desc: "one empty configuration",
			configs: []*dynamic.HTTPConfiguration{
				{
					Routers: map[string]*dynamic.Router{
						"router1": {Rule: "Host(`example.com`)"},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"middleware1": {Headers: &dynamic.Headers{CustomRequestHeaders: map[string]string{"X-Test": "value"}}},
					},
					Services: map[string]*dynamic.Service{
						"service1": {LoadBalancer: &dynamic.ServersLoadBalancer{Servers: []dynamic.Server{{URL: "http://example.com"}}}},
					},
				},
				{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services:    map[string]*dynamic.Service{},
				},
			},
			want: &dynamic.HTTPConfiguration{
				Routers: map[string]*dynamic.Router{
					"router1": {Rule: "Host(`example.com`)"},
				},
				Middlewares: map[string]*dynamic.Middleware{
					"middleware1": {Headers: &dynamic.Headers{CustomRequestHeaders: map[string]string{"X-Test": "value"}}},
				},
				Services: map[string]*dynamic.Service{
					"service1": {LoadBalancer: &dynamic.ServersLoadBalancer{Servers: []dynamic.Server{{URL: "http://example.com"}}}},
				},
			},
		},
		{
			desc: "merging two non-empty configurations",
			configs: []*dynamic.HTTPConfiguration{
				{
					Routers: map[string]*dynamic.Router{
						"router1": {Rule: "Host(`example.com`)"},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"middleware1": {Headers: &dynamic.Headers{CustomRequestHeaders: map[string]string{"X-Test": "value"}}},
					},
					Services: map[string]*dynamic.Service{
						"service1": {LoadBalancer: &dynamic.ServersLoadBalancer{Servers: []dynamic.Server{{URL: "http://example.com"}}}},
					},
				},
				{
					Routers: map[string]*dynamic.Router{
						"router2": {Rule: "PathPrefix(`/test`)"},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"middleware2": {Headers: &dynamic.Headers{CustomRequestHeaders: map[string]string{"X-Test": "value"}}},
					},
					Services: map[string]*dynamic.Service{
						"service2": {LoadBalancer: &dynamic.ServersLoadBalancer{Servers: []dynamic.Server{{URL: "http://example.com"}}}},
					},
				},
			},
			want: &dynamic.HTTPConfiguration{
				Routers: map[string]*dynamic.Router{
					"router1": {Rule: "Host(`example.com`)"},
					"router2": {Rule: "PathPrefix(`/test`)"},
				},
				Middlewares: map[string]*dynamic.Middleware{
					"middleware1": {Headers: &dynamic.Headers{CustomRequestHeaders: map[string]string{"X-Test": "value"}}},
					"middleware2": {Headers: &dynamic.Headers{CustomRequestHeaders: map[string]string{"X-Test": "value"}}},
				},
				Services: map[string]*dynamic.Service{
					"service1": {LoadBalancer: &dynamic.ServersLoadBalancer{Servers: []dynamic.Server{{URL: "http://example.com"}}}},
					"service2": {LoadBalancer: &dynamic.ServersLoadBalancer{Servers: []dynamic.Server{{URL: "http://example.com"}}}},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got := mergeHTTPConfigs(test.configs...)
			assert.Equal(t, test.want, got)
		})
	}
}

func readResources(t *testing.T, paths []string) ([]runtime.Object, []runtime.Object) {
	t.Helper()

	var (
		k8sObjects []runtime.Object
		knObjects  []runtime.Object
	)
	for _, path := range paths {
		yamlContent, err := os.ReadFile(filepath.FromSlash("./fixtures/" + path))
		if err != nil {
			panic(err)
		}

		objects := k8s.MustParseYaml(yamlContent)
		for _, obj := range objects {
			switch obj.GetObjectKind().GroupVersionKind().Group {
			case "networking.internal.knative.dev":
				knObjects = append(knObjects, obj)
			default:
				k8sObjects = append(k8sObjects, obj)
			}
		}
	}

	return k8sObjects, knObjects
}
