package kubernetes

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

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
						onePath(iBackend("service2", intstr.FromInt(802))),
					),
				),
				iRule(iHost("service5"),
					iPaths(
						onePath(iBackend("service5", intstr.FromInt(8888))),
					),
				),
				iRule(iHost("service6"),
					iPaths(
						onePath(iBackend("service6", intstr.FromInt(80))),
					),
				),
				iRule(iHost("*.service7"),
					iPaths(
						onePath(iBackend("service7", intstr.FromInt(80))),
					),
				),
				iRule(iHost(""),
					iPaths(
						onePath(iBackend("service8", intstr.FromInt(80))),
					),
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
		buildService(
			sName("service5"),
			sNamespace("testing"),
			sUID("5"),
			sSpec(
				clusterIP("10.0.0.5"),
				sType("ExternalName"),
				sExternalName("example.com"),
				sPorts(sPort(8888, "http"))),
		),
		buildService(
			sName("service6"),
			sNamespace("testing"),
			sUID("6"),
			sSpec(
				clusterIP("10.0.0.6"),
				sPorts(sPort(80, ""))),
		),
		buildService(
			sName("service7"),
			sNamespace("testing"),
			sUID("7"),
			sSpec(
				clusterIP("10.0.0.7"),
				sPorts(sPort(80, ""))),
		),
		buildService(
			sName("service8"),
			sNamespace("testing"),
			sUID("8"),
			sSpec(
				clusterIP("10.0.0.8"),
				sPorts(sPort(80, ""))),
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
		buildEndpoint(
			eNamespace("testing"),
			eName("service6"),
			eUID("6"),
			subset(
				eAddresses(eAddressWithTargetRef("http://10.15.0.3:80", "10.15.0.3")),
				ePorts(ePort(80, ""))),
		),
		buildEndpoint(
			eNamespace("testing"),
			eName("service7"),
			eUID("7"),
			subset(
				eAddresses(eAddress("10.10.0.7")),
				ePorts(ePort(80, ""))),
		),
		buildEndpoint(
			eNamespace("testing"),
			eName("service8"),
			eUID("8"),
			subset(
				eAddresses(eAddress("10.10.0.8")),
				ePorts(ePort(80, ""))),
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
				servers(
					server("https://example.com", weight(1)),
				),
			),
			backend("bar",
				lbMethod("wrr"),
				servers(
					server("https://10.15.0.1:8443", weight(1)),
					server("https://10.15.0.2:9443", weight(1)),
				),
			),
			backend("service5",
				lbMethod("wrr"),
				servers(
					server("http://example.com:8888", weight(1)),
				),
			),
			backend("service6",
				lbMethod("wrr"),
				servers(
					server("http://10.15.0.3:80", weight(1)),
				),
			),
			backend("*.service7",
				lbMethod("wrr"),
				servers(
					server("http://10.10.0.7:80", weight(1)),
				),
			),
			backend("service8",
				lbMethod("wrr"),
				servers(
					server("http://10.10.0.8:80", weight(1)),
				),
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
			frontend("service5",
				passHostHeader(),
				routes(route("service5", "Host:service5")),
			),
			frontend("service6",
				passHostHeader(),
				routes(route("service6", "Host:service6")),
			),
			frontend("*.service7",
				passHostHeader(),
				routes(route("*.service7", "HostRegexp:{subdomain:[A-Za-z0-9-_]+}.service7")),
			),
			frontend("service8",
				passHostHeader(),
				routes(route("/", "PathPrefix:/")),
			),
		),
	)
	assert.Equal(t, expected, actual)
}

