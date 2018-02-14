package kubernetes

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/tls"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestLoadIngresses(t *testing.T) {
	ingresses := []*extensionsv1beta1.Ingress{
		buildIngress(
			iNamespace("testing"),
			iRules(
				iRule(iHost("foo"),
					iPaths(
						onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80))),
						onePath(iPath("/namedthing"), iBackend("service4", intstr.FromString("https")))),
				),
				iRule(iHost("bar"),
					iPaths(
						onePath(iBackend("service3", intstr.FromString("https"))),
						onePath(iBackend("service2", intstr.FromInt(802)))),
				),
			),
		),
	}

	services := []*corev1.Service{
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
				clusterIP("10.0.0.2"),
				sPorts(sPort(802, ""))),
		),
		buildService(
			sName("service3"),
			sNamespace("testing"),
			sUID("3"),
			sSpec(
				clusterIP("10.0.0.3"),
				sPorts(
					sPort(80, "http"),
					sPort(443, "https")),
			),
		),
		buildService(
			sName("service4"),
			sNamespace("testing"),
			sUID("4"),
			sSpec(
				clusterIP("10.0.0.4"),
				sType("ExternalName"),
				sExternalName("example.com"),
				sPorts(sPort(443, "https"))),
		),
	}

	endpoints := []*corev1.Endpoints{
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
			eName("service3"),
			eUID("3"),
			subset(
				eAddresses(eAddress("10.15.0.1")),
				ePorts(
					ePort(8080, "http"),
					ePort(8443, "https")),
			),
			subset(
				eAddresses(eAddress("10.15.0.2")),
				ePorts(
					ePort(9080, "http"),
					ePort(9443, "https")),
			),
		),
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
	require.NoError(t, err, "error loading ingresses")

	expected := buildConfiguration(
		backends(
			backend("foo/bar",
				lbMethod("wrr"),
				servers(
					server("http://10.10.0.1:8080", weight(1)),
					server("http://10.21.0.1:8080", weight(1))),
			),
			backend("foo/namedthing",
				lbMethod("wrr"),
				servers(server("https://example.com", weight(1))),
			),
			backend("bar",
				lbMethod("wrr"),
				servers(
					server("https://10.15.0.1:8443", weight(1)),
					server("https://10.15.0.2:9443", weight(1))),
			),
		),
		frontends(
			frontend("foo/bar",
				passHostHeader(),
				routes(
					route("/bar", "PathPrefix:/bar"),
					route("foo", "Host:foo")),
			),
			frontend("foo/namedthing",
				passHostHeader(),
				routes(
					route("/namedthing", "PathPrefix:/namedthing"),
					route("foo", "Host:foo")),
			),
			frontend("bar",
				passHostHeader(),
				routes(route("bar", "Host:bar")),
			),
		),
	)

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

			ingress := buildIngress(iRules(iRule(
				iHost("host"),
				iPaths(
					onePath(iPath("/path"), iBackend("service", intstr.FromInt(80))),
				),
			)))

			if test.ingressRuleType != "" {
				ingress.Annotations = map[string]string{
					annotationKubernetesRuleType: test.ingressRuleType,
				}
			}

			service := buildService(
				sName("service"),
				sUID("1"),
				sSpec(sPorts(sPort(801, "http"))),
			)

			watchChan := make(chan interface{})
			client := clientMock{
				ingresses: []*extensionsv1beta1.Ingress{ingress},
				services:  []*corev1.Service{service},
				watchChan: watchChan,
			}
			provider := Provider{DisablePassHostHeaders: true}

			actualConfig, err := provider.loadIngresses(client)
			require.NoError(t, err, "error loading ingresses")

			expected := buildFrontends(frontend("host/path",
				routes(
					route("/path", fmt.Sprintf("%s:/path", test.frontendRuleType)),
					route("host", "Host:host")),
			))

			assert.Equal(t, expected, actualConfig.Frontends)
		})
	}
}

func TestGetPassHostHeader(t *testing.T) {
	ingresses := []*extensionsv1beta1.Ingress{
		buildIngress(
			iNamespace("awesome"),
			iRules(iRule(
				iHost("foo"),
				iPaths(onePath(
					iPath("/bar"),
					iBackend("service1", intstr.FromInt(801)))),
			)),
		),
	}

	services := []*corev1.Service{
		buildService(
			sNamespace("awesome"), sName("service1"), sUID("1"),
			sSpec(sPorts(sPort(801, "http"))),
		),
	}

	watchChan := make(chan interface{})
	client := clientMock{
		ingresses: ingresses,
		services:  services,
		watchChan: watchChan,
	}
	provider := Provider{DisablePassHostHeaders: true}

	actual, err := provider.loadIngresses(client)
	require.NoError(t, err, "error loading ingresses")

	expected := buildConfiguration(
		backends(backend("foo/bar", lbMethod("wrr"), servers())),
		frontends(
			frontend("foo/bar",
				routes(
					route("/bar", "PathPrefix:/bar"),
					route("foo", "Host:foo")),
			),
		),
	)

	assert.Equal(t, expected, actual)
}

