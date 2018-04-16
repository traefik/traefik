package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func buildEndpoint(opts ...func(*corev1.Endpoints)) *corev1.Endpoints {
	e := &corev1.Endpoints{}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func eNamespace(value string) func(*corev1.Endpoints) {
	return func(i *corev1.Endpoints) {
		i.Namespace = value
	}
}

func eName(value string) func(*corev1.Endpoints) {
	return func(i *corev1.Endpoints) {
		i.Name = value
	}
}

func eUID(value types.UID) func(*corev1.Endpoints) {
	return func(i *corev1.Endpoints) {
		i.UID = value
	}
}

func subset(opts ...func(*corev1.EndpointSubset)) func(*corev1.Endpoints) {
	return func(e *corev1.Endpoints) {
		s := &corev1.EndpointSubset{}
		for _, opt := range opts {
			opt(s)
		}
		e.Subsets = append(e.Subsets, *s)
	}
}

func eAddresses(opts ...func(*corev1.EndpointAddress)) func(*corev1.EndpointSubset) {
	return func(subset *corev1.EndpointSubset) {
		for _, opt := range opts {
			a := &corev1.EndpointAddress{}
			opt(a)
			subset.Addresses = append(subset.Addresses, *a)
		}
	}
}

func eAddress(ip string) func(*corev1.EndpointAddress) {
	return func(address *corev1.EndpointAddress) {
		address.IP = ip
	}
}

func ePorts(opts ...func(port *corev1.EndpointPort)) func(*corev1.EndpointSubset) {
	return func(spec *corev1.EndpointSubset) {
		for _, opt := range opts {
			p := &corev1.EndpointPort{}
			opt(p)
			spec.Ports = append(spec.Ports, *p)
		}
	}
}

func ePort(port int32, name string) func(*corev1.EndpointPort) {
	return func(sp *corev1.EndpointPort) {
		sp.Port = port
		sp.Name = name
	}
}

// Test

func TestBuildEndpoint(t *testing.T) {
	actual := buildEndpoint(
		eNamespace("testing"),
		eName("service3"),
		eUID("3"),
		subset(
			eAddresses(eAddress("10.15.0.1")),
			ePorts(
				ePort(8080, "http"),
				ePort(8443, "https"),
			),
		),
		subset(
			eAddresses(eAddress("10.15.0.2")),
			ePorts(
				ePort(9080, "http"),
				ePort(9443, "https"),
			),
		),
	)

	assert.EqualValues(t, sampleEndpoint1(), actual)
}

func sampleEndpoint1() *corev1.Endpoints {
	return &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "service3",
			UID:       "3",
			Namespace: "testing",
		},
		Subsets: []corev1.EndpointSubset{
			{
				Addresses: []corev1.EndpointAddress{
					{
						IP: "10.15.0.1",
					},
				},
				Ports: []corev1.EndpointPort{
					{
						Name: "http",
						Port: 8080,
					},
					{
						Name: "https",
						Port: 8443,
					},
				},
			},
			{
				Addresses: []corev1.EndpointAddress{
					{
						IP: "10.15.0.2",
					},
				},
				Ports: []corev1.EndpointPort{
					{
						Name: "http",
						Port: 9080,
					},
					{
						Name: "https",
						Port: 9443,
					},
				},
			},
		},
	}
}
