package kubernetes

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestString(t *testing.T) {
	pv1 := newPercentageValueFromFloat64(0.5)
	pv2 := newPercentageValueFromFloat64(0.2)
	pv3 := newPercentageValueFromFloat64(0.3)
	f := fractionalWeightAllocator(
		map[ingressService]int{
			{
				host:    "host2",
				path:    "path2",
				service: "service2",
			}: int(pv2),
			{
				host:    "host3",
				path:    "path3",
				service: "service3",
			}: int(pv3),
			{
				host:    "host1",
				path:    "path1",
				service: "service1",
			}: int(pv1),
		},
	)

	expected := fmt.Sprintf("[service1: %s service2: %s service3: %s]", pv1, pv2, pv3)
	actual := f.String()
	assert.Equal(t, expected, actual)
}

func TestGetServicesPercentageWeights(t *testing.T) {
	testCases := []struct {
		desc            string
		annotationValue string
		expectError     bool
		expectedWeights map[string]percentageValue
	}{
		{
			desc:            "empty annotation",
			annotationValue: ``,
			expectedWeights: map[string]percentageValue{},
		},
		{
			desc: "50% fraction",
			annotationValue: `
service1: 10%
service2: 20%
service3: 20%
`,
			expectedWeights: map[string]percentageValue{
				"service1": newPercentageValueFromFloat64(0.1),
				"service2": newPercentageValueFromFloat64(0.2),
				"service3": newPercentageValueFromFloat64(0.2),
			},
		},
		{
			desc: "50% fraction with empty fraction",
			annotationValue: `
service1: 10%
service2: 20%
service3: 20%
service4:
`,
			expectError: true,
		},
		{
			desc: "50% fraction float form",
			annotationValue: `
service1: 0.1
service2: 0.2 
service3: 0.2
`,
			expectedWeights: map[string]percentageValue{
				"service1": newPercentageValueFromFloat64(0.001),
				"service2": newPercentageValueFromFloat64(0.002),
				"service3": newPercentageValueFromFloat64(0.002),
			},
		},
		{
			desc: "no fraction",
			annotationValue: `
service1: 10%
service2: 90%
`,
			expectedWeights: map[string]percentageValue{
				"service1": newPercentageValueFromFloat64(0.1),
				"service2": newPercentageValueFromFloat64(0.9),
			},
		},
		{
			desc: "extra weight specification",
			annotationValue: `
service1: 90%
service5: 90%
`,
			expectedWeights: map[string]percentageValue{
				"service1": newPercentageValueFromFloat64(0.9),
				"service5": newPercentageValueFromFloat64(0.9),
			},
		},
		{
			desc: "malformed annotation",
			annotationValue: `
service1- 90%
service5- 90%
`,
			expectError:     true,
			expectedWeights: nil,
		},
		{
			desc: "more than one hundred percentaged service",
			annotationValue: `
service1: 100%
service2: 1%
`,
			expectedWeights: map[string]percentageValue{
				"service1": newPercentageValueFromFloat64(1),
				"service2": newPercentageValueFromFloat64(0.01),
			},
		},
		{
			desc: "incorrect percentage value",
			annotationValue: `
service1: 1000%
`,
			expectedWeights: map[string]percentageValue{
				"service1": newPercentageValueFromFloat64(10),
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			ingress := &extensionsv1beta1.Ingress{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						annotationKubernetesServiceWeights: test.annotationValue,
					},
				},
			}

			weights, err := getServicesPercentageWeights(ingress)

			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectedWeights, weights)
			}
		})
	}
}

