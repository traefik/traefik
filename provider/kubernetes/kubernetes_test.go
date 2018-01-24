package kubernetes

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/tls"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/util/intstr"
)

func TestLoadIngresses(t *testing.T) {
	ingresses := []*v1beta1.Ingress{
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

	services := []*v1.Service{
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

	endpoints := []*v1.Endpoints{
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
				headers(),
				passHostHeader(),
				routes(
					route("/bar", "PathPrefix:/bar"),
					route("foo", "Host:foo")),
			),
			frontend("foo/namedthing",
				headers(),
				passHostHeader(),
				routes(
					route("/namedthing", "PathPrefix:/namedthing"),
					route("foo", "Host:foo")),
			),
			frontend("bar",
				headers(),
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
				ingress.ObjectMeta.Annotations = map[string]string{
					label.TraefikFrontendRuleType: test.ingressRuleType,
				}
			}

			service := buildService(
				sName("service"),
				sUID("1"),
				sSpec(sPorts(sPort(801, "http"))),
			)

			watchChan := make(chan interface{})
			client := clientMock{
				ingresses: []*v1beta1.Ingress{ingress},
				services:  []*v1.Service{service},
				watchChan: watchChan,
			}
			provider := Provider{DisablePassHostHeaders: true}

			actualConfig, err := provider.loadIngresses(client)
			require.NoError(t, err, "error loading ingresses")

			expected := buildFrontends(frontend("host/path",
				headers(),
				routes(
					route("/path", fmt.Sprintf("%s:/path", test.frontendRuleType)),
					route("host", "Host:host")),
			))

			assert.Equal(t, expected, actualConfig.Frontends)
		})
	}
}

func TestGetPassHostHeader(t *testing.T) {
	ingresses := []*v1beta1.Ingress{
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

	services := []*v1.Service{
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
				headers(),
				routes(
					route("/bar", "PathPrefix:/bar"),
					route("foo", "Host:foo")),
			),
		),
	)

	assert.Equal(t, expected, actual)
}

func TestGetPassTLSCert(t *testing.T) {
	ingresses := []*v1beta1.Ingress{
		buildIngress(iNamespace("awesome"),
			iRules(iRule(
				iHost("foo"),
				iPaths(onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
	}

	services := []*v1.Service{
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
			headers(),
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
	ingresses := []*v1beta1.Ingress{
		buildIngress(iNamespace("awesome"),
			iRules(iRule(
				iHost("foo"),
				iPaths(onePath(iBackend("service", intstr.FromInt(80))))),
			),
		),
	}

	services := []*v1.Service{
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
			headers(),
			passHostHeader(),
			routes(route("foo", "Host:foo")),
		)),
	)

	assert.Equal(t, expected, actual)
}

func TestHostlessIngress(t *testing.T) {
	ingresses := []*v1beta1.Ingress{
		buildIngress(iNamespace("awesome"),
			iRules(iRule(
				iPaths(onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(801))))),
			),
		),
	}

	services := []*v1.Service{
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
			headers(),
			routes(route("/bar", "PathPrefix:/bar")))),
	)

	assert.Equal(t, expected, actual)
}

func TestServiceAnnotations(t *testing.T) {
	ingresses := []*v1beta1.Ingress{
		buildIngress(iNamespace("testing"),
			iRules(
				iRule(
					iHost("foo"),
					iPaths(onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80))))),
				iRule(
					iHost("bar"),
					iPaths(onePath(iBackend("service2", intstr.FromInt(802))))),
			),
		),
	}

	services := []*v1.Service{
		buildService(
			sName("service1"),
			sNamespace("testing"),
			sUID("1"),
			sAnnotation(label.TraefikBackendCircuitBreaker, "NetworkErrorRatio() > 0.5"),
			sAnnotation(label.TraefikBackendLoadBalancerMethod, "drr"),
			sSpec(
				clusterIP("10.0.0.1"),
				sPorts(sPort(80, ""))),
		),
		buildService(
			sName("service2"),
			sNamespace("testing"),
			sUID("2"),
			sAnnotation(label.TraefikBackendCircuitBreaker, ""),
			sAnnotation(label.TraefikBackendLoadBalancerSticky, "true"),
			sSpec(
				clusterIP("10.0.0.2"),
				sPorts(sPort(802, ""))),
		),
	}

	endpoints := []*v1.Endpoints{
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
		),
		frontends(
			frontend("foo/bar",
				headers(),
				passHostHeader(),
				routes(
					route("/bar", "PathPrefix:/bar"),
					route("foo", "Host:foo")),
			),
			frontend("bar",
				headers(),
				passHostHeader(),
				routes(route("bar", "Host:bar"))),
		),
	)

	assert.EqualValues(t, expected, actual)
}

