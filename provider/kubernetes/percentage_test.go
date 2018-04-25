package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestPercentageValueParse(t *testing.T) {
	testCases := []struct {
		parseString     string
		parseFloat64    float64
		shouldError     bool
		expectedString  string
		expectedFloat64 float64
	}{
		{
			parseString:     "1%",
			parseFloat64:    0.01,
			shouldError:     false,
			expectedString:  "1.000%",
			expectedFloat64: 0.01,
		},
		{
			parseString:     "0.5",
			parseFloat64:    0.5,
			shouldError:     false,
			expectedString:  "50.000%",
			expectedFloat64: 0.5,
		},
		{
			parseString:     "99%",
			parseFloat64:    0.99,
			shouldError:     false,
			expectedString:  "99.000%",
			expectedFloat64: 0.99,
		},
		{
			parseString:     "99.999%",
			parseFloat64:    0.99999,
			shouldError:     false,
			expectedString:  "99.999%",
			expectedFloat64: 0.99999,
		},
		{
			parseString:     "-99.999%",
			parseFloat64:    -0.99999,
			shouldError:     false,
			expectedString:  "-99.999%",
			expectedFloat64: -0.99999,
		},
		{
			parseString:     "-99.9990%",
			parseFloat64:    -0.99999,
			shouldError:     false,
			expectedString:  "-99.999%",
			expectedFloat64: -0.99999,
		},
		{
			parseString:     "0%",
			parseFloat64:    0,
			shouldError:     false,
			expectedString:  "0.000%",
			expectedFloat64: 0,
		},
		{
			parseString:     "%",
			parseFloat64:    0,
			shouldError:     true,
			expectedString:  "",
			expectedFloat64: 0,
		},
		{
			parseString:     "foo",
			parseFloat64:    0,
			shouldError:     true,
			expectedString:  "",
			expectedFloat64: 0,
		},
		{
			parseString:     "",
			parseFloat64:    0,
			shouldError:     true,
			expectedString:  "",
			expectedFloat64: 0,
		},
	}
	for _, testCase := range testCases {
		pvFromString, err := percentageValueFromString(testCase.parseString)
		pvFromFloat64 := percentageValueFromFloat64(testCase.parseFloat64)
		if testCase.shouldError {
			assert.Error(t, err, "expecting error but not happening")
			continue
		}
		assert.Equal(t, pvFromString, pvFromFloat64)
		assert.NoError(t, err, "fail to parse percentage value")
		assert.Equal(t, testCase.expectedString, pvFromFloat64.toString(), "percentage string value mismatched")
		assert.Equal(t, testCase.expectedFloat64, pvFromFloat64.toFloat64(), "percentage float64 value mismatched")
	}
}

func TestGetServicesPercentageWeights(t *testing.T) {

	testCases := []struct {
		name                              string
		servicePercentageWeightAnnotation string
		shouldErrorOnParseAnnotation      bool
		expectedWeightMap                 map[string]*percentageValue
	}{
		{
			name: "empty annotation",
			servicePercentageWeightAnnotation: ``,
			expectedWeightMap:                 map[string]*percentageValue{},
		},
		{
			name: "50% fraction",
			servicePercentageWeightAnnotation: `
service1: 10%
service2: 20%
service3: 20%
`,
			expectedWeightMap: map[string]*percentageValue{
				"service1": percentageValueFromFloat64(0.1),
				"service2": percentageValueFromFloat64(0.2),
				"service3": percentageValueFromFloat64(0.2),
			},
		},
		{
			name: "50% fraction float form",
			servicePercentageWeightAnnotation: `
service1: 0.1
service2: 0.2 
service3: 0.2
`,
			expectedWeightMap: map[string]*percentageValue{
				"service1": percentageValueFromFloat64(0.1),
				"service2": percentageValueFromFloat64(0.2),
				"service3": percentageValueFromFloat64(0.2),
			},
		},
		{
			name: "no fraction",
			servicePercentageWeightAnnotation: `
service1: 10%
service2: 90%
`,
			expectedWeightMap: map[string]*percentageValue{
				"service1": percentageValueFromFloat64(0.1),
				"service2": percentageValueFromFloat64(0.9),
			},
		},
		{
			name: "extra weight specification",
			servicePercentageWeightAnnotation: `
service1: 90%
service5: 90%
`,
			expectedWeightMap: map[string]*percentageValue{
				"service1": percentageValueFromFloat64(0.9),
				"service5": percentageValueFromFloat64(0.9),
			},
		},
		{
			name: "malformed annotation",
			servicePercentageWeightAnnotation: `
service1- 90%
service5- 90%
`,
			shouldErrorOnParseAnnotation: true,
			expectedWeightMap:            nil,
		},
		{
			name: "more than one hundred percentaged service",
			servicePercentageWeightAnnotation: `
service1: 100%
service2: 1%
`,
			expectedWeightMap: map[string]*percentageValue{
				"service1": percentageValueFromFloat64(1),
				"service2": percentageValueFromFloat64(0.01),
			},
		},
		{
			name: "incorrect percentage value",
			servicePercentageWeightAnnotation: `
service1: 1000%
`,
			expectedWeightMap: map[string]*percentageValue{
				"service1": percentageValueFromFloat64(10),
			},
		},
	}
	for _, testCase := range testCases {
		ing := &extensionsv1beta1.Ingress{}
		ing.Annotations = make(map[string]string)
		ing.Annotations[annotationKubernetesPercentageWeights] = testCase.servicePercentageWeightAnnotation
		weightMap, err := getServicesPercentageWeights(ing)
		if !assert.Equal(t, testCase.shouldErrorOnParseAnnotation, err != nil) {
			t.Error(err)
		}
		if !testCase.shouldErrorOnParseAnnotation {
			assert.Equal(t, testCase.expectedWeightMap, weightMap, "%s weight map mismatched", testCase.name)
		}
	}
}

