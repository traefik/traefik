package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/types"
)

func buildEndpoint(opts ...func(*v1.Endpoints)) *v1.Endpoints {
	e := &v1.Endpoints{}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func eNamespace(value string) func(*v1.Endpoints) {
	return func(i *v1.Endpoints) {
		i.Namespace = value
	}
}

func eName(value string) func(*v1.Endpoints) {
	return func(i *v1.Endpoints) {
		i.Name = value
	}
}

func eUID(value types.UID) func(*v1.Endpoints) {
	return func(i *v1.Endpoints) {
		i.UID = value
	}
}

func subset(opts ...func(*v1.EndpointSubset)) func(*v1.Endpoints) {
	return func(e *v1.Endpoints) {
		s := &v1.EndpointSubset{}
		for _, opt := range opts {
			opt(s)
		}
		e.Subsets = append(e.Subsets, *s)
	}
}

func eAddresses(opts ...func(*v1.EndpointAddress)) func(*v1.EndpointSubset) {
	return func(subset *v1.EndpointSubset) {
		a := &v1.EndpointAddress{}
		for _, opt := range opts {
			opt(a)
		}
		subset.Addresses = append(subset.Addresses, *a)
	}
}

func eAddress(ip string) func(*v1.EndpointAddress) {
	return func(address *v1.EndpointAddress) {
		address.IP = ip
	}
}

func ePorts(opts ...func(port *v1.EndpointPort)) func(*v1.EndpointSubset) {
	return func(spec *v1.EndpointSubset) {
		for _, opt := range opts {
			p := &v1.EndpointPort{}
			opt(p)
			spec.Ports = append(spec.Ports, *p)
		}
	}
}

func ePort(port int32, name string) func(*v1.EndpointPort) {
	return func(sp *v1.EndpointPort) {
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

func sampleEndpoint1() *v1.Endpoints {
	return &v1.Endpoints{
		ObjectMeta: v1.ObjectMeta{
			Name:      "service3",
			UID:       "3",
			Namespace: "testing",
		},
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{
						IP: "10.15.0.1",
					},
				},
				Ports: []v1.EndpointPort{
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
				Addresses: []v1.EndpointAddress{
					{
						IP: "10.15.0.2",
					},
				},
				Ports: []v1.EndpointPort{
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
