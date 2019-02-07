package kubernetes

import (
	"context"
	"errors"
	"math"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/containous/traefik/config"
	"github.com/containous/traefik/tls"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
)

func TestLoadConfigurationFromIngresses(t *testing.T) {
	testCases := []struct {
		desc          string
		ingressClass  string
		ingresses     []*v1beta1.Ingress
		services      []*corev1.Service
		secrets       []*corev1.Secret
		endpoints     []*corev1.Endpoints
		expected      *config.Configuration
		expectedError bool
	}{
		{
			desc: "Empty ingresses",
			expected: &config.Configuration{
				Routers:     map[string]*config.Router{},
				Middlewares: map[string]*config.Middleware{},
				Services:    map[string]*config.Service{},
			},
		},
		{
			desc: "Ingress with a basic rule on one path",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iNamespace("testing"),
					iRules(
						iRule(
							iPaths(
								onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80)))),
						),
					),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("service1"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(80, ""))),
				),
			},
			endpoints: []*corev1.Endpoints{
				buildEndpoint(
					eNamespace("testing"),
					eName("service1"),
					eUID("1"),
					subset(
						eAddresses(eAddress("10.10.0.1")),
						ePorts(ePort(8080, ""))),
					subset(
						eAddresses(eAddress("10.21.0.1")),
						ePorts(ePort(8080, ""))),
				),
			},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers: map[string]*config.Router{
					"/bar": {
						Rule:    "PathPrefix(`/bar`)",
						Service: "testing/service1/80",
					},
				},
				Services: map[string]*config.Service{
					"testing/service1/80": {
						LoadBalancer: &config.LoadBalancerService{
							Method:         "wrr",
							PassHostHeader: true,
							Servers: []config.Server{
								{
									URL:    "http://10.10.0.1:8080",
									Weight: 1,
								},
								{
									URL:    "http://10.21.0.1:8080",
									Weight: 1,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with two different rules with one path",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iNamespace("testing"),
					iRules(
						iRule(
							iPaths(
								onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80)))),
						),
						iRule(
							iPaths(
								onePath(iPath("/foo"), iBackend("service1", intstr.FromInt(80)))),
						),
					),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("service1"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(80, ""))),
				),
			},
			endpoints: []*corev1.Endpoints{
				buildEndpoint(
					eNamespace("testing"),
					eName("service1"),
					eUID("1"),
					subset(
						eAddresses(eAddress("10.10.0.1")),
						ePorts(ePort(8080, ""))),
					subset(
						eAddresses(eAddress("10.21.0.1")),
						ePorts(ePort(8080, ""))),
				),
			},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers: map[string]*config.Router{
					"/bar": {
						Rule:    "PathPrefix(`/bar`)",
						Service: "testing/service1/80",
					},
					"/foo": {
						Rule:    "PathPrefix(`/foo`)",
						Service: "testing/service1/80",
					},
				},
				Services: map[string]*config.Service{
					"testing/service1/80": {
						LoadBalancer: &config.LoadBalancerService{
							Method:         "wrr",
							PassHostHeader: true,
							Servers: []config.Server{
								{
									URL:    "http://10.10.0.1:8080",
									Weight: 1,
								},
								{
									URL:    "http://10.21.0.1:8080",
									Weight: 1,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress one rule with two paths",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iNamespace("testing"),
					iRules(
						iRule(
							iPaths(
								onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80))),
								onePath(iPath("/foo"), iBackend("service1", intstr.FromInt(80))),
							),
						),
					),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("service1"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(80, ""))),
				),
			},
			endpoints: []*corev1.Endpoints{
				buildEndpoint(
					eNamespace("testing"),
					eName("service1"),
					eUID("1"),
					subset(
						eAddresses(eAddress("10.10.0.1")),
						ePorts(ePort(8080, ""))),
					subset(
						eAddresses(eAddress("10.21.0.1")),
						ePorts(ePort(8080, ""))),
				),
			},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers: map[string]*config.Router{
					"/bar": {
						Rule:    "PathPrefix(`/bar`)",
						Service: "testing/service1/80",
					},
					"/foo": {
						Rule:    "PathPrefix(`/foo`)",
						Service: "testing/service1/80",
					},
				},
				Services: map[string]*config.Service{
					"testing/service1/80": {
						LoadBalancer: &config.LoadBalancerService{
							Method:         "wrr",
							PassHostHeader: true,
							Servers: []config.Server{
								{
									URL:    "http://10.10.0.1:8080",
									Weight: 1,
								},
								{
									URL:    "http://10.21.0.1:8080",
									Weight: 1,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress one rule with one path and one host",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iNamespace("testing"),
					iRules(
						iRule(
							iHost("traefik.tchouk"),
							iPaths(
								onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80))),
							),
						),
					),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("service1"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(80, ""))),
				),
			},
			endpoints: []*corev1.Endpoints{
				buildEndpoint(
					eNamespace("testing"),
					eName("service1"),
					eUID("1"),
					subset(
						eAddresses(eAddress("10.10.0.1")),
						ePorts(ePort(8080, ""))),
					subset(
						eAddresses(eAddress("10.21.0.1")),
						ePorts(ePort(8080, ""))),
				),
			},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers: map[string]*config.Router{
					"traefik-tchouk/bar": {
						Rule:    "Host(`traefik.tchouk`) && PathPrefix(`/bar`)",
						Service: "testing/service1/80",
					},
				},
				Services: map[string]*config.Service{
					"testing/service1/80": {
						LoadBalancer: &config.LoadBalancerService{
							Method:         "wrr",
							PassHostHeader: true,
							Servers: []config.Server{
								{
									URL:    "http://10.10.0.1:8080",
									Weight: 1,
								},
								{
									URL:    "http://10.21.0.1:8080",
									Weight: 1,
								},
							},
						},
					},
				},
			},
		}, {
			desc: "Ingress with one host without path",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iNamespace("testing"),
					iRules(
						iRule(iHost("example.com"), iPaths(
							onePath(iBackend("example-com", intstr.FromInt(80))),
						)),
					),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("example-com"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sType("ClusterIP"),
						sPorts(sPort(80, "http"))),
				),
			},
			endpoints: []*corev1.Endpoints{
				buildEndpoint(
					eNamespace("testing"),
					eName("example-com"),
					eUID("1"),
					subset(
						ePorts(
							ePort(80, "http"),
						),
						eAddresses(eAddress("10.11.0.1")),
					),
				),
			},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers: map[string]*config.Router{
					"example-com": {
						Rule:    "Host(`example.com`)",
						Service: "testing/example-com/80",
					},
				},
				Services: map[string]*config.Service{
					"testing/example-com/80": {
						LoadBalancer: &config.LoadBalancerService{
							Method:         "wrr",
							PassHostHeader: true,
							Servers: []config.Server{
								{
									URL:    "http://10.11.0.1:80",
									Weight: 1,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress one rule with one host and two paths",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iNamespace("testing"),
					iRules(
						iRule(
							iHost("traefik.tchouk"),
							iPaths(
								onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80))),
								onePath(iPath("/foo"), iBackend("service1", intstr.FromInt(80))),
							),
						),
					),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("service1"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(80, ""))),
				),
			},
			endpoints: []*corev1.Endpoints{
				buildEndpoint(
					eNamespace("testing"),
					eName("service1"),
					eUID("1"),
					subset(
						eAddresses(eAddress("10.10.0.1")),
						ePorts(ePort(8080, ""))),
					subset(
						eAddresses(eAddress("10.21.0.1")),
						ePorts(ePort(8080, ""))),
				),
			},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers: map[string]*config.Router{
					"traefik-tchouk/bar": {
						Rule:    "Host(`traefik.tchouk`) && PathPrefix(`/bar`)",
						Service: "testing/service1/80",
					},
					"traefik-tchouk/foo": {
						Rule:    "Host(`traefik.tchouk`) && PathPrefix(`/foo`)",
						Service: "testing/service1/80",
					},
				},
				Services: map[string]*config.Service{
					"testing/service1/80": {
						LoadBalancer: &config.LoadBalancerService{
							Method:         "wrr",
							PassHostHeader: true,
							Servers: []config.Server{
								{
									URL:    "http://10.10.0.1:8080",
									Weight: 1,
								},
								{
									URL:    "http://10.21.0.1:8080",
									Weight: 1,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress Two rules with one host and one path",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iNamespace("testing"),
					iRules(
						iRule(
							iHost("traefik.tchouk"),
							iPaths(
								onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80))),
							),
						),
						iRule(
							iHost("traefik.courgette"),
							iPaths(
								onePath(iPath("/carotte"), iBackend("service1", intstr.FromInt(80))),
							),
						),
					),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("service1"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(80, ""))),
				),
			},
			endpoints: []*corev1.Endpoints{
				buildEndpoint(
					eNamespace("testing"),
					eName("service1"),
					eUID("1"),
					subset(
						eAddresses(eAddress("10.10.0.1")),
						ePorts(ePort(8080, ""))),
					subset(
						eAddresses(eAddress("10.21.0.1")),
						ePorts(ePort(8080, ""))),
				),
			},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers: map[string]*config.Router{
					"traefik-tchouk/bar": {
						Rule:    "Host(`traefik.tchouk`) && PathPrefix(`/bar`)",
						Service: "testing/service1/80",
					},
					"traefik-courgette/carotte": {
						Rule:    "Host(`traefik.courgette`) && PathPrefix(`/carotte`)",
						Service: "testing/service1/80",
					},
				},
				Services: map[string]*config.Service{
					"testing/service1/80": {
						LoadBalancer: &config.LoadBalancerService{
							Method:         "wrr",
							PassHostHeader: true,
							Servers: []config.Server{
								{
									URL:    "http://10.10.0.1:8080",
									Weight: 1,
								},
								{
									URL:    "http://10.21.0.1:8080",
									Weight: 1,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with a bad path syntax",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iNamespace("testing"),
					iRules(
						iRule(
							iPaths(
								onePath(iPath(`/foo`), iBackend("service1", intstr.FromInt(80))),
								onePath(iPath(`/bar-"0"`), iBackend("service1", intstr.FromInt(80))),
								onePath(iPath(`/bar`), iBackend("service1", intstr.FromInt(80))),
							),
						),
					),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("service1"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(80, ""))),
				),
			},
			endpoints: []*corev1.Endpoints{
				buildEndpoint(
					eNamespace("testing"),
					eName("service1"),
					eUID("1"),
					subset(
						eAddresses(eAddress("10.10.0.1")),
						ePorts(ePort(8080, ""))),
					subset(
						eAddresses(eAddress("10.21.0.1")),
						ePorts(ePort(8080, ""))),
				),
			},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers: map[string]*config.Router{
					"/bar": {
						Rule:    "PathPrefix(`/bar`)",
						Service: "testing/service1/80",
					},
					"/foo": {
						Rule:    "PathPrefix(`/foo`)",
						Service: "testing/service1/80",
					},
				},
				Services: map[string]*config.Service{
					"testing/service1/80": {
						LoadBalancer: &config.LoadBalancerService{
							Method:         "wrr",
							PassHostHeader: true,
							Servers: []config.Server{
								{
									URL:    "http://10.10.0.1:8080",
									Weight: 1,
								},
								{
									URL:    "http://10.21.0.1:8080",
									Weight: 1,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with only a bad path syntax",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iNamespace("testing"),
					iRules(
						iRule(
							iPaths(
								onePath(iPath(`/bar-"0"`), iBackend("service1", intstr.FromInt(80))),
							),
						),
					),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("service1"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(80, ""))),
				),
			},
			endpoints: []*corev1.Endpoints{
				buildEndpoint(
					eNamespace("testing"),
					eName("service1"),
					eUID("1"),
					subset(
						eAddresses(eAddress("10.10.0.1")),
						ePorts(ePort(8080, ""))),
					subset(
						eAddresses(eAddress("10.21.0.1")),
						ePorts(ePort(8080, ""))),
				),
			},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers:     map[string]*config.Router{},
				Services:    map[string]*config.Service{},
			},
		},
		{
			desc: "Ingress with a bad host syntax",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iNamespace("testing"),
					iRules(
						iRule(
							iHost(`traefik.tchouk"0"`),
							iPaths(
								onePath(iPath(`/foo`), iBackend("service1", intstr.FromInt(80))),
							),
						),
						iRule(
							iHost("traefik.courgette"),
							iPaths(
								onePath(iPath("/carotte"), iBackend("service1", intstr.FromInt(80))),
							),
						),
					),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("service1"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(80, ""))),
				),
			},
			endpoints: []*corev1.Endpoints{
				buildEndpoint(
					eNamespace("testing"),
					eName("service1"),
					eUID("1"),
					subset(
						eAddresses(eAddress("10.10.0.1")),
						ePorts(ePort(8080, ""))),
					subset(
						eAddresses(eAddress("10.21.0.1")),
						ePorts(ePort(8080, ""))),
				),
			},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers: map[string]*config.Router{
					"traefik-courgette/carotte": {
						Rule:    "Host(`traefik.courgette`) && PathPrefix(`/carotte`)",
						Service: "testing/service1/80",
					},
				},
				Services: map[string]*config.Service{
					"testing/service1/80": {
						LoadBalancer: &config.LoadBalancerService{
							Method:         "wrr",
							PassHostHeader: true,
							Servers: []config.Server{
								{
									URL:    "http://10.10.0.1:8080",
									Weight: 1,
								},
								{
									URL:    "http://10.21.0.1:8080",
									Weight: 1,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with only a bad host syntax",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iNamespace("testing"),
					iRules(
						iRule(
							iHost(`traefik.tchouk"0"`),
							iPaths(
								onePath(iPath(`/foo`), iBackend("service1", intstr.FromInt(80))),
							),
						),
					),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("service1"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(80, ""))),
				),
			},
			endpoints: []*corev1.Endpoints{
				buildEndpoint(
					eNamespace("testing"),
					eName("service1"),
					eUID("1"),
					subset(
						eAddresses(eAddress("10.10.0.1")),
						ePorts(ePort(8080, ""))),
					subset(
						eAddresses(eAddress("10.21.0.1")),
						ePorts(ePort(8080, ""))),
				),
			},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers:     map[string]*config.Router{},
				Services:    map[string]*config.Service{},
			},
		},
		{
			desc: "Ingress with two services",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iNamespace("testing"),
					iRules(
						iRule(
							iHost(`traefik.tchouk`),
							iPaths(
								onePath(iPath(`/bar`), iBackend("service1", intstr.FromInt(80))),
							),
						),
						iRule(
							iHost("traefik.courgette"),
							iPaths(
								onePath(iPath("/carotte"), iBackend("service2", intstr.FromInt(8082))),
							),
						),
					),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("service1"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(80, ""))),
				),
				buildService(
					sName("service2"),
					sNamespace("testing"),
					sUID("2"),
					sSpec(
						clusterIP("10.1.0.1"),
						sPorts(sPort(8082, ""))),
				),
			},
			endpoints: []*corev1.Endpoints{
				buildEndpoint(
					eNamespace("testing"),
					eName("service1"),
					eUID("1"),
					subset(
						eAddresses(eAddress("10.10.0.1")),
						ePorts(ePort(8080, ""))),
					subset(
						eAddresses(eAddress("10.21.0.1")),
						ePorts(ePort(8080, ""))),
				),
				buildEndpoint(
					eNamespace("testing"),
					eName("service2"),
					eUID("2"),
					subset(
						eAddresses(eAddress("10.10.0.2")),
						ePorts(ePort(8080, ""))),
					subset(
						eAddresses(eAddress("10.21.0.2")),
						ePorts(ePort(8080, ""))),
				),
			},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers: map[string]*config.Router{
					"traefik-tchouk/bar": {
						Rule:    "Host(`traefik.tchouk`) && PathPrefix(`/bar`)",
						Service: "testing/service1/80",
					},
					"traefik-courgette/carotte": {
						Rule:    "Host(`traefik.courgette`) && PathPrefix(`/carotte`)",
						Service: "testing/service2/8082",
					},
				},
				Services: map[string]*config.Service{
					"testing/service1/80": {
						LoadBalancer: &config.LoadBalancerService{
							Method:         "wrr",
							PassHostHeader: true,
							Servers: []config.Server{
								{
									URL:    "http://10.10.0.1:8080",
									Weight: 1,
								},
								{
									URL:    "http://10.21.0.1:8080",
									Weight: 1,
								},
							},
						},
					},
					"testing/service2/8082": {
						LoadBalancer: &config.LoadBalancerService{
							Method:         "wrr",
							PassHostHeader: true,
							Servers: []config.Server{
								{
									URL:    "http://10.10.0.2:8080",
									Weight: 1,
								},
								{
									URL:    "http://10.21.0.2:8080",
									Weight: 1,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with one service without endpoints subset",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iNamespace("testing"),
					iRules(
						iRule(
							iHost(`traefik.tchouk`),
							iPaths(
								onePath(iPath(`/bar`), iBackend("service1", intstr.FromInt(80))),
							),
						),
					),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("service1"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(80, ""))),
				),
			},
			endpoints: []*corev1.Endpoints{
				buildEndpoint(
					eNamespace("testing"),
					eName("service1"),
					eUID("1"),
				),
			},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers:     map[string]*config.Router{},
				Services:    map[string]*config.Service{},
			},
		},
		{
			desc: "Ingress with one service without endpoint",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iNamespace("testing"),
					iRules(
						iRule(
							iHost(`traefik.tchouk`),
							iPaths(
								onePath(iPath(`/bar`), iBackend("service1", intstr.FromInt(80))),
							),
						),
					),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("service1"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(80, ""))),
				),
			},
			endpoints: []*corev1.Endpoints{},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers:     map[string]*config.Router{},
				Services:    map[string]*config.Service{},
			},
		},
		{
			desc: "Single Service Ingress (without any rules)",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iNamespace("testing"),
					iSpecBackends(iSpecBackend(iIngressBackend("service1", intstr.FromInt(80)))),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("service1"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(80, ""))),
				),
			},
			endpoints: []*corev1.Endpoints{
				buildEndpoint(
					eNamespace("testing"),
					eName("service1"),
					eUID("1"),
					subset(
						eAddresses(eAddress("10.10.0.1")),
						ePorts(ePort(8080, ""))),
					subset(
						eAddresses(eAddress("10.21.0.1")),
						ePorts(ePort(8080, ""))),
				),
			},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers: map[string]*config.Router{
					"/": {
						Rule:     "PathPrefix(`/`)",
						Service:  "default-backend",
						Priority: math.MinInt32,
					},
				},
				Services: map[string]*config.Service{
					"default-backend": {
						LoadBalancer: &config.LoadBalancerService{
							Method:         "wrr",
							PassHostHeader: true,
							Servers: []config.Server{
								{
									URL:    "http://10.10.0.1:8080",
									Weight: 1,
								},
								{
									URL:    "http://10.21.0.1:8080",
									Weight: 1,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with port value in backend and no pod replica",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iNamespace("testing"),
					iRules(
						iRule(
							iHost(`traefik.tchouk`),
							iPaths(
								onePath(iPath(`/bar`), iBackend("service1", intstr.FromInt(80))),
							),
						),
					),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("service1"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(8082, "carotte")),
						sPorts(sPort(80, "tchouk")),
					),
				),
			},
			endpoints: []*corev1.Endpoints{
				buildEndpoint(
					eNamespace("testing"),
					eName("service1"),
					eUID("1"),
					subset(
						ePorts(
							ePort(8090, "carotte"),
							ePort(8089, "tchouk"),
						),
						eAddresses(eAddress("10.10.0.1")),
					),
					subset(
						ePorts(
							ePort(8090, "carotte"),
							ePort(8089, "tchouk"),
						),
						eAddresses(eAddress("10.21.0.1")),
					),
				),
			},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers: map[string]*config.Router{
					"traefik-tchouk/bar": {
						Rule:    "Host(`traefik.tchouk`) && PathPrefix(`/bar`)",
						Service: "testing/service1/80",
					},
				},
				Services: map[string]*config.Service{
					"testing/service1/80": {
						LoadBalancer: &config.LoadBalancerService{
							Method:         "wrr",
							PassHostHeader: true,
							Servers: []config.Server{
								{
									URL:    "http://10.10.0.1:8089",
									Weight: 1,
								},
								{
									URL:    "http://10.21.0.1:8089",
									Weight: 1,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with port name in backend and no pod replica",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iNamespace("testing"),
					iRules(
						iRule(
							iHost(`traefik.tchouk`),
							iPaths(
								onePath(iPath(`/bar`), iBackend("service1", intstr.FromString("tchouk"))),
							),
						),
					),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("service1"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(8082, "carotte")),
						sPorts(sPort(80, "tchouk")),
					),
				),
			},
			endpoints: []*corev1.Endpoints{
				buildEndpoint(
					eNamespace("testing"),
					eName("service1"),
					eUID("1"),
					subset(
						ePorts(
							ePort(8090, "carotte"),
							ePort(8089, "tchouk"),
						),
						eAddresses(eAddress("10.10.0.1")),
					),
					subset(
						ePorts(
							ePort(8090, "carotte"),
							ePort(8089, "tchouk"),
						),
						eAddresses(eAddress("10.21.0.1")),
					),
				),
			},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers: map[string]*config.Router{
					"traefik-tchouk/bar": {
						Rule:    "Host(`traefik.tchouk`) && PathPrefix(`/bar`)",
						Service: "testing/service1/tchouk",
					},
				},
				Services: map[string]*config.Service{
					"testing/service1/tchouk": {
						LoadBalancer: &config.LoadBalancerService{
							Method:         "wrr",
							PassHostHeader: true,
							Servers: []config.Server{
								{
									URL:    "http://10.10.0.1:8089",
									Weight: 1,
								},
								{
									URL:    "http://10.21.0.1:8089",
									Weight: 1,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with with port name in backend and 2 pod replica",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iNamespace("testing"),
					iRules(
						iRule(
							iHost(`traefik.tchouk`),
							iPaths(
								onePath(iPath(`/bar`), iBackend("service1", intstr.FromString("tchouk"))),
							),
						),
					),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("service1"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(8082, "carotte")),
						sPorts(sPort(80, "tchouk")),
					),
				),
			},
			endpoints: []*corev1.Endpoints{
				buildEndpoint(
					eNamespace("testing"),
					eName("service1"),
					eUID("1"),
					subset(
						ePorts(
							ePort(8090, "carotte"),
							ePort(8089, "tchouk"),
						),
						eAddresses(eAddress("10.10.0.1"), eAddress("10.10.0.2")),
					),
				),
			},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers: map[string]*config.Router{
					"traefik-tchouk/bar": {
						Rule:    "Host(`traefik.tchouk`) && PathPrefix(`/bar`)",
						Service: "testing/service1/tchouk",
					},
				},
				Services: map[string]*config.Service{
					"testing/service1/tchouk": {
						LoadBalancer: &config.LoadBalancerService{
							Method:         "wrr",
							PassHostHeader: true,
							Servers: []config.Server{
								{
									URL:    "http://10.10.0.1:8089",
									Weight: 1,
								},
								{
									URL:    "http://10.10.0.2:8089",
									Weight: 1,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with two paths using same service and different port name",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iNamespace("testing"),
					iRules(
						iRule(
							iHost(`traefik.tchouk`),
							iPaths(
								onePath(iPath(`/bar`), iBackend("service1", intstr.FromString("tchouk"))),
								onePath(iPath(`/foo`), iBackend("service1", intstr.FromString("carotte"))),
							),
						),
					),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("service1"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(8082, "carotte")),
						sPorts(sPort(80, "tchouk")),
					),
				),
			},
			endpoints: []*corev1.Endpoints{
				buildEndpoint(
					eNamespace("testing"),
					eName("service1"),
					eUID("1"),
					subset(
						ePorts(
							ePort(8090, "carotte"),
							ePort(8089, "tchouk"),
						),
						eAddresses(eAddress("10.10.0.1"), eAddress("10.10.0.2")),
					),
				),
			},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers: map[string]*config.Router{
					"traefik-tchouk/bar": {
						Rule:    "Host(`traefik.tchouk`) && PathPrefix(`/bar`)",
						Service: "testing/service1/tchouk",
					},
					"traefik-tchouk/foo": {
						Rule:    "Host(`traefik.tchouk`) && PathPrefix(`/foo`)",
						Service: "testing/service1/carotte",
					},
				},
				Services: map[string]*config.Service{
					"testing/service1/tchouk": {
						LoadBalancer: &config.LoadBalancerService{
							Method:         "wrr",
							PassHostHeader: true,
							Servers: []config.Server{
								{
									URL:    "http://10.10.0.1:8089",
									Weight: 1,
								},
								{
									URL:    "http://10.10.0.2:8089",
									Weight: 1,
								},
							},
						},
					},
					"testing/service1/carotte": {
						LoadBalancer: &config.LoadBalancerService{
							Method:         "wrr",
							PassHostHeader: true,
							Servers: []config.Server{
								{
									URL:    "http://10.10.0.1:8090",
									Weight: 1,
								},
								{
									URL:    "http://10.10.0.2:8090",
									Weight: 1,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "2 ingresses in different namespace with same service name",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iNamespace("testing"),
					iRules(
						iRule(
							iHost(`traefik.tchouk`),
							iPaths(
								onePath(iPath(`/bar`), iBackend("service1", intstr.FromString("tchouk"))),
								onePath(iPath(`/foo`), iBackend("service1", intstr.FromString("carotte"))),
							),
						),
					),
				),
				buildIngress(
					iNamespace("toto"),
					iRules(
						iRule(
							iHost(`toto.traefik.tchouk`),
							iPaths(
								onePath(iPath(`/bar`), iBackend("service1", intstr.FromString("tchouk"))),
							),
						),
					),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("service1"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(80, "tchouk")),
					),
				),
				buildService(
					sName("service1"),
					sNamespace("toto"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(80, "tchouk")),
					),
				),
			},
			endpoints: []*corev1.Endpoints{
				buildEndpoint(
					eNamespace("testing"),
					eName("service1"),
					eUID("1"),
					subset(
						ePorts(
							ePort(8089, "tchouk"),
						),
						eAddresses(eAddress("10.10.0.1"), eAddress("10.10.0.2")),
					),
				),
				buildEndpoint(
					eNamespace("toto"),
					eName("service1"),
					eUID("1"),
					subset(
						ePorts(
							ePort(8089, "tchouk"),
						),
						eAddresses(eAddress("10.11.0.1"), eAddress("10.11.0.2")),
					),
				),
			},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers: map[string]*config.Router{
					"traefik-tchouk/bar": {
						Rule:    "Host(`traefik.tchouk`) && PathPrefix(`/bar`)",
						Service: "testing/service1/tchouk",
					},
					"toto-traefik-tchouk/bar": {
						Rule:    "Host(`toto.traefik.tchouk`) && PathPrefix(`/bar`)",
						Service: "toto/service1/tchouk",
					},
				},
				Services: map[string]*config.Service{
					"testing/service1/tchouk": {
						LoadBalancer: &config.LoadBalancerService{
							Method:         "wrr",
							PassHostHeader: true,
							Servers: []config.Server{
								{
									URL:    "http://10.10.0.1:8089",
									Weight: 1,
								},
								{
									URL:    "http://10.10.0.2:8089",
									Weight: 1,
								},
							},
						},
					},
					"toto/service1/tchouk": {
						LoadBalancer: &config.LoadBalancerService{
							Method:         "wrr",
							PassHostHeader: true,
							Servers: []config.Server{
								{
									URL:    "http://10.11.0.1:8089",
									Weight: 1,
								},
								{
									URL:    "http://10.11.0.2:8089",
									Weight: 1,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with unknown service port name",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iNamespace("testing"),
					iRules(
						iRule(
							iHost(`traefik.tchouk`),
							iPaths(
								onePath(iPath(`/bar`), iBackend("service1", intstr.FromString("toto"))),
							),
						),
					),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("service1"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(80, ""))),
				),
			},
			endpoints: []*corev1.Endpoints{
				buildEndpoint(
					eNamespace("testing"),
					eName("service1"),
					eUID("1"),
					subset(
						ePorts(
							ePort(8089, ""),
						),
						eAddresses(eAddress("10.11.0.1"), eAddress("10.11.0.2")),
					),
				),
			},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers:     map[string]*config.Router{},
				Services:    map[string]*config.Service{},
			},
		},
		{
			desc: "Ingress with unknown service port",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iNamespace("testing"),
					iRules(
						iRule(
							iHost(`traefik.tchouk`),
							iPaths(
								onePath(iPath(`/bar`), iBackend("service1", intstr.FromInt(21))),
							),
						),
					),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("service1"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(80, ""))),
				),
			},
			endpoints: []*corev1.Endpoints{
				buildEndpoint(
					eNamespace("testing"),
					eName("service1"),
					eUID("1"),
					subset(
						ePorts(
							ePort(8089, ""),
						),
						eAddresses(eAddress("10.11.0.1"), eAddress("10.11.0.2")),
					),
				),
			},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers:     map[string]*config.Router{},
				Services:    map[string]*config.Service{},
			},
		},
		{
			desc: "Ingress with service with externalName",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iNamespace("testing"),
					iRules(
						iRule(
							iHost(`traefik.tchouk`),
							iPaths(
								onePath(iPath(`/bar`), iBackend("service1", intstr.FromInt(8080))),
							),
						),
					),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("service1"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(8080, "")),
						sType(corev1.ServiceTypeExternalName), sExternalName("traefik.wtf")),
				),
			},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers: map[string]*config.Router{
					"traefik-tchouk/bar": {
						Rule:    "Host(`traefik.tchouk`) && PathPrefix(`/bar`)",
						Service: "testing/service1/8080",
					},
				},
				Services: map[string]*config.Service{
					"testing/service1/8080": {
						LoadBalancer: &config.LoadBalancerService{
							Method:         "wrr",
							PassHostHeader: true,
							Servers: []config.Server{
								{
									URL:    "http://traefik.wtf:8080",
									Weight: 1,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "TLS support",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iNamespace("testing"),
					iRules(
						iRule(iHost("example.com"), iPaths(
							onePath(iBackend("example-com", intstr.FromInt(80))),
						)),
					),
					iTLSes(
						iTLS("myTlsSecret"),
					),
				),
				buildIngress(
					iNamespace("testing"),
					iRules(
						iRule(iHost("example.fail"), iPaths(
							onePath(iBackend("example-fail", intstr.FromInt(80))),
						)),
					),
					iTLSes(
						iTLS("myUndefinedSecret"),
					),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("example-com"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sType("ClusterIP"),
						sPorts(sPort(80, "http"))),
				),
				buildService(
					sName("example-org"),
					sNamespace("testing"),
					sUID("2"),
					sSpec(
						clusterIP("10.0.0.2"),
						sType("ClusterIP"),
						sPorts(sPort(80, "http"))),
				),
			},
			endpoints: []*corev1.Endpoints{
				buildEndpoint(
					eNamespace("testing"),
					eName("example-com"),
					eUID("1"),
					subset(
						ePorts(
							ePort(80, "http"),
						),
						eAddresses(eAddress("10.11.0.1")),
					),
				),
			},
			secrets: []*corev1.Secret{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "myTlsSecret",
						UID:       "1",
						Namespace: "testing",
					},
					Data: map[string][]byte{
						"tls.crt": []byte("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
						"tls.key": []byte("-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----"),
					},
				},
			},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers: map[string]*config.Router{
					"example-com": {
						Rule:    "Host(`example.com`)",
						Service: "testing/example-com/80",
					},
				},
				Services: map[string]*config.Service{
					"testing/example-com/80": {
						LoadBalancer: &config.LoadBalancerService{
							Method:         "wrr",
							PassHostHeader: true,
							Servers: []config.Server{
								{
									URL:    "http://10.11.0.1:80",
									Weight: 1,
								},
							},
						},
					},
				},
				TLS: []*tls.Configuration{
					{
						Certificate: &tls.Certificate{
							CertFile: tls.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
							KeyFile:  tls.FileOrContent("-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----"),
						},
					},
				},
			},
		},
		{
			desc: "Ingress with a basic rule on one path with https (port == 443)",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iNamespace("testing"),
					iRules(
						iRule(
							iPaths(
								onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(443)))),
						),
					),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("service1"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(443, ""))),
				),
			},
			endpoints: []*corev1.Endpoints{
				buildEndpoint(
					eNamespace("testing"),
					eName("service1"),
					eUID("1"),
					subset(
						eAddresses(eAddress("10.10.0.1")),
						ePorts(ePort(443, ""))),
					subset(
						eAddresses(eAddress("10.21.0.1")),
						ePorts(ePort(443, ""))),
				),
			},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers: map[string]*config.Router{
					"/bar": {
						Rule:    "PathPrefix(`/bar`)",
						Service: "testing/service1/443",
					},
				},
				Services: map[string]*config.Service{
					"testing/service1/443": {
						LoadBalancer: &config.LoadBalancerService{
							Method:         "wrr",
							PassHostHeader: true,
							Servers: []config.Server{
								{
									URL:    "https://10.10.0.1:443",
									Weight: 1,
								},
								{
									URL:    "https://10.21.0.1:443",
									Weight: 1,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with a basic rule on one path with https (portname == https)",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iNamespace("testing"),
					iRules(
						iRule(
							iPaths(
								onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(8443)))),
						),
					),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("service1"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(8443, "https"))),
				),
			},
			endpoints: []*corev1.Endpoints{
				buildEndpoint(
					eNamespace("testing"),
					eName("service1"),
					eUID("1"),
					subset(
						eAddresses(eAddress("10.10.0.1")),
						ePorts(ePort(8443, "https"))),
					subset(
						eAddresses(eAddress("10.21.0.1")),
						ePorts(ePort(8443, "https"))),
				),
			},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers: map[string]*config.Router{
					"/bar": {
						Rule:    "PathPrefix(`/bar`)",
						Service: "testing/service1/8443",
					},
				},
				Services: map[string]*config.Service{
					"testing/service1/8443": {
						LoadBalancer: &config.LoadBalancerService{
							Method:         "wrr",
							PassHostHeader: true,
							Servers: []config.Server{
								{
									URL:    "https://10.10.0.1:8443",
									Weight: 1,
								},
								{
									URL:    "https://10.21.0.1:8443",
									Weight: 1,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Double Single Service Ingress",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iNamespace("testing"),
					iSpecBackends(iSpecBackend(iIngressBackend("service1", intstr.FromInt(80)))),
				),
				buildIngress(
					iNamespace("testing"),
					iSpecBackends(iSpecBackend(iIngressBackend("service2", intstr.FromInt(80)))),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("service1"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(80, ""))),
				),
				buildService(
					sName("service2"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(80, ""))),
				),
			},
			endpoints: []*corev1.Endpoints{
				buildEndpoint(
					eNamespace("testing"),
					eName("service1"),
					eUID("1"),
					subset(
						eAddresses(eAddress("10.30.0.1")),
						ePorts(ePort(8080, ""))),
					subset(
						eAddresses(eAddress("10.41.0.1")),
						ePorts(ePort(8080, ""))),
				),
				buildEndpoint(
					eNamespace("testing"),
					eName("service2"),
					eUID("1"),
					subset(
						eAddresses(eAddress("10.10.0.1")),
						ePorts(ePort(8080, ""))),
					subset(
						eAddresses(eAddress("10.21.0.1")),
						ePorts(ePort(8080, ""))),
				),
			},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers: map[string]*config.Router{
					"/": {
						Rule:     "PathPrefix(`/`)",
						Service:  "default-backend",
						Priority: math.MinInt32,
					},
				},
				Services: map[string]*config.Service{
					"default-backend": {
						LoadBalancer: &config.LoadBalancerService{
							Method:         "wrr",
							PassHostHeader: true,
							Servers: []config.Server{
								{
									URL:    "http://10.30.0.1:8080",
									Weight: 1,
								},
								{
									URL:    "http://10.41.0.1:8080",
									Weight: 1,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with default traefik ingressClass",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iAnnotation(annotationKubernetesIngressClass, traefikDefaultIngressClass),
					iNamespace("testing"),
					iRules(
						iRule(
							iPaths(
								onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80)))),
						),
					),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("service1"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(80, ""))),
				),
			},
			endpoints: []*corev1.Endpoints{
				buildEndpoint(
					eNamespace("testing"),
					eName("service1"),
					eUID("1"),
					subset(
						eAddresses(eAddress("10.10.0.1")),
						ePorts(ePort(8080, ""))),
				),
			},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers: map[string]*config.Router{
					"/bar": {
						Rule:    "PathPrefix(`/bar`)",
						Service: "testing/service1/80",
					},
				},
				Services: map[string]*config.Service{
					"testing/service1/80": {
						LoadBalancer: &config.LoadBalancerService{
							Method:         "wrr",
							PassHostHeader: true,
							Servers: []config.Server{
								{
									URL:    "http://10.10.0.1:8080",
									Weight: 1,
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress without provider traefik ingressClass and unknown annotation",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iAnnotation(annotationKubernetesIngressClass, "tchouk"),
					iNamespace("testing"),
					iRules(
						iRule(
							iPaths(
								onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80)))),
						),
					),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("service1"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(80, ""))),
				),
			},
			endpoints: []*corev1.Endpoints{
				buildEndpoint(
					eNamespace("testing"),
					eName("service1"),
					eUID("1"),
					subset(
						eAddresses(eAddress("10.10.0.1")),
						ePorts(ePort(8080, ""))),
				),
			},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers:     map[string]*config.Router{},
				Services:    map[string]*config.Service{},
			},
		},
		{
			desc:         "Ingress with non matching provider traefik ingressClass and annotation",
			ingressClass: "tchouk",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iAnnotation(annotationKubernetesIngressClass, "toto"),
					iNamespace("testing"),
					iRules(
						iRule(
							iPaths(
								onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80)))),
						),
					),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("service1"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(80, ""))),
				),
			},
			endpoints: []*corev1.Endpoints{
				buildEndpoint(
					eNamespace("testing"),
					eName("service1"),
					eUID("1"),
					subset(
						eAddresses(eAddress("10.10.0.1")),
						ePorts(ePort(8080, ""))),
				),
			},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers:     map[string]*config.Router{},
				Services:    map[string]*config.Service{},
			},
		},
		{
			desc:         "Ingress with ingressClass without annotation",
			ingressClass: "tchouk",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iNamespace("testing"),
					iRules(
						iRule(
							iPaths(
								onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80)))),
						),
					),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("service1"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(80, ""))),
				),
			},
			endpoints: []*corev1.Endpoints{
				buildEndpoint(
					eNamespace("testing"),
					eName("service1"),
					eUID("1"),
					subset(
						eAddresses(eAddress("10.10.0.1")),
						ePorts(ePort(8080, ""))),
				),
			},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers:     map[string]*config.Router{},
				Services:    map[string]*config.Service{},
			},
		},
		{
			desc:         "Ingress with ingressClass without annotation",
			ingressClass: "toto",
			ingresses: []*v1beta1.Ingress{
				buildIngress(
					iAnnotation(annotationKubernetesIngressClass, traefikDefaultIngressClass),
					iNamespace("testing"),
					iRules(
						iRule(
							iPaths(
								onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80)))),
						),
					),
				),
			},
			services: []*corev1.Service{
				buildService(
					sName("service1"),
					sNamespace("testing"),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(80, ""))),
				),
			},
			endpoints: []*corev1.Endpoints{
				buildEndpoint(
					eNamespace("testing"),
					eName("service1"),
					eUID("1"),
					subset(
						eAddresses(eAddress("10.10.0.1")),
						ePorts(ePort(8080, ""))),
				),
			},
			expected: &config.Configuration{
				Middlewares: map[string]*config.Middleware{},
				Routers:     map[string]*config.Router{},
				Services:    map[string]*config.Service{},
			},
		},
	}

	for _, test := range testCases {
		if test.desc != "Ingress with default traefik ingressClass" {
			//continue
		}
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			clientMock := &clientMock{
				ingresses: test.ingresses,
				services:  test.services,
				endpoints: test.endpoints,
				secrets:   test.secrets,
			}

			p := Provider{IngressClass: test.ingressClass}
			conf := p.loadConfigurationFromIngresses(context.Background(), clientMock)

			assert.Equal(t, test.expected, conf)
		})
	}
}