func TestGetLeftFraction(t *testing.T) {
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
	buildPath := func(path string, f func(path *extensionsv1beta1.HTTPIngressPath)) extensionsv1beta1.HTTPIngressPath {
		pa := &extensionsv1beta1.HTTPIngressPath{}
		pa.Path = path
		f(pa)
		return *pa
	}
	testCases := []struct {
		name                        string
		weightMap                   map[string]*percentageValue
		ingressPaths                []extensionsv1beta1.HTTPIngressPath
		shouldError                 bool
		expectedLeftInstanceCount   map[string]int
		expectedLeftPercentageValue map[string]*percentageValue
	}{
		{
			name: "1 path 2 service",
			weightMap: map[string]*percentageValue{
				"service1": percentageValueFromFloat64(0.5),
			},
			ingressPaths: []extensionsv1beta1.HTTPIngressPath{
				buildPath("/foo", iBackend("service1", intstr.FromInt(8080))),
				buildPath("/foo", iBackend("service2", intstr.FromInt(8080))),
			},
			shouldError: false,
			expectedLeftInstanceCount: map[string]int{
				"/foo": 1,
			},
			expectedLeftPercentageValue: map[string]*percentageValue{
				"/foo": percentageValueFromFloat64(0.5),
			},
		},
		{
			name: "2 path 4 service",
			weightMap: map[string]*percentageValue{
				"service1": percentageValueFromFloat64(0.1),
				"service2": percentageValueFromFloat64(0.1),
			},
			ingressPaths: []extensionsv1beta1.HTTPIngressPath{
				buildPath("/foo", iBackend("service1", intstr.FromInt(8080))),
				buildPath("/foo", iBackend("service2", intstr.FromInt(8080))),
				buildPath("/bar", iBackend("service1", intstr.FromInt(8080))),
				buildPath("/bar", iBackend("service2", intstr.FromInt(8080))),
			},
			shouldError: false,
			expectedLeftInstanceCount: map[string]int{
				"/foo": 0,
				"/bar": 0,
			},
			expectedLeftPercentageValue: map[string]*percentageValue{
				"/foo": percentageValueFromFloat64(0.8),
				"/bar": percentageValueFromFloat64(0.8),
			},
		},
		{
			name: "2 path 8 service",
			weightMap: map[string]*percentageValue{
				"service1": percentageValueFromFloat64(0.1),
				"service2": percentageValueFromFloat64(0.1),
			},
			ingressPaths: []extensionsv1beta1.HTTPIngressPath{
				buildPath("/foo", iBackend("service1", intstr.FromInt(8080))),
				buildPath("/foo", iBackend("service2", intstr.FromInt(8080))),
				buildPath("/foo", iBackend("service3", intstr.FromInt(8080))),
				buildPath("/foo", iBackend("service4", intstr.FromInt(8080))),
				buildPath("/bar", iBackend("service1", intstr.FromInt(8080))),
				buildPath("/bar", iBackend("service2", intstr.FromInt(8080))),
				buildPath("/bar", iBackend("service3", intstr.FromInt(8080))),
				buildPath("/bar", iBackend("service4", intstr.FromInt(8080))),
			},
			shouldError: false,
			expectedLeftInstanceCount: map[string]int{
				"/foo": 5,
				"/bar": 5,
			},
			expectedLeftPercentageValue: map[string]*percentageValue{
				"/foo": percentageValueFromFloat64(0.8),
				"/bar": percentageValueFromFloat64(0.8),
			},
		},
		{
			name: "2 path 7 service",
			weightMap: map[string]*percentageValue{
				"service1": percentageValueFromFloat64(0.1),
			},
			ingressPaths: []extensionsv1beta1.HTTPIngressPath{
				buildPath("/foo", iBackend("service1", intstr.FromInt(8080))),
				buildPath("/foo", iBackend("service2", intstr.FromInt(8080))),
				buildPath("/foo", iBackend("service3", intstr.FromInt(8080))),
				buildPath("/foo", iBackend("service4", intstr.FromInt(8080))),
				buildPath("/bar", iBackend("service2", intstr.FromInt(8080))),
				buildPath("/bar", iBackend("service3", intstr.FromInt(8080))),
				buildPath("/bar", iBackend("service4", intstr.FromInt(8080))),
			},
			shouldError: false,
			expectedLeftInstanceCount: map[string]int{
				"/foo": 6,
				"/bar": 6,
			},
			expectedLeftPercentageValue: map[string]*percentageValue{
				"/foo": percentageValueFromFloat64(0.9),
				"/bar": percentageValueFromFloat64(1),
			},
		},
		{
			name: "2 path 4 service",
			weightMap: map[string]*percentageValue{
				"service1": percentageValueFromFloat64(0.1),
			},
			ingressPaths: []extensionsv1beta1.HTTPIngressPath{
				buildPath("/foo", iBackend("service1", intstr.FromInt(8080))),
				buildPath("/foo", iBackend("service2", intstr.FromInt(8080))),
				buildPath("/foo", iBackend("service3", intstr.FromInt(8080))),
				buildPath("/foo", iBackend("service4", intstr.FromInt(8080))),
				buildPath("/bar", iBackend("service2", intstr.FromInt(8080))),
				buildPath("/bar", iBackend("service3", intstr.FromInt(8080))),
				buildPath("/bar", iBackend("service4", intstr.FromInt(8080))),
			},
			shouldError: false,
			expectedLeftInstanceCount: map[string]int{
				"/foo": 6,
				"/bar": 6,
			},
			expectedLeftPercentageValue: map[string]*percentageValue{
				"/foo": percentageValueFromFloat64(0.9),
				"/bar": percentageValueFromFloat64(1),
			},
		},
		{
			name: "2 path no service",
			weightMap: map[string]*percentageValue{
				"service1": percentageValueFromFloat64(0.1),
			},
			ingressPaths: []extensionsv1beta1.HTTPIngressPath{
				buildPath("/foo", iBackend("noservice", intstr.FromInt(8080))),
				buildPath("/bar", iBackend("noservice", intstr.FromInt(8080))),
			},
			shouldError: false,
			expectedLeftInstanceCount: map[string]int{
				"/foo": 0,
				"/bar": 0,
			},
			expectedLeftPercentageValue: map[string]*percentageValue{
				"/foo": percentageValueFromFloat64(1),
				"/bar": percentageValueFromFloat64(1),
			},
		},
		{
			name:      "2 path without weight",
			weightMap: map[string]*percentageValue{},
			ingressPaths: []extensionsv1beta1.HTTPIngressPath{
				buildPath("/foo", iBackend("service1", intstr.FromInt(8080))),
				buildPath("/bar", iBackend("service2", intstr.FromInt(8080))),
			},
			shouldError: false,
			expectedLeftInstanceCount: map[string]int{
				"/foo": 2,
				"/bar": 1,
			},
			expectedLeftPercentageValue: map[string]*percentageValue{
				"/foo": percentageValueFromFloat64(1),
				"/bar": percentageValueFromFloat64(1),
			},
		},
		{
			name: "2 path overflow",
			weightMap: map[string]*percentageValue{
				"service1": percentageValueFromFloat64(0.7),
				"service2": percentageValueFromFloat64(0.8),
			},
			ingressPaths: []extensionsv1beta1.HTTPIngressPath{
				buildPath("/foo", iBackend("service1", intstr.FromInt(8080))),
				buildPath("/foo", iBackend("service2", intstr.FromInt(8080))),
				buildPath("/bar", iBackend("service2", intstr.FromInt(8080))),
			},
			shouldError:                 true,
			expectedLeftInstanceCount:   nil,
			expectedLeftPercentageValue: nil,
		},
	}
	for _, testCase := range testCases {
		leftPercentageMap, leftInstanceMap, err := getLeftFraction(client, "testing", testCase.ingressPaths, testCase.weightMap)
		assert.Equal(t, testCase.shouldError, err != nil)
		assert.Equal(t, testCase.expectedLeftPercentageValue, leftPercentageMap, "%s left percentage mismatch", testCase.name)
		assert.Equal(t, testCase.expectedLeftInstanceCount, leftInstanceMap, "%s left instance count mismatch", testCase.name)
	}

}