func TestComputeServiceWeights(t *testing.T) {
	client := clientMock{
		endpoints: []*corev1.Endpoints{
			buildEndpoint(
				eNamespace("testing"),
				eName("service1"),
				eUID("1"),
				subset(
					eAddresses(eAddress("10.10.0.1")),
					ePorts(ePort(8080, ""))),
				subset(
					eAddresses(eAddress("10.21.0.2")),
					ePorts(ePort(8080, ""))),
			),
			buildEndpoint(
				eNamespace("testing"),
				eName("service2"),
				eUID("2"),
				subset(
					eAddresses(eAddress("10.10.0.3")),
					ePorts(ePort(8080, ""))),
			),
			buildEndpoint(
				eNamespace("testing"),
				eName("service3"),
				eUID("3"),
				subset(
					eAddresses(eAddress("10.10.0.4")),
					ePorts(ePort(8080, ""))),
				subset(
					eAddresses(eAddress("10.21.0.5")),
					ePorts(ePort(8080, ""))),
				subset(
					eAddresses(eAddress("10.21.0.6")),
					ePorts(ePort(8080, ""))),
				subset(
					eAddresses(eAddress("10.21.0.7")),
					ePorts(ePort(8080, ""))),
			),
			buildEndpoint(
				eNamespace("testing"),
				eName("service4"),
				eUID("4"),
				subset(
					eAddresses(eAddress("10.10.0.7")),
					ePorts(ePort(8080, ""))),
			),
		},
	}

	testCases := []struct {
		desc            string
		ingress         *extensionsv1beta1.Ingress
		expectError     bool
		expectedWeights map[ingressService]percentageValue
	}{
		{
			desc: "1 path 2 service",
			ingress: buildIngress(
				iNamespace("testing"),
				iAnnotation(annotationKubernetesServiceWeights, `
service1: 10%
`),
				iRules(
					iRule(iHost("foo.test"), iPaths(
						onePath(iPath("/foo"), iBackend("service1", intstr.FromInt(8080))),
						onePath(iPath("/foo"), iBackend("service2", intstr.FromInt(8080))),
					)),
				),
			),
			expectError: false,
			expectedWeights: map[ingressService]percentageValue{
				{
					host:    "foo.test",
					path:    "/foo",
					service: "service1",
				}: newPercentageValueFromFloat64(0.05),
				{
					host:    "foo.test",
					path:    "/foo",
					service: "service2",
				}: newPercentageValueFromFloat64(0.90),
			},
		},
		{
			desc: "2 path 2 service",
			ingress: buildIngress(
				iNamespace("testing"),
				iAnnotation(annotationKubernetesServiceWeights, `
service1: 60%
`),
				iRules(
					iRule(iHost("foo.test"), iPaths(
						onePath(iPath("/foo"), iBackend("service1", intstr.FromInt(8080))),
						onePath(iPath("/foo"), iBackend("service2", intstr.FromInt(8080))),
						onePath(iPath("/bar"), iBackend("service1", intstr.FromInt(8080))),
						onePath(iPath("/bar"), iBackend("service3", intstr.FromInt(8080))),
					)),
				),
			),
			expectError: false,
			expectedWeights: map[ingressService]percentageValue{
				{
					host:    "foo.test",
					path:    "/foo",
					service: "service1",
				}: newPercentageValueFromFloat64(0.30),
				{
					host:    "foo.test",
					path:    "/foo",
					service: "service2",
				}: newPercentageValueFromFloat64(0.40),
				{
					host:    "foo.test",
					path:    "/bar",
					service: "service1",
				}: newPercentageValueFromFloat64(0.30),
				{
					host:    "foo.test",
					path:    "/bar",
					service: "service3",
				}: newPercentageValueFromFloat64(0.10),
			},
		},
		{
			desc: "2 path 3 service",
			ingress: buildIngress(
				iNamespace("testing"),
				iAnnotation(annotationKubernetesServiceWeights, `
service1: 20%
service3: 20%
`),
				iRules(
					iRule(iHost("foo.test"), iPaths(
						onePath(iPath("/foo"), iBackend("service1", intstr.FromInt(8080))),
						onePath(iPath("/foo"), iBackend("service2", intstr.FromInt(8080))),
						onePath(iPath("/bar"), iBackend("service2", intstr.FromInt(8080))),
						onePath(iPath("/bar"), iBackend("service3", intstr.FromInt(8080))),
					)),
				),
			),
			expectError: false,
			expectedWeights: map[ingressService]percentageValue{
				{
					host:    "foo.test",
					path:    "/foo",
					service: "service1",
				}: newPercentageValueFromFloat64(0.10),
				{
					host:    "foo.test",
					path:    "/foo",
					service: "service2",
				}: newPercentageValueFromFloat64(0.80),
				{
					host:    "foo.test",
					path:    "/bar",
					service: "service3",
				}: newPercentageValueFromFloat64(0.05),
				{
					host:    "foo.test",
					path:    "/bar",
					service: "service2",
				}: newPercentageValueFromFloat64(0.80),
			},
		},
		{
			desc: "1 path 4 service",
			ingress: buildIngress(
				iNamespace("testing"),
				iAnnotation(annotationKubernetesServiceWeights, `
service1: 20%
service2: 40%
service3: 40%
`),
				iRules(
					iRule(iHost("foo.test"), iPaths(
						onePath(iPath("/foo"), iBackend("service1", intstr.FromInt(8080))),
						onePath(iPath("/foo"), iBackend("service2", intstr.FromInt(8080))),
						onePath(iPath("/foo"), iBackend("service3", intstr.FromInt(8080))),
						onePath(iPath("/foo"), iBackend("service4", intstr.FromInt(8080))),
					)),
				),
			),
			expectError: false,
			expectedWeights: map[ingressService]percentageValue{
				{
					host:    "foo.test",
					path:    "/foo",
					service: "service1",
				}: newPercentageValueFromFloat64(0.10),
				{
					host:    "foo.test",
					path:    "/foo",
					service: "service2",
				}: newPercentageValueFromFloat64(0.40),
				{
					host:    "foo.test",
					path:    "/foo",
					service: "service3",
				}: newPercentageValueFromFloat64(0.10),
				{
					host:    "foo.test",
					path:    "/foo",
					service: "service4",
				}: newPercentageValueFromFloat64(0.00),
			},
		},
		{
			desc: "2 path no service",
			ingress: buildIngress(
				iNamespace("testing"),
				iAnnotation(annotationKubernetesServiceWeights, `
service1: 20%
service2: 40%
service3: 40%
`),
				iRules(
					iRule(iHost("foo.test"), iPaths(
						onePath(iPath("/foo"), iBackend("noservice", intstr.FromInt(8080))),
						onePath(iPath("/bar"), iBackend("noservice", intstr.FromInt(8080))),
					)),
				),
			),
			expectError: true,
		},
		{
			desc: "2 path without weight",
			ingress: buildIngress(
				iNamespace("testing"),
				iAnnotation(annotationKubernetesServiceWeights, ``),
				iRules(
					iRule(iHost("foo.test"), iPaths(
						onePath(iPath("/foo"), iBackend("service1", intstr.FromInt(8080))),
						onePath(iPath("/bar"), iBackend("service2", intstr.FromInt(8080))),
					)),
				),
			),
			expectError: false,
			expectedWeights: map[ingressService]percentageValue{
				{
					host:    "foo.test",
					path:    "/foo",
					service: "service1",
				}: newPercentageValueFromFloat64(0.50),
				{
					host:    "foo.test",
					path:    "/bar",
					service: "service2",
				}: newPercentageValueFromFloat64(1.00),
			},
		},
		{
			desc: "2 path overflow",
			ingress: buildIngress(
				iNamespace("testing"),
				iAnnotation(annotationKubernetesServiceWeights, `
service1: 70%
service2: 80%
`),
				iRules(
					iRule(iHost("foo.test"), iPaths(
						onePath(iPath("/foo"), iBackend("service1", intstr.FromInt(8080))),
						onePath(iPath("/foo"), iBackend("service2", intstr.FromInt(8080))),
					)),
				),
			),
			expectError: true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			weightAllocator, err := newFractionalWeightAllocator(test.ingress, client)
			if test.expectError {
				require.Error(t, err)
			} else {
				for ingSvc, percentage := range test.expectedWeights {
					assert.Equal(t, int(percentage), weightAllocator.getWeight(ingSvc.host, ingSvc.path, ingSvc.service))
				}
			}
		})
	}
}