func TestGetTLS(t *testing.T) {
	testIngressWithoutHostname := buildIngress(
		iNamespace("testing"),
		iRules(
			iRule(iHost("ep1.example.com")),
			iRule(iHost("ep2.example.com")),
		),
		iTLSes(
			iTLS("test-secret"),
		),
	)

	testIngressWithoutSecret := buildIngress(
		iNamespace("testing"),
		iRules(
			iRule(iHost("ep1.example.com")),
		),
		iTLSes(
			iTLS("", "foo.com"),
		),
	)

	testCases := []struct {
		desc      string
		ingress   *v1beta1.Ingress
		client    Client
		result    map[string]*tls.Configuration
		errResult string
	}{
		{
			desc:    "api client returns error",
			ingress: testIngressWithoutHostname,
			client: clientMock{
				apiSecretError: errors.New("api secret error"),
			},
			errResult: "failed to fetch secret testing/test-secret: api secret error",
		},
		{
			desc:      "api client doesn't find secret",
			ingress:   testIngressWithoutHostname,
			client:    clientMock{},
			errResult: "secret testing/test-secret does not exist",
		},
		{
			desc:    "entry 'tls.crt' in secret missing",
			ingress: testIngressWithoutHostname,
			client: clientMock{
				secrets: []*corev1.Secret{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-secret",
							Namespace: "testing",
						},
						Data: map[string][]byte{
							"tls.key": []byte("tls-key"),
						},
					},
				},
			},
			errResult: "secret testing/test-secret is missing the following TLS data entries: tls.crt",
		},
		{
			desc:    "entry 'tls.key' in secret missing",
			ingress: testIngressWithoutHostname,
			client: clientMock{
				secrets: []*corev1.Secret{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-secret",
							Namespace: "testing",
						},
						Data: map[string][]byte{
							"tls.crt": []byte("tls-crt"),
						},
					},
				},
			},
			errResult: "secret testing/test-secret is missing the following TLS data entries: tls.key",
		},
		{
			desc:    "secret doesn't provide any of the required fields",
			ingress: testIngressWithoutHostname,
			client: clientMock{
				secrets: []*corev1.Secret{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-secret",
							Namespace: "testing",
						},
						Data: map[string][]byte{},
					},
				},
			},
			errResult: "secret testing/test-secret is missing the following TLS data entries: tls.crt, tls.key",
		},
		{
			desc: "add certificates to the configuration",
			ingress: buildIngress(
				iNamespace("testing"),
				iRules(
					iRule(iHost("ep1.example.com")),
					iRule(iHost("ep2.example.com")),
					iRule(iHost("ep3.example.com")),
				),
				iTLSes(
					iTLS("test-secret"),
					iTLS("test-secret2"),
				),
			),
			client: clientMock{
				secrets: []*corev1.Secret{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-secret2",
							Namespace: "testing",
						},
						Data: map[string][]byte{
							"tls.crt": []byte("tls-crt"),
							"tls.key": []byte("tls-key"),
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-secret",
							Namespace: "testing",
						},
						Data: map[string][]byte{
							"tls.crt": []byte("tls-crt"),
							"tls.key": []byte("tls-key"),
						},
					},
				},
			},
			result: map[string]*tls.Configuration{
				"testing/test-secret": {
					Certificate: &tls.Certificate{
						CertFile: tls.FileOrContent("tls-crt"),
						KeyFile:  tls.FileOrContent("tls-key"),
					},
				},
				"testing/test-secret2": {
					Certificate: &tls.Certificate{
						CertFile: tls.FileOrContent("tls-crt"),
						KeyFile:  tls.FileOrContent("tls-key"),
					},
				},
			},
		},
		{
			desc:    "return nil when no secret is defined",
			ingress: testIngressWithoutSecret,
			client:  clientMock{},
			result:  map[string]*tls.Configuration{},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			tlsConfigs := map[string]*tls.Configuration{}
			err := getTLS(context.Background(), test.ingress, test.client, tlsConfigs)

			if test.errResult != "" {
				assert.EqualError(t, err, test.errResult)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.result, tlsConfigs)
			}
		})
	}
}
