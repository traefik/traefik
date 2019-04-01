package kubernetes

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
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

			ingress := buildIngress(
				iAnnotation(annotationKubernetesServiceWeights, test.annotationValue),
			)

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
	client := newClientMock(
		filepath.Join("fixtures", "computeServiceWeights_endpoints.yml"),
		filepath.Join("fixtures", "computeServiceWeights_services.yml"),
	)

	testCases := []struct {
		desc            string
		ingress         string
		expectError     bool
		expectedWeights map[ingressService]percentageValue
	}{
		{
			desc:        "1 path 2 service",
			ingress:     filepath.Join("fixtures", "computeServiceWeights", "1_path_2_service_ingresses.yml"),
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
			desc:        "2 path 2 service",
			ingress:     filepath.Join("fixtures", "computeServiceWeights", "2_path_2_service_ingresses.yml"),
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
			desc:        "2 path 3 service",
			ingress:     filepath.Join("fixtures", "computeServiceWeights", "2_path_3_service_ingresses.yml"),
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
			desc:        "1 path 4 service",
			ingress:     filepath.Join("fixtures", "computeServiceWeights", "1_path_4_service_ingresses.yml"),
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
			desc:        "2 path no service",
			ingress:     filepath.Join("fixtures", "computeServiceWeights", "2_path_no_service_ingresses.yml"),
			expectError: true,
		},
		{
			desc:        "2 path without weight",
			ingress:     filepath.Join("fixtures", "computeServiceWeights", "2_path_without_weight_ingresses.yml"),
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
			desc:        "2 path overflow",
			ingress:     filepath.Join("fixtures", "computeServiceWeights", "2_path_overflow_ingresses.yml"),
			expectError: true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			yamlContent, err := ioutil.ReadFile(test.ingress)
			require.NoError(t, err)
			k8sObjects := MustDecodeYaml(yamlContent)
			require.Len(t, k8sObjects, 1)

			var ingress *extensionsv1beta1.Ingress
			switch o := k8sObjects[0].(type) {
			case *extensionsv1beta1.Ingress:
				ingress = o
			default:
				require.Fail(t, fmt.Sprintf("Unknown runtime object %+v %T", o, o))
			}

			weightAllocator, err := newFractionalWeightAllocator(ingress, client)
			if test.expectError {
				require.Error(t, err)
			} else {
				if err != nil {
					t.Errorf("%v failed: %v", test.desc, err)
				} else {
					for ingSvc, percentage := range test.expectedWeights {
						assert.Equal(t, int(percentage), weightAllocator.getWeight(ingSvc.host, ingSvc.path, ingSvc.service))
					}
				}
			}
		})
	}
}