func TestGetPassTLSCert(t *testing.T) {
	ingresses := []*extensionsv1beta1.Ingress{
		buildIngress(iNamespace("awesome"),
			iRules(iRule(
				iHost("foo"),
				iPaths(onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
	}

	services := []*corev1.Service{
		buildService(
			sName("service1"),
			sNamespace("awesome"),
			sUID("1"),
			sSpec(sPorts(sPort(801, "http"))),
		),
	}

	watchChan := make(chan interface{})
	client := clientMock{
		ingresses: ingresses,
		services:  services,
		watchChan: watchChan,
	}
	provider := Provider{EnablePassTLSCert: true}

	actual, err := provider.loadIngresses(client)
	require.NoError(t, err, "error loading ingresses")

	expected := buildConfiguration(
		backends(backend("foo/bar", lbMethod("wrr"), servers())),
		frontends(frontend("foo/bar",
			passHostHeader(),
			passTLSCert(),
			routes(
				route("/bar", "PathPrefix:/bar"),
				route("foo", "Host:foo")),
		)),
	)

	assert.Equal(t, expected, actual)
}

func TestOnlyReferencesServicesFromOwnNamespace(t *testing.T) {
	ingresses := []*extensionsv1beta1.Ingress{
		buildIngress(iNamespace("awesome"),
			iRules(iRule(
				iHost("foo"),
				iPaths(onePath(iBackend("service", intstr.FromInt(80))))),
			),
		),
	}

	services := []*corev1.Service{
		buildService(
			sNamespace("awesome"),
			sName("service"),
			sUID("1"),
			sSpec(
				clusterIP("10.0.0.1"),
				sPorts(sPort(80, "http"))),
		),
		buildService(
			sNamespace("not-awesome"),
			sName("service"),
			sUID("2"),
			sSpec(
				clusterIP("10.0.0.2"),
				sPorts(sPort(80, "http"))),
		),
	}

	watchChan := make(chan interface{})
	client := clientMock{
		ingresses: ingresses,
		services:  services,
		watchChan: watchChan,
	}
	provider := Provider{}

	actual, err := provider.loadIngresses(client)
	require.NoError(t, err, "error loading ingresses")

	expected := buildConfiguration(
		backends(backend("foo", lbMethod("wrr"), servers())),
		frontends(frontend("foo",
			passHostHeader(),
			routes(route("foo", "Host:foo")),
		)),
	)

	assert.Equal(t, expected, actual)
}

func TestHostlessIngress(t *testing.T) {
	ingresses := []*extensionsv1beta1.Ingress{
		buildIngress(iNamespace("awesome"),
			iRules(iRule(
				iPaths(onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(801))))),
			),
		),
	}

	services := []*corev1.Service{
		buildService(
			sName("service1"),
			sNamespace("awesome"),
			sUID("1"),
			sSpec(
				clusterIP("10.0.0.1"),
				sPorts(sPort(801, "http"))),
		),
	}

	watchChan := make(chan interface{})
	client := clientMock{
		ingresses: ingresses,
		services:  services,
		watchChan: watchChan,
	}
	provider := Provider{DisablePassHostHeaders: true}

	actual, err := provider.loadIngresses(client)
	require.NoError(t, err, "error loading ingresses")

	expected := buildConfiguration(
		backends(backend("/bar", lbMethod("wrr"), servers())),
		frontends(frontend("/bar",
			routes(route("/bar", "PathPrefix:/bar")))),
	)

	assert.Equal(t, expected, actual)
}

func TestServiceAnnotations(t *testing.T) {
	ingresses := []*extensionsv1beta1.Ingress{
		buildIngress(iNamespace("testing"),
			iRules(
				iRule(
					iHost("foo"),
					iPaths(onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80))))),
				iRule(
					iHost("bar"),
					iPaths(onePath(iBackend("service2", intstr.FromInt(802))))),
				iRule(
					iHost("baz"),
					iPaths(onePath(iBackend("service3", intstr.FromInt(803))))),
				iRule(
					iHost("max-conn"),
					iPaths(onePath(iBackend("service4", intstr.FromInt(804))))),
			),
		),
	}

	services := []*corev1.Service{
		buildService(
			sName("service1"),
			sNamespace("testing"),
			sUID("1"),
			sAnnotation(annotationKubernetesCircuitBreakerExpression, "NetworkErrorRatio() > 0.5"),
			sAnnotation(annotationKubernetesLoadBalancerMethod, "drr"),
			sSpec(
				clusterIP("10.0.0.1"),
				sPorts(sPort(80, ""))),
		),
		buildService(
			sName("service2"),
			sNamespace("testing"),
			sUID("2"),
			sAnnotation(annotationKubernetesCircuitBreakerExpression, ""),
			sAnnotation(label.TraefikBackendLoadBalancerSticky, "true"),
			sSpec(
				clusterIP("10.0.0.2"),
				sPorts(sPort(802, ""))),
		),
		buildService(
			sName("service3"),
			sNamespace("testing"),
			sUID("3"),
			sAnnotation(annotationKubernetesBuffering, `
maxrequestbodybytes: 10485760
memrequestbodybytes: 2097153
maxresponsebodybytes: 10485761
memresponsebodybytes: 2097152
retryexpression: IsNetworkError() && Attempts() <= 2
`),
			sSpec(
				clusterIP("10.0.0.3"),
				sPorts(sPort(803, ""))),
		),
		buildService(
			sName("service4"),
			sNamespace("testing"),
			sUID("4"),
			sAnnotation(annotationKubernetesMaxConnExtractorFunc, "client.ip"),
			sAnnotation(annotationKubernetesMaxConnAmount, "6"),
			sSpec(
				clusterIP("10.0.0.4"),
				sPorts(sPort(804, ""))),
		),
	}

	endpoints := []*corev1.Endpoints{
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
				eAddresses(eAddress("10.15.0.1")),
				ePorts(ePort(8080, "http"))),
			subset(
				eAddresses(eAddress("10.15.0.2")),
				ePorts(ePort(8080, "http"))),
		),
		buildEndpoint(
			eNamespace("testing"),
			eName("service3"),
			eUID("3"),
			subset(
				eAddresses(eAddress("10.14.0.1")),
				ePorts(ePort(8080, "http"))),
			subset(
				eAddresses(eAddress("10.12.0.1")),
				ePorts(ePort(8080, "http"))),
		),
		buildEndpoint(
			eNamespace("testing"),
			eName("service4"),
			eUID("4"),
			subset(
				eAddresses(eAddress("10.4.0.1")),
				ePorts(ePort(8080, "http"))),
			subset(
				eAddresses(eAddress("10.4.0.2")),
				ePorts(ePort(8080, "http"))),
		),
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
	require.NoError(t, err, "error loading ingresses")

	expected := buildConfiguration(
		backends(
			backend("foo/bar",
				servers(
					server("http://10.10.0.1:8080", weight(1)),
					server("http://10.21.0.1:8080", weight(1))),
				lbMethod("drr"),
				circuitBreaker("NetworkErrorRatio() > 0.5"),
			),
			backend("bar",
				servers(
					server("http://10.15.0.1:8080", weight(1)),
					server("http://10.15.0.2:8080", weight(1))),
				lbMethod("wrr"), lbSticky(),
			),
			backend("baz",
				servers(
					server("http://10.14.0.1:8080", weight(1)),
					server("http://10.12.0.1:8080", weight(1))),
				lbMethod("wrr"),
				buffering(
					maxRequestBodyBytes(10485760),
					memRequestBodyBytes(2097153),
					maxResponseBodyBytes(10485761),
					memResponseBodyBytes(2097152),
					retrying("IsNetworkError() && Attempts() <= 2"),
				),
			),
			backend("max-conn",
				servers(
					server("http://10.4.0.1:8080", weight(1)),
					server("http://10.4.0.2:8080", weight(1))),
				maxConnExtractorFunc("client.ip"),
				maxConnAmount(6),
				lbMethod("wrr"),
			),
		),
		frontends(
			frontend("foo/bar",
				passHostHeader(),
				routes(
					route("/bar", "PathPrefix:/bar"),
					route("foo", "Host:foo")),
			),
			frontend("bar",
				passHostHeader(),
				routes(route("bar", "Host:bar"))),
			frontend("baz",
				passHostHeader(),
				routes(route("baz", "Host:baz"))),
			frontend("max-conn",
				passHostHeader(),
				routes(
					route("max-conn", "Host:max-conn"))),
		),
	)

	assert.EqualValues(t, expected, actual)
}

