package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/types"
)

func buildService(opts ...func(*v1.Service)) *v1.Service {
	s := &v1.Service{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func sNamespace(value string) func(*v1.Service) {
	return func(i *v1.Service) {
		i.Namespace = value
	}
}

func sName(value string) func(*v1.Service) {
	return func(i *v1.Service) {
		i.Name = value
	}
}

func sUID(value types.UID) func(*v1.Service) {
	return func(i *v1.Service) {
		i.UID = value
	}
}

func sAnnotation(name string, value string) func(*v1.Service) {
	return func(s *v1.Service) {
		if s.Annotations == nil {
			s.Annotations = make(map[string]string)
		}
		s.Annotations[name] = value
	}
}

func sSpec(opts ...func(*v1.ServiceSpec)) func(*v1.Service) {
	return func(i *v1.Service) {
		spec := &v1.ServiceSpec{}
		for _, opt := range opts {
			opt(spec)
		}
		i.Spec = *spec
	}
}

func clusterIP(ip string) func(*v1.ServiceSpec) {
	return func(spec *v1.ServiceSpec) {
		spec.ClusterIP = ip
	}
}

func sType(value v1.ServiceType) func(*v1.ServiceSpec) {
	return func(spec *v1.ServiceSpec) {
		spec.Type = value
	}
}

func sExternalName(name string) func(*v1.ServiceSpec) {
	return func(spec *v1.ServiceSpec) {
		spec.ExternalName = name
	}
}

func sPorts(opts ...func(*v1.ServicePort)) func(*v1.ServiceSpec) {
	return func(spec *v1.ServiceSpec) {
		for _, opt := range opts {
			p := &v1.ServicePort{}
			opt(p)
			spec.Ports = append(spec.Ports, *p)
		}
	}
}

func sPort(port int32, name string) func(*v1.ServicePort) {
	return func(sp *v1.ServicePort) {
		sp.Port = port
		sp.Name = name
	}
}

// Test

func TestBuildService(t *testing.T) {
	actual1 := buildService(
		sName("service1"),
		sNamespace("testing"),
		sUID("1"),
		sSpec(
			clusterIP("10.0.0.1"),
			sPorts(sPort(80, "")),
		),
	)

	assert.EqualValues(t, sampleService1(), actual1)

	actual2 := buildService(
		sName("service3"),
		sNamespace("testing"),
		sUID("3"),
		sSpec(
			clusterIP("10.0.0.3"),
			sType("ExternalName"),
			sExternalName("example.com"),
			sPorts(
				sPort(80, "http"),
				sPort(443, "https"),
			),
		),
	)

	assert.EqualValues(t, sampleService2(), actual2)
}

func sampleService1() *v1.Service {
	return &v1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:      "service1",
			UID:       "1",
			Namespace: "testing",
		},
		Spec: v1.ServiceSpec{
			ClusterIP: "10.0.0.1",
			Ports: []v1.ServicePort{
				{
					Port: 80,
				},
			},
		},
	}
}

func sampleService2() *v1.Service {
	return &v1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:      "service3",
			UID:       "3",
			Namespace: "testing",
		},
		Spec: v1.ServiceSpec{
			ClusterIP:    "10.0.0.3",
			Type:         "ExternalName",
			ExternalName: "example.com",
			Ports: []v1.ServicePort{
				{
					Name: "http",
					Port: 80,
				},
				{
					Name: "https",
					Port: 443,
				},
			},
		},
	}
}