func TestLoadGlobalIngressWithPortNumbers(t *testing.T) {
	ingresses := []*extensionsv1beta1.Ingress{
		buildIngress(
			iNamespace("testing"),
			iSpecBackends(iSpecBackend(iIngressBackend("service1", intstr.FromInt(80)))),
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

	endpoints := []*corev1.Endpoints{
		buildEndpoint(
			eNamespace("testing"),
			eName("service1"),
			eUID("1"),
			subset(
				eAddresses(eAddress("10.10.0.1")),
				ePorts(ePort(8080, ""))),
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
			backend("global-default-backend",
				lbMethod("wrr"),
				servers(
					server("http://10.10.0.1:8080", weight(1)),
				),
			),
		),
		frontends(
			frontend("global-default-backend",
				frontendName("global-default-frontend"),
				passHostHeader(),
				routes(
					route("/", "PathPrefix:/"),
				),
			),
		),
	)
	assert.Equal(t, expected, actual)
}

func TestLoadGlobalIngressWithHttpsPortNames(t *testing.T) {
	ingresses := []*extensionsv1beta1.Ingress{
		buildIngress(
			iNamespace("testing"),
			iSpecBackends(iSpecBackend(iIngressBackend("service1", intstr.FromString("https-global")))),
		),
	}

	services := []*corev1.Service{
		buildService(
			sName("service1"),
			sNamespace("testing"),
			sUID("1"),
			sSpec(
				clusterIP("10.0.0.1"),
				sPorts(sPort(8443, "https-global"))),
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
			backend("global-default-backend",
				lbMethod("wrr"),
				servers(
					server("https://10.10.0.1:8080", weight(1)),
				),
			),
		),
		frontends(
			frontend("global-default-backend",
				frontendName("global-default-frontend"),
				passHostHeader(),
				routes(
					route("/", "PathPrefix:/"),
				),
			),
		),
	)
	assert.Equal(t, expected, actual)
}

func TestRuleType(t *testing.T) {
	testCases := []struct {
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
			desc:             "PathPrefixStrip rule type annotation set",
			ingressRuleType:  "PathPrefixStrip",
			frontendRuleType: "PathPrefixStrip",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			ingress := buildIngress(iRules(iRule(
				iHost("host"),
				iPaths(
					onePath(iPath("/path"), iBackend("service", intstr.FromInt(80))),
				),
			)))

			ingress.Annotations = map[string]string{
				annotationKubernetesRuleType: test.ingressRuleType,
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

func TestRuleFails(t *testing.T) {
	testCases := []struct {
		desc                      string
		ruletypeAnnotation        string
		requestModifierAnnotation string
	}{
		{
			desc:               "Rule-type using unknown rule",
			ruletypeAnnotation: "Foo: /bar",
		},
		{
			desc:               "Rule type full of spaces",
			ruletypeAnnotation: "  ",
		},
		{
			desc:               "Rule type missing both parts of rule",
			ruletypeAnnotation: "  :  ",
		},
		{
			desc:                      "Rule type combined with replacepath modifier",
			ruletypeAnnotation:        "ReplacePath",
			requestModifierAnnotation: "ReplacePath:/foo",
		},
		{
			desc:                      "Rule type combined with replacepathregex modifier",
			ruletypeAnnotation:        "ReplacePath",
			requestModifierAnnotation: "ReplacePathRegex:/foo /bar",
		},
		{
			desc:               "Empty Rule",
			ruletypeAnnotation: " : /bar",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			ingress := buildIngress(iRules(iRule(
				iHost("host"),
				iPaths(
					onePath(iPath("/path"), iBackend("service", intstr.FromInt(80))),
				),
			)))

			ingress.Annotations = map[string]string{
				annotationKubernetesRuleType:        test.ruletypeAnnotation,
				annotationKubernetesRequestModifier: test.requestModifierAnnotation,
			}

			_, err := getRuleForPath(extensionsv1beta1.HTTPIngressPath{Path: "/path"}, ingress)
			assert.Error(t, err)
		})
	}
}

func TestModifierType(t *testing.T) {
	testCases := []struct {
		desc                      string
		requestModifierAnnotation string
		expectedModifierRule      string
	}{
		{
			desc:                      "Request modifier annotation missing",
			requestModifierAnnotation: "",
			expectedModifierRule:      "",
		},
		{
			desc:                      "AddPrefix modifier annotation",
			requestModifierAnnotation: " AddPrefix: /foo",
			expectedModifierRule:      "AddPrefix:/foo",
		},
		{
			desc:                      "ReplacePath modifier annotation",
			requestModifierAnnotation: " ReplacePath: /foo",
			expectedModifierRule:      "ReplacePath:/foo",
		},
		{
			desc:                      "ReplacePathRegex modifier annotation",
			requestModifierAnnotation: " ReplacePathRegex: /foo /bar",
			expectedModifierRule:      "ReplacePathRegex:/foo /bar",
		},
		{
			desc:                      "AddPrefix modifier annotation",
			requestModifierAnnotation: "AddPrefix:/foo",
			expectedModifierRule:      "AddPrefix:/foo",
		},
		{
			desc:                      "ReplacePath modifier annotation",
			requestModifierAnnotation: "ReplacePath:/foo",
			expectedModifierRule:      "ReplacePath:/foo",
		},
		{
			desc:                      "ReplacePathRegex modifier annotation",
			requestModifierAnnotation: "ReplacePathRegex:/foo /bar",
			expectedModifierRule:      "ReplacePathRegex:/foo /bar",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			ingress := buildIngress(iRules(iRule(
				iHost("host"),
				iPaths(
					onePath(iPath("/path"), iBackend("service", intstr.FromInt(80))),
				),
			)))

			ingress.Annotations = map[string]string{
				annotationKubernetesRequestModifier: test.requestModifierAnnotation,
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

			expectedRules := []string{"PathPrefix:/path"}
			if len(test.expectedModifierRule) > 0 {
				expectedRules = append(expectedRules, test.expectedModifierRule)
			}

			expected := buildFrontends(frontend("host/path",
				routes(
					route("/path", strings.Join(expectedRules, ";")),
					route("host", "Host:host")),
			))

			assert.Equal(t, expected, actualConfig.Frontends)
		})
	}
}

func TestModifierFails(t *testing.T) {
	testCases := []struct {
		desc                      string
		requestModifierAnnotation string
	}{
		{
			desc:                      "Request modifier missing part of annotation",
			requestModifierAnnotation: "AddPrefix: ",
		},
		{
			desc:                      "Request modifier full of spaces annotation",
			requestModifierAnnotation: "    ",
		},
		{
			desc:                      "Request modifier missing both parts of annotation",
			requestModifierAnnotation: "  :  ",
		},
		{
			desc:                      "Request modifier using unknown rule",
			requestModifierAnnotation: "Foo: /bar",
		},
		{
			desc:                      "Missing Rule",
			requestModifierAnnotation: " : /bar",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			ingress := buildIngress(iRules(iRule(
				iHost("host"),
				iPaths(
					onePath(iPath("/path"), iBackend("service", intstr.FromInt(80))),
				),
			)))

			ingress.Annotations = map[string]string{
				annotationKubernetesRequestModifier: test.requestModifierAnnotation,
			}

			_, err := getRuleForPath(extensionsv1beta1.HTTPIngressPath{Path: "/path"}, ingress)
			assert.Error(t, err)
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

// Deprecated
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

func TestInvalidRedirectAnnotation(t *testing.T) {
	ingresses := []*extensionsv1beta1.Ingress{
		buildIngress(iNamespace("awesome"),
			iAnnotation(annotationKubernetesRedirectRegex, `bad\.regex`),
			iAnnotation(annotationKubernetesRedirectReplacement, "test"),
			iRules(iRule(
				iHost("foo"),
				iPaths(onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
		buildIngress(iNamespace("awesome"),
			iAnnotation(annotationKubernetesRedirectRegex, `test`),
			iAnnotation(annotationKubernetesRedirectReplacement, `bad\.replacement`),
			iRules(iRule(
				iHost("foo"),
				iPaths(onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
	}

	for _, ingress := range ingresses {
		actual := getFrontendRedirect(ingress, "test", "/")
		var expected *types.Redirect

		assert.Equal(t, expected, actual)
	}
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
				iRule(
					iHost("flush"),
					iPaths(onePath(iBackend("service5", intstr.FromInt(805))))),
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
			sAnnotation(annotationKubernetesAffinity, "true"),
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
				sPorts(sPort(803, "http"))),
		),
		buildService(
			sName("service4"),
			sNamespace("testing"),
			sUID("4"),
			sAnnotation(annotationKubernetesMaxConnExtractorFunc, "client.ip"),
			sAnnotation(annotationKubernetesMaxConnAmount, "6"),
			sSpec(
				clusterIP("10.0.0.4"),
				sPorts(sPort(804, "http"))),
		),
		buildService(
			sName("service5"),
			sNamespace("testing"),
			sUID("5"),
			sAnnotation(annotationKubernetesResponseForwardingFlushInterval, "10ms"),
			sSpec(
				clusterIP("10.0.0.5"),
				sPorts(sPort(80, ""))),
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
				ePorts(ePort(8080, ""))),
			subset(
				eAddresses(eAddress("10.15.0.2")),
				ePorts(ePort(8080, ""))),
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
		buildEndpoint(
			eNamespace("testing"),
			eName("service5"),
			eUID("5"),
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
			backend("flush",
				servers(),
				lbMethod("wrr"),
				responseForwarding("10ms"),
			),
			backend("bar",
				servers(
					server("http://10.15.0.1:8080", weight(1)),
					server("http://10.15.0.2:8080", weight(1))),
				lbMethod("wrr"), lbStickiness(),
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
			frontend("flush",
				passHostHeader(),
				routes(
					route("flush", "Host:flush"))),
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
			iAnnotation(annotationKubernetesPassTLSClientCert, `
pem: true
infos:
  notafter: true
  notbefore: true
  sans: true
  subject:
    country: true
    province: true
    locality: true
    organization: true
    commonname: true
    serialnumber: true
`),
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
			iAnnotation(annotationKubernetesWhiteListSourceRange, "1.1.1.1/24, 1234:abcd::42/32"),
			iAnnotation(annotationKubernetesWhiteListIPStrategyExcludedIPs, "1.1.1.1/24, 1234:abcd::42/32"),
			iAnnotation(annotationKubernetesWhiteListIPStrategyDepth, "5"),
			iRules(
				iRule(
					iHost("test"),
					iPaths(onePath(iPath("/whitelist-source-range"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesWhiteListSourceRange, "1.1.1.1/24, 1234:abcd::42/32"),
			iAnnotation(annotationKubernetesWhiteListIPStrategy, "true"),
			iRules(
				iRule(
					iHost("test"),
					iPaths(onePath(iPath("/whitelist-remote-addr"), iBackend("service1", intstr.FromInt(80))))),
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
			iAnnotation(annotationKubernetesAppRoot, "/root"),
			iRules(
				iRule(
					iHost("root"),
					iPaths(
						onePath(iPath("/"), iBackend("service1", intstr.FromInt(80))),
						onePath(iPath("/root1"), iBackend("service1", intstr.FromInt(80))),
					),
				),
			),
		),
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesAppRoot, "/root2"),
			iAnnotation(annotationKubernetesRewriteTarget, "/abc"),
			iRules(
				iRule(
					iHost("root2"),
					iPaths(
						onePath(iPath("/"), iBackend("service2", intstr.FromInt(80))),
					),
				),
			),
		),
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesRuleType, ruleTypeReplacePath),
			iAnnotation(annotationKubernetesRewriteTarget, "/abc"),
			iRules(
				iRule(
					iHost("root2"),
					iPaths(
						onePath(iPath("/"), iBackend("service2", intstr.FromInt(80))),
					),
				),
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
			iAnnotation(annotationKubernetesSSLForceHost, "true"),
			iAnnotation(annotationKubernetesSSLRedirect, "true"),
			iAnnotation(annotationKubernetesSSLTemporaryRedirect, "true"),
			iAnnotation(annotationKubernetesHSTSIncludeSubdomains, "true"),
			iAnnotation(annotationKubernetesForceHSTSHeader, "true"),
			iAnnotation(annotationKubernetesHSTSPreload, "true"),
			iAnnotation(annotationKubernetesFrameDeny, "true"),
			iAnnotation(annotationKubernetesContentTypeNosniff, "true"),
			iAnnotation(annotationKubernetesBrowserXSSFilter, "true"),
			iAnnotation(annotationKubernetesCustomBrowserXSSValue, "foo"),
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
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesProtocol, "h2c"),
			iRules(
				iRule(
					iHost("protocol"),
					iPaths(onePath(iPath("/valid"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesProtocol, "foobar"),
			iRules(
				iRule(
					iHost("protocol"),
					iPaths(onePath(iPath("/notvalid"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesProtocol, "http"),
			iRules(
				iRule(
					iHost("protocol"),
					iPaths(onePath(iPath("/missmatch"), iBackend("serviceHTTPS", intstr.FromInt(443))))),
			),
		),
		buildIngress(
			iNamespace("testing"),
			iRules(
				iRule(
					iHost("protocol"),
					iPaths(onePath(iPath("/noAnnotation"), iBackend("serviceHTTPS", intstr.FromInt(443))))),
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
		buildService(
			sName("serviceHTTPS"),
			sNamespace("testing"),
			sUID("2"),
			sSpec(
				clusterIP("10.0.0.3"),
				sType("ExternalName"),
				sExternalName("example.com"),
				sPorts(sPort(443, "https"))),
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
			backend("test/whitelist-remote-addr",
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
			backend("root/",
				servers(
					server("http://example.com", weight(1))),
				lbMethod("wrr"),
			),
			backend("root/root1",
				servers(
					server("http://example.com", weight(1))),
				lbMethod("wrr"),
			),
			backend("root2/",
				servers(),
				lbMethod("wrr"),
			),
			backend("protocol/valid",
				servers(
					server("h2c://example.com", weight(1)),
					server("h2c://example.com", weight(1))),
				lbMethod("wrr"),
			),
			backend("protocol/notvalid",
				servers(),
				lbMethod("wrr"),
			),
			backend("protocol/missmatch",
				servers(
					server("http://example.com", weight(1)),
					server("http://example.com", weight(1))),
				lbMethod("wrr"),
			),
			backend("protocol/noAnnotation",
				servers(
					server("https://example.com", weight(1)),
					server("https://example.com", weight(1))),
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
				passTLSClientCert(),
				passTLSCert(),
				routes(
					route("/sslstuff", "PathPrefix:/sslstuff"),
					route("other", "Host:other")),
			),
			frontend("basic/auth",
				passHostHeader(),
				basicAuthDeprecated("myUser:myEncodedPW"),
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
				whiteList(
					whiteListRange("1.1.1.1/24", "1234:abcd::42/32"),
					whiteListIPStrategy(5, "1.1.1.1/24", "1234:abcd::42/32")),
				routes(
					route("/whitelist-source-range", "PathPrefix:/whitelist-source-range"),
					route("test", "Host:test")),
			),
			frontend("test/whitelist-remote-addr",
				passHostHeader(),
				whiteList(
					whiteListRange("1.1.1.1/24", "1234:abcd::42/32"),
					whiteListIPStrategy(0)),
				routes(
					route("/whitelist-remote-addr", "PathPrefix:/whitelist-remote-addr"),
					route("test", "Host:test")),
			),
			frontend("rewrite/api",
				passHostHeader(),
				routes(
					route("/api", "PathPrefix:/api;ReplacePathRegex: ^/api(.*) $1"),
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
					SSLForceHost:            true,
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
					CustomBrowserXSSValue:   "foo",
				}),
				routes(
					route("/customheaders", "PathPrefix:/customheaders"),
					route("custom-headers", "Host:custom-headers")),
			),
			frontend("root/",
				passHostHeader(),
				redirectRegex("root/$", "root/root"),
				routes(
					route("/", "PathPrefix:/"),
					route("root", "Host:root"),
				),
			),
			frontend("root2/",
				passHostHeader(),
				redirectRegex("root2/$", "root2/root2"),
				routes(
					route("/", "PathPrefix:/;ReplacePathRegex: ^/(.*) /abc$1"),
					route("root2", "Host:root2"),
				),
			),
			frontend("root/root1",
				passHostHeader(),
				routes(
					route("/root1", "PathPrefix:/root1"),
					route("root", "Host:root"),
				),
			),
			frontend("protocol/valid",
				passHostHeader(),
				routes(
					route("/valid", "PathPrefix:/valid"),
					route("protocol", "Host:protocol"),
				),
			),
			frontend("protocol/notvalid",
				passHostHeader(),
				routes(
					route("/notvalid", "PathPrefix:/notvalid"),
					route("protocol", "Host:protocol"),
				),
			),
			frontend("protocol/missmatch",
				passHostHeader(),
				routes(
					route("/missmatch", "PathPrefix:/missmatch"),
					route("protocol", "Host:protocol"),
				),
			),
			frontend("protocol/noAnnotation",
				passHostHeader(),
				routes(
					route("/noAnnotation", "PathPrefix:/noAnnotation"),
					route("protocol", "Host:protocol"),
				),
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
					iHost("foo"),
					iPaths(onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesIngressClass, "custom"),
			iRules(
				iRule(
					iHost("foo"),
					iPaths(onePath(iPath("/bar"), iBackend("service2", intstr.FromInt(80))))),
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
				sPorts(sPort(80, "http"))),
		),
	}

	endpoints := []*corev1.Endpoints{
		buildEndpoint(
			eName("service2"),
			eUID("1"),
			eNamespace("testing"),
			subset(
				eAddresses(eAddress("10.10.0.1")),
				ePorts(ePort(80, "http"))),
		),
	}

	watchChan := make(chan interface{})
	client := clientMock{
		ingresses: ingresses,
		services:  services,
		endpoints: endpoints,
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
					backend("foo/bar",
						servers(
							server("http://example.com", weight(1))),
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
			),
		},
		{
			desc:     "Provided IngressClass annotation",
			provider: Provider{IngressClass: "custom"},
			expected: buildConfiguration(
				backends(
					backend("foo/bar",
						servers(
							server("http://10.10.0.1:80", weight(1))),
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
			backend("1337-foo/bar",
				servers(server("http://example.com", weight(1))),
				lbMethod("wrr"),
			),
		),
		frontends(
			frontend("1337-foo/bar",
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

			if _, err := provider.loadIngresses(client); err != nil {
				if client.apiServiceError != nil {
					assert.EqualError(t, err, "failed kube api call")
				}
				if client.apiEndpointsError != nil {
					assert.EqualError(t, err, "failed kube api call")
				}
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

func TestLoadIngressesBasicAuth(t *testing.T) {
	ingresses := []*extensionsv1beta1.Ingress{
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesAuthType, "basic"),
			iAnnotation(annotationKubernetesAuthSecret, "mySecret"),
			iAnnotation(annotationKubernetesAuthRemoveHeader, "true"),
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
	actualBasicAuth := actual.Frontends["basic/auth"].Auth.Basic
	assert.Equal(t, types.Users{"myUser:myEncodedPW"}, actualBasicAuth.Users)
	assert.True(t, actualBasicAuth.RemoveHeader, "Bad RemoveHeader flag")
}

func TestLoadIngressesForwardAuth(t *testing.T) {
	ingresses := []*extensionsv1beta1.Ingress{
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesAuthType, "forward"),
			iAnnotation(annotationKubernetesAuthForwardURL, "https://auth.host"),
			iAnnotation(annotationKubernetesAuthForwardTrustHeaders, "true"),
			iAnnotation(annotationKubernetesAuthForwardResponseHeaders, "X-Auth,X-Test,X-Secret"),
			iRules(
				iRule(iHost("foo"),
					iPaths(
						onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80))))),
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

	endpoints := []*corev1.Endpoints{
		buildEndpoint(
			eNamespace("testing"),
			eName("service1"),
			eUID("1"),
			subset(
				eAddresses(eAddress("10.10.0.1")),
				ePorts(ePort(8080, ""))),
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
					server("http://10.10.0.1:8080", weight(1))),
			),
		),
		frontends(
			frontend("foo/bar",
				passHostHeader(),
				auth(forwardAuth("https://auth.host",
					fwdTrustForwardHeader(),
					fwdAuthResponseHeaders("X-Auth", "X-Test", "X-Secret"))),
				routes(
					route("/bar", "PathPrefix:/bar"),
					route("foo", "Host:foo")),
			),
		),
	)

	assert.Equal(t, expected, actual)
}

func TestLoadIngressesForwardAuthMissingURL(t *testing.T) {
	ingresses := []*extensionsv1beta1.Ingress{
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesAuthType, "forward"),
			iRules(
				iRule(iHost("foo"),
					iPaths(
						onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80))))),
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

	endpoints := []*corev1.Endpoints{
		buildEndpoint(
			eNamespace("testing"),
			eName("service1"),
			eUID("1"),
			subset(
				eAddresses(eAddress("10.10.0.1")),
				ePorts(ePort(8080, ""))),
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
				servers(),
			),
		),
		frontends(),
	)

	assert.Equal(t, expected, actual)
}

func TestLoadIngressesForwardAuthWithTLSSecret(t *testing.T) {
	ingresses := []*extensionsv1beta1.Ingress{
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesAuthType, "forward"),
			iAnnotation(annotationKubernetesAuthForwardURL, "https://auth.host"),
			iAnnotation(annotationKubernetesAuthForwardTLSSecret, "secret"),
			iAnnotation(annotationKubernetesAuthForwardTLSInsecure, "true"),
			iRules(
				iRule(iHost("foo"),
					iPaths(
						onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
	}

	secrets := []*corev1.Secret{{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret",
			UID:       "1",
			Namespace: "testing",
		},
		Data: map[string][]byte{
			"tls.crt": []byte("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
			"tls.key": []byte("-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----"),
		},
	}}

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

	endpoints := []*corev1.Endpoints{
		buildEndpoint(
			eNamespace("testing"),
			eName("service1"),
			eUID("1"),
			subset(
				eAddresses(eAddress("10.10.0.1")),
				ePorts(ePort(8080, ""))),
		),
	}

	watchChan := make(chan interface{})
	client := clientMock{
		ingresses: ingresses,
		services:  services,
		endpoints: endpoints,
		secrets:   secrets,
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
					server("http://10.10.0.1:8080", weight(1))),
			),
		),
		frontends(
			frontend("foo/bar",
				passHostHeader(),
				auth(
					forwardAuth("https://auth.host",
						fwdAuthTLS(
							"-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----",
							"-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----",
							true))),
				routes(
					route("/bar", "PathPrefix:/bar"),
					route("foo", "Host:foo")),
			),
		),
	)

	assert.Equal(t, expected, actual)
}

func TestLoadIngressesForwardAuthWithTLSSecretFailures(t *testing.T) {
	testCases := []struct {
		desc       string
		secretName string
		certName   string
		certData   string
		keyName    string
		keyData    string
	}{
		{
			desc:       "empty certificate and key",
			secretName: "secret",
			certName:   "",
			certData:   "",
			keyName:    "",
			keyData:    "",
		},
		{
			desc:       "wrong secret name, empty certificate and key",
			secretName: "wrongSecret",
			certName:   "",
			certData:   "",
			keyName:    "",
			keyData:    "",
		},
		{
			desc:       "empty certificate data",
			secretName: "secret",
			certName:   "tls.crt",
			certData:   "",
			keyName:    "tls.key",
			keyData:    "-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----",
		},
		{
			desc:       "empty key data",
			secretName: "secret",
			certName:   "tls.crt",
			certData:   "-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----",
			keyName:    "tls.key",
			keyData:    "",
		},
		{
			desc:       "wrong cert name",
			secretName: "secret",
			certName:   "cert.crt",
			certData:   "-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE----",
			keyName:    "tls.key",
			keyData:    "-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----",
		},
		{
			desc:       "wrong key name",
			secretName: "secret",
			certName:   "tls.crt",
			certData:   "-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----",
			keyName:    "cert.key",
			keyData:    "-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----",
		},
	}

	ingresses := []*extensionsv1beta1.Ingress{
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesAuthType, "forward"),
			iAnnotation(annotationKubernetesAuthForwardURL, "https://auth.host"),
			iAnnotation(annotationKubernetesAuthForwardTLSSecret, "secret"),
			iRules(
				iRule(iHost("foo"),
					iPaths(
						onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80))))),
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

	endpoints := []*corev1.Endpoints{
		buildEndpoint(
			eNamespace("testing"),
			eName("service1"),
			eUID("1"),
			subset(
				eAddresses(eAddress("10.10.0.1")),
				ePorts(ePort(8080, ""))),
		),
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			secrets := []*corev1.Secret{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      test.secretName,
					UID:       "1",
					Namespace: "testing",
				},
				Data: map[string][]byte{
					test.certName: []byte(test.certData),
					test.keyName:  []byte(test.keyData),
				},
			}}

			watchChan := make(chan interface{})
			client := clientMock{
				ingresses: ingresses,
				services:  services,
				endpoints: endpoints,
				secrets:   secrets,
				watchChan: watchChan,
			}
			provider := Provider{}

			actual, err := provider.loadIngresses(client)
			require.NoError(t, err, "error loading ingresses")

			expected := buildConfiguration(
				backends(
					backend("foo/bar",
						lbMethod("wrr"),
						servers(),
					),
				),
				frontends(),
			)

			assert.Equal(t, expected, actual)
		})

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

	testCases := []struct {
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

	for _, test := range testCases {
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

func TestMultiPortServices(t *testing.T) {
	ingresses := []*extensionsv1beta1.Ingress{
		buildIngress(
			iNamespace("testing"),
			iRules(
				iRule(iPaths(
					onePath(iPath("/cheddar"), iBackend("service", intstr.FromString("cheddar"))),
					onePath(iPath("/stilton"), iBackend("service", intstr.FromString("stilton"))),
				)),
			),
		),
	}

	services := []*corev1.Service{
		buildService(
			sName("service"),
			sNamespace("testing"),
			sUID("1"),
			sSpec(
				clusterIP("10.0.0.1"),
				sPorts(sPort(80, "cheddar")),
				sPorts(sPort(81, "stilton")),
			),
		),
	}

	endpoints := []*corev1.Endpoints{
		buildEndpoint(
			eNamespace("testing"),
			eName("service"),
			eUID("1"),
			subset(
				eAddresses(
					eAddress("10.10.0.1"),
					eAddress("10.10.0.2"),
				),
				ePorts(ePort(8080, "cheddar")),
			),
			subset(
				eAddresses(
					eAddress("10.20.0.1"),
					eAddress("10.20.0.2"),
				),
				ePorts(ePort(8081, "stilton")),
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
			backend("/cheddar",
				lbMethod("wrr"),
				servers(
					server("http://10.10.0.1:8080", weight(1)),
					server("http://10.10.0.2:8080", weight(1)),
				),
			),
			backend("/stilton",
				lbMethod("wrr"),
				servers(
					server("http://10.20.0.1:8081", weight(1)),
					server("http://10.20.0.2:8081", weight(1)),
				),
			),
		),
		frontends(
			frontend("/cheddar",
				passHostHeader(),
				routes(route("/cheddar", "PathPrefix:/cheddar")),
			),
			frontend("/stilton",
				passHostHeader(),
				routes(route("/stilton", "PathPrefix:/stilton")),
			),
		),
	)

	assert.Equal(t, expected, actual)
}

func TestProviderUpdateIngressStatus(t *testing.T) {
	testCases := []struct {
		desc                  string
		ingressEndpoint       *IngressEndpoint
		apiServiceError       error
		apiIngressStatusError error
		expectedError         bool
	}{
		{
			desc:          "without IngressEndpoint configuration",
			expectedError: false,
		},
		{
			desc:            "without any IngressEndpoint option",
			ingressEndpoint: &IngressEndpoint{},
			expectedError:   true,
		},
		{
			desc: "PublishedService - invalid format",
			ingressEndpoint: &IngressEndpoint{
				PublishedService: "foo",
			},
			expectedError: true,
		},
		{
			desc: "PublishedService - missing service",
			ingressEndpoint: &IngressEndpoint{
				PublishedService: "foo/bar",
			},
			expectedError: true,
		},
		{
			desc: "PublishedService - get service error",
			ingressEndpoint: &IngressEndpoint{
				PublishedService: "foo/bar",
			},
			apiServiceError: errors.New("error"),
			expectedError:   true,
		},
		{
			desc: "PublishedService - Skipping empty LoadBalancerIngress",
			ingressEndpoint: &IngressEndpoint{
				PublishedService: "testing/service-empty-status",
			},
			expectedError: false,
		},
		{
			desc: "PublishedService - with update error",
			ingressEndpoint: &IngressEndpoint{
				PublishedService: "testing/service",
			},
			apiIngressStatusError: errors.New("error"),
			expectedError:         true,
		},
		{
			desc: "PublishedService - right service",
			ingressEndpoint: &IngressEndpoint{
				PublishedService: "testing/service",
			},
			expectedError: false,
		},
		{
			desc: "IP - valid",
			ingressEndpoint: &IngressEndpoint{
				IP: "127.0.0.1",
			},
			expectedError: false,
		},
		{
			desc: "IP - with update error",
			ingressEndpoint: &IngressEndpoint{
				IP: "127.0.0.1",
			},
			apiIngressStatusError: errors.New("error"),
			expectedError:         true,
		},
		{
			desc: "hostname - valid",
			ingressEndpoint: &IngressEndpoint{
				Hostname: "foo",
			},
			expectedError: false,
		},
		{
			desc: "hostname - with update error",
			ingressEndpoint: &IngressEndpoint{
				Hostname: "foo",
			},
			apiIngressStatusError: errors.New("error"),
			expectedError:         true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := &Provider{
				IngressEndpoint: test.ingressEndpoint,
			}

			services := []*corev1.Service{
				buildService(
					sName("service-empty-status"),
					sNamespace("testing"),
					sLoadBalancerStatus(),
					sUID("1"),
					sSpec(
						clusterIP("10.0.0.1"),
						sPorts(sPort(80, ""))),
				),
				buildService(
					sName("service"),
					sNamespace("testing"),
					sLoadBalancerStatus(sLoadBalancerIngress("127.0.0.1", "")),
					sUID("2"),
					sSpec(
						clusterIP("10.0.0.2"),
						sPorts(sPort(80, ""))),
				),
			}

			client := clientMock{
				services:              services,
				apiServiceError:       test.apiServiceError,
				apiIngressStatusError: test.apiIngressStatusError,
			}

			i := &extensionsv1beta1.Ingress{}

			err := p.updateIngressStatus(i, client)

			if test.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPercentageWeightServiceAnnotation(t *testing.T) {
	ingresses := []*extensionsv1beta1.Ingress{
		buildIngress(
			iAnnotation(annotationKubernetesServiceWeights, `
service1: 10%
`),
			iNamespace("testing"),
			iRules(
				iRule(
					iHost("host1"),
					iPaths(
						onePath(iPath("/foo"), iBackend("service1", intstr.FromString("8080"))),
						onePath(iPath("/foo"), iBackend("service2", intstr.FromString("7070"))),
						onePath(iPath("/bar"), iBackend("service2", intstr.FromString("7070"))),
					)),
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
				sPorts(sPort(8080, "")),
			),
		),
		buildService(
			sName("service2"),
			sNamespace("testing"),
			sUID("1"),
			sSpec(
				clusterIP("10.0.0.1"),
				sPorts(sPort(7070, "")),
			),
		),
	}

	endpoints := []*corev1.Endpoints{
		buildEndpoint(
			eNamespace("testing"),
			eName("service1"),
			eUID("1"),
			subset(
				eAddresses(
					eAddress("10.10.0.1"),
					eAddress("10.10.0.2"),
				),
				ePorts(ePort(8080, "")),
			),
		),
		buildEndpoint(
			eNamespace("testing"),
			eName("service2"),
			eUID("1"),
			subset(
				eAddresses(
					eAddress("10.10.0.3"),
					eAddress("10.10.0.4"),
				),
				ePorts(ePort(7070, "")),
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
			backend("host1/foo",
				servers(
					server("http://10.10.0.1:8080", weight(int(newPercentageValueFromFloat64(0.05)))),
					server("http://10.10.0.2:8080", weight(int(newPercentageValueFromFloat64(0.05)))),
					server("http://10.10.0.3:7070", weight(int(newPercentageValueFromFloat64(0.45)))),
					server("http://10.10.0.4:7070", weight(int(newPercentageValueFromFloat64(0.45)))),
				),
				lbMethod("wrr"),
			),
			backend("host1/bar",
				servers(
					server("http://10.10.0.3:7070", weight(int(newPercentageValueFromFloat64(0.5)))),
					server("http://10.10.0.4:7070", weight(int(newPercentageValueFromFloat64(0.5)))),
				),
				lbMethod("wrr"),
			),
		),
		frontends(
			frontend("host1/bar",
				passHostHeader(),
				routes(
					route("/bar", "PathPrefix:/bar"),
					route("host1", "Host:host1")),
			),
			frontend("host1/foo",
				passHostHeader(),
				routes(
					route("/foo", "PathPrefix:/foo"),
					route("host1", "Host:host1")),
			),
		),
	)

	assert.Equal(t, expected, actual, "error loading percentage weight annotation")
}

func TestProviderNewK8sInClusterClient(t *testing.T) {
	p := Provider{}
	os.Setenv("KUBERNETES_SERVICE_HOST", "localhost")
	os.Setenv("KUBERNETES_SERVICE_PORT", "443")
	defer os.Clearenv()
	_, err := p.newK8sClient("")
	assert.EqualError(t, err, "failed to create in-cluster configuration: open /var/run/secrets/kubernetes.io/serviceaccount/token: no such file or directory")
}

func TestProviderNewK8sInClusterClientFailLabelSel(t *testing.T) {
	p := Provider{}
	os.Setenv("KUBERNETES_SERVICE_HOST", "localhost")
	os.Setenv("KUBERNETES_SERVICE_PORT", "443")
	defer os.Clearenv()
	_, err := p.newK8sClient("%")
	assert.EqualError(t, err, "invalid ingress label selector: \"%\"")
}

func TestProviderNewK8sOutOfClusterClient(t *testing.T) {
	p := Provider{}
	p.Endpoint = "localhost"
	_, err := p.newK8sClient("")
	assert.NoError(t, err)
}

func TestAddGlobalBackendDuplicateFailures(t *testing.T) {
	testCases := []struct {
		desc           string
		previousConfig *types.Configuration
		err            string
	}{
		{
			desc: "Duplicate Frontend",
			previousConfig: buildConfiguration(
				frontends(
					frontend("global-default-backend",
						frontendName("global-default-frontend"),
						passHostHeader(),
						routes(
							route("/", "PathPrefix:/"),
						),
					),
				),
			),
			err: "duplicate frontend: global-default-frontend",
		},
		{
			desc: "Duplicate Backend",
			previousConfig: buildConfiguration(
				backends(
					backend("global-default-backend",
						lbMethod("wrr"),
						servers(
							server("http://10.10.0.1:8080", weight(1)),
						),
					),
				)),
			err: "duplicate backend: global-default-backend",
		},
	}

	ingress := buildIngress(
		iNamespace("testing"),
		iSpecBackends(iSpecBackend(iIngressBackend("service1", intstr.FromInt(80)))),
	)

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			watchChan := make(chan interface{})
			client := clientMock{
				watchChan: watchChan,
			}
			provider := Provider{}

			err := provider.addGlobalBackend(client, ingress, test.previousConfig)
			assert.EqualError(t, err, test.err)
		})
	}
}

func TestAddGlobalBackendServiceMissing(t *testing.T) {
	ingresses := buildIngress(
		iNamespace("testing"),
		iSpecBackends(iSpecBackend(iIngressBackend("service1", intstr.FromInt(80)))),
	)

	config := buildConfiguration(
		frontends(),
		backends(),
	)
	watchChan := make(chan interface{})
	client := clientMock{
		watchChan: watchChan,
	}
	provider := Provider{}

	err := provider.addGlobalBackend(client, ingresses, config)
	assert.Error(t, err)
}

func TestAddGlobalBackendServiceAPIError(t *testing.T) {
	ingresses := buildIngress(
		iNamespace("testing"),
		iSpecBackends(iSpecBackend(iIngressBackend("service1", intstr.FromInt(80)))),
	)

	config := buildConfiguration(
		frontends(),
		backends(),
	)

	apiErr := errors.New("failed kube api call")

	watchChan := make(chan interface{})
	client := clientMock{
		apiServiceError: apiErr,
		watchChan:       watchChan,
	}
	provider := Provider{}
	err := provider.addGlobalBackend(client, ingresses, config)
	assert.Error(t, err)
}

func TestAddGlobalBackendEndpointMissing(t *testing.T) {
	ingresses := buildIngress(
		iNamespace("testing"),
		iSpecBackends(iSpecBackend(iIngressBackend("service", intstr.FromInt(80)))),
	)

	services := []*corev1.Service{
		buildService(
			sName("service"),
			sNamespace("testing"),
			sUID("1"),
			sSpec(
				clusterIP("10.0.0.1"),
				sPorts(sPort(80, "")),
			),
		),
	}

	config := buildConfiguration(
		frontends(),
		backends(),
	)
	watchChan := make(chan interface{})
	client := clientMock{
		services:  services,
		watchChan: watchChan,
	}
	provider := Provider{}

	err := provider.addGlobalBackend(client, ingresses, config)
	assert.Error(t, err)
}

func TestAddGlobalBackendEndpointAPIError(t *testing.T) {
	ingresses := buildIngress(
		iNamespace("testing"),
		iSpecBackends(iSpecBackend(iIngressBackend("service", intstr.FromInt(80)))),
	)

	config := buildConfiguration(
		frontends(),
		backends(),
	)

	services := []*corev1.Service{
		buildService(
			sName("service"),
			sNamespace("testing"),
			sUID("1"),
			sSpec(
				clusterIP("10.0.0.1"),
				sPorts(sPort(80, "")),
			),
		),
	}

	apiErr := errors.New("failed kube api call")

	watchChan := make(chan interface{})
	client := clientMock{
		apiEndpointsError: apiErr,
		services:          services,
		watchChan:         watchChan,
	}
	provider := Provider{}
	err := provider.addGlobalBackend(client, ingresses, config)
	assert.Error(t, err)
}

func TestTemplateBreakingIngresssValues(t *testing.T) {
	ingresses := []*extensionsv1beta1.Ingress{
		buildIngress(
			iNamespace("testing"),
			iAnnotation(annotationKubernetesIngressClass, "testing-\"foo\""),
			iRules(
				iRule(
					iHost("foo"),
					iPaths(onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
		buildIngress(
			iNamespace("testing"),
			iRules(
				iRule(
					iHost("testing-\"foo\""),
					iPaths(onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(80))))),
			),
		),
		buildIngress(
			iNamespace("testing"),
			iRules(
				iRule(
					iHost("foo"),
					iPaths(onePath(iPath("/testing-\"foo\""), iBackend("service1", intstr.FromInt(80))))),
			),
		),
	}

	client := clientMock{
		ingresses: ingresses,
	}
	provider := Provider{}

	actual, err := provider.loadIngresses(client)
	require.NoError(t, err, "error loading ingresses")

	expected := buildConfiguration(
		backends(),
		frontends(),
	)

	assert.Equal(t, expected, actual)
}

func TestDivergingIngressDefinitions(t *testing.T) {
	ingresses := []*extensionsv1beta1.Ingress{
		buildIngress(
			iNamespace("testing"),
			iRules(
				iRule(
					iHost("host-a"),
					iPaths(
						onePath(iBackend("service1", intstr.FromString("80"))),
					)),
			),
		),
		buildIngress(
			iNamespace("testing"),
			iRules(
				iRule(
					iHost("host-a"),
					iPaths(
						onePath(iBackend("missing", intstr.FromString("80"))),
					)),
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
				sPorts(sPort(80, "http")),
			),
		),
	}

	endpoints := []*corev1.Endpoints{
		buildEndpoint(
			eNamespace("testing"),
			eName("service1"),
			eUID("1"),
			subset(
				eAddresses(
					eAddress("10.10.0.1"),
				),
				ePorts(ePort(80, "http")),
			),
			subset(
				eAddresses(
					eAddress("10.10.0.2"),
				),
				ePorts(ePort(80, "http")),
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
			backend("host-a",
				servers(
					server("http://10.10.0.1:80", weight(1)),
					server("http://10.10.0.2:80", weight(1)),
				),
				lbMethod("wrr"),
			),
		),
		frontends(
			frontend("host-a",
				passHostHeader(),
				routes(
					route("host-a", "Host:host-a")),
			),
		),
	)

	assert.Equal(t, expected, actual, "error merging multiple backends")
}