func TestIngressAnnotations(t *testing.T) {
	ingresses := []*extensionsv1beta1.Ingress{
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesPreserveHost, "false"),
			iRules(
				iRule(
					iHost("foo"),
					iPaths(onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesPreserveHost, "true"),
			iAnnotation(annotationKubernetesIngressClass, traefikDefaultRealm),
			iRules(
				iRule(
					iHost("other"),
					iPaths(onePath(iPath("/stuff"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesPassTLSCert, "true"),
			iAnnotation(annotationKubernetesIngressClass, traefikDefaultRealm),
			iRules(
				iRule(
					iHost("other"),
					iPaths(onePath(iPath("/sslstuff"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesFrontendEntryPoints, "http,https"),
			iAnnotation(annotationKubernetesIngressClass, traefikDefaultRealm),
			iRules(
				iRule(
					iHost("other"),
					iPaths(onePath(iPath("/"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesAuthType, "basic"),
			iAnnotation(annotationKubernetesAuthSecret, "mySecret"),
			iRules(
				iRule(
					iHost("basic"),
					iPaths(onePath(iPath("/auth"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesIngressClass, traefikDefaultRealm+"-other"),
			iRules(
				iRule(
					iHost("herp"),
					iPaths(onePath(iPath("/derp"), iBackend("service2", intstr.FromInt(80))))),
			),
		),
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesWhitelistSourceRange, "1.1.1.1/24, 1234:abcd::42/32"),
			iRules(
				iRule(
					iHost("test"),
					iPaths(onePath(iPath("/whitelist-source-range"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesRewriteTarget, "/"),
			iRules(
				iRule(
					iHost("rewrite"),
					iPaths(onePath(iPath("/api"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesAuthRealm, "customized"),
			iRules(
				iRule(
					iHost("auth-realm-customized"),
					iPaths(onePath(iPath("/auth-realm-customized"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesRedirectEntryPoint, "https"),
			iRules(
				iRule(
					iHost("redirect"),
					iPaths(onePath(iPath("/https"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesIngressClass, "traefik"),
			iAnnotation(annotationKubernetesErrorPages, `
foo:
  status:
  - "123"
  - "456"
  backend: bar
  query: /bar
`),
			iRules(
				iRule(
					iHost("error-pages"),
					iPaths(onePath(iPath("/errorpages"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesIngressClass, "traefik"),
			iAnnotation(annotationKubernetesRateLimit, `
extractorfunc: client.ip
rateset:
  bar:
    period: 3s
    average: 6
    burst: 9
  foo:
    period: 6s
    average: 12
    burst: 18
`),
			iRules(
				iRule(
					iHost("rate-limit"),
					iPaths(onePath(iPath("/ratelimit"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesIngressClass, "traefik"),
			iAnnotation(annotationKubernetesCustomRequestHeaders, "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8"),
			iAnnotation(annotationKubernetesCustomResponseHeaders, "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8"),
			iAnnotation(annotationKubernetesSSLProxyHeaders, "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8"),
			iAnnotation(annotationKubernetesAllowedHosts, "foo, fii, fuu"),
			iAnnotation(annotationKubernetesProxyHeaders, "foo, fii, fuu"),
			iAnnotation(annotationKubernetesHSTSMaxAge, "666"),
			iAnnotation(annotationKubernetesSSLRedirect, "true"),
			iAnnotation(annotationKubernetesSSLTemporaryRedirect, "true"),
			iAnnotation(annotationKubernetesHSTSIncludeSubdomains, "true"),
			iAnnotation(annotationKubernetesForceHSTSHeader, "true"),
			iAnnotation(annotationKubernetesHSTSPreload, "true"),
			iAnnotation(annotationKubernetesFrameDeny, "true"),
			iAnnotation(annotationKubernetesContentTypeNosniff, "true"),
			iAnnotation(annotationKubernetesBrowserXSSFilter, "true"),
			iAnnotation(annotationKubernetesIsDevelopment, "true"),
			iAnnotation(annotationKubernetesSSLHost, "foo"),
			iAnnotation(annotationKubernetesCustomFrameOptionsValue, "foo"),
			iAnnotation(annotationKubernetesContentSecurityPolicy, "foo"),
			iAnnotation(annotationKubernetesPublicKey, "foo"),
			iAnnotation(annotationKubernetesReferrerPolicy, "foo"),
			iRules(
				iRule(
					iHost("custom-headers"),
					iPaths(onePath(iPath("/customheaders"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
	}

	services := []*corev1.Service{
		buildService(
			sName("service1"),
			sNamespace("testing"),
			sUID("1"),
			sSpec(
				clusterIP("10.0.0.1"),
				sType("ExternalName"),
				sExternalName("example.com"),
				sPorts(sPort(80, "http"))),
		),
		buildService(
			sName("service2"),
			sNamespace("testing"),
			sUID("2"),
			sSpec(
				clusterIP("10.0.0.2"),
				sPorts(sPort(802, ""))),
		),
	}

	secrets := []*corev1.Secret{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "mySecret",
				UID:       "1",
				Namespace: "testing",
			},
			Data: map[string][]byte{"auth": []byte("myUser:myEncodedPW")},
		},
	}

	watchChan := make(chan interface{})
	client := clientMock{
		ingresses: ingresses,
		services:  services,
		secrets:   secrets,
		watchChan: watchChan,
	}
	provider := Provider{}

	actual, err := provider.loadIngresses(client)
	require.NoError(t, err, "error loading ingresses")

	expected := buildConfiguration(
		backends(
			backend("foo/bar",
				servers(
					server("http://example.com", weight(1)),
					server("http://example.com", weight(1))),
				lbMethod("wrr"),
			),
			backend("other/stuff",
				servers(
					server("http://example.com", weight(1)),
					server("http://example.com", weight(1))),
				lbMethod("wrr"),
			),
			backend("other/",
				servers(
					server("http://example.com", weight(1)),
					server("http://example.com", weight(1))),
				lbMethod("wrr"),
			),
			backend("other/sslstuff",
				servers(
					server("http://example.com", weight(1)),
					server("http://example.com", weight(1))),
				lbMethod("wrr"),
			),
			backend("basic/auth",
				servers(
					server("http://example.com", weight(1)),
					server("http://example.com", weight(1))),
				lbMethod("wrr"),
			),
			backend("redirect/https",
				servers(
					server("http://example.com", weight(1)),
					server("http://example.com", weight(1))),
				lbMethod("wrr"),
			),
			backend("test/whitelist-source-range",
				servers(
					server("http://example.com", weight(1)),
					server("http://example.com", weight(1))),
				lbMethod("wrr"),
			),
			backend("rewrite/api",
				servers(
					server("http://example.com", weight(1)),
					server("http://example.com", weight(1))),
				lbMethod("wrr"),
			),
			backend("error-pages/errorpages",
				servers(
					server("http://example.com", weight(1)),
					server("http://example.com", weight(1))),
				lbMethod("wrr"),
			),
			backend("rate-limit/ratelimit",
				servers(
					server("http://example.com", weight(1)),
					server("http://example.com", weight(1))),
				lbMethod("wrr"),
			),
			backend("custom-headers/customheaders",
				servers(
					server("http://example.com", weight(1)),
					server("http://example.com", weight(1))),
				lbMethod("wrr"),
			),
		),
		frontends(
			frontend("foo/bar",
				routes(
					route("/bar", "PathPrefix:/bar"),
					route("foo", "Host:foo")),
			),
			frontend("other/stuff",
				passHostHeader(),
				routes(
					route("/stuff", "PathPrefix:/stuff"),
					route("other", "Host:other")),
			),
			frontend("other/",
				passHostHeader(),
				entryPoints("http", "https"),
				routes(
					route("/", "PathPrefix:/"),
					route("other", "Host:other")),
			),
			frontend("other/sslstuff",
				passHostHeader(),
				passTLSCert(),
				routes(
					route("/sslstuff", "PathPrefix:/sslstuff"),
					route("other", "Host:other")),
			),
			frontend("other/sslstuff",
				passHostHeader(),
				passTLSCert(),
				routes(
					route("/sslstuff", "PathPrefix:/sslstuff"),
					route("other", "Host:other")),
			),
			frontend("basic/auth",
				passHostHeader(),
				basicAuth("myUser:myEncodedPW"),
				routes(
					route("/auth", "PathPrefix:/auth"),
					route("basic", "Host:basic")),
			),
			frontend("redirect/https",
				passHostHeader(),
				redirectEntryPoint("https"),
				routes(
					route("/https", "PathPrefix:/https"),
					route("redirect", "Host:redirect")),
			),
			frontend("test/whitelist-source-range",
				passHostHeader(),
				whitelistSourceRange("1.1.1.1/24", "1234:abcd::42/32"),
				routes(
					route("/whitelist-source-range", "PathPrefix:/whitelist-source-range"),
					route("test", "Host:test")),
			),
			frontend("rewrite/api",
				passHostHeader(),
				routes(
					route("/api", "PathPrefix:/api;ReplacePath:/"),
					route("rewrite", "Host:rewrite")),
			),
			frontend("error-pages/errorpages",
				passHostHeader(),
				errorPage("foo", errorQuery("/bar"), errorStatus("123", "456"), errorBackend("bar")),
				routes(
					route("/errorpages", "PathPrefix:/errorpages"),
					route("error-pages", "Host:error-pages")),
			),
			frontend("rate-limit/ratelimit",
				passHostHeader(),
				rateLimit(rateExtractorFunc("client.ip"),
					rateSet("foo", limitPeriod(6*time.Second), limitAverage(12), limitBurst(18)),
					rateSet("bar", limitPeriod(3*time.Second), limitAverage(6), limitBurst(9))),
				routes(
					route("/ratelimit", "PathPrefix:/ratelimit"),
					route("rate-limit", "Host:rate-limit")),
			),
			frontend("custom-headers/customheaders",
				passHostHeader(),
				headers(&types.Headers{
					CustomRequestHeaders: map[string]string{
						"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
						"Content-Type":                 "application/json; charset=utf-8",
					},
					CustomResponseHeaders: map[string]string{
						"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
						"Content-Type":                 "application/json; charset=utf-8",
					},
					SSLProxyHeaders: map[string]string{
						"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
						"Content-Type":                 "application/json; charset=utf-8",
					},
					AllowedHosts:            []string{"foo", "fii", "fuu"},
					HostsProxyHeaders:       []string{"foo", "fii", "fuu"},
					STSSeconds:              666,
					SSLRedirect:             true,
					SSLTemporaryRedirect:    true,
					STSIncludeSubdomains:    true,
					STSPreload:              true,
					ForceSTSHeader:          true,
					FrameDeny:               true,
					ContentTypeNosniff:      true,
					BrowserXSSFilter:        true,
					IsDevelopment:           true,
					CustomFrameOptionsValue: "foo",
					SSLHost:                 "foo",
					ContentSecurityPolicy:   "foo",
					PublicKey:               "foo",
					ReferrerPolicy:          "foo",
				}),
				routes(
					route("/customheaders", "PathPrefix:/customheaders"),
					route("custom-headers", "Host:custom-headers")),
			),
		),
	)

	assert.Equal(t, expected, actual)
}

func TestIngressClassAnnotation(t *testing.T) {
	ingresses := []*extensionsv1beta1.Ingress{
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesIngressClass, traefikDefaultIngressClass),
			iRules(
				iRule(
					iHost("other"),
					iPaths(onePath(iPath("/stuff"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesIngressClass, ""),
			iRules(
				iRule(
					iHost("other"),
					iPaths(onePath(iPath("/sslstuff"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
		buildIngress(
			iNamespace("testing"),
			iRules(
				iRule(
					iHost("other"),
					iPaths(onePath(iPath("/"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesIngressClass, traefikDefaultIngressClass+"-other"),
			iRules(
				iRule(
					iHost("herp"),
					iPaths(onePath(iPath("/derp"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
	}

	services := []*corev1.Service{
		buildService(
			sName("service1"),
			sNamespace("testing"),
			sUID("1"),
			sSpec(
				clusterIP("10.0.0.1"),
				sType("ExternalName"),
				sExternalName("example.com"),
				sPorts(sPort(80, "http"))),
		),
	}

	watchChan := make(chan interface{})
	client := clientMock{
		ingresses: ingresses,
		services:  services,
		watchChan: watchChan,
	}

	testCases := []struct {
		desc     string
		provider Provider
		expected *types.Configuration
	}{
		{
			desc:     "Empty IngressClass annotation",
			provider: Provider{},
			expected: buildConfiguration(
				backends(
					backend("other/stuff",
						servers(
							server("http://example.com", weight(1)),
							server("http://example.com", weight(1))),
						lbMethod("wrr"),
					),
					backend("other/",
						servers(
							server("http://example.com", weight(1)),
							server("http://example.com", weight(1))),
						lbMethod("wrr"),
					),
					backend("other/sslstuff",
						servers(
							server("http://example.com", weight(1)),
							server("http://example.com", weight(1))),
						lbMethod("wrr"),
					),
				),
				frontends(
					frontend("other/stuff",
						passHostHeader(),
						routes(
							route("/stuff", "PathPrefix:/stuff"),
							route("other", "Host:other")),
					),
					frontend("other/",
						passHostHeader(),
						routes(
							route("/", "PathPrefix:/"),
							route("other", "Host:other")),
					),
					frontend("other/sslstuff",
						passHostHeader(),
						routes(
							route("/sslstuff", "PathPrefix:/sslstuff"),
							route("other", "Host:other")),
					),
				),
			),
		},
		{
			desc:     "Provided IngressClass annotation",
			provider: Provider{IngressClass: traefikDefaultRealm + "-other"},
			expected: buildConfiguration(
				backends(
					backend("herp/derp",
						servers(
							server("http://example.com", weight(1)),
							server("http://example.com", weight(1))),
						lbMethod("wrr"),
					),
				),
				frontends(
					frontend("herp/derp",
						passHostHeader(),
						routes(
							route("/derp", "PathPrefix:/derp"),
							route("herp", "Host:herp")),
					),
				),
			),
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual, err := test.provider.loadIngresses(client)
			require.NoError(t, err, "error loading ingresses")

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestPriorityHeaderValue(t *testing.T) {
	ingresses := []*extensionsv1beta1.Ingress{
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesPriority, "1337"),
			iRules(
				iRule(
					iHost("foo"),
					iPaths(onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
	}

	services := []*corev1.Service{
		buildService(
			sName("service1"),
			sNamespace("testing"),
			sUID("1"),
			sSpec(
				clusterIP("10.0.0.1"),
				sType("ExternalName"),
				sExternalName("example.com"),
				sPorts(sPort(80, "http"))),
		),
	}

	var endpoints []*corev1.Endpoints
	watchChan := make(chan interface{})
	client := clientMock{
		ingresses: ingresses,
		services:  services,
		endpoints: endpoints,
		watchChan: watchChan,
	}
	provider := Provider{}

	actual, err := provider.loadIngresses(client)
	require.NoError(t, err, "error loading ingresses")

	expected := buildConfiguration(
		backends(
			backend("foo/bar",
				servers(server("http://example.com", weight(1))),
				lbMethod("wrr"),
			),
		),
		frontends(
			frontend("foo/bar",
				passHostHeader(),
				priority(1337),
				routes(
					route("/bar", "PathPrefix:/bar"),
					route("foo", "Host:foo")),
			),
		),
	)

	assert.Equal(t, expected, actual)
}

func TestInvalidPassTLSCertValue(t *testing.T) {
	ingresses := []*extensionsv1beta1.Ingress{
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesPassTLSCert, "herpderp"),
			iRules(
				iRule(
					iHost("foo"),
					iPaths(onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
	}

	services := []*corev1.Service{
		buildService(
			sName("service1"),
			sNamespace("testing"),
			sUID("1"),
			sSpec(
				clusterIP("10.0.0.1"),
				sType("ExternalName"),
				sExternalName("example.com"),
				sPorts(sPort(80, "http"))),
		),
	}

	watchChan := make(chan interface{})
	client := clientMock{
		ingresses: ingresses,
		services:  services,
		watchChan: watchChan,
	}
	provider := Provider{}

	actual, err := provider.loadIngresses(client)
	require.NoError(t, err, "error loading ingresses")

	expected := buildConfiguration(
		backends(
			backend("foo/bar",
				servers(server("http://example.com", weight(1))),
				lbMethod("wrr"),
			),
		),
		frontends(
			frontend("foo/bar",
				passHostHeader(),
				routes(
					route("/bar", "PathPrefix:/bar"),
					route("foo", "Host:foo")),
			),
		),
	)

	assert.Equal(t, expected, actual)
}

func TestInvalidPassHostHeaderValue(t *testing.T) {
	ingresses := []*extensionsv1beta1.Ingress{
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesPreserveHost, "herpderp"),
			iRules(
				iRule(
					iHost("foo"),
					iPaths(onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
	}

	services := []*corev1.Service{
		buildService(
			sName("service1"),
			sNamespace("testing"),
			sUID("1"),
			sSpec(
				clusterIP("10.0.0.1"),
				sType("ExternalName"),
				sExternalName("example.com"),
				sPorts(sPort(80, "http"))),
		),
	}

	watchChan := make(chan interface{})
	client := clientMock{
		ingresses: ingresses,
		services:  services,
		watchChan: watchChan,
	}
	provider := Provider{}

	actual, err := provider.loadIngresses(client)
	require.NoError(t, err, "error loading ingresses")

	expected := buildConfiguration(
		backends(
			backend("foo/bar",
				servers(server("http://example.com", weight(1))),
				lbMethod("wrr"),
			),
		),
		frontends(
			frontend("foo/bar",
				passHostHeader(),
				routes(
					route("/bar", "PathPrefix:/bar"),
					route("foo", "Host:foo")),
			),
		),
	)

	assert.Equal(t, expected, actual)
}

func TestKubeAPIErrors(t *testing.T) {
	ingresses := []*extensionsv1beta1.Ingress{
		buildIngress(
			iNamespace("testing"),
			iRules(
				iRule(
					iHost("foo"),
					iPaths(onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
	}

	services := []*corev1.Service{
		buildService(
			sName("service1"),
			sNamespace("testing"),
			sUID("1"),
			sSpec(
				clusterIP("10.0.0.1"),
				sPorts(sPort(80, ""))),
		),
	}

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
	ingresses := []*extensionsv1beta1.Ingress{
		buildIngress(
			iNamespace("testing"),
			iRules(
				iRule(
					iHost("fully_working"),
					iPaths(onePath(iBackend("fully_working_service", intstr.FromInt(80))))),
				iRule(
					iHost("missing_service"),
					iPaths(onePath(iBackend("missing_service_service", intstr.FromInt(80))))),
				iRule(
					iHost("missing_endpoints"),
					iPaths(onePath(iBackend("missing_endpoints_service", intstr.FromInt(80))))),
				iRule(
					iHost("missing_endpoint_subsets"),
					iPaths(onePath(iBackend("missing_endpoint_subsets_service", intstr.FromInt(80))))),
			),
		),
	}

	services := []*corev1.Service{
		buildService(
			sName("fully_working_service"),
			sNamespace("testing"),
			sUID("1"),
			sSpec(
				clusterIP("10.0.0.1"),
				sPorts(sPort(80, ""))),
		),
		buildService(
			sName("missing_endpoints_service"),
			sNamespace("testing"),
			sUID("3"),
			sSpec(
				clusterIP("10.0.0.3"),
				sPorts(sPort(80, ""))),
		),
		buildService(
			sName("missing_endpoint_subsets_service"),
			sNamespace("testing"),
			sUID("4"),
			sSpec(
				clusterIP("10.0.0.4"),
				sPorts(sPort(80, ""))),
		),
	}

	endpoints := []*corev1.Endpoints{
		buildEndpoint(
			eName("fully_working_service"),
			eUID("1"),
			eNamespace("testing"),
			subset(
				eAddresses(eAddress("10.10.0.1")),
				ePorts(ePort(8080, ""))),
		),
		buildEndpoint(
			eName("missing_endpoint_subsets_service"),
			eUID("4"),
			eNamespace("testing"),
			subset(),
		),
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
	require.NoError(t, err, "error loading ingresses")

	expected := buildConfiguration(
		backends(
			backend("fully_working",
				servers(server("http://10.10.0.1:8080", weight(1))),
				lbMethod("wrr"),
			),
			backend("missing_service",
				servers(),
				lbMethod("wrr"),
			),
			backend("missing_endpoints",
				servers(),
				lbMethod("wrr"),
			),
			backend("missing_endpoint_subsets",
				servers(),
				lbMethod("wrr"),
			),
		),
		frontends(
			frontend("fully_working",
				passHostHeader(),
				routes(route("fully_working", "Host:fully_working")),
			),
			frontend("missing_endpoints",
				passHostHeader(),
				routes(route("missing_endpoints", "Host:missing_endpoints")),
			),
			frontend("missing_endpoint_subsets",
				passHostHeader(),
				routes(route("missing_endpoint_subsets", "Host:missing_endpoint_subsets")),
			),
		),
	)

	assert.Equal(t, expected, actual)
}

func TestBasicAuthInTemplate(t *testing.T) {
	ingresses := []*extensionsv1beta1.Ingress{
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesAuthType, "basic"),
			iAnnotation(annotationKubernetesAuthSecret, "mySecret"),
			iRules(
				iRule(
					iHost("basic"),
					iPaths(onePath(iPath("/auth"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
	}

	services := []*corev1.Service{
		buildService(
			sName("service1"),
			sNamespace("testing"),
			sUID("1"),
			sSpec(
				clusterIP("10.0.0.1"),
				sType("ExternalName"),
				sExternalName("example.com"),
				sPorts(sPort(80, "http"))),
		),
	}

	secrets := []*corev1.Secret{{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mySecret",
			UID:       "1",
			Namespace: "testing",
		},
		Data: map[string][]byte{
			"auth": []byte("myUser:myEncodedPW"),
		},
	}}

	var endpoints []*corev1.Endpoints
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
	require.NoError(t, err, "error loading ingresses")

	actual = provider.loadConfig(*actual)
	require.NotNil(t, actual)
	got := actual.Frontends["basic/auth"].BasicAuth
	if !reflect.DeepEqual(got, []string{"myUser:myEncodedPW"}) {
		t.Fatalf("unexpected credentials: %+v", got)
	}
}

func TestTLSSecretLoad(t *testing.T) {
	ingresses := []*extensionsv1beta1.Ingress{
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesFrontendEntryPoints, "ep1,ep2"),
			iRules(
				iRule(iHost("example.com"), iPaths(
					onePath(iBackend("example-com", intstr.FromInt(80))),
				)),
				iRule(iHost("example.org"), iPaths(
					onePath(iBackend("example-org", intstr.FromInt(80))),
				)),
			),
			iTLSes(
				iTLS("myTlsSecret"),
			),
		),
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesFrontendEntryPoints, "ep3"),
			iRules(
				iRule(iHost("example.fail"), iPaths(
					onePath(iBackend("example-fail", intstr.FromInt(80))),
				)),
			),
			iTLSes(
				iTLS("myUndefinedSecret"),
			),
		),
	}
	services := []*corev1.Service{
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
	}
	secrets := []*corev1.Secret{
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
	}
	endpoints := []*corev1.Endpoints{}
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

	expected := buildConfiguration(
		backends(
			backend("example.com",
				servers(),
				lbMethod("wrr"),
			),
			backend("example.org",
				servers(),
				lbMethod("wrr"),
			),
		),
		frontends(
			frontend("example.com",
				entryPoints("ep1", "ep2"),
				passHostHeader(),
				routes(
					route("example.com", "Host:example.com"),
				),
			),
			frontend("example.org",
				entryPoints("ep1", "ep2"),
				passHostHeader(),
				routes(
					route("example.org", "Host:example.org"),
				),
			),
		),
		tlsesSection(
			tlsSection(
				tlsEntryPoints("ep1", "ep2"),
				certificate(
					"-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----",
					"-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----"),
			),
		),
	)

	assert.Equal(t, expected, actual)
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

	tests := []struct {
		desc      string
		ingress   *extensionsv1beta1.Ingress
		client    Client
		result    []*tls.Configuration
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
					iTLS("test-secret"),
				),
			),
			client: clientMock{
				secrets: []*corev1.Secret{
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
			result: []*tls.Configuration{
				{
					Certificate: &tls.Certificate{
						CertFile: tls.FileOrContent("tls-crt"),
						KeyFile:  tls.FileOrContent("tls-key"),
					},
				},
				{
					Certificate: &tls.Certificate{
						CertFile: tls.FileOrContent("tls-crt"),
						KeyFile:  tls.FileOrContent("tls-key"),
					},
				},
			},
		},
		{
			desc: "pass the endpoints defined in the annotation to the certificate",
			ingress: buildIngress(
				iNamespace("testing"),
				iAnnotation(annotationKubernetesFrontendEntryPoints, "https,api-secure"),
				iRules(iRule(iHost("example.com"))),
				iTLSes(iTLS("test-secret")),
			),
			client: clientMock{
				secrets: []*corev1.Secret{
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
			result: []*tls.Configuration{
				{
					EntryPoints: []string{"https", "api-secure"},
					Certificate: &tls.Certificate{
						CertFile: tls.FileOrContent("tls-crt"),
						KeyFile:  tls.FileOrContent("tls-key"),
					},
				},
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			tlsConfigs, err := getTLS(test.ingress, test.client)

			if test.errResult != "" {
				assert.EqualError(t, err, test.errResult)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.result, tlsConfigs)
			}
		})
	}
}