func TestIngressAnnotations(t *testing.T) {
	ingresses := []*v1beta1.Ingress{
		buildIngress(
			iNamespace("testing"),
			iAnnotation(label.TraefikFrontendPassHostHeader, "false"),
			iRules(
				iRule(
					iHost("foo"),
					iPaths(onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
		buildIngress(
			iNamespace("testing"),
			iAnnotation(label.TraefikFrontendPassHostHeader, "true"),
			iAnnotation(annotationKubernetesIngressClass, "traefik"),
			iRules(
				iRule(
					iHost("other"),
					iPaths(onePath(iPath("/stuff"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
		buildIngress(
			iNamespace("testing"),
			iAnnotation(label.TraefikFrontendPassTLSCert, "true"),
			iAnnotation(annotationKubernetesIngressClass, "traefik"),
			iRules(
				iRule(
					iHost("other"),
					iPaths(onePath(iPath("/sslstuff"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
		buildIngress(
			iNamespace("testing"),
			iAnnotation(label.TraefikFrontendEntryPoints, "http,https"),
			iAnnotation(annotationKubernetesIngressClass, "traefik"),
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
			iAnnotation(annotationKubernetesIngressClass, "somethingOtherThanTraefik"),
			iRules(
				iRule(
					iHost("herp"),
					iPaths(onePath(iPath("/derp"), iBackend("service2", intstr.FromInt(80))))),
			),
		),
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesIngressClass, "traefik"),
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
			iAnnotation(annotationKubernetesIngressClass, "traefik"),
			iAnnotation(label.TraefikFrontendRedirectEntryPoint, "https"),
			iRules(
				iRule(
					iHost("redirect"),
					iPaths(onePath(iPath("/https"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
	}

	services := []*v1.Service{
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

	secrets := []*v1.Secret{{
		ObjectMeta: v1.ObjectMeta{
			Name:      "mySecret",
			UID:       "1",
			Namespace: "testing",
		},
		Data: map[string][]byte{"auth": []byte("myUser:myEncodedPW")},
	}}

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
		),
		frontends(
			frontend("foo/bar",
				headers(),
				routes(
					route("/bar", "PathPrefix:/bar"),
					route("foo", "Host:foo")),
			),
			frontend("other/stuff",
				headers(),
				passHostHeader(),
				routes(
					route("/stuff", "PathPrefix:/stuff"),
					route("other", "Host:other")),
			),
			frontend("other/",
				headers(),
				passHostHeader(),
				entryPoints("http", "https"),
				routes(
					route("/", "PathPrefix:/"),
					route("other", "Host:other")),
			),
			frontend("other/sslstuff",
				headers(),
				passHostHeader(),
				passTLSCert(),
				routes(
					route("/sslstuff", "PathPrefix:/sslstuff"),
					route("other", "Host:other")),
			),
			frontend("other/sslstuff",
				headers(),
				passHostHeader(),
				passTLSCert(),
				routes(
					route("/sslstuff", "PathPrefix:/sslstuff"),
					route("other", "Host:other")),
			),
			frontend("basic/auth",
				headers(),
				passHostHeader(),
				basicAuth("myUser:myEncodedPW"),
				routes(
					route("/auth", "PathPrefix:/auth"),
					route("basic", "Host:basic")),
			),
			frontend("redirect/https",
				headers(),
				passHostHeader(),
				redirectEntryPoint("https"),
				routes(
					route("/https", "PathPrefix:/https"),
					route("redirect", "Host:redirect")),
			),
			frontend("test/whitelist-source-range",
				headers(),
				passHostHeader(),
				whitelistSourceRange("1.1.1.1/24", "1234:abcd::42/32"),
				routes(
					route("/whitelist-source-range", "PathPrefix:/whitelist-source-range"),
					route("test", "Host:test")),
			),
			frontend("rewrite/api",
				headers(),
				passHostHeader(),
				routes(
					route("/api", "PathPrefix:/api;ReplacePath:/"),
					route("rewrite", "Host:rewrite")),
			),
		),
	)

	assert.Equal(t, expected, actual)
}

func TestPriorityHeaderValue(t *testing.T) {
	ingresses := []*v1beta1.Ingress{
		buildIngress(
			iNamespace("testing"),
			iAnnotation(label.TraefikFrontendPriority, "1337"),
			iRules(
				iRule(
					iHost("foo"),
					iPaths(onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
	}

	services := []*v1.Service{
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

	var endpoints []*v1.Endpoints
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
				headers(),
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
	ingresses := []*v1beta1.Ingress{
		buildIngress(
			iNamespace("testing"),
			iAnnotation(label.TraefikFrontendPassTLSCert, "herpderp"),
			iRules(
				iRule(
					iHost("foo"),
					iPaths(onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
	}

	services := []*v1.Service{
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
				headers(),
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
	ingresses := []*v1beta1.Ingress{
		buildIngress(
			iNamespace("testing"),
			iAnnotation(label.TraefikFrontendPassHostHeader, "herpderp"),
			iRules(
				iRule(
					iHost("foo"),
					iPaths(onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
	}

	services := []*v1.Service{
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
				headers(),
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
	ingresses := []*v1beta1.Ingress{
		buildIngress(
			iNamespace("testing"),
			iRules(
				iRule(
					iHost("foo"),
					iPaths(onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
	}

	services := []*v1.Service{
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
	ingresses := []*v1beta1.Ingress{
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

	services := []*v1.Service{
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

	endpoints := []*v1.Endpoints{
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
				headers(),
				passHostHeader(),
				routes(route("fully_working", "Host:fully_working")),
			),
			frontend("missing_endpoints",
				headers(),
				passHostHeader(),
				routes(route("missing_endpoints", "Host:missing_endpoints")),
			),
			frontend("missing_endpoint_subsets",
				headers(),
				passHostHeader(),
				routes(route("missing_endpoint_subsets", "Host:missing_endpoint_subsets")),
			),
		),
	)

	assert.Equal(t, expected, actual)
}

func TestBasicAuthInTemplate(t *testing.T) {
	ingresses := []*v1beta1.Ingress{
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

	services := []*v1.Service{
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

	secrets := []*v1.Secret{{
		ObjectMeta: v1.ObjectMeta{
			Name:      "mySecret",
			UID:       "1",
			Namespace: "testing",
		},
		Data: map[string][]byte{
			"auth": []byte("myUser:myEncodedPW"),
		},
	}}

	var endpoints []*v1.Endpoints
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
	ingresses := []*v1beta1.Ingress{
		buildIngress(
			iNamespace("testing"),
			iAnnotation(label.TraefikFrontendEntryPoints, "ep1,ep2"),
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
			iAnnotation(label.TraefikFrontendEntryPoints, "ep3"),
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
	services := []*v1.Service{
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
	secrets := []*v1.Secret{
		{
			ObjectMeta: v1.ObjectMeta{
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
				headers(),
				entryPoints("ep1", "ep2"),
				passHostHeader(),
				routes(
					route("example.com", "Host:example.com"),
				),
			),
			frontend("example.org",
				headers(),
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
		ingress   *v1beta1.Ingress
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
				secrets: []*v1.Secret{
					{
						ObjectMeta: v1.ObjectMeta{
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
				secrets: []*v1.Secret{
					{
						ObjectMeta: v1.ObjectMeta{
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
				secrets: []*v1.Secret{
					{
						ObjectMeta: v1.ObjectMeta{
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
				secrets: []*v1.Secret{
					{
						ObjectMeta: v1.ObjectMeta{
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
				iAnnotation(label.TraefikFrontendEntryPoints, "https,api-secure"),
				iRules(iRule(iHost("example.com"))),
				iTLSes(iTLS("test-secret")),
			),
			client: clientMock{
				secrets: []*v1.Secret{
					{
						ObjectMeta: v1.ObjectMeta{
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
